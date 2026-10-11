# Implementation Plan: Annualized Return (TWR) Valuation & Ingestion Remediation

> **Investigation Report**: [annualized-twr-return-investigation-report.md](file:///Users/oscargarcia/workspace/graphfolio/docs/annualized-twr-return-investigation-report.md)  
> **Status**: Completed  
> **Roadmap Reference**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)  
> **Financial Standards**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md) (GIPS Standard 2.A.20, US GAAP ASC 820 Level 2/3 Fair Value Hierarchy)  
> **Target Layers**: Domain Microservice (`services/portfolio-api`), Repository (`internal/repository/`), Ingestion Pipeline (`internal/service/ingestion.go`), Calculation Engine (`internal/domain/calculator.go`, `internal/service/valuation.go`), BFF (`bff/`), Frontend Presentation (`apps/main-app`)

---

## 1. Executive Summary & Objective

In GraphFolio, the primary investor dashboard displays an erroneous **`Annualized Return (TWR): -93.9%`**, despite the investor holding a profitable equity portfolio valued at **`A$450,724.53`** against an invested capital cost basis of **`~A$270,000.00`** (representing an actual economic gain of **`+60%`**).

As proven in the [Annualized Return Investigation Report](file:///Users/oscargarcia/workspace/graphfolio/docs/annualized-twr-return-investigation-report.md), this defect originates from a cascade of four interrelated vulnerabilities:
1. **Asymmetric Historical Market Data**: Real transactions dating back to May 10, 2023 were imported without pre-October 2025 closing price quotes for newly introduced instruments in `portfolio.instrument_prices`.
2. **Zero-Price Defaulting in Valuation Pipeline**: In [`valuation.go`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/valuation.go#L373-L376), unpriced holdings evaluate to `$0.00`, artificially deducting cash while generating zero market value.
3. **Catastrophic TWR Chaining on Inflow Days**: Under the Time-Weighted Return formula $R_t = \frac{(V_t - F_t) - V_{t-1}}{V_{t-1}}$, unvalued asset acquisitions on May 11, 2023 generated a phantom single-day return of **`-573.35%`**, dragging the TWR index to negative values and compounding permanently down to **`0.061055145281`**.
4. **Metric Conflation in Portfolio Summary**: In [`calculator.go`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go#L163-L166), the service assigns raw cumulative period return $(0.061055 - 1) \times 100 = \mathbf{-93.9\%}$ directly to the `AnnualizedReturnPercent` field without applying geometric CAGR annualization over the 1,250 elapsed days.

The objective of this implementation plan is to rectify all four layers, eliminating artificial write-offs, backfilling missing market data, calculating true GIPS-compliant CAGR annualization, and re-running the valuation pipeline to restore accurate performance analytics.

---

## 2. Technical Architecture & System Data Flow

```mermaid
flowchart TD
    subgraph MarketData ["Market Data Ingestion Pipeline"]
        YF["Yahoo Finance / Twelve Data Provider"]
        BF["IngestionService.BackfillInstrumentPrices<br/>Pulls 2023-05 to 2025-10 quotes"]
        DB_Prices[("portfolio.instrument_prices<br/>Complete daily close history")]
    end

    subgraph ValuationReplay ["Historical Valuation Pipeline"]
        TXS["portfolio.transactions (2023-05 to Present)"]
        PL["ProcessLedger<br/>Computes Holdings, CostBasis, Cash"]
        VAL["ValuationService.RebuildValuations"]
        FALLBACK{"Price in DB?"}
        P_MKT["Use Level 1 Market Close"]
        P_COST["Level 2/3 Fallback:<br/>CostBasis / Quantity (or CostBasisBase)"]
        RET["CalculateDailyReturn<br/>R_t = (V_t - F_t - V_t-1) / V_t-1"]
        TWR["ChainTWR<br/>TWR_t = TWR_t-1 * (1 + R_t)"]
        DB_VAL[("portfolio.portfolio_valuations<br/>Accurate TWR Index History")]
    end

    subgraph SummaryCalc ["Portfolio Summary Calculator"]
        LATEST["GetLatestValuation<br/>Retrieves latest twr_index"]
        EARLIEST["GetEarliestValuationDate<br/>Calculates elapsed days D"]
        GIPS["domain.CalculateAnnualizedReturn<br/>If D >= 365: ((TWR)^(365.25/D) - 1) * 100"]
        SUMMARY["PortfolioSummary.AnnualizedReturnPercent"]
    end

    subgraph Presentation ["Consumer Layers"]
        GRPC["gRPC PortfolioService (:50051)"]
        BFF["GraphQL BFF (:8080)"]
        UI["web/apps/main-app (:5173)<br/>Renders true positive CAGR"]
    end

    YF --> BF --> DB_Prices
    TXS --> PL
    PL --> VAL
    DB_Prices --> VAL
    VAL --> FALLBACK
    FALLBACK -->|Yes| P_MKT --> RET
    FALLBACK -->|No (Data Gap)| P_COST --> RET
    RET --> TWR --> DB_VAL

    DB_VAL --> LATEST --> GIPS
    DB_VAL --> EARLIEST --> GIPS
    GIPS --> SUMMARY --> GRPC --> BFF --> UI
```

---

## 3. Mathematical & Accounting Specifications

### 3.1 Valuation Fallback Hierarchy (US GAAP ASC 820 / GIPS Standard 2.A)
When computing the market value of holding $i$ on valuation date $t$:
$$V_{i,t} = Q_{i,t} \times P_{i,t} \times \text{FX}(C_i \to C_{\text{base}}, t)$$

The price $P_{i,t}$ is resolved via the following hierarchy:
1. **Level 1 (Quoted Market Price)**: The official closing price from `portfolio.instrument_prices` on date $t$.
2. **Level 2 (Last Observation Carried Forward - LOCF)**: If date $t$ is a weekend, exchange holiday, or missing price, carry forward the most recent previous close:
   $$P_{i,t} = P_{i, \tau} \quad \text{where } \tau = \max \{ d < t \mid \text{price exists} \}$$
3. **Level 3 (Acquisition Cost Basis Fallback)**: If date $t$ precedes all recorded market price quotes (e.g., historical transaction imported prior to quotes availability), the asset's fair value is its volume-weighted purchase cost basis:
   $$P_{i,t} = \frac{\text{CostBasis}_{i,t}}{Q_{i,t}} \quad \implies \quad V_{i,t} = \text{CostBasisBase}_{i,t}$$

#### Proof of Stability on Cash Inflow Days:
When an investor deposits cash $F_t$ and buys an unquoted asset:
- $\Delta \text{CashValue}_t = -F_t$
- $\Delta \text{MarketValue}_t = +F_t$
- $\Delta V_t = 0 \implies V_t = V_{t-1} + F_t$
- Daily return numerator:
  $$(V_t - F_t) - V_{t-1} = (V_{t-1} + F_t - F_t) - V_{t-1} = 0.00$$
- Resulting daily return: $R_t = 0.00\%$.
- The artificial $-573\%$ phantom loss is completely prevented.

---

### 3.2 GIPS-Compliant Annualized Return (CAGR)
In accordance with **GIPS Standard 2.A.20**:
- **Period $< 365$ calendar days**: Compounding is prohibited. The metric represents cumulative period return:
  $$R_{\text{period}} = (\text{TWRIndex} - 1) \times 100$$
- **Period $\ge 365$ calendar days**: Compound Annual Growth Rate (CAGR) using the standard Actual/365.25 day count convention:
  $$R_{\text{annualized}} = \left( (\text{TWRIndex})^{\frac{365.25}{D}} - 1 \right) \times 100$$
  Where $D = \text{ValuationDate} - \text{InceptionDate}$.

---

## 4. Phase-by-Phase Implementation Plan

### Phase 1: Valuation Engine Resiliency (Cost Basis Fallback)

#### 1.1 Update `internal/service/valuation.go`
In `RebuildValuations` and `SnapshotValuation`:
- When iterating through `holdings`:
  ```go
  price := decimal.Zero
  if prices, ok := priceMatrix[h.InstrumentID]; ok {
      price = prices[dateStr]
  }

  // Resilient Fallback (US GAAP ASC 820 Level 3):
  // If price is missing or zero, fallback to cost basis per share (or CostBasisBase directly)
  if price.IsZero() && h.CostBasis.IsPositive() && h.Quantity.IsPositive() {
      price = h.CostBasis.Div(h.Quantity)
  }
  ```
- If local currency requires conversion and local price was derived from `h.CostBasis`, compute `holdingVal` using `h.CostBasisBase`:
  ```go
  var holdingVal decimal.Decimal
  if price.IsZero() && h.CostBasisBase.IsPositive() {
      holdingVal = h.CostBasisBase
  } else {
      holdingVal = domain.CalculateHoldingMarketValue(h.Quantity, price, fx)
  }
  marketValBase = marketValBase.Add(holdingVal)
  ```

#### 1.2 Update Repository Matrix Queries (`internal/repository/postgres.go`)
In `GetHistoricalPriceMatrix`:
- If an instrument has observations but the matrix loop starts before the first observation (`date < obsList[0].date`), enable backwards LOCF fallback so the earliest known price is available backwards to the query start date as an alternative market estimate.

---

### Phase 2: Inception Tracking & True GIPS CAGR in `calculator.go`

#### 2.1 Add Inception Date / Earliest Valuation Query
In `internal/repository/repository.go` and `internal/repository/postgres.go`:
- Add `GetEarliestValuation(ctx context.Context, portfolioID uuid.UUID) (*domain.PortfolioValuation, error)`:
  ```sql
  SELECT portfolio_id, valuation_date, market_value_base, cash_value_base, net_flow_base, daily_return, twr_index
  FROM portfolio.portfolio_valuations
  WHERE portfolio_id = $1
  ORDER BY valuation_date ASC
  LIMIT 1;
  ```

#### 2.2 Wire Inception Date into `CalculatePortfolioSummary`
In `internal/service/service.go`:
- Call `s.repo.GetEarliestValuation(ctx, portfolio.ID)` alongside `GetLatestValuation`.
- Pass `earliestValuation` (or elapsed days) into `domain.CalculatePortfolioSummary`.

#### 2.3 Apply `domain.CalculateAnnualizedReturn` in `calculator.go`
In `internal/domain/calculator.go`:
```go
// Annualized Return / TWR from latest valuation snapshot using GIPS Standard 2.A.20
var annualizedReturnPercent decimal.Decimal
if latestValuation != nil && latestValuation.TWRIndex.IsPositive() {
    elapsedDays := 0
    if earliestValuation != nil {
        elapsedDays = int(latestValuation.ValuationDate.Sub(earliestValuation.ValuationDate).Hours() / 24)
    }
    
    annReturn, _, err := CalculateAnnualizedReturn(latestValuation.TWRIndex, elapsedDays)
    if err == nil {
        annualizedReturnPercent = annReturn.Round(1)
    } else {
        annualizedReturnPercent = latestValuation.TWRIndex.Sub(one).Mul(oneHundred).Round(1)
    }
}
```

---

### Phase 3: Historical Market Data Backfill Automation

#### 3.1 Historical Instrument Price Backfill
Provide an administrative command / worker trigger to backfill prices for all instruments in `portfolio.portfolios` holding active positions:
- Query all instruments held by portfolio `018f0000-0002-7000-8000-000000000001` with inception `2023-05-01`.
- Execute `IngestionService.BackfillInstrumentPrices` via Yahoo Finance provider from `2023-05-01` to `2025-10-10`.
- Populate `portfolio.instrument_prices` with real historical daily close quotes.

#### 3.2 Automated Ingestion Hook on Broker Import
In `internal/service/transaction.go` (`BatchImportTransactions`):
- When importing historical transactions whose trade dates precede the minimum price date of the instrument, automatically trigger or schedule a background historical price backfill for that instrument across `[earliestTradeDate, today]`.

---

### Phase 4: Rebuild Historical Valuations & Verification

#### 4.1 Trigger Rebuild
Invoke the valuation replay engine:
```bash
make run-valuation-job
```
Or via gRPC / worker CLI to replay all 1,250 daily points for portfolio `018f0000-0002-7000-8000-000000000001`.

#### 4.2 Verify Database Records
Execute verification query:
```sql
SELECT valuation_date, market_value_base, cash_value_base, net_flow_base, daily_return, twr_index 
FROM portfolio.portfolio_valuations 
WHERE portfolio_id = '018f0000-0002-7000-8000-000000000001' 
  AND valuation_date BETWEEN '2023-05-10' AND '2023-05-15'
ORDER BY valuation_date ASC;
```
**Verification Success Criterion**:
- On `2023-05-11`, $R_t$ must be within normal market daily movements (e.g., between `-2.0%` and `+2.0%`).
- `twr_index` remains strictly positive ($> 0.95$).
- Latest `twr_index` on `2026-10-11` reflects cumulative positive growth ($> 1.30$).
- Annualized return on dashboard reflects positive **`+12.0%` to `+18.0% p.a.`**.

---

## 5. Testing & Quality Assurance Plan

### 5.1 Unit Tests (`services/portfolio-api/internal/service/valuation_test.go`)
- **`TestRebuildValuations_MissingPrices_FallbackToCostBasis`**:
  - Setup: Portfolio with initial deposit and subsequent trade where instrument has zero rows in `instrument_prices`.
  - Expectation: Holding market value falls back to cost basis. Daily return does not plunge. TWR index remains stable ($1.000$).
- **`TestCalculatePortfolioSummary_GIPSAnnualization`**:
  - Test period $< 365$ days: returns cumulative period return with `isAnnualized = false`.
  - Test period $\ge 365$ days: returns geometric CAGR $((1+R)^{\frac{365.25}{D}} - 1) \times 100$.

### 5.2 Integration Tests
- Verify `GetEarliestValuation` repository query across empty and populated portfolio tables.
- Verify full ledger replay from CSV import through valuation snapshot generation.

---

## 6. Rollback & Contingency Strategy

- The valuation table `portfolio.portfolio_valuations` is a deterministic projection table derived from immutable transactions in `portfolio.transactions`.
- If an issue occurs during rebuild, `s.repo.DeleteValuationsFromDate` allows re-running the valuation engine cleanly from inception at any time without data loss.
- All database operations are wrapped in PostgreSQL transactions or idempotent upsert batches (`ON CONFLICT (portfolio_id, valuation_date) DO UPDATE`).
