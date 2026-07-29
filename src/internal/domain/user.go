package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
	ID       primitive.ObjectID `json:"id"`
	Name     string             `json:"name"`
	Email    string             `json:"email"`
	Password string             `json:"password"`
}

type UserService interface {
	CreateUser(ctx context.Context, name, email, password string) (*UserModel, error)
	GetUser(ctx context.Context, userID string) (*UserModel, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *UserModel) (*UserModel, error)
	GetUser(ctx context.Context, userID string) (*UserModel, error)
}
