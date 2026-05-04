package user

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `db:"id" json:"id"`
	FullName string `db:"full_name" json:"full_name"`
	Email    string `db:"email" json:"email"`
}

type UserWithPassword struct {
	ID           string `db:"id"`
	FullName     string `db:"full_name"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type UpdateRequest struct {
	Name string `json:"name"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

var ErrUserNotFound = errors.New("kullanici bulunamadi")
var ErrInvalidPassword = errors.New("eski sifre hatali")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMe(ctx context.Context, userID string) (*User, int, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if user == nil {
		return nil, http.StatusNotFound, ErrUserNotFound
	}
	return user, http.StatusOK, nil
}

func (s *Service) UpdateMe(ctx context.Context, userID string, req UpdateRequest) (*User, int, error) {
	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["full_name"] = req.Name
	}

	user, err := s.repo.Update(ctx, userID, fields)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return user, http.StatusOK, nil
}

func (s *Service) UpdatePassword(ctx context.Context, userID string, req UpdatePasswordRequest) (int, error) {
	user, err := s.repo.GetByIDWithPassword(ctx, userID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if user == nil {
		return http.StatusNotFound, ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return http.StatusUnauthorized, ErrInvalidPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = s.repo.UpdatePassword(ctx, userID, string(newHash))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Şifre değişince diğer cihazlardaki tüm oturumları geçersiz kıl
	_ = s.repo.DeleteAllRefreshTokens(ctx, userID)

	return http.StatusOK, nil
}

func (s *Service) DeleteMe(ctx context.Context, userID string) (int, error) {
	err := s.repo.Delete(ctx, userID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}