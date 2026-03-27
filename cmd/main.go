package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Posinowa/FinbudApp/internal/auth"
	"github.com/Posinowa/FinbudApp/pkg/config"
	"github.com/Posinowa/FinbudApp/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config yuklenemedi: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("DB baglantisi kurulamadi: %v", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(r)

	log.Printf("Sunucu :%s portunda baslatiliyor...", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Sunucu baslatılamadi: %v", err)
	}
}