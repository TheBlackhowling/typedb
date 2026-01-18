# typedb

A type-safe, generic database query library for Go that prioritizes SQL-first development with minimal abstraction.

## Features

- ✅ **Type-Safe Generic Queries** - Uses Go 1.18+ generics for compile-time type safety
- ✅ **Database-Agnostic** - Works with any `database/sql` driver (PostgreSQL, MySQL, SQLite, MSSQL, Oracle)
- ✅ **SQL-First Philosophy** - You write SQL, typedb handles type safety and deserialization
- ✅ **Flexible Deserialization** - Custom deserialization via interfaces and struct tags
- ✅ **Built-in Timeout Handling** - Automatic context timeout management
- ✅ **Transaction Support** - Seamless transaction handling with the same API
- ✅ **No Global State** - All operations require explicit database/client instances
- ✅ **Minimal Dependencies** - Only `database/sql` and standard library

## Design Principles

1. **No SQL Generation** - typedb does NOT generate SQL. You write your own SQL queries for maximum flexibility and database-specific optimizations.
2. **Database-Agnostic** - Core package works with any `database/sql` driver without requiring database-specific code. The executor layer is fully database-agnostic.
3. **SQL-First** - Developers write SQL, library handles type safety and deserialization.
4. **No Global State** - All operations require explicit database/client instances for testability.
5. **Minimal Abstraction** - Lightweight library focused on type safety, not a full ORM.

## Database Compatibility

typedb is designed to work with any database that has a `database/sql` driver. The core executor layer (`DB`, `Tx`, query methods) is fully database-agnostic and works with PostgreSQL, MySQL, SQLite, MSSQL, Oracle, and any other database with a compatible driver.

### PostgreSQL-Specific Features

Some serialization helpers are PostgreSQL-specific:

- **Array Serialization** (`SerializeIntArray`, `SerializeStringArray`) - These functions serialize Go slices to PostgreSQL array format (`{1,2,3}`). For other databases, handle arrays in your SQL queries directly.
- **JSONB** - While JSON serialization (`SerializeJSONB`) produces standard JSON strings that work across databases, the "JSONB" naming reflects PostgreSQL's JSONB type. The serialized output is standard JSON compatible with MySQL JSON, SQL Server JSON, etc.

For maximum database portability, write database-specific SQL for arrays and use standard JSON for JSON columns.

## Installation

```bash
go get github.com/TheBlackHowling/typedb
```

## Quick Start

### Basic Usage

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

### Model Load Methods

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

### Composite Keys

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

### Transactions

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

## API Overview

### Query Functions

- `QueryAll[T](ctx, exec, query, args...)` - Returns `[]*T`, empty slice if no results
- `QueryFirst[T](ctx, exec, query, args...)` - Returns `*T`, `nil` if no results
- `QueryOne[T](ctx, exec, query, args...)` - Returns `*T`, errors if not exactly one result

### Load Functions

- `Load(ctx, exec, model)` - Loads model by primary key field
- `LoadByField(ctx, exec, model, fieldName)` - Loads model by unique field
- `LoadByComposite(ctx, exec, model, compositeName)` - Loads model by composite key

### Insert Functions

- `InsertAndReturn[T](ctx, exec, query, args...)` - Inserts with RETURNING/OUTPUT clause, returns full model
- `InsertAndGetId(ctx, exec, query, args...)` - Inserts and returns inserted ID as int64
- `Insert(ctx, exec, model)` - Inserts model by object, automatically builds INSERT query from struct fields

### Update Functions

- `Update(ctx, exec, model)` - Updates model by object, automatically builds UPDATE query from struct fields

#### Insert with RETURNING Clause

For databases that support `RETURNING` (PostgreSQL, SQLite) or `OUTPUT` (SQL Server), use `InsertAndReturn` to get the full model back:

```go
// PostgreSQL/SQLite
user, err := typedb.InsertAndReturn[*User](ctx, db,
    "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at",
    "John", "john@example.com")

// SQL Server
user, err := typedb.InsertAndReturn[*User](ctx, db,
    "INSERT INTO users (name, email) OUTPUT INSERTED.id, INSERTED.name, INSERTED.email, INSERTED.created_at VALUES (@p1, @p2)",
    "John", "john@example.com")
```

#### Insert and Get ID

For a simpler API that just returns the inserted ID:

