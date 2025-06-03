-- name: GetServiceTypes :many
SELECT
    *
FROM
    service_types
ORDER BY
    name;

-- name: GetServiceType :one
SELECT
    st.*,
(
        SELECT
            COUNT(*)::int
        FROM
            services s
        WHERE
            s.service_type_id = st.id) AS usage_count
FROM
    service_types st
WHERE
    st.id = @service_type_id
LIMIT 1;

-- name: GetServiceTypeIDByName :one
SELECT
    id
FROM
    service_types
WHERE
    name = @name
LIMIT 1;

-- name: CreateServiceType :one
INSERT INTO service_types(name, created_at)
    VALUES (@name, now())
RETURNING
    id;

-- name: DeleteServiceType :exec
DELETE FROM service_types
WHERE id = @id;

-- name: GetServiceTypeVariationPropertyLinks :many
WITH usage AS (
    SELECT
        vpv.variation_property_id,
        COUNT(vv.id)::int AS usage_count
    FROM
        variation_property_values vpv
        JOIN variation_context_variation_property_values vcvpv ON vcvpv.variation_property_value_id = vpv.id
        JOIN variation_values vv ON vv.variation_context_id = vcvpv.variation_context_id
        JOIN keys k ON k.id = vv.key_id
        JOIN feature_versions fv ON fv.id = k.feature_version_id
        JOIN feature_version_service_versions fvsv ON fvsv.feature_version_id = fv.id
        JOIN service_versions sv ON sv.id = fvsv.service_version_id
        JOIN services s ON s.id = sv.service_id
    WHERE
        s.service_type_id = @service_type_id
    GROUP BY
        vpv.variation_property_id
)
SELECT
    stvp.id,
    stvp.priority,
    vp.name,
    vp.display_name,
    vp.id AS property_id,
    COALESCE(u.usage_count, 0) AS usage_count
FROM
    service_type_variation_properties stvp
    JOIN variation_properties vp ON stvp.variation_property_id = vp.id
    LEFT JOIN usage u ON stvp.variation_property_id = u.variation_property_id
WHERE
    stvp.service_type_id = @service_type_id
ORDER BY
    stvp.priority;

-- name: IsVariationPropertyLinkedToServiceType :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            service_type_variation_properties
        WHERE
            service_type_id = @service_type_id
            AND variation_property_id = @variation_property_id);

-- name: CreateServiceTypeVariationPropertyLink :one
INSERT INTO service_type_variation_properties(service_type_id, variation_property_id, priority)
    VALUES (@service_type_id, @variation_property_id,(
            SELECT
                COALESCE(MAX(priority), 0) + 1
            FROM
                service_type_variation_properties
            WHERE
                service_type_id = @service_type_id))
RETURNING
    id;

-- name: DeleteServiceTypeVariationPropertyLink :exec
DELETE FROM service_type_variation_properties
WHERE service_type_id = @service_type_id
    AND variation_property_id = @variation_property_id;

-- name: UpdateServiceTypeVariationPropertyPriority :exec
WITH source AS (
    SELECT
        sstvp.id,
        sstvp.priority,
        sstvp.service_type_id
    FROM
        service_type_variation_properties sstvp
    WHERE
        sstvp.id = @id
),
bounds AS (
    SELECT
        MIN(bstvp.priority)::int AS min_priority,
        MAX(bstvp.priority)::int AS max_priority
    FROM
        source s
        JOIN service_type_variation_properties bstvp ON bstvp.service_type_id = s.service_type_id
),
params AS (
    SELECT
        source.id,
        source.service_type_id,
        source.priority AS source_priority,
        GREATEST(bounds.min_priority, LEAST(@target_priority::int, bounds.max_priority)) AS target_priority
    FROM
        source,
        bounds)
UPDATE
    service_type_variation_properties stvp
SET
    priority = CASE WHEN stvp.priority = params.source_priority THEN
        params.target_priority
    WHEN params.source_priority < params.target_priority THEN
        stvp.priority - 1
    ELSE
        stvp.priority + 1
    END,
    updated_at = now()
FROM
    params
WHERE
    stvp.service_type_id = params.service_type_id
    AND stvp.priority BETWEEN LEAST(params.source_priority, params.target_priority)
    AND GREATEST(params.source_priority, params.target_priority);

