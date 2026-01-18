# SQLite Integration Tests

This directory contains comprehensive integration tests for using typedb with SQLite. These tests cover both **happy paths** (successful operations) and **negative paths** (error cases).

For simple example programs demonstrating the happy path, see [`../../examples/sqlite/`](../../examples/sqlite/).

## Prerequisites

- Go 1.18 or later
- SQLite driver: `github.com/mattn/go-sqlite3`

**Note:** SQLite requires CGO to be enabled. Make sure CGO is enabled when building:
```bash
export CGO_ENABLED=1
```

## Setup

1. **Install dependencies:**
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

2. **Set environment variable (optional):**
   ```bash
   export SQLITE_DSN="typedb_examples_test.db"
   ```

   Or use the default DSN: `typedb_examples_test.db`

3. **Run migrations (optional):**
   ```bash
   # Install golang-migrate (if not already installed)
   # macOS: brew install golang-migrate
   # Linux: See https://github.com/golang-migrate/migrate
   
   # Run migrations
   cd integration_tests/sqlite
   CGO_ENABLED=1 migrate -path migrations -database "sqlite3://typedb_examples_test.db" up
   ```

   Tests will create the database automatically if it doesn't exist, but you'll need to run migrations to set up the schema.

## Running Integration Tests

**Run all tests:**
```bash
CGO_ENABLED=1 go test -v
```

**Run with specific test:**
```bash
CGO_ENABLED=1 go test -v -run TestSQLite_Load
```

**Note:** Tests create a temporary database file (`typedb_examples_test.db`) that is cleaned up after tests complete.

## Test Coverage

### Happy Path Tests
- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ SQLite-specific features:
  - File-based database (no server required)
  - JSON stored as TEXT
  - Composite keys

### Negative Path Tests
- ✅ Error handling for invalid queries
- ✅ Error handling for missing records
- ✅ Error handling for constraint violations
- ✅ Error handling for invalid data types

## Database Schema

The tests use three tables:

- **users** - User accounts with primary key and unique email
- **posts** - Blog posts with JSON stored as TEXT
- **user_posts** - Many-to-many relationship with composite key

See `schema.sql` for the complete schema definition.

## Differences from PostgreSQL/MySQL

- Uses `?` placeholders (like MySQL)
- File-based database (no server required)
- `INTEGER PRIMARY KEY AUTOINCREMENT` instead of `SERIAL` or `AUTO_INCREMENT`
- JSON stored as TEXT (no native JSON type)
- Perfect for development and testing

## CGO Requirements

SQLite driver requires CGO. If you encounter build errors:

1. Install SQLite development libraries:
   - **macOS:** `brew install sqlite`
   - **Linux:** `sudo apt-get install libsqlite3-dev` (Debian/Ubuntu) or `sudo yum install sqlite-devel` (RHEL/CentOS)
   - **Windows:** Download from [SQLite website](https://www.sqlite.org/download.html)

2. Ensure CGO is enabled:
   ```bash
   export CGO_ENABLED=1
   ```
