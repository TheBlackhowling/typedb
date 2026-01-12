# Docker Compose Setup for Examples

This directory contains a `docker-compose.yml` file that orchestrates all database services, migrations, and tests for the typedb examples.

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- No need to install database servers locally
- No need to install golang-migrate locally

### Running All Tests

**Linux/macOS:**
```bash
cd examples
chmod +x run-tests.sh
./run-tests.sh
```

**Windows (PowerShell):**
```powershell
cd examples
.\run-tests.ps1
```

**Using docker-compose directly:**
```bash
# Start all databases
docker-compose up -d postgresql mysql mssql

# Wait for databases to be healthy, then run migrations
docker-compose up migrate-postgresql migrate-mysql migrate-mssql

# Run tests
docker-compose run --rm test-postgresql
docker-compose run --rm test-mysql
docker-compose run --rm test-mssql
docker-compose run --rm test-sqlite

# Cleanup
docker-compose down
```

### Running Tests for a Specific Database

**Linux/macOS:**
```bash
./run-tests.sh postgresql
./run-tests.sh mysql
./run-tests.sh mssql
./run-tests.sh sqlite
```

**Windows:**
```powershell
.\run-tests.ps1 postgresql
.\run-tests.ps1 mysql
.\run-tests.ps1 mssql
.\run-tests.ps1 sqlite
```

## Architecture

The docker-compose setup includes:

### Database Services

- **postgresql** - PostgreSQL 16 (Alpine)
- **mysql** - MySQL 8.0
- **mssql** - SQL Server 2022

### Migration Services

- **migrate-postgresql** - Runs migrations for PostgreSQL
- **migrate-mysql** - Runs migrations for MySQL
- **mssql-init** - Creates the SQL Server database (required before migrations)
- **migrate-mssql** - Runs migrations for SQL Server

### Test Services

- **test-postgresql** - Runs integration tests for PostgreSQL
- **test-mysql** - Runs integration tests for MySQL
- **test-mssql** - Runs integration tests for SQL Server
- **test-sqlite** - Runs integration tests for SQLite (includes migrations)

## Service Dependencies

The services are orchestrated with proper dependencies:

1. **Database services** start first and wait for health checks
2. **Migration services** depend on database health checks
3. **Test services** depend on migration completion

Each database type has its own isolated network:

```
PostgreSQL Network:
  postgresql (health check) → migrate-postgresql → test-postgresql

MySQL Network:
  mysql (health check) → migrate-mysql → test-mysql

SQL Server Network:
  mssql (health check) → mssql-init → migrate-mssql → test-mssql

SQLite (standalone, includes migrations):
  test-sqlite
```

This isolation ensures:
- **Security**: Databases cannot communicate with each other
- **Clarity**: Easy to see which services belong together
- **Realism**: More closely matches production environments

## Configuration

### Database Connection Strings

The test services use these environment variables:

- **PostgreSQL:** `POSTGRES_DSN=postgres://user:password@postgresql:5432/typedb_examples?sslmode=disable`
- **MySQL:** `MYSQL_DSN=user:password@tcp(mysql:3306)/typedb_examples?parseTime=true`
- **SQL Server:** `MSSQL_DSN=sqlserver://sa:YourStrong!Passw0rd@mssql:1433?database=master`
- **SQLite:** `SQLITE_DSN=typedb_examples_test.db`

### Ports

Default ports (can be overridden in `docker-compose.override.yml`):

- PostgreSQL: `5432`
- MySQL: `3306`
- SQL Server: `1433`

### Volumes

Data persistence volumes:
- `postgresql_data` - PostgreSQL data directory
- `mysql_data` - MySQL data directory
- `mssql_data` - SQL Server data directory
- `go_cache` - Go module cache (shared across test containers)

## Health Checks

Each database service includes health checks:

- **PostgreSQL:** Uses `pg_isready`
- **MySQL:** Uses `mysqladmin ping`
- **SQL Server:** Uses `sqlcmd` to execute a simple query

Health checks run every 5-10 seconds and wait up to 50-100 seconds for the database to be ready.

## Development Workflow

### Running Tests Locally

