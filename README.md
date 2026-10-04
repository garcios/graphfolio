# GraphFolio

A modern portfolio tracker built for serious investors. GraphFolio accurately measures your total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, consolidated dashboard.

## Key Features

- **Consolidated Dashboard**: View all your investments across multiple asset classes in a single, unified view.
- **Accurate Return Metrics**: Calculate total return, annualized returns (TWR), and daily fluctuations accurately.
- **Microservice Architecture**: A robust backend architecture powered by Go, gRPC, and GraphQL.
- **Responsive UI**: A beautiful, dark-themed, glassmorphic UI built with React and Vite.

## Tech Stack

- **Backend (Microservices)**: Go, gRPC, Protocol Buffers
- **Backend-for-Frontend (BFF)**: Go, GraphQL (`gqlgen`)
- **Frontend**: React, TypeScript, Vite, `genql`
- **Orchestration**: Docker Compose, Make

## Prerequisites

- Go 1.21 or higher
- Node.js 20+ and npm
- Make
- Protobuf Compiler (`protoc`) - *Optional, for regenerating protobufs*

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
This will run `npm install` for the frontend and `go mod download` for the backend modules.

### 3. Generate Models and API Clients

GraphFolio uses a strictly typed contract-first approach. Generate the GraphQL BFF models and the frontend typed client:

```bash
make generate
```
This runs `gqlgen` in the BFF and `@genql/cli` in the Web frontend, ensuring the entire stack is perfectly typed.

### 4. Start Development Servers

You can start both the Go GraphQL server and the Vite React frontend simultaneously with a single command:

```bash
make run
```
- The frontend will be available at [http://localhost:5173](http://localhost:5173) (or whichever port Vite selects).
- The GraphQL Playground (BFF) will be available at [http://localhost:8080/](http://localhost:8080/).

## Architecture

GraphFolio uses a layered microservice monorepo structure. For deep dive architectural rules, see [`AGENTS.md`](./AGENTS.md).

### Directory Structure

```text
├── proto/                        # Single Source of Truth for APIs (Protocol Buffers)
├── pkg/                          # Shared Go Infrastructure (logging, DB connections)
├── services/                     # Core gRPC Microservices
│   ├── portfolio-api/            # Portfolio business logic
│   └── user-api/                 # User management logic
├── bff/                          # GraphQL Backend-for-Frontend (gqlgen)
│   ├── graph/                    # Schema and resolvers mapping to gRPC
│   └── cmd/server/               # BFF entrypoint
├── web/                          # Frontend React Application
│   ├── src/components/           # UI Components (e.g. Dashboard)
│   └── src/generated/            # genql auto-generated typed client
├── docs/                         # Additional project documentation
└── Makefile                      # Standardized commands (install, generate, run)
```

### Data Flow

```text
React Component (Dashboard) 
  → GenQL Typed Client 
    → GraphQL Query (http://localhost:8080/query)
      → BFF Resolver (schema.resolvers.go)
        → [Future: gRPC Call to services/portfolio-api]
          → Data returned to frontend
```

## Available Scripts

We rely heavily on the `Makefile` at the root of the project to orchestrate tasks across the monorepo:

| Command | Description |
| --- | --- |
| `make install` | Installs dependencies for both Go (backend) and Node (frontend) |
| `make generate` | Regenerates backend GraphQL models and frontend `genql` SDK |
| `make run-bff` | Starts only the Go GraphQL backend on port 8080 |
| `make run-web` | Starts only the Vite React dev server |
| `make run` | Starts both the BFF and Web servers concurrently |

## Documentation Reference

For more detailed technical documentation:
- **Repository Architecture**: [`AGENTS.md`](./AGENTS.md)
- **Frontend GraphQL Setup**: [`docs/genql-usage.md`](./docs/genql-usage.md)
