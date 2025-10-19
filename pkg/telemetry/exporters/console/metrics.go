package console

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// MetricExporter implements a console metric exporter
type MetricExporter struct {
	writer    Writer
	formatter MetricFormatter
}

// MetricFormatter formats metrics for console output
type MetricFormatter interface {
	Format(metrics *metricdata.ResourceMetrics) string
}

// NewMetricExporter creates a new console metric exporter
func NewMetricExporter(opts ...MetricExporterOption) *MetricExporter {
	exporter := &MetricExporter{
		writer:    &defaultWriter{},
		formatter: &defaultMetricFormatter{},
	}

	for _, opt := range opts {
		opt(exporter)
	}

	return exporter
}

// MetricExporterOption configures a MetricExporter
type MetricExporterOption func(*MetricExporter)

// WithMetricWriter sets the writer for the exporter
func WithMetricWriter(w Writer) MetricExporterOption {
	return func(e *MetricExporter) {
		e.writer = w
	}
}

// WithMetricFormatter sets the formatter for the exporter
func WithMetricFormatter(f MetricFormatter) MetricExporterOption {
	return func(e *MetricExporter) {
		e.formatter = f
	}
}

// Export exports metrics to the console
func (e *MetricExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	output := e.formatter.Format(metrics)
	if output != "" {
		_, err := e.writer.Write([]byte(output))
		return err
	}
	return nil
}

// ForceFlush forces a flush of the exporter
func (e *MetricExporter) ForceFlush(ctx context.Context) error {
	return nil
}

// Shutdown shuts down the exporter
func (e *MetricExporter) Shutdown(ctx context.Context) error {
	return nil
}

// Temporality returns the temporality preference for the exporter
func (e *MetricExporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

// Aggregation returns the aggregation preference for the exporter
func (e *MetricExporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return metric.DefaultAggregationSelector(kind)
}

// defaultMetricFormatter provides the default metric formatting
type defaultMetricFormatter struct{}

// Format formats metrics in a human-readable format similar to the JS version
func (f *defaultMetricFormatter) Format(rm *metricdata.ResourceMetrics) string {
	if rm == nil || len(rm.ScopeMetrics) == 0 {
		return ""
	}

	var builder strings.Builder

	// Group metrics by type for better presentation
	hostMetrics := make([]metricdata.Metrics, 0)
	dbPoolMetrics := make([]metricdata.Metrics, 0)
	queueMetrics := make([]metricdata.Metrics, 0)
	customMetrics := make([]metricdata.Metrics, 0)

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch {
			case strings.HasPrefix(m.Name, "process.") || strings.HasPrefix(m.Name, "runtime."):
				hostMetrics = append(hostMetrics, m)
			case strings.HasPrefix(m.Name, "db.pool"):
				dbPoolMetrics = append(dbPoolMetrics, m)
			case strings.HasPrefix(m.Name, "queue"):
				queueMetrics = append(queueMetrics, m)
			default:
				customMetrics = append(customMetrics, m)
			}
		}
	}

	// Define colors
	labelColor := color.New(color.FgGreen, color.Bold).SprintFunc()
	sectionColor := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Format host metrics
	if len(hostMetrics) > 0 {
		builder.WriteString(fmt.Sprintf("%s - %s:\n", labelColor("[telemetry]"), sectionColor("host metrics")))
		f.formatHostMetrics(&builder, hostMetrics)
		builder.WriteString("\n")
	}

	// Format DB pool metrics
	if len(dbPoolMetrics) > 0 {
		builder.WriteString(fmt.Sprintf("%s - %s:\n", labelColor("[telemetry]"), sectionColor("db.pool")))
		f.formatDBPoolMetrics(&builder, dbPoolMetrics)
		builder.WriteString("\n")
	}

	// Format queue metrics
	if len(queueMetrics) > 0 {
		builder.WriteString(fmt.Sprintf("%s - %s:\n", labelColor("[telemetry]"), sectionColor("queue")))
		f.formatQueueMetrics(&builder, queueMetrics)
		builder.WriteString("\n")
	}

	// Format custom metrics
	if len(customMetrics) > 0 {
		builder.WriteString(fmt.Sprintf("%s - %s:\n", labelColor("[telemetry]"), sectionColor("custom metrics")))
		f.formatCustomMetrics(&builder, customMetrics)
		builder.WriteString("\n")
	}

	return builder.String()
}

// formatHostMetrics formats host-related metrics
func (f *defaultMetricFormatter) formatHostMetrics(builder *strings.Builder, metrics []metricdata.Metrics) {
	for _, m := range metrics {
		switch m.Name {
		case "process.cpu.time":
			f.formatCPUTime(builder, m)
		case "process.memory.usage":
			f.formatMemoryUsage(builder, m)
		case "runtime.go.gc.count":
			f.formatGCCount(builder, m)
		default:
			f.formatGenericMetric(builder, m)
		}
	}
}

// formatCPUTime formats CPU time metrics
func (f *defaultMetricFormatter) formatCPUTime(builder *strings.Builder, m metricdata.Metrics) {
	if sum, ok := m.Data.(metricdata.Sum[float64]); ok {
		userTime, systemTime := 0.0, 0.0
		for _, dp := range sum.DataPoints {
			for _, attr := range dp.Attributes.ToSlice() {
				if string(attr.Key) == "state" {
					if attr.Value.AsString() == "user" {
						userTime = dp.Value
					} else if attr.Value.AsString() == "system" {
						systemTime = dp.Value
					}
				}
			}
		}
		builder.WriteString(fmt.Sprintf("  Process Cpu time in seconds: { user: %.3f, system: %.3f }\n",
			userTime, systemTime))
	}
}

