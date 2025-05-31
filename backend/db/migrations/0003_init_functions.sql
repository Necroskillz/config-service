-- migrate:up
CREATE FUNCTION valid_service_versions_in_changeset(in_changeset_id bigint)
    RETURNS TABLE(id bigint)
    AS $$
    SELECT sv.id
    FROM service_versions sv
    WHERE (
        -- Currently valid and not being deleted
        (sv.valid_from IS NOT NULL 
         AND sv.valid_to IS NULL 
         AND NOT EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.kind = 'service_version' 
               AND csc.type = 'delete' 
               AND csc.service_version_id = sv.id
         ))
        OR
        -- Being created in this changeset
        (sv.valid_from IS NULL 
         AND EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.type = 'create' 
               AND csc.kind = 'service_version' 
               AND csc.service_version_id = sv.id
         ))
    )
$$
LANGUAGE sql
STABLE;

CREATE FUNCTION valid_feature_versions_in_changeset(in_changeset_id bigint)
    RETURNS TABLE(id bigint)
    AS $$
    SELECT fv.id
    FROM feature_versions fv
    WHERE (
        (fv.valid_from IS NOT NULL 
         AND fv.valid_to IS NULL 
         AND NOT EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.kind = 'feature_version' 
               AND csc.type = 'delete' 
               AND csc.feature_version_id = fv.id
         ))
        OR
        (fv.valid_from IS NULL 
         AND EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.type = 'create' 
               AND csc.kind = 'feature_version' 
               AND csc.feature_version_id = fv.id
         ))
    )
$$
LANGUAGE sql
STABLE;

CREATE FUNCTION valid_links_in_changeset(in_changeset_id bigint)
    RETURNS TABLE(id bigint)
    AS $$
    SELECT fvsv.id
    FROM feature_version_service_versions fvsv
    WHERE (
        (fvsv.valid_from IS NOT NULL 
         AND fvsv.valid_to IS NULL 
         AND NOT EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.kind = 'feature_version_service_version' 
               AND csc.type = 'delete' 
               AND csc.feature_version_service_version_id = fvsv.id
         ))
        OR
        (fvsv.valid_from IS NULL 
         AND EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.type = 'create' 
               AND csc.kind = 'feature_version_service_version' 
               AND csc.feature_version_service_version_id = fvsv.id
         ))
    )
$$
LANGUAGE sql
STABLE;

CREATE FUNCTION valid_keys_in_changeset(in_changeset_id bigint)
    RETURNS TABLE(id bigint)
    AS $$
    SELECT k.id
    FROM keys k
    WHERE (
        (k.valid_from IS NOT NULL 
         AND k.valid_to IS NULL 
         AND NOT EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.kind = 'key' 
               AND csc.type = 'delete' 
               AND csc.key_id = k.id
         ))
        OR
        (k.valid_from IS NULL 
         AND EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.type = 'create' 
               AND csc.kind = 'key' 
               AND csc.key_id = k.id
         ))
    )
$$
LANGUAGE sql
STABLE;

CREATE FUNCTION valid_variation_values_in_changeset(in_changeset_id bigint)
    RETURNS TABLE(id bigint)
    AS $$
    SELECT vv.id
    FROM variation_values vv
    WHERE (
        (vv.valid_from IS NOT NULL 
         AND vv.valid_to IS NULL 
         AND NOT EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.kind = 'variation_value' 
               AND csc.old_variation_value_id = vv.id
         ))
        OR
        (vv.valid_from IS NULL 
         AND EXISTS(
             SELECT 1 FROM changeset_changes csc 
             WHERE csc.changeset_id = in_changeset_id 
               AND csc.kind = 'variation_value' 
               AND csc.new_variation_value_id = vv.id
         ))
    )
$$
LANGUAGE sql
STABLE;

-- migrate:down
DROP FUNCTION valid_variation_values_in_changeset(bigint);
DROP FUNCTION valid_keys_in_changeset(bigint);
DROP FUNCTION valid_links_in_changeset(bigint);
DROP FUNCTION valid_feature_versions_in_changeset(bigint);
DROP FUNCTION valid_service_versions_in_changeset(bigint); 