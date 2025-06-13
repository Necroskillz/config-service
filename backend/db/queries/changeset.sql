-- name: GetChangesets :many
WITH changeset_services AS (
    SELECT DISTINCT
        cs.id,
        sv.service_id
    FROM
        changesets cs
        JOIN changeset_changes csc ON csc.changeset_id = cs.id
        JOIN service_versions sv ON sv.id = csc.service_version_id
    WHERE
        cs.state = 'committed'
),
user_service_permissions AS (
    SELECT DISTINCT
        service_id
    FROM
        permissions p
    WHERE
        p.kind = 'service'
        AND p.user_id = sqlc.narg('approver_id')::bigint
        AND p.permission = 'admin'
),
filtered_changesets AS (
    SELECT
        cs.id
    FROM
        changesets cs
    WHERE (sqlc.narg('user_id')::bigint IS NULL
        OR cs.user_id = sqlc.narg('user_id')::bigint)
    AND (sqlc.narg('approver_id')::bigint IS NULL
        OR (cs.state = 'committed'
            AND NOT EXISTS (
                SELECT
                    1
                FROM
                    changeset_services sub
                WHERE
                    sub.id = cs.id
                    AND sub.service_id NOT IN (
                        SELECT
                            service_id
                        FROM
                            user_service_permissions))))
),
last_actions AS (
    SELECT DISTINCT ON (changeset_id)
        changeset_id,
        created_at AS last_action_at
    FROM
        changeset_actions
    ORDER BY
        changeset_id,
        created_at DESC
),
action_counts AS (
    SELECT
        changeset_id,
        COUNT(*)::integer AS action_count
    FROM
        changeset_actions
    GROUP BY
        changeset_id
)
SELECT
    cs.*,
    COALESCE(la.last_action_at, cs.created_at) AS last_action_at,
    COALESCE(ac.action_count, 0)::integer AS action_count,
    u.name AS user_name,
    COUNT(*) OVER ()::integer AS total_count
    FROM
        filtered_changesets fc
        JOIN changesets cs ON cs.id = fc.id
        JOIN users u ON u.id = cs.user_id
        LEFT JOIN last_actions la ON la.changeset_id = cs.id
        LEFT JOIN action_counts ac ON ac.changeset_id = cs.id
    ORDER BY
        cs.id DESC
    LIMIT sqlc.arg('limit')::integer OFFSET sqlc.arg('offset')::integer;

-- name: CreateChangeset :one
INSERT INTO changesets(user_id, state)
    VALUES (@user_id, 'open')
RETURNING
    id;

-- name: GetOpenChangesetIDForUser :one
SELECT
    id
FROM
    changesets
WHERE
    user_id = @user_id
    AND state = 'open'
LIMIT 1;

-- name: GetRelatedServiceVersionChangesCount :one
SELECT
    COUNT(*)::integer
FROM
    changeset_changes csc
WHERE
    csc.service_version_id = @service_version_id
    AND csc.changeset_id = @changeset_id;

-- name: GetRelatedFeatureVersionChangesCount :one
SELECT
    COUNT(*)::integer
FROM
    changeset_changes csc
WHERE
    csc.feature_version_id = @feature_version_id::bigint
    AND csc.changeset_id = @changeset_id
    AND (csc.kind = 'feature_version'
        OR csc.kind = 'key'
        OR csc.kind = 'variation_value');

-- name: GetRelatedKeyChangesCount :one
SELECT
    COUNT(*)::integer
FROM
    changeset_changes csc
WHERE
    csc.key_id = @key_id::bigint
    AND csc.changeset_id = @changeset_id;

-- name: GetChangeset :one
SELECT
    cs.id,
    cs.state,
    cs.applied_at,
    u.id AS user_id,
    u.name AS user_name
FROM
    changesets cs
    JOIN users u ON u.id = cs.user_id
WHERE
    cs.id = @changeset_id
LIMIT 1;

-- name: LockChangesetForUpdate :one
SELECT
    cs.id
