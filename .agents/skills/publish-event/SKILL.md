---
name: publish-event
description: >-
  Guide and step-by-step procedures for implementing and publishing domain events via the Transactional Outbox Pattern in the Spend Limit project.
  Use when the user asks to "publish an event", "add an event", "implement event publishing", "outbox event",
  "publish-event", or emit domain events from the service layer.
---

# Publishing Domain Events (Transactional Outbox Pattern)

This skill provides the architectural conventions, step-by-step implementation guide, and testing patterns for publishing domain events in the Spend Limit project using the **Transactional Outbox Pattern**.

---

## 1. Event Publishing Architecture

In the Spend Limit service, domain events are stored in the `event_outbox` table rather than directly published to external brokers (e.g., Kafka, RabbitMQ). This guarantees:
- **Atomic Operations**: The event is written inside the **same database transaction** (`WithinTransaction`) as the business state change. If the database transaction rolls back, the event is never written.
- **At-Least-Once Delivery**: Downstream workers/relays poll or stream `event_outbox` where `status = 'PENDING'` to broadcast externally.

### Key Components

- **Contract & Implementation**: [service/publisher.go](file:///Users/oscargarcia/workspace/spend-limit/service/publisher.go)
  - `EventPublisher`: Interface defining typed publishing methods.
  - `OutboxPublisher`: Concrete implementation writing to `repository.Repository`.
- **Database Table**: `event_outbox` defined in [db/schema.sql](file:///Users/oscargarcia/workspace/spend-limit/db/schema.sql).
- **Outbox Model**: `domain.OutboxEventRecord` in [domain/domain.go](file:///Users/oscargarcia/workspace/spend-limit/domain/domain.go).

---

## 2. Step-by-Step Implementation Workflow

Follow these 5 steps whenever adding a new domain event:

### Step 1: Define the Event Subject Constant
In [service/publisher.go](file:///Users/oscargarcia/workspace/spend-limit/service/publisher.go), declare a new constant for the event subject following the `noun.verb` format:

```go
const (
    EventSpendLimitCreated = "spend_limit.created"
    EventSpendLimitUpdated = "spend_limit.updated"
    EventSpendLimitReached = "spend_limit.reached"
    EventSpendLimitDeleted = "spend_limit.deleted" // New event
)
```

### Step 2: Define the Event Payload Struct
Create a dedicated payload struct with JSON serialization tags. Always use explicit types, integer cents for monetary values, and UTC timestamps:

```go
// SpendLimitDeletedPayload is the payload for spend_limit.deleted.
type SpendLimitDeletedPayload struct {
    ClientID    string    `json:"client_id"`
    Period      string    `json:"period"`
    RequestedBy string    `json:"requested_by"`
    DeletedAt   time.Time `json:"deleted_at"`
}
```

### Step 3: Add Method to `EventPublisher` Interface
Add the typed publish signature to `EventPublisher` in [service/publisher.go](file:///Users/oscargarcia/workspace/spend-limit/service/publisher.go):

```go
type EventPublisher interface {
    PublishSpendLimitCreated(ctx context.Context, detail *SpendLimitDetail) (string, error)
    PublishSpendLimitUpdated(...) (string, error)
    PublishSpendLimitReached(...) (string, error)
    PublishSpendLimitDeleted(ctx context.Context, clientID, period, requestedBy string) (string, error) // Add here
    PublishEvent(ctx context.Context, subject string, payload interface{}, headers map[string]string) (string, error)
}
```

### Step 4: Implement Method on `OutboxPublisher`
Implement the method by assembling the payload, headers (e.g. `client-id`, `period`), and delegating to `p.PublishEvent(...)`:

```go
// PublishSpendLimitDeleted publishes the spend_limit.deleted event.
func (p *OutboxPublisher) PublishSpendLimitDeleted(ctx context.Context, clientID, period, requestedBy string) (string, error) {
    payload := SpendLimitDeletedPayload{
        ClientID:    clientID,
        Period:      period,
        RequestedBy: requestedBy,
        DeletedAt:   time.Now().UTC(),
    }
    headers := map[string]string{
        "client-id": clientID,
        "period":    period,
    }
    return p.PublishEvent(ctx, EventSpendLimitDeleted, payload, headers)
}
```

> **Note**: `p.PublishEvent(...)` automatically attaches standard headers (`content-type: application/json`, `event-type: <subject>`), generates a UUID `event_id`, and sets initial status to `'PENDING'`.

### Step 5: Invoke from `Service` within Transaction
In [service/service.go](file:///Users/oscargarcia/workspace/spend-limit/service/service.go), call the publisher method inside the transactional block (`WithinTransaction`).

> [!IMPORTANT]
> **Always pass the transactional context `txCtx`** (not the outer `ctx`) to ensure the event record is committed atomically with the table updates!

```go
err := s.repo.WithinTransaction(ctx, func(txCtx context.Context) error {
    // 1. Perform database state mutation using txCtx
    if err := s.repo.DeleteSpendLimit(txCtx, clientID, period); err != nil {
        return err
    }

    // 2. Publish outbox event using txCtx
    if s.publisher != nil {
        if _, err := s.publisher.PublishSpendLimitDeleted(txCtx, clientID, period, requestedBy); err != nil {
            return fmt.Errorf("failed to publish delete event: %w", err)
        }
    }

    return nil
})
```

---

## 3. Unit Testing Event Publishing

Unit tests must verify:
1. The publisher properly serializes JSON payload and headers.
2. The event record is appended to `MockRepository.Events`.
3. The event rolls back if the business transaction fails.

### Example Unit Test in `service/service_test.go`

```go
func TestPublishSpendLimitDeleted(t *testing.T) {
    ctx := context.Background()
    repo := repository.NewMockRepository()
    pub := NewOutboxPublisher(repo)

    clientID := "11111111-2222-3333-4444-555555555555"
    eventID, err := pub.PublishSpendLimitDeleted(ctx, clientID, PeriodHours24, "agent-007")
    if err != nil {
        t.Fatalf("unexpected publish error: %v", err)
    }
    if eventID == "" {
        t.Errorf("expected non-empty eventID")
    }

    // Verify event in mock repository
    if len(repo.Events) != 1 {
        t.Fatalf("expected 1 event in outbox, found %d", len(repo.Events))
    }

    event := repo.Events[0]
    if event.Subject != EventSpendLimitDeleted {
        t.Errorf("expected subject %s, got %s", EventSpendLimitDeleted, event.Subject)
    }
    if event.Status != "PENDING" {
        t.Errorf("expected status PENDING, got %s", event.Status)
    }

    // Unmarshal and verify payload
    var payload SpendLimitDeletedPayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        t.Fatalf("failed to unmarshal event payload: %v", err)
    }
    if payload.ClientID != clientID || payload.Period != PeriodHours24 {
        t.Errorf("unexpected payload values: %+v", payload)
    }
}
```

---

## 4. Verification & Inspection

1. **Run Unit Tests**:
   ```bash
   go test -v -race -run TestPublish ./service/...
   ```

2. **Inspect Outbox via CLI** (against live database):
   ```bash
   # List all events in event_outbox
   ./spendlimit list-events

   # Filter by client ID
   ./spendlimit list-events -client-id 11111111-2222-3333-4444-555555555555
   ```
