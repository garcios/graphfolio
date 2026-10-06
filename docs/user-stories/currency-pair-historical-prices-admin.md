# User Story: Currency Pair Historical Prices & FX Trend Inspection

> **Target**: Internal Operations Portal (`web/apps/admin-app`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [currency-pair-historical-prices-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/currency-pair-historical-prices-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #16)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L26)

---

## 1. User Story

> **As an** operations administrator and financial controller,  
> **I want to** inspect historical daily fixing rates, trend charts, and reciprocal rates for tracked currency pairs in the Admin Portal,  
> **So that** I can audit central bank FX benchmarks (ECB), identify exchange rate anomalies, and record manual overrides when official feeds experience outages.

---

## 2. Acceptance Criteria

- [ ] **Pair Selector & Discovery**: Display pills/selector for all active currency pairs (`EUR/USD`, `USD/EUR`, `USD/GBP`, `USD/AUD`, etc.) with their latest fixing marks.
- [ ] **Interactive Trend Chart**: Render an interactive SVG trend curve plotting daily rates across selectable timeframes (`1W`, `1M`, `3M`, `1Y`, `ALL`) with hover crosshairs.
- [ ] **Reciprocal Rate Inversion**: Provide a one-click "Invert Pair (Base ⇄ Quote)" toggle computing reciprocal rates with exact 10-decimal precision.
- [ ] **Authoritative Rates Ledger**: Provide a paginated table showing date, direct rate, inverted rate, provider source (`ECB`, `Twelve Data`), and override indicators.
- [ ] **Manual Rate Override**: Allow authorized operators to input a manual override with mandatory audit justification notes.

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [currency-pair-historical-prices-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/currency-pair-historical-prices-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
