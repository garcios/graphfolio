-- Migration: 000003_add_auth_and_sessions.down.sql

DROP TABLE IF EXISTS users.sessions;

ALTER TABLE users.users
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS last_login_at;
