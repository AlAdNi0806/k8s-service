// cmd/main.go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/internal/utils"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

const serviceName = "order-service"
const serviceVersion = "1.0.0"

func main() {
	// Setup OpenTelemetry SDK
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.Load()

	otelShutdown, err := utils.SetupOTelSDK(ctx, serviceName, serviceVersion, cfg.OtelExporterURL)
	if err != nil {
		log.Fatalf("Error setting up OpenTelemetry SDK: %v", err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
		if err != nil {
			log.Fatalf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

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

	// Kafka
	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers...),
		Topic:        "order.created",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	defer kafkaWriter.Close()

	// Order Service
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, kafkaWriter)

	// Echo
	e := echo.New()

	// OpenTelemetry Middleware
	e.Use(otelecho.Middleware(serviceName,
		otelecho.WithSkipper(func(c echo.Context) bool {
			return c.Path() == "/health"
		}),
	))

	// Custom logger with IP
	e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}, "latency":"${latency_human}", "ip":"${remote_ip}"}` + "\n",
	}))
	e.Use(echomw.Recover())

	// Auth middleware
	authMid := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.ErrUnauthorized
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.ErrUnauthorized
			}
			userID, err := orderService.ValidateToken(parts[1])
			if err != nil {
				return echo.ErrUnauthorized
			}
			c.Set("user_id", userID)
			return next(c)
		}
	}

	// Routes
	orderHandler := handler.NewOrderHandler(orderService)
	e.POST("/orders", orderHandler.CreateOrder, authMid)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	addr := ":8082"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("Order service starting on %s", addr)

	// Graceful shutdown
	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Echo server Shutdown error:", err)
	}
	log.Println("Echo server shut down.")
}
