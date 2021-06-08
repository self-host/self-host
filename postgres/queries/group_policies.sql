-- name: FindPoliciesByGroup :many
SELECT * FROM group_policies
WHERE group_policies.group_uuid = sqlc.arg(uuid)
ORDER BY priority;

-- name: FindPoliciesByUser :many
SELECT group_policies.*
FROM group_policies, user_groups
WHERE user_groups.group_uuid = group_policies.group_uuid
AND user_groups.user_uuid = sqlc.arg(uuid)
ORDER BY priority;
