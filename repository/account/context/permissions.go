package context

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/account/models"
)

// PermissionCheckAccountFromContext returns the original account for permission checks
// from the given context, or nil if not found.
func PermissionCheckAccountFromContext(ctx context.Context) *models.Account {
	switch acc := ctx.Value(models.CtxPermissionCheckAccount).(type) {
	case nil:
	case *models.Account:
		return acc
	}
	return nil
}
