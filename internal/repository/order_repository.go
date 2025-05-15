package repository

import (
	"database/sql"
	"fmt"
	"frappuchino/internal/apperrors"
	"frappuchino/internal/models"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) Close() error {
	return r.db.Close()
}

func (r *OrderRepository) AddOrderRepository(order models.Order, orderItems []*models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Add Order: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	orderQuery := `
		INSERT INTO orders (customer_id, total_amount, status, special_instructions, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var orderID int
	err = tx.QueryRow(orderQuery, order.CustomerID, order.TotalAmount, order.Status, order.SpecialInstructions, order.PaymentMethod, order.CreatedAt, order.UpdatedAt).Scan(&orderID)
	if err != nil {
		slog.Error("Repository error from Add Order: failed to add order", "customer_id", order.CustomerID, "error", err)
		return err
	}
	itemQuery := `
		INSERT INTO order_items (order_id, quantity, price_at_order, menu_item_id)
		VALUES ($1, $2, $3, $4)
	`
	for _, item := range orderItems {
		_, err := tx.Exec(itemQuery, orderID, item.Quantity, item.Price, item.MenuItemID)
		if err != nil {
			slog.Error("Repository error from Add Order: failed to add order item", "order_id", orderID, "menu_item_id", item.MenuItemID, "error", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Add Order: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: order added successfully", "order_id", orderID)
	return nil
}

func (r *OrderRepository) GetAllOrdersRepository() ([]*models.Order, error) {
	query := `
		SELECT id, customer_id, total_amount, status, special_instructions, payment_method, created_at, updated_at
		FROM orders;
	`

	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("Repository error from Get Orders: failed to retrieve all orders", "error", err)
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.SpecialInstructions, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt); err != nil {
			slog.Error("Repository error from Get Orders: failed to scan order row", "error", err)
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository error from Get Orders: failed iterating over rows", "error", err)
		return nil, err
	}

	slog.Info("Repository info: retrieved all orders successfully", "count", len(orders))
	return orders, nil
}

func (r *OrderRepository) GetOrderRepository(id int) (*models.Order, error) {
	orderQuery := `
		SELECT id, customer_id, total_amount, status, payment_method, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var order models.Order
	err := r.db.QueryRow(orderQuery, id).Scan(
		&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		slog.Error("Repository error from Get Order: order not found", "id", id)
		return nil, apperrors.ErrNotExistConflict
	} else if err != nil {
		slog.Error("Repository error from Update Order: failed to retrieve order", "id", id, "error", err)
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}
	slog.Info("Order retrieved successfully", "id", id)
	return &order, nil
}

func (r *OrderRepository) UpdateOrderRepository(id int, order models.Order, orderItems []*models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Update Order: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	orderQuery := `
		UPDATE orders
		SET customer_id = $1, total_amount = $2, special_instructions = $3, payment_method = $4, status = $5, updated_at = NOW() 
		WHERE id = $6
	`
	result, err := tx.Exec(orderQuery, order.CustomerID, order.TotalAmount, order.SpecialInstructions, order.PaymentMethod, order.Status, id)
	if err != nil {
		slog.Error("Repository error from Update Order: failed to update order", "id", id, "error", err)
		return err
	}

	if err := checkRowsAffected(result, id); err != nil {
		slog.Error("Repository error from Update Order: order not found", "id", id, "error", err)
		return err
	}

	if err = r.updateOrderItems(tx, id, orderItems); err != nil {
		slog.Error("Repository error from Update Order: failed update order items", "order id", id, "error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Update Order: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: order updated successfully", "id", id)
	return nil
}

func (r *OrderRepository) updateOrderItems(tx *sql.Tx, id int, orderItems []*models.OrderItem) error {
	itemDeleteQuery := `DELETE FROM order_items WHERE order_id = $1`

	if _, err := tx.Exec(itemDeleteQuery, id); err != nil {
		slog.Error("Repository error from update order items: failed to delete old order items", "order_id", id, "error", err)
		return err
	}

	itemQuery := `
		INSERT INTO order_items (order_id, quantity, price_at_order, menu_item_id)
		VALUES ($1, $2, $3, $4);
	`
	for _, item := range orderItems {
		if _, err := tx.Exec(itemQuery, id, item.Quantity, item.Price, item.MenuItemID); err != nil {
			slog.Error("Repository error from update order items: failed to add order item", "order_id", id, "menu_item_id", item.MenuItemID, "error", err)
			return err
		}
	}

	slog.Info("Repository info: order items updated successfully", "order id", id)
	return nil
}

func (r *OrderRepository) CloseOrderRepository(id int) error {
	query := `
	SELECT status FROM orders WHERE id = $1
	`

	var status string
	if err := r.db.QueryRow(query, id).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			slog.Error("Repository error from Close Order: order not found", "id", id)
			return apperrors.ErrNotExistConflict
		} else {
			slog.Error("Repository error from Close Order: failed to retrieve order", "id", id, "error", err)
			return fmt.Errorf("failed to check status order: %w", err)
		}
	}

	if status == "close" {
		slog.Error("Repository error from Close Order: order already closed", "id", id)
		return apperrors.ErrOrderClosed
	}

	queryUpdate := `
		UPDATE orders
		SET status = 'close', updated_at = NOW()
		WHERE id = $1 AND status = 'open'
	`

	_, err := r.db.Exec(queryUpdate, id)
	if err != nil {
		slog.Error("Repository error from Close Order: failed to close order", "id", id, "error", err)
		return err
	}

	slog.Info("Repository info: order closed successfully", "id", id)
	return nil
}

