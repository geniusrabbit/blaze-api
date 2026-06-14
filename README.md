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

# Start Postgres + run migrations + start the API
make run-api

# API is available at http://localhost:8581
# GraphQL playground: http://localhost:8581/
# Prometheus metrics: http://localhost:8581/metrics
# pprof profiler:     http://localhost:8583/debug/pprof/
```

The `run-api` target builds the Docker image, runs migrations, and starts the full stack via docker-compose.

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

# Dev mode (skip auth with a static token)
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

A full annotated example lives in [example/api/.env](example/api/.env) and [example/api/deploy/develop/.api.env](example/api/deploy/develop/.api.env).

### Wiring it together (`main.go`)

```go
// example/api/cmd/api/main.go
package main

import (
  "github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
  "github.com/geniusrabbit/blaze-api/pkg/database"
  "github.com/geniusrabbit/blaze-api/pkg/permissions"
  "github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
  "github.com/geniusrabbit/blaze-api/pkg/auth/oauth2"
  "github.com/geniusrabbit/blaze-api/repository/account/authorizer"
  "github.com/geniusrabbit/blaze-api/repository/historylog/middleware/gormlog"
)

func main() {
  // Connect master + slave databases
  masterDB, slaveDB, _ := database.ConnectMasterSlave(ctx,
    conf.System.Storage.MasterConnect,
    conf.System.Storage.SlaveConnect)

  // Register GORM callback — writes a HistoryAction row for every mutation
  gormlog.Register(masterDB)

  // Build permission manager (RBAC, cached)
  permissionManager := permissions.NewManager(masterDB, conf.Permissions.RoleCacheLifetime)
  appinit.InitModelPermissions(permissionManager)  // register all domain models

  // Build OAuth2 + JWT providers
  oauth2provider, jwtProvider := appinit.Auth(ctx, conf, masterDB)

  // Attach services to context (propagated to every request handler)
  ctx = ctxlogger.WithLogger(ctx, logger)
  ctx = database.WithDatabase(ctx, masterDB, slaveDB)
  ctx = permissions.WithManager(ctx, permissionManager)

  httpServer := server.HTTPServer{
    Logger:      logger,
    JWTProvider: jwtProvider,
    Authorizers: []auth.Authorizer[*user.User, *account.Account]{
      jwt.NewAuthorizer(jwtProvider),
      oauth2.NewAuthorizer(oauth2provider),
      authorizer.NewDevTokenAuthorizer(...), // dev-only static token
    },
    ContextWrap: func(ctx context.Context) context.Context {
      ctx = ctxlogger.WithLogger(ctx, logger)
      ctx = database.WithDatabase(ctx, masterDB, slaveDB)
      ctx = permissions.WithManager(ctx, permissionManager)
      return ctx
    },
  }
  httpServer.Run(ctx, conf.Server.HTTP.Listen)
}
```

### Registering permissions for a domain model

Every model that should participate in ACL must be registered with the permission manager:

```go
// example/api/cmd/api/appinit/acl.go
func InitModelPermissions(pm *permissions.Manager) {
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

A complete `gqlgen.yml` that maps blaze-api connection types:

```yaml
schema:
  - ../../../../protocol/graphql/schemas/*.graphql
  - ../../../../repository/**/*.graphql

skip_mod_tidy: true

exec:
  filename: ../../internal/server/graphql/generated/exec.go
  package: generated

model:
  filename: ../../internal/server/graphql/models/generated.go
  package: models

resolver:
  layout: follow-schema
  dir: ../../internal/server/graphql/resolvers
  package: resolvers

omit_slice_element_pointers: false
skip_validation: true

autobind:
  - github.com/geniusrabbit/blaze-api/server/graphql/models

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  Int64:
    model:
      - github.com/99designs/gqlgen/graphql.Int64
  Time:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.Time
  TimeDuration:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.TimeDuration
  DateTime:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.DateTime
  JSON:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.JSON
  NullableJSON:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.NullableJSON
  UUID:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.UUID
  ID64:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.ID64
  # Connection types — use blaze-api's built-in implementations
  UserConnection:
    model: github.com/geniusrabbit/blaze-api/repository/user/delivery/graphql.UserConnection
  AccountConnection:
    model: github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql.AccountConnection
  MemberConnection:
    model: github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql.MemberConnection
  RBACRoleConnection:
    model: github.com/geniusrabbit/blaze-api/repository/rbac/delivery/graphql.RBACRoleConnection
  AuthClientConnection:
    model: github.com/geniusrabbit/blaze-api/repository/authclient/delivery/graphql.AuthClientConnection
  HistoryActionConnection:
    model: github.com/geniusrabbit/blaze-api/repository/historylog/delivery/graphql.HistoryActionConnection
  OptionConnection:
    model: github.com/geniusrabbit/blaze-api/repository/option/delivery/graphql.OptionConnection
  DirectAccessTokenConnection:
    model: github.com/geniusrabbit/blaze-api/repository/directaccesstoken/delivery/graphql.DirectAccessTokenConnection
```

## Development

```bash
# Run all tests
make test

# Run tests with race detector + coverage report
make cover

# Regenerate mocks (go generate ./...)
make generate-code

# Regenerate GraphQL server code (gqlgen)
cd example/api && make build-gql

# Build the example API binary
cd example/api && make build-api

# Run full stack via Docker Compose (Postgres + migrations + API)
cd example/api && make run-api

# Lint
make lint
```

## TODO

- [ ] OAuth2 social providers: Google, LinkedIn, X.com (endpoints are already wired; need full handler)
- [ ] REST API interface
- [ ] Swagger / OpenAPI documentation
- [ ] OpenTelemetry tracing ([opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go/))

**Features**

- [x] Users: Manage user data and interactions.
- [x] Accounts: Handle account operations and storage.
- [x] Roles: Role-Based Access Control (RBAC) for managing user permissions.
- [x] Permissions: Define and manage access rights for different roles.
- [x] JWT Authentication: Secure your API with JWT-based authentication.
- [x] GraphQL API: Integrated GraphQL support for building flexible APIs.
- [x] OAuth2: Server and client support with remote authorization.
- [x] Object history log: Track all mutations with a per-request message.
- [x] Auth clients: OAuth2 client management (token issuance, revocation).
- [x] Generic repository/usecase layer: Type-safe CRUD with compile-time model constraints.
- [x] Tests: Comprehensive test suite for maintaining code quality.
- [x] Logging: Structured logging with context propagation.
- [ ] REST API: RESTful API interface for your application.
- [ ] Swagger API documentation: Generate comprehensive API documentation with Swagger.

## Quick Start

### Installation

```bash
go get github.com/geniusrabbit/blaze-api
```

### Example Usage

```go
// @see example/api/cmd/api/main.go
package main

import (
  ...
  "github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
  "github.com/geniusrabbit/blaze-api/pkg/permissions"
  "github.com/geniusrabbit/blaze-api/pkg/database"
  "github.com/geniusrabbit/blaze-api/pkg/middleware"
  "github.com/geniusrabbit/blaze-api/repository/historylog/middleware/gormlog"
)

func main() {
  // Register callback for history log (only for modifications)
  gormlog.Register(masterDatabase)

  // Init permission manager
  permissionManager := permissions.NewManager(masterDatabase, conf.Permissions.RoleCacheLifetime)
  appinit.InitModelPermissions(permissionManager)

  // Init OAuth2 + JWT providers
  oauth2provider, jwtProvider := appinit.Auth(ctx, conf, masterDatabase)

  // Init HTTP server
  httpServer := server.HTTPServer{
    OAuth2provider: oauth2provider,
    JWTProvider:    jwtProvider,
    SessionManager: appinit.SessionManager("session", 60*time.Minute),
    AuthOption: gocast.IfThen(conf.IsDebug(), &middleware.AuthOption{
      DevToken:     conf.Session.DevToken,
      DevUserID:    conf.Session.DevUserID,
      DevAccountID: conf.Session.DevAccountID,
    }, nil),
    ContextWrap: func(ctx context.Context) context.Context {
      ctx = ctxlogger.WithLogger(ctx, loggerObj)
      ctx = database.WithDatabase(ctx, masterDatabase, slaveDatabase)
      ctx = permissionmanager.WithManager(ctx, permissionManager)
      return ctx
    },
  }
  httpServer.Run(ctx, ":8080")
}
```

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
```

Then create the repository:

```go
type Repository struct {
    generated.Repository[Widget, uint64]
}

func New() *Repository {
    return &Repository{Repository: *generated.NewRepository[Widget, uint64]()}
}
```

If your model already has `ID`, `CreatedAt`, `UpdatedAt` fields but no embeds, add the methods explicitly (value receiver required for `GetID`):

```go
func (m Widget) GetID() uint64          { return m.ID }
func (m *Widget) SetID(id uint64)       { m.ID = id }
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

err = roleRepo.Delete(ctx, id,
    historylog.Message("cleanup"),
)
```

### History log

Every write that goes through a GORM master connection registered with `gormlog.Register(db)` records a `HistoryAction` row. The optional `historylog.Message(msg)` option attaches a human-readable reason:

```go
gormlog.Register(masterDatabase)

// later, in a usecase or resolver:
repo.Delete(ctx, id, historylog.Message("user requested account deletion"))
```

### Mock generation

Mocks are generated with [mockgen](https://github.com/uber-go/mock) and committed as source code. **Never edit mock files by hand** — regenerate them:

```bash
make generate-code   # runs: go generate ./...
```

Each mock package carries the directive:

Each mock package carries the directive:

```go
//go:generate mockgen -source=../repository.go -destination=../mocks/repository.go
```

## Extending the GraphQL API

1. Add a schema file to `protocol/graphql/schemas/` (or your app's `schemas/` folder).
2. Reference it in `gqlgen.yml`:

```yaml
schema:
  - ./schemas/*.graphql
  - ../../vendor/github.com/geniusrabbit/blaze-api/protocol/graphql/schemas/*.graphql
  - ../../vendor/github.com/geniusrabbit/blaze-api/repository/**/*.graphql
```

1. Regenerate the server code:

```bash
make build-gql   # runs: go run github.com/99designs/gqlgen
```

1. Implement the generated resolver stubs in `internal/server/graphql/resolvers/`.

A minimal `gqlgen.yml` for an application that imports blaze-api:

```yaml
schema:
  - ./schemas/*.graphql
  - ../../vendor/github.com/geniusrabbit/blaze-api/protocol/graphql/schemas/*.graphql
  - ../../vendor/github.com/geniusrabbit/blaze-api/repository/**/*.graphql

skip_mod_tidy: yes

exec:
  filename: ../../internal/server/graphql/generated/exec.go
  package: generated

model:
  filename: ../../internal/server/graphql/models/generated.go
  package: models

resolver:
  layout: follow-schema
  dir: ../../internal/server/graphql/resolvers
  package: resolvers

omit_slice_element_pointers: false
skip_validation: true

autobind:
  - github.com/geniusrabbit/blaze-api/server/graphql/models

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int64
  Int64:
    model:
      - github.com/99designs/gqlgen/graphql.Int64
  Time:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.Time
  JSON:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.JSON
  NullableJSON:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.NullableJSON
  UUID:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.UUID
  ID64:
    model: github.com/geniusrabbit/blaze-api/server/graphql/types.ID64
  UserConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.UserConnection
  AccountConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.AccountConnection
  RBACRoleConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.RBACRoleConnection
  AuthClientConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.AuthClientConnection
  HistoryActionConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.HistoryActionConnection
  OptionConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.OptionConnection
```

## Development

```bash
# Run all tests
make test

# Run tests with coverage report
make cover

# Regenerate mocks and gqlgen code
make generate-code

# Build the example API
cd example/api && make build-api

# Lint
make lint
```

## TODO

- [ ] OAuth2 social providers: Google, Facebook, LinkedIn, GitHub
- [ ] REST API interface
- [ ] Swagger / OpenAPI documentation
- [ ] OpenTelemetry tracing ([opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go/))
