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

type RepositoryIface[T Model[TID], TID comparable] interface {
	Get(ctx context.Context, id TID, qops ...Option) (*T, error)
	FetchList(ctx context.Context, qops ...Option) ([]*T, error)
	Count(ctx context.Context, qops ...Option) (int64, error)
	Create(ctx context.Context, obj *T, opts ...Option) (TID, error)
	Update(ctx context.Context, id TID, obj *T, opts ...Option) error
	Delete(ctx context.Context, id TID, opts ...Option) error
}

type RepositoryApproveIface[TID comparable] interface {
	Approve(ctx context.Context, id TID, opts ...Option) error
	Reject(ctx context.Context, id TID, opts ...Option) error
}

type RepositoryIfaceWithApprove[T Model[TID], TID comparable] interface {
	RepositoryIface[T, TID]
	RepositoryApproveIface[TID]
}
