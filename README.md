# GraphFolio

A modern portfolio tracker built for serious investors. GraphFolio accurately measures your total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, consolidated dashboard.

## Key Features

- **Consolidated Dashboard**: View all your investments across multiple asset classes in a single, unified view.
- **Accurate Return Metrics**: Calculate total return, annualized returns (TWR), and daily fluctuations accurately using exact decimal precision.
- **Microservice Architecture**: A robust backend architecture powered by Go, gRPC, and GraphQL.
- **PostgreSQL Persistence**: Transaction-ledger source of truth, tax-lot accounting (AVERAGE_COST & FIFO), and projections.
- **Responsive UI**: A beautiful, dark-themed, glassmorphic UI built with React and Vite.

## Tech Stack

- **Backend (Microservices)**: Go, gRPC, Protocol Buffers
- **Database & Persistence**: PostgreSQL 17+ (18+ recommended), `pgx/v5`, `golang-migrate`
- **Backend-for-Frontend (BFF)**: Go, GraphQL (`gqlgen`)
- **Frontend**: React, TypeScript, Vite, `genql`
- **Orchestration**: Make

## Prerequisites

- Go 1.26 or higher
- Node.js 20+ and npm
- PostgreSQL 17+ (18+ recommended) running locally at `localhost:5432` (`psql`, `pg_isready`)
- `golang-migrate` CLI (`brew install golang-migrate`)
- Protobuf Compiler (`protoc`) - *Optional, for regenerating protobufs*
- Make

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/garcios/graphfolio.git
cd graphfolio
```

### 2. Install Dependencies

We use a unified `Makefile` to simplify local development.

```bash
make install
```
This will run `npm install` for the frontend and download backend Go module dependencies.

### 3. Configure and Bootstrap Database

Ensure your local PostgreSQL instance is running on `localhost:5432`:

```bash
# Optional: copy and inspect environment configuration
cp .env.example .env

# Bootstrap roles, database, and service schemas
make db-bootstrap

# Run all schema migrations
make migrate-up

# Seed development data
make db-seed

# (Or run all of the above in one step)
make db-reset
```

### 4. Generate Models and API Clients

GraphFolio uses a strictly typed contract-first approach. Generate Protocol Buffers, GraphQL BFF models, and the frontend typed client:

```bash
make generate
```

### 5. Start Development Servers

You can start the backend services and the Vite React frontend simultaneously:

```bash
make run
```
- The frontend will be available at [http://localhost:5173](http://localhost:5173).
- The GraphQL Playground (BFF) will be available at [http://localhost:8080/](http://localhost:8080/).
- The Portfolio gRPC service will be available on `:50051`.

## Architecture

GraphFolio uses a layered microservice monorepo structure. For deep dive architectural rules, see [`AGENTS.md`](./AGENTS.md).

### Directory Structure

```text
├── proto/                        # Single Source of Truth for APIs (Protocol Buffers)
│   ├── common/v1/                # Shared types (Decimal, Money)
│   ├── portfolio/v1/             # Portfolio service definition
│   └── user/v1/                  # User service definition
├── pkg/                          # Shared Go Infrastructure (pkg/database, pkg/decimalpb)
├── scripts/db/                   # Database bootstrap and teardown scripts
├── services/                     # Core gRPC Microservices
│   ├── portfolio-api/            # Portfolio business logic, migrations, and seeds
│   └── user-api/                 # User management logic and migrations
├── bff/                          # GraphQL Backend-for-Frontend (gqlgen)
│   ├── graph/                    # Schema, custom scalars, resolvers mapping to gRPC
│   └── cmd/server/               # BFF entrypoint
├── web/                          # Frontend React Application
│   ├── src/components/           # UI Components (e.g. Dashboard)
│   └── src/generated/            # genql auto-generated typed client
├── docs/                         # Additional project documentation
└── Makefile                      # Standardized commands (bootstrap, migrate, run)
```

### Data Flow

```text
React Component (Dashboard) 
  → GenQL Typed Client 
    → GraphQL Query (http://localhost:8080/query)
      → BFF Resolver (schema.resolvers.go)
        → gRPC Call to services/portfolio-api (localhost:50051)
          → PostgreSQL (graphfolio.portfolio schema via pgx/v5)
            → Exact Decimal/Money returned to frontend
```

## Available Scripts

| Command | Description |
| --- | --- |
| `make install` | Installs dependencies for Go and Node modules |
| `make proto` | Compiles `.proto` definitions into Go packages |
| `make generate` | Runs proto, gqlgen, and genql code generation |
| `make db-check` | Verifies local PostgreSQL is reachable on localhost:5432 |
| `make db-bootstrap` | Idempotently creates service roles, DB, and schemas |
| `make db-drop` | Drops local development database |
| `make migrate-up` | Applies all pending migrations across services |
| `make migrate-down` | Rolls back one migration step across services |
| `make db-seed` | Populates development reference data, demo portfolios, and market prices |
| `make db-reset` | Runs db-drop, db-bootstrap, migrate-up, and db-seed |
| `make test` | Runs Go test suites |
| `make run-portfolio` | Starts the Go gRPC Portfolio API on port 50051 |
| `make run-user` | Starts the Go gRPC User API on port 50052 |
| `make run-bff` | Starts the Go GraphQL backend on port 8080 |
| `make run-web` | Starts the Vite React dev server |
| `make run` | Starts Portfolio API, BFF, and Web servers concurrently |

## Documentation Reference

- **Repository Architecture**: [`AGENTS.md`](./AGENTS.md)
- **Database Implementation Plan**: [`docs/db-implementation-plan.md`](./docs/db-implementation-plan.md)
- **Frontend GraphQL Setup**: [`docs/genql-usage.md`](./docs/genql-usage.md)
