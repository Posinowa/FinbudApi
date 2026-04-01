package transaction

import (
	"context"
        "fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new transaction into the database
func (r *Repository) Create(ctx context.Context, t *Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, category_id, amount, type, date, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		t.ID,
		t.UserID,
		t.CategoryID,
		t.Amount,
		t.Type,
		t.Date,
		t.Description,
		t.CreatedAt,
		t.UpdatedAt,
	)
	return err
}

// GetByID retrieves a transaction by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	var t Transaction
	query := `
		SELECT id, user_id, category_id, amount, type, date, description, created_at, updated_at
		FROM transactions WHERE id = $1
	`
	err := r.db.GetContext(ctx, &t, query, id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
// GetByIDWithCategory retrieves a transaction with its category by ID
func (r *Repository) GetByIDWithCategory(ctx context.Context, id string) (*TransactionWithCategory, error) {
	var t Transaction
	query := `
		SELECT id, user_id, category_id, amount, type, date, description, created_at, updated_at
		FROM transactions WHERE id = $1
	`
	err := r.db.GetContext(ctx, &t, query, id)
	if err != nil {
		return nil, err
	}

	return &TransactionWithCategory{
		Transaction: t,
		Category:    nil, // Category will be fetched by service
	}, nil
}
// GetAll retrieves transactions with filters and pagination
func (r *Repository) GetAll(ctx context.Context, userID string, filter TransactionFilter) ([]Transaction, int, error) {
	// Base query
	query := `
		SELECT id, user_id, category_id, amount, type, date, description, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`

	args := []interface{}{userID}
	countArgs := []interface{}{userID}
	argIndex := 2

	// Type filter
	if filter.Type != nil && *filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filter.Type)
		countArgs = append(countArgs, *filter.Type)
		argIndex++
	}

	// Category filter
	if filter.CategoryID != nil && *filter.CategoryID != "" {
		query += fmt.Sprintf(" AND category_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, *filter.CategoryID)
		countArgs = append(countArgs, *filter.CategoryID)
		argIndex++
	}

	// Month filter (YYYY-MM format)
	if filter.Month != nil && *filter.Month != "" {
		query += fmt.Sprintf(" AND TO_CHAR(date, 'YYYY-MM') = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND TO_CHAR(date, 'YYYY-MM') = $%d", argIndex)
		args = append(args, *filter.Month)
		countArgs = append(countArgs, *filter.Month)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	// Order and pagination
	query += " ORDER BY date DESC, created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var transactions []Transaction
	err = r.db.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
// Update updates a transaction in the database
func (r *Repository) Update(ctx context.Context, t *Transaction) error {
	query := `
		UPDATE transactions 
		SET amount = $1, category_id = $2, description = $3, date = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query,
		t.Amount,
		t.CategoryID,
		t.Description,
		t.Date,
		t.UpdatedAt,
		t.ID,
	)
	return err
}
