package messanger

import "context"

var ctxMessangerKey = &struct{ s string }{"messanger"}
var globalDummyMessanger = &DummyMessanger{}

// WithMessanger sets the messanger in the context
func WithMessanger(ctx context.Context, messanger Messanger) context.Context {
	return context.WithValue(ctx, ctxMessangerKey, messanger)
}

// Get messanger from the context
func Get(ctx context.Context) Messanger {
	if v := ctx.Value(ctxMessangerKey); v != nil {
		return v.(Messanger)
	}
	return globalDummyMessanger
}
