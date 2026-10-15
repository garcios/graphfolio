# User Story: Exchange Code Dropdown in Admin Asset Registration

> **Story ID**: US-ADM-104  
> **Epic**: Internal Administration & Asset Management  
> **Target**: Admin Portal Application (`web/apps/admin-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), Protobuf (`proto/portfolio/v1/portfolio.proto`)  
> **Status**: **DONE**  
> **Implementation Plan**: [exchange-dropdown-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/exchange-dropdown-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #10)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L20)  
> **Standards Reference**: [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** operations administrator registering a new tradable instrument,  
> **I want to** select the exchange from an authoritative dropdown list populated dynamically from supported market exchanges,  
> **So that** I cannot enter invalid free-text abbreviations or typos (such as "US") that violate foreign key constraints on the exchange code.

---

## 2. Business Context & Usability Need

1. **Elimination of Foreign Key Violations**:
   - `portfolio.instruments.exchange_code` has a strict foreign key reference to `portfolio.exchanges.code` (ISO 10383 Market Identifier Codes, e.g. `XNAS`, `XNYS`, `XASX`).
   - The initial modal presented a free-text input with default text `"US"`, which is an ISO country code rather than an exchange MIC code. Submitting with `"US"` caused foreign key insert failures.
2. **Dynamic Authoritative Listing**:
   - Rather than hardcoding exchange codes on the client (which risks schema drift), the dropdown is backed by the authoritative `portfolio.exchanges` database table via gRPC `ListExchanges` and GraphQL `exchanges`.
   - New market exchanges added to the reference table automatically appear in the UI dropdown without frontend code redeployments.

---

## 3. Acceptance Criteria

### 3.1 Backend Contract & Service Layer
- [x] **AC-1 (ListExchanges RPC)**: `proto/portfolio/v1/portfolio.proto` defines `rpc ListExchanges(ListExchangesRequest) returns (ListExchangesResponse)` returning an array of `Exchange` messages (`code`, `name`, `country`, `timezone`).
- [x] **AC-2 (Database Query)**: Repository retrieves all supported exchanges from `portfolio.exchanges` ordered alphabetically by code.
- [x] **AC-3 (BFF GraphQL Query)**: BFF exposes `exchanges: [Exchange!]!` mapping directly to the gRPC client.

### 3.2 Admin Portal UI Integration
- [x] **AC-4 (Dropdown Replacement)**: In `AddInstrumentModal.tsx`, the free-text `<Input>` for Exchange Code is replaced with a `<Select>` dropdown.
- [x] **AC-5 (Informative Option Formatting)**: Each dropdown option displays both code and exchange name (e.g. `XNAS — NASDAQ Stock Market`, `XNYS — New York Stock Exchange`, `XASX — Australian Securities Exchange`).
- [x] **AC-6 (Placeholder & Mandatory Selection)**: The select control displays `"Select Exchange"` as an initial placeholder and marks the field as required.
- [x] **AC-7 (Form Submission Integrity)**: The selected exchange option submits the ISO 10383 MIC string (`XNAS`) directly into `createInstrument(input: { exchangeCode: ... })`, guaranteeing foreign key validity.

---

## 4. Technical & Implementation Notes

- **Shared UI**: `web/packages/ui/src/components/Select/Select.tsx` updated to support placeholders and typed option objects.
- **Admin Modal**: `web/apps/admin-app/src/components/AddInstrumentModal.tsx` queries `exchanges` on render and binds the selection to state.
- **Backend Service**: `services/portfolio-api/internal/service/admin.go` implements `ListExchanges` fetching from `portfolio.exchanges`.
