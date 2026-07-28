# Database migrations

This project supports **two** independent ways to manage database schema
migrations. They are not mutually exclusive on the tooling level, but for any
given deployment you pick **one** of them as the source of truth for a given
table set:

1. **`golang-migrate` / SQL** (existing, primary) — hand-written SQL files
   under [`deploy/migrations`](../deploy/migrations), applied either from
   inside the service binary at startup, or via the standalone
   `migrate/migrate` Docker image. This is what `make run-api` uses.
2. **Atlas / schema-as-HCL** (new, opt-in) — the schema is described
   directly in [Atlas HCL](https://atlasgo.io/atlas-schema/hcl) (`schema.hcl`
   files), using [Atlas](https://atlasgo.io) to diff/apply/version it. This
   is what the rest of this document describes. It has two flavors:
   per-domain **versioned** migrations (one `migrations/` directory per
   domain, mirroring `golang-migrate`'s model), and, for `example/api`
   specifically, a **declarative**, combined-schema shortcut that applies
   directly to a live database with no SQL files at all — see
   [Combined, declarative "app" schema](#combined-declarative-app-schema-exampleapi).

Atlas is **fully additive**: nothing under `pkg/migratedb`,
`example/api/cmd/api/main.go`, or `example/api/deploy/develop/migrate.Dockerfile`
changes. Atlas is never imported by the service — no Go code in this repo
depends on any Atlas package. It is driven purely by the `atlas` CLI,
invoked from your machine, from Docker, or from CI — never from `main.go`'s
`init()`.

## Why HCL, and why not generate from GORM structs

An earlier iteration of this generated the Atlas schema from the existing
GORM struct tags via
[`ariga.io/atlas-provider-gorm`](https://github.com/ariga/atlas-provider-gorm).
That was dropped in favor of hand-written `schema.hcl` files, for a few
reasons:

- It requires a Go toolchain (`go run ...`) just to compute a schema, which
  in turn requires a Go+Atlas Docker image instead of the small official
  `arigaio/atlas` one.
- GORM's reflection-based schema builder doesn't handle every shape well —
  e.g. it panics on many2many join tables owned by a **generic** struct (see
  the note about `account.Member[TUser, TAccount]` below, which used to need
  a hand-written non-generic mirror type just to work around this).
- The GORM tags encode Go-level relations (`gorm:"foreignKey:..."`), not the
  SQL DDL itself — HCL is a more direct, more debuggable representation of
  exactly what will be created, and it's what `atlas migrate diff`/`schema
  inspect` already emit natively.

So `schema.hcl` is now the single source of truth Atlas diffs against.
**It's hand-maintained** — when you add/rename/remove a field on a GORM
model that's backed by one of these tables, update the corresponding
`schema.hcl` too, then run `atlas migrate diff` to generate the migration.

## Why split by repository

Each Atlas-managed domain owns its **own** `schema.hcl` and `migrations/`
directory, colocated with the code that owns the concrete schema — there's
no separate Atlas project file (`atlas.hcl`) per domain; see
[No per-domain `atlas.hcl`](#no-per-domain-atlashcl) for why:

| Domain | Schema/migrations location | Why |
|---|---|---|
| `rbac` | [`repository/rbac/migrations`](../repository/rbac/migrations) | Self-contained: no generics, no consumer-side trait overrides. |
| `historylog` | [`repository/historylog/migrations`](../repository/historylog/migrations) | Self-contained. |
| `option` | [`repository/option/migrations`](../repository/option/migrations) | Self-contained. |
| `directaccesstoken` | [`repository/directaccesstoken/migrations`](../repository/directaccesstoken/migrations) | Self-contained. |
| `authclient` | [`repository/authclient/migrations`](../repository/authclient/migrations) | Self-contained. |
| `socialaccount` | [`repository/socialaccount/migrations`](../repository/socialaccount/migrations) | Self-contained. |
| `user` + `account` | [`example/api/migrations/atlas/migrations`](../example/api/migrations/atlas/migrations) | **Project-dependent** — see below. |

`user` and `account` are handled differently on purpose:

- [`repository/account/member.go`](../repository/account/member.go) defines
  the **generic** `Member[TUser user.Model, TAccount Model]` type, with real
  GORM `belongsTo` associations to the type parameters themselves
  (`Account TAccount gorm:"foreignKey:AccountID;references:ID"`,
  `User TUser gorm:"foreignKey:UserID;references:ID"`). Its concrete schema
  only exists once `TUser`/`TAccount` are instantiated by a consumer —
  here, [`example/api/internal/domain`](../example/api/internal/domain)
  (`User`, `Account`, `AccountMember = account.Member[*User, *Account]`).
- [`example/api/internal/domain/user.go`](../example/api/internal/domain/user.go)
  and [`account.go`](../example/api/internal/domain/account.go) also embed
  **traits** (`UserEmail`, `UserPassword`, `AccountProfile`) that are
  consumer-specific extensions, not part of the base `UserBase`/`AccountBase`
  structs in `repository/user`/`repository/account`.
- `MemberBase.Roles []*rbac.Role` (many2many via `m2m_account_member_role`)
  pulls `repository/rbac` into the same schema — the only real cross-package
  GORM association in the codebase.

So the concrete `user`/`account` schema is **project-dependent**, and its
`schema.hcl`/`migrations` live in `example/api`, not in `repository/user` or
`repository/account` (which get no new files at all). Any other consumer
project that defines its own `User`/`Account` types (with its own traits)
would similarly own its own `schema.hcl` for those two domains, following
the same pattern as `example/api/migrations/atlas`.

**Caveat:** if a project only needs RBAC without the `account`/`user`
domain, it applies the standalone `repository/rbac` schema. Never apply both
`repository/rbac` and `example/api/migrations/atlas` to the same database —
the latter already includes `rbac_role`/`m2m_rbac_role`
(see [`example/api/migrations/atlas/migrations/schema.hcl`](../example/api/migrations/atlas/migrations/schema.hcl)),
mirroring how [`deploy/migrations/initial/002_account.up.sql`](../deploy/migrations/initial/002_account.up.sql)
already bundles `account_user` + `account_base` + `rbac_role` +
`m2m_rbac_role` + `account_member` + `m2m_account_member_role` together for
exactly the same FK-coupling reason.

## Scope boundary vs. `deploy/migrations`

Atlas here describes the schema equivalent of `deploy/migrations/initial/*`
only:

- **`deploy/migrations/fixtures/*`** (seed data) stays exclusively
  `golang-migrate`/SQL — Atlas has nothing to diff seed *data* against, only
  schema.
- **`deploy/migrations/traits/*`** (optional `ALTER TABLE` columns for
  consumer-specific extensions) is largely redundant for `example/api`
  specifically, since its traits (`AccountProfile`, `UserEmail`,
  `UserPassword`) are already expressed directly as columns in
  [`example/api/migrations/atlas/migrations/schema.hcl`](../example/api/migrations/atlas/migrations/schema.hcl).
  Other consumer projects that don't define those same trait columns still
  rely on the SQL trait files.

## How the schema is described

Every domain has a plain `schema.hcl` **inside its own `migrations/`
directory**, written in
[Atlas's native HCL syntax](https://atlasgo.io/atlas-schema/hcl) — no code
generation, no external program, no project file:

```text
repository/rbac/
  migrations/
    schema.hcl   # desired-state source of truth (hand-maintained, tracked in git)
```

Only `schema.hcl` is committed. Running `make atlas-diff DIR=repository/rbac` (see
[Running locally](#running-locally)) adds the other two files Atlas manages in
that same directory — a timestamped `<timestamp>.sql` per change and the
`atlas.sum` checksum — which are **not** committed to this template (see
[No committed migration history](#no-committed-migration-history)):

```text
repository/rbac/
  migrations/
    schema.hcl
    20260728105005.sql   # generated versioned migration(s)
    atlas.sum             # migration directory checksum, managed by Atlas
```

```hcl
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
  # ...
  primary_key {
    columns = [column.id]
  }
}
```

See [`atlas-schema/hcl`](https://atlasgo.io/atlas-schema/hcl) and
[`atlas-schema/hcl-types`](https://atlasgo.io/atlas-schema/hcl-types) for the
full syntax reference (column types, indexes, foreign keys, etc). A few
things worth knowing when editing these files in this repo:

- Postgres array columns use `type = sql("text[]")` (there's no native array
  keyword in the HCL DSL).
- `bigserial`/`serial` are macros for a non-nullable integer column backed by
  a sequence — matches GORM's default `id` column behavior.
- `repository/option/migrations/schema.hcl`'s `option` table intentionally
  has **no** `primary_key` block — it mirrors the GORM model as-is.

### No committed migration history

This is a **template** repository, so none of the `migrations/` directories
above ship a pre-generated `<timestamp>.sql`/`atlas.sum` — only the
hand-maintained `schema.hcl`. A committed initial migration here would just
be an artifact of this template's own testing, tied to whatever schema
existed (and whatever timestamp it happened to run at) when it was
generated — not something a downstream project should build its real
migration history on top of.

Run `make atlas-diff DIR=<domain>` (see [Running locally](#running-locally))
once, in **your own** fork/use of this template, to generate your project's
first migration — after you've made whatever schema changes you need — and
commit the resulting `.sql`/`atlas.sum` there. From that point on, this is a
completely normal versioned Atlas migration history: subsequent
`atlas-diff` runs append new files, `atlas.sum` gets updated by Atlas
itself, and both are meant to be committed like any other `golang-migrate`
SQL file under `deploy/migrations`.

### No per-domain `atlas.hcl`

Unlike a typical Atlas project, there is **no** `atlas.hcl` file next to
each `schema.hcl` — commands are invoked with explicit `--to`/`--dir`/
`--dev-url` flags instead of `--env`. An `atlas.hcl` per domain would mostly
have been repetitive boilerplate (the same `variable "dialect"`/`locals`
dance seven times over), so that little bit of config
(dialect → dev-database URL mapping) lives once, in the root
[`Makefile`](../Makefile), instead. If you invoke `atlas` directly without
`make` (see [Running locally](#running-locally)), just pass the same flags
by hand — see the examples below.

## Installing Atlas

Local install (macOS/Linux):

```bash
curl -sSf https://atlasgo.sh | sh
```

See the [Atlas docs](https://atlasgo.io/getting-started#installation) for
other options. You'll also need Docker available locally — the `dev`
database used for diffing is spun up as an ephemeral container via Atlas's
`docker://` driver (Docker Desktop or an equivalent daemon must be running).

Alternatively, use the bundled `deploy/develop/atlas.Dockerfile`
(see [Running via Docker](#running-via-docker) below) if you don't want to
install anything locally.

## Running locally

```bash
# Generate a new migration for a domain after changing its migrations/schema.hcl
make atlas-diff DIR=repository/rbac
make atlas-diff DIR=example/api/migrations/atlas

# Apply migrations to a real database
make atlas-apply DIR=repository/rbac URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable&search_path=public"

# Check migration status against a real database
make atlas-status DIR=repository/rbac URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable&search_path=public"
```

`DIR` can be any of:

```text
repository/rbac
repository/historylog
repository/option
repository/directaccesstoken
repository/authclient
repository/socialaccount
example/api/migrations/atlas
```

Or, without `make`, straight `atlas` CLI equivalents run from inside the
domain directory (no config file — see
[No per-domain `atlas.hcl`](#no-per-domain-atlashcl)):

```bash
cd repository/rbac
atlas migrate diff --to "file://migrations/schema.hcl" --dir "file://migrations" --dev-url "docker://postgres/16/dev?search_path=public" --format '{{ sql . "  " }}'
atlas migrate apply --dir "file://migrations" --url "$DATABASE_URL"
atlas migrate status --dir "file://migrations" --url "$DATABASE_URL"
```

`ATLAS_DIALECT` (used by the `make` targets above) also accepts `mysql` (and
`sqlite` for the standalone domains); it controls which ephemeral `dev`
database Atlas spins up for diffing — see `ATLAS_DEV_URL_*` in the root
[`Makefile`](../Makefile). Note that `schema.hcl` in this repo is written
using Postgres HCL syntax — targeting another dialect would mean
maintaining a separate `schema.hcl` (or an Atlas
[composite schema](https://atlasgo.io/atlas-schema/projects)) for it.

## Running via Docker

No local Atlas install required. This uses a small dev-only image (the
official `arigaio/atlas` image plus the Docker CLI, so the `docker://`
dev-database driver works) built from
[`deploy/develop/atlas.Dockerfile`](../deploy/develop/atlas.Dockerfile) —
completely separate from the service's own
[`example/api/deploy/develop/api.Dockerfile`](../example/api/deploy/develop/api.Dockerfile).
The repo is bind-mounted into the container at run time, so the image never
needs rebuilding when a `schema.hcl` changes:

```bash
make atlas-docker-diff DIR=repository/rbac
make atlas-docker-apply DIR=repository/rbac URL="postgres://user:pass@host.docker.internal:5432/mydb?sslmode=disable&search_path=public"
```

These mount `/var/run/docker.sock` (so Atlas's `docker://` dev-database
driver can ask the host's Docker daemon to spin up ephemeral containers —
the same mechanism used when running Atlas natively on a machine with Docker
Desktop) and the repository root itself (so generated migration files land
on the host, not just inside the ephemeral container).

## Combined, declarative "app" schema (example/api)

Everything above manages **one domain at a time**, with its own versioned
migration history — good for libraries/packages, but tedious if you just
want to stand up (or update) a whole consumer service's database in one
shot. [`example/api/migrations/atlas-app`](../example/api/migrations/atlas-app)
offers a third, complementary workflow for exactly that case: it combines
every `schema.hcl` this service depends on into one schema and applies it
**directly** to a live database with `atlas schema apply` — a purely
**declarative** diff-and-apply, no per-domain (or per-app) SQL migration
files generated or tracked at all.

```text
example/api/migrations/atlas-app/
  atlas.hcl            # env "app": src = "file://.assembled/"
  assemble-schema.sh   # (re)builds .assembled/ from the schema.hcl files below
  .assembled/          # generated, gitignored — do not edit directly
```

(This is the one place in the Atlas setup that still has an `atlas.hcl` —
it's the merge target Atlas needs a project file for; see
[No per-domain `atlas.hcl`](#no-per-domain-atlashcl) for why the individual
domains don't have one.)

It's assembled from:

- [`example/api/migrations/atlas/migrations/schema.hcl`](../example/api/migrations/atlas/migrations/schema.hcl)
  (`user`/`account`, which already includes `rbac_role`/`m2m_rbac_role`)
- [`repository/historylog/migrations/schema.hcl`](../repository/historylog/migrations/schema.hcl)
- [`repository/option/migrations/schema.hcl`](../repository/option/migrations/schema.hcl)
- [`repository/directaccesstoken/migrations/schema.hcl`](../repository/directaccesstoken/migrations/schema.hcl)
- [`repository/authclient/migrations/schema.hcl`](../repository/authclient/migrations/schema.hcl)
- [`repository/socialaccount/migrations/schema.hcl`](../repository/socialaccount/migrations/schema.hcl)

`repository/rbac/migrations/schema.hcl` is deliberately **not** included —
its tables are already part of the `user`/`account` schema above (see
[Why split by repository](#why-split-by-repository)).

**Why not just list all the files directly in `atlas.hcl`'s `src`?** Atlas
treats a `src` list of files as multiple *separate* schemas, not fragments
of one — it only merges multiple HCL files into a single schema when `src`
points at one **directory**. Since every `schema.hcl` above also declares
its own `schema "public" {}` (so it works standalone too), and Atlas
refuses to merge two files that both declare the same schema, they can't be
combined as-is either. `assemble-schema.sh` resolves both problems: it
copies the table definitions from each file, without their `schema "public"
{}` header, into `.assembled/`, alongside one shared copy of that
declaration.

**Regenerate `.assembled/` whenever any of the source `schema.hcl` files
change** — it's a derived artifact, not something to hand-edit:

```bash
cd example/api && make atlas-app-assemble
```

Diff/apply against a real database:

```bash
cd example/api
make atlas-app-diff  URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable&search_path=public"
make atlas-app-apply URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable&search_path=public"          # prompts for confirmation
make atlas-app-apply URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable&search_path=public" AUTO_APPROVE=1  # non-interactive
```

Or as a Docker image + docker-compose service, analogous to the existing
`migration`/`migration-fixtures` services but declarative instead of
SQL-file-based — the combined schema is assembled and baked into the image
at *build* time (see
[`example/api/deploy/develop/atlas-apply.Dockerfile`](../example/api/deploy/develop/atlas-apply.Dockerfile)),
so no source checkout is needed at run time:

```bash
cd example/api
make migrate-atlas   # docker compose run --rm migration-atlas-app
```

This mounts `/var/run/docker.sock` (for the `docker://` dev-database driver,
same as the other `atlas-docker-*` targets) and applies straight to
`api-template-db` using `--auto-approve` (see the `migration-atlas-app`
service in
[`docker-compose.yml`](../example/api/deploy/develop/docker-compose.yml)).

**Trade-off vs. the per-domain versioned workflow:** declarative apply is
simpler to run and doesn't need any tracked SQL files, but it has no
migration history to review/rollback and no `schema_migrations`-style
ledger — Atlas computes the diff against the live database every time. For
a database anyone else also runs `deploy/migrations`/`migrate` against, or
one you need an auditable change history for, prefer the versioned,
per-domain `atlas-diff`/`atlas-apply` flow (or `golang-migrate`) instead.

## Adding Atlas support to a new domain

1. Create a `migrations/` directory in the domain's package (e.g.
   `repository/<domain>/migrations/`).
2. Write a `migrations/schema.hcl` describing the tables that domain owns,
   matching the GORM models' columns/types/constraints (see
   [`repository/rbac/migrations/schema.hcl`](../repository/rbac/migrations/schema.hcl)
   as a reference). No `atlas.hcl` needed — see
   [No per-domain `atlas.hcl`](#no-per-domain-atlashcl).
3. Run `make atlas-diff DIR=repository/<domain>` (or the Docker equivalent)
   to generate the first migration file, and review the generated SQL.
