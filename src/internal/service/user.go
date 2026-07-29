package service

import (
	"context"
	"fmt"

	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.UserModel, error) {
	user := &domain.UserModel{
		ID:       primitive.NewObjectID(),
		Name:     name,
		Email:    email,
		Password: password,
	}

	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*domain.UserModel, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

var _ domain.UserService = (*UserService)(nil)
