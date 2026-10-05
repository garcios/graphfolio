---
name: financial-calculations
description: "Industry-standard financial and accounting formulas, business rules, and calculation engines conforming to GIPS (Global Investment Performance Standards), CFA Institute, US GAAP, and IFRS. Covers Annualized Return (CAGR), Time-Weighted Return (TWR), Money-Weighted Return (MWR / IRR / XIRR), Modified Dietz, Cost Basis relief (FIFO, Average Cost, Specific Lot), Realized/Unrealized Gain/Loss, Dividends, Corporate Actions, and Exact Fixed-Point Precision."
risk: safe
source: "GraphFolio Architecture & Finance Standards"
---

# Financial & Accounting Calculation Standards

This skill governs the mathematical formulation, accounting business rules, edge-case handling, and precision standards for all financial calculations within GraphFolio. 

Every agent and developer implementing calculations for portfolio valuation, return metrics, tax lot relief, or performance tracking must strictly follow this guide to remain compliant with **GIPS® (Global Investment Performance Standards by CFA Institute)**, **US GAAP (ASC 320 / ASC 820)**, and **IFRS (IFRS 9)**.

---

## 1. Regulatory & Industry Standards Framework

### 1.1 GIPS® (Global Investment Performance Standards) Compliance
GraphFolio calculations must follow the CFA Institute GIPS standards for calculating and presenting investment performance:
1. **The Sub-1-Year Annualization Prohibition (GIPS Standard 2.A.20)**:
   > **Returns for periods of less than one full year (< 365 calendar days) MUST NOT be annualized.**
   - *Rationale*: Annualizing short periods (e.g., compounding a 1-month gain of 3% into $(1.03)^{12} - 1 = 42.6\%$) produces deceptive, statistically invalid volatility exaggerations.
   - *Rule*: If elapsed days $D < 365$, the system **must** return the non-annualized cumulative period return and label it as "Period Return" or "Cumulative Return", or return an explicit indicator that annualization is invalid.
2. **Time-Weighted Return (TWR) as Primary Performance Metric**:
   - TWR isolates investment decisions (asset allocation, security selection) from investor cash flow timing and sizing.
   - External cash flows (deposits and withdrawals) must not distort performance returns.
3. **Daily Valuation**:
   - Valuations must be calculated at least daily at market close, or on dates of large external cash flows.
4. **Gross vs. Net of Fees**:
   - GraphFolio reports net performance after trading commissions and transaction fees.

### 1.2 US GAAP & IFRS Accounting Standards
1. **Trade Date Accounting (US GAAP ASC 946 / IFRS 9)**:
   - Security transactions are recognized on the **Trade Date** (execution date), not the settlement date (T+1/T+2).
   - Valuations, cash deductions, and lot creations occur on trade date $t$.
2. **Fair Value Measurement (ASC 820 / IFRS 13)**:
   - Valuations are based on official Level 1 quoted market closing prices on active exchanges.
   - For non-trading days (weekends, holidays), use **Last Observation Carried Forward (LOCF)**.
3. **Matching Principle & Realization Principle**:
   - Capital gains and losses are only *realized* upon an actual disposal event (sale, merger cash-out).
   - Unrealized gains/losses remain mark-to-market adjustments on open balance sheets.

### 1.3 Strict Zero Floating-Point Arithmetic Policy
- **Never use `float32` or `float64`** for currency, quantities, prices, or percentages. IEEE 754 floating-point introduces binary representation errors (e.g. `0.1 + 0.2 = 0.30000000000000004`).
- All calculations must use arbitrary-precision fixed-point math (`github.com/shopspring/decimal` in Go, `Decimal` scalar in GraphQL).
- **Scale Requirements**:
  - Currency Balances & Valuations: Round to **2 decimal places** for presentation; maintain full precision during intermediary steps.
  - Asset Quantities: Up to **8 decimal places** (supporting fractional shares and crypto).
  - Security Prices & FX Rates: At least **6 to 8 decimal places**.
  - Performance Indices (TWR Index): **12 decimal places**.
  - Return Percentages: Calculated to **6 decimal places**, rounded to **2 decimal places** for display (e.g., `14.25%`).

---

## 2. Portfolio Valuation & Cash Flow Categorization

### 2.1 Valuation Formulations
At valuation date $t$ for a portfolio with base currency $C_{\text{base}}$:

$$\text{TotalValue}_t = \text{MarketValue}_t + \text{CashValue}_t$$

$$\text{MarketValue}_t = \sum_{i \in \text{Holdings}_t} \left( Q_{i,t} \times P_{i,t} \times \text{FX}(C_i \to C_{\text{base}}, t) \right)$$

