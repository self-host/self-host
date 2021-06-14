BEGIN;

CREATE TABLE things (
  uuid UUID NOT NULL DEFAULT uuid_generate_v4 () PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT,
  state thing_state DEFAULT 'inactive' NOT NULL,

  created_by UUID REFERENCES users(uuid) ON DELETE SET NULL,

  tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[]
);

CREATE INDEX things_created_by_idx ON things(created_by);

-- Must use the array operators for this index to work
-- https://www.postgresql.org/docs/current/functions-array.html#ARRAY-OPERATORS-TABLE
-- fastupdate = false to spread out the load
CREATE INDEX things_tags_idx ON things USING GIN("tags") WITH (fastupdate = false);

COMMIT;
