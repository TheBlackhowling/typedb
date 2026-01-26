# PostgreSQL Integration Tests

This directory contains comprehensive integration tests for using typedb with PostgreSQL. These tests cover both **happy paths** (successful operations) and **negative paths** (error cases).

For simple example programs demonstrating the happy path, see [`../../examples/postgresql/`](../../examples/postgresql/).

## Prerequisites

- PostgreSQL database server
- Go 1.18 or later
- PostgreSQL driver: `github.com/lib/pq`

## Setup

1. **Install dependencies:**
   ```bash
   go get github.com/lib/pq
   ```

2. **Create the database:**
   ```bash
   createdb typedb_examples
   ```

3. **Run migrations:**
   ```bash
   # Install golang-migrate (if not already installed)
   # macOS: brew install golang-migrate
   # Linux: See https://github.com/golang-migrate/migrate
   
   # Run migrations
   cd integration_tests/postgresql
   migrate -path migrations -database "postgres://user:password@localhost/typedb_examples?sslmode=disable" up
   ```

4. **Set environment variable (optional):**
   ```bash
   export POSTGRES_DSN="postgres://user:password@localhost/typedb_examples?sslmode=disable"
   ```

   Or use the default DSN: `postgres://user:password@localhost/typedb_examples?sslmode=disable`

## Running Integration Tests

**Run all tests:**
```bash
go test -v
```

**Run with specific test:**
```bash
go test -v -run TestPostgreSQL_Load
```

**Note:** Tests will fail if database connection fails, ensuring databases are properly configured in CI/CD environments.

## Test Coverage

### Happy Path Tests
- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ PostgreSQL-specific features:
  - Arrays (`TEXT[]`)
  - JSONB columns
  - Composite keys
  - Comprehensive type handling

### Negative Path Tests
- ✅ Error handling for invalid queries
- ✅ Error handling for missing records
- ✅ Error handling for constraint violations
- ✅ Error handling for invalid data types

## Database Schema

The tests use three tables:

- **users** - User accounts with primary key and unique email
- **posts** - Blog posts with PostgreSQL arrays and JSONB
- **user_posts** - Many-to-many relationship with composite key
- **type_examples** - Comprehensive type examples for PostgreSQL-specific types

See `schema.sql` for the complete schema definition.
