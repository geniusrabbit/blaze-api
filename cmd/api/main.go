package main

import (
	"context"
	"fmt"
	"log"

	"github.com/demdxx/goconfig"
	"go.uber.org/zap"

	"github.com/geniusrabbit/api-template-base/cmd/api/appcontext"
	"github.com/geniusrabbit/api-template-base/cmd/api/appinit"
	"github.com/geniusrabbit/api-template-base/cmd/api/migratedb"
	"github.com/geniusrabbit/api-template-base/cmd/api/server"
	"github.com/geniusrabbit/api-template-base/internal/context/ctxlogger"
	"github.com/geniusrabbit/api-template-base/internal/context/permissionmanager"
	"github.com/geniusrabbit/api-template-base/internal/database"
	_ "github.com/geniusrabbit/api-template-base/internal/gopentracing"
	"github.com/geniusrabbit/api-template-base/internal/middleware"
	"github.com/geniusrabbit/api-template-base/internal/permissions"
	"github.com/geniusrabbit/api-template-base/internal/profiler"
	"github.com/geniusrabbit/api-template-base/internal/zlogger"
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

	sources := []string{"file:///data/migrations/prod", "file:///data/migrations/fixtures"}
	if len(sources) > 0 {
		fatalError(migratedb.Migrate(conf.System.Storage.MasterConnect, sources), "migrate database")
	}
}

func main() {
	conf := &appcontext.Config

	// Init new logger object
	loggerObj, err := zlogger.New(conf.ServiceName, conf.LogEncoder,
		conf.LogLevel, conf.LogAddr, zap.Fields(
			zap.String("commit", buildCommit),
			zap.String("version", buildVersion),
			zap.String("build_date", buildDate),
		))
	fatalError(err, "init logger")

	// Register global logger
	zap.ReplaceGlobals(loggerObj)

	// Define cancelation context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Profiling server of collector
	profiler.Run(conf.Server.Profile.Mode,
		conf.Server.Profile.Listen, loggerObj, true)

	// Establich connect to the database
	masterDatabase, err := database.Connect(ctx,
		conf.System.Storage.MasterConnect, conf.IsDebug())
	fatalError(err, "connect to master database")

	slaveDatabase, err := database.Connect(ctx,
		conf.System.Storage.SlaveConnect, conf.IsDebug())
	fatalError(err, "connect to slave database")

	// Init permission manager
	permissionManager := permissions.NewManager(masterDatabase, conf.Permissions.RoleCacheLifetime)
	appinit.InitModelPermissions(permissionManager)

	// Init OAuth2 provider
	oauth2provider, jwtProvider := appinit.Auth(ctx, conf, masterDatabase)
	authOption := (*middleware.AuthOption)(nil)

	if conf.IsDebug() {
		authOption = &middleware.AuthOption{
			DevToken:     conf.Session.DevToken,
			DevUserID:    conf.Session.DevUserID,
			DevAccountID: conf.Session.DevAccountID,
		}
	}

	// Prepare context
	ctx = ctxlogger.WithLogger(ctx, loggerObj)
	ctx = database.WithDatabase(ctx, masterDatabase, slaveDatabase)
	ctx = permissionmanager.WithManager(ctx, permissionManager)

	httpServer := server.HTTPServer{
		Logger:         loggerObj,
		OAuth2provider: oauth2provider,
		JWTProvider:    jwtProvider,
		SessionManager: appinit.SessionManager(conf.Session.CookieName, conf.Session.Lifetime),
		AuthOption:     authOption,
		ContextWrap: func(ctx context.Context) context.Context {
			ctx = ctxlogger.WithLogger(ctx, loggerObj)
			ctx = database.WithDatabase(ctx, masterDatabase, slaveDatabase)
			ctx = permissionmanager.WithManager(ctx, permissionManager)
			return ctx
		},
	}
	fatalError(httpServer.Run(ctx, conf.Server.HTTP.Listen), "HTTP server")
}

func fatalError(err error, msgs ...any) {
	if err != nil {
		log.Fatalln(append(msgs, err)...)
	}
}
