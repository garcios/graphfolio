---
name: db-migrate
description: >-
  Run, apply, verify, or troubleshoot database schema migrations and table setup for the Spend Limit service.
  Use when the user asks to "migrate database", "run migrations", "apply schema", "db-migrate",
  "initialize tables", or reset/truncate database tables for testing.
---

# Database Migration Runbook (`db-migrate`)

This skill guides the agent on how to apply schema migrations, verify tables, and manage database lifecycle for the Spend Limit application.

## Overview

The Spend Limit service uses embedded SQL migrations defined in [db/schema.sql](file:///Users/oscargarcia/workspace/spend-limit/db/schema.sql) and executed via [db/db.go](file:///Users/oscargarcia/workspace/spend-limit/db/db.go) through the `spendlimit db-migrate` Cobra CLI command.

---

## Prerequisites & Environment Configuration

Ensure the MySQL instance is reachable. The service uses the following environment variables (with defaults):

| Variable | Description | Default |
| :--- | :--- | :--- |
| `DB_HOST` | MySQL hostname or IP | `127.0.0.1` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL user | `root` |
| `DB_PASSWORD` | MySQL password | `Password123` |
| `DB_NAME` | Target database name | `spend_limit` |

If MySQL is running locally via Homebrew:
```bash
brew services start mysql
```

Ensure the target database exists prior to migration:
```bash
mysql -u root -pPassword123 -e "CREATE DATABASE IF NOT EXISTS spend_limit;"
```

---

## Applying Migrations

Execute the migration command using either `go run` or the compiled binary:

### Option A: Using `go run` (Development)
```bash
go run main.go db-migrate
```

### Option B: Using compiled binary
```bash
# Build if binary does not exist
go build -o spendlimit main.go

# Run migration
./spendlimit db-migrate
```

### Expected Output
```text
✓ Database migration applied successfully. Tables created / verified.
```

---

## Schema & Tables Created

All statements in [db/schema.sql](file:///Users/oscargarcia/workspace/spend-limit/db/schema.sql) use `CREATE TABLE IF NOT EXISTS`, making migrations idempotent. The following tables are created/verified:

1. **`SpendLimits`**: Active customer spend limits (`clientID`, `period`, `amountCents`, `anchorAt`, `source`).
2. **`SpendLimitRequests`**: Pending limit increases/deletions queued under the 24-hour cooling-off rule (`effectiveAt`).
3. **`SpendCounters`**: Rolling window spend trackers (`windowStart`, `windowEnd`, `spentCents`).
4. **`SpendLimitLocks`**: Per-client concurrency lock table.
5. **`Transactions`**: Audit log of customer bets, refunds, and bonus transactions.
6. **`event_outbox`**: Transactional outbox table for reliable asynchronous domain event publishing.

---

## Verification Steps

After running `db-migrate`, verify the database state:

1. **Verify Table Creation**:
   ```bash
   mysql -u root -pPassword123 -D spend_limit -e "SHOW TABLES;"
   ```

2. **Run Unit & Integration Tests**:
   ```bash
   go test -v ./repository/...
   ```

3. **Check Service Connectivity**:
   ```bash
   go run main.go list-limits -client-id 10000000-0000-0000-0000-000000000001
   ```

---

## Related Database Operations

- **Truncate All Tables** (Wipe data while keeping schema intact):
  ```bash
  go run main.go db-truncate
  # or
  ./spendlimit db-truncate
  ```
- **Drop All Tables** (Remove all tables from database):
  ```bash
  go run main.go db-drop-tables
  # or
  ./spendlimit db-drop-tables
  ```
- **Apply Eligible Pending Requests**:
  ```bash
  go run main.go apply-pending-changes
  ```
- **Time-Travel for Testing**:
  ```bash
  go run main.go time-travel -client-id <UUID> -period HOURS_24 -field effectiveAt
  ```
