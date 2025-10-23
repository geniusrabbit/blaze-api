package generated

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository"
)

// List select options
type (
	Option  = repository.QOption
	Options = repository.ListOptions
)

type RepositoryIface[T any, TID any] interface {
	Get(ctx context.Context, id TID, qops ...Option) (*T, error)
	FetchList(ctx context.Context, qops ...Option) ([]*T, error)
	Count(ctx context.Context, qops ...Option) (int64, error)
	Create(ctx context.Context, obj *T, message string) (TID, error)
	Update(ctx context.Context, id TID, obj *T, message string) error
	Delete(ctx context.Context, id TID, message string) error
}

type RepositoryApproveIface[TID any] interface {
	Approve(ctx context.Context, id TID, message string) error
	Reject(ctx context.Context, id TID, message string) error
}

type RepositoryIfaceWithApprove[T any, TID any] interface {
	RepositoryIface[T, TID]
	RepositoryApproveIface[TID]
}
