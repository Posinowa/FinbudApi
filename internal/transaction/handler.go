package transaction

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

// CreateTransactionRequest represents the request body for creating a transaction
type CreateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	CategoryID  string  `json:"category_id" binding:"required,uuid"`
	Date        string  `json:"date" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// Create godoc
// @Summary      Create a new transaction
// @Description  Creates a new income or expense transaction
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        request body CreateTransactionRequest true "Transaction data"
// @Success      201 {object} TransactionResponse
// @Failure      400 {object} apperror.ErrorResponse "Validation error"
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure      404 {object} apperror.ErrorResponse "Category not found"
// @Security     BearerAuth
// @Router       /transactions [post]
func (h *Handler) Create(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.NewErrorResponse("unauthorized", "User not authenticated"))
		return
	}

	// Handle both string and uuid.UUID types
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

	// Bind and validate request
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	// Create input for service
	input := CreateTransactionInput{
		UserID:      userID,
		Amount:      req.Amount,
		Type:        TransactionType(req.Type),
		CategoryID:  req.CategoryID,
		Date:        req.Date,
		Description: req.Description,
	}

	// Call service
	transaction, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		switch err {
		case ErrCategoryNotFound:
			c.JSON(http.StatusNotFound, apperror.NewErrorResponse("not_found", "Category not found"))
		case ErrInvalidDate:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Invalid date format. Use YYYY-MM-DD"))
		case ErrInvalidAmount:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Amount must be positive"))
		case ErrInvalidType:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Type must be 'income' or 'expense'"))
		default:
			c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to create transaction"))
		}
		return
	}

	c.JSON(http.StatusCreated, ToTransactionResponse(transaction))
}