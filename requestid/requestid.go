package requestid

import (
	"net/http"
)

// RequestID middleware adds a unique request id to the request context
func RequestID(next http.Handler, opts ...Option) http.Handler {
	var headerName string
	var idGetter func(r *http.Request) string
	var requestIDGenerator func(r *http.Request) string

	// Isolate configuration object to release it after the function ends
	{
		var options config
		for _, opt := range opts {
			opt(&options)
		}
		headerName = options.getHeaderName()
		idGetter = options.getExternalIDGetter()
		requestIDGenerator = options.getRequestIDGenerator()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryid := idGetter(r)
		if queryid == "" {
			queryid = requestIDGenerator(r)
		}
		r.Header.Add(headerName, queryid)
		next.ServeHTTP(w, r.WithContext(
			WithQueryID(r.Context(), queryid),
		))
	})
}