$$\text{CashValue}_t = \sum_{c \in \text{Currencies}} \left( B_{c,t} \times \text{FX}(c \to C_{\text{base}}, t) \right)$$

- $Q_{i,t}$: Quantity of instrument $i$ held at close of date $t$.
- $P_{i,t}$: Official market close price of instrument $i$ on date $t$.
- $B_{c,t}$: Cash balance in currency $c$ at close of date $t$.
- $\text{FX}(c \to C_{\text{base}}, t)$: Exchange rate converting currency $c$ to base currency.

### 2.2 Cash Flow Classification (External vs. Internal)
Correct classification of cash flows is essential to avoid corrupting performance metrics:

| Transaction Type | Flow Classification | Impact on Portfolio Total Value | Impact on TWR Return ($R_t$) |
|---|---|---|---|
| **DEPOSIT** | **External Flow ($+F_t$)** | Increases total capital | Adjusted out (no return impact) |
| **WITHDRAWAL** | **External Flow ($-F_t$)** | Decreases total capital | Adjusted out (no return impact) |
| **BUY** | **Internal Flow** | Neutral (Cash $\to$ Asset) | None (trading fee impacts PnL) |
| **SELL** | **Internal Flow** | Neutral (Asset $\to$ Cash) | None (trading fee impacts PnL) |
| **DIVIDEND (Cash)** | **Internal Flow** | Neutral (Asset drop $\to$ Cash gain) | Retained as portfolio income |
| **INTEREST INCOME** | **Internal Flow** | Increases cash from yield | Retained as portfolio income |
| **MANAGEMENT / CUSTODY FEE**| **Internal Outflow** | Decreases cash | Decreases net return |

$$\text{NetExternalFlow}_t (F_t) = \sum \text{Deposits}_t - \sum \text{Withdrawals}_t$$

---

## 3. Return Metric Formulas

### 3.1 Holding Period Return (Simple Return)
Used for single instruments or static periods with **zero external cash flows**:

$$R = \frac{V_{\text{end}} - V_{\text{start}}}{V_{\text{start}}}$$

$$\text{ReturnPercentage} = R \times 100$$

> ⚠️ **Limitation**: Cannot be used for an entire portfolio across multiple periods if deposits or withdrawals occur, because cash flows distort the denominator.

---

### 3.2 Time-Weighted Return (TWR) — Daily Valuation Method
The **gold standard** under GIPS for measuring investment management performance.

#### Daily Sub-Period Return ($R_t$):
Assuming external flows $F_t$ occur at the **end of day**:

$$R_t = \frac{V_t - F_t - V_{t-1}}{V_{t-1}}$$

Assuming external flows $F_t$ occur at **mid-day** (or day-weighted):

$$R_t = \frac{V_t - V_{t-1} - F_t}{V_{t-1} + 0.5 \times F_t}$$

*GraphFolio Default*: End-of-day flow convention: $R_t = \frac{(V_t - F_t) - V_{t-1}}{V_{t-1}}$.

#### Compound Cumulative TWR ($\text{TWR}_{0 \to N}$):
The cumulative return over $N$ trading periods:

$$\text{TWR} = \prod_{t=1}^N (1 + R_t) - 1$$

#### Unitized TWR Index ($\text{TWRIndex}_t$):
To allow fast $O(1)$ interval return queries, GraphFolio maintains a daily cumulative index:

$$\text{TWRIndex}_0 = 1.000000000000$$

$$\text{TWRIndex}_t = \text{TWRIndex}_{t-1} \times (1 + R_t)$$

The cumulative return between any date $A$ and date $B$ ($B > A$) is then:

$$\text{TWR}_{A \to B} = \frac{\text{TWRIndex}_B}{\text{TWRIndex}_A} - 1$$

#### Edge Cases & Business Rules:
- **Zero Previous Valuation ($V_{t-1} = 0$)**:
  - Occurs on initial portfolio funding: $V_{t-1} = 0$, deposit $F_t = \$10,000$, $V_t = \$10,000$.
  - Formula yields $0/0$. **Rule: $R_t = 0$**.
- **Full Liquidation / Zero Balance ($V_t = 0$)**:
  - If entire account is withdrawn: $V_t = 0$, $F_t = -\$10,000$.
  - Formula yields $\frac{0 - (-10000) - 10000}{10000} = 0$. Return is $0$.
- **Negative Valuation (Margin/Borrowing)**:
  - If $V_{t-1} \le 0$, percentage returns become undefined or mathematically inverted. Flag account as margin-constrained; freeze TWR compounding until equity becomes positive.

---

