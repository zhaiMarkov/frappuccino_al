package router

import (
	"frappuchino/internal/handler"
	"net/http"
)

func OrderRouter(h *handler.OrderHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Используем стандартные пути для маршрутов
	mux.HandleFunc("POST /orders", h.CreateOrder)
	mux.HandleFunc("GET /orders", h.GetAllOrders)
	mux.HandleFunc("GET /orders/{id}", h.GetOrder)
	mux.HandleFunc("PUT /orders/{id}", h.UpdateOrder)
	mux.HandleFunc("DELETE /orders/{id}", h.DeleteOrder)
	mux.HandleFunc("POST /orders/{id}/close", h.CloseOrder)
	mux.HandleFunc("GET /orders/numberOfOrderedItems", h.NumberOfOrderedItems)
	mux.HandleFunc("POST /orders/batch-process", h.BatchCreateOrders)

	return mux
}
