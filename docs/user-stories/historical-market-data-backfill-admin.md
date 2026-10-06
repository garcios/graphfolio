# User Story: Historical Market Data Backfill (Admin Portal)

> **Target**: Internal Operations Portal (`web/apps/admin-app`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [admin-historical-backfill-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-historical-backfill-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #17)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L27)

---

## 1. User Story

> **As an** Admin Portal user and data operations manager,  
> **I want to** trigger on-demand historical price and currency rate backfills for customizable date ranges and instrument scopes,  
> **So that** newly registered assets and unpriced transaction histories have full daily closing mark coverage without relying solely on manual CLI jobs.

---

## 2. Acceptance Criteria

- [ ] **Date Range Specification**: Provide interactive date pickers (`Start Date`, `End Date`) with quick presets (`30 Days`, `90 Days`, `YTD`, `1 Year`, `Max`).
- [ ] **Scope Selection**: Allow operators to toggle between backfilling "All Active" entities or selecting specific ticker symbols (`AAPL`, `MSFT`) and currency pairs (`EUR/USD`).
- [ ] **Target Toggles**: Support independent checkboxes for backfilling asset prices and/or central bank FX rates.
- [ ] **Rate Limit Telemetry Projection**: Display estimated outbound API calls against current token-bucket budget (`rateLimitRemaining`) before execution.
- [ ] **Valuation Replay Option**: Provide an optional checkbox to automatically recompute affected historical portfolio valuations upon backfill completion.

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [admin-historical-backfill-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/admin-historical-backfill-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
