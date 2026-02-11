package usecase

import (
	"context"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/modules/product/entity"
	"voyago/core-api/internal/modules/product/repository"
	baserepo "voyago/core-api/internal/pkg/repository"
)

type ReadCategoryDetailRepositories struct {
	CategoryCmd repository.CategoryCommandRepository
	CategoryQry repository.CategoryQueryRepository
}

type readCategoryDetailUseCase struct {
	Log    logger.Logger
	Tracer tracer.Tracer
	Runner baserepo.TransactionManager
	Repo   ReadCategoryDetailRepositories
}

var _ ReadCategoryDetailUseCase = (*readCategoryDetailUseCase)(nil)

func NewReadCategoryDetailUseCase(log logger.Logger, trc tracer.Tracer, runner baserepo.TransactionManager, repo ReadCategoryDetailRepositories) ReadCategoryDetailUseCase {
	return &readCategoryDetailUseCase{
		Log:    log.WithField("action", "usecase:product.read_category_detail"),
		Tracer: trc,
		Runner: runner,
		Repo:   repo,
	}
}

func (uc *readCategoryDetailUseCase) Execute(ctx context.Context, id string) (*CategoryResponse, error) {
	span, ctx := uc.Tracer.StartSpan(ctx, useCaseName)

	defer span.Finish()

	category, err := uc.Repo.CategoryQry.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, entity.ErrCategoryNotFound
	}

	return &CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
	}, nil
}
