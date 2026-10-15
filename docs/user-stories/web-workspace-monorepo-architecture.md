# User Story: Frontend Workspace Monorepo Architecture (`main-app` & `admin-app`)

> **Story ID**: US-FE-101  
> **Epic**: Frontend Workspace Infrastructure & Design System  
> **Target**: Frontend Monorepo (`web/`), Applications (`web/apps/main-app`, `web/apps/admin-app`), Shared Packages (`web/packages/ui`, `web/packages/api-client`), Build Tooling (Vite, TypeScript, Oxlint)  
> **Status**: **DONE**  
> **Implementation Plan**: [web-workspace-refactoring-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/web-workspace-refactoring-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #9)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L19)  
> **Standards Reference**: [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statements

### 1.1 Multi-Application Decoupling & Security
> **As a** platform engineering team,  
> **I want to** decouple the public-facing investor web application (`main-app`) and internal operations portal (`admin-app`) into separate applications within an npm workspace monorepo,  
> **So that** administrative workflows, internal endpoints, and sensitive forms are never shipped in the client bundle delivered to public retail investors.

### 1.2 Shared Design System & API Client
> **As a** frontend developer,  
> **I want to** share reusable atomic UI primitives, dark glassmorphic design tokens (`@graphfolio/ui`), and a typed auto-generated GraphQL client (`@graphfolio/api-client`) across all web applications,  
> **So that** UI components and API access patterns maintain 100% visual consistency and zero code duplication without breaking hot module reloading.

---

## 2. Business Context & Architectural Motivation

1. **Monolithic Bundle Segregation**:
   - Previously, `web/` was a single monolithic Vite app. Bundling internal administrative capabilities (asset creation, price overrides, feed ingestion diagnostics) with the public investor dashboard exposed internal administrative APIs and bloated client bundle sizes.
2. **Modular Monorepo Structure**:
   - `web/` is refactored into an npm workspace:
     - `apps/main-app`: Investor dashboard on port `:5173`.
     - `apps/admin-app`: Administrative operations portal on port `:5174`.
     - `packages/ui`: Shared design system (`@graphfolio/ui`) providing design tokens, CSS reset, and atomic components (`Button`, `Modal`, `Card`, `Badge`, `Table`, `Input`, `Select`).
     - `packages/api-client`: Shared GenQL generated client (`@graphfolio/api-client`) with bearer auth injection.
3. **Unified Developer Experience**:
   - Single root `package-lock.json`, shared `tsconfig.base.json`, standardized Oxlint linting, and consolidated Makefile targets (`run-web`, `run-admin`, `run-all-web`, `build-web`).

---

## 3. Acceptance Criteria

### 3.1 Monorepo Workspace Configuration
- [x] **AC-1 (Root Workspaces Definition)**: `web/package.json` declares `"workspaces": ["apps/*", "packages/*"]` with hoisted dependencies.
- [x] **AC-2 (Shared Base Configurations)**: Shared `tsconfig.base.json` provides strict TypeScript compiler options extended by apps and packages.
- [x] **AC-3 (Independent Ports & Dev Servers)**: `main-app` runs on `http://localhost:5173` and `admin-app` runs on `http://localhost:5174`.

### 3.2 Shared Packages
- [x] **AC-4 (Design System `@graphfolio/ui`)**: Implements atomic dark glassmorphism components (`Button`, `Modal`, `Card`, `Badge`, `Table`, `Input`, `Select`) and precision currency/percentage formatters.
- [x] **AC-5 (API Client `@graphfolio/api-client`)**: Provides GenQL client factory with credentials inclusion and dynamic in-memory access token injection.

### 3.3 Application Separation & Bundle Independence
- [x] **AC-6 (Clean Bundle Isolation)**: Production build of `main-app` contains zero code or routes from `admin-app`.
- [x] **AC-7 (Makefile Automation)**: Root Makefile provides targets `make run-web`, `make run-admin`, `make run-all-web`, `make build-web`, and `make install-web`.

---

## 4. Technical & Implementation Notes

- **Workspace Layout**:
  - `web/apps/main-app/`
  - `web/apps/admin-app/`
  - `web/packages/ui/`
  - `web/packages/api-client/`
- **Build Tooling**: Vite with TypeScript path aliases for instant HMR without prior package build steps during local development.
- **Linters**: Oxlint configured across all packages and apps.
