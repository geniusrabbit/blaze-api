package middleware

import (
	"net/http"

	scs "github.com/alexedwards/scs/v2"
	"github.com/opentracing/opentracing-go"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
)

// HTTPSession middleware wrapper
func HTTPSession(h http.Handler, manager *scs.SessionManager) http.Handler {
	return manager.LoadAndSave(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span, ctx := opentracing.StartSpanFromContext(r.Context(), "middleware.httpsession")
			if span != nil {
				defer span.Finish()
			} else {
				ctx = r.Context()
			}
			h.ServeHTTP(w, r.WithContext(
				session.WithSession(ctx, manager),
			))
		}),
	)
}