### 3.3 Annualized Return (Compound Annual Growth Rate / CAGR)

#### Mathematical Formula:
Given cumulative return $R_{\text{cum}}$ over $D$ calendar days ($D \ge 365$):

$$R_{\text{annualized}} = \left( (1 + R_{\text{cum}})^{\frac{365.25}{D}} - 1 \right) \times 100$$

Or using the cumulative $\text{TWRIndex}$:

$$R_{\text{annualized}} = \left( \left(\frac{\text{TWRIndex}_{\text{end}}}{\text{TWRIndex}_{\text{start}}}\right)^{\frac{365.25}{D}} - 1 \right) \times 100$$

*Day Count Convention*: **Actual/365.25** (365.25 days per year accounts for leap year cadence; Actual/365 is also permitted if specified in portfolio settings).

#### Strict GIPS Annualization Business Rules:
```text
IF ElapsedDays < 365:
    AnnualizedReturn is PROHIBITED by GIPS Standard 2.A.20.
    Output: Cumulative Period Return (non-annualized)
    UI Display: "14.2% (Period Return)" or "N/A (< 1 Year)"
ELSE:
    AnnualizedReturn = ((1 + CumulativeReturn) ^ (365.25 / ElapsedDays) - 1) * 100
    UI Display: "14.2% p.a."
```

---

### 3.4 Money-Weighted Return (MWR / Internal Rate of Return / XIRR)
Measures the investor's personal bottom-line return, reflecting both asset performance and the investor's market timing of deposits/withdrawals.

#### Mathematical Equation:
Find discount rate $r$ (annualized internal rate of return) such that the Net Present Value (NPV) of all portfolio cash flows equals zero:

$$\sum_{i=0}^N \frac{C_i}{(1 + r)^{\frac{d_i - d_0}{365.25}}} = 0$$

Where:
- $d_0$: Portfolio start date.
- $d_i$: Date of transaction $i$.
- $C_0 = -V_{\text{start}}$ (initial portfolio value treated as cash outflow).
- $C_i = -F_i$ (deposits are negative cash outflows; withdrawals are positive inflows to investor).
- $C_N = +V_{\text{end}}$ (final portfolio value treated as terminal liquidation inflow).

