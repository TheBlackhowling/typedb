#!/bin/bash
# Check benchmark results against threshold
# Usage: ./scripts/check-benchmark-threshold.sh [current-file] [baseline-file] [threshold-percent]

set -e

CURRENT_FILE="${1:-benchmark-results/current.txt}"
BASELINE_FILE="${2:-benchmark-results/baseline.txt}"
THRESHOLD="${3:-10}"  # Default 10% regression threshold

if [ ! -f "$BASELINE_FILE" ]; then
    echo "No baseline file found at $BASELINE_FILE"
    echo "Creating baseline from current run..."
    cp "$CURRENT_FILE" "$BASELINE_FILE"
    exit 0
fi

if [ ! -f "$CURRENT_FILE" ]; then
    echo "Error: Current benchmark file not found: $CURRENT_FILE"
    exit 1
fi

# Install benchstat if not available
if ! command -v benchstat &> /dev/null; then
    echo "Installing benchstat..."
    go install golang.org/x/perf/cmd/benchstat@latest
fi

# Run comparison
echo "Comparing benchmarks..."
benchstat "$BASELINE_FILE" "$CURRENT_FILE" > /tmp/benchstat_output.txt 2>&1 || true

# Check for regressions
REGRESSIONS=0
while IFS= read -r line; do
    # Look for lines with negative percentage changes (regressions)
    if echo "$line" | grep -qE "^-.*[0-9]+\.[0-9]+%"; then
        PERCENT=$(echo "$line" | grep -oE "[0-9]+\.[0-9]+%" | sed 's/%//')
        if (( $(echo "$PERCENT > $THRESHOLD" | bc -l) )); then
            echo "❌ Regression detected: $line"
            REGRESSIONS=$((REGRESSIONS + 1))
        fi
    fi
done < /tmp/benchstat_output.txt

echo ""
echo "=== Benchmark Comparison ==="
cat /tmp/benchstat_output.txt

if [ $REGRESSIONS -gt 0 ]; then
    echo ""
    echo "❌ Performance regression detected: $REGRESSIONS benchmark(s) degraded by more than ${THRESHOLD}%"
    exit 1
else
    echo ""
    echo "✅ No significant regressions detected (threshold: ${THRESHOLD}%)"
    exit 0
fi
