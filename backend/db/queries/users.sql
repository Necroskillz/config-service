-- name: GetUserByName :one
SELECT *
FROM users
WHERE name = @name
    AND deleted_at IS NULL
LIMIT 1;
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = @user_id
    AND deleted_at IS NULL;
-- name: GetUserPermissions :many
SELECT up.id,
    up.service_id,
    up.feature_id,
    up.key_id,
    up.permission,
    up.variation_context_id
FROM user_permissions up
WHERE up.user_id = @user_id;
-- name: CreateUser :one
INSERT INTO users (name, password, global_administrator)
VALUES (@name, @password, @global_administrator)
RETURNING id;
-- name: CreatePermission :one
INSERT INTO user_permissions (user_id, service_id, feature_id, key_id, permission, variation_context_id)
VALUES (@user_id, @service_id, @feature_id, @key_id, @permission, @variation_context_id)
RETURNING id;
