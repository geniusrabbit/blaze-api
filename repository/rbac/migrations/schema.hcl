// Schema for the "rbac" repository domain, described directly in Atlas HCL
// (the source of truth for `atlas migrate diff`), rather than generated
// from the Go/GORM structs in ./models. Keep this in sync with ./models by
// hand when the Go structs change. See docs/MIGRATIONS.md.

schema "public" {}

table "rbac_role" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "name" {
    null = true
    type = text
  }
  column "title" {
    null = true
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "context" {
    null = true
    type = text
  }
  column "permissions" {
    null = true
    type = sql("text[]")
  }
  column "access_level" {
    null = true
    type = bigint
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

table "m2m_rbac_role" {
  schema = schema.public

  column "parent_role_id" {
    null = false
    type = bigint
  }
  column "child_role_id" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.parent_role_id, column.child_role_id]
  }
  foreign_key "fk_m2m_rbac_role_role" {
    columns     = [column.parent_role_id]
    ref_columns = [table.rbac_role.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_m2m_rbac_role_child_roles" {
    columns     = [column.child_role_id]
    ref_columns = [table.rbac_role.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
