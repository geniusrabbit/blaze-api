package historylog

import "context"

var ctxMessageKey = &struct{ name string }{"message"}

// WithMessage returns context with message
func WithMessage(ctx context.Context, msg string) context.Context {
	if msg == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxMessageKey, msg)
}

// MessageFromContext returns message from context
func MessageFromContext(ctx context.Context) string {
	if v := ctx.Value(ctxMessageKey); v != nil {
		return v.(string)
	}
	return ""
}
