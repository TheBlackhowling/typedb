# Checkptr Fix - Work Summary and Status

## Problem Statement

**Error**: `fatal error: checkptr: pointer arithmetic result points to invalid allocation` occurring in `buildFieldMap.func1` when calling `reflect.Value.Field()`.

**Scope**: Failing across ALL Go versions 1.18-1.25 in CI, not version-specific.

**Root Cause**: Values created via `reflect.NewAt` (used in `Model.Deserialize`) trigger checkptr validation when accessing fields via `reflect.Value.Field()`, even when `CanAddr()` returns true.

## Solution Implemented

### Core Fix (Commit f191c50)
- **Changed**: Always use `buildFieldMapFromPtr` instead of conditional `CanAddr()` check
- **Location**: `deserialize.go` line 58
- **Rationale**: Bypasses `reflect.Value.Field()` entirely using unsafe pointer arithmetic, avoiding checkptr validation

### Bug Fix in buildFieldMapFromPtr
- **Changed**: Store `fieldValuePtr` directly instead of calling `.Addr()` on non-addressable value
- **Location**: `deserialize.go` line 193
- **Before**: `fieldMap[dbTag] = fieldValuePtr.Elem().Addr()` ❌ (fails: value not addressable)
- **After**: `fieldMap[dbTag] = fieldValuePtr` ✅ (stores pointer directly)

### Additional Fixes (Commit 3c3e19f)
- **Changed**: Register `InsertedId` globally in `insert.go` init function
- **Reason**: `Model.Deserialize` requires models to be registered to find outer struct type
- **Changed**: Updated test comments to reflect that we always use unsafe path

## Current Status

### ✅ Completed
1. Main checkptr fix implemented and committed
2. Bug fix in `buildFieldMapFromPtr` (storing pointer directly)
3. `InsertedId` registered globally
4. Test comments updated
5. Changes pushed to `remove-redundant-deserialize-methods` branch

### ❌ Known Issues (Test Failures)