```go
// PostgreSQL/SQLite/SQL Server/Oracle (with RETURNING/OUTPUT)
id, err := typedb.InsertAndGetId(ctx, db,
    "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
    "John", "john@example.com")

// MySQL (uses LastInsertId)
id, err := typedb.InsertAndGetId(ctx, db,
    "INSERT INTO users (name, email) VALUES (?, ?)",
    "John", "john@example.com")
```

#### Insert by Object

Automatically build INSERT queries from your model struct. Requires:
- `TableName()` method on the model
- A field with `load:"primary"` tag
- Model must not have dot notation in db tags (single-table models only)

```go
type User struct {
    typedb.Model
    ID    int64  `db:"id" load:"primary"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

func (u *User) TableName() string {
    return "users"
}

// Note: Models that embed typedb.Model automatically get Deserialize() functionality
// You only need to override Deserialize() if you have custom deserialization logic

// Usage - automatically builds INSERT query
user := &User{Name: "John", Email: "john@example.com"}
err := typedb.Insert(ctx, db, user)
// user.ID is now set with the inserted ID

// Zero/nil fields are automatically excluded
user2 := &User{Name: "Jane"} // Email is empty, will be skipped
err = typedb.Insert(ctx, db, user2)
```

**Database Support:**
- **PostgreSQL/SQLite**: Uses `RETURNING` clause
- **SQL Server/MSSQL**: Uses `OUTPUT INSERTED.id` clause
- **Oracle**: Uses `RETURNING` clause (handled via InsertAndReturn)
- **MySQL**: Uses `LastInsertId()` (no RETURNING support)
- **Unknown drivers**: Defaults to PostgreSQL-style `RETURNING`

#### Update by Object

Automatically build UPDATE queries from your model struct. Requires:
- `TableName()` method on the model
- A field with `load:"primary"` tag (must be set/non-zero)
- Model must not have dot notation in db tags (single-table models only)
- At least one non-zero field to update (besides primary key)

**Auto-Updated Timestamp Fields:**

Fields with `dbUpdate:"auto-timestamp"` tag are automatically populated with database timestamp functions (e.g., `CURRENT_TIMESTAMP`, `NOW()`, `GETDATE()`) and do not need to be set in the model. The appropriate database function is used based on the driver:

- PostgreSQL/SQLite/Oracle: `CURRENT_TIMESTAMP`
- MySQL: `NOW()`
- SQL Server: `GETDATE()`

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

**Field Exclusion Tags:**
- `db:"-"` - Excludes field from all database operations (INSERT, UPDATE, SELECT)
- `dbInsert:"false"` - Excludes field from INSERT operations only
- `dbUpdate:"false"` - Excludes field from UPDATE operations only
- `dbUpdate:"auto"` - Automatically populates field with database timestamp function (e.g., `CURRENT_TIMESTAMP`, `NOW()`, `GETDATE()`) during UPDATE
- Fields with `dbUpdate:"false"` can still be read via SELECT queries
- Fields with `dbUpdate:"auto-timestamp"` are automatically included in UPDATE queries using database functions, even if not set in the model

**Database Support:**
- All supported databases (PostgreSQL, MySQL, SQLite, SQL Server, Oracle)
- Uses database-specific identifier quoting and parameter placeholders

### Connection Management

- `Open(driverName, dsn, opts...)` - Opens database connection with validation
- `NewDB(db *sql.DB, timeout)` - Creates DB instance from existing connection

## Requirements

- Go 1.18 or later (for generics support)
- Any `database/sql` driver

## Documentation

- [Design Draft](docs/backlog/typedb-design-draft.md) - Complete design documentation
- [Complex Models](docs/backlog/typedb-complex-models-design.md) - Multi-table models and JOINs
- [Loader Pattern](docs/backlog/typedb-loader-pattern-discussion.md) - Model loading patterns

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Status

🚧 **Early Development** - This project is in active development. API may change before v1.0.0.

## Why typedb?

The Go ecosystem has many database libraries, but typedb fills a specific gap:

- **sqlx** - Great, but doesn't use generics
- **pgx** - PostgreSQL-specific, not database-agnostic
- **GORM/Bun** - Full ORMs with heavy abstractions
- **typedb** - Type-safe generics, SQL-first, database-agnostic, minimal abstraction

If you want type-safe queries without ORM overhead, typedb is for you.
