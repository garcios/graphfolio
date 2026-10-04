# Agent Guidelines & Repository Architecture (`AGENTS.md`)

This file contains the architectural rules, coding standards, and verification procedures that all AI agents operating in the **GraphFolio** repository must adhere to.


## 1. Project Overview

A modern portfolio tracker built for serious investors. Move beyond simple price charts to accurately measure your 
total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, 
consolidated dashboard.

## 2. Layered Architecture & Separation of Concerns

### Layer Breakdown
- **`proto/` (The Contract)**: Storing Protocol Buffers at the root ensures all microservices and the BFF generate their code from the exact same definitions. This prevents schema drift between the gRPC servers and the BFF client wrappers.

- **`pkg/` (Shared Infrastructure)**: Contains strictly non-domain-specific code. If it handles business logic, it belongs in a service. If it handles database connection pooling, logging formats, or JWT parsing interceptors, it lives here to be imported by all backend modules.

- **`services/` (The Domain)**: Each gRPC microservice is its own module with isolated domain logic (`internal/`). They only communicate with the database and, if necessary, other gRPC services. They do not know about GraphQL.

- **`bff/` (The Orchestrator)**: The Backend-for-Frontend acts as the translation layer. Using a code-first or schema-first GraphQL generator (like gqlgen) here is ideal. The resolvers map GraphQL queries to one or more gRPC calls, stitch the data together, and return the exact payload the frontend needs, shielding the UI from the microservice topology.

- **`web/` (The Consumer)**: A standalone frontend directory (React, Vue, etc.) that interacts exclusively with the GraphQL BFF. Code generation tools (like Apollo Codegen or GraphQL Code Generator) can point directly to the `bff/graph/schema.graphqls` file to generate strongly-typed frontend hooks.

## 3. Project Structure

```text
├── proto/                        # 1. Single Source of Truth for APIs
│   ├── portfolio/v1/
│   │   └── portfolio.proto
│   └── user/v1/
│       └── user.proto
│
├── pkg/                          # 2. Shared Libraries (Backend)
│   ├── logger/                   # Standardized logging
│   ├── middleware/               # gRPC interceptors (auth, metrics)
│   ├── database/                 # Postgres/DB connection managers
│   └── go.mod
│
├── services/                     # 3. Core gRPC Microservices
│   ├── portfolio-api/
│   │   ├── cmd/server/           # Application entrypoint
│   │   ├── internal/             # Domain logic, handlers, repository
│   │   └── go.mod
│   └── user-api/
│       ├── cmd/server/
│       ├── internal/
│       └── go.mod
│
├── bff/                          # 4. GraphQL Backend-for-Frontend
│   ├── cmd/server/               # BFF entrypoint
│   ├── graph/                    # GraphQL definitions (e.g., via gqlgen)
│   │   ├── schema.graphqls       # Federated or standalone schema
│   │   ├── schema.resolvers.go   # Resolver implementations
│   │   └── model/                # Generated models
│   ├── internal/
│   │   └── clients/              # gRPC client wrappers to communicate with services/
│   └── go.mod
│
├── web/                          # 5. Frontend Application
│   ├── src/
│   │   ├── components/           # Reusable UI elements
│   │   ├── features/             # Domain-specific frontend modules
│   │   ├── graphql/              # Frontend queries/mutations and generated hooks
│   │   └── pages/                # Route components
│   ├── package.json
│   └── tsconfig.json
│
├── tools/                        # Code generation and dev scripts
│   └── tools.go                  # Version locking for protoc-gen-go, gqlgen, etc.
│
├── Makefile                      # Standardized commands (make proto, make build)
├── docker-compose.yml            # Local dev orchestration
└── go.work                       # Go workspace mapping (pkg, services/*, bff)
```

## 4. Monorepo Tooling Strategy

For a polyglot monorepo like this, a standard Makefile is often enough to orchestrate protoc generation, gqlgen execution, and Docker builds. If the frontend tooling becomes complex, adopting a build system like Turborepo or Bazel allows you to cache tasks—ensuring the frontend only rebuilds when `web/` or `bff/graph/` changes, and gRPC services only rebuild when their specific `services/` directory or the `proto/` directory changes.