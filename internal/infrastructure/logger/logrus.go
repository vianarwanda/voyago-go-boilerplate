package logger

import (
	"context"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/ctxkey"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/pkg/utils"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logrusLogger struct {
	log    *logrus.Entry
	tracer tracer.Tracer
}

var _ Logger = (*logrusLogger)(nil)

func NewLogrus(cfg *config.Config, trc tracer.Tracer) Logger {
	baseLogger := logrus.New()
	baseLogger.SetFormatter(&logrus.JSONFormatter{})
	baseLogger.SetLevel(logrus.Level(cfg.Log.Level))

	baseLogger.SetOutput(&lumberjack.Logger{
		Filename:   cfg.Log.Path,
		MaxSize:    cfg.Log.Rotation.MaxSize,
		MaxBackups: cfg.Log.Rotation.MaxBackup,
		MaxAge:     cfg.Log.Rotation.MaxAge,
		Compress:   cfg.Log.Rotation.Compress,
	})

	baseLogger.AddHook(NewMaskingHook())

	return &logrusLogger{
		log:    logrus.NewEntry(baseLogger),
		tracer: trc,
	}
}

func (l *logrusLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	requestID := ctxkey.GetRequestID(ctx)
	fields := logrus.Fields{}

	if requestID != "" && requestID != "unknown" {
		fields["request_id"] = requestID
	}

	// Extract Trace & Span IDs for log correlation
	if l.tracer != nil {
		traceID, spanID, ok := l.tracer.ExtractTraceInfo(ctx)
		if ok {
			fields["trace_id"] = traceID
			fields["span_id"] = spanID
		}
	}

	if len(fields) > 0 {
		return &logrusLogger{
			log:    l.log.WithFields(fields),
			tracer: l.tracer,
		}
	}

	return l
}

func (l *logrusLogger) WithField(key string, value any) Logger {
	return &logrusLogger{
		log:    l.log.WithField(key, value),
		tracer: l.tracer,
	}
}

func (l *logrusLogger) WithFields(fields map[string]any) Logger {
	return &logrusLogger{
		log:    l.log.WithFields(fields),
		tracer: l.tracer,
	}
}

func (l *logrusLogger) Debug(message string) { l.log.Debug(message) }
func (l *logrusLogger) Info(message string)  { l.log.Info(message) }
func (l *logrusLogger) Warn(message string)  { l.log.Warn(message) }
func (l *logrusLogger) Error(message string) { l.log.Error(message) }

// --- Masking Hook Implementation ---

type MaskingHook struct{}

func NewMaskingHook() *MaskingHook {
	return &MaskingHook{}
}

func (h *MaskingHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *MaskingHook) Fire(entry *logrus.Entry) error {
	for k, v := range entry.Data {
		if utils.IsSensitiveKey(k) {
			entry.Data[k] = "******** [REDACTED]"
			continue
		}
		entry.Data[k] = utils.MaskSensitive(v)
	}

	if len(entry.Message) > utils.MaxFieldSize {
		entry.Message = "[message too large to log]"
	} else {
		if utils.ContainsSensitiveToken(entry.Message) {
			entry.Message = "******** [REDACTED]"
		}
	}

	return nil
}
