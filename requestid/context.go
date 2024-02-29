package requestid

import "context"

var ctxQueryIDKey = &struct{ s string }{"queryid:queryid"}

// WithQueryID puts to the context queryid
func WithQueryID(ctx context.Context, queryid string) context.Context {
	return context.WithValue(ctx, ctxQueryIDKey, queryid)
}

// WithDefaultQueryID puts to the context default queryid
func WithDefaultQueryID(ctx context.Context) context.Context {
	return WithQueryID(ctx, "0")
}

// Get returns queryid
func Get(ctx context.Context) string {
	if q, _ := ctx.Value(ctxQueryIDKey).(string); q != "" {
		return q
	}
	return "0"
}
