# MySQL Examples

This directory contains example programs demonstrating how to use typedb with MySQL. These examples show the **happy path** - successful operations that demonstrate typedb's features.

For comprehensive integration tests (including happy and negative paths), see [`../../integration_tests/mysql/`](../../integration_tests/mysql/).

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
   cd examples/mysql
   migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/typedb_examples?parseTime=true" up
   ```

3. **Set environment variable (optional):**
   ```bash
   export MYSQL_DSN="user:password@tcp(localhost:3306)/typedb_examples?parseTime=true"
   ```

   Or use the default DSN: `user:password@tcp(localhost:3306)/typedb_examples?parseTime=true`

## Running Examples

**Run the example program:**
```bash
go run example.go models.go
```

This will demonstrate the happy path - successful operations with typedb.

**Note:** For comprehensive integration tests (including error cases), see [`../../integration_tests/mysql/`](../../integration_tests/mysql/).

## Features Demonstrated

- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ MySQL-specific features:
  - JSON columns
  - Composite keys

## Database Schema

The example uses three tables:

- **users** - User accounts with primary key and unique email
- **posts** - Blog posts with MySQL JSON columns
- **user_posts** - Many-to-many relationship with composite key

See `schema.sql` for the complete schema definition.

## Differences from PostgreSQL

- Uses `?` placeholders instead of `$1`, `$2`
- Uses `AUTO_INCREMENT` instead of `SERIAL`
- JSON type instead of JSONB
- `parseTime=true` parameter required in DSN for proper timestamp handling
