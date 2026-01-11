# typedb Implementation Dependency Chart

This document outlines the dependency order for implementing typedb. Work through items in order, as each layer depends on the previous layers.

## Layer 0: Foundation (No Dependencies)

### 1. `errors.go`
**Purpose**: Define error types used throughout the package
**Dependencies**: None
**Exports**:
- `ErrNotFound` - returned when a query returns no rows
- Other error types as needed

### 2. `types.go` - Core Types Only
**Purpose**: Define basic types and interfaces (without implementations)
**Dependencies**: `errors.go`
**Exports**:
- `Executor` interface (method signatures only, no implementation)
- `ModelInterface` interface
- `Model` struct (empty base struct)
- `Config` struct
- `DB` struct (fields only, no methods)
- `Tx` struct (fields only, no methods)

**Note**: Keep implementations separate - this file should only have type definitions.

---

## Layer 1: Registration System

### 3. `registry.go`
**Purpose**: Track registered models for validation
**Dependencies**: `types.go`
**Exports**:
- `RegisterModel[T ModelInterface]()` - register a model type
- `GetRegisteredModels()` - get all registered models (for validation)
- Internal: model registry map/slice

**Implementation Notes**:
- Use `init()` function pattern - models call `RegisterModel` in their `init()`
- Store model types (reflect.Type) for later validation
- Thread-safe registration

---

## Layer 2: Reflection Utilities

### 4. `reflect.go` (or `internal/reflect.go`)
**Purpose**: Reflection helpers for working with models
**Dependencies**: `types.go`, `registry.go`
**Exports**:
- `FindFieldByTag(model, tagKey, tagValue)` - find struct field by tag
- `GetFieldValue(model, fieldName)` - get field value
- `SetFieldValue(model, fieldName, value)` - set field value
- `FindMethod(model, methodName)` - find method by name
- `CallMethod(model, methodName, args)` - call method via reflection
- `GetModelType(model)` - get reflect.Type of model

**Implementation Notes**:
- Handle embedded structs (Model base)
- Handle pointer vs value types
- Error handling for missing fields/methods

---

## Layer 3: Deserialization

### 5. `deserialize.go`
**Purpose**: Convert database rows (map[string]any) into model structs
**Dependencies**: `types.go`, `reflect.go`
**Exports**:
- `DeserializeForType[T ModelInterface](row map[string]any) T` - deserialize row into model
- `Deserialize(row map[string]any, dest ModelInterface) error` - deserialize into existing model
- Internal: type conversion helpers (string→int, etc.)

**Implementation Notes**:
- Handle `db` tags for field mapping
- Support dot notation in tags (`users.id`)
- Type conversion (database types → Go types)
- Handle pointer fields, slices, etc.

---

## Layer 4: Executor Implementation

### 6. `executor.go` - Core Methods
**Purpose**: Implement Executor interface for DB and Tx
**Dependencies**: `types.go`, `errors.go`
**Exports**:
- `NewDB(*sql.DB, time.Duration) *DB` - create DB wrapper
- `DB.Exec()` - implement Executor.Exec
- `DB.QueryAll()` - implement Executor.QueryAll
- `DB.QueryRowMap()` - implement Executor.QueryRowMap
- `DB.GetInto()` - implement Executor.GetInto
- `DB.QueryDo()` - implement Executor.QueryDo
- `Tx.Exec()`, `Tx.QueryAll()`, etc. - same for transactions

**Implementation Notes**:
- Wrap `database/sql` calls
- Add timeout handling via context
- Error wrapping (convert sql.ErrNoRows to ErrNotFound)

### 7. `executor.go` - Connection Management
**Purpose**: Database connection and pool management
**Dependencies**: `executor.go` (core methods), `types.go`
**Exports**:
- `Open(driverName, dsn string, opts ...Option) (*DB, error)` - open connection
- `OpenWithoutValidation()` - open without validation
- `DB.Close()` - close connection
- `DB.Ping()` - ping database
- `DB.Begin()` - start transaction
- `Tx.Commit()` - commit transaction
- `Tx.Rollback()` - rollback transaction
- `Option` type and option functions

**Implementation Notes**:
- Connection pooling configuration
- Timeout handling
- **Calls `MustValidateAllRegistered()` in `Open()`** (depends on validation)

---

## Layer 5: Validation System

### 8. `validate.go`
**Purpose**: Validate registered models have required methods and tags
**Dependencies**: `registry.go`, `reflect.go`, `types.go`
**Exports**:
- `ValidateModel[T ModelInterface](model T) error` - validate single model
- `ValidateAllRegistered() error` - validate all registered models
- `MustValidateAllRegistered()` - panic if validation fails
- Internal: validation rules checking

