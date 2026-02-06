package tracer

import (
	"context"
	"strconv"

	gormtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorm.io/gorm.v1"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gorm.io/gorm"
)

type datadogTracer struct {
	serviceName string
}

type datadogSpan struct {
	span tracer.Span
}

var _ Tracer = (*datadogTracer)(nil)

func NewDatadogTracer(serviceName, env, addr string, sampleRate float64) Tracer {
	tracer.Start(
		tracer.WithService(serviceName),
		tracer.WithEnv(env),
		tracer.WithAgentAddr(addr),
		tracer.WithSampler(tracer.NewRateSampler(sampleRate)),
	)
	return &datadogTracer{serviceName: serviceName}
}

func (t *datadogTracer) StartSpan(ctx context.Context, name string) (Span, context.Context) {
	span, ctx := tracer.StartSpanFromContext(ctx, name)
	return &datadogSpan{span: span}, ctx
}

func (t *datadogTracer) ExtractTraceInfo(ctx context.Context) (traceID, spanID string, ok bool) {
	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return "", "", false
	}
	return strconv.FormatUint(span.Context().TraceID(), 10), strconv.FormatUint(span.Context().SpanID(), 10), true
}

func (t *datadogTracer) UseGorm(db *gorm.DB) {
	db.Use(gormtrace.NewTracePlugin(gormtrace.WithServiceName(t.serviceName + "-db")))
}

func (t *datadogTracer) Close() error {
	tracer.Stop()
	return nil
}

func (s *datadogSpan) SetOperationName(name string) {
	s.span.SetOperationName(name)
}

func (s *datadogSpan) Finish() {
	s.span.Finish()
}

func (s *datadogSpan) SetTag(key string, value any) {
	s.span.SetTag(key, value)
}
