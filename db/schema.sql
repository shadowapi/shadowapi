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
  name              VARCHAR NOT NULL,
  "type"            VARCHAR NOT NULL,
  is_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
  settings          JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE
);