**Validation Rules**:
- Models with `load:"primary"` tag must have `QueryBy{Field}()` method
- Models with `load:"unique"` tag should have `QueryBy{Field}()` method (optional)
- Models with `load:"composite:name"` tags must have `QueryBy{Field1}{Field2}...()` method
- Composite key fields must be sorted alphabetically in method name
- Only one field can have `load:"primary"`
- Method signatures must match expected patterns

**Implementation Notes**:
- Collect all errors, don't fail fast during collection
- Group errors by model
- Return comprehensive error list
- Called automatically in `Open()`

---

## Layer 6: Query Helpers

### 9. `query.go`
**Purpose**: Type-safe query helpers that use deserialization
**Dependencies**: `executor.go`, `deserialize.go`, `types.go`
**Exports**:
- `QueryAll[T ModelInterface](ctx, exec, query, args...) ([]T, error)` - query all rows
- `QueryFirst[T ModelInterface](ctx, exec, query, args...) (T, error)` - query first row
- `QueryOne[T ModelInterface](ctx, exec, query, args...) (T, error)` - query exactly one row (returns ErrNotFound if none)

**Implementation Notes**:
- Use Executor.QueryAll/QueryRowMap internally
- Use DeserializeForType to convert rows to models
- Type-safe via generics

---

## Layer 7: Load Methods

### 10. `load.go` - Core Load Logic
**Purpose**: Load models from database using their query methods
**Dependencies**: `query.go`, `reflect.go`, `validate.go`, `types.go`
**Exports**:
- `Load[T ModelInterface](ctx, exec, model T) error` - load by primary key
- `LoadByField[T ModelInterface](ctx, exec, model T, fieldName string) error` - load by any field
- `LoadByComposite[T ModelInterface](ctx, exec, model T, compositeName string) error` - load by composite key

**Implementation Notes**:
- Use reflection to find `load` tags
- Use reflection to find and call query methods (`QueryBy{Field}()`)
- Use QueryOne to execute query and deserialize
- Update model in-place
- Handle embedded Model struct

### 11. `model.go` - Model Methods
**Purpose**: Methods on Model struct that delegate to load functions
**Dependencies**: `load.go`, `types.go`
**Exports**:
- `Model.Load(ctx, exec) error` - calls `Load()` helper
- `Model.Deserialize(row) error` - calls `DeserializeForType()`

**Implementation Notes**:
- Model.Load() uses reflection to get concrete type
- Delegates to typedb.Load() function

---

## Dependency Graph Summary

```
errors.go (Layer 0)
    ↓
types.go (Layer 0) ──┐
    ↓                │
registry.go (Layer 1)│
    ↓                │
reflect.go (Layer 2) │
    ↓                │
deserialize.go (Layer 3)
    ↓                │
executor.go (Layer 4)│
    ↓                │
validate.go (Layer 5)│
    ↓                │
query.go (Layer 6) ──┘
    ↓
load.go (Layer 7)
    ↓
model.go (Layer 7)
```

## Implementation Order

1. ✅ **errors.go** - Define error types
2. ✅ **types.go** (types only) - Define interfaces and structs
3. ✅ **registry.go** - Model registration system
4. ✅ **reflect.go** - Reflection utilities
5. ✅ **deserialize.go** - Row → Model conversion
6. ✅ **executor.go** (core) - Executor interface implementation
7. ✅ **executor.go** (connection) - Open/Close/Begin/Commit
8. ✅ **validate.go** - Model validation
9. ✅ **query.go** - Query helpers (QueryAll, QueryFirst, QueryOne)
10. ✅ **load.go** - Load methods (Load, LoadByField, LoadByComposite)
11. ✅ **model.go** - Model struct methods

## Testing Strategy

Test each layer as you build it:
- **Layer 0**: Unit tests for error types, type definitions
- **Layer 1**: Unit tests for registration (register, retrieve)
- **Layer 2**: Unit tests for reflection helpers (mock structs)
- **Layer 3**: Unit tests for deserialization (mock rows)
- **Layer 4**: Integration tests with real database (or sqlmock)
- **Layer 5**: Unit tests for validation (mock models)
- **Layer 6**: Integration tests with real database
- **Layer 7**: Integration tests with real database

## Notes

- Each layer should be testable independently
- Use interfaces to allow mocking dependencies
- Keep implementations simple and focused
- Document exported functions
- Follow Go naming conventions
