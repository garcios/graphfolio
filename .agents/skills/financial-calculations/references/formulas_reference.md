# Financial Calculations Cheat Sheet & Verification Test Cases

This reference provides rapid lookups for formulas and expected numerical outputs for unit test suites across GraphFolio.

---

## 1. Quick Formula Cheat Sheet

| Metric | Formula | GIPS / Accounting Rule |
|---|---|---|
| **Simple / Period Return** | $R = \frac{V_{\text{end}} - V_{\text{start}}}{V_{\text{start}}}$ | Zero external cash flows only |
| **Annualized Return (CAGR)** | $R_{\text{ann}} = \left( (1 + R)^{\frac{365.25}{D}} - 1 \right) \times 100$ | **Prohibited if $D < 365$ days** |
| **Sub-Period Return (EOD Flow)** | $R_t = \frac{V_t - F_t - V_{t-1}}{V_{t-1}}$ | $F_t$ = Deposits $-$ Withdrawals only |
| **Sub-Period Return (Mid-Day Flow)** | $R_t = \frac{V_t - V_{t-1} - F_t}{V_{t-1} + 0.5 \times F_t}$ | Weighting flows at middle of trading day |
| **TWR Cumulative Index** | $\text{TWR}_t = \text{TWR}_{t-1} \times (1 + R_t)$ | Baseline $\text{TWR}_0 = 1.000000000000$ |
| **Modified Dietz** | $R = \frac{V_E - V_S - F}{V_S + \sum (W_i \times F_i)}$ | $W_i = \frac{CD - D_i}{CD}$ |
| **Realized PnL** | $\text{NetProceeds} - \text{RelievedCostBasis}$ | Triggered on SELL execution date |
| **Unrealized PnL** | $\text{CurrentMarketValue} - \text{RemainingCostBasis}$ | Mark-to-Market balance sheet adjustment |
| **Stock Split Basis Adjustment** | $Q_{\text{new}} = Q \times \frac{A}{B}$, $\text{Cost}_{\text{new}} = \text{Cost} \times \frac{B}{A}$ | Non-taxable; total cost basis unchanged |
| **FX Decomposed Return** | $R_{\text{base}} = R_{\text{asset}} + R_{\text{fx}} + (R_{\text{asset}} \times R_{\text{fx}})$ | Multi-currency holding performance |

---

## 2. Test Verification Cases

### Test Case 1: Annualized Return for Multi-Year Investment
- **Start Date**: 2023-01-01
- **End Date**: 2026-01-01 (1,096 calendar days = 3.00 years)
- **Cumulative TWR**: $+40.4928\%$ (TWRIndex = $1.404928$)
- **Calculation**:
  $$\text{Exponent} = \frac{365.25}{1096} = 0.3332573$$
  $$(1.404928)^{0.3332573} - 1 = 1.120000 - 1 = 0.1200 = 12.00\%$$
- **Expected Annualized Return**: **12.00%**

### Test Case 2: Sub-1-Year GIPS Prohibition
- **Start Date**: 2026-01-01
- **End Date**: 2026-04-01 (90 calendar days)
- **Cumulative Return**: $+5.00\%$
- **Naive Compounding (INVALID)**: $(1.05)^{\frac{365.25}{90}} - 1 = 22.03\%$ ❌ *(Violates GIPS Standard 2.A.20)*
- **Expected Compliant Output**: **5.00% (Cumulative Period Return, not annualized)** ✅

### Test Case 3: Daily Return with Mid-Stream Deposit
- **Day $t-1$ Ending Valuation**: $\$100,000$
- **Day $t$ Deposit**: $\$20,000$ (Cash transferred into account)
- **Day $t$ Market Movement**: Stocks gained 2% on pre-deposit valuation ($\$2,000$)
- **Day $t$ Ending Valuation**: $\$100,000 + \$2,000 + \$20,000 = \$122,000$
- **Calculation**:
  $$R_t = \frac{\$122,000 - \$20,000 - \$100,000}{\$100,000} = \frac{\$2,000}{\$100,000} = 0.0200 = +2.00\%$$
- **Verification**: The $+2.00\%$ daily return correctly isolates investment performance; the $\$20,000$ deposit does not produce an artificial $+22.00\%$ gain.

### Test Case 4: Stock Split (3-for-1 Forward Split)
- **Original Lot**: 100 shares @ \$300/share (Cost Basis = \$30,000)
- **Corporate Action**: 3:1 forward split ($A=3, B=1$)
- **New Quantity**: $100 \times \frac{3}{1} = 300$ shares
- **New Unit Cost**: $\$300 \times \frac{1}{3} = \$100$/share
- **New Total Cost Basis**: $300 \times \$100 = \$30,000$ (Invariant preserved)
- **Realized Gain**: $\$0.00$ (Non-taxable event)
