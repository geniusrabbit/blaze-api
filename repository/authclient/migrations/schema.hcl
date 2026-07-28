// Schema for the "authclient" repository domain, described directly in
// Atlas HCL (the source of truth for `atlas migrate diff`), rather than
// generated from the Go/GORM structs in ./models. Keep this in sync with
// ./models by hand when the Go structs change. See docs/MIGRATIONS.md.

schema "public" {}

table "auth_client" {
  schema = schema.public

  column "id" {
    null = false
    type = text
  }
  column "account_id" {
    null = true
    type = bigint
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "title" {
    null = true
    type = text
  }
  column "secret" {
    null = true
    type = text
  }
  column "redirect_uris" {
    null = true
    type = sql("text[]")
  }
  column "grant_types" {
    null = true
    type = sql("text[]")
  }
  column "response_types" {
    null = true
    type = sql("text[]")
  }
  column "scope" {
    null = true
    type = text
  }
  column "audience" {
    null = true
    type = sql("text[]")
  }
  column "subject_type" {
    null = true
    type = text
  }
  column "allowed_cors_origins" {
    null = true
    type = sql("text[]")
  }
  column "public" {
    null = true
    type = boolean
  }
  column "expires_at" {
    null = true
    type = timestamptz
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

table "auth_session" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "active" {
    null = true
    type = boolean
  }
  column "client_id" {
    null = true
    type = text
  }
  column "username" {
    null = true
    type = text
  }
  column "subject" {
    null = true
    type = text
  }
  column "request_id" {
    null = true
    type = text
  }
  column "access_token" {
    null = true
    type = text
  }
  column "access_token_expires_at" {
    null = true
    type = timestamptz
  }
  column "refresh_token" {
    null = true
    type = text
  }
  column "refresh_token_expires_at" {
    null = true
    type = timestamptz
  }
  column "form" {
    null = true
    type = text
  }
  column "requested_scope" {
    null = true
    type = sql("text[]")
  }
  column "granted_scope" {
    null = true
    type = sql("text[]")
  }
  column "requested_audience" {
    null = true
    type = sql("text[]")
  }
  column "granted_audience" {
    null = true
    type = sql("text[]")
  }
  column "created_at" {
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
