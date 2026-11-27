package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
)

// Recover turns a panic in the handler into a 500 response and logs it,
// instead of dropping the connection. http.ErrAbortHandler is passed through.
func Recover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(rec)
			}
			ctxlogger.Get(r.Context()).Error("panic in HTTP handler",
				zap.Any("panic", rec), zap.ByteString("stack", debug.Stack()))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Internal Server Error","code":500}]}`))
		}()
		h.ServeHTTP(w, r)
	})
}

// MaxBodySize limits the request body; reading more fails and the server
// answers 413 when the handler reports it. A non-positive limit is off.
func MaxBodySize(h http.Handler, limit int64) http.Handler {
	if limit <= 0 {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > limit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Request Entity Too Large","code":413}]}`))
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}
		h.ServeHTTP(w, r)
	})
}
