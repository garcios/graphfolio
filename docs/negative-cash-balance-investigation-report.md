# Negative Cash Balance Investigation Report

---

## 1. Executive Summary

In GraphFolio, the investor portfolio currently reports a consolidated cash balance of **`-A$250,187.80`** (composed of **`-$105,830.47 USD`** and **`-$98,591.27 AUD`**).

Because total portfolio value is calculated as:
$$\text{Total Portfolio Value} = \text{Invested Holdings Market Value} + \text{Consolidated Cash Balance}$$

The system treats this negative cash balance as an overdraft / margin liability. Consequently, the reported total portfolio value is depressed from **`A$411,252.22`** (the actual market value of open positions) down to **`A$161,064.42`**.

The primary root cause is that **the transaction ledger contains zero `DEPOSIT` transactions**. Across all 542 transactions imported from brokerage statements (covering the period May 2023 to October 2026), 310 stock purchase (`BUY`) orders were debited against a cash ledger that begins at `$0.00`. Without recorded capital funding events, every purchase pushed the portfolio deeper into negative cash.

---

## 2. Ledger Architecture & Computation Walkthrough

In GraphFolio, cash balances are non-persistent ledger projections computed deterministically by [`ProcessLedger`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go#L55-L180) in `services/portfolio-api`.

When a portfolio's ledger is processed:
1. `cashBalances := make(map[string]decimal.Decimal)` initializes every currency bucket at zero ($0.00$).
2. Transactions are replayed chronologically.
3. Debits and credits are applied based on transaction type, fee currency, and withholding tax.
4. Resulting balances are stored in `portfolio.cash_balances` and consolidated to the portfolio's base currency via [`CalculatePortfolioSummary`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go#L57-L125).

---

### A. AUD Cash Ledger Computation

| Transaction Type | Count | Ledger Cash Impact Formula | Gross Consideration | Incurred Fees | Net Cash Impact |
| :--- | :---: | :--- | :---: | :---: | :---: |
| **DEPOSIT** | 0 | $+\text{Amount}$ | $0.00 | $0.00 | $0.00 |
| **WITHDRAWAL** | 0 | $-\text{Amount}$ | $0.00 | $0.00 | $0.00 |
| **BUY** | 179 | $-(\text{Amount} + \text{Fee})$ | $262,601.84 | $1,955.54 | **-$264,557.38** |
| **SELL** | 39 | $+(\text{Amount} - \text{Fee})$ | $166,612.93 | $646.82 | **+$165,966.11** |
| **DIVIDEND** | 0 | $+(\text{Amount} - \text{WHT})$ | $0.00 | $0.00 | $0.00 |
| **Total AUD Balance** | **218** | | | | **-$98,591.27** |

$$\begin{aligned}
\text{AUD Balance} &= \text{Starting Balance} + \sum \text{Credits} - \sum \text{Debits} \\
&= 0.00 + 165,966.11 - 264,557.38 \\
&= \mathbf{-A\$98,591.27}
\end{aligned}$$

---

### B. USD Cash Ledger Computation

| Transaction Type | Count | Ledger Cash Impact Formula | Gross Consideration | Incurred Fees | Net Cash Impact |
| :--- | :---: | :--- | :---: | :---: | :---: |
| **DEPOSIT** | 0 | $+\text{Amount}$ | $0.00 | $0.00 | $0.00 |
| **WITHDRAWAL** | 0 | $-\text{Amount}$ | $0.00 | $0.00 | $0.00 |
| **BUY** | 131 | $-(\text{Amount} + \text{Fee})$* | $135,329.22 | $1,054.75 | **-$136,383.97** |
| **SELL** | 11 | $+(\text{Amount} - \text{Fee})$ | $28,353.42 | $0.00 | **+$28,353.42** |
| **DIVIDEND** | 180 | $+(\text{Amount} - \text{WHT})$ | $2,200.08 | $0.00 | **+$2,200.08** |
| **SPLIT** | 2 | None (Share count multiplier) | $0.00 | $0.00 | $0.00 |
| **Total USD Balance** | **324** | | | | **-$105,830.47** |

$$\begin{aligned}
\text{USD Balance} &= \text{Starting Balance} + \sum \text{Credits} - \sum \text{Debits} \\
&= 0.00 + (28,353.42 + 2,200.08) - 136,383.97 \\
&= 30,553.50 - 136,383.97 \\
&= \mathbf{-\$105,830.47\text{ USD}}
\end{aligned}$$

*\*Brokerage Fee Currency Allocation:*
In the persisted `portfolio.cash_balances` table, the $1,054.75 USD buy fees were deducted from the USD cash bucket prior to the introduction of PR #25 (`fee_currency_code = 'AUD'`). When replayed with the new cross-currency fee logic, the USD balance recalibrates to `-$104,775.72 USD` and the AUD balance adjusts to `-$99,646.02 AUD`. In both scenarios, both balances remain deeply negative.

---

### C. Base Currency (AUD) Consolidation

In [`CalculatePortfolioSummary`](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/domain/calculator.go#L89-L99), multi-currency balances are unified into the portfolio base currency (`AUD`):

```go
totalCashBase := decimal.Zero
for _, c := range cashBalances {
    fxRate := one
    if c.CurrencyCode != baseCurrency {
        if rate, ok := cashFXRates[c.CurrencyCode]; ok && rate.IsPositive() {
            fxRate = rate
        }
    }
    totalCashBase = totalCashBase.Add(c.Balance.Mul(fxRate))
}
```

Using the latest market FX rate from `portfolio.fx_rates` (`USD/AUD = 1.4324469034`):

$$\begin{aligned}
\text{USD Cash in AUD} &= -\$105,830.47 \times 1.4324469034 = -\text{A}\$151,596.525\dots \\
\text{Total Cash in AUD} &= -\text{A}\$98,591.27 + (-\text{A}\$151,596.53) = \mathbf{-A\$250,187.80}
\end{aligned}$$

---

## 3. Root Cause Analysis

### Root Cause 1: Zero Capital Inflow (`DEPOSIT`) Transactions
In strict double-entry investment accounting, cash and equity are dual legs:
$$\text{Cash Balance} = \sum \text{Deposits} - \sum \text{Withdrawals} - \sum \text{Purchases} + \sum \text{Sales} + \sum \text{Dividends}$$
Because the database contains only trade execution transactions (`BUY`, `SELL`, `DIVIDEND`), the cumulative net cost of acquiring all active holdings was subtracted directly from a starting cash baseline of `$0.00`.

### Root Cause 2: CSV Import Scope (Trade Orders vs. Cash Account Statements)
All transactions in this portfolio originated from **nabtrade** CSV statement exports:
- Trade activity exports (confirmations) only list order executions (shares bought/sold). They do not record when money was deposited from an external bank account into the broker cash account.
- The nabtrade CSV parser ([`nabtrade.ts`](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/services/csv/parsers/nabtrade.ts#L33-L144)) is designed specifically to match regex patterns for `BUY`, `SELL`, and `DIVIDEND`. Any non-trade bank/cash account rows (e.g. `Direct Credit - Deposit`, `Funds Transfer`, `Interest`) are filtered out or rejected with validation warnings.

### Root Cause 3: Dual-Currency Accounts vs. Automatic Currency Sweeps
When Australian investors purchase US equities through brokers like nabtrade:
- International trades typically convert AUD from the Australian cash account on-the-fly (as reflected in transaction notes: `Imported from nabtrade International (Order: ..., FX: 0.7059)`).
- Because trades specify `currency_code = 'USD'`, GraphFolio created a dedicated **USD cash bucket**, debiting $135,329.22 USD.
- Without a paired currency conversion record (e.g., Transfer: Debit AUD, Credit USD) or initial USD funding, the USD ledger falls into an artificial deficit.

---

## 4. Impact on Portfolio Valuation & Performance

1. **Total Portfolio Value**:
   - Market Value of Holdings: **`A$411,252.22`**
   - Cash Balance: **`-A$250,187.80`**
   - Displayed Portfolio Value: **`A$161,064.42`**
2. **Performance / TWR / Money-Weighted Return**:
   - Because the initial asset purchases are treated as financed by debt/overdraft rather than funded by an external deposit, cash-flow weighted return calculations (MWR/IRR) and daily valuation snapshots treat subsequent debt repayments or holdings growth under distorted equity denominators.

---

## 5. Recommendations to Fix the Problem

Depending on how you wish to manage cash tracking within GraphFolio, here are three recommended paths:

### Option 1: Record Initial & Historical Deposits (Recommended for Exact Accounting)
If you want accurate cash tracking alongside your stock holdings:
1. **Add Initial / Lump-Sum Deposit**:
   Record one or more `DEPOSIT` transactions (using the existing UI **+ Add Transaction** modal) matching your actual capital contributions.
2. **Current Cash Balance Calibration**:
   If your real-world brokerage cash balance is currently, for example, `A$5,000.00 AUD`, you can add a `DEPOSIT` transaction for `A$255,187.80 AUD` (or the equivalent in USD/AUD).
   - This immediately restores your cash balance to your real uninvested cash level and brings your Total Portfolio Value to its true market value (`~A$416,252.22`).

---

### Option 2: Add Cash Statement Parsing to the CSV Importer
Enhance the CSV import engine ([`nabtrade.ts`](file:///Users/oscargarcia/workspace/graphfolio/web/apps/main-app/src/services/csv/parsers/nabtrade.ts)):
1. Support importing raw **nabtrade Cash Account statements** (the bank account statement rather than order confirmations).
2. Recognize cash inflow patterns (`DEPOSIT`, `FUNDS TRANSFER IN`, `INTEREST`) and cash outflow patterns (`WITHDRAWAL`).
3. For international trades where nabtrade swept AUD to fund USD orders, parse the paired currency transfer legs so AUD is debited and USD is credited, preventing the USD ledger from dropping into an artificial deficit.

---

### Option 3: Configurable Cash Tracking Mode ("Unlinked Cash" Setting)
In platforms like Sharesight and Navexa, investors can choose whether cash tracking is **Linked** or **Unlinked**:
1. **Linked Cash Account (Current behavior)**: Share purchases debit cash, share sales credit cash. Requires tracking deposits.
2. **Unlinked / Floating Cash**:
   - Stock purchases and sales do not affect a cash balance (or cash balance only reflects explicit cash transactions / user-defined static balance).
   - Total Portfolio Value equals Invested Holdings Value ($411,252.22) plus any manually configured cash balance, preventing unlinked purchases from artificially pulling portfolio net worth into the negative.
