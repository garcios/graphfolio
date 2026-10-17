-- Migration: 000004_add_role_and_password_policy.up.sql
-- Description: Add role check constraint, must_change_password flag, and auto-provision admin user

-- 1. Add role and must_change_password columns to users.users
ALTER TABLE users.users
    ADD COLUMN IF NOT EXISTS role                 text    NOT NULL DEFAULT 'USER',
    ADD COLUMN IF NOT EXISTS must_change_password boolean NOT NULL DEFAULT false;

-- 2. Add check constraint for permitted roles
ALTER TABLE users.users
    DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users.users
    ADD CONSTRAINT users_role_check CHECK (role IN ('USER', 'ADMIN'));

-- 3. High-performance index for role filtering
CREATE INDEX IF NOT EXISTS users_role_idx ON users.users(role);

-- 4. Auto-provision default administrator
-- Email: admin@graphfolio.app
-- Default Password: Password123!
INSERT INTO users.users (
    id,
    email,
    password_hash,
    display_name,
    base_currency,
    theme,
    role,
    must_change_password,
    is_active
)
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
ON CONFLICT (lower(email)) DO UPDATE
SET role = 'ADMIN'
WHERE users.users.role != 'ADMIN';
