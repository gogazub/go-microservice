package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogazub/myapp/internal/orders"
	"github.com/gogazub/myapp/internal/server/consumer"
)

func main() {
	cfg := consumer.Config{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders",
		GroupID:  "orders-consumer-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}

	handler := consumer.NewFuncHandler(handleOrder)
	kafkaConsumer := consumer.NewConsumer(cfg, handler)
	defer kafkaConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go kafkaConsumer.Start(ctx)

	log.Println("Consumer started. Press Ctrl+C to stop.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down consumer...")
}

func handleOrder(ctx context.Context, order *orders.ModelOrder) error {
	log.Printf("Handling order: %s", order.OrderUID)
	log.Printf("Track number: %s", order.TrackNumber)
	log.Printf("Customer: %s", order.Delivery.Name)
	log.Printf("Items count: %d", len(order.Items))

	time.Sleep(100 * time.Millisecond)
	return nil
}
