-- account.up.sql
-- Adds common account traits/metadata table for the blaze-api account domain.
-- Run with: migrate -path deploy/migrations -database $DATABASE_URL up

BEGIN;

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT '';

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS description TEXT;

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS logo_uri VARCHAR(1024) NOT NULL DEFAULT '';

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS policy_uri VARCHAR(1024) NOT NULL DEFAULT '';

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS tos_uri VARCHAR(1024) NOT NULL DEFAULT '';

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS client_uri VARCHAR(1024) NOT NULL DEFAULT '';

ALTER TABLE account_base
  ADD COLUMN IF NOT EXISTS contacts TEXT[];

COMMIT;
