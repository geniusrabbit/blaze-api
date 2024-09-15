package graphql

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/opentracing/opentracing-go"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/server/graphql/directives"
	"github.com/geniusrabbit/blaze-api/server/graphql/generated"
	"github.com/geniusrabbit/blaze-api/server/graphql/resolvers"
)

// GraphQL mux handler
func GraphQL(provider *jwt.Provider) http.Handler {
	srv := handler.New(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: resolvers.NewResolver(provider),
			Directives: generated.DirectiveRoot{
				HasPermissions:    directives.HasPermissions,
				Auth:              directives.Auth,
				Acl:               directives.HasPermissions,
				SkipNoPermissions: directives.SkipNoPermissions,
			},
		}),
	)
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(r.Context(), "graphql.request")
		defer span.Finish()
		srv.ServeHTTP(rw, r.WithContext(ctx))
	})
}
