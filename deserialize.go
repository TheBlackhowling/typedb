package typedb

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// deserializeForType creates a new instance of T and deserializes the row into it.
// Returns a pointer to the deserialized model.
// T must be a pointer type (e.g., *User).
// This is an internal function - users should deserialize via Query, InsertAndLoad, etc.
func deserializeForType[T ModelInterface](row map[string]any) (T, error) {
	var model T
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Ptr {
		var zero T
		return zero, fmt.Errorf("typedb: DeserializeForType requires a pointer type (e.g., *User)")
	}

	// Create a new instance of the underlying type
	elemType := modelType.Elem()
	modelPtr := reflect.New(elemType)
	modelInterface, ok := modelPtr.Interface().(ModelInterface)
	if !ok {
		var zero T
		return zero, fmt.Errorf("typedb: type %T does not implement ModelInterface", modelPtr.Interface())
	}

	if err := deserialize(row, modelInterface); err != nil {
		var zero T
		return zero, err
	}

	result, ok := modelPtr.Interface().(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("typedb: failed to convert %T to %T", modelPtr.Interface(), *new(T))
	}

	return result, nil
}

// deserialize deserializes a row into an existing model.
// Uses reflection to map database column names (from db tags) to struct fields.
// This is an internal function - users should deserialize via Query, InsertAndLoad, etc.
func deserialize(row map[string]any, dest ModelInterface) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: dest must be a pointer type")
	}
	if destValue.IsNil() {
		return fmt.Errorf("typedb: cannot deserialize into nil pointer")
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("typedb: dest must be a pointer to struct")
	}

	// Always use buildFieldMapFromPtr to bypass checkptr validation.
	// Even when CanAddr() returns true, values from reflect.NewAt can trigger
	// checkptr errors when accessing fields via reflect.Value.Field().
	// This happens across all Go versions 1.18-1.25, so we always use the unsafe
	// path which bypasses reflect.Value.Field() entirely.
	fieldMap := buildFieldMapFromPtr(destValue, structValue)

	for key, value := range row {
		if value == nil {
			continue
		}

		if fieldValue, ok := fieldMap[key]; ok {
			// Work directly with reflect.Value instead of converting to interface
			// This avoids issues with reflect.NewAt pointers losing type information
			if err := deserializeToFieldValue(fieldValue, value); err != nil {
				return fmt.Errorf("field %s: %w", key, err)
			}
		}
	}

	// Save original copy if partial update is enabled for this model
	if err := saveOriginalCopyIfEnabled(dest); err != nil {
		return fmt.Errorf("typedb: failed to save original copy: %w", err)
	}

	return nil
}

// saveOriginalCopyIfEnabled saves a deep copy of the model if partial update tracking is enabled.
// The copy is stored in the Model.originalCopy field for later comparison during Update operations.
func saveOriginalCopyIfEnabled(model ModelInterface) error {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return nil // Not a valid model pointer
	}

	structType := modelValue.Elem().Type()
	opts := GetModelOptions(structType)
	if !opts.PartialUpdate {
		return nil // Partial update not enabled for this model
	}

	// Create a deep copy of the model
	originalCopy := deepCopyModel(model)
	if originalCopy == nil {
		return fmt.Errorf("failed to create deep copy")
	}

	// Store the copy in the Model.originalCopy field
	// Find the Model field (typically first embedded field)
	structValue := modelValue.Elem()
	if structValue.Kind() == reflect.Struct && structValue.NumField() > 0 {
		// Look for embedded Model field
		for i := 0; i < structValue.NumField(); i++ {
			field := structValue.Type().Field(i)
			if field.Anonymous && field.Type == reflect.TypeOf(Model{}) {
				// Found the Model field, access its originalCopy field
				modelFieldValue := structValue.Field(i)
				// Use unsafe to set unexported field
				modelFieldPtr := unsafe.Pointer(modelFieldValue.UnsafeAddr())
				originalCopyFieldType := field.Type.Field(0) // Model.originalCopy field
				originalCopyFieldPtr := unsafe.Pointer(uintptr(modelFieldPtr) + originalCopyFieldType.Offset)
				*(*interface{})(originalCopyFieldPtr) = originalCopy
				return nil
			}
		}
	}

	return nil
}

