package models

import (
	"frappuchino/internal/apperrors"
)

// Структура запроса на создание товара на складе
type CreateInventoryRequest struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	StockLevel float64 `json:"stock_level"` // количество на складе
	Price      float64 `json:"price"`       // цена за единицу
	UnitType   string  `json:"unit_type"`   // тип единицы (например, кг, л, шт)
}

// Конструктор с валидацией данных и генерацией ID
func NewCreateInventoryRequest(inventoryRequest CreateInventoryRequest) (*CreateInventoryRequest, error) {
	// Проверка обязательных полей
	if inventoryRequest.Name == "" || inventoryRequest.UnitType == "" || inventoryRequest.StockLevel <= 0 || inventoryRequest.Price <= 0 {
		return nil, apperrors.ErrInvalidInput
	}

	// Если ID не задан — генерируем его из имени
	if inventoryRequest.ID == "" {
		inventoryRequest.ID = fromNameToID(inventoryRequest.Name)
	}

	return &CreateInventoryRequest{
		ID:         inventoryRequest.ID,
		Name:       inventoryRequest.Name,
		StockLevel: inventoryRequest.StockLevel,
		Price:      inventoryRequest.Price,
		UnitType:   inventoryRequest.UnitType,
	}, nil
}
