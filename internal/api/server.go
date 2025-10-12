package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	svc "github.com/gogazub/myapp/internal/service"
)

type Server struct {
	service svc.Service
}

func NewServer(service svc.Service) *Server {
	return &Server{
		service: service,
	}
}

// Запускает сервер
func (s *Server) Start(address string) error {
	http.HandleFunc("/orders/", s.handleGetOrderByID)
	http.Handle("/", http.FileServer(http.Dir("./internal/server/web")))
	log.Printf("Server is running on %s\n", address)
	return http.ListenAndServe(address, nil)
}

// Обработчик GET-запросов по order_id
func (s *Server) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Path[len("/orders/"):]

	// Используем контекст из http запроса. Он канселится, если запрос был отменен или разорвано соединение
	ctx_base := r.Context()
	// Навешиваем на него таймаут и прокидываем во все слои
	ctx, cancel := context.WithTimeout(ctx_base, 1*time.Minute)
	defer cancel()
	order, err := s.service.GetOrderByID(ctx, orderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get order: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode order: %v", err), http.StatusInternalServerError)
	}
}
