# User Story: Delete Asset & Cascading Price Cleanup (Admin Portal)

> **Story ID**: US-ADM-102  
> **Epic**: Internal Administration & Asset Management  
> **Target**: Admin Portal Application (`web/apps/admin-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), PostgreSQL (`portfolio` schema)  
> **Status**: **DONE**  
> **Implementation Plan**: [delete-asset-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/delete-asset-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #18)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L28)  
> **Standards Reference**: [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** operations administrator,  
> **I want to** permanently remove an erroneously created tradable asset from the master directory along with all its historical market prices,  
> **So that** obsolete test fixtures, incorrect ticker symbols, and invalid assets are purged completely while referential integrity is guaranteed by blocking deletion if the asset is actively held or referenced in user transactions.

---

## 2. Business Context & Safety Constraints

1. **Purging Erroneous Assets**:
   - Administrators occasionally register assets with typographical errors (e.g. wrong ticker, incorrect exchange MIC, or duplicate records). Previously, administrators could only toggle `isActive = false`, leaving obsolete assets cluttering searches and pricing jobs.
2. **Atomic Price Cascade**:
   - Deleting an asset requires atomic deletion of all corresponding market closing prices in `portfolio.instrument_prices` to prevent orphaned price rows.
3. **Ledger Integrity Guard (In-Use Rejection)**:
   - If an asset is actively referenced by user transactions (`portfolio.transactions`) or open holdings (`portfolio.holdings`), deletion must be strictly rejected with `ErrInstrumentInUse` (`codes.FailedPrecondition`). Operators must deactivate rather than delete such assets.
4. **Destructive Action Safety**:
   - Deleting an asset is permanent and unrecoverable. The Admin Portal must require an explicit confirmation modal before executing the deletion.

---

## 3. Acceptance Criteria

### 3.1 Referential Integrity & Database Protection
- [x] **AC-1 (In-Use Verification Guard)**: If `portfolio.transactions` contains any transaction referencing the instrument ID, deletion is blocked, returning `ErrInstrumentInUse` / `codes.FailedPrecondition` with the explanation:  
  `"cannot delete instrument in use by transactions; deactivate it instead"`.
- [x] **AC-2 (Open Holdings Guard)**: If `portfolio.holdings` contains an open position for the instrument, deletion is blocked.
- [x] **AC-3 (Atomic Price Cascade)**: For unused assets, deletion of the instrument and all linked rows in `portfolio.instrument_prices` executes within a single atomic database transaction via `ON DELETE CASCADE`.

### 3.2 Protocol & GraphQL API Layer
- [x] **AC-4 (Proto RPC)**: `proto/portfolio/v1/portfolio.proto` defines `DeleteInstrument(DeleteInstrumentRequest)` returning `DeleteInstrumentResponse`.
- [x] **AC-5 (GraphQL Mutation)**: BFF exposes `deleteInstrument(id: ID!): DeleteInstrumentPayload!` mapping to the gRPC client with comprehensive error translation (`NOT_FOUND`, `FAILED_PRECONDITION`).

### 3.3 Admin Portal User Experience
- [x] **AC-6 (Delete Action Button)**: `AssetManagement.tsx` table actions include a "Delete" button (`variant="danger"`).
- [x] **AC-7 (Confirmation Modal)**: Clicking "Delete" opens `DeleteAssetModal.tsx` displaying the asset symbol, name, exchange, and a warning that all associated pricing records will be wiped.
- [x] **AC-8 (Client-Side State Update)**: Upon successful deletion, the modal closes, displays a success toast notification, and dynamically removes the asset row from the directory table without a full page reload.
- [x] **AC-9 (In-Use Error Presentation)**: If deletion is rejected due to active references, an error modal or toast alerts the admin to deactivate the asset instead.

---

## 4. Technical & Implementation Notes

- **Database Migration**: `services/portfolio-api/migrations/000007_cascade_instrument_prices.up.sql` updates `portfolio.instrument_prices` foreign key to `ON DELETE CASCADE`.
- **Backend Service**: `services/portfolio-api/internal/service/admin.go` coordinates `DeleteInstrument`, executing transactional in-use validation against `portfolio.transactions`.
- **Admin Modal**: `web/apps/admin-app/src/components/DeleteAssetModal.tsx` and `.css` implemented using `@graphfolio/ui` modal tokens.