FROM
    changesets cs
WHERE
    cs.id = @changeset_id
FOR UPDATE;

-- name: SetChangesetState :exec
UPDATE
    changesets
SET
    state = @state,
    applied_at = @applied_at,
    updated_at = now()
WHERE
    id = @changeset_id;

-- name: AddChangesetAction :exec
INSERT INTO changeset_actions(changeset_id, user_id, type, comment)
    VALUES (@changeset_id, @user_id, @type, sqlc.narg('comment'));

-- name: GetChangesetActions :many
SELECT
    ca.id,
    ca.type,
    ca.comment,
    ca.created_at,
    u.id AS user_id,
    u.name AS user_name
FROM
    changeset_actions ca
    JOIN users u ON u.id = ca.user_id
WHERE
    ca.changeset_id = @changeset_id
ORDER BY
    ca.id;

-- name: GetChangesetChanges :many
WITH links AS (
    SELECT
        fvsv.id AS feature_version_service_version_id,
        fv.feature_id AS feature_id,
        fvsv.service_version_id
    FROM
        feature_version_service_versions fvsv
        JOIN feature_versions fv ON fv.id = fvsv.feature_version_id
    WHERE
        fvsv.valid_from IS NOT NULL
        AND fvsv.valid_to IS NULL
),
last_feature_versions AS (
    SELECT
        fv.feature_id,
        MAX(fv.version)::int AS last_version
    FROM
        feature_versions fv
    WHERE
        fv.valid_from IS NOT NULL
        AND fv.valid_to IS NULL
    GROUP BY
        fv.feature_id
),
last_service_versions AS (
    SELECT
        sv.service_id,
        MAX(sv.version)::int AS last_version
    FROM
        service_versions sv
    WHERE
        sv.valid_from IS NOT NULL
        AND sv.valid_to IS NULL
    GROUP BY
        sv.service_id
)
SELECT
    csc.id,
    csc.type,
    csc.kind,
    csc.created_at,
    sv.id AS service_version_id,
    csc.previous_service_version_id,
    s.name AS service_name,
    s.id AS service_id,
    sv.version AS service_version,
    sv.published AS service_version_published,
    fv.id AS feature_version_id,
    csc.previous_feature_version_id,
    f.name AS feature_name,
    f.id AS feature_id,
    fv.version AS feature_version,
    fv.valid_to AS feature_version_valid_to,
    k.id AS key_id,
    k.name AS key_name,
    k.valid_to AS key_valid_to,
    k.validators_updated_at AS key_validators_updated_at,
    nv.id AS new_variation_value_id,
    nv.data AS new_variation_value_data,
    ov.id AS old_variation_value_id,
    ov.data AS old_variation_value_data,
    ov.valid_to AS old_variation_value_valid_to,
    vc.id AS variation_context_id,
    fvsv.id AS feature_version_service_version_id,
    fvsv.valid_to AS feature_version_service_version_valid_to,
    evv.variation_context_id AS existing_variation_context_id,
    evv.data AS existing_value_data,
    ek.id AS existing_key_id,
    el.feature_version_service_version_id AS existing_feature_version_service_version_id,
    COALESCE(lfv.last_version, 0) AS last_feature_version_version,
    COALESCE(lsv.last_version, 0) AS last_service_version_version
