package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/profiler"
)

func TestProbeRouterBypassesTheApp(t *testing.T) {
	t.Parallel()

	// The app stands for the session/auth chain, which panics or fails while
	// the database is down
	appCalls := 0
	app := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		appCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})
	dbDown := profiler.Check{Name: "db", Check: func(context.Context) error { return errors.New("connection refused") }}
	h := probeRouter(app, []profiler.Check{dbDown})

	for path, want := range map[string]int{
		"/healthcheck": http.StatusOK,                 // liveness: the process is up
		"/readiness":   http.StatusServiceUnavailable, // readiness: the DB is not
		"/metrics":     http.StatusOK,
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != want {
			t.Errorf("%s: status = %d, want %d", path, rec.Code, want)
		}
	}
	if appCalls != 0 {
		t.Fatalf("probes went through the app %d times", appCalls)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/graphql", nil))
	if appCalls != 1 {
		t.Fatal("other paths must reach the app")
	}
}
