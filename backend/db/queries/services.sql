-- name: GetActiveServiceVersions :many
SELECT sv.*,
    s.name as service_name,
    s.description as service_description,
    s.service_type_id as service_type_id,
    st.name as service_type_name
FROM service_versions sv
    JOIN services s ON s.id = sv.service_id
    JOIN service_types st ON st.id = s.service_type_id
WHERE (
        sv.valid_from IS NOT NULL
        AND sv.valid_to IS NULL
        AND NOT EXISTS (
            SELECT csc.id
            FROM changeset_changes csc
            WHERE csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.previous_service_version_id = sv.id
            LIMIT 1
        )
    )
    OR (
        sv.valid_from IS NULL
        AND EXISTS (
            SELECT csc.id
            FROM changeset_changes csc
            WHERE csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.service_version_id = sv.id
            LIMIT 1
        )
    )
ORDER BY s.name;
-- name: GetServiceVersionsForService :many
SELECT sv.id,
    sv.version
FROM service_versions sv
    JOIN services s ON s.id = sv.service_id
WHERE sv.service_id = @service_id
    AND (
        sv.valid_from IS NOT NULL
        OR (
            sv.valid_from IS NULL
            AND EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.type = 'create'
                    AND csc.service_version_id = sv.id
                LIMIT 1
            )
        )
    )
ORDER BY sv.version;
-- name: GetServiceVersion :one
SELECT sv.*,
    s.name as service_name,
    s.description as service_description,
    s.service_type_id as service_type_id,
    st.name as service_type_name
FROM service_versions sv
    JOIN services s ON s.id = sv.service_id
    JOIN service_types st ON st.id = s.service_type_id
WHERE sv.id = @service_version_id
LIMIT 1;
-- name: GetServiceTypes :many
SELECT *
FROM service_types
ORDER BY name;
-- name: GetServiceType :one
SELECT *
FROM service_types
WHERE id = @service_type_id
LIMIT 1;
-- name: GetServiceIDByName :one
SELECT id
FROM services
WHERE name = @name
LIMIT 1;
-- name: CreateService :one
INSERT INTO services (name, description, service_type_id)
VALUES (@name, @description, @service_type_id)
RETURNING id;
-- name: UpdateService :exec
UPDATE services
SET description = @description
WHERE id = @service_id;
-- name: CreateServiceVersion :one
INSERT INTO service_versions (service_id, version)
VALUES (@service_id, @version)
RETURNING id;
-- name: PublishServiceVersion :exec
UPDATE service_versions
SET published = TRUE
WHERE id = @service_version_id;
-- name: CreateServiceType :one
INSERT INTO service_types (name)
VALUES (@name)
RETURNING id;
-- name: EndServiceVersionValidity :exec
UPDATE service_versions
SET valid_to = @valid_to
WHERE id = @service_version_id;
-- name: StartServiceVersionValidity :exec
UPDATE service_versions
SET valid_from = @valid_from
WHERE id = @service_version_id;
-- name: DeleteServiceVersion :exec
DELETE FROM service_versions
WHERE id = @service_version_id;
-- name: DeleteService :exec
DELETE FROM services
WHERE id = @service_id;