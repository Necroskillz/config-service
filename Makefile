.PHONY: dev build test clean install-deps seed sqlc

# Default target
all: build

# Install dependencies
install-deps:
	go mod tidy
	go install github.com/cosmtrek/air@latest

sqlc:
	sqlc generate

# Run the server with hot reload using Air
dev-server:
	air

dev-css:
	tailwindcss -i ./views/assets/css/input.css -o ./views/assets/css/output.css --watch

dev-views:
	templ generate --watch

dev:
	make -j3 dev-server dev-css dev-views

seed:
	go run cmd/seed/main.go

# Build the application
build:
	templ generate
	tailwindcss -i ./views/assets/css/input.css -o ./views/assets/css/output.css -m
	go build -o bin/server.exe ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean