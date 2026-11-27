package profiler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
)

// HealthCheckHandler is the liveness probe: the process is up and serving.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]any{"status": "OK"})
}

// Check reports whether a dependency is ready.
type Check struct {
	Name  string
	Check func(ctx context.Context) error
}

// DatabaseCheck pings the database.
func DatabaseCheck(name string, db *gorm.DB) Check {
	return Check{Name: name, Check: func(ctx context.Context) error {
		if db == nil {
			return errors.New("no database connection")
		}
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}}
}

// ReadinessHandler is the readiness probe: it runs the checks (each within
// timeout) and answers 503 when any fails. Error details are logged, not
// returned, so the endpoint doesn't leak internals.
func ReadinessHandler(timeout time.Duration, checks ...Check) http.HandlerFunc {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		results := make(map[string]string, len(checks))
		for _, c := range checks {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			err := c.Check(ctx)
			cancel()
			if err != nil {
				status = http.StatusServiceUnavailable
				results[c.Name] = "FAIL"
				ctxlogger.Get(r.Context()).Error("readiness check failed", zap.String("check", c.Name), zap.Error(err))
				continue
			}
			results[c.Name] = "OK"
		}
		writeJSON(w, r, status, map[string]any{
			"status": map[bool]string{true: "OK", false: "FAIL"}[status == http.StatusOK],
			"checks": results,
		})
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	// Headers must be set before WriteHeader
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		ctxlogger.Get(r.Context()).Error("write HTTP response", zap.Error(err))
	}
}
