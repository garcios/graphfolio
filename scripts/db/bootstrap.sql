-- GraphFolio local database bootstrap.
--
-- Idempotent: safe to run repeatedly. Must be run as a local superuser:
--   psql "postgres://$USER@localhost:5432/postgres" -v db=graphfolio -f scripts/db/bootstrap.sql
--
-- Creates the cluster-wide service roles, the database (named by :db, default
-- "graphfolio") and one schema per service. Tables are created by each
-- service's golang-migrate migrations, never here.

\set ON_ERROR_STOP on

\if :{?db}
\else
  \set db graphfolio
\endif

-- Fail fast on an unsupported server version. PG 18 provides uuidv7() natively;
-- on PG 17 the first migration of each schema installs a compatible fallback.
DO $$
BEGIN
  IF current_setting('server_version_num')::int < 170000 THEN
    RAISE EXCEPTION 'PostgreSQL 17+ required (18+ recommended), found %',
      current_setting('server_version');
  END IF;
END
$$;

-- Roles are cluster-wide: create only if missing.
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'portfolio_svc') THEN
    CREATE ROLE portfolio_svc LOGIN PASSWORD 'portfolio';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user_svc') THEN
    CREATE ROLE user_svc LOGIN PASSWORD 'user';
  END IF;
END
$$;

-- CREATE DATABASE cannot run inside a DO block / transaction, so use \gexec.
SELECT format('CREATE DATABASE %I', :'db')
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'db') \gexec

\connect :"db"

REVOKE ALL ON DATABASE :"db" FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db" TO portfolio_svc, user_svc;

CREATE SCHEMA IF NOT EXISTS portfolio AUTHORIZATION portfolio_svc;
CREATE SCHEMA IF NOT EXISTS users     AUTHORIZATION user_svc;

-- Each service role only ever sees its own schema.
ALTER ROLE portfolio_svc IN DATABASE :"db" SET search_path = portfolio;
ALTER ROLE user_svc      IN DATABASE :"db" SET search_path = users;

\echo 'Bootstrap complete for database' :"db"
