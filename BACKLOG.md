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
**Status:** In Progress

Significant code duplication exists between `DB` and `Tx` methods in `executor.go`:
- `getLogger()` - identical implementations (lines 37-42 vs 380-385)
- `withTimeout()` - identical implementations (lines 47-56 vs 388-397)
- `Exec()` - nearly identical (lines 60-70 vs 248-257)
- `QueryAll()` - nearly identical (lines 74-92 vs 260-278)
- `QueryRowMap()` - similar, but Tx has less logging (lines 97-130 vs 281-309)
- `GetInto()` - nearly identical (lines 135-150 vs 312-327)
- `QueryDo()` - nearly identical (lines 154-178 vs 330-354)

**Approach:**
- Extract shared logic into helper functions that accept an interface or executor helper
- Create a common executor interface/helper to reduce duplication
- Ensure logging consistency between DB and Tx methods

**Related Code:**
- `executor.go`: DB and Tx method implementations (~150+ lines of duplication)

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
