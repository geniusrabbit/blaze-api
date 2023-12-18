CREATE TABLE IF NOT EXISTS option
( name               VARCHAR(256)                NOT NULL
, type               VARCHAR(32)                 NOT NULL
, target_id          BIGINT                      NOT NULL
, value              JSONB                       NOT NULL

, created_at         TIMESTAMP                   NOT NULL      DEFAULT NOW()
, updated_at         TIMESTAMP                   NOT NULL      DEFAULT NOW()
, deleted_at         TIMESTAMP

, PRIMARY KEY (name, type, target_id)
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON option FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER notify_update_event_trigger
  AFTER INSERT OR UPDATE OR DELETE ON option
      FOR EACH ROW EXECUTE PROCEDURE notify_update_event();