func (r *OrderRepository) DeleteOrderRepository(id int) error {
	query := `
		DELETE FROM orders
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("Repository error from Delete Order: failed to delete order", "order id", id, "error", err)
		return err
	}

	if err := checkRowsAffected(result, id); err != nil {
		slog.Error("Repository error from Delete Order: now rows affected", "order id", id, "error", err)
		return err
	}

	slog.Info("Repository info: order deleted successfully", "id", id)
	return nil
}

func (r *OrderRepository) NumberOfOrderedItemsRepository(startDate, endDate time.Time) (map[string]int, error) {
	query := `
		SELECT m.name, SUM(oi.quantity) AS count
		FROM order_items oi
		JOIN menu_items m ON oi.menu_item_id = m.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY m.name
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		slog.Error("Repository error from Number of Ordered Items: failed to retrieve ordered items", "error", err)
		return nil, err
	}
	defer rows.Close()

	orderedItems := make(map[string]int)
	for rows.Next() {
		var key string
		var value int
		if err := rows.Scan(&key, &value); err != nil {
			slog.Error("Repository error from Number of Ordered Items: failed to scan order row", "error", err)
			return nil, err
		}
		orderedItems[key] = value
	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository error from Number of Ordered Items: failed iterating over rows", "error", err)
		return nil, err
	}

	slog.Info("Repository info: retrieved ordered items successfully", "count", len(orderedItems))
	return orderedItems, nil
}

func (r *OrderRepository) AddOrdersRepository(orders []*models.Order, orderItems [][]*models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Add Orders: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	orderQuery := `
		INSERT INTO orders (customer_id, total_amount, status, special_instructions, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	itemQuery := `
		INSERT INTO order_items (order_id, quantity, price_at_order, menu_item_id)
		VALUES ($1, $2, $3, $4)
	`

	for i := range orders {
		var orderID int
		err = tx.QueryRow(orderQuery, orders[i].CustomerID, orders[i].TotalAmount, orders[i].Status, orders[i].SpecialInstructions, orders[i].PaymentMethod, orders[i].CreatedAt, orders[i].UpdatedAt).Scan(&orderID)
		if err != nil {
			slog.Error("Repository error from Add Orders: failed to add order", "customer_id", orders[i].CustomerID, "error", err)
			return err
		}

		for _, item := range orderItems[i] {
			_, err := tx.Exec(itemQuery, orderID, item.Quantity, item.Price, item.MenuItemID)
			if err != nil {
				slog.Error("Repository error from Add Orders: failed to add order item", "order_id", orderID, "menu_item_id", item.MenuItemID, "error", err)
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Add Orders: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: order added successfully", "count", len(orders))
	return nil
}
