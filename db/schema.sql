CREATE TABLE "user" (
  "uuid"     UUID PRIMARY KEY,
  email      VARCHAR NOT NULL,
  password   VARCHAR NOT NULL,
  first_name VARCHAR NOT NULL,
  last_name  VARCHAR NOT NULL,
  is_enabled BOOLEAN NOT NULL,
  is_admin   BOOLEAN NOT NULL DEFAULT FALSE,
  meta JSONB,

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
  user_uuid    UUID,
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

-- 2025-03-28 @reactima
-- Add missing column to oauth2_token for "name"
ALTER TABLE oauth2_token
    ADD COLUMN IF NOT EXISTS name VARCHAR;

-- Create table "file" (FileObject)
CREATE TABLE IF NOT EXISTS "file" (
                                      uuid         UUID PRIMARY KEY,
                                      storage_type VARCHAR NOT NULL,
                                      storage_uuid  VARCHAR NOT NULL,
                                      name         VARCHAR NOT NULL,
                                      mime_type    VARCHAR,
                                      size         BIGINT,
                                      created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                      updated_at   TIMESTAMP WITH TIME ZONE
);

-- Create table "message" (Message)
CREATE TABLE IF NOT EXISTS message (
                                       uuid                       UUID PRIMARY KEY,
                                       source                     VARCHAR NOT NULL,
                                       type                       VARCHAR NOT NULL,
                                       chat_uuid                  VARCHAR,
                                       thread_uuid                VARCHAR,
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
