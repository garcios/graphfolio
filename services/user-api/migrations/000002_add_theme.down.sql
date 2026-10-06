-- 000002_add_theme.down.sql
ALTER TABLE users.users DROP COLUMN IF EXISTS theme;
