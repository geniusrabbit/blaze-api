package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/goconfig"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/appcontext"
	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/appinit"
	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/migratedb"
	"github.com/geniusrabbit/blaze-api/example/api/internal/server"
	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/auth/devtoken"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin/facebook"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/auth/oauth2"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/version"
	"github.com/geniusrabbit/blaze-api/pkg/database"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/pkg/profiler"
	"github.com/geniusrabbit/blaze-api/pkg/zlogger"
	"github.com/geniusrabbit/blaze-api/repository/historylog/middleware/gormlog"
	"github.com/geniusrabbit/blaze-api/repository/socialauth/delivery/rest"
)

var (
	buildDate    = ""
	buildCommit  = ""
	buildVersion = "develop"
)

func init() {
	conf := &appcontext.Config
	fatalError(goconfig.Load(conf), "load config:")

	if conf.IsDebug() {
		fmt.Println(conf)
	}

	// Migrate database schemas
	fatalError(migratedb.Migrate(conf.System.Storage.MasterConnect, []migratedb.MigrateSource{
		{
			URI:                   []string{"file:///data/migrations/initial"},
			SchemaMigrationsTable: "schema_migrations_prod",
		},
		{
			URI:                   []string{"file:///data/migrations/fixtures"},
			SchemaMigrationsTable: "schema_migrations_dev",
		},
	}), "migrate database")
}

func initZapLogger() *zap.Logger {
	conf := &appcontext.Config
	loggerObj, err := zlogger.New(conf.ServiceName, conf.LogEncoder,
		conf.LogLevel, conf.LogAddr, zap.Fields(
			zap.String("commit", buildCommit),
			zap.String("version", buildVersion),
			zap.String("build_date", buildDate),
		))
	fatalError(err, "init logger")

	// Register global logger
	zap.ReplaceGlobals(loggerObj)

	return loggerObj
}

func main() {
	conf := &appcontext.Config
	loggerObj := initZapLogger()

	// Define cancelation context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ctx = version.WithContext(ctx, &version.Version{
		Version: buildVersion,
		Commit:  buildCommit,
		Date:    buildDate,
	})

	// Profiling server of collector
	profiler.Run(conf.Server.Profile.Mode,
		conf.Server.Profile.Listen, loggerObj, true)

	// Establich connect to the database
	masterDatabase, slaveDatabase, err := database.ConnectMasterSlave(ctx,
		conf.System.Storage.MasterConnect,
		conf.System.Storage.SlaveConnect,
		conf.IsDebug())
	fatalError(err, "connect to database")

	// Register callback for history log (only for modifications)
	fatalError(gormlog.Register(masterDatabase), "register history log")

	// Init permission manager
	permissionManager := permissions.NewManager(masterDatabase, conf.Permissions.RoleCacheLifetime)
	appinit.InitModelPermissions(permissionManager)

	// Init OAuth2 provider
	oauth2provider, jwtProvider := appinit.Auth(ctx, conf, masterDatabase)

	// Prepare context
	ctx = ctxlogger.WithLogger(ctx, loggerObj)
	ctx = database.WithDatabase(ctx, masterDatabase, slaveDatabase)
	ctx = permissions.WithManager(ctx, permissionManager)

	httpServer := server.HTTPServer{
		Logger:         loggerObj,
		JWTProvider:    jwtProvider,
		SessionManager: appinit.SessionManager(conf.Session.CookieName, conf.Session.Lifetime),
		Authorizers: []auth.Authorizer{
			jwt.NewAuthorizer(jwtProvider),
			oauth2.NewAuthorizer(oauth2provider),
			devtoken.NewAuthorizer(gocast.IfThen(conf.IsDebug(), &devtoken.AuthOption{
				DevToken:     conf.Session.DevToken,
				DevUserID:    conf.Session.DevUserID,
				DevAccountID: conf.Session.DevAccountID,
			}, nil)),
		},
		ContextWrap: func(ctx context.Context) context.Context {
			ctx = ctxlogger.WithLogger(ctx, loggerObj)
			ctx = database.WithDatabase(ctx, masterDatabase, slaveDatabase)
			ctx = permissions.WithManager(ctx, permissionManager)
			return ctx
		},
		InitWrap: func(mux *chi.Mux) {
			if conf.SocialAuth.Facebook.IsValid() {
				oa2conf := conf.SocialAuth.Facebook.OAuth2Config("facebook")
				mux.Handle("/auth/facebook/*",
					rest.NewWrapper(facebook.NewFacebookConfig(oa2conf), rest.WithSessionProvider(jwtProvider)).
						HandleWrapper("/auth/facebook"),
				)
			}
		},
	}
	fatalError(httpServer.Run(ctx, conf.Server.HTTP.Listen), "HTTP server")
}

func fatalError(err error, msgs ...any) {
	if err != nil {
		log.Fatalln(append(msgs, err)...)
	}
}