1. **Start databases:**
   ```bash
   docker-compose up -d postgresql mysql mssql
   ```

2. **Run migrations manually (if needed):**
   ```bash
   docker-compose run --rm migrate-postgresql
   # For SQL Server, create database first:
   docker-compose run --rm mssql-init
   docker-compose run --rm migrate-mssql
   ```

3. **Run tests:**
   ```bash
   docker-compose run --rm test-postgresql
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f postgresql
   ```

### Debugging

**View database logs:**
```bash
docker-compose logs postgresql
docker-compose logs mysql
docker-compose logs mssql
```

**Connect to database directly:**
```bash
# PostgreSQL
docker-compose exec postgresql psql -U user -d typedb_examples

# MySQL
docker-compose exec mysql mysql -u user -ppassword typedb_examples

# SQL Server
docker-compose exec mssql /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P 'YourStrong!Passw0rd'
```

**Run a single test:**
```bash
docker-compose run --rm test-postgresql go test -v -run TestPostgreSQL_Load
```

## Cleanup

**Stop all services:**
```bash
docker-compose down
```

**Stop and remove volumes (clean slate):**
```bash
docker-compose down -v
```

**Remove only specific volumes:**
```bash
docker volume rm examples_postgresql_data
docker volume rm examples_mysql_data
docker volume rm examples_mssql_data
```

## Troubleshooting

### SQL Server Takes Too Long to Start

SQL Server can take 30+ seconds to start. The health check includes a `start_period` to account for this. If tests fail, wait a bit longer and try again.

### Migration Fails

Check the migration logs:
```bash
docker-compose logs migrate-postgresql
```

Ensure the database is healthy before migrations run:
```bash
docker-compose ps
```

### Port Conflicts

If ports are already in use, modify `docker-compose.override.yml`:
```yaml
services:
  postgresql:
    ports:
      - "5433:5432"  # Use different host port
```

### Permission Issues (Linux)

If you get permission errors, ensure your user is in the `docker` group:
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

## Oracle Database

Oracle Database Free is **enabled by default** in `docker-compose.yml`. Oracle Database Free is free for development and testing purposes.

### Automated Setup

**Quick Setup (Recommended):**

```bash
# Linux/macOS
./setup-oracle.sh

# Windows PowerShell
.\setup-oracle.ps1
```

The setup script will:
1. Guide you through Oracle Container Registry authentication
2. Pull the Oracle Database Free image
3. Verify license acceptance

### Manual Setup

If you prefer manual setup:

1. **Create Oracle Account**:
   - Visit https://container-registry.oracle.com/
   - Create a free account

2. **Accept License Terms**:
   - Navigate to `database/free` in the registry
   - Accept the license terms (free for development/testing)

3. **Login to Oracle Container Registry**:
   ```bash
   docker login container-registry.oracle.com
   ```

4. **Pull Oracle Image**:
   ```bash
   docker pull container-registry.oracle.com/database/free:23.4.0-free
   ```

5. **Run Oracle Tests**:
   ```bash
   docker-compose up oracle migrate-oracle test-oracle
   ```

### Oracle Database Free

Oracle Database Free is Oracle's free tier database, perfect for:
- ✅ Development
- ✅ Testing
- ✅ Learning
- ✅ CI/CD pipelines

**Note**: Oracle Database Free has some limitations compared to Enterprise Edition, but it's fully sufficient for testing typedb functionality.

### Running Tests

Oracle tests are included when running all tests:

```bash
./run-tests.sh all        # Includes Oracle if image is available
.\run-tests.ps1 all       # PowerShell version
```

Or run Oracle tests specifically:

```bash
./run-tests.sh oracle
.\run-tests.ps1 oracle
```

### Alternative: Local Oracle Setup

For Oracle testing without Docker:
- Use the GitHub Actions workflow (which skips Oracle if unavailable)
- Set up Oracle locally following instructions in `examples/oracle/README.md`

## CI/CD Integration

This docker-compose setup can be used in CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Run PostgreSQL tests
  run: |
    cd examples
    docker-compose up -d postgresql
    docker-compose up migrate-postgresql
    docker-compose run --rm test-postgresql
```

The setup is designed to be idempotent and can be run multiple times safely.
