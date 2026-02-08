package booking

import (
	"voyago/core-api/internal/infrastructure/config"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/infrastructure/validator"
	"voyago/core-api/internal/modules/booking/delivery/http"
	"voyago/core-api/internal/modules/booking/repository/command"
	"voyago/core-api/internal/modules/booking/repository/query"
	"voyago/core-api/internal/modules/booking/usecase"

	"github.com/gofiber/fiber/v2"
)

type HttpModuleConfig struct {
	Config *config.Config
	Server *fiber.App
	DB     database.Database
	Log    logger.Logger
	Val    validator.Validator
	Tracer tracer.Tracer
}

func RegisterHttpModule(cfg HttpModuleConfig) {
	ucLogger := cfg.Log.WithField("component", "usecase")
	hdlrLogger := cfg.Log.WithField("component", "handler")

	// setup repositories
	bookingCmdRepository := command.NewBookingRepository(cfg.DB)
	bookingQryRepository := query.NewBookingRepository(cfg.DB)

	// setup use cases
	createBookingUseCase := usecase.NewCreateBookingUseCase(
		ucLogger,
		cfg.Tracer,
		cfg.DB,
		usecase.CreateBookingRepositories{
			BookingCmd: bookingCmdRepository,
			BookingQry: bookingQryRepository,
		},
	)

	// setup handler
	h := http.NewHandler(
		cfg.Config,
		hdlrLogger,
		cfg.Val,
		http.HandlerUseCases{
			CreateBookingUseCase: createBookingUseCase,
		},
	)

	routeConfig := http.RouteConfig{
		Server:  cfg.Server,
		Config:  cfg.Config,
		Handler: h,
	}
	routeConfig.Setup()
}
