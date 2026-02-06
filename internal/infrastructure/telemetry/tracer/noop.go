package tracer

import (
	"context"

	"gorm.io/gorm"
)

type noOpTracer struct{}

type noOpSpan struct{}

var _ Tracer = (*noOpTracer)(nil)

func NewNoOpTracer() Tracer {
	return &noOpTracer{}
}

func (t *noOpTracer) StartSpan(ctx context.Context, name string) (Span, context.Context) {
	return &noOpSpan{}, ctx
}

func (t *noOpTracer) UseGorm(db *gorm.DB) {}

func (t *noOpTracer) ExtractTraceInfo(ctx context.Context) (traceID, spanID string, ok bool) {
	return "", "", false
}

func (t *noOpTracer) Close() error {
	return nil
}

func (s *noOpSpan) SetOperationName(name string) {}

func (s *noOpSpan) Finish() {}

func (s *noOpSpan) SetTag(key string, value any) {}