// formatMemoryUsage formats memory usage metrics
func (f *defaultMetricFormatter) formatMemoryUsage(builder *strings.Builder, m metricdata.Metrics) {
	if gauge, ok := m.Data.(metricdata.Gauge[int64]); ok {
		for _, dp := range gauge.DataPoints {
			builder.WriteString(fmt.Sprintf("  Process Memory usage in bytes: %d\n", dp.Value))
		}
	}
}

// formatGCCount formats garbage collection count
func (f *defaultMetricFormatter) formatGCCount(builder *strings.Builder, m metricdata.Metrics) {
	if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
		for _, dp := range sum.DataPoints {
			builder.WriteString(fmt.Sprintf("  Runtime GC count: %d\n", dp.Value))
		}
	}
}

// formatDBPoolMetrics formats database pool metrics
func (f *defaultMetricFormatter) formatDBPoolMetrics(builder *strings.Builder, metrics []metricdata.Metrics) {
	// Define colors
	headerColor := color.New(color.FgYellow, color.Bold).SprintFunc()
	valueColor := color.New(color.FgCyan).SprintFunc()

	// Example format:     size | available | pending
	//                      1/1 |       1/1 |       0
	builder.WriteString(fmt.Sprintf("     %s | %s | %s\n",
		headerColor("size"), headerColor("available"), headerColor("pending")))

	size, available, pending := "0/0", "0/0", "0"

	for _, m := range metrics {
		switch m.Name {
		case "db.pool.size":
			if gauge, ok := m.Data.(metricdata.Gauge[int64]); ok {
				for _, dp := range gauge.DataPoints {
					size = fmt.Sprintf("%d/%d", dp.Value, dp.Value) // Current/Max
				}
			}
		case "db.pool.available":
			if gauge, ok := m.Data.(metricdata.Gauge[int64]); ok {
				for _, dp := range gauge.DataPoints {
					available = fmt.Sprintf("%d/%d", dp.Value, dp.Value)
				}
			}
		case "db.pool.pending":
			if gauge, ok := m.Data.(metricdata.Gauge[int64]); ok {
				for _, dp := range gauge.DataPoints {
					pending = fmt.Sprintf("%d", dp.Value)
				}
			}
		}
	}

	builder.WriteString(fmt.Sprintf("     %s |      %s |      %s\n",
		valueColor(size), valueColor(available), valueColor(pending)))
}

// formatQueueMetrics formats queue metrics
func (f *defaultMetricFormatter) formatQueueMetrics(builder *strings.Builder, metrics []metricdata.Metrics) {
	// Example format: cold | remaining | min storage time | med storage time | max storage time | incoming | outgoing
	//                   2  |       32  |                2 |               16 |              128 |      256 |      512
	builder.WriteString("     cold | remaining | min storage time | med storage time | max storage time | incoming | outgoing\n")

	values := map[string]string{
		"cold": "0", "remaining": "0", "min": "0", "med": "0", "max": "0", "incoming": "0", "outgoing": "0",
	}

	for _, m := range metrics {
		if gauge, ok := m.Data.(metricdata.Gauge[int64]); ok {
			for _, dp := range gauge.DataPoints {
				switch m.Name {
				case "queue.cold":
					values["cold"] = fmt.Sprintf("%d", dp.Value)
				case "queue.remaining":
					values["remaining"] = fmt.Sprintf("%d", dp.Value)
				case "queue.incoming":
					values["incoming"] = fmt.Sprintf("%d", dp.Value)
				case "queue.outgoing":
					values["outgoing"] = fmt.Sprintf("%d", dp.Value)
				}
			}
		}
	}

	builder.WriteString(fmt.Sprintf("     %4s |      %4s |             %4s |             %4s |             %4s |     %4s |     %4s\n",
		values["cold"], values["remaining"], values["min"], values["med"], values["max"], values["incoming"], values["outgoing"]))
}

// formatCustomMetrics formats custom application metrics
func (f *defaultMetricFormatter) formatCustomMetrics(builder *strings.Builder, metrics []metricdata.Metrics) {
	for _, m := range metrics {
		f.formatGenericMetric(builder, m)
	}
}

// formatGenericMetric formats any metric in a generic way
func (f *defaultMetricFormatter) formatGenericMetric(builder *strings.Builder, m metricdata.Metrics) {
	builder.WriteString(fmt.Sprintf("  %s: ", m.Name))

	switch data := m.Data.(type) {
	case metricdata.Gauge[int64]:
		for _, dp := range data.DataPoints {
			builder.WriteString(fmt.Sprintf("%d ", dp.Value))
		}
	case metricdata.Gauge[float64]:
		for _, dp := range data.DataPoints {
			builder.WriteString(fmt.Sprintf("%.3f ", dp.Value))
		}
	case metricdata.Sum[int64]:
		for _, dp := range data.DataPoints {
			builder.WriteString(fmt.Sprintf("%d ", dp.Value))
		}
	case metricdata.Sum[float64]:
		for _, dp := range data.DataPoints {
			builder.WriteString(fmt.Sprintf("%.3f ", dp.Value))
		}
	case metricdata.Histogram[int64]:
		builder.WriteString(fmt.Sprintf("count: %d ", data.DataPoints[0].Count))
	case metricdata.Histogram[float64]:
		builder.WriteString(fmt.Sprintf("count: %d ", data.DataPoints[0].Count))
	}

	builder.WriteString("\n")
}
