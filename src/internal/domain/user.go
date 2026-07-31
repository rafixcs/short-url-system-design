package domain

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

type UserModel struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
}

type UserService interface {
	CreateUser(ctx context.Context, name, email, password string) (*UserModel, error)
	GetUser(ctx context.Context, userID string) (*UserModel, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *UserModel) (*UserModel, error)
	GetUser(ctx context.Context, userID string) (*UserModel, error)
}
