package messanger

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"go.uber.org/zap"
)

type DummyMessanger struct{}

func (d *DummyMessanger) IsEnabled() bool {
	return false
}

func (d *DummyMessanger) Send(ctx context.Context, name string, recipients []string, vars map[string]any) error {
	ctxlogger.Get(ctx).Info("Dummy messanger:",
		zap.String("name", name),
		zap.Strings("recipients", recipients),
		zap.Any("vars", vars),
	)
	return nil
}
