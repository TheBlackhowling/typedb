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
- Layer 6: Query Helpers (`query.go`)
  - `QueryAll[T]` - Execute query and return all rows as slice of model pointers
  - `QueryFirst[T]` - Execute query and return first row (returns nil if no rows, no error)
  - `QueryOne[T]` - Execute query and return exactly one row (returns ErrNotFound if none)
  - Comprehensive test coverage for all query functions (100% coverage)
  - Tests cover success cases, error handling, and edge cases
- Cursor command for code coverage (`/coverage`)
  - Generates coverage profile and displays summary
  - Documents PowerShell-specific quoting requirements
- Testing and coverage documentation in `CONTEXT.md`
  - Commands for running tests with race detection
  - Coverage report generation instructions
  - Note about coverage files being ignored by `.gitignore`
