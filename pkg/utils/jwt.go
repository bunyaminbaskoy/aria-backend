package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token süreleri.
const (
	AccessTokenDuration  = 15 * time.Minute // Access token: 15 dakika
	RefreshTokenDuration = 7 * 24 * time.Hour // Refresh token: 7 gün
)

// Claims — JWT içindeki veriler.
type Claims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" veya "refresh"
	jwt.RegisteredClaims
}

// TokenPair — Token çifti.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateToken — Tek access token üret.
func GenerateToken(userID uint, email string) (string, error) {
	// Sadece access token döndür
	accessToken, err := generateTokenWithType(userID, email, "access", AccessTokenDuration)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

// GenerateTokenPair — Token çifti üret.
func GenerateTokenPair(userID uint, email string) (*TokenPair, error) {
	accessToken, err := generateTokenWithType(userID, email, "access", AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateTokenWithType(userID, email, "refresh", RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateAccessToken — Yeni access token üret.
func GenerateAccessToken(userID uint, email string) (string, error) {
	return generateTokenWithType(userID, email, "access", AccessTokenDuration)
}

// generateTokenWithType — JWT oluştur.
func generateTokenWithType(userID uint, email string, tokenType string, duration time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken — Token'ı doğrula.
func ValidateToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
