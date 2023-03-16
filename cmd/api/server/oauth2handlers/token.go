package oauth2handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ory/fosite"

	"github.com/geniusrabbit/api-template-base/internal/context/ctxlogger"
	"github.com/geniusrabbit/api-template-base/internal/oauth2"
)

// Token endpoint
func Token(oauth2provider fosite.OAuth2Provider) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// This context will be passed to all methods.
		ctx := oauth2.NewContext()

		// Create an empty session object which will be passed to the request handlers
		mySessionData := newSession("")

		// Parse the income request
		err := req.ParseMultipartForm(1 << 10)
		if err != nil && err != http.ErrNotMultipart {
			ctxlogger.Get(req.Context()).Error("Error occurred in ParseMultipartForm", zap.Error(err))
		} else if err = req.ParseForm(); err != nil {
			ctxlogger.Get(req.Context()).Error("Error occurred in ParseForm", zap.Error(err))
		}
		if err != nil {
			http.Error(rw, "bad request", http.StatusBadRequest)
			return
		}

		// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
		accessRequest, err := oauth2provider.NewAccessRequest(ctx, req, mySessionData)

		// Catch any errors, e.g.:
		// * unknown client
		// * invalid redirect
		// * ...
		if err != nil {
			ctxlogger.Get(req.Context()).Error("Error occurred in NewAccessRequest", zap.Error(err))
			oauth2provider.WriteAccessError(ctx, rw, accessRequest, err)
			return
		}

		// If this is a client_credentials grant, grant all scopes the client is allowed to perform.
		if accessRequest.GetGrantTypes().ExactOne("client_credentials") {
			for _, scope := range accessRequest.GetRequestedScopes() {
				if fosite.HierarchicScopeStrategy(accessRequest.GetClient().GetScopes(), scope) {
					accessRequest.GrantScope(scope)
				}
			}
		}

		// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
		// and aggregate the result in response.
		response, err := oauth2provider.NewAccessResponse(ctx, accessRequest)
		if err != nil {
			ctxlogger.Get(req.Context()).Error("Error occurred in NewAccessResponse", zap.Error(err))
			oauth2provider.WriteAccessError(ctx, rw, accessRequest, err)
			return
		}

		// All done, send the response.
		oauth2provider.WriteAccessResponse(ctx, rw, accessRequest, response)

		// The client now has a valid access token
	}
}
