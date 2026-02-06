package metrics

import "time"

type noOpMetrics struct{}

var _ Metrics = (*noOpMetrics)(nil)

func NewNoOpMetrics() Metrics                                                 { return &noOpMetrics{} }
func (m *noOpMetrics) Incr(name string, tags []string)                        {}
func (m *noOpMetrics) Distribution(name string, value float64, tags []string) {}
func (m *noOpMetrics) Timing(name string, value time.Duration, tags []string) {}
func (m *noOpMetrics) RecordHTTP(method string, path string, routePath string, status int, duration float64) {
}
func (m *noOpMetrics) Close() error { return nil }
