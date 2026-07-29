package domain

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type JWTToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type AuthClaims struct {
	UserID string
	Email  string
}

type AuthService interface {
	Auth(ctx context.Context, user, email, password string) (*JWTToken, error)
}

type TokenValidator interface {
	ValidateToken(token string) (*AuthClaims, error)
}

type AuthRepository interface {
	GetUserByEmailOrName(ctx context.Context, email, name string) (*UserModel, error)
}
