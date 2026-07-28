// Schema for the example/api consumer service, described directly in
// Atlas HCL (the source of truth for `atlas migrate diff`), rather than
// generated from Go/GORM structs.
//
// Covers the "user"/"account" tables (project-dependent — see atlas.hcl)
// plus the "rbac" tables they reference. Keep this in sync by hand with
// ../../internal/domain and repository/{account,rbac}/models when they
// change. See docs/MIGRATIONS.md.

schema "public" {}

table "account_base" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "approve_status" {
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
  column "title" {
    null = true
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "logo_uri" {
    null = true
    type = text
  }
  column "policy_uri" {
    null = true
    type = text
  }
  column "tos_uri" {
    null = true
    type = text
  }
  column "client_uri" {
    null = true
    type = text
  }
  column "contacts" {
    null = true
    type = sql("text[]")
  }
  primary_key {
    columns = [column.id]
  }
}

table "account_user" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "approve_status" {
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
  column "email" {
    null    = false
    type    = text
    default = ""
  }
  column "password" {
    null    = false
    type    = text
    default = ""
  }
  column "required_password_reset" {
    null    = false
    type    = boolean
    default = false
  }
  primary_key {
    columns = [column.id]
  }
}

table "account_member" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }
  column "approve_status" {
    null = true
    type = bigint
  }
  column "account_id" {
    null = true
    type = bigint
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "is_admin" {
    null = true
    type = boolean
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
  foreign_key "fk_account_member_account" {
    columns     = [column.account_id]
    ref_columns = [table.account_base.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_account_member_user" {
    columns     = [column.user_id]
    ref_columns = [table.account_user.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}

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

table "m2m_account_member_role" {
  schema = schema.public

  column "member_id" {
    null = false
    type = bigint
  }
  column "role_id" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.member_id, column.role_id]
  }
  foreign_key "fk_m2m_account_member_role_account_member" {
    columns     = [column.member_id]
    ref_columns = [table.account_member.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_m2m_account_member_role_role" {
    columns     = [column.role_id]
    ref_columns = [table.rbac_role.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
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
