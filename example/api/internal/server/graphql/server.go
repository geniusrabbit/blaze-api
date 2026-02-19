package graphql

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/opentracing/opentracing-go"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/generated"
	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/resolvers"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/server/graphql/directives"
)

// GraphQL mux handler
func GraphQL(provider *jwt.Provider, options option.Usecase) http.Handler {
	srv := handler.New(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: resolvers.NewResolver(provider, options),
			Directives: generated.DirectiveRoot{
				HasPermissions:    directives.HasPermissions,
				Acl:               directives.HasPermissions,
				Auth:              directives.Auth,
				SkipNoPermissions: directives.SkipNoPermissions,
				CacheData:         directives.CacheData,
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
	srv.AddTransport(transport.UrlEncodedForm{})
	srv.AddTransport(transport.GRAPHQL{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	srv.SetRecoverFunc(recoverHandler)

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(r.Context(), "graphql.request")
		defer span.Finish()
		srv.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func recoverHandler(ctx context.Context, err any) error {
	switch verr := err.(type) {
	case error:
		if errors.Is(verr, acl.ErrNoPermissions) {
			return verr
		}
	}
	return graphql.DefaultRecover(ctx, err)
}
