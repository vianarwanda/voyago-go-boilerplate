package metrics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type otelMetrics struct {
	provider *sdkmetric.MeterProvider
	meter    metric.Meter
	counters sync.Map
	histos   sync.Map
}

var _ Metrics = (*otelMetrics)(nil)

func NewOTelMetrics(addr, namespace string, tags []string) (Metrics, error) {
	ctx := context.Background()

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(addr),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithTimeout(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metrics exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(namespace),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	httpView := sdkmetric.NewView(
		sdkmetric.Instrument{
			Name: "http_request_duration", // Pastikan ini sesuai hasil sanitizeName (http.request.duration -> http_request_duration)
		},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{
					0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
				},
			},
		},
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(5*time.Second),
		)),
		sdkmetric.WithResource(res),
		sdkmetric.WithView(httpView),
	)

	otel.SetMeterProvider(mp)

	if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
	}

	return &otelMetrics{
		provider: mp,
		meter:    mp.Meter(namespace),
	}, nil
}

func (m *otelMetrics) sanitizeName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

func (m *otelMetrics) parseAttributes(tags []string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(tags))
	for _, t := range tags {
		parts := strings.SplitN(t, ":", 2)
		if len(parts) == 2 {
			attrs = append(attrs, attribute.String(parts[0], parts[1]))
		} else {
			attrs = append(attrs, attribute.String("tag", t))
		}
	}
	return attrs
}

func (m *otelMetrics) recordDistributionWithAttributes(name string, val float64, attrs []attribute.KeyValue) {
	cleanName := m.sanitizeName(name)
	var histogram metric.Float64Histogram
	if v, ok := m.histos.Load(cleanName); ok {
		histogram = v.(metric.Float64Histogram)
	} else {
		histogram, _ = m.meter.Float64Histogram(cleanName)
		m.histos.Store(cleanName, histogram)
	}
	histogram.Record(context.Background(), val, metric.WithAttributes(attrs...))
}

func (m *otelMetrics) recordWithAttributes(name string, val int64, attrs []attribute.KeyValue) {
	cleanName := m.sanitizeName(name)
	var counter metric.Int64Counter
	if v, ok := m.counters.Load(cleanName); ok {
		counter = v.(metric.Int64Counter)
	} else {
		counter, _ = m.meter.Int64Counter(cleanName)
		m.counters.Store(cleanName, counter)
	}
	counter.Add(context.Background(), val, metric.WithAttributes(attrs...))
}

func (m *otelMetrics) Incr(name string, tags []string) {
	cleanName := m.sanitizeName(name)

	var counter metric.Int64Counter
	if val, ok := m.counters.Load(cleanName); ok {
		counter = val.(metric.Int64Counter)
	} else {
		var err error
		counter, err = m.meter.Int64Counter(cleanName, metric.WithDescription("Total count of "+name))
		if err != nil {
			return
		}
		m.counters.Store(cleanName, counter)
	}

	counter.Add(context.Background(), 1, metric.WithAttributes(m.parseAttributes(tags)...))
}

func (m *otelMetrics) Timing(name string, value time.Duration, tags []string) {
	m.Distribution(name+"_duration", value.Seconds(), tags)
}

func (m *otelMetrics) Distribution(name string, value float64, tags []string) {
	cleanName := m.sanitizeName(name)

	var histogram metric.Float64Histogram
	if val, ok := m.histos.Load(cleanName); ok {
		histogram = val.(metric.Float64Histogram)
	} else {
		var err error
		histogram, err = m.meter.Float64Histogram(cleanName, metric.WithDescription("Distribution of "+name))
		if err != nil {
			return
		}
		m.histos.Store(cleanName, histogram)
	}

	histogram.Record(context.Background(), value, metric.WithAttributes(m.parseAttributes(tags)...))
}

func (m *otelMetrics) RecordHTTP(method string, path string, routePath string, statusCode int, duration float64) {
	// Standard attributes based on OTel semantic conventions
	tags := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", routePath),
		// attribute.String("http.route_path", routePath),
		attribute.Int("http.status_code", statusCode),
	}

	// m.Incr("http.request.total", nil)
	m.recordWithAttributes("http.request.total", 1, tags)
	m.recordDistributionWithAttributes("http.request.duration", duration, tags)
}

func (m *otelMetrics) Close() error {
	if m.provider != nil {
		return m.provider.Shutdown(context.Background())
	}
	return nil
}
