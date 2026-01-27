# API Stability Review for v1.0.0

This document reviews all public APIs in typedb to ensure stability for v1.0.0 release.

**Review Date:** 2026-01-27  
**Current Version:** 0.1.47  
**Target Version:** 1.0.0

---

## Public API Inventory

### Connection Management

#### Functions
- ✅ `Open(driverName, dsn string, opts ...Option) (*DB, error)` - **STABLE**
- ✅ `OpenWithoutValidation(driverName, dsn string, opts ...Option) (*DB, error)` - **STABLE**
- ✅ `NewDB(db *sql.DB, driverName string, timeout time.Duration) *DB` - **STABLE**
- ✅ `NewDBWithLogger(db *sql.DB, driverName string, timeout time.Duration, logger Logger) *DB` - **STABLE**

#### Types
- ✅ `DB` struct - **STABLE** (fields may change but methods are stable)
- ✅ `Tx` struct - **STABLE** (fields may change but methods are stable)
- ✅ `Config` struct - **STABLE**
- ✅ `Option` type (func(*Config)) - **STABLE**

#### Option Functions
- ✅ `WithMaxOpenConns(n int) Option` - **STABLE**
- ✅ `WithMaxIdleConns(n int) Option` - **STABLE**
- ✅ `WithConnMaxLifetime(d time.Duration) Option` - **STABLE**
- ✅ `WithConnMaxIdleTime(d time.Duration) Option` - **STABLE**
- ✅ `WithTimeout(d time.Duration) Option` - **STABLE**
- ✅ `WithLogger(logger Logger) Option` - **STABLE**
- ✅ `WithLogQueries(enabled bool) Option` - **STABLE**
- ✅ `WithLogArgs(enabled bool) Option` - **STABLE**

### Query Functions

#### Functions
- ✅ `QueryAll[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) ([]T, error)` - **STABLE**
- ✅ `QueryFirst[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) (T, error)` - **STABLE**
- ✅ `QueryOne[T ModelInterface](ctx context.Context, exec Executor, query string, args ...any) (T, error)` - **STABLE**

**Note:** All query functions use generics and require pointer types (e.g., `*User`). This is stable.

### Load Functions

#### Functions
- ✅ `Load[T ModelInterface](ctx context.Context, exec Executor, model T) error` - **STABLE**
- ✅ `LoadByField[T ModelInterface](ctx context.Context, exec Executor, model T, fieldName string) error` - **STABLE**
- ✅ `LoadByComposite[T ModelInterface](ctx context.Context, exec Executor, model T, fieldNames ...string) error` - **STABLE**

**Note:** Load functions require models with `load:"primary"`, `load:"unique"`, or `load:"composite"` tags and corresponding `QueryBy*()` methods. This is stable.

### Insert Functions

#### Functions
- ✅ `Insert[T ModelInterface](ctx context.Context, exec Executor, model T) error` - **STABLE**
- ✅ `InsertAndGetID(ctx context.Context, exec Executor, insertQuery string, args ...any) (int64, error)` - **STABLE**
- ✅ `InsertAndLoad[T ModelInterface](ctx context.Context, exec Executor, model T) error` - **STABLE**

**Note:** Insert functions require models with `TableName()` method. This is stable.

### Update Functions

#### Functions
- ✅ `Update[T ModelInterface](ctx context.Context, exec Executor, model T) error` - **STABLE**

**Note:** Update requires `TableName()` method and `load:"primary"` tag. Partial update behavior is controlled by `RegisterModelWithOptions`. This is stable.

### Executor Interface

#### Interface
- ✅ `Executor` interface - **STABLE**
  - `Exec(ctx context.Context, query string, args ...any) (sql.Result, error)`
  - `QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error)`
  - `QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error)`
  - `GetInto(ctx context.Context, query string, args []any, dest ...any) error`
  - `QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error`

**Note:** Both `DB` and `Tx` implement `Executor`. This interface is stable.

### DB Methods

#### Methods
- ✅ `(*DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error)` - **STABLE**
- ✅ `(*DB) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error)` - **STABLE**
- ✅ `(*DB) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error)` - **STABLE**
- ✅ `(*DB) GetInto(ctx context.Context, query string, args []any, dest ...any) error` - **STABLE**
- ✅ `(*DB) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error` - **STABLE**
- ✅ `(*DB) Close() error` - **STABLE**
- ✅ `(*DB) Ping(ctx context.Context) error` - **STABLE**
- ✅ `(*DB) Begin(ctx context.Context, opts *sql.TxOptions) (*Tx, error)` - **STABLE**
- ✅ `(*DB) WithTx(ctx context.Context, fn func(*Tx) error, opts *sql.TxOptions) error` - **STABLE**

### Tx Methods

