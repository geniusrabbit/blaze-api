// Atlas project file for a single, combined "app" schema for the
// example/api service — the "user"/"account" schema
// (../atlas/migrations/schema.hcl, which already includes
// rbac_role/m2m_rbac_role) plus every other self-contained
// repository/<domain> schema this service depends on (historylog, option,
// directaccesstoken, authclient, socialaccount).
//
// Unlike ../atlas/migrations (and the repository/<domain>/migrations
// directories), which each own a *versioned* migration history, this env
// is purely **declarative**: `atlas schema apply` diffs the combined
// desired schema directly against a real database's live state and
// applies the delta on the spot — no SQL migration files are generated or
// tracked for it.
//
// The combined schema is assembled by ./assemble-schema.sh into
// ./.assembled (gitignored — regenerate it any time the source
// schema.hcl files change) since each source file also declares its own
// `schema "public" {}` for standalone use, and Atlas refuses to combine
// two files that both declare the same schema. See docs/MIGRATIONS.md.

variable "dialect" {
  type    = string
  default = "postgres"
}

locals {
  dev_url = {
    postgres = "docker://postgres/16/dev?search_path=public"
    mysql    = "docker://mysql/8/dev"
  }[var.dialect]
}

env "app" {
  src = "file://.assembled/"
  dev = local.dev_url
}
