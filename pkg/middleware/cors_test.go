package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

func corsRequest(h http.Handler, method, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/graphql", nil)
	req.Header.Set("Origin", origin)
	if method == http.MethodOptions {
		req.Header.Set("Access-Control-Request-Method", "POST")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCORSAllowlist(t *testing.T) {
	t.Parallel()

	h := CORS(okHandler, CORSConfig{
		AllowedOrigins:   []string{"https://app.example.com", "https://*.corp.example"},
		AllowCredentials: true,
	})
	for origin, allowed := range map[string]bool{
		"https://app.example.com":         true,
		"https://admin.corp.example":      true,
		"https://evil.example":            false,
		"https://app.example.com.evil.io": false,
		"http://admin.corp.example":       false,
		"https://corp.example.evil":       false,
	} {
		rec := corsRequest(h, http.MethodGet, origin)
		got := rec.Header().Get("Access-Control-Allow-Origin")
		if allowed && (got != origin || rec.Header().Get("Access-Control-Allow-Credentials") != "true") {
			t.Errorf("%s: allow-origin = %q, want allowed with credentials", origin, got)
		}
		// Before the fix every origin was reflected with credentials
		if !allowed && (got != "" || rec.Header().Get("Access-Control-Allow-Credentials") != "") {
			t.Errorf("%s: allow-origin = %q, want no CORS headers", origin, got)
		}
		if rec.Header().Get("Vary") != "Origin" {
			t.Errorf("%s: Vary = %q, want Origin", origin, rec.Header().Get("Vary"))
		}
	}

	rec := corsRequest(h, http.MethodOptions, "https://app.example.com")
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("preflight: status = %d, headers = %v", rec.Code, rec.Header())
	}
}

func TestCORSWildcardNeverAllowsCredentials(t *testing.T) {
	t.Parallel()

	rec := corsRequest(CORS(okHandler, CORSConfig{AllowedOrigins: []string{"*"}, AllowCredentials: true}), http.MethodGet, "https://any.example")
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" || rec.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("headers = %v", rec.Header())
	}
}

func TestRecover(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	defer func() {
		if r := recover(); r != http.ErrAbortHandler {
			t.Fatalf("recovered %v, want http.ErrAbortHandler to pass through", r)
		}
	}()
	Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) })).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestMaxBodySize(t *testing.T) {
	t.Parallel()

	var readErr error
	h := MaxBodySize(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}), 8)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789")))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("declared length over the limit: status = %d, want 413", rec.Code)
	}

	// Chunked body without Content-Length: the reader stops at the limit
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789"))
	req.ContentLength = -1
	h.ServeHTTP(httptest.NewRecorder(), req)
	if readErr == nil {
		t.Fatal("reading past the limit succeeded")
	}
}
