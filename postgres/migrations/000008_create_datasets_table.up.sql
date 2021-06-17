BEGIN;

CREATE OR REPLACE FUNCTION update_dataset_content_change()
RETURNS TRIGGER AS $$
BEGIN
   NEW.size = length(NEW.content);
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE datasets (
	uuid UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
	name TEXT NOT NULL,
	format TEXT NOT NULL,
	content BYTEA DEFAULT ''::bytea NOT NULL,
	checksum BYTEA DEFAULT ''::bytea NOT NULL,
	size INTEGER DEFAULT 0 NOT NULL,
	belongs_to UUID REFERENCES things(uuid) ON DELETE SET NULL,
	created TIMESTAMPTZ DEFAULT NOW() NOT NULL,
	updated TIMESTAMPTZ DEFAULT NOW() NOT NULL,
	created_by UUID REFERENCES users(uuid) ON DELETE SET NULL,
	updated_by UUID REFERENCES users(uuid) ON DELETE SET NULL,
	tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[]
);

CREATE INDEX datasets_created_by_idx ON datasets(created_by);
CREATE INDEX datasets_updated_by_idx ON datasets(updated_by);
CREATE INDEX datasets_belong_to_idx ON datasets(belongs_to);

-- Must use the array operators for this index to work
-- https://www.postgresql.org/docs/current/functions-array.html#ARRAY-OPERATORS-TABLE
-- fastupdate = false to spread out the load
CREATE INDEX datasets_tags_idx ON datasets USING GIN("tags") WITH (fastupdate = false);

CREATE TRIGGER update_dataset_content_trigger
BEFORE UPDATE OF content ON datasets
FOR EACH ROW EXECUTE PROCEDURE
update_dataset_content_change();

COMMIT;
