-- migrate:up
CREATE TABLE users(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp with time zone,
    name text NOT NULL UNIQUE,
    password TEXT NOT NULL,
    global_administrator boolean NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_users_deleted_at ON users(deleted_at);

CREATE INDEX idx_users_name ON users USING btree(name);

CREATE TABLE value_types(
    id bigserial PRIMARY KEY,
    kind value_type_kind NOT NULL,
    name text NOT NULL UNIQUE
);

CREATE TABLE service_types(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name text NOT NULL UNIQUE
);

CREATE TABLE services(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name text NOT NULL UNIQUE,
    description text NOT NULL,
    service_type_id bigint NOT NULL REFERENCES service_types(id)
);

CREATE TABLE features(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name text NOT NULL UNIQUE,
    description text NOT NULL,
    service_id bigint NOT NULL REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE feature_versions(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    version integer NOT NULL,
    feature_id bigint NOT NULL REFERENCES features(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_feature_versions_unique_version_per_feature ON feature_versions(feature_id, version) WHERE valid_from IS NOT NULL AND valid_to IS NULL;

CREATE INDEX idx_feature_versions_valid_from ON feature_versions(valid_from);

CREATE INDEX idx_feature_versions_valid_to ON feature_versions(valid_to);

CREATE TABLE service_versions(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    service_id bigint NOT NULL REFERENCES services(id),
    version integer NOT NULL,
    published boolean NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX idx_service_versions_unique_version_per_service ON service_versions(service_id, version) WHERE valid_from IS NOT NULL AND valid_to IS NULL;

CREATE INDEX idx_service_versions_valid_from ON service_versions(valid_from);

CREATE INDEX idx_service_versions_valid_to ON service_versions(valid_to);

CREATE TABLE feature_version_service_versions(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    feature_version_id bigint NOT NULL REFERENCES feature_versions(id) ON DELETE CASCADE,
    service_version_id bigint NOT NULL REFERENCES service_versions(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_feature_version_service_versions_unique ON feature_version_service_versions(feature_version_id, service_version_id) WHERE valid_from IS NOT NULL AND valid_to IS NULL;

CREATE INDEX idx_fvsv_valid_from ON feature_version_service_versions(valid_from);

CREATE INDEX idx_fvsv_valid_to ON feature_version_service_versions(valid_to);

CREATE TABLE keys(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    name text NOT NULL,
    description text,
    value_type_id bigint NOT NULL REFERENCES value_types(id),
    feature_version_id bigint NOT NULL REFERENCES feature_versions(id) ON DELETE CASCADE,
    validators_updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_keys_unique_name_per_feature_version ON keys(feature_version_id, name) WHERE valid_from IS NOT NULL AND valid_to IS NULL;

CREATE INDEX idx_keys_valid_from ON keys(valid_from);

CREATE INDEX idx_keys_valid_to ON keys(valid_to);

CREATE TABLE value_validators(
    id bigserial PRIMARY KEY,
    value_type_id bigint REFERENCES value_types(id),
    key_id bigint REFERENCES keys(id) ON DELETE CASCADE,
    validator_type value_validator_type NOT NULL,
    parameter text,
    error_text text,
    CHECK ((value_type_id IS NOT NULL AND key_id IS NULL) OR (value_type_id IS NULL AND key_id IS NOT NULL))
);

CREATE TABLE variation_properties(
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    display_name text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE service_type_variation_properties(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    priority integer NOT NULL,
    service_type_id bigint NOT NULL REFERENCES service_types(id) ON DELETE CASCADE,
    variation_property_id bigint NOT NULL REFERENCES variation_properties(id),
    UNIQUE (service_type_id, variation_property_id)
);

CREATE TABLE variation_property_values(
    id bigserial PRIMARY KEY,
    variation_property_id bigint NOT NULL REFERENCES variation_properties(id) ON DELETE CASCADE,
    value text NOT NULL,
    parent_id bigint REFERENCES variation_property_values(id) ON DELETE CASCADE,
    order_index integer NOT NULL,
    archived boolean NOT NULL DEFAULT FALSE,
    UNIQUE (variation_property_id, value)
);

CREATE INDEX idx_variation_property_values_order_index ON variation_property_values(order_index);

CREATE TABLE variation_contexts(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE variation_values(
    id bigserial PRIMARY KEY,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    key_id bigint NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    variation_context_id bigint NOT NULL REFERENCES variation_contexts(id),
    data text NOT NULL
);

CREATE UNIQUE INDEX idx_one_value_per_key_and_context ON variation_values(key_id, variation_context_id) WHERE valid_from IS NOT NULL AND valid_to IS NULL;

CREATE INDEX idx_variation_values_valid_from ON variation_values(valid_from);

CREATE INDEX idx_variation_values_valid_to ON variation_values(valid_to);

CREATE TABLE variation_context_variation_property_values(
    variation_context_id bigint NOT NULL REFERENCES variation_contexts(id),
    variation_property_value_id bigint NOT NULL REFERENCES variation_property_values(id) ON DELETE CASCADE,
    PRIMARY KEY (variation_context_id, variation_property_value_id)
);

CREATE TABLE user_permissions(
    id bigserial PRIMARY KEY,
    kind user_permission_kind NOT NULL,
    user_id bigint NOT NULL REFERENCES users(id),
    service_id bigint NOT NULL REFERENCES services(id),
    feature_id bigint REFERENCES features(id),
    key_id bigint REFERENCES keys(id),
    variation_context_id bigint REFERENCES variation_contexts(id),
    permission permission_level NOT NULL,
    UNIQUE (user_id, service_id, feature_id, key_id, variation_context_id)
);

CREATE INDEX idx_user_permissions_kind ON user_permissions(kind);

CREATE TABLE changesets(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id bigint NOT NULL REFERENCES users(id),
    state changeset_state NOT NULL,
    applied_at timestamp with time zone
);

CREATE UNIQUE INDEX idx_changesets_one_open_per_user ON changesets(user_id) WHERE state = 'open';

CREATE TABLE changeset_changes(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changeset_id bigint NOT NULL REFERENCES changesets(id),
    type changeset_change_type NOT NULL,
    kind changeset_change_kind NOT NULL,
    feature_version_id bigint REFERENCES feature_versions(id) ON DELETE CASCADE,
    previous_feature_version_id bigint REFERENCES feature_versions(id),
    service_version_id bigint NOT NULL REFERENCES service_versions(id) ON DELETE CASCADE,
    previous_service_version_id bigint REFERENCES service_versions(id),
    feature_version_service_version_id bigint REFERENCES feature_version_service_versions(id) ON DELETE CASCADE,
    key_id bigint REFERENCES keys(id) ON DELETE CASCADE,
    new_variation_value_id bigint REFERENCES variation_values(id) ON DELETE CASCADE,
    old_variation_value_id bigint REFERENCES variation_values(id)
);

CREATE INDEX idx_changeset_changes_kind ON changeset_changes(kind);

CREATE TABLE changeset_actions(
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changeset_id bigint NOT NULL REFERENCES changesets(id),
    user_id bigint NOT NULL REFERENCES users(id),
    type changeset_action_type NOT NULL,
    comment text
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

