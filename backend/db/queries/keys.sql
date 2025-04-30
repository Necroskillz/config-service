-- name: GetKey :one
SELECT k.*,
    vt.editor as value_type_editor,
    vt.name as value_type_name
FROM keys k
    JOIN value_types vt ON vt.id = k.value_type_id
WHERE k.id = @key_id
LIMIT 1;
-- name: GetActiveKeysForFeatureVersion :many
SELECT k.*,
    vt.editor as value_type_editor,
    vt.name as value_type_name
FROM keys k
    JOIN value_types vt ON vt.id = k.value_type_id
WHERE k.feature_version_id = @feature_version_id
    AND (
        (
            k.valid_from IS NOT NULL
            AND k.valid_to IS NULL
            AND NOT EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'key'
                    AND csc.type = 'delete'
                    AND csc.key_id = k.id
                LIMIT 1
            )
        )
        OR (
            k.valid_from IS NULL
            AND EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'key'
                    AND csc.type = 'create'
                    AND csc.key_id = k.id
                LIMIT 1
            )
        )
    )
ORDER BY k.name;
-- name: GetValueTypes :many
SELECT *
FROM value_types;
-- name: CreateKey :one
INSERT INTO keys (
        name,
        description,
        value_type_id,
        feature_version_id
    )
VALUES (
        @name,
        @description,
        @value_type_id,
        @feature_version_id
    )
RETURNING id;
-- name: UpdateKey :exec
UPDATE keys
SET description = @description
WHERE id = @key_id;
-- name: GetKeyIDByName :one
SELECT id
FROM keys
WHERE name = @name
    AND feature_version_id = @feature_version_id
LIMIT 1;
-- name: CreateValueType :one
INSERT INTO value_types (name, editor)
VALUES (@name, @editor)
RETURNING id;
-- name: EndKeyValidity :exec
UPDATE keys
SET valid_to = @valid_to
WHERE id = @key_id;
-- name: StartKeyValidity :exec
UPDATE keys
SET valid_from = @valid_from
WHERE id = @key_id;
-- name: DeleteKey :exec
DELETE FROM keys
WHERE id = @key_id;