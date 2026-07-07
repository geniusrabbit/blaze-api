-- account.up.sql
-- Adds common account traits/metadata table for the blaze-api account domain.
-- Run with: migrate -path deploy/migrations -database $DATABASE_URL up

CREATE TABLE IF NOT EXISTS account_traits (
    id           BIGSERIAL    PRIMARY KEY,
    account_id   BIGINT       NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    value        TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, name)
);

CREATE INDEX IF NOT EXISTS idx_account_traits_account_id ON account_traits (account_id);
