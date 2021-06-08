BEGIN;

DROP FUNCTION user_has_access;

DROP TABLE group_policies;
DROP TABLE user_groups;
DROP TABLE groups;

DROP TYPE policy_effect;
DROP TYPE policy_action;

COMMIT;
