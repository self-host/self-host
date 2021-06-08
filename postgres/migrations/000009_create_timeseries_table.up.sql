BEGIN;

CREATE TABLE timeseries (
  uuid UUID NOT NULL DEFAULT uuid_generate_v4 () PRIMARY KEY,
  thing_uuid UUID REFERENCES things(uuid) ON DELETE SET NULL,

  name TEXT NOT NULL,
  si_unit TEXT NOT NULL,
  lower_bound double precision,
  upper_bound double precision,

  created_by UUID REFERENCES users(uuid) ON DELETE SET NULL,

  tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[]
);

-- Must use the array operators for this index to work
-- https://www.postgresql.org/docs/current/functions-array.html#ARRAY-OPERATORS-TABLE
-- fastupdate = false to spread out the load
CREATE INDEX timeseries_tags_idx ON timeseries USING GIN("tags") WITH (fastupdate = false);

CREATE INDEX timeseries_created_by_idx ON timeseries(created_by);
-- CREATE INDEX tuneseries_thing_uuid_idx ON timeseries(thing_uuid);

COMMIT;
