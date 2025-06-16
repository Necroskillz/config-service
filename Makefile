dev-backend-rest:
	cd backend && make dev-rest

dev-backend-grpc:
	cd backend && make dev-grpc

dev-frontend:
	cd frontend && pnpm dev

swag:
	cd backend && make swag
	cd frontend && pnpm generate

sqlc:
	cd backend && make sqlc

proto:
	cd backend && make proto
	cd go-client && make proto

test:
	cd backend && make test
	cd frontend && pnpm test || echo "No frontend tests defined"

build:
	cd backend && make build
	cd frontend && pnpm build

install:
	cd frontend && pnpm install
	cd backend && go mod tidy
