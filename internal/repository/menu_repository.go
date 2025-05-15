package repository

import (
	"database/sql"
	"frappuchino/internal/apperrors"
	"frappuchino/internal/models"
	"log/slog"

	"github.com/lib/pq"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{
		db: db,
	}
}

func (r *MenuRepository) Close() error {
	return r.db.Close()
}

func (r *MenuRepository) AddMenuItemRepository(menuItem models.MenuItem, menuItemIngredients []*models.MenuItemIngredient) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Add Menu: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	orderQuery := `
		INSERT INTO menu_items (id, name, description, price, allergens, size)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(orderQuery, menuItem.ID, menuItem.Name, menuItem.Description, menuItem.Price, pq.Array(menuItem.Allergens), menuItem.Size)
	if err != nil {
		slog.Error("Repository error from Add Menu: failed to add menu", "menu_id", menuItem.ID, "error", err)
		return err
	}

	itemQuery := `
		INSERT INTO menu_item_ingredients (menu_item_id, quantity, ingredient_id)
		VALUES ($1, $2, $3)
	`
	for _, item := range menuItemIngredients {
		_, err := tx.Exec(itemQuery, item.MenuItemID, item.Quantity, item.IngredientID)
		if err != nil {
			slog.Error("Repository error from Add Menu: failed to add menu item", "menu_item_id", item.MenuItemID, "ingredient_id", item.IngredientID, "error", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Add Menu: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: menu item added successfully", "order_id", menuItem.ID)
	return nil
}

func (r *MenuRepository) GetAllMenuItemsRepository() ([]*models.MenuItem, error) {
	query := `
		SELECT id, name, description, price, allergens, size
		FROM menu_items
	`

	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("Repository error from Get Menu: failed to retrieve all menu items", "error", err)
		return nil, err
	}
	defer rows.Close()

	var menuItems []*models.MenuItem
	for rows.Next() {
		var menuItem models.MenuItem
		if err := rows.Scan(&menuItem.ID, &menuItem.Name, &menuItem.Description, &menuItem.Price, pq.Array(&menuItem.Allergens), &menuItem.Size); err != nil {
			slog.Error("Repository error from Get Menu: failed to scan menu item row", "error", err)
			return nil, err
		}
		menuItems = append(menuItems, &menuItem)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository error from Get Menu: failed iterating over rows", "error", err)
		return nil, err
	}

	slog.Info("Repository info: retrieved all menu items successfully", "count", len(menuItems))
	return menuItems, nil
}

func (r *MenuRepository) GetMenuItemRepository(id string) (*models.MenuItem, error) {
	query := `
		SELECT id, name, description, price, allergens, size
		FROM menu_items
		WHERE id = $1
	`
	var menuItem models.MenuItem
	err := r.db.QueryRow(query, id).Scan(&menuItem.ID, &menuItem.Name, &menuItem.Description, &menuItem.Price, pq.Array(&menuItem.Allergens), &menuItem.Size)
	if err == sql.ErrNoRows {
		slog.Error("Repository error from Get Menu: menu item not found", "id", id)
		return nil, apperrors.ErrNotExistConflict
	} else if err != nil {
		slog.Error("Repository error from Get Menu: failed to retrieve menu item", "id", id, "error", err)
		return nil, err
	}

	slog.Info("Repository info: menu item retrieved successfully", "id", id)
	return &menuItem, nil
}

func (r *MenuRepository) UpdateMenuItemRepository(id string, menuItem models.MenuItem, menuItemIngredients []*models.MenuItemIngredient) error {
	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("Repository error from Update Menu: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	if err := r.addPriceHistory(tx, id, menuItem.Price); err != nil {
		slog.Error("Repository error from Update Menu: failed add price history", "menu id", menuItem.ID, "error", err)
		return err
	}

	itemQuery := `
		UPDATE menu_items
		SET name = $1, description = $2, price = $3, allergens = $4, size = $5
		WHERE id = $6;
	`
	result, err := tx.Exec(itemQuery, menuItem.Name, menuItem.Description, menuItem.Price, pq.Array(menuItem.Allergens), menuItem.Size, id)
	if err != nil {
		slog.Error("Repository error from Update Menu: failed to update menu item", "id", id, "error", err)
		return err
	}

	if err := checkRowsAffected(result, id); err != nil {
		slog.Error("Repository error from Update Menu: menu item not found", "id", id, "error", err)
		return err
	}

	if err := r.updateMenuItemIngredients(tx, id, menuItemIngredients); err != nil {
		slog.Error("Repository error from Update Menu: failed update menu ingredients", "menu id", id, "error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Repository error from Update Menu: failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Repository info: menu item updated successfully", "id", id)
	return nil
}

func (r *MenuRepository) addPriceHistory(tx *sql.Tx, id string, newPrice float64) error {
	var oldPrice float64
	priceQuery := `SELECT price FROM menu_items WHERE id = $1`
	if err := tx.QueryRow(priceQuery, id).Scan(&oldPrice); err != nil {
		slog.Error("Repository error from add price history: failed to fetch current price", "id", id, "error", err)
		return err
	}

	priceHistoryQuery := `
		INSERT INTO price_history (menu_item_id, old_price, new_price, changed_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := tx.Exec(priceHistoryQuery, id, oldPrice, newPrice)
	if err != nil {
		slog.Error("Repository error from add price history: failed to insert price history", "menu_item_id", id, "error", err)
		return err
	}
	slog.Info("Repository info: price history add successfully", "menu item ID", id)
	return nil
}