// deepCopyModel creates a deep copy of a model using JSON marshaling/unmarshaling.
// This ensures all fields are properly copied, including nested structures.
func deepCopyModel(model ModelInterface) interface{} {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return nil
	}

	// Use JSON marshaling for deep copy
	data, err := json.Marshal(model)
	if err != nil {
		return nil
	}

	// Create a new instance of the same type
	structType := modelValue.Elem().Type()
	newPtr := reflect.New(structType)
	newModel := newPtr.Interface()

	// Unmarshal into the new instance
	if err := json.Unmarshal(data, newModel); err != nil {
		return nil
	}

	return newModel
}

// buildFieldMapFromPtr creates a field map by accessing fields through the pointer
// using unsafe operations. This bypasses reflect.Value.Field() entirely, avoiding
// checkptr validation issues that occur when values come from reflect.NewAt.
// This is used for all deserialization to ensure consistent behavior across
// all Go versions (1.18-1.25).
//
//go:nocheckptr
func buildFieldMapFromPtr(ptrValue, structValue reflect.Value) map[string]reflect.Value {
	structType := structValue.Type()
	fieldMap := make(map[string]reflect.Value)

	// Get the address of the struct that the pointer points to
	// ptrValue is a reflect.Value of a pointer type
	// We can use Pointer() to get the actual pointer value (what the pointer points to)
	// Then convert it to unsafe.Pointer to use for field access
	structAddr := unsafe.Pointer(ptrValue.Pointer())

	var processFields func(reflect.Type, unsafe.Pointer, []int)
	processFields = func(t reflect.Type, basePtr unsafe.Pointer, indexPath []int) {
		if t.Kind() != reflect.Struct {
			return
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			// Calculate field address using unsafe pointer arithmetic
			fieldOffset := field.Offset
			fieldPtr := unsafe.Add(basePtr, fieldOffset)

			// Create reflect.Value for the field using reflect.NewAt
			// This gives us a pointer to the field (*fieldType)
			fieldType := field.Type
			fieldValuePtr := reflect.NewAt(fieldType, fieldPtr)

			currentIndex := make([]int, len(indexPath), len(indexPath)+1)
			copy(currentIndex, indexPath)
			currentIndex = append(currentIndex, i)

			// Handle embedded structs
			if field.Anonymous {
				embeddedType := field.Type
				// For embedded structs, we need the value (not pointer) to check IsNil/Set/Elem
				fieldValue := fieldValuePtr.Elem()
				if embeddedType.Kind() == reflect.Ptr {
					if fieldValue.IsNil() {
						// Initialize pointer embedded struct
						fieldValue.Set(reflect.New(embeddedType.Elem()))
						// Recalculate fieldPtr after setting the value
						fieldPtr = unsafe.Pointer(fieldValue.Pointer())
					} else {
						// Get the address of what the pointer points to
						fieldPtr = unsafe.Pointer(fieldValue.Pointer())
					}
					embeddedType = embeddedType.Elem()
				}
				if embeddedType.Kind() == reflect.Struct {
					// fieldPtr already points to the embedded struct (or its pointer value)
					processFields(embeddedType, fieldPtr, currentIndex)
					continue
				}
			}

			// Get db tag
			dbTag := field.Tag.Get("db")
			if dbTag == "" || dbTag == "-" {
				continue
			}

			// Store the pointer directly (fieldValuePtr is already *fieldType)
			// This matches what buildFieldMap does: fieldValue.Addr() → *fieldType
			fieldMap[dbTag] = fieldValuePtr
		}
	}

	processFields(structType, structAddr, nil)
	return fieldMap
}

