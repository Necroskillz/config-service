-- name: DeleteVariationValue :exec
DELETE FROM variation_values
WHERE id = @variation_value_id;
-- name: CreateVariationValue :one
INSERT INTO variation_values (key_id, variation_context_id, data)
VALUES (@key_id, @variation_context_id, @data)
RETURNING id;
-- name: GetActiveVariationValuesForKey :many
SELECT vv.*
FROM variation_values vv
WHERE vv.key_id = @key_id
    AND (
        vv.valid_from IS NOT NULL
        AND vv.valid_to IS NULL
        AND NOT EXISTS (
            SELECT csc.id
            FROM changeset_changes csc
            WHERE csc.changeset_id = @changeset_id
                AND csc.old_variation_value_id = vv.id
            LIMIT 1
        )
    )
    OR (
        vv.valid_from IS NULL
        AND EXISTS (
            SELECT csc.id
            FROM changeset_changes csc
            WHERE csc.changeset_id = @changeset_id
                AND csc.new_variation_value_id = vv.id
            LIMIT 1
        )
    );
-- name: GetVariationValueIDByVariationContextID :one
SELECT id
FROM variation_values
WHERE variation_context_id = @variation_context_id
LIMIT 1;