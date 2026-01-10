-- Global users table (no workspace reference - users are global)
CREATE TABLE "user" (
  "uuid"     UUID PRIMARY KEY,
  email      VARCHAR NOT NULL UNIQUE,  -- Global unique email
  password   VARCHAR NOT NULL,         -- bcrypt hashed
  first_name VARCHAR NOT NULL,
  last_name  VARCHAR NOT NULL,
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  meta JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_user_email ON "user"(email);

-- Workspace table (replaces tenant)
CREATE TABLE workspace (
    uuid UUID PRIMARY KEY,
    slug VARCHAR(63) NOT NULL,              -- URL path segment (lowercase alphanumeric + hyphens)
    display_name VARCHAR(255) NOT NULL,     -- human-readable name
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    settings JSONB,                          -- workspace-specific settings
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_workspace_slug UNIQUE(slug),
    CONSTRAINT chk_workspace_slug CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$')
);

CREATE INDEX idx_workspace_slug ON workspace(slug);

-- Workspace membership (policy sets are assigned via user_policy_set table)
CREATE TABLE workspace_member (
    uuid UUID PRIMARY KEY,
    workspace_uuid UUID NOT NULL REFERENCES workspace(uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_workspace_member UNIQUE(workspace_uuid, user_uuid)
);

CREATE INDEX idx_workspace_member_user ON workspace_member(user_uuid);
CREATE INDEX idx_workspace_member_workspace ON workspace_member(workspace_uuid);

-- User invitations to workspaces
CREATE TABLE user_invite (
    uuid UUID PRIMARY KEY,
    workspace_uuid UUID NOT NULL REFERENCES workspace(uuid) ON DELETE CASCADE,
    email VARCHAR NOT NULL,
    policy_set_name VARCHAR(100) NOT NULL DEFAULT 'workspace_member',  -- policy set to assign on accept
    token_hash VARCHAR(64) NOT NULL,  -- SHA256 hex of invite token (for lookup)
    invited_by_user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_user_invite_email UNIQUE(email)  -- One active pending invite per email globally
);

CREATE INDEX idx_user_invite_token_hash ON user_invite(token_hash);
CREATE INDEX idx_user_invite_workspace ON user_invite(workspace_uuid);
CREATE INDEX idx_user_invite_expires ON user_invite(expires_at);

-- Password reset tokens
CREATE TABLE password_reset (
    uuid UUID PRIMARY KEY,
    user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    email VARCHAR NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_password_reset_token_hash ON password_reset(token_hash);
CREATE INDEX idx_password_reset_email ON password_reset(email);

CREATE TABLE oauth2_client (
  uuid UUID PRIMARY KEY, -- Internal unique ID for the client
  workspace_uuid UUID NOT NULL REFERENCES workspace(uuid) ON DELETE CASCADE, -- Workspace this client belongs to
  name VARCHAR NOT NULL, -- Friendly name for admin UI
  provider VARCHAR NOT NULL, -- e.g. "github", "google"
  client_id VARCHAR NOT NULL, -- OAuth2 client ID (unique per workspace)
  secret VARCHAR NOT NULL, -- OAuth2 client secret (store encrypted in production)

  created_at TIMESTAMP WITH TIME ZONE  DEFAULT NOW(), -- When the client was registered
  updated_at TIMESTAMP WITH TIME ZONE  -- Last time client config was updated
);

CREATE INDEX idx_oauth2_client_workspace ON oauth2_client(workspace_uuid);
CREATE UNIQUE INDEX idx_oauth2_client_workspace_client_id ON oauth2_client(workspace_uuid, client_id);

-- Stores issued tokens (access + refresh) per client and optionally user
CREATE TABLE oauth2_token (
  uuid UUID PRIMARY KEY, -- Internal unique ID for token record
  client_uuid UUID NOT NULL REFERENCES oauth2_client(uuid) ON DELETE CASCADE, -- OAuth2 client that issued this token
  user_uuid UUID, -- Optional: the user this token was issued for (nullable for machine tokens)

  token JSONB, -- Raw OAuth2 token response (useful for debugging or extra metadata)
  expires_at TIMESTAMP WITH TIME ZONE, -- When the token expires
  created_at TIMESTAMP WITH TIME ZONE  DEFAULT NOW(), -- When the token was stored
  updated_at TIMESTAMP WITH TIME ZONE  -- Last refresh/update of this token
);

-- Stores OAuth2 state values to prevent CSRF during login flows
CREATE TABLE oauth2_state (
  uuid UUID PRIMARY KEY, -- Internal ID
  client_uuid UUID NOT NULL REFERENCES oauth2_client(uuid) ON DELETE CASCADE, -- Client the state is related to

  state JSONB, -- Actual state value (may include nonce, redirect_uri, etc.)
  created_at TIMESTAMP WITH TIME ZONE  DEFAULT NOW(), -- Created when auth flow started
  updated_at TIMESTAMP WITH TIME ZONE , -- Optional updates (rarely used)
  expired_at TIMESTAMP WITH TIME ZONE  NOT NULL -- When this state should be invalidated
);

-- Represents a persistent subject relationship (user ↔ client), e.g. for user linking or re-authentication
CREATE TABLE oauth2_subject (
  uuid UUID PRIMARY KEY, -- Internal ID
  user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE, -- The user in your system
  token_uuid UUID NOT NULL REFERENCES oauth2_token(uuid) ON DELETE CASCADE, -- Link to their most recent token

  created_at TIMESTAMP WITH TIME ZONE  DEFAULT NOW(), -- When the user linked this client
  updated_at TIMESTAMP WITH TIME ZONE  -- Last update to the relationship
);

CREATE TABLE datasource(
                           "uuid"            UUID PRIMARY KEY,
                           workspace_uuid UUID REFERENCES workspace(uuid) ON DELETE CASCADE,
                           user_uuid   UUID NOT NULL,
                           name              VARCHAR NOT NULL,
                           "type"            VARCHAR NOT NULL,
                           is_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
                           provider VARCHAR NOT NULL,
                           settings          JSONB,

                           created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                           updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_datasource_workspace ON datasource(workspace_uuid);

-- Pipelines table
CREATE TABLE pipeline (
                          uuid UUID PRIMARY KEY,
                          workspace_uuid UUID REFERENCES workspace(uuid) ON DELETE CASCADE,
                          datasource_uuid UUID NOT NULL,
                          storage_uuid UUID NOT NULL,
                          worker_uuid UUID,
                          name VARCHAR NOT NULL,
                          type VARCHAR NOT NULL,
                          is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
                          flow JSONB NOT NULL,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                          updated_at TIMESTAMP WITH TIME ZONE,
  CONSTRAINT fk_pipeline_datasource FOREIGN KEY("datasource_uuid") REFERENCES datasource("uuid") ON DELETE CASCADE
);

CREATE INDEX idx_pipeline_workspace ON pipeline(workspace_uuid);
CREATE INDEX idx_pipeline_worker ON pipeline(worker_uuid);


CREATE TABLE storage(
  "uuid"            UUID PRIMARY KEY,
  workspace_uuid UUID REFERENCES workspace(uuid) ON DELETE CASCADE,
  name              VARCHAR NOT NULL,
  "type"            VARCHAR NOT NULL,
  is_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
  settings          JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_storage_workspace ON storage(workspace_uuid);

CREATE TABLE "sync_policy" (
                               "uuid" UUID PRIMARY KEY,
                               pipeline_uuid   UUID NOT NULL,
                               name              VARCHAR NOT NULL,
                               "type" VARCHAR NOT NULL,
                               "blocklist" TEXT[],
                               "exclude_list" TEXT[],
                               "sync_all" BOOLEAN NOT NULL,
                               is_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
                               "settings" JSONB,
                               "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                               "updated_at" TIMESTAMP WITH TIME ZONE,
                               CONSTRAINT fk_sync_policy_pipeline FOREIGN KEY("pipeline_uuid") REFERENCES "pipeline"("uuid") ON DELETE CASCADE
);

-- 2025-03-28 @reactima
-- Add missing column to oauth2_token for "name"
ALTER TABLE oauth2_token
    ADD COLUMN IF NOT EXISTS name VARCHAR;

-- Create table "file" (FileObject)
CREATE TABLE IF NOT EXISTS "file" (
                                      uuid          UUID PRIMARY KEY,
                                      storage_type  VARCHAR NOT NULL,        -- "postgres", "s3", or "hostfiles"
                                      storage_uuid  UUID,                    -- references some row in table "storage" if you want
                                      name          VARCHAR NOT NULL,        -- e.g. "invoice.pdf" or "email.raw"
                                      mime_type     VARCHAR,                 -- e.g. "application/pdf" or "message/rfc822"
                                      size          BIGINT,
                                      data          BYTEA,                   -- if storage_type='postgres', actual file stored here
                                      path          VARCHAR,                 -- if hostfiles or s3, optional path or object key
                                      is_raw        BOOLEAN DEFAULT false,   -- indicates if this is the entire raw email
                                      raw_headers   TEXT,                   -- optional raw headers if needed
                                      has_raw_email BOOLEAN DEFAULT false,  -- indicates whether the raw email is contained
                                      is_inline     BOOLEAN DEFAULT false,  -- indicates if this is an inline/embedded file
                                      created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                      updated_at    TIMESTAMP WITH TIME ZONE
);


-- Create table "message" (Message)
CREATE TABLE IF NOT EXISTS message (
                                       uuid                       UUID PRIMARY KEY,
                                       format                     VARCHAR NOT NULL,
                                       type                       VARCHAR NOT NULL,
                                       chat_uuid                  VARCHAR,
                                       thread_uuid                VARCHAR,
                                       external_message_id TEXT,  -- original system's message ID (e.g. Gmail "messageId")
                                       sender                     VARCHAR NOT NULL,
                                       recipients                 TEXT[] NOT NULL,
                                       subject                    TEXT,
                                       body                       TEXT NOT NULL,
                                       body_parsed                JSONB,
                                       reactions                  JSONB,
                                       attachments                JSONB,
                                       forward_from               VARCHAR,
                                       reply_to_message_uuid      VARCHAR,
                                       forward_from_chat_uuid     VARCHAR,
                                       forward_from_message_uuid  VARCHAR,
                                       forward_meta               JSONB,
                                       meta                       JSONB,


                                       created_at                 TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                       updated_at                 TIMESTAMP WITH TIME ZONE
);
CREATE TABLE IF NOT EXISTS contact (
    -- Basic ID fields
                                       uuid                        text PRIMARY KEY,
                                       workspace_uuid              UUID REFERENCES workspace(uuid) ON DELETE CASCADE,
                                       user_uuid                   text,
                                       instance_uuid               text,

    -- Status and name fields
                                       status                      text,
                                       names                       jsonb,
                                       names_search                text,
                                       last                        text,
                                       first                       text,
                                       middle                      text,

    -- Birthday and salary fields
                                       birthday                    timestamptz,
                                       birthday_type               text,
                                       salary                      text,
                                       salary_data                 jsonb,

    -- Last positions
                                       last_positions              jsonb,
                                       last_position_id            integer,
                                       last_position_company_id    integer,
                                       last_position_company_name  text,
                                       last_position_title         text,
                                       last_position_start_date    timestamptz,
                                       last_position_end_date      timestamptz,
                                       last_position_end_now       boolean,
                                       last_position_description   text,

    -- Note-related
                                       note_search                 text,
                                       note_kpi_id                 jsonb,

    -- Phones
                                       phones                      jsonb,
                                       phone_search                text,
                                       phone1                      text,
                                       phone1_type                 text,
                                       phone1_country              text,
                                       phone2                      text,
                                       phone2_type                 text,
                                       phone2_country              text,
                                       phone3                      text,
                                       phone3_type                 text,
                                       phone3_country              text,
                                       phone4                      text,
                                       phone4_type                 text,
                                       phone4_country              text,
                                       phone5                      text,
                                       phone5_type                 text,
                                       phone5_country              text,

    -- Emails
                                       emails                      jsonb,
                                       email_search                text,
                                       email1                      text,
                                       email1_type                 text,
                                       email2                      text,
                                       email2_type                 text,
                                       email3                      text,
                                       email3_type                 text,
                                       email4                      text,
                                       email4_type                 text,
                                       email5                      text,
                                       email5_type                 text,

    -- Messengers
                                       messengers                  jsonb,
                                       messengers_search           text,
                                       skype_uuid                  text,
                                       skype                       text,
                                       whatsapp_uuid               text,
                                       whatsapp                    text,
                                       telegram_uuid               text,
                                       telegram                    text,
                                       wechat_uuid                 text,
                                       wechat                      text,
                                       line_uuid                   text,
                                       line                        text,

    -- Socials
                                       socials                     jsonb,
                                       socials_search              text,
                                       linkedin_uuid               text,
                                       linkedin_url                text,
                                       facebook_uuid               text,
                                       facebook_url                text,
                                       twitter_uuid                text,
                                       twitter_url                 text,
                                       github_uuid                 text,
                                       github_url                  text,
                                       vk_uuid                     text,
                                       vk_url                      text,
                                       odno_uuid                   text,
                                       odno_url                    text,
                                       hhru_uuid                   text,
                                       hhru_url                    text,
                                       habr_uuid                   text,
                                       habr_url                    text,
                                       moikrug_uuid                text,
                                       moikrug_url                 text,
                                       instagram_uuid              text,
                                       instagram_url               text,

    -- Additional socials 1..9
                                       social1_uuid                text,
                                       social1_url                 text,
                                       social1_type                text,
                                       social2_uuid                text,
                                       social2_url                 text,
                                       social2_type                text,
                                       social3_uuid                text,
                                       social3_url                 text,
                                       social3_type                text,
                                       social4_uuid                text,
                                       social4_url                 text,
                                       social4_type                text,
                                       social5_uuid                text,
                                       social5_url                 text,
                                       social5_type                text,
                                       social6_uuid                text,
                                       social6_url                 text,
                                       social6_type                text,
                                       social7_uuid                text,
                                       social7_url                 text,
                                       social7_type                text,
                                       social8_uuid                text,
                                       social8_url                 text,
                                       social8_type                text,
                                       social9_uuid                text,
                                       social9_url                 text,
                                       social9_type                text,

    -- Tracking info
                                       tracking_source             text,
                                       tracking_slug               text,

    -- Caching info
                                       cached_img                  text,
                                       cached_img_data             jsonb,
                                       crawl                       jsonb,

    -- Duplicate / housekeeping
                                       duplicate_user_id           text,
                                       duplicate_alternative_id    text,
                                       duplicate_report_date       timestamptz,

    -- Timestamps
                                       entry_date                  timestamptz,
                                       edit_date                   timestamptz,
                                       last_kpi_entry_date         timestamptz
);

CREATE INDEX idx_contact_workspace ON contact(workspace_uuid);

CREATE TABLE "scheduler" (
                             "uuid"             UUID PRIMARY KEY,
                             pipeline_uuid      UUID NOT NULL,
                             schedule_type      VARCHAR NOT NULL, -- 'cron' or 'one_time'
                             cron_expression    VARCHAR,
                             run_at             TIMESTAMP WITH TIME ZONE,
                             timezone           VARCHAR NOT NULL DEFAULT 'UTC',
                             next_run           TIMESTAMP WITH TIME ZONE,
                             last_run           TIMESTAMP WITH TIME ZONE,
                             last_uid           BIGINT NOT NULL DEFAULT 0, -- last processed IMAP UID for incremental fetch
                             is_enabled         BOOLEAN NOT NULL DEFAULT TRUE,
                             is_paused         BOOLEAN NOT NULL DEFAULT FALSE,
                             batch_size         INTEGER NOT NULL DEFAULT 100, -- number of items to process per execution

                             -- Timestamp-based sync tracking (replaces last_uid)
                             sync_state              VARCHAR(50) DEFAULT 'initial', -- 'initial', 'sync_recent', 'sync_historical', 'sync_complete'
                             last_sync_timestamp     TIMESTAMP WITH TIME ZONE,      -- most recent message timestamp synced
                             oldest_sync_timestamp   TIMESTAMP WITH TIME ZONE,      -- oldest message timestamp synced (for historical backfill)
                             cutoff_date             TIMESTAMP WITH TIME ZONE,      -- stop historical sync before this date

                             created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                             updated_at         TIMESTAMP WITH TIME ZONE,

                             CONSTRAINT fk_scheduler_pipeline FOREIGN KEY(pipeline_uuid) REFERENCES pipeline("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS worker_jobs (
                                           uuid             UUID PRIMARY KEY,
                                           scheduler_uuid      UUID,  -- NULL for ad-hoc jobs like test connections
                                           job_uuid      UUID NOT NULL,
                                           subject     VARCHAR NOT NULL,
                                           status      VARCHAR NOT NULL,            -- e.g. "running", "completed", "failed", "retry"
                                           data        JSONB DEFAULT '{}'::jsonb,-- used for error details, logs, or metadata
                                           started_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                           finished_at         TIMESTAMP WITH TIME ZONE
);

-- ============================================================================
-- Policy-Based Access Control Tables - Ory Ladon
-- ============================================================================

-- Ladon policy storage
CREATE TABLE IF NOT EXISTS ladon_policy (
    id VARCHAR(255) PRIMARY KEY,
    description TEXT,
    effect VARCHAR(10) NOT NULL CHECK (effect IN ('allow', 'deny')),
    conditions JSONB DEFAULT '{}'::jsonb,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_effect ON ladon_policy(effect);

-- Policy subjects (many-to-many)
CREATE TABLE IF NOT EXISTS ladon_policy_subject (
    id SERIAL PRIMARY KEY,
    policy_id VARCHAR(255) NOT NULL REFERENCES ladon_policy(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    CONSTRAINT uq_ladon_policy_subject UNIQUE(policy_id, subject)
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_subject_policy ON ladon_policy_subject(policy_id);
CREATE INDEX IF NOT EXISTS idx_ladon_policy_subject_subject ON ladon_policy_subject(subject);

-- Policy resources (many-to-many)
CREATE TABLE IF NOT EXISTS ladon_policy_resource (
    id SERIAL PRIMARY KEY,
    policy_id VARCHAR(255) NOT NULL REFERENCES ladon_policy(id) ON DELETE CASCADE,
    resource VARCHAR(512) NOT NULL,
    CONSTRAINT uq_ladon_policy_resource UNIQUE(policy_id, resource)
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_resource_policy ON ladon_policy_resource(policy_id);
CREATE INDEX IF NOT EXISTS idx_ladon_policy_resource_resource ON ladon_policy_resource(resource);

-- Policy actions (many-to-many)
CREATE TABLE IF NOT EXISTS ladon_policy_action (
    id SERIAL PRIMARY KEY,
    policy_id VARCHAR(255) NOT NULL REFERENCES ladon_policy(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    CONSTRAINT uq_ladon_policy_action UNIQUE(policy_id, action)
);

CREATE INDEX IF NOT EXISTS idx_ladon_policy_action_policy ON ladon_policy_action(policy_id);
CREATE INDEX IF NOT EXISTS idx_ladon_policy_action_action ON ladon_policy_action(action);

-- User policy set assignments (user-to-policy-set mappings with optional workspace scope)
CREATE TABLE IF NOT EXISTS user_policy_set (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    policy_set_name VARCHAR(100) NOT NULL,
    workspace_slug VARCHAR(63),  -- NULL for global policy sets
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_user_policy_set UNIQUE(user_uuid, policy_set_name, workspace_slug)
);

CREATE INDEX IF NOT EXISTS idx_user_policy_set_user ON user_policy_set(user_uuid);
CREATE INDEX IF NOT EXISTS idx_user_policy_set_policy_set ON user_policy_set(policy_set_name);
CREATE INDEX IF NOT EXISTS idx_user_policy_set_workspace ON user_policy_set(workspace_slug);

-- Policy set definitions for API management
CREATE TABLE IF NOT EXISTS policy_set (
    uuid UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    scope VARCHAR(50) NOT NULL DEFAULT 'workspace', -- 'global' or 'workspace'
    is_system BOOLEAN NOT NULL DEFAULT FALSE, -- System policy sets cannot be deleted
    permissions JSONB DEFAULT '[]'::jsonb, -- Cached list of permissions for display
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT chk_policy_set_scope CHECK (scope IN ('global', 'workspace'))
);

CREATE INDEX IF NOT EXISTS idx_policy_set_scope ON policy_set(scope);
CREATE INDEX IF NOT EXISTS idx_policy_set_name ON policy_set(name);

-- Permission definitions (for documentation/UI)
CREATE TABLE IF NOT EXISTS permission (
    uuid UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE, -- e.g., "datasource:read", "workspace:admin"
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL, -- e.g., "datasource", "pipeline", "workspace"
    action VARCHAR(50) NOT NULL, -- e.g., "read", "write", "delete", "admin"
    scope VARCHAR(50) NOT NULL DEFAULT 'workspace', -- 'global' or 'workspace'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_permission_scope CHECK (scope IN ('global', 'workspace'))
);

CREATE INDEX IF NOT EXISTS idx_permission_resource ON permission(resource);
CREATE INDEX IF NOT EXISTS idx_permission_scope ON permission(scope);
CREATE INDEX IF NOT EXISTS idx_permission_name ON permission(name);

-- ============================================================================
-- Distributed Worker Tables
-- ============================================================================

-- Worker registry - stores registered workers that connect via gRPC
CREATE TABLE IF NOT EXISTS registered_worker (
    uuid UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    secret_hash VARCHAR(255) NOT NULL,  -- bcrypt hashed secret for authentication
    status VARCHAR(50) NOT NULL DEFAULT 'offline',  -- online, offline, draining
    is_global BOOLEAN NOT NULL DEFAULT FALSE,  -- can process any workspace
    version VARCHAR(100),  -- worker binary version
    labels JSONB DEFAULT '{}'::jsonb,  -- metadata labels
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    last_connected_at TIMESTAMP WITH TIME ZONE,
    connected_from VARCHAR(255),  -- IP address
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT chk_registered_worker_status CHECK (status IN ('online', 'offline', 'draining'))
);

CREATE INDEX IF NOT EXISTS idx_registered_worker_status ON registered_worker(status);

-- Worker-to-workspace assignments (many-to-many)
CREATE TABLE IF NOT EXISTS worker_workspace (
    uuid UUID PRIMARY KEY,
    worker_uuid UUID NOT NULL REFERENCES registered_worker(uuid) ON DELETE CASCADE,
    workspace_uuid UUID NOT NULL REFERENCES workspace(uuid) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_worker_workspace UNIQUE(worker_uuid, workspace_uuid)
);

CREATE INDEX IF NOT EXISTS idx_worker_workspace_worker ON worker_workspace(worker_uuid);
CREATE INDEX IF NOT EXISTS idx_worker_workspace_workspace ON worker_workspace(workspace_uuid);

-- One-time enrollment tokens for worker registration
CREATE TABLE IF NOT EXISTS worker_enrollment_token (
    uuid UUID PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL UNIQUE,  -- bcrypt hashed token
    name VARCHAR(255) NOT NULL,  -- pre-assigned worker name
    is_global BOOLEAN NOT NULL DEFAULT FALSE,
    workspace_uuids UUID[] DEFAULT '{}',  -- workspaces to assign on enrollment
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    used_by_worker_uuid UUID REFERENCES registered_worker(uuid) ON DELETE SET NULL,
    created_by_user_uuid UUID REFERENCES "user"(uuid),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_worker_enrollment_token_expires ON worker_enrollment_token(expires_at);

-- FK for pipeline.worker_uuid (added after registered_worker table is created)
ALTER TABLE pipeline ADD CONSTRAINT fk_pipeline_worker
    FOREIGN KEY (worker_uuid) REFERENCES registered_worker(uuid) ON DELETE SET NULL;

-- ============================================================================
-- Usage Limits
-- ============================================================================

-- User limits defined at policy set level (defaults)
CREATE TABLE IF NOT EXISTS usage_limit (
    uuid UUID PRIMARY KEY,
    policy_set_name VARCHAR(100) NOT NULL REFERENCES policy_set(name) ON DELETE CASCADE,
    limit_type VARCHAR(50) NOT NULL,  -- 'messages_fetch', 'messages_push'
    limit_value BIGINT,  -- NULL = unlimited
    reset_period VARCHAR(20) NOT NULL DEFAULT 'monthly',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_usage_limit_policy_type UNIQUE(policy_set_name, limit_type),
    CONSTRAINT chk_usage_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push')),
    CONSTRAINT chk_usage_limit_period CHECK (reset_period IN ('daily', 'weekly', 'monthly', 'rolling_24h', 'rolling_7d', 'rolling_30d'))
);

CREATE INDEX IF NOT EXISTS idx_usage_limit_policy_set ON usage_limit(policy_set_name);

-- Per-user limit overrides, scoped to workspace
CREATE TABLE IF NOT EXISTS user_usage_limit_override (
    uuid UUID PRIMARY KEY,
    user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    workspace_slug VARCHAR(63) NOT NULL,
    limit_type VARCHAR(50) NOT NULL,
    limit_value BIGINT,  -- NULL = unlimited
    reset_period VARCHAR(20),  -- NULL = inherit from policy set
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_user_usage_limit UNIQUE(user_uuid, workspace_slug, limit_type),
    CONSTRAINT chk_user_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push')),
    CONSTRAINT chk_user_limit_period CHECK (reset_period IS NULL OR reset_period IN ('daily', 'weekly', 'monthly', 'rolling_24h', 'rolling_7d', 'rolling_30d'))
);

CREATE INDEX IF NOT EXISTS idx_user_usage_limit_user ON user_usage_limit_override(user_uuid);
CREATE INDEX IF NOT EXISTS idx_user_usage_limit_workspace ON user_usage_limit_override(workspace_slug);

-- Worker limits, scoped to workspace
CREATE TABLE IF NOT EXISTS worker_usage_limit (
    uuid UUID PRIMARY KEY,
    worker_uuid UUID NOT NULL REFERENCES registered_worker(uuid) ON DELETE CASCADE,
    workspace_slug VARCHAR(63) NOT NULL,
    limit_type VARCHAR(50) NOT NULL,
    limit_value BIGINT,  -- NULL = unlimited
    reset_period VARCHAR(20) NOT NULL DEFAULT 'monthly',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_worker_usage_limit UNIQUE(worker_uuid, workspace_slug, limit_type),
    CONSTRAINT chk_worker_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push')),
    CONSTRAINT chk_worker_limit_period CHECK (reset_period IN ('daily', 'weekly', 'monthly', 'rolling_24h', 'rolling_7d', 'rolling_30d'))
);

CREATE INDEX IF NOT EXISTS idx_worker_usage_limit_worker ON worker_usage_limit(worker_uuid);
CREATE INDEX IF NOT EXISTS idx_worker_usage_limit_workspace ON worker_usage_limit(workspace_slug);

-- Tracks user usage per workspace
CREATE TABLE IF NOT EXISTS user_usage_tracking (
    uuid UUID PRIMARY KEY,
    user_uuid UUID NOT NULL REFERENCES "user"(uuid) ON DELETE CASCADE,
    workspace_slug VARCHAR(63) NOT NULL,
    limit_type VARCHAR(50) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    current_usage BIGINT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_user_usage_tracking UNIQUE(user_uuid, workspace_slug, limit_type, period_start),
    CONSTRAINT chk_user_tracking_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push'))
);

CREATE INDEX IF NOT EXISTS idx_user_usage_tracking_user ON user_usage_tracking(user_uuid);
CREATE INDEX IF NOT EXISTS idx_user_usage_tracking_workspace ON user_usage_tracking(workspace_slug);
CREATE INDEX IF NOT EXISTS idx_user_usage_tracking_lookup ON user_usage_tracking(user_uuid, workspace_slug, limit_type, period_start);

-- Tracks worker usage per workspace
CREATE TABLE IF NOT EXISTS worker_usage_tracking (
    uuid UUID PRIMARY KEY,
    worker_uuid UUID NOT NULL REFERENCES registered_worker(uuid) ON DELETE CASCADE,
    workspace_slug VARCHAR(63) NOT NULL,
    limit_type VARCHAR(50) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    current_usage BIGINT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_worker_usage_tracking UNIQUE(worker_uuid, workspace_slug, limit_type, period_start),
    CONSTRAINT chk_worker_tracking_limit_type CHECK (limit_type IN ('messages_fetch', 'messages_push'))
);

CREATE INDEX IF NOT EXISTS idx_worker_usage_tracking_worker ON worker_usage_tracking(worker_uuid);
CREATE INDEX IF NOT EXISTS idx_worker_usage_tracking_workspace ON worker_usage_tracking(workspace_slug);
CREATE INDEX IF NOT EXISTS idx_worker_usage_tracking_lookup ON worker_usage_tracking(worker_uuid, workspace_slug, limit_type, period_start);

