package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"github.com/rafixcs/shorter-url-design-system/src/pkg/utils"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{service: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var reqBody LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "bad request body format", http.StatusBadRequest)
		return
	}
	if (reqBody.User == "" && reqBody.Email == "") || reqBody.Password == "" {
		http.Error(w, "user or email and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.service.Auth(r.Context(), reqBody.User, reqBody.Email, reqBody.Password)
	if errors.Is(err, domain.ErrInvalidCredentials) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Printf("[Login]: failed to authenticate user: %v", err)
		http.Error(w, "failed to authenticate", http.StatusInternalServerError)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, token); err != nil {
		log.Printf("[Login]: failed to encode response: %v", err)
	}
}
