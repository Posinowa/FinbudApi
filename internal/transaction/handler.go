package transaction

import (
	"net/http"
        "strconv"

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
// GetByID godoc
// @Summary      Get a transaction by ID
// @Description  Retrieves a single transaction with category details
// @Tags         transactions
// @Produce      json
// @Param        id path string true "Transaction ID"
// @Success      200 {object} TransactionResponse
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure      403 {object} apperror.ErrorResponse "Forbidden"
// @Failure      404 {object} apperror.ErrorResponse "Not found"
// @Security     BearerAuth
// @Router       /transactions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
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

	// Get transaction ID from path
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Transaction ID is required"))
		return
	}

	// Call service
	transaction, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, apperror.NewErrorResponse("not_found", "Transaction not found"))
		case ErrUnauthorized:
			c.JSON(http.StatusForbidden, apperror.NewErrorResponse("forbidden", "You don't have access to this transaction"))
		default:
			c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to get transaction"))
		}
		return
	}

	c.JSON(http.StatusOK, ToTransactionResponse(transaction))
}

// GetAll godoc
// @Summary      List all transactions
// @Description  Retrieves transactions with filtering and pagination
// @Tags         transactions
// @Produce      json
// @Param        type query string false "Filter by type (income/expense)"
// @Param        category_id query string false "Filter by category ID"
// @Param        month query string false "Filter by month (YYYY-MM)"
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Items per page (default: 20, max: 100)"
// @Success      200 {object} TransactionListResponse
// @Failure      400 {object} apperror.ErrorResponse "Invalid query parameter"
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /transactions [get]
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

	// Parse query parameters
	filter := TransactionFilter{
		Page:  1,
		Limit: 20,
	}

	// Type filter
	if typeParam := c.Query("type"); typeParam != "" {
		if typeParam != "income" && typeParam != "expense" {
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Type must be 'income' or 'expense'"))
			return
		}
		filter.Type = &typeParam
	}

	// Category filter
	if categoryID := c.Query("category_id"); categoryID != "" {
		filter.CategoryID = &categoryID
	}

	// Month filter
	if month := c.Query("month"); month != "" {
		filter.Month = &month
	}

	// Page
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Invalid page number"))
			return
		}
		filter.Page = page
	}

	// Limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Invalid limit"))
			return
		}
		filter.Limit = limit
	}

	// Call service
	result, err := h.service.GetAll(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to get transactions"))
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update godoc
// @Summary      Update a transaction
// @Description  Updates an existing transaction (type cannot be changed)
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        id path string true "Transaction ID"
// @Param        request body UpdateTransactionRequest true "Fields to update"
// @Success      200 {object} TransactionResponse
// @Failure      400 {object} apperror.ErrorResponse "Validation error"
// @Failure      401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure      403 {object} apperror.ErrorResponse "Forbidden"
// @Failure      404 {object} apperror.ErrorResponse "Not found"
// @Security     BearerAuth
// @Router       /transactions/{id} [put]
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

	// Get transaction ID from path
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Transaction ID is required"))
		return
	}

	// Bind request body
	var req UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewValidationErrorResponse(err))
		return
	}

	// Validate amount if provided
	if req.Amount != nil && *req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Amount must be positive"))
		return
	}

	// Call service
	transaction, err := h.service.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, apperror.NewErrorResponse("not_found", "Transaction not found"))
		case ErrUnauthorized:
			c.JSON(http.StatusForbidden, apperror.NewErrorResponse("forbidden", "You don't have access to this transaction"))
		case ErrCategoryNotFound:
			c.JSON(http.StatusNotFound, apperror.NewErrorResponse("not_found", "Category not found"))
		case ErrInvalidDate:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Invalid date format. Use YYYY-MM-DD"))
		case ErrInvalidAmount:
			c.JSON(http.StatusBadRequest, apperror.NewErrorResponse("validation_error", "Amount must be positive"))
		default:
			c.JSON(http.StatusInternalServerError, apperror.NewErrorResponse("internal_error", "Failed to update transaction"))
		}
		return
	}

	c.JSON(http.StatusOK, ToTransactionResponse(transaction))
}
