package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	golangJWT "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	repo        *Repository
	emailSender emailSender
}

type emailSender interface {
	Send(to, subject, htmlBody string) error
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithEmail(repo *Repository, sender emailSender) *Service {
	return &Service{repo: repo, emailSender: sender}
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
var ErrAppleTokenInvalid = errors.New("geçersiz Apple token")
var ErrInvalidResetToken = errors.New("geçersiz veya süresi dolmuş sıfırlama bağlantısı")

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
// ForgotPasswordRequest represents the request body for forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the request body for resetting password
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ForgotPassword sends a password reset email
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (int, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Güvenlik: kullanıcı yoksa da başarılı dön (email enumeration önleme)
	if user == nil {
		return http.StatusOK, nil
	}

	// Token oluştur
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := s.repo.CreatePasswordResetToken(ctx, user.ID, token, expiresAt); err != nil {
		return http.StatusInternalServerError, err
	}

	// Email gönder
	if s.emailSender != nil {
		resetLink := fmt.Sprintf("https://finbud-api-197224562444.europe-west1.run.app/reset-password?token=%s", token)
		subject := "Finbud - Şifre Sıfırlama"
		body := buildResetEmailHTML(user.FullName, resetLink)
		if emailErr := s.emailSender.Send(req.Email, subject, body); emailErr != nil {
			fmt.Printf("[EMAIL ERROR] Sifre sifirlama maili gonderilemedi (%s): %v\n", req.Email, emailErr)
		} else {
			fmt.Printf("[EMAIL OK] Sifre sifirlama maili gonderildi: %s\n", req.Email)
		}
	}

	return http.StatusOK, nil
}

// ResetPassword resets a user's password using a valid token
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) (int, error) {
	if err := validator.ValidatePasswordStrength(req.NewPassword); err != nil {
		return http.StatusBadRequest, err
	}

	rt, err := s.repo.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if rt == nil || rt.UsedAt != nil || rt.ExpiresAt.Before(time.Now()) {
		return http.StatusBadRequest, ErrInvalidResetToken
	}

	// Yeni şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Şifreyi güncelle
	if err := s.repo.UpdatePassword(ctx, rt.UserID, string(hashedPassword)); err != nil {
		return http.StatusInternalServerError, err
	}

	// Token'ı kullanıldı olarak işaretle
	_ = s.repo.MarkResetTokenUsed(ctx, req.Token)

	// Tüm aktif oturumları geçersiz kıl
	_ = s.repo.DeleteAllRefreshTokens(ctx, rt.UserID)

	return http.StatusOK, nil
}

// buildResetEmailHTML generates the HTML email body
func buildResetEmailHTML(name, resetLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f5f5f5; margin: 0; padding: 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #fff; border-radius: 12px; padding: 32px;">
    <h2 style="color: #2D3748; margin-bottom: 8px;">Şifre Sıfırlama</h2>
    <p style="color: #4A5568;">Merhaba %s,</p>
    <p style="color: #4A5568;">Finbud hesabınız için şifre sıfırlama talebinde bulundunuz.</p>
    <p style="color: #4A5568;">Aşağıdaki butona telefonu ile tıklayarak yeni şifrenizi oluşturabilirsiniz. Bu bağlantı <strong>1 saat</strong> geçerlidir.</p>
    <div style="text-align: center; margin: 32px 0;">
      <a href="%s"
         style="background: #4F5D75; color: #fff; padding: 14px 32px; border-radius: 8px; text-decoration: none; font-weight: bold; font-size: 16px;">
        Şifremi Sıfırla
      </a>
    </div>
    <p style="color: #718096; font-size: 13px;">Bu talebi siz yapmadıysanız bu e-postayı görmezden gelebilirsiniz.</p>
    <hr style="border: none; border-top: 1px solid #E2E8F0; margin: 24px 0;">
    <p style="color: #A0AEC0; font-size: 12px; text-align: center;">Finbud &mdash; Kişisel Finans Yönetimi</p>
  </div>
</body>
</html>`, name, resetLink)
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

// ── Apple Sign-In ──────────────────────────────────────────────────────────────

// AppleLoginRequest represents the request body for Apple login
type AppleLoginRequest struct {
	IdentityToken string `json:"identity_token" binding:"required"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
}

// appleTokenInfo holds verified claims from Apple identity token
type appleTokenInfo struct {
	Sub           string // Apple user ID (stable, always present)
	Email         string // Only on first sign-in
	EmailVerified bool
}

// appleJWKS represents Apple's public key set
type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// appleClaims holds the JWT claims from Apple's identity token
type appleClaims struct {
	golangJWT.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"` // Apple sends bool or string
}

// fetchApplePublicKeys fetches Apple's JWKS from their endpoint
func fetchApplePublicKeys() (*appleJWKS, error) {
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks appleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}
	return &jwks, nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(key appleJWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("n decode hatası: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("e decode hatası: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

// getJWTHeaderKid extracts the kid from JWT header without full verification
func getJWTHeaderKid(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", errors.New("geçersiz JWT formatı")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("header decode hatası: %w", err)
	}

	var header struct {
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("header parse hatası: %w", err)
	}

	return header.Kid, nil
}

// verifyAppleToken verifies an Apple identity token and returns its claims
func verifyAppleToken(identityToken string) (*appleTokenInfo, error) {
	// 1. JWT header'dan kid'i al
	kid, err := getJWTHeaderKid(identityToken)
	if err != nil {
		return nil, ErrAppleTokenInvalid
	}

	// 2. Apple'ın public key'lerini çek
	jwks, err := fetchApplePublicKeys()
	if err != nil {
		return nil, fmt.Errorf("Apple public keys alınamadı: %w", err)
	}

	// 3. Eşleşen key'i bul
	var matchedKey *appleJWK
	for _, k := range jwks.Keys {
		if k.Kid == kid {
			kCopy := k
			matchedKey = &kCopy
			break
		}
	}
	if matchedKey == nil {
		return nil, ErrAppleTokenInvalid
	}

	// 4. RSA public key'e çevir
	rsaKey, err := jwkToRSAPublicKey(*matchedKey)
	if err != nil {
		return nil, ErrAppleTokenInvalid
	}

	// 5. JWT'yi doğrula ve claim'leri çıkar
	token, err := golangJWT.ParseWithClaims(
		identityToken,
		&appleClaims{},
		func(t *golangJWT.Token) (interface{}, error) {
			if _, ok := t.Method.(*golangJWT.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("beklenmeyen signing method: %v", t.Header["alg"])
			}
			return rsaKey, nil
		},
		golangJWT.WithIssuer("https://appleid.apple.com"),
		golangJWT.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, ErrAppleTokenInvalid
	}

	claims, ok := token.Claims.(*appleClaims)
	if !ok || claims.Subject == "" {
		return nil, ErrAppleTokenInvalid
	}

	// email_verified bool veya string olarak gelebilir
	emailVerified := false
	switch v := claims.EmailVerified.(type) {
	case bool:
		emailVerified = v
	case string:
		emailVerified = v == "true"
	}

	return &appleTokenInfo{
		Sub:           claims.Subject,
		Email:         claims.Email,
		EmailVerified: emailVerified,
	}, nil
}

// AppleLogin verifies an Apple identity token and returns JWT tokens
func (s *Service) AppleLogin(ctx context.Context, req AppleLoginRequest) (*AuthResponse, int, error) {
	// 1. Apple identity token'ı doğrula
	info, err := verifyAppleToken(req.IdentityToken)
	if err != nil {
		return nil, http.StatusUnauthorized, ErrAppleTokenInvalid
	}

	// 2. Apple ID'ye göre kullanıcı ara
	appleUser, err := s.repo.GetUserByAppleID(ctx, info.Sub)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var userID string

	if appleUser != nil {
		// Kullanıcı zaten var
		userID = appleUser.ID
	} else {
		// Email ile mevcut kullanıcı var mı kontrol et (aynı email ile kayıtlı olabilir)
		email := info.Email
		if email == "" {
			email = req.Email // Flutter'dan gelen email (sadece ilk girişte Apple gönderir)
		}

		if email != "" {
			existingUser, err := s.repo.GetUserByEmail(ctx, email)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			if existingUser != nil {
				// Mevcut hesaba Apple ID bağla
				if linkErr := s.repo.LinkAppleID(ctx, existingUser.ID, info.Sub); linkErr != nil {
					return nil, http.StatusInternalServerError, linkErr
				}
				userID = existingUser.ID
			}
		}

		// Kullanıcı hiç yoksa yeni oluştur
		if userID == "" {
			name := req.FullName
			if name == "" {
				name = email
			}
			if name == "" {
				name = "Apple Kullanıcısı"
			}
			if email == "" {
				// Apple private relay veya email gönderilmedi — bu durumda giriş yapılamaz
				return nil, http.StatusUnauthorized, errors.New("email bilgisi alınamadı")
			}

			newID, err := s.repo.CreateAppleUser(ctx, name, email, info.Sub)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			userID = newID
		}
	}

	// 3. JWT token'ları oluştur
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

// ── Logout ─────────────────────────────────────────────────────────────────────

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