# User Story: Include Closed Positions in "Your Investments" Dashboard

> **Target**: Primary Investor Application (`web/apps/main-app`), Backend-for-Frontend (`bff/`), Portfolio Service (`services/portfolio-api`), Protobuf (`proto/portfolio/v1/`)  
> **Status**: Ready for Implementation  
> **Story ID**: US-INV-105  
> **Implementation Plan**: [include-closed-positions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/include-closed-positions-implementation-plan.md)  
> **Epic**: Portfolio Holdings Management & Return Attribution  
> **Calculation Standard**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md)  
> **Related Stories**: [investment-holding-capital-gain-income-currency.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/investment-holding-capital-gain-income-currency.md), [average_buy_price_column.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/average_buy_price_column.md)  

---

## 1. Story Statement

**As an** active investor monitoring my portfolio performance,  
**I want to** check a toggle option labeled **"Include closed positions"** directly within the "Your Investments" section on the dashboard,  
**So that** I can review the lifetime returns, realized capital gains, and historical dividend income of fully liquidated assets alongside my active holdings without losing sight of my historical investment decisions.

---

## 2. Background & Problem Context

In investment portfolio tracking, investors manage two distinct categories of holdings:
1. **Open Positions**: Assets currently owned with an active share quantity ($Q > 0$) subject to daily mark-to-market fluctuations.
2. **Closed Positions**: Assets previously purchased and subsequently 100% liquidated ($Q = 0$), where all open tax lots have been relieved.

### The Problem with Hiding Closed Positions by Default
Currently, GraphFolio's "Your Investments" table strictly surfaces open positions ($Q > 0$). When an investor liquidates their entire stake in an asset (e.g. sells all shares of `NVDA` or `BHP.AX`):
- The asset vanishes from the dashboard view entirely.
- The investor loses visibility into **how much total profit or loss was realized** on that asset (realized capital gain + cumulative cash dividends received).
- In multi-currency investing, the investor cannot inspect the realized **foreign exchange gain/loss** that occurred between acquisition and disposal dates.
- To evaluate historical investment decisions, the investor is forced to manually sift through dozens of individual transactions in the Transaction Ledger or generate external reports.

By introducing an intuitive **"Include closed positions"** control on the dashboard, investors gain complete flexibility: maintaining a clean, focused view of active capital by default, while being able to instantly inspect lifetime performance on exited positions with a single click.

---

## 3. Financial Attribution & Mathematical Formulation for Closed Positions

All calculations conform strictly to [financial-calculations standards](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), CFA Institute GIPS® standards, and US GAAP trade-date accounting using exact fixed-point arithmetic (`shopspring/decimal`).

### 3.1 Closed Position Identification Criteria
An asset $i$ qualifies as a **Closed Position** in a portfolio if and only if:
1. It has at least one historical acquisition transaction (`BUY`, `TRANSFER_IN`).
2. Its current remaining share quantity is zero:
   $$Q_i = 0$$
3. All tax lots associated with instrument $i$ have been closed ($Q_{\text{remaining}} = 0$).
4. The asset has non-zero historical activity (e.g. cumulative realized PnL $\ne 0$, dividends received $> 0$, or historical transaction volume $> 0$).

---

### 3.2 Metric Definitions for Closed Positions in the Holdings Table

The 10-column "Your Investments" table presents financial metrics adjusted for zero open inventory ($Q = 0$):