// deserializeBasicType handles basic types: *int, *int64, *int32, *bool, *string
func deserializeBasicType(target, value any) error {
	switch ptr := target.(type) {
	case *int:
		val, err := deserializeInt(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	case *int64:
		val, err := deserializeInt64(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	case *int32:
		val, err := deserializeInt32(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	case *bool:
		val, err := deserializeBool(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	case *string:
		*ptr = deserializeString(value)
		return nil
	default:
		return errNotMyType
	}
}

// deserializeUintType handles unsigned integer types: *uint64, *uint32, *uint
func deserializeUintType(target, value any) error {
	switch ptr := target.(type) {
	case *uint64:
		val, err := deserializeUint64(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	case *uint32:
		val, err := deserializeUint32(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	case *uint:
		val, err := deserializeUint(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	default:
		return errNotMyType
	}
}

// deserializePointerType handles pointer-to-pointer types: **int, **bool, **string, **time.Time
func deserializePointerType(target, value any) error {
	switch ptr := target.(type) {
	case **int:
		val, err := deserializeInt(value)
		if err != nil {
			return err
		}
		*ptr = &val
		return nil
	case **bool:
		val, err := deserializeBool(value)
		if err != nil {
			return err
		}
		*ptr = &val
		return nil
	case **string:
		str := deserializeString(value)
		*ptr = &str
		return nil
	case **time.Time:
		val, err := deserializeTime(value)
		if err != nil {
			return err
		}
		*ptr = &val
		return nil
	default:
		return errNotMyType
	}
}

// deserializeTimeType handles time.Time type: *time.Time
func deserializeTimeType(target, value any) error {
	switch ptr := target.(type) {
	case *time.Time:
		val, err := deserializeTime(value)
		if err != nil {
			return err
		}
		*ptr = val
		return nil
	default:
		return errNotMyType
	}
}

// deserializeArrayType handles array types: *[]int, *[]string
func deserializeArrayType(target, value any) error {
	switch ptr := target.(type) {
	case *[]int:
		arr, err := deserializeIntArray(value)
		if err != nil {
			return err
		}
		*ptr = arr
		return nil
	case *[]string:
		arr, err := deserializeStringArray(value)
		if err != nil {
			return err
		}
		*ptr = arr
		return nil
	default:
		return errNotMyType
	}
}

// deserializeMapType handles map types: *map[string]any, *map[string]string
func deserializeMapType(target, value any) error {
	switch ptr := target.(type) {
	case *map[string]any:
		jsonb, err := deserializeJSONB(value)
		if err != nil {
			return err
		}
		*ptr = jsonb
		return nil
	case *map[string]string:
		m, err := deserializeMap(value)
		if err != nil {
			return err
		}
		*ptr = m
		return nil
	default:
		return errNotMyType
	}
}

// deserializeToField deserializes a value to the appropriate type.
// Handles type conversion for common Go types and uses reflection for complex types.
func deserializeToField(target, value any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: target must be a pointer")
	}

	targetElem := targetValue.Elem()
	targetType := targetElem.Type()

	// Handle nil values
	if value == nil {
		// For pointer types, set to nil
		if targetType.Kind() == reflect.Ptr {
			targetElem.Set(reflect.Zero(targetType))
			return nil
		}
		// For non-pointer types, set to zero value
		targetElem.Set(reflect.Zero(targetType))
		return nil
	}

	// Try basic types first
	if err := deserializeBasicType(target, value); err != errNotMyType {
		return err
	}

	// Try uint types
	if err := deserializeUintType(target, value); err != errNotMyType {
		return err
	}

	// Try pointer types
	if err := deserializePointerType(target, value); err != errNotMyType {
		return err
	}

	// Try time type
	if err := deserializeTimeType(target, value); err != errNotMyType {
		return err
	}

	// Try array types
	if err := deserializeArrayType(target, value); err != errNotMyType {
		return err
	}

	// Try map types
	if err := deserializeMapType(target, value); err != errNotMyType {
		return err
	}

	// Fallback to reflection for other types
	return deserializeWithReflection(targetElem, value)
}

// deserializeToFieldValue deserializes a value directly using reflect.Value.
// This avoids issues with reflect.NewAt pointers when converting to interface{}.
func deserializeToFieldValue(fieldValuePtr reflect.Value, value any) error {
	if fieldValuePtr.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: fieldValuePtr must be a pointer")
	}

	fieldElem := fieldValuePtr.Elem()
	fieldType := fieldElem.Type()

	// Handle nil values
	if value == nil {
		if fieldType.Kind() == reflect.Ptr {
			fieldElem.Set(reflect.Zero(fieldType))
			return nil
		}
		fieldElem.Set(reflect.Zero(fieldType))
		return nil
	}

	valueValue := reflect.ValueOf(value)

	// Try direct assignment first
	if valueValue.Type().AssignableTo(fieldType) {
		fieldElem.Set(valueValue)
		return nil
	}

	// Try conversion
	if valueValue.Type().ConvertibleTo(fieldType) {
		fieldElem.Set(valueValue.Convert(fieldType))
		return nil
	}

	// Use the existing DeserializeToField for type-specific handling
	// Convert to interface for the type switch
	return deserializeToField(fieldValuePtr.Interface(), value)
}

// deserializeWithReflection handles complex types using reflection.
func deserializeWithReflection(targetElem reflect.Value, value any) error {
	valueValue := reflect.ValueOf(value)
	valueType := valueValue.Type()
	targetType := targetElem.Type()

	// If types are directly assignable, use assignment
	if valueType.AssignableTo(targetType) {
		targetElem.Set(valueValue)
		return nil
	}

	// If value can be converted to target type, use conversion
	if valueType.ConvertibleTo(targetType) {
		targetElem.Set(valueValue.Convert(targetType))
		return nil
	}

	// Handle pointer types
	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		if valueType.AssignableTo(elemType) || valueType.ConvertibleTo(elemType) {
			ptrValue := reflect.New(elemType)
			if valueType.AssignableTo(elemType) {
				ptrValue.Elem().Set(valueValue)
			} else {
				ptrValue.Elem().Set(valueValue.Convert(elemType))
			}
			targetElem.Set(ptrValue)
			return nil
		}
	}

	return fmt.Errorf("typedb: cannot deserialize %T to %s", value, targetType)
}

// convertInt64ToInt safely converts int64 to int with overflow check
func convertInt64ToInt(v int64) (int, error) {
	maxInt := int64(^uint(0) >> 1)
	minInt := ^maxInt
	if v < minInt || v > maxInt {
		return 0, fmt.Errorf("typedb: int64 value %d overflows int", v)
	}
	return int(v), nil
}

// convertUintToInt safely converts uint to int with overflow check
func convertUintToInt(v uint) (int, error) {
	if v > uint(^uint(0)>>1) {
		return 0, fmt.Errorf("typedb: uint value %d overflows int", v)
	}
	return int(v), nil
}

// convertUint64ToInt safely converts uint64 to int with overflow check
func convertUint64ToInt(v uint64) (int, error) {
	if v > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("typedb: uint64 value %d overflows int", v)
	}
	return int(v), nil
}

// deserializeInt converts a value to int
func deserializeInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return convertInt64ToInt(v)
	case int32:
		return int(v), nil
	case int16:
		return int(v), nil
	case int8:
		return int(v), nil
	case uint:
		return convertUintToInt(v)
	case uint64:
		return convertUint64ToInt(v)
	case uint32:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint8:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return strconv.Atoi(fmt.Sprintf("%v", value))
	}
}

// convertUint64ToInt64 safely converts uint64 to int64 with overflow check
func convertUint64ToInt64(v uint64) (int64, error) {
	if v > uint64(^uint64(0)>>1) {
		return 0, fmt.Errorf("typedb: uint64 value %d overflows int64", v)
	}
	return int64(v), nil
}

// convertUintToInt64 safely converts uint to int64 with overflow check
func convertUintToInt64(v uint) (int64, error) {
	if v > uint(^uint(0)>>1) {
		return 0, fmt.Errorf("typedb: uint value %d overflows int64", v)
	}
	return int64(v), nil
}

// deserializeInt64 converts a value to int64
func deserializeInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint64:
		return convertUint64ToInt64(v)
	case uint:
		return convertUintToInt64(v)
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
	}
}

