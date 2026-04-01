package budget

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Posinowa/FinbudApp/internal/apperror"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAll godoc
// @Summary      List all budgets
// @Description  Retrieves budgets for a specific month with calculated spent values
// @Tags         budgets
// @Produce      json
// @Param        month query string false "Month filter (YYYY-MM format, defaults to current month)"
// @Success      200 {object} BudgetListResponse
// @Failure      400 {object} apperror.ErrorResponse "Invalid month format"
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /budgets [get]
func (h *Handler) GetAll(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.NewErrorResponse("unauthorized", "User not authenticated"))
		return
	}

	var userID string
	switch v := userIDValue.(type) {
	case string:
		userID = v
	case uuid.UUID:
		userID = v.String()
	default:
		c.JSON(http.StatusUnauthorized, apperror.NewErrorResponse("unauthorized", "Invalid user ID"))
		return
	}

	// Get month parameter
	month := c.Query("month")

	// Call service
	result, err := h.service.GetAll(c.Request.Context(), userID, month)
	if err != nil {
		switch err {
		case ErrInvalidMonth:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Invalid month format. Use YYYY-MM"))
		default:
			c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to get budgets"))
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
