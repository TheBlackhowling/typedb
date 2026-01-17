# Unexport API Functions - Planning Document

## Goal
Hide/unexport deserialization helper functions and reflection utilities that should not be part of the public API. Users should deserialize via Query, InsertAndReturn, Load, etc., not directly via helper functions.

## Completed Work

### 1. Core Deserialization Functions (✅ Complete)
- `Deserialize` → `deserialize` (unexported)
- `DeserializeForType` → `deserializeForType` (unexported)
- Updated all internal calls in:
  - `query.go` (QueryAll, QueryFirst, QueryOne)
  - `insert.go` (InsertAndReturn)
  - `model.go` (Model.Deserialize)

### 2. Deserialization Helper Functions (✅ Mostly Complete)
- `DeserializeToField` → `deserializeToField` (unexported)
- `DeserializeInt` → `deserializeInt` (unexported)
- `DeserializeInt64` → `deserializeInt64` (unexported)
- `DeserializeInt32` → `deserializeInt32` (unexported)
- `DeserializeUint64` → `deserializeUint64` (unexported)
- `DeserializeUint32` → `deserializeUint32` (unexported)
- `DeserializeUint` → `deserializeUint` (unexported)
- `DeserializeBool` → `deserializeBool` (unexported)
- `DeserializeString` → `deserializeString` (unexported)
- `DeserializeTime` → `deserializeTime` (unexported)
- `DeserializeIntArray` → `deserializeIntArray` (unexported)
- `DeserializeStringArray` → `deserializeStringArray` (unexported)
- `DeserializeJSONB` → `deserializeJSONB` (unexported)
- `DeserializeMap` → `deserializeMap` (unexported)

- Updated all internal calls in `deserialize.go`

### 3. Reflection Utilities (✅ Complete)
- `GetModelType` → `getModelType` (unexported)
- `FindFieldByTag` → `findFieldByTag` (unexported)
- `GetFieldValue` → `getFieldValue` (unexported)
- `SetFieldValue` → `setFieldValue` (unexported)
- `FindMethod` → `findMethod` (unexported)
- `CallMethod` → `callMethod` (unexported)

- Updated all internal calls in:
  - `insert.go`
  - `load.go`
  - `update.go`
  - `validate.go`
  - `model.go`

## Remaining Work

### Test Files Need Updates (⚠️ In Progress)
The test files still reference the old exported function names. Need to update:

1. **`deserialize_test.go`** - Still has some references:
   - Line 507: `DeserializeString` → `deserializeString`
   - Line 556: `deserializeIntArray` (already lowercase, but may be wrong context)
   - Line 563: `DeserializeStringArray` → `deserializeStringArray`
   - Line 583: `DeserializeMap` → `deserializeMap`
   - Line 595: `DeserializeMap` → `deserializeMap`
   - Line 617: `deserializeInt64` (already lowercase, but may be wrong context)
   - Line 1056: `deserializeInt64` (already lowercase, but may be wrong context)
   - Line 1302: `deserializeIntArray` (already lowercase, but may be wrong context)
   - Line 1314: `deserializeIntArray` (already lowercase, but may be wrong context)
   - Line 1322: `deserializeIntArray` (already lowercase, but may be wrong context)
   - Lines 1711, 1767, 1823: `DeserializeUint64`, `DeserializeUint32`, `DeserializeUint` → lowercase versions

2. **`reflect_test.go`** - Already updated (✅)

## Current Status

- ✅ Build passes (`go build ./...`)
- ❌ Tests fail due to remaining references in test files
- ⚠️ Some test function names were changed (e.g., `TestDeserializeUint64` → `TestdeserializeUint64`) - this may need to be reverted or test names should remain capitalized

## Next Steps

1. **Fix remaining test file references:**
   - Search `deserialize_test.go` for all remaining capitalized function names
   - Replace with lowercase versions
   - Verify test function names follow Go conventions (should start with `Test` + capitalized name)

2. **Verify test function naming:**
   - Test functions should be `TestDeserializeUint64` (capitalized)
   - But call `deserializeUint64` (lowercase) inside the test
   - The user's changes show test names were lowercased - may need to revert this

3. **Run full test suite:**
   - `go test ./...` should pass
   - Verify no external code breaks (check examples, etc.)

## Files Modified

### Source Files:
- `deserialize.go` - All helper functions unexported, internal calls updated
- `query.go` - Uses `deserializeForType`
- `insert.go` - Uses `findFieldByTag`, `setFieldValue`, `findMethod`, `getModelType`
- `load.go` - Uses `findFieldByTag`, `getFieldValue`, `findMethod`, `getModelType`
- `update.go` - Uses `findFieldByTag`, `getFieldValue`, `setFieldValue`
- `validate.go` - Uses `findMethod`
- `model.go` - Uses `deserialize`, `GetRegisteredModels` (still exported)
- `reflect.go` - All reflection utilities unexported

### Test Files:
- `deserialize_test.go` - ✅ All references updated, test function names fixed
- `reflect_test.go` - ✅ All references updated, test function names fixed

## Notes

- `GetRegisteredModels()` is still exported (used by validation and Model.Deserialize)
- `Model.Deserialize()` is still exported (convenience method, though has type detection issues)
- Serialization helpers (`Serialize`, `SerializeJSONB`, etc.) are still exported - not part of this task
- Error variables (`ErrNotFound`, `ErrFieldNotFound`, `ErrMethodNotFound`) are still exported - appropriate

## Verification Checklist

- [x] All test files updated
- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] No external breaking changes (check examples)
- [x] Documentation updated if needed

## Summary

All work to unexport internal API functions has been completed successfully. All deserialization helper functions and reflection utilities are now unexported, and all test files have been updated to use the new lowercase function names. Test function names follow Go conventions (capitalized after `Test`), and all tests pass.
