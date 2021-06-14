-- name: ExistsProgram :one
SELECT COUNT(*) AS count
FROM programs
WHERE programs.uuid = sqlc.arg(uuid);

-- name: CreateProgram :one
WITH p AS (
	INSERT INTO programs (name, type, state, schedule, deadline, language, tags)
	VALUES (
		sqlc.arg(name),
		sqlc.arg(type),
		sqlc.arg(state),
		sqlc.arg(schedule),
		sqlc.arg(deadline),
		sqlc.arg(language),
		sqlc.arg(tags)
	) RETURNING *
), grp AS (
	SELECT groups.uuid
	FROM groups, user_groups
	WHERE user_groups.group_uuid = groups.uuid
	AND user_groups.user_uuid = sqlc.arg(created_by)
	AND groups.uuid = (
		SELECT users.uuid
		FROM users
		WHERE users.name = groups.name
	)
	LIMIT 1
), grp_policies AS (
	INSERT INTO group_policies(group_uuid, priority, effect, action, resource)
	VALUES (
		(SELECT uuid FROM grp), 0, 'allow', 'create','programs/'||(SELECT uuid FROM p)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'read','programs/'||(SELECT uuid FROM p)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'update','programs/'||(SELECT uuid FROM p)||'/%'
	), (
		(SELECT uuid FROM grp), 0, 'allow', 'delete','programs/'||(SELECT uuid FROM p)||'/%'
	)
)
SELECT *
FROM p LIMIT 1;

-- name: CreateCodeRevision :one
INSERT INTO program_code_revisions (program_uuid, revision, created_by, code, checksum)
VALUES (
	sqlc.arg(program_uuid),
	COALESCE((
		SELECT MAX(pcr.revision) + 1
		FROM program_code_revisions AS pcr
		WHERE pcr.program_uuid = sqlc.arg(program_uuid)
	), 0)::INTEGER,
	sqlc.arg(created_by),
	sqlc.arg(code),
	sha256(sqlc.arg(code))
) RETURNING
	program_uuid,
	revision,
	created,
	created_by,
	signed,
	signed_by,
	encode(checksum, 'hex') AS checksum;

-- name: FindPrograms :many
WITH usr AS (
	SELECT users.uuid
	FROM users, user_tokens
	WHERE user_tokens.user_uuid = users.uuid
	AND user_tokens.token_hash = sha256(sqlc.arg(token))
	LIMIT 1
), policies AS (
	SELECT group_policies.effect, group_policies.priority, group_policies.resource
	FROM group_policies, user_groups
	WHERE user_groups.group_uuid = group_policies.group_uuid
	AND user_groups.user_uuid = (SELECT uuid FROM usr)
	AND action = 'read'
)
SELECT
	*
FROM programs
WHERE 'programs/'||programs.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
EXCEPT
SELECT
	*
FROM programs
WHERE 'programs/'||programs.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindProgramsByTags :many
WITH usr AS (
	SELECT users.uuid
	FROM users, user_tokens
	WHERE user_tokens.user_uuid = users.uuid
	AND user_tokens.token_hash = sha256(sqlc.arg(token))
	LIMIT 1
), policies AS (
	SELECT group_policies.effect, group_policies.priority, group_policies.resource
	FROM group_policies, user_groups
	WHERE user_groups.group_uuid = group_policies.group_uuid
	AND user_groups.user_uuid = (SELECT uuid FROM usr)
	AND action = 'read'
)
SELECT *
FROM programs
WHERE 'programs/'||programs.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'allow')
)
AND sqlc.arg(tags) && programs.tags
EXCEPT
SELECT *
FROM programs
WHERE 'programs/'||programs.uuid LIKE ANY(
	(SELECT resource FROM policies WHERE effect = 'deny')
)
AND sqlc.arg(tags) && programs.tags
ORDER BY name
LIMIT sqlc.arg(arg_limit)::BIGINT
OFFSET sqlc.arg(arg_offset)::BIGINT
;

-- name: FindAllRoutineRevisions :many
WITH p AS (
	SELECT
        	pcr.*,
		programs.*,
		RANK () OVER (
			PARTITION BY pcr.program_uuid
			ORDER BY revision DESC
		) revision_rank
	FROM program_code_revisions AS pcr
	INNER JOIN programs ON programs.uuid = pcr.program_uuid
	WHERE pcr.signed IS NOT NULL
	AND programs.type IN ('routine', 'webhook')
	AND programs.state = 'active'
)
SELECT
	p.name,
	p.program_uuid,
        p.type,
        p.schedule,
        p.deadline,
        p.language,
        p.revision,
        p.code,
        encode(p.checksum, 'hex') AS checksum
