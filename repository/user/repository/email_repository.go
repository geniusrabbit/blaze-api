package repository

import (
	"context"
	"database/sql"
	"strings"

	baseRepo "github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type emailRepository[T user.EmailCapableModel] struct {
	baseRepo.Repository
	core     user.Repository[T]
	newModel func() T
}

// NewEmailRepository creates email lookup repository.
func NewEmailRepository[T user.EmailCapableModel](core user.Repository[T], newModel func() T) user.EmailRepository[T] {
	return &emailRepository[T]{core: core, newModel: newModel}
}

// GetByEmail retrieves user by email (case-insensitive).
func (r *emailRepository[T]) GetByEmail(ctx context.Context, email string) (T, error) {
	var zero T
	object := r.newModel()
	if err := r.Slave(ctx).First(object, `lower(email)=?`, strings.ToLower(email)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			return zero, nil
		}
		return zero, err
	}
	return object, nil
}
