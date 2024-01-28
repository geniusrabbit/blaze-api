package main

import (
	"context"
	"fmt"
	"log"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/goconfig"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/context/permissionmanager"
	"github.com/geniusrabbit/blaze-api/database"
	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/appcontext"
	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/appinit"
	"github.com/geniusrabbit/blaze-api/example/api/cmd/api/migratedb"
	"github.com/geniusrabbit/blaze-api/example/api/internal/server"
	"github.com/geniusrabbit/blaze-api/middleware"
	"github.com/geniusrabbit/blaze-api/permissions"
	"github.com/geniusrabbit/blaze-api/profiler"
	"github.com/geniusrabbit/blaze-api/repository/historylog/middleware/gormlog"
	"github.com/geniusrabbit/blaze-api/zlogger"
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
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
	gormlog.Register(masterDatabase)

	// Init permission manager
	permissionManager := permissions.NewManager(masterDatabase, conf.Permissions.RoleCacheLifetime)
	appinit.InitModelPermissions(permissionManager)

	// Init OAuth2 provider
	oauth2provider, jwtProvider := appinit.Auth(ctx, conf, masterDatabase)

	// Prepare context
	ctx = ctxlogger.WithLogger(ctx, loggerObj)
	ctx = database.WithDatabase(ctx, masterDatabase, slaveDatabase)
	ctx = permissionmanager.WithManager(ctx, permissionManager)

	httpServer := server.HTTPServer{
		Logger:         loggerObj,
		OAuth2provider: oauth2provider,
		JWTProvider:    jwtProvider,
		SessionManager: appinit.SessionManager(conf.Session.CookieName, conf.Session.Lifetime),
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
	fatalError(httpServer.Run(ctx, conf.Server.HTTP.Listen), "HTTP server")
}

func fatalError(err error, msgs ...any) {
	if err != nil {
		log.Fatalln(append(msgs, err)...)
	}
}