// convertSignedToUint64 converts signed integers to uint64 with validation
func convertSignedToUint64(value int64) (uint64, error) {
	if value < 0 {
		return 0, fmt.Errorf("typedb: cannot convert negative int64 to uint64")
	}
	return uint64(value), nil
}

// convertFloatToUint64 converts floats to uint64 with validation
func convertFloatToUint64(value float64) (uint64, error) {
	if value < 0 {
		return 0, fmt.Errorf("typedb: cannot convert negative float64 to uint64")
	}
	return uint64(value), nil
}

// deserializeUint64 converts a value to uint64
// Handles MySQL unsigned BIGINT which is returned as string to avoid overflow
func deserializeUint64(value any) (uint64, error) {
	switch v := value.(type) {
	case uint64:
		return v, nil
	case uint:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case int64:
		return convertSignedToUint64(v)
	case int:
		return convertSignedToUint64(int64(v)) //nolint:unconvert // int must be converted to int64 for function signature
	case int32:
		return convertSignedToUint64(int64(v)) //nolint:unconvert // int32 must be converted to int64 for function signature
	case string:
		return strconv.ParseUint(v, 10, 64)
	case float64:
		return convertFloatToUint64(v)
	case float32:
		return convertFloatToUint64(float64(v))
	default:
		return strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
	}
}

