# Run benchmarks and output results in both JSON and standard format for benchstat
# Usage: .\scripts\run-benchmarks.ps1 [output-dir]

param(
    [string]$OutputDir = "benchmark-results"
)

$ErrorActionPreference = "Stop"

# Create output directory if it doesn't exist
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

$Timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$JsonFile = Join-Path $OutputDir "benchmarks-$Timestamp.json"
$BenchstatFile = Join-Path $OutputDir "benchmarks-$Timestamp.txt"

Write-Host "Running benchmarks..." -ForegroundColor Cyan
Write-Host "  JSON output: $JsonFile" -ForegroundColor Gray
Write-Host "  Benchstat output: $BenchstatFile" -ForegroundColor Gray
Write-Host ""

# Run benchmarks and capture standard output for benchstat
Write-Host "Running benchmarks (standard format for benchstat)..." -ForegroundColor Yellow
$benchmarkOutput = go test -run=^$ -bench=BenchmarkDeserialize -benchmem 2>&1 | Out-String
$benchmarkOutput | Out-File -FilePath $BenchstatFile -Encoding utf8

# Run again with JSON output for programmatic processing
Write-Host "Running benchmarks (JSON format for analysis)..." -ForegroundColor Yellow
go test -run=^$ -bench=BenchmarkDeserialize -benchmem -json | Out-File -FilePath $JsonFile -Encoding utf8

Write-Host ""
Write-Host "Benchmarks completed!" -ForegroundColor Green
Write-Host ""
Write-Host "To compare results with benchstat:" -ForegroundColor Yellow
Write-Host "  benchstat $BenchstatFile <previous-run>.txt" -ForegroundColor White
Write-Host ""
Write-Host "To view JSON results:" -ForegroundColor Yellow
Write-Host "  Get-Content $JsonFile | ConvertFrom-Json | Select-Object -First 10" -ForegroundColor White
Write-Host ""
Write-Host "Files created:" -ForegroundColor Cyan
Write-Host "  $BenchstatFile" -ForegroundColor Gray
Write-Host "  $JsonFile" -ForegroundColor Gray
