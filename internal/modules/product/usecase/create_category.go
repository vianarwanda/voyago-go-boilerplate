package usecase

import (
	"context"
	"time"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/modules/product/entity"
	"voyago/core-api/internal/modules/product/repository"
	baserepo "voyago/core-api/internal/pkg/repository"
	"voyago/core-api/internal/pkg/uid"
	"voyago/core-api/internal/pkg/utils"
)

type CreateCategoryRepositories struct {
	CategoryCmd repository.CategoryCommandRepository
	CategoryQry repository.CategoryQueryRepository
}

type createCategoryUseCase struct {
	Log    logger.Logger
	Tracer tracer.Tracer
	Runner baserepo.TransactionManager
	Repo   CreateCategoryRepositories
}

var _ CreateCategoryUseCase = (*createCategoryUseCase)(nil)

func NewCreateCategoryUseCase(log logger.Logger, trc tracer.Tracer, runner baserepo.TransactionManager, repo CreateCategoryRepositories) CreateCategoryUseCase {
	return &createCategoryUseCase{
		Log:    log.WithField("action", useCaseName),
		Tracer: trc,
		Runner: runner,
		Repo:   repo,
	}
}

func (uc *createCategoryUseCase) Execute(ctx context.Context, req *CategoryRequest) (*CategoryResponse, error) {
	span, ctx := uc.Tracer.StartSpan(ctx, useCaseName)
	defer span.Finish()

	e := entity.Category{
		ID:          uid.NewUUID(),
		ParentID:    req.ParentId,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		IconURL:     nil,
		SortOrder:   0,
		CreatedAt:   time.Time{},
	}

	errRunner := uc.Runner.Atomic(ctx, func(txCtx context.Context) error {
		if err := uc.Repo.CategoryCmd.Create(txCtx, &e); err != nil {
			return err
		}
		return nil
	})

	if errRunner != nil {
		// [STANDARD ERROR HANDLING]: BUBBLE UP
		// We only record the span error to ensure the trace reflects the failure.
		// Logging is already handled by the Repository/DB bridge.
		utils.RecordSpanError(span, errRunner)
		return nil, errRunner
	}

	return &CategoryResponse{
		ID:          e.ID,
		Name:        e.Name,
		Slug:        e.Slug,
		Description: e.Description,
		Children:    nil,
	}, nil
}
