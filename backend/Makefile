.PHONY: dev build test clean install-deps seed sqlc swag testdata

# Default target
all: build

# Install dependencies
install-deps:
	go mod tidy
	go install github.com/cosmtrek/air@latest

sqlc:
	sqlc generate

dev:
	air

testdata:
	go run cmd/datagenerator/main.go $(if $(SEED),--seed=$(SEED))

swag:
	swag init -g ./routes.go -d ./handler --parseDependency

# Build the application
build:
	make swag
	go build -o bin/server.exe ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean