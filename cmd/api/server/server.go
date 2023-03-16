package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/ory/fosite"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/geniusrabbit/api-template-base/cmd/api/server/jwthandler"
	"github.com/geniusrabbit/api-template-base/cmd/api/server/oauth2handlers"
	"github.com/geniusrabbit/api-template-base/internal/jwt"
	"github.com/geniusrabbit/api-template-base/internal/middleware"
	"github.com/geniusrabbit/api-template-base/internal/profiler"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql"
)

type contextWrapper func(context.Context) context.Context

// HTTPServer wrapper object
type HTTPServer struct {
	RequestTimeout time.Duration
	ContextWrap    contextWrapper
	OAuth2provider fosite.OAuth2Provider
	JWTProvider    *jwt.Provider
	SessionManager *scs.SessionManager
	AuthOption     *middleware.AuthOption
	Logger         *zap.Logger
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func (s *HTTPServer) Run(ctx context.Context, address string) (err error) {
	s.Logger.Info("Start balance HTTP API: " + address)

	// mux := http.NewServeMux()
	mux := chi.NewRouter()
	mux.HandleFunc("/healthcheck", profiler.HealthCheckHandler)
	mux.Handle("/metrics", promhttp.Handler())

	mux.With(basicauth.NewFromEnv("Graph", "GRAPHQL_USERS_")).
		Handle("/", playground.Handler("Query console", "/graphql"))

	// OAuth2 handlers
	swh := graphql.GraphQL(s.JWTProvider)
	swh = middleware.AuthHTTP("http_", swh, s.OAuth2provider, s.JWTProvider, s.AuthOption)

	mux.Handle("/graphql", swh)
	mux.HandleFunc("/oauth2/auth", oauth2handlers.Auth(s.OAuth2provider))
	mux.HandleFunc("/oauth2/token", oauth2handlers.Token(s.OAuth2provider))
	mux.HandleFunc("/oauth2/revoke", oauth2handlers.Revoke(s.OAuth2provider))
	mux.HandleFunc("/oauth2/introspect", oauth2handlers.Introspect(s.OAuth2provider))
	mux.HandleFunc("/authenticate", jwthandler.AuthHandler(s.JWTProvider))

	h := middleware.HTTPContextWrapper(mux, s.ContextWrap)
	h = middleware.HTTPSession(h, s.SessionManager)
	h = middleware.RealIP(h)
	h = middleware.AllowCORS(h)
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
