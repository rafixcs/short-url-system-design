package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

const tokenIssuer = "shorter-url-api"

type AuthService struct {
	repo       domain.AuthRepository
	secret     []byte
	expiration time.Duration
}

type tokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(repo domain.AuthRepository, secret string, expiration time.Duration) *AuthService {
	return &AuthService{
		repo:       repo,
		secret:     []byte(secret),
		expiration: expiration,
	}
}

func (s *AuthService) Auth(ctx context.Context, user, email, password string) (*domain.JWTToken, error) {
	account, err := s.repo.GetUserByEmailOrName(ctx, email, user)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now()
	expiresAt := now.Add(s.expiration)
	claims := tokenClaims{
		UserID: account.ID.Hex(),
		Email:  account.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID.Hex(),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("sign JWT: %w", err)
	}

	return &domain.JWTToken{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *AuthService) ValidateToken(rawToken string) (*domain.AuthClaims, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return s.secret, nil
		},
		jwt.WithIssuer(tokenIssuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return &domain.AuthClaims{UserID: claims.UserID, Email: claims.Email}, nil
}

var _ domain.AuthService = (*AuthService)(nil)
var _ domain.TokenValidator = (*AuthService)(nil)
