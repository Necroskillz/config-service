-- migrate:up
DO $$
DECLARE
    string_type_id bigint;
    boolean_type_id bigint;
    integer_type_id bigint;
    decimal_type_id bigint;
    json_type_id bigint;
BEGIN
    INSERT INTO value_types(kind, name) VALUES ('string', 'String')
    RETURNING id INTO string_type_id;
    
    INSERT INTO value_types(kind, name) VALUES ('boolean', 'Boolean')
    RETURNING id INTO boolean_type_id;
    
    INSERT INTO value_types(kind, name) VALUES ('integer', 'Integer')
    RETURNING id INTO integer_type_id;
    
    INSERT INTO value_types(kind, name) VALUES ('decimal', 'Decimal')
    RETURNING id INTO decimal_type_id;
    
    INSERT INTO value_types(kind, name) VALUES ('json', 'JSON')
    RETURNING id INTO json_type_id;

    INSERT INTO value_validators(value_type_id, validator_type, parameter, error_text) VALUES
        (boolean_type_id, 'required', NULL, NULL),
        (boolean_type_id, 'regex', '^TRUE|FALSE$', 'Value must be TRUE or FALSE'),
        (integer_type_id, 'required', NULL, NULL),
        (integer_type_id, 'valid_integer', NULL, 'Value must be an integer'),
        (decimal_type_id, 'required', NULL, NULL),
        (decimal_type_id, 'valid_decimal', NULL, 'Value must be a number with optional decimal part'),
        (json_type_id, 'required', NULL, NULL),
        (json_type_id, 'valid_json', NULL, 'Value must be valid JSON: {0}');

    INSERT INTO users(name, password, global_administrator, created_at, updated_at) VALUES
        ('admin', '$2a$10$rm9BWRgCVrt.rIPAEHzFKOTZHxUj93oFjlo1CwSGG7S3.dkXd8dBq', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
END $$;

-- migrate:down
TRUNCATE TABLE value_types CASCADE;

