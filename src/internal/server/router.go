package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httptransport "github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/http"
)

func NewRouter(userHandler *httptransport.UserHandler, shorterHandler *httptransport.ShorterHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/{shortURL}", shorterHandler.RedirectToSourceUrl)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"its alive!"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)
			r.Get("/{userID}", userHandler.GetUser)
		})

		r.Route("/urls", func(r chi.Router) {
			r.Post("/", shorterHandler.CreateShortUrl)
		})
	})

	return r
}
