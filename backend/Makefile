.PHONY: dev-rest dev-grpc build test clean testdata sqlc swag proto

all: build

sqlc:
	sqlc generate

dev-rest:
	air

dev-grpc:
	air -c .air-grpc.toml

testdata:
	go run cmd/datagenerator/main.go $(if $(SEED),--seed=$(SEED)) $(if $(ITERATIONS),--iterations=$(ITERATIONS))

swag:
	swag init -g ./routes.go -d ./handler --parseDependency

proto:
	protoc --go_out=./grpc/gen --go_opt=paths=source_relative \
		--go-grpc_out=./grpc/gen --go-grpc_opt=paths=source_relative \
		--proto_path=./proto \
		configuration.proto

build:
	make swag
	go build -o bin/server.exe ./cmd/server

test:
	go test ./...

cover:
	go test -coverprofile=tmp/coverage.out ./...
	go tool cover -html=tmp/coverage.out

clean:
	rm -rf bin/
	go clean