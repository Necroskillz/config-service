package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/necroskillz/config-service/grpc"
)

func main() {
	server := grpc.NewServer()

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
		log.Println("Server closed")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Fatalf("failed to gracefully stop server: %v", err)
	}
}
