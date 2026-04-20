package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = time.Hour
	RefreshTokenDuration = time.Hour * 24 * 7
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var secretKey []byte

func Init(secret string) {
	secretKey = []byte(secret)
}

func GenerateAccessToken(userID string) (string, error) {
	return generate(userID, AccessTokenDuration)
}

func GenerateRefreshToken(userID string) (string, error) {
	return generate(userID, RefreshTokenDuration)
}

func generate(userID string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("beklenmeyen imzalama metodu")
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("gecersiz token")
	}
	return claims, nil
}