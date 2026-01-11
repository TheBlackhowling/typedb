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
- Comprehensive sqlmock tests for Layer 4 (executor.go)
  - DB methods: Exec, QueryAll, QueryRowMap, GetInto, QueryDo (100% coverage)
  - Tx methods: Exec, QueryAll, QueryRowMap, GetInto, QueryDo
  - Tests cover success cases, error handling, empty results, and edge cases
  - Improved helper function coverage (scanRowsToMaps, scanRowToMap, scanRowToMapWithCols)
- sqlmock dependency for database mocking in tests
  - Added `github.com/DATA-DOG/go-sqlmock v1.5.2` as test dependency

## Changed
- Improved overall code coverage from 77.2% to 90.1% (+12.9%)
  - DB.Exec: 0% → 100%
  - DB.QueryAll: 0% → 100%
  - DB.QueryRowMap: 0% → 87.5%
  - DB.GetInto: 0% → 100%
  - DB.QueryDo: 0% → 100%
  - scanRowsToMaps: 0% → 75%
  - scanRowToMapWithCols: 0% → 84.6%
  - scanRowToMap: 0% → 75%
- Removed `Model.Load()` method - use `typedb.Load(ctx, exec, model)` instead
  - Due to Go's method promotion limitations with embedded structs, `Model.Load()` cannot reliably determine the outer struct type
  - This aligns with TabulaRasa's approach of using helper functions rather than methods on Model
  - Updated README.md examples to use `typedb.Load()` directly
