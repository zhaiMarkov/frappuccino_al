package main

import (
	"frappuchino/internal/config"
	"frappuchino/internal/db"
	"frappuchino/internal/router"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	// Setting up Config: Подготовить все настройки базы данных. Порт подключения, хост, API и так далее
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Configuration loading error", "error", err)
		os.Exit(1)
	}
	slog.Info("Configuration loaded successfully")

	// Уже подключиться к базе данных
	dataBase, err := db.InitDataBase(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		slog.Error("Database connection failed", "host", cfg.DBHost, "port", cfg.DBPort, "error", err)
		os.Exit(1)
	}
	defer dataBase.Close()
	slog.Info("Database connection successfully")

	// Подготовить енд пойнты
	mux, err := router.LoadRoutes(dataBase)
	if err != nil {
		slog.Error("Failed to set up routes", "error", err)
		os.Exit(1)
	}
	slog.Info("Setap router successfully")

	slog.Info("Starting server", "port", cfg.APIPort)
	if err := http.ListenAndServe(":"+cfg.APIPort, mux); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
