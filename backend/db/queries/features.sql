-- name: GetFeatureVersion :one
SELECT fv.*,
    f.name as feature_name,
    f.description as feature_description,
    f.service_id
FROM feature_versions fv
    JOIN features f ON f.id = fv.feature_id
WHERE fv.id = @feature_version_id
LIMIT 1;
-- name: GetFeatureIDByName :one
SELECT id
FROM features
WHERE name = @name;
-- name: GetFeatureVersionsForServiceVersion :many
SELECT fv.id,
    fv.feature_id,
    fv.version,
    f.name as feature_name,
    f.description as feature_description,
    csc.changeset_id
FROM feature_version_service_versions fvsv
    JOIN feature_versions fv ON fv.id = fvsv.feature_version_id
    JOIN features f ON f.id = fv.feature_id
    JOIN changeset_changes csc ON csc.feature_version_service_version_id = fvsv.id AND csc.type = 'create' AND csc.kind = 'feature_version_service_version'
WHERE fvsv.service_version_id = @service_version_id
    AND (
        (
            fvsv.valid_from IS NOT NULL
            AND fvsv.valid_to IS NULL
            AND NOT EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'feature_version_service_version'
                    AND csc.type = 'delete'
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
                    AND csc.kind = 'feature_version_service_version'
                    AND csc.type = 'create'
                    AND csc.feature_version_service_version_id = fvsv.id
                LIMIT 1
            )
        )
    )
ORDER BY f.name;
-- name: GetFeatureVersionsLinkedToServiceVersionForFeature :many
SELECT fv.id,
    fv.version
FROM feature_version_service_versions fvsv
    JOIN feature_versions fv ON fvsv.feature_version_id = fv.id
    JOIN features f ON f.id = fv.feature_id
WHERE fv.feature_id = @feature_id
    AND fvsv.service_version_id = @service_version_id
    AND (
        (
            fvsv.valid_from IS NOT NULL
            AND NOT EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'feature_version_service_version'
                    AND csc.type = 'delete'
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
                    AND csc.kind = 'feature_version_service_version'
                    AND csc.type = 'create'
                    AND csc.feature_version_service_version_id = fvsv.id
                LIMIT 1
            )
        )
    )
ORDER BY fv.version;
-- name: GetFeatureVersionsLinkableToServiceVersion :many
SELECT fv.id,
    fv.version,
    f.name as feature_name,
    f.description as feature_description
FROM feature_versions fv
    JOIN features f ON f.id = fv.feature_id
WHERE f.service_id = @service_id
    AND (
        (
            fv.valid_from IS NOT NULL
            AND NOT EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'feature_version'
                    AND csc.type = 'delete'
                    AND csc.feature_version_id = fv.id
                LIMIT 1
            )
        )
        OR (
            fv.valid_from IS NULL
            AND EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'feature_version'
                    AND csc.type = 'create'
                    AND csc.feature_version_id = fv.id
                LIMIT 1
            )
        )
    )
    AND NOT EXISTS (
        SELECT 1
        FROM feature_version_service_versions fvsv
        WHERE fvsv.feature_version_id = fv.id
            AND fvsv.service_version_id = @service_version_id
            AND (
                (
                    fvsv.valid_from IS NOT NULL
                    AND NOT EXISTS (
                        SELECT csc.id
                        FROM changeset_changes csc
                        WHERE csc.changeset_id = @changeset_id
                            AND csc.kind = 'feature_version_service_version'
                            AND csc.type = 'delete'
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
                            AND csc.kind = 'feature_version_service_version'
                            AND csc.type = 'create'
                            AND csc.feature_version_service_version_id = fvsv.id
                        LIMIT 1
                    )
                )
            )
    )
ORDER BY f.name,
    fv.version;
-- name: GetFeatureVersionServiceVersionLink :one
SELECT fvsv.id, csc.changeset_id
FROM feature_version_service_versions fvsv
    JOIN changeset_changes csc ON csc.feature_version_service_version_id = fvsv.id AND csc.type = 'create' AND csc.kind = 'feature_version_service_version'
WHERE fvsv.feature_version_id = @feature_version_id
    AND fvsv.service_version_id = @service_version_id
    AND (
        (
            fvsv.valid_from IS NOT NULL
            AND NOT EXISTS (
                SELECT csc.id
                FROM changeset_changes csc
                WHERE csc.changeset_id = @changeset_id
                    AND csc.kind = 'feature_version_service_version'
                    AND csc.type = 'delete'
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
                    AND csc.kind = 'feature_version_service_version'
                    AND csc.type = 'create'
                    AND csc.feature_version_service_version_id = fvsv.id
                LIMIT 1
            )
        )
    );
-- name: IsFeatureLinkedToServiceVersion :one
SELECT EXISTS (
        SELECT 1
        FROM feature_version_service_versions fvsv
            JOIN feature_versions fv ON fv.id = fvsv.feature_version_id
        WHERE fv.feature_id = @feature_id
            AND fvsv.service_version_id = @service_version_id
            AND (
                (
                    fvsv.valid_from IS NOT NULL
                    AND NOT EXISTS (
                        SELECT csc.id
                        FROM changeset_changes csc
                        WHERE csc.changeset_id = @changeset_id
                            AND csc.kind = 'feature_version_service_version'
                            AND csc.type = 'delete'
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
                            AND csc.kind = 'feature_version_service_version'
                            AND csc.type = 'create'
                            AND csc.feature_version_service_version_id = fvsv.id
                        LIMIT 1
                    )
                )
            )
    );
-- name: CreateFeature :one
INSERT INTO features (name, description, service_id)
VALUES (@name, @description, @service_id)
RETURNING id;
-- name: UpdateFeature :exec
UPDATE features
SET description = @description
WHERE id = @feature_id;
-- name: CreateFeatureVersion :one
INSERT INTO feature_versions (feature_id, version, valid_from)
VALUES (@feature_id, @version, @valid_from)
RETURNING id;
-- name: CreateFeatureVersionServiceVersion :one
INSERT INTO feature_version_service_versions (service_version_id, feature_version_id)
VALUES (@service_version_id, @feature_version_id)
RETURNING id;
-- name: EndFeatureVersionValidity :exec
UPDATE feature_versions
SET valid_to = @valid_to
WHERE id = @feature_version_id;
-- name: StartFeatureVersionValidity :exec
UPDATE feature_versions
SET valid_from = @valid_from
WHERE id = @feature_version_id;
-- name: EndFeatureVersionServiceVersionValidity :exec
UPDATE feature_version_service_versions
SET valid_to = @valid_to
WHERE id = @feature_version_service_version_id;
-- name: StartFeatureVersionServiceVersionValidity :exec
UPDATE feature_version_service_versions
SET valid_from = @valid_from
WHERE id = @feature_version_service_version_id;
-- name: DeleteFeatureVersionServiceVersion :exec
DELETE FROM feature_version_service_versions
WHERE id = @feature_version_service_version_id;
-- name: DeleteFeatureVersion :exec
DELETE FROM feature_versions
WHERE id = @feature_version_id;
-- name: DeleteFeature :exec
DELETE FROM features
WHERE id = @feature_id;