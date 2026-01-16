-- Global users table (no workspace reference - users are global)
CREATE TABLE "user" (
    "uuid" uuid PRIMARY KEY,
    email varchar NOT NULL UNIQUE, -- Global unique email
    password VARCHAR NOT NULL, -- bcrypt hashed
    first_name varchar NOT NULL,
    last_name varchar NOT NULL,
    is_enabled boolean NOT NULL DEFAULT TRUE,
    meta jsonb,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone
);

CREATE INDEX idx_user_email ON "user" (email);

-- Workspace table (replaces tenant)
CREATE TABLE workspace (
    uuid uuid PRIMARY KEY,
    slug varchar(63) NOT NULL, -- URL path segment (lowercase alphanumeric + hyphens)
    display_name varchar(255) NOT NULL, -- human-readable name
    is_enabled boolean NOT NULL DEFAULT TRUE,
    settings jsonb, -- workspace-specific settings
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT uq_workspace_slug UNIQUE (slug),
    CONSTRAINT chk_workspace_slug CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$')
);

CREATE INDEX idx_workspace_slug ON workspace (slug);

-- Workspace membership (policy sets are assigned via user_policy_set table)
CREATE TABLE workspace_member (
    uuid uuid PRIMARY KEY,
    workspace_uuid uuid NOT NULL REFERENCES workspace (uuid) ON DELETE CASCADE,
    user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT uq_workspace_member UNIQUE (workspace_uuid, user_uuid)
);

CREATE INDEX idx_workspace_member_user ON workspace_member (user_uuid);

CREATE INDEX idx_workspace_member_workspace ON workspace_member (workspace_uuid);

-- User invitations to workspaces
CREATE TABLE user_invite (
    uuid uuid PRIMARY KEY,
    workspace_uuid uuid NOT NULL REFERENCES workspace (uuid) ON DELETE CASCADE,
    email varchar NOT NULL,
    policy_set_name varchar(100) NOT NULL DEFAULT 'workspace_member', -- policy set to assign on accept
    token_hash varchar(64) NOT NULL, -- SHA256 hex of invite token (for lookup)
    invited_by_user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE,
    expires_at timestamp with time zone NOT NULL,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT NOW(),
    CONSTRAINT uq_user_invite_email UNIQUE (email) -- One active pending invite per email globally
);

CREATE INDEX idx_user_invite_token_hash ON user_invite (token_hash);

CREATE INDEX idx_user_invite_workspace ON user_invite (workspace_uuid);

CREATE INDEX idx_user_invite_expires ON user_invite (expires_at);

-- Password reset tokens
CREATE TABLE password_reset (
    uuid uuid PRIMARY KEY,
    user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE,
    email varchar NOT NULL,
    token_hash varchar(64) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT NOW()
);

CREATE INDEX idx_password_reset_token_hash ON password_reset (token_hash);

CREATE INDEX idx_password_reset_email ON password_reset (email);

CREATE TABLE oauth2_client (
    uuid uuid PRIMARY KEY, -- Internal unique ID for the client
    workspace_uuid uuid NOT NULL REFERENCES workspace (uuid) ON DELETE CASCADE, -- Workspace this client belongs to
    name varchar NOT NULL, -- Friendly name for admin UI
    provider varchar NOT NULL, -- e.g. "github", "google"
    client_id varchar NOT NULL, -- OAuth2 client ID (unique per workspace)
    secret varchar NOT NULL, -- OAuth2 client secret (store encrypted in production)
    created_at timestamp with time zone DEFAULT NOW(), -- When the client was registered
    updated_at timestamp with time zone -- Last time client config was updated
);

CREATE INDEX idx_oauth2_client_workspace ON oauth2_client (workspace_uuid);

CREATE UNIQUE INDEX idx_oauth2_client_workspace_client_id ON oauth2_client (workspace_uuid, client_id);

-- Stores issued tokens (access + refresh) per client and optionally user
CREATE TABLE oauth2_token (
    uuid uuid PRIMARY KEY, -- Internal unique ID for token record
    client_uuid uuid NOT NULL REFERENCES oauth2_client (uuid) ON DELETE CASCADE, -- OAuth2 client that issued this token
    user_uuid uuid, -- Optional: the user this token was issued for (nullable for machine tokens)
    token jsonb, -- Raw OAuth2 token response (useful for debugging or extra metadata)
    expires_at timestamp with time zone, -- When the token expires
    created_at timestamp with time zone DEFAULT NOW(), -- When the token was stored
    updated_at timestamp with time zone -- Last refresh/update of this token
);

