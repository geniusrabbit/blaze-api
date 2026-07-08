# Configuring example/api — Minimal Project Setup

`example/api` is a reference implementation of a minimal API project built on top of `blaze-api`.
This document explains which parts are mandatory and which are optional "traits" that can be removed.

---

## Project Structure

```
example/api/
├── cmd/api/
│   ├── appcontext/config.go     # All config structs + env var bindings
│   ├── appinit/                 # Startup helpers (deps, auth, superuser)
│   └── main.go                  # Entry point: init → migrations → HTTP server
├── deploy/migrations/
│   ├── initial/                 # Core tables (always required)
│   ├── fixtures/                # Seed data
│   └── traits/                  # Optional column additions (user_email, user_password, …)
├── internal/
│   ├── domain/                  # Your User and Account model definitions
│   └── server/graphql/          # Generated GraphQL code + resolvers wiring
└── protocol/graphql/
    ├── gqlgen.yml               # Schema file list → controls generated models
    └── extensions/*.graphql     # Project-specific schema extensions
```

---

## Domain Models — the Core Decision

The two models `User` and `Account` in `internal/domain/` are composed by embedding **trait structs**
from the `repository/` layer. You control which features exist by choosing which traits to embed.

### User traits

| Trait      | Struct to embed           | What it adds                             |
| ---------- | ------------------------- | ---------------------------------------- |
| _(always)_ | `userModels.UserBase`     | `id`, `status`, `createdAt`, `updatedAt` |
| Email auth | `userModels.UserEmail`    | `email` field + email-based login        |
| Password   | `userModels.UserPassword` | bcrypt password + `CreateWithPassword`   |
| Username   | `userModels.UserUsername` | separate `username` field                |

**Minimal `domain/user.go` (email + password login only):**

```go
type User struct {
    userModels.UserBase
    userModels.UserEmail
    userModels.UserPassword
    // userModels.UserUsername  ← uncomment to add username support
}
```

### Account traits

| Trait      | Struct to embed             | What it adds                              |
| ---------- | --------------------------- | ----------------------------------------- |
| _(always)_ | `accountModels.AccountBase` | `id`, `status`, admins, `createdAt`       |
| Profile    | `domain.AccountProfile`     | title, description, logo, policy/tos URLs |

---

## Trait Activation Checklist

Every trait has **three places** that must be kept in sync:

### 1. Domain model (`internal/domain/user.go` or `account.go`)

Embed or comment out the struct:

```go
// userModels.UserUsername  ← commented out = disabled
```

### 2. Migration (`deploy/migrations/traits/`)

Each trait has a corresponding SQL file:

| File                   | Adds                             |
| ---------------------- | -------------------------------- |
| `user_email.up.sql`    | `email` column + unique index    |
| `user_password.up.sql` | `password_hash` column           |
| `user_username.up.sql` | `username` column + unique index |
| `account.up.sql`       | account profile columns          |

These run under the `traits` migration source (separate `schema_migrations_traits` table).
If you don't embed a trait, simply omit its migration file — the column won't exist and nothing
references it.

### 3. GraphQL schema (`protocol/graphql/gqlgen.yml`)

Each trait has a corresponding `.graphql` file that adds fields to `User` or `Account` via
`extend type`. Add or remove the corresponding line:

```yaml
schema:
  # Core schemas (always required)
  - ../../../../protocol/graphql/schemas/*.graphql
  - ../../../../repository/*/*/graphql/*.graphql # flat domain schemas (rbac, historylog, …)
  - ../../../../repository/account/*/graphql/account_base/*.graphql
  - ../../../../repository/account/*/graphql/account_login/*.graphql
  - ../../../../repository/account/*/graphql/account_member.graphql
  - ../../../../repository/user/*/graphql/user_base/*.graphql

  # Optional traits — add the line to activate, remove to disable:
  - ../../../../repository/user/*/graphql/user_email/*.graphql
  - ../../../../repository/user/*/graphql/user_password/*.graphql
  - ../../../../repository/user/*/graphql/user_password_reset/*.graphql
  # - ../../../../repository/user/*/graphql/user_username/*.graphql  ← disabled

  # Project-specific extensions
  - ./extensions/*.graphql
```