FROM
    changeset_changes csc
    JOIN service_versions sv ON sv.id = csc.service_version_id
    JOIN services s ON s.id = sv.service_id
    LEFT JOIN feature_version_service_versions fvsv ON fvsv.id = csc.feature_version_service_version_id
    LEFT JOIN feature_versions fv ON fv.id = csc.feature_version_id
    LEFT JOIN features f ON f.id = fv.feature_id
    LEFT JOIN keys k ON k.id = csc.key_id
    LEFT JOIN variation_values nv ON nv.id = csc.new_variation_value_id
    LEFT JOIN variation_values ov ON ov.id = csc.old_variation_value_id
    LEFT JOIN variation_contexts vc ON vc.id = COALESCE(nv.variation_context_id, ov.variation_context_id)
    LEFT JOIN variation_values evv ON evv.variation_context_id = vc.id
        AND evv.key_id = k.id
        AND evv.valid_from IS NOT NULL
        AND evv.valid_to IS NULL
    LEFT JOIN keys ek ON ek.id <> k.id
        AND ek.name = k.name
        AND ek.feature_version_id = k.feature_version_id
        AND ek.valid_from IS NOT NULL
        AND ek.valid_to IS NULL
    LEFT JOIN links el ON el.service_version_id = sv.id
        AND el.feature_id = f.id
    LEFT JOIN last_feature_versions lfv ON lfv.feature_id = f.id
    LEFT JOIN last_service_versions lsv ON lsv.service_id = sv.service_id
WHERE
    changeset_id = @changeset_id
ORDER BY
    csc.id;

-- name: GetChangesetChange :one
SELECT
    csc.*,
    vv.variation_context_id AS variation_context_id
FROM
    changeset_changes csc
    LEFT JOIN variation_values vv ON vv.id = COALESCE(csc.new_variation_value_id, csc.old_variation_value_id)
WHERE
    csc.id = @change_id
LIMIT 1;

-- name: GetChangesetChangesCount :one
SELECT
    COUNT(*)::integer
FROM
    changeset_changes csc
WHERE
    csc.changeset_id = @changeset_id;

-- name: GetApprovableChangesetCount :one
WITH changeset_services AS (
    SELECT DISTINCT
        cs.id,
        sv.service_id
    FROM
        changesets cs
        JOIN changeset_changes csc ON csc.changeset_id = cs.id
        JOIN service_versions sv ON sv.id = csc.service_version_id
    WHERE
        cs.state = 'committed'
),
user_service_permissions AS (
    SELECT DISTINCT
        service_id
    FROM
        users u
        LEFT JOIN user_group_memberships ugm ON ugm.user_id = u.id
        JOIN permissions p ON p.user_group_id = ugm.user_group_id
            OR p.user_id = u.id
    WHERE
        p.kind = 'service'
        AND u.id = @user_id
        AND p.permission = 'admin'
)
SELECT
    COUNT(DISTINCT cs.id)::integer
FROM
    changeset_services cs
WHERE
    NOT EXISTS (
        SELECT
            1
        FROM
            changeset_services sub
        WHERE
            sub.id = cs.id
            AND sub.service_id NOT IN (
                SELECT
                    service_id
                FROM
                    user_service_permissions));

-- name: GetChangeForKey :one
SELECT
    csc.id,
    csc.type,
    csc.key_id,
    csc.feature_version_id,
    csc.service_version_id
FROM
    changeset_changes csc
WHERE
    csc.changeset_id = @changeset_id
    AND csc.kind = 'key'
    AND csc.key_id = @key_id::bigint
LIMIT 1;

-- name: GetChangeForVariationValue :one
SELECT
    csc.id,
    csc.type,
    nvv.id AS new_variation_value_id,
    ovv.id AS old_variation_value_id,
    COALESCE(nvv.variation_context_id, ovv.variation_context_id) AS variation_context_id,
    nvv.data AS new_variation_value_data,
    ovv.data AS old_variation_value_data
FROM
    changeset_changes csc
    LEFT JOIN variation_values nvv ON nvv.id = csc.new_variation_value_id
    LEFT JOIN variation_values ovv ON ovv.id = csc.old_variation_value_id
WHERE
    csc.changeset_id = @changeset_id
    AND (csc.old_variation_value_id = @variation_value_id::bigint
        OR csc.new_variation_value_id = @variation_value_id::bigint)
LIMIT 1;

-- name: GetDeleteChangeForVariationContextID :one
SELECT
    csc.id,
    csc.type,
    vv.id AS variation_value_id,
    vv.data AS variation_value_data
FROM
    changeset_changes csc
    JOIN variation_values vv ON vv.id = csc.old_variation_value_id
