package app

import (
	"fmt"
	"time"
	"voyago/core-api/internal/infrastructure/config"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/middleware"
	"voyago/core-api/internal/infrastructure/telemetry/metrics"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/infrastructure/validator"
	"voyago/core-api/internal/modules/booking"

	"github.com/gofiber/fiber/v2"
)

var domains = [1]string{
	"booking",
	// "merchant",
}

type BootstrapApiConfig struct {
	App     *fiber.App
	Val     validator.Validator
	Log     logger.Logger
	Tracer  tracer.Tracer
	Metrics metrics.Metrics

	configs map[string]*config.Config
	loggers map[string]logger.Logger
	dbs     map[string]database.Database
}

func (b *BootstrapApiConfig) Run() {
	b.setupMiddleware()
	b.setupInfrastructureModules()
	b.setupModules()
	b.setupHealthRoute()
}

func (b *BootstrapApiConfig) Stop() {
	for _, domain := range domains {
		log, okLog := b.loggers[domain]
		db, okDb := b.dbs[domain]

		if !okLog || log == nil {
			log = b.Log // Fallback to global logger
		}

		if !okDb || db == nil {
			log.WithFields(map[string]any{
				"domain":    domain,
				"component": "database",
			}).Warn("Database connection not found during shutdown")
			continue
		}

		if err := db.Close(); err != nil {
			log.WithFields(map[string]any{
				"domain":       domain,
				"component":    "database",
				"error_detail": err.Error(),
			}).Error("Failed to close database connection")
		} else {
			log.WithFields(map[string]any{
				"domain":    domain,
				"component": "database",
			}).Info("Database connection closed gracefully")
		}
	}
}

func (b *BootstrapApiConfig) setupMiddleware() {
	t := middleware.NewTelemetrist(b.Log, b.Tracer, b.Metrics)

	b.App.Use(middleware.RequestID())
	b.App.Use(t.HandleMetrics())
	b.App.Use(t.HandleTrace())
	b.App.Use(t.HandleLog())
}

func (b *BootstrapApiConfig) setupInfrastructureModules() {
	domainCount := len(domains)
	b.configs = make(map[string]*config.Config, domainCount)
	b.loggers = make(map[string]logger.Logger, domainCount)
	b.dbs = make(map[string]database.Database, domainCount)

	for _, domain := range domains {
		path := fmt.Sprintf("config/%s/config.yaml", domain)
		domainCfg := config.LoadDomainConfig(path)

		// 1. Logger
		domainLogger := logger.
			New(domainCfg, b.Tracer).
			WithFields(map[string]any{
				"service": domainCfg.App.Name,
				"version": domainCfg.App.Version,
				"env":     domainCfg.App.Env,
				"port":    domainCfg.Http.Port,
				"domain":  domain,
			})

		// 2. Database
		db := database.NewDatabase(&domainCfg.Database, domainLogger, b.Tracer)

		b.configs[domain] = domainCfg
		b.loggers[domain] = domainLogger
		b.dbs[domain] = db
	}
}

func (b *BootstrapApiConfig) setupModules() {
	var m string

	// --- Booking Module ---
	m = "booking"
	if cfg, ok := b.configs[m]; ok {
		booking.RegisterModule(booking.ModuleConfig{
			Config: cfg,
			Server: b.App,
			DB:     b.dbs[m],
			Log:    b.loggers[m],
			Val:    b.Val,
			Tracer: b.Tracer,
		})
	}
}

func (b *BootstrapApiConfig) setupHealthRoute() {
	h := func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
		})
	}

	b.App.Get("/", h)
	b.App.Get("/health", h)
}
