package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gogazub/myapp/internal/api"
	"github.com/gogazub/myapp/internal/consumer"
	repo "github.com/gogazub/myapp/internal/repository"
	"github.com/gogazub/myapp/internal/service"
	svc "github.com/gogazub/myapp/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	service, err := createService()
	if err != nil {
		log.Fatalf("Error creating service: %v", err)
	}

	go startConsumer(service)
	go startServer(service)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down gracefully...")
}

// startConsumer запускает Kafka consumer и передает данные в сервис
func startConsumer(service svc.IService) {
	consumerConfig := consumer.Config{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders",
		GroupID:  "order-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}

	kafkaConsumer := consumer.NewConsumer(consumerConfig, service)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumer.Start(ctx)
}

// startServer запускает HTTP сервер, который обслуживает запросы по order_id
func startServer(service svc.IService) {
	srv := api.NewServer(service)

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %w", err)
	}

	address := os.Getenv("SERVER_PORT")
	if address == "" {
		log.Fatalf("SERVER_PORT not set in .env file")
	}

	log.Printf("Starting HTTP server on %s...", address)
	if err := srv.Start(address); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
		return
	}
}

// Создает подключение к БД
func connectToDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// createService инициализирует репозитории и сервис для обработки заказов
func createService() (*svc.Service, error) {
	db, err := connectToDB()
	if err != nil {
		return &svc.Service{}, err
	}

	psqlRepo := repo.NewOrderRepository(db)
	cacheRepo := repo.NewCacheRepository()
	cacheRepo.LoadFromDB(psqlRepo)

	service := service.NewService(psqlRepo, cacheRepo)
	return service, nil
}
