// Точка входа. Инициализирует все зависимости и запускает сервер
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogazub/myapp/internal/api"
	"github.com/gogazub/myapp/internal/consumer"
	repo "github.com/gogazub/myapp/internal/repository"
	svc "github.com/gogazub/myapp/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func main() {
	service, err := createService()
	if err != nil {
		log.Printf("create service error:%v", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := startConsumer(service); err != nil {
			errCh <- err
		}
	}()
	go func() {
		if err := startServer(service); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, fmt.Errorf("stop consumer")) || errors.Is(err, fmt.Errorf("stop server")) {
			break
		}
		log.Printf("application launch err:%v", err)
		os.Exit(1)

	// Можно добавить healthcheck на consumer и service
	// И по готовности вывести сообщение об успешном запуске
	default:
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
	}

}

// startConsumer запускает Kafka consumer и передает данные в сервис
func startConsumer(service svc.IService) error {
	config := consumer.Config{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders",
		GroupID:  "order-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}
	kafkaCfg := kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  config.GroupID,
		MinBytes: config.MinBytes,
		MaxBytes: config.MaxBytes,
		MaxWait:  1 * time.Second,
	}
	reader := kafka.NewReader(kafkaCfg)
	kafkaConsumer := consumer.NewConsumer(service, reader)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Можно добавить контекст для цепочки ошибок
	return kafkaConsumer.Start(ctx)
}

// startServer запускает HTTP сервер, который обслуживает запросы по order_id
func startServer(service svc.IService) error {
	srv := api.NewServer(service)

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("starting server error: %w", err)
	}

	address := os.Getenv("SERVER_PORT")
	if address == "" {
		return fmt.Errorf("address not specified")
	}

	// Можно попробовать добавить healthcheck сервера и выводить сообщение о запуске в main,
	//  а не здесь, потому что startServer не должен заниматься выводом в терминал.
	log.Printf("Starting HTTP server :%s...", address)
	if err := srv.Start(address); err != nil {
		return fmt.Errorf("starting server error: %w", err)
	}
	return nil
}

// Создает подключение к БД
func connectToDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("connect to db error: %w", err)
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
		return nil, fmt.Errorf("connect to db error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connect to db error: %w", err)
	}

	return db, nil
}

// createService инициализирует репозитории и сервис для обработки заказов
func createService() (*svc.Service, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, fmt.Errorf("create service error:%w", err)
	}

	psqlRepo := repo.NewOrderRepository(db)
	cacheRepo := repo.NewCacheRepository()

	err = cacheRepo.LoadFromDB(psqlRepo)
	if err != nil {
		return nil, fmt.Errorf("create service error:%w", err)
	}

	service := svc.NewService(psqlRepo, cacheRepo)
	return service, nil
}
