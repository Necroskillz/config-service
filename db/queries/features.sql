-- name: GetFeatureVersion :one
SELECT fv.*,
    f.name as feature_name,
    f.description as feature_description
FROM feature_versions fv
    JOIN features f ON f.id = fv.feature_id
WHERE fv.id = @feature_version_id
LIMIT 1;
-- name: GetFeatureIDByName :one
SELECT id
FROM features
WHERE name = @name;
-- name: GetActiveFeatureVersionsForServiceVersion :many
SELECT fv.id,
    fv.feature_id,
    fv.version,
    f.name as feature_name,
    f.description as feature_description
FROM feature_version_service_versions fvsv
    JOIN feature_versions fv ON fv.id = fvsv.feature_version_id
    JOIN features f ON f.id = fv.feature_id
WHERE fvsv.service_version_id = @service_version_id
    AND (
        fvsv.valid_from IS NOT NULL
        AND fvsv.valid_to IS NULL
        AND NOT EXISTS (
            SELECT csc.id
            FROM changeset_changes csc
            WHERE csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.feature_version_service_version_id = fvsv.id
            LIMIT 1
        )
    )
    OR (
        fvsv.valid_from IS NULL
        AND EXISTS (
            SELECT csc.id
            FROM changeset_changes csc
            WHERE csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.feature_version_service_version_id = fvsv.id
            LIMIT 1
        )
    )
ORDER BY f.name;
-- name: GetFeatureVersionsLinkedToServiceVersion :many
SELECT fv.id,
    fv.version
FROM feature_version_service_versions fvsv
    JOIN feature_versions fv ON fvsv.feature_version_id = fv.id
    JOIN features f ON f.id = fv.feature_id
WHERE fv.feature_id = @feature_id
    AND fvsv.service_version_id = @service_version_id
    AND (
        fv.valid_from IS NOT NULL
        OR (
            fv.valid_from IS NULL
            AND EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.type = 'create'
                    AND csc.feature_version_id = fv.id
                LIMIT 1
            )
        )
    )
ORDER BY fv.version;
-- name: CreateFeature :one
INSERT INTO features (name, description, service_id)
VALUES (@name, @description, @service_id)
RETURNING id;
-- name: CreateFeatureVersion :one
INSERT INTO feature_versions (feature_id, version, valid_from)
VALUES (@feature_id, @version, @valid_from)
RETURNING id;
-- name: CreateFeatureVersionServiceVersion :one
INSERT INTO feature_version_service_versions (service_version_id, feature_version_id)
VALUES (@service_version_id, @feature_version_id)
RETURNING id;