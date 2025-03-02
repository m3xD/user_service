package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
	"user_service/api/middleware"
	"user_service/internal/delivery/rest"
	"user_service/internal/repository/postgres"
	"user_service/internal/service"
	pkg "user_service/pkg/logger"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger
	logger := pkg.NewLogger().Logger

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "changeme")
	dbName := getEnv("DB_NAME", "user_db")

	logger.Info("User service starting", zap.String("version", "1.0.0"))

	// Connect to database
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Successfully connected to database", zap.String("host", dbHost), zap.String("port", dbPort))

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, logger)

	// Initialize handlers
	userHandler := rest.NewUserHandler(userService, logger)

	// Setup router with logging middleware
	router := mux.NewRouter()
	router.Use(middleware.NewLogMiddleware(logger).LoggingMiddleware)

	// Register routes
	userHandler.RegisterRoutes(router)

	// Start server
	port := getEnv("PORT", "8083")
	logger.Info("Server starting", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
