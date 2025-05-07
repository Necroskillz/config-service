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
    csc.changeset_id as linked_in_changeset_id
FROM feature_version_service_versions fvsv
    JOIN feature_versions fv ON fv.id = fvsv.feature_version_id
    JOIN features f ON f.id = fv.feature_id
    JOIN changeset_changes csc ON csc.feature_version_service_version_id = fvsv.id
    AND csc.type = 'create'
    AND csc.kind = 'feature_version_service_version'
WHERE fvsv.service_version_id = @service_version_id
    AND is_link_valid_in_changeset(fvsv, @changeset_id)
ORDER BY f.name;
-- name: GetFeatureVersionsLinkedToServiceVersionForFeature :many
SELECT fv.id,
    fv.version
FROM feature_version_service_versions fvsv
    JOIN feature_versions fv ON fvsv.feature_version_id = fv.id
    JOIN features f ON f.id = fv.feature_id
WHERE fv.feature_id = @feature_id
    AND fvsv.service_version_id = @service_version_id
    AND is_link_valid_in_changeset(fvsv, @changeset_id)
ORDER BY fv.version;
-- name: GetFeatureVersionsLinkableToServiceVersion :many
SELECT fv.id,
    fv.version,
    f.name as feature_name,
    f.description as feature_description
FROM feature_versions fv
    JOIN features f ON f.id = fv.feature_id
WHERE f.service_id = @service_id
    AND is_feature_version_valid_in_changeset(fv, @changeset_id)
    AND NOT EXISTS (
        SELECT 1
        FROM feature_version_service_versions fvsv
        WHERE fvsv.feature_version_id = fv.id
            AND fvsv.service_version_id = @service_version_id
            AND is_link_valid_in_changeset(fvsv, @changeset_id)
    )
ORDER BY f.name,
    fv.version;
-- name: GetFeatureVersionServiceVersionLink :one
SELECT fvsv.id,
    csc.changeset_id as created_in_changeset_id
FROM feature_version_service_versions fvsv
    JOIN changeset_changes csc ON csc.feature_version_service_version_id = fvsv.id
    AND csc.type = 'create'
    AND csc.kind = 'feature_version_service_version'
WHERE fvsv.feature_version_id = @feature_version_id
    AND fvsv.service_version_id = @service_version_id
    AND is_link_valid_in_changeset(fvsv, @changeset_id);
-- name: IsFeatureLinkedToServiceVersion :one
SELECT EXISTS (
        SELECT 1
        FROM feature_version_service_versions fvsv
            JOIN feature_versions fv ON fv.id = fvsv.feature_version_id
        WHERE fv.feature_id = @feature_id
            AND fvsv.service_version_id = @service_version_id
            AND is_link_valid_in_changeset(fvsv, @changeset_id)
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