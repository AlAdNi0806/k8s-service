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

	KafkaBrokers []string
	KafkaGroupID string
	KafkaTopic   string

	OtelExporterURL string
}

func Load() *Config {
	kafkaBrokers := []string{"localhost:9092"}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kafkaBrokers = []string{}
		for _, b := range strings.Split(brokers, ",") {
			kafkaBrokers = append(kafkaBrokers, strings.TrimSpace(b))
		}
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "inventory_user"),
		DBPassword: getEnv("DB_PASSWORD", "inventory_pass"),
		DBName:     getEnv("DB_NAME", "inventory_db"),

		KafkaBrokers: kafkaBrokers,
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "inventory-group"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "order.created"),

		OtelExporterURL: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
