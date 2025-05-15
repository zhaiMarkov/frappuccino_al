package models

import (
	"encoding/json"
	"frappuchino/internal/apperrors"
)

// Структура покупателя
type Customer struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Preferences json.RawMessage `json:"preferences"` // хранит произвольные JSON-настройки
}

// Конструктор покупателя с валидацией и автогенерацией email
func NewCustomer(name string, instructions json.RawMessage) (*Customer, error) {
	if name == "" {
		return nil, apperrors.ErrInvalidInput // имя обязательно
	}

	// если настройки не переданы — установить значение по умолчанию
	if instructions == nil {
		instructions = json.RawMessage(`{"preferences": true}`)
	}

	email := fromNameToID(name) + "@gmail.com" // генерация email из имени

	return &Customer{
		Name:        name,
		Email:       email,
		Preferences: instructions,
	}, nil
}
