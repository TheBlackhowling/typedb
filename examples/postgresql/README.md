# PostgreSQL Examples

This directory contains example programs demonstrating how to use typedb with PostgreSQL. These examples show the **happy path** - successful operations that demonstrate typedb's features.

For comprehensive integration tests (including happy and negative paths), see [`../../integration_tests/postgresql/`](../../integration_tests/postgresql/).

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
   cd examples/postgresql
   migrate -path migrations -database "postgres://user:password@localhost/typedb_examples?sslmode=disable" up
   ```

4. **Set environment variable (optional):**
   ```bash
   export POSTGRES_DSN="postgres://user:password@localhost/typedb_examples?sslmode=disable"
   ```

   Or use the default DSN: `postgres://user:password@localhost/typedb_examples?sslmode=disable`

## Running Examples

**Run the example program:**
```bash
go run example.go models.go
```

This will demonstrate the happy path - successful operations with typedb.

**Note:** For comprehensive integration tests (including error cases), see [`../../integration_tests/postgresql/`](../../integration_tests/postgresql/).

## Features Demonstrated

- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ PostgreSQL-specific features:
  - Arrays (`TEXT[]`)
  - JSONB columns
  - Composite keys

## Database Schema

The example uses three tables:

- **users** - User accounts with primary key and unique email
- **posts** - Blog posts with PostgreSQL arrays and JSONB
- **user_posts** - Many-to-many relationship with composite key

See `schema.sql` for the complete schema definition.
