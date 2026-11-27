# blaze-api

[![Tests](https://github.com/geniusrabbit/blaze-api/actions/workflows/tests.yml/badge.svg)](https://github.com/geniusrabbit/blaze-api/actions?workflow=Tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/geniusrabbit/blaze-api)](https://goreportcard.com/report/github.com/geniusrabbit/blaze-api)
[![Coverage Status](https://coveralls.io/repos/github/geniusrabbit/blaze-api/badge.svg?branch=main)](https://coveralls.io/github/geniusrabbit/blaze-api?branch=main)

Blaze-API is a foundational template for building and deploying APIs in Go. It provides a production-ready structure for creating GraphQL APIs with user management, account handling, role-based access control (RBAC), OAuth2, and JWT authentication.

**Features**

- [x] Users: Manage user data and interactions.
- [x] Accounts: Handle account operations and storage.
- [x] Roles: Role-Based Access Control (RBAC) for managing user permissions.
- [x] Permissions: Define and manage access rights for different roles.
- [x] JWT Authentication: Secure your API with JWT-based authentication.
- [x] GraphQL API: Integrated GraphQL support for building flexible APIs.
- [x] OAuth2: Server and client support with remote authorization.
- [x] Social auth: Facebook OAuth2 login out of the box (Google, LinkedIn, X.com ready to configure).
- [x] Object history log: Track all mutations with a per-request message.
- [x] Auth clients: OAuth2 client management (token issuance, revocation).
- [x] Direct access tokens: Long-lived tokens for service-to-service auth.
- [x] Generic repository/usecase layer: Type-safe CRUD with compile-time model constraints.
- [x] Tests: Comprehensive test suite for maintaining code quality.
- [x] Logging: Structured logging (Zap) with context propagation.
- [x] Profiler & metrics: pprof + Prometheus endpoints built in.
- [ ] REST API: RESTful API interface for your application.
- [ ] Swagger API documentation: Generate comprehensive API documentation with Swagger.

## Quick Start

### Installation

```bash
go get github.com/geniusrabbit/blaze-api
```

### Run the example locally (Docker)

```bash
cd example/api

# Local settings: copy the template and fill in the secrets (.env is git-ignored)
cp .env.example .env

# Start Postgres + run migrations + start the API
make run-api

# API (GraphQL):      http://localhost:8581/graphql
# GraphQL playground: http://localhost:8581/playground (basic auth: GRAPHQL_USERS_<name>=<password>)
# Liveness/readiness: http://localhost:8581/healthcheck, http://localhost:8581/readiness
# Prometheus metrics: http://localhost:8581/metrics
# pprof profiler:     http://localhost:8583/debug/pprof/ (dev image only; localhost:6060 by default)
```

The `run-api` target builds the Docker image and starts the full stack via docker-compose. The API applies the SQL migrations itself on start (`MIGRATE_ON_START=true`, from `MIGRATIONS_DIR`, default `/data/migrations`). A migration that failed half-way leaves the database "dirty": the API then refuses to start until the schema is fixed and the version is marked with `migrate force <version>`.

For an alternative, schema-as-HCL migration workflow based on [Atlas](https://atlasgo.io) (opt-in, additional to the SQL migrations above), see [docs/MIGRATIONS.md](docs/MIGRATIONS.md).

### Configuration

All settings are read from environment variables (or a `.env` file). The key ones:

```bash
# Database (PostgreSQL)
SYSTEM_STORAGE_DATABASE_MASTER_CONNECT=postgres://dbuser:password@localhost:5432/project?sslmode=disable
SYSTEM_STORAGE_DATABASE_SLAVE_CONNECT=postgres://dbuser:password@localhost:5432/project?sslmode=disable

# OAuth2 / JWT
OAUTH2_SECRET=your-secret-min-32-chars
OAUTH2_ACCESS_TOKEN_LIFESPAN=1h
OAUTH2_REFRESH_TOKEN_LIFESPAN=720h

# Session
SESSION_COOKIE_NAME=sessid
SESSION_LIFETIME=1h

# HTTP server limits (zero read/write timeouts keep websocket subscriptions alive)
SERVER_HTTP_READ_HEADER_TIMEOUT=10s
SERVER_HTTP_IDLE_TIMEOUT=120s
SERVER_HTTP_SHUTDOWN_TIMEOUT=30s
SERVER_HTTP_MAX_BODY_SIZE=10485760
SERVER_HTTP_CORS_ORIGINS=https://app.example.com,https://*.example.com

# GraphQL query limits; introspection is off unless enabled
GRAPHQL_MAX_COMPLEXITY=1000
GRAPHQL_MAX_DEPTH=12
GRAPHQL_MAX_ALIASES=30
GRAPHQL_INTROSPECTION=false

# Dev mode (skip auth with a static token). Off by default; never in production.
DEBUG=true
LOG_LEVEL=debug
SESSION_DEV_TOKEN=develop
SESSION_DEV_USER_ID=1
SESSION_DEV_ACCOUNT_ID=1

# Social auth (optional)
FACEBOOK_CLIENT_ID=...
FACEBOOK_CLIENT_SECRET=...
FACEBOOK_REDIRECT_URL=http://localhost:8581/auth/facebook/callback
```

A full annotated example lives in [example/api/.env.example](example/api/.env.example) and [example/api/deploy/develop/.api.env](example/api/deploy/develop/.api.env).

### Wiring it together (`main.go`)

A condensed version of [example/api/cmd/api/main.go](example/api/cmd/api/main.go); see it for the full wiring.

```go
func main() {
  conf := &appcontext.Config // loaded with goconfig from env / files
  logger := initZapLogger()
  ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer cancel()

  // Apply SQL migrations (fails on a dirty database instead of resetting it)
  if conf.MigrateOnStart {
    fatalError(migrate(ctx, conf), "migrate database")
  }

  // pprof/metrics on a separate listener; stop it on exit
  stopProfiler := profiler.Run(conf.Server.Profile.Mode, conf.Server.Profile.Listen, logger, true)
  defer stopProfiler()

  masterDB, slaveDB, err := database.ConnectMasterSlave(ctx,
    conf.System.Storage.MasterConnect, conf.System.Storage.SlaveConnect)
  fatalError(err, "connect to database")
  fatalError(gormlog.Register(masterDB), "register history log") // history rows in the same transaction

  deps := appinit.NewDeps() // repositories, usecases, auth loader
  permissionManager := permissions.NewManager(masterDB, conf.Permissions.RoleCacheLifetime)
  appinit.InitModelPermissions(permissionManager, deps)

  // OAuth2 server + JWT provider (the JWT secret must be at least 32 bytes)
  oauth2provider, jwtProvider := appinit.Auth(ctx, conf, masterDB, deps)

  httpServer := server.HTTPServer{
    Config: server.Config{ // timeouts, body limit, CORS allowlist, GraphQL limits
      ReadHeaderTimeout: conf.Server.HTTP.ReadHeaderTimeout,
      ShutdownTimeout:   conf.Server.HTTP.ShutdownTimeout,
      MaxBodySize:       conf.Server.HTTP.MaxBodySize,
      CORS:              middleware.CORSConfig{AllowedOrigins: conf.CORSOrigins(), AllowCredentials: true},
      GraphQLLimits:     limits.Config{Introspection: conf.Server.GraphQL.Introspection},
    },
    ReadinessChecks: []profiler.Check{profiler.DatabaseCheck("master", masterDB)},
    Logger:          logger,
    JWTProvider:     jwtProvider,
    AuthLoader:      deps.AuthLoader,
    Authorizers: []auth.Authorizer[*domain.User, *domain.Account]{
      jwt.NewAuthorizer(jwtProvider, deps.AuthLoader),
      oauth2.NewAuthorizer(oauth2provider, deps.AccountRepo),
      authorizer.NewDevTokenAuthorizer(devTokenOptions, deps.AuthLoader), // only with DEBUG=true
    },
    ContextWrap: func(ctx context.Context) context.Context {
      ctx = ctxlogger.WithLogger(ctx, logger)
      ctx = database.WithDatabase(ctx, masterDB, slaveDB)
      return permissions.WithManager(ctx, permissionManager)
    },
  }
  fatalError(httpServer.Run(ctx, conf.Server.HTTP.Listen), "HTTP server") // graceful shutdown on ctx cancel
}
```

### Security defaults

The framework is meant to be safe by default; most of these can be tuned, few should be turned off.

| Area | Default |
|---|---|
| JWT | HS256 only, `exp` required, `iss`/`aud` checked when set, secret ≥ 32 bytes; token from the `Authorization` header only (`tokenextractor.HeaderOrWebsocketQuery` allows `?access_token=` on websocket handshakes); optional revocation (`jwt.NewCacheRevocationStore`, used by `logout`) |
| GraphQL | complexity ≤ 1000, depth ≤ 12, aliases ≤ 30, introspection off (`server/graphql/limits`); page size ≤ `repository.MaxPageSize` (1000); `@cacheData` keys are per user and account and the directive can't be set from queries |
| HTTP | header timeout, idle timeout, body size limit, panic recovery, graceful shutdown; CORS only for `SERVER_HTTP_CORS_ORIGINS` (`middleware.CORS`) |
| Sessions | `auth.cross.account` (`AccountID[:UserID]`) needs the `auth.cross` permission for any switch; API keys can't switch accounts, create keys or change passwords |
| Social login | signed, browser-bound OAuth2 `state`; `redirect` only to allowed URLs (`rest.WithAllowedRedirectURLs`); linking to the logged-in user requires the `Authorization` header; provider tokens are never returned by GraphQL |
| Passwords | bcrypt over HMAC-SHA256, any length; reset tokens from `crypto/rand`; login and reset don't reveal whether an email is registered |
| Debug | off unless `DEBUG=true`; the config dump masks secrets; tokens are never logged |

### Registering permissions for a domain model

Every model that should participate in ACL must be registered with the permission manager:

```go
// example/api/cmd/api/appinit/acl.go
func InitModelPermissions(pm *permissions.Manager, deps *Deps) {
  acl.InitModelPermissions(pm,
    &user.User{},
    &rbacModels.Role{},
    &authclient.AuthClient{},
    &account.Account{},
    &historylog.HistoryAction{},
    // ... add your own models here
  )

  // Standard CRUD permissions
  _ = pm.RegisterNewOwningPermissions(&user.User{},
    []string{acl.PermView, acl.PermList, acl.PermCreate, acl.PermUpdate, acl.PermDelete})

  // With approval workflow
  _ = pm.RegisterNewOwningPermissions(&account.Account{},
    append(crudPermissions, acl.PermApprove, acl.PermReject),
    rbac.WithCustomCheck(accountCustomCheck))
}
```

## User / Account / Member templates

User, Account, and Member are **composable embeddable bases + optional traits**, not monolithic structs. User remains bundled in the library; Account and Member are wired by the consumer (see `example/api/internal/domain/account.go`).

| Layer | Bundled / reference type | Compose from |
| ----- | ------------------------ | ------------ |
| User | `example/api/internal/domain.User` | `UserBase` + `UserEmail` + `UserPassword` |
| Account | `example/api/internal/domain.Account` | `AccountBase` + consumer `AccountProfile` (see `domain/account_profile.go`) |
| Member | `account.Member[TUser, TAccount]` | `MemberBase` + preload `User`/`Account` |

**`account.Model`** covers core tenant fields only (`GetID`, permissions, `GetApprove`, `GetCreatedAt`, `GetUpdatedAt`). Profile fields (title, logo, …) live in the **consumer GraphQL schema** via `extend type` / `extend input`, not in the base library schema.

Base GraphQL types in `repository/account/delivery/graphql/account_base.graphql` and `repository/user/delivery/graphql/user.graphql` contain **minimal core fields only**. Consumer apps add profile/filter fields in separate extension files (see `example/api/protocol/graphql/extensions/`).

Account and User **QueryResolver** implementations are **generics** parameterized by consumer GraphQL types:

```go
// repository/account/delivery/graphql/base_resolver.go
type QueryResolver[TUser, TDomain, TGQLAccount, TGQLAccountInput, ...] struct { ... }

// Base wiring (library default types):
accountgraphql.NewBaseQueryResolver(cfg)   // TGQLAccount = *gqlmodels.Account

// Example consumer (extended Account):
wiring.NewExampleAccountQueryResolver(cfg) // TGQLAccount = exmodels.Account
```

Wire extended resolvers in `example/api/internal/server/graphql/wiring/` and map domain → GraphQL in `example/api/internal/domain/graphql.go`:

```go
// example/api/internal/domain/graphql.go
func AccountToGraphQL(acc *Account) exmodels.Account { /* core + profile fields */ }
func FillAccountFromInput(dest *Account, input *exmodels.AccountInput, ...) *Account { /* ... */ }

// example/api/cmd/api/appinit/deps.go
ExampleResolverDeps = resolvers.Deps[*UserType, *AccountType, exmodels.Account, *exmodels.AccountInput]
```

`RBACResourceName()` / `TableName()` stay on concrete structs (ACL/GORM), not on `account.Model`.

**Reference consumer wiring** — see [`example/api/cmd/api/appinit/deps.go`](example/api/cmd/api/appinit/deps.go) and [`example/api/internal/user/stack.go`](example/api/internal/user/stack.go):

```go
deps := appinit.NewDeps()
// deps.UserModule  — composite Repository + Core/Email/Password usecases (example-only helper)
// deps.Resolver    — passed to server/graphql/resolvers
```

The framework exposes **separate** interfaces only: `user.Repository`, `user.EmailRepository`, `user.PasswordRepository`, and matching usecases. Compose them in the consumer (example uses `userstack.NewModule`).

`Deps[TUser, TAccount]` is parameterized by **your** user/account types. GraphQL wiring uses `UserModuleDeps` (core + email + password + repos) and `BaseDeps` (account/member). Auth resolvers take `EmailRepository` + `PasswordRepository` separately — not a bundled repo type.

```go
baseResolver.UserModuleDeps[*MyUser]{
    NewUser:      func() *MyUser { return new(MyUser) },
    Core:         userCoreUC,
    Email:        userEmailUC,
    Password:     userPasswordUC,
    EmailRepo:    userEmailRepo,
    PasswordRepo: userPasswordRepo,
}
```

### Custom User type (consumer)

```go
type User struct {
    userModels.UserBase
    userModels.UserEmail
    userModels.UserPassword
    Country string `gorm:"column:country"`
}

func (u *User) TableName() string       { return "adnet_user" }
func (u *User) RBACResourceName() string { return "adnet.user" }
```

Wire once at startup (example pattern):

```go
newModel := func() *User { return new(User) }
userModule := userstack.NewModule(newModel) // example/api only
userRepo := userModule.Repo                 // satisfies Repository + Email + Password
userCoreUC := userModule.Core
```

**Minimal UserBase-only** deployments (no email/password) require a custom migration and must not use `PasswordRepository` / password GraphQL modules — compile-time generic constraints enforce this.

### Adnetapi migration

When migrating an existing consumer (e.g. adnetapi):

1. Replace `type User = blzuser.User` with a local struct embedding `UserBase` + traits + custom fields.
2. Replace legacy user repository wiring with `NewDefaultUserRepository()` or `NewBundledRepository(newModel)`.
3. Pass `server/graphql/resolvers.Deps` (or legacy interfaces) into resolvers — do not construct repos inside resolver packages.
4. Update struct literals: promoted embed fields cannot be set in composite literals — use stubs (`UserStub(id)`) or assign after `&User{}`.
5. ACL registration stays the same — register your concrete `&User{}` / `&Account{}` types.

## Architecture

### Repository / Usecase layer (`repository/generated`)

All domain entities follow the same layered pattern:

```
repository/<domain>/
  models/        — domain structs (must implement generated.Model[TID])
  repository.go  — domain Repository/Usecase interface
  repository/    — GORM implementation (embeds generated.Repository[T, TID])
  usecase/       — business logic (embeds generated.Usecase[T, TID])
  mocks/         — generated mocks (go:generate mockgen, DO NOT EDIT)
  delivery/      — transport adapters (GraphQL resolvers, REST handlers)
```

The generic base types live in `repository/generated`:

| Type                      | Description                                                         |
| ------------------------- | ------------------------------------------------------------------- |
| `Repository[T, TID]`      | GORM CRUD implementation for any model satisfying `Model[TID]`      |
| `Usecase[T, TID]`         | ACL-checked business logic delegating to `RepositoryIface[T, TID]`  |
| `UsecaseApprover[T, TID]` | Approve/reject workflow with ACL checks                             |
| `BaseModel[TID]`          | Convenience embed — provides `GetID`/`SetID` for free               |
| `BaseTimestamps`          | Convenience embed — provides `SetCreatedAt`/`SetUpdatedAt` for free |

#### Defining a new domain model

A model type `T` must satisfy the `generated.Model[TID]` constraint — i.e., expose `GetID() TID` via a **value receiver**. The easiest way is to embed `generated.BaseModel`:

```go
import "github.com/geniusrabbit/blaze-api/repository/generated"

type Widget struct {
    generated.BaseModel[uint64]  // GetID() + SetID() for free
    generated.BaseTimestamps     // SetCreatedAt() + SetUpdatedAt() for free
    gorm.DeletedAt

    Name string
}

func (w *Widget) TableName() string        { return "widget" }
func (w *Widget) RBACResourceName() string { return "widget" }
```

Then create the repository and usecase:

```go
// repository/widget/repository/repository.go
type Repository struct {
    generated.Repository[widget.Widget, uint64]
}

func New() *Repository {
    return &Repository{Repository: *generated.NewRepository[widget.Widget, uint64]()}
}

// repository/widget/usecase/usecase.go
type Usecase struct {
    generated.Usecase[widget.Widget, uint64]
}

func New(repo widget.Repository) *Usecase {
    return &Usecase{Usecase: generated.Usecase[widget.Widget, uint64]{Repo: repo}}
}
```

If your model already has `ID`, `CreatedAt`, `UpdatedAt` fields but no embeds, add the methods explicitly (value receiver required for `GetID`):

```go
func (m Widget) GetID() uint64             { return m.ID }
func (m *Widget) SetID(id uint64)          { m.ID = id }
func (m *Widget) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *Widget) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }
```

### Query options (`repository.QOption`)

All mutation and query methods accept `...QOption` instead of positional parameters. Options compose freely:

```go
type QOption interface {
    PrepareQuery(query *gorm.DB) *gorm.DB
}
```

Built-in options:

| Option                                                      | Package                 | Effect                                                                        |
| ----------------------------------------------------------- | ----------------------- | ----------------------------------------------------------------------------- |
| `historylog.Message("reason")`                              | `repository/historylog` | Attaches a human-readable message to the mutation recorded in the history log |
| `&repository.PreloadOption{Fields: []string{"ChildRoles"}}` | `repository`            | Adds GORM `.Preload(...)` calls                                               |
| `filter` (`*Filter` implementing `QOption`)                 | domain package          | Adds WHERE conditions                                                         |
| `order` (`*Order` implementing `QOption`)                   | domain package          | Adds ORDER BY                                                                 |

Example:

```go
id, err := roleRepo.Create(ctx, role, historylog.Message("initial seed"))

err = roleRepo.Delete(ctx, id, historylog.Message("cleanup"))
```

### History log

Every write that goes through a GORM master connection registered with `gormlog.Register(db)` records a `HistoryAction` row. The optional `historylog.Message(msg)` option attaches a human-readable reason:

```go
gormlog.Register(masterDatabase)

// in a usecase or resolver:
repo.Delete(ctx, id, historylog.Message("user requested account deletion"))
```

### Mock generation

Mocks are generated with [mockgen](https://github.com/uber-go/mock) and committed as source code. **Never edit mock files by hand** — regenerate them:

```bash
make generate-code   # runs: go generate ./...
```

Each mock package carries the directive:

```go
//go:generate mockgen -source=../repository.go -destination=../mocks/repository.go
```

## Extending the GraphQL API

### Minimal base schema + `extend type` (recommended)

Base types ship with core fields only. Extend them in your app without forking library schemas:

```graphql
# example/api/protocol/graphql/extensions/extend_account.graphql
extend type Account {
  title: String!
  description: String!
  logoURI: String!
  # ...
}

extend input AccountInput { title: String }
extend input AccountListFilter { title: [String!] }
extend input AccountListOrder { title: Ordering }
```

Point gqlgen at base schemas **and** your extensions:

```yaml
# example/api/protocol/graphql/gqlgen.yml
schema:
  - ../../../../protocol/graphql/schemas/*.graphql
  - ../../../../repository/**/*.graphql
  - ./extensions/*.graphql
```

**Model binding rules:**

| Type | Bind to | Why |
| ---- | ------- | --- |
| Extended `Account`, `AccountInput`, filters | Generate in `example/.../models/` | Merged type from base + extend |
| Shared `User`, `SessionToken`, RBAC, … | `blaze-api/server/graphql/models` | Unchanged across consumers |
| `AccountConnection`, `UserConnection`, `MemberConnection` | `example/.../models/connections.go` | Uses extended node types |

Regenerate and wire generic resolvers:

```bash
cd example/api/protocol/graphql && go run github.com/99designs/gqlgen
cd example/api && make build
```

Resolver wiring lives in `example/api/internal/server/graphql/wiring/` (`NewExampleAccountQueryResolver`, `NewExampleUserQueryResolver`, `NewExampleMemberQueryResolver`).

### Adding new domain types

1. Add a `.graphql` schema file to `protocol/graphql/schemas/` (or your app's `schemas/` folder).
2. Point `gqlgen.yml` at the schemas — the example app uses:

```yaml
# example/api/protocol/graphql/gqlgen.yml
schema:
  - ../../../../protocol/graphql/schemas/*.graphql
  - ../../../../repository/**/*.graphql
```

1. Regenerate:

```bash
cd example/api && make build-gql   # runs: go run github.com/99designs/gqlgen
```

1. Implement the generated resolver stubs in `internal/server/graphql/resolvers/`.

```yaml
# example/api/protocol/graphql/gqlgen.yml (excerpt — see repo for full file)
schema:
  - ../../../../protocol/graphql/schemas/*.graphql
  - ../../../../repository/**/*.graphql
  - ./extensions/*.graphql

models:
  User:
    model: github.com/geniusrabbit/blaze-api/server/graphql/models.User
  AccountConnection:
    model: github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models.AccountConnection
  UserConnection:
    model: github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models.UserConnection
  MemberConnection:
    model: github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models.MemberConnection
  # ... other shared types bind to blaze-api/server/graphql/models
  # Account / AccountInput / filters: NOT bound — generated with extensions in example/models
```

## Development

```bash
# Run all tests
make test

# Run tests with race detector + coverage report (written to .build/tmp)
make cover

# Install the pinned tools (golangci-lint, mockgen) used by CI
make tools

# Check dependencies for known vulnerabilities
make vuln

# Regenerate mocks (go generate ./...)
make generate-code

# Regenerate GraphQL server code (gqlgen)
cd example/api && make build-gql

# Build the example API binary
cd example/api && make build

# Run full stack via Docker Compose (Postgres + API, which applies the migrations)
cd example/api && make run-api

# Lint
make lint
```

## TODO

- [ ] OAuth2 social providers: Google, LinkedIn, X.com (endpoints are already wired; need full handler)
- [ ] REST API interface
- [ ] Swagger / OpenAPI documentation
- [ ] OpenTelemetry tracing ([opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go/)) instead of OpenTracing/Jaeger/APM
- [ ] Move the example app into its own Go module, so its dependencies don't reach consumers
- [ ] Keep the Atlas HCL schemas in sync with the SQL migrations (CI drift check)
