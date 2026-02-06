// Package config handles multi-level configuration loading, environment expansion,
// and domain-specific configuration merging.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// globalViper holds the base configuration state to be used as a template
// for all domain-specific configurations.
var globalViper *viper.Viper

// InitGlobalConfig initializes the base configuration from the provided globalPath.
// It parses the YAML file, expands environment variables, and stores the state internally.
// Use the returned *Config for global infrastructure setup like Telemetry or Global App settings.
//
// Example:
//
//	globalCfg := config.InitGlobalConfig("config/config.yaml")
func InitGlobalConfig(globalPath string) *Config {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	content, err := processingFile(globalPath)
	if err != nil {
		panic(fmt.Errorf("error reading global config: %w", err))
	}

	v.SetConfigType("yaml")
	v.ReadConfig(strings.NewReader(content))

	globalViper = v

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode global config into struct: %v", err))
	}

	return &cfg
}

// LoadDomainConfig creates a domain-specific configuration by merging the global settings
// with the specific settings found in the domainPath.
// It performs a deep copy of the global configuration, ensuring that domain-specific
// overrides do not pollute the global state or other domains.
//
// Example:
//
//	bookingCfg := config.LoadDomainConfig("config/booking/config.yaml")
func LoadDomainConfig(domainPath string) *Config {
	if globalViper == nil {
		panic(fmt.Errorf("ERROR: Global Config is nil! InitGlobalConfig must be called first from the same package."))
	}

	domainViper := viper.New()
	domainViper.AutomaticEnv()
	domainViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := domainViper.MergeConfigMap(globalViper.AllSettings()); err != nil {
		panic(fmt.Errorf("Error merging global settings: %v", err))
	}

	if domainPath != "" {
		content, err := processingFile(domainPath)
		if err != nil {
			panic(fmt.Errorf("failed to load domain config %s: %w", domainPath, err))
		}
		domainViper.SetConfigType("yaml")
		domainViper.MergeConfig(strings.NewReader(content))
	}

	var cfg Config
	if err := domainViper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode domain config into struct: %v", err))
	}
	return &cfg
}

func processingFile(path string) (string, error) {
	actualPath := findActualPath(path)

	content, err := os.ReadFile(actualPath)
	if err != nil {
		return "", err
	}

	return os.Expand(string(content), func(s string) string {
		parts := strings.SplitN(s, ":", 2)
		val := os.Getenv(parts[0])
		if val == "" && len(parts) > 1 {
			return parts[1]
		}
		return val
	}), nil
}

func findActualPath(configPath string) string {
	finalPath := configPath
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		climbPath := fmt.Sprintf("../../%s", configPath)
		if _, err := os.Stat(climbPath); err == nil {
			return climbPath
		}
		parts := strings.Split(configPath, "/")
		flatPath := parts[len(parts)-1]
		if _, err := os.Stat(flatPath); err == nil {
			return flatPath
		}
	}
	return finalPath
}
