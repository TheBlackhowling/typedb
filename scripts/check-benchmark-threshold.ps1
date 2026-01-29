# Check benchmark results against threshold
# Usage: .\scripts\check-benchmark-threshold.ps1 [current-file] [baseline-file] [threshold-percent]

param(
    [string]$CurrentFile = "benchmark-results/current.txt",
    [string]$BaselineFile = "benchmark-results/baseline.txt",
    [int]$Threshold = 10  # Default 10% regression threshold
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $BaselineFile)) {
    Write-Host "No baseline file found at $BaselineFile" -ForegroundColor Yellow
    Write-Host "Creating baseline from current run..." -ForegroundColor Yellow
    Copy-Item $CurrentFile $BaselineFile
    exit 0
}

if (-not (Test-Path $CurrentFile)) {
    Write-Host "Error: Current benchmark file not found: $CurrentFile" -ForegroundColor Red
    exit 1
}

# Check if benchstat is available
$benchstatPath = Get-Command benchstat -ErrorAction SilentlyContinue
if (-not $benchstatPath) {
    Write-Host "Installing benchstat..." -ForegroundColor Yellow
    go install golang.org/x/perf/cmd/benchstat@latest
}

# Run comparison
Write-Host "Comparing benchmarks..." -ForegroundColor Cyan
$comparisonOutput = & benchstat $BaselineFile $CurrentFile 2>&1 | Out-String

# Parse for regressions
$regressions = @()
$lines = $comparisonOutput -split "`n"
foreach ($line in $lines) {
    # Look for lines with negative percentage changes (regressions)
    if ($line -match "-\s+.*?(\d+\.\d+)%") {
        $percent = [double]$matches[1]
        if ($percent -gt $Threshold) {
            $regressions += $line
            Write-Host "❌ Regression detected: $line" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "=== Benchmark Comparison ===" -ForegroundColor Cyan
Write-Host $comparisonOutput

if ($regressions.Count -gt 0) {
    Write-Host ""
    Write-Host "❌ Performance regression detected: $($regressions.Count) benchmark(s) degraded by more than ${Threshold}%" -ForegroundColor Red
    exit 1
} else {
    Write-Host ""
    Write-Host "✅ No significant regressions detected (threshold: ${Threshold}%)" -ForegroundColor Green
    exit 0
}
