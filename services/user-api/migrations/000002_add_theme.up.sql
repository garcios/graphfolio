-- 000002_add_theme.up.sql
ALTER TABLE users.users
    ADD COLUMN IF NOT EXISTS theme text NOT NULL DEFAULT 'DARK' CHECK (theme IN ('DARK', 'LIGHT', 'SYSTEM'));
