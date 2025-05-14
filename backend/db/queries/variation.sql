-- name: GetVariationContextValues :many
SELECT vpv.value,
    vpv.id as value_id,
    vpv.variation_property_id as property_id
FROM variation_contexts vc
    JOIN variation_context_variation_property_values vcvpv on vcvpv.variation_context_id = vc.id
    JOIN variation_property_values vpv on vpv.id = vcvpv.variation_property_value_id
WHERE vc.id = @variation_context_id;
-- name: GetVariationContextId :one
SELECT vc.id
FROM variation_contexts vc
WHERE (
        SELECT COUNT(variation_property_value_id)
        FROM variation_context_variation_property_values
        WHERE variation_context_id = vc.id
    ) = @property_count::int
    AND (
        SELECT COUNT(vcvpv.variation_property_value_id)
        FROM variation_context_variation_property_values vcvpv
        WHERE vcvpv.variation_context_id = vc.id
            AND vcvpv.variation_property_value_id = ANY(@variation_property_value_ids::bigint [])
    ) = @property_count::int;
-- name: CreateVariationContext :one
INSERT INTO variation_contexts DEFAULT
VALUES
RETURNING id;
-- name: CreateVariationContextValue :exec
INSERT INTO variation_context_variation_property_values (
        variation_context_id,
        variation_property_value_id
    )
VALUES (
        @variation_context_id,
        @variation_property_value_id
    );
-- name: GetVariationPropertyValues :many
SELECT vpv.id,
    vpv.value,
    vpv.parent_id,
    vp.name as property_name,
    vp.display_name as property_display_name,
    vp.id as property_id
FROM variation_properties vp
    LEFT JOIN variation_property_values vpv ON vpv.variation_property_id = vp.id
ORDER BY vpv.id;
-- name: GetServiceTypeVariationProperties :many
SELECT stvp.service_type_id,
    stvp.variation_property_id
FROM service_type_variation_properties stvp
ORDER BY stvp.service_type_id,
    stvp.priority;
-- name: AddPropertyToServiceType :exec
INSERT INTO service_type_variation_properties (service_type_id, variation_property_id, priority)
VALUES (
        @service_type_id,
        @variation_property_id,
        @priority
    );
-- name: CreateVariationProperty :one
INSERT INTO variation_properties (name, display_name)
VALUES (@name, @display_name)
RETURNING id;
-- name: UpdateVariationProperty :exec
UPDATE variation_properties
SET display_name = @display_name
WHERE id = @id;
-- name: GetVariationProperty :one
SELECT *
FROM variation_properties
WHERE id = @id;
-- name: GetVariationPropertyIDByName :one
SELECT id
FROM variation_properties
WHERE name = @name;
-- name: GetVariationPropertyValueIDByName :one
SELECT id
FROM variation_property_values
WHERE variation_property_id = @variation_property_id
    AND value = @value
    AND NOT archived;
-- name: CreateVariationPropertyValue :one
INSERT INTO variation_property_values (variation_property_id, value, parent_id)
VALUES (@variation_property_id, @value, @parent_id)
RETURNING id;
-- name: GetVariationPropertyValueIDByValue :one
SELECT id
FROM variation_property_values
WHERE variation_property_id = @variation_property_id
    AND value = @value
    AND NOT archived;