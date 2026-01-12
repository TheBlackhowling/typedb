# PowerShell script for automated Oracle Database setup
# Oracle Database Free is free for development and testing purposes

$ErrorActionPreference = "Stop"

$OracleRegistry = "container-registry.oracle.com"
$OracleImage = "database/free:latest"

Write-Host "🔐 Oracle Database Setup for Testing" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Oracle Database Free is free for development and testing purposes." -ForegroundColor Yellow
Write-Host "This script will:" -ForegroundColor Yellow
Write-Host "  1. Check if you're logged into Oracle Container Registry" -ForegroundColor Yellow
Write-Host "  2. Pull the Oracle Database Free image" -ForegroundColor Yellow
Write-Host "  3. Verify license acceptance" -ForegroundColor Yellow
Write-Host ""

# Check if already logged in (basic check)
$dockerInfo = docker info 2>&1
if ($dockerInfo -match $OracleRegistry) {
    Write-Host "✅ Already authenticated to Oracle Container Registry" -ForegroundColor Green
} else {
    Write-Host "📝 Oracle Container Registry Authentication Required" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To use Oracle Database Free (free for testing):" -ForegroundColor Yellow
    Write-Host "  1. Create a free account at: https://container-registry.oracle.com/" -ForegroundColor Yellow
    Write-Host "  2. Accept the license terms for 'database/free'" -ForegroundColor Yellow
    Write-Host "  3. Run: docker login $OracleRegistry" -ForegroundColor Yellow
    Write-Host ""
    $response = Read-Host "Have you created an account and accepted the license? (y/n)"
    
    if ($response -ne "y" -and $response -ne "Y") {
        Write-Host ""
        Write-Host "Please visit https://container-registry.oracle.com/ to:" -ForegroundColor Yellow
        Write-Host "  1. Create a free account" -ForegroundColor Yellow
        Write-Host "  2. Navigate to 'database/free'" -ForegroundColor Yellow
        Write-Host "  3. Accept the license terms" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Then run this script again." -ForegroundColor Yellow
        exit 1
    }
    
    Write-Host ""
    Write-Host "🔑 Logging into Oracle Container Registry..." -ForegroundColor Cyan
    docker login $OracleRegistry
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Login failed. Please check your credentials." -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "📥 Pulling Oracle Database Free image..." -ForegroundColor Cyan
docker pull "$OracleRegistry/$OracleImage"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Oracle Database Free image pulled successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You can now use Oracle in docker-compose:" -ForegroundColor Green
    Write-Host "  docker-compose up oracle migrate-oracle test-oracle" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Or run all tests including Oracle:" -ForegroundColor Green
    Write-Host "  .\run-tests.ps1 all" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "❌ Failed to pull Oracle image." -ForegroundColor Red
    Write-Host "Make sure you've accepted the license terms at:" -ForegroundColor Yellow
    Write-Host "  https://container-registry.oracle.com/" -ForegroundColor Yellow
    exit 1
}