// convertUintToUint32 safely converts uint to uint32 with overflow check
func convertUintToUint32(v uint) (uint32, error) {
	if v > uint(^uint32(0)) {
		return 0, fmt.Errorf("typedb: uint value %d overflows uint32", v)
	}
	return uint32(v), nil
}

// convertIntToUint32 safely converts int to uint32 with overflow check
func convertIntToUint32(v int) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("typedb: cannot convert negative int to uint32")
	}
	maxUint32 := int(math.MaxUint32)
	if v > maxUint32 {
		return 0, fmt.Errorf("typedb: int value %d overflows uint32", v)
	}
	return uint32(v), nil
}

// convertInt32ToUint32 safely converts int32 to uint32 with overflow check
func convertInt32ToUint32(v int32) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("typedb: cannot convert negative int32 to uint32")
	}
	return uint32(v), nil
}

// convertFloatToUint32 safely converts float to uint32 with overflow check
func convertFloatToUint32(v float64) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("typedb: cannot convert negative float64 to uint32")
	}
	return uint32(v), nil
}

// deserializeUint32 converts a value to uint32
func deserializeUint32(value any) (uint32, error) {
	switch v := value.(type) {
	case uint32:
		return v, nil
	case uint:
		return convertUintToUint32(v)
	case uint16:
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case int32:
		return convertInt32ToUint32(v)
	case int:
		return convertIntToUint32(v)
	case string:
		val, err := strconv.ParseUint(v, 10, 32)
		return uint32(val), err
	case float64:
		return convertFloatToUint32(v)
	case float32:
		return convertFloatToUint32(float64(v))
	default:
		val, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 32)
		return uint32(val), err
	}
}

// deserializeUint converts a value to uint
func deserializeUint(value any) (uint, error) {
	switch v := value.(type) {
	case uint:
		return v, nil
	case uint64:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint8:
		return uint(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("typedb: cannot convert negative int to uint")
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("typedb: cannot convert negative int64 to uint")
		}
		return uint(v), nil
	case string:
		val, err := strconv.ParseUint(v, 10, 64)
		return uint(val), err
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("typedb: cannot convert negative float64 to uint")
		}
		return uint(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("typedb: cannot convert negative float32 to uint")
		}
		return uint(v), nil
	default:
		val, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
		return uint(val), err
	}
}

