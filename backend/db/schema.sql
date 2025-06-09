SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: changeset_action_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.changeset_action_type AS ENUM (
    'apply',
    'discard',
    'stash',
    'commit',
    'reopen',
    'comment'
);


--
-- Name: changeset_change_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.changeset_change_kind AS ENUM (
    'feature_version',
    'service_version',
    'feature_version_service_version',
    'key',
    'variation_value'
);


--
-- Name: changeset_change_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.changeset_change_type AS ENUM (
    'create',
    'update',
    'delete'
);


--
-- Name: changeset_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.changeset_state AS ENUM (
    'open',
    'committed',
    'applied',
    'rejected',
    'discarded',
    'stashed'
);


--
-- Name: permission_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.permission_kind AS ENUM (
    'service',
    'feature',
    'key',
    'variation'
);


--
-- Name: permission_level; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.permission_level AS ENUM (
    'editor',
    'admin'
);


--
-- Name: value_type_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.value_type_kind AS ENUM (
    'string',
    'integer',
    'decimal',
    'boolean',
    'json'
);


--
-- Name: value_validator_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.value_validator_type AS ENUM (
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


--
-- Name: valid_feature_versions_in_changeset(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.valid_feature_versions_in_changeset(in_changeset_id bigint) RETURNS TABLE(id bigint)
    LANGUAGE sql STABLE
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
$$;


--
-- Name: valid_keys_in_changeset(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.valid_keys_in_changeset(in_changeset_id bigint) RETURNS TABLE(id bigint)
    LANGUAGE sql STABLE
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
$$;


--
-- Name: valid_links_in_changeset(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.valid_links_in_changeset(in_changeset_id bigint) RETURNS TABLE(id bigint)
    LANGUAGE sql STABLE
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
$$;


--
-- Name: valid_service_versions_in_changeset(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.valid_service_versions_in_changeset(in_changeset_id bigint) RETURNS TABLE(id bigint)
    LANGUAGE sql STABLE
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
$$;


--
-- Name: valid_variation_values_in_changeset(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.valid_variation_values_in_changeset(in_changeset_id bigint) RETURNS TABLE(id bigint)
    LANGUAGE sql STABLE
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
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: changeset_actions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.changeset_actions (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    changeset_id bigint NOT NULL,
    user_id bigint NOT NULL,
    type public.changeset_action_type NOT NULL,
    comment text
);


--
-- Name: changeset_actions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.changeset_actions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: changeset_actions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.changeset_actions_id_seq OWNED BY public.changeset_actions.id;


--
-- Name: changeset_changes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.changeset_changes (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    changeset_id bigint NOT NULL,
    type public.changeset_change_type NOT NULL,
    kind public.changeset_change_kind NOT NULL,
    feature_version_id bigint,
    previous_feature_version_id bigint,
    service_version_id bigint NOT NULL,
    previous_service_version_id bigint,
    feature_version_service_version_id bigint,
    key_id bigint,
    new_variation_value_id bigint,
    old_variation_value_id bigint
);


--
-- Name: changeset_changes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.changeset_changes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: changeset_changes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.changeset_changes_id_seq OWNED BY public.changeset_changes.id;


--
-- Name: changesets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.changesets (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id bigint NOT NULL,
    state public.changeset_state NOT NULL,
    applied_at timestamp with time zone
);


--
-- Name: changesets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.changesets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: changesets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.changesets_id_seq OWNED BY public.changesets.id;


--
-- Name: feature_version_service_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feature_version_service_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    feature_version_id bigint NOT NULL,
    service_version_id bigint NOT NULL
);


--
-- Name: feature_version_service_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.feature_version_service_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: feature_version_service_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.feature_version_service_versions_id_seq OWNED BY public.feature_version_service_versions.id;


--
-- Name: feature_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feature_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    version integer NOT NULL,
    feature_id bigint NOT NULL
);


--
-- Name: feature_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.feature_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: feature_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.feature_versions_id_seq OWNED BY public.feature_versions.id;


--
-- Name: features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.features (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    service_id bigint NOT NULL
);


--
-- Name: features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.features_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: features_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.features_id_seq OWNED BY public.features.id;


--
-- Name: keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.keys (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    name text NOT NULL,
    description text,
    value_type_id bigint NOT NULL,
    feature_version_id bigint NOT NULL,
    validators_updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: keys_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: keys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.keys_id_seq OWNED BY public.keys.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    kind public.permission_kind NOT NULL,
    user_id bigint,
    user_group_id bigint,
    service_id bigint NOT NULL,
    feature_id bigint,
    key_id bigint,
    variation_context_id bigint,
    permission public.permission_level NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT permissions_check CHECK ((((user_id IS NOT NULL) AND (user_group_id IS NULL)) OR ((user_id IS NULL) AND (user_group_id IS NOT NULL))))
);


--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: service_type_variation_properties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.service_type_variation_properties (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    priority integer NOT NULL,
    service_type_id bigint NOT NULL,
    variation_property_id bigint NOT NULL
);


--
-- Name: service_type_variation_properties_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.service_type_variation_properties_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: service_type_variation_properties_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.service_type_variation_properties_id_seq OWNED BY public.service_type_variation_properties.id;


--
-- Name: service_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.service_types (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL
);


--
-- Name: service_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.service_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: service_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.service_types_id_seq OWNED BY public.service_types.id;


--
-- Name: service_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.service_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    service_id bigint NOT NULL,
    version integer NOT NULL,
    published boolean DEFAULT false NOT NULL
);


--
-- Name: service_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.service_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: service_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.service_versions_id_seq OWNED BY public.service_versions.id;


--
-- Name: services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.services (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    service_type_id bigint NOT NULL
);


--
-- Name: services_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.services_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.services_id_seq OWNED BY public.services.id;


--
-- Name: user_group_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_group_memberships (
    user_group_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: user_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_groups (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp with time zone,
    name text NOT NULL
);


--
-- Name: user_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_groups_id_seq OWNED BY public.user_groups.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    password text NOT NULL,
    global_administrator boolean DEFAULT false NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: value_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.value_types (
    id bigint NOT NULL,
    kind public.value_type_kind NOT NULL,
    name text NOT NULL
);


--
-- Name: value_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.value_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: value_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.value_types_id_seq OWNED BY public.value_types.id;


--
-- Name: value_validators; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.value_validators (
    id bigint NOT NULL,
    value_type_id bigint,
    key_id bigint,
    validator_type public.value_validator_type NOT NULL,
    parameter text,
    error_text text,
    CONSTRAINT value_validators_check CHECK ((((value_type_id IS NOT NULL) AND (key_id IS NULL)) OR ((value_type_id IS NULL) AND (key_id IS NOT NULL))))
);


--
-- Name: value_validators_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.value_validators_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: value_validators_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.value_validators_id_seq OWNED BY public.value_validators.id;


--
-- Name: variation_context_variation_property_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.variation_context_variation_property_values (
    variation_context_id bigint NOT NULL,
    variation_property_value_id bigint NOT NULL
);


--
-- Name: variation_contexts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.variation_contexts (
    id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: variation_contexts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.variation_contexts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: variation_contexts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.variation_contexts_id_seq OWNED BY public.variation_contexts.id;


--
-- Name: variation_properties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.variation_properties (
    id bigint NOT NULL,
    name text NOT NULL,
    display_name text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: variation_properties_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.variation_properties_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: variation_properties_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.variation_properties_id_seq OWNED BY public.variation_properties.id;


--
-- Name: variation_property_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.variation_property_values (
    id bigint NOT NULL,
    variation_property_id bigint NOT NULL,
    value text NOT NULL,
    parent_id bigint,
    order_index integer NOT NULL,
    archived boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: variation_property_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.variation_property_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: variation_property_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.variation_property_values_id_seq OWNED BY public.variation_property_values.id;


--
-- Name: variation_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.variation_values (
    id bigint NOT NULL,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    key_id bigint NOT NULL,
    variation_context_id bigint NOT NULL,
    data text NOT NULL
);


--
-- Name: variation_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.variation_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: variation_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.variation_values_id_seq OWNED BY public.variation_values.id;


--
-- Name: changeset_actions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_actions ALTER COLUMN id SET DEFAULT nextval('public.changeset_actions_id_seq'::regclass);


--
-- Name: changeset_changes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes ALTER COLUMN id SET DEFAULT nextval('public.changeset_changes_id_seq'::regclass);


--
-- Name: changesets id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changesets ALTER COLUMN id SET DEFAULT nextval('public.changesets_id_seq'::regclass);


--
-- Name: feature_version_service_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_version_service_versions ALTER COLUMN id SET DEFAULT nextval('public.feature_version_service_versions_id_seq'::regclass);


--
-- Name: feature_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_versions ALTER COLUMN id SET DEFAULT nextval('public.feature_versions_id_seq'::regclass);


--
-- Name: features id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features ALTER COLUMN id SET DEFAULT nextval('public.features_id_seq'::regclass);


--
-- Name: keys id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keys ALTER COLUMN id SET DEFAULT nextval('public.keys_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: service_type_variation_properties id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_type_variation_properties ALTER COLUMN id SET DEFAULT nextval('public.service_type_variation_properties_id_seq'::regclass);


--
-- Name: service_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_types ALTER COLUMN id SET DEFAULT nextval('public.service_types_id_seq'::regclass);


--
-- Name: service_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_versions ALTER COLUMN id SET DEFAULT nextval('public.service_versions_id_seq'::regclass);


--
-- Name: services id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.services ALTER COLUMN id SET DEFAULT nextval('public.services_id_seq'::regclass);


--
-- Name: user_groups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_groups ALTER COLUMN id SET DEFAULT nextval('public.user_groups_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: value_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_types ALTER COLUMN id SET DEFAULT nextval('public.value_types_id_seq'::regclass);


--
-- Name: value_validators id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_validators ALTER COLUMN id SET DEFAULT nextval('public.value_validators_id_seq'::regclass);


--
-- Name: variation_contexts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_contexts ALTER COLUMN id SET DEFAULT nextval('public.variation_contexts_id_seq'::regclass);


--
-- Name: variation_properties id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_properties ALTER COLUMN id SET DEFAULT nextval('public.variation_properties_id_seq'::regclass);


--
-- Name: variation_property_values id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_property_values ALTER COLUMN id SET DEFAULT nextval('public.variation_property_values_id_seq'::regclass);


--
-- Name: variation_values id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_values ALTER COLUMN id SET DEFAULT nextval('public.variation_values_id_seq'::regclass);


--
-- Name: changeset_actions changeset_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_actions
    ADD CONSTRAINT changeset_actions_pkey PRIMARY KEY (id);


--
-- Name: changeset_changes changeset_changes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_pkey PRIMARY KEY (id);


--
-- Name: changesets changesets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changesets
    ADD CONSTRAINT changesets_pkey PRIMARY KEY (id);


--
-- Name: feature_version_service_versions feature_version_service_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_version_service_versions
    ADD CONSTRAINT feature_version_service_versions_pkey PRIMARY KEY (id);


--
-- Name: feature_versions feature_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_versions
    ADD CONSTRAINT feature_versions_pkey PRIMARY KEY (id);


--
-- Name: features features_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features
    ADD CONSTRAINT features_name_key UNIQUE (name);


--
-- Name: features features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features
    ADD CONSTRAINT features_pkey PRIMARY KEY (id);


--
-- Name: keys keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keys
    ADD CONSTRAINT keys_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_user_id_user_group_id_service_id_feature_id_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_user_id_user_group_id_service_id_feature_id_key_key UNIQUE (user_id, user_group_id, service_id, feature_id, key_id, variation_context_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: service_type_variation_properties service_type_variation_proper_service_type_id_variation_pro_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_type_variation_properties
    ADD CONSTRAINT service_type_variation_proper_service_type_id_variation_pro_key UNIQUE (service_type_id, variation_property_id);


--
-- Name: service_type_variation_properties service_type_variation_properties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_type_variation_properties
    ADD CONSTRAINT service_type_variation_properties_pkey PRIMARY KEY (id);


--
-- Name: service_types service_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_types
    ADD CONSTRAINT service_types_name_key UNIQUE (name);


--
-- Name: service_types service_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_types
    ADD CONSTRAINT service_types_pkey PRIMARY KEY (id);


--
-- Name: service_versions service_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_versions
    ADD CONSTRAINT service_versions_pkey PRIMARY KEY (id);


--
-- Name: services services_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_name_key UNIQUE (name);


--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);


--
-- Name: user_group_memberships user_group_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_group_memberships
    ADD CONSTRAINT user_group_memberships_pkey PRIMARY KEY (user_group_id, user_id);


--
-- Name: user_groups user_groups_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_groups
    ADD CONSTRAINT user_groups_name_key UNIQUE (name);


--
-- Name: user_groups user_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_groups
    ADD CONSTRAINT user_groups_pkey PRIMARY KEY (id);


--
-- Name: users users_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_name_key UNIQUE (name);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: value_types value_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_types
    ADD CONSTRAINT value_types_name_key UNIQUE (name);


--
-- Name: value_types value_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_types
    ADD CONSTRAINT value_types_pkey PRIMARY KEY (id);


--
-- Name: value_validators value_validators_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_validators
    ADD CONSTRAINT value_validators_pkey PRIMARY KEY (id);


--
-- Name: variation_context_variation_property_values variation_context_variation_property_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_context_variation_property_values
    ADD CONSTRAINT variation_context_variation_property_values_pkey PRIMARY KEY (variation_context_id, variation_property_value_id);


--
-- Name: variation_contexts variation_contexts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_contexts
    ADD CONSTRAINT variation_contexts_pkey PRIMARY KEY (id);


--
-- Name: variation_properties variation_properties_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_properties
    ADD CONSTRAINT variation_properties_name_key UNIQUE (name);


--
-- Name: variation_properties variation_properties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_properties
    ADD CONSTRAINT variation_properties_pkey PRIMARY KEY (id);


--
-- Name: variation_property_values variation_property_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_property_values
    ADD CONSTRAINT variation_property_values_pkey PRIMARY KEY (id);


--
-- Name: variation_property_values variation_property_values_variation_property_id_value_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_property_values
    ADD CONSTRAINT variation_property_values_variation_property_id_value_key UNIQUE (variation_property_id, value);


--
-- Name: variation_values variation_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_values
    ADD CONSTRAINT variation_values_pkey PRIMARY KEY (id);


--
-- Name: idx_changeset_changes_feature_version_create; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_feature_version_create ON public.changeset_changes USING btree (changeset_id, feature_version_id) WHERE ((kind = 'feature_version'::public.changeset_change_kind) AND (type = 'create'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_feature_version_delete; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_feature_version_delete ON public.changeset_changes USING btree (changeset_id, feature_version_id) WHERE ((kind = 'feature_version'::public.changeset_change_kind) AND (type = 'delete'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_key_create; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_key_create ON public.changeset_changes USING btree (changeset_id, key_id) WHERE ((kind = 'key'::public.changeset_change_kind) AND (type = 'create'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_key_delete; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_key_delete ON public.changeset_changes USING btree (changeset_id, key_id) WHERE ((kind = 'key'::public.changeset_change_kind) AND (type = 'delete'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_link_create; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_link_create ON public.changeset_changes USING btree (changeset_id, feature_version_service_version_id) WHERE ((kind = 'feature_version_service_version'::public.changeset_change_kind) AND (type = 'create'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_link_delete; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_link_delete ON public.changeset_changes USING btree (changeset_id, feature_version_service_version_id) WHERE ((kind = 'feature_version_service_version'::public.changeset_change_kind) AND (type = 'delete'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_service_version_create; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_service_version_create ON public.changeset_changes USING btree (changeset_id, service_version_id) WHERE ((kind = 'service_version'::public.changeset_change_kind) AND (type = 'create'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_service_version_delete; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_service_version_delete ON public.changeset_changes USING btree (changeset_id, service_version_id) WHERE ((kind = 'service_version'::public.changeset_change_kind) AND (type = 'delete'::public.changeset_change_type));


--
-- Name: idx_changeset_changes_variation_value_new; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_variation_value_new ON public.changeset_changes USING btree (changeset_id, new_variation_value_id) WHERE (kind = 'variation_value'::public.changeset_change_kind);


--
-- Name: idx_changeset_changes_variation_value_old; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changeset_changes_variation_value_old ON public.changeset_changes USING btree (changeset_id, old_variation_value_id) WHERE (kind = 'variation_value'::public.changeset_change_kind);


--
-- Name: idx_changesets_applied_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_changesets_applied_at ON public.changesets USING btree (applied_at);


--
-- Name: idx_changesets_one_open_per_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_changesets_one_open_per_user ON public.changesets USING btree (user_id) WHERE (state = 'open'::public.changeset_state);


--
-- Name: idx_feature_version_service_versions_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_feature_version_service_versions_unique ON public.feature_version_service_versions USING btree (feature_version_id, service_version_id) WHERE ((valid_from IS NOT NULL) AND (valid_to IS NULL));


--
-- Name: idx_feature_versions_unique_version_per_feature; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_feature_versions_unique_version_per_feature ON public.feature_versions USING btree (feature_id, version) WHERE ((valid_from IS NOT NULL) AND (valid_to IS NULL));


--
-- Name: idx_feature_versions_valid_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_feature_versions_valid_from ON public.feature_versions USING btree (valid_from);


--
-- Name: idx_feature_versions_valid_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_feature_versions_valid_to ON public.feature_versions USING btree (valid_to);


--
-- Name: idx_fvsv_valid_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fvsv_valid_from ON public.feature_version_service_versions USING btree (valid_from);


--
-- Name: idx_fvsv_valid_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fvsv_valid_to ON public.feature_version_service_versions USING btree (valid_to);


--
-- Name: idx_keys_unique_name_per_feature_version; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_keys_unique_name_per_feature_version ON public.keys USING btree (feature_version_id, name) WHERE ((valid_from IS NOT NULL) AND (valid_to IS NULL));


--
-- Name: idx_keys_valid_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_keys_valid_from ON public.keys USING btree (valid_from);


--
-- Name: idx_keys_valid_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_keys_valid_to ON public.keys USING btree (valid_to);


--
-- Name: idx_one_value_per_key_and_context; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_one_value_per_key_and_context ON public.variation_values USING btree (key_id, variation_context_id) WHERE ((valid_from IS NOT NULL) AND (valid_to IS NULL));


--
-- Name: idx_permissions_kind; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permissions_kind ON public.permissions USING btree (kind);


--
-- Name: idx_service_versions_unique_version_per_service; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_service_versions_unique_version_per_service ON public.service_versions USING btree (service_id, version) WHERE ((valid_from IS NOT NULL) AND (valid_to IS NULL));


--
-- Name: idx_service_versions_valid_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_service_versions_valid_from ON public.service_versions USING btree (valid_from);


--
-- Name: idx_service_versions_valid_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_service_versions_valid_to ON public.service_versions USING btree (valid_to);


--
-- Name: idx_user_groups_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_groups_deleted_at ON public.user_groups USING btree (deleted_at);


--
-- Name: idx_user_groups_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_groups_name ON public.user_groups USING btree (name);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_name ON public.users USING btree (name);


--
-- Name: idx_variation_property_values_order_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_variation_property_values_order_index ON public.variation_property_values USING btree (order_index);


--
-- Name: idx_variation_values_valid_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_variation_values_valid_from ON public.variation_values USING btree (valid_from);


--
-- Name: idx_variation_values_valid_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_variation_values_valid_to ON public.variation_values USING btree (valid_to);


--
-- Name: changeset_actions changeset_actions_changeset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_actions
    ADD CONSTRAINT changeset_actions_changeset_id_fkey FOREIGN KEY (changeset_id) REFERENCES public.changesets(id);


--
-- Name: changeset_actions changeset_actions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_actions
    ADD CONSTRAINT changeset_actions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: changeset_changes changeset_changes_changeset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_changeset_id_fkey FOREIGN KEY (changeset_id) REFERENCES public.changesets(id);


--
-- Name: changeset_changes changeset_changes_feature_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_feature_version_id_fkey FOREIGN KEY (feature_version_id) REFERENCES public.feature_versions(id) ON DELETE CASCADE;


--
-- Name: changeset_changes changeset_changes_feature_version_service_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_feature_version_service_version_id_fkey FOREIGN KEY (feature_version_service_version_id) REFERENCES public.feature_version_service_versions(id) ON DELETE CASCADE;


--
-- Name: changeset_changes changeset_changes_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_key_id_fkey FOREIGN KEY (key_id) REFERENCES public.keys(id) ON DELETE CASCADE;


--
-- Name: changeset_changes changeset_changes_new_variation_value_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_new_variation_value_id_fkey FOREIGN KEY (new_variation_value_id) REFERENCES public.variation_values(id) ON DELETE CASCADE;


--
-- Name: changeset_changes changeset_changes_old_variation_value_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_old_variation_value_id_fkey FOREIGN KEY (old_variation_value_id) REFERENCES public.variation_values(id);


--
-- Name: changeset_changes changeset_changes_previous_feature_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_previous_feature_version_id_fkey FOREIGN KEY (previous_feature_version_id) REFERENCES public.feature_versions(id);


--
-- Name: changeset_changes changeset_changes_previous_service_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_previous_service_version_id_fkey FOREIGN KEY (previous_service_version_id) REFERENCES public.service_versions(id);


--
-- Name: changeset_changes changeset_changes_service_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changeset_changes
    ADD CONSTRAINT changeset_changes_service_version_id_fkey FOREIGN KEY (service_version_id) REFERENCES public.service_versions(id) ON DELETE CASCADE;


--
-- Name: changesets changesets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changesets
    ADD CONSTRAINT changesets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: feature_version_service_versions feature_version_service_versions_feature_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_version_service_versions
    ADD CONSTRAINT feature_version_service_versions_feature_version_id_fkey FOREIGN KEY (feature_version_id) REFERENCES public.feature_versions(id) ON DELETE CASCADE;


--
-- Name: feature_version_service_versions feature_version_service_versions_service_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_version_service_versions
    ADD CONSTRAINT feature_version_service_versions_service_version_id_fkey FOREIGN KEY (service_version_id) REFERENCES public.service_versions(id) ON DELETE CASCADE;


--
-- Name: feature_versions feature_versions_feature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_versions
    ADD CONSTRAINT feature_versions_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES public.features(id) ON DELETE CASCADE;


--
-- Name: features features_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features
    ADD CONSTRAINT features_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.services(id) ON DELETE CASCADE;


--
-- Name: keys keys_feature_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keys
    ADD CONSTRAINT keys_feature_version_id_fkey FOREIGN KEY (feature_version_id) REFERENCES public.feature_versions(id) ON DELETE CASCADE;


--
-- Name: keys keys_value_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keys
    ADD CONSTRAINT keys_value_type_id_fkey FOREIGN KEY (value_type_id) REFERENCES public.value_types(id);


--
-- Name: permissions permissions_feature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES public.features(id);


--
-- Name: permissions permissions_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_key_id_fkey FOREIGN KEY (key_id) REFERENCES public.keys(id);


--
-- Name: permissions permissions_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.services(id);


--
-- Name: permissions permissions_user_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_user_group_id_fkey FOREIGN KEY (user_group_id) REFERENCES public.user_groups(id);


--
-- Name: permissions permissions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: permissions permissions_variation_context_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_variation_context_id_fkey FOREIGN KEY (variation_context_id) REFERENCES public.variation_contexts(id);


--
-- Name: service_type_variation_properties service_type_variation_properties_service_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_type_variation_properties
    ADD CONSTRAINT service_type_variation_properties_service_type_id_fkey FOREIGN KEY (service_type_id) REFERENCES public.service_types(id) ON DELETE CASCADE;


--
-- Name: service_type_variation_properties service_type_variation_properties_variation_property_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_type_variation_properties
    ADD CONSTRAINT service_type_variation_properties_variation_property_id_fkey FOREIGN KEY (variation_property_id) REFERENCES public.variation_properties(id);


--
-- Name: service_versions service_versions_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.service_versions
    ADD CONSTRAINT service_versions_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.services(id);


--
-- Name: services services_service_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_service_type_id_fkey FOREIGN KEY (service_type_id) REFERENCES public.service_types(id);


--
-- Name: user_group_memberships user_group_memberships_user_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_group_memberships
    ADD CONSTRAINT user_group_memberships_user_group_id_fkey FOREIGN KEY (user_group_id) REFERENCES public.user_groups(id);


--
-- Name: user_group_memberships user_group_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_group_memberships
    ADD CONSTRAINT user_group_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: value_validators value_validators_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_validators
    ADD CONSTRAINT value_validators_key_id_fkey FOREIGN KEY (key_id) REFERENCES public.keys(id) ON DELETE CASCADE;


--
-- Name: value_validators value_validators_value_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.value_validators
    ADD CONSTRAINT value_validators_value_type_id_fkey FOREIGN KEY (value_type_id) REFERENCES public.value_types(id);


--
-- Name: variation_context_variation_property_values variation_context_variation_pr_variation_property_value_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_context_variation_property_values
    ADD CONSTRAINT variation_context_variation_pr_variation_property_value_id_fkey FOREIGN KEY (variation_property_value_id) REFERENCES public.variation_property_values(id) ON DELETE CASCADE;


--
-- Name: variation_context_variation_property_values variation_context_variation_property__variation_context_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_context_variation_property_values
    ADD CONSTRAINT variation_context_variation_property__variation_context_id_fkey FOREIGN KEY (variation_context_id) REFERENCES public.variation_contexts(id);


--
-- Name: variation_property_values variation_property_values_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_property_values
    ADD CONSTRAINT variation_property_values_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.variation_property_values(id) ON DELETE CASCADE;


--
-- Name: variation_property_values variation_property_values_variation_property_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_property_values
    ADD CONSTRAINT variation_property_values_variation_property_id_fkey FOREIGN KEY (variation_property_id) REFERENCES public.variation_properties(id) ON DELETE CASCADE;


--
-- Name: variation_values variation_values_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_values
    ADD CONSTRAINT variation_values_key_id_fkey FOREIGN KEY (key_id) REFERENCES public.keys(id) ON DELETE CASCADE;


--
-- Name: variation_values variation_values_variation_context_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variation_values
    ADD CONSTRAINT variation_values_variation_context_id_fkey FOREIGN KEY (variation_context_id) REFERENCES public.variation_contexts(id);


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('0001'),
    ('0002'),
    ('0003'),
    ('0004'),
    ('0005');
