---
name: write-unit-test
description: >-
  Standard guidelines, idioms, and canonical templates for writing Go unit tests in the Spend Limit codebase.
  Use when the user asks to "write a unit test", "add unit tests", "create Go test", "test function",
  "table-driven test", "write-unit-test", or ensure Go testing convention compliance.
---

# Go Unit Testing Guidelines & Patterns

This skill documents the conventions, testing idioms, mock strategies, and canonical templates for implementing unit tests in the **Spend Limit** repository.

---

## 1. Invocation Context

Activate or reference this skill whenever:
- Creating new `*_test.go` files or adding test cases to existing test suites (`service`, `repository`, `db`, `cli`, `domain`).
- Converting ad-hoc tests into table-driven unit tests.
- Testing domain rules, boundary constraints, decimal calculations, or Responsible Gaming validations.
- Simulating repository failures, database transactions, or event outbox publishing.
- Reviewing PR test coverage against project architecture and testing standards.

---

## 2. Core Testing Principles & Conventions

### 2.1 Testing & Assertion Libraries
- **Standard Library Only**: Use Go's built-in `testing` package exclusively (`t.Run`, `t.Errorf`, `t.Fatalf`).
- **No Third-Party Assertion Frameworks**: Do **NOT** import `github.com/stretchr/testify` (`assert`/`require`) or `github.com/google/go-cmp`.
- **Decimal Assertions**: Never use `==` or `!=` on `decimal.Decimal`. Use `.Equal()`, `.IsZero()`, or `.IsNegative()` from `github.com/shopspring/decimal`.
- **Failures vs Errors**:
  - `t.Fatalf`: Use when a failure invalidates subsequent assertions (e.g., unexpected error during setup, nil pointer dereference prevention, missing expected record).
  - `t.Errorf`: Use for non-fatal value mismatches so all failed fields in a test case are reported in a single run.

### 2.2 Package Placement & Visibility
- **Internal / White-Box Testing**: Declare test files under the same package name as the code being tested (e.g., `package service`, `package repository`, `package db`).
- Do **NOT** use black-box package naming (`package service_test`). Same-package placement allows direct access to unexported domain constants, error helpers, and constructor functions without polluting public APIs.

### 2.3 Subtest Naming & Parallelism
- **Naming Convention**: Subtest names passed to `t.Run()` should be concise, lowercase phrases describing the specific input condition or scenario:
  - Good: `"valid whole dollar"`, `"invalid negative amount"`, `"cooling off already expired"`, `"success cancel pending increase"`
  - Avoid: `"Test_Success_1"`, `"CaseA"`, `"ShouldWork"`
- **No `t.Parallel()`**: Do **NOT** call `t.Parallel()` in unit tests. Tests run sequentially to prevent race conditions on shared mock repository maps, failure hooks, and time-travel modifications. The full test suite is verified with `-race`.

### 2.4 Mocking & Stubbing Strategy
- **Service Layer (`service/`)**:
  - Use the handwritten in-memory double: `repository.NewMockRepository()`.
  - Do **NOT** use mock generation tools (`mockery`, `gomock`).
  - Pre-seed state using high-level service methods (`svc.CreateSpendLimit`) for standard flows, or direct mock injection (`repo.Limits[...]`, `repo.UpsertSpendLimitRequest`) for simulated clock drift, expired cooling-off periods, or edge states.
  - Test failure propagation and transaction rollbacks using mock hooks (`repo.OnCreateSpendLimit`, `repo.OnCreateTransaction`).
- **Repository / Infrastructure Layer (`repository/`, `db/`)**:
  - Use `github.com/DATA-DOG/go-sqlmock` to mock `*sql.DB`.
  - Always quote regexes via `regexp.QuoteMeta()`.
  - Always assert `mock.ExpectationsWereMet()` at the end of each test.

### 2.5 Context & Time Handling
- Always use `context.Background()` for unit test contexts.
- Always use `time.Now().UTC()` when seeding timestamps or testing rolling windows.

---

## 3. Canonical Test Skeletons

### Pattern A: Table-Driven Test (Validation & Pure Logic)

Use this pattern for input validators, pure domain functions, and boundary checks:

```go
package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  decimal.Decimal
		wantErr bool
	}{
		{
			name:    "valid whole dollar",
			amount:  decimal.NewFromInt(100),
			wantErr: false,
		},
		{
			name:    "valid two decimal places",
			amount:  decimal.NewFromFloat(10.50),
			wantErr: false,
		},
		{
			name:    "invalid zero",
			amount:  decimal.Zero,
			wantErr: true,
		},
		{
			name:    "invalid negative",
			amount:  decimal.NewFromFloat(-10.50),
			wantErr: true,
		},
		{
			name:    "invalid three decimal places",
			amount:  decimal.RequireFromString("10.123"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%s) err = %v, wantErr %v", tt.amount.String(), err, tt.wantErr)
			}
		})
	}
}
```

