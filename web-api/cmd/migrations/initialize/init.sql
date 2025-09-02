BEGIN;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version text PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE schema_migrations IS 'History of applied migrations';
COMMENT ON COLUMN schema_migrations.version IS 'Unique migration version';

COMMIT;
