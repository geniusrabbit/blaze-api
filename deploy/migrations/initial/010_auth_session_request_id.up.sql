-- PKCE and authorize-code sessions share a fosite request_id but are stored as
-- separate auth_session rows (different access_token signatures). Token exchange
-- then inserts a third row with the same request_id for the access token.
-- Unique request_id made the second INSERT fail (SQLSTATE 23505).
DROP INDEX IF EXISTS idx_auth_session_uniq_request_id;

CREATE INDEX IF NOT EXISTS idx_auth_session_request_id
    ON auth_session (request_id) WHERE deleted_at IS NULL;
