package messanger

import "context"

type Messanger interface {
	Send(ctx context.Context, name string, recipients []string, vars map[string]any) error
}

type MessangerFunc func(ctx context.Context, name string, recipients []string, vars map[string]any) error

func (f MessangerFunc) Send(ctx context.Context, name string, recipients []string, vars map[string]any) error {
	return f(ctx, name, recipients, vars)
}
