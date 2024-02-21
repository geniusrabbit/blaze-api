-- Social account connection table with the user
CREATE TABLE IF NOT EXISTS account_social
( ID                      BIGSERIAL                   PRIMARY KEY
, user_id                 BIGINT                      NOT NULL      REFERENCES account_user (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT
, social_id               VARCHAR(128)                NOT NULL
, provider                VARCHAR(128)                NOT NULL

-- Basic info
, email                   VARCHAR(128)                NOT NULL
, first_name              VARCHAR(128)                NOT NULL
, last_name               VARCHAR(128)                NOT NULL
, username                VARCHAR(128)                NOT NULL
, avatar                  VARCHAR(512)                NOT NULL
, link                    VARCHAR(1024)               NOT NULL -- Link to the social profile
, scope                   TEXT[]                      NOT NULL

-- Additional info data from the social network
, data                    JSONB                       NOT NULL

, created_at              TIMESTAMP                   NOT NULL      DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL      DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON account_social FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON account_social
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();

CREATE UNIQUE INDEX idx_account_social_uniq_social_id
    ON account_social (social_id, provider) WHERE deleted_at IS NULL;

-- Social account session
CREATE TABLE IF NOT EXISTS account_social_session
( account_social_id      BIGINT                      NOT NULL      REFERENCES account_social (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT
, token_type            VARCHAR(128)                NOT NULL
, access_token          VARCHAR(512)                NOT NULL
, refresh_token         VARCHAR(512)                NOT NULL

, expires_at            TIMESTAMP                   NOT NULL
, created_at            TIMESTAMP                   NOT NULL      DEFAULT NOW()
, updated_at            TIMESTAMP                   NOT NULL      DEFAULT NOW()
, deleted_at            TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON account_social_session FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON account_social_session
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();