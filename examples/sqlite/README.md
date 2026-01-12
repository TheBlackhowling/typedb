# SQLite Examples

This directory contains example programs demonstrating how to use typedb with SQLite. These examples show the **happy path** - successful operations that demonstrate typedb's features.

For comprehensive integration tests (including happy and negative paths), see [`../../integration_tests/sqlite/`](../../integration_tests/sqlite/).

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
   export SQLITE_DSN="typedb_examples.db"
   ```

   Or use the default DSN: `typedb_examples.db`

3. **Run migrations (optional):**
   ```bash
   # Install golang-migrate (if not already installed)
   # macOS: brew install golang-migrate
   # Linux: See https://github.com/golang-migrate/migrate
   
   # Run migrations
   cd examples/sqlite
   CGO_ENABLED=1 migrate -path migrations -database "sqlite3://typedb_examples.db" up
   ```

   The example program will create the database automatically if it doesn't exist, but you'll need to run migrations to set up the schema.

## Running Examples

**Run the example program:**
```bash
CGO_ENABLED=1 go run example.go models.go
```

**Note:** Make sure CGO is enabled! This will demonstrate the happy path - successful operations with typedb.

**Note:** For comprehensive integration tests (including error cases), see [`../../integration_tests/sqlite/`](../../integration_tests/sqlite/).

## Features Demonstrated

- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ SQLite-specific features:
  - File-based database (no server required)
  - JSON stored as TEXT
  - Composite keys

## Database Schema

The example uses three tables:

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
