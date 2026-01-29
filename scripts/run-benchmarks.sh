#!/bin/bash
# Run benchmarks and output results in both JSON and standard format for benchstat
# Usage: ./scripts/run-benchmarks.sh [output-dir]

set -e

OUTPUT_DIR="${1:-benchmark-results}"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
JSON_FILE="$OUTPUT_DIR/benchmarks-$TIMESTAMP.json"
BENCHSTAT_FILE="$OUTPUT_DIR/benchmarks-$TIMESTAMP.txt"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Running benchmarks..."
echo "  JSON output: $JSON_FILE"
echo "  Benchstat output: $BENCHSTAT_FILE"
echo ""

# Run benchmarks and capture standard output for benchstat
echo "Running benchmarks (standard format for benchstat)..."
go test -run=^$ -bench=BenchmarkDeserialize -benchmem > "$BENCHSTAT_FILE" 2>&1

# Run again with JSON output for programmatic processing
echo "Running benchmarks (JSON format for analysis)..."
go test -run=^$ -bench=BenchmarkDeserialize -benchmem -json > "$JSON_FILE" 2>&1

echo ""
echo "Benchmarks completed!"
echo ""
echo "To compare results with benchstat:"
echo "  benchstat $BENCHSTAT_FILE <previous-run>.txt"
echo ""
echo "To view JSON results:"
echo "  cat $JSON_FILE | jq . | head -20"
echo ""
echo "Files created:"
echo "  $BENCHSTAT_FILE"
echo "  $JSON_FILE"
