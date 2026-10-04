---
name: pr-code-review
description: >-
  Review code quality, architecture compliance, Responsible Gaming rules, and test coverage for a pull request (PR) or git diff in the Spend Limit project.
  Use when the user asks to "review PR", "review code quality", "PR review", "review changes",
  "check diff", "code review", or audit code against project standards.
---

# Pull Request (PR) Code Review Guide

This skill defines the quality checklist, architectural rules, regulatory standards, clean code guidelines, and verification steps for reviewing code changes in the Spend Limit repository.

---

## 1. Core Review Principles

When reviewing a PR or git diff, evaluate against these essential pillars:

1. **Layered Architecture & Separation of Concerns**: Strict boundary separation across the presentation layer (`cli/` and REST API `server/`), application composition (`app/`), service orchestration (`service/`), storage/caching (`store/`), shared domain models (`domain/`), repository data access (`repository/`), and database persistence (`db/`).
2. **Clean Code & Go Idioms**: High readability, early returns (left-aligned happy path), no magic literals, structured logging (`slog`), and standard Go docstrings.
3. **Responsible Gaming (RG) & Financial Correctness**: Precision in whole NZD currency calculations (stored as integer cents) and strict enforcement of cooling-off periods.
4. **Data Integrity & Transaction Safety**: Correct use of database transactions (`WithinTransaction`), row locks (`FOR UPDATE`), and the Transactional Outbox Pattern for domain events.
5. **Security & Data Sanitization**: Parameterized SQL queries, payload bounds checking (1MB ceiling), input scrubbing, and zero credential leakage.
6. **Deterministic Testing & Edge-Case Coverage**: Zero-dependency unit testing in `service/` using `MockRepository` and `server/` using `net/http/httptest`, exhaustive edge-case and boundary coverage (flooring, clamping, multi-window interactions, failure injection), and 100% clean race-detection runs (`-race`).

---

## 2. PR Review Checklist

### A. Architectural Boundaries & Dependencies
- [ ] **CLI Layer (`cli/`)**:
  - Contains NO raw SQL queries or direct repository calls.
  - Interacts strictly via `app.App` or `*service.Service`.
  - Normalizes flags (supports both `-flag` and `--flag`).
  - Supports persistent `--cache-type` flag (`"memory"` or `"redis"`).
  - Exposes server execution via `spendlimit serve` or `spendlimit --server` with host, port, timeout, and cache options.
  - Validates inputs (`ValidateUUID`, `CleanPeriod`, `ValidatePeriod`) before calling `InitService(cmd)` (which silences usage display for subsequent business exceptions).
  - Commands use `RunE` and return errors rather than calling `os.Exit` directly.
