package main

import (
	"log"
	"net/http"

	"github.com/Posinowa/FinbudApp/pkg/config"
	"github.com/Posinowa/FinbudApp/pkg/database"
	"github.com/Posinowa/FinbudApp/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config yuklenemedi: %v", err)
	}

	_, err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("DB baglantisi kurulamadi: %v", err)
	}

	jwt.Init(cfg.JWTSecret)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Sunucu :%s portunda baslatiliyor...", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Sunucu baslatılamadi: %v", err)
	}
}
