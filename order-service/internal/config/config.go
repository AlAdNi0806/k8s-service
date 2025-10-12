// internal/config/config.go
package config

import (
	"os"
	"strings"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string

	KafkaBrokers []string

	OtelExporterURL string
}

func Load() *Config {
	kafkaBrokers := []string{"192.168.0.176:9092"}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kafkaBrokers = []string{}
		for _, b := range strings.Split(brokers, ",") {
			kafkaBrokers = append(kafkaBrokers, strings.TrimSpace(b))
		}
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "192.168.0.176"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "order_user"),
		DBPassword: getEnv("DB_PASSWORD", "order_pass"),
		DBName:     getEnv("DB_NAME", "order_db"),

		JWTSecret: getEnv("JWT_SECRET", "super-secret-jwt-key"),

		KafkaBrokers: kafkaBrokers,

		OtelExporterURL: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "192.168.0.176:4317"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
