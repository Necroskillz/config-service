-- name: GetServiceVersions :many
SELECT
    sv.*,
    s.name AS service_name,
    s.description AS service_description,
    s.service_type_id AS service_type_id,
    st.name AS service_type_name
FROM
    service_versions sv
    JOIN services s ON s.id = sv.service_id
    JOIN service_types st ON st.id = s.service_type_id
    JOIN valid_service_versions_in_changeset(@changeset_id) vsv ON vsv.id = sv.id
ORDER BY
    s.name,
    sv.version ASC;

-- name: GetServiceAdmins :many
SELECT
    s.id AS service_id,
    s.name AS service_name,
    u.id AS user_id,
    u.name AS user_name
FROM
    services s
    JOIN user_permissions up ON up.service_id = s.id
        AND up.kind = 'service'
        AND up.permission = 'admin'
    JOIN users u ON u.id = up.user_id
WHERE (sqlc.narg('service_id')::bigint IS NULL
    OR s.id = sqlc.narg('service_id')::bigint)
AND (sqlc.narg('user_id')::bigint IS NULL
    OR u.id = sqlc.narg('user_id')::bigint);

-- name: GetServiceVersionsForService :many
SELECT
    sv.id,
    sv.version
FROM
    service_versions sv
    JOIN valid_service_versions_in_changeset(@changeset_id) vsv ON vsv.id = sv.id
WHERE
    sv.service_id = @service_id
ORDER BY
    sv.version ASC;

-- name: GetServiceVersion :one
WITH last_service_versions AS (
    SELECT
        sv.service_id,
        MAX(sv.version)::int AS last_version
    FROM
        service_versions sv
        JOIN valid_service_versions_in_changeset(@changeset_id) vsv ON vsv.id = sv.id
    GROUP BY
        sv.service_id
)
SELECT
    sv.*,
    s.name AS service_name,
    s.description AS service_description,
    s.service_type_id AS service_type_id,
    st.name AS service_type_name,
    lsv.last_version AS last_version,
    csc.changeset_id AS changeset_id
FROM
    service_versions sv
    JOIN services s ON s.id = sv.service_id
    JOIN service_types st ON st.id = s.service_type_id
    JOIN last_service_versions lsv ON lsv.service_id = sv.service_id
    JOIN changeset_changes csc ON csc.service_version_id = sv.id
        AND csc.type = 'create'
        AND csc.kind = 'service_version'
    JOIN valid_service_versions_in_changeset(@changeset_id) vsv ON vsv.id = sv.id
WHERE
    sv.id = @service_version_id
LIMIT 1;

-- name: GetServiceIDByName :one
SELECT
    id
FROM
    services
WHERE
    name = @name
LIMIT 1;

-- name: CreateService :one
INSERT INTO services(name, description, service_type_id)
    VALUES (@name, @description, @service_type_id)
RETURNING
    id;

-- name: UpdateService :exec
UPDATE
    services
SET
    description = @description,
    updated_at = now()
WHERE
    id = @service_id;

-- name: CreateServiceVersion :one
INSERT INTO service_versions(service_id, version)
    VALUES (@service_id, @version)
RETURNING
    id;

-- name: PublishServiceVersion :exec
UPDATE
    service_versions
SET
    published = TRUE,
    updated_at = now()
WHERE
    id = @service_version_id;

-- name: EndServiceVersionValidity :exec
UPDATE
    service_versions
SET
    valid_to = @valid_to
WHERE
    id = @service_version_id;

-- name: StartServiceVersionValidity :exec
UPDATE
    service_versions
SET
    valid_from = @valid_from
WHERE
    id = @service_version_id;

-- name: DeleteServiceVersion :exec
DELETE FROM service_versions
WHERE id = @service_version_id;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = @service_id;

-- name: GetServiceVersionByNameAndVersion :one
SELECT
    sv.*
FROM
    service_versions sv
    JOIN services s ON s.id = sv.service_id
WHERE
    s.name = @name
    AND sv.version = @version
    AND sv.valid_from IS NOT NULL
    AND sv.valid_to IS NULL
LIMIT 1;

