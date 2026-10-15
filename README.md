# GraphFolio

A modern portfolio tracker built for serious investors. GraphFolio accurately measures your total annualized returns—including dividends, currency fluctuations, and corporate actions—all in one clear, consolidated dashboard.

---

## Key Features

- **Consolidated Dashboard**: View all your investments across multiple asset classes in a single, unified view.
- **Exact Decimal Arithmetic**: Zero floating-point drift. All money, quantities, prices, and rates use fixed-point decimal arithmetic from database to browser.
- **Interactive Performance Chart**: Timeframe-selectable SVG performance chart (1D, 1W, 1M, YTD, 1Y, ALL) rendering cubic Bezier curves, dynamic profit/loss color gradients, crosshair tracking, inspection tooltips, and exact period returns.
- **Interactive Transaction Logging & Multi-Currency Fees**: Record Buys, Sells, Cash Deposits, Withdrawals, Dividends, and Stock Splits directly in the UI (`AddTransactionModal`). Investors can specify custom currencies for brokerage fees (defaulting to the investor's preferred display currency, e.g. paying AUD brokerage for USD share purchases), with exact multi-currency cash balance deduction and tax lot cost basis calibration. Holding quantities, cost bases, cash balances, and returns re-project deterministically in real time.
- **Corporate Action Stock Splits & Reverse Splits**: Native handling of forward and reverse stock splits (`SPLIT`). Features ratio presets (`2:1`, `3:1`, `4:1`, `10:1`, `1:2 Reverse`, `1:10 Reverse`) or custom ratios with live multiplier calculation. Preserves overall holding cost basis and cash balances while proportionally scaling share quantities across all open tax lots ($Q_{\text{new}} = Q_{\text{old}} \times \text{multiplier}$), conforming to US GAAP, IFRS, and GIPS standards.
- **Transaction History & Ledger Management**: Paginated transaction history with type filtering, exact execution prices, fee tracking, and safe deletion with immediate atomic ledger replay.
- **Transaction-Ledger Architecture**: Immutable transaction ledger acts as the authoritative source of truth, deterministically projecting holdings, cash balances, and valuations.
- **Flexible Cost Basis Accounting**: Native support for both **Average Cost** (`AVERAGE_COST`, default) and **FIFO** (`FIFO`) tax lot relief strategies.
- **Daily Portfolio Valuation Engine & GIPS Annualized Returns**: Automated daily End-of-Day (EOD) portfolio valuation engine and scheduled background worker (`make run-valuation-job`). Evaluates daily market values, net cash flows, sub-period returns, and cumulative Time-Weighted Return (TWR) indexes with zero float drift. Features true GIPS-compliant Compound Annual Growth Rate (CAGR) annualization ($(\text{TWR}^{\frac{365.25}{D}} - 1) \times 100$ for periods $\ge 365$ days), US GAAP ASC 820 Level 2/3 cost-basis fallback for unquoted or historically backfilled assets (eliminating phantom write-offs and false return drops on inflow dates), multi-day historical price/FX matrix replay with Last Observation Carried Forward (LOCF), and automated backfill triggers on past-dated trade logging and deletions.
- **Foreign Exchange (FX) & Currency Pair Management**: Dedicated operations view in the Admin Portal (`:5174`). Inspect tracked currency pairs (`EUR/USD`, `USD/GBP`, `USD/AUD`, `USD/JPY`, `EUR/GBP`, `USD/CAD`), toggle direct and reciprocal quotes (`Base ⇄ Quote`) with exact 10-decimal precision, explore interactive SVG historical trend curves across timeframes (`1W`, `1M`, `1Y`, `ALL`) with glowing area fills and hover crosshairs, browse the authoritative rates ledger, and apply audit-justified manual exchange rate overrides.
- **Master Asset Directory & Exchange Reference Data**: Authoritative instrument catalog management in the Admin Portal. Operators register tradable assets with symbol, exchange selection powered by live `portfolio.exchanges` reference data (`ListExchanges` gRPC and `exchanges` GraphQL query), asset class categorization, currency, and ISO 6166 ISIN validation.
- **Authoritative Closing Price Ledger & Date Range Filtering**: Dedicated pricing ledger in the internal Admin Portal (`:5174`). Filter closing marks across custom date ranges with dual date pickers and one-click preset buttons (`7D`, `30D`, `90D`, `YTD`, `1Y`), client-side date validation guards, multi-page pagination controls (`50` records per page), manual audit-justified price overrides with retroactive valuation recalibrations, and on-demand EOD market sync and backfill triggers.
- **Asset Deletion & Historical Price Cascading**: Permanently remove obsolete or erroneous assets from the master directory in the Admin Portal (`:5174`). Automatically cascades deletion to all associated closing marks in `portfolio.instrument_prices` within an atomic database transaction. Protects referential integrity by rejecting deletion if an asset is actively held or referenced by user transactions (`ErrInstrumentInUse`, `FailedPrecondition`), accompanied by a safety confirmation dialog (`DeleteAssetModal`).
- **Admin Historical Market Data Backfill**: Operators can backfill daily closing prices and ECB FX fixings over custom date ranges (up to ~5 years) from the Admin Portal's `BackfillModal`. Offers date presets (`30D`, `90D`, `YTD`, `1Y`, `ALL`), All-Active vs. Specific symbol/pair scopes, API token budget impact projections, and optional retroactive valuation recompute for affected portfolios. Available from the Ingestion Pipeline console and the Price and FX management toolbars. Per-asset provider failures are reported as non-fatal warnings and do not abort the batch.
- **Missing Prices Diagnostic Report & Date Range Backfill (Admin Portal)**: Dedicated operational diagnostic tool in the internal Admin Portal (`:5174`). Detects and aggregates trade transactions (`BUY`, `SELL`, `SPLIT`, excluding cash distributions like `DIVIDEND`) that lack corresponding market closing marks in `portfolio.instrument_prices` on execution dates. Displays summary KPI cards (Missing Transactions count, Impacted Assets, Date Span), date range presets (`30D`, `90D`, `YTD`, `1Y`, `ALL`), symbol and transaction type filters, and one-click backfill integrations: row-level "Backfill Range" and bulk "Backfill All Missing" opening `BackfillModal` pre-populated with affected instruments and date boundaries, plus inline manual closing price override shortcuts (`AddAssetPriceModal`).
- **Missing FX Rates Diagnostic Report & Date Range Backfill (Admin Portal)**: Dedicated operational diagnostic tool in the internal Admin Portal (`:5174`). Detects and aggregates cross-currency trade transactions ($t.\text{currency\_code} \neq p.\text{base\_currency}$) lacking authoritative daily closing foreign exchange rates in `portfolio.fx_rates` on execution dates. Supports bidirectional quote lookup (direct vs. reciprocal inversion $1/\text{rate}$), summary KPI cards (Total Missing Transactions, Impacted Currency Pairs, Earliest Trade Date, Date Range Span), date range presets (`30D`, `90D`, `1Y`, `ALL`), base and quote currency filters, and one-click backfill integrations: row-level "Backfill FX" opening `BackfillModal` pre-populated with affected currency pairs and a 3-day buffer on `fromDate` (accounting for weekend/holiday trade dates and preceding Friday central bank fixings), and inline manual rate override shortcut (`FXOverrideModal`).
- **Multi-Broker CSV Ingestion & Cash Account Statement Parsing**: Seamless batch statement ingestion supporting leading Australian retail brokers (CommSec and nabtrade). Features automated header signature detection with manual override, dialect parsing for Australian dates (`DD/MM/YYYY`) and currency formatting, domestic and international cash account trade matching with native brokerage fee derivation, full cash account statement ingestion (capturing `DEPOSIT`, `WITHDRAWAL`, `INTEREST`, and domestic ETF/stock dividends while ignoring informational interest rate updates to eliminate negative cash balances), idempotent deduplication via pre-flight external reference checking, and an interactive glassmorphic preview modal (`ImportTransactionsModal`) with validation diagnostics. Backend microservices support flexible ticker suffix matching (`IVV` ↔ `IVV.AX` ↔ `IVV.ASX`), unmapped instrument auto-provisioning, atomic batch persistence, single-pass projection rebuilds, and retroactive valuation backfills.
- **User Preferences & Dynamic Multi-Currency Re-anchoring**: Manage investor profile display name, UI theme (`DARK`, `LIGHT`, `SYSTEM`), and base display currency (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`). Changing preferred currency automatically triggers atomic base currency re-anchoring on `portfolio-api`, re-scaling valuations and holdings cost bases via live FX triangulation without data drift.
- **Investment Holding Return Attribution (Capital Gain, Income & Currency Gain/Loss)**: Full 3-pillar return decomposition for every investment holding using exact fixed-point decimal arithmetic:
  - **Capital Gain/Loss**: Local asset price movement converted at the effective weighted-average acquisition exchange rate: $\text{CapGain}_{\text{base}} = (V_{\text{local}} - C_{\text{local}}) \times \overline{\text{FX}}_0$.
  - **Income (Dividends & Distributions)**: Cumulative gross dividends received in base currency, accompanied by Yield on Cost ($\text{YOC}\%$).
  - **Currency Gain/Loss**: Foreign exchange rate fluctuation between the acquisition rate and current market rate: $\text{FXGain}_{\text{base}} = V_{\text{local}} \times (\text{FX}_t - \overline{\text{FX}}_0)$ for international assets, with a clean neutral indicator (`—`) for domestic assets.
  - **Verified Mathematical Identity**: $\text{CapGain}_{\text{base}} + \text{FXGain}_{\text{base}} \equiv V_{\text{base}} - C_{\text{base}}$ invariant preserved across all multi-currency assets.
  - **Expanded Desktop Layout**: Widened dashboard layout (`max-width: 1680px`) with 10 columns (`Asset`, `Price`, `Avg Buy Price`, `Quantity`, `Total Value`, `Capital Gain`, `Income`, `Currency Gain`, `Total Return`, `Today's Return`), `INTL` badges, itemized subtotal footer row, and responsive card views on mobile.
- **Closed Positions Visibility & Lifetime Return Attribution**: View fully exited or liquidated investments ($Q = 0$) directly in the dashboard "Your Investments" table via an interactive `"Include closed positions"` toggle. Displays a distinct `CLOSED` badge, reconstructs volume-weighted average acquisition buy price from historical tax lots and disposals, surfaces lifetime realized capital gains and historical dividend income, and calculates net return percentages with zero division risk and complete valuation neutrality ($V = \$0.00$). Preference is persisted in browser local storage across sessions.
- **Portfolio Cash Flow Report & Unified Reports Hub**: Dedicated investor reporting workspace accessible via the dashboard "Reports" tab (`ReportsView`). Features a comprehensive Cash Flow Report (`CashFlowReport`) with dynamic timeframe selection (`MTD`, `QTD`, `YTD`, `Last Year`, `All-Time`, `Custom Date Range`), 4 summary KPI cards (Opening Cash Balance, Total Inflows, Total Outflows, Closing Cash Balance, Net Cash Flow), categorized breakdowns (Inflows: Deposits, Dividends, Asset Sales; Outflows: Withdrawals, Asset Purchases, Brokerage Fees), mathematical reconciliation verification ($\text{Closing Balance} \equiv \text{Opening Balance} + \text{Net Cash Flow}$), chronological transaction ledger with running cash balances, and client-side CSV export. Designed as an extensible hub for upcoming tax and dividend reports.
- **Modified Dietz Annualized Return & Cash-Flow Weighting**: Accurate, GIPS Standard 2.A.20-compliant money-weighted rate of return calculation ($R_{\text{Dietz}} = \frac{V_1 - V_0 - F}{V_0 + \sum (F_i \times W_i)}$) with Actual/Actual calendar day weighting ($W_i = \frac{CD - D_i}{CD}$). Employs compound geometric annualization for multi-year periods ($\ge 365$ days) with sub-annual protection, optional monthly sub-period linking to prevent cash flow dilution, and an interactive glassmorphic modal (`ModifiedDietzModal`) featuring capital balance decomposition and an itemized cash-flow weight inspection table.
- **Dual Return Methodology Switcher (`[ TWR | DIETZ ]`)**: Dashboard metric card toggle empowering investors to switch seamlessly between pure investment performance (Time-Weighted Return) and dollar-weighted personal performance (Modified Dietz). Includes `localStorage` persistence, inline methodology badges (`GIPS TWR` vs `MODIFIED DIETZ`), quick inspection launch, and automatic pre-selection of the portfolio's true inception date.
- **Portfolio Benchmark Comparison & Relative Performance Analytics**: Overlay and compare portfolio cumulative and annualized performance against industry-standard market benchmarks (S&P 500 Total Return `SP500TR`, NASDAQ-100 `QQQ`, and ASX 200 `VAS`) on the interactive SVG performance chart across any selected timeframe (`1D`, `1W`, `1M`, `1Y`, `ALL`). Conforms strictly to CFA Institute and GIPS Standard 2.A:
  - **Common Origin Re-Basing ($T_0$ Normalization)**: Indexes both the portfolio (via unitized daily TWR index) and benchmark closing prices to $0.00\%$ at horizon start date $T_0$, ensuring cash-flow-neutral comparability.
  - **Dual Currency Perspectives**: Toggle seamlessly between **Portfolio Base Currency Mode** (converts foreign benchmark prices into portfolio base currency via daily historical spot FX to measure true opportunity cost to the local investor) and **Benchmark Native Currency Mode** (evaluates unhedged local index return).
  - **Total Return Prioritization**: Eliminates deceptive dividend bias by benchmarking against total return series (`TOTAL_RETURN`) where cash distributions are reinvested.
  - **GIPS Sub-1-Year Annualization Protection**: Enforces GIPS Standard 2.A.20, strictly prohibiting compound annualization for periods under 365 calendar days ($CD < 365$) while geometrically compounding multi-year horizons ($CD \ge 365$).
  - **Active Risk Attribution & Metrics Strip**: Computes Active Return / Alpha ($R_p - R_b$), Portfolio Beta ($\beta$), Annualized Tracking Error ($\text{TE}$), and Information Ratio ($\text{IR}$) for horizons with $\ge 30$ trading days, displayed in an interactive metrics strip with glowing outperformance/underperformance badges.
  - **Interactive Dual-Line SVG Chart & Synchronized Crosshairs**: Renders a secondary dashed benchmark curve alongside the solid portfolio curve, dual-metric hover inspection crosshairs, benchmark selector dropdown, currency mode toggle, and `localStorage` persistence.
- **User Registration, Credential Authentication & Multi-Tenant Session Management**: Institutional-grade authentication and session security conforming to OWASP ASVS v4.0 Level 2, NIST SP 800-63B, and RFC 9106:
  - **Argon2id Cryptographic Password Hashing**: Passwords hashed using RFC 9106 Argon2id ($m=64\text{MB}, t=3\text{ iterations}, p=4\text{ lanes}$) with 16-byte random salts, constant-time verification, and automatic dummy hash evaluation to protect against user enumeration timing attacks. Real-time complexity validation requires $\ge 10$ characters, upper and lowercase letters, numbers, and special symbols.
  - **Dual-Token Architecture**: Short-lived (15-minute) standard-library HMAC-SHA256 JWT access tokens stored exclusively in browser memory (`AuthContext`), paired with long-lived (7- to 30-day) cryptographically hashed opaque refresh tokens transmitted via `httpOnly`, `Secure`, `SameSite=Strict` cookies.
  - **Token Family Rotation & Reuse Detection**: Every refresh invalidates the current token and issues a new pair. If a revoked or replayed refresh token is detected, the entire session family is revoked immediately to prevent session hijacking.
  - **Automatic Default Portfolio Auto-Provisioning**: Registering a new account atomically provisions an initial default portfolio (`"Main Portfolio"`) in `services/portfolio-api`, anchored to the investor's selected base currency (`USD`, `EUR`, `AUD`, `GBP`, `CAD`, `JPY`, `CHF`).
  - **Decoupled Multi-Tenant BFF**: Contextual HTTP authorization middleware extracts and verifies bearer JWTs, injecting `ctx.UserID` to scope all portfolio, cash flow, benchmark, and preference resolvers to the authenticated tenant.
  - **Dark Glassmorphic UI & Session Lifecycle**: Centered `AuthView` with `[ Sign In | Create Account ]` tabs, live `PasswordStrengthMeter` checklist, silent background token renewal 60 seconds prior to expiration, and an investor `UserProfileMenu` dropdown with profile details, preferences shortcut, and one-click sign out.
  - **Credential Management & Password Change**: In-app password rotation with mandatory current-password re-authentication, live complexity checklist, prevention of identical password reuse, and optional invalidation of all other active sessions across devices.
- **Clean Microservice Monorepo**: Contract-first gRPC services with a Go GraphQL Backend-for-Frontend (BFF) and strongly-typed frontend queries.
- **Responsive UI**: Glassmorphic, dark-mode dashboard built with React 19, TypeScript, Vite, modal transaction entry (`AddTransactionModal`), statement import (`ImportTransactionsModal`), investor preferences dialog (`UserPreferencesModal`), reports hub (`ReportsView`), return calculator (`ModifiedDietzModal`), and instant reactive state refresh.

> 📋 *For a comprehensive list of completed milestones, in-progress components, and planned roadmap items, see [`FEATURES.md`](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md).*

---

## Tech Stack

| Layer | Technologies |
| --- | --- |
| **Frontend** | React 19, TypeScript, Vite, Vanilla CSS, GenQL |
| **Backend-for-Frontend (BFF)** | Go 1.26+, GraphQL (`gqlgen`), gRPC Client, HTTP JWT Auth Middleware |
| **Core Microservices** | Go 1.26+, gRPC, Protocol Buffers (`protoc-gen-go`, `protoc-gen-go-grpc`) |
| **Authentication & Security** | Argon2id (RFC 9106), HMAC-SHA256 JWT, SHA-256 session token hashing, `httpOnly` `SameSite=Strict` cookies |
| **Database & Persistence** | PostgreSQL 17+ (18+ supported), `pgx/v5`, `golang-migrate` |
| **Mocking & Testing** | `go.uber.org/mock` (`mockgen`), standard library Go test runner with `-race` |
| **Orchestration** | GNU Make, Go Workspaces (`go.work`) |

---

## Prerequisites

Ensure the following tools are installed on your machine:

- **Go 1.26+**
- **Node.js 20+** and **npm**
- **PostgreSQL 17+** running locally at `localhost:5432` (`psql`, `pg_isready`)
- **`golang-migrate` CLI**: `brew install golang-migrate`
- **Protobuf Compiler (`protoc`)**: `brew install protobuf` *(only required if editing `.proto` files)*

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/garcios/graphfolio.git
cd graphfolio
```

### 2. Configure Environment

Copy the example environment file:

```bash
cp .env.example .env
```

The default `.env` is pre-configured for a local PostgreSQL instance:

```env
PG_ADMIN_URL=postgres://$(USER)@localhost:5432/postgres
PORTFOLIO_DB_URL=postgres://portfolio_svc:portfolio@localhost:5432/graphfolio?sslmode=disable
USER_DB_URL=postgres://user_svc:user@localhost:5432/graphfolio?sslmode=disable
MIGRATE_ON_START=false
JWT_SECRET=graphfolio-default-dev-secret-key-256bit-ok!
```

### 3. Install Dependencies

Install all frontend npm packages and backend Go module dependencies:

```bash
make install
```

### 4. Bootstrap and Migrate Database

Set up database roles, schemas, migrations, and development seed data:

```bash
# Verify PostgreSQL is running
make db-check

# Bootstrap roles (portfolio_svc, user_svc), database (graphfolio), and schemas
make db-bootstrap

# Run all schema migrations for portfolio-api and user-api
make migrate-up

# Seed development market data, instruments, demo portfolios, and transactions
make db-seed

# (Alternatively, run all four steps with: make db-reset)
```

### 5. Generate Code Contracts, Models, and Mocks

Compile protobufs, generate GraphQL models, regenerate the typed GenQL frontend client, and generate Go unit test mocks:

```bash
make generate
```

### 6. Start Development Servers

Start the Portfolio gRPC API, the User gRPC API, the GraphQL BFF, and the Vite frontend concurrently:

```bash
make run
```

- **Primary Investor Web App**: [http://localhost:5173](http://localhost:5173)
- **Internal Admin Portal**: [http://localhost:5174](http://localhost:5174) *(run separately via `make run-admin`)*
- **GraphQL Playground (BFF)**: [http://localhost:8080/](http://localhost:8080/)
- **Portfolio gRPC API**: `localhost:50051`
- **User gRPC API**: `localhost:50052`

> 💡 **Default Development Credentials** (seeded via `make db-seed`):
> - **Email**: `oscar@graphfolio.com` (or `tom@cruz.com`)
> - **Password**: `Password123!` *(Note: case-sensitive with capital `P` and exclamation mark `!`)*

---

## Architecture & Data Flow

GraphFolio enforces a strict separation of concerns across contract-first, microservice, and frontend boundaries:

```mermaid
flowchart TB
    subgraph WebMonorepo ["Frontend Workspace Monorepo (web/)"]
        subgraph MainApp ["Primary Investor Application (:5173)"]
            MAIN_UI["Investor Views and Dashboards<br/>• AuthView (Login and Register)<br/>• 10-Col Holdings Return Attribution<br/>• Benchmark SVG Chart and Metrics Strip<br/>• Cash Flow Report and Modified Dietz Modal<br/>• User Preferences and Change Password Modal"]
            AUTH_CTX["AuthContext and State<br/>• In-Memory Bearer Token<br/>• Silent Refresh Timer (~60s Pre-Expiry)"]
        end

        subgraph AdminApp ["Internal Operations Portal (:5174)"]
            ADMIN_UI["Administrative Portals<br/>• Price Management and Override Presets<br/>• Missing Price and FX Diagnostics<br/>• FX Rates and Backfill Range Dispatcher"]
        end

        subgraph SharedPackages ["Shared Workspace Packages"]
            CLIENT_PKG["@graphfolio/api-client<br/>• Typed GenQL Client<br/>• Dynamic Bearer Token Injection<br/>• Automatic Cookie Credentials"]
            UI_PKG["@graphfolio/ui<br/>• Dark Glassmorphic Design System<br/>• Atomic Components and Formatters"]
        end
    end

    subgraph BFFLayer ["GraphQL Backend-for-Frontend (:8080)"]
        BFF_MW["HTTP Auth Middleware<br/>• HMAC-SHA256 JWT Verification<br/>• httpOnly Cookie Management<br/>• Scoped ctx.UserID Injection"]
        BFF_RESOLVER["GraphQL Resolvers and Orchestrator<br/>• Auth: register, login, refresh, changePassword<br/>• Portfolio: holdings, transactions, CSV import<br/>• Performance: history, Dietz, benchmarks<br/>• Admin: prices, FX, diagnostics, backfills<br/>• Dynamic Base Currency Re-anchoring"]
    end

    subgraph ProtoLayer ["Protocol Buffers Contract Layer (proto/)"]
        COMMON_PB["common/v1/decimal.proto<br/>• Exact Decimal and Money Contracts"]
        PORT_PB["portfolio/v1/portfolio.proto<br/>• Portfolio, Benchmarks and Ledger RPCs"]
        USER_PB["user/v1/user.proto<br/>• User Auth, Sessions and Preferences RPCs"]
    end

    subgraph ServiceLayer ["Core Domain Microservices (services/)"]
        subgraph PortService ["Portfolio API (:50051) and Worker"]
            PORT_SERVER["Portfolio gRPC Server<br/>• Deterministic Ledger Replay Engine<br/>• 3-Pillar Return Attribution Calculator<br/>• Benchmark Comparison and Risk Engine<br/>• Modified Dietz Actual/Actual Engine<br/>• Cash Flow Reconciliation Engine<br/>• Historical Valuation Snapshots and Backfill"]
        end

        subgraph UserService ["User API (:50052)"]
            USER_SERVER["User gRPC Server<br/>• Argon2id Cryptographic Hasher (RFC 9106)<br/>• Session Family Rotation and Reuse Detection<br/>• Credential Rotation and Multi-Session Revocation<br/>• User Preferences and Supported Currencies"]
        end
    end

    subgraph DBLayer ["PostgreSQL 17 Database (:5432/graphfolio)"]
        PORT_DB["Portfolio Schema [portfolio_svc]<br/>• portfolios, transactions, holding_projections<br/>• tax_lots, lot_disposals, portfolio_valuations<br/>• daily_prices, fx_rates, benchmarks"]
        USER_DB["Users Schema [user_svc]<br/>• users (password_hash, preferences, theme)<br/>• sessions (token_hash, family_id, is_revoked)"]
    end

    %% Web Monorepo Internal Bindings
    MAIN_UI --> CLIENT_PKG
    ADMIN_UI --> CLIENT_PKG
    MAIN_UI -.-> UI_PKG
    ADMIN_UI -.-> UI_PKG
    MAIN_UI --- AUTH_CTX

    %% Frontend to BFF Communication
    CLIENT_PKG -->|"GraphQL HTTP"| BFF_MW
    BFF_MW --> BFF_RESOLVER

    %% BFF to Microservices gRPC
    BFF_RESOLVER -->|"gRPC on :50051"| PORT_SERVER
    BFF_RESOLVER -->|"gRPC on :50052"| USER_SERVER

    %% Protobuf Contract Implementations
    PORT_SERVER -.-> PORT_PB
    PORT_SERVER -.-> COMMON_PB
    USER_SERVER -.-> USER_PB

    %% Microservices to Database
    PORT_SERVER -->|"pgxpool / SQL"| PORT_DB
    USER_SERVER -->|"pgxpool / SQL"| USER_DB
```

### Exact Decimal Precision

Financial applications cannot tolerate IEEE 754 binary floating-point rounding errors:
- Contract layer defines fixed-point types in [`proto/common/v1/decimal.proto`](file:///Users/oscargarcia/workspace/graphfolio/proto/common/v1/decimal.proto).
- Shared package [`pkg/decimalpb`](file:///Users/oscargarcia/workspace/graphfolio/pkg/decimalpb/decimalpb.go) converts losslessly between protobuf messages and `shopspring/decimal.Decimal`.
- PostgreSQL drivers automatically register the `pgx-shopspring-decimal` extension on every pool connection.
- GraphQL exposes a dedicated `Decimal` scalar in [`bff/graph/model/decimal.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/model/decimal.go).

### Ledger Replay & Cost Basis Methods

1. **Transaction Ledger**: All investment events (buys, sells, dividends, transfers, deposits, withdrawals, stock splits) are appended to `portfolio.transactions`.
2. **Deterministic Projections**: When transactions are added, deleted, or cost-basis methods switch, the replay engine in [`services/portfolio-api/internal/service/projection.go`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go) re-evaluates all transactions to produce:
   - `tax_lots` and `lot_disposals`
   - `holding_projections` (quantity, cost basis, realized PnL, dividend income)
   - `cash_balances`
3. **Multi-Currency Brokerage Fees**: When trade fees are paid in a currency different from the asset (e.g. AUD brokerage for USD shares), the engine deducts gross trade proceeds from the asset currency cash account and the fee from the fee currency cash account. In base currency, tax lot cost bases and disposal net proceeds properly incorporate the fee via live or historical exchange rates.
4. **Stock Splits (`SPLIT`)**: Corporate action stock splits scale the share quantities of all open tax lots acquired prior to the split date ($Q_{\text{new}} = Q_{\text{old}} \times \text{multiplier}$) while keeping total cost basis and cash balance unchanged. Subsequent sales relieve split-adjusted lots.
5. **Atomic Replacement**: Rebuild runs under a row lock (`SELECT ... FOR UPDATE` on `portfolio.portfolios`) to serialize concurrent updates and atomically save recomputed projections in `SaveProjectionsTx`.
6. **Accounting Methods**:
   - `AVERAGE_COST`: Disposals proportionally relieve cost basis across all open tax lots.
   - `FIFO`: Disposals relieve the earliest acquired tax lots first.

### User Identity, Preferences & Dynamic Currency Re-anchoring

1. **User Domain Microservice**: `services/user-api` (Port `:50052`) owns the `users.users` table, isolating user profile identity, display name, UI theme (`DARK`, `LIGHT`, `SYSTEM`), and preferred display currency (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`).
2. **Atomic Base Currency Synchronization**: When an investor updates their display currency in `UserPreferencesModal`:
   - BFF dispatches `UpdateUserPreferences` to `user-api` to persist user profile settings.
   - BFF dispatches `UpdatePortfolioBaseCurrency` to `portfolio-api` within an atomic database transaction.
   - `portfolio.portfolios.base_currency` is updated, and historical valuation snapshots (`portfolio.portfolio_valuations`) and holding cost bases (`portfolio.holdings`) are re-scaled proportionally via live FX rates with triangulation fallback.
   - The GraphQL mutation returns the re-calculated portfolio, instantly updating all cards, tables, and metrics without floating-point error.
3. **Query-Level Consistency**: The GraphQL `portfolio` query automatically verifies the investor's preferred display currency with `user-api` and aligns `portfolio-api`'s base currency if necessary.

### User Authentication, Multi-Tenancy & Session Management

1. **Cryptographic Identity Protection**:
   - `services/user-api` hashes passwords using **Argon2id** (RFC 9106: 64 MB memory, 3 iterations, 4 parallelism lanes, 16-byte cryptographically secure random salt).
   - Verifications execute in constant time (`subtle.ConstantTimeCompare`). To prevent user enumeration via timing discrepancies, login requests for non-existent emails compute a dummy Argon2id hash before returning generic credential errors.
2. **Dual-Token Session Lifecycle**:
   - **Access Token (JWT)**: Short-lived (15 minutes), signed with standard-library HMAC-SHA256 containing `sub` (User UUID), `email`, `display_name`, and `base_currency`. Held exclusively in frontend memory (`AuthContext`); never stored in `localStorage` or `sessionStorage` (mitigating XSS extraction).
   - **Refresh Token**: Long-lived (7 days by default, 30 days if "Remember me" is checked), cryptographically secure opaque token (32 random bytes, base64url-encoded). The SHA-256 hash is recorded in `users.sessions` alongside token family identifiers (`family_id`).
   - **Cookie Transport**: Stored in a browser cookie (`graphfolio_refresh`) with `HttpOnly`, `Secure`, `SameSite=Strict`, and `Path=/`.
3. **Token Family Rotation & Hijack Detection**:
   - Every invocation of the `refreshToken` mutation generates a new refresh token and rotates the session record.
   - If an already-rotated or revoked refresh token is presented (indicating a potential replay attack or token leak), the system revokes all sessions in that `family_id` immediately, requiring re-authentication across all compromised devices.
4. **Tenant Isolation & Contextual Authorization**:
   - BFF's `auth.Middleware` extracts incoming `Authorization: Bearer <token>` headers, verifies token signatures and expiration, and binds the authenticated user ID into the Go request context (`ctx.UserID`).
   - All portfolio queries and mutations resolve user identity strictly from `r.resolveUserID(ctx)`, ensuring complete tenant isolation and multi-user data segregation.
5. **Initial Portfolio Auto-Provisioning**:
   - When a new investor completes registration via `register`, the BFF coordinates with `portfolio-api`'s `CreatePortfolio` gRPC endpoint to automatically instantiate an initial default `"Main Portfolio"` anchored in the investor's designated base currency.

---

## Directory Structure

```text
├── proto/                        # Single Source of Truth for APIs (Protobuf definitions)
│   ├── common/v1/decimal.proto   # Decimal and Money contracts
│   ├── portfolio/v1/             # Portfolio gRPC service (GetPortfolio, CreatePortfolio, UpdatePortfolioBaseCurrency, AddTransaction, GetCashFlowReport, ListMissingFXRateTransactions, DeleteInstrument, TriggerBackfill, etc.)
│   └── user/v1/user.proto        # User gRPC service (Register, Login, RefreshToken, Logout, GetMe, GetUserPreferences, UpdateUserPreferences, ListSupportedCurrencies)
├── pkg/                          # Shared Go infrastructure
│   ├── database/                 # pgx connection pooling, auto .env loading, migration runner
│   └── decimalpb/                # Decimal/Money conversions between proto and shopspring
├── scripts/db/                   # Database bootstrap and teardown scripts
├── services/                     # Domain Microservices
│   ├── portfolio-api/            # Portfolio business logic, ledger replay, valuations, migrations, seeds
│   │   ├── cmd/server/           # Application entrypoint (:50051)
│   │   ├── cmd/worker/           # Scheduled EOD & on-demand valuation worker (make run-valuation-job)
│   │   ├── cmd/market-ingest/    # Market data ingestion & historical backfill CLI (make ingest-market-data)
│   │   ├── internal/             # Domain entities (incl. missing_fx.go), repository, service (incl. cash flow report & engine), and gRPC server
│   │   ├── migrations/           # Versioned schema migrations (000001 - 000010)
│   │   └── seeds/                # Seed fixtures (dev_seed.sql with 365-day history)
│   └── user-api/                 # User domain microservice (:50052)
│       ├── cmd/server/           # Application entrypoint (:50052)
│       ├── internal/             # User domain, crypto (Argon2id), auth (JWT/hashing), repository (pgx), service, and gRPC server
│       ├── migrations/           # Versioned schema migrations (000001, 000002_add_theme, 000003_add_auth_and_sessions)
│       └── seeds/                # User dev fixtures (dev_seed.sql with demo investor & Argon2id hash)
├── bff/                          # GraphQL Backend-for-Frontend (gqlgen on :8080)
│   ├── cmd/server/               # BFF entrypoint
│   ├── internal/auth/            # JWT authorization middleware and httpOnly cookie management
│   └── graph/                    # Schema, resolvers, helpers, model
├── web/                          # Frontend Workspace Monorepo
│   │   ├── main-app/             # Primary Investor React App (:5173)
│   │   │   ├── src/context/      # AuthContext.tsx (in-memory access token, silent refresh, session state)
│   │   │   └── src/components/   # AuthView, LoginForm, RegisterForm, PasswordStrengthMeter, UserProfileMenu, Dashboard, PerformanceChart, ModifiedDietzModal, TransactionLedger, ReportsView, CashFlowReport, AddTransactionModal, ImportTransactionsModal, UserPreferencesModal
│   │   └── admin-app/            # Internal Operations Portal (:5174)
│   │       └── src/components/   # AssetManagement, DeleteAssetModal, PriceManagement, MissingPricesReport, MissingFXRatesReport, FXManagement, FXTrendChart, FXOverrideModal, BackfillModal, IngestionPipeline
│   └── packages/
│       ├── ui/                   # Shared Design System (@graphfolio/ui)
│       └── api-client/           # Shared GraphQL Client (@graphfolio/api-client with dynamic access token & credentials)
├── docs/                         # Implementation plans and guides
│   ├── plans/                    # System implementation plans
│   │   ├── db-implementation-plan.md
│   │   ├── portfolio-service-implementation-plan.md
│   │   ├── add-transactions-implementation-plan.md
│   │   ├── performance-chart-implementation-plan.md
│   │   ├── portfolio-valuation-engine-implementation-plan.md
│   │   ├── market-data-ingestion-implementation-plan.md
│   │   ├── transaction-history-implementation-plan.md
│   │   ├── cost-basis-switching-implementation-plan.md
│   │   ├── user-preferences-implementation-plan.md
│   │   ├── fundamental-cash-flow-engine-implementation-plan.md
│   │   ├── currency-pair-historical-prices-implementation-plan.md
│   │   ├── admin-historical-backfill-implementation-plan.md
│   │   ├── asset-prices-date-range-filter-implementation-plan.md
│   │   ├── admin-missing-prices-report-implementation-plan.md
│   │   ├── admin-missing-fx-rates-report-implementation-plan.md
│   │   ├── delete-asset-implementation-plan.md
│   │   ├── exchange-dropdown-implementation-plan.md
│   │   ├── ingestion-job-history-implementation-plan.md
│   │   ├── multi-broker-csv-ingestion-implementation-plan.md
│   │   ├── cash-statement-csv-ingestion-implementation-plan.md
│   │   ├── holding-return-attribution-implementation-plan.md
│   │   ├── portfolio-cash-flow-report-implementation-plan.md
│   │   ├── include-closed-positions-implementation-plan.md
│   │   ├── annualized-return-modified-dietz-implementation-plan.md
│   │   ├── user-registration-and-login-implementation-plan.md
│   │   └── web-workspace-refactoring-plan.md
│   ├── user-stories/             # Product specifications & acceptance criteria
│   │   ├── investment-holding-capital-gain-income-currency.md
│   │   ├── portfolio-cash-flow-report-generation.md
│   │   ├── admin-missing-prices-report.md
│   │   ├── admin-missing-fx-rates-report.md
│   │   ├── include-closed-positions-dashboard.md
│   │   ├── annualized-return-modified-dietz.md
│   │   └── user-registration-and-login.md
│   ├── negative-cash-balance-investigation-report.md # Cash ledger overdraft root cause & remediation report
│   ├── competitive-analysis.md
│   └── genql-usage.md
├── FEATURES.md                   # Master features matrix & roadmap (completed & planned)
├── Makefile                      # Standardized development workflows
└── go.work                       # Go workspace mapping modules
```

---

## Testing & Quality Assurance

### Running Tests

Execute unit tests across all Go workspace modules with race condition detection:

```bash
make test
```

### Unit Tests with Mocks (`go.uber.org/mock`)

Service and transport layer unit tests run with zero external dependencies using mocks generated by `mockgen`:
- **Repository Mock**: [`MockRepository`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/repository/mocks/mock_repository.go) allows verifying `PortfolioService` business calculations, error handling, and projection rebuilds without a database.
- **Service Mock**: [`MockPortfolioService`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/mocks/mock_service.go) allows verifying gRPC server request routing, error code mapping (`codes.NotFound`, `codes.Internal`), and fallback behavior.
- **Ingestion Mock**: [`MockIngestionService`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/mocks/mock_ingestion_service.go) allows verifying historical backfill orchestration (`TriggerBackfill`) across asset price and FX provider calls, including partial-failure warning collection.

To regenerate mocks after modifying Go interfaces:

```bash
make generate
```

---

## Available Make Commands

| Command | Description |
| --- | --- |
| `make install` | Installs npm packages and downloads all Go module dependencies |
| `make proto` | Compiles `.proto` contracts into Go packages |
| `make generate` | Runs proto compilation, gqlgen generation, genql client build, and mockgen |
| `make db-check` | Tests if local PostgreSQL is accepting connections on localhost:5432 |
| `make db-bootstrap` | Idempotently creates database roles (`portfolio_svc`, `user_svc`), DB, and schemas |
| `make db-drop` | Drops local development database |
| `make migrate-up` | Applies all pending migrations across all microservices |
| `make migrate-down` | Rolls back the latest migration step across services |
| `make db-seed` | Seeds development instruments, market data, and demo portfolios |
| `make db-reset` | Recreates database from scratch: drop, bootstrap, migrate, and seed |
| `make test` | Runs unit tests across all modules in `go.work` with race detection |
| `make ingest-market-data` | Runs the market data closing price and ECB FX fixing ingestion pipeline |
| `make run-valuation-job` | Runs the scheduled EOD valuation snapshot across active portfolios or historical backfills |
| `make run-portfolio` | Starts the Portfolio gRPC microservice on port 50051 |
| `make run-user` | Starts the User gRPC microservice on port 50052 |
| `make run-bff` | Starts the GraphQL BFF on port 8080 |
| `make run-web` | Starts the Primary Investor frontend (`main-app`) on port 5173 |
| `make run-admin` | Starts the Internal Operations Portal (`admin-app`) on port 5174 |
| `make run-all-web` | Concurrently starts both `main-app` and `admin-app` |
| `make build-web` | Builds production bundles for all web workspaces |
| `make run` | Concurrently starts Portfolio API, User API, BFF, and React web servers |

---

## Documentation Links

- **Repository Rules & Guidelines**: [`AGENTS.md`](./AGENTS.md)
- **Frontend Workspace Architecture Plan**: [`docs/plans/web-workspace-refactoring-plan.md`](./docs/plans/web-workspace-refactoring-plan.md)
- **Database Architecture & Schema Design**: [`docs/plans/db-implementation-plan.md`](./docs/plans/db-implementation-plan.md)
- **Portfolio Service Implementation Plan**: [`docs/plans/portfolio-service-implementation-plan.md`](./docs/plans/portfolio-service-implementation-plan.md)
- **Add Transactions Implementation Plan**: [`docs/plans/add-transactions-implementation-plan.md`](./docs/plans/add-transactions-implementation-plan.md)
- **Performance Chart Implementation Plan**: [`docs/plans/performance-chart-implementation-plan.md`](./docs/plans/performance-chart-implementation-plan.md)
- **Transaction History Implementation Plan**: [`docs/plans/transaction-history-implementation-plan.md`](./docs/plans/transaction-history-implementation-plan.md)
- **Cost Basis Switching Plan**: [`docs/plans/cost-basis-switching-implementation-plan.md`](./docs/plans/cost-basis-switching-implementation-plan.md)
- **Portfolio Valuation Engine Plan**: [`docs/plans/portfolio-valuation-engine-implementation-plan.md`](./docs/plans/portfolio-valuation-engine-implementation-plan.md)
- **Market Data Ingestion Plan**: [`docs/plans/market-data-ingestion-implementation-plan.md`](./docs/plans/market-data-ingestion-implementation-plan.md)
- **User Preferences Implementation Plan**: [`docs/plans/user-preferences-implementation-plan.md`](./docs/plans/user-preferences-implementation-plan.md)
- **Fundamental & Cash Flow Quality Engine Plan**: [`docs/plans/fundamental-cash-flow-engine-implementation-plan.md`](./docs/plans/fundamental-cash-flow-engine-implementation-plan.md)
- **Currency Pair Historical Prices Plan**: [`docs/plans/currency-pair-historical-prices-implementation-plan.md`](./docs/plans/currency-pair-historical-prices-implementation-plan.md)
- **Admin Historical Backfill Plan**: [`docs/plans/admin-historical-backfill-implementation-plan.md`](./docs/plans/admin-historical-backfill-implementation-plan.md)
- **Asset Deletion Implementation Plan**: [`docs/plans/delete-asset-implementation-plan.md`](./docs/plans/delete-asset-implementation-plan.md)
- **Exchange Code Dropdown Plan**: [`docs/plans/exchange-dropdown-implementation-plan.md`](./docs/plans/exchange-dropdown-implementation-plan.md)
- **Ingestion Job History Plan**: [`docs/plans/ingestion-job-history-implementation-plan.md`](./docs/plans/ingestion-job-history-implementation-plan.md)
- **Multi-Broker CSV Ingestion Plan**: [`docs/plans/multi-broker-csv-ingestion-implementation-plan.md`](./docs/plans/multi-broker-csv-ingestion-implementation-plan.md)
- **Cash Statement CSV Ingestion Plan**: [`docs/plans/cash-statement-csv-ingestion-implementation-plan.md`](./docs/plans/cash-statement-csv-ingestion-implementation-plan.md)
- **Holding Return Attribution Implementation Plan**: [`docs/plans/holding-return-attribution-implementation-plan.md`](./docs/plans/holding-return-attribution-implementation-plan.md)
- **Holding Return Attribution User Story**: [`docs/user-stories/investment-holding-capital-gain-income-currency.md`](./docs/user-stories/investment-holding-capital-gain-income-currency.md)
- **Portfolio Cash Flow Report Implementation Plan**: [`docs/plans/portfolio-cash-flow-report-implementation-plan.md`](./docs/plans/portfolio-cash-flow-report-implementation-plan.md)
- **Portfolio Cash Flow Report User Story**: [`docs/user-stories/portfolio-cash-flow-report-generation.md`](./docs/user-stories/portfolio-cash-flow-report-generation.md)
- **Admin Missing Prices Report Implementation Plan**: [`docs/plans/admin-missing-prices-report-implementation-plan.md`](./docs/plans/admin-missing-prices-report-implementation-plan.md)
- **Admin Missing Prices Report User Story**: [`docs/user-stories/admin-missing-prices-report.md`](./docs/user-stories/admin-missing-prices-report.md)
- **Admin Missing FX Rates Report Implementation Plan**: [`docs/plans/admin-missing-fx-rates-report-implementation-plan.md`](./docs/plans/admin-missing-fx-rates-report-implementation-plan.md)
- **Admin Missing FX Rates Report User Story**: [`docs/user-stories/admin-missing-fx-rates-report.md`](./docs/user-stories/admin-missing-fx-rates-report.md)
- **Include Closed Positions Implementation Plan**: [`docs/plans/include-closed-positions-implementation-plan.md`](./docs/plans/include-closed-positions-implementation-plan.md)
- **Include Closed Positions User Story**: [`docs/user-stories/include-closed-positions-dashboard.md`](./docs/user-stories/include-closed-positions-dashboard.md)
- **Modified Dietz Annualized Return Implementation Plan**: [`docs/plans/annualized-return-modified-dietz-implementation-plan.md`](./docs/plans/annualized-return-modified-dietz-implementation-plan.md)
- **Modified Dietz Annualized Return User Story**: [`docs/user-stories/annualized-return-modified-dietz.md`](./docs/user-stories/annualized-return-modified-dietz.md)
- **User Registration & Login Implementation Plan**: [`docs/plans/user-registration-and-login-implementation-plan.md`](./docs/plans/user-registration-and-login-implementation-plan.md)
- **User Registration & Login User Story**: [`docs/user-stories/user-registration-and-login.md`](./docs/user-stories/user-registration-and-login.md)
- **Edit Transaction in Ledger Implementation Plan**: [`docs/plans/edit-transaction-implementation-plan.md`](./docs/plans/edit-transaction-implementation-plan.md)
- **Edit Transaction in Ledger User Story**: [`docs/user-stories/edit-transaction-ledger.md`](./docs/user-stories/edit-transaction-ledger.md)
- **Negative Cash Balance Investigation Report**: [`docs/negative-cash-balance-investigation-report.md`](./docs/negative-cash-balance-investigation-report.md)
- **Competitive Strategy Analysis**: [`docs/competitive-analysis.md`](./docs/competitive-analysis.md)
- **Frontend GraphQL Setup**: [`docs/genql-usage.md`](./docs/genql-usage.md)


