package transaction

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/Posinowa/FinbudApp/internal/category"
)

// RegisterRoutes registers transaction routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB, categoryRepo *category.Repository) {
	repo := NewRepository(db)
	service := NewService(repo, categoryRepo)
	handler := NewHandler(service)

	transactions := router.Group("/transactions")
	{
		transactions.POST("", handler.Create)
		transactions.GET("", handler.GetAll)
		transactions.GET("/:id", handler.GetByID)
		transactions.PUT("/:id", handler.Update)
	}
}