**Failing Tests**:
1. `TestDeserialize_NonAddressableValue` - `Name` field not being set (ID works, Name doesn't)
2. `TestModel_Deserialize` - Fields not being deserialized correctly

**Symptoms**:
- First field (`ID`) works correctly
- Subsequent fields (`Name`, `Email`) are not being set
- Suggests issue with how pointers from `reflect.NewAt` are being used

**Hypothesis**:
The pointer from `reflect.NewAt(fieldType, fieldPtr)` may not be usable correctly when calling `.Interface()` on it. The pointer value might not properly track the memory location when passed through the interface conversion.

## Code Changes Made

### deserialize.go
```go
// Line 53-58: Always use unsafe path
fieldMap := buildFieldMapFromPtr(destValue, structValue)

// Line 157: Create pointer using reflect.NewAt
fieldValuePtr := reflect.NewAt(fieldType, fieldPtr)

// Line 193: Store pointer directly (FIXED)
fieldMap[dbTag] = fieldValuePtr  // Was: fieldValuePtr.Elem().Addr()
```

### insert.go
```go
// Lines 72-77: Register InsertedId globally
func init() {
    RegisterModel[*InsertedId]()
}
```

## What's Left To Do

### 1. Fix Test Failures
**Problem**: Fields after the first one aren't being set when using `buildFieldMapFromPtr`.

**Investigation Needed**:
- Verify that `fieldValuePtr.Interface()` returns a usable pointer
- Check if `DeserializeToField` can properly use pointers from `reflect.NewAt`
- Compare behavior between `buildFieldMap` (works) and `buildFieldMapFromPtr` (fails for some fields)

**Possible Solutions**:
- Ensure pointer from `reflect.NewAt` is properly addressable
- Check if we need to use `.Addr()` differently
- Verify field offset calculations are correct
- Test if the issue is with how `.Interface()` works on pointers from `reflect.NewAt`

### 2. Verify CI Passes
- Once test failures are fixed, verify CI passes across all Go versions 1.18-1.25
- The checkptr error should be resolved, but need to ensure functionality works correctly

## Key Files Modified

1. **deserialize.go**
   - Line 58: Always use `buildFieldMapFromPtr`
   - Line 193: Store pointer directly (bug fix)
   - Line 203-205: Updated comment for `buildFieldMap`

2. **insert.go**
   - Lines 72-77: Added init function to register `InsertedId`

3. **insert_test.go**
   - Updated test comments to reflect new approach

## Technical Details

### How buildFieldMapFromPtr Works
1. Gets struct address via `ptrValue.Pointer()`
2. Calculates field addresses using `unsafe.Pointer(uintptr(basePtr) + uintptr(fieldOffset))`
3. Creates pointers using `reflect.NewAt(fieldType, fieldPtr)`
4. Stores pointers directly in fieldMap

### Why This Should Work
- `reflect.NewAt(fieldType, fieldPtr)` creates `*fieldType` (pointer to field)
- This matches what `buildFieldMap` stores: `fieldValue.Addr()` → `*fieldType`
- Both paths should store identical values: pointers to fields
- `DeserializeToField` receives `fieldValue.Interface()` which unwraps the pointer

### Why It Might Not Be Working
- Pointers from `reflect.NewAt` might not be properly usable with `.Interface()`
- The pointer value might not track memory location correctly through interface conversion
- There might be an issue with how `DeserializeToField` uses these pointers

## Next Steps

1. **Debug the test failures**:
   - Add logging to see what `fieldValuePtr.Interface()` returns
   - Verify pointers are correct before storing in fieldMap
   - Check if `DeserializeToField` receives valid pointers
   - Test if calling `.Interface()` on pointer from `reflect.NewAt` works correctly

2. **Compare working vs non-working cases**:
   - ✅ `TestInsertedId_Deserialize` passes (single field: `ID`)
   - ✅ `TestDeserialize_AddressableValue` passes (direct `Deserialize` call)
   - ❌ `TestDeserialize_NonAddressableValue` fails (`ID` works, `Name` empty)
   - ❌ `TestModel_Deserialize` fails (all fields empty except `ID`)
   - Pattern: First field works, subsequent fields don't

3. **Potential fixes to try**:
   - Verify `fieldValuePtr` from `reflect.NewAt` is actually a pointer type
   - Check if `.Interface()` on `reflect.NewAt` pointer preserves memory reference
   - Ensure pointer is addressable before storing
   - Compare exact behavior: `buildFieldMap` stores `fieldValue.Addr()` vs `buildFieldMapFromPtr` stores `fieldValuePtr`
   - Test if issue is with how `DeserializeToField` uses pointers from `reflect.NewAt`
   - Verify field offset calculations are correct for all fields

4. **Debugging approach**:
   - Add debug output in `buildFieldMapFromPtr` to log field values before storing
   - Add debug output in `Deserialize` to log what `fieldValue.Interface()` returns
   - Compare pointer addresses between working (`buildFieldMap`) and non-working (`buildFieldMapFromPtr`) paths

## Commits Made

1. **f191c50**: "fix: resolve checkptr error by always using unsafe path for field access"
2. **3c3e19f**: "fix: register InsertedId and update test comments"

Both commits are pushed to `remove-redundant-deserialize-methods` branch.

## Test Details

### Failing Test: TestDeserialize_NonAddressableValue
- **Structure**: `TestModelNonAddr` with `Model`, `ID int64`, `Name string`
- **Test**: Calls `model.Deserialize(row)` where row has `id` and `name`
- **Result**: `ID` is set correctly (123), but `Name` is empty string
- **Pattern**: First field works, second field doesn't

### Failing Test: TestModel_Deserialize  
- **Structure**: `LoadTestUser` (need to check exact structure)
- **Test**: Calls `user.Deserialize(row)` where row has `id`, `name`, `email`
- **Result**: Only `ID` is set (123), `Name` and `Email` are empty
- **Pattern**: Same as above - first field works, others don't

### Working Test: TestInsertedId_Deserialize
- **Structure**: `InsertedId` with `Model`, `ID int64` (only one field)
- **Test**: Calls `insertedId.Deserialize(row)` where row has `id`
- **Result**: ✅ Passes - `ID` is set correctly

**Key Insight**: Single-field structs work, multi-field structs fail on fields after the first one.

## References

- Planning document: `docs/CHECKPTR_FIX_PLANNING.md`
- Main fix implements "Option D" from planning document
- Fix addresses checkptr errors across all Go versions 1.18-1.25