// convertIntToInt32 safely converts int to int32 with overflow check
func convertIntToInt32(v int) (int32, error) {
	maxInt32 := int64(math.MaxInt32)
	minInt32 := int64(math.MinInt32)
	v64 := int64(v)
	if v64 < minInt32 || v64 > maxInt32 {
		return 0, fmt.Errorf("typedb: int value %d overflows int32", v)
	}
	return int32(v), nil
}

// convertInt64ToInt32 safely converts int64 to int32 with overflow check
func convertInt64ToInt32(v int64) (int32, error) {
	maxInt32 := int64(math.MaxInt32)
	minInt32 := int64(math.MinInt32)
	if v < minInt32 || v > maxInt32 {
		return 0, fmt.Errorf("typedb: int64 value %d overflows int32", v)
	}
	return int32(v), nil
}

// convertUint32ToInt32 safely converts uint32 to int32 with overflow check
func convertUint32ToInt32(v uint32) (int32, error) {
	if v > uint32(^uint32(0)>>1) {
		return 0, fmt.Errorf("typedb: uint32 value %d overflows int32", v)
	}
	return int32(v), nil
}

// convertUintToInt32 safely converts uint to int32 with overflow check
func convertUintToInt32(v uint) (int32, error) {
	if v > uint(^uint32(0)>>1) {
		return 0, fmt.Errorf("typedb: uint value %d overflows int32", v)
	}
	return int32(v), nil
}

// deserializeInt32 converts a value to int32
func deserializeInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int32:
		return v, nil
	case int:
		return convertIntToInt32(v)
	case int64:
		return convertInt64ToInt32(v)
	case int16:
		return int32(v), nil
	case int8:
		return int32(v), nil
	case uint32:
		return convertUint32ToInt32(v)
	case uint:
		return convertUintToInt32(v)
	case uint16:
		return int32(v), nil
	case uint8:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case float32:
		return int32(v), nil
	case string:
		val, err := strconv.ParseInt(v, 10, 32)
		return int32(val), err
	default:
		val, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 32)
		return int32(val), err
	}
}

// parseBoolString parses various string formats to bool
func parseBoolString(s string) (bool, error) {
	lower := strings.ToLower(strings.TrimSpace(s))

	// Common true values
	if lower == "t" || lower == "true" || lower == "1" {
		return true, nil
	}

	// Common false values
	if lower == "f" || lower == "false" || lower == "0" {
		return false, nil
	}

	// Fallback to strconv.ParseBool
	return strconv.ParseBool(s)
}

// deserializeBool converts a value to bool
func deserializeBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return parseBoolString(v)
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case int32:
		return v != 0, nil
	default:
		// Convert to string and parse
		return parseBoolString(fmt.Sprintf("%v", value))
	}
}

// deserializeString converts a value to string
func deserializeString(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}

// deserializeTime converts a value to time.Time
func deserializeTime(value any) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseTime(v)
	case []byte:
		return parseTime(string(v))
	default:
		return parseTime(fmt.Sprintf("%v", value))
	}
}

// deserializeIntArray converts a value to []int
func deserializeIntArray(value any) ([]int, error) {
	switch v := value.(type) {
	case []int:
		return v, nil
	case []int64:
		result := make([]int, len(v))
		for i, val := range v {
			result[i] = int(val)
		}
		return result, nil
	case []int32:
		result := make([]int, len(v))
		for i, val := range v {
			result[i] = int(val)
		}
		return result, nil
	case []any:
		result := make([]int, len(v))
		for i, item := range v {
			val, err := deserializeInt(item)
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			result[i] = val
		}
		return result, nil
	case string:
		// Handle PostgreSQL array format "{1,2,3}"
		v = strings.Trim(v, "{}")
		if v == "" {
			return []int{}, nil
		}
		parts := strings.Split(v, ",")
		result := make([]int, len(parts))
		for i, part := range parts {
			val, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			result[i] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported type for int array: %T", value)
	}
}

// deserializeStringArray converts a value to []string
func deserializeStringArray(value any) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = deserializeString(item)
		}
		return result, nil
	case string:
		// Handle PostgreSQL array format "{a,b,c}"
		v = strings.Trim(v, "{}")
		if v == "" {
			return []string{}, nil
		}
		parts := strings.Split(v, ",")
		result := make([]string, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported type for string array: %T", value)
	}
}

