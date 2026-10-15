# User Story: Portfolio Service & Deterministic Ledger Projections

> **Story ID**: US-SVC-101  
> **Epic**: Portfolio Engine & Domain Microservices  
> **Target**: Portfolio Microservice (`services/portfolio-api`), Repository (`internal/repository/`), Calculation Engine (`internal/domain/calculator.go`), BFF GraphQL Orchestrator (`bff/`), Primary Investor Application (`web/apps/main-app`)  
> **Status**: **DONE**  
> **Implementation Plan**: [portfolio-service-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-service-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Features #2, #3, #4, #5)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L12-L15)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** investor accessing GraphFolio,  
> **I want to** have my portfolio holdings, open tax lots, cash balances, and performance metrics calculated deterministically from my transaction history by an event-sourced gRPC service and served via GraphQL,  
> **So that** my portfolio view reflects real-time, audited financial data with exact zero floating-point loss across all currency conversions and return calculations.

---

## 2. Business Context & Core Principles

1. **Transition from Mock to Live Domain Microservice**:
   - Replaces initial mock responses in `services/portfolio-api` with live PostgreSQL repository queries (`pgx/v5`) and domain calculations.
2. **Event-Sourced Deterministic Ledger Projections**:
   - The transaction ledger (`portfolio.transactions`) is the single immutable source of truth.
   - Projections for open tax lots (`portfolio.tax_lots`), lot disposals (`portfolio.lot_disposals`), current holdings (`portfolio.holdings`), and cash balances (`portfolio.cash_balances`) are computed deterministically by replaying historical transactions through `ProcessLedger` under atomic database row locks (`SELECT ... FOR UPDATE`).
3. **Exact Fixed-Point Arithmetic (Zero Floating-Point Drift)**:
   - All monetary amounts, stock quantities, and return percentages use `shopspring/decimal.Decimal` in Go, custom `common.v1.Decimal` and `Money` in Protobuf, and custom `Decimal` scalars in GraphQL.
4. **End-to-End Orchestration via GraphQL BFF**:
   - A dedicated Go BFF (`bff/` using `gqlgen`) translates frontend GraphQL queries (`me`, `portfolio`, `portfolioHistory`) into gRPC calls, insulating the React client from microservice network topologies.

---

## 3. Acceptance Criteria

### 3.1 Repository & Query Layer
- [x] **AC-1 (Hand-Written SQL Queries)**: `internal/repository/postgres.go` implements strongly-typed, parameterized SQL queries using `pgxpool.Pool` for transaction retrieval, holding queries, and projection persistence.
- [x] **AC-2 (Atomic Projection Persistence)**: `SaveProjectionsTx` executes within an atomic transaction, replacing active tax lots, holdings, and cash balances in a single commit.

### 3.2 Domain Engine & Return Math
- [x] **AC-3 (Deterministic Ledger Replay)**: `projection.go` (`ProcessLedger`) iterates transactions chronologically, creating lots for buys, allocating lots for sells (supporting `AVERAGE_COST` and `FIFO`), updating cash for deposits/withdrawals, and adjusting quantities for stock splits.
- [x] **AC-4 (Exact Financial Formulas)**:
  - Portfolio Value: $\sum (\text{Holding Value} \times \text{FX}) + \text{Cash Balance}$.
  - Today's Return: $(\text{Latest Price} - \text{Previous Close}) \times \text{Quantity} \times \text{FX}$.
  - Total Return: $\text{Current Value} - \text{Cost Basis} + \text{Realized PnL} + \text{Dividends}$.
- [x] **AC-5 (Strict Decimal Types)**: Zero `float32` or `float64` types utilized across domain entities, repositories, or calculation helpers.

### 3.3 gRPC Server & BFF GraphQL Integration
- [x] **AC-6 (gRPC PortfolioServiceServer)**: `services/portfolio-api/internal/server.go` implements `GetPortfolio`, `AddTransaction`, `ListTransactions`, and `DeleteTransaction`.
- [x] **AC-7 (BFF GraphQL Schema & Resolvers)**: `bff/graph/schema.resolvers.go` resolves `portfolio` queries, maps Protobuf decimal types to GraphQL `Decimal` strings, and forwards claims to microservices.
- [x] **AC-8 (Dashboard UI Data Binding)**: Investor dashboard in `web/apps/main-app` renders live holdings, portfolio hero cards, and cash balances from GraphQL queries.

---

## 4. Technical & Implementation Notes

- **Microservice Structure**: `services/portfolio-api/` follows clean architecture (`internal/domain/`, `internal/repository/`, `internal/service/`, `internal/server.go`).
- **BFF Orchestrator**: `bff/cmd/server/main.go` and `bff/graph/`.
- **Decimal Mapping**: `pkg/decimalpb/decimalpb.go` handles conversions between Protobuf messages and `shopspring/decimal.Decimal`.
