package profiler

import (
	"context"
	"errors"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/pkg/profile"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// DefaultListen is used by the "net" mode when no address is given.
// It is bound to localhost: pprof exposes internals and must not be public.
const DefaultListen = "localhost:6060"

// Run starts profiling in the given mode and returns the function that stops
// it; call it on shutdown (e.g. `defer stop()` in main).
//
// Modes:
//   - "cpu", "mem"/"memory", "mutex", "block": write the profile to a file in
//     a temporary directory until stop is called;
//   - "net": serve /debug/pprof/ (and /healthcheck, /metrics when
//     withDefHandlers is true) on listenAddr, DefaultListen when empty;
//   - anything else: no profiling.
func Run(mode, listenAddr string, logger *zap.Logger, withDefHandlers ...bool) (stop func()) {
	if logger == nil {
		logger = zap.NewNop()
	}
	// Profiles go to a temporary directory; pkg/profile logs the file path
	opts := []func(*profile.Profile){profile.NoShutdownHook}
	switch mode {
	case "cpu":
		return profile.Start(append(opts, profile.CPUProfile)...).Stop
	case "mem", "memory":
		return profile.Start(append(opts, profile.MemProfile)...).Stop
	case "mutex":
		return profile.Start(append(opts, profile.MutexProfile)...).Stop
	case "block":
		return profile.Start(append(opts, profile.BlockProfile)...).Stop
	case "net":
		if listenAddr == "" {
			listenAddr = DefaultListen
		}
		srv := &http.Server{
			Addr:              listenAddr,
			Handler:           Handler(len(withDefHandlers) > 0 && withDefHandlers[0]),
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			logger.Info("profiler listening", zap.String("addr", listenAddr))
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("profile server error", zap.Error(err))
			}
		}()
		return func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
		}
	}
	return func() {}
}

// Handler returns a mux with the pprof endpoints under /debug/pprof/ and,
// optionally, /healthcheck and /metrics. It doesn't touch http.DefaultServeMux.
func Handler(withDefHandlers bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	if withDefHandlers {
		mux.HandleFunc("/healthcheck", HealthCheckHandler)
		mux.Handle("/metrics", promhttp.Handler())
	}
	return mux
}
