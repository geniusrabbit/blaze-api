// Package repository contains control entety repositories
package repository

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
)

// Repository with basic functionality
type Repository struct{}

// PermissionManager returns permission-manager object from context
func (r *Repository) PermissionManager(ctx context.Context) *permissions.Manager {
	return permissions.FromContext(ctx)
}

// Logger returns logger object from context
func (r *Repository) Logger(ctx context.Context) *zap.Logger {
	return ctxlogger.Get(ctx)
}

// Slave returns readonly database connection
// TODO: rename to ReadOnly
func (r *Repository) Slave(ctx context.Context) *gorm.DB {
	return database.Readonly(ctx).WithContext(ctx)
}

// Master returns master database executor
// TODO: rename to Leader
func (r *Repository) Master(ctx context.Context) *gorm.DB {
	return database.ContextExecutor(ctx).WithContext(ctx)
}

// Transaction returns new or exists transaction executor
func (r *Repository) Transaction(ctx context.Context) (*gorm.DB, context.Context, bool, error) {
	return database.ContextTransaction(ctx)
}

// TransactionExec executes function in transaction
func (r *Repository) TransactionExec(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error, dbs ...*gorm.DB) error {
	return database.ContextTransactionExec(ctx, fn, dbs...)
}
