#!/bin/bash
# Script to run all database tests using docker-compose
# Usage: ./run-tests.sh [database]
# If database is specified, only run tests for that database
# Valid databases: postgresql, mysql, mssql, sqlite, oracle, all

set -e

DATABASE="${1:-all}"

echo "🚀 Starting typedb example tests..."
echo "Database: $DATABASE"

# Function to run tests for a specific database
run_database_tests() {
    local db=$1
    echo ""
    echo "=========================================="
    echo "Testing $db..."
    echo "=========================================="
    
    case $db in
        postgresql)
            docker-compose up -d postgresql
            echo "⏳ Waiting for PostgreSQL to be healthy..."
            docker-compose up migrate-postgresql
            echo "✅ Migrations completed, running tests..."
            docker-compose run --rm test-postgresql
            ;;
        mysql)
            docker-compose up -d mysql
            echo "⏳ Waiting for MySQL to be healthy..."
            docker-compose up migrate-mysql
            echo "✅ Migrations completed, running tests..."
            docker-compose run --rm test-mysql
            ;;
        mssql)
            docker-compose up -d mssql
            echo "⏳ Waiting for SQL Server to be healthy..."
            docker-compose up mssql-init migrate-mssql
            echo "✅ Migrations completed, running tests..."
            docker-compose run --rm test-mssql
            ;;
        sqlite)
            echo "Running SQLite tests (no container needed)..."
            docker-compose run --rm test-sqlite
            ;;
        oracle)
            echo "⚠️  Oracle requires authentication and license acceptance"
            echo "See README-DOCKER.md for setup instructions"
            docker-compose up -d oracle
            echo "⏳ Waiting for Oracle to be healthy..."
            docker-compose up migrate-oracle
            echo "✅ Migrations completed, running tests..."
            docker-compose run --rm test-oracle
            ;;
        *)
            echo "❌ Unknown database: $db"
            echo "Valid databases: postgresql, mysql, mssql, sqlite, oracle, all"
            exit 1
            ;;
    esac
}

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    docker-compose down
}

trap cleanup EXIT

if [ "$DATABASE" = "all" ]; then
    echo "Running tests for all databases..."
    run_database_tests postgresql
    run_database_tests mysql
    run_database_tests mssql
    run_database_tests sqlite
    
    # Check if Oracle is set up
    if docker images | grep -q "container-registry.oracle.com/database/free"; then
        echo ""
        echo "Oracle image found. Running Oracle tests..."
        run_database_tests oracle
    else
        echo ""
        echo "⚠️  Oracle not set up. Run ./setup-oracle.sh to enable Oracle testing."
    fi
else
    run_database_tests "$DATABASE"
fi

echo ""
echo "✅ All tests completed!"
