package transaction

import (
	"context"

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