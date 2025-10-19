package config

import (
	"time"
)

// Config represents the main telemetry configuration
type Config struct {
	// Global settings
	Disabled    bool   `mapstructure:"disabled" yaml:"disabled" json:"disabled"`
	ServiceName string `mapstructure:"service_name" yaml:"service_name" json:"service_name"`
	Kind        string `mapstructure:"kind" yaml:"kind" json:"kind"`

	// Telemetry signals
	Tracing *TracingConfig `mapstructure:"tracing" yaml:"tracing" json:"tracing"`
	Metrics *MetricsConfig `mapstructure:"metrics" yaml:"metrics" json:"metrics"`
	Logging *LoggingConfig `mapstructure:"logging" yaml:"logging" json:"logging"`

	// Instrumentations
	Instrumentations map[string]*InstrumentationConfig `mapstructure:"instrumentations" yaml:"instrumentations" json:"instrumentations"`
}

// TracingConfig configures distributed tracing
type TracingConfig struct {
	Enabled    bool            `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Sampler    *SamplerConfig  `mapstructure:"sampler" yaml:"sampler" json:"sampler"`
	Exporter   *ExporterConfig `mapstructure:"exporter" yaml:"exporter" json:"exporter"`
	HRTime     bool            `mapstructure:"hrtime" yaml:"hrtime" json:"hrtime"`
	TxEnabled  bool            `mapstructure:"_tx" yaml:"_tx" json:"_tx"`
	HanaPrompt bool            `mapstructure:"_hana_prom" yaml:"_hana_prom" json:"_hana_prom"`
}

// MetricsConfig configures metrics collection
type MetricsConfig struct {
	Enabled        bool                 `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Exporter       *ExporterConfig      `mapstructure:"exporter" yaml:"exporter" json:"exporter"`
	Config         *MetricsExportConfig `mapstructure:"config" yaml:"config" json:"config"`
	DBPool         bool                 `mapstructure:"_db_pool" yaml:"_db_pool" json:"_db_pool"`
	Queue          bool                 `mapstructure:"_queue" yaml:"_queue" json:"_queue"`
	HostMetrics    bool                 `mapstructure:"host_metrics" yaml:"host_metrics" json:"host_metrics"`
	RuntimeMetrics bool                 `mapstructure:"runtime_metrics" yaml:"runtime_metrics" json:"runtime_metrics"`
}

// LoggingConfig configures logging export
type LoggingConfig struct {
	Enabled  bool            `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Exporter *ExporterConfig `mapstructure:"exporter" yaml:"exporter" json:"exporter"`
}

// SamplerConfig configures trace sampling
type SamplerConfig struct {
	Kind                string   `mapstructure:"kind" yaml:"kind" json:"kind"`
	Root                string   `mapstructure:"root" yaml:"root" json:"root"`
	Ratio               float64  `mapstructure:"ratio" yaml:"ratio" json:"ratio"`
	IgnoreIncomingPaths []string `mapstructure:"ignore_incoming_paths" yaml:"ignore_incoming_paths" json:"ignore_incoming_paths"`
}

// ExporterConfig configures telemetry exporters
type ExporterConfig struct {
	Module string                 `mapstructure:"module" yaml:"module" json:"module"`
	Class  string                 `mapstructure:"class" yaml:"class" json:"class"`
	Config map[string]interface{} `mapstructure:"config" yaml:"config" json:"config"`
}

// MetricsExportConfig configures metrics export behavior
type MetricsExportConfig struct {
	ExportIntervalMillis int `mapstructure:"export_interval_millis" yaml:"export_interval_millis" json:"export_interval_millis"`
}

// InstrumentationConfig configures individual instrumentations
type InstrumentationConfig struct {
	Module  string                 `mapstructure:"module" yaml:"module" json:"module"`
	Class   string                 `mapstructure:"class" yaml:"class" json:"class"`
	Config  map[string]interface{} `mapstructure:"config" yaml:"config" json:"config"`
	Enabled bool                   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
}

// PredefinedKind represents a predefined telemetry configuration
type PredefinedKind struct {
	Name      string         `yaml:"name" json:"name"`
	Tracing   *TracingConfig `yaml:"tracing" json:"tracing"`
	Metrics   *MetricsConfig `yaml:"metrics" json:"metrics"`
	Logging   *LoggingConfig `yaml:"logging" json:"logging"`
	VCAP      *VCAPConfig    `yaml:"vcap" json:"vcap"`
	TokenName string         `yaml:"token_name" json:"token_name"`
}

// VCAPConfig for cloud foundry service binding
type VCAPConfig struct {
	Label string `yaml:"label" json:"label"`
}

// GetExportInterval returns the metrics export interval as a duration
func (m *MetricsExportConfig) GetExportInterval() time.Duration {
	if m.ExportIntervalMillis <= 0 {
		return 60 * time.Second // Default to 60 seconds
	}
	return time.Duration(m.ExportIntervalMillis) * time.Millisecond
}

// IsEnabled returns whether the given configuration is enabled
func (c *Config) IsEnabled() bool {
	return !c.Disabled
}

// IsTracingEnabled returns whether tracing is enabled
func (c *Config) IsTracingEnabled() bool {
	return c.IsEnabled() && c.Tracing != nil && c.Tracing.Enabled
}

// IsMetricsEnabled returns whether metrics are enabled
func (c *Config) IsMetricsEnabled() bool {
	return c.IsEnabled() && c.Metrics != nil && c.Metrics.Enabled
}

// IsLoggingEnabled returns whether logging is enabled
func (c *Config) IsLoggingEnabled() bool {
	return c.IsEnabled() && c.Logging != nil && c.Logging.Enabled
}
