package server

import (
	"context"
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

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/middleware"
	"github.com/geniusrabbit/blaze-api/pkg/profiler"
	"github.com/geniusrabbit/blaze-api/server/graphql"
)

type (
	contextWrapper func(context.Context) context.Context
	muxInitWrapper func(mux *chi.Mux)
)

// HTTPServer wrapper object
type HTTPServer struct {
	RequestTimeout time.Duration
	ContextWrap    contextWrapper
	InitWrap       muxInitWrapper
	Authorizers    []auth.Authorizer
	JWTProvider    *jwt.Provider
	SessionManager *scs.SessionManager
	Logger         *zap.Logger
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func (s *HTTPServer) Run(ctx context.Context, address string) (err error) {
	s.Logger.Info("Start balance HTTP API: " + address)

	mux := chi.NewRouter()
	mux.With(basicauth.NewFromEnv("Graph", "GRAPHQL_USERS_")).
		Handle("/", playground.Handler("Query console", "/graphql"))
	mux.Handle("/healthcheck", http.HandlerFunc(profiler.HealthCheckHandler))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/graphql", graphql.GraphQL(s.JWTProvider))

	if s.InitWrap != nil {
		s.InitWrap(mux)
	}

	h := http.Handler(mux)

	// Add middleware's
	h = auth.Middelware(h, s.Authorizers...)
	h = middleware.HTTPContextWrapper(h, s.ContextWrap)
	h = middleware.HTTPSession(h, s.SessionManager)
	h = middleware.RealIP(h)
	h = middleware.AllowCORS(h)
	h = middleware.RequestID(h)
	h = nethttp.Middleware(opentracing.GlobalTracer(), h)

	srv := &http.Server{Addr: address, Handler: h}
	go func() {
		<-ctx.Done()
		s.Logger.Info("Shutting down the http server")
		if err := srv.Shutdown(context.Background()); err != nil {
			s.Logger.Error("Failed to shutdown http server", zap.Error(err))
		}
	}()

	s.Logger.Info(fmt.Sprintf("Starting listening at %s", address))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		s.Logger.Error("Failed to listen and serve", zap.Error(err))
		return err
	}
	return nil
}
