-- Drop test tables and remove the "Test Internal" storage record.
-- Order matters: children before parents (FK constraints).

DROP TABLE IF EXISTS test_calendar_event_participants;
DROP TABLE IF EXISTS test_calendar_events;
DROP TABLE IF EXISTS test_message_participants;
DROP TABLE IF EXISTS test_messages;

DELETE FROM storage WHERE uuid = 'a0000000-0000-0000-0000-000000000001';