- [ ] **REST API Server Layer (`server/`)**:
  - Managed by [server/server.go](file:///Users/oscargarcia/workspace/spend-limit/server/server.go) using standard Go 1.22+ `http.ServeMux` route definitions (`"METHOD /path"` pattern, e.g. `"POST /api/v1/customers/{id}/limits"`).
  - Contains NO business logic, raw SQL queries, or direct repository/store calls; delegates strictly to `*service.Service`.
  - Request body decoding enforces safety via `readJSON`:
    - Enforces a 1MB payload ceiling via `http.MaxBytesReader` to protect against memory exhaustion.
    - Prohibits unknown fields via `dec.DisallowUnknownFields()`.
    - Verifies no trailing data exists after the JSON value.
  - Path and query parameters (`{id}`, `{period}`) are sanitized and validated with domain helpers (`domain.ValidateUUID`, `domain.CleanPeriod`, `domain.ValidatePeriod`) before calling service methods.
  - Standardizes HTTP status codes and error mappings:
    - `200 OK` / `201 Created` for successful resource operations.
    - `400 Bad Request` (`ErrCodeBadRequest`, `ErrCodeInvalidAmount`) for malformed JSON, unparseable bodies, invalid UUIDs, or invalid period/amount formats.
    - `404 Not Found` (`ErrCodeNotFound`) for non-existent entities or customers without limits.
    - `409 Conflict` (`ErrCodeConflict`) for active pending requests or state conflicts.
    - `422 Unprocessable Entity` (`ErrCodeLimitReached`) for bets blocked by limit restrictions or RG rule violations.
    - `500 Internal Server Error` (`ErrCodeInternal`) for database failures, transaction rollbacks, or recovered panics.
  - Emits uniform JSON responses via `writeJSON` and `writeError` using the standard envelope structure:
    `{"error": {"code": "...", "message": "...", "details": [...]}}`.
  - Middleware pipeline is ordered and configured correctly in `server/server.go`:
    `RecoveryMiddleware` (recovers from panics, logs stack trace via `slog`, returns 500) -> `LoggingMiddleware` (structured request logging with method, URI, status, and latency) -> `RequestIDMiddleware` (injects or propagates `X-Request-ID` across headers and context) -> `CORSMiddleware` (handles cross-origin headers and `OPTIONS` preflight requests).
  - Lifecycle & Graceful Shutdown:
    - Traps `SIGINT`/`SIGTERM` and honors `ShutdownTimeout`.
    - Drains in-flight HTTP connections before calling `app.Close()` to safely terminate database pools and store connections.
- [ ] **Service Layer (`service/`)**:
  - Contains pure business logic, cooling-off rules, multi-window limit hierarchy, and transaction boundaries.
  - Contains NO raw SQL queries, database driver imports, or presentation code (CLI or HTTP).
  - Coordinates between `repository.Repository` for persistent storage and `store.Store` for fast counter caching.
  - Interacts strictly via `repository.Repository` interface.
  - Uses structured logging (`slog`) exclusively; no console printing (`fmt.Print*`).
- [ ] **Domain Layer (`domain/`)**:
  - Contains domain models, view models, and domain validators.
  - Has ZERO dependencies on other project packages (prevents circular imports).
- [ ] **Repository Layer (`repository/`)**:
  - Implements the `Repository` interface.
  - `mysql.go` encapsulates all SQL queries, parameter bindings, row scanning, and database dialect specifics.
  - Any new repository method added to `repository.Repository` must be implemented in both `mysql.go` AND `mock.go`.

### B. Clean Code & Go Idiomatic Practices
- [ ] **Guard Clauses & Early Returns**:
  - Functions return early on errors or invalid inputs to avoid deeply nested `if/else` ladders.
  - The successful execution path remains flat and left-aligned.
- [ ] **No Magic Numbers or Magic Strings**:
  - Forbid raw string literals like `"HOURS_24"`, `"CUSTOMER"`, `"BET_CASH"`.
  - Use domain constants: `domain.PeriodHours24`, `domain.SourceCustomer`, `domain.TxnTypeBetCash`.
  - Use domain helper functions (e.g. `domain.PeriodDuration(period)`) rather than hardcoded durations.
- [ ] **Separation of Presentation from Business Logic**:
  - No `fmt.Print*` or console formatting in `service/`, `server/`, `repository/`, or `store/` layers. Console output belongs exclusively in `cli/`.
  - All operational diagnostics in `server/`, `service/`, and `app/` must use structured logging (`log/slog`) with contextual attributes (`slog.String`, `slog.Int64`, `slog.Duration`).
- [ ] **Doc Comments on Exported Identifiers**:
  - Every exported struct, interface, function, and constant must have a Go docstring beginning with the identifier's name.
- [ ] **Organized & Grouped Imports**:
  - Imports must be organized into 3 distinct groups separated by empty lines, and sorted alphabetically within each group:
    ```go
    import (
        // 1. Standard library packages
        "context"
        "database/sql"
        "fmt"
        "log/slog"
        "net/http"
        "time"

        // 2. Third-party packages (external dependencies)
        "github.com/google/uuid"
        "github.com/shopspring/decimal"
        "github.com/spf13/cobra"

        // 3. Internal / local project packages
        "spend-limit/app"
        "spend-limit/config"
        "spend-limit/domain"
        "spend-limit/repository"
        "spend-limit/server"
        "spend-limit/service"
        "spend-limit/store"
    )
    ```
- [ ] **Single Responsibility Principle (SRP)**:
  - Command handlers in `cli/` only parse flags, validate parameters, invoke the service, and render terminal output.
  - HTTP handlers in `server/` only parse HTTP request envelopes/parameters, validate inputs, invoke the service, and serialize structured JSON responses with appropriate HTTP status codes. Business logic belongs strictly in `service/`.

### C. Responsible Gaming & Financial Logic
- [ ] **Currency Handling**:
  - All monetary values are whole New Zealand Dollars (NZD) represented as integer cents (`int64`, `amountCents % 100 == 0`).
  - No floating-point math used for persisting balances or evaluating limits.
- [ ] **Limit Changes**:
  - **Decreasing limit** (stricter): Takes effect **immediately**, restarts the rolling window (`anchorAt = now`), and updates counter.
  - **Increasing limit** (less restrictive): Must queue into `SpendLimitRequests` with a mandatory **24-hour cooling-off period** (`effectiveAt = now + 24h`).
  - **Deleting limit** (removal request): Must queue into `SpendLimitRequests` with `amountCents = NULL` and a mandatory **24-hour cooling-off period**.
  - **Cancellation of Pending Requests**: Can only be cancelled while still in cooling-off (`effectiveAt > now`).
- [ ] **Betting & Transactions**:
  - Cash bets are rejected if bet amount > remaining limit allowance or if any active limit is reached.
  - Rejection responses must clearly inform the user of remaining budget and restricting period.
  - Bonus bets (`BONUS_BET`) do not consume limit budget, but are blocked if customer has reached limit ($0 remaining).
  - Refunds subtract from active window spent balance (floored at $0).

### D. Concurrency, Transactions & Outbox Pattern
- [ ] **Atomic Demarcation**:
  - Multi-table writes (e.g. creating a limit + initializing counter, or recording a bet + updating counter) must run within `s.repo.WithinTransaction(ctx, func(txCtx context.Context) error { ... })`.
  - All repo calls inside the transaction callback must use `txCtx`.
- [ ] **Outbox Pattern**:
  - Domain events (`spend_limit.created`, `spend_limit.updated`, `spend_limit.reached`, `spend_limit.request_cancelled`, `spend_limit.deleted`) must be recorded into `event_outbox` inside the *same* database transaction as the business operation.
- [ ] **Locking**:
  - Race-sensitive reads during mutation operations must acquire row locks (`FOR UPDATE` via `GetActiveLimitsForUpdate` or `GetTransactionForUpdate`).

### E. Security, Data Protection & Sanitization
- [ ] **Parameterized SQL**:
  - All MySQL queries in `repository/mysql.go` must strictly use `?` placeholders. No string concatenation for user-supplied data in queries.
- [ ] **Zero Credential / Secret Leaks**:
  - Database connection strings, passwords, or authentication tokens must never be written to logs, error messages, or console output.
- [ ] **Input Scrubbing**:
  - String inputs (e.g., period values, UUIDs) are trimmed and normalized (`CleanPeriod`, `ValidateUUID`).

### F. Error Handling & Resource Management
- [ ] **Error Wrapping**:
  - Errors are wrapped with contextual messages: `fmt.Errorf("failed to ...: %w", err)`.
- [ ] **Context Propagation**:
  - `context.Context` is the first parameter of all service and repository methods.
  - HTTP handlers propagate `r.Context()` directly to service calls.
  - Use `time.Now().UTC()` for consistent timestamps across all database tables.
- [ ] **Resource Cleanup & Payload Protection**:
  - Request bodies are read with `http.MaxBytesReader` to guard against memory exhaustion.
  - Database connections, rows, and statements are properly closed with `defer`.
  - Server gracefully drains in-flight connections on shutdown before invoking `app.Close()` to release database and cache connections.

### G. Unit Testing & Edge-Case Coverage
- [ ] **Zero External Dependencies**:
  - Service-layer tests in [service/service_test.go](file:///Users/oscargarcia/workspace/spend-limit/service/service_test.go) and [service/publisher_test.go](file:///Users/oscargarcia/workspace/spend-limit/service/publisher_test.go) use `repository.NewMockRepository()`. No live MySQL or Docker containers needed.
  - Server-layer tests in [server/server_test.go](file:///Users/oscargarcia/workspace/spend-limit/server/server_test.go) use `net/http/httptest` (`httptest.NewRequest`, `httptest.NewRecorder` or `httptest.NewServer`) without binding external host ports.
- [ ] **Subtest Organization (`t.Run`)**:
  - Both happy paths and negative validation paths use `t.Run` subtests for clear reporting and isolation.
- [ ] **REST API & Handler Testing**:
  - **HTTP Status Codes & Response Envelopes**: Verify exact status codes (`200 OK`, `201 Created`, `400 Bad Request`, `404 Not Found`, `409 Conflict`, `422 Unprocessable Entity`, `500 Internal Server Error`) and structured JSON error format (`code`, `message`, `details`).
  - **Path Parameter & Input Validation**: Test invalid UUIDs (`/api/v1/customers/invalid-id/limits`), unsupported periods (`YEARS_1`), and invalid currency formats.
  - **Body Parsing & Malformed Payloads**: Test invalid JSON syntax, empty request bodies, unexpected/disallowed JSON fields (`DisallowUnknownFields`), and payload size limit enforcement (>1MB).
  - **Middleware Behaviors**:
    - `RecoveryMiddleware`: Panic in handler recovers safely, logs stack trace, and returns 500 JSON envelope without terminating the server.
    - `RequestIDMiddleware`: Echoes existing `X-Request-ID` or generates a valid UUID when absent.
    - `CORSMiddleware`: Verifies CORS headers on responses and preflight `OPTIONS` requests (204 No Content).
  - **Health Probe**: Test `/health` returns `200 OK` with database status `connected`, or `503 Service Unavailable` with `degraded` when database ping fails.
- [ ] **Responsible Gaming & Financial Edge Cases**:
  - **Limit Clamping & Reached Trigger**: Reducing limit below current spent balance immediately clamps remaining allowance to `$0.00` NZD (`remainingCents = 0`), sets `IsReached = true`, and emits `spend_limit.reached` with `immediate-limit-reduction`.
  - **Financial Flooring**: Refunds exceeding active spent balances floor at `$0.00` NZD (`spentCents = 0`), never negative.
  - **Multi-Window Hierarchy**: Bets evaluated against simultaneous active windows (e.g. 24h vs. 7d) are blocked if *any* window is exceeded, correctly identifying the restricting period.
  - **Bonus Bet Non-Consumption**: Generosity bets (`BONUS_BET`) succeed when budget > $0, but leave active spend counters untouched (`SpentCents` unchanged).
- [ ] **Operational & Boundary Edge Cases**:
  - **Unmatured Cooling-Off Periods**: `ApplyPendingChanges` ignores requests where cooling-off window is still active (`effectiveAt > now`).
  - **Tie-Breaker Logic**: Equal remaining budgets across different periods resolve in favor of the shorter duration period.
  - **Zero / Negative Input Rejection**: Zero or negative amounts (`amountNZD <= 0`) for bets, generosity, and refunds are rejected immediately with validation errors.
  - **Limit-Free Customers**: Operations on clients with no active limits handle nil states gracefully (bets succeed, limit lookup returns `nil, nil`).
  - **Non-Existent Entities**: Updating, clearing, or cancelling non-existent limits/requests returns appropriate not-found errors.
  - **Transaction Failure & Rollbacks**: Simulated repository errors (via `MockRepository` hooks) abort transactions without corrupting counters or persisting outbox events.
- [ ] **Race Conditions**:
  - All tests pass without data races (`go test -v -race ./...`).

---

## 3. Step-by-Step Review Procedure

When performing a PR review:

1. **Examine Changed Files & Git Diff**:
   ```bash
   git status -s
   git diff main...HEAD  # or git diff HEAD~1
   ```
2. **Code Formatting & Import Check**:
   ```bash
   # Check gofmt formatting
   test -z "$(gofmt -s -l .)" || (echo "Unformatted files found:" && gofmt -s -l .)
   ```
   - Visually inspect changed files to ensure imports follow the 3-tier grouping (Standard Library, Third-party, Local packages).
3. **Audit Against Checklists** (Sections 2A - 2G above).
4. **Run Static Analysis & Build**:
   ```bash
   go vet ./...
   go build -o /dev/null main.go
   ```
5. **Execute Full Test Suite with Race Detector**:
   ```bash
   go test -v -race -count=1 ./...
   ```
6. **Verify Documentation Sync**:
   - Check if new/modified CLI commands and REST API endpoints are documented in [README.md](file:///Users/oscargarcia/workspace/spend-limit/README.md) and Cobra's `Example` strings.
7. **Formulate the Review Report** using the template below.

---

## 4. PR Review Report Template

```markdown
## PR Review Summary

**PR Title / Change**: <Short description of PR>
**Verdict**: [APPROVE | REQUEST CHANGES | COMMENT]

### 1. Architectural & Design Compliance
- [x] Layering boundaries respected (`cli` / `server` -> `app` -> `service` -> `repository` / `store` -> `db`)
- [x] REST API standards followed (Go 1.22 mux routing, middleware pipeline, status codes, standard JSON envelopes)
- [x] Interface updated in both `mysql.go` and `mock.go`

### 2. Clean Code & Go Idioms
- [x] Imports properly grouped (Standard Lib / Third-party / Local) and sorted
- [x] Early returns / guard clauses used (flat happy path)
- [x] No magic numbers or strings (domain constants used)
- [x] Structured logging (`slog`) used across server and service layers (no `fmt.Print*`)
- [x] Exported symbols documented with standard Go docstrings

### 3. Responsible Gaming & Financial Correctness
- [x] Currency validated in whole NZD / integer cents
- [x] 24-hour cooling-off rule preserved for increases/removals
- [x] Immediate activation for limit decreases

### 4. Concurrency, Security & Transactions
- [x] Context-aware transaction (`WithinTransaction`) applied
- [x] Domain events written to `event_outbox` within transaction
- [x] Parameterized SQL queries (zero concatenation)
- [x] Payload size limits enforced (1MB MaxBytesReader, DisallowUnknownFields)
- [x] No secret or sensitive credential leakage

### 5. Unit Testing & Edge-Case Coverage
- [x] Zero-dependency tests via `MockRepository` and `net/http/httptest` (no live DB or Redis required)
- [x] REST API endpoints tested via `httptest` (status codes, JSON envelopes, validation, middleware)
- [x] Subtests (`t.Run`) used for happy and failure paths
- [x] Financial & RG edge cases verified (flooring at $0, limit clamping, multi-window constraints, bonus bet non-consumption)
- [x] Operational edge cases tested (unmatured cooling-off, tie-breaker logic, nil/empty limits)
- [x] Repository failure hooks (`OnCreateSpendLimit`, etc.) tested for rollback safety

### 6. Findings & Recommendations

#### 🔴 Blocker (Must Fix Before Merge)
- *File: `path/to/file.go:line`* — Description of critical bug, security flaw, or RG rule violation.

#### 🟡 Improvements / Suggestions (Non-blocking)
- *File: `path/to/file.go:line`* — Refactoring suggestion, performance enhancement, or naming improvement.

#### 🟢 Positive Highlights
- Clean abstraction, thorough unit test coverage, good use of subtests.

### 7. Automated Verification Results
- `gofmt -s`: PASS
- `go vet ./...`: PASS
- `go build`: PASS
- `go test -v -race -count=1 ./...`: PASS (X tests passed, 0 failures)
```
