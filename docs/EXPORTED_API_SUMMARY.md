# Exported API Summary

This document lists all currently exported functions, types, and variables from the `typedb` package.

## Core Database Operations

### Query Functions
- `QueryAll[T]` - Query multiple rows, returns slice
- `QueryFirst[T]` - Query first row, returns nil if not found
- `QueryOne[T]` - Query exactly one row, returns ErrNotFound if not found

### Insert Functions
- `Insert[T]` - Auto-build INSERT query
- `InsertAndReturn[T]` - INSERT with RETURNING/OUTPUT clause
- `InsertAndGetId` - Convenience for getting just the ID (int64)

### Update Functions
- `Update[T]` - Auto-build UPDATE query

### Load Functions
- `Load[T]` - Load by primary key
- `LoadByField[T]` - Load by any field with load tag
- `LoadByComposite[T]` - Load by composite key

## Connection & Transaction Management

### DB Creation
- `Open` - Open DB with validation
- `OpenWithoutValidation` - Open DB without validation
- `NewDB` - Create DB from existing *sql.DB

### DB Methods (on *DB)
- `Exec` - Execute non-query SQL
- `QueryAll` - Query all rows as []map[string]any
- `QueryRowMap` - Query single row as map[string]any
- `GetInto` - Scan single row into dest pointers
- `QueryDo` - Streaming query with callback
- `Close` - Close database connection
- `Ping` - Verify connection
- `Begin` - Start transaction
- `WithTx` - Execute function within transaction

### Transaction Methods (on *Tx)
- `Exec` - Execute non-query SQL
- `QueryAll` - Query all rows as []map[string]any
- `QueryRowMap` - Query single row as map[string]any
- `GetInto` - Scan single row into dest pointers
- `QueryDo` - Streaming query with callback
- `Commit` - Commit transaction
- `Rollback` - Rollback transaction

### Configuration Options
- `WithMaxOpenConns` - Set max open connections
- `WithMaxIdleConns` - Set max idle connections
- `WithConnMaxLifetime` - Set connection max lifetime
- `WithConnMaxIdleTime` - Set connection max idle time
- `WithTimeout` - Set default operation timeout

## Types & Interfaces

### Core Types
- `Executor` - Interface for executing database queries
- `DB` - Database connection wrapper
- `Tx` - Transaction wrapper
- `Config` - Database connection configuration
- `Model` - Base struct to embed in models
- `ModelInterface` - Interface models must implement (requires `deserialize` method - unexported)
- `Option` - Function type for configuring DB

### Error Types
- `ValidationError` - Single model validation error
- `ValidationErrors` - Multiple validation errors

### Internal Types
- `InsertedId` - Internal model used by `InsertAndGetId` (exported but documented as internal)

## Registration & Validation

- `RegisterModel[T]` - Register a model type
- `GetRegisteredModels` - Get all registered models (used by validation)
- `ValidateModel[T]` - Validate a single model
- `ValidateAllRegistered` - Validate all registered models
- `MustValidateAllRegistered` - Validate all models and panic on failure

## Serialization Helpers

These are utility functions for converting Go values to database-compatible formats:

- `Serialize` - Generic serialization
- `SerializeJSONB` - Serialize to JSONB format
- `SerializeIntArray` - Serialize int slice to PostgreSQL array format
- `SerializeStringArray` - Serialize string slice to PostgreSQL array format

**Note**: These are PostgreSQL-specific helpers. They're exported because they're useful utilities for preparing data before database operations.

## Errors

- `ErrNotFound` - No rows found
- `ErrFieldNotFound` - Field not found (from reflect.go)
- `ErrMethodNotFound` - Method not found (from reflect.go)

## Model Methods

### Model.deserialize (UNEXPORTED)
- The `Model` struct has an unexported `deserialize` method
- Users should NOT call this directly
- Use `QueryAll`, `QueryFirst`, `QueryOne`, `InsertAndReturn`, `Load`, etc. instead

## Summary

### ✅ Appropriately Exported
- All query/insert/update/load operations
- Connection management (Open, DB, Tx)
- Configuration options
- Model registration and validation
- Serialization helpers (useful utilities)
- Error types and variables

### ✅ Appropriately Unexported (Recently Hidden)
- All deserialization helper functions (`deserializeInt`, `deserializeString`, etc.)
- `DeserializeToField` - Internal field deserialization
- `DeserializeForType` - Internal type-safe deserialization
- `deserialize` - Core deserialization function
- All reflection utilities (`getModelType`, `findFieldByTag`, etc.)
- `Model.deserialize` - Method on Model struct

### ⚠️ Potentially Questionable
- `InsertedId` - Exported but documented as internal. Used by `InsertAndGetId`. Could be unexported if we want stricter API.
- `GetRegisteredModels` - Exported, used by validation and `Model.deserialize`. Could be unexported if we want stricter API.
- Serialization helpers (`Serialize`, `SerializeJSONB`, etc.) - Exported utility functions. These are PostgreSQL-specific but useful. Could be unexported if we want stricter API.
