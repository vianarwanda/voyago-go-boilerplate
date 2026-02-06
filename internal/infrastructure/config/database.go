package config

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Pool     struct {
		Idle     int `mapstructure:"idle"`
		Max      int `mapstructure:"max"`
		Lifetime int `mapstructure:"lifetime"`
	} `mapstructure:"pool"`
}
