package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"voyago/core-api/internal/app"
	"voyago/core-api/internal/infrastructure/config"
	server "voyago/core-api/internal/infrastructure/http"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/metrics"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/infrastructure/validator"
)

func main() {
	// ----- Load config -----
	globalCfgPath := "config/config.yaml"
	globalCfg := config.InitGlobalConfig(globalCfgPath)
	// ----- Load config -----

	// ----- Initialize validator -----
	val := validator.NewPlaygroundValidator()
	// ----- Initialize validator -----

	// ----- Initialize global logger -----
	log := logger.New(globalCfg, nil)
	appLogger := log.WithFields(map[string]any{
		"service": globalCfg.App.Name,
		"version": globalCfg.App.Version,
		"env":     globalCfg.App.Env,
		"port":    globalCfg.Http.Port,
		"domain":  "main",
	})
	// ----- Initialize global logger -----

	// ----- Initialize metrics -----
	metrics, err := metrics.New(
		&globalCfg.Telemetry,
		globalCfg.App.Env,
	)
	if err != nil {
		panic(err)
	}
	defer metrics.Close()
	// ----- Initialize metrics -----

	// ----- Initialize tracer -----
	tracer, err := tracer.New(
		&globalCfg.Telemetry,
		globalCfg.App.Env,
	)
	if err != nil {
		panic(err)
	}
	defer tracer.Close()
	// ----- Initialize tracer -----

	l := appLogger.WithField("component", "app")
	l.Info("Application starting")

	if globalCfg.Telemetry.Enabled {
		l.Info(fmt.Sprintf("Telemetry config: metrics=%s, tracer=%s, sample_rate=%f",
			globalCfg.Telemetry.MetricsAddress,
			globalCfg.Telemetry.TracerAddress,
			globalCfg.Telemetry.SampleRate))
	}

	srv := server.NewServer(globalCfg, appLogger)
	bootstrap := app.BootstrapHttpConfig{
		App:     srv.App,
		Val:     val,
		Log:     appLogger,
		Tracer:  tracer,
		Metrics: metrics,
	}
	bootstrap.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Stop(ctx); err != nil {
			l.WithFields(map[string]any{
				"error_detail": err.Error(),
			}).Error("Server forced to shutdown")
		}

		// Stop all domain connections (databases, loggers, etc.)
		bootstrap.Stop()
	}()

	if err := srv.Start(); err != nil {
		l.WithFields(map[string]any{
			"error_detail": err.Error(),
		}).Error("failed to start server")
	}
}
