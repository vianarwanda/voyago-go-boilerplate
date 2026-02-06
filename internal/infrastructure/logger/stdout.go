package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/ctxkey"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/pkg/utils"

	"github.com/lmittmann/tint"
)

type stdoutLogger struct {
	handler slog.Handler
	logger  *slog.Logger
	tracer  tracer.Tracer
}

var _ Logger = (*stdoutLogger)(nil)

func NewStdoutLogger(config *config.Config, trc tracer.Tracer) Logger {
	var slogLevel slog.Level
	switch config.Log.Level {
	case 6: // Trace
		slogLevel = slog.LevelDebug - 4
	case 5: // Debug
		slogLevel = slog.LevelDebug
	case 4: // Info
		slogLevel = slog.LevelInfo
	case 3: // Warn
		slogLevel = slog.LevelWarn
	case 2: // Error
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	baseHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slogLevel,
		TimeFormat: time.RFC1123,
	})
	maskingHandler := NewMaskingHandler(baseHandler)

	return &stdoutLogger{
		handler: maskingHandler,
		logger:  slog.New(maskingHandler),
		tracer:  trc,
	}
}

func (l *stdoutLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	var args []any
	if requestID := ctxkey.GetRequestID(ctx); requestID != "" {
		args = append(args, slog.String("request_id", requestID))
	}

	if l.tracer != nil {
		if traceID, spanID, ok := l.tracer.ExtractTraceInfo(ctx); ok {
			args = append(args,
				slog.String("trace_id", traceID),
				slog.String("span_id", spanID),
			)
		}
	}

	if len(args) > 0 {
		return &stdoutLogger{
			handler: l.handler,
			logger:  l.logger.With(args...),
			tracer:  l.tracer,
		}
	}

	return l
}

func (l *stdoutLogger) WithField(key string, value any) Logger {
	newLogger := l.logger.With(slog.Any(key, value))
	return &stdoutLogger{handler: l.handler, logger: newLogger, tracer: l.tracer}
}

func (l *stdoutLogger) WithFields(fields map[string]any) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	newLogger := l.logger.With(args...)
	return &stdoutLogger{handler: l.handler, logger: newLogger, tracer: l.tracer}
}

func (l *stdoutLogger) Debug(msg string) { l.logger.Debug(msg) }
func (l *stdoutLogger) Info(msg string)  { l.logger.Info(msg) }
func (l *stdoutLogger) Warn(msg string)  { l.logger.Warn(msg) }
func (l *stdoutLogger) Error(msg string) { l.logger.Error(msg) }

// ------- MASKING HANDLER -------

type MaskingHandler struct {
	next slog.Handler
}

func NewMaskingHandler(next slog.Handler) *MaskingHandler {
	return &MaskingHandler{next: next}
}

func (h *MaskingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *MaskingHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. Mask the Message
	if len(r.Message) > utils.MaxFieldSize {
		r.Message = "[message too large to log]"
	} else if utils.ContainsSensitiveToken(r.Message) {
		r.Message = "******** [REDACTED]"
	}

	// 2. Create a new record to hold masked attributes
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	r.Attrs(func(a slog.Attr) bool {
		maskedAttr := h.maskAttr(a)
		newRecord.AddAttrs(maskedAttr)
		return true
	})

	return h.next.Handle(ctx, newRecord)
}

func (h *MaskingHandler) maskAttr(a slog.Attr) slog.Attr {
	// If key is sensitive, redact immediately
	if utils.IsSensitiveKey(a.Key) {
		return slog.String(a.Key, "******** [REDACTED]")
	}

	// For nested groups, mask recursively
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		maskedGroup := make([]slog.Attr, len(attrs))
		for i, attr := range attrs {
			maskedGroup[i] = h.maskAttr(attr)
		}
		return slog.Group(a.Key, anyToAnySlice(maskedGroup)...)
	}

	// Use your utils.MaskSensitive for everything else
	maskedValue := utils.MaskSensitive(a.Value.Any())
	return slog.Any(a.Key, maskedValue)
}

// Helper to convert Attrs for slog.Group
func anyToAnySlice(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, v := range attrs {
		result[i] = v
	}
	return result
}

func (h *MaskingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Mask attributes when .With() or .WithFields() is called
	masked := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		masked[i] = h.maskAttr(a)
	}
	return &MaskingHandler{next: h.next.WithAttrs(masked)}
}

func (h *MaskingHandler) WithGroup(name string) slog.Handler {
	return &MaskingHandler{next: h.next.WithGroup(name)}
}
