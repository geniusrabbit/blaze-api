// +build jaeger

package gopentracing

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

func init() {
	cfg, err := config.FromEnv()
	if err != nil {
		zap.L().Warn("cannot parse Jaeger env vars", zap.Error(err))
		return
	}

	tracer, closer, err := cfg.NewTracer(
		config.Logger(otErrorWrapper{
			logger: zap.L().With(zap.String("module", "opentracing")),
		}),
		// jaeger.TracerOptions.PoolSpans(true),
	)

	if err != nil {
		// ! Attention, we can't connect to the tracer, but we don't have to fall down any case
		zap.L().Error("cannot initialize tracer", zap.Error(err))
		return
	}

	// * Register tracer as global interface to make it work inside the application
	opentracing.SetGlobalTracer(tracer)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		_ = closer.Close()
	}()
}

type otErrorWrapper struct {
	logger *zap.Logger
}

func (l otErrorWrapper) Error(s string) {
	l.logger.Error(s)
}

func (l otErrorWrapper) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}
