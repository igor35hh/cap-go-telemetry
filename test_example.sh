#!/bin/bash

# Test script for cap-go-telemetry

echo "=== Cap-go-telemetry Test Script ==="
echo

# Build the project
echo "1. Building the project..."
make build
echo

# Run a quick test
echo "2. Running basic functionality test..."
cd examples/basic

# Start the server in background
echo "Starting telemetry example server..."
go run main.go &
SERVER_PID=$!

# Wait a moment for the server to start
sleep 2

# Make a test request
echo "Making test HTTP request..."
curl -s http://localhost:8080/ > /dev/null 2>&1

# Wait to see some telemetry output
sleep 3

# Stop the server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null

echo
echo "=== Test completed ==="
echo "Check the output above for telemetry traces and metrics"