// deserializeJSONB unmarshals a value to map[string]any
func deserializeJSONB(value any) (map[string]any, error) {
	switch v := value.(type) {
	case map[string]any:
		return v, nil
	case string:
		var result map[string]any
		err := json.Unmarshal([]byte(v), &result)
		return result, err
	case []byte:
		var result map[string]any
		err := json.Unmarshal(v, &result)
		return result, err
	default:
		return nil, fmt.Errorf("unsupported type for JSONB: %T", value)
	}
}

// deserializeMap converts a value to map[string]string
func deserializeMap(value any) (map[string]string, error) {
	switch v := value.(type) {
	case map[string]string:
		return v, nil
	case map[string]any:
		result := make(map[string]string)
		for key, val := range v {
			result[key] = deserializeString(val)
		}
		return result, nil
	case string:
		var result map[string]string
		err := json.Unmarshal([]byte(v), &result)
		return result, err
	case []byte:
		var result map[string]string
		err := json.Unmarshal(v, &result)
		return result, err
	default:
		return nil, fmt.Errorf("unsupported type for map: %T", value)
	}
}

// parseTime tries to parse a string into time.Time using common formats
func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try common time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999999999",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Return error if all formats fail
	return time.Time{}, fmt.Errorf("unable to parse time: %q", s)
}

// serializeJSONB serializes a Go value to JSON format (JSON string or bytes).
// Converts map[string]any, map[string]string, or any JSON-marshalable value to JSON.
// Returns the value as-is if it's already a string or []byte.
//
// Note: While named "JSONB" (reflecting PostgreSQL's JSONB type), this function
// produces standard JSON strings that are compatible with JSON columns in MySQL,
// SQL Server, and other databases that support JSON types.
func serializeJSONB(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		// Already a JSON string
		return v, nil
	case []byte:
		// Already JSON bytes
		return v, nil
	case map[string]any, map[string]string, []any, []string, []int:
		// JSON-marshalable types
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("typedb: failed to marshal JSONB: %w", err)
		}
		return string(jsonBytes), nil
	default:
		// Try to marshal any other type
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("typedb: failed to marshal JSONB: %w", err)
		}
		return string(jsonBytes), nil
	}
}

// convertInt64Slice converts []int64 to []int
func convertInt64Slice(v []int64) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertInt32Slice converts []int32 to []int
func convertInt32Slice(v []int32) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertInt16Slice converts []int16 to []int
func convertInt16Slice(v []int16) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertInt8Slice converts []int8 to []int
func convertInt8Slice(v []int8) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertUintSlice converts []uint to []int
func convertUintSlice(v []uint) []int {
	result := make([]int, len(v))
	for i, val := range v {
		if val > uint(^uint(0)>>1) {
			// Skip overflow values or use 0 - this shouldn't happen in practice
			result[i] = 0
			continue
		}
		result[i] = int(val)
	}
	return result
}

// convertUint64Slice converts []uint64 to []int
func convertUint64Slice(v []uint64) []int {
	result := make([]int, len(v))
	for i, val := range v {
		if val > uint64(^uint(0)>>1) {
			// Skip overflow values or use 0 - this shouldn't happen in practice
			result[i] = 0
			continue
		}
		result[i] = int(val)
	}
	return result
}

// convertUint32Slice converts []uint32 to []int
func convertUint32Slice(v []uint32) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertUint16Slice converts []uint16 to []int
func convertUint16Slice(v []uint16) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertUint8Slice converts []uint8 to []int
func convertUint8Slice(v []uint8) []int {
	result := make([]int, len(v))
	for i, val := range v {
		result[i] = int(val)
	}
	return result
}

