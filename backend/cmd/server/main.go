package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/necroskillz/config-service/server"
)

func main() {
	server := server.NewServer()

	go func() {
		if err := server.Start(); err != nil {
			if err == http.ErrServerClosed {
				log.Println("Server closed")
			} else {
				log.Fatalf("failed to start server: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to gracefully stop server: %v", err)
	}
}
