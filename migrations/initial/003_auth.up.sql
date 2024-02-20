
CREATE TABLE auth_client
( id                      VARCHAR(128)                NOT NULL      PRIMARY KEY
, account_id              BIGINT                      NOT NULL      REFERENCES account_base (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT
, user_id                 BIGINT                      NOT NULL      REFERENCES account_user (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT

, title                   TEXT                        NOT NULL
, secret                  VARCHAR(256)                NOT NULL

-- RedirectURIs is an array of allowed redirect urls for the client, for example http://mydomain/oauth/callback .
, redirect_uris           TEXT[]
-- List of available auntification methods
-- https://www.oauth.com/oauth2-servers/access-tokens/
-- password           - https://www.oauth.com/oauth2-servers/access-tokens/password-grant/
-- authorization_code - https://www.oauth.com/oauth2-servers/access-tokens/authorization-code-request/
-- client_credentials - https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/
-- Pattern: client_credentials|authorization_code|implicit|refresh_token
, grant_types             TEXT[]
-- Pattern: id_token|code|token
, response_types          TEXT[]

--  Scope is a string containing a space-separated list of scope values (as
--  described in Section 3.3 of OAuth 2.0 [RFC6749]) that the client
--  can use when requesting access tokens.
-- 
--  Pattern: ([a-zA-Z0-9\.\*]+\s?)+
, scope                   TEXT

-- Audience is a whitelist defining the audiences this client is allowed to request tokens for. An audience limits
-- the applicability of an OAuth 2.0 Access Token to, for example, certain API endpoints. The value is a list
-- of URLs. URLs MUST NOT contain whitespaces.
, audience                TEXT[]

-- SubjectType requested for responses to this Client. The subject_types_supported Discovery parameter contains a
-- list of the supported subject_type values for this server. Valid types include `pairwise` and `public`.
, subject_type            VARCHAR(128)

-- AllowedCORSOrigins are one or more URLs (scheme://host[:port]) which are allowed to make CORS requests
-- to the /oauth/token endpoint. If this array is empty, the sever's CORS origin configuration (`CORS_ALLOWED_ORIGINS`)
-- will be used instead. If this array is set, the allowed origins are appended to the server's CORS origin configuration.
-- Be aware that environment variable `CORS_ENABLED` MUST be set to `true` for this to work.
, allowed_cors_origins    TEXT[]

, public                  BOOLEAN                     NOT NULL        DEFAULT FALSE

-- There is no way to issue the auth-gate without expiration time but can be issued really long-living access
, expires_at              TIMESTAMP                   NOT NULL

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON auth_client FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON auth_client
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();

-- NOTE: for specific user permissions need to implement additional table like connect between user and gateway

CREATE TABLE auth_session
( id                      BIGSERIAL                   PRIMARY KEY
, active                  BOOLEAN                     NOT NULL      DEFAULT TRUE

, client_id               VARCHAR(128)                NOT NULL      REFERENCES auth_client (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT
, username                VARCHAR(128)                NOT NULL      DEFAULT ''
, subject                 VARCHAR(128)                NOT NULL      DEFAULT ''

, request_id                VARCHAR(256)                NOT NULL
, access_token              VARCHAR(256)                NOT NULL    -- The main session code
, access_token_expires_at   TIMESTAMP                   NOT NULL
, refresh_token             VARCHAR(256)
, refresh_token_expires_at  TIMESTAMP                   NOT NULL

, form                    TEXT                        NOT NULL      DEFAULT ''
, requested_scope         TEXT[]
, granted_scope           TEXT[]
, requested_audience      TEXT[]
, granted_audience        TEXT[]

, created_at              TIMESTAMP                   NOT NULL      DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE UNIQUE INDEX idx_auth_session_uniq_request_id
    ON auth_session (request_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_auth_session_uniq_access_token_id
    ON auth_session (access_token) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_auth_session_uniq_refresh_token_id
    ON auth_session (refresh_token) WHERE LENGTH(COALESCE(refresh_token, '')) > 0 AND deleted_at IS NULL;

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON auth_session FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON auth_session
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();
