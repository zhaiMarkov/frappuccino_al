package models

// Структура для хранения общей суммы продаж
type TotalPrice struct {
	TotalSale float64 `json:"total-sales"` // общая сумма продаж
}

// Структура для хранения популярного товара
type PopularItem struct {
	ItemName        string `json:"item-name"`         // название товара
	QuantityOfSales int    `json:"quantity_of_sales"` // количество продаж
}

// Конструктор для TotalPrice
func NewTotalPrice() *TotalPrice {
	return &TotalPrice{}
}

// Конструктор для PopularItem
func NewPopularItem(name string, quantity int) PopularItem {
	return PopularItem{
		ItemName:        name,
		QuantityOfSales: quantity,
	}
}
