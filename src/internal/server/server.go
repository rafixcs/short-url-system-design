package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/database"
	httptransport "github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/http"
	"github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/repository"
	"github.com/rafixcs/shorter-url-design-system/src/internal/service"
	"github.com/rafixcs/shorter-url-design-system/src/pkg/env"
	"go.mongodb.org/mongo-driver/mongo"
)

const startupTimeout = 10 * time.Second

type AppServer struct {
	Port        string
	server      http.Server
	mongoClient *mongo.Client
}

func NewServer() (*AppServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	log.Printf("Starting new server")
	a := env.GetString("MONGO_URI", "mongodb://localhost:27017")
	b := env.GetString("MONGO_DATABASE", "shorter_url")
	log.Println(a)
	log.Println(b)

	client, mongoDatabase, err := database.ConnectMongo(ctx, database.MongoConfig{
		URI:      env.GetString("MONGO_URI", "mongodb://localhost:27017"),
		Database: env.GetString("MONGO_DATABASE", "shorter_url"),
	})
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	repo, err := repository.NewMongoRepository(ctx, mongoDatabase)
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("initialize repository: %w", err)
	}

	port := env.GetString("PORT", "8080")
	jwtSecret := env.GetString("JWT_SECRET", "development-only-change-me")
	userService := service.NewUserService(repo)
	shorterService := service.NewService(repo)
	authService := service.NewAuthService(repo, jwtSecret, 24*time.Hour)
	handler := NewRouter(
		httptransport.NewUserHandler(userService),
		httptransport.NewShorterHandler(shorterService),
		httptransport.NewAuthHandler(authService),
		httptransport.NewAuthMiddleware(authService),
	)

	return &AppServer{
		Port:        port,
		mongoClient: client,
		server: http.Server{
			Addr:         fmt.Sprintf(":%s", port),
			Handler:      handler,
			IdleTimeout:  time.Minute,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}, nil
}

func (a *AppServer) Run() error {
	log.Printf("Running server on PORT %s", a.Port)
	return a.server.ListenAndServe()
}

func (a *AppServer) Shutdown(ctx context.Context) error {
	serverErr := a.server.Shutdown(ctx)
	databaseErr := a.mongoClient.Disconnect(ctx)
	if serverErr != nil {
		return fmt.Errorf("shutdown HTTP server: %w", serverErr)
	}
	if databaseErr != nil {
		return fmt.Errorf("disconnect MongoDB: %w", databaseErr)
	}
	return nil
}
