-- name: CreateChangeset :one
INSERT INTO changesets (user_id, state)
VALUES (@user_id, 'open')
RETURNING id;
-- name: GetOpenChangesetIDForUser :one
SELECT id
FROM changesets
WHERE user_id = @user_id
    AND state = 'open'
LIMIT 1;
-- name: GetChangeset :one
SELECT cs.id,
    cs.state,
    u.id as user_id,
    u.name as user_name
FROM changesets cs
    JOIN users u ON u.id = cs.user_id
WHERE cs.id = @changeset_id
LIMIT 1;
-- name: SetChangesetState :exec
UPDATE changesets
SET state = @state
WHERE id = @changeset_id;
-- name: GetChangesetChanges :many
SELECT csc.id,
    csc.type,
    sv.id as service_version_id,
    s.name as service_name,
    fv.id as feature_version_id,
    f.name as feature_name,
    k.id as key_id,
    k.name as key_name,
    nv.id as new_variation_value_id,
    nv.data as new_variation_value_data,
    ov.id as old_variation_value_id,
    ov.data as old_variation_value_data,
    vc.id as variation_context_id,
    csc.feature_version_service_version_id
FROM changeset_changes csc
    LEFT JOIN service_versions sv ON sv.id = csc.service_version_id
    LEFT JOIN services s ON s.id = sv.service_id
    LEFT JOIN feature_versions fv ON fv.id = csc.feature_version_id
    LEFT JOIN features f ON f.id = fv.feature_id
    LEFT JOIN keys k ON k.id = csc.key_id
    LEFT JOIN variation_values nv ON nv.id = csc.new_variation_value_id
    LEFT JOIN variation_values ov ON ov.id = csc.old_variation_value_id
    LEFT JOIN variation_contexts vc ON vc.id = COALESCE(nv.variation_context_id, ov.variation_context_id)
WHERE changeset_id = @changeset_id
ORDER BY csc.id;
-- name: GetChangeForVariationValue :one
SELECT id,
    type,
    new_variation_value_id,
    old_variation_value_id
FROM changeset_changes
WHERE changeset_id = @changeset_id
    AND (
        old_variation_value_id = @variation_value_id::bigint
        OR new_variation_value_id = @variation_value_id::bigint
    )
LIMIT 1;
-- name: DeleteChange :exec
DELETE FROM changeset_changes
WHERE id = @change_id;
-- name: AddCreateServiceVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        service_version_id,
        previous_service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @service_version_id::bigint,
        sqlc.narg('previous_service_version_id'),
        'create'
    );
-- name: AddCreateFeatureVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        feature_version_id,
        previous_feature_version_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @feature_version_id::bigint,
        sqlc.narg('previous_feature_version_id'),
        @service_version_id::bigint,
        'create'
    );
-- name: AddCreateFeatureVersionServiceVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        feature_version_service_version_id,
        feature_version_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @feature_version_service_version_id::bigint,
        @feature_version_id::bigint,
        @service_version_id::bigint,
        'create'
    );
-- name: AddDeleteFeatureVersionServiceVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        feature_version_service_version_id,
        feature_version_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @feature_version_service_version_id::bigint,
        @feature_version_id::bigint,
        @service_version_id::bigint,
        'delete'
    );
-- name: AddCreateKeyChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        key_id,
        feature_version_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @key_id::bigint,
        @feature_version_id::bigint,
        @service_version_id::bigint,
        'create'
    );
-- name: AddCreateVariationValueChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        new_variation_value_id,
        feature_version_id,
        key_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @new_variation_value_id::bigint,
        @feature_version_id::bigint,
        @key_id::bigint,
        @service_version_id::bigint,
        'create'
    );
-- name: AddDeleteVariationValueChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        old_variation_value_id,
        feature_version_id,
        key_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @old_variation_value_id::bigint,
        @feature_version_id::bigint,
        @key_id::bigint,
        @service_version_id::bigint,
        'delete'
    );
-- name: AddUpdateVariationValueChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        new_variation_value_id,
        old_variation_value_id,
        feature_version_id,
        key_id,
        service_version_id,
        type
    )
VALUES (
        @changeset_id,
        @new_variation_value_id::bigint,
        @old_variation_value_id::bigint,
        @feature_version_id::bigint,
        @key_id::bigint,
        @service_version_id::bigint,
        'update'
    );