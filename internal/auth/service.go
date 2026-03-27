package auth

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `db:"id"`
	FullName     string `db:"full_name"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

var ErrEmailAlreadyExists = errors.New("email already exists")

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, int, error) {
	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if existing != nil {
		return nil, http.StatusConflict, ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	userID, err := s.repo.CreateUser(ctx, req.Name, req.Email, string(hashedPassword))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &RegisterResponse{
		Message: "Kullanici basariyla olusturuldu",
		UserID:  userID,
	}, http.StatusCreated, nil
}