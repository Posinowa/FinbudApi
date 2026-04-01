package dashboard

import (
	"errors"
	"time"
)

// Error definitions
var (
	ErrInvalidMonth = errors.New("invalid month format")
)

// CategoryInfo represents category in response
type CategoryInfo struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon,omitempty"`
	Type string  `json:"type"`
}

// BudgetSummary represents a budget in dashboard
type BudgetSummary struct {
	ID          string       `json:"id"`
	Category    CategoryInfo `json:"category"`
	Limit       float64      `json:"limit"`
	Spent       float64      `json:"spent"`
	Remaining   float64      `json:"remaining"`
	PercentUsed float64      `json:"percent_used"`
}

// RecentTransaction represents a transaction in dashboard
type RecentTransaction struct {
	ID          string       `json:"id"`
	Amount      float64      `json:"amount"`
	Type        string       `json:"type"`
	Category    CategoryInfo `json:"category"`
	Description *string      `json:"description,omitempty"`
	Date        time.Time    `json:"date"`
}

// DashboardSummary represents the dashboard response
type DashboardSummary struct {
	Month              string              `json:"month"`
	TotalIncome        float64             `json:"total_income"`
	TotalExpense       float64             `json:"total_expense"`
	Balance            float64             `json:"balance"`
	BudgetSummary      []BudgetSummary     `json:"budget_summary"`
	RecentTransactions []RecentTransaction `json:"recent_transactions"`
}