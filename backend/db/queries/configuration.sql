-- name: GetConfiguration :many
SELECT
    f.id AS feature_id,
    k.id AS key_id,
    s.service_type_id AS service_type_id,
    f.name AS feature_name,
    k.name AS key_name,
    vt.kind AS value_type,
    vv.data AS data,
    vv.variation_context_id
FROM
    variation_values vv
    JOIN keys k ON k.id = vv.key_id
    JOIN feature_versions fv ON fv.id = k.feature_version_id
    JOIN features f ON f.id = fv.feature_id
    JOIN feature_version_service_versions fvsv ON fvsv.feature_version_id = fv.id
    JOIN service_versions sv ON sv.id = fvsv.service_version_id
    JOIN value_types vt ON vt.id = k.value_type_id
    JOIN services s ON s.id = sv.service_id
WHERE
    sv.id = ANY (@service_version_ids::bigint[])
    AND CASE WHEN @is_applied = TRUE THEN
        vv.valid_from <= @timestamp::timestamptz
        AND (@timestamp::timestamptz < vv.valid_to
            OR vv.valid_to IS NULL)
        AND fv.valid_from <= @timestamp::timestamptz
        AND (@timestamp::timestamptz < fv.valid_to
            OR fv.valid_to IS NULL)
        AND fvsv.valid_from <= @timestamp::timestamptz
        AND (@timestamp::timestamptz < fvsv.valid_to
            OR fvsv.valid_to IS NULL)
        AND sv.valid_from <= @timestamp::timestamptz
        AND (@timestamp::timestamptz < sv.valid_to
            OR sv.valid_to IS NULL)
    ELSE
        is_variation_value_valid_in_changeset(vv, @changeset_id)
        AND is_feature_version_valid_in_changeset(fv, @changeset_id)
        AND is_link_valid_in_changeset(fvsv, @changeset_id)
        AND is_service_version_valid_in_changeset(sv, @changeset_id)
    END
ORDER BY
    f.name,
    k.name;
