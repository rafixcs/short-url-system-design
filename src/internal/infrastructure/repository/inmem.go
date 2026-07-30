package repository

import (
	"context"
	"strings"
	"sync"

	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
)

var ErrShortURLNotFound = domain.ErrShortURLNotFound
var ErrUserNotFound = domain.ErrUserNotFound

type InMemRepository struct {
	mu    sync.RWMutex
	urls  map[string]*domain.ShorterUrlModel
	users map[string]*domain.UserModel
}

func NewInMemRepository() *InMemRepository {
	return &InMemRepository{
		urls:  make(map[string]*domain.ShorterUrlModel),
		users: make(map[string]*domain.UserModel),
	}
}

func (r *InMemRepository) CreateShortUrl(ctx context.Context, shorterUrl *domain.ShorterUrlModel) (*domain.ShorterUrlModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[shorterUrl.ShortURL] = shorterUrl
	return shorterUrl, nil
}

func (r *InMemRepository) GetLongUrl(ctx context.Context, shortURL string) (*domain.ShorterUrlModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, value := range r.urls {
		if value.ShortURL == shortURL {
			return value, nil
		}
	}

	return nil, domain.ErrShortURLNotFound
}

func (r *InMemRepository) CreateUser(ctx context.Context, user *domain.UserModel) (*domain.UserModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, storedUser := range r.users {
		if strings.EqualFold(storedUser.Email, user.Email) || strings.EqualFold(storedUser.Name, user.Name) {
			return nil, domain.ErrUserAlreadyExists
		}
	}

	r.users[user.ID.Hex()] = user
	return user, nil
}

func (r *InMemRepository) GetUser(ctx context.Context, userID string) (*domain.UserModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[userID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (r *InMemRepository) GetUserByEmailOrName(ctx context.Context, email, name string) (*domain.UserModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if email != "" && strings.EqualFold(user.Email, email) {
			return user, nil
		}
		if name != "" && strings.EqualFold(user.Name, name) {
			return user, nil
		}
	}

	return nil, domain.ErrUserNotFound
}
