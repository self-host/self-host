BEGIN;

DROP FUNCTION IF EXISTS array_distinct;
DROP FUNCTION IF EXISTS generate_token;

DROP TYPE thing_state;
DROP TYPE account_state;

COMMIT;
