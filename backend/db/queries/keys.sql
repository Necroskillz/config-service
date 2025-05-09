-- name: GetKey :one
SELECT k.*,
    vt.kind as value_type_kind,
    vt.name as value_type_name,
    csc.changeset_id as created_in_changeset_id
FROM keys k
    JOIN value_types vt ON vt.id = k.value_type_id
    JOIN changeset_changes csc ON csc.key_id = k.id
    AND csc.type = 'create'
    AND csc.kind = 'key'
WHERE k.id = @key_id
    AND is_key_valid_in_changeset(k, @changeset_id)
LIMIT 1;
-- name: GetKeysForFeatureVersion :many
SELECT k.*,
    vt.kind as value_type_kind,
    vt.name as value_type_name
FROM keys k
    JOIN value_types vt ON vt.id = k.value_type_id
WHERE k.feature_version_id = @feature_version_id
    AND is_key_valid_in_changeset(k, @changeset_id)
ORDER BY k.name;
-- name: GetValueTypes :many
SELECT *
FROM value_types;
-- name: GetValueType :one
SELECT *
FROM value_types
WHERE id = @id;
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
SET description = @description, validators_updated_at = @validators_updated_at
WHERE id = @key_id;
-- name: GetKeyIDByName :one
SELECT id
FROM keys
WHERE name = @name
    AND feature_version_id = @feature_version_id
LIMIT 1;
-- name: CreateValueType :one
INSERT INTO value_types (name, kind)
VALUES (@name, @kind)
RETURNING id;
-- name: GetValueValidators :many
SELECT *
FROM value_validators
WHERE value_type_id = @value_type_id
    OR key_id = @key_id;
-- name: GetValueTypeValueValidators :many
SELECT *
FROM value_validators
WHERE value_type_id IS NOT NULL;
-- name: CreateValueValidatorForValueType :one
INSERT INTO value_validators (
        value_type_id,
        validator_type,
        parameter,
        error_text
    )
VALUES (
        @value_type_id,
        @validator_type,
        sqlc.narg('parameter'),
        sqlc.narg('error_text')
    )
RETURNING id;
-- name: CreateValueValidatorForKey :one
INSERT INTO value_validators (key_id, validator_type, parameter, error_text)
VALUES (
        @key_id,
        @validator_type,
        sqlc.narg('parameter'),
        sqlc.narg('error_text')
    )
RETURNING id;
-- name: DeleteValueValidator :exec
DELETE FROM value_validators
WHERE id = @value_validator_id;
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