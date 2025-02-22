CREATE TABLE "user" (
  "uuid"     UUID PRIMARY KEY,
  email      VARCHAR NOT NULL,
  password   VARCHAR NOT NULL,
  first_name VARCHAR NOT NULL,
  last_name  VARCHAR NOT NULL,
  is_enabled BOOLEAN NOT NULL,
  is_admin   BOOLEAN NOT NULL DEFAULT FALSE,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE,

  CONSTRAINT uq_users_email unique(email)
);

CREATE TABLE "oauth2_client"(
  id         VARCHAR PRIMARY KEY,
  name       VARCHAR NOT NULL,
  provider   VARCHAR NOT NULL,
  secret     VARCHAR NOT NULL,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE "oauth2_token"(
  "uuid"          UUID PRIMARY KEY,
  client_id       VARCHAR NOT NULL,
  token           JSONB,

  created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at  TIMESTAMP WITH TIME ZONE
);

CREATE TABLE "oauth2_state"(
  "uuid"      UUID PRIMARY KEY,
  client_name VARCHAR NOT NULL,
  client_id   VARCHAR NOT NULL,
  state       JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE,
  expired_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE "oauth2_subject"(
  "uuid"      UUID PRIMARY KEY,
  user_uuid   UUID NOT NULL,
  client_name VARCHAR NOT NULL,
  client_id   VARCHAR NOT NULL,
  token       JSONB,

  created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at  TIMESTAMP WITH TIME ZONE,
  expired_at  TIMESTAMP WITH TIME ZONE NOT NULL,

  CONSTRAINT fk_oauth2_subject_user_uuid
  FOREIGN KEY(user_uuid) REFERENCES "user"("uuid") ON DELETE CASCADE
);

CREATE TABLE datasource(
  "uuid"            UUID PRIMARY KEY,
  user_uuid         UUID NOT NULL,
  name              VARCHAR NOT NULL,
  "type"            VARCHAR NOT NULL,
  is_enabled        BOOLEAN NOT NULL,
  oauth2_client_id  VARCHAR,
  oauth2_token_uuid UUID,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE,

  CONSTRAINT fk_datasource_oauth2_client_id
  FOREIGN KEY(oauth2_client_id) REFERENCES oauth2_client("id")
  ON DELETE SET NULL,

  CONSTRAINT fk_datasource_oauth2_token_id
  FOREIGN KEY(oauth2_token_uuid) REFERENCES oauth2_token("uuid")
  ON DELETE SET NULL
);

CREATE TABLE datasource_email(
  "uuid"           UUID PRIMARY KEY,
  datasource_uuid  UUID NOT NULL,
  email            VARCHAR NOT NULL,
  password         VARCHAR,
  imap_server      VARCHAR,
  smtp_server      VARCHAR,
  smtp_tls         BOOLEAN,
  provider         VARCHAR NOT NULL DEFAULT 'imap',

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE,

  CONSTRAINT fk_datasource_email_datasource_uuid
  FOREIGN KEY(datasource_uuid) REFERENCES datasource("uuid") ON DELETE CASCADE
);

CREATE TABLE pipeline(
  "uuid"            UUID PRIMARY KEY,
  name              VARCHAR NOT NULL,
  flow              JSONB NOT NULL,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE pipeline_entry(
  "uuid"            UUID PRIMARY KEY,
  pipeline_uuid     UUID NOT NULL,
  parent_uuid       UUID,
  "type"            TEXT NOT NULL,
  params            JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE,

  CONSTRAINT fk_pipeline_entry_pipeline_uuid
  FOREIGN KEY(pipeline_uuid) REFERENCES pipeline("uuid") ON DELETE CASCADE
);

CREATE TABLE storage(
  "uuid"            UUID PRIMARY KEY,
  user_uuid         UUID NOT NULL,
  name              VARCHAR NOT NULL,
  "type"            VARCHAR NOT NULL,
  is_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
  settings          JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE,

  CONSTRAINT fk_storage_user_uuid
  FOREIGN KEY(user_uuid) REFERENCES "user"("uuid") ON DELETE CASCADE
);
-- @reactima added tg_ prefix
-- tg_accounts table
CREATE TABLE IF NOT EXISTS tg_accounts
(
    id       BIGSERIAL PRIMARY KEY,
    username VARCHAR(32) UNIQUE -- @reactima Added UNIQUE constraint for username
);

-- tg_sessions table, stores session states for phone numbers
CREATE TABLE IF NOT EXISTS tg_sessions
(
    id            BIGSERIAL PRIMARY KEY,
    phone         VARCHAR(16) NOT NULL,
    account_id    BIGINT         NOT NULL REFERENCES tg_accounts (id) ON DELETE CASCADE,
    session       JSON,
    contacts_hash BIGINT,
    description   VARCHAR(256),

    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE,

    CONSTRAINT created_at_non_negative CHECK (created_at >= 0),
    CONSTRAINT updated_at_non_negative CHECK (updated_at >= 0)
);

-- tg_sessions_states table, stores state for each session
CREATE TABLE IF NOT EXISTS tg_sessions_states
(
    id   BIGINT NOT NULL,
    pts  BIGINT NOT NULL DEFAULT 0,
    qts  BIGINT NOT NULL DEFAULT 0,
    date BIGINT NOT NULL DEFAULT 0,
    seq  BIGINT NOT NULL DEFAULT 0,
    FOREIGN KEY (id) REFERENCES tg_sessions (id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS tg_peers
(
    id            BIGINT      NOT NULL,
    fk_session_id BIGINT         NOT NULL REFERENCES tg_sessions (id) ON DELETE CASCADE,
    peer_type     VARCHAR(32) NOT NULL,
    access_hash   BIGINT,
    PRIMARY KEY (id, fk_session_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tg_peers_type_peer_id ON tg_peers (id, peer_type, fk_session_id);

CREATE TABLE IF NOT EXISTS tg_peers_users
(
    id            BIGINT NOT NULL,
    fk_session_id BIGINT    NOT NULL,
    phone         VARCHAR(16),
    FOREIGN KEY (id, fk_session_id) REFERENCES tg_peers (id, fk_session_id) ON DELETE CASCADE,
    PRIMARY KEY (id, fk_session_id)
);

CREATE INDEX IF NOT EXISTS idx_tg_peers_phone ON tg_peers_users (phone);

CREATE TABLE IF NOT EXISTS tg_peers_channels
(
    id            BIGINT NOT NULL,
    fk_session_id BIGINT    NOT NULL,
    pts           BIGINT    NOT NULL DEFAULT 0,
    FOREIGN KEY (id, fk_session_id) REFERENCES tg_peers (id, fk_session_id) ON DELETE CASCADE,
    PRIMARY KEY (id, fk_session_id)
);

CREATE TABLE IF NOT EXISTS tg_cached_users
(
    id            BIGINT,
    first_name    VARCHAR(64),
    last_name     VARCHAR(64),
    username      VARCHAR(32),
    phone         VARCHAR(16),
    raw           BYTEA,
    raw_full      BYTEA,
    fk_session_id BIGINT NOT NULL,
    FOREIGN KEY (fk_session_id) REFERENCES tg_sessions (id) ON DELETE CASCADE,
    PRIMARY KEY (fk_session_id, id)
);

CREATE TABLE IF NOT EXISTS tg_cached_channels
(
    id            BIGINT,
    title         VARCHAR(128),
    username      VARCHAR(32),
    broadcast     BOOLEAN,
    forum         BOOLEAN,
    megagroup     BOOLEAN,
    raw           BYTEA,
    raw_full      BYTEA,
    fk_session_id BIGINT NOT NULL,
    FOREIGN KEY (fk_session_id) REFERENCES tg_sessions (id) ON DELETE CASCADE,
    PRIMARY KEY (fk_session_id, id)
);

CREATE TABLE IF NOT EXISTS tg_cached_chats
(
    id            BIGINT,
    title         VARCHAR(128),
    raw           BYTEA,
    raw_full      BYTEA,
    fk_session_id BIGINT NOT NULL,
    FOREIGN KEY (fk_session_id) REFERENCES tg_sessions (id) ON DELETE CASCADE,
    PRIMARY KEY (fk_session_id, id)
);
