package models

import (
	"frappuchino/internal/apperrors"
)

// Запрос на создание пункта меню
type CreateMenuRequest struct {
	ID          string                    `json:"product_id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Price       float64                   `json:"price"`
	Size        string                    `json:"size"`
	Ingredients []MenuItemIngredientInput `json:"ingredients"` // список ингредиентов
}

// Ингредиент для позиции меню
type MenuItemIngredientInput struct {
	IngredientID string  `json:"ingredient_id"` // id товара на складе
	Quantity     float64 `json:"quantity"`      // количество на одну порцию
}

// Конструктор с валидацией и автозаполнением
func NewCreateMenuRequest(menuRequest CreateMenuRequest) (*CreateMenuRequest, error) {
	if menuRequest.Name == "" || menuRequest.Price <= 0 || menuRequest.Size == "" {
		return nil, apperrors.ErrInvalidInput // обязательные поля
	}

	// генерация ID по имени, если не указан
	if menuRequest.ID == "" {
		menuRequest.ID = fromNameToID(menuRequest.Name)
	}

	// если описание отсутствует — ставим по умолчанию
	if menuRequest.Description == "" {
		menuRequest.Description = "No description"
	}

	// проверка ингредиентов
	for _, ingredient := range menuRequest.Ingredients {
		if ingredient.IngredientID == "" || ingredient.Quantity <= 0 {
			return nil, apperrors.ErrInvalidInput
		}
	}

	return &CreateMenuRequest{
		ID:          menuRequest.ID,
		Name:        menuRequest.Name,
		Description: menuRequest.Description,
		Price:       menuRequest.Price,
		Size:        menuRequest.Size,
		Ingredients: menuRequest.Ingredients,
	}, nil
}
