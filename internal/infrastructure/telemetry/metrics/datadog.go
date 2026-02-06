package metrics

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type datadogMetrics struct {
	client *statsd.Client
}

var _ Metrics = (*datadogMetrics)(nil)

func NewDatadogMetrics(addr string, namespace string, tags []string) (Metrics, error) {
	client, err := statsd.New(addr,
		statsd.WithNamespace(namespace),
		statsd.WithTags(tags),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dogstatsd: %w", err)
	}

	return &datadogMetrics{client: client}, nil
}

func (m *datadogMetrics) Incr(name string, tags []string) {
	_ = m.client.Incr(name, tags, 1.0)
}

func (m *datadogMetrics) Distribution(name string, value float64, tags []string) {
	_ = m.client.Distribution(name, value, tags, 1.0)
}

func (m *datadogMetrics) Timing(name string, value time.Duration, tags []string) {
	_ = m.client.Timing(name, value, tags, 1.0)
}

func (m *datadogMetrics) RecordHTTP(method string, path string, routePath string, statusCode int, duration float64) {
	tags := []string{
		fmt.Sprintf("method:%s", method),
		fmt.Sprintf("resource:%s", routePath),
		// fmt.Sprintf("route:%s", routePath),
		fmt.Sprintf("status:%d", statusCode),
		fmt.Sprintf("status_group:%dxx", statusCode/100),
	}
	_ = m.client.Incr("http.request.total", tags, 1.0)
	_ = m.client.Distribution("http.request.duration", duration, tags, 1.0)
}

func (m *datadogMetrics) Close() error {
	return m.client.Close()
}
