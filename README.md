# Cap-go-telemetry

A Go implementation of the CAP.js telemetry plugin, providing observability features with automatic OpenTelemetry instrumentation.

## Overview

Cap-go-telemetry is a comprehensive telemetry solution for Go applications that provides:

- **Distributed Tracing**: Track requests across microservices
- **Metrics Collection**: Host metrics, custom metrics, and application performance metrics
- **Logging Integration**: Structured logging with trace correlation
- **Multiple Exporters**: Console, OTLP, Dynatrace, SAP Cloud Logging, Jaeger
- **Auto-instrumentation**: Automatic instrumentation for popular Go frameworks

## Features

### Telemetry Signals
- **Traces**: Distributed tracing with automatic span creation
- **Metrics**: Host metrics, runtime metrics, custom metrics
- **Logs**: Structured logging with OpenTelemetry integration

### Supported Exporters
- **Console**: Pretty-printed output for development
- **OTLP**: gRPC, HTTP, and Proto formats
- **Dynatrace**: Direct integration with Dynatrace
- **SAP Cloud Logging**: Integration with SAP BTP Cloud Logging
- **Jaeger**: Direct export to Jaeger

### Auto-instrumentation (Planned)
- HTTP frameworks (net/http, Gin, Echo, Chi)
- Database drivers (database/sql, GORM, Redis, MongoDB)
- gRPC (server and client)
- Message queues (Kafka, RabbitMQ)

## Quick Start

### Installation

```bash
go get github.com/iklimetscisco/cap-go-telemetry
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/iklimetscisco/cap-go-telemetry/pkg/telemetry"
    "go.opentelemetry.io/otel"
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
        tel.Shutdown(ctx)
    }()

    // Create a tracer
    tracer := otel.Tracer("my-service")

    // Start a span
    ctx, span := tracer.Start(context.Background(), "my-operation")
    defer span.End()

    // Your application code here
    doSomeWork(ctx)
}
```

### Running the Example

```bash
# Run the basic example
cd examples/basic
go run main.go

# Visit http://localhost:8080/ to see telemetry in action
# Check the console for trace and metric output
```

## Configuration

Cap-go-telemetry supports configuration through:

1. Environment variables
2. Configuration files (YAML, JSON)
3. Programmatic configuration

### Environment Variables

```bash
# Disable telemetry
export NO_TELEMETRY=true

# Set service name
export OTEL_SERVICE_NAME="my-application"

# Set telemetry kind
export TELEMETRY_KIND="telemetry-to-console"
```

### Configuration File

Create a `telemetry.yaml` file:

```yaml
service_name: "my-application"
kind: "telemetry-to-console"

tracing:
  enabled: true
  hrtime: true
  sampler:
    kind: "ParentBasedSampler"
    root: "AlwaysOnSampler"
    ignore_incoming_paths:
      - "/health"
      - "/metrics"

metrics:
  enabled: true
  host_metrics: true
  runtime_metrics: true
  config:
    export_interval_millis: 60000

logging:
  enabled: false
```

### Predefined Kinds

Cap-go-telemetry includes several predefined configurations:

- `telemetry-to-console`: Development-friendly console output
- `telemetry-to-dynatrace`: Dynatrace integration
- `telemetry-to-cloud-logging`: SAP Cloud Logging integration
- `telemetry-to-jaeger`: Jaeger integration
- `telemetry-to-otlp`: Generic OTLP endpoint

## Development Status

This is the initial implementation of cap-go-telemetry. Current status:

### ✅ Completed
- Core configuration system
- Console exporters for traces and metrics
- Basic OpenTelemetry integration
- Project structure and foundation
- Basic example application

### 🚧 In Progress
- OTLP exporters (gRPC, HTTP, Proto)
- Cloud-specific exporters (Dynatrace, SAP Cloud Logging, Jaeger)
- Host and runtime metrics collection
- Logging integration

### 📋 Planned
- Auto-instrumentation for HTTP frameworks
- Database instrumentation
- gRPC instrumentation
- Message queue instrumentation
- CAP framework integration
- Performance optimizations
- Comprehensive testing

## Project Structure

```
cap-go-telemetry/
├── pkg/telemetry/           # Public API
│   ├── config/             # Configuration management
│   ├── exporters/          # Telemetry exporters
│   │   └── console/        # Console exporters
│   └── telemetry.go        # Main telemetry API
├── internal/               # Internal packages
│   └── version/            # Version information
├── examples/               # Example applications
│   └── basic/              # Basic usage example
├── docs/                   # Documentation
└── test/                   # Tests
```

## Comparison with JavaScript Version

| Feature | JavaScript (@cap-js/telemetry) | Go (cap-go-telemetry) | Status |
|---------|--------------------------------|----------------------|---------|
| Console Exporters | ✅ | ✅ | Complete |
| OTLP Exporters | ✅ | 🚧 | In Progress |
| Dynatrace Export | ✅ | 📋 | Planned |
| Cloud Logging Export | ✅ | 📋 | Planned |
| Jaeger Export | ✅ | 📋 | Planned |
| Host Metrics | ✅ | 🚧 | In Progress |
| Database Pool Metrics | ✅ | 📋 | Planned |
| Queue Metrics | ✅ | 📋 | Planned |
| HTTP Instrumentation | ✅ | 📋 | Planned |
| Database Instrumentation | ✅ | 📋 | Planned |
| Custom Spans | ✅ | ✅ | Complete |
| Configuration System | ✅ | ✅ | Complete |

## Contributing

This project is under active development. Contributions are welcome!

### Development Setup

```bash
# Clone the repository
git clone https://github.com/iklimetscisco/cap-go-telemetry.git
cd cap-go-telemetry

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run the example
cd examples/basic
go run main.go
```

### Building

```bash
# Build all examples
make build

# Run tests
make test

# Run linting
make lint
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original [CAP.js telemetry plugin](https://github.com/cap-js/telemetry) for inspiration and feature reference
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go) for the underlying instrumentation
- SAP CAP framework for the plugin architecture concepts

## Next Steps (Following the Roadmap):
OTLP Exporters (4-6 weeks) - gRPC, HTTP, Proto formats
Cloud Exporters (3-4 weeks) - Dynatrace, SAP Cloud Logging, Jaeger
Auto-Instrumentation (6-8 weeks) - HTTP frameworks, databases, gRPC
Advanced Metrics (4-6 weeks) - Host metrics, DB pool, queue metrics