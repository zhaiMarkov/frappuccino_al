package handler

import (
	"encoding/json"
	"frappuchino/internal/models"
	"log/slog"
	"net/http"
)

// Интерфейс MenuService определяет контракт для работы с меню.
// Имплементация этого интерфейса будет использоваться в обработчиках.
type MenuService interface {
	CreateMenuItemService(menuNew models.CreateMenuRequest) error
	GetAllMenuItemsService() ([]*models.MenuItem, error)
	GetMenuItemService(id string) (*models.MenuItem, error)
	UpdateMenuItemService(id string, menuItem models.CreateMenuRequest) error
	DeleteMenuItemService(id string) error
}

// Структура MenuHandler инкапсулирует сервис меню,
// и используется для обработки HTTP-запросов.
type MenuHandler struct {
	menuService MenuService
}

// Конструктор MenuHandler — инициализирует новый обработчик с переданным сервисом меню.
func NewMenuHandler(mS MenuService) *MenuHandler {
	return &MenuHandler{menuService: mS}
}

// Обработчик для создания нового элемента меню.
// Декодирует JSON из тела запроса и передает данные в сервис.
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	if !isJSONFile(w, r) {
		slog.Error("Data is not JSON format")
		return
	}

	var inputMenu models.CreateMenuRequest

	// Декодирование JSON → структура inputMenu
	if err := json.NewDecoder(r.Body).Decode(&inputMenu); err != nil {
		slog.Error("Handler error in Create Menu: decoding JSON data ", "error", err)
		writeError(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Валидация и приведение структуры к модели
	menu, err := models.NewCreateMenuRequest(inputMenu)
	if err != nil {
		slog.Error("Handler error in Create Menu: invalid input data", "input item", inputMenu, "error", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Вызов сервиса для создания
	if err := h.menuService.CreateMenuItemService(*menu); err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Create Menu: creating menu", "menu item", menu, "error", err)
		writeError(w, err.Error(), status)
		return
	}

	slog.Info("Menu item created successfully", "id", menu.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// Обработчик для получения всех элементов меню
func (h *MenuHandler) GetAllMenuItems(w http.ResponseWriter, r *http.Request) {
	menu, err := h.menuService.GetAllMenuItemsService()
	if err != nil {
		slog.Error("Handler error in Get Menu: retrieving all menu", "error", err)
		writeError(w, "Failed to retrieve all menu", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, menu)
	slog.Info("All menu items retrieved successfully", "count", len(menu))
}

// Обработчик для получения одного элемента меню по ID
func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	menu, err := h.menuService.GetMenuItemService(id)
	if err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Get Menu: retrieving menu item", "id", id, "error", err)
		writeError(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, menu)
	slog.Info("Menu item retrieved successfully", "id", id)
}

// Обработчик для обновления элемента меню по ID
func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	if !isJSONFile(w, r) {
		slog.Error("Data is not JSON format")
		return
	}
	id := r.PathValue("id")

	var inputMenu models.CreateMenuRequest

	// Декодирование JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&inputMenu); err != nil {
		slog.Error("Handler error in Update Menu: decoding JSON data", "error", err)
		writeError(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Валидация и преобразование в модель
	menu, err := models.NewCreateMenuRequest(inputMenu)
	if err != nil {
		slog.Error("Handler error in Update Menu: invalid input data", "item", inputMenu, "error", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Вызов сервиса обновления
	if err := h.menuService.UpdateMenuItemService(id, *menu); err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Update Menu: updating menu", "menu item", menu, "error", err)
		writeError(w, err.Error(), status)
		return
	}

	slog.Info("Menu item updated successfully", "id", id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// Обработчик для удаления элемента меню по ID
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Вызов сервиса удаления
	if err := h.menuService.DeleteMenuItemService(id); err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Delete Menu: deleting menu item", "id", id, "error", err)
		writeError(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	slog.Info("Menu item deleted successfully", "id", id)
}
