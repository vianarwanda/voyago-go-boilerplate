package config

type TelemetryConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	Type           string  `mapstructure:"type"`
	MetricsAddress string  `mapstructure:"metrics_address"`
	TracerAddress  string  `mapstructure:"tracer_address"`
	Namespace      string  `mapstructure:"namespace"`
	SampleRate     float64 `mapstructure:"sample_rate"`
}
