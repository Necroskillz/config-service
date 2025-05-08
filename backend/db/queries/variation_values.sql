-- name: DeleteVariationValue :exec
DELETE FROM variation_values
WHERE id = @variation_value_id;
-- name: CreateVariationValue :one
INSERT INTO variation_values (key_id, variation_context_id, data)
VALUES (@key_id, @variation_context_id, @data)
RETURNING id;
-- name: UpdateVariationValue :exec
UPDATE variation_values
SET data = @data,
    variation_context_id = @variation_context_id
WHERE id = @variation_value_id;
-- name: GetVariationValue :one
SELECT vv.*
FROM variation_values vv
WHERE vv.id = @variation_value_id
    AND is_variation_value_valid_in_changeset(vv, @changeset_id)
LIMIT 1;
-- name: GetVariationValuesForKey :many
SELECT vv.*
FROM variation_values vv
WHERE vv.key_id = @key_id
    AND is_variation_value_valid_in_changeset(vv, @changeset_id);
-- name: GetVariationValueIDByVariationContextID :one
SELECT id
FROM variation_values vv
WHERE vv.key_id = @key_id
    AND vv.variation_context_id = @variation_context_id
    AND is_variation_value_valid_in_changeset(vv, @changeset_id)
LIMIT 1;
-- name: EndValueValidity :exec
UPDATE variation_values
SET valid_to = @valid_to
WHERE id = @variation_value_id;
-- name: StartValueValidity :exec
UPDATE variation_values
SET valid_from = @valid_from
WHERE id = @variation_value_id;