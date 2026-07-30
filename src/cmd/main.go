package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rafixcs/shorter-url-design-system/src/internal/server"
)

func main() {
	app, err := server.NewServer()
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	log.Print("Created new server")

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.Run()
	}()

	signalContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server stopped unexpectedly: %v", err)
		}
	case <-signalContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.Shutdown(shutdownContext); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}
}
