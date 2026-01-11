# Unreleased Changes

This file tracks changes that have been made but not yet released in a version.

When a version is released, the content from this file will be moved to the version file and this file will be cleared.

## How to Use

Add changelog entries here as you make changes. When ready to release:

1. The changelog action will automatically use this content when creating a version
2. Or include changelog content in your PR description
3. Or manually trigger a release via workflow_dispatch

---

## Current Unreleased Changes

## Added
- Insert functionality with ID retrieval (`insert.go`)
  - `InsertAndReturn[T]`: Executes INSERT with RETURNING/OUTPUT clause and deserializes returned row into model
  - `InsertAndGetId`: Convenience helper that returns inserted ID as int64
  - Supports PostgreSQL, SQLite, SQL Server (RETURNING/OUTPUT clauses)
  - Supports MySQL and SQLite using `sql.Result.LastInsertId()` (safe with connection pooling)
  - Includes internal `InsertedId` model for ID retrieval
- Comprehensive unit tests for insert functionality (`insert_test.go`)
  - Tests for `InsertAndReturn` with success and error cases
  - Tests for `InsertAndGetId` across all database types
  - Tests for MySQL/SQLite `LastInsertId()` path
  - Tests for error handling (database errors, deserialization failures, missing RETURNING clauses)
  - Tests for transaction integration
  - 100% statement coverage for insert functions
- Comprehensive unit tests for uint deserialization functions
  - `DeserializeUint64`: 23 test cases covering all type conversions, error cases, and edge cases
  - `DeserializeUint32`: 22 test cases including overflow validation
  - `DeserializeUint`: 22 test cases covering all conversion paths
  - Improved coverage from 8.7-9.5% to 100% for all three functions
- Driver name tracking in DB and Tx structs
  - Added `driverName` field to `DB` and `Tx` structs for database-specific logic
  - Updated `NewDB` signature to accept `driverName` parameter
  - Updated `Open` and `OpenWithoutValidation` to pass driver name
  - Updated `DB.Begin` to propagate driver name to transactions

## Changed
- Updated `DeserializeToField` to support uint types
  - Added handling for `*uint64`, `*uint32`, and `*uint` pointer fields
  - Uses new `DeserializeUint64`, `DeserializeUint32`, and `DeserializeUint` functions
- Improved overall code coverage from 84.4% to 90.2% (+5.8%)
  - `DeserializeUint64`: 8.7% → 100%
  - `DeserializeUint32`: 9.5% → 100%
  - `DeserializeUint`: 9.1% → 100%
  - `InsertAndReturn`: 100% coverage
  - `InsertAndGetId`: 100% coverage
- Updated Go version requirement from 1.18 to 1.23
- Updated `go-sqlmock` dependency from indirect to direct dependency
