package config

type LogConfig struct {
	Path     string `mapstructure:"path"`
	Level    int    `mapstructure:"level"`
	Rotation struct {
		MaxSize   int  `mapstructure:"max_size"`
		MaxBackup int  `mapstructure:"max_backup"`
		MaxAge    int  `mapstructure:"max_age"`
		Compress  bool `mapstructure:"compress"`
	} `mapstructure:"rotation"`
}
