package profiler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestHealthCheckHandler(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HealthCheckHandler(rec, httptest.NewRequest(http.MethodGet, "/healthcheck", nil))
	// Before the fix Content-Type was set after WriteHeader and got lost
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status = %d, content-type = %q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestReadinessHandler(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	ok := ReadinessHandler(time.Second, DatabaseCheck("db", db))
	rec := httptest.NewRecorder()
	ok(rec, httptest.NewRequest(http.MethodGet, "/readiness", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"db":"OK"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}

	failing := ReadinessHandler(time.Second, DatabaseCheck("db", db),
		Check{Name: "cache", Check: func(context.Context) error { return errors.New("redis: password=secret refused") }})
	rec = httptest.NewRecorder()
	failing(rec, httptest.NewRequest(http.MethodGet, "/readiness", nil))
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"cache":"FAIL"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("error details leaked: %s", rec.Body)
	}
}

func TestProfilerHandlerServesPprof(t *testing.T) {
	t.Parallel()

	// Before the fix net/http/pprof was never imported: /debug/pprof was 404
	rec := httptest.NewRecorder()
	Handler(true).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/debug/pprof/: status = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	Handler(false).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("/metrics without default handlers: status = %d", rec.Code)
	}
}

func TestRunNetModeStops(t *testing.T) {
	t.Parallel()

	stop := Run("net", "127.0.0.1:0", nil)
	stop()
	Run("none", "", nil)()
}
