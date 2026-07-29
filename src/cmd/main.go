package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/rafixcs/shorter-url-design-system/src/internal/server"
)

func main() {
	app := server.NewServer()
	if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped unexpectedly: %v", err)
	}
}
