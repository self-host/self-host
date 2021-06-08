BEGIN;

CREATE TYPE thing_state AS ENUM ('active', 'inactive', 'passive', 'archived');
CREATE TYPE account_state AS ENUM ('active', 'inactive', 'archived');

CREATE FUNCTION array_distinct(anyarray) RETURNS anyarray AS $BODY$
  SELECT array_agg(DISTINCT x) FROM unnest($1) t(x);
$BODY$ LANGUAGE SQL IMMUTABLE;

CREATE FUNCTION generate_token(integer) RETURNS TEXT AS $BODY$
  SELECT string_agg (substr('abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789', ceil (random() * 62)::integer, 1), '')
  FROM generate_series(1, $1);
$BODY$ LANGUAGE SQL IMMUTABLE;

COMMIT;
