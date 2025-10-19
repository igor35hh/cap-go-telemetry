package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iklimetscisco/cap-go-telemetry/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	// Initialize telemetry
	tel, err := telemetry.New()
	if err != nil {
		log.Fatalf("failed to initialize telemetry: %v", err)
	}

	// Shutdown telemetry when the application exits
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tel.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown telemetry: %v", err)
		}
	}()

	// Create a tracer
	tracer := otel.Tracer("example-service")

	// Create a meter for custom metrics
	meter := otel.Meter("example-service")
	requestCounter, err := meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		log.Fatalf("failed to create counter: %v", err)
	}

	// HTTP handler with tracing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handle_request")
		defer span.End()

		// Add some attributes to the span
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("user_agent", r.UserAgent()),
		)

		// Increment counter
		requestCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", r.Method),
			attribute.String("path", r.URL.Path),
		))

		// Simulate some work
		time.Sleep(100 * time.Millisecond)

		// Simulate a database call
		_, dbSpan := tracer.Start(ctx, "database_query")
		dbSpan.SetAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", "SELECT * FROM users WHERE id = $1"),
		)
		time.Sleep(50 * time.Millisecond)
		dbSpan.End()

		span.SetAttributes(attribute.Int("http.status_code", 200))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello, World! Request processed with telemetry.\n")
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
	})

	// Start background work to generate metrics
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Create some background activity
			_, span := tracer.Start(context.Background(), "background_task")
			span.SetAttributes(attribute.String("task.type", "cleanup"))

			// Simulate work
			time.Sleep(200 * time.Millisecond)

			span.End()
		}
	}()

	fmt.Println("Server starting on :8080")
	fmt.Println("Visit http://localhost:8080/ to see telemetry in action")
	fmt.Println("Metrics will be exported every 60 seconds to the console")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
