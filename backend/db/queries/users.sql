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

-- name: GetUserPermissions :many
SELECT
    up.id,
    up.service_id,
    up.feature_id,
    up.key_id,
    up.permission,
    up.variation_context_id
FROM
    user_permissions up
WHERE
    up.user_id = @user_id;

-- name: GetUsers :many
WITH filtered_users AS (
    SELECT
        *
    FROM
        users
    WHERE
        deleted_at IS NULL
        AND (sqlc.narg('name')::text IS NULL
            OR name ILIKE sqlc.narg('name')::text || '%'))
SELECT
    filtered_users.*,
    COUNT(*) OVER ()::integer AS total_count
FROM
    filtered_users
ORDER BY
    filtered_users.name ASC
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

-- name: CreatePermission :one
INSERT INTO user_permissions(user_id, kind, service_id, feature_id, key_id, permission, variation_context_id)
    VALUES (@user_id, @kind, @service_id, @feature_id, @key_id, @permission, @variation_context_id)
RETURNING
    id;

