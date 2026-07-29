package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	httptransport "github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/http"
	"github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/repository"
	"github.com/rafixcs/shorter-url-design-system/src/internal/service"
	"github.com/rafixcs/shorter-url-design-system/src/pkg/env"
)

type AppSever struct {
	Port   string
	server http.Server
}

func NewServer() *AppSever {
	server := &AppSever{}
	server.initializeServer()
	return server
}

func (a *AppSever) initializeServer() {
	a.Port = env.GetString("PORT", "8080")
	jwtSecret := env.GetString("JWT_SECRET", "development-only-change-me")

	repo := repository.NewInMemRepository()
	userService := service.NewUserService(repo)
	shorterService := service.NewService(repo)
	authService := service.NewAuthService(repo, jwtSecret, 24*time.Hour)
	userHandler := httptransport.NewUserHandler(userService)
	shorterHandler := httptransport.NewShorterHandler(shorterService)
	authHandler := httptransport.NewAuthHandler(authService)
	authMiddleware := httptransport.NewAuthMiddleware(authService)
	handler := NewRouter(userHandler, shorterHandler, authHandler, authMiddleware)

	a.server = http.Server{
		Addr:         fmt.Sprintf(":%s", a.Port),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func (a *AppSever) Run() error {
	log.Printf("Running server on PORT %s", a.Port)
	return a.server.ListenAndServe()
}
