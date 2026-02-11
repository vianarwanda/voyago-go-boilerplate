package usecase

import (
	"context"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/modules/product/entity"
	"voyago/core-api/internal/modules/product/repository"
	baserepo "voyago/core-api/internal/pkg/repository"
)

type ReadCategoryRepositories struct {
	CategoryCmd repository.CategoryCommandRepository
	CategoryQry repository.CategoryQueryRepository
}

type readCategoryUseCase struct {
	Log    logger.Logger
	Tracer tracer.Tracer
	Runner baserepo.TransactionManager
	Repo   ReadCategoryRepositories
}

const (
	useCaseName = "usecase:product.read_category"
)

var _ ReadCategoryUseCase = (*readCategoryUseCase)(nil)

func NewReadCategoryUseCase(log logger.Logger, trc tracer.Tracer, runner baserepo.TransactionManager, repo ReadCategoryRepositories) ReadCategoryUseCase {
	return &readCategoryUseCase{
		Log:    log.WithField("action", useCaseName),
		Tracer: trc,
		Runner: runner,
		Repo:   repo,
	}
}

func (uc *readCategoryUseCase) Execute(ctx context.Context) ([]CategoryResponse, error) {
	span, ctx := uc.Tracer.StartSpan(ctx, useCaseName)
	defer span.Finish()

	log := uc.Log.WithField("method", "execute")
	log.Info("usecase started")

	parents, err := uc.Repo.CategoryQry.FindParent(ctx)
	if err != nil {
		return nil, err
	}

	children, err := uc.Repo.CategoryQry.FindChildren(ctx)
	if err != nil {
		return nil, err
	}

	// Group children by parent_id
	childMap := make(map[string][]entity.Category)
	for _, c := range children {
		if c.ParentID == nil {
			continue
		}
		childMap[*c.ParentID] = append(childMap[*c.ParentID], c)
	}

	// Build response tree
	result := make([]CategoryResponse, 0)
	for _, p := range parents {
		resp := CategoryResponse{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Description: p.Description,
		}

		for _, child := range childMap[p.ID] {
			resp.Children = append(resp.Children, CategoryResponse{
				ID:          child.ID,
				Name:        child.Name,
				Slug:        child.Slug,
				Description: child.Description,
			})
		}

		result = append(result, resp)
	}
	return result, nil
}
