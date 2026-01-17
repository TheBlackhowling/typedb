# Checkptr Fix - Next Steps for Completion

## Current Status

✅ **Main fix implemented**: Always use `buildFieldMapFromPtr` to bypass checkptr errors
✅ **Bug fix implemented**: Store `fieldValuePtr` directly instead of calling `.Addr()` on non-addressable value
✅ **InsertedId registered**: Added init function to register `InsertedId` globally
✅ **Changes pushed**: Commits f191c50 and 3c3e19f pushed to `remove-redundant-deserialize-methods` branch

❌ **Test failures remain**: Two tests failing with pattern where first field works but subsequent fields don't

## The Problem

**Failing Tests**:
1. `TestDeserialize_NonAddressableValue` - `ID` works, `Name` is empty
2. `TestModel_Deserialize` - `ID` works, `Name` and `Email` are empty

**Pattern**: First field works correctly, subsequent fields are not being set.

## Root Cause Analysis Needed

### Hypothesis
The pointer from `reflect.NewAt(fieldType, fieldPtr)` may not be properly usable when calling `.Interface()` on it. The pointer value might not preserve the memory reference through interface conversion, causing `DeserializeToField` to fail for fields after the first one.

### Code Flow
1. `buildFieldMapFromPtr` creates: `fieldValuePtr := reflect.NewAt(fieldType, fieldPtr)` (line 157)
2. Stores: `fieldMap[dbTag] = fieldValuePtr` (line 193)
3. `Deserialize` retrieves: `fieldValue.Interface()` (line 66)
4. `DeserializeToField` receives pointer and calls: `reflect.ValueOf(target).Elem()` (line 269)

### Key Difference
- `buildFieldMap`: Stores `fieldValue.Addr()` where `fieldValue` is addressable from `v.Field(i)`
- `buildFieldMapFromPtr`: Stores `fieldValuePtr` from `reflect.NewAt` which creates pointer to memory location

## Debugging Steps

### 1. Verify Pointer Type
Add debug logging to check:
```go
fmt.Printf("fieldValuePtr.Kind(): %v\n", fieldValuePtr.Kind())
fmt.Printf("fieldValuePtr.Type(): %v\n", fieldValuePtr.Type())
fmt.Printf("fieldValuePtr.CanAddr(): %v\n", fieldValuePtr.CanAddr())
```

### 2. Verify Interface Conversion
Add debug logging in `Deserialize`:
```go
if fieldValue, ok := fieldMap[key]; ok {
    ptr := fieldValue.Interface()
    fmt.Printf("key=%s, ptr=%v, ptr type=%T\n", key, ptr, ptr)
    // Check if ptr is actually a pointer
    if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
        return fmt.Errorf("field %s: not a pointer, got %v", key, reflect.ValueOf(ptr).Kind())
    }
    if err := DeserializeToField(ptr, value); err != nil {
        return fmt.Errorf("field %s: %w", key, err)
    }
}
```

### 3. Compare Working vs Non-Working
- Run `TestInsertedId_Deserialize` (works) with debug logging
- Run `TestDeserialize_NonAddressableValue` (fails) with debug logging
- Compare the pointer values and types

### 4. Test Pointer Usability
Create a test to verify if pointer from `reflect.NewAt` works:
```go
type TestStruct struct {
    ID   int64
    Name string
}

s := &TestStruct{}
basePtr := unsafe.Pointer(s)

// Test first field
idPtr := unsafe.Pointer(uintptr(basePtr) + unsafe.Offsetof(s.ID))
idValuePtr := reflect.NewAt(reflect.TypeOf(int64(0)), idPtr)
idPtrInterface := idValuePtr.Interface()
// Try to set value via DeserializeToField

// Test second field  
namePtr := unsafe.Pointer(uintptr(basePtr) + unsafe.Offsetof(s.Name))
nameValuePtr := reflect.NewAt(reflect.TypeOf(""), namePtr)
namePtrInterface := nameValuePtr.Interface()
// Try to set value via DeserializeToField
```

## Potential Fixes to Try

### Fix 1: Ensure Pointer is Addressable
Maybe we need to verify the pointer is usable before storing:
```go
if !fieldValuePtr.CanAddr() {
    // Handle error or use different approach
}
```

### Fix 2: Use Pointer Value Directly
Maybe we need to get the actual pointer value:
```go
// Instead of storing fieldValuePtr, store the actual pointer value
ptrValue := fieldValuePtr.Interface()
fieldMap[dbTag] = reflect.ValueOf(ptrValue)
```

### Fix 3: Check Field Offset Calculation
Verify that field offsets are calculated correctly for all fields:
```go
fmt.Printf("Field %s: offset=%d, type=%v\n", dbTag, fieldOffset, fieldType)
```

### Fix 4: Compare with buildFieldMap
Add logging to compare what `buildFieldMap` stores vs what `buildFieldMapFromPtr` stores:
```go
// In buildFieldMap
fmt.Printf("buildFieldMap: key=%s, value.Kind()=%v, value.Type()=%v\n", dbTag, fieldValue.Addr().Kind(), fieldValue.Addr().Type())

// In buildFieldMapFromPtr  
fmt.Printf("buildFieldMapFromPtr: key=%s, value.Kind()=%v, value.Type()=%v\n", dbTag, fieldValuePtr.Kind(), fieldValuePtr.Type())
```

## Files to Modify

1. **deserialize.go** - Add debug logging and test fixes
2. **insert_test.go** - May need test adjustments once fix is found
3. **load_test.go** - May need test adjustments once fix is found

## Expected Outcome

Once fixed:
- All tests should pass
- CI should pass across all Go versions 1.18-1.25
- Checkptr errors should be resolved
- All fields should be deserialized correctly

## Quick Test Command

```bash
go test -v -run "TestDeserialize_NonAddressableValue|TestModel_Deserialize|TestInsertedId_Deserialize"
```
