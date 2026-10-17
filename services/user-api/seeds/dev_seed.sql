-- GraphFolio User API Development Seed Data
-- Idempotent: inserts or updates demo user record.
-- Password: Password123!

BEGIN;

INSERT INTO users.users (id, email, password_hash, display_name, base_currency, theme, role, must_change_password, is_active)
VALUES (
    '018f0000-0000-7000-8000-000000000001',
    'oscar@graphfolio.app',
    '$argon2id$v=19$m=65536,t=3,p=4$z48VmX+qFO9X2FkAfV/Z6Q$WW1NOG+O6TwEwlSDkDIfae3Xi0xOePDpNg+rd7lkWsw',
    'Oscar Garcia',
    'AUD',
    'DARK',
    'USER',
    false,
    true
)
ON CONFLICT (id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    display_name = EXCLUDED.display_name,
    base_currency = EXCLUDED.base_currency,
    theme = EXCLUDED.theme,
    role = EXCLUDED.role,
    must_change_password = EXCLUDED.must_change_password,
    is_active = EXCLUDED.is_active;

-- Default Administrator
INSERT INTO users.users (id, email, password_hash, display_name, base_currency, theme, role, must_change_password, is_active)
VALUES (
    '018f0000-0000-7000-8000-000000000002',
    'admin@graphfolio.app',
    '$argon2id$v=19$m=65536,t=3,p=4$z48VmX+qFO9X2FkAfV/Z6Q$WW1NOG+O6TwEwlSDkDIfae3Xi0xOePDpNg+rd7lkWsw',
    'System Administrator',
    'USD',
    'DARK',
    'ADMIN',
    true,
    true
)
ON CONFLICT (lower(email)) DO NOTHING;

COMMIT;
