-- name: GetUserByName :one
SELECT
    *
FROM
    users
WHERE
    name = @name
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByID :one
SELECT
    *
FROM
    users
WHERE
    id = @user_id
    AND deleted_at IS NULL;

-- name: GetPermissions :many
SELECT
    p.*
FROM
    users u
    LEFT JOIN user_group_memberships ugm ON ugm.user_id = u.id
    LEFT JOIN user_groups ug ON ug.id = ugm.user_group_id
    JOIN permissions p ON p.user_group_id = ugm.user_group_id
        OR p.user_id = u.id
WHERE
    u.id = @user_id
    AND (ug.id IS NULL
        OR ug.deleted_at IS NULL);

-- name: GetPermissionsForMembershipObject :many
SELECT
    p.id,
    p.kind,
    p.user_id,
    p.user_group_id,
    p.service_id,
    p.feature_id,
    p.key_id,
    p.variation_context_id,
    p.permission,
    s.name AS service_name,
    f.name AS feature_name,
    k.name AS key_name
FROM
    permissions p
    JOIN services s ON s.id = p.service_id
    LEFT JOIN features f ON f.id = p.feature_id
    LEFT JOIN keys k ON k.id = p.key_id
WHERE (sqlc.narg('user_id')::bigint IS NULL
    OR p.user_id = sqlc.narg('user_id')::bigint
    OR p.user_group_id IN (
        SELECT
            ugm.user_group_id
        FROM
            user_group_memberships ugm
            JOIN user_groups ug ON ug.id = ugm.user_group_id
        WHERE
            ugm.user_id = sqlc.narg('user_id')::bigint
            AND ug.deleted_at IS NULL))
AND (sqlc.narg('group_id')::bigint IS NULL
    OR p.user_group_id = sqlc.narg('group_id')::bigint)
ORDER BY
    p.id ASC;

-- name: GetPermissionsForEntity :many
SELECT
    p.id,
    p.permission,
    u.id AS user_id,
    u.name AS user_name,
    ug.id AS group_id,
    ug.name AS group_name
FROM
    permissions p
    LEFT JOIN users u ON u.id = p.user_id
    LEFT JOIN user_groups ug ON ug.id = p.user_group_id
WHERE
    service_id = @service_id
    AND feature_id IS NOT DISTINCT FROM @feature_id
    AND key_id IS NOT DISTINCT FROM @key_id
    AND variation_context_id IS NOT DISTINCT FROM @variation_context_id
ORDER BY
    p.id ASC;

-- name: GetPermissionByID :one
SELECT
    *
FROM
    permissions
WHERE
    id = @id;

-- name: GetPermission :one
SELECT
    *
FROM
    permissions
WHERE
    user_id IS NOT DISTINCT FROM @user_id
    AND user_group_id IS NOT DISTINCT FROM @user_group_id
    AND service_id = @service_id
    AND feature_id IS NOT DISTINCT FROM @feature_id
    AND key_id IS NOT DISTINCT FROM @key_id
    AND variation_context_id IS NOT DISTINCT FROM @variation_context_id;

-- name: GetUserGroups :many
SELECT
    ug.id,
    ug.name
FROM
    user_groups ug
    JOIN user_group_memberships ugm ON ugm.user_group_id = ug.id
WHERE
    ugm.user_id = @user_id;

-- name: GetGroupUsers :many
SELECT
    u.id,
    u.name,
    COUNT(*) OVER ()::integer AS total_count
FROM
    users u
    JOIN user_group_memberships ugm ON ugm.user_id = u.id
WHERE
    ugm.user_group_id = @id
ORDER BY
    u.name ASC
LIMIT sqlc.arg('limit')::integer OFFSET sqlc.arg('offset')::integer;

-- name: GetUsersAndGroups :many
WITH filtered_users AS (
    SELECT
        u.id,
        u.name,
        CASE WHEN u.global_administrator THEN
            'global_administrator'
        ELSE
            'user'
        END AS type
    FROM
        users u
    WHERE
        u.deleted_at IS NULL
        AND (sqlc.narg('name')::text IS NULL
            OR name ILIKE sqlc.narg('name')::text || '%')
        AND (sqlc.narg('type')::text IS NULL
            OR sqlc.narg('type')::text = 'user')
),
filtered_groups AS (
    SELECT
        ug.id,
        ug.name,
        'group' AS type
    FROM
        user_groups ug
    WHERE
        deleted_at IS NULL
        AND (sqlc.narg('name')::text IS NULL
            OR name ILIKE sqlc.narg('name')::text || '%')
        AND (sqlc.narg('type')::text IS NULL
            OR sqlc.narg('type') = 'group')
),
combined AS (
    SELECT
        *
    FROM
        filtered_users
    UNION ALL
    SELECT
        *
    FROM
        filtered_groups
)
SELECT
    combined.*,
    COUNT(*) OVER ()::integer AS total_count
    FROM
        combined
    ORDER BY
        combined.name ASC
    LIMIT sqlc.arg('limit')::integer OFFSET sqlc.arg('offset')::integer;

-- name: CreateUser :one
INSERT INTO users(name, password, global_administrator, created_at)
    VALUES (@name, @password, @global_administrator, now())
RETURNING
    id;

-- name: CreateUsers :copyfrom
INSERT INTO users(name, password, global_administrator, created_at)
    VALUES ($1, $2, $3, $4);

-- name: UpdateUser :exec
UPDATE
    users
SET
    global_administrator = @global_administrator,
    updated_at = now()
WHERE
    id = @id;

-- name: DeleteUser :exec
UPDATE
    users
SET
    deleted_at = now()
WHERE
    id = @id;

-- name: CreatePermission :one
INSERT INTO permissions(user_id, user_group_id, kind, service_id, feature_id, key_id, permission, variation_context_id)
    VALUES (@user_id, @user_group_id, @kind, @service_id, @feature_id, @key_id, @permission, @variation_context_id)
RETURNING
    id;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE id = @id;

-- name: GetGroupByID :one
SELECT
    *
FROM
    user_groups
WHERE
    id = @id;

-- name: GetGroupIDByName :one
SELECT
    id
FROM
    user_groups
WHERE
    name = @name;

-- name: GetUserGroupMembership :one
SELECT
    *
FROM
    user_group_memberships
WHERE
    user_id = @user_id
    AND user_group_id = @user_group_id;

-- name: CreateGroup :one
INSERT INTO user_groups(name, created_at)
    VALUES (@name, now())
RETURNING
    id;

-- name: DeleteGroup :exec
UPDATE
    user_groups
SET
    deleted_at = now()
WHERE
    id = @id;

-- name: CreateUserGroupMembership :exec
INSERT INTO user_group_memberships(user_id, user_group_id, created_at)
    VALUES (@user_id, @user_group_id, now());

-- name: DeleteUserGroupMembership :exec
DELETE FROM user_group_memberships
WHERE user_id = @user_id
    AND user_group_id = @user_group_id;

