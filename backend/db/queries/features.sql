-- name: GetFeatureVersion :one
WITH last_feature_versions AS (
    SELECT fv.feature_id,
        MAX(fv.version)::int as last_version
    FROM feature_versions fv
    WHERE is_feature_version_valid_in_changeset(fv, @changeset_id)
    GROUP BY fv.feature_id
),
links AS (
    SELECT fvsv.feature_version_id,
        BOOL_OR(sv.published) as published,
        COUNT(*) as link_count
    FROM feature_version_service_versions fvsv
        JOIN service_versions sv ON sv.id = fvsv.service_version_id
    WHERE is_link_valid_in_changeset(fvsv, @changeset_id)
    GROUP BY fvsv.feature_version_id
)
SELECT fv.*,
    f.name as feature_name,
    f.description as feature_description,
    f.service_id,
    lfv.last_version as last_version,
    l.published as linked_to_published_service_version,
    l.link_count as service_version_link_count
FROM feature_versions fv
    JOIN features f ON f.id = fv.feature_id
    JOIN last_feature_versions lfv ON lfv.feature_id = fv.feature_id
    JOIN links l ON l.feature_version_id = fv.id
WHERE fv.id = @feature_version_id
    AND is_feature_version_valid_in_changeset(fv, @changeset_id)
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
-- name: GetVersionsOfFeatureForServiceVersion :many
WITH latest_links AS (
    SELECT fvsv.feature_version_id,
        MAX(fvsv.service_version_id)::bigint as service_version_id
    FROM feature_version_service_versions fvsv
    WHERE is_link_valid_in_changeset(fvsv, @changeset_id)
    GROUP BY fvsv.feature_version_id
)
SELECT fv.id,
    fv.version,
    ll.service_version_id as service_version_id
FROM feature_versions fv
    JOIN latest_links ll ON ll.feature_version_id = fv.id
WHERE fv.feature_id = @feature_id
    AND is_feature_version_valid_in_changeset(fv, @changeset_id)
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
        FROM feature_version_service_versions ifvsv
            JOIN feature_versions ifv ON ifv.id = ifvsv.feature_version_id
        WHERE ifv.feature_id = f.id
            AND ifvsv.service_version_id = @service_version_id
            AND is_link_valid_in_changeset(ifvsv, @changeset_id)
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
-- name: GetFeatureVersionValuesData :many
SELECT vv.data,
    vv.variation_context_id,
    k.id as key_id,
    k.name as key_name,
    k.value_type_id as key_value_type_id,
    k.description as key_description
FROM variation_values vv
    JOIN keys k ON k.id = vv.key_id
WHERE k.feature_version_id = @feature_version_id
    AND is_key_valid_in_changeset(k, @changeset_id)
    AND is_variation_value_valid_in_changeset(vv, @changeset_id);
-- name: GetFeatureVersionValidatorData :many
SELECT k.id as key_id,
    v.validator_type,
    v.parameter,
    v.error_text
FROM value_validators v
    JOIN keys k ON k.id = v.key_id
WHERE k.feature_version_id = @feature_version_id
    AND is_key_valid_in_changeset(k, @changeset_id);