BEGIN;

ALTER TABLE user_tokens
ADD COLUMN created TIMESTAMPTZ NOT NULL DEFAULT NOW();

COMMIT;