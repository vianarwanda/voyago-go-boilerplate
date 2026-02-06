package database

import (
	"context"
	"fmt"
	"time"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/ctxkey"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

type gormDatabase struct {
	db *gorm.DB
}

var _ Database = (*gormDatabase)(nil)

func NewGormDatabase(cfg *config.DatabaseConfig, log logger.Logger, trc tracer.Tracer) Database {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger:                 NewGormLoggerBridge(log),
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)

	if err != nil {
		log.Error(fmt.Sprintf("failed to connect database: %v", err))
		panic(err)
	}

	if trc != nil {
		trc.UseGorm(db)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(cfg.Pool.Idle)
	sqlDB.SetMaxOpenConns(cfg.Pool.Max)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(cfg.Pool.Lifetime))

	return &gormDatabase{db: db}
}

func (g *gormDatabase) GetDB() *gorm.DB {
	return g.db
}

func (g *gormDatabase) WithContext(ctx context.Context) *gorm.DB {
	if tx := ctxkey.GetTransaction(ctx); tx != nil {
		if gormTx, ok := tx.(*gorm.DB); ok {
			return gormTx.WithContext(ctx)
		}
	}
	return g.db.WithContext(ctx)
}

func (g *gormDatabase) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (g *gormDatabase) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := ctxkey.SetTransaction(ctx, tx)
		return fn(txCtx)
	})
}

// ----- GORM Logger Bridge -----

type gormLoggerBridge struct {
	Log           logger.Logger
	SlowThreshold time.Duration
}

func NewGormLoggerBridge(l logger.Logger) gormlog.Interface {
	return &gormLoggerBridge{
		Log:           l.WithField("component", "database").WithField("source", "gorm"),
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *gormLoggerBridge) LogMode(level gormlog.LogLevel) gormlog.Interface {
	return l
}

func (l *gormLoggerBridge) Info(ctx context.Context, msg string, data ...any) {
	l.Log.WithContext(ctx).Info(fmt.Sprintf(msg, data...))
}

func (l *gormLoggerBridge) Warn(ctx context.Context, msg string, data ...any) {
	l.Log.WithContext(ctx).Warn(fmt.Sprintf(msg, data...))
}

func (l *gormLoggerBridge) Error(ctx context.Context, msg string, data ...any) {
	l.Log.WithContext(ctx).Error(fmt.Sprintf(msg, data...))
}

func (l *gormLoggerBridge) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	isSlow := elapsed > l.SlowThreshold

	log := l.Log.WithContext(ctx).
		WithFields(map[string]any{
			"db_sql":        utils.MaskSensitive(sql),
			"db_rows":       rows,
			"db_elapsed":    elapsed.String(),
			"db_latency_ms": float64(elapsed.Nanoseconds()) / 1e6,
			"db_slow":       isSlow,
		})

	if err != nil && err != gorm.ErrRecordNotFound {
		log.WithFields(map[string]any{
			"error_type": "database",
			"db_error":   err.Error(),
		}).Error("database query error")
		return
	}

	if isSlow {
		log.Warn("SLOW SQL DETECTED")
		return
	}

	log.Debug("SQL TRACE")
}
