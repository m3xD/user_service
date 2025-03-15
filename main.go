package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"user_service/api/middleware"
	"user_service/internal/delivery/rest"
	"user_service/internal/repository/postgres"
	"user_service/internal/service"
	"user_service/internal/util"
	pkg "user_service/pkg/logger"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "user_service/docs"
)

// @title           User Service API
// @version         1.0
// @description     API for user management.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Bui Duy Khanh
// @contact.email  iam@m3xd.dev

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      auth-service-6f3ceb0b5b52.herokuapp.com
// @BasePath  /

//  @securityDefinitions.apiKey  JWT
//  @in                          header
//  @name                        Authorization
//  @description                 JWT security accessToken. Please add it in the format "Bearer {token}"

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	// Initialize logger
	logger := pkg.NewLogger().Logger

	// Database configuration
	/*dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "user_db")*/

	os.Setenv("SECRET_KEY", "sap-secrets")

	logger.Info("User service starting", zap.String("version", "1.0.0"))

	// Connect to database
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/user_db?sslmode=disable")

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
	logger.Info("Successfully connected to database")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)

	// jwt service
	jwtService := util.NewJwtImpl()

	// Initialize services
	userService := service.NewUserService(userRepo, logger)
	authService := service.NewAuthService(userRepo, jwtService, logger)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	port := getEnv("PORT", "8083")

	// route to docs
	// link, _ := strings.CutSuffix(getEnv("SYS_URL", ""), "/api")

	// Setup router with logging middleware
	router := mux.NewRouter()
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(getEnv("SYS_URL", "http://localhost:"+port)+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)
	router.Use(middleware.NewLogMiddleware(logger).LoggingMiddleware)

	// Initialize handlers
	userHandler := rest.NewUserHandler(userService, logger, authMiddleware)
	authHandler := rest.NewAuthHandler(authService, userService, router, logger, *jwtService)

	// Register routes
	userHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes()

	fmt.Println(os.Getenv("SECRET_KEY"))
	// Start server
	logger.Info("Server starting", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)(router)); err != nil {
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
