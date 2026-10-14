# User Story: User Password Change & Session Invalidation

> **Story ID**: US-AUTH-102  
> **Epic**: Identity, Authentication & Credential Security  
> **Target**: Primary Investor Application (`web/apps/main-app`), BFF GraphQL Orchestrator (`bff/`), User Domain Microservice (`services/user-api`), Proto (`proto/user/v1/`), Database (`users.users`, `users.sessions`)  
> **Status**: **DONE**  
> **Implementation Plan**: [user-change-password-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/user-change-password-implementation-plan.md)  
> **Security Baseline**: OWASP Application Security Verification Standard (ASVS v4.0 L2), NIST SP 800-63B Section 5.1.1.2 (Memorized Secret Verifiers), RFC 9106 (Argon2id)  
> **Related Documentation**: [user-registration-and-login.md](file:///Users/oscargarcia/workspace/graphfolio/docs/user-stories/user-registration-and-login.md), [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)  

---

## 1. Story Statements

### 1.1 Password Change & Re-Authentication
> **As an** authenticated investor in GraphFolio,  
> **I want to** change my account password by providing my current password and a new, compliant password,  
> **So that** I can maintain strong personal security hygiene, rotate credentials periodically, or recover from potential credential exposure without losing access to my portfolio history.

### 1.2 Session Invalidation on Other Devices
> **As an** investor changing my credentials,  
> **I want to** have the option to revoke all other active sessions across devices,  
> **So that** any previous unattended logins, background sessions, or compromised tokens on other devices are terminated immediately while my current active session remains uninterrupted.

---

## 2. Business Context & Security Requirements

1. **Defense Against Unattended Session Takeover**:
   - Re-authenticating with the current password ensures that an unauthorized party with temporary physical access to an unlocked workstation cannot alter the account credentials.
2. **NIST SP 800-63B & OWASP ASVS v4.0 Compliance**:
   - Enforce minimum length ($\ge 10$ characters) and character diversity (uppercase, lowercase, number, special character).
   - Reject new passwords identical to the current password.
   - Re-hash new passwords using RFC 9106 Argon2id ($m=64\text{MB}, t=3, p=4$).
3. **Session Revocation Hygiene**:
   - If selected (default `true`), invalidate all refresh tokens in `users.sessions` belonging to the user where `refresh_token_hash != current_session_token_hash`.

---

## 3. Acceptance Criteria

### 3.1 Credential Verification & Validation
- [x] **AC-1 (Current Password Check)**: Submitting an invalid current password fails with `codes.Unauthenticated` and displays a clear error message `"Current password is incorrect"`.
- [x] **AC-2 (Complexity Enforcement)**: The new password must satisfy all 5 criteria evaluated by the live `PasswordStrengthMeter` component (length $\ge 10$, uppercase, lowercase, digit, special character).
- [x] **AC-3 (Identical Password Rejection)**: If the new password matches the current password, submission is blocked with an informative message.
- [x] **AC-4 (Confirmation Match)**: The modal verifies that the "Confirm New Password" field matches the "New Password" before enabling the submit button.

### 3.2 Backend Security & Persistence
- [x] **AC-5 (Argon2id Re-Hashing)**: Upon successful verification, the new password is hashed with Argon2id ($m=65536, t=3, p=4$) and updated in `users.users (password_hash, updated_at)`.
- [x] **AC-6 (Session Invalidation)**: If "Sign out of all other devices" is selected, all sessions in `users.sessions` for the user except the current session are revoked (`is_revoked = true`).
- [x] **AC-7 (Timing Attack Mitigation)**: Invalid user lookups execute dummy hash verification to preserve constant-time execution.

### 3.3 User Experience & UI
- [x] **AC-8 (Modal Launch)**: Accessible from the `UserProfileMenu` dropdown via a `Change Password` action item.
- [x] **AC-9 (Dark Glassmorphic UI)**: Styled to match `@graphfolio/ui` with password visibility toggles, loading state, error banners, and success confirmation.

