// Schema for the "directaccesstoken" repository domain, described
// directly in Atlas HCL (the source of truth for `atlas migrate diff`),
// rather than generated from the Go/GORM structs in ./models. Keep this in
// sync with ./models by hand when the Go structs change. See
// docs/MIGRATIONS.md.

schema "public" {}

table "direct_access_tokens" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "token" {
    null = true
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "account_id" {
    null = true
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "expires_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
