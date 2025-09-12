package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gogazub/myapp/internal/orders"
	"github.com/gogazub/myapp/internal/server"
	"github.com/gogazub/myapp/internal/server/consumer"
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
func startConsumer(service orders.Service) {
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
func startServer(service orders.Service) error {
	srv := server.NewServer(service)

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Error loading .env file: %w", err)
	}

	address := os.Getenv("SERVER_PORT")
	if address == "" {
		return fmt.Errorf("SERVER_PORT not set in .env file")
	}

	log.Printf("Starting HTTP server on %s...", address)
	if err := srv.Start(address); err != nil {
		log.Printf("Error starting HTTP server: %v", err)
		return err
	}

	return nil
}

// Создает подключение к БД
func connectToDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file: %w", err)
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
func createService() (orders.Service, error) {
	db, err := connectToDB()
	if err != nil {
		return orders.Service{}, err
	}

	psqlRepo := orders.NewOrderRepository(db)
	cacheRepo := orders.NewCacheRepository(psqlRepo)

	service := orders.NewService(psqlRepo, cacheRepo)
	return *service, nil
}
