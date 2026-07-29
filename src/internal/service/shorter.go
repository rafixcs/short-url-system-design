package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrGetLongURLFromShortURL = errors.New("failed to get long url from short url")

type UrlService struct {
	repo domain.ShorterUrlRepository
}

func NewService(repo domain.ShorterUrlRepository) *UrlService {
	return &UrlService{
		repo: repo,
	}
}

func (s *UrlService) CreateShortUrl(ctx context.Context, userID string, longURL string) (*domain.ShorterUrlModel, error) {

	shorURL, err := s.generate(longURL)
	if err != nil {
		return nil, fmt.Errorf("[URLService.CreateShorUrl] Error generating short url: %v", err)
	}

	model := domain.ShorterUrlModel{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		LongURL:   longURL,
		ShortURL:  shorURL,
		CreatedAt: time.Now(),
	}
	return s.repo.CreateShortUrl(ctx, &model)
}

func (s *UrlService) GetLongUrl(ctx context.Context, shortURL string) (string, error) {
	content, err := s.repo.GetLongUrl(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("get long URL: %w", err)
	}

	return content.LongURL, nil
}

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	idLength = 6
)

func (s *UrlService) generate(rawURL string) (string, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid URL")
	}

	id := make([]byte, idLength)
	random := make([]byte, idLength)

	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	for i, value := range random {
		id[i] = alphabet[int(value)%len(alphabet)]
	}

	return string(id), nil
}

func (s *UrlService) generateShortURL(
	rawURL string,
	exists func(string) (bool, error),
) (string, error) {
	for attempts := 0; attempts < 10; attempts++ {
		id, err := s.generate(rawURL)
		if err != nil {
			return "", err
		}

		taken, err := exists(id)
		if err != nil {
			return "", err
		}

		if !taken {
			return id, nil
		}
	}

	return "", errors.New("could not generate a unique short ID")
}
