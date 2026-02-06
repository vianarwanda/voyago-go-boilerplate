package config

type Config struct {
	// Global configuration
	App       AppConfig       `mapstructure:"app"`
	Http      HttpConfig      `mapstructure:"http"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`

	// Domain configuration
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
}
