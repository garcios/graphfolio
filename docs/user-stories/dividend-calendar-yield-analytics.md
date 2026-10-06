# User Story: Dividend Calendar & Yield Analytics

> **Target**: Primary Investor Application (`web/apps/main-app`)  
> **Status**: Planned (Roadmap)  
> **Implementation Plan**: *(To be drafted)*  
> **Roadmap Reference**: [FEATURES.md (Feature #20)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L30)

---

## 1. User Story

> **As an** income-focused investor relying on passive cash flows,  
> **I want to** view an interactive dividend calendar, projected payment schedules, and Yield on Cost (YOC) analytics,  
> **So that** I can anticipate cash inflows, track dividend safety, and plan reinvestment strategies.

---

## 2. Acceptance Criteria

- [ ] **Interactive Calendar**: Display a monthly/quarterly calendar highlighting upcoming ex-dividend dates, payment dates, and declared amounts per holding.
- [ ] **Forward Income Projections**: Calculate projected annualized forward dividend cash flow based on currently held share quantities and latest declared payouts.
- [ ] **Yield on Cost (YOC)**: Display both Current Dividend Yield (`Annual Dividend / Current Price`) and Yield on Cost (`Annual Dividend / Purchase Cost Basis`) per holding.
- [ ] **Reinvestment (DRIP) Modeling**: Project future compounding curves comparing cash payouts vs automated dividend reinvestment plans.
- [ ] **Historical Dividend Stream**: Provide a chronological log of all received dividend cash flows credited to portfolio cash balances.

---

## 3. Related Documentation

- **Transaction Ingestion**: [add-transactions-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/add-transactions-implementation-plan.md)
- **Corporate Actions**: [db-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/db-implementation-plan.md) (table `portfolio.corporate_actions`)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
