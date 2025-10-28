// Producer - это вспомогательный модуль, который генерирует по заданным настройкам order`ы и отпарвляет в кафку
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"order-service/producer/model"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	brokers := flag.String("brokers", "localhost:9092", "comma-separated list of kafka brokers")
	topic := flag.String("topic", "orders", "kafka topic")
	groupID := flag.String("group", "order-group", "consumer group id")
	count := flag.Int("n", 10, "how many messages to send (0 - infinite)")
	interval := flag.Duration("interval", time.Second, "interval between messages")
	flag.Parse()

	brokerList := []string{}
	for _, b := range splitAndTrim(*brokers) {
		if b != "" {
			brokerList = append(brokerList, b)
		}
	}
	if len(brokerList) == 0 {
		log.Fatal("no brokers provided")
	}

	cfg := consumerConfig{
		Brokers:  brokerList,
		Topic:    *topic,
		GroupID:  *groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}
	log.Printf("Producer config: brokers=%v topic=%s", cfg.Brokers, cfg.Topic)

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		Balancer: &kafka.Hash{},
	})
	defer func() {
		if err := w.Close(); err != nil {
			log.Println("error closing kafka writer:", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		log.Println("signal received, stopping...")
		cancel()
	}()

	rand.Seed(time.Now().UnixNano())

	sent := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("context canceled, exiting loop")
			return
		default:
		}

		if *count > 0 && sent >= *count {
			log.Printf("sent %d messages, done", sent)
			return
		}

		o := generateOrder()
		b, err := json.Marshal(o)
		if err != nil {
			log.Println("json marshal error:", err)
			continue
		}

		msg := kafka.Message{
			Key:   []byte(o.OrderUID),
			Value: b,
			Time:  time.Now(),
		}

		writeCtx, writeCancel := context.WithTimeout(ctx, 10*time.Second)
		err = w.WriteMessages(writeCtx, msg)
		writeCancel()
		if err != nil {
			log.Printf("failed to write message: %v", err)
		} else {
			sent++
			log.Printf("sent message #%d order_uid=%s size=%d", sent, o.OrderUID, len(b))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(*interval):
		}
	}
}

type consumerConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

func splitAndTrim(s string) []string {
	var out []string
	for _, p := range kafkaSplit(s) {
		out = append(out, p)
	}
	return out
}

func kafkaSplit(s string) []string {
	var res []string
	cur := ""
	for _, r := range s {
		if r == ',' {
			res = append(res, trimSpace(cur))
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		res = append(res, trimSpace(cur))
	}
	return res
}
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func generateOrder() *model.Order {
	now := time.Now().UTC()
	uid := uuid.New().String()

	itemCount := rand.Intn(3) + 1
	items := make([]model.Item, 0, itemCount)
	var total float64
	for i := 0; i < itemCount; i++ {
		price := float64(rand.Intn(2000)+100) / 100.0
		count := 1
		tp := price * float64(count)
		items = append(items, model.Item{
			ChrtID:      int64(rand.Intn(1_000_000) + 1),
			TrackNumber: fmt.Sprintf("TRK%06d", rand.Intn(1_000_000)),
			Price:       price,
			Rid:         fmt.Sprintf("RID%04d", rand.Intn(10000)),
			Name:        fmt.Sprintf("Item-%d", rand.Intn(1000)),
			Sale:        rand.Intn(50),
			Size:        "M",
			TotalPrice:  tp,
			NmID:        int64(rand.Intn(1_000_000)),
			Brand:       "BrandX",
			Status:      1,
		})
		total += tp
	}

	order := &model.Order{
		OrderUID:          uid,
		TrackNumber:       fmt.Sprintf("TRACK-%s", uid[:8]),
		Entry:             "WEB",
		Locale:            "en-GB",
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("cust-%d", rand.Intn(100000)),
		DeliveryService:   "dhl",
		Shardkey:          fmt.Sprintf("%d", rand.Intn(100)),
		SmID:              rand.Intn(1000),
		DateCreated:       now,
		OofShard:          "1",
		Delivery: model.Delivery{
			Name:    "Ivan Ivanov",
			Phone:   "+79991234567",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina, 1",
			Region:  "Moscow",
			Email:   "ivan@example.com",
		},
		Payment: model.Payment{
			Transaction:  fmt.Sprintf("tx-%s", uid[:8]),
			RequestID:    "",
			Currency:     "RUB",
			Provider:     "bank",
			Amount:       total,
			PaymentDt:    now.Unix(),
			Bank:         "BigBank",
			DeliveryCost: 0,
			GoodsTotal:   int(total),
			CustomFee:    0,
		},
		Items: items,
	}
	return order
}
