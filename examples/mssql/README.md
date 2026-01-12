# SQL Server (MSSQL) Examples

This directory contains example programs demonstrating how to use typedb with Microsoft SQL Server. These examples show the **happy path** - successful operations that demonstrate typedb's features.

For comprehensive integration tests (including happy and negative paths), see [`../../integration_tests/mssql/`](../../integration_tests/mssql/).

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
   cd examples/mssql
   migrate -path migrations -database "sqlserver://sa:YourPassword123@localhost:1433?database=typedb_examples" up
   ```

3. **Set environment variable (optional):**
   ```bash
   export MSSQL_DSN="server=localhost;user id=sa;password=YourPassword123;database=typedb_examples"
   ```

   Or use the default DSN: `server=localhost;user id=sa;password=YourPassword123;database=typedb_examples`

## Running Examples

**Run the example program:**
```bash
go run example.go models.go
```

This will demonstrate the happy path - successful operations with typedb.

**Note:** For comprehensive integration tests (including error cases), see [`../../integration_tests/mssql/`](../../integration_tests/mssql/).

## Features Demonstrated

- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ SQL Server-specific features:
  - Named parameters (`@p1`, `@p2`)
  - JSON stored as NVARCHAR(MAX)
  - Composite keys

## Database Schema

The example uses three tables:

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
