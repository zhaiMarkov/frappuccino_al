package handler

import (
	"encoding/json"
	"frappuchino/internal/models"
	"log/slog"
	"net/http"
)

// InventoryService определяет интерфейс бизнес-логики для работы с инвентарем.
type InventoryService interface {
	CreateInventoryItemService(invent models.CreateInventoryRequest) error
	GetAllInventoryItemsService() ([]*models.InventoryItem, error)
	GetInventoryItemService(id string) (*models.InventoryItem, error)
	UpdateInventoryItemService(id string, inventoryItem models.CreateInventoryRequest) error
	DeleteInventoryItemService(id string) error
	GetLeftOversService(sortBy, page, pageSize string) (map[string]interface{}, error)
}

// InventoryHandler — HTTP-обработчик, взаимодействующий с InventoryService.
type InventoryHandler struct {
	inventoryService InventoryService
}

// NewInventHandler создает новый экземпляр InventoryHandler.
func NewInventHandler(iS InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: iS}
}

// CreateInventoryItem обрабатывает POST-запрос для создания нового элемента инвентаря.
func (h *InventoryHandler) CreateInventoryItem(w http.ResponseWriter, r *http.Request) {
	// Проверка, что тело запроса — JSON
	if !isJSONFile(w, r) {
		slog.Error("Data is not JSON format")
		return
	}

	// Декодирование запроса
	var inputInvent models.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&inputInvent); err != nil {
		slog.Error("Handler error in Create Inventory: decoding JSON data", "error", err)
		writeError(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Валидация данных
	invent, err := models.NewCreateInventoryRequest(inputInvent)
	if err != nil {
		slog.Error("Handler error in Create Inventory: invalid input data", "item", inputInvent, "error", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Сохранение в БД
	if err := h.inventoryService.CreateInventoryItemService(*invent); err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Create Inventory: creating inventory item", "inventory item", invent, "Error", err)
		writeError(w, err.Error(), status)
		return
	}

	slog.Info("Inventory created successfully", "inventory ID", invent.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GetAllInventoryItems обрабатывает GET-запрос для получения всех элементов инвентаря.
func (h *InventoryHandler) GetAllInventoryItems(w http.ResponseWriter, r *http.Request) {
	allInvents, err := h.inventoryService.GetAllInventoryItemsService()
	if err != nil {
		slog.Error("Handler error in Get Inventory: retrieving all inventory items", "error", err)
		writeError(w, "Failed to retrieve inventory items", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, allInvents)
	slog.Info("Inventory items retrieved successfully", "count", len(allInvents))
}

// GetInventoryItem обрабатывает GET-запрос для получения конкретного элемента инвентаря по ID.
func (h *InventoryHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	inventId, err := h.inventoryService.GetInventoryItemService(id)
	if err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Get Inventory: retrieving inventory item", "id", id, "error", err)
		writeError(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, inventId)
	slog.Info("Inventory item retrieved successfully", "id", id)
}

// UpdateInventoryItem обрабатывает PUT-запрос для обновления элемента инвентаря по ID.
func (h *InventoryHandler) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	// Проверка на JSON
	if !isJSONFile(w, r) {
		slog.Error("Data is not JSON format")
		return
	}
	id := r.PathValue("id")

	// Декодирование запроса
	var inputInvent models.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&inputInvent); err != nil {
		slog.Error("Handler error in Update Inventory: decoding JSON data", "error", err)
		writeError(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Валидация данных
	inventoryItem, err := models.NewCreateInventoryRequest(inputInvent)
	if err != nil {
		slog.Error("Handler error in Update Inventory: invalid input data", "item", inputInvent, "error", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Обновление в БД
	if err := h.inventoryService.UpdateInventoryItemService(id, *inventoryItem); err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Update Inventory: updating inventory", "inventory item", inventoryItem, "error", err)
		writeError(w, err.Error(), status)
		return
	}
	slog.Info("Inventory updated successfully", "id", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// DeleteInventoryItem обрабатывает DELETE-запрос для удаления элемента инвентаря по ID.
func (h *InventoryHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.inventoryService.DeleteInventoryItemService(id); err != nil {
		status := mapAppErrorToStatus(err)
		slog.Error("Handler error in Delete Inventory: deleting inventory", "id", id, "error", err)
		writeError(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	slog.Info("Inventory item deleted successfully", "id", id)
}

// GetLeftOvers обрабатывает GET-запрос для получения остатков инвентаря с пагинацией и сортировкой.
func (h *InventoryHandler) GetLeftItems(w http.ResponseWriter, r *http.Request) {
	// Чтение query-параметров
	sortBy := r.URL.Query().Get("sortBy")
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageSize := r.URL.Query().Get("pageSize")
	if pageSize == "" {
		pageSize = "10"
	}

	// Получение данных из сервиса
	leftOvers, err := h.inventoryService.GetLeftOversService(sortBy, page, pageSize)
	if err != nil {
		slog.Error("Handler error in Get LeftOvers: retrieving left overs", "sortBy", sortBy, "page", page, "pageSize", pageSize, "error", err)
		writeError(w, "Failed to retrieve left overs", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, leftOvers)
	slog.Info("Left overs retrieved successfully")
}
