package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iklimetscisco/cap-go-telemetry/pkg/telemetry/exporters/console"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func main() {
	fmt.Println("=== Console Log Exporter Demo ===")
	fmt.Println()

	// Create a simple logger provider with console exporter
	exporter := console.NewLogExporter()

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("logging-demo"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		fmt.Printf("failed to create resource: %v\n", err)
		return
	}

	// Create logger provider
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	// Shutdown when done
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := loggerProvider.Shutdown(ctx); err != nil {
			fmt.Printf("failed to shutdown logger provider: %v\n", err)
		}
	}()

	fmt.Println("Emitting log records with different severity levels...")
	fmt.Println()

	// Create some test log records manually for the exporter
	records := createTestLogRecords()

	// Export them directly to see the formatting
	if err := exporter.Export(context.Background(), records); err != nil {
		fmt.Printf("failed to export logs: %v\n", err)
	}

	fmt.Println("\nDemo completed!")
}

// createTestLogRecords creates sample log records for demonstration
func createTestLogRecords() []sdklog.Record {
	records := make([]sdklog.Record, 0)

	// DEBUG log
	debugRec := sdklog.Record{}
	debugRec.SetTimestamp(time.Now())
	debugRec.SetSeverity(log.SeverityDebug)
	debugRec.SetBody(log.StringValue("This is a debug message"))
	debugRec.AddAttributes(
		log.String("component", "main"),
		log.Int64("attempt", 1),
	)
	records = append(records, debugRec)

	time.Sleep(100 * time.Millisecond)

	// INFO log
	infoRec := sdklog.Record{}
	infoRec.SetTimestamp(time.Now())
	infoRec.SetSeverity(log.SeverityInfo)
	infoRec.SetBody(log.StringValue("Application started successfully"))
	infoRec.AddAttributes(
		log.String("version", "1.0.0"),
		log.String("environment", "development"),
	)
	records = append(records, infoRec)

	time.Sleep(100 * time.Millisecond)

	// WARN log
	warnRec := sdklog.Record{}
	warnRec.SetTimestamp(time.Now())
	warnRec.SetSeverity(log.SeverityWarn)
	warnRec.SetBody(log.StringValue("Configuration value missing, using default"))
	warnRec.AddAttributes(
		log.String("config_key", "timeout"),
		log.Int64("default_value", 30),
	)
	records = append(records, warnRec)

	time.Sleep(100 * time.Millisecond)

	// ERROR log
	errorRec := sdklog.Record{}
	errorRec.SetTimestamp(time.Now())
	errorRec.SetSeverity(log.SeverityError)
	errorRec.SetBody(log.StringValue("Failed to connect to database"))
	errorRec.AddAttributes(
		log.String("database", "postgres"),
		log.String("host", "localhost:5432"),
		log.String("error", "connection refused"),
	)
	records = append(records, errorRec)

	return records
}
