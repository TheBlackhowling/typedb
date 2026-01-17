# Unexport Internal Functions - Planning Document

## Objective
Hide/unexport internal deserialization helpers and reflection utilities that should not be part of the public API. Users should interact with typedb through Query, Insert, Update, Load, and other high-level database operations, not through low-level deserialization or reflection utilities.

## Status: In Progress

### Completed ✅

1. **Core Deserialization Functions** - Made unexported:
   - `Deserialize` → `deserialize` ✅
   - `DeserializeForType` → `deserializeForType` ✅

2. **Deserialization Helper Functions** - Made unexported:
   - `DeserializeToField` → `deserializeToField` ✅
   - `DeserializeInt` → `deserializeInt` ✅
   - `DeserializeInt64` → `deserializeInt64` ✅
   - `DeserializeInt32` → `deserializeInt32` ✅
   - `DeserializeUint64` → `deserializeUint64` ✅
   - `DeserializeUint32` → `deserializeUint32` ✅
   - `DeserializeUint` → `deserializeUint` ✅
   - `DeserializeBool` → `deserializeBool` ✅
   - `DeserializeString` → `deserializeString` ✅
   - `DeserializeTime` → `deserializeTime` ✅
   - `DeserializeIntArray` → `deserializeIntArray` ✅
   - `DeserializeStringArray` → `deserializeStringArray` ✅
   - `DeserializeJSONB` → `deserializeJSONB` ✅
   - `DeserializeMap` → `deserializeMap` ✅

3. **Reflection Utilities** - Made unexported:
   - `GetModelType` → `getModelType` ✅
   - `FindFieldByTag` → `findFieldByTag` ✅
   - `GetFieldValue` → `getFieldValue` ✅
   - `SetFieldValue` → `setFieldValue` ✅
   - `FindMethod` → `findMethod` ✅
   - `CallMethod` → `callMethod` ✅

4. **Internal Code Updates** - Updated all internal calls:
   - `deserialize.go` - All internal calls updated ✅
   - `insert.go` - All internal calls updated ✅
   - `load.go` - All internal calls updated ✅
   - `update.go` - All internal calls updated ✅
   - `validate.go` - All internal calls updated ✅
   - `model.go` - Uses `GetRegisteredModels()` (still exported, which is correct) ✅

5. **Build Status**:
   - `go build ./...` passes ✅

### In Progress 🔄

**Test Files** - Need to update test files to use unexported functions:
   - `deserialize_test.go` - Partially updated, some references remain:
     - Line 507: `DeserializeString` → needs `deserializeString`
     - Line 556: `deserializeIntArray` → already updated (but may have case issue)
     - Line 563: `DeserializeStringArray` → needs `deserializeStringArray`
     - Line 583: `DeserializeMap` → needs `deserializeMap`
     - Line 595: `DeserializeMap` → needs `deserializeMap`
     - Line 617: `deserializeInt64` → already updated (but may have case issue)
     - Line 1056: `deserializeInt64` → already updated (but may have case issue)
     - Line 1302: `deserializeIntArray` → already updated (but may have case issue)
     - Line 1314: `deserializeIntArray` → already updated (but may have case issue)
     - Line 1322: `deserializeIntArray` → already updated (but may have case issue)
   - `reflect_test.go` - Partially updated, some references may remain
   - User has started updating `deserialize_test.go` (TestDeserializeUint64, TestDeserializeUint)

### Remaining Work 📋

1. **Complete Test File Updates**:
   - Fix remaining references in `deserialize_test.go`:
     - `DeserializeString` → `deserializeString` (line 507)
     - `DeserializeStringArray` → `deserializeStringArray` (line 563)
     - `DeserializeMap` → `deserializeMap` (lines 583, 595)
   - Verify all test function names are correct (user started renaming some)
   - Ensure all test files compile and pass

2. **Verify Test Coverage**:
   - Run full test suite: `go test ./...`
   - Ensure all tests pass
   - Check for any external dependencies on unexported functions

3. **Documentation** (if needed):
   - Update any public documentation that references these functions
   - Ensure README/examples don't reference unexported functions

## Current Public API (After Completion)

### ✅ Still Exported (Correct):
- **Database Operations**: `QueryAll`, `QueryFirst`, `QueryOne`, `Insert`, `InsertAndReturn`, `InsertAndGetId`, `Update`, `Load`, `LoadByField`, `LoadByComposite`
- **Connection Management**: `Open`, `OpenWithoutValidation`, `NewDB`, `DB`, `Tx`, `Executor`, `Config`, Option functions
- **Model Infrastructure**: `Model`, `ModelInterface`, `Model.Deserialize` (convenience method)
- **Registration**: `RegisterModel`, `GetRegisteredModels`
- **Validation**: `ValidateModel`, `MustValidateAllRegistered`, `ValidationError`, `ValidationErrors`
- **Errors**: `ErrNotFound`, `ErrFieldNotFound`, `ErrMethodNotFound`
- **Serialization Helpers**: `Serialize`, `SerializeJSONB`, `SerializeIntArray`, `SerializeStringArray` (these remain exported)

### ✅ Now Unexported (Correct):
- All deserialization helpers (DeserializeInt, DeserializeString, etc.)
- `DeserializeToField`
- All reflection utilities (GetModelType, FindFieldByTag, etc.)

## Next Steps

1. **Immediate**: Fix remaining test file references
   - Update `DeserializeString` → `deserializeString` in `deserialize_test.go`
   - Update `DeserializeStringArray` → `deserializeStringArray` in `deserialize_test.go`
   - Update `DeserializeMap` → `deserializeMap` in `deserialize_test.go`
   - Verify all function names are lowercase

2. **Verify**: Run full test suite
   ```bash
   go test ./...
   ```

3. **Final Check**: Ensure no external code depends on these functions
   - Check examples directory
   - Check for any imports in other projects

## Notes

- The user has started updating test files (TestDeserializeUint64, TestDeserializeUint)
- Build passes but tests are failing due to incomplete test file updates
- All internal code has been updated correctly
- The task is nearly complete - just need to finish test file updates
