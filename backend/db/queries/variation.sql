-- name: GetVariationContextValues :many
SELECT
    vpv.value,
    vpv.id AS value_id,
    vpv.variation_property_id AS property_id
FROM
    variation_contexts vc
    JOIN variation_context_variation_property_values vcvpv ON vcvpv.variation_context_id = vc.id
    JOIN variation_property_values vpv ON vpv.id = vcvpv.variation_property_value_id
WHERE
    vc.id = @variation_context_id;

-- name: GetVariationContextId :one
SELECT
    vc.id
FROM
    variation_contexts vc
WHERE (
    SELECT
        COUNT(variation_property_value_id)
    FROM
        variation_context_variation_property_values
    WHERE
        variation_context_id = vc.id) = @property_count::int
    AND (
        SELECT
            COUNT(vcvpv.variation_property_value_id)
        FROM
            variation_context_variation_property_values vcvpv
        WHERE
            vcvpv.variation_context_id = vc.id
            AND vcvpv.variation_property_value_id = ANY (@variation_property_value_ids::bigint[])) = @property_count::int;

-- name: CreateVariationContext :one
INSERT INTO variation_contexts DEFAULT
    VALUES
    RETURNING
        id;

-- name: CreateVariationContextValue :exec
INSERT INTO variation_context_variation_property_values(variation_context_id, variation_property_value_id)
    VALUES (@variation_context_id, @variation_property_value_id);

-- name: GetVariationPropertyValues :many
SELECT
    vpv.id,
    vpv.value,
    COALESCE(vpv.parent_id, 0) AS parent_id,
    vpv.archived,
    vp.name AS property_name,
    vp.display_name AS property_display_name,
    vp.id AS property_id
FROM
    variation_properties vp
    LEFT JOIN variation_property_values vpv ON vpv.variation_property_id = vp.id
ORDER BY
    vp.id,
    parent_id,
    vpv.order_index;

-- name: GetServiceTypeVariationProperties :many
SELECT
    stvp.service_type_id,
    stvp.variation_property_id
FROM
    service_type_variation_properties stvp
ORDER BY
    stvp.service_type_id,
    stvp.priority;

-- name: AddPropertyToServiceType :exec
INSERT INTO service_type_variation_properties(service_type_id, variation_property_id, priority)
    VALUES (@service_type_id, @variation_property_id, @priority);

-- name: CreateVariationProperty :one
INSERT INTO variation_properties(name, display_name)
    VALUES (@name, @display_name)
RETURNING
    id;

-- name: UpdateVariationProperty :exec
UPDATE
    variation_properties
SET
    display_name = @display_name
WHERE
    id = @id;

-- name: GetVariationProperty :one
SELECT
    *
FROM
    variation_properties
WHERE
    id = @id;

-- name: GetVariationPropertyIDByName :one
SELECT
    id
FROM
    variation_properties
WHERE
    name = @name;

-- name: GetVariationPropertyValueIDByName :one
SELECT
    id
FROM
    variation_property_values
WHERE
    variation_property_id = @variation_property_id
    AND value = @value;

-- name: CreateVariationPropertyValue :one
INSERT INTO variation_property_values(variation_property_id, value, parent_id, order_index)
SELECT
    @variation_property_id,
    @value,
    @parent_id,
(
        SELECT
            COALESCE(MAX(order_index), 0) + 1
        FROM
            variation_property_values
        WHERE
            variation_property_id = @variation_property_id
            AND parent_id IS NOT DISTINCT FROM @parent_id)
RETURNING
    id;

-- name: UpdateVariationPropertyValueOrder :exec
WITH source AS (
    SELECT
        svpv.id,
        svpv.order_index,
        svpv.parent_id,
        svpv.variation_property_id
    FROM
        variation_property_values svpv
    WHERE
        svpv.id = @id
),
bounds AS (
    SELECT
        MIN(bvpv.order_index)::int AS min_index,
        MAX(bvpv.order_index)::int AS max_index
    FROM
        source s
        JOIN variation_property_values bvpv ON bvpv.variation_property_id = s.variation_property_id
            AND bvpv.parent_id IS NOT DISTINCT FROM s.parent_id
),
params AS (
    SELECT
        source.id,
        source.variation_property_id,
        source.parent_id,
        source.order_index AS source_index,
        GREATEST(bounds.min_index, LEAST(@target_index::int, bounds.max_index)) AS target_index
    FROM
        source,
        bounds)
UPDATE
    variation_property_values vpv
SET
    order_index = CASE WHEN vpv.order_index = params.source_index THEN
        params.target_index
    WHEN params.source_index < params.target_index THEN
        vpv.order_index - 1
    ELSE
        vpv.order_index + 1
    END
FROM
    params
WHERE
    vpv.variation_property_id = params.variation_property_id
    AND vpv.parent_id IS NOT DISTINCT FROM params.parent_id
    AND vpv.order_index BETWEEN LEAST(params.source_index, params.target_index)
    AND GREATEST(params.source_index, params.target_index);

-- name: GetVariationPropertyValueIDByValue :one
SELECT
    id
FROM
    variation_property_values
WHERE
    variation_property_id = @variation_property_id
    AND value = @value
    AND NOT archived;

-- name: GetVariationPropertyValuesUsage :many
SELECT
    vpv.id,
    COUNT(vv.id)::int AS usage_count
FROM
    variation_property_values vpv
    JOIN variation_context_variation_property_values vcvpv ON vcvpv.variation_property_value_id = vpv.id
    JOIN variation_values vv ON vv.variation_context_id = vcvpv.variation_context_id
WHERE
    vpv.variation_property_id = @variation_property_id
GROUP BY
    vpv.id;

-- name: GetVariationPropertyUsage :one
SELECT
    COUNT(vv.id)::int AS usage_count
FROM
    variation_property_values vpv
    JOIN variation_context_variation_property_values vcvpv ON vcvpv.variation_property_value_id = vpv.id
    JOIN variation_values vv ON vv.variation_context_id = vcvpv.variation_context_id
WHERE
    vpv.variation_property_id = @variation_property_id;

-- name: DeleteVariationProperty :exec
DELETE FROM variation_properties
WHERE id = @id;

-- name: DeleteVariationPropertyValue :exec
DELETE FROM variation_property_values
WHERE id = @id;

-- name: SetVariationPropertyValueArchived :exec
UPDATE
    variation_property_values
SET
    archived = @archived
WHERE
    id = @id;