#### Numerical Solution Algorithm:
The equation has no closed-form algebraic solution and must be solved iteratively:
1. **Newton-Raphson Method**:
   $$r_{k+1} = r_k - \frac{\text{NPV}(r_k)}{\text{NPV}'(r_k)}$$
   $$\text{NPV}'(r) = \sum_{i=0}^N -\frac{d_i - d_0}{365.25} \cdot C_i \cdot (1 + r)^{-\left(\frac{d_i - d_0}{365.25} + 1\right)}$$
2. **Convergence Criteria**: Stop when $|\text{NPV}(r)| < 10^{-6}$ or max iterations (100) reached.
3. **Fallback**: If Newton-Raphson diverges (e.g., extreme cash flows or non-monotonic derivative), fall back to **Bisection Method** on interval $[-0.9999, 10.0]$.

---

### 3.5 Modified Dietz Method
Standard approximation when daily valuations are not accessible (e.g. monthly valuation periods with mid-month flows):

$$R_{\text{Dietz}} = \frac{V_{\text{end}} - V_{\text{start}} - F}{V_{\text{start}} + \sum_{i=1}^M (W_i \times F_i)}$$

Where:
- $F = \sum_{i=1}^M F_i$ (sum of all net external cash flows in the period).
- $CD$: Total number of calendar days in the period.
- $D_i$: Number of calendar days from period start until cash flow $i$ occurred.
- $W_i = \frac{CD - D_i}{CD}$: Time-weight factor of cash flow $i$ ($0 \le W_i \le 1$).

---

## 4. Cost Basis & Capital Gains (Accounting Standards)

### 4.1 Cost Basis Allocation Methods

#### 1. FIFO (First-In, First-Out)
- **Standard**: US IRS (IRC §1012), US GAAP.
- **Rule**: Shares sold are matched against the earliest acquired open tax lots first.
- **Formulation**:
  When selling quantity $Q_{\text{sell}}$:
  Loop through open lots sorted by `acquisition_date ASC`:
  $$\Delta Q = \min(Q_{\text{lot}}, Q_{\text{remaining\_sell}})$$
  $$\text{RelievedCost} = \Delta Q \times \text{UnitCost}_{\text{lot}}$$

#### 2. Average Cost Method (Weighted Average Cost / WAC)
- **Standard**: Mutual funds (IRC Reg 1.1012-1(e)), UK Section 104 Holding Pools, Canadian ACB (Adjusted Cost Base).
- **Rule**: All shares share a single pooled average cost basis.
- **Formulation**:
  Upon each purchase:
  $$\text{PoolQty}_{\text{new}} = \text{PoolQty}_{\text{old}} + Q_{\text{buy}}$$
  $$\text{PoolCost}_{\text{new}} = \text{PoolCost}_{\text{old}} + (Q_{\text{buy}} \times P_{\text{buy}}) + \text{Fees}_{\text{buy}}$$
  $$\text{AvgUnitCost} = \frac{\text{PoolCost}_{\text{new}}}{\text{PoolQty}_{\text{new}}}$$
  Upon each sale:
  $$\text{RelievedCost} = Q_{\text{sell}} \times \text{AvgUnitCost}$$
  $$\text{PoolCost}_{\text{remaining}} = \text{PoolCost}_{\text{old}} - \text{RelievedCost}$$

---

### 4.2 Realized vs. Unrealized Gain/Loss

#### Realized Gain/Loss:
Triggered strictly upon execution of a disposal (SELL):

$$\text{GrossProceeds} = Q_{\text{sell}} \times P_{\text{sell}}$$

$$\text{NetProceeds} = \text{GrossProceeds} - \text{SellingFees}$$

$$\text{RealizedPnL} = \text{NetProceeds} - \text{RelievedCostBasis}$$

- If $\text{RealizedPnL} > 0$: Capital Gain.
- If $\text{RealizedPnL} < 0$: Capital Loss.

#### Unrealized Gain/Loss (Paper Return):
Mark-to-market valuation on open inventory at current price $P_{\text{current}}$:

$$\text{CurrentMarketValue} = Q_{\text{remaining}} \times P_{\text{current}} \times \text{FX}$$

$$\text{UnrealizedPnL} = \text{CurrentMarketValue} - \text{RemainingCostBasis}$$

$$\text{UnrealizedReturnPercent} = \frac{\text{UnrealizedPnL}}{\text{RemainingCostBasis}} \times 100$$

---

### 4.3 Total Return Calculation
Total return on an instrument or portfolio must reflect all components of economic value:

$$\text{TotalReturnAmount} = \text{UnrealizedPnL} + \text{RealizedPnL} + \text{CumulativeDividends}$$

$$\text{TotalReturnPercent} = \frac{\text{TotalReturnAmount}}{\text{CumulativeCapitalInvested}} \times 100$$

---

## 5. Corporate Actions Business Rules

### 5.1 Stock Splits & Reverse Splits
A stock split changes the number of shares without altering the total capital invested or economic value:

- **Split Ratio**: $A : B$ (e.g. 2:1 forward split: $A = 2, B = 1$; 1:10 reverse split: $A = 1, B = 10$).
- **Tax Lot Adjustment**:
  $$Q_{\text{new}} = Q_{\text{old}} \times \frac{A}{B}$$
  $$\text{UnitCost}_{\text{new}} = \text{UnitCost}_{\text{old}} \times \frac{B}{A}$$
- **Invariant**:
  $$Q_{\text{new}} \times \text{UnitCost}_{\text{new}} \equiv Q_{\text{old}} \times \text{UnitCost}_{\text{old}} = \text{TotalLotCostBasis}$$
- **Capital Gain Event**: A stock split is **never** a taxable realization event under US GAAP / IFRS. Realized PnL is zero.

### 5.2 Cash Dividends
- Under US GAAP, receiving a cash dividend does **not** adjust the cost basis of the underlying stock.
- The cash is recognized as investment income and increases portfolio cash balance.
- Total portfolio valuation is unchanged at the instant of payment (stock price drops by dividend amount on ex-dividend date; cash increases by dividend amount on pay date).

---

## 6. Currency Fluctuations & FX Decomposition

For investments held in a foreign currency $C_{\text{inst}}$ with portfolio base currency $C_{\text{base}}$:

### 6.1 Total Return Multi-Currency Decomposition
Total return in base currency is the product of asset price movement and currency movement:

$$(1 + R_{\text{base}}) = (1 + R_{\text{asset}}) \times (1 + R_{\text{FX}})$$

$$R_{\text{base}} = R_{\text{asset}} + R_{\text{FX}} + (R_{\text{asset}} \times R_{\text{FX}})$$

Where:
- $R_{\text{asset}} = \frac{P_{t, \text{inst}} - P_{0, \text{inst}}}{P_{0, \text{inst}}}$ (Asset return in local currency)
- $R_{\text{FX}} = \frac{\text{FX}_t - \text{FX}_0}{\text{FX}_0}$ (Currency movement relative to base)
- $(R_{\text{asset}} \times R_{\text{FX}})$: Cross-product interaction term.

---

## 7. Risk-Adjusted Metrics

### 7.1 Standard Deviation (Annualized Volatility)
Given a sample of $N$ daily returns $R_1, R_2, \dots, R_N$:

$$\bar{R} = \frac{1}{N} \sum_{t=1}^N R_t$$

$$s_{\text{daily}} = \sqrt{\frac{1}{N-1} \sum_{t=1}^N (R_t - \bar{R})^2}$$

$$\sigma_{\text{annualized}} = s_{\text{daily}} \times \sqrt{252}$$

*(Use $\sqrt{365}$ for 24/7 trading markets like cryptocurrency).*

### 7.2 Sharpe Ratio
Measures excess return per unit of total risk:

$$\text{Sharpe} = \frac{R_{\text{annualized}} - R_{f}}{\sigma_{\text{annualized}}}$$

- $R_f$: Annualized risk-free rate (e.g. 10-Year US Treasury yield).

### 7.3 Maximum Drawdown (MDD)
The peak-to-trough decline over a specified timeframe:

$$\text{Drawdown}_t = \frac{V_t - \max_{0 \le s \le t} V_s}{\max_{0 \le s \le t} V_s}$$

$$\text{MDD} = \min_{0 \le t \le T} \text{Drawdown}_t$$

---

## 8. Go Reference Implementation (`shopspring/decimal`)

Here are the canonical, production-ready Go implementations adhering strictly to these standards:

### 8.1 GIPS-Compliant Annualized Return
```go
package domain

import (
	"errors"
	"math"
	"time"

	"github.com/shopspring/decimal"
)

var (
	oneHundred = decimal.NewFromInt(100)
	one        = decimal.NewFromInt(1)
	zero       = decimal.Zero
)

// CalculateAnnualizedReturn computes the GIPS-compliant annualized return percentage.
// If the period between startDate and endDate is less than 365 days, it returns an error
// or returns non-annualized period return per GIPS Standard 2.A.20.
func CalculateAnnualizedReturn(
	twrIndex decimal.Decimal,
	startDate time.Time,
	endDate time.Time,
) (annualized decimal.Decimal, isAnnualized bool, err error) {
	if !twrIndex.IsPositive() {
		return zero, false, errors.New("twrIndex must be positive")
	}

	days := endDate.Sub(startDate).Hours() / 24.0
	if days <= 0 {
		return zero, false, errors.New("endDate must be after startDate")
	}

	// GIPS Standard 2.A.20: Periods under 1 year (< 365 days) MUST NOT be annualized.
	if days < 365.0 {
		// Return cumulative period return: (TWRIndex - 1) * 100
		periodReturn := twrIndex.Sub(one).Mul(oneHundred).Round(2)
		return periodReturn, false, nil
	}

	// Geometric compound annualization: (TWRIndex ^ (365.25 / days) - 1) * 100
	twrFloat, _ := twrIndex.Float64()
	exponent := 365.25 / days
	annFactor := math.Pow(twrFloat, exponent)

	annDec := decimal.NewFromFloat(annFactor).Sub(one).Mul(oneHundred).Round(2)
	return annDec, true, nil
}
```

### 8.2 Daily Sub-Period Return with End-of-Day External Flow
```go
// CalculateDailyReturn computes sub-period daily return adjusting for external net flows.
func CalculateDailyReturn(
	prevValuation decimal.Decimal,
	currentValuation decimal.Decimal,
	netExternalFlow decimal.Decimal,
) decimal.Decimal {
	// Boundary: If prior valuation was 0 (e.g. brand new portfolio deposit)
	if !prevValuation.IsPositive() {
		return zero
	}

	// Isolated market movement = CurrentValuation - NetExternalFlow - PrevValuation
	netGain := currentValuation.Sub(netExternalFlow).Sub(prevValuation)

	// Return = NetGain / PrevValuation
	return netGain.DivRound(prevValuation, 8)
}
```

### 8.3 Safe Division & Zero Handling Checklists
When writing financial calculation code in GraphFolio:
- [ ] Always check `.IsPositive()` or `!divisor.IsZero()` before calling `.Div()` or `.DivRound()`.
- [ ] If initial cost basis or starting value is zero, return `decimal.Zero` (never return `+Inf` or NaN).
- [ ] Always pass an explicit precision scale (e.g. 6 to 8 places) to `.DivRound()`.
- [ ] Verify that external cash flows are isolated from internal asset swaps (BUY/SELL) and income (DIVIDENDS).
- [ ] Enforce the **365-day check** before presenting any metric labeled "Annualized Return".
