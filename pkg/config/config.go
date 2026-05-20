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
	UpdateAvailable bool
	AndroidStoreURL string
	IOSStoreURL     string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPassword    string
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
		UpdateAvailable: os.Getenv("UPDATE_AVAILABLE") == "true",
		AndroidStoreURL: os.Getenv("ANDROID_STORE_URL"),
		IOSStoreURL:     os.Getenv("IOS_STORE_URL"),
		SMTPHost:        getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:        getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser:        os.Getenv("SMTP_USER"),
		SMTPPassword:    os.Getenv("SMTP_PASSWORD"),
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

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}