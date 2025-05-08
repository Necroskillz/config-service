# Config Service

## Overview

Config Service is a microservice for managing configuration data, with a RESTful API and an admin frontend. It stores and retrieves configuration data from a database. The configuration structure is hierarchical, starting from a Service, with ServiceVersions, Features, FeatureVersions, Keys and Values with variations of properties.

## Tech Stack
- Go (Echo, REST API, swagger, jwt)
- Postgres (sqlc)
- Frontend: pnpm, @tanstack/react-start, tailwindcss, shadcn, @tanstack/react-form, react-query, react-table, @kubb/cli

## Prerequisites

### Installing Dependencies on Windows (install with [Scoop](https://scoop.sh/))

Open PowerShell and run:

to install `scoop`:
```
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
```

to install dependencies:
```
scoop install go postgresql dbmate fnm make
go install github.com/air-verse/air@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

- **Go**: Backend language
- **PostgreSQL**: Database
- **db-mate**: Database migration runner
- **fnm**: Node.js version manager (or you can install node directly or with some other version manager)
- **make**: Script runner
- **air**: Live reload for Go
- **sqlc**: SQL to Go code generator
- **swag**: OpenAPI doc and schema generator

### Node.js Setup (with fnm)

After installing `fnm`, set up your shell (PowerShell example):

include this in your PS profile:
```
fnm env --use-on-cd --shell powershell | Out-String | Invoke-Expression
```

to install specific node version:
```
fnm use --install-if-missing 23
corepack enable pnpm
```

## Environment Setup

Copy the example environment file to create your local configuration:

```
cd backend
cp .env.example .env
```

Edit `.env` as needed for your local database and environment settings.

## Install Dependencies

From the project root:

```
make install
```

This installs frontend and backend dependencies.

## Database Setup

- Ensure PostgreSQL is running

```
cd backend
dbmate create
make seed
```

## Running the Project

Use VSCode tasks (recommended):
- Open the Command Palette (Ctrl+Shift+P)
- `Tasks: Run Task` > `dev (all)` to start both backend and frontend

Or from the command line:

```
make dev-backend
make dev-frontend
```

Frontend will start at http://localhost:3000. Use username `admin`, password `admin` to login.

## Regenerating Code

- **sqlc (DB codegen):**
  - `make sqlc` (or run the "sqlc" VSCode task)
- **OpenAPI/Swagger:**
  - `make swag` (or run the "swag" VSCode task)
