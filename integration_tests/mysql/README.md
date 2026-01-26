# MySQL Integration Tests

This directory contains comprehensive integration tests for using typedb with MySQL. These tests cover both **happy paths** (successful operations) and **negative paths** (error cases).

For simple example programs demonstrating the happy path, see [`../../examples/mysql/`](../../examples/mysql/).

## Prerequisites

- MySQL database server (5.7+ or 8.0+)
- Go 1.18 or later
- MySQL driver: `github.com/go-sql-driver/mysql`

## Setup

1. **Install dependencies:**
   ```bash
   go get github.com/go-sql-driver/mysql
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
   cd integration_tests/mysql
   migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/typedb_examples?parseTime=true" up
   ```

4. **Set environment variable (optional):**
   ```bash
   export MYSQL_DSN="user:password@tcp(localhost:3306)/typedb_examples?parseTime=true"
   ```

   Or use the default DSN: `user:password@tcp(localhost:3306)/typedb_examples?parseTime=true`

## Running Integration Tests

**Run all tests:**
```bash
go test -v
```

**Run with specific test:**
```bash
go test -v -run TestMySQL_Load
```

**Note:** Tests will fail if database connection fails, ensuring databases are properly configured in CI/CD environments.

## Test Coverage

### Happy Path Tests
- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ MySQL-specific features:
  - JSON columns
  - Composite keys

### Negative Path Tests
- ✅ Error handling for invalid queries
- ✅ Error handling for missing records
- ✅ Error handling for constraint violations
- ✅ Error handling for invalid data types

## Database Schema

The tests use three tables:

- **users** - User accounts with primary key and unique email
- **posts** - Blog posts with MySQL JSON columns
- **user_posts** - Many-to-many relationship with composite key

See `schema.sql` for the complete schema definition.

## Differences from PostgreSQL

- Uses `?` placeholders instead of `$1`, `$2`
- Uses `AUTO_INCREMENT` instead of `SERIAL`
- JSON type instead of JSONB
- `parseTime=true` parameter required in DSN for proper timestamp handling
