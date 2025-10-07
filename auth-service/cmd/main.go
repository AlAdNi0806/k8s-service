package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handler"
	authmw "auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// OpenTelemetry Imports
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/redis/go-redis/v9"
)

const serviceName = "auth-service"
const serviceVersion = "1.0.0"

func main() {
	// Setup OpenTelemetry SDK (Traces, Metrics, Logs)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.Load()

	otelShutdown, err := utils.SetupOTelSDK(ctx, serviceName, serviceVersion, cfg.OtelExporterURL)
	if err != nil {
		log.Fatalf("Error setting up OpenTelemetry SDK: %v", err)
	}
	// Call cleanup on exit to ensure all spans/metrics are flushed
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
		if err != nil {
			log.Fatalf("Error shutting down OpenTelemetry: %v", err)
		}
	}()
	// End of OTel Setup

	// --- Your existing application logic starts here ---

	utils.InitJWT(cfg.JWTSecret)

	// Подключение к MariaDB (Consider instrumenting your SQL connection)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	// Подключение к Redis (Consider instrumenting your Redis client)
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
	defer redisClient.Close()

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		logger := utils.NewHelperLogger("auth-service.service.general")
		logger.LogError(ctx, "Could not get into redis", err)
	}

	// Инициализация сервисов
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, redisClient)

	// Echo
	e := echo.New()

	// ⭐️ ADD OPENTELEMETRY MIDDLEWARE HERE ⭐️
	e.Use(otelecho.Middleware(serviceName,
		// You can optionally skip tracing for certain endpoints like health checks
		otelecho.WithSkipper(func(c echo.Context) bool {
			return c.Path() == "/health" || c.Path() == "/metrics"
		}),
	))

	// Middleware: логирование (включает IP)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}, "latency":"${latency_human}", "ip":"${remote_ip}"}` + "\n",
	}))
	// e.Use(middleware.Recover())

	// Prometheus middleware — ДОБАВЛЕНО
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e) // регистрирует /metrics

	// Роуты
	authHandler := handler.NewAuthHandler(authService)
	e.POST("/register", authHandler.Register)
	e.POST("/login", authHandler.Login)

	// Защищённый эндпоинт
	e.GET("/me", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Authenticated!"})
	}, authmw.AuthMiddleware(authService))

	// Health-check (Skipped from tracing via WithSkipper above)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Порт
	port := "8081"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	addr := ":" + port

	log.Printf("Auth service starting on %s", addr)

	// Use the graceful shutdown context for Echo's Start as well
	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// Wait for an interrupt signal for graceful shutdown (like your initial example)
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	// Perform graceful shutdown of the Echo server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Echo server Shutdown error:", err)
	}
	log.Println("Echo server shut down.")
}
