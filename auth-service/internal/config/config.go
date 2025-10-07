package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisAddr string

	JWTSecret string

	OtelExporterURL string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "192.168.0.176"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "auth_user"),
		DBPassword: getEnv("DB_PASSWORD", "auth_pass"),
		DBName:     getEnv("DB_NAME", "auth_db"),

		RedisAddr: getEnv("REDIS_ADDR", "192.168.0.176:6379"),

		JWTSecret: getEnv("JWT_SECRET", "super-secret-jwt-key"),

		OtelExporterURL: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.0.176:4318"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
