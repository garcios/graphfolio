# User Story: Cash Flow Report Brokerage Fee & Tax Attribution

> **Story ID**: US-REP-102  
> **Epic**: Portfolio Cash Flow & Reporting Hub  
> **Target**: Domain Microservice (`services/portfolio-api`), BFF GraphQL Orchestrator (`bff/`), Primary Investor Application (`web/apps/main-app`)  
> **Status**: **DONE**  
> **Implementation Plan**: [cash-flow-fee-and-tax-attribution-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cash-flow-fee-and-tax-attribution-implementation-plan.md)  
> **Related Stories**: [portfolio-cash-flow-report-generation.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/portfolio-cash-flow-report-generation.md)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** investor reviewing my portfolio cash flow statement,  
> **I want to** see trading commissions (brokerage fees) and dividend withholding taxes explicitly attributed to their dedicated outflow categories ("Brokerage & Custody Fees" and "Taxes & Withholding"),  
> **So that** I have full visibility into my investment frictions and drag without having those expenses obscured inside asset purchases, sale proceeds, or gross dividend payouts.

---

## 2. Business Context & Financial Reconciliation

1. **Category Isolation & Frictional Drag Visibility**:
   - In retail investing, transactions incur trade execution commissions (e.g. $9.95 / $19.95 per order on CommSec / IBKR) and foreign withholding taxes (e.g. 15% or 30% US WHT on dividend payments).
   - In previous iterations, trade fees were bundled into gross `BUY` purchases and deducted from gross `SELL` proceeds, while withholding taxes were deducted directly from dividend inflows. Consequently, the "Brokerage & Custody Fees" and "Taxes & Withholding" summary cards remained perpetually at `$0.00` (0%).
2. **Mathematical Attribution Matrix**:
   - `BUY`: Outflows attribute asset consideration to `Purchases` and trading fee to `Brokerage Fees`.
   - `SELL`: Inflows attribute gross sale consideration to `Sale Proceeds`, and trading fee attributes to `Brokerage Fees` outflow.
   - `DIVIDEND`: Inflows attribute gross dividend amount to `Dividends`, and withholding tax attributes to `Taxes & Withholding` outflow.
3. **100% Mathematical Reconciliation Identity**:
   $$\text{Ending Cash Balance} = \text{Starting Cash Balance} + \text{Total Inflows} - \text{Total Outflows}$$
   where:
   $$\text{Total Inflows} = \text{Deposits} + \text{Dividends}_{\text{gross}} + \text{Interest} + \text{Sales Proceeds}_{\text{gross}}$$
   $$\text{Total Outflows} = \text{Withdrawals} + \text{Purchases}_{\text{gross}} + \text{Brokerage Fees} + \text{Taxes \& Withholding}$$

---

## 3. Acceptance Criteria

### 3.1 Fee & Tax Decoupling
- [x] **AC-1 (Trade Fee Outflows)**: For all `BUY` and `SELL` transactions in the reporting period with a non-zero `fee`, the fee amount (converted to portfolio base currency via `fx_rate_to_base`) is accumulated into the `Brokerage & Custody Fees` outflow category.
- [x] **AC-2 (Gross Purchase Consideration)**: For `BUY` transactions, the `Purchases` category reflects pure gross asset consideration ($\text{Quantity} \times \text{Price} \times \text{FX}$), preventing fee double-counting.
- [x] **AC-3 (Gross Sale Proceeds)**: For `SELL` transactions, `Sale Proceeds` reflects gross market value proceeds, with the commission captured under `Brokerage & Custody Fees`.
- [x] **AC-4 (Dividend Withholding Tax Attribution)**: For `DIVIDEND` transactions with a non-zero `withholding_tax`, the gross dividend is credited to `Dividends` inflow, and the withholding tax is attributed to `Taxes & Withholding` outflow.

### 3.2 Reconciliation & Presentation
- [x] **AC-5 (Mathematical Reconciliation Bar)**: The cash flow reconciliation strip continues to verify that $\Delta\text{Cash} = \text{Inflows} - \text{Outflows}$ equals the difference between starting and ending cash balances with zero precision error.
- [x] **AC-6 (Summary Cards & Category Breakdown)**: In `CashFlowReport.tsx`, the "Brokerage & Custody Fees" and "Taxes & Withholding" metric cards and breakdown tables accurately populate non-zero amounts and relative percentages.
- [x] **AC-7 (CSV Export Integrity)**: Cash Flow Report CSV export itemizes fee and tax line items with accurate category assignments.

---

## 4. Technical & Implementation Notes

- **Domain Model**: `services/portfolio-api/internal/domain/cash_flow.go` defines `CalculateTransactionCashImpact` decoupling fee and tax cash components.
- **Service Aggregator**: `services/portfolio-api/internal/service/cash_flow.go` aggregates separate inflow and outflow buckets during timeframe iterations.
- **Unit Tests**: `cash_flow_test.go` verifies multi-currency fee conversions, dividend tax deductions, and reconciliation integrity.
