# Check that Go version meets minimum requirement (1.23.12+)

$MIN_VERSION = "1.23.12"
$CURRENT_VERSION = (go version).Split(' ')[2].TrimStart('go')

Write-Host "Checking Go version..."
Write-Host "Current: $CURRENT_VERSION"
Write-Host "Required: $MIN_VERSION or later"

function Compare-Versions {
    param(
        [string]$Version1,
        [string]$Version2
    )
    
    $v1 = [version]$Version1
    $v2 = [version]$Version2
    
    return $v1 -ge $v2
}

if (Compare-Versions -Version1 $CURRENT_VERSION -Version2 $MIN_VERSION) {
    Write-Host "✅ Go version meets requirement" -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ Go version $CURRENT_VERSION is below minimum requirement $MIN_VERSION" -ForegroundColor Red
    Write-Host "Please upgrade to Go $MIN_VERSION or later to address security vulnerabilities:"
    Write-Host "  - GO-2025-3849 (requires Go 1.23.12+)"
    Write-Host "  - GO-2025-3750 (requires Go 1.23.10+)"
    exit 1
}
