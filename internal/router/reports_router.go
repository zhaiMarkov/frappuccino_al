package router

import (
	"frappuchino/internal/handler"
	"net/http"
)

func ReportRouter(h *handler.ReportsHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Используем стандартные пути для маршрутов
	mux.HandleFunc("GET /reports/total-sales", h.TotalSalesReportHandler)
	mux.HandleFunc("GET /reports/popular-items", h.PopularItemsReportHandler)
	mux.HandleFunc("GET /reports/search", h.SearchHandler)
	mux.HandleFunc("GET /reports/orderedItemsByPeriod", h.OrderedItemsByPeriodHandler)

	return mux
}