-- Stores OAuth2 state values to prevent CSRF during login flows
CREATE TABLE oauth2_state (
    uuid uuid PRIMARY KEY, -- Internal ID
    client_uuid uuid NOT NULL REFERENCES oauth2_client (uuid) ON DELETE CASCADE, -- Client the state is related to
    state jsonb, -- Actual state value (may include nonce, redirect_uri, etc.)
    created_at timestamp with time zone DEFAULT NOW(), -- Created when auth flow started
    updated_at timestamp with time zone, -- Optional updates (rarely used)
    expired_at timestamp with time zone NOT NULL -- When this state should be invalidated
);

-- Represents a persistent subject relationship (user ↔ client), e.g. for user linking or re-authentication
CREATE TABLE oauth2_subject (
    uuid uuid PRIMARY KEY, -- Internal ID
    user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE, -- The user in your system
    token_uuid uuid NOT NULL REFERENCES oauth2_token (uuid) ON DELETE CASCADE, -- Link to their most recent token
    created_at timestamp with time zone DEFAULT NOW(), -- When the user linked this client
    updated_at timestamp with time zone -- Last update to the relationship
);

CREATE TABLE datasource (
    "uuid" uuid PRIMARY KEY,
    workspace_uuid uuid REFERENCES workspace (uuid) ON DELETE CASCADE,
    user_uuid uuid NOT NULL,
    name varchar NOT NULL,
    "type" varchar NOT NULL,
    is_enabled boolean NOT NULL DEFAULT TRUE,
    provider varchar NOT NULL,
    settings jsonb,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone
);

CREATE INDEX idx_datasource_workspace ON datasource (workspace_uuid);

-- Pipelines table
CREATE TABLE pipeline (
    uuid uuid PRIMARY KEY,
    workspace_uuid uuid REFERENCES workspace (uuid) ON DELETE CASCADE,
    datasource_uuid uuid NOT NULL,
    storage_uuid uuid NOT NULL,
    worker_uuid uuid,
    name varchar NOT NULL,
    type VARCHAR NOT NULL,
    is_enabled boolean NOT NULL DEFAULT FALSE,
    flow jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT fk_pipeline_datasource FOREIGN KEY ("datasource_uuid") REFERENCES datasource ("uuid") ON DELETE CASCADE
);

CREATE INDEX idx_pipeline_workspace ON pipeline (workspace_uuid);

CREATE INDEX idx_pipeline_worker ON pipeline (worker_uuid);

CREATE TABLE storage (
    "uuid" uuid PRIMARY KEY,
    workspace_uuid uuid REFERENCES workspace (uuid) ON DELETE CASCADE,
    name varchar NOT NULL,
    "type" varchar NOT NULL,
    is_enabled boolean NOT NULL DEFAULT TRUE,
    settings jsonb,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone
);

CREATE INDEX idx_storage_workspace ON storage (workspace_uuid);

CREATE TABLE "sync_policy" (
    "uuid" uuid PRIMARY KEY,
    pipeline_uuid uuid NOT NULL,
    name varchar NOT NULL,
    "type" varchar NOT NULL,
    "blocklist" text[],
    "exclude_list" text[],
    "sync_all" boolean NOT NULL,
    is_enabled boolean NOT NULL DEFAULT TRUE,
    "settings" jsonb,
    "created_at" timestamp with time zone DEFAULT NOW(),
    "updated_at" timestamp with time zone,
    CONSTRAINT fk_sync_policy_pipeline FOREIGN KEY ("pipeline_uuid") REFERENCES "pipeline" ("uuid") ON DELETE CASCADE
);

-- 2025-03-28 @reactima
-- Add missing column to oauth2_token for "name"
ALTER TABLE oauth2_token
    ADD COLUMN IF NOT EXISTS name varchar;

