package repository

import (
	"database/sql"
	"fmt"
	"frappuchino/internal/apperrors"
	"frappuchino/internal/models"
	"log/slog"

	_ "github.com/lib/pq"
)

type InventoryRepository struct {
	db *sql.DB // База данных
}

// Создает новый экземпляр InventoryRepository
func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

// Закрывает подключение к базе данных
func (r *InventoryRepository) Close() error {
	return r.db.Close()
}

// Добавляет новый элемент в инвентарь и фиксирует транзакцию
func (r *InventoryRepository) AddInventoryItemRepository(inventoryItem models.InventoryItem, inventoryTransaction models.InventoryTransaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Add Inventory: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback() // Откат транзакции в случае ошибки

	// Вставляем новый элемент в инвентарь
	orderQuery := `
		INSERT INTO inventory (id, name, stock, price, unit_type, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err = tx.Exec(orderQuery, inventoryItem.ID, inventoryItem.Name, inventoryItem.StockLevel, inventoryItem.Price, inventoryItem.UnitType, inventoryItem.LastUpdated)
	if err != nil {
		slog.Error("Repository error from Add Inventory: failed to add inventory", "inventory ID", inventoryItem.ID, "error", err)
		return err
	}

	// Вставляем транзакцию изменения инвентаря
	itemQuery := `
		INSERT INTO inventory_transactions (inventory_id, change_amount, transaction_type, changed_at)
		VALUES ($1, $2, $3, $4);
	`
	_, err = tx.Exec(itemQuery, inventoryTransaction.InventoryID, inventoryTransaction.ChangeAmount, inventoryTransaction.TransactionType, inventoryTransaction.ChangeAt)
	if err != nil {
		slog.Error("Repository error from Add Inventory: failed to add inventory item", "inventory_id", inventoryItem.ID, "error", err)
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Add Inventory: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: inventory item added successfully", "inventory_id", inventoryItem.ID)
	return nil
}

// Получает все элементы инвентаря
func (r *InventoryRepository) GetAllInventoryItemsRepository() ([]*models.InventoryItem, error) {
	query := `
		SELECT id, name, stock, price, unit_type, last_updated
		FROM inventory;
	`

	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("Repository error from Get Inventory: failed to retrieve all inventory", "error", err)
		return nil, err
	}
	defer rows.Close()

	var inventoryItems []*models.InventoryItem
	for rows.Next() {
		var inventoryItem models.InventoryItem
		if err := rows.Scan(&inventoryItem.ID, &inventoryItem.Name, &inventoryItem.StockLevel, &inventoryItem.Price, &inventoryItem.UnitType, &inventoryItem.LastUpdated); err != nil {
			slog.Error("Repository error from Get Inventory: failed to scan inventory row", "error", err)
			return nil, err
		}
		inventoryItems = append(inventoryItems, &inventoryItem)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository error from Get Inventory: failed iterating over rows", "error", err)
		return nil, err
	}

	slog.Info("Repository info: retrieved all orders successfully", "count", len(inventoryItems))
	return inventoryItems, nil
}

// Получает элемент инвентаря по ID
func (r *InventoryRepository) GetInventoryItemRepository(id string) (*models.InventoryItem, error) {
	query := `
	SELECT id, name, stock, price, unit_type, last_updated
	FROM inventory
	WHERE id = $1;
	`

	var inventoryItem models.InventoryItem
	err := r.db.QueryRow(query, id).Scan(&inventoryItem.ID, &inventoryItem.Name, &inventoryItem.StockLevel, &inventoryItem.Price, &inventoryItem.UnitType, &inventoryItem.LastUpdated)
	if err == sql.ErrNoRows {
		slog.Error("Repository error from Get Inventory: no inventory found", "id", id)
		return nil, apperrors.ErrNotExistConflict
	} else if err != nil {
		slog.Error("Repository error from Get Inventory: failed to retrieve inventory", "id", id, "error", err)
		return nil, err
	}

	slog.Info("Repository info: retrieved inventory successfully", "id", id)
	return &inventoryItem, nil
}

// Обновляет элемент инвентаря и фиксирует транзакцию
func (r *InventoryRepository) UpdateInventoryItemRepository(id string, inventoryItem models.InventoryItem, inventoryTransaction models.InventoryTransaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Update Inventory: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	// Обновляем данные инвентаря
	itemQuery := `
		UPDATE inventory
		SET name = $1, stock = stock + $2, unit_type = $3, price = $4, last_updated = NOW()
		WHERE id = $5
	`
	result, err := tx.Exec(itemQuery, inventoryItem.Name, inventoryItem.StockLevel, inventoryItem.UnitType, inventoryItem.Price, id)
	if err != nil {
		slog.Error("Repository error from Update Inventory: failed to update inventory", "id", id, "error", err)
		return err
	}

	// Проверяем, что обновление затронуло хотя бы одну строку
	if err := checkRowsAffected(result, id); err != nil {
		slog.Error("Repository error from Update Inventory: inventory not found", "id", id, "error", err)
		return err
	}

	// Вставляем транзакцию изменения инвентаря
	transactionQuery := `
	INSERT INTO inventory_transactions (inventory_id, change_amount, transaction_type, changed_at)
	VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(transactionQuery, id, inventoryTransaction.ChangeAmount, inventoryTransaction.TransactionType, inventoryTransaction.ChangeAt)
	if err != nil {
		slog.Error("Repository error from Update Inventory: failed to update inventory_transactions", "id", id, "error", err)
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Update Inventory: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: inventory updated successfully", "id", id)
	return nil
}

// Удаляет элемент из инвентаря
func (r *InventoryRepository) DeleteInventoryItemRepository(id string) error {
	query := `
		DELETE FROM inventory
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("Repository error from Delete Inventory: failed to delete inventory", "id", id, "error", err)
		return err
	}

	// Проверяем, что удаление затронуло хотя бы одну строку
	if err := checkRowsAffected(result, id); err != nil {
		slog.Error("Repository error from Delete Inventory: inventory not found", "id", id, "error", err)
		return err
	}

	slog.Info("Repository info: inventory deleted successfully", "id", id)
	return nil
}

// Обновляет инвентарь при продаже
func (r *InventoryRepository) UpdateInventoryForSale(quantities map[string]float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Update Inventory for Sale: failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	// Для каждого ингредиента обновляем количество и записываем транзакцию
	for ingredientID, quantity := range quantities {
		updateInventoryQuery := `
			UPDATE inventory
			SET stock = stock - $1, last_updated = NOW()
			WHERE id = $2
		`
		_, err = tx.Exec(updateInventoryQuery, quantity, ingredientID)
		if err != nil {
			slog.Error("Repository error from Update Inventory for Sale: failed to update inventory", "ingredient ID", ingredientID, "error", err)
			return err
		}

		insertTransactionQuery := `
			INSERT INTO inventory_transactions (inventory_id, change_amount, transaction_type, changed_at)
			VALUES ($1, $2, $3, $4)
		`
		transaction, err := models.NewInventoryTransaction(ingredientID, quantity, "sale")
		if err != nil {
			slog.Error("Repository error from Update Inventory for Sale: invalid input data", "ingredient ID", ingredientID, "error", err)
			return err
		}

		_, err = tx.Exec(insertTransactionQuery, transaction.InventoryID, transaction.ChangeAmount, transaction.TransactionType, transaction.ChangeAt)
		if err != nil {
			slog.Error("Repository error from Update Inventory for Sale: failed to insert transaction", "ingredient ID", transaction.InventoryID, "error", err)
			return err
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Update Inventory for Sale: failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Repository info: inventory update for sale successfully")
	return nil
}

// Получает остатки инвентаря с пагинацией и сортировкой
func (r *InventoryRepository) GetLeftOversRepository(sortBy string, page, offset, pageSize int) (map[string]interface{}, error) {
	query := `
		SELECT name, stock, price
		FROM inventory
		ORDER BY 
    	CASE 
        WHEN $1 = '' THEN NULL
        WHEN $1 = 'price' THEN price
        WHEN $1 = 'quantity' THEN stock
    	END
		OFFSET $2
		LIMIT $3;
		`

	rows, err := r.db.Query(query, sortBy, offset, pageSize)
	if err != nil {
		slog.Error("Repository error from Get Leftovers: failed to retrieve leftovers", "sort by", sortBy, "offset", offset, "page size", pageSize, "error", err)
		return nil, err
	}
	defer rows.Close()

	var leftovers []map[string]interface{}
	for rows.Next() {
		var name string
		var stock, price float64
		if err := rows.Scan(&name, &stock, &price); err != nil {
			slog.Error("Repository error from Get Leftovers: failed to scan leftovers row", "error", err)
			return nil, err
		}
		leftovers = append(leftovers, map[string]interface{}{
			"name":     name,
			"quantity": stock,
			"price":    price,
		})
	}

	var totalItems int
	err = r.db.QueryRow("SELECT COUNT(*) FROM inventory").Scan(&totalItems)
	if err != nil {
		slog.Warn("Repository error from Get Leftovers: failed to retrieve total items", "error", err)
		return nil, err
	}
	totalPages := (totalItems + pageSize - 1) / pageSize
	hasNextPage := page < totalPages

	response := map[string]interface{}{
		"currentPage": page,
		"hasNextPage": hasNextPage,
		"pageSize":    pageSize,
		"totalPages":  totalPages,
		"data":        leftovers,
	}

	return response, nil
}
