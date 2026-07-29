package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
)

type authClaimsContextKey struct{}

type AuthMiddleware struct {
	validator domain.TokenValidator
}

func NewAuthMiddleware(validator domain.TokenValidator) *AuthMiddleware {
	return &AuthMiddleware{validator: validator}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := m.validator.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), authClaimsContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthClaimsFromContext(ctx context.Context) (*domain.AuthClaims, bool) {
	claims, ok := ctx.Value(authClaimsContextKey{}).(*domain.AuthClaims)
	return claims, ok
}
