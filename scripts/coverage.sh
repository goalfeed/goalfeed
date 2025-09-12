#!/bin/bash
# Script to generate coverage report locally

echo "Running tests and generating coverage report..."

# Run tests that work without network access and generate coverage
echo "Running package tests..."
go test -coverprofile=coverage.out ./models ./targets/homeassistant ./targets/memoryStore ./services/leagues/nhl ./services/leagues/mlb ./services/leagues/iihf ./utils ./config ./clients/...

echo "Running main package tests..."
# Run all main package tests (tests avoid network by design)
go test -coverprofile=main_coverage.out -v .

echo "Combining coverage files..."
# Combine coverage files
echo "mode: set" > combined_coverage.out
grep -h -v "^mode:" coverage.out main_coverage.out >> combined_coverage.out 2>/dev/null || true

echo "Coverage report:"
go tool cover -func=combined_coverage.out | tail -1

echo ""
echo "To view detailed coverage in browser, run:"
echo "go tool cover -html=combined_coverage.out"

# Clean up temporary files
rm -f coverage.out main_coverage.out 2>/dev/null || true

echo ""
echo "Combined coverage report saved as: combined_coverage.out"