---

### Pattern B: Service Method with MockRepository & Subtests

Use this pattern when testing business logic, outbox events, limits, and transactions:

```go
package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"spend-limit/domain"
	"spend-limit/repository"
)

func TestFeatureServiceOperation(t *testing.T) {
	ctx := context.Background()
	clientID := "11111111-2222-3333-4444-555555555555"

	t.Run("success updates limit and outbox", func(t *testing.T) {
		repo := repository.NewMockRepository()
		svc := NewService(repo)

		// 1. Arrange - Seed prerequisite state
		initialAmount := decimal.NewFromInt(100)
		_, err := svc.CreateSpendLimit(ctx, clientID, PeriodHours24, initialAmount, SourceCustomer, "user-1")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// 2. Act - Execute business logic
		newAmount := decimal.NewFromInt(50)
		detail, isPending, err := svc.UpdateSpendLimit(ctx, clientID, PeriodHours24, newAmount, SourceCustomer)

		// 3. Assert - Verify return values
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isPending {
			t.Errorf("expected immediate update (isPending=false), got true")
		}
		if !detail.Amount.Equal(newAmount) {
			t.Errorf("expected amount %s, got %s", newAmount.String(), detail.Amount.String())
		}

		// 4. Assert - Verify repository state mutation
		stored, err := repo.GetSpendLimit(ctx, clientID, PeriodHours24)
		if err != nil || stored == nil {
			t.Fatalf("expected stored limit, got nil: %v", err)
		}
		if !stored.Amount.Equal(newAmount) {
			t.Errorf("expected stored amount %s, got %s", newAmount.String(), stored.Amount.String())
		}

		// 5. Assert - Verify transactional outbox events
		if len(repo.Events) == 0 {
			t.Fatalf("expected outbox event to be recorded")
		}
		lastEvent := repo.Events[len(repo.Events)-1]
		if lastEvent.Subject != EventSpendLimitUpdated {
			t.Errorf("expected event subject %s, got %s", EventSpendLimitUpdated, lastEvent.Subject)
		}
	})

	t.Run("simulated repository failure returns wrapped error", func(t *testing.T) {
		repo := repository.NewMockRepository()
		svc := NewService(repo)

		// Inject repository failure hook
		repo.OnCreateSpendLimit = func(limit *domain.SpendLimit) error {
			return fmt.Errorf("simulated db failure")
		}

		_, err := svc.CreateSpendLimit(ctx, clientID, PeriodHours24, decimal.NewFromInt(100), SourceCustomer, "user-1")
		if err == nil {
			t.Fatalf("expected error on repository failure, got nil")
		}
	})
}
```

---

### Pattern C: SQL Repository Test with `sqlmock`

Use this pattern for SQL data access and raw query testing in `repository/` or `db/`:

```go
package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"

	"spend-limit/domain"
)

func TestMySQLRepository_GetSpendLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewMySQLRepository(db)
	ctx := context.Background()
	clientID := "11111111-2222-3333-4444-555555555555"
	now := time.Now().UTC()
	expectedAmount := decimal.NewFromInt(100)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT clientID, period, amount, anchorAt, source, created, updated FROM SpendLimits WHERE clientID = ? AND period = ?")).
		WithArgs(clientID, domain.PeriodHours24).
		WillReturnRows(sqlmock.NewRows([]string{"clientID", "period", "amount", "anchorAt", "source", "created", "updated"}).
			AddRow(clientID, domain.PeriodHours24, "100.00", now, domain.SourceCustomer, now, now))

	fetched, err := repo.GetSpendLimit(ctx, clientID, domain.PeriodHours24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched == nil {
		t.Fatalf("expected limit, got nil")
	}
	if !fetched.Amount.Equal(expectedAmount) {
		t.Errorf("expected amount %s, got %s", expectedAmount.String(), fetched.Amount.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}
```

---

## 4. Test Verification & CLI Commands

Always run the standard verification pipeline to ensure formatting, static checks, race safety, and clean passes:

```bash
# 1. Run all unit tests with race detection and fresh count
go test -v -race -count=1 ./...

# 2. Run tests for a specific package
go test -v -race -count=1 ./service/...

# 3. Run a specific test or subtest by pattern
go test -v -race -run TestValidateAmount ./service/...
go test -v -race -run "TestValidateAmount/valid_whole_dollar" ./service/...

# 4. Check test coverage
go test -cover ./service/...

# 5. Full pre-commit check (format, vet, tests)
test -z "$(gofmt -s -l .)" || (echo "Unformatted files found:" && gofmt -s -l . && exit 1)
go vet ./...
go test -v -race -count=1 ./...
```

### Pass Criteria
- Exit status `0`.
- Zero race detector warnings (`WARNING: DATA RACE`).
- All table subtests pass without unhandled assertions.
