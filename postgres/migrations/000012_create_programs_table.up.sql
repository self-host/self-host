BEGIN;

CREATE TABLE programs (
        uuid UUID NOT NULL DEFAULT uuid_generate_v4 () PRIMARY KEY,
        name TEXT NOT NULL,
        type TEXT NOT NULL DEFAULT 'program',
        state TEXT NOT NULL DEFAULT 'inactive',
        schedule TEXT NOT NULL DEFAULT '0s',
        deadline INTEGER NOT NULL DEFAULT 1000,
        language TEXT NOT NULL DEFAULT 'tengo',

        tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[]
);

CREATE TABLE program_code_revisions (
        program_uuid UUID NOT NULL REFERENCES programs(uuid) ON DELETE CASCADE,
        revision INTEGER NOT NULL DEFAULT 0,
        created TIMESTAMPTZ NOT NULL DEFAULT now(),
        signed TIMESTAMPTZ,
        created_by UUID REFERENCES users(uuid) ON DELETE SET NULL,
        signed_by UUID REFERENCES users(uuid) ON DELETE SET NULL,
	code BYTEA,
        checksum BYTEA,
        PRIMARY KEY (program_uuid, revision)
);

-- Must use the array operators for this index to work
-- https://www.postgresql.org/docs/current/functions-array.html#ARRAY-OPERATORS-TABLE
-- fastupdate = false to spread out the load
CREATE INDEX programs_tags_idx ON programs USING GIN("tags") WITH (fastupdate = false);

COMMIT;
