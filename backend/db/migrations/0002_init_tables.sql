-- migrate:up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    global_administrator BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE TABLE value_types (
    id BIGSERIAL PRIMARY KEY,
    kind value_type_kind NOT NULL,
    name TEXT NOT NULL
);
CREATE TABLE service_types (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL
);
CREATE TABLE services (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    service_type_id BIGINT NOT NULL REFERENCES service_types(id)
);
CREATE TABLE features (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    service_id BIGINT NOT NULL REFERENCES services(id) ON DELETE CASCADE
);
CREATE TABLE feature_versions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL,
    feature_id BIGINT NOT NULL REFERENCES features(id) ON DELETE CASCADE
);
CREATE INDEX idx_feature_versions_valid_from ON feature_versions(valid_from);
CREATE INDEX idx_feature_versions_valid_to ON feature_versions(valid_to);
CREATE TABLE service_versions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    service_id BIGINT NOT NULL REFERENCES services(id),
    version INTEGER NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_service_versions_valid_from ON service_versions(valid_from);
CREATE INDEX idx_service_versions_valid_to ON service_versions(valid_to);
CREATE TABLE feature_version_service_versions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    feature_version_id BIGINT NOT NULL REFERENCES feature_versions(id) ON DELETE CASCADE,
    service_version_id BIGINT NOT NULL REFERENCES service_versions(id) ON DELETE CASCADE
);
CREATE INDEX idx_fvsv_valid_from ON feature_version_service_versions(valid_from);
CREATE INDEX idx_fvsv_valid_to ON feature_version_service_versions(valid_to);
CREATE TABLE keys (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    name TEXT NOT NULL,
    description TEXT,
    value_type_id BIGINT NOT NULL REFERENCES value_types(id),
    feature_version_id BIGINT NOT NULL REFERENCES feature_versions(id) ON DELETE CASCADE,
    validators_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_keys_valid_from ON keys(valid_from);
CREATE INDEX idx_keys_valid_to ON keys(valid_to);
CREATE TABLE value_validators (
    id BIGSERIAL PRIMARY KEY,
    value_type_id BIGINT REFERENCES value_types(id),
    key_id BIGINT REFERENCES keys(id) ON DELETE CASCADE,
    validator_type value_validator_type NOT NULL,
    parameter TEXT,
    error_text TEXT,
    CHECK (
        (
            value_type_id IS NOT NULL
            AND key_id IS NULL
        )
        OR (
            value_type_id IS NULL
            AND key_id IS NOT NULL
        )
    )
);
CREATE TABLE variation_properties (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TABLE service_type_variation_properties (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    priority INTEGER NOT NULL,
    service_type_id BIGINT NOT NULL REFERENCES service_types(id),
    variation_property_id BIGINT NOT NULL REFERENCES variation_properties(id),
    archived BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TABLE variation_property_values (
    id BIGSERIAL PRIMARY KEY,
    variation_property_id BIGINT NOT NULL REFERENCES variation_properties(id),
    value TEXT NOT NULL,
    parent_id BIGINT REFERENCES variation_property_values(id),
    order_index INTEGER NOT NULL,
    archived BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_variation_property_values_order_index ON variation_property_values(order_index);
CREATE TABLE variation_contexts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE variation_values (
    id BIGSERIAL PRIMARY KEY,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    key_id BIGINT NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    variation_context_id BIGINT NOT NULL REFERENCES variation_contexts(id),
    data TEXT NOT NULL
);
CREATE INDEX idx_variation_values_valid_from ON variation_values(valid_from);
CREATE INDEX idx_variation_values_valid_to ON variation_values(valid_to);
CREATE TABLE variation_context_variation_property_values (
    variation_context_id BIGINT NOT NULL REFERENCES variation_contexts(id),
    variation_property_value_id BIGINT NOT NULL REFERENCES variation_property_values(id) ON DELETE CASCADE,
    PRIMARY KEY (
        variation_context_id,
        variation_property_value_id
    )
);
CREATE TABLE user_permissions (
    id BIGSERIAL PRIMARY KEY,
    kind user_permission_kind NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    service_id BIGINT NOT NULL REFERENCES services(id),
    feature_id BIGINT REFERENCES features(id),
    key_id BIGINT REFERENCES keys(id),
    variation_context_id BIGINT REFERENCES variation_contexts(id),
    permission permission_level NOT NULL
);
CREATE TABLE changesets (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id BIGINT NOT NULL REFERENCES users(id),
    state changeset_state NOT NULL
);
CREATE TABLE changeset_changes (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changeset_id BIGINT NOT NULL REFERENCES changesets(id),
    type changeset_change_type NOT NULL,
    kind changeset_change_kind NOT NULL,
    feature_version_id BIGINT REFERENCES feature_versions(id) ON DELETE CASCADE,
    previous_feature_version_id BIGINT REFERENCES feature_versions(id),
    service_version_id BIGINT NOT NULL REFERENCES service_versions(id) ON DELETE CASCADE,
    previous_service_version_id BIGINT REFERENCES service_versions(id),
    feature_version_service_version_id BIGINT REFERENCES feature_version_service_versions(id) ON DELETE CASCADE,
    key_id BIGINT REFERENCES keys(id) ON DELETE CASCADE,
    new_variation_value_id BIGINT REFERENCES variation_values(id) ON DELETE CASCADE,
    old_variation_value_id BIGINT REFERENCES variation_values(id)
);
CREATE INDEX idx_changeset_changes_kind ON changeset_changes(kind);
CREATE TABLE changeset_actions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changeset_id BIGINT NOT NULL REFERENCES changesets(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    type changeset_action_type NOT NULL,
    comment TEXT
);
-- migrate:down
DROP TABLE value_validators;
DROP TABLE changeset_actions;
DROP TABLE changeset_changes;
DROP TABLE changesets;
DROP TABLE user_permissions;
DROP TABLE variation_context_variation_property_values;
DROP TABLE variation_values;
DROP TABLE variation_contexts;
DROP TABLE variation_property_values;
DROP TABLE service_type_variation_properties;
DROP TABLE variation_properties;
DROP TABLE keys;
DROP TABLE feature_version_service_versions;
DROP TABLE service_versions;
DROP TABLE feature_versions;
DROP TABLE features;
DROP TABLE services;
DROP TABLE service_types;
DROP TABLE value_types;
DROP TABLE users;