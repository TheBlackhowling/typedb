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
2. **Database-Agnostic** - Core package works with any `database/sql` driver without requiring database-specific code.
3. **SQL-First** - Developers write SQL, library handles type safety and deserialization.
4. **No Global State** - All operations require explicit database/client instances for testability.
5. **Minimal Abstraction** - Lightweight library focused on type safety, not a full ORM.

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
    typedb.RegisterModel[User]()
}

// Usage
func main() {
    db, _ := typedb.Open("postgres", "postgres://...")
    
    // Load by primary key
    user := &User{ID: 123}
    err := typedb.Load(user, ctx, db)
    
    // Load by unique field
    user2 := &User{Email: "test@example.com"}
    err = typedb.LoadByEmail(user2, ctx, db)
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

- `Load(model, ctx, exec)` - Loads model by primary key field
- `LoadBy{Field}(model, ctx, exec)` - Loads model by unique field

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
