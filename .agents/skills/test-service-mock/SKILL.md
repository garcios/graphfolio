---
name: test-service-mock
description: >-
  Guide and best practices for implementing service-layer unit tests using MockRepository in the Spend Limit project.
  Use when the user asks to "write a unit test for service", "test service layer", "add unit test using mock",
  "mock repository test", "test-service-mock", or test business logic without a live database.
---

# Service-Layer Unit Testing with MockRepository

This skill provides the architectural guidelines, step-by-step workflow, and code patterns for writing zero-dependency unit tests for the `service.Service` using `MockRepository`.

---

## 1. Testing Philosophy & Mock Architecture

The Spend Limit application uses the **Repository Pattern** to completely isolate business logic from database infrastructure:

- **Target under test**: [service/service.go](file:///Users/oscargarcia/workspace/spend-limit/service/service.go) (`*service.Service`).
- **In-memory mock**: [repository/mock.go](file:///Users/oscargarcia/workspace/spend-limit/repository/mock.go) (`*repository.MockRepository`).
- **Test location**: [service/service_test.go](file:///Users/oscargarcia/workspace/spend-limit/service/service_test.go).
- **Zero external dependencies**: Tests run fast without requiring MySQL, Docker, or external mock generators.

### Capabilities of `MockRepository`

```go
type MockRepository struct {
    Limits       map[string]*domain.SpendLimit          // Key: "clientID:period"
    Counters     map[string]*domain.SpendCounter        // Key: "clientID:period"
    Requests     map[string]*domain.SpendLimitRequest   // Key: "clientID:period"
    Transactions map[string]*domain.Transaction         // Key: txnID
    Events       []*domain.OutboxEventRecord            // Appended on CreateOutboxEvent

    // Failure injection hooks
    OnCreateSpendLimit  func(limit *domain.SpendLimit) error
    OnCreateTransaction func(txn *domain.Transaction) error
}
```

---

## 2. Step-by-Step Implementation Workflow

### Step 1: Instantiate MockRepository and Service
Every test or subtest should create a fresh in-memory repository to guarantee test isolation:

```go
repo := repository.NewMockRepository()
svc := service.NewService(repo)
ctx := context.Background()
clientID := "11111111-2222-3333-4444-555555555555"
```

### Step 2: Seed Initial State (if needed)
There are two ways to set up state prior to testing:

#### Approach A: Using High-Level Service Methods (Recommended for standard flows)
```go
// Creates the limit and initializes the counter
detail, err := svc.CreateSpendLimit(ctx, clientID, service.PeriodHours24, 100, service.SourceCustomer, "customer-1")
if err != nil {
    t.Fatalf("failed setup: %v", err)
}
```

#### Approach B: Direct Mock Injection (Recommended for time travel, expired windows, or edge states)
```go
now := time.Now().UTC()
_ = repo.UpsertSpendLimitRequest(ctx, &domain.SpendLimitRequest{
    ClientID:    clientID,
    Period:      service.PeriodHours24,
    AmountCents: sql.NullInt64{Int64: 20000, Valid: true},
    EffectiveAt: now.Add(-1 * time.Hour), // Expired cooling-off period
    Created:     now.Add(-25 * time.Hour),
})
```

### Step 3: Execute the Service Method
Invoke the business logic method being tested:

```go
result, err := svc.CreateBetCashTxn(ctx, clientID, 40.0, "txn-001")
```

### Step 4: Assert Returns and Errors
Check both positive outputs and negative error messages:

```go
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
if !result.Allowed {
    t.Errorf("expected bet to be allowed")
}
if result.RemainingNZD != 60.0 {
    t.Errorf("expected remaining NZD to be 60.0, got %f", result.RemainingNZD)
}
```

### Step 5: Verify Underlying Mock State Mutations
Verify that side-effects were recorded correctly in repository tables:

```go
// 1. Verify updated counter
counter, err := repo.GetSpendCounter(ctx, clientID, service.PeriodHours24)
if err != nil || counter == nil {
    t.Fatalf("counter not found: %v", err)
}
if counter.SpentCents != 4000 {
    t.Errorf("expected spentCents 4000, got %d", counter.SpentCents)
}

// 2. Verify audit transaction
txn, err := repo.GetTransactionForUpdate(ctx, "txn-001")
if err != nil || txn == nil {
    t.Fatalf("transaction not recorded")
}

// 3. Verify domain events published to outbox (if applicable)
if len(repo.Events) == 0 {
    t.Errorf("expected outbox event to be created")
}
```

### Step 6: Test Failure Simulation (Hooks)
To test how `Service` handles repository errors and transaction rollbacks:

```go
repo.OnCreateTransaction = func(txn *domain.Transaction) error {
    return fmt.Errorf("simulated database disk error")
}

_, err := svc.CreateBetCashTxn(ctx, clientID, 50.0, "txn-fail")
if err == nil {
    t.Fatalf("expected error when repository fails, got nil")
}
```

---

## 3. Standard Test Template

Use Go subtests (`t.Run`) for testing multiple scenarios (happy path, validation failures, boundary conditions):

```go
func TestFeatureName(t *testing.T) {
    ctx := context.Background()
    clientID := "11111111-2222-3333-4444-555555555555"

    t.Run("success scenario", func(t *testing.T) {
        repo := repository.NewMockRepository()
        svc := service.NewService(repo)

        // 1. Arrange
        _, _ = svc.CreateSpendLimit(ctx, clientID, service.PeriodHours24, 100, service.SourceCustomer, "cust")

        // 2. Act
        res, err := svc.SomeOperation(ctx, clientID, ...)

        // 3. Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if res == nil {
            t.Fatalf("expected result, got nil")
        }
    })

    t.Run("boundary or restriction exceeded", func(t *testing.T) {
        repo := repository.NewMockRepository()
        svc := service.NewService(repo)

        // Arrange & Act
        _, err := svc.SomeOperation(ctx, clientID, ...)

        // Assert error
        if err == nil {
            t.Fatalf("expected error, got nil")
        }
    })
}
```

---

## 4. Verification

Run the service test suite with race detection enabled:

```bash
# Run all service tests
go test -v -race ./service/...

# Run a specific test
go test -v -race -run TestFeatureName ./service/...
```
