// Schema for the "option" repository domain, described directly in Atlas
// HCL (the source of truth for `atlas migrate diff`), rather than
// generated from the Go/GORM structs in ./models. Keep this in sync with
// ./models by hand when the Go structs change. See docs/MIGRATIONS.md.
//
// NOTE: this table intentionally has no primary key — it mirrors the
// GORM model as-is.

schema "public" {}

table "option" {
  schema = schema.public

  column "type" {
    null = true
    type = text
  }
  column "target_id" {
    null = true
    type = bigint
  }
  column "name" {
    null = true
    type = text
  }
  column "value" {
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
}
