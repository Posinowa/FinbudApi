package budget

import (
	"errors"
        "fmt"
	"time"

	"github.com/Posinowa/FinbudApp/internal/category"
)

// Error definitions
var (
	ErrNotFound      = errors.New("budget not found")
	ErrUnauthorized  = errors.New("unauthorized access")
	ErrInvalidMonth  = errors.New("invalid month format")
	ErrAlreadyExists = errors.New("budget already exists for this category and month")
)

// Budget represents a budget for a category
type Budget struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	CategoryID string    `db:"category_id" json:"category_id"`
	Amount     float64   `db:"amount" json:"amount"`
	Month      int       `db:"month" json:"month"`
	Year       int       `db:"year" json:"year"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// BudgetWithCategory includes category details
type BudgetWithCategory struct {
	Budget
	Category *category.Category `json:"category,omitempty"`
}

// BudgetResponse represents the API response for a budget
type BudgetResponse struct {
	ID          string           `json:"id"`
	Category    CategoryResponse `json:"category"`
	Limit       float64          `json:"limit"`
	Spent       float64          `json:"spent"`
	Remaining   float64          `json:"remaining"`
	PercentUsed float64          `json:"percent_used"`
}

// CategoryResponse represents nested category in budget response
type CategoryResponse struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon,omitempty"`
	Type string  `json:"type"`
}

// BudgetListResponse represents the list response
type BudgetListResponse struct {
	Month string           `json:"month"`
	Data  []BudgetResponse `json:"data"`
}

// BudgetWithSpent includes spent calculation
type BudgetWithSpent struct {
	Budget
	Category *category.Category `db:"-"`
	Spent    float64            `db:"spent"`
}

// ToBudgetResponse converts BudgetWithSpent to BudgetResponse
func ToBudgetResponse(b *BudgetWithSpent) BudgetResponse {
	remaining := b.Amount - b.Spent
	percentUsed := 0.0
	if b.Amount > 0 {
		percentUsed = (b.Spent / b.Amount) * 100
	}

	response := BudgetResponse{
		ID:          b.ID,
		Limit:       b.Amount,
		Spent:       b.Spent,
		Remaining:   remaining,
		PercentUsed: percentUsed,
	}

	if b.Category != nil {
		response.Category = CategoryResponse{
			ID:   b.Category.ID,
			Name: b.Category.Name,
			Icon: b.Category.Icon,
			Type: b.Category.Type,
		}
	}

	return response
}
// CreateBudgetRequest represents the request body for creating a budget
type CreateBudgetRequest struct {
	CategoryID string  `json:"category_id" binding:"required,uuid"`
	Limit      float64 `json:"limit" binding:"required,gt=0"`
	Month      string  `json:"month" binding:"required"`
}

// CreateBudgetResponse represents the response for created budget
type CreateBudgetResponse struct {
	ID        string           `json:"id"`
	Category  CategoryResponse `json:"category"`
	Limit     float64          `json:"limit"`
	Month     string           `json:"month"`
	CreatedAt time.Time        `json:"created_at"`
}

// ToCreateBudgetResponse converts Budget to CreateBudgetResponse
func ToCreateBudgetResponse(b *Budget, cat *category.Category) CreateBudgetResponse {
	response := CreateBudgetResponse{
		ID:        b.ID,
		Limit:     b.Amount,
		Month:     fmt.Sprintf("%d-%02d", b.Year, b.Month),
		CreatedAt: b.CreatedAt,
	}

	if cat != nil {
		response.Category = CategoryResponse{
			ID:   cat.ID,
			Name: cat.Name,
			Icon: cat.Icon,
			Type: cat.Type,
		}
	}

	return response
}

// UpdateBudgetRequest represents the request body for updating a budget
type UpdateBudgetRequest struct {
	Limit float64 `json:"limit" binding:"required,gt=0"`
}
