# blaze-api

[![Tests](https://github.com/geniusrabbit/blaze-api/actions/workflows/tests.yml/badge.svg)](https://github.com/geniusrabbit/blaze-api/actions?workflow=Tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/geniusrabbit/blaze-api)](https://goreportcard.com/report/github.com/geniusrabbit/blaze-api)
[![Coverage Status](https://coveralls.io/repos/github/geniusrabbit/blaze-api/badge.svg?branch=main)](https://coveralls.io/github/geniusrabbit/blaze-api?branch=main)

Blaze-API is a foundational template for building and deploying APIs in Go. This template is designed to provide a basic structure for creating both RESTful and GraphQL APIs. It includes essential features such as user management, account handling, role-based access control (RBAC), and JWT authentication.

- [x] Users: Manage user data and interactions.
- [x] Accounts: Handle account operations and storage.
- [x] Roles: Implement Role-Based Access Control (RBAC) for managing user permissions.
- [x] Permissions: Define and manage access rights for different roles.
- [x] JWT Authentication: Secure your API with JWT-based authentication.
- [x] GraphQL API: Integrated GraphQL support for building flexible APIs.
- [ ] REST API: RESTful API interface for your application.
- [ ] Swagger API documentation: Generate comprehensive API documentation with Swagger.
- [x] Tests: Comprehensive test suite for maintaining code quality.
- [x] Logging: Robust logging for monitoring and debugging.

## Quick Start

### Installation

To install Blaze-API, run the following command:

```bash
go get github.com/geniusrabbit/blaze-api
```

### Example Usage

Below is a brief example to get you started with Blaze-API.

```go
// @see example/api/main.go
package main

import (
  ...
  "github.com/geniusrabbit/blaze-api/context/ctxlogger"
  "github.com/geniusrabbit/blaze-api/context/permissionmanager"
  "github.com/geniusrabbit/blaze-api/database"
  "github.com/geniusrabbit/blaze-api/middleware"
  "github.com/geniusrabbit/blaze-api/permissions"
  "github.com/geniusrabbit/blaze-api/profiler"
  "github.com/geniusrabbit/blaze-api/repository/historylog/middleware/gormlog"
)

func main() {
  ...

  // Register callback for history log (only for modifications)
  gormlog.Register(masterDatabase)

  // Init permission manager
  permissionManager := permissions.NewManager(masterDatabase, conf.Permissions.RoleCacheLifetime)
  appinit.InitModelPermissions(permissionManager)

  // Init OAuth2 provider
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

### Extend existing GraphQL API

To extend the existing GraphQL API, you need to create a new schema file in the `src/graphql/schemas` folder. Then, you need to add the new schema file to the `gqlgen.yml` configuration file. Finally, you need to run the `go generate` command to generate the new GraphQL API.

```yaml
# Generate the server inside the folder
# > go run github.com/99designs/gqlgen

# Where are all the schema files located? globs are supported eg  src/**/*.graphqls
schema:
  - ./schemas/*.graphql
  - ../../vendor/github.com/geniusrabbit/blaze-api/protocol/graphql/schemas/*.graphql

skip_mod_tidy: yes

# Where should the generated server code go?
exec:
  filename: ../../internal/server/graphql/generated/exec.go
  package: generated

# federation:
#   filename: ../../lib/server/graphql/generated/federation.go
#   package: generated

model:
  filename: ../../internal/server/graphql/models/generated.go
  package: models

# Where should the resolver implementations go?
resolver:
  layout: follow-schema
  dir: ../../internal/server/graphql/resolvers
  package: resolvers

# Optional: turn on use ` + "`" + `gqlgen:"fieldName"` + "`" + ` tags in your models
# struct_tag: json

# Optional: turn on to use []Thing instead of []*Thing
omit_slice_element_pointers: false

# Optional: set to speed up generation time by not performing a final validation pass.
skip_validation: true

# gqlgen will search for any type names in the schema in these go packages
# if they match it will use them, otherwise it will generate them.
autobind:
  - github.com/geniusrabbit/blaze-api/server/graphql/models

# This section declares type mapping between the GraphQL and go type systems
#
# The first line in each type will be used as defaults for resolver arguments and
# modelgen, the others will be allowed when binding to fields. Configure them to
# your liking
models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  Int:
    model:
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
  # Connectors
  UserConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.UserConnection
  AccountConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.AccountConnection
  MemberConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.MemberConnection
  RBACRoleConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.RBACRoleConnection
  AuthClientConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.AuthClientConnection
  HistoryActionConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.HistoryActionConnection
  OptionConnection:
    model: github.com/geniusrabbit/blaze-api/server/graphql/connectors.OptionConnection
```

## TODO features

- [ ] OAuth2 add authorization providers: Google, Facebook, LinkedIn, GitHub, etc.
- [x] OAuth2 remote authorization.
- [ ] REST API: RESTful API interface for your application.
- [ ] Swagger API documentation: Generate comprehensive API documentation with Swagger.
- [x] GraphQL API: Integrated GraphQL support for building flexible APIs.
- [x] Support different databases (PostgreSQL, MySQL, SQLite, etc.)
- [x] OAuth2 server and client support.
- [x] RBAC: Role-Based Access Control (RBAC) for managing user permissions.
- [x] JWT Authentication: Secure your API with JWT-based authentication.
- [x] Object modification history log.
- [x] Profiler: Integrated profiler for monitoring and debugging.
- [x] Mailer/messanger support (abstract interface layer).
- [ ] Add support [OpenTelemetry-Go](https://github.com/open-telemetry/opentelemetry-go/)
