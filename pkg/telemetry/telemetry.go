package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iklimetscisco/cap-go-telemetry/pkg/telemetry/config"
	"github.com/iklimetscisco/cap-go-telemetry/pkg/telemetry/exporters/console"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Telemetry represents the main telemetry instance
type Telemetry struct {
	config         *config.Config
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
	resource       *resource.Resource
	logger         *log.Logger
}

// New creates a new telemetry instance
func New(opts ...Option) (*Telemetry, error) {
	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	t := &Telemetry{
		config: cfg,
		logger: log.New(os.Stdout, "[telemetry] ", log.LstdFlags),
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	// Check if telemetry is disabled
	if !cfg.IsEnabled() {
		t.logger.Println("telemetry is disabled")
		return t, nil
	}

	// Initialize resource
	if err := t.initResource(); err != nil {
		return nil, fmt.Errorf("failed to initialize resource: %w", err)
	}

	// Initialize tracing if enabled
	if cfg.IsTracingEnabled() {
		if err := t.initTracing(); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	// Initialize metrics if enabled
	if cfg.IsMetricsEnabled() {
		if err := t.initMetrics(); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	t.logger.Printf("telemetry initialized with kind: %s", cfg.Kind)
	return t, nil
}

// Option configures the telemetry instance
type Option func(*Telemetry)

// WithConfig sets a custom configuration
func WithConfig(cfg *config.Config) Option {
	return func(t *Telemetry) {
		t.config = cfg
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *log.Logger) Option {
	return func(t *Telemetry) {
		t.logger = logger
	}
}

// initResource initializes the OpenTelemetry resource
func (t *Telemetry) initResource() error {
	serviceName := t.config.ServiceName
	if serviceName == "" {
		serviceName = "CAP Application"
	}

	// Try to get service version from environment or default
	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	t.resource = r
	return nil
}

// initTracing initializes the tracing provider
func (t *Telemetry) initTracing() error {
	var exporter trace.SpanExporter

	// Create exporter based on configuration
	exporterConfig := t.config.Tracing.Exporter
	switch exporterConfig.Module {
	case "console":
		exporter = console.NewSpanExporter()
	default:
		return fmt.Errorf("unsupported trace exporter: %s", exporterConfig.Module)
	}

	// Create sampler
	sampler := t.createSampler()

	// Create tracer provider
	opts := []trace.TracerProviderOption{
		trace.WithBatcher(exporter),
		trace.WithResource(t.resource),
		trace.WithSampler(sampler),
	}

	t.tracerProvider = trace.NewTracerProvider(opts...)

	// Set global tracer provider
	otel.SetTracerProvider(t.tracerProvider)

	// Set global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// initMetrics initializes the metrics provider
func (t *Telemetry) initMetrics() error {
	var exporter metric.Exporter

	// Create exporter based on configuration
	exporterConfig := t.config.Metrics.Exporter
	switch exporterConfig.Module {
	case "console":
		exporter = console.NewMetricExporter()
	default:
		return fmt.Errorf("unsupported metric exporter: %s", exporterConfig.Module)
	}

	// Create meter provider
	exportInterval := t.config.Metrics.Config.GetExportInterval()
	opts := []metric.Option{
		metric.WithResource(t.resource),
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(exportInterval))),
	}

	t.meterProvider = metric.NewMeterProvider(opts...)

	// Set global meter provider
	otel.SetMeterProvider(t.meterProvider)

	return nil
}

// createSampler creates a sampler based on configuration
func (t *Telemetry) createSampler() trace.Sampler {
	samplerConfig := t.config.Tracing.Sampler
	if samplerConfig == nil {
		return trace.AlwaysSample()
	}

	switch samplerConfig.Kind {
	case "AlwaysOnSampler":
		return trace.AlwaysSample()
	case "AlwaysOffSampler":
		return trace.NeverSample()
	case "TraceIdRatioBasedSampler":
		ratio := samplerConfig.Ratio
		if ratio <= 0 {
			ratio = 1.0
		}
		return trace.TraceIDRatioBased(ratio)
	case "ParentBasedSampler":
		var root trace.Sampler
		switch samplerConfig.Root {
		case "AlwaysOnSampler":
			root = trace.AlwaysSample()
		case "AlwaysOffSampler":
			root = trace.NeverSample()
		default:
			root = trace.AlwaysSample()
		}
		return trace.ParentBased(root)
	default:
		return trace.AlwaysSample()
	}
}

// Shutdown gracefully shuts down the telemetry providers
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errors []error

	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
	}

	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown meter provider: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	t.logger.Println("telemetry shutdown completed")
	return nil
}

// TracerProvider returns the tracer provider
func (t *Telemetry) TracerProvider() *trace.TracerProvider {
	return t.tracerProvider
}

// MeterProvider returns the meter provider
func (t *Telemetry) MeterProvider() *metric.MeterProvider {
	return t.meterProvider
}

// Config returns the configuration
func (t *Telemetry) Config() *config.Config {
	return t.config
}
