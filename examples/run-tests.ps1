# PowerShell script to run all database tests using docker-compose
# Usage: .\run-tests.ps1 [database]
# If database is specified, only run tests for that database
# Valid databases: postgresql, mysql, mssql, sqlite, oracle, all

param(
    [string]$Database = "all"
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 Starting typedb example tests..." -ForegroundColor Cyan
Write-Host "Database: $Database" -ForegroundColor Cyan

function Run-DatabaseTests {
    param([string]$db)
    
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Yellow
    Write-Host "Testing $db..." -ForegroundColor Yellow
    Write-Host "==========================================" -ForegroundColor Yellow
    
    switch ($db) {
        "postgresql" {
            docker-compose up -d postgresql
            Write-Host "⏳ Waiting for PostgreSQL to be healthy..." -ForegroundColor Yellow
            docker-compose up migrate-postgresql
            Write-Host "✅ Migrations completed, running tests..." -ForegroundColor Green
            docker-compose run --rm test-postgresql
        }
        "mysql" {
            docker-compose up -d mysql
            Write-Host "⏳ Waiting for MySQL to be healthy..." -ForegroundColor Yellow
            docker-compose up migrate-mysql
            Write-Host "✅ Migrations completed, running tests..." -ForegroundColor Green
            docker-compose run --rm test-mysql
        }
        "mssql" {
            docker-compose up -d mssql
            Write-Host "⏳ Waiting for SQL Server to be healthy..." -ForegroundColor Yellow
            docker-compose up mssql-init migrate-mssql
            Write-Host "✅ Migrations completed, running tests..." -ForegroundColor Green
            docker-compose run --rm test-mssql
        }
        "sqlite" {
            Write-Host "Running SQLite tests (no container needed)..." -ForegroundColor Yellow
            docker-compose run --rm test-sqlite
        }
        "oracle" {
            Write-Host "⚠️  Oracle requires authentication and license acceptance" -ForegroundColor Yellow
            Write-Host "See README-DOCKER.md for setup instructions" -ForegroundColor Yellow
            docker-compose up -d oracle
            Write-Host "⏳ Waiting for Oracle to be healthy..." -ForegroundColor Yellow
            docker-compose up migrate-oracle
            Write-Host "✅ Migrations completed, running tests..." -ForegroundColor Green
            docker-compose run --rm test-oracle
        }
        default {
            Write-Host "❌ Unknown database: $db" -ForegroundColor Red
            Write-Host "Valid databases: postgresql, mysql, mssql, sqlite, oracle, all" -ForegroundColor Yellow
            exit 1
        }
    }
}

function Cleanup {
    Write-Host ""
    Write-Host "🧹 Cleaning up..." -ForegroundColor Yellow
    docker-compose down
}

try {
    if ($Database -eq "all") {
        Write-Host "Running tests for all databases..." -ForegroundColor Cyan
        Run-DatabaseTests "postgresql"
        Run-DatabaseTests "mysql"
        Run-DatabaseTests "mssql"
        Run-DatabaseTests "sqlite"
        
        # Check if Oracle is set up
        $oracleImage = docker images --format "{{.Repository}}:{{.Tag}}" | Select-String "container-registry.oracle.com/database/free"
        if ($oracleImage) {
            Write-Host ""
            Write-Host "Oracle image found. Running Oracle tests..." -ForegroundColor Cyan
            Run-DatabaseTests "oracle"
        } else {
            Write-Host ""
            Write-Host "⚠️  Oracle not set up. Run .\setup-oracle.ps1 to enable Oracle testing." -ForegroundColor Yellow
        }
    } else {
        Run-DatabaseTests $Database
    }
    
    Write-Host ""
    Write-Host "✅ All tests completed!" -ForegroundColor Green
} finally {
    Cleanup
}