| Column | Open Position Behavior ($Q > 0$) | Closed Position Behavior ($Q = 0$) | Mathematical Definition & Notes |
| :--- | :--- | :--- | :--- |
| **Asset** | Ticker symbol + company name + `INTL` badge (if foreign) | Ticker symbol + company name + `CLOSED` badge + `INTL` badge (if foreign) | Visual `CLOSED` pill indicates the position is fully liquidated. |
| **Price** | Latest market closing price in local currency | Latest market closing price in local currency (or last available close) | Quoted Level 1 market price from `portfolio.instrument_prices`. |
| **Avg Buy Price** | Volume-weighted cost per share of remaining open lots | Volume-weighted average purchase price across all historical purchases | Split-adjusted historical acquisition price: $\frac{\sum (\text{Qty}_{\text{buy}} \times P_{\text{buy}})}{\sum \text{Qty}_{\text{buy}}}$. Render `—` if cost basis was zero/unavailable. |
| **Quantity** | Current open units ($Q > 0$) | `0` (or `0.00`) | Exact zero. |
| **Total Value** | Mark-to-market value: $Q \times P \times \text{FX}$ | **$0.00** | $0 \times P \times \text{FX} = 0.00$. Zero current capital allocation. |
| **Capital Gain** | Unrealized capital gain/loss in base currency: $(V_{\text{local}} - C_{\text{local}}) \times \overline{\text{FX}}_0$ | **Realized Capital Gain / Loss** in base currency (`realized_pnl_base`) | Reflects net cash proceeds minus relieved cost basis across all closed lots. Percentage: $\frac{\text{RealizedPnL}_{\text{base}}}{\text{HistoricalCostBasis}_{\text{base}}} \times 100\%$. |
| **Income** | Cumulative dividends received in base currency + Yield on Cost (YOC) | Cumulative dividends received while holding the asset (`dividends_base`) | Total net cash distributions received. Yield displayed relative to historical cost basis: $\frac{\text{Dividends}_{\text{base}}}{\text{HistoricalCostBasis}_{\text{base}}} \times 100\%$. |
| **Currency Gain** | Unrealized FX movement: $V_{\text{local}} \times (\text{FX}_t - \overline{\text{FX}}_0)$ | Realized currency gain/loss at disposal dates (or `—` for domestic holdings) | Captures foreign exchange divergence between acquisition date FX and liquidation date FX. |
| **Total Return** | $\text{Unrealized Gain} + \text{Income} + \text{Realized PnL}$ | $\text{Realized PnL}_{\text{base}} + \text{Dividends}_{\text{base}}$ | Net economic profit/loss generated by this asset over its entire holding lifecycle. Percentage: $\frac{\text{TotalReturn}_{\text{base}}}{\text{HistoricalCostBasis}_{\text{base}}} \times 100\%$. |
| **Today's Return** | $Q \times (P_t - P_{t-1}) \times \text{FX}_t$ | **$0.00** / `0.00%` (or neutral `—`) | Since no shares are held today, current market movements do not alter investor wealth. Tooltip: *"Position closed — no open shares held."* |

---

## 4. UI/UX Layout & Interaction Design

### 4.1 Header Placement & Toggle Control
In [Dashboard.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/components/Dashboard.tsx), within the `.table-card` section header:
- Convert the static `<h3>Your Investments</h3>` into a flex header row:
  ```html
  <div class="investments-header">
    <div class="header-title-group">
      <h3>Your Investments</h3>
      <span class="active-count-badge">8 active</span>
    </div>
    <div class="header-controls">
      <label class="closed-positions-toggle">
        <input type="checkbox" id="include-closed-positions-checkbox" checked={includeClosed} />
        <span class="toggle-label">Include closed positions</span>
      </label>
    </div>
  </div>
  ```

### 4.2 Visual Styling of Closed Position Rows
To ensure investors never confuse closed positions with active holdings:
1. **`CLOSED` Badge**: Displayed prominently next to the asset ticker (e.g. `[TSLA] [CLOSED]`). Badge style: subtle slate/neutral background (`background: rgba(255, 255, 255, 0.08)`, border: `1px solid rgba(255, 255, 255, 0.15)`, text: `var(--text-secondary)`).
2. **Muted Row Appearance**: Closed rows use subtle transparency (`opacity: 0.75`), transitioning to full opacity on hover (`:hover { opacity: 1; background: rgba(255, 255, 255, 0.03); }`).
3. **Zero Quantity & Value Clarity**: `$0.00` total value and `0` quantity are rendered in muted typography (`var(--text-muted)`).
4. **Positive/Negative Realized PnL Colors**: Realized gains and returns maintain standard semantic coloring (green for positive realized return, red for negative realized loss, neutral gray for zero).

