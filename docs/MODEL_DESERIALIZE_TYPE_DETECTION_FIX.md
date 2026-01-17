# Model.Deserialize Type Detection Fix Options

## Problem

`Model.Deserialize` is incorrectly identifying the outer struct type when called on an embedded Model. It picks the first registered model that has Model as the first field, without verifying the pointer actually points to that type.

### Current Behavior
- Test calls `model.Deserialize(row)` where `model` is `*TestModelNonAddr`
- `Model.Deserialize` finds `InsertedId` first (also has Model as first field)
- Uses `InsertedId` instead of `TestModelNonAddr`
- Result: `Name` field is missing because `InsertedId` doesn't have it

### Root Cause
The code uses `reflect.NewAt` to create a pointer to any type at the address, then checks if it implements `ModelInterface`. Since `reflect.NewAt` can create a pointer to any type at that address, the type assertion always succeeds, making it impossible to verify the actual type.

## Options for Fix

### Option 1: Verify Type by Checking Struct Size
**Approach**: Check if the struct size matches what we expect at that address.

**Pros**:
- Simple to implement
- Doesn't require additional type information

**Cons**:
- Not foolproof - different structs could have the same size
- Doesn't verify field layout, just size

**Implementation**:
```go
// Check if struct size matches
if structType.Size() != getActualStructSize(outerPtr) {
    continue // Wrong type
}
```

### Option 2: Use Runtime Type Information from Interface
**Approach**: When `model.Deserialize(row)` is called, the original value has type information. We need to preserve or recover this.

**Pros**:
- Most accurate - uses actual runtime type
- No guessing

**Cons**:
- Requires changing how `Model.Deserialize` is called or how we get type info
- May require passing type information through

**Implementation**:
- Option 2a: Store type information in Model struct (adds overhead)
- Option 2b: Use reflection to get type from the original value before it becomes `*Model`
- Option 2c: Require caller to pass type information

### Option 3: Verify Type by Checking Field Layout
**Approach**: Verify that fields at expected offsets match the expected types.

**Pros**:
- More reliable than size checking
- Verifies actual structure

**Cons**:
- Complex to implement
- Requires knowing field offsets
- Still not 100% foolproof

**Implementation**:
```go
// Check if first few fields match expected types
if !verifyFieldLayout(structType, outerPtr) {
    continue // Wrong type
}
```

### Option 4: Iterate All Registered Models and Find Best Match
**Approach**: Try all registered models and verify which one is correct by checking multiple criteria (size, field count, field types).

**Pros**:
- More robust than single check
- Can use multiple heuristics

**Cons**:
- Slower (iterates all models)
- Still heuristic-based, not guaranteed

**Implementation**:
```go
var bestMatch reflect.Type
var bestScore int
for _, structType := range registeredModels {
    score := calculateMatchScore(structType, outerPtr)
    if score > bestScore {
        bestMatch = structType
        bestScore = score
    }
}
```

### Option 5: Use Type Assertion with Actual Value
**Approach**: Get the actual type by using the original value's type information before it's converted to `*Model`.

**Pros**:
- Most accurate
- Uses Go's type system

**Cons**:
- Requires access to original value type
- May require refactoring how Model.Deserialize works

**Implementation**:
- Need to get type from the original `model` value before `Model.Deserialize` is called
- Could use `reflect.TypeOf(originalValue).Elem()` if we can access it

### Option 6: Store Type Information in Model Struct
**Approach**: Add a field to Model that stores the outer struct type.

**Pros**:
- Direct access to correct type
- No guessing needed

**Cons**:
- Adds memory overhead to every Model
- Requires initialization
- Changes Model struct definition

**Implementation**:
```go
type Model struct {
    outerType reflect.Type // Store the outer struct type
}
```

### Option 7: Use Reflection to Get Type from Interface Value
**Approach**: When the method is called via interface, use reflection to get the actual concrete type.

**Pros**:
- Uses Go's reflection capabilities
- No struct changes needed

**Cons**:
- May not work if called directly (not via interface)
- Complex reflection code

**Implementation**:
```go
// Try to get actual type from the interface value
// This might work if called via ModelInterface
actualType := getActualTypeFromInterface(m)
```

### Option 8: Require Explicit Type Registration with Instance
**Approach**: Change the API so callers must explicitly pass the type or register it with the instance.

**Pros**:
- Explicit and clear
- No guessing

**Cons**:
- Breaking API change
- More verbose for users

**Implementation**:
```go
// Option 8a: Pass type explicitly
func (m *Model) DeserializeWithType[T ModelInterface](row map[string]any) error

// Option 8b: Register type with instance
model.RegisterType[*TestModelNonAddr]()
model.Deserialize(row)
```

## Recommended Approach

**Option 2b** seems most promising: Use reflection to get the actual type from the original value. However, this requires understanding how Go's method dispatch works when calling methods on embedded fields.

When `model.Deserialize(row)` is called where `model` is `*TestModelNonAddr`:
1. Go finds the `Deserialize` method on the embedded `Model`
2. The receiver `m` is `*Model` (pointing to the embedded Model field)
3. We need to recover that `model` was originally `*TestModelNonAddr`

**Potential Solution**:
- Use `reflect.ValueOf(m).Type()` to get `*Model`
- Use the pointer address to find which registered model matches
- But we need a way to verify the match...

**Alternative**: Check struct size AND field count as a heuristic:
```go
// Get actual struct info from memory
actualSize := getStructSizeAt(outerPtr)
actualFieldCount := getFieldCountAt(outerPtr)

// Match against registered models
if structType.Size() == actualSize && structType.NumField() == actualFieldCount {
    // Likely match, but still not guaranteed
}
```

## Current Debug Output Analysis

From the test run:
```
[Model.Deserialize] Checking structType=typedb.InsertedId, actualType=typedb.InsertedId, match=true
```

The `match=true` is misleading because:
- `structType` is `InsertedId` (what we're checking)
- `actualType` is also `InsertedId` (because we just created a pointer to InsertedId)
- But the REAL actual type is `TestModelNonAddr`

The type assertion `outerModelInterface.(ModelInterface)` succeeds for ANY type that implements ModelInterface, so it doesn't help us distinguish.

## Next Steps

1. Investigate if we can get the actual type from the call stack or method receiver
2. Try Option 2b: Get type from original value before conversion
3. Fallback to Option 4: Use multiple heuristics (size + field count + field types)
4. Consider Option 6 if performance allows: Store type in Model struct
