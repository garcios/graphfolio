# Database Layer Implementation Plan (PostgreSQL)

> Branch: `feat/db-layer`
> Status: **Approved decisions incorporated** — see [§10 Decision Log](#10-decision-log).

## 1. Goals

- Add PostgreSQL persistence for the gRPC microservices (`services/portfolio-api`, `services/user-api`).
- Model the domain well enough to compute **total annualized returns**, including **dividends**, **currency fluctuations**, and **corporate actions** (see `AGENTS.md` §1).
- Support **two cost-basis methods** per portfolio: **Average Cost (default)** and **FIFO**.
- Switch money and percentages from `double`/`Float` to an **exact decimal type** end to end (DB → gRPC → GraphQL → web).
- Respect the layered architecture. Each service owns its own schema. The BFF and web never touch the DB. Shared connection code lives in `pkg/database`.

### Non-Goals (for this iteration)

- Multiple brokerage accounts per portfolio (a portfolio is the single unit of ownership).
- Market-data ingestion jobs (price/FX feeds). The tables are created; populating them is out of scope.
- Authentication/authorization (`user-api` gets only a minimal `users` table).
- Production HA, backups, and partitioning.
- Containerization. Local development uses a **natively installed PostgreSQL at `localhost:5432`** (no Docker).

## 2. Key Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Engine | **PostgreSQL 18**, installed locally, listening on `localhost:5432` | Native `uuidv7()` (time-ordered PKs), mature `NUMERIC`, partitioning available later. No container runtime needed. |
| Topology | **One database `graphfolio`, one schema per service**: `portfolio` (portfolio-api), `users` (user-api) | Simple ops and ad-hoc querying. Isolation comes from per-service roles that only have privileges on their own schema. (`user` is a reserved word, hence `users`.) |
| Cross-schema rules | **No cross-schema FKs, joins, or grants** | Keeps the services separable. A schema could move to its own DB later without code changes. |
| Primary keys | `uuid DEFAULT uuidv7()` | Globally unique across services, index-friendly ordering, safe to expose through the API. |
| Money & quantities | `NUMERIC`, **never** `float` | Amounts `NUMERIC(20,6)`, quantities `NUMERIC(28,10)` (fractional shares), prices `NUMERIC(20,8)`, FX `NUMERIC(20,10)`. |
| Timestamps | `timestamptz` for events, `date` for trade/price/ex dates | Trade and valuation semantics are date-based. Audit columns are instants. |
| Source of truth | **Transaction ledger** (append-only) | Lots, holdings, cash, and valuations are *derived projections*, rebuildable from `transactions` + `corporate_actions`. |
| Cost basis | Per-portfolio `cost_basis_method` enum, default `AVERAGE_COST`; `FIFO` supported | Both methods share one `tax_lots` structure (see §4.7), so switching method is just a projection rebuild. |
| Enums | Postgres `ENUM` types, created inside each service's schema | Type-safe and compact. New values can be added with `ALTER TYPE ... ADD VALUE`. |
| Migrations | **golang-migrate**, plain SQL `up`/`down`, one migration set and one `schema_migrations` table per schema | SQL-first, with a CLI for the Makefile and a library for the optional run-on-startup. |
| Driver / query layer | **pgx/v5** (`pgxpool`) with **hand-written SQL** in each service's `internal/repository` | Full control over SQL, no codegen step. `NUMERIC` ↔ `decimal.Decimal` via `pgx-shopspring-decimal`. |
| Go decimal type | `github.com/shopspring/decimal` | De-facto standard, with pgx integration and JSON/string round-tripping. |

## 3. Entity-Relationship Overview

### Schema `users` (owned by `user-api`)

```mermaid
erDiagram
    users {
        uuid id PK
        text email UK
        text display_name
        char3 base_currency
        timestamptz created_at
        timestamptz updated_at
    }
```

### Schema `portfolio` (owned by `portfolio-api`)

```mermaid
erDiagram
    currencies ||--o{ instruments : "denominated in"
    exchanges ||--o{ instruments : "listed on"
    currencies ||--o{ portfolios : "base currency"
    portfolios ||--o{ transactions : has
    instruments ||--o{ transactions : "traded in"
    instruments ||--o{ corporate_actions : "subject to"
    instruments ||--o{ instrument_prices : "priced daily"
    currencies ||--o{ fx_rates : "base/quote"
    portfolios ||--o{ tax_lots : "projection"
    transactions ||--o{ tax_lots : "opens"
    tax_lots ||--o{ lot_disposals : "consumed by"
    transactions ||--o{ lot_disposals : "SELL"
    portfolios ||--o{ holdings : "projection"
    portfolios ||--o{ cash_balances : "projection"
    portfolios ||--o{ portfolio_valuations : "daily snapshot"

    portfolios {
        uuid id PK
        uuid user_id "logical ref to users schema (no FK)"
        text name
        char3 base_currency FK
        cost_basis_method cost_basis_method "default AVERAGE_COST"
    }
    transactions {
        uuid id PK
        uuid portfolio_id FK
        uuid instrument_id FK "nullable"
        transaction_type type
        date trade_date
        numeric quantity
        numeric price
        numeric amount
    }
    tax_lots {
        uuid id PK
        uuid open_transaction_id FK
        date acquired_date
        numeric remaining_quantity
        numeric cost_basis
    }
    lot_disposals {
        uuid id PK
        uuid tax_lot_id FK
        uuid sell_transaction_id FK
        numeric quantity
        numeric cost_basis_released
    }
```

> [!NOTE]
> `portfolio.portfolios.user_id` deliberately has **no foreign key** to `users.users`. Even though both schemas share one database, cross-schema references are forbidden by convention and by privileges (`portfolio_svc` has no access to the `users` schema).

## 4. Table Definitions

All DDL is **schema-qualified** so that it doesn't depend on `search_path`.

### 4.0 Local PostgreSQL prerequisite

PostgreSQL **18+** must be installed natively and listening on **`localhost:5432`**. Version 18 is required for `uuidv7()`. On macOS:

```bash
brew install postgresql@18
brew services start postgresql@18
pg_isready -h localhost -p 5432     # => accepting connections
```

The bootstrap and teardown scripts connect as a **local superuser**. For Homebrew installs that's your macOS user. The Makefile variable `PG_ADMIN_URL` (default `postgres://$(USER)@localhost:5432/postgres`) can be overridden, e.g. `PG_ADMIN_URL=postgres://postgres@localhost:5432/postgres`.

### 4.1 Bootstrap — `scripts/db/bootstrap.sql` (run once with `make db-bootstrap`)

The script is **idempotent** and **parameterized by database name** (`-v db=...`). The same script therefore creates both the dev database (`graphfolio`) and the integration-test database (`graphfolio_test`).

```sql
\set ON_ERROR_STOP on
\if :{?db} \else \set db graphfolio \endif

-- Fail fast on an unsupported server version (uuidv7() needs PG 18).
DO $$ BEGIN
  IF current_setting('server_version_num')::int < 180000 THEN
    RAISE EXCEPTION 'PostgreSQL 18+ required, found %', current_setting('server_version');
  END IF;
END $$;

-- Roles are cluster-wide: create only if missing.
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'portfolio_svc') THEN
    CREATE ROLE portfolio_svc LOGIN PASSWORD 'portfolio';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user_svc') THEN
    CREATE ROLE user_svc LOGIN PASSWORD 'user';
  END IF;
END $$;

-- CREATE DATABASE can't run inside DO/transactions, so use \gexec.
SELECT format('CREATE DATABASE %I', :'db')
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'db') \gexec

\connect :"db"

REVOKE ALL ON DATABASE :"db" FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db" TO portfolio_svc, user_svc;

CREATE SCHEMA IF NOT EXISTS portfolio AUTHORIZATION portfolio_svc;
CREATE SCHEMA IF NOT EXISTS users     AUTHORIZATION user_svc;

ALTER ROLE portfolio_svc IN DATABASE :"db" SET search_path = portfolio;
ALTER ROLE user_svc      IN DATABASE :"db" SET search_path = users;
```

The companion `scripts/db/teardown.sql` runs `DROP DATABASE IF EXISTS :"db" WITH (FORCE);`. It drops the roles only when invoked with `-v drop_roles=1`, because roles are shared across every database in the local cluster.

Schemas and roles are infrastructure, so they are created **here and not in migrations**. golang-migrate stores its `schema_migrations` table *inside* the target schema, so the schema must exist before the first migration runs.

> [!NOTE]
> The role passwords are for local development only. On a default Homebrew install, `pg_hba.conf` trusts local connections. Either way, the services connect with these roles and never as the superuser.

### 4.2 `users` — `services/user-api/migrations/000001_create_users.up.sql`

```sql
CREATE TABLE users.users (
    id             uuid        PRIMARY KEY DEFAULT uuidv7(),
    email          text        NOT NULL,
    display_name   text        NOT NULL,
    base_currency  char(3)     NOT NULL DEFAULT 'USD' CHECK (base_currency ~ '^[A-Z]{3}$'),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_lower_uq ON users.users (lower(email));
```

### 4.3 `portfolio` — reference data (`000001_reference_data`)

```sql
CREATE TYPE portfolio.asset_class AS ENUM ('EQUITY', 'ETF', 'FUND', 'BOND', 'CRYPTO', 'CASH_EQUIVALENT');

CREATE TABLE portfolio.currencies (
    code         char(3)  PRIMARY KEY,          -- ISO 4217
    name         text     NOT NULL,
    minor_units  smallint NOT NULL DEFAULT 2
);

CREATE TABLE portfolio.exchanges (
    code      text    PRIMARY KEY,              -- ISO 10383 MIC, e.g. XNAS, XASX
    name      text    NOT NULL,
    country   char(2) NOT NULL,                 -- ISO 3166-1 alpha-2
    timezone  text    NOT NULL                  -- IANA tz, e.g. America/New_York
);

CREATE TABLE portfolio.instruments (
    id             uuid                  PRIMARY KEY DEFAULT uuidv7(),
    symbol         text                  NOT NULL,     -- maps to Investment.ticker
    exchange_code  text                  NOT NULL REFERENCES portfolio.exchanges(code),
    name           text                  NOT NULL,
    asset_class    portfolio.asset_class NOT NULL,
    currency_code  char(3)               NOT NULL REFERENCES portfolio.currencies(code),
    isin           char(12)              UNIQUE,
    is_active      boolean               NOT NULL DEFAULT true,
    created_at     timestamptz           NOT NULL DEFAULT now(),
    updated_at     timestamptz           NOT NULL DEFAULT now(),
    UNIQUE (symbol, exchange_code)
);
```

### 4.4 Core ledger (`000002_portfolios_and_transactions`)

```sql
CREATE TYPE portfolio.cost_basis_method AS ENUM ('AVERAGE_COST', 'FIFO');

CREATE TYPE portfolio.transaction_type AS ENUM (
    'BUY', 'SELL',                       -- instrument trades
    'DIVIDEND', 'INTEREST',              -- income
    'DEPOSIT', 'WITHDRAWAL',             -- external cash flows (matter for TWR)
    'FEE', 'TAX',                        -- standalone charges
    'TRANSFER_IN', 'TRANSFER_OUT',       -- in-kind security transfers
    'FX_CONVERSION'                      -- cash moved between currencies
);

CREATE TABLE portfolio.portfolios (
    id                 uuid                        PRIMARY KEY DEFAULT uuidv7(),
    user_id            uuid                        NOT NULL,     -- users.users.id; no FK
    name               text                        NOT NULL,
    base_currency      char(3)                     NOT NULL REFERENCES portfolio.currencies(code),
    cost_basis_method  portfolio.cost_basis_method NOT NULL DEFAULT 'AVERAGE_COST',
    created_at         timestamptz                 NOT NULL DEFAULT now(),
    updated_at         timestamptz                 NOT NULL DEFAULT now(),
    archived_at        timestamptz,
    UNIQUE (user_id, name)
);
CREATE INDEX portfolios_user_id_idx ON portfolio.portfolios (user_id) WHERE archived_at IS NULL;

CREATE TABLE portfolio.transactions (
    id               uuid                       PRIMARY KEY DEFAULT uuidv7(),
    portfolio_id     uuid                       NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    instrument_id    uuid                       REFERENCES portfolio.instruments(id),
    type             portfolio.transaction_type NOT NULL,
    trade_date       date                       NOT NULL,
    settle_date      date,
    quantity         numeric(28,10),             -- shares/units (positive)
    price            numeric(20,8),              -- per unit, in currency_code
    amount           numeric(20,6)              NOT NULL,  -- signed net cash impact in currency_code
    currency_code    char(3)                    NOT NULL REFERENCES portfolio.currencies(code),
    fee              numeric(20,6)              NOT NULL DEFAULT 0,
    withholding_tax  numeric(20,6)              NOT NULL DEFAULT 0,  -- e.g. on dividends
    fx_rate_to_base  numeric(20,10),             -- rate at trade time; NULL = look up fx_rates
    external_ref     text,                       -- broker import id (idempotency)
    notes            text,
    created_at       timestamptz                NOT NULL DEFAULT now(),

    CONSTRAINT instrument_required CHECK (
        type NOT IN ('BUY','SELL','DIVIDEND','TRANSFER_IN','TRANSFER_OUT')
        OR instrument_id IS NOT NULL
    ),
    CONSTRAINT trade_fields_required CHECK (
        type NOT IN ('BUY','SELL','TRANSFER_IN','TRANSFER_OUT') OR (quantity > 0 AND price >= 0)
    ),
    CONSTRAINT non_negative_charges CHECK (fee >= 0 AND withholding_tax >= 0),
    UNIQUE (portfolio_id, external_ref)
);
CREATE INDEX transactions_portfolio_date_idx       ON portfolio.transactions (portfolio_id, trade_date, id);
CREATE INDEX transactions_portfolio_instrument_idx ON portfolio.transactions (portfolio_id, instrument_id, trade_date, id);
```

> [!NOTE]
> `(trade_date, id)` gives a **deterministic replay order**. Because `uuidv7` is time-ordered, transactions on the same day are processed in insertion order. This matters for FIFO lot matching.

### 4.5 Corporate actions (`000003_corporate_actions`)

Corporate actions are **instrument-level facts** shared by all portfolios. They are applied when the projections are rebuilt and are never copied into each user's ledger.

```sql
CREATE TYPE portfolio.corporate_action_type AS ENUM (
    'SPLIT', 'REVERSE_SPLIT', 'SPIN_OFF', 'MERGER', 'SYMBOL_CHANGE', 'RETURN_OF_CAPITAL'
);

CREATE TABLE portfolio.corporate_actions (
    id                 uuid                            PRIMARY KEY DEFAULT uuidv7(),
    instrument_id      uuid                            NOT NULL REFERENCES portfolio.instruments(id),
    type               portfolio.corporate_action_type NOT NULL,
    ex_date            date                            NOT NULL,
    ratio_from         numeric(20,10),        -- e.g. 1 (old shares)
    ratio_to           numeric(20,10),        -- e.g. 4 (new shares) => 4:1 split
    new_instrument_id  uuid                            REFERENCES portfolio.instruments(id),
    cash_amount        numeric(20,6),         -- per share, for cash components / RoC
    currency_code      char(3)                         REFERENCES portfolio.currencies(code),
    cost_basis_pct     numeric(9,6),          -- share of cost basis moved to new instrument
    notes              text,
    created_at         timestamptz                     NOT NULL DEFAULT now(),
    UNIQUE (instrument_id, type, ex_date),
    CONSTRAINT ratio_positive CHECK (
        type NOT IN ('SPLIT','REVERSE_SPLIT') OR (ratio_from > 0 AND ratio_to > 0)
    )
);
```

### 4.6 Market data (`000004_market_data`)

```sql
CREATE TABLE portfolio.instrument_prices (
    instrument_id  uuid          NOT NULL REFERENCES portfolio.instruments(id),
    price_date     date          NOT NULL,
    close          numeric(20,8) NOT NULL CHECK (close >= 0),  -- unadjusted close
    source         text          NOT NULL DEFAULT 'manual',
    created_at     timestamptz   NOT NULL DEFAULT now(),
    PRIMARY KEY (instrument_id, price_date)
);

CREATE TABLE portfolio.fx_rates (
    base_currency   char(3)        NOT NULL REFERENCES portfolio.currencies(code),
    quote_currency  char(3)        NOT NULL REFERENCES portfolio.currencies(code),
    rate_date       date           NOT NULL,
    rate            numeric(20,10) NOT NULL CHECK (rate > 0),  -- 1 base = rate quote
    source          text           NOT NULL DEFAULT 'manual',
    PRIMARY KEY (base_currency, quote_currency, rate_date),
    CHECK (base_currency <> quote_currency)
);
```

> [!TIP]
> Prices are stored **unadjusted**. Split adjustment is computed from `corporate_actions`, which avoids rewriting history whenever a split occurs.

### 4.7 Cost basis: tax lots (`000005_tax_lots`)

Lots are recorded for **every** portfolio, whatever its method. Only the **disposal-matching rule** differs:

| Method | On SELL / TRANSFER_OUT of `s` units |
|---|---|
| `AVERAGE_COST` (default) | Release units from **every open lot pro-rata** by remaining quantity. Lot *i* releases `s × qᵢ/Q` units at its own per-unit cost. Summed over lots, this equals exactly `s × (C/Q)`, the average cost. |
| `FIFO` | Consume the oldest open lots first, ordered by `(acquired_date, id)`, until `s` units are matched. |

Because both methods use the same tables, **changing a portfolio's method is just a projection rebuild**. The ledger is untouched. Rounding residuals from pro-rata allocation are assigned to the last lot, so totals reconcile to the cent.

```sql
CREATE TABLE portfolio.tax_lots (
    id                    uuid           PRIMARY KEY DEFAULT uuidv7(),
    portfolio_id          uuid           NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    instrument_id         uuid           NOT NULL REFERENCES portfolio.instruments(id),
    open_transaction_id   uuid           NOT NULL REFERENCES portfolio.transactions(id) ON DELETE CASCADE,
    acquired_date         date           NOT NULL,
    original_quantity     numeric(28,10) NOT NULL CHECK (original_quantity > 0),   -- split-adjusted
    remaining_quantity    numeric(28,10) NOT NULL CHECK (remaining_quantity >= 0),
    cost_basis            numeric(20,6)  NOT NULL,   -- remaining cost, instrument currency (incl. fees)
    cost_basis_base       numeric(20,6)  NOT NULL,   -- remaining cost, base currency at historical FX
    closed_date           date,
    CHECK (remaining_quantity <= original_quantity)
);
CREATE INDEX tax_lots_open_idx
    ON portfolio.tax_lots (portfolio_id, instrument_id, acquired_date, id)
    WHERE remaining_quantity > 0;

CREATE TABLE portfolio.lot_disposals (
    id                        uuid           PRIMARY KEY DEFAULT uuidv7(),
    tax_lot_id                uuid           NOT NULL REFERENCES portfolio.tax_lots(id) ON DELETE CASCADE,
    sell_transaction_id       uuid           NOT NULL REFERENCES portfolio.transactions(id) ON DELETE CASCADE,
    quantity                  numeric(28,10) NOT NULL CHECK (quantity > 0),
    cost_basis_released       numeric(20,6)  NOT NULL,
    cost_basis_released_base  numeric(20,6)  NOT NULL,
    proceeds_base             numeric(20,6)  NOT NULL,   -- net of fees, allocated pro-rata
    realized_pnl_base         numeric(20,6)  NOT NULL,   -- proceeds_base - cost_basis_released_base
    UNIQUE (tax_lot_id, sell_transaction_id)
);
CREATE INDEX lot_disposals_sell_idx ON portfolio.lot_disposals (sell_transaction_id);
```

Corporate actions applied to lots during a rebuild:
- **Split / reverse split**: multiply `original_quantity` and `remaining_quantity` by `ratio_to/ratio_from`. Cost stays the same.
- **Spin-off**: move `cost_basis_pct` of each lot's cost into a new lot on `new_instrument_id`, keeping the original `acquired_date`.
- **Return of capital**: reduce lot cost by `cash_amount × remaining_quantity`, floored at 0.

### 4.8 Projections & analytics (`000006_projections`)

These tables are **rebuildable caches** that keep `GetPortfolio` to a single indexed read. `holdings.cost_basis` always equals the sum of the instrument's open `tax_lots`.

```sql
CREATE TABLE portfolio.holdings (
    portfolio_id         uuid           NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    instrument_id        uuid           NOT NULL REFERENCES portfolio.instruments(id),
    quantity             numeric(28,10) NOT NULL,
    cost_basis           numeric(20,6)  NOT NULL,   -- Σ open lots, instrument currency
    cost_basis_base      numeric(20,6)  NOT NULL,   -- Σ open lots, base currency
    realized_pnl_base    numeric(20,6)  NOT NULL DEFAULT 0,   -- Σ lot_disposals
    dividends_base       numeric(20,6)  NOT NULL DEFAULT 0,   -- net of withholding tax
    last_transaction_id  uuid           REFERENCES portfolio.transactions(id) ON DELETE SET NULL,
    updated_at           timestamptz    NOT NULL DEFAULT now(),
    PRIMARY KEY (portfolio_id, instrument_id)
);

CREATE TABLE portfolio.cash_balances (
    portfolio_id   uuid          NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    currency_code  char(3)       NOT NULL REFERENCES portfolio.currencies(code),
    balance        numeric(20,6) NOT NULL,
    updated_at     timestamptz   NOT NULL DEFAULT now(),
    PRIMARY KEY (portfolio_id, currency_code)
);

-- Daily snapshot for Time-Weighted Return (TWR) / annualized return.
CREATE TABLE portfolio.portfolio_valuations (
    portfolio_id        uuid           NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    valuation_date      date           NOT NULL,
    market_value_base   numeric(20,6)  NOT NULL,   -- holdings only
    cash_value_base     numeric(20,6)  NOT NULL,
    net_flow_base       numeric(20,6)  NOT NULL DEFAULT 0,  -- deposits - withdrawals that day
    daily_return        numeric(20,12),             -- sub-period return
    twr_index           numeric(20,12) NOT NULL,    -- cumulative growth of 1.0
    PRIMARY KEY (portfolio_id, valuation_date)
);
```

## 5. Decimal Money in the API

The switch from `double` happens in this branch. The contract is pre-release and every consumer lives in this monorepo, so the fields are changed **in place** in `portfolio.v1` and deployed together.

### 5.1 Proto — new `proto/common/v1/decimal.proto`

```proto
syntax = "proto3";
package common.v1;
option go_package = "graphfolio/proto/common/v1;commonpb";

// Arbitrary-precision decimal encoded as a canonical string, e.g. "-1234.56".
message Decimal { string value = 1; }

message Money {
  Decimal amount        = 1;
  string  currency_code = 2;   // ISO 4217
}
```

Changes to `portfolio.proto`: monetary fields such as `total_value`, `cash_balance`, `price` and `*_return_amount` become `common.v1.Money`. `quantity` and the `*_percent` fields become `common.v1.Decimal`. Percentages keep their current meaning (`14.2` means 14.2 %).

### 5.2 Go conversion helpers — `pkg/decimalpb`

Domain-agnostic helpers, used by both the services and the BFF:

```go
func ToProto(d decimal.Decimal) *commonpb.Decimal
func FromProto(p *commonpb.Decimal) (decimal.Decimal, error)
func MoneyToProto(d decimal.Decimal, ccy string) *commonpb.Money
```

### 5.3 GraphQL (BFF)

```graphql
"Arbitrary-precision decimal serialized as a string, e.g. \"1234.56\"."
scalar Decimal

type Money {
  amount: Decimal!
  currencyCode: String!
}

type Portfolio {
  totalValue: Money!
  todayReturnAmount: Money!
  todayReturnPercent: Decimal!
  annualizedReturnPercent: Decimal!
  cashBalance: Money!
  investments: [Investment!]!
}
```

- `gqlgen.yml` maps `Decimal` to `github.com/shopspring/decimal.Decimal`, with a custom marshaller in `bff/graph/scalars/decimal.go`. It **serializes to a JSON string**, never a JSON number.
- The resolvers pass the proto string straight through, with no float conversion.

### 5.4 Web

- `make generate` regenerates the genql client, where `Decimal` becomes `string`.
- Formatting uses `Intl.NumberFormat(...).format(value)` with the **string** (supported in modern browsers), so the value is formatted exactly.
- The dashboard does no client-side arithmetic. If that's ever needed, use `big.js`.

## 6. Mapping to the API Contract

| API field | Source |
|---|---|
| `Portfolio.totalValue` | Σ(`holdings.quantity` × latest `close` × FX→base) + `cashBalance` |
| `Portfolio.cashBalance` | Σ(`cash_balances.balance` × FX→base) |
| `Portfolio.todayReturn*` | Latest vs. previous `instrument_prices.close` (and FX), weighted by quantity |
| `Portfolio.annualizedReturnPercent` | `(twr_index^(365/days) − 1) × 100` from `portfolio_valuations` |
| `Investment.id` | `instruments.id` (UUID string) |
| `Investment.ticker` / `name` | `instruments.symbol` / `instruments.name` |
| `Investment.price` | Latest `instrument_prices.close`, in the instrument's currency |
| `Investment.quantity` | `holdings.quantity` |
| `Investment.totalReturn*` | market value − `cost_basis_base` + `dividends_base` + `realized_pnl_base`. The cost basis depends on the portfolio's method |

## 7. Repository Layout Changes

```text
├── scripts/db/
│   ├── bootstrap.sql                          # NEW: idempotent roles + DB + schemas (§4.1)
│   └── teardown.sql                           # NEW: drop DB (optionally roles)
├── .env.example                               # NEW: PORTFOLIO_DB_URL, USER_DB_URL
├── proto/common/v1/decimal.proto              # NEW: Decimal, Money (§5.1)
├── pkg/
│   ├── database/
│   │   ├── config.go                          # NEW: env-driven config (DSN, pool sizes, timeouts)
│   │   ├── postgres.go                        # NEW: NewPool(); registers shopspring decimal in AfterConnect
│   │   └── migrate.go                         # NEW: RunMigrations(fs embed.FS, dsn)
│   └── decimalpb/                             # NEW: decimal <-> proto helpers (§5.2)
├── services/portfolio-api/
│   ├── migrations/                            # NEW: 000001..000006 *.up.sql / *.down.sql + embed.go
│   ├── seeds/dev_seed.sql                     # NEW: AAPL, MSFT, TSLA, NVDA, V (matches current mock)
│   └── internal/repository/                   # NEW (follow-up): hand-written pgx queries
├── services/user-api/
│   └── migrations/000001_create_users.*.sql   # NEW
└── bff/graph/scalars/decimal.go               # NEW: Decimal scalar marshaller (§5.3)
```

### Connection strings (`.env.example`)

Each service connects with **its own role** to the local server at `localhost:5432`. `search_path` and the migrations table keep each service's migration history inside its own schema:

```dotenv
PG_ADMIN_URL=postgres://<your-mac-user>@localhost:5432/postgres   # superuser, bootstrap/teardown only
PORTFOLIO_DB_URL=postgres://portfolio_svc:portfolio@localhost:5432/graphfolio?sslmode=disable&search_path=portfolio&x-migrations-table=schema_migrations
USER_DB_URL=postgres://user_svc:user@localhost:5432/graphfolio?sslmode=disable&search_path=users&x-migrations-table=schema_migrations
```

> [!IMPORTANT]
> `x-migrations-table` is a golang-migrate parameter. `pkg/database.NewPool` must **strip `x-*` parameters** before handing the DSN to pgx.

### Repository conventions (hand-written pgx)

- One repository struct per aggregate (`PortfolioRepo`, `TransactionRepo`, `LotRepo`, ...), constructed with a `*pgxpool.Pool`.
- SQL is kept as `const` strings next to the method that uses it, always schema-qualified, always with `$n` placeholders (never string concatenation).
- Scan with `pgx.CollectRows` + `pgx.RowToStructByName`.
- Multi-statement writes (insert transaction → rebuild lots/holdings) go through `pgx.BeginFunc`.
- Projection rebuilds take `SELECT ... FROM portfolio.portfolios WHERE id = $1 FOR UPDATE` to serialize concurrent rebuilds for the same portfolio.

### Makefile additions

```make
PG_ADMIN_URL ?= postgres://$(USER)@localhost:5432/postgres
DB_NAME      ?= graphfolio

db-check:         ## Verify local Postgres is running on localhost:5432
	pg_isready -h localhost -p 5432
db-bootstrap: db-check   ## Create roles, database and schemas (idempotent)
	psql "$(PG_ADMIN_URL)" -v db=$(DB_NAME) -f scripts/db/bootstrap.sql
db-drop: db-check        ## Drop the database (keeps cluster-wide roles)
	psql "$(PG_ADMIN_URL)" -v db=$(DB_NAME) -f scripts/db/teardown.sql
migrate-up:       ## Apply all migrations for every service
	migrate -path services/portfolio-api/migrations -database "$(PORTFOLIO_DB_URL)" up
	migrate -path services/user-api/migrations      -database "$(USER_DB_URL)" up
migrate-down:     ## Roll back one step per service
	migrate -path services/portfolio-api/migrations -database "$(PORTFOLIO_DB_URL)" down 1
	migrate -path services/user-api/migrations      -database "$(USER_DB_URL)" down 1
migrate-create:   ## make migrate-create svc=portfolio-api name=add_x
	migrate create -ext sql -dir services/$(svc)/migrations -seq $(name)
db-seed:
	psql "$(PORTFOLIO_DB_URL_PSQL)" -f services/portfolio-api/seeds/dev_seed.sql
db-reset: db-drop db-bootstrap migrate-up db-seed
db-test-setup:    ## Fresh graphfolio_test DB for integration tests
	$(MAKE) db-drop db-bootstrap DB_NAME=graphfolio_test
```

The `golang-migrate` CLI version is pinned in `tools/tools.go`, as `AGENTS.md` §3 describes.

## 8. Implementation Phases

| # | Phase | Deliverables | Done when |
|---|---|---|---|
| 1 | **Infrastructure** | `scripts/db/bootstrap.sql`, `scripts/db/teardown.sql`, `.env.example`, Makefile `db-*` targets | Against the local server, `make db-bootstrap` succeeds twice in a row (idempotent), and both schemas and roles exist |
| 2 | **Shared packages** | `pkg/database` (pool, decimal registration, DSN sanitizing, migration runner) and `pkg/decimalpb` | Unit tests pass, and the services/BFF import them via `go.work` |
| 3 | **users schema** | `000001_create_users` up/down | `migrate up` → `down` → `up` round-trips cleanly |
| 4 | **portfolio schema** | Migrations `000001`–`000006` (§4.3–§4.8), each with a matching `down` | Round-trip succeeds, and `\dt portfolio.*` matches the plan |
| 5 | **Seed data** | `dev_seed.sql`: currencies (USD, AUD, EUR, GBP), exchanges (XNAS, XNYS, XASX), 5 instruments, one `AVERAGE_COST` and one `FIFO` demo portfolio, multi-lot BUYs, a partial SELL, prices, FX, a dividend and a split | The seeded data covers both methods, and the expected results are documented in the seed file |
| 6 | **Decimal API contract** | `decimal.proto`, updated `portfolio.proto`, regenerated stubs, mock server updated to `Money`/`Decimal`, `Decimal` scalar + `Money` type in the BFF, regenerated genql, web formatting | `make generate && make run` works, and the dashboard renders the same values as before |
| 7 | **Service wiring** | `portfolio-api` and `user-api` `main.go` read their DB URL, open the pool, and optionally run migrations (`MIGRATE_ON_START=true`) | The service starts, fails fast on a bad DSN, and closes the pool on shutdown |
| 8 | **Docs** | Update `README.md` (prerequisites: local PostgreSQL 18 on `localhost:5432`, `migrate`, `psql`; first-time `make db-bootstrap`) and the `AGENTS.md` structure tree | Docs reviewed |

> Follow-up PR (out of scope here): `internal/repository` with pgx queries, the projection rebuild engine (ledger → lots → holdings/cash, both cost-basis methods), the TWR valuation job, and replacing the mock in `server.go` with DB reads.

## 9. Verification Plan

1. **Migration integrity**: `make db-test-setup`, then `migrate-up`, roll everything back, and `migrate-up` again, all against `graphfolio_test` on the local server. There must be no drift, and both `schema_migrations` tables must sit in their own schemas.
2. **Constraint tests**: insert invalid rows (BUY without instrument, negative fee, zero split ratio, duplicate `external_ref`, `remaining_quantity > original_quantity`) and assert they're rejected.
3. **Schema isolation**: as `portfolio_svc`, `SELECT * FROM users.users` must fail with *permission denied*, and vice versa. `CREATE TABLE public.x` must fail for both roles.
4. **Cost-basis fixtures** (implemented with the follow-up rebuild engine; the fixtures are defined now in the seed): buy 10 @ 100, buy 10 @ 200, sell 10 @ 250.
   - `AVERAGE_COST`: released cost 1,500, realized P&L 1,000, remaining cost 1,500.
   - `FIFO`: released cost 1,000, realized P&L 1,500, remaining cost 2,000.
5. **Decimal round-trip**: a value such as `0.1 + 0.2` stored in the DB comes back from GraphQL as the exact string `"0.300000"`, not `0.30000000000000004`. Covered by BFF scalar unit tests and `pkg/decimalpb` tests.
6. **Integration tests**: Go tests read `TEST_PORTFOLIO_DB_URL` / `TEST_USER_DB_URL`, which point to `graphfolio_test` on `localhost:5432`. They run migrations and the seed, then assert row counts. Each test runs inside a transaction that is rolled back. Tests call `t.Skip` when the variables are unset or the server is unreachable, so `go test ./...` still passes on machines without Postgres. In CI, PostgreSQL 18 is installed natively on the runner (e.g. the PGDG apt package), not as a container.
7. **Smoke test**: `make db-reset && make run`, and the dashboard renders with the decimal-based API. The dev database `graphfolio` is never touched by the tests.

## 10. Decision Log

| # | Question | Decision |
|---|---|---|
| 1 | Database topology | **Single database `graphfolio`, one schema per service** (`portfolio`, `users`), isolated by per-service roles |
| 2 | Query layer | **Hand-written SQL with pgx/v5** |
| 3 | Migration tool | **golang-migrate** |
| 4 | Multiple brokerage accounts | **No.** Not in scope; a portfolio is the unit of ownership |
| 5 | Cost-basis method | **Both supported**, `AVERAGE_COST` default, `FIFO` optional per portfolio, via shared `tax_lots` + `lot_disposals` |
| 6 | Money in the API | **Switch to decimal now**: proto `Decimal`/`Money`, GraphQL `Decimal` scalar (string) + `Money` type |
| 7 | Database runtime | **Locally installed PostgreSQL 18 at `localhost:5432`**, no Docker. Bootstrapped by an idempotent `scripts/db/bootstrap.sql` |
