# User Story: Tax Lot Inspector & Capital Gains Reports

> **Target**: Primary Investor Application (`web/apps/main-app`)  
> **Status**: Planned (Roadmap)  
> **Implementation Plan**: *(To be drafted)*  
> **Roadmap Reference**: [FEATURES.md (Feature #19)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L29)

---

## 1. User Story

> **As an** investor preparing tax filings or planning portfolio rebalancing,  
> **I want to** inspect individual open tax lots and historical closed lot disposals with holding period classifications,  
> **So that** I can track short-term vs long-term capital gains and identify tax-loss harvesting opportunities.

---

## 2. Acceptance Criteria

- [ ] **Open Tax Lot Inspection**: Display a granular table of open purchase lots per instrument with purchase date, cost basis per share, remaining shares, and unrealized gain/loss.
- [ ] **Holding Period Classification**: Categorize lots into Short-Term (< 1 year) and Long-Term (≥ 1 year) according to applicable tax jurisdiction rules.
- [ ] **Disposal Traceability**: For every SELL transaction, display the matched purchase lots that were relieved, the cost basis relieved, and the realized capital gain.
- [ ] **Tax-Loss Harvesting Indicators**: Highlight underwater lots with high cost bases to easily identify potential tax-loss harvesting candidates.
- [ ] **Exportable Reports**: Allow exporting capital gains reports to CSV/JSON format for tax preparation software.

---

## 3. Related Documentation

- **Database Foundations**: [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/db-implementation-plan.md) (tables `portfolio.tax_lots` and `portfolio.lot_disposals`)
- **Cost Basis Relieve Logic**: [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/cost-basis-switching-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
