package repository

import (
	"context"
	"voyago/core-api/internal/modules/product/entity"
)

type CategoryCommandRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, category *entity.Category) error
}

type CategoryQueryRepository interface {
	FindParent(ctx context.Context) ([]entity.Category, error)
	FindChildren(ctx context.Context) ([]entity.Category, error)
	Retrieve(ctx context.Context, id string) (*entity.Category, error)
}
