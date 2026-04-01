package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Posinowa/FinbudApp/internal/auth"
	"github.com/Posinowa/FinbudApp/internal/budget"
	"github.com/Posinowa/FinbudApp/internal/category"
	"github.com/Posinowa/FinbudApp/internal/transaction"
	"github.com/Posinowa/FinbudApp/internal/user"
	"github.com/Posinowa/FinbudApp/pkg/config"
	"github.com/Posinowa/FinbudApp/pkg/database"
	jwtpkg "github.com/Posinowa/FinbudApp/pkg/jwt"
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

	jwtpkg.Init(cfg.JWTSecret)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(r)

	// User routes
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	userHandler.RegisterRoutes(r)

	// Category routes
	categoryRepo := category.NewRepository(db)
	categoryService := category.NewService(categoryRepo)
	categoryHandler := category.NewHandler(categoryService)
	categoryHandler.RegisterRoutes(r)

	// Transaction routes
	transaction.RegisterRoutes(r.Group("/api/v1"), db, categoryRepo)

	// Budget routes
	budget.RegisterRoutes(r.Group("/api/v1"), db, categoryRepo)

	log.Printf("Sunucu :%s portunda baslatiliyor...", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Sunucu baslatilamadi: %v", err)
	}
}