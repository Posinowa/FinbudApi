package category

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Posinowa/FinbudApp/pkg/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	categories := r.Group("/categories")
	categories.Use(middleware.AuthMiddleware())
	{
		categories.GET("", h.GetAll)
                categories.GET("/:id", h.GetByID)
		categories.POST("", h.Create)
		categories.PUT("/:id", h.Update)
		categories.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) GetAll(c *gin.Context) {
	userID := c.GetString("user_id")

	// Query parametresi: ?type=income veya ?type=expense
	var categoryType *string
	if t := c.Query("type"); t != "" {
		categoryType = &t
	}

	categories, statusCode, err := h.service.GetAll(c.Request.Context(), userID, categoryType)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, gin.H{"data": categories})
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, statusCode, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, category)
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	categoryID := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, statusCode, err := h.service.Update(c.Request.Context(), userID, categoryID, req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, category)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	categoryID := c.Param("id")

	statusCode, err := h.service.Delete(c.Request.Context(), userID, categoryID)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(statusCode)
}
func (h *Handler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	categoryID := c.Param("id")

	category, statusCode, err := h.service.GetByID(c.Request.Context(), userID, categoryID)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, category)
}