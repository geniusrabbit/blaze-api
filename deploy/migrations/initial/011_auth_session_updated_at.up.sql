-- Add updated_at column to auth_session table
-- This column was missing despite having the updated_at_column() trigger
ALTER TABLE auth_session ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT NOW();
