package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/repository"
	"github.com/rafixcs/shorter-url-design-system/src/pkg/utils"
)

type UserHandler struct {
	service domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{service: userService}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("[CreateUser]: failed to decode request body: %v", err)
		http.Error(w, "bad request body format", http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(
		r.Context(),
		reqBody.Name,
		reqBody.Email,
		reqBody.Password,
	)
	if err != nil {
		log.Printf("[CreateUser]: failed to create user: %v", err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	h.writeUserResponse(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		userID = chi.URLParam(r, "user_id")
	}
	if userID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("[GetUser]: failed to get user: %v", err)
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	h.writeUserResponse(w, http.StatusOK, user)
}

func (h *UserHandler) writeUserResponse(w http.ResponseWriter, status int, user *domain.UserModel) {
	response := UserResponse{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
	}

	if err := utils.WriteJSON(w, status, response); err != nil {
		log.Printf("[UserResponse]: failed to encode response: %v", err)
	}
}
