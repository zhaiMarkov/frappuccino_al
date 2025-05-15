package models

import (
	"encoding/json"
	"frappuchino/internal/apperrors"
	"time"
)

// Заказ
type Order struct {
	ID                  int             `json:"id"`                   // ID заказа
	CustomerID          int             `json:"customer_id"`          // ID клиента
	TotalAmount         float64         `json:"total_amount"`         // общая сумма
	Status              string          `json:"status"`               // статус заказа
	SpecialInstructions json.RawMessage `json:"special_instructions"` // особые пожелания
	PaymentMethod       string          `json:"payment_method"`       // способ оплаты
	CreatedAt           time.Time       `json:"created_at"`           // время создания
	UpdatedAt           time.Time       `json:"updated_at"`           // время обновления
}

// Позиция заказа
type OrderItem struct {
	ID         string  `json:"id"`             // ID позиции
	OrderID    int     `json:"order_id"`       // ID заказа
	Quantity   int     `json:"quantity"`       // количество
	Price      float64 `json:"price_at_order"` // цена на момент заказа
	MenuItemID string  `json:"menu_item_id"`   // ID товара из меню
}

// Создание заказа
func NewOrder(customerID int, totalAmount float64, dto CreateOrderRequest) (*Order, error) {
	if customerID < 1 || totalAmount == 0 || dto.PaymentMethod == "" {
		return nil, apperrors.ErrInvalidInput
	}

	// допустимые методы оплаты
	if !(dto.PaymentMethod == "card" || dto.PaymentMethod == "cash" || dto.PaymentMethod == "kaspi_qr") {
		return nil, apperrors.ErrInvalidInput
	}

	// значение по умолчанию
	if dto.Instructions == nil {
		dto.Instructions = json.RawMessage(`{"special_request":"no"}`)
	}

	return &Order{
		CustomerID:          customerID,
		TotalAmount:         totalAmount,
		Status:              "open",
		SpecialInstructions: dto.Instructions,
		PaymentMethod:       dto.PaymentMethod,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}

// Создание позиций заказа
func NewOrderItems(items []OrderItemInput, productPrices map[string]float64) ([]*OrderItem, error) {
	if len(items) < 1 || len(productPrices) < 1 {
		return nil, apperrors.ErrInvalidInput
	}

	orderItems := []*OrderItem{}
	for _, item := range items {
		// проверка наличия цены на товар
		price, ok := productPrices[item.ProductID]
		if !ok {
			return nil, apperrors.ErrInvalidInput
		}

		orderItems = append(orderItems, &OrderItem{
			MenuItemID: item.ProductID,
			Quantity:   item.Quantity,
			Price:      price,
		})
	}

	return orderItems, nil
}
