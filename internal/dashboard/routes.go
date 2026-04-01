package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/Posinowa/FinbudApp/internal/category"
	"github.com/Posinowa/FinbudApp/pkg/middleware"
)

// RegisterRoutes registers dashboard routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB, categoryRepo *category.Repository) {
	service := NewService(db, categoryRepo)
	handler := NewHandler(service)

	dashboardGroup := router.Group("/dashboard")
	dashboardGroup.Use(middleware.AuthMiddleware())
	{
		dashboardGroup.GET("/summary", handler.GetSummary)
	}
}