package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	JWTSecret       string
	AllowedOrigins  string
	MaintenanceMode bool
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env") // .env production ortamında olmayabilir, hata kasıtlı olarak görmezden gelinir

	// Cloud Run PORT env var'ını da destekle
	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = os.Getenv("APP_PORT")
	}

	cfg := &Config{
		AppPort:         appPort,
		DBHost:          os.Getenv("DB_HOST"),
		DBPort:          os.Getenv("DB_PORT"),
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AllowedOrigins:  os.Getenv("ALLOWED_ORIGINS"),
		MaintenanceMode: os.Getenv("MAINTENANCE_MODE") == "true",
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("eksik zorunlu DB config degiskenleri")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET zorunludur")
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	return cfg, nil
}