-- 000001_create_users.up.sql

DO $$
BEGIN
  IF current_setting('server_version_num')::int < 180000 THEN
    CREATE OR REPLACE FUNCTION users.uuidv7() RETURNS uuid AS $func$
    DECLARE
      unix_time_ms bytea;
      uuid_bytes bytea;
    BEGIN
      unix_time_ms := substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3 for 6);
      uuid_bytes := unix_time_ms || substring(uuid_send(gen_random_uuid()) from 7 for 10);
      uuid_bytes := set_byte(uuid_bytes, 6, (get_byte(uuid_bytes, 6) & 15) | 112);
      uuid_bytes := set_byte(uuid_bytes, 8, (get_byte(uuid_bytes, 8) & 63) | 128);
      RETURN encode(uuid_bytes, 'hex')::uuid;
    END;
    $func$ LANGUAGE plpgsql VOLATILE;
  END IF;
END $$;

CREATE TABLE users.users (
    id             uuid        PRIMARY KEY DEFAULT uuidv7(),
    email          text        NOT NULL,
    display_name   text        NOT NULL,
    base_currency  char(3)     NOT NULL DEFAULT 'USD' CHECK (base_currency ~ '^[A-Z]{3}$'),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_lower_uq ON users.users (lower(email));
