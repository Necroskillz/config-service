dev-backend:
	cd backend && make dev

dev-frontend:
	cd frontend && pnpm dev

swag:
	cd backend && make swag
	cd frontend && pnpm generate

sqlc:
	cd backend && make sqlc

test:
	cd backend && make test
	cd frontend && pnpm test || echo "No frontend tests defined"

build:
	cd backend && make build
	cd frontend && pnpm build

install:
	cd frontend && pnpm install
	cd backend && go mod tidy
