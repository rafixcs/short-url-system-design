package domain

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrShortURLNotFound = errors.New("short URL not found")

type ShorterUrlModel struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	LongURL   string             `bson:"long_url" json:"long_url"`
	ShortURL  string             `bson:"short_url" json:"short_url"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type ShorterUrlService interface {
	CreateShortUrl(ctx context.Context, userID string, longURL string) (*ShorterUrlModel, error)
	GetLongUrl(ctx context.Context, shortURL string) (string, error)
}

type ShorterUrlRepository interface {
	CreateShortUrl(ctx context.Context, shorterUrl *ShorterUrlModel) (*ShorterUrlModel, error)
	GetLongUrl(ctx context.Context, shortURL string) (*ShorterUrlModel, error)
}
