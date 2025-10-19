package console

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

func TestLogExporter_Export(t *testing.T) {
	buf := &bytes.Buffer{}
	exporter := NewLogExporter(WithLogWriter(buf))

	// Create test log records
	records := []sdklog.Record{
		createTestLogRecord(log.SeverityInfo, "Test info message"),
		createTestLogRecord(log.SeverityError, "Test error message"),
	}

	err := exporter.Export(context.Background(), records)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test info message") {
		t.Error("Output doesn't contain info message")
	}
	if !strings.Contains(output, "Test error message") {
		t.Error("Output doesn't contain error message")
	}
}

func TestCompactLogFormatter(t *testing.T) {
	formatter := &CompactLogFormatter{}
	records := []sdklog.Record{
		createTestLogRecord(log.SeverityInfo, "Compact test"),
	}

	output := formatter.Format(records)
	if !strings.Contains(output, "INF") {
		t.Error("Compact format doesn't contain severity")
	}
	if !strings.Contains(output, "Compact test") {
		t.Error("Compact format doesn't contain message")
	}
}

func TestJSONLogFormatter(t *testing.T) {
	formatter := &JSONLogFormatter{}
	records := []sdklog.Record{
		createTestLogRecord(log.SeverityWarn, "JSON test"),
	}

	output := formatter.Format(records)
	if !strings.Contains(output, `"severity"`) {
		t.Error("JSON format doesn't contain severity field")
	}
	if !strings.Contains(output, `"body": "JSON test"`) {
		t.Error("JSON format doesn't contain body field")
	}
}

func TestLogExporter_WithTraceContext(t *testing.T) {
	buf := &bytes.Buffer{}
	exporter := NewLogExporter(WithLogWriter(buf))

	record := createTestLogRecord(log.SeverityInfo, "Traced message")
	// Add trace context
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	record.SetTraceID(traceID)
	record.SetSpanID(spanID)

	err := exporter.Export(context.Background(), []sdklog.Record{record})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Trace ID") {
		t.Error("Output doesn't contain trace ID")
	}
	if !strings.Contains(output, "Span ID") {
		t.Error("Output doesn't contain span ID")
	}
}

func TestLogExporter_EmptyRecords(t *testing.T) {
	buf := &bytes.Buffer{}
	exporter := NewLogExporter(WithLogWriter(buf))

	err := exporter.Export(context.Background(), []sdklog.Record{})
	if err != nil {
		t.Fatalf("Export of empty records failed: %v", err)
	}

	if buf.Len() > 0 {
		t.Error("Expected no output for empty records")
	}
}

func TestDefaultLogFormatter_SeverityLevels(t *testing.T) {
	formatter := &defaultLogFormatter{}

	tests := []struct {
		severity log.Severity
		expected string
	}{
		{log.SeverityFatal, "💀 FATAL"},
		{log.SeverityError, "❌ ERROR"},
		{log.SeverityWarn, "⚠️  WARN"},
		{log.SeverityInfo, "ℹ️  INFO"},
		{log.SeverityDebug, "🐛 DEBUG"},
		{log.SeverityTrace, "📝 TRACE"},
	}

	for _, tt := range tests {
		result := formatter.formatSeverity(tt.severity)
		if !strings.Contains(result, tt.expected) {
			t.Errorf("formatSeverity(%v) = %q, want to contain %q", tt.severity, result, tt.expected)
		}
	}
}

func TestLogExporter_Shutdown(t *testing.T) {
	exporter := NewLogExporter()
	err := exporter.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestLogExporter_ForceFlush(t *testing.T) {
	exporter := NewLogExporter()
	err := exporter.ForceFlush(context.Background())
	if err != nil {
		t.Errorf("ForceFlush failed: %v", err)
	}
}

// Helper function to create test log records
func createTestLogRecord(severity log.Severity, message string) sdklog.Record {
	record := sdklog.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetBody(log.StringValue(message))
	record.AddAttributes(
		log.String("test.key", "test.value"),
	)
	return record
}
