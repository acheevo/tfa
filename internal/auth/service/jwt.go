package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/shared/config"
)

// JWTService handles JWT token operations
type JWTService struct {
	config *config.Config
}

// NewJWTService creates a new JWT service
func NewJWTService(config *config.Config) *JWTService {
	return &JWTService{
		config: config,
	}
}

// GenerateAccessToken generates a new access token for the user
func (j *JWTService) GenerateAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.JWTAccessTokenDurationParsed())

	claims := &domain.JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role, // Include role in JWT claims for stateless authorization
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.JWTSecret))
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTService) GenerateRefreshToken() (string, error) {
	// Generate a random UUID for the refresh token
	tokenUUID := uuid.New()
	return tokenUUID.String(), nil
}

// ValidateAccessToken validates an access token and returns the claims
func (j *JWTService) ValidateAccessToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Make sure token method conforms to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.JWTSecret), nil
	})
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	// Check if token type is access
	if claims.TokenType != "access" {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

// GenerateRandomToken generates a random token for email verification and password reset
func (j *JWTService) GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetAccessTokenDuration returns the access token duration
func (j *JWTService) GetAccessTokenDuration() time.Duration {
	return j.config.JWTAccessTokenDurationParsed()
}

// GetRefreshTokenDuration returns the refresh token duration
func (j *JWTService) GetRefreshTokenDuration() time.Duration {
	return j.config.JWTRefreshTokenDurationParsed()
}
