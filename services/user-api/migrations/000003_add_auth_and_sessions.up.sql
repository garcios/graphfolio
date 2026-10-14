-- Migration: 000003_add_auth_and_sessions.up.sql
-- Description: Add authentication credentials and session tracking to users schema

-- 1. Extend users.users table
ALTER TABLE users.users
    ADD COLUMN IF NOT EXISTS password_hash       text        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS is_active           boolean     NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS email_verified_at   timestamptz,
    ADD COLUMN IF NOT EXISTS last_login_at       timestamptz;

-- 2. Create session & refresh token tracking table
CREATE TABLE IF NOT EXISTS users.sessions (
    id                 uuid        PRIMARY KEY DEFAULT users.uuidv7(),
    user_id            uuid        NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    refresh_token_hash text        NOT NULL UNIQUE,
    family_id          uuid        NOT NULL DEFAULT users.uuidv7(),
    is_revoked         boolean     NOT NULL DEFAULT false,
    user_agent         text,
    ip_address         inet,
    expires_at         timestamptz NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

-- 3. High-performance indexes
CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON users.sessions(user_id) WHERE NOT is_revoked;
CREATE INDEX IF NOT EXISTS sessions_family_id_idx ON users.sessions(family_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON users.sessions(expires_at);
