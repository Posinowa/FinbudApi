package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, name, email, passwordHash string) (string, error) {
	var userID string
	query := `
		INSERT INTO users (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query, name, email, passwordHash).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `SELECT id, full_name, email, password_hash FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	return err
}