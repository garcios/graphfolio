# User Story: User Preferences & Multi-Currency Display

> **Target**: User Service (`services/user-api`), BFF (`bff/`), and Primary Investor Application (`web/apps/main-app`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-preferences-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #12)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L22)

---

## 1. User Story

> **As an** international investor holding foreign assets across different markets,  
> **I want to** configure my personal profile preferences, theme, and preferred portfolio display currency (e.g., USD, EUR, GBP, AUD, CAD, JPY, CHF),  
> **So that** my entire portfolio valuation, cash balances, and performance metrics are dynamically consolidated and reported in my home currency.

---

## 2. Acceptance Criteria

- [ ] **Preferences Modal**: Provide a modal accessible from the user navigation avatar allowing updates to display name, theme (`DARK`, `LIGHT`, `SYSTEM`), and display currency.
- [ ] **Multi-Currency Conversion**: When a non-native display currency is chosen, all holding valuations, daily returns, and cash balances are dynamically converted using official FX rates.
- [ ] **Currencies Supported**: Support standard major currencies (`USD`, `EUR`, `GBP`, `AUD`, `CAD`, `JPY`, `CHF`) with proper ISO-4217 symbol formatting.
- [ ] **Persistence**: User preferences are saved in the `users.user_preferences` database schema via `user-api` gRPC service.
- [ ] **Visual Distinction**: Clearly display the active reporting currency code next to all top-level money metrics.

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [user-preferences-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-preferences-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
