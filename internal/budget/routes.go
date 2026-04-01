package budget

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/Posinowa/FinbudApp/internal/category"
)

// RegisterRoutes registers budget routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB, categoryRepo *category.Repository) {
	repo := NewRepository(db)
	service := NewService(repo, categoryRepo)
	handler := NewHandler(service)

	budgets := router.Group("/budgets")
	{
		budgets.GET("", handler.GetAll)
		budgets.POST("", handler.Create)
		budgets.PUT("/:id", handler.Update)
	}
}