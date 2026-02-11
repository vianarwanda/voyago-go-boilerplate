package command

import (
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/modules/product/entity"
	"voyago/core-api/internal/modules/product/repository"
)

type categoryRepository struct {
	*database.GormBaseRepository[entity.Category]
}

var _ repository.CategoryCommandRepository = (*categoryRepository)(nil)

func NewCategoryRepository(db database.Database) repository.CategoryCommandRepository {
	return &categoryRepository{
		GormBaseRepository: &database.GormBaseRepository[entity.Category]{
			DB:          db,
			ErrorMapper: database.MapDBError,
		},
	}
}
