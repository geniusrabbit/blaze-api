// Package repository implements methods of working with the repository objects
package repository

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/generated"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
)

// Repository DAO which provides functionality of working with RBAC roles
type Repository struct {
	generated.Repository[rbac.Role, uint64]
}

// New role repository
func New() *Repository {
	return &Repository{
		Repository: *generated.NewRepository[rbac.Role, uint64](),
	}
}

// GetByName returns RBAC role model by name
func (r *Repository) GetByName(ctx context.Context, name string) (*rbac.Role, error) {
	object := new(rbac.Role)
	if err := r.Slave(ctx).Find(object, `name=?`, name).Error; err != nil {
		return nil, err
	}
	return object, nil
}
