// Schema for the "historylog" repository domain, described directly in
// Atlas HCL (the source of truth for `atlas migrate diff`), rather than
// generated from the Go/GORM structs in ./models. Keep this in sync with
// ./models by hand when the Go structs change. See docs/MIGRATIONS.md.

schema "public" {}

table "history_actions" {
  schema = schema.public

  column "id" {
    null = false
    type = uuid
  }
  column "request_id" {
    null = false
    type = character_varying(255)
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "account_id" {
    null = false
    type = bigint
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  column "message" {
    null = false
    type = text
  }
  column "object_type" {
    null = false
    type = character_varying(255)
  }
  column "object_id" {
    null = false
    type = bigint
  }
  column "object_ids" {
    null = false
    type = character_varying(255)
  }
  column "data" {
    null = false
    type = jsonb
  }
  column "action_at" {
    null = false
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_history_actions_account_id" {
    columns = [column.account_id]
  }
  index "idx_history_actions_at" {
    columns = [column.action_at]
  }
  index "idx_history_actions_name" {
    columns = [column.name]
  }
  index "idx_history_actions_object_id" {
    columns = [column.object_id]
  }
  index "idx_history_actions_object_ids" {
    columns = [column.object_ids]
  }
  index "idx_history_actions_object_type" {
    columns = [column.object_type]
  }
  index "idx_history_actions_request_id" {
    columns = [column.request_id]
  }
  index "idx_history_actions_user_id" {
    columns = [column.user_id]
  }
}
