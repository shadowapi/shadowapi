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
