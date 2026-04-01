package category

import "time"

type Category struct {
	ID        string    `json:"id" db:"id"`
	UserID    *string   `json:"user_id,omitempty" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Icon      *string   `json:"icon,omitempty" db:"icon"`
	Type      string    `json:"type" db:"type"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateCategoryRequest struct {
	Name string  `json:"name" binding:"required,min=1"`
	Icon *string `json:"icon"`
	Type string  `json:"type" binding:"required,oneof=income expense"`
}

type UpdateCategoryRequest struct {
	Name *string `json:"name"`
	Icon *string `json:"icon"`
}