# Static analysis script for typedb (PowerShell version)
# Run all static analysis tools before 1.0.0 release

$ErrorActionPreference = "Stop"
$errors = 0

Write-Host "Running static analysis for typedb..." -ForegroundColor Cyan
Write-Host ""

function Test-Tool {
    param([string]$ToolName, [string]$InstallCommand)
    
    $exists = Get-Command $ToolName -ErrorAction SilentlyContinue
    if (-not $exists) {
        Write-Host "WARNING: $ToolName not found. Install with: $InstallCommand" -ForegroundColor Yellow
        return $false
    }
    return $true
}

# 1. go vet (built-in)
Write-Host "1. Running go vet..." -ForegroundColor Cyan
try {
    go vet ./...
    Write-Host "PASS: go vet passed" -ForegroundColor Green
} catch {
    Write-Host "FAIL: go vet failed" -ForegroundColor Red
    $errors++
}
Write-Host ""

# 2. golangci-lint
Write-Host "2. Running golangci-lint..." -ForegroundColor Cyan
if (Test-Tool "golangci-lint" "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest") {
    try {
        golangci-lint run
        Write-Host "PASS: golangci-lint passed" -ForegroundColor Green
    } catch {
        Write-Host "FAIL: golangci-lint failed" -ForegroundColor Red
        $errors++
    }
} else {
    Write-Host "SKIP: Skipping golangci-lint (not installed)" -ForegroundColor Yellow
}
Write-Host ""

# 3. staticcheck
Write-Host "3. Running staticcheck..." -ForegroundColor Cyan
if (Test-Tool "staticcheck" "go install honnef.co/go/tools/cmd/staticcheck@latest") {
    try {
        staticcheck ./...
        Write-Host "PASS: staticcheck passed" -ForegroundColor Green
    } catch {
        Write-Host "FAIL: staticcheck failed" -ForegroundColor Red
        $errors++
    }
} else {
    Write-Host "SKIP: Skipping staticcheck (not installed)" -ForegroundColor Yellow
}
Write-Host ""

# 4. gosec (security)
Write-Host "4. Running gosec (security analysis)..." -ForegroundColor Cyan
if (Test-Tool "gosec" "go install github.com/securego/gosec/v2/cmd/gosec@latest") {
    try {
        gosec -quiet ./...
        Write-Host "PASS: gosec passed" -ForegroundColor Green
    } catch {
        Write-Host "FAIL: gosec found security issues" -ForegroundColor Red
        $errors++
    }
} else {
    Write-Host "SKIP: Skipping gosec (not installed)" -ForegroundColor Yellow
}
Write-Host ""

# 5. govulncheck (vulnerability scanning)
Write-Host "5. Running govulncheck..." -ForegroundColor Cyan
if (Test-Tool "govulncheck" "go install golang.org/x/vuln/cmd/govulncheck@latest") {
    try {
        govulncheck ./...
        Write-Host "PASS: govulncheck passed" -ForegroundColor Green
    } catch {
        Write-Host "FAIL: govulncheck found vulnerabilities" -ForegroundColor Red
        $errors++
    }
} else {
    Write-Host "SKIP: Skipping govulncheck (not installed)" -ForegroundColor Yellow
}
Write-Host ""

# 6. errcheck
Write-Host "6. Running errcheck..." -ForegroundColor Cyan
if (Test-Tool "errcheck" "go install github.com/kisielk/errcheck@latest") {
    try {
        errcheck ./...
        Write-Host "PASS: errcheck passed" -ForegroundColor Green
    } catch {
        Write-Host "FAIL: errcheck found unchecked errors" -ForegroundColor Red
        $errors++
    }
} else {
    Write-Host "SKIP: Skipping errcheck (not installed)" -ForegroundColor Yellow
}
Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
if ($errors -eq 0) {
    Write-Host "All static analysis checks passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Static analysis found $errors issue(s)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Install missing tools with:" -ForegroundColor Yellow
    Write-Host '  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'
    Write-Host '  go install honnef.co/go/tools/cmd/staticcheck@latest'
    Write-Host '  go install github.com/securego/gosec/v2/cmd/gosec@latest'
    Write-Host '  go install golang.org/x/vuln/cmd/govulncheck@latest'
    Write-Host '  go install github.com/kisielk/errcheck@latest'
    exit 1
}
