BEGIN;

ALTER TABLE platforms ADD COLUMN IF NOT EXISTS old_username varchar(255) NOT NULL DEFAULT '';
ALTER TABLE platforms ADD COLUMN IF NOT EXISTS old_password bytea NOT NULL DEFAULT '';
ALTER TABLE platforms ADD COLUMN IF NOT EXISTS credentials_active boolean NOT NULL DEFAULT '1';

COMMIT;