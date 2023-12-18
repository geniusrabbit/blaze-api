
CREATE TABLE IF NOT EXISTS history_actions (
  id            UUID            PRIMARY KEY DEFAULT uuid_generate_v4()
, user_id       BIGINT          NOT NULL
, account_id    BIGINT          NOT NULL

, name          VARCHAR(255)    NOT NULL
, message       TEXT            NOT NULL

, object_type   VARCHAR(255)    NOT NULL
, object_id     BIGINT          NOT NULL
, object_ids    VARCHAR(255)    NOT NULL

, data         JSONB           NOT NULL

, action_at     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);
