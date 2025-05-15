package repository

import (
	"database/sql"
	"frappuchino/internal/models"
	"log/slog"
	"strconv"

	"github.com/lib/pq"
)

type ReportsRepository struct {
	db *sql.DB
}

func NewReportsRepository(db *sql.DB) *ReportsRepository {
	return &ReportsRepository{
		db: db,
	}
}

func (r *ReportsRepository) Close() error {
	return r.db.Close()
}

func (r *ReportsRepository) GetTotalSales() (*models.TotalPrice, error) {
	query := `
	SELECT SUM(total_amount) AS total_sales
	FROM orders
`

	var totalSales models.TotalPrice
	if err := r.db.QueryRow(query).Scan(&totalSales.TotalSale); err != nil {
		slog.Error("Repository error from Get Total Sales: failed retrieve total amount", "error", err)
		return nil, err
	}

	slog.Info("Repository info: calculating total sales successfully", "total sales", totalSales)
	return &totalSales, nil
}

func (r *ReportsRepository) GetPopularItems() ([]*models.PopularItem, error) {
	query := `
	SELECT menu_item_id, SUM(quantity) AS count
	FROM order_items
	GROUP BY menu_item_id
	ORDER BY count DESC
	LIMIT 3
	`

	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("Repository error from Get Popular Item: failed to retrieve popular menu items", "error", err)
		return nil, err
	}
	defer rows.Close()

	var popularItems []*models.PopularItem
	for rows.Next() {
		var popularItem models.PopularItem
		if err := rows.Scan(&popularItem.ItemName, &popularItem.QuantityOfSales); err != nil {
			slog.Error("Repository error from Get Popular Item: failed to scan menu item row", "error", err)
			return nil, err
		}
		popularItems = append(popularItems, &popularItem)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository error from Get Popular Item: failed  iterating over rows", "error", err)
		return nil, err
	}

	slog.Info("Repository info: retrieved popular items successfully")
	return popularItems, nil
}

func (r *ReportsRepository) SearchMenuItems(q string, minPrice, maxPrice float64) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, description, price,
			   ts_rank(to_tsvector('english', name || ' ' || description), to_tsquery('english', REPLACE($1, ' ', '&'))) AS relevance
		FROM menu_items
		WHERE to_tsvector('english', name || ' ' || description) @@ to_tsquery('english', REPLACE($1, ' ', '&')) 
					AND price >= $2 AND price <= $3
		ORDER BY relevance DESC`

	rows, err := r.db.Query(query, q, minPrice, maxPrice)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuItems []*models.MenuItem
	relevanceValues := make(map[string]float64)
	for rows.Next() {
		var relevance float64
		item := &models.MenuItem{}
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &relevance)
		if err != nil {
			return nil, err
		}
		menuItems = append(menuItems, item)
		relevanceValues[item.ID] = relevance
	}

	var result []map[string]interface{}
	for _, item := range menuItems {
		itemMap := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"price":       item.Price,
			"relevance":   relevanceValues[item.ID],
		}
		result = append(result, itemMap)
	}

	return result, nil
}

func (r *ReportsRepository) SearchOrders(q string, minPrice, maxPrice float64) ([]map[string]interface{}, error) {
	query := `
		SELECT o.id, c.name AS customer_name, 
				ARRAY_AGG(mi.name) AS items, 
				o.total_amount, 
				ts_rank(to_tsvector('english', c.name || ' ' || mi.name), to_tsquery('english', REPLACE($1, ' ', '&'))) AS relevance
		FROM orders o
		JOIN customers c 
			ON o.customer_id = c.id
		JOIN order_items oi 
			ON o.id = oi.order_id
		JOIN menu_items mi
			ON oi.menu_item_id = mi.id
		WHERE to_tsvector('english', c.name || ' ' || mi.name) @@ to_tsquery('english', REPLACE($1, ' ', '&'))
					AND o.total_amount >= $2 AND o.total_amount <= $3
		GROUP BY o.id, c.name, o.total_amount, relevance
		ORDER BY relevance DESC`

	rows, err := r.db.Query(query, q, minPrice, maxPrice)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type orderResp struct {
		ID           int
		CustomerName string
		Total        float64
		Items        []string
		Relevance    float64
	}
	var orders []*orderResp
	for rows.Next() {
		order := &orderResp{}
		err := rows.Scan(&order.ID, &order.CustomerName, pq.Array(&order.Items), &order.Total, &order.Relevance)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	var result []map[string]interface{}
	for _, item := range orders {
		itemMap := map[string]interface{}{
			"id":            item.ID,
			"customer_name": item.CustomerName,
			"items":         item.Items,
			"total":         item.Total,
			"relevance":     item.Relevance,
		}
		result = append(result, itemMap)
	}

	return result, nil
}

func (r *ReportsRepository) OrderedItemByDayRepository(month string) (map[string]interface{}, error) {
	query := `
		SELECT EXTRACT(DAY FROM created_at) AS day, COUNT(*) AS orders
		FROM orders 
		WHERE LOWER(TO_CHAR(created_at, 'FMMonth')) = $1
		GROUP BY day
		ORDER BY day
	`
	rows, err := r.db.Query(query, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderedItems []map[string]int
	for rows.Next() {
		var day int
		var orders int
		if err := rows.Scan(&day, &orders); err != nil {
			return nil, err
		}
		orderedItems = append(orderedItems, map[string]int{strconv.Itoa(day): orders})
	}

	result := map[string]interface{}{
		"period":       "day",
		"month":        month,
		"orderedItems": orderedItems,
	}

	return result, nil
}

func (r *ReportsRepository) OrderedItemByMonthRepository(year int) (map[string]interface{}, error) {
	slog.Info("year", "y", year)
	query := `
		SELECT LOWER(TO_CHAR(created_at, 'FMMonth')) AS month, COUNT(*) AS orders
		FROM orders
		WHERE EXTRACT(YEAR FROM created_at) = $1
		GROUP BY month
		ORDER BY month;
	`

	rows, err := r.db.Query(query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderedItems []map[string]int
	for rows.Next() {
		var month string
		var orders int
		if err := rows.Scan(&month, &orders); err != nil {
			return nil, err
		}
		orderedItems = append(orderedItems, map[string]int{month: orders})
	}

	result := map[string]interface{}{
		"period":       "year",
		"year":         year,
		"orderedItems": orderedItems,
	}

	return result, nil
}
