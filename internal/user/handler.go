package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Posinowa/FinbudApp/internal/apperror"
	"github.com/Posinowa/FinbudApp/pkg/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware())
	users.Use(middleware.APIUserRateLimiter.UserMiddleware())
	{
		users.GET("/me", h.GetMe)
		users.PUT("/me", h.UpdateMe)
		users.PUT("/me/password", middleware.PasswordChangeRateLimiter.UserMiddleware(), h.UpdatePassword)
		users.DELETE("/me", h.DeleteMe)
	}
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")

	user, statusCode, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, user)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	user, statusCode, err := h.service.UpdateMe(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, user)
}

func (h *Handler) UpdatePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	statusCode, err := h.service.UpdatePassword(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, gin.H{"message": "Sifre basariyla guncellendi"})
}

func (h *Handler) DeleteMe(c *gin.Context) {
	userID := c.GetString("user_id")

	statusCode, err := h.service.DeleteMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(statusCode, apperror.NewErrorResponse("error", err.Error()))
		return
	}

	c.JSON(statusCode, gin.H{"message": "Hesap basariyla silindi"})
}