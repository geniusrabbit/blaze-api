package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig configures cross-origin access.
type CORSConfig struct {
	// AllowedOrigins lists the allowed origins, like "https://app.example.com".
	// An entry "https://*.example.com" allows any subdomain over https.
	// A single "*" allows any origin, but then credentials are never allowed.
	AllowedOrigins []string
	// AllowCredentials lets the browser send cookies and Authorization to
	// allowed origins. Never combined with "*".
	AllowCredentials bool
	// AllowedHeaders for preflight requests; a default set is used when empty.
	AllowedHeaders []string
	// AllowedMethods for preflight requests; a default set is used when empty.
	AllowedMethods []string
}

var (
	defaultCORSHeaders = []string{"Content-Type", "Accept", "Authorization", "X-Request-ID", "auth.cross.account"}
	defaultCORSMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
)

// CORS allows cross-origin requests from the configured origins only.
func CORS(h http.Handler, cfg CORSConfig) http.Handler {
	headers := strings.Join(orDefault(cfg.AllowedHeaders, defaultCORSHeaders), ",")
	methods := strings.Join(orDefault(cfg.AllowedMethods, defaultCORSMethods), ",")
	anyOrigin := len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Origin")
		if !anyOrigin && !originAllowed(origin, cfg.AllowedOrigins) {
			// Not an allowed origin: no CORS headers, the browser blocks the response
			h.ServeHTTP(w, r)
			return
		}
		if anyOrigin {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// AllowCORS allows Cross Origin Resource Sharing from any origin, with credentials.
//
// Deprecated: any site can then make authenticated requests with the user's
// cookies. Use CORS with an explicit list of origins.
func AllowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Accept,Authorization,X-Tunnel-Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,POST,PUT,DELETE")
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(a, origin) {
			return true
		}
		// "https://*.example.com" matches "https://x.example.com"
		if scheme, host, ok := strings.Cut(a, "://*."); ok {
			prefix := scheme + "://"
			if len(origin) > len(prefix) && strings.EqualFold(origin[:len(prefix)], prefix) &&
				strings.HasSuffix(strings.ToLower(origin), "."+strings.ToLower(host)) {
				return true
			}
		}
	}
	return false
}

func orDefault(v, def []string) []string {
	if len(v) == 0 {
		return def
	}
	return v
}
