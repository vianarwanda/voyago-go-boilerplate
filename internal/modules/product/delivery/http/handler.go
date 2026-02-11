package http

import (
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/validator"
	"voyago/core-api/internal/modules/product/entity"
	"voyago/core-api/internal/modules/product/usecase"
	"voyago/core-api/internal/pkg/apperror"
	"voyago/core-api/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
)

const (
	handlerName = "http:handler.product"
)

type HandlerUseCases struct {
	ReadCategoryUseCase       usecase.ReadCategoryUseCase
	ReadCategoryDetailUseCase usecase.ReadCategoryDetailUseCase
	CreateCategoryUseCase     usecase.CreateCategoryUseCase
}

type Handler struct {
	Cfg *config.Config
	Log logger.Logger
	Val validator.Validator
	Uc  HandlerUseCases
}

func NewHandler(cfg *config.Config, log logger.Logger, validator validator.Validator, useCases HandlerUseCases) *Handler {
	return &Handler{
		Cfg: cfg,
		Log: log,
		Val: validator,
		Uc:  useCases,
	}
}

func (h *Handler) ReadCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	log := h.Log.WithContext(ctx).WithField("method", "ReadCategory")

	log.Info("request received")

	category, err := h.Uc.ReadCategoryUseCase.Execute(ctx)
	if err != nil {
		return err
	}

	return response.NewResponseApi(c).OK(response.ResponseApi{
		Message: "success",
		Data:    category,
	})

}

func (h *Handler) ReadCategoryDetail(c *fiber.Ctx) error {
	ctx := c.Context()
	log := h.Log.WithContext(ctx).WithField("method", "ReadCategoryDetail")
	log.Info("request received")

	id := c.Params("id")

	if id == "" {
		return entity.ErrCategoryIdRequired
	}

	category, err := h.Uc.ReadCategoryDetailUseCase.Execute(ctx, id)
	if err != nil {
		return err
	}

	return response.NewResponseApi(c).OK(response.ResponseApi{
		Message: "success",
		Data:    category,
	})
}

func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	ctx := c.Context()
	log := h.Log.WithContext(ctx).WithField("method", "ReadCategoryDetail")
	log.Info("request received")

	request := new(usecase.CategoryRequest)
	if err := c.BodyParser(request); err != nil {
		return apperror.ErrCodeMalformedRequest.WithError(err)
	}

	if err := h.Val.Validate(request); err != nil {
		return apperror.ErrCodeInvalidRequest.WithError(err).AddValidationErrors(h.Val.ToDetails(err))
	}

	category, err := h.Uc.CreateCategoryUseCase.Execute(ctx, request)
	if err != nil {
		return err
	}

	return response.NewResponseApi(c).Created(response.ResponseApi{
		Message: "success",
		Data:    category,
	})
}
