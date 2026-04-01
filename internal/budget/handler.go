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

// Create godoc
// @Summary      Create a new budget
// @Description  Creates a new budget for a category and month
// @Tags         budgets
// @Accept       json
// @Produce      json
// @Param        request body CreateBudgetRequest true "Budget data"
// @Success      201 {object} CreateBudgetResponse
// @Failure      400 {object} apperror.ErrorResponse "Validation error"
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure      404 {object} apperror.ErrorResponse "Category not found"
// @Failure      409 {object} apperror.ErrorResponse "Conflict"
// @Security     BearerAuth
// @Router       /budgets [post]
func (h *Handler) Create(c *gin.Context) {
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

	// Bind request body
	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	// Call service
	result, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		switch err {
		case ErrInvalidMonth:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Invalid month format. Use YYYY-MM"))
		case ErrNotFound:
			c.JSON(http.StatusNotFound, apperror.NewErrorResponse("not_found", "Category not found"))
		case ErrAlreadyExists:
			c.JSON(http.StatusConflict, apperror.NewErrorResponse("conflict", "Budget already exists for this category and month"))
		default:
			c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to create budget"))
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Update godoc
// @Summary      Update a budget
// @Description  Updates a budget limit (category and month cannot be changed)
// @Tags         budgets
// @Accept       json
// @Produce      json
// @Param        id path string true "Budget ID"
// @Param        request body UpdateBudgetRequest true "Budget data"
// @Success      200 {object} CreateBudgetResponse
// @Failure      400 {object} apperror.ErrorResponse "Validation error"
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure      403 {object} apperror.ErrorResponse "Forbidden"
// @Failure      404 {object} apperror.ErrorResponse "Not found"
// @Security     BearerAuth
// @Router       /budgets/{id} [put]
func (h *Handler) Update(c *gin.Context) {
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

	// Get budget ID from path
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Budget ID is required"))
		return
	}

	// Bind request body
	var req UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	// Call service
	result, err := h.service.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, apperror.NewErrorResponse("not_found", "Budget not found"))
		case ErrUnauthorized:
			c.JSON(http.StatusForbidden, apperror.NewErrorResponse("forbidden", "You don't have access to this budget"))
		default:
			c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to update budget"))
		}
		return
	}

	c.JSON(http.StatusOK, result)
}