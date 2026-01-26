# typedb Integration Tests

This directory contains comprehensive integration tests for using typedb with different database systems. These tests cover both **happy paths** (successful operations) and **negative paths** (error cases) to verify typedb's functionality.

For simple example programs demonstrating the happy path, see the [`../examples/`](../examples/) directory.

## Overview

Each database type has its own subdirectory with:
- **Integration tests** (`integration_test.go`) - Comprehensive tests covering happy and negative paths
- **Migrations** - Database migration files (using golang-migrate format)
- **Schema files** - Legacy schema.sql files (for reference, migrations are preferred)
- **README** - Database-specific setup and usage instructions

## Supported Databases

- **[PostgreSQL](postgresql/)** - Full-featured with arrays and JSONB
- **[MySQL](mysql/)** - Popular open-source database
- **[SQLite](sqlite/)** - File-based, perfect for development
- **[SQL Server (MSSQL)](mssql/)** - Microsoft SQL Server
- **[Oracle](oracle/)** - Enterprise Oracle Database

## Quick Start

1. **Choose a database** from the list above
2. **Install golang-migrate:**
   ```bash
   # macOS
   brew install golang-migrate
   
   # Linux (or download from https://github.com/golang-migrate/migrate)
   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
   sudo mv migrate /usr/local/bin/migrate
   ```
3. **Navigate to the database directory:**
   ```bash
   cd integration_tests/postgresql  # or mysql, sqlite, mssql, oracle
   ```
4. **Follow the README** in that directory for setup instructions
5. **Run migrations** to set up the database schema
6. **Run integration tests:**
   ```bash
   go test -v
   ```

## Test Coverage

Integration tests verify:

### Happy Path Tests
- ✅ Basic queries (`QueryAll`, `QueryFirst`, `QueryOne`)
- ✅ Model loading (`Load`, `LoadByField`, `LoadByComposite`)
- ✅ Transactions (`WithTx`)
- ✅ Database-specific features (arrays, JSONB, etc.)

### Negative Path Tests
- ✅ Error handling for invalid queries
- ✅ Error handling for missing records
- ✅ Error handling for constraint violations
- ✅ Error handling for invalid data types
- ✅ Transaction rollback scenarios

## CI/CD Integration

All integration tests are designed to:
- **Fail clearly** if database is unavailable (ensures proper CI/CD setup)
- **Clean up** after themselves (where applicable)
- **Work independently** without affecting other tests

This ensures CI/CD pipelines properly configure databases before running tests.

## Running All Tests

To run integration tests for all databases (requires all databases to be set up):

```bash
# From the integration_tests directory
for dir in */; do
    echo "Testing $dir"
    cd "$dir"
    go test -v
    cd ..
done
```

**Note:** Tests will fail if database connections fail, ensuring databases are properly configured in CI/CD environments.

## Contributing

When adding integration tests for a new database:

1. Create a new subdirectory
2. Include:
   - `schema.sql` - Database schema
   - `models.go` - Model definitions
   - `integration_test.go` - Comprehensive integration tests (happy and negative paths)
   - `migrations/` - Database migration files
   - `README.md` - Setup and usage instructions
3. Follow the patterns established in existing tests
4. Ensure tests fail clearly if database is unavailable (for proper CI/CD validation)
5. Cover both happy and negative paths

## License

Integration tests are provided under the same license as typedb (MIT License).
