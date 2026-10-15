# User Story: Asset Prices Date Range Filter & Pagination (Admin Portal)

> **Story ID**: US-ADM-103  
> **Epic**: Internal Administration & Market Data Operations  
> **Target**: Admin Portal Application (`web/apps/admin-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`)  
> **Status**: **DONE**  
> **Implementation Plan**: [asset-prices-date-range-filter-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/asset-prices-date-range-filter-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #20)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L30)  
> **Standards Reference**: [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** operations administrator managing closing market data,  
> **I want to** filter the Closing Prices Ledger by custom start and end dates or quick preset buttons (`Today`, `7D`, `30D`, `90D`, `YTD`, `1Y`, `ALL`) and navigate through multi-page results,  
> **So that** I can inspect multi-week or multi-month historical price curves, audit price feeds over specific market cycles, and navigate large pricing datasets seamlessly.

---

## 2. Business Context & Usability Requirements

1. **Dual Date Range Querying**: Previously, the Admin Closing Prices Ledger (`PriceManagement`) only accepted a single date filter, forcing `fromDate = dateFilter, toDate = dateFilter`. Administrators could not view an asset's continuous time series or verify historical backfills across date spans.
2. **Quick Date Presets**: Fast preset chips allow operators to inspect relevant rolling windows (`7D`, `30D`, `90D`, `YTD`, `1Y`, `ALL`) with a single click, automatically calculating `fromDate` and `toDate` relative to today's date.
3. **Client-Side Validation Guard**: If an operator inputs an inverted date range where `fromDate > toDate`, client-side validation prevents unnecessary server roundtrips and displays an informative validation banner.
4. **Authoritative Pagination**: Large datasets exceeding the 50-item page limit must render a responsive pagination bar displaying total matching records, current page number, total pages, and navigation controls.

---

## 3. Acceptance Criteria

### 3.1 Date Range Filtering & Presets
- [x] **AC-1 (Dual Date Inputs)**: The filter bar in `PriceManagement.tsx` provides distinct `From Date` and `To Date` input fields formatted as `YYYY-MM-DD`.
- [x] **AC-2 (Quick Preset Buttons)**: Preset buttons (`Today`, `7D`, `30D`, `90D`, `YTD`, `1Y`, `ALL`) dynamically populate both `fromDate` and `toDate` and reset pagination to page 1.
- [x] **AC-3 (Inverted Date Validation)**: If `fromDate > toDate`, the UI disables query execution, flags the input borders in warning amber/red, and surfaces an error notice.

### 3.2 Pagination & Result Navigation
- [x] **AC-4 (Pagination State)**: The table displays a glassmorphic pagination toolbar showing:  
  `Showing X - Y of Z prices (Page N of M)`.
- [x] **AC-5 (Page Controls)**: "Previous" and "Next" buttons allow navigating pages, disabling appropriately at first and last pages.
- [x] **AC-6 (Page Reset on Filter Change)**: Changing the symbol filter, preset, or date inputs automatically resets the active page to 1.

### 3.3 Backend Query & Data Layer
- [x] **AC-7 (SQL Window Function)**: PostgreSQL repository utilizes `COUNT(*) OVER()` window function in `portfolio.instrument_prices` queries to retrieve total matching records alongside the page slice.
- [x] **AC-8 (Half-Open & Bounded Ranges)**: Supports querying bounded ranges (`fromDate` and `toDate`), half-open ranges (only `fromDate` or only `toDate`), or unbounded queries (`ALL`).
- [x] **AC-9 (Backfill Modal Synergy)**: Launching `BackfillModal` from the price ledger auto-populates the modal's date range inputs with the active ledger filter bounds.

---

## 4. Technical & Implementation Notes

- **Admin Component**: `web/apps/admin-app/src/components/PriceManagement.tsx` manages dual date states, preset handlers, and pagination offset calculation (`offset = (page - 1) * pageSize`).
- **Styling**: `web/apps/admin-app/src/components/PriceManagement.css` defines responsive preset pill styles and glassmorphic pagination layout.
- **BFF GraphQL**: Resolves `instrumentPrices(symbol, fromDate, toDate, limit, offset)`.
- **Backend Service**: `services/portfolio-api/internal/service/admin.go` and `repository/postgres.go` execute bounded SQL pagination queries.
