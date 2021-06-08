BEGIN;

DROP TRIGGER IF EXISTS before_insert_things_deps_trg ON thing_deps;
DROP FUNCTION deps_insert_trigger_func;

DROP TRIGGER IF EXISTS before_update_thing_deps_trg ON thing_deps;
DROP FUNCTION deps_update_trigger_func;

DROP TABLE thing_deps;

COMMIT;
