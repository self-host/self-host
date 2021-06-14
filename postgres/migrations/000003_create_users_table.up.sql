BEGIN;

CREATE TABLE users (
  uuid UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  state account_state DEFAULT 'active' NOT NULL
);

CREATE INDEX users_policy_map_idx ON users (('users/'||uuid) text_pattern_ops); -- Is this useful?

CREATE TABLE user_tokens (
  uuid UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
  user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
  name TEXT NOT NULL,
  token_hash BYTEA NOT NULL,
  created TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX user_tokens_user_uuid_idx ON user_tokens(user_uuid);

INSERT INTO users(uuid, name)
VALUES
  ('00000000-0000-1000-8000-000000000000', 'root')
;

--
-- This should not be here but should be managed by the setup-procedure
--
INSERT INTO user_tokens(user_uuid, name, token_hash)
VALUES
  ('00000000-0000-1000-8000-000000000000', 'Main admin token', sha256('root')) -- root
;

COMMIT;
