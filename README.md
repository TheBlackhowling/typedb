# typedb

![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)
![Status](https://img.shields.io/badge/status-active-success.svg?style=flat-square)

**Type-safe SQL queries for Go without the ORM overhead.**

A type-safe, generic database query library for Go that prioritizes SQL-first development with minimal abstraction.

## Table of Contents

- [What is typedb?](#what-is-typedb)
- [Why typedb Instead of an ORM?](#why-typedb-instead-of-an-orm)
  - [When to Use typedb](#when-to-use-typedb)
  - [Real-World Use Cases](#real-world-use-cases)
  - [When NOT to Use typedb](#when-not-to-use-typedb)
- [Features](#features)
- [Design Principles](#design-principles)
- [Database Compatibility](#database-compatibility)
- [Installation](#installation)
- [Getting Started](#getting-started)
  - [Requirements](#requirements)
  - [Quick Start](#quick-start)
  - [Examples](#examples)
- [API Reference](API.md) - Complete API documentation
- [API Overview](#api-overview)
  - [Query Functions](#query-functions)
  - [Load Functions](#load-functions)
  - [Insert Functions](#insert-functions)
  - [Update Functions](#update-functions)
- [Performance](#performance)
- [Testing](#testing)
- [Feature Requests & Suggestions](#feature-requests--suggestions)
- [License](#license)
- [Status](#status)

## What is typedb?

**typedb is NOT an ORM**, but it provides some ORM-like convenience features while maintaining SQL-first development. typedb is a lightweight library that adds type safety and convenient deserialization to your SQL queries without hiding SQL behind abstractions.

## Why typedb Instead of an ORM?

The Go ecosystem has many database libraries. Here's why you might choose typedb over a full ORM:

### typedb vs ORMs (GORM, Bun, Ent, etc.)

**Advantages of typedb:**
- ✅ **SQL Transparency** - You write SQL, so you know exactly what queries are executed
- ✅ **No Query Generation Surprises** - No hidden N+1 queries or unexpected JOINs
- ✅ **Database-Specific Features** - Use PostgreSQL arrays, MySQL JSON functions, SQL Server window functions, etc. without ORM limitations
- ✅ **Performance Control** - Optimize queries yourself rather than fighting with ORM query builders
- ✅ **Minimal Learning Curve** - If you know SQL, you can use typedb immediately
- ✅ **Lightweight** - Small dependency footprint, no code generation, no migrations framework
- ✅ **Testing Flexibility** - Test ORM-based applications without coupling to the ORM layer

**When to Choose an ORM Instead:**
- You need automatic migrations and schema management
- You want relationship management (has-many, belongs-to, etc.) handled automatically
- You prefer query builders over writing SQL
- Your team is more comfortable with ORM abstractions

### typedb vs Other Libraries

| Feature | typedb | sqlx | pgx | database/sql | ORMs |
|---------|--------|------|-----|--------------|------|
| Type Safety (Generics) | ✅ | ❌ | ✅ | ❌ | ✅ |
| SQL-First | ✅ | ✅ | ✅ | ✅ | ❌ |
| Database-Agnostic | ✅ | ✅ | ❌ | ✅ | ✅ |
| Minimal Abstraction | ✅ | ✅ | ✅ | ✅ | ❌ |
| No Code Generation | ✅ | ✅ | ✅ | ✅ | ❌* |
| Relationship Management | ❌ | ❌ | ❌ | ❌ | ✅ |
| Migrations | ❌ | ❌ | ❌ | ❌ | ✅ |

\* Some ORMs require code generation

**If you want type-safe queries without ORM overhead, typedb is for you.**

### When to Use typedb

**typedb works well for production use cases** where you want:

- **Type Safety Without ORM Overhead** - Get compile-time type safety with Go generics without the complexity and abstraction layers of a full ORM
- **SQL-First Development** - Write and control your own SQL queries for maximum flexibility and database-specific optimizations
- **Minimal Abstraction** - Stay close to SQL while getting convenient deserialization and type safety
- **Database Portability** - Work with any `database/sql` driver (PostgreSQL, MySQL, SQLite, MSSQL, Oracle) without vendor lock-in
- **Performance Control** - Understand exactly what SQL is being executed without ORM query generation surprises
- **API Response Design** - typedb makes it easy to plan and design API responses with models that map 1:1 with website views. Write SQL queries with easy-to-deserialize interfaces that return exactly the data structure your frontend needs, creating a clean separation between your database schema and your API contract.
- **Testing ORM Applications** - When testing applications that use ORMs, typedb provides a lightweight alternative without building a second ORM application. This is particularly useful for integration tests where you need to verify database state without coupling to the application's ORM layer.

### Real-World Use Cases

typedb shines in these common scenarios:

**1. API Development**
- Models map 1:1 to API responses - no DTO layer needed
- Write SQL queries that return exactly what your frontend needs
- Clean separation between database schema and API contract

**2. Microservices & Web Applications**
- Most applications handle 1K-100K requests/day where database latency dominates
- typedb overhead (~0.1-0.5ms) is negligible compared to database queries (10-50ms)
- Developer productivity gains outweigh minimal performance cost

**3. Testing ORM-Based Applications**
- Verify database state independently without coupling to ORM layer
- Write integration tests that work regardless of which ORM the app uses
- Test complex SQL queries generated by ORMs

**4. Rapid Prototyping & MVPs**
- Get to market faster with less boilerplate code
- Type safety catches errors at compile time
- Easy to refactor later if needed

**5. Teams That Want SQL Control**
- Write SQL yourself for maximum flexibility
- Use database-specific features (PostgreSQL arrays, MySQL JSON functions, etc.)
- Avoid ORM query generation surprises

**6. Internal Tools & Admin Dashboards**
- Correctness and maintainability matter more than microsecond performance
- Less code means fewer bugs
- Easy to understand and modify

### When NOT to Use typedb

**Feature Requirements:**
- You need automatic SQL generation for complex queries
- You want full ORM features like migrations, relationships, and query builders
- You prefer maximum abstraction over SQL control

**Performance Considerations:**
- **Extreme Scale** - Applications handling 100K+ queries/second where every microsecond counts
- **Ultra-Low Latency** - Systems requiring sub-millisecond p99 latency (<1ms)
- **CPU-Bound Workloads** - Applications where CPU is the bottleneck and reflection overhead becomes significant
- **Memory-Constrained Environments** - When partial update doubles memory usage and causes issues

**Note:** For most real-world applications (1K-100K requests/day), typedb's overhead (~0.1-0.5ms per query) is negligible compared to database latency (10-50ms). The convenience and type safety benefits typically outweigh the performance cost.

## Features

- ✅ **Type-Safe Generic Queries** - Uses Go 1.23.12+ generics for compile-time type safety
- ✅ **Database-Agnostic** - Works with any `database/sql` driver (PostgreSQL, MySQL, SQLite, MSSQL, Oracle)
- ✅ **SQL-First Philosophy** - You write SQL, typedb handles type safety and deserialization
- ✅ **Flexible Deserialization** - Custom deserialization via interfaces and struct tags
- ✅ **Built-in Timeout Handling** - Automatic context timeout management
- ✅ **Transaction Support** - Seamless transaction handling with the same API
- ✅ **No Global State** - All operations require explicit database/client instances
- ✅ **Minimal Dependencies** - Only `database/sql` and standard library
- ✅ **Partial Update Support** - Track changes and update only modified fields (optional)
- ✅ **Auto-Timestamp Fields** - Automatic `updated_at` timestamp management with database-specific functions
- ✅ **Composite Key Support** - Full support for multi-column primary keys
- ✅ **Object-Based CRUD** - Insert and update by object with automatic query generation
- ✅ **Structured Logging** - Pluggable logging interface for debugging and monitoring (optional, zero overhead when disabled)

## Design Principles

1. **No SQL Generation** - typedb does NOT generate SQL. You write your own SQL queries for maximum flexibility and database-specific optimizations.
2. **Database-Agnostic** - Core package works with any `database/sql` driver without requiring database-specific code. The executor layer is fully database-agnostic.
3. **SQL-First** - Developers write SQL, library handles type safety and deserialization.
4. **No Global State** - All operations require explicit database/client instances for testability.
5. **Minimal Abstraction** - Lightweight library focused on type safety, not a full ORM.

## Database Compatibility

typedb is designed to work with any database that has a `database/sql` driver. The core executor layer (`DB`, `Tx`, query methods) is fully database-agnostic and works with PostgreSQL, MySQL, SQLite, MSSQL, Oracle, and any other database with a compatible driver.

## Installation

```bash
go get github.com/TheBlackHowling/typedb
```

## Getting Started

### Requirements

- **Go 1.23.12 or later** (minimum required for security fixes)
- **Go 1.24+ recommended** for best security and performance
- Any `database/sql` driver

**Security Note:** Go 1.23.12+ is required to address vulnerabilities GO-2025-3849 and GO-2025-3750. We test on Go 1.23.12, 1.24, and 1.25.

**Important:** While `go.mod` specifies `go 1.23`, you must use Go 1.23.12 or later (patch version) to avoid security vulnerabilities. Run `go version` to verify your Go version meets this requirement.

### Quick Start

Get started with typedb in 5 minutes. This example shows connection, query, insert, and update:

- [Basic Usage](#basic-usage)
- [Model Load Methods](#model-load-methods)
- [Composite Keys](#composite-keys)
- [Transactions](#transactions)

#### Basic Usage

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    "github.com/TheBlackHowling/typedb"
    _ "github.com/lib/pq" // PostgreSQL driver
)

type User struct {
    typedb.Model
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

func main() {
    ctx := context.Background()

    // Open database connection
    db, err := typedb.Open("postgres", "postgres://user:pass@localhost/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Query all users
    users, err := typedb.QueryAll[User](ctx, db, "SELECT id, name, email FROM users")
    if err != nil {
        log.Fatal(err)
    }
    
    for _, user := range users {
        fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
    }

    // Query single user
    user, err := typedb.QueryFirst[User](ctx, db, "SELECT id, name, email FROM users WHERE id = $1", 123)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found user: %s\n", user.Name)
}
```

#### Model Load Methods

```go
package main

import (
    "context"

    "github.com/TheBlackHowling/typedb"
)

type User struct {
    typedb.Model
    ID    int    `db:"id" load:"primary"`
    Email string `db:"email" load:"unique"`
    Name  string `db:"name"`
}

// Define query methods
func (u *User) QueryByID() string {
    return "SELECT id, name, email FROM users WHERE id = $1"
}

func (u *User) QueryByEmail() string {
    return "SELECT id, name, email FROM users WHERE email = $1"
}

// Register model for validation
func init() {
    typedb.RegisterModel[*User]()
}

// Note: Models that embed typedb.Model automatically get Deserialize() functionality
// You only need to override Deserialize() if you have custom deserialization logic

// Usage
func main() {
    ctx := context.Background()
    db, _ := typedb.Open("postgres", "postgres://...")

    // Load by primary key
    user := &User{ID: 123}
    err := typedb.Load(ctx, db, user)

    // Load by unique field
    user2 := &User{Email: "test@example.com"}
    err = typedb.LoadByField(ctx, db, user2, "Email")
}
```

#### Composite Keys

```go
package main

import (
    "context"
    "log"

    "github.com/TheBlackHowling/typedb"
)

type UserPost struct {
    typedb.Model
    UserID int `db:"user_id" load:"composite:userpost"`
    PostID int `db:"post_id" load:"composite:userpost"`
    // UserID + PostID together uniquely identify a UserPost
}

// Query method: fields sorted alphabetically (PostID, UserID)
// IMPORTANT: SQL parameters must match alphabetical field order
func (up *UserPost) QueryByPostIDUserID() string {
    // PostID comes before UserID alphabetically, so $1 = PostID, $2 = UserID
    return "SELECT user_id, post_id FROM user_posts WHERE post_id = $1 AND user_id = $2"
}

// Register model for validation
func init() {
    typedb.RegisterModel[*UserPost]()
}

// Usage - must populate all fields in composite key
func main() {
    ctx := context.Background()
    db, _ := typedb.Open("postgres", "postgres://...")

    userPost := &UserPost{UserID: 123, PostID: 456}
    err := typedb.LoadByComposite(ctx, db, userPost, "userpost")
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Transactions

```go
err := db.WithTx(ctx, func(tx *typedb.Tx) error {
    // All queries use the same transaction
    users, err := typedb.QueryAll[User](ctx, tx, "SELECT * FROM users")
    if err != nil {
        return err
    }

    // More operations...
    return nil
})
```

### Examples

Database-specific examples demonstrating typedb usage are available for all supported databases:

- **[PostgreSQL Examples](examples/postgresql/)** - Full-featured examples including arrays and JSONB
- **[MySQL Examples](examples/mysql/)** - Examples for MySQL database
- **[SQLite Examples](examples/sqlite/)** - File-based database examples
- **[SQL Server (MSSQL) Examples](examples/mssql/)** - Microsoft SQL Server examples
- **[Oracle Examples](examples/oracle/)** - Oracle Database examples

Each example directory includes:
- Complete working examples demonstrating typedb features
- Database schema and migration files
- Setup instructions and usage patterns

For comprehensive test coverage (including error cases), see the [Integration Tests](#testing) section below.

## API Overview

### Query Functions

Query data from your database with type-safe generics:

- `QueryAll[T](ctx, exec, query, args...)` - Returns `[]*T`, empty slice if no results
- `QueryFirst[T](ctx, exec, query, args...)` - Returns `*T`, `nil` if no results
- `QueryOne[T](ctx, exec, query, args...)` - Returns `*T`, errors if not exactly one result

### Load Functions

Load models by primary key, unique field, or composite key:

- `Load(ctx, exec, model)` - Loads model by primary key field
- `LoadByField(ctx, exec, model, fieldName)` - Loads model by unique field
- `LoadByComposite(ctx, exec, model, compositeName)` - Loads model by composite key

### Insert Functions

Insert data with multiple options:

- `Insert(ctx, exec, model)` - Inserts model by object, automatically builds INSERT query from struct fields, sets ID on model
- `InsertAndLoad(ctx, exec, model)` - Inserts model by object, then loads full object from database, returns fully populated model
- `InsertAndGetId(ctx, exec, query, args...)` - Inserts with raw SQL and returns inserted ID as int64 (convenience for raw SQL)

### Update Functions

Update models with automatic query generation:

- `Update(ctx, exec, model)` - Updates model by object, automatically builds UPDATE query from struct fields

#### Insert by Object

Automatically build INSERT queries from your model struct. Requires:
- `TableName()` method on the model
- A field with `load:"primary"` tag
- Model must not have dot notation in db tags (single-table models only)

```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email"`
    CreatedAt string `db:"created_at"`
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) QueryByID() string {
    return "SELECT id, name, email, created_at FROM users WHERE id = $1"
}

// Insert only - sets ID on model
user := &User{Name: "John", Email: "john@example.com"}
err := typedb.Insert(ctx, db, user)
// user.ID is now set with the inserted ID
// user.CreatedAt is NOT populated (still zero value)

// Insert and Load - returns fully populated model
user2 := &User{Name: "Jane", Email: "jane@example.com"}
returnedUser, err := typedb.InsertAndLoad[*User](ctx, db, user2)
// returnedUser.ID, returnedUser.CreatedAt, etc. are all populated

// Zero/nil fields are automatically excluded from INSERT
user3 := &User{Name: "Bob"} // Email is empty, will be skipped
err = typedb.Insert(ctx, db, user3)
```

**Database Support:**
- **PostgreSQL/SQLite**: Uses `RETURNING` clause
- **SQL Server/MSSQL**: Uses `OUTPUT INSERTED.id` clause
- **MySQL**: Uses `LastInsertId()` (no RETURNING support)
- **Oracle**: Uses `RETURNING INTO` with `sql.Out` parameter
- **Unknown drivers**: Defaults to PostgreSQL-style `RETURNING`

#### Insert and Get ID (Raw SQL)

For raw SQL queries where you only need the inserted ID:

```go
// PostgreSQL/SQLite/SQL Server/Oracle (with RETURNING/OUTPUT)
id, err := typedb.InsertAndGetId(ctx, db,
    "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
    "John", "john@example.com")

// MySQL (uses LastInsertId - no RETURNING needed)
id, err := typedb.InsertAndGetId(ctx, db,
    "INSERT INTO users (name, email) VALUES (?, ?)",
    "John", "john@example.com")
```

**Note:** For raw SQL with full model return, write your own SQL and use `QueryRowMap()` or `QueryFirst()` to deserialize the result.

#### Update by Object

Automatically build UPDATE queries from your model struct. Requires:
- `TableName()` method on the model
- A field with `load:"primary"` tag (must be set/non-zero)
- Model must not have dot notation in db tags (single-table models only)
- At least one non-zero field to update (besides primary key)

**Auto-Updated Timestamp Fields:**

Fields with `dbUpdate:"auto-timestamp"` tag are automatically populated with database timestamp functions (e.g., `CURRENT_TIMESTAMP`, `NOW()`, `GETDATE()`) and do not need to be set in the model. The appropriate database function is used based on the driver:

- PostgreSQL/SQLite/Oracle: `CURRENT_TIMESTAMP`
- MySQL: `NOW()` (uses standard `TIMESTAMP` column type)
- SQL Server: `GETDATE()`

**MySQL Timestamp Precision:**

MySQL uses `NOW()` with standard `TIMESTAMP` columns by default, which provides second-level precision. This is sufficient for most use cases. If you need microsecond precision, you can use `TIMESTAMP(6)` columns in your schema, but you'll need to manually specify `NOW(6)` in your UPDATE queries. Custom timestamp precision configuration may be added in the future if there's sufficient demand.

```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email"`
    CreatedAt string `db:"created_at" dbUpdate:"false"` // Excluded from UPDATE
    UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`  // Auto-populated with database timestamp
}

func (u *User) TableName() string {
    return "users"
}

// Usage - automatically builds UPDATE query
user := &User{ID: 123, Name: "John Updated", Email: "john.updated@example.com"}
err := typedb.Update(ctx, db, user)
// Generates: UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3
// UpdatedAt is automatically populated by the database

// Zero/nil fields are automatically excluded
user2 := &User{ID: 123, Name: "Jane Updated"} // Email is empty, will be skipped
err = typedb.Update(ctx, db, user2)
// Generates: UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
```

**Partial Update (Change Tracking):**

When partial update is enabled for a model, `Update` will only modify fields that have changed since the model was last loaded from the database. This is useful for:
- Optimizing UPDATE queries to only include changed fields
- Preventing accidental overwrites of unchanged fields
- Reducing database load by updating only what changed

To enable partial update for a model, register it with `RegisterModelWithOptions`:

```go
func init() {
    typedb.RegisterModel[*User]()
    // Enable partial update for User model
    typedb.RegisterModelWithOptions[*User](typedb.ModelOptions{PartialUpdate: true})
}
```

**How Partial Update Works:**

1. When a model is deserialized (via `Load`, `QueryFirst`, `QueryOne`, `QueryAll`), a deep copy of the model is saved internally
2. When `Update` is called, the current model state is compared with the saved copy
3. Only fields that have changed are included in the UPDATE statement
4. After a successful update, the saved copy is refreshed with the new state

**Example:**

```go
// Load user - this saves the original state internally
user := &User{ID: 123}
err := typedb.Load(ctx, db, user)
// user.Name = "John", user.Email = "john@example.com"

// Modify only the name
user.Name = "John Updated"
// user.Email remains "john@example.com"

// Update - only name will be updated, email remains unchanged
err = typedb.Update(ctx, db, user)
// Generates: UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
// Email is NOT included because it hasn't changed

// If you modify both fields, both will be updated
user.Name = "John Updated Again"
user.Email = "john.new@example.com"
err = typedb.Update(ctx, db, user)
// Generates: UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3
```

**Important Notes:**

- Partial update requires the model to be loaded from the database first (via `Load`, `QueryFirst`, `QueryOne`, or `QueryAll`) before calling `Update`
- If a model hasn't been loaded, `Update` will behave as if partial update is disabled (all non-null fields will be updated)
- **Memory Overhead**: Partial update stores a deep copy of the model internally, effectively doubling memory usage for the duration the model object is in memory. Consider this when enabling partial update for large models or high-volume scenarios
- Partial update works with all field types including strings, numbers, booleans, slices, maps, and nested structs
- The comparison uses `reflect.DeepEqual`, so be aware that JSON round-trip conversions (e.g., `int` to `float64`) may be detected as changes for map/slice fields
- Partial update is optional and disabled by default - models registered with `RegisterModel` will use full updates

**Field Exclusion Tags:**
- `db:"-"` - Excludes field from all database operations (INSERT, UPDATE, SELECT)
- `dbInsert:"false"` - Excludes field from INSERT operations only
- `dbUpdate:"false"` - Excludes field from UPDATE operations only
- `dbUpdate:"auto-timestamp"` - Automatically populates field with database timestamp function (e.g., `CURRENT_TIMESTAMP`, `NOW()`, `GETDATE()`) during UPDATE
- Fields with `dbUpdate:"false"` can still be read via SELECT queries
- Fields with `dbUpdate:"auto-timestamp"` are automatically included in UPDATE queries using database functions, even if not set in the model

**Database Support:**
- All supported databases (PostgreSQL, MySQL, SQLite, SQL Server, Oracle)
- Uses database-specific identifier quoting and parameter placeholders

### Connection Management

- `Open(driverName, dsn, opts...)` - Opens database connection with validation
- `NewDB(db *sql.DB, timeout)` - Creates DB instance from existing connection

## Logging

typedb provides a pluggable logging interface that allows you to integrate with your preferred logging library. Logging is completely optional and has zero overhead when disabled (uses a no-op logger by default).

### Logger Interface

The `Logger` interface provides four log levels:

```go
type Logger interface {
    Debug(msg string, keyvals ...any)
    Info(msg string, keyvals ...any)
    Warn(msg string, keyvals ...any)
    Error(msg string, keyvals ...any)
}
```

### Setting a Logger

**Global Logger** (applies to all DB instances):

```go
import (
    "github.com/TheBlackHowling/typedb"
    "github.com/your-logger/logger"
)

// Set global logger
typedb.SetLogger(yourLogger)

// Open database - will use global logger
db, err := typedb.Open("postgres", dsn)
```

**Per-Instance Logger** (overrides global logger for specific DB instance):

```go
import (
    "github.com/TheBlackHowling/typedb"
    "github.com/your-logger/logger"
)

// Open with specific logger
db, err := typedb.Open("postgres", dsn, 
    typedb.WithLogger(yourLogger))
```

### Log Levels

**Debug Logs** (detailed operation tracking):
- Query execution (SQL and arguments)
- Transaction begin
- Row scanning operations
- "No rows found" cases

**Info Logs** (important lifecycle events):
- Connection opened/closed
- Transaction commit/rollback
- Model validation (during `Open()`)

**Error Logs** (all failures):
- Query execution failures
- Connection failures
- Transaction failures
- Scan/deserialization errors

### Example: Integrating with a Logging Library

```go
package main

import (
    "github.com/TheBlackHowling/typedb"
    "github.com/rs/zerolog"
)

// Implement typedb.Logger interface
type ZerologAdapter struct {
    logger zerolog.Logger
}

func (z *ZerologAdapter) Debug(msg string, keyvals ...any) {
    z.logger.Debug().Fields(keyvalsToMap(keyvals)).Msg(msg)
}

func (z *ZerologAdapter) Info(msg string, keyvals ...any) {
    z.logger.Info().Fields(keyvalsToMap(keyvals)).Msg(msg)
}

func (z *ZerologAdapter) Warn(msg string, keyvals ...any) {
    z.logger.Warn().Fields(keyvalsToMap(keyvals)).Msg(msg)
}

func (z *ZerologAdapter) Error(msg string, keyvals ...any) {
    z.logger.Error().Fields(keyvalsToMap(keyvals)).Msg(msg)
}

func keyvalsToMap(keyvals []any) map[string]any {
    m := make(map[string]any)
    for i := 0; i < len(keyvals)-1; i += 2 {
        key := keyvals[i].(string)
        m[key] = keyvals[i+1]
    }
    return m
}

func main() {
    // Create logger adapter
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    adapter := &ZerologAdapter{logger: logger}
    
    // Set global logger
    typedb.SetLogger(adapter)
    
    // Use typedb - all operations will be logged
    db, err := typedb.Open("postgres", dsn)
    // ...
}
```

### Zero Overhead

When no logger is set, typedb uses a no-op logger that discards all log messages. This ensures zero performance overhead when logging is not needed:

```go
// No logger set - zero overhead
db, err := typedb.Open("postgres", dsn)
// All log calls are no-ops, no performance impact
```

## Performance

typedb prioritizes developer productivity and type safety over raw performance. Understanding the performance characteristics helps you make informed decisions.

### Performance Characteristics

**Overhead per Query (Measured on Linux, AMD EPYC 7763):**
- **Simple Structs** (5-10 fields): ~2μs per row overhead
- **Complex Structs** (15-20 fields, JSONB, arrays, nested types): ~15-18μs per row overhead
- **Partial Update Overhead**: ~150-700μs per load when enabled (JSON marshaling/unmarshaling for deep copy)
- **Memory Overhead**: Partial update doubles memory usage for loaded models

**Single Row Overhead:**
- typedb overhead: 0.002-0.018ms (simple to complex structs, without partial update)
- Negligible compared to typical database query and network latency

**Bulk Query Performance (Measured Results):**
- **1,000 rows**: ~2ms overhead (simple) or ~16ms overhead (complex)
- **10,000 rows**: ~20ms overhead (simple) or ~165ms overhead (complex)
- **100,000 rows**: ~200ms overhead (simple) or ~1.6s overhead (complex)
- **1,000,000 rows**: ~2s overhead (simple) or ~18s overhead (complex)
- **Example**: Exporting 50,000 records adds ~100ms overhead (simple) or ~800ms overhead (complex)

**Very Large Dataset Observations (1M+ rows):**
- **Simple structs (1M rows)**: Overhead remains consistent at ~2s, demonstrating linear scaling
- **Complex structs (1M rows)**: Overhead increases to ~18s due to JSONB parsing, nested struct deserialization, and array processing
- **Memory usage**: Scales linearly with row count (~624KB per 1K simple rows, ~5.6MB per 1K complex rows)
- **Performance characteristics**: Overhead is predictable and consistent across runs, making it suitable for batch processing
- **When to use**: Ideal for background jobs, ETL pipelines, and data exports where 2-18 seconds overhead is acceptable
- **When to consider alternatives**: For real-time processing of 1M+ rows or when memory constraints are a concern

**For single-row or small result sets (<100 rows):**
- typedb overhead is negligible (microseconds per row)
- The overhead is typically much smaller than database query and network latency

**For bulk operations (100+ rows):**
- Reflection overhead scales linearly but remains modest: ~2μs per row (simple) or ~15-18μs per row (complex)
- Consider using `QueryDo` for streaming large result sets to reduce memory usage
- For very large exports (100K+ rows), overhead is typically 200ms-2s depending on struct complexity

### When Bulk Queries (10K+ Rows) Are Reasonable

Even with 50K+ rows, typedb can be a good choice when:

**✅ Background Jobs & Batch Processing**
- Latency isn't critical (runs overnight, weekly reports)
- Correctness and maintainability matter more than speed
- Example: Monthly financial reports, data warehouse ETL jobs

**✅ Admin & Reporting Tools**
- Human-facing interfaces where sub-second to few seconds is acceptable
- Type safety prevents bugs that could corrupt reports
- Example: Admin dashboards, analytics exports, audit logs (100K rows: ~200ms-1.6s overhead)

**✅ One-Time Operations**
- Migrations, data exports, or one-off scripts
- Developer productivity > performance optimization
- Example: Database migrations, data exports for compliance

**✅ When Query Execution Time Dominates**
- If the overall query execution takes 30+ seconds, typedb's 200ms-2s overhead is negligible
- Complex joins, aggregations, or slow queries make reflection overhead negligible
- Example: Complex analytics queries, multi-table joins with aggregations

**✅ Type Safety Prevents Costly Bugs**
- When manual scanning errors could cause data corruption or compliance issues
- The overhead cost is less than the cost of bugs
- Example: Financial transactions, medical records, legal compliance data

**❌ Avoid typedb for bulk queries when:**
- Real-time user-facing endpoints (API responses, web pages)
- High-frequency bulk operations (100K+ queries/hour)
- Memory-constrained environments
- When every millisecond matters for user experience

### When Performance Matters

**For most applications (1K-100K requests/day):**
- ✅ typedb overhead is negligible (~0.002-0.018ms per row)
- ✅ Overhead is typically much smaller than database query and network latency
- ✅ Developer productivity gains outweigh minimal cost

**Consider alternatives if:**
- ❌ Handling 100K+ queries/second (extreme scale)
- ❌ Requiring sub-millisecond p99 latency (<1ms)
- ❌ CPU-bound workloads where reflection overhead becomes significant
- ❌ Memory-constrained environments (especially with partial update enabled)

### Performance Optimization Tips

1. **Disable Partial Update Unless Needed**
   - Saves ~150-700μs per load
   - Cuts memory usage in half
   - Only enable for models that benefit from change tracking

2. **Memory Considerations for Large Datasets**
   - Memory usage scales linearly with row count
   - For 1M+ rows, consider streaming with `QueryDo` to reduce memory footprint
   - Simple structs use ~624KB per 1K rows; complex structs use ~5.6MB per 1K rows

3. **Profile Your Application**
   - Reflection overhead varies by struct complexity
   - Measure your specific workload to identify bottlenecks

4. **Consider Alternatives for Hot Paths**
   - Use `database/sql` directly for ultra-high-performance endpoints
   - Use typedb for convenience in less critical paths

5. **Handle Bulk Queries Efficiently**
   - For queries returning 100+ rows, overhead is ~0.2-1.8ms (simple to complex)
   - Use `QueryDo` for streaming large result sets to reduce memory usage
   - For exports/reports with 100K+ rows, overhead is ~200ms-2s depending on struct complexity
   - Example: Exporting 50K records with typedb adds ~100ms-800ms overhead (simple to complex)

### Comparison with Alternatives

| Library | Query Overhead (per row) | Type Safety | Convenience |
|---------|-------------------------|-------------|-------------|
| `database/sql` | ~0.2-0.3μs | ❌ Manual scanning | Low |
| `sqlx` | ~0.3-0.5μs | ❌ Manual scanning | Medium |
| `typedb` (simple) | ~2μs | ✅ Generics | High |
| `typedb` (complex) | ~15-18μs | ✅ Generics | High |
| `typedb` (partial update) | +150-700μs per load | ✅ Generics | High |

**Bottom Line:** For most real-world applications, typedb's performance overhead is acceptable given the productivity and type safety benefits. If you're at extreme scale or have strict latency requirements, consider `database/sql` or `pgx` directly.

## Testing

typedb is comprehensively tested across all supported databases to ensure full functionality and database compatibility. Our testing approach includes:

- **Integration Tests** - Comprehensive test suites for PostgreSQL, MySQL, SQLite, SQL Server, and Oracle covering both happy paths and error cases
- **Example Programs** - Working examples for each database demonstrating real-world usage patterns
- **Cross-Database Validation** - All features are tested against each supported database to ensure consistent behavior

See the [integration_tests/](integration_tests/) directory for complete test coverage.

**Additional Database Support:** We are open to expanding our test coverage to include additional databases. If you need support for a database not currently covered, please open an issue to discuss adding it to our test matrix.

## Feature Requests & Suggestions

We welcome feature requests and suggestions! While we're not currently accepting code contributions, we'd love to hear your ideas for improving typedb.

- **Feature Requests**: Open an issue describing the feature you'd like to see
- **Suggestions**: Share your ideas for improvements, optimizations, or new capabilities
- **Bug Reports**: Report any issues you encounter

Your feedback helps shape the future of typedb!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Status

✅ **Production Ready** - typedb v1.0.0 provides a stable API for type-safe SQL queries in Go. All public APIs are stable and backward compatible.
