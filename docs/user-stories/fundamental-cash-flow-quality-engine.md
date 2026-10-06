# User Story: Look-Through Fundamental & Cash Flow Quality Engine

> **Target**: Portfolio Domain (`services/portfolio-api`), BFF (`bff/`), and Primary Investor Application (`web/apps/main-app`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [fundamental-cash-flow-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/fundamental-cash-flow-engine-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #15)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L25)

---

## 1. User Story

> **As an** equity investor adopting a business-owner investment mindset,  
> **I want to** analyze my portfolio's look-through fundamental financial metrics (Owner Earnings, Free Cash Flow, ROIC, and Sloan Accrual Ratios),  
> **So that** I can track whether the intrinsic business earnings power of my portfolio is compounding independently of short-term market price fluctuations.

---

## 2. Acceptance Criteria

- [ ] **Look-Through Aggregation**: Accurately calculate the investor's exact proportional look-through share of Operating Cash Flow, Owner Earnings, and Maintenance CapEx based on share count ownership.
- [ ] **Capital Allocation Scorecard**: Display weighted portfolio Return on Invested Capital (ROIC), Reinvestment Rate, and Intrinsic Compounding Rate ($g = \text{ROIC} \times \text{RR}$).
- [ ] **Earnings Quality Detection**: Compute the Sloan Accrual Ratio to highlight divergence between reported GAAP Net Income and genuine cash realizations.
- [ ] **Business Owner Dashboard**: Provide a visual chart contrasting portfolio Market Value growth against Business Intrinsic Value growth over multi-year horizons.
- [ ] **Zero Float Arithmetic**: All per-share fundamentals, shares outstanding, and look-through cash flows computed with exact fixed-point decimal arithmetic.

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [fundamental-cash-flow-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/fundamental-cash-flow-engine-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
