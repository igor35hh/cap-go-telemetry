package console

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SpanExporter implements a console span exporter that mimics the JavaScript version
type SpanExporter struct {
	writer    Writer
	formatter SpanFormatter
}

// Writer interface for output
type Writer interface {
	Write([]byte) (int, error)
}

// SpanFormatter formats spans for console output
type SpanFormatter interface {
	Format(spans []trace.ReadOnlySpan) string
}

// NewSpanExporter creates a new console span exporter
func NewSpanExporter(opts ...SpanExporterOption) *SpanExporter {
	exporter := &SpanExporter{
		writer:    &defaultWriter{},
		formatter: &defaultSpanFormatter{},
	}

	for _, opt := range opts {
		opt(exporter)
	}

	return exporter
}

// SpanExporterOption configures a SpanExporter
type SpanExporterOption func(*SpanExporter)

// WithWriter sets the writer for the exporter
func WithWriter(w Writer) SpanExporterOption {
	return func(e *SpanExporter) {
		e.writer = w
	}
}

// WithSpanFormatter sets the formatter for the exporter
func WithSpanFormatter(f SpanFormatter) SpanExporterOption {
	return func(e *SpanExporter) {
		e.formatter = f
	}
}

// ExportSpans exports spans to the console
func (e *SpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}

	output := e.formatter.Format(spans)
	_, err := e.writer.Write([]byte(output))
	return err
}

// Shutdown shuts down the exporter
func (e *SpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

// defaultSpanFormatter provides the default span formatting
type defaultSpanFormatter struct{}

// Format formats spans in a tree-like structure similar to the JS version
func (f *defaultSpanFormatter) Format(spans []trace.ReadOnlySpan) string {
	if len(spans) == 0 {
		return ""
	}

	var builder strings.Builder

	// Group spans by trace ID and build hierarchy
	traceGroups := make(map[string][]trace.ReadOnlySpan)
	for _, span := range spans {
		traceID := span.SpanContext().TraceID().String()
		traceGroups[traceID] = append(traceGroups[traceID], span)
	}

	// Define colors
	labelColor := color.New(color.FgGreen, color.Bold).SprintFunc()
	traceIDColor := color.New(color.FgMagenta).SprintFunc()

	for traceID, traceSpans := range traceGroups {
		builder.WriteString(fmt.Sprintf("%s - %s (trace: %s):\n",
			labelColor("[telemetry]"),
			color.GreenString("elapsed times"),
			traceIDColor(traceID[:8])))

		// Sort spans by start time
		sortedSpans := sortSpansByStartTime(traceSpans)

		// Find the root span (the one with the earliest start time)
		if len(sortedSpans) > 0 {
			f.formatSpanHierarchy(&builder, sortedSpans, 0)
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// formatSpanHierarchy formats spans in a hierarchical manner
func (f *defaultSpanFormatter) formatSpanHierarchy(builder *strings.Builder, spans []trace.ReadOnlySpan, depth int) {
	// Define colors
	timeColor := color.New(color.FgHiBlack).SprintFunc()
	durationColor := color.New(color.FgYellow, color.Bold).SprintFunc()
	spanNameColor := color.New(color.FgCyan).SprintFunc()
	attributeKeyColor := color.New(color.FgMagenta).SprintFunc()

	for _, span := range spans {
		indent := strings.Repeat("  ", depth)
		duration := span.EndTime().Sub(span.StartTime())

		// Format: start → end = duration ms  operation_name
		startMs := float64(span.StartTime().UnixNano()) / 1e6
		endMs := float64(span.EndTime().UnixNano()) / 1e6
		durationMs := float64(duration.Nanoseconds()) / 1e6

		// Use modulo with int conversion for display
		builder.WriteString(fmt.Sprintf("%s%s → %s = %s  %s\n",
			indent,
			timeColor(fmt.Sprintf("%8.2f", float64(int64(startMs)%10000))),
			timeColor(fmt.Sprintf("%8.2f", float64(int64(endMs)%10000))),
			durationColor(fmt.Sprintf("%8.2f ms", durationMs)),
			spanNameColor(span.Name())))

		// Add attributes if present
		attrs := span.Attributes()
		for _, attr := range attrs {
			if isImportantAttribute(string(attr.Key)) {
				builder.WriteString(fmt.Sprintf("%s    %s: %v\n",
					indent, attributeKeyColor(string(attr.Key)), attr.Value.AsString()))
			}
		}
	}
}

// isImportantAttribute determines if an attribute should be displayed
func isImportantAttribute(key string) bool {
	importantKeys := []string{
		"http.method",
		"http.url",
		"http.status_code",
		"db.statement",
		"db.system",
		"error",
	}

	keyStr := string(key)
	for _, important := range importantKeys {
		if keyStr == important {
			return true
		}
	}
	return false
}

// sortSpansByStartTime sorts spans by their start time
func sortSpansByStartTime(spans []trace.ReadOnlySpan) []trace.ReadOnlySpan {
	sorted := make([]trace.ReadOnlySpan, len(spans))
	copy(sorted, spans)

	// Simple bubble sort - good enough for console output
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].StartTime().After(sorted[j+1].StartTime()) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// defaultWriter writes to stdout
type defaultWriter struct{}

func (w *defaultWriter) Write(p []byte) (int, error) {
	return fmt.Print(string(p))
}