func (r *MenuRepository) updateMenuItemIngredients(tx *sql.Tx, id string, ingredients []*models.MenuItemIngredient) error {
	ingredientDeleteQuery := `DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`
	if _, err := tx.Exec(ingredientDeleteQuery, id); err != nil {
		slog.Error("Repository error from update menu ingredients: failed to delete old menu item ingredients", "menu_item_id", id, "error", err)
		return err
	}

	ingredientInsertQuery := `
		INSERT INTO menu_item_ingredients (menu_item_id, quantity, ingredient_id)
		VALUES ($1, $2, $3)
	`
	for _, item := range ingredients {
		if _, err := tx.Exec(ingredientInsertQuery, id, item.Quantity, item.IngredientID); err != nil {
			slog.Error("Repository error from update menu ingredients: failed to insert updated menu item ingredients", "menu_item_id", id, "ingredient_id", item.IngredientID, "error", err)
			return err
		}
	}

	slog.Info("Repository info: update menu ingredients successfully", "count", len(ingredients))
	return nil
}

func (r *MenuRepository) DeleteMenuItemRepository(id string) error {
	query := `
		DELETE FROM menu_items
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("Repository error from Delete Menu: failed to delete menu item", "id", id, "error", err)
		return err
	}

	if err := checkRowsAffected(result, id); err != nil {
		slog.Error("Repository error from Delete Menu: menu item not found", "id", id, "error", err)
		return err
	}

	slog.Info("Repository info: menu item deleted successfully", "id", id)
	return nil
}

func (r *MenuRepository) GetMenuItemsAndPrice(productIDs []string) (map[string]float64, error) {
	query := "SELECT id, price FROM menu_items WHERE id = ANY($1)"
	rows, err := r.db.Query(query, pq.Array(productIDs))
	if err != nil {
		slog.Error("Repository error from Get Menu and Price: failed to fetch menu items", "error", err)
		return nil, err
	}
	defer rows.Close()

	menuItems := make(map[string]float64)
	for rows.Next() {
		var id string
		var price float64
		if err := rows.Scan(&id, &price); err != nil {
			slog.Error("Repository error from Get Menu and Pice: failed to scan menu item", "error", err)
			return nil, err
		}
		menuItems[id] = price
	}

	slog.Info("Repository info: menu items and Price retrieved successfully")
	return menuItems, nil
}

func (r *MenuRepository) CalculateIngredientsForOrder(menuQuantities map[string]int) (map[string]float64, error) {
	menuItemIDs := make([]string, 0, len(menuQuantities))
	for menuID := range menuQuantities {
		menuItemIDs = append(menuItemIDs, menuID)
	}

	query := `
		SELECT menu_item_id, ingredient_id, quantity
		FROM menu_item_ingredients
		WHERE menu_item_id = ANY($1)
	`
	rows, err := r.db.Query(query, pq.Array(menuItemIDs))
	if err != nil {
		slog.Error("Repository error from Calculate Ingredients for Order: failed to fetch ingredients for menu items", "error", err)
		return nil, err
	}
	defer rows.Close()

	ingredients := make(map[string]float64)

	for rows.Next() {
		var menuItemID string
		var ingredientID string
		var amountRequired float64

		if err := rows.Scan(&menuItemID, &ingredientID, &amountRequired); err != nil {
			slog.Error("Repository error from Calculate Ingredients for Order: failed to scan row for ingredients", "error", err)
			return nil, err
		}

		if quantity, exists := menuQuantities[menuItemID]; exists {
			ingredients[ingredientID] += amountRequired * float64(quantity)
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository error from Calculate Ingredients for Order: error while iterating over rows", "error", err)
		return nil, err
	}

	slog.Info("Repository info: calculate ingredients and price successfully")
	return ingredients, nil
}
