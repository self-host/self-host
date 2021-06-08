BEGIN;

CREATE TABLE thing_deps (
  parent UUID REFERENCES things(uuid),
  child UUID REFERENCES things(uuid),

  PRIMARY KEY (parent, child),
  CHECK (parent <> child)
);

CREATE INDEX thing_deps_child_idx ON thing_deps(child);

---
-- Insert trigger
---
CREATE OR REPLACE FUNCTION deps_insert_trigger_func() RETURNS trigger AS $BODY$
    DECLARE
        results bigint;
    BEGIN
        WITH RECURSIVE p(id) AS (
            SELECT parent
                FROM thing_deps
                WHERE child=NEW.parent
            UNION
            SELECT parent
                FROM p, thing_deps d
                WHERE p.id = d.child
            )
        SELECT * INTO results
        FROM p
        WHERE id=NEW.child;

        IF FOUND THEN
            RAISE EXCEPTION 'circular dependencies are not allowed.';
        END IF;
        RETURN NEW;
    END;
$BODY$ LANGUAGE plpgsql;

CREATE TRIGGER before_insert_things_deps_trg BEFORE INSERT ON thing_deps
    FOR EACH ROW
    EXECUTE PROCEDURE deps_insert_trigger_func();

---
-- Update trigger
---
CREATE FUNCTION deps_update_trigger_func() RETURNS trigger AS $BODY$
    DECLARE
        results bigint;
    BEGIN
        WITH RECURSIVE p(id) AS (
            SELECT parent
                FROM thing_deps
                WHERE child=NEW.parent
                    AND NOT (child = OLD.child AND parent = OLD.parent) -- hide old row
            UNION
            SELECT parent
                FROM p, thing_deps d
                WHERE p.id = d.child
                    AND NOT (child = OLD.child AND parent = OLD.parent) -- hide old row
            )
        SELECT * INTO results
        FROM p
        WHERE id=NEW.child;

        IF FOUND THEN
            RAISE EXCEPTION 'circular dependencies are not allowed.';
        END IF;
        RETURN NEW;
    END;
$BODY$ LANGUAGE plpgsql;

CREATE TRIGGER before_update_thing_deps_trg BEFORE UPDATE ON thing_deps
    FOR EACH ROW
    EXECUTE PROCEDURE deps_update_trigger_func();

COMMIT;