### 4.3 Table Footers & Subtotals
When "Include closed positions" is checked:
- **Invested Assets Subtotal (Market Value)**: Remains strictly the sum of active positions (adding $0.00$ does not distort total portfolio market value).
- **Asset Count**: Displays breakdown: e.g. `"8 active, 3 closed"` (or `"11 assets"`).
- **Subtotal Total Return**: Accumulates both open unrealized returns and closed realized returns to reflect overall asset performance.

---

## 5. Acceptance Criteria

### AC 1: Toggle Control Display & Default State
- **Given** I am on the GraphFolio Dashboard page viewing the "Your Investments" section,
- **When** the page first loads,
- **Then** a checkbox control labeled **"Include closed positions"** is visible in the section header,
- **And** the checkbox is **unchecked by default**, displaying only active holdings ($Q > 0$).

### AC 2: Dynamic Inclusion of Closed Holdings
- **Given** I have at least one fully liquidated asset in my portfolio (e.g., bought and later sold all shares of `AAPL`),
- **When** I check the **"Include closed positions"** toggle,
- **Then** the table immediately re-renders to include the closed assets without a full page reload,
- **And** unchecking the toggle instantly removes closed assets, restoring the active-only view.

### AC 3: Closed Position Row Presentation & Badging
- **Given** closed positions are included in the table,
- **Then** each closed position row:
  - Displays a distinct **`CLOSED`** badge next to the ticker symbol.
  - Displays `0` in the **Quantity** column.
  - Displays `$0.00` in the **Total Value** column.
  - Displays the historical volume-weighted average purchase price in **Avg Buy Price** (or `—` if cost basis is unavailable).
  - Displays the total realized gain/loss in the **Capital Gain** column.
  - Displays cumulative net cash dividends received in the **Income** column.
  - Displays the net lifetime profit/loss ($\text{Realized PnL} + \text{Dividends}$) and percentage return in the **Total Return** column.
  - Displays `$0.00` / `0.00%` or a neutral dash `—` in **Today's Return** with an explanatory tooltip (*"Position closed — no open shares held"*).
  - Uses subtle opacity styling to visually differentiate closed positions from active positions.

### AC 4: Sorting & Interactivity
- **Given** closed positions are displayed in the table,
- **When** I click on any sortable column header (e.g., Asset, Avg Buy Price, Capital Gain, Total Return):
  - Sorting by **Total Value** or **Quantity** (descending) correctly positions closed rows ($0.00$) at the bottom.
  - Sorting by **Total Return** or **Capital Gain** sorts all rows (open and closed) strictly by their numeric return values.
  - Sorting by **Asset** sorts alphabetically across all displayed assets.

### AC 5: Preference Persistence
- **Given** I have checked "Include closed positions",
- **When** I navigate to another tab (e.g., "Ledger" or "Reports") and return to "Overview", or refresh the browser,
- **Then** my preference is preserved (via `localStorage` or component session state) so I do not have to re-check the box on every visit.

### AC 6: Empty State & Portfolios with No Closed Positions
- **Given** an investor whose portfolio has zero closed positions (all purchased assets remain open),
- **When** the investor checks "Include closed positions",
- **Then** the table continues to display all active holdings normally without errors, blank screens, or layout shifts.

---

## 6. Technical Architecture & Component Changes

