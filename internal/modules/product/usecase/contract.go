package usecase

import (
	"context"
	"voyago/core-api/internal/modules/product/entity"
)

type CategoryRequest struct {
	ID          string            `json:"id" validate:"omitempty,uuid" label:"ID"`
	ParentId    *string           `json:"parent_id" validate:"omitempty,uuid" label:"Parent ID"`
	Name        entity.Localized  `json:"name" validate:"required" label:"Name"`
	Slug        entity.Localized  `json:"slug" validate:"required" label:"Slug"`
	Description *entity.Localized `json:"description" validate:"omitempty" label:"Description"`
}

type CategoryResponse struct {
	ID          string             `json:"id"`
	Name        entity.Localized   `json:"name"`
	Slug        entity.Localized   `json:"slug"`
	Description *entity.Localized  `json:"description"`
	Children    []CategoryResponse `json:"children,omitempty"`
}

type ReadCategoryUseCase interface {
	Execute(ctx context.Context) ([]CategoryResponse, error)
}

type ReadCategoryDetailUseCase interface {
	Execute(ctx context.Context, id string) (*CategoryResponse, error)
}

type CreateCategoryUseCase interface {
	Execute(ctx context.Context, req *CategoryRequest) (*CategoryResponse, error)
}
