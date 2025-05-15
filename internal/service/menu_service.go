package service

import (
	"fmt"
	"frappuchino/internal/apperrors"
	"frappuchino/internal/models"
	"log/slog"
)

// MenuRepository интерфейс определяет методы для работы с хранилищем меню
type MenuRepository interface {
	AddMenuItemRepository(menuItem models.MenuItem, menuItemIngredients []*models.MenuItemIngredient) error
	GetMenuItemRepository(id string) (*models.MenuItem, error)
	GetAllMenuItemsRepository() ([]*models.MenuItem, error)
	UpdateMenuItemRepository(id string, menuItem models.MenuItem, menuItemIngredients []*models.MenuItemIngredient) error
	DeleteMenuItemRepository(id string) error
}

// InventoryRepoForMenu интерфейс для доступа к инвентарю из сервиса меню
type InventoryRepoForMenu interface {
	GetAllInventoryItemsRepository() ([]*models.InventoryItem, error)
}

// MenuService реализует бизнес-логику для управления меню
type MenuService struct {
	menuRepo      MenuRepository
	inventoryRepo InventoryRepoForMenu
}

// NewMenuService создает новый экземпляр сервиса меню
func NewMenuService(mR MenuRepository, iD InventoryRepoForMenu) *MenuService {
	return &MenuService{
		menuRepo:      mR,
		inventoryRepo: iD,
	}
}

// CreateMenuItemService создает новый элемент меню и его ингредиенты
func (s *MenuService) CreateMenuItemService(menuItemRequest models.CreateMenuRequest) error {
	if err := s.validateMenuInventory(menuItemRequest.Ingredients); err != nil {
		slog.Error("Service error in Create Menu: failed to validate ingredients", "ingredients", menuItemRequest.Ingredients, "error", err)
		return err
	}

	menuItem, menuItemIngredients, err := s.createMenuObjects(menuItemRequest)
	if err != nil {
		slog.Error("Service error in Create Menu: failed to creating objects", "input item", menuItemRequest, "error", err)
		return err
	}

	err = s.menuRepo.AddMenuItemRepository(*menuItem, menuItemIngredients)
	if err != nil {
		slog.Error("Service error in Create Menu: failed to adding objects", "menu item", menuItem, "menu ingredients", menuItemIngredients, "error", err)
		return err
	}

	return nil
}

// GetAllMenuItemsService возвращает все элементы меню
func (s *MenuService) GetAllMenuItemsService() ([]*models.MenuItem, error) {
	menuItems, err := s.menuRepo.GetAllMenuItemsRepository()
	if err != nil {
		slog.Error("Service error in Create Menu: failed to retrieving all menu", "error", err)
		return nil, err
	}
	return menuItems, err
}

// GetMenuItemService возвращает элемент меню по ID
func (s *MenuService) GetMenuItemService(id string) (*models.MenuItem, error) {
	menuItem, err := s.menuRepo.GetMenuItemRepository(id)
	if err != nil {
		slog.Error("Service error in Get Menu: failed to retrieving menu item", "id", id, "error", err)
		return nil, err
	}
	return menuItem, err
}

// UpdateMenuItemService обновляет существующий элемент меню
func (s *MenuService) UpdateMenuItemService(id string, menuItemRequest models.CreateMenuRequest) error {
	if err := s.validateMenuInventory(menuItemRequest.Ingredients); err != nil {
		slog.Error("Service error in Update Menu: failed to validate ingredients", "ingredients", menuItemRequest.Ingredients, "error", err)
		return err
	}

	menuItemRequest.ID = id
	menuItem, menuItemIngredients, err := s.createMenuObjects(menuItemRequest)
	if err != nil {
		slog.Error("Service error in Update Menu: failed to create objects", "input item", menuItemRequest, "error", err)
		return err
	}

	err = s.menuRepo.UpdateMenuItemRepository(id, *menuItem, menuItemIngredients)
	if err != nil {
		slog.Error("Service error in Update Menu: failed to update objects", "menu item", menuItem, "menu ingredients", menuItemIngredients, "error", err)
		return err
	}

	return nil
}

// DeleteMenuItemService удаляет элемент меню по ID
func (s *MenuService) DeleteMenuItemService(id string) error {
	err := s.menuRepo.DeleteMenuItemRepository(id)
	if err != nil {
		slog.Error("Service error in Delete Menu: failed to delete item", "id", id, "error", err)
		return err
	}
	return nil
}

// validateMenuInventory проверяет наличие всех ингредиентов в инвентаре
func (s *MenuService) validateMenuInventory(ingredients []models.MenuItemIngredientInput) error {
	inventory, err := s.inventoryRepo.GetAllInventoryItemsRepository()
	if err != nil {
		slog.Error("Service error in validate Menu Inventory: there are no ingredients", "ingredients", ingredients, "error", err)
		return err
	}

	inventMap := make(map[string]*models.InventoryItem)
	for _, item := range inventory {
		inventMap[item.ID] = item
	}

	for _, ingredient := range ingredients {
		if _, exists := inventMap[ingredient.IngredientID]; !exists {
			slog.Error("Service error in validate ingredients: doesn't exist", "ingredient ID", ingredient.IngredientID)
			return fmt.Errorf("%w", apperrors.ErrNotExistConflict)
		}
	}

	return nil
}

// indentAllergens определяет аллергены на основе списка ингредиентов
func (s *MenuService) indentAllergens(items []models.MenuItemIngredientInput) ([]string, error) {
	allergensMap := map[string][]string{
		"gluten":    {"wheat", "barley", "rye"},
		"lactose":   {"milk", "cheese", "butter"},
		"nuts":      {"almond", "walnut", "hazelnut"},
		"soy":       {"soy", "soybean", "tofu"},
		"shellfish": {"shrimp", "crab", "lobster"},
		"eggs":      {"egg", "albumin"},
		"fish":      {"salmon", "tuna", "cod"},
		"peanuts":   {"peanut", "groundnut"},
		"sesame":    {"sesame", "tahini"},
		"mustard":   {"mustard", "mustard seed"},
		"caffeine":  {"espresso", "coffee"},
	}

	var allergens []string
	for _, item := range items {
		for allergen, keywords := range allergensMap {
			found := false
			for _, keyword := range keywords {
				if item.IngredientID == keyword {
					allergens = append(allergens, allergen)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	return allergens, nil
}

// createMenuObjects создает объекты элемента меню и его ингредиентов
func (s *MenuService) createMenuObjects(menuItemRequest models.CreateMenuRequest) (*models.MenuItem, []*models.MenuItemIngredient, error) {
	allergens, err := s.indentAllergens(menuItemRequest.Ingredients)
	if err != nil {
		slog.Error("Service error in create menu objects: failed to ident allergens", "error", err)
		return nil, nil, err
	}

	menuItem, err := models.NewMenuItem(allergens, menuItemRequest)
	if err != nil {
		return nil, nil, err
	}

	menuItemIngredients, err := models.NewMenuItemIngredients(menuItem.ID, menuItemRequest.Ingredients)
	if err != nil {
		return nil, nil, err
	}

	return menuItem, menuItemIngredients, nil
}
