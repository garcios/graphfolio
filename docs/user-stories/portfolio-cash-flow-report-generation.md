# User Story: Portfolio Cash Flow Report Generation

## 1. Story Statement
**As an** investor tracking my stock portfolio,  
**I want to** generate and view a comprehensive Cash Flow Report covering all inflows and outflows (deposits, withdrawals, dividends, interest, and trade settlements),  
**So that** I can accurately understand my portfolio's net cash movement, monitor passive income, and reconcile my uninvested cash balance over any selected timeframe.

---

## 2. Scope & Cash Event Taxonomy

| Flow Direction | Event Type | Description |
| :--- | :--- | :--- |
| **Inflow (+)** | **Deposit** | External cash transferred into the portfolio account. |
| **Inflow (+)** | **Dividend Received** | Cash dividend payouts from equity/ETF holdings (gross and net of withholding tax). |
| **Inflow (+)** | **Interest Earned** | Cash sweep yields or interest earned on uninvested cash balances. |
| **Inflow (+)** | **Stock Sale Proceeds** | Net cash received from selling securities (post-brokerage/fees). |
| **Outflow (-)** | **Withdrawal** | Cash transferred out to external bank accounts. |
| **Outflow (-)** | **Stock Purchase** | Cash deducted for buying securities (including trade fees/commissions). |
| **Outflow (-)** | **Fees & Taxes** | Custody fees, subscription fees, margin loan interest, or withholding tax. |

---

## 3. Acceptance Criteria

### AC 1: Filter & Date Range Selection
- **Scenario: Choosing the reporting period**
  - **Given** I am on the Cash Flow Report page,
  - **When** I select a date range (e.g., *MTD, YTD, Trailing 12 Months, Custom Date Range*),
  - **Then** the report updates to only display cash events where `settlement_date` falls within the selected window.

- **Scenario: Multi-currency / Multi-account filtering**
  - **Given** I hold multi-currency cash or multiple broker accounts,
  - **When** I apply an account or currency filter,
  - **Then** the report aggregates values in the selected base currency with appropriate FX rates at event time.

### AC 2: Summary KPI Cards
- **Scenario: Viewing high-level aggregates**
  - **Given** a filtered date range is applied,
  - **Then** the system displays the following key metrics at the top of the report:
    - **Starting Cash Balance** (as of day 0 of the window)
    - **Total Inflows** (sum of deposits, dividends, interest, sales)
    - **Total Outflows** (sum of withdrawals, buys, fees)
    - **Net Cash Flow** ($\text{Total Inflows} - \text{Total Outflows}$)
    - **Ending Cash Balance** ($\text{Starting Balance} + \text{Net Cash Flow}$)

### AC 3: Breakdown by Category
- **Scenario: Inspecting sub-totals**
  - **Given** the report is loaded,
  - **Then** the system presents grouped category totals (e.g., Total Dividends received, Total Capital Deposited, Total Brokerage Fees incurred) alongside visual indicators (e.g., green for inflows, red/neutral for outflows).

### AC 4: Itemized Transaction Ledger
- **Scenario: Reviewing granular transactions**
  - **Given** the summary view,
  - **When** I inspect the detailed transaction table,
  - **Then** each row displays:
    1. `Date` (Settlement date)
    2. `Type` (Deposit, Withdrawal, Dividend, Interest, Buy, Sell, Fee)
    3. `Asset / Ticker` (if applicable, e.g., AAPL for a dividend, or N/A for cash interest)
    4. `Description` (e.g., "Quarterly Dividend - 50 shares @ $0.25")
    5. `Amount` (positive for inflow, negative for outflow)
    6. `Running Cash Balance`
  - **And** I can sort the table by date, amount, or transaction category.

### AC 5: Export Functionality
- **Scenario: Exporting report data**
  - **Given** the generated report,
  - **When** I click the "Export" button,
  - **Then** I can download the dataset as CSV or PDF containing both the summary totals and the itemized transaction list.

---

## 4. Non-Functional Requirements & Edge Cases

* **Negative Cash Balances:** If margin trading or overdraft is supported, negative running cash balances must be handled without breaking reconciliation equations.
* **Pending vs. Settled Transactions:** Inflows like dividend declarations with future record dates must not appear in realized cash flow until the settlement/payable date.
* **Dividend Reinvestment Plans (DRIP):** If a dividend is automatically reinvested, it must be represented as a two-leg transaction (Dividend Inflow + Stock Purchase Outflow) to keep cash balance true while accurately recording passive income.
