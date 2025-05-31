-- name: DeleteVariationValue :exec
DELETE FROM variation_values
WHERE id = @variation_value_id;

-- name: CreateVariationValue :one
INSERT INTO variation_values(key_id, variation_context_id, data)
    VALUES (@key_id, @variation_context_id, @data)
RETURNING
    id;

-- name: CreateVariationValues :copyfrom
INSERT INTO variation_values(key_id, variation_context_id, data)
    VALUES ($1, $2, $3);

-- name: UpdateVariationValue :exec
UPDATE
    variation_values
SET
    data = @data,
    variation_context_id = @variation_context_id
WHERE
    id = @variation_value_id;

-- name: GetVariationValue :one
SELECT
    vv.*
FROM
    variation_values vv
    JOIN valid_variation_values_in_changeset(@changeset_id) vvv ON vvv.id = vv.id
WHERE
    vv.id = @variation_value_id
LIMIT 1;

-- name: GetVariationValuesForKey :many
SELECT
    vv.*
FROM
    variation_values vv
    JOIN valid_variation_values_in_changeset(@changeset_id) vvv ON vvv.id = vv.id
WHERE
    vv.key_id = @key_id;

-- name: GetVariationValuesForWipFeatureVersion :many
SELECT
    vv.*
FROM
    variation_values vv
    JOIN keys k ON k.id = vv.key_id
WHERE
    k.feature_version_id = @feature_version_id;

-- name: GetVariationValueIDByVariationContextID :one
SELECT
    vv.id
FROM
    variation_values vv
    JOIN valid_variation_values_in_changeset(@changeset_id) vvv ON vvv.id = vv.id
WHERE
    vv.key_id = @key_id
    AND vv.variation_context_id = @variation_context_id
LIMIT 1;

-- name: EndValueValidity :exec
UPDATE
    variation_values
SET
    valid_to = @valid_to
WHERE
    id = @variation_value_id;

-- name: StartValueValidity :exec
UPDATE
    variation_values
SET
    valid_from = @valid_from
WHERE
    id = @variation_value_id;

