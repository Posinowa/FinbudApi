package budget

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetAllByMonth retrieves all budgets for a user in a specific month with spent amounts
func (r *Repository) GetAllByMonth(ctx context.Context, userID string, year int, month int) ([]BudgetWithSpent, error) {
	query := `
		SELECT
			b.id, b.user_id, b.category_id, b.amount, b.month, b.year, b.created_at, b.updated_at,
			COALESCE(SUM(
				CASE WHEN t.type = 'expense' AND EXTRACT(YEAR FROM t.date) = $2 AND EXTRACT(MONTH FROM t.date) = $3
				THEN t.amount ELSE 0 END
			), 0) as spent
		FROM budgets b
		LEFT JOIN transactions t ON t.category_id = b.category_id AND t.user_id = b.user_id
		WHERE b.user_id = $1::uuid AND b.year = $2 AND b.month = $3
		GROUP BY b.id, b.user_id, b.category_id, b.amount, b.month, b.year, b.created_at, b.updated_at
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.QueryxContext(ctx, query, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []BudgetWithSpent
	for rows.Next() {
		var b BudgetWithSpent
		err := rows.Scan(
			&b.ID, &b.UserID, &b.CategoryID, &b.Amount, &b.Month, &b.Year, &b.CreatedAt, &b.UpdatedAt,
			&b.Spent,
		)
		if err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}

	return budgets, nil
}

// GetByID retrieves a budget by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Budget, error) {
	var b Budget
	query := `
		SELECT id, user_id, category_id, amount, month, year, created_at, updated_at
		FROM budgets WHERE id = $1::uuid
	`
	err := r.db.GetContext(ctx, &b, query, id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetSpentAmount calculates the spent amount for a budget
func (r *Repository) GetSpentAmount(ctx context.Context, userID string, categoryID string, year int, month int) (float64, error) {
	var spent float64
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1::uuid
		AND category_id = $2::uuid
		AND EXTRACT(YEAR FROM date) = $3
		AND EXTRACT(MONTH FROM date) = $4
		AND type = 'expense'
	`
	err := r.db.GetContext(ctx, &spent, query, userID, categoryID, year, month)
	if err != nil {
		return 0, err
	}
	return spent, nil
}

// Create inserts a new budget
func (r *Repository) Create(ctx context.Context, b *Budget) error {
	query := `
		INSERT INTO budgets (id, user_id, category_id, amount, month, year, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.UserID, b.CategoryID, b.Amount, b.Month, b.Year, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

// Exists checks if a budget already exists for a category and month
func (r *Repository) Exists(ctx context.Context, userID string, categoryID string, year int, month int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM budgets
		WHERE user_id = $1::uuid AND category_id = $2::uuid AND year = $3 AND month = $4
	`
	err := r.db.GetContext(ctx, &count, query, userID, categoryID, year, month)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update updates a budget
func (r *Repository) Update(ctx context.Context, b *Budget) error {
	query := `
		UPDATE budgets SET amount = $1, updated_at = $2 WHERE id = $3::uuid
	`
	_, err := r.db.ExecContext(ctx, query, b.Amount, b.UpdatedAt, b.ID)
	return err
}

// Delete deletes a budget
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM budgets WHERE id = $1::uuid`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Dummy usage of fmt to avoid unused import error
var _ = fmt.Sprint