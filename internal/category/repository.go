package category

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

func (r *Repository) GetAll(ctx context.Context, userID string, categoryType *string) ([]Category, error) {
	query := `
		SELECT id, user_id, name, icon, type, is_default, created_at
		FROM categories
		WHERE (user_id = $1 OR is_default = true)
	`
	args := []interface{}{userID}

	if categoryType != nil {
		query += " AND type = $2"
		args = append(args, *categoryType)
	}

	query += " ORDER BY is_default DESC, name ASC"

	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query, args...)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Category, error) {
	var category Category
	err := r.db.GetContext(ctx, &category,
		"SELECT id, user_id, name, icon, type, is_default, created_at FROM categories WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *Repository) Create(ctx context.Context, userID, name string, icon *string, categoryType string) (*Category, error) {
	var category Category
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO categories (user_id, name, icon, type, is_default)
		VALUES ($1, $2, $3, $4, false)
		RETURNING id, user_id, name, icon, type, is_default, created_at`,
		userID, name, icon, categoryType,
	).StructScan(&category)

	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *Repository) Update(ctx context.Context, id string, name *string, icon *string, categoryType *string) (*Category, error) {
	var category Category

	query := "UPDATE categories SET "
	args := []interface{}{}
	argIndex := 1

	if name != nil {
		query += fmt.Sprintf("name = $%d, ", argIndex)
		args = append(args, *name)
		argIndex++
	}

	if icon != nil {
		query += fmt.Sprintf("icon = $%d, ", argIndex)
		args = append(args, *icon)
		argIndex++
	}

	if categoryType != nil {
		query += fmt.Sprintf("type = $%d, ", argIndex)
		args = append(args, *categoryType)
		argIndex++
	}

	if len(args) == 0 {
		return r.GetByID(ctx, id)
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d RETURNING id, user_id, name, icon, type, is_default, created_at", argIndex)
	args = append(args, id)

	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", id)
	return err
}