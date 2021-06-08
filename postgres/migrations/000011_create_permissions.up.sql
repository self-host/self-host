/*
BEGIN;

CREATE TABLE api_endpoints (
  path TEXT NOT NULL,

  user_uuid UUID REFERENCES users(uuid) NOT NULL,
  group_uuid UUID REFERENCES groups(uuid) NOT NULL,
  ugo crud_perm DEFAULT 15 NOT NULL,

  UNIQUE(path, resource)
);

INSERT INTO api_endpoints(path, user_uuid, group_uuid, ugo)
VALUES
  'things/%/notify' ('00000000-0000-1000-8000-000000000000', '00000000-0000-1000-8000-000000000000', )
;

CREATE VIEW v_permissions AS
SELECT
'things' AS path, uuid, user_uuid, group_uuid, ugo
FROM things
UNION ALL
SELECT
'datasets' AS path, uuid, user_uuid, group_uuid, ugo
FROM datasets
UNION ALL
SELECT
'timeseries' AS path, uuid, user_uuid, group_uuid, ugo
FROM timeseries
UNION ALL
SELECT
path, NULL, user_uuid, group_uuid, ugo
FROM api_endpoints;

COMMIT;
*/
