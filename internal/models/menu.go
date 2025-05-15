package models

import (
	"frappuchino/internal/apperrors"
)

// Элемент меню
type MenuItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Allergens   []string `json:"allergens"` // возможные аллергены
	Size        string   `json:"size"`
}

// Связь ингредиентов с пунктом меню
type MenuItemIngredient struct {
	ID           int     `json:"id"`
	MenuItemID   string  `json:"menu_item_id"`
	Quantity     float64 `json:"quantity"`      // количество на порцию
	IngredientID string  `json:"ingredient_id"` // ID товара на складе
}

// Конструктор MenuItem с валидацией
func NewMenuItem(allergens []string, dto CreateMenuRequest) (*MenuItem, error) {
	if dto.Name == "" || dto.Price <= 0 || dto.Size == "" {
		return nil, apperrors.ErrInvalidInput
	}

	// значение по умолчанию для описания
	description := dto.Description
	if description == "" {
		description = "No description"
	}

	return &MenuItem{
		ID:          dto.ID,
		Name:        dto.Name,
		Description: description,
		Price:       dto.Price,
		Allergens:   allergens,
		Size:        dto.Size,
	}, nil
}

// Генерация списка ингредиентов для пункта меню
func NewMenuItemIngredients(menuItemID string, items []MenuItemIngredientInput) ([]*MenuItemIngredient, error) {
	if menuItemID == "" || len(items) < 1 {
		return nil, apperrors.ErrInvalidInput
	}

	menuItems := []*MenuItemIngredient{}
	for _, item := range items {
		if item.Quantity <= 0 || item.IngredientID == "" {
			return nil, apperrors.ErrInvalidInput
		}
		menuItems = append(menuItems, &MenuItemIngredient{
			MenuItemID:   menuItemID,
			Quantity:     item.Quantity,
			IngredientID: item.IngredientID,
		})
	}
	return menuItems, nil
}
