-- migrate:up
CREATE FUNCTION is_service_version_valid_in_changeset(sv service_versions, changeset_id bigint)
    RETURNS boolean
    AS $$
    SELECT
((sv.valid_from IS NOT NULL
                AND sv.valid_to IS NULL
                AND NOT EXISTS(
                    SELECT
                        csc.id
                    FROM
                        changeset_changes csc
                    WHERE
                        csc.changeset_id = @changeset_id
                        AND csc.kind = 'service_version'
                        AND csc.type = 'delete'
                        AND csc.service_version_id = sv.id
                    LIMIT 1)))
    OR(sv.valid_from IS NULL
        AND EXISTS(
            SELECT
                csc.id
            FROM
                changeset_changes csc
            WHERE
                csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.kind = 'service_version'
                AND csc.service_version_id = sv.id
            LIMIT 1))
$$
LANGUAGE sql
IMMUTABLE;

CREATE FUNCTION is_feature_version_valid_in_changeset(fv feature_versions, changeset_id bigint)
    RETURNS boolean
    AS $$
    SELECT
((fv.valid_from IS NOT NULL
                AND fv.valid_to IS NULL
                AND NOT EXISTS(
                    SELECT
                        csc.id
                    FROM
                        changeset_changes csc
                    WHERE
                        csc.changeset_id = @changeset_id
                        AND csc.kind = 'feature_version'
                        AND csc.type = 'delete'
                        AND csc.feature_version_id = fv.id
                    LIMIT 1)))
    OR(fv.valid_from IS NULL
        AND EXISTS(
            SELECT
                csc.id
            FROM
                changeset_changes csc
            WHERE
                csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.kind = 'feature_version'
                AND csc.feature_version_id = fv.id
            LIMIT 1))
$$
LANGUAGE sql
IMMUTABLE;

CREATE FUNCTION is_link_valid_in_changeset(fvsv feature_version_service_versions, changeset_id bigint)
    RETURNS boolean
    AS $$
    SELECT
((fvsv.valid_from IS NOT NULL
                AND fvsv.valid_to IS NULL
                AND NOT EXISTS(
                    SELECT
                        csc.id
                    FROM
                        changeset_changes csc
                    WHERE
                        csc.changeset_id = @changeset_id
                        AND csc.kind = 'feature_version_service_version'
                        AND csc.type = 'delete'
                        AND csc.feature_version_service_version_id = fvsv.id
                    LIMIT 1)))
    OR(fvsv.valid_from IS NULL
        AND EXISTS(
            SELECT
                csc.id
            FROM
                changeset_changes csc
            WHERE
                csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.kind = 'feature_version_service_version'
                AND csc.feature_version_service_version_id = fvsv.id
            LIMIT 1))
$$
LANGUAGE sql
IMMUTABLE;

CREATE FUNCTION is_key_valid_in_changeset(k keys, changeset_id bigint)
    RETURNS boolean
    AS $$
    SELECT
((k.valid_from IS NOT NULL
                AND k.valid_to IS NULL
                AND NOT EXISTS(
                    SELECT
                        csc.id
                    FROM
                        changeset_changes csc
                    WHERE
                        csc.changeset_id = @changeset_id
                        AND csc.kind = 'key'
                        AND csc.type = 'delete'
                        AND csc.key_id = k.id
                    LIMIT 1)))
    OR(k.valid_from IS NULL
        AND EXISTS(
            SELECT
                csc.id
            FROM
                changeset_changes csc
            WHERE
                csc.changeset_id = @changeset_id
                AND csc.type = 'create'
                AND csc.kind = 'key'
                AND csc.key_id = k.id
            LIMIT 1))
$$
LANGUAGE sql
IMMUTABLE;

CREATE FUNCTION is_variation_value_valid_in_changeset(vv variation_values, changeset_id bigint)
    RETURNS boolean
    AS $$
    SELECT
((vv.valid_from IS NOT NULL
                AND vv.valid_to IS NULL
                AND NOT EXISTS(
                    SELECT
                        csc.id
                    FROM
                        changeset_changes csc
                    WHERE
                        csc.changeset_id = @changeset_id
                        AND csc.kind = 'variation_value'
                        AND csc.old_variation_value_id = vv.id
                    LIMIT 1)))
    OR(vv.valid_from IS NULL
        AND EXISTS(
            SELECT
                csc.id
            FROM
                changeset_changes csc
            WHERE
                csc.changeset_id = @changeset_id
                AND csc.kind = 'variation_value'
                AND csc.new_variation_value_id = vv.id
            LIMIT 1))
$$
LANGUAGE sql
IMMUTABLE;

-- migrate:down
DROP FUNCTION is_service_version_valid_in_changeset(service_versions, bigint);

DROP FUNCTION is_feature_version_valid_in_changeset(feature_versions, bigint);

DROP FUNCTION is_link_valid_in_changeset(feature_version_service_versions, bigint);

DROP FUNCTION is_key_valid_in_changeset(keys, bigint);

DROP FUNCTION is_variation_value_valid_in_changeset(variation_values, bigint);

