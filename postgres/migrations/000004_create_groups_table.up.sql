BEGIN;

CREATE TYPE policy_effect AS ENUM ('allow', 'deny');
CREATE TYPE policy_action AS ENUM ('create', 'read', 'update', 'delete');



CREATE TABLE groups (
  uuid UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
);



CREATE TABLE user_groups (
  user_uuid UUID REFERENCES users(uuid) ON DELETE CASCADE,
  group_uuid UUID REFERENCES groups(uuid) ON DELETE CASCADE,

  PRIMARY KEY(user_uuid, group_uuid)
);

CREATE INDEX user_group_group_uuid_idx ON user_groups(group_uuid);

CREATE TABLE group_policies (
  uuid UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
  group_uuid UUID NOT NULL REFERENCES groups(uuid) ON DELETE CASCADE,
  priority INTEGER DEFAULT 0 NOT NULL,
  effect policy_effect DEFAULT 'deny' NOT NULL,

  action policy_action NOT NULL,
  resource TEXT NOT NULL,
  UNIQUE(group_uuid, priority, effect, action, resource)
);

CREATE INDEX group_policies_effect_grp_uuid_idx ON group_policies(group_uuid);
CREATE INDEX group_policies_effect_idx ON group_policies(effect);
CREATE INDEX group_policies_action_idx ON group_policies(action);
CREATE INDEX group_policies_resource_ops_idx ON group_policies(resource text_pattern_ops);
-- CREATE INDEX group_policies_resource_idx ON group_policies(resource);
-- CREATE INDEX group_policies_resource_idx ON group_policies USING GIST(resource gist_trgm_ops);

INSERT INTO groups(uuid, name)
VALUES
  ('00000000-0000-1000-8000-000000000000', 'root'),

  ('00000000-0000-1000-8000-000000000001', 'things'), -- C, R
  ('00000000-0000-1000-8000-000000000002', 'timeseries'), -- C, R
  ('00000000-0000-1000-8000-000000000003', 'tsdata'), -- C, R  
  ('00000000-0000-1000-8000-000000000004', 'datasets'),

  ('00000000-0000-1000-8000-000000000100', 'operator'),
  ('00000000-0000-1000-8000-000000000101', 'admin')
;

INSERT INTO user_groups(user_uuid, group_uuid)
VALUES
  ('00000000-0000-1000-8000-000000000000', '00000000-0000-1000-8000-000000000000') -- root user and root group
;

INSERT INTO group_policies(group_uuid, priority, effect, action, resource)
VALUES
  ('00000000-0000-1000-8000-000000000000', 0, 'allow', 'create', '%'), -- allow all on root group
  ('00000000-0000-1000-8000-000000000000', 0, 'allow', 'read', '%'),   -- allow all on root group
  ('00000000-0000-1000-8000-000000000000', 0, 'allow', 'update', '%'), -- allow all on root group
  ('00000000-0000-1000-8000-000000000000', 0, 'allow', 'delete', '%'), -- allow all on root group


  ('00000000-0000-1000-8000-000000000001', 0, 'allow', 'create', 'things'),
  ('00000000-0000-1000-8000-000000000001', 0, 'allow', 'read', 'things'),
  ('00000000-0000-1000-8000-000000000001', 0, 'allow', 'read', 'things/%'),
  ('00000000-0000-1000-8000-000000000001', 0, 'allow', 'update', 'things/%'),
  ('00000000-0000-1000-8000-000000000001', 0, 'allow', 'delete', 'things/%'),

  ('00000000-0000-1000-8000-000000000002', 0, 'allow', 'create', 'timeseries'),
  ('00000000-0000-1000-8000-000000000002', 0, 'allow', 'read', 'timeseries'),
  ('00000000-0000-1000-8000-000000000002', 0, 'allow', 'create', 'timeseries/%'),
  ('00000000-0000-1000-8000-000000000002', 0, 'allow', 'read', 'timeseries/%'),
  ('00000000-0000-1000-8000-000000000002', 0, 'allow', 'update', 'timeseries/%'),
  ('00000000-0000-1000-8000-000000000002', 0, 'allow', 'delete', 'timeseries/%'),

  ('00000000-0000-1000-8000-000000000003', 0, 'allow', 'create', 'tsdata'),
  ('00000000-0000-1000-8000-000000000003', 0, 'allow', 'read', 'tsdata'),
  ('00000000-0000-1000-8000-000000000003', 0, 'allow', 'read', 'tsdata/%'),
  ('00000000-0000-1000-8000-000000000003', 0, 'allow', 'update', 'tsdata/%'),
  ('00000000-0000-1000-8000-000000000003', 0, 'allow', 'delete', 'tsdata/%'),

  ('00000000-0000-1000-8000-000000000004', 0, 'allow', 'create', 'datasets'),
  ('00000000-0000-1000-8000-000000000004', 0, 'allow', 'read', 'datasets'),
  ('00000000-0000-1000-8000-000000000004', 0, 'allow', 'read', 'datasets/%'),
  ('00000000-0000-1000-8000-000000000004', 0, 'allow', 'update', 'datasets/%'),
  ('00000000-0000-1000-8000-000000000004', 0, 'allow', 'delete', 'datasets/%')
;

CREATE FUNCTION user_has_access(uid UUID, act policy_action, res TEXT) RETURNS boolean AS $$
	WITH policies AS (
	        SELECT group_policies.effect, group_policies.priority, group_policies.resource
        	FROM group_policies, user_groups
	        WHERE user_groups.group_uuid = group_policies.group_uuid
	        AND user_groups.user_uuid = $1
        	AND action = $2::policy_action
	), c AS (
		SELECT res AS resource
	), has_access AS (
		SELECT *
		FROM c
		WHERE c.resource LIKE ANY((SELECT resource FROM policies WHERE effect = 'allow'))
		EXCEPT
		SELECT *
		FROM c
		WHERE c.resource LIKE ANY((SELECT resource FROM policies WHERE effect = 'deny'))
	)
	SELECT COUNT(*) > 0 FROM has_access;
$$ LANGUAGE sql;

COMMIT;
