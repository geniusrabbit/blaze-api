package generated

import "context"

type UsecaseIface[T any, TID any] interface {
	Get(ctx context.Context, id TID, qops ...Option) (*T, error)
	FetchList(ctx context.Context, qops ...Option) ([]*T, error)
	Count(ctx context.Context, qops ...Option) (int64, error)
	Create(ctx context.Context, obj *T, message string) (TID, error)
	Update(ctx context.Context, id TID, obj *T, message string) error
	Delete(ctx context.Context, id TID, message string) error
}

type UsecaseApproveIface[TID any] interface {
	Approve(ctx context.Context, id TID, message string) error
	Reject(ctx context.Context, id TID, message string) error
}
