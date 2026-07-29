package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShorterUrlModel struct {
	ID        primitive.ObjectID `json:"id"`
	UserID    string             `json:"user_id"`
	LongURL   string             `json:"long_url"`
	ShortURL  string             `json:"short_url"`
	CreatedAt time.Time          `json:"created_at"`
}

type ShorterUrlService interface {
	CreateShortUrl(ctx context.Context, userID string, longURL string) (*ShorterUrlModel, error)
	GetLongUrl(ctx context.Context, shortURL string) (string, error)
}

type ShorterUrlRepository interface {
	CreateShortUrl(ctx context.Context, shorterUrl *ShorterUrlModel) (*ShorterUrlModel, error)
	GetLongUrl(ctx context.Context, shortURL string) (*ShorterUrlModel, error)
}
