package product

import (
	"voyago/core-api/internal/infrastructure/config"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/infrastructure/validator"
	"voyago/core-api/internal/modules/product/delivery/http"
	"voyago/core-api/internal/modules/product/repository/command"
	"voyago/core-api/internal/modules/product/repository/query"
	"voyago/core-api/internal/modules/product/usecase"

	"github.com/gofiber/fiber/v2"
)

type ModuleConfig struct {
	Config *config.Config
	Server *fiber.App
	DB     database.Database
	Log    logger.Logger
	Val    validator.Validator
	Tracer tracer.Tracer
}

func RegisterModule(cfg ModuleConfig) {
	ucLogger := cfg.Log.WithField("component", "usecase")
	hdlrLogger := cfg.Log.WithField("component", "handler")

	categoryCmdRepository := command.NewCategoryRepository(cfg.DB)
	categoryQryRepository := query.NewCategoryRepository(cfg.DB)

	readCategoryUseCase := usecase.NewReadCategoryUseCase(
		ucLogger,
		cfg.Tracer,
		cfg.DB,
		usecase.ReadCategoryRepositories{
			CategoryCmd: categoryCmdRepository,
			CategoryQry: categoryQryRepository,
		},
	)

	readCategoryDetailUseCase := usecase.NewReadCategoryDetailUseCase(
		ucLogger,
		cfg.Tracer,
		cfg.DB,
		usecase.ReadCategoryDetailRepositories{
			CategoryCmd: categoryCmdRepository,
			CategoryQry: categoryQryRepository,
		},
	)

	createCategoryUseCase := usecase.NewCreateCategoryUseCase(
		ucLogger,
		cfg.Tracer,
		cfg.DB,
		usecase.CreateCategoryRepositories{
			CategoryCmd: categoryCmdRepository,
			CategoryQry: categoryQryRepository,
		},
	)

	h := http.NewHandler(
		cfg.Config,
		hdlrLogger,
		cfg.Val,
		http.HandlerUseCases{
			ReadCategoryUseCase:       readCategoryUseCase,
			ReadCategoryDetailUseCase: readCategoryDetailUseCase,
			CreateCategoryUseCase:     createCategoryUseCase,
		},
	)

	routeConfig := http.RouteConfig{
		Server:  cfg.Server,
		Config:  cfg.Config,
		Handler: h,
	}

	routeConfig.Setup()
}
