package config

import "time"

type HttpConfig struct {
	Port         int           `mapstructure:"port"`
	Prefork      bool          `mapstructure:"prefork"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}
