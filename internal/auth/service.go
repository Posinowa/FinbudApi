package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwtpkg "github.com/Posinowa/FinbudApp/pkg/jwt"
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

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid or expired token")

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

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, int, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if user == nil {
		return nil, http.StatusUnauthorized, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, http.StatusUnauthorized, ErrInvalidCredentials
	}

	accessToken, err := jwtpkg.GenerateAccessToken(0)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken(0)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.repo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, http.StatusOK, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, int, error) {
	rt, err := s.repo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if rt == nil || rt.ExpiresAt.Before(time.Now()) {
		return nil, http.StatusUnauthorized, ErrInvalidToken
	}

	err = s.repo.DeleteRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	accessToken, err := jwtpkg.GenerateAccessToken(0)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	newRefreshToken, err := jwtpkg.GenerateRefreshToken(0)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.repo.SaveRefreshToken(ctx, rt.UserID, newRefreshToken, expiresAt)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, http.StatusOK, nil
}