# User Story: GIPS-Compliant Annualized Return (TWR) & Valuation Cost-Basis Resiliency

> **Story ID**: US-VAL-102  
> **Epic**: Portfolio Valuation & Performance Analytics  
> **Target**: Domain Microservice (`services/portfolio-api`), Repository (`internal/repository/`), Valuation Engine (`internal/service/valuation.go`), Calculation Engine (`internal/domain/calculator.go`), BFF (`bff/`), Investor UI (`web/apps/main-app`)  
> **Status**: **DONE**  
> **Implementation Plan**: [annualized-return-twr-fix-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/annualized-return-twr-fix-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #28)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L38)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md) (GIPS Standard 2.A.20, US GAAP ASC 820 Fair Value Hierarchy)  

---

## 1. User Story Statement

> **As an** investor with a multi-year investment history,  
> **I want to** have my Annualized Time-Weighted Return (TWR) calculated with strict GIPS CAGR annualization and resilient valuation mark fallbacks for unpriced assets,  
> **So that** my performance metrics reflect true economic compounding and eliminate catastrophic negative return distortions caused by missing historical quotes on past trade dates.

---

## 2. Business Context & Financial Defect Remediation

1. **Defect Remediation**:
   - Previously, transactions dating back years imported into GraphFolio encountered periods where daily closing quotes in `portfolio.instrument_prices` were missing.
   - The valuation pipeline evaluated unpriced assets at `$0.00`, triggering phantom daily returns (e.g. $-573\%$) on trade entry days that dragged the geometric TWR compounding chain permanently down to near-zero ($0.061$, or $-93.9\%$), despite the investor possessing massive real portfolio gains ($+60\%$).
   - Furthermore, raw cumulative multi-year return percentages were assigned directly to the `AnnualizedReturnPercent` field without geometric CAGR time annualization.
2. **US GAAP ASC 820 Fair Value Fallback Hierarchy**:
   - Level 1: Authoritative exchange daily closing price.
   - Level 2/3 Fallback: When no market quote is available for an open position on date $t$, evaluate the position using its known volume-weighted average cost basis ($\text{CostBasis}_{\text{base}} / \text{Quantity}$) or acquisition cost basis, preventing artificial asset write-downs.
3. **GIPS Standard 2.A.20 Annualization Rule**:
   - For investment periods $\ge 365$ days:
     $$\text{CAGR} = \left( (1 + R_{\text{cum}})^{\frac{365.25}{D}} - 1 \right) \times 100$$
     where $D$ is the elapsed day count from inception to evaluation date.
   - For investment periods $< 365$ days: GIPS forbids annualizing sub-annual returns; display raw cumulative period return.

---

## 3. Acceptance Criteria

### 3.1 Valuation Fallback & Zero-Price Prevention
- [x] **AC-1 (Fair Value Fallback Hierarchy)**: In `valuation.go`, if an instrument has no market price on date $t$, the valuation engine falls back to the position's cost basis per share ($\text{CostBasis}_{\text{base}}$), ensuring holding value $> 0$.
- [x] **AC-2 (Cash & Capital Flow Alignment)**: On asset acquisition dates where cash is debited and shares are received, net portfolio market value reconciles without generating phantom single-day drawdowns.
- [x] **AC-3 (Historical Price Backfilling)**: External ingestion providers (Yahoo Finance / Twelve Data) support backfilling historical closing prices covering the entire trade history span.

### 3.2 GIPS CAGR Annualization Engine
- [x] **AC-4 (Geometric Time Annualization)**: In `domain/calculator.go`, `CalculateAnnualizedReturn` calculates exact CAGR when the elapsed holding duration is $\ge 365$ days using the GIPS 365.25-day convention.
- [x] **AC-5 (Sub-Annual Guard)**: If the investment history spans fewer than 365 days, annualization is suppressed and cumulative return is presented.
- [x] **AC-6 (Geometric Compounding Continuity)**: Cumulative TWR index values stored in `portfolio.portfolio_valuations` compound strictly without arithmetic distortion.

### 3.3 Revaluation & Presentation
- [x] **AC-7 (Valuation Replay)**: Calling `RebuildValuations` reconstructs daily snapshots from the portfolio's inception date with cost-basis fallbacks active.
- [x] **AC-8 (Dashboard Metric Presentation)**: Primary dashboard hero displays true positive annualized return percentage matching real investor economics.

---

## 4. Technical & Implementation Notes

- **Investigation**: Documented in `docs/annualized-twr-return-investigation-report.md`.
- **Valuation Logic**: `services/portfolio-api/internal/service/valuation.go` implements cost basis fallback when `price == 0`.
- **Domain Math**: `services/portfolio-api/internal/domain/calculator.go` implements `CalculateAnnualizedReturn`.
- **Worker & Script**: `cmd/worker/main.go` / `make run-valuation-job` re-evaluates all active portfolios.
