package budget

import (
	"errors"
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
	Category *category.Category
	Spent    float64
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
