---
name: memory-cache
description: >-
  Guidelines, patterns, and operational procedures for in-memory LRU caching, thread safety, eviction policies,
  active client indexing, out-of-order update protection, and metrics telemetry in the Spend Limit project.
  Use when working with store/memory/cache.go, modifying MemoryCache, adding cache telemetry, or testing in-memory counter stores.
---

# In-Memory Cache Store (`MemoryCache`)

This skill documents the architecture, concurrency model, lifecycle rules, testing patterns, and implementation guidelines for the in-memory LRU counter cache implemented in [store/memory/cache.go](file:///Users/oscargarcia/workspace/spend-limit/store/memory/cache.go).

---

## 1. Overview & Architectural Role

The Spend Limit service requires ultra-low latency validation of cash bets and refunds against multi-window spend limits (`HOURS_24`, `DAYS_7`, `DAYS_30`). To avoid database bottlenecks on hot paths, volatile spend counters are cached in memory.

- **Interface Contract**: Implements `store.Store` defined in [store/store.go](file:///Users/oscargarcia/workspace/spend-limit/store/store.go).
- **Primary Implementation**: `MemoryCache` (also aliased as `MemoryStore`) in [store/memory/cache.go](file:///Users/oscargarcia/workspace/spend-limit/store/memory/cache.go).
- **Core Technology**: HashiCorp's LRU cache (`github.com/hashicorp/golang-lru`) augmented with:
  - `sync.RWMutex` protecting cache operations and client index lookups.
  - Active client-periods index (`clientPeriods`) to track which periods are active per client.
  - Out-of-order update protection based on counter timestamps.
  - Eviction hook (`onEvict`) that automatically cleans up client period sets and tracks eviction telemetry.
  - Pluggable `AfterAddFunc` callback for logging or downstream replication.
  - Atomic telemetry counters (`hits`, `misses`, `evictions`) exposed via JSON metrics snapshot.

---

## 2. Component Architecture & Data Structures

```
┌────────────────────────────────────────────────────────────────────────┐
│                        MemoryCache Struct                              │
│                                                                        │
│   mu sync.RWMutex                                                      │
│   ├── cache *lru.Cache  ──► key: "spend_counter:<clientID>:<period>"   │
│   │                         value: *counterEntry                       │
│   │                                ├── counter: domain.SpendCounter    │
│   │                                ├── limitAmount: decimal.Decimal    │
│   │                                └── at: time.Time                   │
│   │                                                                    │
│   ├── clientPeriods map[string]map[string]struct{}                     │
│   │   └── clientID ──► { "HOURS_24": {}, "DAYS_7": {} }                │
│   │                                                                    │
│   ├── afterAdd AfterAddFunc                                            │
│   │                                                                    │
│   └── Observability (sync/atomic int64)                                │
│       ├── hits                                                         │
│       ├── misses                                                       │
│       └── evictions                                                    │
└────────────────────────────────────────────────────────────────────────┘
```

### Struct Definition & Options

```go
type MemoryCache struct {
    mu            sync.RWMutex
    cache         *lru.Cache
    capacity      int
    clientPeriods map[string]map[string]struct{}
    afterAdd      AfterAddFunc

    // Observability metrics
    hits      int64
    misses    int64
    evictions int64
}

// MemoryStore is an alias for MemoryCache
type MemoryStore = MemoryCache
```

### Constructors & Functional Options

The package supports configuration via the **Functional Options Pattern**:

```go
// Option defines a functional option for configuring a MemoryCache instance.
type Option func(*MemoryCache)

// WithCapacity sets the maximum capacity for the LRU cache (default: 10,000).
func WithCapacity(capacity int) Option

// WithAfterAdd sets the callback invoked after a counter is added or updated.
func WithAfterAdd(fn AfterAddFunc) Option

// NewMemoryCache initializes a default MemoryCache with optional functional options.
func NewMemoryCache(opts ...Option) *MemoryCache

// NewMemoryCacheWithSize creates a cache with an explicit capacity and afterAdd hook.
func NewMemoryCacheWithSize(size int, afterAdd AfterAddFunc) (*MemoryCache, error)

// NewMemoryStore creates a new MemoryStore instance (alias of NewMemoryCache).
func NewMemoryStore(opts ...Option) *MemoryStore
```

---

## 3. Core Mechanisms & Invariants

### 3.1 Cache Key Scheme
Keys in the LRU cache are formatted deterministically:
```go
func counterKey(clientID, period string) string {
    return fmt.Sprintf("spend_counter:%s:%s", clientID, period)
}
```

### 3.2 Active Client-Periods Index (`clientPeriods`)
`lru.Cache` stores flat key-value pairs. To validate multi-window constraints during a bet or refund without scanning all keys, `MemoryCache` maintains a secondary index:
```go
clientPeriods map[string]map[string]struct{}
```
- **Registration**: When a counter is added (`Add`, `Receive`, `TimeTravel`), the period is registered under `clientPeriods[clientID][period]`.
- **Deregistration on Delete**: When `DeleteCounter` is called, the period is removed from `clientPeriods[clientID]`. If no periods remain for the client, the client entry is removed from the map.
- **Deregistration on LRU Eviction**: When the cache reaches capacity and an entry is evicted, `onEvict` runs:
  ```go
  onEvict := func(key interface{}, value interface{}) {
      atomic.AddInt64(&c.evictions, 1)
      entry, ok := value.(*counterEntry)
      if !ok || entry == nil {
          return
      }
      if periods, exists := c.clientPeriods[entry.counter.ClientID]; exists {
          delete(periods, entry.counter.Period)
          if len(periods) == 0 {
              delete(c.clientPeriods, entry.counter.ClientID)
          }
      }
  }
  ```

### 3.3 Out-of-Order Update Protection
In high-concurrency or distributed environments, updates may arrive out of chronological order. Both `Add` and `Receive` guard against overwriting newer state with stale data:
```go
cKey := counterKey(counter.ClientID, counter.Period)
if curr, has := c.peek(cKey); has && !at.IsZero() && !curr.at.IsZero() && at.Before(curr.at) {
    // Incoming counter update is older than stored counter, so ignore it
    return false, nil
}
```
*Note: `peek` inspects the existing entry without altering its LRU eviction order.*

### 3.4 Hook Execution: `Add` vs `Receive`
- **`Add(counter, limitAmount, at)`**: Updates the cache and executes `afterAdd` callback if registered. Used when the local service node generates a state update.
- **`Receive(counter, limitAmount, at)`**: Updates the cache **without** firing `afterAdd`. Used when synchronizing updates received from external events or replicas to prevent infinite ping-pong loops.
- **`SetCounter(ctx, counter, limitAmount)`**: Implements `store.Store` by delegating to `c.Add(counter, limitAmount, counter.Updated)`.

### 3.5 Defensive Copying on Reads
To prevent external callers from mutating cached pointers without holding locks, `GetCounter` returns a shallow copy of `domain.SpendCounter`:
```go
atomic.AddInt64(&c.hits, 1)
entry := v.(*counterEntry)
copied := entry.counter
return &copied, true, nil
```

### 3.6 Two-Phase Atomic Bet Validation (`BetCash`)
When evaluating a cash bet:
1. **Acquire write lock**: `c.mu.Lock()` / `defer c.mu.Unlock()`.
2. **Lookup active periods**: Retrieve `c.clientPeriods[clientID]`. If empty, the bet is unconditionally allowed.
3. **Sort periods**: Period keys are sorted alphabetically (`sort.Strings(periods)`) for deterministic validation order.
4. **Validation Phase (Read-only check)**:
   - For each period, calculate `spent`:
     - **Lazy Window Rollover**: If `now.Unix() >= windowEnd.Unix()`, `spent` is treated as `decimal.Zero`.
   - Calculate remaining balance: `rem = limitAmount - spent`.
   - If `rem <= 0`: Block immediately (`"spend limit reached for period %s ($0.00 NZD remaining until reset)"`).
   - If `amount > rem`: Block immediately (`"spend limit exceeded for period %s. Remaining limit is $%s NZD..."`).
   - Prepare atomic update plan: `newSpent = spent + amount`.
   - Check if the new spent reaches or exceeds limit: mark period in `reachedPeriods`.
5. **Mutation Phase (Write)**:
   - If all active periods passed validation, commit the update plans to the cache entries.
   - Update `entry.counter.Spent = plan.newSpent`, `entry.counter.Updated = now`, and `entry.at = now`.
   - Re-insert into LRU cache (`c.cache.Add(plan.key, entry)`).
6. **Return**: `(allowed=true, reachedPeriods, blockReason="", remaining=0, nil)`.

### 3.7 Floored Cash Refunds (`RefundCash`)
Refunds subtract the refunded amount across all active client counters, strictly floored at zero:
```go
newSpent := entry.counter.Spent.Sub(amount)
if newSpent.IsNegative() {
    newSpent = decimal.Zero
}
entry.counter.Spent = newSpent
entry.counter.Updated = now
entry.at = now
c.cache.Add(cKey, entry)
```

---

## 4. Observability & Telemetry

### 4.1 Metrics Tracking
`MemoryCache` tracks telemetry using atomic operations (`sync/atomic`):
- `hits`: Incremented whenever `GetCounter` finds an existing entry.
- `misses`: Incremented whenever `GetCounter` fails to find an entry.
- `evictions`: Incremented inside `onEvict` whenever the LRU cache purges an entry.

### 4.2 Snapshot API
Exported methods allow inspecting cache performance:
```go
type CacheMetrics struct {
    Hits          int64   `json:"hits"`
    Misses        int64   `json:"misses"`
    Evictions     int64   `json:"evictions"`
    Len           int     `json:"len"`
    Capacity      int     `json:"capacity"`
    HitRate       float64 `json:"hit_rate"`
    ActiveClients int     `json:"active_clients"`
}

func (c *MemoryCache) Metrics() CacheMetrics
func (c *MemoryCache) Hits() int64
func (c *MemoryCache) Misses() int64
func (c *MemoryCache) Evictions() int64
func (c *MemoryCache) Len() int
func (c *MemoryCache) Capacity() int
func (c *MemoryCache) ActiveClients() int
func (c *MemoryCache) ResetMetrics()
```

### 4.3 REST API Endpoints
The metrics snapshot is exposed via the HTTP server in [server/handlers.go](file:///Users/oscargarcia/workspace/spend-limit/server/handlers.go):
- `GET /api/v1/admin/cache/metrics`
- `GET /api/v1/cache/metrics`
- `GET /api/v1/metrics/cache`

Response envelope:
```json
{
  "success": true,
  "data": {
    "hits": 1420,
    "misses": 38,
    "evictions": 12,
    "len": 850,
    "capacity": 10000,
    "hit_rate": 0.9739,
    "active_clients": 412
  }
}
```

---

## 5. Factory Integration

`MemoryCache` is integrated into the storage factory in [store/factory.go](file:///Users/oscargarcia/workspace/spend-limit/store/factory.go):

```go
case config.CacheTypeMemory:
    afterAdd := func(clientID, period string, counter domain.SpendCounter, limitAmount decimal.Decimal, at time.Time) {
        log.Printf("[memory-cache] afterAdd: clientID=%s period=%s counter=%+v limitAmount=%s at=%s",
            clientID, period, counter, limitAmount, at.Format(time.RFC3339))
    }
    return memory.NewMemoryCache(memory.WithAfterAdd(afterAdd)), nil
```

To configure memory cache via environment variables:
```bash
CACHE_TYPE=memory
```

---

## 6. Testing Patterns for `MemoryCache`

All tests in [store/memory/cache_test.go](file:///Users/oscargarcia/workspace/spend-limit/store/memory/cache_test.go) run in-process with zero external dependencies.

### 6.1 Basic Set & Get Test Pattern
```go
func TestMemoryCache_SetAndGetCounter(t *testing.T) {
    cache := memory.NewMemoryCache()
    ctx := context.Background()
    clientID := "c123"
    period := domain.PeriodHours24
    now := time.Now().UTC()

    counter := &domain.SpendCounter{
        ClientID:    clientID,
        Period:      period,
        WindowStart: now,
        WindowEnd:   now.Add(24 * time.Hour),
        Spent:       decimal.NewFromFloat(25.50),
        Updated:     now,
    }
    limitAmount := decimal.NewFromInt(100)

    if err := cache.SetCounter(ctx, counter, limitAmount); err != nil {
        t.Fatalf("SetCounter failed: %v", err)
    }

    got, exists, err := cache.GetCounter(ctx, clientID, period)
    if err != nil || !exists {
        t.Fatalf("expected counter to exist, err: %v", err)
    }

    if !got.Spent.Equal(decimal.NewFromFloat(25.50)) {
        t.Errorf("expected spent 25.50, got %s", got.Spent)
    }
}
```

### 6.2 Testing LRU Eviction & Client Deregistration
```go
func TestMemoryCache_Eviction(t *testing.T) {
    // Create cache with tiny capacity
    cache := memory.NewMemoryCache(memory.WithCapacity(2))
    ctx := context.Background()
    now := time.Now().UTC()

    // Add 3 entries to trigger eviction of the 1st
    for i := 1; i <= 3; i++ {
        cID := fmt.Sprintf("client-%d", i)
        cntr := &domain.SpendCounter{
            ClientID:    cID,
            Period:      domain.PeriodHours24,
            WindowStart: now,
            WindowEnd:   now.Add(24 * time.Hour),
            Spent:       decimal.Zero,
            Updated:     now,
        }
        _ = cache.SetCounter(ctx, cntr, decimal.NewFromInt(100))
    }

    if cache.Evictions() != 1 {
        t.Errorf("expected 1 eviction, got %d", cache.Evictions())
    }
    if cache.Len() != 2 {
        t.Errorf("expected cache len 2, got %d", cache.Len())
    }
}
```

### 6.3 Concurrency & Race Verification
When testing concurrent operations, always run with the Go race detector:
```bash
go test -v -race -run TestMemoryCache ./store/memory/...
```

---

## 7. Common Gotchas & Best Practices

1. **Deadlock Prevention**:
   - `Add` and `Receive` acquire `c.mu.Lock()`.
   - Never call `c.Add(...)` or `c.Receive(...)` from within a function that is already holding `c.mu` (use internal unexported helpers if nested lock acquisition is needed).
   - In `SetCounter`, do **not** lock `c.mu` before delegating to `c.Add`.
2. **Defensive Copies**:
   - `GetCounter` must **never** return a direct pointer to internal entry structures. Modifying the returned struct must not affect internal cache state.
3. **Decimal Arithmetic**:
   - Never use float arithmetic. Always use `decimal.Decimal` methods (`.Add()`, `.Sub()`, `.Equal()`, `.LessThanOrEqual()`).
4. **Window Rollovers**:
   - Remember that window rollovers are evaluated lazily in `BetCash`. If `nowUnix >= windowEndUnix`, `spent` is treated as `0` for validation, but persistent resets are executed via `ResetCounter`.
5. **Client Periods Map**:
   - Any method that creates a new entry MUST register the period in `c.clientPeriods[clientID]`.
   - Any method that removes or purges entries MUST clean up `c.clientPeriods[clientID]` and remove the client key when no periods remain.
