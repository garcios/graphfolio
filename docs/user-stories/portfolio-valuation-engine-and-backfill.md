# User Story: Portfolio Valuation Engine & Historical Replay

> **Target**: Database & Portfolio Engine (`services/portfolio-api`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-valuation-engine-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #13)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L23)

---

## 1. User Story

> **As an** investor with a complex multi-year transaction history,  
> **I want to** have daily portfolio valuation snapshots, cash balances, and Time-Weighted Returns (TWR) automatically computed and reconstructed over time,  
> **So that** my performance curve reflects accurate day-by-day wealth progression independent of cash deposit and withdrawal distortion.

---

## 2. Acceptance Criteria

- [ ] **Daily Valuation Snapshots**: Persist daily snapshots in `portfolio.portfolio_valuations` recording total market value, cash value, net cash flows, and cumulative TWR index.
- [ ] **Automated Daily Valuation Job**: Execute an end-of-day batch worker evaluating holding balances against closing prices and foreign exchange marks.
- [ ] **Historical Ledger Replay**: Automatically trigger historical daily valuation replays from the earliest transaction date when past-dated transactions are ingested.
- [ ] **Time-Weighted Return (TWR) Precision**: Compute daily sub-period returns and link compounding factors using exact zero-drift fixed-point arithmetic conforming to GIPS standards.
- [ ] **Performance Chart Compatibility**: Provide daily points formatted for high-performance rendering in the investor SVG performance chart.

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/portfolio-valuation-engine-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
