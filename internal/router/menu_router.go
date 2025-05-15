package router

import (
	"frappuchino/internal/handler"
	"net/http"
)

func MenuRouter(h *handler.MenuHandler) *http.ServeMux {
	mux := http.NewServeMux()
	// Используем стандартные пути для маршрутов
	mux.HandleFunc("POST /menu", h.CreateMenuItem)
	mux.HandleFunc("GET /menu", h.GetAllMenuItems)
	mux.HandleFunc("GET /menu/{id}", h.GetMenuItem)
	mux.HandleFunc("PUT /menu/{id}", h.UpdateMenuItem)
	mux.HandleFunc("DELETE /menu/{id}", h.DeleteMenuItem)

	return mux
}
