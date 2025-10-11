package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gogazub/myapp/internal/service"
)

type Server struct {
	service service.Service
}

func NewServer(service orders.Service) *Server {
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

	order, err := s.service.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get order: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode order: %v", err), http.StatusInternalServerError)
	}
}
