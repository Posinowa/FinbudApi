package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Posinowa/FinbudApp/internal/validator"
	"github.com/Posinowa/FinbudApp/pkg/blacklist"
	jwtpkg "github.com/Posinowa/FinbudApp/pkg/jwt"
)

type User struct {
	ID           string  `db:"id"`
	FullName     string  `db:"full_name"`
	Email        string  `db:"email"`
	PasswordHash *string `db:"password_hash"`
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
	Password string `json:"password" binding:"required"`
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
var ErrGoogleTokenInvalid = errors.New("geçersiz Google token")

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, int, error) {
	if err := validator.ValidatePasswordStrength(req.Password); err != nil {
		return nil, http.StatusBadRequest, err
	}

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

	if user.PasswordHash == nil {
		return nil, http.StatusUnauthorized, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, http.StatusUnauthorized, ErrInvalidCredentials
	}

	accessToken, err := jwtpkg.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken(user.ID)
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

	accessToken, err := jwtpkg.GenerateAccessToken(rt.UserID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	newRefreshToken, err := jwtpkg.GenerateRefreshToken(rt.UserID)
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
// GoogleLoginRequest represents the request body for Google login
type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// googleTokenInfo represents the response from Google's tokeninfo endpoint
type googleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Error         string `json:"error"`
}

// GoogleLogin verifies a Google ID token and returns JWT tokens
func (s *Service) GoogleLogin(ctx context.Context, req GoogleLoginRequest) (*AuthResponse, int, error) {
	// Verify token with Google's tokeninfo endpoint
	info, err := verifyGoogleToken(req.IDToken)
	if err != nil || info.Email == "" {
		return nil, http.StatusUnauthorized, ErrGoogleTokenInvalid
	}

	// Email verified check
	if info.EmailVerified != "true" {
		return nil, http.StatusUnauthorized, ErrGoogleTokenInvalid
	}

	// Find or create user by email
	user, err := s.repo.GetUserByEmail(ctx, info.Email)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var userID string
	if user == nil {
		// Create new user (no password for Google users)
		name := info.Name
		if name == "" {
			name = info.Email
		}
		userID, err = s.repo.CreateGoogleUser(ctx, name, info.Email)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	} else {
		userID = user.ID
	}

	// Generate tokens
	accessToken, err := jwtpkg.GenerateAccessToken(userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken(userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.repo.SaveRefreshToken(ctx, userID, refreshToken, expiresAt); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, http.StatusOK, nil
}

// verifyGoogleToken calls Google's tokeninfo endpoint to validate an ID token
func verifyGoogleToken(idToken string) (*googleTokenInfo, error) {
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || info.Error != "" {
		return nil, ErrGoogleTokenInvalid
	}

	return &info, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string, accessToken string) (int, error) {
	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if rt == nil {
		return http.StatusUnauthorized, ErrInvalidToken
	}

	err = s.repo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Access token'ı blacklist'e ekle
	if accessToken != "" {
		claims, err := jwtpkg.ValidateToken(accessToken)
		if err == nil && claims.ID != "" {
			blacklist.Add(claims.ID)
		}
	}

	return http.StatusOK, nil
}