> **Note — gqlgen glob behaviour**: gqlgen converts `*` to the regex `.+` which matches `/`
> (path separators). This means `graphql/*.graphql` also matches files inside subdirectories.
> Use `*/graphql/*.graphql` (one intermediate `*`) for flat schema files, and explicit
> subdirectory paths for trait subdirectories such as `user_username/`.

After changing `gqlgen.yml` regenerate:

```bash
make build-gql
```

---

## Minimum Required Setup

The absolute minimum for a working authenticated API:

**Domain embeds:**

```go
// User
userModels.UserBase
userModels.UserEmail
userModels.UserPassword

// Account
accountModels.AccountBase
domain.AccountProfile
```

**Migrations to run:**

- `initial/` — all files (core tables)
- `traits/user_email.up.sql`
- `traits/user_password.up.sql`
- `traits/account.up.sql`

**`gqlgen.yml` schema lines (minimum):**

```yaml
- ../../../../protocol/graphql/schemas/*.graphql
- ../../../../repository/*/*/graphql/*.graphql
- ../../../../repository/account/*/graphql/account_base/*.graphql
- ../../../../repository/account/*/graphql/account_login/*.graphql
- ../../../../repository/account/*/graphql/account_member.graphql
- ../../../../repository/user/*/graphql/user_base/*.graphql
- ../../../../repository/user/*/graphql/user_email/*.graphql
- ../../../../repository/user/*/graphql/user_password/*.graphql
- ./extensions/*.graphql
```

---

## Environment Variables

### Required

```bash
# PostgreSQL
SYSTEM_STORAGE_DATABASE_MASTER_CONNECT=postgres://user:pass@localhost:5432/db?sslmode=disable
SYSTEM_STORAGE_DATABASE_SLAVE_CONNECT=postgres://user:pass@localhost:5432/db?sslmode=disable

# OAuth2 / JWT (secret must be ≥ 32 characters)
OAUTH2_SECRET=change-me-to-a-random-32-char-secret
```

### Superuser (first boot)

```bash
SUPERUSER_EMAIL=admin@example.com
SUPERUSER_PASSWORD=strongpassword
```

`EnsureSuperuser` runs at startup, creates the user + system account + `system:admin` role,
and is idempotent — safe to leave set after first boot.

### Development mode

```bash
LOG_LEVEL=debug
SESSION_DEV_TOKEN=develop          # static token accepted as Authorization header
SESSION_DEV_USER_ID=1
SESSION_DEV_ACCOUNT_ID=1
```

### Optional

```bash
OAUTH2_ACCESS_TOKEN_LIFESPAN=1h
OAUTH2_REFRESH_TOKEN_LIFESPAN=720h
SESSION_COOKIE_NAME=sessid
SESSION_LIFETIME=1h
PERMISSIONS_CACHE_LIFETIME=10s
SERVER_HTTP_LISTEN=:8080
SERVER_PROFILE_LISTEN=:8083        # pprof + metrics port

# Social auth (all optional — only Facebook wired in main.go by default)
FACEBOOK_CLIENT_ID=...
FACEBOOK_CLIENT_SECRET=...
FACEBOOK_REDIRECT_URL=http://localhost:8080/auth/facebook/callback
```

---

## Running Locally

```bash
cd example/api

# 1. Copy and edit env
cp deploy/develop/.api.env .env
# edit .env — set DB connection string + OAUTH2_SECRET at minimum

# 2. Build + start (Docker, includes migrations)
make run-api

# Endpoints:
# GraphQL playground  http://localhost:8581/
# Prometheus metrics  http://localhost:8581/metrics
# pprof profiler      http://localhost:8583/debug/pprof/
```

---

## Adding a Custom Schema Extension

Place project-specific GraphQL extensions in `protocol/graphql/extensions/`.
Example — adding a `notes` field to `User`:

```graphql
# protocol/graphql/extensions/extend_user.graphql
extend type User {
  notes: String
}
```

Map the field to a Go resolver or model binding in `gqlgen.yml` if needed, then regenerate:

```bash
make build-gql
```

---

## Regenerating GraphQL Code

```bash
cd example/api
make build-gql
# equivalent: cd protocol/graphql && go run github.com/99designs/gqlgen
```

Generated files — do **not** edit manually:
github.com/99designs/gqlgen

```

Generated files — do **not** edit manually:

- `internal/server/graphql/generated/exec.go`
- `internal/server/graphql/models/generated.go`
``
- `internal/server/graphql/models/generated.go`
