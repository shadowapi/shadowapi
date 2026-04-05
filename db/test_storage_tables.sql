-- ============================================================
-- ShadowAPI: Test Storage Tables
--
-- Test-prefixed versions of the 4 core CRM tables that
-- ShadowAPI will write to. Used for dry-run testing before
-- CRM integration.
--
-- Based on: reactima-crm/z/attio_email_calendar/
--           email_messenger_calendar_with_shadowapi.sql
--
-- Usage:
--   make test-init-test-tables    — create tables + register storage
--   make test-destroy-test-tables — drop tables + remove storage
-- ============================================================


-- ------------------------------------------------------------
-- 1. test_messages (unified message store — all channels)
--
-- CRM equivalent: hh_outbound_messages
-- Stores email, LinkedIn, WhatsApp, Telegram messages.
-- Both inbound (synced) and outbound (composed/sent).
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS test_messages (
    id              SERIAL PRIMARY KEY,
    channel         text DEFAULT 'email',       -- 'email' | 'linkedin' | 'whatsapp' | 'telegram' | 'sms'
    provider_id     text NOT NULL,              -- Gmail msg ID / Graph ID / LinkedIn URN / WA ID / TG ID
    thread_id       text,                       -- provider thread / conversation / chat ID
    subject         text,                       -- email only (NULL for messengers)
    body_preview    text,                       -- first 500 chars (email) or full text (messengers)
    body_s3_key     text,                       -- S3 key for HTML body (email only)
    sent_at         timestamp with time zone,
    received_at     timestamp with time zone,
    direction       text NOT NULL,              -- 'inbound' | 'outbound'
    source          text DEFAULT 'sync',        -- sync | composed | forwarded | sequence
    is_draft        boolean DEFAULT false,
    options         jsonb,                      -- AI summary, labels, provider metadata
    created_at      timestamp with time zone DEFAULT now(),
    updated_at      timestamp with time zone DEFAULT now(),

    CONSTRAINT test_messages_provider_unique UNIQUE (channel, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_test_messages_thread ON test_messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_test_messages_sent ON test_messages(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_messages_channel ON test_messages(channel);


-- ------------------------------------------------------------
-- 2. test_message_participants (To/CC/BCC/From)
--
-- CRM equivalent: hh_outbound_message_participants
-- Separate table because we need index on address for
-- record matching (email → contact).
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS test_message_participants (
    id              SERIAL PRIMARY KEY,
    message_id      integer NOT NULL REFERENCES test_messages(id) ON DELETE CASCADE,
    role            text NOT NULL,              -- 'from' | 'to' | 'cc' | 'bcc'
    address         text NOT NULL,              -- email address, phone, or username
    name            text
);

CREATE INDEX IF NOT EXISTS idx_test_msg_part_addr ON test_message_participants(address);
CREATE INDEX IF NOT EXISTS idx_test_msg_part_msg ON test_message_participants(message_id);


-- ------------------------------------------------------------
-- 3. test_calendar_events
--
-- CRM equivalent: hh_calendar_events
-- Synced from Google Calendar / Outlook Calendar via ShadowAPI.
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS test_calendar_events (
    id              SERIAL PRIMARY KEY,
    provider_id     text NOT NULL,
    title           text,
    description     text,
    location        text,
    start_at        timestamp with time zone NOT NULL,
    end_at          timestamp with time zone NOT NULL,
    is_all_day      boolean DEFAULT false,
    video_link      text,
    is_private      boolean DEFAULT false,
    options         jsonb,
    created_at      timestamp with time zone DEFAULT now(),
    updated_at      timestamp with time zone DEFAULT now(),

    CONSTRAINT test_calendar_events_provider_unique UNIQUE (provider_id)
);

CREATE INDEX IF NOT EXISTS idx_test_calendar_events_start ON test_calendar_events(start_at DESC);


-- ------------------------------------------------------------
-- 4. test_calendar_event_participants
--
-- CRM equivalent: hh_calendar_event_participants
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS test_calendar_event_participants (
    id              SERIAL PRIMARY KEY,
    event_id        integer NOT NULL REFERENCES test_calendar_events(id) ON DELETE CASCADE,
    address         text NOT NULL,
    name            text,
    rsvp            text                        -- accepted | declined | tentative | pending
);
