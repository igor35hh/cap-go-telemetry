package config

import (
	"os"
	"strings"
)

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		Disabled:    getEnvBool("NO_TELEMETRY", false),
		ServiceName: getEnvString("OTEL_SERVICE_NAME", "CAP Application"),
		Kind:        getEnvString("TELEMETRY_KIND", "telemetry-to-console"),
		Tracing:     NewDefaultTracingConfig(),
		Metrics:     NewDefaultMetricsConfig(),
		Logging:     NewDefaultLoggingConfig(),
		Instrumentations: map[string]*InstrumentationConfig{
			"http": {
				Module:  "otelhttp",
				Class:   "HTTPInstrumentation",
				Enabled: true,
				Config:  make(map[string]interface{}),
			},
		},
	}
}

// NewDefaultTracingConfig creates default tracing configuration
func NewDefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		Enabled:    true,
		HRTime:     getEnvBool("TELEMETRY_HRTIME", false),
		TxEnabled:  false,
		HanaPrompt: true,
		Sampler: &SamplerConfig{
			Kind: "ParentBasedSampler",
			Root: "AlwaysOnSampler",
			IgnoreIncomingPaths: []string{
				"/health",
				"/metrics",
				"/ready",
			},
		},
		Exporter: &ExporterConfig{
			Module: "console",
			Class:  "ConsoleSpanExporter",
			Config: make(map[string]interface{}),
		},
	}
}

// NewDefaultMetricsConfig creates default metrics configuration
func NewDefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:        true,
		DBPool:         true,
		Queue:          true,
		HostMetrics:    getEnvBool("HOST_METRICS_ENABLED", true),
		RuntimeMetrics: true,
		Config: &MetricsExportConfig{
			ExportIntervalMillis: 60000, // 60 seconds
		},
		Exporter: &ExporterConfig{
			Module: "console",
			Class:  "ConsoleMetricExporter",
			Config: make(map[string]interface{}),
		},
	}
}

// NewDefaultLoggingConfig creates default logging configuration
func NewDefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Enabled: false, // Disabled by default, opt-in
		Exporter: &ExporterConfig{
			Module: "console",
			Class:  "ConsoleLogExporter",
			Config: make(map[string]interface{}),
		},
	}
}

// GetPredefinedKinds returns all predefined telemetry kinds
func GetPredefinedKinds() map[string]*PredefinedKind {
	return map[string]*PredefinedKind{
		"telemetry-to-console": {
			Name: "telemetry-to-console",
			Tracing: &TracingConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "console",
					Class:  "ConsoleSpanExporter",
				},
			},
			Metrics: &MetricsConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "console",
					Class:  "ConsoleMetricExporter",
				},
			},
		},
		"telemetry-to-dynatrace": {
			Name:      "telemetry-to-dynatrace",
			TokenName: "ingest_apitoken",
			VCAP: &VCAPConfig{
				Label: "dynatrace",
			},
			Tracing: &TracingConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "otlp",
					Class:  "OTLPTraceExporter",
				},
			},
			Metrics: &MetricsConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "otlp",
					Class:  "OTLPMetricExporter",
				},
			},
		},
		"telemetry-to-cloud-logging": {
			Name: "telemetry-to-cloud-logging",
			VCAP: &VCAPConfig{
				Label: "cloud-logging",
			},
			Tracing: &TracingConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "otlp-grpc",
					Class:  "OTLPTraceExporter",
				},
			},
			Metrics: &MetricsConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "otlp-grpc",
					Class:  "OTLPMetricExporter",
				},
			},
		},
		"telemetry-to-jaeger": {
			Name: "telemetry-to-jaeger",
			Tracing: &TracingConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "jaeger",
					Class:  "JaegerExporter",
				},
			},
		},
		"telemetry-to-otlp": {
			Name: "telemetry-to-otlp",
			Tracing: &TracingConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "otlp-env",
					Class:  "OTLPTraceExporter",
				},
			},
			Metrics: &MetricsConfig{
				Enabled: true,
				Exporter: &ExporterConfig{
					Module: "otlp-env",
					Class:  "OTLPMetricExporter",
				},
			},
		},
	}
}

// Helper functions for environment variable parsing

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		// Consider "false", "0", "" as false, everything else as true
		switch strings.ToLower(value) {
		case "false", "0", "":
			return false
		default:
			return true
		}
	}
	return defaultValue
}
