package db

import (
	"database/sql"
	"fmt"
)

// Создаёт готовое соединение с базой данных с которым впредь можно будет работать
func InitDataBase(dbHost, dbPort, dbUser, dbPassword, dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}
