package oauth2handlers

import (
	"net/http"

	"github.com/ory/fosite"
	"go.uber.org/zap"

	"github.com/geniusrabbit/api-template-base/internal/context/ctxlogger"
	"github.com/geniusrabbit/api-template-base/internal/oauth2"
)

// Introspect endpoint
func Introspect(oauth2provider fosite.OAuth2Provider) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var (
			ctx           = oauth2.NewContext()
			mySessionData = newSession("")
			ir, err       = oauth2provider.NewIntrospectionRequest(ctx, req, mySessionData)
		)
		if err != nil {
			ctxlogger.Get(req.Context()).Error("Error occurred in NewAuthorizeRequest", zap.Error(err))
			oauth2provider.WriteIntrospectionError(ctx, rw, err)
			return
		}
		oauth2provider.WriteIntrospectionResponse(ctx, rw, ir)
	}
}
