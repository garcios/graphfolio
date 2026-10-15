# User Story: Investment Preferred Currency Conversion in Holdings Table

> **Story ID**: US-INV-105  
> **Epic**: Portfolio Holdings & Multi-Currency Presentation  
> **Target**: Domain Microservice (`services/portfolio-api`), Repository (`internal/repository/queries.go`), Calculation Engine (`internal/domain/calculator.go`), Investor Web Application (`web/apps/main-app`)  
> **Status**: **DONE**  
> **Implementation Plan**: [investment-currency-conversion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/investment-currency-conversion-implementation-plan.md)  
> **Related Stories**: [user-preferences-and-display-currency.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/user-preferences-and-display-currency.md), [investment-holding-capital-gain-income-currency.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/investment-holding-capital-gain-income-currency.md)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** international investor with a preferred display currency (e.g. AUD or EUR),  
> **I want to** have each holding's "Total Value" and "Today's Return" displayed in my portfolio's base currency while retaining the asset's traded unit price in its native exchange currency,  
> **So that** holding valuations and daily returns sum consistently to the hero metrics shown at the top of the dashboard without currency confusion.

---

## 2. Business Context & Currency Conversion Rules

1. **Elimination of Visual Currency Inconsistencies**:
   - When an investor sets their preferred currency to Australian Dollars (`AUD`) or Euros (`EUR`), the dashboard hero cards (Total Portfolio Value, Today's Return, Cash Balance) correctly calculate and format in that preferred currency.
   - Previously, in the "Your Investments" table, "Total Value" and "Today's Return" were formatted in the asset's listing currency (e.g. USD for Apple or Microsoft), while "Total Return" was formatted in the preferred currency (e.g. AUD).
   - This caused holding totals not to sum to the portfolio value, creating confusion.
2. **Dual-Currency Presentation Principle**:
   - **Traded Asset Unit Price**: Retained in the native instrument currency (e.g. `$228.45 USD`) to preserve real exchange quotation context.
   - **Total Value**: Converted to the portfolio base currency using the latest spot FX rate ($\text{TotalValue}_{\text{base}} = \text{Quantity} \times \text{Price}_{\text{local}} \times \text{FX}_{\text{local}\to\text{base}}$).
   - **Today's Return**: Converted to the portfolio base currency ($\Delta\text{Price} \times \text{Quantity} \times \text{FX}_{\text{local}\to\text{base}}$).
   - **Cost Basis & Returns**: Maintained in portfolio base currency to evaluate true investor returns.

---

## 3. Acceptance Criteria

### 3.1 Domain Calculation Engine
- [x] **AC-1 (Total Value in Base Currency)**: In `CalculateInvestment`, `TotalValue` is constructed as `NewMoney(totalValueBase, baseCurrency)`, converting foreign holdings using the latest closing foreign exchange rate.
- [x] **AC-2 (Today's Return in Base Currency)**: `TodayReturnAmount` is converted to base currency via `todayReturnAmountInst.Mul(fxRate)`, tagged with the portfolio's base currency code.
- [x] **AC-3 (Traded Price Preserved in Native Currency)**: Traded `Price` remains tagged with the instrument's native currency (e.g. `USD` for US equities, `AUD` for ASX equities).
- [x] **AC-4 (Domestic Instrument Identity)**: If the instrument currency equals the portfolio base currency, the FX rate is exactly `1.0000000000`, producing zero currency conversion discrepancy.

### 3.2 SQL Repository & FX Rates Joining
- [x] **AC-5 (Latest FX Rate Join)**: SQL query `GetHoldingsWithMarketData` joins the latest rate from `portfolio.fx_rates` for the pair `(instrument_currency, portfolio_base_currency)` or computes reciprocal rates where appropriate.

### 3.3 Dashboard Holdings Table Presentation
- [x] **AC-6 (Visual Currency Alignment)**: In `Dashboard.tsx`, the "Total Value" and "Today's Return" columns render the base currency symbol (e.g. `A$` or `€`), aligning with portfolio totals.
- [x] **AC-7 (Summability)**: The sum of all holding "Total Value" rows plus the cash balance matches the hero "Portfolio Value" card.

---

## 4. Technical & Implementation Notes

- **Domain Calculator**: `services/portfolio-api/internal/domain/calculator.go` calculates `totalValueBase` and `todayReturnAmountBase` using `shopspring/decimal`.
- **Repository**: `services/portfolio-api/internal/repository/queries.go` (`GetHoldingsWithMarketData`) joins `portfolio.fx_rates`.
- **Unit Tests**: `calculator_test.go` verifies multi-currency holdings calculations across AUD, USD, and EUR portfolios.
