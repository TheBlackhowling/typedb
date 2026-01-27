# typedb Backlog

This file tracks planned improvements and features for typedb.

## Composite Key Improvements

### 1. Improve Error Messages for Missing Query Methods
**Priority:** Medium  
**Status:** Backlog

When `LoadByComposite` is called but the required `QueryBy{Field1}{Field2}...()` method is missing, improve the error message to be more helpful:

- Include the expected method name
- Show the field names used (sorted alphabetically)
- Provide guidance on the expected method signature
- Example: `"typedb: QueryByPostIDUserID() method not found. Composite key "userpost" requires a method named QueryByPostIDUserID() that returns a SQL query string. Fields are sorted alphabetically: PostID, UserID."`

**Related Code:**
- `load.go` line 164: Current error message
- `load.go` lines 158-165: Method name building and error handling

### 2. Add Validation at Model Registration Time
**Priority:** Medium  
**Status:** Backlog

Add validation during `RegisterModel` or `RegisterModelWithOptions` to catch missing `QueryBy` methods early:

- Validate that all required `QueryBy` methods exist for:
  - Primary keys (`QueryBy{PrimaryField}()`)
  - Unique fields (`QueryBy{UniqueField}()`)
  - Composite keys (`QueryBy{Field1}{Field2}...()`)
- Fail fast at registration time rather than at runtime
- Provide clear error messages indicating which methods are missing

**Related Code:**
- `registry.go`: Model registration functions
- `validate.go`: Model validation logic
- `load.go`: Load method implementations

## Code Duplication Refactoring

### 3. Refactor DB vs Tx Method Duplication
**Priority:** High  
**Status:** ✅ Complete

~~Significant code duplication exists between `DB` and `Tx` methods in `executor.go`~~

**Completed:** All DB and Tx methods now use shared helper functions:
- `execHelper()` - used by both DB.Exec and Tx.Exec
- `queryAllHelper()` - used by both DB.QueryAll and Tx.QueryAll
- `queryRowMapHelper()` - used by both DB.QueryRowMap and Tx.QueryRowMap
- `getIntoHelper()` - used by both DB.GetInto and Tx.GetInto
- `queryDoHelper()` - used by both DB.QueryDo and Tx.QueryDo
- `getLoggerHelper()` - shared helper for logger access
- `withTimeoutHelper()` - shared helper for timeout handling

**Result:** Code duplication eliminated, ~150+ lines of duplication removed. All methods now delegate to shared helpers that accept the `sqlQueryExecutor` interface.

**Related Code:**
- `executor.go`: All helper functions implemented and used by both DB and Tx methods

### 4. Refactor Open() vs OpenWithoutValidation() Duplication
**Priority:** Medium  
**Status:** Backlog

`Open()` and `OpenWithoutValidation()` functions have ~35 lines of duplicated code (lines 463-498 vs 502-533).

**Differences:**
- Logger message text
- Call to `MustValidateAllRegistered()` (present vs absent)
- Final logger message text

**Approach:**
- Extract common connection setup logic into a helper function
- Pass a flag or parameter to control validation behavior

**Related Code:**
- `executor.go`: Open() and OpenWithoutValidation() functions

### 5. Extract Shared processFields Logic
**Priority:** Medium  
**Status:** Backlog

Multiple `processFields` closures exist with similar field iteration logic:
- `insert.go` line 398 - `serializeModelFields()`
- `update.go` line 222 - `serializeModelFieldsForUpdate()`
- `update.go` line 401 - `buildFieldMapForComparison()`
- `deserialize.go` line 172 - different signature (uses `unsafe.Pointer`)

**Common Patterns:**
- Iterating struct fields
- Handling embedded structs
- Extracting column names from db tags
- Checking for dot notation
- Skipping unexported fields

**Approach:**
- Extract shared field iteration logic into reusable helper functions
- Use callback/visitor pattern to parameterize field processing behavior
- Maintain type safety and performance characteristics

**Related Code:**
- `insert.go`: serializeModelFields()
- `update.go`: serializeModelFieldsForUpdate(), buildFieldMapForComparison()
- `deserialize.go`: Field iteration logic

## Pre-1.0.0 Release Checklist

### 6. API Stability Review
**Priority:** High  
**Status:** Backlog

Before releasing v1.0.0, conduct a comprehensive API stability review:

- Review all public APIs for breaking changes
- Document any planned breaking changes
- Ensure backward compatibility plan is in place
- Update API documentation to reflect stable APIs
- Consider deprecation warnings for any APIs that will change

**Related Code:**
- All public functions and types in the codebase
- `API.md`: API documentation

### 7. Complete Documentation for 1.0.0
**Priority:** High  
**Status:** Backlog

Ensure all documentation is complete and ready for 1.0.0:

- Update README.md to remove "Early Development" warnings
- Complete API.md with all public APIs documented
- Add migration guide if any breaking changes are planned
- Update examples to reflect stable API
- Add performance benchmarks and guidance

**Related Files:**
- `README.md`
- `API.md`
- `examples/` directory

### 8. Performance Benchmarks Documentation
**Priority:** Medium  
**Status:** Backlog

Document performance characteristics and benchmarks:

- Add benchmark results to documentation
- Document performance trade-offs
- Provide guidance on when typedb is appropriate vs. raw database/sql
- Include real-world performance examples

**Related Code:**
- Performance section in README.md
- Consider adding benchmark tests

### 9. Final Code Quality Pass
**Priority:** Medium  
**Status:** Backlog

Before 1.0.0, ensure code quality standards are met:

- Run all static analysis tools and fix any remaining issues
- Ensure test coverage is comprehensive
- Review and fix any remaining linter warnings
- Verify all examples compile and run correctly
- Check that all integration tests pass

**Related Code:**
- `scripts/static-analysis.sh` and `scripts/static-analysis.ps1`
- All test files
- All example files