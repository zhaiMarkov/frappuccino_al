package router

import (
	"frappuchino/internal/handler"
	"net/http"
)

func InventoryRouter(h *handler.InventoryHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Используем стандартные пути для маршрутов
	mux.HandleFunc("POST /inventory", h.CreateInventoryItem)
	mux.HandleFunc("GET /inventory", h.GetAllInventoryItems)
	mux.HandleFunc("GET /inventory/{id}", h.GetInventoryItem)
	mux.HandleFunc("PUT /inventory/{id}", h.UpdateInventoryItem)
	mux.HandleFunc("DELETE /inventory/{id}", h.DeleteInventoryItem)
	mux.HandleFunc("GET /inventory/getLeftOvers", h.GetLeftItems)

	return mux
}
