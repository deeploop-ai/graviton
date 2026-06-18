package crud

import "context"

type Repository[I any, T any] interface {
	Create(ctx context.Context, v T) error
	Update(ctx context.Context, v T, updateColumns ...string) error
	Delete(ctx context.Context, v T) error
	List(ctx context.Context, param ListParams) ([]T, int, string, error)
	Get(ctx context.Context, id I) (T, error)
	BatchGet(ctx context.Context, ids []I) ([]T, error)
	BatchDelete(ctx context.Context, ids []I) error
}
