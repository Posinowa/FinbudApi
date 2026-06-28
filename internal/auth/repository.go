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

// AppleUser Apple Sign-In ile bulunan/oluşturulan kullanıcı
type AppleUser struct {
	ID      string  `db:"id"`
	AppleID string  `db:"apple_id"`
	Email   *string `db:"email"`
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

// CreateGoogleUser creates a user without a password (Google OAuth)
func (r *Repository) CreateGoogleUser(ctx context.Context, name, email string) (string, error) {
	var userID string
	query := `
		INSERT INTO users (full_name, email)
		VALUES ($1, $2)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query, name, email).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	return err
}

type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	var rt RefreshToken
	query := `SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = $1`
	err := r.db.GetContext(ctx, &rt, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// UpdatePassword updates a user's password hash
func (r *Repository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// DeleteAllRefreshTokens deletes all refresh tokens for a user
func (r *Repository) DeleteAllRefreshTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// CreatePasswordResetToken stores a password reset token for a user
func (r *Repository) CreatePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	// Önce bu kullanıcıya ait eski kullanılmamış tokenları sil
	_, _ = r.db.ExecContext(ctx,
		`DELETE FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NULL`, userID)

	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	return err
}

type PasswordResetToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
}

// GetPasswordResetToken retrieves a reset token by its value
func (r *Repository) GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	var t PasswordResetToken
	query := `SELECT id, user_id, token, expires_at, used_at FROM password_reset_tokens WHERE token = $1`
	err := r.db.GetContext(ctx, &t, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// MarkResetTokenUsed marks a reset token as used
func (r *Repository) MarkResetTokenUsed(ctx context.Context, token string) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// GetUserByAppleID retrieves a user by their Apple ID
func (r *Repository) GetUserByAppleID(ctx context.Context, appleID string) (*AppleUser, error) {
	var user AppleUser
	query := `SELECT id, apple_id, email FROM users WHERE apple_id = $1`
	err := r.db.GetContext(ctx, &user, query, appleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateAppleUser creates a new user from Apple Sign-In
func (r *Repository) CreateAppleUser(ctx context.Context, name, email, appleID string) (string, error) {
	var userID string
	query := `
		INSERT INTO users (full_name, email, apple_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query, name, email, appleID).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

// LinkAppleID links an Apple ID to an existing user (email match)
func (r *Repository) LinkAppleID(ctx context.Context, userID, appleID string) error {
	query := `UPDATE users SET apple_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, appleID, userID)
	return err
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	query := `SELECT id, full_name, email, password_hash FROM users WHERE id = $1::uuid`
	err := r.db.GetContext(ctx, &user, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}