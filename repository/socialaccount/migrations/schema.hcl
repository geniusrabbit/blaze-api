// Schema for the "socialaccount" repository domain, described directly in
// Atlas HCL (the source of truth for `atlas migrate diff`), rather than
// generated from the Go/GORM structs in ./models. Keep this in sync with
// ./models by hand when the Go structs change. See docs/MIGRATIONS.md.

schema "public" {}

table "account_social" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "social_id" {
    null = true
    type = text
  }
  column "provider" {
    null = true
    type = text
  }
  column "email" {
    null = true
    type = text
  }
  column "first_name" {
    null = true
    type = text
  }
  column "last_name" {
    null = true
    type = text
  }
  column "username" {
    null = true
    type = text
  }
  column "avatar" {
    null = true
    type = text
  }
  column "link" {
    null = true
    type = text
  }
  column "data" {
    null = true
    type = jsonb
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}

table "account_social_session" {
  schema = schema.public

  column "name" {
    null = false
    type = text
  }
  column "account_social_id" {
    null = false
    type = bigint
  }
  column "token_type" {
    null = true
    type = text
  }
  column "access_token" {
    null = true
    type = text
  }
  column "refresh_token" {
    null = true
    type = text
  }
  column "scopes" {
    null = true
    type = sql("text[]")
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "expires_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.name, column.account_social_id]
  }
  foreign_key "fk_account_social_sessions" {
    columns     = [column.account_social_id]
    ref_columns = [table.account_social.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