CREATE TABLE "scheduler" (
    "uuid" uuid PRIMARY KEY,
    pipeline_uuid uuid NOT NULL,
    schedule_type varchar NOT NULL, -- 'cron' or 'one_time'
    cron_expression varchar,
    run_at timestamp with time zone,
    timezone varchar NOT NULL DEFAULT 'UTC',
    next_run timestamp with time zone,
    last_run timestamp with time zone,
    last_uid bigint NOT NULL DEFAULT 0, -- last processed IMAP UID for incremental fetch
    is_enabled boolean NOT NULL DEFAULT TRUE,
    is_paused boolean NOT NULL DEFAULT FALSE,
    batch_size integer NOT NULL DEFAULT 100, -- number of items to process per execution
    -- Timestamp-based sync tracking (replaces last_uid)
    sync_state varchar(50) DEFAULT 'initial', -- 'initial', 'sync_recent', 'sync_historical', 'sync_complete'
    last_sync_timestamp timestamp with time zone, -- most recent message timestamp synced
    oldest_sync_timestamp timestamp with time zone, -- oldest message timestamp synced (for historical backfill)
    cutoff_date timestamp with time zone, -- stop historical sync before this date
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT fk_scheduler_pipeline FOREIGN KEY (pipeline_uuid) REFERENCES pipeline ("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS worker_jobs (
    uuid uuid PRIMARY KEY,
    scheduler_uuid uuid, -- NULL for ad-hoc jobs like test connections
    job_uuid uuid NOT NULL,
    subject varchar NOT NULL,
    status varchar NOT NULL, -- e.g. "running", "completed", "failed", "retry"
    data jsonb DEFAULT '{}'::jsonb, -- used for error details, logs, or metadata
    started_at timestamp with time zone DEFAULT NOW(),
    finished_at timestamp with time zone
);

-- ============================================================================
-- Policy-Based Access Control Tables - Ory Ladon
-- ============================================================================
-- Ladon policy storage
CREATE TABLE IF NOT EXISTS ladon_policy (
    id varchar(255) PRIMARY KEY,
    description text,
    effect varchar(10) NOT NULL CHECK (effect IN ('allow', 'deny')),
    conditions jsonb DEFAULT '{}'::jsonb,
    meta jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_effect ON ladon_policy (effect);

-- Policy subjects (many-to-many)
CREATE TABLE IF NOT EXISTS ladon_policy_subject (
    id serial PRIMARY KEY,
    policy_id varchar(255) NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    subject varchar(255) NOT NULL,
    CONSTRAINT uq_ladon_policy_subject UNIQUE (policy_id, subject)
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_subject_policy ON ladon_policy_subject (policy_id);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_subject_subject ON ladon_policy_subject (subject);

-- Policy resources (many-to-many)
CREATE TABLE IF NOT EXISTS ladon_policy_resource (
    id serial PRIMARY KEY,
    policy_id varchar(255) NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    resource varchar(512) NOT NULL,
    CONSTRAINT uq_ladon_policy_resource UNIQUE (policy_id, resource)
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_resource_policy ON ladon_policy_resource (policy_id);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_resource_resource ON ladon_policy_resource (resource);

-- Policy actions (many-to-many)
CREATE TABLE IF NOT EXISTS ladon_policy_action (
    id serial PRIMARY KEY,
    policy_id varchar(255) NOT NULL REFERENCES ladon_policy (id) ON DELETE CASCADE,
    action varchar(100) NOT NULL,
    CONSTRAINT uq_ladon_policy_action UNIQUE (policy_id, action)
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_action_policy ON ladon_policy_action (policy_id);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_action_action ON ladon_policy_action (action);

-- User policy set assignments (user-to-policy-set mappings with optional workspace scope)
CREATE TABLE IF NOT EXISTS user_policy_set (
    uuid uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE,
    policy_set_name varchar(100) NOT NULL,
    workspace_slug varchar(63), -- NULL for global policy sets
    created_at timestamp with time zone DEFAULT NOW(),
    CONSTRAINT uq_user_policy_set UNIQUE (user_uuid, policy_set_name, workspace_slug)
);

CREATE INDEX IF NOT EXISTS idx_user_policy_set_user ON user_policy_set (user_uuid);

CREATE INDEX IF NOT EXISTS idx_user_policy_set_policy_set ON user_policy_set (policy_set_name);

CREATE INDEX IF NOT EXISTS idx_user_policy_set_workspace ON user_policy_set (workspace_slug);

-- Policy set definitions for API management
CREATE TABLE IF NOT EXISTS policy_set (
    uuid uuid PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE,
    display_name varchar(255) NOT NULL,
    description text,
    scope varchar(50) NOT NULL DEFAULT 'workspace', -- 'global' or 'workspace'
    is_system boolean NOT NULL DEFAULT FALSE, -- System policy sets cannot be deleted
    permissions jsonb DEFAULT '[]'::jsonb, -- Cached list of permissions for display
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT chk_policy_set_scope CHECK (scope IN ('global', 'workspace'))
);

CREATE INDEX IF NOT EXISTS idx_policy_set_scope ON policy_set (scope);

CREATE INDEX IF NOT EXISTS idx_policy_set_name ON policy_set (name);

-- Permission definitions (for documentation/UI)
CREATE TABLE IF NOT EXISTS permission (
    uuid uuid PRIMARY KEY,
    name varchar(100) NOT NULL UNIQUE, -- e.g., "datasource:read", "workspace:admin"
    display_name varchar(255) NOT NULL,
    description text,
    resource varchar(100) NOT NULL, -- e.g., "datasource", "pipeline", "workspace"
    action varchar(50) NOT NULL, -- e.g., "read", "write", "delete", "admin"
    scope varchar(50) NOT NULL DEFAULT 'workspace', -- 'global' or 'workspace'
    created_at timestamp with time zone DEFAULT NOW(),
    CONSTRAINT chk_permission_scope CHECK (scope IN ('global', 'workspace'))
);

CREATE INDEX IF NOT EXISTS idx_permission_resource ON permission (resource);

CREATE INDEX IF NOT EXISTS idx_permission_scope ON permission (scope);

CREATE INDEX IF NOT EXISTS idx_permission_name ON permission (name);

-- ============================================================================
-- Distributed Worker Tables
-- ============================================================================
-- Worker registry - stores registered workers that connect via gRPC
CREATE TABLE IF NOT EXISTS registered_worker (
    uuid uuid PRIMARY KEY,
    name varchar(255) NOT NULL,
    secret_hash varchar(255) NOT NULL, -- bcrypt hashed secret for authentication
    status varchar(50) NOT NULL DEFAULT 'offline', -- online, offline, draining
    is_global boolean NOT NULL DEFAULT FALSE, -- can process any workspace
    version varchar(100), -- worker binary version
    labels jsonb DEFAULT '{}'::jsonb, -- metadata labels
    last_heartbeat timestamp with time zone,
    last_connected_at timestamp with time zone,
    connected_from varchar(255), -- IP address
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT chk_registered_worker_status CHECK (status IN ('online', 'offline', 'draining'))
);

CREATE INDEX IF NOT EXISTS idx_registered_worker_status ON registered_worker (status);

-- Worker-to-workspace assignments (many-to-many)
CREATE TABLE IF NOT EXISTS worker_workspace (
    uuid uuid PRIMARY KEY,
    worker_uuid uuid NOT NULL REFERENCES registered_worker (uuid) ON DELETE CASCADE,
    workspace_uuid uuid NOT NULL REFERENCES workspace (uuid) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT NOW(),
    CONSTRAINT uq_worker_workspace UNIQUE (worker_uuid, workspace_uuid)
);

CREATE INDEX IF NOT EXISTS idx_worker_workspace_worker ON worker_workspace (worker_uuid);

CREATE INDEX IF NOT EXISTS idx_worker_workspace_workspace ON worker_workspace (workspace_uuid);

-- One-time enrollment tokens for worker registration
CREATE TABLE IF NOT EXISTS worker_enrollment_token (
    uuid uuid PRIMARY KEY,
    token_hash varchar(255) NOT NULL UNIQUE, -- bcrypt hashed token
    name varchar(255) NOT NULL, -- pre-assigned worker name
    is_global boolean NOT NULL DEFAULT FALSE,
    workspace_uuids uuid[] DEFAULT '{}', -- workspaces to assign on enrollment
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    used_by_worker_uuid uuid REFERENCES registered_worker (uuid) ON DELETE SET NULL,
    created_by_user_uuid uuid REFERENCES "user" (uuid),
    created_at timestamp with time zone DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_worker_enrollment_token_expires ON worker_enrollment_token (expires_at);

-- FK for pipeline.worker_uuid (added after registered_worker table is created)
ALTER TABLE pipeline
    ADD CONSTRAINT fk_pipeline_worker FOREIGN KEY (worker_uuid) REFERENCES registered_worker (uuid) ON DELETE SET NULL;

-- ============================================================================
-- Usage Limits
-- ============================================================================
-- User limits defined at policy set level (defaults)
CREATE TABLE IF NOT EXISTS usage_limit (
    uuid uuid PRIMARY KEY,
    policy_set_name varchar(100) NOT NULL REFERENCES policy_set (name) ON DELETE CASCADE,
    limit_type varchar(50) NOT NULL, -- 'messages_fetch', 'messages_push'
    limit_value bigint, -- NULL = unlimited
    reset_period varchar(20) NOT NULL DEFAULT 'monthly',
    is_enabled boolean NOT NULL DEFAULT TRUE,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT uq_usage_limit_policy_type UNIQUE (policy_set_name, limit_type),
    CONSTRAINT chk_usage_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push')),
    CONSTRAINT chk_usage_limit_period CHECK (reset_period IN ('daily', 'weekly', 'monthly', 'rolling_24h', 'rolling_7d', 'rolling_30d'))
);

CREATE INDEX IF NOT EXISTS idx_usage_limit_policy_set ON usage_limit (policy_set_name);

-- Per-user limit overrides, scoped to workspace
CREATE TABLE IF NOT EXISTS user_usage_limit_override (
    uuid uuid PRIMARY KEY,
    user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE,
    workspace_slug varchar(63) NOT NULL,
    limit_type varchar(50) NOT NULL,
    limit_value bigint, -- NULL = unlimited
    reset_period varchar(20), -- NULL = inherit from policy set
    is_enabled boolean NOT NULL DEFAULT TRUE,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT uq_user_usage_limit UNIQUE (user_uuid, workspace_slug, limit_type),
    CONSTRAINT chk_user_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push')),
    CONSTRAINT chk_user_limit_period CHECK (reset_period IS NULL OR reset_period IN ('daily', 'weekly', 'monthly', 'rolling_24h', 'rolling_7d', 'rolling_30d'))
);

CREATE INDEX IF NOT EXISTS idx_user_usage_limit_user ON user_usage_limit_override (user_uuid);

CREATE INDEX IF NOT EXISTS idx_user_usage_limit_workspace ON user_usage_limit_override (workspace_slug);

-- Worker limits, scoped to workspace
CREATE TABLE IF NOT EXISTS worker_usage_limit (
    uuid uuid PRIMARY KEY,
    worker_uuid uuid NOT NULL REFERENCES registered_worker (uuid) ON DELETE CASCADE,
    workspace_slug varchar(63) NOT NULL,
    limit_type varchar(50) NOT NULL,
    limit_value bigint, -- NULL = unlimited
    reset_period varchar(20) NOT NULL DEFAULT 'monthly',
    is_enabled boolean NOT NULL DEFAULT TRUE,
    created_at timestamp with time zone DEFAULT NOW(),
    updated_at timestamp with time zone,
    CONSTRAINT uq_worker_usage_limit UNIQUE (worker_uuid, workspace_slug, limit_type),
    CONSTRAINT chk_worker_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push')),
    CONSTRAINT chk_worker_limit_period CHECK (reset_period IN ('daily', 'weekly', 'monthly', 'rolling_24h', 'rolling_7d', 'rolling_30d'))
);

CREATE INDEX IF NOT EXISTS idx_worker_usage_limit_worker ON worker_usage_limit (worker_uuid);

CREATE INDEX IF NOT EXISTS idx_worker_usage_limit_workspace ON worker_usage_limit (workspace_slug);

-- Tracks user usage per workspace
CREATE TABLE IF NOT EXISTS user_usage_tracking (
    uuid uuid PRIMARY KEY,
    user_uuid uuid NOT NULL REFERENCES "user" (uuid) ON DELETE CASCADE,
    workspace_slug varchar(63) NOT NULL,
    limit_type varchar(50) NOT NULL,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    current_usage bigint NOT NULL DEFAULT 0,
    last_updated timestamp with time zone DEFAULT NOW(),
    created_at timestamp with time zone DEFAULT NOW(),
    CONSTRAINT uq_user_usage_tracking UNIQUE (user_uuid, workspace_slug, limit_type, period_start),
    CONSTRAINT chk_user_tracking_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push'))
);

CREATE INDEX IF NOT EXISTS idx_user_usage_tracking_user ON user_usage_tracking (user_uuid);

CREATE INDEX IF NOT EXISTS idx_user_usage_tracking_workspace ON user_usage_tracking (workspace_slug);

CREATE INDEX IF NOT EXISTS idx_user_usage_tracking_lookup ON user_usage_tracking (user_uuid, workspace_slug, limit_type, period_start);

-- Tracks worker usage per workspace
CREATE TABLE IF NOT EXISTS worker_usage_tracking (
    uuid uuid PRIMARY KEY,
    worker_uuid uuid NOT NULL REFERENCES registered_worker (uuid) ON DELETE CASCADE,
    workspace_slug varchar(63) NOT NULL,
    limit_type varchar(50) NOT NULL,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    current_usage bigint NOT NULL DEFAULT 0,
    last_updated timestamp with time zone DEFAULT NOW(),
    created_at timestamp with time zone DEFAULT NOW(),
    CONSTRAINT uq_worker_usage_tracking UNIQUE (worker_uuid, workspace_slug, limit_type, period_start),
    CONSTRAINT chk_worker_tracking_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push'))
);

CREATE INDEX IF NOT EXISTS idx_worker_usage_tracking_worker ON worker_usage_tracking (worker_uuid);

CREATE INDEX IF NOT EXISTS idx_worker_usage_tracking_workspace ON worker_usage_tracking (workspace_slug);

CREATE INDEX IF NOT EXISTS idx_worker_usage_tracking_lookup ON worker_usage_tracking (worker_uuid, workspace_slug, limit_type, period_start);