FROM p
WHERE revision_rank = 1;

-- name: FindAllModules :many
SELECT
	programs.name,
	pcr.program_uuid,
        programs.type,
        programs.schedule,
        programs.deadline,
        programs.language,
        pcr.revision,
        pcr.code,
	encode(pcr.checksum, 'hex') AS checksum
FROM program_code_revisions AS pcr, programs
WHERE programs.uuid = pcr.program_uuid
AND programs.state = 'active'
AND programs.type = 'module'
AND pcr.signed IS NOT NULL
ORDER BY
	pcr.program_uuid,
	pcr.revision DESC;


-- name: FindProgramByUUID :one
SELECT
	*
FROM programs
WHERE programs.uuid = sqlc.arg(uuid)
LIMIT 1;

-- name: FindProgramCodeRevisions :many
SELECT
	revision, created, created_by, signed, signed_by, encode(checksum, 'hex') AS checksum
FROM
	program_code_revisions
WHERE
	program_uuid = sqlc.arg(program_uuid)
ORDER BY revision ASC;

-- name: GetNamedModuleCodeAtRevision :one
SELECT
	code
FROM program_code_revisions, programs
WHERE program_uuid = programs.uuid
AND programs.name = sqlc.arg(name)
AND programs.type = 'module'
AND programs.state = 'active'
AND programs.language = sqlc.arg(language)
AND program_code_revisions.signed IS NOT NULL
AND program_code_revisions.revision = sqlc.arg(revision)
LIMIT 1;

-- name: GetNamedModuleCodeAtHead :one
SELECT
	code
FROM program_code_revisions, programs
WHERE program_uuid = programs.uuid
AND programs.name = sqlc.arg(name)
AND programs.type = 'module'
AND programs.state = 'active'
AND programs.language = sqlc.arg(language)
AND program_code_revisions.signed IS NOT NULL
ORDER BY revision DESC
LIMIT 1;

-- name: GetProgramCodeAtRevision :one
SELECT
	code
FROM program_code_revisions
WHERE program_uuid = sqlc.arg(program_uuid)
AND revision = sqlc.arg(revision)
LIMIT 1;

-- name: GetProgramCodeAtHead :one
SELECT
	code, revision
FROM program_code_revisions
WHERE program_uuid = sqlc.arg(program_uuid)
ORDER BY revision DESC
LIMIT 1;

-- name: GetSignedProgramCodeAtHead :one
SELECT
	code, revision
FROM program_code_revisions
WHERE program_uuid = sqlc.arg(program_uuid)
AND signed IS NOT NULL
ORDER BY revision DESC
LIMIT 1;

-- name: SetProgramNameByUUID :execrows
UPDATE programs
SET name = sqlc.arg(name)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SetProgramTypeByUUID :execrows
UPDATE programs
SET type = sqlc.arg(type)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SetProgramStateByUUID :execrows
UPDATE programs
SET state = sqlc.arg(state)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SetProgramScheduleByUUID :execrows
UPDATE programs
SET schedule = sqlc.arg(schedule)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SetProgramDeadlineByUUID :execrows
UPDATE programs
SET deadline = sqlc.arg(deadline)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SetProgramLanguageByUUID :execrows
UPDATE programs
SET language = sqlc.arg(language)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SetProgramTags :execrows
UPDATE programs
SET tags = sqlc.arg(tags)
WHERE programs.uuid = sqlc.arg(uuid);

-- name: SignProgramCodeRevision :execrows
UPDATE program_code_revisions
SET signed_by = sqlc.arg(signed_by), signed = now()
WHERE program_uuid = sqlc.arg(program_uuid)
AND revision = sqlc.arg(revision);

-- name: DeleteProgram :execrows
DELETE FROM programs
WHERE programs.uuid = sqlc.arg(uuid);

-- name: DeleteProgramCodeRevision :execrows
DELETE FROM program_code_revisions
WHERE program_uuid = sqlc.arg(program_uuid)
AND revision = sqlc.arg(revision);
