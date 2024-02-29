CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS history_actions (
  id            UUID            PRIMARY KEY DEFAULT uuid_generate_v4()
, request_id    VARCHAR(255)    NOT NULL
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

CREATE INDEX IF NOT EXISTS idx_history_actions_request_id ON history_actions(request_id);
CREATE INDEX IF NOT EXISTS idx_history_actions_at ON history_actions(action_at);
CREATE INDEX IF NOT EXISTS idx_history_actions_user_id ON history_actions(user_id);
CREATE INDEX IF NOT EXISTS idx_history_actions_account_id ON history_actions(account_id);

CREATE INDEX IF NOT EXISTS idx_history_actions_name ON history_actions(name);
CREATE INDEX IF NOT EXISTS idx_history_actions_object_type ON history_actions(object_type);
CREATE INDEX IF NOT EXISTS idx_history_actions_object_id ON history_actions(object_id);
CREATE INDEX IF NOT EXISTS idx_history_actions_object_ids ON history_actions(object_ids);
