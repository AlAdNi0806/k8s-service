// cmd/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-service/internal/config"
	"inventory-service/internal/consumer"
	"inventory-service/internal/repository"
	"inventory-service/internal/service"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
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
			semconv.ServiceName("inventory-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	cfg := config.Load()

	// PostgreSQL
	pgOpts := pg.Options{
		Addr:     cfg.DBHost + ":" + cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
	}
	db := pg.Connect(&pgOpts)
	defer db.Close()

	// OpenTelemetry
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := newOTelProvider(ctx, cfg.OtelExporterURL)
	if err != nil {
		log.Fatal("Failed to create OTel provider:", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shutdown OTel:", err)
		}
	}()

	// Repository & Service
	repo := repository.NewInventoryRepository(db)
	invService := service.NewInventoryService(repo)

	// Предзаполним склад для демо (в реальности — отдельный сервис каталога)
	repo.EnsureStock(ctx, 123, 100) // product_id=123, qty=100

	// Kafka Consumer
	kafkaConsumer := consumer.NewKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, cfg.KafkaTopic, invService)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP-сервера (для health-check и метрик)
	e := echo.New()
	e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}, "latency":"${latency_human}", "ip":"${remote_ip}"}` + "\n",
	}))
	e.Use(echomw.Recover())
	e.Use(otelecho.Middleware("inventory-service"))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok", "service": "inventory"})
	})

	// Запуск в горутинах
	go func() {
		if err := e.Start(":8083"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go kafkaConsumer.Start(ctx)

	log.Println("Inventory service started")

	// Ожидание сигнала завершения
	<-sigCh
	log.Println("Shutting down...")

	// Закрытие consumer
	kafkaConsumer.Close()

	// Закрытие HTTP сервера
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	e.Shutdown(ctxShutdown)
}
