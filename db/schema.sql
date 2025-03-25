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
  provider VARCHAR NOT NULL,
  settings          JSONB,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE
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

CREATE TABLE "sync_policy" (
                               "uuid" UUID PRIMARY KEY,
                               "user_id" UUID NOT NULL,
                               "service" VARCHAR NOT NULL,
                               "blocklist" TEXT[],
                               "exclude_list" TEXT[],
                               "sync_all" BOOLEAN NOT NULL,
                               "settings" JSONB,
                               "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                               "updated_at" TIMESTAMP WITH TIME ZONE,
                               CONSTRAINT fk_sync_policy_user FOREIGN KEY("user_id") REFERENCES "user"("uuid") ON DELETE CASCADE
);
