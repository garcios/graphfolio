-- GraphFolio User API Development Seed Data
-- Idempotent: inserts or updates demo user record.

BEGIN;

INSERT INTO users.users (id, email, display_name, base_currency, theme)
VALUES (
    '018f0000-0000-7000-8000-000000000001',
    'demo@graphfolio.internal',
    'Oscar Garcia',
    'USD',
    'DARK'
)
ON CONFLICT (id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    base_currency = EXCLUDED.base_currency,
    theme = EXCLUDED.theme;

COMMIT;
