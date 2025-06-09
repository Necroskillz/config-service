-- migrate:up
CREATE TYPE changeset_state AS ENUM(
    'open',
    'committed',
    'applied',
    'rejected',
    'discarded',
    'stashed'
);

CREATE TYPE changeset_action_type AS ENUM(
    'apply',
    'discard',
    'stash',
    'commit',
    'reopen',
    'comment'
);

CREATE TYPE changeset_change_kind AS ENUM(
    'feature_version',
    'service_version',
    'feature_version_service_version',
    'key',
    'variation_value'
);

CREATE TYPE changeset_change_type AS ENUM(
    'create',
    'update',
    'delete'
);

CREATE TYPE permission_kind AS ENUM(
    'service',
    'feature',
    'key',
    'variation'
);

CREATE TYPE permission_level AS ENUM(
    'editor',
    'admin'
);

CREATE TYPE value_validator_type AS ENUM(
    'required',
    'min_length',
    'max_length',
    'min',
    'max',
    'min_decimal',
    'max_decimal',
    'regex',
    'json_schema',
    'valid_json',
    'valid_integer',
    'valid_decimal',
    'valid_regex'
);

CREATE TYPE value_type_kind AS ENUM(
    'string',
    'integer',
    'decimal',
    'boolean',
    'json'
);

-- migrate:down
DROP TYPE permission_level;

DROP TYPE changeset_change_type;

DROP TYPE changeset_state;

DROP TYPE changeset_action_type;

DROP TYPE changeset_change_kind;

DROP TYPE user_permission_kind;

DROP TYPE value_type_kind;

DROP TYPE value_validator_type;

