package dashboard

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Posinowa/FinbudApp/internal/category"
)

type Service struct {
	db           *sqlx.DB
	categoryRepo *category.Repository
}

func NewService(db *sqlx.DB, categoryRepo *category.Repository) *Service {
	return &Service{
		db:           db,
		categoryRepo: categoryRepo,
	}
}

// GetSummary returns the dashboard summary for a given month
func (s *Service) GetSummary(ctx context.Context, userID string, monthStr string) (*DashboardSummary, error) {
	// Parse month
	year, month, err := parseMonth(monthStr)
	if err != nil {
		return nil, ErrInvalidMonth
	}

	// Get total income
	totalIncome, err := s.getTotalByType(ctx, userID, year, month, "income")
	if err != nil {
		return nil, err
	}

	// Get total expense
	totalExpense, err := s.getTotalByType(ctx, userID, year, month, "expense")
	if err != nil {
		return nil, err
	}

	// Get budget summary
	budgetSummary, err := s.getBudgetSummary(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Get recent transactions
	recentTransactions, err := s.getRecentTransactions(ctx, userID, 10)
	if err != nil {
		return nil, err
	}

	return &DashboardSummary{
		Month:              fmt.Sprintf("%d-%02d", year, month),
		TotalIncome:        totalIncome,
		TotalExpense:       totalExpense,
		Balance:            totalIncome - totalExpense,
		BudgetSummary:      budgetSummary,
		RecentTransactions: recentTransactions,
	}, nil
}

// getTotalByType gets sum of transactions by type for a month
func (s *Service) getTotalByType(ctx context.Context, userID string, year int, month int, txType string) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1::uuid
		AND EXTRACT(YEAR FROM date) = $2
		AND EXTRACT(MONTH FROM date) = $3
		AND type = $4
	`
	err := s.db.GetContext(ctx, &total, query, userID, year, month, txType)
	return total, err
}

// getBudgetSummary gets all budgets with spent amounts for a month
func (s *Service) getBudgetSummary(ctx context.Context, userID string, year int, month int) ([]BudgetSummary, error) {
	query := `
		SELECT
			b.id, b.category_id, b.amount as budget_limit,
			COALESCE(SUM(
				CASE WHEN t.type = 'expense' AND EXTRACT(YEAR FROM t.date) = $2 AND EXTRACT(MONTH FROM t.date) = $3
				THEN t.amount ELSE 0 END
			), 0) as spent
		FROM budgets b
		LEFT JOIN transactions t ON t.category_id = b.category_id AND t.user_id = b.user_id
		WHERE b.user_id = $1::uuid AND b.year = $2 AND b.month = $3
		GROUP BY b.id, b.category_id, b.amount
		ORDER BY b.created_at DESC
	`

	rows, err := s.db.QueryxContext(ctx, query, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []BudgetSummary
	for rows.Next() {
		var id, categoryID string
		var limit, spent float64

		err := rows.Scan(&id, &categoryID, &limit, &spent)
		if err != nil {
			return nil, err
		}

		remaining := limit - spent
		percentUsed := 0.0
		if limit > 0 {
			percentUsed = (spent / limit) * 100
		}

		budget := BudgetSummary{
			ID:          id,
			Limit:       limit,
			Spent:       spent,
			Remaining:   remaining,
			PercentUsed: percentUsed,
		}

		// Get category
		cat, err := s.categoryRepo.GetByID(ctx, categoryID)
		if err == nil {
			budget.Category = CategoryInfo{
				ID:   cat.ID,
				Name: cat.Name,
				Icon: cat.Icon,
				Type: cat.Type,
			}
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// getRecentTransactions gets the most recent transactions
func (s *Service) getRecentTransactions(ctx context.Context, userID string, limit int) ([]RecentTransaction, error) {
	query := `
		SELECT id, amount, type, category_id, description, date
		FROM transactions
		WHERE user_id = $1::uuid
		ORDER BY date DESC, created_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryxContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []RecentTransaction
	for rows.Next() {
		var id, txType, categoryID string
		var amount float64
		var description *string
		var date time.Time

		err := rows.Scan(&id, &amount, &txType, &categoryID, &description, &date)
		if err != nil {
			return nil, err
		}

		tx := RecentTransaction{
			ID:          id,
			Amount:      amount,
			Type:        txType,
			Description: description,
			Date:        date,
		}

		// Get category
		cat, err := s.categoryRepo.GetByID(ctx, categoryID)
		if err == nil {
			tx.Category = CategoryInfo{
				ID:   cat.ID,
				Name: cat.Name,
				Icon: cat.Icon,
				Type: cat.Type,
			}
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// parseMonth parses YYYY-MM format or returns current month if empty
func parseMonth(monthStr string) (int, int, error) {
	if monthStr == "" {
		now := time.Now()
		return now.Year(), int(now.Month()), nil
	}

	parts := strings.Split(monthStr, "-")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidMonth
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 2000 || year > 2100 {
		return 0, 0, ErrInvalidMonth
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return 0, 0, ErrInvalidMonth
	}

	return year, month, nil
}