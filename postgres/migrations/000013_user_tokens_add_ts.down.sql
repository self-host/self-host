BEGIN;

ALTER TABLE user_tokens DROP COLUMN created;

COMMIT;