WHERE
    csc.changeset_id = @changeset_id
    AND vv.variation_context_id = @variation_context_id
    AND vv.key_id = @key_id
LIMIT 1;

-- name: GetChangeForFeatureVersionServiceVersion :one
SELECT
    csc.id,
    csc.type,
    csc.feature_version_service_version_id
FROM
    changeset_changes csc
WHERE
    csc.changeset_id = @changeset_id
    AND csc.service_version_id = @service_version_id::bigint
    AND csc.feature_version_id = @feature_version_id::bigint
    AND csc.kind = 'feature_version_service_version'
LIMIT 1;

-- name: DeleteChange :exec
DELETE FROM changeset_changes
WHERE id = @change_id;

-- name: DeleteChangesForChangeset :exec
DELETE FROM changeset_changes
WHERE changeset_id = @changeset_id;

-- name: AddCreateServiceVersionChange :exec
INSERT INTO changeset_changes(changeset_id, service_version_id, previous_service_version_id, type, kind)
    VALUES (@changeset_id, @service_version_id::bigint, sqlc.narg('previous_service_version_id'), 'create', 'service_version');

-- name: AddCreateFeatureVersionChange :exec
INSERT INTO changeset_changes(changeset_id, feature_version_id, previous_feature_version_id, service_version_id, type, kind)
    VALUES (@changeset_id, @feature_version_id::bigint, sqlc.narg('previous_feature_version_id'), @service_version_id::bigint, 'create', 'feature_version');

-- name: AddCreateFeatureVersionServiceVersionChange :exec
INSERT INTO changeset_changes(changeset_id, feature_version_service_version_id, feature_version_id, service_version_id, type, kind)
    VALUES (@changeset_id, @feature_version_service_version_id::bigint, @feature_version_id::bigint, @service_version_id::bigint, 'create', 'feature_version_service_version');

-- name: AddDeleteFeatureVersionServiceVersionChange :exec
INSERT INTO changeset_changes(changeset_id, feature_version_service_version_id, feature_version_id, service_version_id, type, kind)
    VALUES (@changeset_id, @feature_version_service_version_id::bigint, @feature_version_id::bigint, @service_version_id::bigint, 'delete', 'feature_version_service_version');

-- name: AddCreateKeyChange :exec
INSERT INTO changeset_changes(changeset_id, key_id, feature_version_id, service_version_id, type, kind)
    VALUES (@changeset_id, @key_id::bigint, @feature_version_id::bigint, @service_version_id::bigint, 'create', 'key');

-- name: AddDeleteKeyChange :exec
INSERT INTO changeset_changes(changeset_id, key_id, feature_version_id, service_version_id, type, kind)
    VALUES (@changeset_id, @key_id::bigint, @feature_version_id::bigint, @service_version_id::bigint, 'delete', 'key');

-- name: AddCreateVariationValueChange :exec
INSERT INTO changeset_changes(changeset_id, new_variation_value_id, feature_version_id, key_id, service_version_id, type, kind)
    VALUES (@changeset_id, @new_variation_value_id::bigint, @feature_version_id::bigint, @key_id::bigint, @service_version_id::bigint, 'create', 'variation_value');

-- name: AddDeleteVariationValueChange :exec
INSERT INTO changeset_changes(changeset_id, old_variation_value_id, feature_version_id, key_id, service_version_id, type, kind)
    VALUES (@changeset_id, @old_variation_value_id::bigint, @feature_version_id::bigint, @key_id::bigint, @service_version_id::bigint, 'delete', 'variation_value');

-- name: AddUpdateVariationValueChange :exec
INSERT INTO changeset_changes(changeset_id, new_variation_value_id, old_variation_value_id, feature_version_id, key_id, service_version_id, type, kind)
    VALUES (@changeset_id, @new_variation_value_id::bigint, @old_variation_value_id::bigint, @feature_version_id::bigint, @key_id::bigint, @service_version_id::bigint, 'update', 'variation_value');