// convertAnySlice converts []any to []int by deserializing each element
func convertAnySlice(v []any) ([]int, error) {
	result := make([]int, len(v))
	for i, item := range v {
		val, err := deserializeInt(item)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result[i] = val
	}
	return result, nil
}

// convertToIntSlice converts various integer slice types to []int
func convertToIntSlice(value any) ([]int, error) {
	switch v := value.(type) {
	case []int:
		return v, nil
	case []int64:
		return convertInt64Slice(v), nil
	case []int32:
		return convertInt32Slice(v), nil
	case []int16:
		return convertInt16Slice(v), nil
	case []int8:
		return convertInt8Slice(v), nil
	case []uint:
		return convertUintSlice(v), nil
	case []uint64:
		return convertUint64Slice(v), nil
	case []uint32:
		return convertUint32Slice(v), nil
	case []uint16:
		return convertUint16Slice(v), nil
	case []uint8:
		return convertUint8Slice(v), nil
	case []any:
		return convertAnySlice(v)
	default:
		return nil, fmt.Errorf("typedb: unsupported type for int array serialization: %T", value)
	}
}

// serializeIntArray serializes a Go slice to PostgreSQL array format.
// Converts []int, []int64, []int32, etc. to PostgreSQL array string "{1,2,3}".
//
// Note: This function is PostgreSQL-specific. For other databases, handle arrays
// directly in your SQL queries (e.g., using JSON, comma-separated values, or
// database-specific array syntax).
func serializeIntArray(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	ints, err := convertToIntSlice(value)
	if err != nil {
		return "", err
	}

	if len(ints) == 0 {
		return "{}", nil
	}

	parts := make([]string, len(ints))
	for i, val := range ints {
		parts[i] = strconv.Itoa(val)
	}

	return "{" + strings.Join(parts, ",") + "}", nil
}

// serializeStringArray serializes a Go slice to PostgreSQL array format.
// Converts []string or []any to PostgreSQL array string "{a,b,c}".
//
// Note: This function is PostgreSQL-specific. For other databases, handle arrays
// directly in your SQL queries (e.g., using JSON, comma-separated values, or
// database-specific array syntax).
func serializeStringArray(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	var strs []string
	switch v := value.(type) {
	case []string:
		strs = v
	case []any:
		strs = make([]string, len(v))
		for i, item := range v {
			strs[i] = deserializeString(item)
		}
	default:
		return "", fmt.Errorf("typedb: unsupported type for string array serialization: %T", value)
	}

	if len(strs) == 0 {
		return "{}", nil
	}

	// Escape strings that contain commas, quotes, or backslashes
	parts := make([]string, len(strs))
	for i, s := range strs {
		// PostgreSQL array format: escape quotes and backslashes, quote if contains special chars
		escaped := strings.ReplaceAll(s, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		if strings.ContainsAny(escaped, `,"{}\`) {
			parts[i] = `"` + escaped + `"`
		} else {
			parts[i] = escaped
		}
	}

	return "{" + strings.Join(parts, ",") + "}", nil
}

// serialize converts a Go value to a database-compatible format.
// Handles JSON, arrays, and other types that need conversion for database operations.
// Returns the value as-is for types that databases handle natively (int, string, bool, time.Time, etc.).
//
// Note: Array serialization uses PostgreSQL array format. For other databases,
// handle arrays directly in your SQL queries or use database-specific serialization.
func serialize(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	// Check if it's already a database-compatible type
	switch value.(type) {
	case int, int64, int32, int16, int8,
		uint, uint64, uint32, uint16, uint8,
		float64, float32,
		bool,
		string,
		time.Time,
		[]byte:
		return value, nil
	}

	// Handle JSONB types
	switch value.(type) {
	case map[string]any, map[string]string:
		return serializeJSONB(value)
	}

	// Handle array types
	switch value.(type) {
	case []int, []int64, []int32, []int16, []int8,
		[]uint, []uint64, []uint32, []uint16, []uint8:
		return serializeIntArray(value)
	case []string:
		return serializeStringArray(value)
	}

	// For other types, try JSONB serialization
	return serializeJSONB(value)
}
