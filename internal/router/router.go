package router

import (
	"database/sql"
	"frappuchino/internal/handler"
	"frappuchino/internal/repository"
	"frappuchino/internal/service"
	"net/http"
)

// LoadRoutes настраивает маршрутизацию HTTP-запросов, инициализируя репозитории,
// сервисы и обработчики для инвентаря, меню, заказов и отчетов системы frappuchino
func LoadRoutes(db *sql.DB) (*http.ServeMux, error) {
	// Инициализация компонентов инвентаря
	inventRepo := repository.NewInventoryRepository(db)
	inventService := service.NewInventoryService(inventRepo)
	inventHandler := handler.NewInventHandler(inventService)

	// Инициализация компонентов меню
	menuRepo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(menuRepo, inventRepo)
	menuHandler := handler.NewMenuHandler(menuService)

	// Инициализация компонентов заказов
	customerRepo := repository.NewCustomerRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, menuRepo, inventRepo, customerRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	// Инициализация компонентов отчетов
	reportRepo := repository.NewReportsRepository(db)
	serviceReports := service.NewReportsService(reportRepo)
	handlerReports := handler.NewReportsHandler(serviceReports)

	// Создание маршрутизатора и регистрация обработчиков
	mux := http.NewServeMux()
	addRoutes(mux, "/inventory", InventoryRouter(inventHandler))
	addRoutes(mux, "/menu", MenuRouter(menuHandler))
	addRoutes(mux, "/orders", OrderRouter(orderHandler))
	addRoutes(mux, "/reports", ReportRouter(handlerReports))

	return mux, nil
}

// addRoutes регистрирует обработчик для пути с учетом и без завершающего слеша
func addRoutes(mux *http.ServeMux, path string, router http.Handler) {
	mux.Handle(path, router)
	mux.Handle(path+"/", router)
}
