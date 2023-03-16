// Package repository contains control entety repositories
package repository

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/api-template-base/internal/context/ctxlogger"
	"github.com/geniusrabbit/api-template-base/internal/context/database"
	"github.com/geniusrabbit/api-template-base/internal/context/permissionmanager"
	"github.com/geniusrabbit/api-template-base/internal/permissions"
)

// Repository with basic functionality
type Repository struct {
}

// PermissionManager returns permission-manager object from context
func (r *Repository) PermissionManager(ctx context.Context) *permissions.Manager {
	return permissionmanager.Get(ctx)
}

// Logger returns logger object from context
func (r *Repository) Logger(ctx context.Context) *zap.Logger {
	return ctxlogger.Get(ctx)
}

// Slave returns readonly database connection
func (r *Repository) Slave(ctx context.Context) *gorm.DB {
	return database.Readonly(ctx)
}

// Master returns master database executor
func (r *Repository) Master(ctx context.Context) *gorm.DB {
	return database.ContextExecutor(ctx)
}

// Transaction returns new or exists transaction executor
func (r *Repository) Transaction(ctx context.Context) (*gorm.DB, context.Context, bool, error) {
	return database.ContextTransaction(ctx)
}
