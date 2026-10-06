# User Story: Dynamic Cost Basis Method Switching (Average Cost ↔ FIFO)

> **Target**: Primary Investor Application (`web/apps/main-app`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cost-basis-switching-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #11)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L21)

---

## 1. User Story

> **As an** active investor and tax-conscious portfolio owner,  
> **I want to** toggle between Average Cost and FIFO (First-In, First-Out) cost basis relieve methods on demand,  
> **So that** I can accurately evaluate my taxable realized capital gains and compare performance outcomes under different accounting standards.

---

## 2. Acceptance Criteria

- [ ] **Method Selector**: Provide a clear segmented toggle control (`[ Average Cost | FIFO ]`) accessible in the portfolio header or settings.
- [ ] **Ledger Replay**: Replaying the transaction ledger immediately recalculates holding cost bases, realized capital gains, and total return percentages when switched.
- [ ] **Tax Lot Precision**: Ensure lot-level disposals under FIFO strictly deplete oldest matching purchase lots before newer acquisitions with exact decimal precision.
- [ ] **Persistence**: The selected cost basis method is persisted to `portfolio.portfolios.cost_basis_method` and retained across sessions.
- [ ] **Educational Tooltip**: Include an informational indicator explaining tax implications (e.g., pooling average acquisition costs vs selling oldest acquired shares).

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cost-basis-switching-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
