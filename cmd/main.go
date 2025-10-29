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
	"sync"
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

var (
	ErrStopConsumer = errors.New("stop consumer")
	ErrStopServer   = errors.New("stop server")
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: no .env loaded: %v", err)
	}

	service, err := createService()
	if err != nil {
		log.Printf("starting app error: %v", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startConsumer(rootCtx, service); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startServer(rootCtx, service); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("component error: %v - initiating shutdown", err)
		cancel()
	case sig := <-sigCh:
		log.Printf("signal %v received - initiating shutdown", sig)
		cancel()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	timeout := 40 * time.Second
	select {
	case <-done:
		log.Println("all components stopped gracefully")
	case <-time.After(timeout):
		log.Printf("timeout (%s) waiting for components to stop, exiting", timeout)
	}

	log.Println("shutdown complete")

}

// startConsumer запускает Kafka consumer и передает данные в сервис
func startConsumer(ctx context.Context, service svc.IService) error {
	broker := os.Getenv("KAFKA_BROKER")
	config := consumer.Config{
		Brokers:  []string{broker},
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
	// Можно добавить контекст для цепочки ошибок
	return kafkaConsumer.Start(ctx)
}

// startServer запускает HTTP сервер, который обслуживает запросы по order_id
func startServer(ctx context.Context, service svc.IService) error {
	srv := api.NewServer(service)

	address := ":" + os.Getenv("SERVER_PORT")
	if address == "" {
		return fmt.Errorf("address not specified")
	}

	// Можно попробовать добавить healthcheck сервера и выводить сообщение о запуске в main,
	//  а не здесь, потому что startServer не должен заниматься выводом в терминал.
	log.Printf("Starting HTTP server %s...", address)
	if err := srv.Start(ctx, address); err != nil {
		return fmt.Errorf("starting server error: %w", err)
	}
	return nil
}

// Создает подключение к БД
func connectToDB() (*sql.DB, error) {

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
