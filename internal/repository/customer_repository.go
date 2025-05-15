package repository

import (
	"database/sql"
	"encoding/json"
	"frappuchino/internal/models"
	"log/slog"
)

type CustomerRepository struct {
	db *sql.DB // База данных
}

// Создает новый экземпляр CustomerRepository
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

// Закрывает подключение к базе данных
func (r *CustomerRepository) Close() error {
	return r.db.Close()
}

// Находит или создает клиента, возвращая его ID
func (r *CustomerRepository) IndentCustomerID(customerName string, instructions json.RawMessage) (int, error) {
	var customerID int
	query := `
		SELECT id
		FROM customers
		WHERE name = $1
	`
	// Ищем клиента в базе
	err := r.db.QueryRow(query, customerName).Scan(&customerID)
	if err != nil {
		if err == sql.ErrNoRows { // Если клиент не найден, создаем нового
			insertQuery := `
				INSERT INTO customers (name, email, preferences)
				VALUES ($1, $2, $3)
				RETURNING id
			`
			customer, err := models.NewCustomer(customerName, instructions)
			if err != nil {
				slog.Error("Repository error from Ident Customer ID: invalid input data", "customer name", customerName, "error", err)
				return 0, err
			}

			// Вставляем нового клиента в базу данных
			if err := r.db.QueryRow(insertQuery, customer.Name, customer.Email, customer.Preferences).Scan(&customerID); err != nil {
				slog.Error("Repository error from Ident Customer ID: failed to insert into table", "customer", customer, "error", err)
				return 0, err
			}
		} else {
			// Ошибка при выборке клиента из базы данных
			slog.Error("Repository error from Ident Customer ID: failed to select from table", "customer name", customerName, "error", err)
			return 0, err
		}
	}

	slog.Info("Repository info: ident customer ID successfully", "customer ID", customerID)
	return customerID, nil
}
