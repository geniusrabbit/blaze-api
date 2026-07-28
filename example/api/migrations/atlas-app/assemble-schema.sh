#!/usr/bin/env bash
# Assembles a single combined Atlas schema directory (./.assembled) out of
# every repository/<domain>/migrations/schema.hcl this service depends on,
# plus its own user/account schema
# (../atlas/migrations/schema.hcl, which already includes
# rbac_role/m2m_rbac_role — see docs/MIGRATIONS.md for why repository/rbac
# is intentionally excluded here).
#
# Each source file is also usable completely standalone (via plain `atlas
# migrate diff/apply/status --dir file://migrations ...`, no project file
# needed — see docs/MIGRATIONS.md), which means each one declares its own
# `schema "public" {}`. Atlas errors out ("schema "public" is defined more
# than once") if a combined directory contains more than one such
# declaration, so this script strips it from every fragment and adds
# exactly one shared copy (./.assembled/public.hcl).
#
# Used by both `make atlas-app-*` (see ../../Makefile) and
# ../../deploy/develop/atlas-apply.Dockerfile.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

REPO_ROOT="../../../.."
OUT=".assembled"

rm -rf "$OUT"
mkdir -p "$OUT"

echo 'schema "public" {}' >"$OUT/public.hcl"

strip_schema_decl() {
	sed '/^schema "public" {}$/d' "$1" >"$2"
}

strip_schema_decl "../atlas/migrations/schema.hcl" "$OUT/local.hcl"
strip_schema_decl "$REPO_ROOT/repository/historylog/migrations/schema.hcl" "$OUT/historylog.hcl"
strip_schema_decl "$REPO_ROOT/repository/option/migrations/schema.hcl" "$OUT/option.hcl"
strip_schema_decl "$REPO_ROOT/repository/directaccesstoken/migrations/schema.hcl" "$OUT/directaccesstoken.hcl"
strip_schema_decl "$REPO_ROOT/repository/authclient/migrations/schema.hcl" "$OUT/authclient.hcl"
strip_schema_decl "$REPO_ROOT/repository/socialaccount/migrations/schema.hcl" "$OUT/socialaccount.hcl"

echo "Assembled combined app schema into $(pwd)/$OUT/"
