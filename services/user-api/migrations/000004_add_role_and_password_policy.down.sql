-- Migration: 000004_add_role_and_password_policy.down.sql
DELETE FROM users.users WHERE lower(email) = 'admin@graphfolio.app';
DROP INDEX IF EXISTS users.users_role_idx;
ALTER TABLE users.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users.users DROP COLUMN IF EXISTS must_change_password;
ALTER TABLE users.users DROP COLUMN IF EXISTS role;
