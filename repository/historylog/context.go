package historylog

import "context"

var (
	ctxMessageKey = &struct{ name string }{"message"}
	ctxActionKey  = &struct{ name string }{"action"}
	ctxPKKey      = &struct{ name string }{"pk"}
)

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

// WithAction returns context with action
func WithAction(ctx context.Context, action string) context.Context {
	if action == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxActionKey, action)
}

// ActionFromContext returns action from context
func ActionFromContext(ctx context.Context) string {
	if v := ctx.Value(ctxActionKey); v != nil {
		return v.(string)
	}
	return ""
}

// WithPK returns context with primary key
func WithPK(ctx context.Context, pk any) context.Context {
	if pk == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxPKKey, pk)
}

// PKFromContext returns primary key from context
func PKFromContext(ctx context.Context) any {
	if v := ctx.Value(ctxPKKey); v != nil {
		return v
	}
	return nil
}

// WithMessageAndPK returns context with message and primary key
func WithMessageAndPK(ctx context.Context, msg string, pk any) context.Context {
	return WithMessage(WithPK(ctx, pk), msg)
}
