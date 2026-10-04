-- GraphFolio local database teardown.
--
--   psql "postgres://$USER@localhost:5432/postgres" -v db=graphfolio -f scripts/db/teardown.sql
--
-- Drops the database named by :db (default "graphfolio"). The cluster-wide
-- service roles are kept unless invoked with -v drop_roles=1, because other
-- databases on the same local server (e.g. graphfolio_test) may still use them.

\set ON_ERROR_STOP on

\if :{?db}
\else
  \set db graphfolio
\endif

DROP DATABASE IF EXISTS :"db" WITH (FORCE);

\if :{?drop_roles}
  DROP ROLE IF EXISTS portfolio_svc;
  DROP ROLE IF EXISTS user_svc;
\endif

\echo 'Teardown complete for database' :"db"
