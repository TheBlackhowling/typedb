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

// Usage
func main() {
    ctx := context.Background()
    db, _ := typedb.Open("postgres", "postgres://...")

    // Load by primary key (using Load method on model)
    user := &User{ID: 123}
    err := user.Load(ctx, db)

    // Load by unique field (using LoadByField helper)
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

- `model.Load(ctx, exec)` - Loads model by primary key field (method on Model)
- `LoadByField(ctx, exec, model, fieldName)` - Loads model by unique field
- `LoadByComposite(ctx, exec, model, compositeName)` - Loads model by composite key

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
