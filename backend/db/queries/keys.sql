-- name: GetKey :one
SELECT
    k.*,
    vt.kind AS value_type_kind,
    vt.name AS value_type_name,
    csc.changeset_id AS created_in_changeset_id
FROM
    keys k
    JOIN value_types vt ON vt.id = k.value_type_id
    JOIN changeset_changes csc ON csc.key_id = k.id
        AND csc.type = 'create'
        AND csc.kind = 'key'
    JOIN valid_keys_in_changeset(@changeset_id) vk ON vk.id = k.id
WHERE
    k.id = @key_id
LIMIT 1;

-- name: GetKeysForFeatureVersion :many
SELECT
    k.*,
    vt.kind AS value_type_kind,
    vt.name AS value_type_name
FROM
    keys k
    JOIN value_types vt ON vt.id = k.value_type_id
    JOIN valid_keys_in_changeset(@changeset_id) vk ON vk.id = k.id
WHERE
    k.feature_version_id = @feature_version_id
ORDER BY
    k.name;

-- name: GetAppliedKeys :many
SELECT DISTINCT
    k.name
FROM
    keys k
WHERE
    k.valid_from IS NOT NULL
    AND (sqlc.narg('feature_version_id')::bigint IS NULL
        OR k.feature_version_id = sqlc.arg('feature_version_id')::bigint)
    AND (sqlc.narg('feature_id')::bigint IS NULL
        OR EXISTS (
            SELECT
                1
            FROM
                feature_versions fv
            WHERE
                fv.feature_id = sqlc.arg('feature_id')::bigint
                AND fv.id = k.feature_version_id));

-- name: GetKeysForWipFeatureVersion :many
SELECT
    *
FROM
    keys
WHERE
    feature_version_id = @feature_version_id;

-- name: GetValueTypes :many
SELECT
    *
FROM
    value_types;

-- name: GetValueType :one
SELECT
    *
FROM
    value_types
WHERE
    id = @id;

-- name: CreateKey :one
INSERT INTO keys(name, description, value_type_id, feature_version_id)
    VALUES (@name, @description, @value_type_id, @feature_version_id)
RETURNING
    id;

-- name: CreateKeys :copyfrom
INSERT INTO keys(name, description, value_type_id, feature_version_id)
    VALUES ($1, $2, $3, $4);

-- name: UpdateKey :exec
UPDATE
    keys
SET
    description = @description,
    validators_updated_at = @validators_updated_at,
    updated_at = now()
WHERE
    id = @key_id;

-- name: GetKeyIDByName :one
SELECT
    k.id
FROM
    keys k
    JOIN valid_keys_in_changeset(@changeset_id) vk ON vk.id = k.id
WHERE
    k.name = @name
    AND k.feature_version_id = @feature_version_id
LIMIT 1;

-- name: CreateValueType :one
INSERT INTO value_types(name, kind)
    VALUES (@name, @kind)
RETURNING
    id;

-- name: GetValueValidators :many
SELECT
    *
FROM
    value_validators
WHERE
    value_type_id = @value_type_id
    OR key_id = @key_id;

-- name: GetValueTypeValueValidators :many
SELECT
    *
FROM
    value_validators
WHERE
    value_type_id IS NOT NULL;

-- name: CreateValueValidatorForValueType :one
INSERT INTO value_validators(value_type_id, validator_type, parameter, error_text)
    VALUES (@value_type_id, @validator_type, sqlc.narg('parameter'), sqlc.narg('error_text'))
RETURNING
    id;

-- name: CreateValueValidatorForKey :one
INSERT INTO value_validators(key_id, validator_type, parameter, error_text)
    VALUES (@key_id, @validator_type, sqlc.narg('parameter'), sqlc.narg('error_text'))
RETURNING
    id;

-- name: CreateValueValidators :copyfrom
INSERT INTO value_validators(key_id, value_type_id, validator_type, parameter, error_text)
    VALUES ($1, $2, $3, $4, $5);

-- name: DeleteValueValidatorsForKey :exec
DELETE FROM value_validators
WHERE key_id = @key_id::bigint;

-- name: EndKeyValidity :exec
UPDATE
    keys
SET
    valid_to = @valid_to
WHERE
    id = @key_id;

-- name: StartKeyValidity :exec
UPDATE
    keys
SET
    valid_from = @valid_from
WHERE
    id = @key_id;

-- name: DeleteKey :exec
DELETE FROM keys
WHERE id = @key_id;

