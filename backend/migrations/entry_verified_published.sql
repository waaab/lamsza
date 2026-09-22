-- Split admin quality flag (verified) from ownership (claimed).
-- Applied on backend start by handlers.MigrateEntryVerified (idempotent).
ALTER TABLE entries ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS published BOOLEAN NOT NULL DEFAULT true;
UPDATE entries SET verified = claimed WHERE verified = false AND claimed = true;
