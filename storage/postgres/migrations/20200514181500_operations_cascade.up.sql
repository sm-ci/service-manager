BEGIN;

ALTER TABLE operations ADD COLUMN IF NOT EXISTS root BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS cascade BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS parent varchar(100);

ALTER TYPE operation_state ADD VALUE 'pending';

COMMIT;