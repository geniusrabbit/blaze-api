package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql"
	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/middleware"
	"github.com/geniusrabbit/blaze-api/pkg/profiler"
	accAuth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/option/repository"
	"github.com/geniusrabbit/blaze-api/repository/option/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/limits"
)

type (
	contextWrapper func(context.Context) context.Context
	muxInitWrapper func(mux *chi.Mux)
)

// Config of the HTTP server limits. Zero durations are off.
type Config struct {
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxBodySize       int64
	CORS              middleware.CORSConfig
	GraphQLLimits     limits.Config
}

// HTTPServer wrapper object
type HTTPServer struct {
	Config         Config
	ContextWrap    contextWrapper
	InitWrap       muxInitWrapper
	Authorizers    []auth.Authorizer[*domain.User, *domain.Account]
	JWTProvider    *jwt.Provider
	SessionManager *scs.SessionManager
	AuthLoader     *accAuth.Loader[*domain.User, *domain.Account]
	Logger         *zap.Logger
	GraphqlOptions graphql.Options
	// ReadinessChecks back the /readiness probe (e.g. the database ping)
	ReadinessChecks []profiler.Check
}

// Run starts a HTTP server and blocks while running if successful.
// The server is shut down gracefully when "ctx" is canceled: Run returns
// after in-flight requests finish or Config.ShutdownTimeout passes.
func (s *HTTPServer) Run(ctx context.Context, address string) (err error) {
	s.Logger.Info("Start HTTP API: " + address)

	mux := chi.NewRouter()
	mux.With(basicauth.NewFromEnv("Graph", "GRAPHQL_USERS_")).
		Handle("/playground", playground.Handler("Query console", "/graphql"))
	mux.Handle("/graphql", graphql.GraphQL(s.JWTProvider,
		usecase.NewUsecase(repository.NewOptionRepository(nil)),
		s.Config.GraphQLLimits, s.Config.MaxBodySize, s.GraphqlOptions...))

	if s.InitWrap != nil {
		s.InitWrap(mux)
	}

	h := http.Handler(mux)

	// Add middleware's
	h = accAuth.Middleware(h, s.AuthLoader, s.Authorizers...)
	h = middleware.HTTPContextWrapper(h, s.ContextWrap)
	h = middleware.HTTPSession(h, s.SessionManager)
	h = middleware.MaxBodySize(h, s.Config.MaxBodySize)
	h = middleware.Recover(h)
	h = middleware.RealIP(h)
	h = middleware.CORS(h, s.Config.CORS)
	h = middleware.RequestID(h)
	h = nethttp.Middleware(opentracing.GlobalTracer(), h)

	// Probes and metrics don't go through sessions and authorization: they
	// must answer while the database (used to load roles) is down, otherwise
	// a DB outage makes the liveness probe fail and restarts the pods.
	h = probeRouter(h, s.ReadinessChecks)

	srv := &http.Server{
		Addr:              address,
		Handler:           h,
		ReadTimeout:       s.Config.ReadTimeout,
		WriteTimeout:      s.Config.WriteTimeout,
		ReadHeaderTimeout: s.Config.ReadHeaderTimeout,
		IdleTimeout:       s.Config.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		s.Logger.Info("Shutting down the http server")
		shutdownCtx := context.Background()
		if s.Config.ShutdownTimeout > 0 {
			var cancel context.CancelFunc
			shutdownCtx, cancel = context.WithTimeout(shutdownCtx, s.Config.ShutdownTimeout)
			defer cancel()
		}
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.Logger.Error("Failed to shutdown http server", zap.Error(err))
		}
	}()

	s.Logger.Info(fmt.Sprintf("Starting listening at %s", address))
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.Logger.Error("Failed to listen and serve", zap.Error(err))
		return err
	}
	// Wait for in-flight requests before returning
	<-shutdownDone
	return nil
}

// probeRouter serves /healthcheck (liveness), /readiness and /metrics
// directly and passes everything else to the application handler.
func probeRouter(app http.Handler, readiness []profiler.Check) http.Handler {
	probes := http.NewServeMux()
	probes.HandleFunc("/healthcheck", profiler.HealthCheckHandler)
	if len(readiness) > 0 {
		probes.Handle("/readiness", profiler.ReadinessHandler(2*time.Second, readiness...))
	}
	probes.Handle("/metrics", promhttp.Handler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := probes.Handler(r); pattern != "" {
			probes.ServeHTTP(w, r)
			return
		}
		app.ServeHTTP(w, r)
	})
}
