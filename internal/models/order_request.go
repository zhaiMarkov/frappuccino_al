package models

import (
	"encoding/json"
	"frappuchino/internal/apperrors"
)

// Запрос на создание заказа
type CreateOrderRequest struct {
	CustomerName  string           `json:"customer_name"`  // имя клиента
	PaymentMethod string           `json:"payment_method"` // способ оплаты
	Items         []OrderItemInput `json:"items"`          // список товаров
	Instructions  json.RawMessage  `json:"instructions"`   // дополнительные пожелания
}

// Один товар в заказе
type OrderItemInput struct {
	ProductID string `json:"product_id"` // ID продукта
	Quantity  int    `json:"quantity"`   // количество
}

// Конструктор CreateOrderRequest с валидацией
func NewCreateOrder(createOrder CreateOrderRequest) (*CreateOrderRequest, error) {
	if createOrder.CustomerName == "" || createOrder.PaymentMethod == "" {
		return nil, apperrors.ErrInvalidInput
	}

	// проверка всех позиций заказа
	for _, createOrderItem := range createOrder.Items {
		if createOrderItem.ProductID == "" || createOrderItem.Quantity <= 0 {
			return nil, apperrors.ErrInvalidInput
		}
	}

	return &CreateOrderRequest{
		CustomerName:  createOrder.CustomerName,
		PaymentMethod: createOrder.PaymentMethod,
		Items:         createOrder.Items,
		Instructions:  createOrder.Instructions,
	}, nil
}
