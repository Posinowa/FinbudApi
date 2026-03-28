package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(ctx context.Context, userID string) (*User, error) {
	var user User
	query := `SELECT id, full_name, email FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByIDWithPassword(ctx context.Context, userID string) (*UserWithPassword, error) {
	var user UserWithPassword
	query := `SELECT id, full_name, email, password_hash FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Update(ctx context.Context, userID string, fields map[string]interface{}) (*User, error) {
	if len(fields) == 0 {
		return r.GetByID(ctx, userID)
	}

	query := "UPDATE users SET "
	args := []interface{}{}
	i := 1
	for col, val := range fields {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", col, i)
		args = append(args, val)
		i++
	}
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, full_name, email", i)
	args = append(args, userID)

	var user User
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}

func (r *Repository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}