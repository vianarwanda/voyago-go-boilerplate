package query

import (
	"context"
	"errors"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/modules/product/entity"
	"voyago/core-api/internal/modules/product/repository"

	"gorm.io/gorm"
)

type categoryRepository struct {
	DB database.Database
}

var _ repository.CategoryQueryRepository = (*categoryRepository)(nil)

func NewCategoryRepository(db database.Database) repository.CategoryQueryRepository {
	return &categoryRepository{
		DB: db,
	}
}

func (r *categoryRepository) Retrieve(ctx context.Context, id string) (*entity.Category, error) {
	var category *entity.Category
	err := r.DB.WithContext(ctx).
		Model(&entity.Category{}).
		Select("id",
			"name",
			"parent_id",
			"slug",
			"description",
		).
		Where("id = ?", id).
		First(&category).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, database.MapDBError(err)
	}

	return category, nil
}

func (r *categoryRepository) FindParent(ctx context.Context) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.DB.WithContext(ctx).
		Model(&entity.Category{}).
		Select("id",
			"parent_id",
			"name",
			"slug",
			"description",
		).
		Where("parent_id is null").
		Order("sort_order asc").
		Find(&categories).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, database.MapDBError(err)
	}

	return categories, nil
}

func (r *categoryRepository) FindChildren(ctx context.Context) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.DB.WithContext(ctx).
		Model(&entity.Category{}).
		Select("id",
			"parent_id",
			"name",
			"slug",
			"description",
		).
		Where("parent_id is not null").
		Order("sort_order asc").
		Find(&categories).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, database.MapDBError(err)
	}

	return categories, nil
}
