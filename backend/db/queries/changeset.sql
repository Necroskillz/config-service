-- name: GetChangesets :many
WITH changeset_services AS (
    SELECT DISTINCT cs.id,
        sv.service_id
    FROM changesets cs
        JOIN changeset_changes csc ON csc.changeset_id = cs.id
        JOIN service_versions sv ON sv.id = csc.service_version_id
    WHERE cs.state = 'committed'
),
user_service_permissions AS (
    SELECT DISTINCT service_id
    FROM user_permissions up
    WHERE up.kind = 'service'
        AND up.user_id = sqlc.narg('approver_id')::bigint
        AND up.permission = 'admin'
),
filtered_changesets AS (
    SELECT cs.id
    FROM changesets cs
    WHERE (
            sqlc.narg('user_id')::bigint IS NULL
            OR cs.user_id = sqlc.narg('user_id')::bigint
        )
        AND (
            sqlc.narg('approver_id')::bigint IS NULL
            OR (
                cs.state = 'committed'
                AND NOT EXISTS (
                    SELECT 1
                    FROM changeset_services sub
                    WHERE sub.id = cs.id
                        AND sub.service_id NOT IN (
                            SELECT service_id
                            FROM user_service_permissions
                        )
                )
            )
        )
),
last_actions AS (
    SELECT DISTINCT ON (changeset_id) changeset_id,
        created_at as last_action_at
    FROM changeset_actions
    ORDER BY changeset_id,
        created_at DESC
),
action_counts AS (
    SELECT changeset_id,
        COUNT(*)::integer as action_count
    FROM changeset_actions
    GROUP BY changeset_id
),
total_count AS (
    SELECT COUNT(*)::integer as total
    FROM filtered_changesets
)
SELECT cs.*,
    COALESCE(la.last_action_at, cs.created_at) as last_action_at,
    COALESCE(ac.action_count, 0)::integer as action_count,
    u.name as user_name,
    tc.total as total_count
FROM filtered_changesets fc
    JOIN changesets cs ON cs.id = fc.id
    JOIN users u ON u.id = cs.user_id
    LEFT JOIN last_actions la ON la.changeset_id = cs.id
    LEFT JOIN action_counts ac ON ac.changeset_id = cs.id
    CROSS JOIN total_count tc
ORDER BY cs.id DESC
LIMIT sqlc.arg('limit')::integer OFFSET sqlc.arg('offset')::integer;
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
-- name: GetRelatedServiceVersionChangesCount :one
SELECT COUNT(*)::integer
FROM changeset_changes csc
WHERE csc.service_version_id = @service_version_id
    AND csc.changeset_id = @changeset_id;
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
-- name: AddChangesetAction :exec
INSERT INTO changeset_actions (changeset_id, user_id, type, comment)
VALUES (
        @changeset_id,
        @user_id,
        @type,
        sqlc.narg('comment')
    );
-- name: GetChangesetActions :many
SELECT ca.id,
    ca.type,
    ca.comment,
    ca.created_at,
    u.id as user_id,
    u.name as user_name
FROM changeset_actions ca
    JOIN users u ON u.id = ca.user_id
WHERE ca.changeset_id = @changeset_id
ORDER BY ca.id;
-- name: GetChangesetChanges :many
SELECT csc.id,
    csc.type,
    csc.kind,
    sv.id as service_version_id,
    csc.previous_service_version_id,
    s.name as service_name,
    sv.version as service_version,
    fv.id as feature_version_id,
    csc.previous_feature_version_id,
    f.name as feature_name,
    fv.version as feature_version,
    k.id as key_id,
    k.name as key_name,
    nv.id as new_variation_value_id,
    nv.data as new_variation_value_data,
    ov.id as old_variation_value_id,
    ov.data as old_variation_value_data,
    vc.id as variation_context_id,
    csc.feature_version_service_version_id
FROM changeset_changes csc
    JOIN service_versions sv ON sv.id = csc.service_version_id
    JOIN services s ON s.id = sv.service_id
    LEFT JOIN feature_versions fv ON fv.id = csc.feature_version_id
    LEFT JOIN features f ON f.id = fv.feature_id
    LEFT JOIN keys k ON k.id = csc.key_id
    LEFT JOIN variation_values nv ON nv.id = csc.new_variation_value_id
    LEFT JOIN variation_values ov ON ov.id = csc.old_variation_value_id
    LEFT JOIN variation_contexts vc ON vc.id = COALESCE(nv.variation_context_id, ov.variation_context_id)
