# Deserialize Field Failure Analysis

## Problem Summary

Two tests are failing:
1. `TestDeserialize_NonAddressableValue` - Only the first field (`ID`) is set correctly, subsequent fields (`Name`) are not set
2. `TestModel_Deserialize` - Only the first field (`ID`) is set correctly, subsequent fields (`Name`, `Email`) are not set

## Test Details

### TestDeserialize_NonAddressableValue
- Struct: `TestModelNonAddr` with fields `ID int64` and `Name string`
- Expected: Both `ID=789` and `Name="Non-Addressable Test"` should be set
- Actual: `ID=789` works, `Name=""` (empty string)

### TestModel_Deserialize  
- Struct: `LoadTestUser` with fields `ID int`, `Name string`, `Email string`
- Expected: All three fields should be set (`ID=123`, `Name="Alice"`, `Email="alice@example.com"`)
- Actual: `ID=123` works, `Name=""` and `Email=""` (empty strings)

## Current Implementation

### Deserialize Flow
1. `Deserialize()` calls `buildFieldMapFromPtr()` to create field map
2. For each row key, looks up field in map
3. Calls `deserializeToFieldValue()` (or `DeserializeToField()` via interface conversion)
4. Sets the value

### buildFieldMapFromPtr Implementation
- Uses `reflect.NewAt(fieldType, fieldPtr)` to create pointers to fields
- Stores `fieldValuePtr` (which is `*fieldType`) directly in the map
- Uses unsafe pointer arithmetic to calculate field addresses

## Hypothesis

The issue appears to be related to how `reflect.NewAt` pointers work when converted to `interface{}` and used in type switches or reflection operations.

### Possible Causes

1. **Interface Conversion Issue**: When `fieldValuePtr.Interface()` is called on a `reflect.Value` created by `reflect.NewAt`, the resulting interface may not preserve the correct type information for type assertions in `DeserializeToField()`.

2. **Pointer Validity**: The pointers created by `reflect.NewAt` may become invalid or lose their connection to the original memory when stored in the map and retrieved later.

3. **Type Switch Mismatch**: The type switch in `DeserializeToField()` may not match correctly for pointers created via `reflect.NewAt`, causing it to fall through to the reflection path which might not work correctly.

4. **Field Order Dependency**: The fact that the first field works but subsequent fields don't suggests there might be an issue with how field addresses are calculated or stored for fields after the first one.

## Current Attempts

1. **Created `deserializeToFieldValue()`**: A new function that works directly with `reflect.Value` instead of converting to `interface{}` first. This attempts to avoid the interface conversion issue.

2. **Direct Assignment Path**: The new function tries direct assignment and conversion before falling back to `DeserializeToField()`.

## Next Steps for Investigation

1. **Add Debug Logging**: Log what `fieldValuePtr.Interface()` returns for each field to see if the type information is preserved.

2. **Verify Pointer Validity**: Check if the pointers remain valid when retrieved from the map.

3. **Test Field Address Calculation**: Verify that `fieldOffset` calculations are correct for all fields, especially after the first one.

4. **Compare with Working Path**: Compare how `buildFieldMap` (the old path) worked vs `buildFieldMapFromPtr` to identify what's different.

5. **Check reflect.NewAt Behavior**: Research/document how `reflect.NewAt` pointers behave when:
   - Stored in maps
   - Converted to interfaces
   - Used in type switches
   - Used with `.Set()` operations

## Key Code Locations

- `deserialize.go:60-72` - Main Deserialize loop
- `deserialize.go:82-153` - `buildFieldMapFromPtr` function
- `deserialize.go:155-281` - `DeserializeToField` function
- `deserialize.go:283-330` - `deserializeToFieldValue` function (new)
- `insert_test.go:579-608` - `TestDeserialize_NonAddressableValue`
- `load_test.go:261-277` - `TestModel_Deserialize`

## Related Context

This issue emerged after implementing Option D from `CHECKPTR_FIX_PLANNING.md`, which involved:
- Always using `buildFieldMapFromPtr` instead of `buildFieldMap`
- Using unsafe pointer arithmetic to bypass `reflect.Value.Field()` and avoid checkptr errors
- Fixing a bug where `fieldValuePtr.Elem().Addr()` was incorrectly called

The checkptr error is resolved, but now we have this field assignment issue.