#### Methods
- ✅ `(*Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error)` - **STABLE**
- ✅ `(*Tx) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error)` - **STABLE**
- ✅ `(*Tx) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error)` - **STABLE**
- ✅ `(*Tx) GetInto(ctx context.Context, query string, args []any, dest ...any) error` - **STABLE**
- ✅ `(*Tx) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error` - **STABLE**
- ✅ `(*Tx) Commit() error` - **STABLE**
- ✅ `(*Tx) Rollback() error` - **STABLE**

### Registration & Validation

#### Functions
- ✅ `RegisterModel[T ModelInterface]()` - **STABLE**
- ✅ `RegisterModelWithOptions[T ModelInterface](opts ModelOptions)` - **STABLE**
- ✅ `ValidateAllRegistered() error` - **STABLE**
- ✅ `MustValidateAllRegistered()` - **STABLE**

#### Types
- ✅ `ModelOptions` struct - **STABLE**
  - `PartialUpdate bool` - **STABLE**

### Model Types & Interfaces

#### Types
- ✅ `Model` struct - **STABLE** (embed this in your models)
- ✅ `ModelInterface` interface - **STABLE** (satisfied by embedding Model)

**Note:** The `Model` struct has an unexported `originalCopy` field used for partial updates. This is internal and stable.

### Errors

#### Variables
- ✅ `ErrNotFound` - **STABLE**
- ✅ `ErrFieldNotFound` - **STABLE**
- ✅ `ErrMethodNotFound` - **STABLE**

**Note:** These errors are stable. New errors may be added in the future but these will not be removed.

### Logging

#### Interface
- ✅ `Logger` interface - **STABLE**
  - `Debug(msg string, args ...any)`
  - `Info(msg string, args ...any)`
  - `Error(msg string, args ...any)`

#### Functions
- ✅ `WithMaskIndices(ctx context.Context, indices []int) context.Context` - **STABLE**

**Note:** Logger interface is stable. Implementations can be swapped via `WithLogger()` option.

---

## Struct Tags

### Database Tags

- ✅ `db:"column_name"` - **STABLE** - Maps struct field to database column
- ✅ `db:"-"` - **STABLE** - Excludes field from database operations
- ✅ `db:"table.column"` - **STABLE** - Dot notation for joined tables (read-only)

### Load Tags

- ✅ `load:"primary"` - **STABLE** - Marks primary key field
- ✅ `load:"unique"` - **STABLE** - Marks unique field
- ✅ `load:"composite"` - **STABLE** - Marks composite key field

### Insert/Update Tags

- ✅ `dbInsert:"false"` - **STABLE** - Excludes field from INSERT
- ✅ `dbUpdate:"false"` - **STABLE** - Excludes field from UPDATE
- ✅ `dbUpdate:"auto-timestamp"` - **STABLE** - Auto-populates timestamp on UPDATE

### Security Tags

- ✅ `nolog:"true"` - **STABLE** - Masks field value in logs

---

## Required Model Methods

### Required Methods (Stable Contract)

- ✅ `TableName() string` - **STABLE** - Required for Insert/Update operations
- ✅ `QueryBy{Field}() string` - **STABLE** - Required for Load operations
  - Format: `QueryBy{FieldName}()` where FieldName matches the struct field name
  - Must return a SQL query string
  - Must accept no parameters
  - Example: `QueryByID() string` for field `ID`

**Note:** These method signatures are stable. The naming convention is stable.

---

## Potential Breaking Changes (None Identified)

After reviewing all public APIs, **no breaking changes are planned for v1.0.0**.

### Areas Reviewed for Stability:

1. ✅ **Function Signatures** - All stable, no changes planned
2. ✅ **Type Definitions** - All stable, internal fields may change but public API is stable
3. ✅ **Interface Contracts** - All stable
4. ✅ **Error Values** - All stable
5. ✅ **Struct Tags** - All stable
6. ✅ **Method Requirements** - All stable

---

## Deprecation Warnings

**None.** No APIs are being deprecated for v1.0.0.

---

## Backward Compatibility

### Compatibility Guarantee

For v1.0.0, typedb guarantees:

1. ✅ All public functions, types, and interfaces will remain stable
2. ✅ Struct tag semantics will remain stable
3. ✅ Required model methods will remain stable
4. ✅ Error values will remain stable

### What May Change (Internal Only)

- Internal helper functions (unexported)
- Internal struct fields (unexported)
- Implementation details
- Performance optimizations

---

## Migration Guide

**No migration required.** Since no breaking changes are planned, existing code using typedb 0.1.x will work with 1.0.0 without changes.

---

## Recommendations for 1.0.0

1. ✅ **API Stability** - All public APIs are stable and ready for 1.0.0
2. ✅ **Documentation** - API.md is comprehensive and up-to-date
3. ✅ **Testing** - Comprehensive test coverage exists
4. ⚠️ **Examples** - Ensure all examples work with stable API
5. ⚠️ **Performance** - Document performance characteristics

---

## Sign-off

**Status:** ✅ **READY FOR 1.0.0**

All public APIs are stable and no breaking changes are planned. The API surface is well-defined and documented.
