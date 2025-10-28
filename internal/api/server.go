// Package api одержит http-сервер
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	svc "github.com/gogazub/myapp/internal/service"
)

// IServer - интерфейс http-сервера. Требует реализации handleGetOrderByID
type IServer interface {
	handleGetOrderByID(w http.ResponseWriter, r *http.Request)
}

// Server - реализация http-сервера.
type Server struct {
	service svc.IService
}

// NewServer - конструктор.
func NewServer(service svc.IService) *Server {
	return &Server{
		service: service,
	}
}

// Обработчик ошибок вынесен в отдельный модуль. На это две причины
// 1. обработчики запросов не должны заниматься логированием
// 2. даем гибкость настройки логирования
func (s *Server) handleError(msg string, err error) {
	log.Printf("%s:%v", msg, err)
}

// Start запускает сервер
func (s *Server) Start(address string) error {
	http.HandleFunc("/orders/", s.handleGetOrderByID)
	http.Handle("/", http.FileServer(http.Dir("./internal/api/web")))
	http.HandleFunc("/healt", handleHealth)
	// TODO: вывод сообщений в терминал должен быть в main
	log.Printf("Server is running on %s\n", address)
	return http.ListenAndServe(address, nil)
}

// Обработчик GET-запросов по order_id
func (s *Server) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Path[len("/orders/"):]

	// Используем контекст из http запроса. Он канселится, если запрос был отменен или разорвано соединение
	ctxBase := r.Context()
	// Навешиваем на него таймаут и прокидываем во все слои
	ctx, cancel := context.WithTimeout(ctxBase, 1*time.Minute)
	defer cancel()
	order, err := s.service.GetOrderByID(ctx, orderID)
	if err != nil {
		// Если получили error от GetOrderById, то даем пользователю 404,а саму ошибку логируем
		w.WriteHeader(http.StatusNotFound)

		s.handleError("Failed to get order", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.handleError("Failed to encode order", err)
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	ok := map[string]string{
		"status": "ok",
	}
	err := json.NewEncoder(w).Encode(&ok)
	if err != nil {
		log.Println(err.Error())
	}
}
