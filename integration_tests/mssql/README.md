# SQL Server (MSSQL) Integration Tests

This directory contains comprehensive integration tests for using typedb with Microsoft SQL Server. These tests cover both **happy paths** (successful operations) and **negative paths** (error cases).

For simple example programs demonstrating the happy path, see [`../../examples/mssql/`](../../examples/mssql/).

## Prerequisites

- SQL Server database server (2012+)
- Go 1.18 or later
- SQL Server driver: `github.com/microsoft/go-mssqldb`

## Setup

1. **Install dependencies:**
   ```bash
   go get github.com/microsoft/go-mssqldb
   ```

2. **Create the database:**
   ```sql
   CREATE DATABASE typedb_examples;
   ```

3. **Run migrations:**
   ```bash
   # Install golang-migrate (if not already installed)
   # macOS: brew install golang-migrate
   # Linux: See https://github.com/golang-migrate/migrate
   
   # Run migrations
   cd integration_tests/mssql
   migrate -path migrations -database "sqlserver://sa:YourPassword123@localhost:1433?database=typedb_examples" up
   ```

4. **Set environment variable (optional):**
   ```bash
   export MSSQL_DSN="server=localhost;user id=sa;password=YourPassword123;database=typedb_examples"
   ```

   Or use the default DSN: `server=localhost;user id=sa;password=YourPassword123;database=typedb_examples`

## Running Integration Tests

**Run all tests:**
```bash
go test -v
```

**Run with specific test:**
```bash
go test -v -run TestMSSQL_Load
```

**Note:** Tests will fail if database connection fails, ensuring databases are properly configured in CI/CD environments.

## Test Coverage

### Happy Path Tests
- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ SQL Server-specific features:
  - Named parameters (`@p1`, `@p2`)
  - JSON stored as NVARCHAR(MAX)
  - Composite keys

### Negative Path Tests
- ✅ Error handling for invalid queries
- ✅ Error handling for missing records
- ✅ Error handling for constraint violations
- ✅ Error handling for invalid data types

## Database Schema

The tests use three tables:

- **users** - User accounts with primary key and unique email
- **posts** - Blog posts with JSON stored as NVARCHAR(MAX)
- **user_posts** - Many-to-many relationship with composite key

See `schema.sql` for the complete schema definition.

## Differences from PostgreSQL/MySQL

- Uses named parameters (`@p1`, `@p2`) instead of `?` or `$1`
- Uses `IDENTITY(1,1)` instead of `SERIAL` or `AUTO_INCREMENT`
- Uses `NVARCHAR` for Unicode strings
- Uses `DATETIME2` for timestamps
- JSON stored as `NVARCHAR(MAX)` (no native JSON type in older versions)

## Connection String Format

SQL Server uses a connection string format:
```
server=localhost;user id=sa;password=YourPassword123;database=typedb_examples
```

Common parameters:
- `server` - Server name or IP address
- `user id` - SQL Server login
- `password` - SQL Server password
- `database` - Database name
- `encrypt=disable` - Disable encryption (for local development)
- `trustservercertificate=true` - Trust server certificate
