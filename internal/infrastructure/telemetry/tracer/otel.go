package tracer

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type otelTracer struct {
	provider    *sdktrace.TracerProvider
	tracer      trace.Tracer
	serviceName string
}

type otelSpan struct {
	span trace.Span
}

var _ Tracer = (*otelTracer)(nil)

func NewOTelTracer(serviceName, env, addr string, sampleRate float64) (Tracer, error) {
	ctx := context.Background()

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(addr),
		otlptracegrpc.WithInsecure(), // Use TLS in production!
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service info
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &otelTracer{
		provider:    tp,
		tracer:      tp.Tracer(serviceName),
		serviceName: serviceName,
	}, nil
}

func (t *otelTracer) StartSpan(ctx context.Context, name string) (Span, context.Context) {
	ctx, span := t.tracer.Start(ctx, name)
	return &otelSpan{span: span}, ctx
}

func (t *otelTracer) ExtractTraceInfo(ctx context.Context) (traceID, spanID string, ok bool) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return "", "", false
	}

	sc := span.SpanContext()
	if !sc.IsValid() {
		return "", "", false
	}

	return sc.TraceID().String(), sc.SpanID().String(), true
}

func (t *otelTracer) UseGorm(db *gorm.DB) {
	// Register GORM callbacks for OpenTelemetry tracing
	db.Callback().Create().Before("gorm:create").Register("otel:before_create", t.beforeCallback)
	db.Callback().Query().Before("gorm:query").Register("otel:before_query", t.beforeCallback)
	db.Callback().Update().Before("gorm:update").Register("otel:before_update", t.beforeCallback)
	db.Callback().Delete().Before("gorm:delete").Register("otel:before_delete", t.beforeCallback)

	db.Callback().Create().After("gorm:create").Register("otel:after_create", t.afterCallback)
	db.Callback().Query().After("gorm:query").Register("otel:after_query", t.afterCallback)
	db.Callback().Update().After("gorm:update").Register("otel:after_update", t.afterCallback)
	db.Callback().Delete().After("gorm:delete").Register("otel:after_delete", t.afterCallback)
}

func (t *otelTracer) beforeCallback(db *gorm.DB) {
	ctx, span := t.tracer.Start(db.Statement.Context, "gorm:"+db.Statement.Table,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.table", db.Statement.Table),
		),
	)
	db.Statement.Context = ctx
	db.InstanceSet("otel:span", span)
}

func (t *otelTracer) afterCallback(db *gorm.DB) {
	if val, ok := db.InstanceGet("otel:span"); ok {
		if span, ok := val.(trace.Span); ok {
			if db.Error != nil {
				span.RecordError(db.Error)
				span.SetStatus(codes.Error, db.Error.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}
			sqlStatement := db.Statement.SQL.String()
			// sqlStatementMask := utils.MaskSensitive(sqlStatement)
			span.SetAttributes(
				// attribute.String("db.statement", db.Explain(db.Statement.SQL.String(), db.Statement.Vars...)),
				attribute.String("db.statement", sqlStatement),
				attribute.Int64("db.rows_affected", db.RowsAffected),
			)
			span.End()
		}
	}
}

func (t *otelTracer) Close() error {
	if t.provider != nil {
		return t.provider.Shutdown(context.Background())
	}
	return nil
}

func (s *otelSpan) SetOperationName(name string) {
	s.span.SetName(name)
}

func (s *otelSpan) Finish() {
	s.span.End()
}

func (s *otelSpan) SetTag(key string, value any) {
	switch v := value.(type) {
	case string:
		s.span.SetAttributes(attribute.String(key, v))
	case int:
		s.span.SetAttributes(attribute.Int(key, v))
	case int64:
		s.span.SetAttributes(attribute.Int64(key, v))
	case float64:
		s.span.SetAttributes(attribute.Float64(key, v))
	case bool:
		s.span.SetAttributes(attribute.Bool(key, v))
	default:
		s.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}
