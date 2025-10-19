package console

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// LogExporter implements a console log exporter
type LogExporter struct {
	writer    io.Writer
	formatter LogFormatter
}

// LogFormatter formats log records for console output
type LogFormatter interface {
	Format(records []sdklog.Record) string
}

// NewLogExporter creates a new console log exporter
func NewLogExporter(opts ...LogExporterOption) *LogExporter {
	exporter := &LogExporter{
		writer:    os.Stdout,
		formatter: &defaultLogFormatter{},
	}

	for _, opt := range opts {
		opt(exporter)
	}

	return exporter
}

// LogExporterOption configures a LogExporter
type LogExporterOption func(*LogExporter)

// WithLogWriter sets the writer for the exporter
func WithLogWriter(w io.Writer) LogExporterOption {
	return func(e *LogExporter) {
		e.writer = w
	}
}

// WithLogFormatter sets the formatter for the exporter
func WithLogFormatter(f LogFormatter) LogExporterOption {
	return func(e *LogExporter) {
		e.formatter = f
	}
}

// Export exports log records to the console
func (e *LogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if len(records) == 0 {
		return nil
	}

	output := e.formatter.Format(records)
	_, err := fmt.Fprint(e.writer, output)
	return err
}

// Shutdown shuts down the exporter
func (e *LogExporter) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush flushes any buffered log records
func (e *LogExporter) ForceFlush(ctx context.Context) error {
	return nil
}

// defaultLogFormatter provides the default log formatting
type defaultLogFormatter struct{}

// Format formats log records in a structured, readable format
func (f *defaultLogFormatter) Format(records []sdklog.Record) string {
	var builder strings.Builder

	// Color for header
	headerColor := color.New(color.FgCyan, color.Bold).SprintFunc()

	builder.WriteString("\n")
	builder.WriteString(headerColor("╔══════════════════════════════════════════════════════════════════════════════╗\n"))
	builder.WriteString(headerColor("║                              📋 LOG RECORDS                                  ║\n"))
	builder.WriteString(headerColor("╚══════════════════════════════════════════════════════════════════════════════╝\n\n"))

	for i, record := range records {
		if i > 0 {
			builder.WriteString("\n")
		}
		f.formatLogRecord(&builder, record)
	}

	builder.WriteString("\n")
	return builder.String()
}

// formatLogRecord formats a single log record
func (f *defaultLogFormatter) formatLogRecord(builder *strings.Builder, record sdklog.Record) {
	// Define colors
	timestampColor := color.New(color.FgHiBlack).SprintFunc()
	attributeKeyColor := color.New(color.FgCyan).SprintFunc()
	traceColor := color.New(color.FgMagenta).SprintFunc()
	treeColor := color.New(color.FgHiBlack).SprintFunc()

	// Format timestamp
	timestamp := record.Timestamp()
	timeStr := timestamp.Format("2006-01-02 15:04:05.000")

	// Get severity level
	severity := record.Severity()
	severityStr := f.formatSeverity(severity)

	// Get log body
	body := record.Body().AsString()

	// Format: [timestamp] LEVEL: message
	builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", timestampColor(timeStr), severityStr, body))

	// Add trace context if present
	if record.TraceID().IsValid() {
		builder.WriteString(fmt.Sprintf("%s Trace ID: %s\n", treeColor("  ├─"), traceColor(record.TraceID().String())))
	}
	if record.SpanID().IsValid() {
		builder.WriteString(fmt.Sprintf("%s Span ID:  %s\n", treeColor("  ├─"), traceColor(record.SpanID().String())))
	}

	// Add attributes
	hasAttributes := false
	record.WalkAttributes(func(kv log.KeyValue) bool {
		if !hasAttributes {
			builder.WriteString(fmt.Sprintf("%s Attributes:\n", treeColor("  ├─")))
			hasAttributes = true
		}
		// Use String() method which handles all types
		builder.WriteString(fmt.Sprintf("%s %s: %v\n", treeColor("  │  •"), attributeKeyColor(kv.Key), kv.Value.String()))
		return true
	})
}

// formatSeverity formats severity level with emoji indicators and colors
func (f *defaultLogFormatter) formatSeverity(severity log.Severity) string {
	// Define colors
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	switch {
	case severity >= log.SeverityFatal:
		return red("💀 FATAL  ")
	case severity >= log.SeverityError:
		return red("❌ ERROR  ")
	case severity >= log.SeverityWarn:
		return yellow("⚠️  WARN   ")
	case severity >= log.SeverityInfo:
		return cyan("ℹ️  INFO   ")
	case severity >= log.SeverityDebug:
		return gray("🐛 DEBUG  ")
	default:
		return magenta("📝 TRACE  ")
	}
}

// CompactLogFormatter provides a compact, single-line format
type CompactLogFormatter struct{}

// Format formats log records in a compact format
func (f *CompactLogFormatter) Format(records []sdklog.Record) string {
	var builder strings.Builder

	for _, record := range records {
		timestamp := record.Timestamp().Format("15:04:05.000")
		severity := f.formatSeverity(record.Severity())
		body := record.Body().AsString()

		builder.WriteString(fmt.Sprintf("%s %s %s", timestamp, severity, body))

		// Add trace context inline if present
		if record.TraceID().IsValid() {
			builder.WriteString(fmt.Sprintf(" [trace=%s]", record.TraceID().String()[:8]))
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

func (f *CompactLogFormatter) formatSeverity(severity log.Severity) string {
	switch {
	case severity >= log.SeverityFatal:
		return "FTL"
	case severity >= log.SeverityError:
		return "ERR"
	case severity >= log.SeverityWarn:
		return "WRN"
	case severity >= log.SeverityInfo:
		return "INF"
	case severity >= log.SeverityDebug:
		return "DBG"
	default:
		return "TRC"
	}
}

// JSONLogFormatter provides JSON-formatted output
type JSONLogFormatter struct{}

// Format formats log records as JSON
func (f *JSONLogFormatter) Format(records []sdklog.Record) string {
	var builder strings.Builder

	builder.WriteString("[\n")
	for i, record := range records {
		if i > 0 {
			builder.WriteString(",\n")
		}
		builder.WriteString("  {\n")
		builder.WriteString(fmt.Sprintf("    \"timestamp\": \"%s\",\n", record.Timestamp().Format(time.RFC3339Nano)))
		builder.WriteString(fmt.Sprintf("    \"severity\": \"%s\",\n", record.Severity().String()))
		builder.WriteString(fmt.Sprintf("    \"body\": %q", record.Body().AsString()))

		if record.TraceID().IsValid() {
			builder.WriteString(",\n")
			builder.WriteString(fmt.Sprintf("    \"traceId\": \"%s\"", record.TraceID().String()))
		}
		if record.SpanID().IsValid() {
			builder.WriteString(",\n")
			builder.WriteString(fmt.Sprintf("    \"spanId\": \"%s\"", record.SpanID().String()))
		}

		// Add attributes as JSON object
		hasAttributes := false
		record.WalkAttributes(func(kv log.KeyValue) bool {
			if !hasAttributes {
				builder.WriteString(",\n    \"attributes\": {\n")
				hasAttributes = true
			} else {
				builder.WriteString(",\n")
			}
			builder.WriteString(fmt.Sprintf("      %q: %q", kv.Key, kv.Value.String()))
			return true
		})
		if hasAttributes {
			builder.WriteString("\n    }")
		}

		builder.WriteString("\n  }")
	}
	builder.WriteString("\n]\n")

	return builder.String()
}
