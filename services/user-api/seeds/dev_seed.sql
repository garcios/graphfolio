-- GraphFolio User API Development Seed Data
-- Idempotent: inserts or updates demo user record.
-- Password: Password123!

BEGIN;

INSERT INTO users.users (id, email, password_hash, display_name, base_currency, theme, is_active)
VALUES (
    '018f0000-0000-7000-8000-000000000001',
    'oscar@graphfolio.com',
    '$argon2id$v=19$m=65536,t=3,p=4$z48VmX+qFO9X2FkAfV/Z6Q$WW1NOG+O6TwEwlSDkDIfae3Xi0xOePDpNg+rd7lkWsw',
    'Oscar Garcia',
    'AUD',
    'DARK',
    true
)
ON CONFLICT (id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    display_name = EXCLUDED.display_name,
    base_currency = EXCLUDED.base_currency,
    theme = EXCLUDED.theme,
    is_active = EXCLUDED.is_active;

COMMIT;
