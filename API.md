# typedb API Reference

Complete API reference for the typedb package. For usage examples and tutorials, see [README.md](README.md).

## Table of Contents

- [Query Functions](#query-functions)
- [Load Functions](#load-functions)
- [Insert Functions](#insert-functions)
- [Update Functions](#update-functions)
- [Connection Management](#connection-management)
- [Configuration Options](#configuration-options)
- [Logging](#logging)
- [Struct Tags](#struct-tags)
- [Types & Interfaces](#types--interfaces)
- [Registration & Validation](#registration--validation)
- [Serialization Helpers](#serialization-helpers)
- [Errors](#errors)

---

## Query Functions

Query data from your database with type-safe generics.

### QueryAll

```go
func QueryAll[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) ([]*T, error)
```

Returns all rows matching the query as a slice of model pointers. Returns an empty slice if no results are found.

**Example Model:**
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

func (u *User) QueryByID() string {
    return "SELECT id, name, email FROM users WHERE id = $1"
}
```

**Example Usage:**
```go
users, err := typedb.QueryAll[*User](ctx, db, "SELECT id, name, email FROM users WHERE active = $1", true)
// users is []*User, empty slice if no results
```

### QueryFirst

```go
func QueryFirst[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) (*T, error)
```

Returns the first row matching the query. Returns `nil` if no results are found (no error).

**Example Model:**
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

func (u *User) QueryByID() string {
    return "SELECT id, name, email FROM users WHERE id = $1"
}
```

**Example Usage:**
```go
user := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email FROM users WHERE email = $1", "john@example.com")
// user is *User or nil if not found (no error)
if user != nil {
    fmt.Printf("Found user: %s\n", user.Name)
}
```

### QueryOne

```go
func QueryOne[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) (T, error)
```

Returns exactly one row matching the query. Returns `ErrNotFound` if no results are found, or an error if multiple rows are returned.

**Example Model:**
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

func (u *User) QueryByID() string {
    return "SELECT id, name, email FROM users WHERE id = $1"
}
```

**Example Usage:**
```go
user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email FROM users WHERE id = $1", 123)
if err != nil {
    if err == typedb.ErrNotFound {
        // No user found
    } else {
        // Other error (e.g., multiple rows)
    }
}
```

---

## Load Functions

Load models by primary key, unique field, or composite key. These functions automatically call the appropriate `QueryBy*` method on your model.

### Load

```go
func Load[T ModelInterface](ctx context.Context, exec Executor, model T) error
```

Loads a model by its primary key field. The model must have a field with `load:"primary"` tag, and the primary key field must be set (non-zero).

**Requirements:**
- Model must have a field with `load:"primary"` tag
- Model must have a `QueryBy{PrimaryField}()` method
- Primary key field must be set (non-zero value)

**Example Model:**
```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email"`
    CreatedAt string `db:"created_at"`
    UpdatedAt string `db:"updated_at"`
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) QueryByID() string {
    return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
}
```

**Example Usage:**
```go
user := &User{ID: 123}
err := typedb.Load(ctx, db, user)
// user is now fully populated with data from database:
// user.Name, user.Email, user.CreatedAt, user.UpdatedAt are all set
```

### LoadByField

```go
func LoadByField[T ModelInterface](ctx context.Context, exec Executor, model T, fieldName string) error
```

Loads a model by any field with a `load` tag. The specified field must be set (non-zero).

**Requirements:**
- Model must have a `QueryBy{FieldName}()` method
- The specified field must be set (non-zero value)

**Example Model:**
```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email" load:"unique"`
    CreatedAt string `db:"created_at"`
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) QueryByID() string {
    return "SELECT id, name, email, created_at FROM users WHERE id = $1"
}

func (u *User) QueryByEmail() string {
    return "SELECT id, name, email, created_at FROM users WHERE email = $1"
}
```

**Example Usage:**
```go
user := &User{Email: "john@example.com"}
err := typedb.LoadByField(ctx, db, user, "Email")
// user is now fully populated with data from database
```

### LoadByComposite

```go
func LoadByComposite[T ModelInterface](ctx context.Context, exec Executor, model T, compositeName string) error
```

Loads a model by a composite key. All fields with `load:"composite:name"` tag must be set (non-zero).

**Requirements:**
- Model must have at least 2 fields with `load:"composite:name"` tag
- Model must have a `QueryBy{Field1}{Field2}...()` method (fields sorted alphabetically)
- All composite key fields must be set (non-zero values)

**Example Model:**
```go
type UserPost struct {
    typedb.Model
    UserID    int    `db:"user_id" load:"composite:userpost"`
    PostID    int    `db:"post_id" load:"composite:userpost"`
    Role      string `db:"role"`
    CreatedAt string `db:"created_at"`
}

func (up *UserPost) TableName() string {
    return "user_posts"
}

// Method name uses fields sorted alphabetically: PostID, UserID
func (up *UserPost) QueryByPostIDUserID() string {
    return "SELECT user_id, post_id, role, created_at FROM user_posts WHERE post_id = $1 AND user_id = $2"
}
```

**Example Usage:**
```go
userPost := &UserPost{UserID: 1, PostID: 2}
err := typedb.LoadByComposite(ctx, db, userPost, "userpost")
// userPost is now fully populated with data from database
```

---

## Insert Functions

Insert data with multiple options, from raw SQL to automatic query generation.

### Insert

```go
func Insert[T ModelInterface](ctx context.Context, exec Executor, model T) error
```

Automatically builds and executes an INSERT query from model struct fields. Sets the primary key on the model after insertion.

**Requirements:**
- Model must have `TableName()` method
- Model must have a field with `load:"primary"` tag
- Model must not have dot notation in `db` tags (single-table models only)

**Behavior:**
- Excludes primary key field from INSERT
- Excludes fields with `db:"-"` tag
- Excludes fields with `dbInsert:"false"` tag
- Excludes nil/zero value fields
- Automatically uses `RETURNING`/`OUTPUT` clause or `LastInsertId()` based on database

**Example Model:**
```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email"`
    CreatedAt string `db:"created_at" dbInsert:"false"` // Excluded from INSERT
    UpdatedAt string `db:"updated_at" dbInsert:"false"` // Excluded from INSERT
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) QueryByID() string {
    return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
}
```

**Example Usage:**
```go
user := &User{Name: "John", Email: "john@example.com"}
err := typedb.Insert(ctx, db, user)
// user.ID is now set with the inserted ID
// user.CreatedAt is NOT populated (still zero value)
// user.UpdatedAt is NOT populated (still zero value)
```

**Note:** `Insert()` only sets the primary key. Database-generated fields like `created_at` (set via `DEFAULT CURRENT_TIMESTAMP`) are not populated. Use `InsertAndLoad()` if you need the full object with all database-generated fields.

### InsertAndLoad

```go
func InsertAndLoad[T ModelInterface](ctx context.Context, exec Executor, model T) (T, error)
```

Inserts a model and then loads the full object from the database. This is a convenience function that combines `Insert()` and `Load()`. Use this when you need database-generated fields (like `created_at`, `updated_at`) to be populated.

**Example Model:**
```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email"`
    CreatedAt string `db:"created_at" dbInsert:"false"` // Database sets via DEFAULT CURRENT_TIMESTAMP
    UpdatedAt string `db:"updated_at" dbInsert:"false"` // Database sets via DEFAULT CURRENT_TIMESTAMP
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) QueryByID() string {
    return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
}
```

**Example Usage:**
```go
user := &User{Name: "John", Email: "john@example.com"}
returnedUser, err := typedb.InsertAndLoad[*User](ctx, db, user)
// returnedUser.ID is set with the inserted ID
// returnedUser.CreatedAt is populated from database (e.g., "2026-01-25 10:30:00")
// returnedUser.UpdatedAt is populated from database (e.g., "2026-01-25 10:30:00")
// All fields are now fully populated
```

**When to Use:**
- When you need database-generated fields (timestamps, auto-increment values, computed columns)
- When you want the complete object immediately after insertion
- When you need to return the full object to an API client

**When NOT to Use:**
- When you only need the ID (use `Insert()` instead)
- When performance is critical and you don't need the full object (use `Insert()` instead)

### InsertAndGetId

```go
func InsertAndGetId(ctx context.Context, exec Executor, insertQuery string, args ...any) (int64, error)
```

Executes a raw INSERT query and returns the inserted ID as `int64`. Works with all supported databases.

**Database Support:**
- **PostgreSQL/SQLite/SQL Server/Oracle**: Requires `RETURNING id` or `OUTPUT INSERTED.id` clause
- **MySQL**: Uses `LastInsertId()` (no RETURNING needed)

**Example:**
```go
// PostgreSQL/SQLite/SQL Server/Oracle
id, err := typedb.InsertAndGetId(ctx, db,
    "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
    "John", "john@example.com")

// MySQL
id, err := typedb.InsertAndGetId(ctx, db,
    "INSERT INTO users (name, email) VALUES (?, ?)",
    "John", "john@example.com")
```

---

## Update Functions

Update models with automatic query generation.

### Update

```go
func Update[T ModelInterface](ctx context.Context, exec Executor, model T) error
```

Automatically builds and executes an UPDATE query from model struct fields.

**Requirements:**
- Model must have `TableName()` method
- Model must have a field with `load:"primary"` tag (must be set/non-zero)
- Model must not have dot notation in `db` tags (single-table models only)
- At least one non-zero field to update (besides primary key)

**Behavior:**
- Excludes primary key field from SET clause (used in WHERE clause)
- Excludes fields with `db:"-"` tag
- Excludes fields with `dbUpdate:"false"` tag
- Excludes nil/zero value fields
- Fields with `dbUpdate:"auto-timestamp"` are automatically populated with database timestamp functions

**Partial Update:**
When a model is registered with `RegisterModelWithOptions(ModelOptions{PartialUpdate: true})`, `Update()` will only update fields that have changed since the model was last loaded from the database.

**Example:**
```go
user := &User{ID: 123, Name: "John Updated", Email: "john.updated@example.com"}
err := typedb.Update(ctx, db, user)
// Generates: UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3
```

---

## Connection Management

### Opening a Database Connection

#### Open

```go
func Open(driverName, dsn string, opts ...Option) (*DB, error)
```

Opens a database connection with model validation. Validates all registered models before returning.

**Example:**
```go
db, err := typedb.Open("postgres", "postgres://user:pass@localhost/dbname")
```

#### OpenWithoutValidation

```go
func OpenWithoutValidation(driverName, dsn string, opts ...Option) (*DB, error)
```

Opens a database connection without model validation. Use when you want to defer validation or when models aren't registered yet.

**Example:**
```go
db, err := typedb.OpenWithoutValidation("postgres", "postgres://user:pass@localhost/dbname")
```

#### NewDB

```go
func NewDB(db *sql.DB, driverName string, timeout time.Duration) *DB
```

Creates a `DB` instance from an existing `*sql.DB` connection.

**Example:**
```go
sqlDB, _ := sql.Open("postgres", dsn)
typedbDB := typedb.NewDB(sqlDB, "postgres", 5*time.Second)
```

### DB Methods

The `DB` type implements the `Executor` interface and provides the following methods:

#### Exec

```go
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
```

Executes a query that doesn't return rows (INSERT/UPDATE/DELETE/DDL).

#### QueryAll

```go
func (d *DB) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error)
```

Returns all rows as `[]map[string]any`.

#### QueryRowMap

```go
func (d *DB) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error)
```

Returns the first row as `map[string]any`. Returns `ErrNotFound` if no rows are found.

#### GetInto

```go
func (d *DB) GetInto(ctx context.Context, query string, args []any, dest ...any) error
```

Scans a single row into destination pointers. Returns `ErrNotFound` if no rows are found.

#### QueryDo

```go
func (d *DB) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error
```

Executes a query and calls the scan function for each row (streaming).

#### Begin

```go
func (d *DB) Begin(ctx context.Context) (*Tx, error)
```

Starts a new transaction.

#### WithTx

```go
func (d *DB) WithTx(ctx context.Context, fn func(*Tx) error) error
```

Executes a function within a transaction. Automatically commits on success or rolls back on error.

**Example:**
```go
err := db.WithTx(ctx, func(tx *typedb.Tx) error {
    // Perform multiple operations
    err := typedb.Insert(ctx, tx, user)
    if err != nil {
        return err // Transaction will be rolled back
    }
    return typedb.Update(ctx, tx, post) // Transaction will be committed if successful
})
```

#### Close

```go
func (d *DB) Close() error
```

Closes the database connection.

#### Ping

```go
func (d *DB) Ping(ctx context.Context) error
```

Verifies the database connection.

### Transaction Methods

The `Tx` type implements the `Executor` interface and provides the same query methods as `DB`, plus:

#### Commit

```go
func (t *Tx) Commit() error
```

Commits the transaction.

#### Rollback

```go
func (t *Tx) Rollback() error
```

Rolls back the transaction.

---

## Configuration Options

Configuration options are passed to `Open()` or `OpenWithoutValidation()`.

### Connection Pool Options

#### WithMaxOpenConns

```go
func WithMaxOpenConns(n int) Option
```

Sets the maximum number of open connections to the database. Default: 10.

#### WithMaxIdleConns

```go
func WithMaxIdleConns(n int) Option
```

Sets the maximum number of idle connections. Default: 5.

#### WithConnMaxLifetime

```go
func WithConnMaxLifetime(d time.Duration) Option
```

Sets the maximum amount of time a connection may be reused. Default: 30 minutes.

#### WithConnMaxIdleTime

```go
func WithConnMaxIdleTime(d time.Duration) Option
```

Sets the maximum amount of time a connection may be idle. Default: 5 minutes.

#### WithTimeout

```go
func WithTimeout(d time.Duration) Option
```

Sets the default timeout for database operations. Default: 5 seconds.

### Logging Options

#### WithLogger

```go
func WithLogger(logger Logger) Option
```

Sets a custom logger for the database connection.

#### WithLogQueries

```go
func WithLogQueries(enabled bool) Option
```

Enables or disables SQL query logging. Default: `true`.

#### WithLogArgs

```go
func WithLogArgs(enabled bool) Option
```

Enables or disables query argument logging. Default: `true`.

**Example:**
```go
db, err := typedb.Open("postgres", dsn,
    typedb.WithLogQueries(false),  // Disable query logging
    typedb.WithLogArgs(false),      // Disable argument logging
)
```

---

## Logging

### Logger Interface

```go
type Logger interface {
    Debug(msg string, keyvals ...any)
    Info(msg string, keyvals ...any)
    Warn(msg string, keyvals ...any)
    Error(msg string, keyvals ...any)
}
```

### Global Logger

#### SetLogger

```go
func SetLogger(logger Logger)
```

Sets the global logger used by all DB instances (unless overridden with `WithLogger`).

#### GetLogger

```go
func GetLogger() Logger
```

Gets the current global logger.

### Context-Based Logging Overrides

These functions allow you to disable logging for specific operations:

#### WithNoLogging

```go
func WithNoLogging(ctx context.Context) context.Context
```

Disables all logging (queries and arguments) for the specific operation. Overrides global `LogQueries` and `LogArgs` settings.

**Example:**
```go
ctx = typedb.WithNoLogging(ctx)
result, err := db.Exec(ctx, "INSERT INTO users ...", args...)
```

#### WithNoQueryLogging

```go
func WithNoQueryLogging(ctx context.Context) context.Context
```

Disables query logging only for the specific operation. Arguments will still be logged if `LogArgs` is enabled.

**Example:**
```go
ctx = typedb.WithNoQueryLogging(ctx)
result, err := db.Exec(ctx, "INSERT INTO users ...", args...)
```

#### WithNoArgLogging

```go
func WithNoArgLogging(ctx context.Context) context.Context
```

Disables argument logging only for the specific operation. Queries will still be logged if `LogQueries` is enabled.

**Example:**
```go
ctx = typedb.WithNoArgLogging(ctx)
result, err := db.Exec(ctx, "INSERT INTO users ...", args...)
```

---

## Struct Tags

### Database Tags

#### `db:"column_name"`

Maps a struct field to a database column name.

```go
type User struct {
    Name string `db:"name"`
}
```

#### `db:"-"`

Excludes a field from all database operations (INSERT, UPDATE, SELECT).

```go
type User struct {
    Password string `db:"-"` // Never stored in database
}
```

#### `dbInsert:"false"`

Excludes a field from INSERT operations. The field can still be used in UPDATE and SELECT.

```go
type User struct {
    CreatedAt string `db:"created_at" dbInsert:"false"` // Excluded from INSERT
}
```

#### `dbUpdate:"false"`

Excludes a field from UPDATE operations. The field can still be used in INSERT and SELECT.

```go
type User struct {
    CreatedAt string `db:"created_at" dbUpdate:"false"` // Excluded from UPDATE
}
```

#### `dbUpdate:"auto-timestamp"`

Automatically populates a field with a database timestamp function in UPDATE operations. The appropriate function is used based on the driver:
- PostgreSQL/SQLite/Oracle: `CURRENT_TIMESTAMP`
- MySQL: `NOW()`
- SQL Server: `GETDATE()`

```go
type User struct {
    UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
}
```

### Load Tags

#### `load:"primary"`

Marks a field as the primary key. Required for `Load()`, `Insert()`, and `Update()` operations.

```go
type User struct {
    ID int `db:"id" load:"primary"`
}
```

#### `load:"unique"`

Marks a field as unique. Used for `LoadByField()` operations.

```go
type User struct {
    Email string `db:"email" load:"unique"`
}
```

#### `load:"composite:name"`

Marks a field as part of a composite key. Used for `LoadByComposite()` operations. Multiple fields can share the same composite name.

```go
type UserPost struct {
    UserID int `db:"user_id" load:"composite:userpost"`
    PostID int `db:"post_id" load:"composite:userpost"`
}
```

### Logging Tags

#### `nolog:"true"`

Masks a field value in logs (appears as `[REDACTED]`). Works automatically for `Insert()`, `Update()`, and `Load()` operations.

```go
type User struct {
    Password string `db:"password" nolog:"true"` // Masked in logs
}
```

**Example:**
```go
user := &User{Password: "secret123"}
err := typedb.Insert(ctx, db, user)
// Logs will show: args: ["John", "[REDACTED]", "john@example.com"]
```

---

## Types & Interfaces

### Executor

```go
type Executor interface {
    Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error)
    QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error)
    GetInto(ctx context.Context, query string, args []any, dest ...any) error
    QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error
}
```

Interface for executing database queries. Both `DB` and `Tx` implement this interface.

### DB

```go
type DB struct {
    // ... unexported fields
}
```

Database connection wrapper. Provides query execution with timeout handling and logging.

### Tx

```go
type Tx struct {
    // ... unexported fields
}
```

Transaction wrapper. Provides transaction-scoped query execution.

### Config

```go
type Config struct {
    DSN             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
    OpTimeout       time.Duration
    Logger          Logger
    LogQueries      bool
    LogArgs         bool
}
```

Database connection and pool configuration.

### Model

```go
type Model struct {
    // originalCopy stores a deep copy of the model after deserialization
    // Used for partial update tracking when enabled
}
```

Base struct that models should embed. Provides common functionality for model types.

### ModelInterface

```go
type ModelInterface interface {
    deserialize(row map[string]any) error
}
```

Interface that models must implement. Models satisfy this interface by embedding `Model`.

### Option

```go
type Option func(*Config)
```

Function type for configuring DB connection settings. Used with `Open()` and `OpenWithoutValidation()`.

---

## Registration & Validation

### RegisterModel

```go
func RegisterModel[T ModelInterface]()
```

Registers a model type for validation. Models must be registered before they can be validated.

**Example:**
```go
func init() {
    typedb.RegisterModel[*User]()
}
```

### RegisterModelWithOptions

```go
func RegisterModelWithOptions[T ModelInterface](opts ModelOptions) Option
```

Registers a model type with additional options.

**ModelOptions:**
```go
type ModelOptions struct {
    PartialUpdate bool // Enable partial update tracking
}
```

**Example:**
```go
func init() {
    typedb.RegisterModelWithOptions[*User](typedb.ModelOptions{
        PartialUpdate: true,
    })
}
```

### ValidateModel

```go
func ValidateModel[T ModelInterface]() error
```

Validates a single model type. Checks that required methods exist and are correctly implemented.

### ValidateAllRegistered

```go
func ValidateAllRegistered() error
```

Validates all registered models. Returns an error if any model fails validation.

### MustValidateAllRegistered

```go
func MustValidateAllRegistered()
```

Validates all registered models and panics if any model fails validation. Useful for catching configuration errors at startup.

---

## Serialization Helpers

PostgreSQL-specific helper functions for serializing Go values to database-compatible formats.

### Serialize

```go
func Serialize(value any) (string, error)
```

Generic serialization function. Converts Go values to strings suitable for database insertion.

### SerializeJSONB

```go
func SerializeJSONB(value any) (string, error)
```

Serializes a value to JSONB format for PostgreSQL.

### SerializeIntArray

```go
func SerializeIntArray(arr []int) string
```

Serializes an int slice to PostgreSQL array format.

### SerializeStringArray

```go
func SerializeStringArray(arr []string) string
```

Serializes a string slice to PostgreSQL array format.

---

## Errors

### ErrNotFound

```go
var ErrNotFound = errors.New("typedb: not found")
```

Returned when no rows are found (e.g., `QueryOne`, `QueryRowMap`, `GetInto`, `Load`).

### ErrFieldNotFound

```go
var ErrFieldNotFound = errors.New("typedb: field not found")
```

Returned when a struct field cannot be found.

### ErrMethodNotFound

```go
var ErrMethodNotFound = errors.New("typedb: method not found")
```

Returned when a required method cannot be found on a model.

### ValidationError

```go
type ValidationError struct {
    ModelName string
    FieldName string
    Message   string
}
```

Single model validation error.

### ValidationErrors

```go
type ValidationErrors []ValidationError
```

Multiple validation errors. Implements `error` interface.

---

## Model Requirements

### Required Methods

All models must implement:

1. **`TableName() string`** - Returns the database table name
2. **`QueryBy{PrimaryField}() string`** - Returns SQL query for loading by primary key

### Optional Methods

- **`QueryBy{Field}() string`** - Returns SQL query for loading by unique field (for `LoadByField`)
- **`QueryBy{Field1}{Field2}...() string`** - Returns SQL query for loading by composite key (for `LoadByComposite`)

### Required Struct Tags

- At least one field with `db:"column_name"` tag
- One field with `load:"primary"` tag

### Complete Example Model

Here's a complete example model demonstrating all features:

```go
type User struct {
    typedb.Model
    ID        int64  `db:"id" load:"primary"`
    Name      string `db:"name"`
    Email     string `db:"email" load:"unique"`
    Password  string `db:"password" nolog:"true"`                    // Masked in logs
    CreatedAt string `db:"created_at" dbInsert:"false" dbUpdate:"false"` // Database sets via DEFAULT
    UpdatedAt string `db:"updated_at" dbInsert:"false" dbUpdate:"auto-timestamp"` // Auto-updated
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) QueryByID() string {
    return "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
}

func (u *User) QueryByEmail() string {
    return "SELECT id, name, email, created_at, updated_at FROM users WHERE email = $1"
}

func init() {
    typedb.RegisterModel[*User]()
}
```

**Database Schema:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Usage Examples:**

```go
// Insert - only sets ID, database-generated fields are NOT populated
user := &User{Name: "John", Email: "john@example.com", Password: "secret123"}
err := typedb.Insert(ctx, db, user)
// user.ID is set (e.g., 123)
// user.CreatedAt is NOT set (still zero value "")
// user.UpdatedAt is NOT set (still zero value "")

// InsertAndLoad - gets full object with ALL database-generated fields populated
user2 := &User{Name: "Jane", Email: "jane@example.com", Password: "secret456"}
returnedUser, err := typedb.InsertAndLoad[*User](ctx, db, user2)
// returnedUser.ID is set (e.g., 124)
// returnedUser.CreatedAt is populated from database (e.g., "2026-01-25 10:30:00")
// returnedUser.UpdatedAt is populated from database (e.g., "2026-01-25 10:30:00")
// All fields are now fully populated - ready to return to API client

// Load - populate from database
user3 := &User{ID: 123}
err := typedb.Load(ctx, db, user3)
// user3 is fully populated with all fields from database:
// user3.Name, user3.Email, user3.CreatedAt, user3.UpdatedAt are all set

// Update - auto-updates UpdatedAt timestamp
user4 := &User{ID: 123, Name: "John Updated", Email: "john.updated@example.com"}
err := typedb.Update(ctx, db, user4)
// Generates: UPDATE users SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3
// UpdatedAt is automatically set by database
// CreatedAt is NOT updated (excluded via dbUpdate:"false")
```

**Key Differences:**

- **`Insert()`**: Only sets the primary key. Use when you only need the ID.
- **`InsertAndLoad()`**: Sets the primary key AND loads the full object with all database-generated fields. Use when you need timestamps, computed columns, or want to return the complete object to an API client.
- **`Load()`**: Populates an existing model from the database. Use after `Insert()` if you need the full object later.

---

## See Also

- [README.md](README.md) - Project overview, usage examples, and tutorials
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [CHANGELOG.md](CHANGELOG.md) - Version history and changes
