package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles configuration loading from multiple sources
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	v := viper.New()

	// Set default configuration file names and paths
	v.SetConfigName("telemetry")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.cap-go-telemetry")
	v.AddConfigPath("/etc/cap-go-telemetry")

	// Enable environment variable support
	v.SetEnvPrefix("TELEMETRY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	return &Loader{v: v}
}

// Load loads configuration from multiple sources in order of precedence:
// 1. Environment variables
// 2. Configuration file
// 3. Defaults
func (l *Loader) Load() (*Config, error) {
	// Start with defaults
	config := NewDefaultConfig()

	// Try to read config file (optional)
	if err := l.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	// Unmarshal into our config struct
	if err := l.v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply predefined kind if specified
	if config.Kind != "" {
		if err := l.applyPredefinedKind(config); err != nil {
			return nil, fmt.Errorf("failed to apply predefined kind %s: %w", config.Kind, err)
		}
	}

	// Validate configuration
	if err := l.validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadFromFile loads configuration from a specific file
func (l *Loader) LoadFromFile(filename string) (*Config, error) {
	l.v.SetConfigFile(filename)
	return l.Load()
}

// LoadFromJSON loads configuration from JSON string
func (l *Loader) LoadFromJSON(jsonStr string) (*Config, error) {
	config := NewDefaultConfig()

	if err := json.Unmarshal([]byte(jsonStr), config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	if err := l.validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// applyPredefinedKind applies a predefined configuration kind
func (l *Loader) applyPredefinedKind(config *Config) error {
	kinds := GetPredefinedKinds()
	predefined, exists := kinds[config.Kind]
	if !exists {
		return fmt.Errorf("unknown predefined kind: %s", config.Kind)
	}

	// Merge predefined configuration, but don't override explicit settings
	if config.Tracing == nil && predefined.Tracing != nil {
		config.Tracing = predefined.Tracing
	} else if config.Tracing != nil && predefined.Tracing != nil {
		// Merge tracing config
		if config.Tracing.Exporter == nil {
			config.Tracing.Exporter = predefined.Tracing.Exporter
		}
	}

	if config.Metrics == nil && predefined.Metrics != nil {
		config.Metrics = predefined.Metrics
	} else if config.Metrics != nil && predefined.Metrics != nil {
		// Merge metrics config
		if config.Metrics.Exporter == nil {
			config.Metrics.Exporter = predefined.Metrics.Exporter
		}
	}

	if config.Logging == nil && predefined.Logging != nil {
		config.Logging = predefined.Logging
	} else if config.Logging != nil && predefined.Logging != nil {
		// Merge logging config
		if config.Logging.Exporter == nil {
			config.Logging.Exporter = predefined.Logging.Exporter
		}
	}

	return nil
}

// validateConfig validates the loaded configuration
func (l *Loader) validateConfig(config *Config) error {
	if config.ServiceName == "" {
		config.ServiceName = "CAP Application"
	}

	// Validate tracing configuration
	if config.Tracing != nil && config.Tracing.Enabled {
		if config.Tracing.Sampler == nil {
			return fmt.Errorf("tracing sampler configuration is required when tracing is enabled")
		}
		if config.Tracing.Exporter == nil {
			return fmt.Errorf("tracing exporter configuration is required when tracing is enabled")
		}
	}

	// Validate metrics configuration
	if config.Metrics != nil && config.Metrics.Enabled {
		if config.Metrics.Exporter == nil {
			return fmt.Errorf("metrics exporter configuration is required when metrics is enabled")
		}
		if config.Metrics.Config == nil {
			config.Metrics.Config = &MetricsExportConfig{
				ExportIntervalMillis: 60000,
			}
		}
	}

	// Validate logging configuration
	if config.Logging != nil && config.Logging.Enabled {
		if config.Logging.Exporter == nil {
			return fmt.Errorf("logging exporter configuration is required when logging is enabled")
		}
	}

	return nil
}

// GetConfigFile returns the path to the configuration file being used
func (l *Loader) GetConfigFile() string {
	return l.v.ConfigFileUsed()
}

// WriteConfigFile writes the current configuration to a file
func (l *Loader) WriteConfigFile(config *Config, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
