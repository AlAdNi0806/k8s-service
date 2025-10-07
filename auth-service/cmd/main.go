// cmd/main.go
package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handler"
	authmw "auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/utils"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	utils.InitJWT(cfg.JWTSecret)

	// Подключение к MariaDB
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

	// Подключение к Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	// Инициализация сервисов
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, redisClient)

	// Echo
	e := echo.New()

	// Middleware: логирование (включает IP)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}, "latency":"${latency_human}", "ip":"${remote_ip}"}` + "\n",
	}))
	e.Use(middleware.Recover())

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

	// Health-check
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
	log.Fatal(e.Start(addr))
}
