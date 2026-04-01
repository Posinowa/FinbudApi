package transaction

import (
	"errors"
	"time"

	"github.com/Posinowa/FinbudApp/internal/category"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TypeIncome  TransactionType = "income"
	TypeExpense TransactionType = "expense"
)

// Error definitions
var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrInvalidDate      = errors.New("invalid date format")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrInvalidType      = errors.New("invalid transaction type")
	ErrNotFound         = errors.New("transaction not found")
	ErrUnauthorized     = errors.New("unauthorized access")
)

// Transaction represents a financial transaction
type Transaction struct {
	ID          string          `db:"id" json:"id"`
	UserID      string          `db:"user_id" json:"user_id"`
	CategoryID  string          `db:"category_id" json:"category_id"`
	Amount      float64         `db:"amount" json:"amount"`
	Type        TransactionType `db:"type" json:"type"`
	Date        time.Time       `db:"date" json:"date"`
	Description *string         `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// TransactionWithCategory includes category details
type TransactionWithCategory struct {
	Transaction
	Category *category.Category `json:"category"`
}

// CategoryResponse represents the nested category in response
type CategoryResponse struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon,omitempty"`
	Type string  `json:"type"`
}

// TransactionResponse represents the API response for a transaction
type TransactionResponse struct {
	ID          string           `json:"id"`
	Amount      float64          `json:"amount"`
	Type        TransactionType  `json:"type"`
	Date        string           `json:"date"`
	Description *string          `json:"description,omitempty"`
	Category    CategoryResponse `json:"category"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ToTransactionResponse converts TransactionWithCategory to TransactionResponse
func ToTransactionResponse(t *TransactionWithCategory) TransactionResponse {
	response := TransactionResponse{
		ID:          t.ID,
		Amount:      t.Amount,
		Type:        t.Type,
		Date:        t.Date.Format("2006-01-02"),
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}

	if t.Category != nil {
		response.Category = CategoryResponse{
			ID:   t.Category.ID,
			Name: t.Category.Name,
			Icon: t.Category.Icon,
			Type: t.Category.Type,
		}
	}

	return response
}

// CreateTransactionInput represents input for creating a transaction
type CreateTransactionInput struct {
	UserID      string
	Amount      float64
	Type        TransactionType
	CategoryID  string
	Date        string
	Description *string
}
// TransactionFilter represents filter options for listing transactions
type TransactionFilter struct {
	Type       *string
	CategoryID *string
	Month      *string
	Page       int
	Limit      int
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// TransactionListResponse represents paginated transaction list response
type TransactionListResponse struct {
	Data []TransactionResponse `json:"data"`
	Meta PaginationMeta        `json:"meta"`
}

// UpdateTransactionRequest represents the request body for updating a transaction
type UpdateTransactionRequest struct {
	Amount      *float64 `json:"amount,omitempty"`
	CategoryID  *string  `json:"category_id,omitempty"`
	Description *string  `json:"description,omitempty"`
	Date        *string  `json:"date,omitempty"`
}

