package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	usersCollection    = "users"
	shortURLCollection = "short_urls"
)

type MongoRepository struct {
	users     *mongo.Collection
	shortURLs *mongo.Collection
}

func NewMongoRepository(ctx context.Context, database *mongo.Database) (*MongoRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}

	repository := &MongoRepository{
		users:     database.Collection(usersCollection),
		shortURLs: database.Collection(shortURLCollection),
	}
	if err := repository.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("create MongoDB indexes: %w", err)
	}

	return repository, nil
}

func (r *MongoRepository) createIndexes(ctx context.Context) error {
	caseInsensitive := &options.Collation{Locale: "en", Strength: 2}
	_, err := r.users.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetName("users_email_unique").SetUnique(true).SetCollation(caseInsensitive)},
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetName("users_name_unique").SetUnique(true).SetCollation(caseInsensitive)},
	})
	if err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}

	_, err = r.shortURLs.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "short_url", Value: 1}},
		Options: options.Index().SetName("short_urls_code_unique").SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create short URL index: %w", err)
	}

	return nil
}

func (r *MongoRepository) CreateShortUrl(ctx context.Context, shortURL *domain.ShorterUrlModel) (*domain.ShorterUrlModel, error) {
	if _, err := r.shortURLs.InsertOne(ctx, shortURL); err != nil {
		return nil, fmt.Errorf("insert short URL: %w", err)
	}
	return shortURL, nil
}

func (r *MongoRepository) GetLongUrl(ctx context.Context, shortURL string) (*domain.ShorterUrlModel, error) {
	var result domain.ShorterUrlModel
	err := r.shortURLs.FindOne(ctx, bson.M{"short_url": shortURL}).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrShortURLNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find short URL: %w", err)
	}
	return &result, nil
}

func (r *MongoRepository) CreateUser(ctx context.Context, user *domain.UserModel) (*domain.UserModel, error) {
	if _, err := r.users.InsertOne(ctx, user); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *MongoRepository) GetUser(ctx context.Context, userID string) (*domain.UserModel, error) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return r.findUser(ctx, bson.M{"_id": id})
}

func (r *MongoRepository) GetUserByEmailOrName(ctx context.Context, email, name string) (*domain.UserModel, error) {
	filters := make([]bson.M, 0, 2)
	if value := strings.TrimSpace(email); value != "" {
		filters = append(filters, bson.M{"email": value})
	}
	if value := strings.TrimSpace(name); value != "" {
		filters = append(filters, bson.M{"name": value})
	}
	if len(filters) == 0 {
		return nil, domain.ErrUserNotFound
	}
	return r.findUser(ctx, bson.M{"$or": filters})
}

func (r *MongoRepository) findUser(ctx context.Context, filter bson.M) (*domain.UserModel, error) {
	var user domain.UserModel
	err := r.users.FindOne(ctx, filter, options.FindOne().SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 2,
	})).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &user, nil
}

var _ domain.UserRepository = (*MongoRepository)(nil)
var _ domain.AuthRepository = (*MongoRepository)(nil)
var _ domain.ShorterUrlRepository = (*MongoRepository)(nil)