WHERE changeset_id = @changeset_id
ORDER BY csc.id;
-- name: GetChangesetChangesCount :one
SELECT COUNT(*)::integer
FROM changeset_changes csc
WHERE csc.changeset_id = @changeset_id;
-- name: GetApprovableChangesetCount :one
WITH changeset_services AS (
    SELECT DISTINCT cs.id,
        sv.service_id
    FROM changesets cs
        JOIN changeset_changes csc ON csc.changeset_id = cs.id
        JOIN service_versions sv ON sv.id = csc.service_version_id
    WHERE cs.state = 'committed'
),
user_service_permissions AS (
    SELECT DISTINCT service_id
    FROM user_permissions up
    WHERE up.kind = 'service'
        AND up.user_id = @user_id
        AND up.permission = 'admin'
)
SELECT COUNT(DISTINCT cs.id)::integer
FROM changeset_services cs
WHERE NOT EXISTS (
        SELECT 1
        FROM changeset_services sub
        WHERE sub.id = cs.id
            AND sub.service_id NOT IN (
                SELECT service_id
                FROM user_service_permissions
            )
    );
-- name: GetChangeForVariationValue :one
SELECT csc.id,
    csc.type,
    csc.new_variation_value_id,
    csc.old_variation_value_id,
    vv.variation_context_id
FROM changeset_changes csc
    JOIN variation_values vv ON vv.id = COALESCE(
        csc.new_variation_value_id,
        csc.old_variation_value_id
    )
WHERE csc.changeset_id = @changeset_id
    AND (
        csc.old_variation_value_id = @variation_value_id::bigint
        OR csc.new_variation_value_id = @variation_value_id::bigint
    )
LIMIT 1;
-- name: GetChangeForFeatureVersionServiceVersion :one
SELECT csc.id,
    csc.type,
    csc.feature_version_service_version_id
FROM changeset_changes csc
WHERE csc.changeset_id = @changeset_id
    AND csc.service_version_id = @service_version_id::bigint
    AND csc.feature_version_id = @feature_version_id::bigint
    AND csc.kind = 'feature_version_service_version'
LIMIT 1;
-- name: DeleteChange :exec
DELETE FROM changeset_changes
WHERE id = @change_id;
-- name: AddCreateServiceVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        service_version_id,
        previous_service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @service_version_id::bigint,
        sqlc.narg('previous_service_version_id'),
        'create',
        'service_version'
    );
-- name: AddCreateFeatureVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        feature_version_id,
        previous_feature_version_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @feature_version_id::bigint,
        sqlc.narg('previous_feature_version_id'),
        @service_version_id::bigint,
        'create',
        'feature_version'
    );
-- name: AddCreateFeatureVersionServiceVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        feature_version_service_version_id,
        feature_version_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @feature_version_service_version_id::bigint,
        @feature_version_id::bigint,
        @service_version_id::bigint,
        'create',
        'feature_version_service_version'
    );
-- name: AddDeleteFeatureVersionServiceVersionChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        feature_version_service_version_id,
        feature_version_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @feature_version_service_version_id::bigint,
        @feature_version_id::bigint,
        @service_version_id::bigint,
        'delete',
        'feature_version_service_version'
    );
-- name: AddCreateKeyChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        key_id,
        feature_version_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @key_id::bigint,
        @feature_version_id::bigint,
        @service_version_id::bigint,
        'create',
        'key'
    );
-- name: AddCreateVariationValueChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        new_variation_value_id,
        feature_version_id,
        key_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @new_variation_value_id::bigint,
        @feature_version_id::bigint,
        @key_id::bigint,
        @service_version_id::bigint,
        'create',
        'variation_value'
    );
-- name: AddDeleteVariationValueChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        old_variation_value_id,
        feature_version_id,
        key_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @old_variation_value_id::bigint,
        @feature_version_id::bigint,
        @key_id::bigint,
        @service_version_id::bigint,
        'delete',
        'variation_value'
    );
-- name: AddUpdateVariationValueChange :exec
INSERT INTO changeset_changes (
        changeset_id,
        new_variation_value_id,
        old_variation_value_id,
        feature_version_id,
        key_id,
        service_version_id,
        type,
        kind
    )
VALUES (
        @changeset_id,
        @new_variation_value_id::bigint,
        @old_variation_value_id::bigint,
        @feature_version_id::bigint,
        @key_id::bigint,
        @service_version_id::bigint,
        'update',
        'variation_value'
    );