-- name: AddChanges :copyfrom
INSERT INTO changeset_changes(changeset_id, new_variation_value_id, old_variation_value_id, key_id, feature_version_id, service_version_id, type, kind)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetNextChangesetsRelatedToServiceVersions :many
SELECT
    cs.id AS changeset_id
FROM
    changesets cs
    JOIN changeset_changes csc ON csc.changeset_id = cs.id
WHERE
    csc.service_version_id = ANY (@service_version_ids::bigint[])
    AND cs.applied_at > @applied_after
GROUP BY
    cs.id
HAVING
    COUNT(csc.id) > 0
LIMIT 100;

-- name: GetLastAppliedChangeset :one
SELECT
    *
FROM
    changesets
WHERE
    applied_at IS NOT NULL
ORDER BY
    applied_at DESC
LIMIT 1;

-- name: GetChangeHistory :many
SELECT
    csc.id,
    csc.type,
    csc.kind,
    cs.applied_at::timestamptz AS applied_at,
    u.name AS user_name,
    u.id AS user_id,
    cs.id AS changeset_id,
    sv.id AS service_version_id,
    csc.previous_service_version_id,
    s.name AS service_name,
    s.id AS service_id,
    sv.version AS service_version,
    fv.id AS feature_version_id,
    csc.previous_feature_version_id,
    f.name AS feature_name,
    f.id AS feature_id,
    fv.version AS feature_version,
    k.id AS key_id,
    k.name AS key_name,
    nv.id AS new_variation_value_id,
    nv.data AS new_variation_value_data,
    ov.id AS old_variation_value_id,
    ov.data AS old_variation_value_data,
    vc.id AS variation_context_id,
    COUNT(*) OVER ()::integer AS total_count
FROM
    changeset_changes csc
    JOIN changesets cs ON cs.id = csc.changeset_id
    JOIN service_versions sv ON sv.id = csc.service_version_id
    JOIN services s ON s.id = sv.service_id
    JOIN users u ON u.id = cs.user_id
    LEFT JOIN feature_version_service_versions fvsv ON fvsv.id = csc.feature_version_service_version_id
    LEFT JOIN feature_versions fv ON fv.id = csc.feature_version_id
    LEFT JOIN features f ON f.id = fv.feature_id
    LEFT JOIN keys k ON k.id = csc.key_id
    LEFT JOIN variation_values nv ON nv.id = csc.new_variation_value_id
    LEFT JOIN variation_values ov ON ov.id = csc.old_variation_value_id
    LEFT JOIN variation_contexts vc ON vc.id = COALESCE(nv.variation_context_id, ov.variation_context_id)
WHERE
    cs.applied_at IS NOT NULL
    AND (sqlc.narg('service_id')::bigint IS NULL OR s.id = sqlc.narg('service_id')::bigint)
    AND (sqlc.narg('service_version_id')::bigint IS NULL OR sv.id = sqlc.narg('service_version_id')::bigint)
    AND (sqlc.narg('feature_id')::bigint IS NULL OR f.id = sqlc.narg('feature_id')::bigint)
    AND (sqlc.narg('feature_version_id')::bigint IS NULL OR fv.id = sqlc.narg('feature_version_id')::bigint)
    AND (sqlc.narg('key_name')::text IS NULL OR k.name = sqlc.narg('key_name')::text)
    AND (sqlc.narg('variation_context_id')::bigint IS NULL OR vc.id = sqlc.narg('variation_context_id')::bigint)
    AND (sqlc.narg('kinds')::text[] IS NULL OR csc.kind = ANY(sqlc.narg('kinds')::text[]::changeset_change_kind[]))
    AND (sqlc.narg('from')::timestamptz IS NULL OR cs.applied_at >= sqlc.narg('from')::timestamptz)
    AND (sqlc.narg('to')::timestamptz IS NULL OR cs.applied_at <= sqlc.narg('to')::timestamptz)
ORDER BY
    cs.applied_at DESC, csc.id DESC
LIMIT sqlc.arg('limit')::integer OFFSET sqlc.arg('offset')::integer;