-- Register "Test Internal" storage pointing to the same database.
-- Uses a fixed UUID so it's deterministic and idempotent.

INSERT INTO storage (uuid, name, type, is_enabled, settings, created_at, updated_at)
VALUES (
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Test Internal',
    'postgres',
    true,
    '{"name":"Test Internal","is_same_database":true}'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (uuid) DO UPDATE SET
    name = EXCLUDED.name,
    settings = EXCLUDED.settings,
    is_enabled = true,
    updated_at = NOW();
