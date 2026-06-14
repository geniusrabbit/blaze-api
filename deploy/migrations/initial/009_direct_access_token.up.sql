CREATE TABLE IF NOT EXISTS direct_access_tokens
( id                      BIGSERIAL                   PRIMARY KEY
, token                   VARCHAR(256)                NOT NULL       UNIQUE
, description             TEXT                        NOT NULL

, account_id              BIGINT                      NOT NULL      REFERENCES account_base (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT
, user_id                 BIGINT                                    REFERENCES account_user (id) MATCH SIMPLE
                                                                        ON UPDATE NO ACTION
                                                                        ON DELETE RESTRICT

, created_at              TIMESTAMP                   NOT NULL      DEFAULT NOW()
, expires_at              TIMESTAMP
);

CREATE TRIGGER notify_update_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON direct_access_tokens
    FOR EACH ROW EXECUTE PROCEDURE notify_update_event();
