BEGIN;

CREATE TABLE things (
  uuid UUID NOT NULL DEFAULT uuid_generate_v4 () PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT,
  state thing_state DEFAULT 'inactive' NOT NULL,

  created_by UUID REFERENCES users(uuid) ON DELETE SET NULL
);

CREATE INDEX things_created_by_idx ON things(created_by);

COMMIT;
