// cmd/main.go
package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handler"
	authmw "auth-service/internal/middleware" // ← алиас для вашего middleware
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/utils"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware" // ← алиас для echo middleware
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func newOTelProvider(ctx context.Context, endpoint string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("auth-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	cfg := config.Load()
	utils.InitJWT(cfg.JWTSecret)

	// PostgreSQL
	pgOpts := pg.Options{
		Addr:     cfg.DBHost + ":" + cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
	}
	db := pg.Connect(&pgOpts)
	defer db.Close()

	// Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	// Auth Service
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, redisClient)

	// OpenTelemetry
	ctx := context.Background()
	tp, err := newOTelProvider(ctx, cfg.OtelExporterURL)
	if err != nil {
		log.Fatal("Failed to create OTel provider:", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shutdown OTel:", err)
		}
	}()

	// Echo
	e := echo.New()

	// Middleware
	e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}, "latency":"${latency_human}", "ip":"${remote_ip}"}` + "\n",
	}))
	e.Use(echomw.Recover())

	// OpenTelemetry instrumentation
	e.Use(otelecho.Middleware("auth-service"))

	// Routes
	authHandler := handler.NewAuthHandler(authService)
	e.POST("/register", authHandler.Register)
	e.POST("/login", authHandler.Login)

	// Защищённый эндпоинт для теста
	e.GET("/me", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Authenticated!"})
	}, authmw.AuthMiddleware(authService)) // ← ваш middleware

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Запуск
	addr := ":8081"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("Auth service starting on %s", addr)
	log.Fatal(e.Start(addr))
}