```
┌────────────────────────────────────────────────────────────────────────┐
│                        web/apps/main-app                              │
│  [Dashboard.tsx]                                                       │
│    - State: includeClosed (boolean, persisted to localStorage)         │
│    - UI: Checkbox toggle in .table-card-header                         │
│    - Rendering: closed row styling, 'CLOSED' badge, 10-column values   │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ GraphQL Query: portfolio(includeClosed: Boolean)
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                                bff/                                    │
│  [schema.graphqls]                                                     │
│    - Query: portfolio(includeClosed: Boolean): Portfolio!              │
│    - Type Investment: isClosed: Boolean!                               │
│  [schema.resolvers.go] & [helpers.go]                                 │
│    - Maps gRPC GetPortfolioRequest(include_closed_positions: true)     │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ gRPC: GetPortfolio(include_closed_positions: true)
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                      services/portfolio-api                            │
│  [proto/portfolio/v1/portfolio.proto]                                  │
│    - GetPortfolioRequest: bool include_closed_positions                │
│    - Investment: bool is_closed, string closed_date                    │
│  [internal/service/projection.go] & [internal/repository/postgres.go]  │
│    - Retains / queries zero-quantity holdings with historical PnL/divs │
│  [internal/domain/calculator.go]                                       │
│    - CalculateInvestment: Zero-quantity safe calculations              │
└────────────────────────────────────────────────────────────────────────┘
```

### 6.1 Layer Breakdown

1. **Protobuf (`proto/portfolio/v1/portfolio.proto`)**:
   - Add field `bool include_closed_positions = 2;` to `GetPortfolioRequest`.
   - Add field `bool is_closed = 19;` to `Investment`.
   - Re-generate protobuf stubs via `make proto`.

2. **Domain & Service (`services/portfolio-api`)**:
   - `internal/service/projection.go`: Ensure `ProcessLedger` includes instruments with $Q = 0$ that have realized PnL or dividend history when closed projections are requested, or query historical closed holdings.
   - `internal/domain/calculator.go`: Ensure `CalculateInvestment` handles `h.Quantity.IsZero()` cleanly:
     - $V_{\text{base}} = 0.00$.
     - $\text{TotalReturn}_{\text{base}} = \text{RealizedPnL}_{\text{base}} + \text{Dividends}_{\text{base}}$.
     - Avoid division-by-zero errors when calculating percentage returns; use cumulative historical cost basis as denominator.
   - `internal/server.go`: Pass `req.GetIncludeClosedPositions()` to domain service.

3. **Backend-for-Frontend (`bff/`)**:
   - `graph/schema.graphqls`:
     - Update query: `portfolio(includeClosed: Boolean): Portfolio!`
     - Update type `Investment`: add `isClosed: Boolean!`.
   - `graph/schema.resolvers.go`: Forward `includeClosed` argument to portfolio gRPC client.
   - `graph/helpers.go`: Map `is_closed` protobuf field to GraphQL model.

4. **Web Monorepo (`web/apps/main-app`)**:
   - `src/components/Dashboard.tsx`:
     - Add `includeClosed` state initialized from `localStorage.getItem('graphfolio:include_closed_positions') === 'true'`.
     - Pass `includeClosed` to GraphQL query.
     - Add checkbox toggle in table header.
     - Add `CLOSED` badge next to ticker in table body.
     - Add conditional styling for closed rows.
   - `src/components/Dashboard.css`:
     - Styles for `.investments-header`, `.closed-positions-toggle`, `.closed-badge`, and `.closed-row`.

---

## 7. Non-Functional Requirements & Edge Cases

1. **Division by Zero Protection**: If an asset was closed with zero recorded cost basis (e.g. gifted share or missing acquisition record), all percentage return calculations must output `0.00%` or `—` without panic or NaN.
2. **Re-opened Positions**: If an investor completely sold an asset in the past, and later purchased new shares of the same asset, the asset has $Q > 0$ and must be treated as an **active** open position, incorporating both prior realized PnL and current unrealized PnL.
3. **Multiple Liquidations**: Assets sold across multiple partial lots must aggregate total historical proceeds and relieved cost bases accurately.
4. **Performance & Query Latency**: Toggling closed positions must execute within standard SLA (< 50ms roundtrip) by leveraging cached ledger projections.
5. **No Negative Total Value**: Closed positions must always render exactly `$0.00` total value, never negative or residual dust fractions.
