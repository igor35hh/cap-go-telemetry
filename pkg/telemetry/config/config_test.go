package config

import (
	"os"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	if config.ServiceName == "" {
		t.Error("Expected ServiceName to have a default value")
	}

	if config.Kind == "" {
		t.Error("Expected Kind to have a default value")
	}

	if config.Tracing == nil {
		t.Error("Expected Tracing config to be initialized")
	}

	if config.Metrics == nil {
		t.Error("Expected Metrics config to be initialized")
	}

	if config.Logging == nil {
		t.Error("Expected Logging config to be initialized")
	}
}

func TestConfigIsEnabled(t *testing.T) {
	// Test enabled by default
	config := NewDefaultConfig()
	if !config.IsEnabled() {
		t.Error("Expected config to be enabled by default")
	}

	// Test disabled
	config.Disabled = true
	if config.IsEnabled() {
		t.Error("Expected config to be disabled when Disabled=true")
	}
}

func TestEnvVarParsing(t *testing.T) {
	// Test NO_TELEMETRY environment variable
	os.Setenv("NO_TELEMETRY", "true")
	defer os.Unsetenv("NO_TELEMETRY")

	config := NewDefaultConfig()
	if !config.Disabled {
		t.Error("Expected config to be disabled when NO_TELEMETRY=true")
	}
}

func TestPredefinedKinds(t *testing.T) {
	kinds := GetPredefinedKinds()

	expectedKinds := []string{
		"telemetry-to-console",
		"telemetry-to-dynatrace",
		"telemetry-to-cloud-logging",
		"telemetry-to-jaeger",
		"telemetry-to-otlp",
	}

	for _, kind := range expectedKinds {
		if _, exists := kinds[kind]; !exists {
			t.Errorf("Expected predefined kind %s to exist", kind)
		}
	}
}

func TestConfigLoader(t *testing.T) {
	loader := NewLoader()

	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Verify default values are applied
	if config.ServiceName == "" {
		t.Error("Expected ServiceName to have a value")
	}
}

func TestMetricsExportInterval(t *testing.T) {
	config := &MetricsExportConfig{
		ExportIntervalMillis: 30000,
	}

	interval := config.GetExportInterval()
	expected := 30 * 1000 * 1000 * 1000 // 30 seconds in nanoseconds
	if interval.Nanoseconds() != int64(expected) {
		t.Errorf("Expected interval %d, got %d", expected, interval.Nanoseconds())
	}
}
