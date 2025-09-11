package consumer

import (
	"context"

	"github.com/gogazub/myapp/internal/orders"
)

type OrderHandlerFunc func(ctx context.Context, order *orders.ModelOrder) error

type FuncHandler struct {
	handlerFunc OrderHandlerFunc
}

func NewFuncHandler(handlerFunc OrderHandlerFunc) *FuncHandler {
	return &FuncHandler{handlerFunc: handlerFunc}
}

func (h *FuncHandler) HandleOrder(order *orders.ModelOrder) error {
	return h.handlerFunc(context.Background(), order)
}
