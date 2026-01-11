package typedb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// DeserializeForType creates a new instance of T and deserializes the row into it.
// Returns a pointer to the deserialized model.
// T must be a pointer type (e.g., *User).
func DeserializeForType[T ModelInterface](row map[string]any) (T, error) {
	var model T
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Ptr {
		var zero T
		return zero, fmt.Errorf("typedb: DeserializeForType requires a pointer type (e.g., *User)")
	}

	// Create a new instance of the underlying type
	elemType := modelType.Elem()
	modelPtr := reflect.New(elemType)
	modelInterface := modelPtr.Interface().(ModelInterface)

	if err := Deserialize(row, modelInterface); err != nil {
		var zero T
		return zero, err
	}

	return modelPtr.Interface().(T), nil
}

// Deserialize deserializes a row into an existing model.
// Uses reflection to map database column names (from db tags) to struct fields.
func Deserialize(row map[string]any, dest ModelInterface) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: dest must be a pointer type")
	}
	if destValue.IsNil() {
		return fmt.Errorf("typedb: cannot deserialize into nil pointer")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("typedb: dest must be a pointer to struct")
	}

	fieldMap := buildFieldMap(destValue)

	for key, value := range row {
		if value == nil {
			continue
		}

		if fieldValue, ok := fieldMap[key]; ok {
			if err := DeserializeToField(fieldValue.Interface(), value); err != nil {
				return fmt.Errorf("field %s: %w", key, err)
			}
		}
	}

	return nil
}

// buildFieldMap creates a map of database column names to field pointers.
// Handles embedded structs and supports dot notation in db tags (e.g., "users.id").
func buildFieldMap(structValue reflect.Value) map[string]reflect.Value {
	structType := structValue.Type()
	fieldMap := make(map[string]reflect.Value)

	var processFields func(reflect.Type, reflect.Value, []int)
	processFields = func(t reflect.Type, v reflect.Value, indexPath []int) {
		if t.Kind() != reflect.Struct {
			return
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldValue := v.Field(i)
			currentIndex := append(append([]int(nil), indexPath...), i)

			// Handle embedded structs
			if field.Anonymous {
				embeddedType := field.Type
				if embeddedType.Kind() == reflect.Ptr {
					if fieldValue.IsNil() {
						// Initialize pointer embedded struct
						fieldValue.Set(reflect.New(embeddedType.Elem()))
					}
					embeddedType = embeddedType.Elem()
					fieldValue = fieldValue.Elem()
				}
				if embeddedType.Kind() == reflect.Struct {
					processFields(embeddedType, fieldValue, currentIndex)
					continue
				}
			}

			// Get db tag
			dbTag := field.Tag.Get("db")
			if dbTag == "" || dbTag == "-" {
				continue
			}

			// Support dot notation (e.g., "users.id")
			// Use the full tag as the key
			fieldMap[dbTag] = fieldValue.Addr()
		}
	}

	processFields(structType, structValue, nil)
	return fieldMap
}

// DeserializeToField deserializes a value to the appropriate type.
// Handles type conversion for common Go types and uses reflection for complex types.
func DeserializeToField(target any, value any) error {
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

	// Type switch for common types
	switch ptr := target.(type) {
	case *int:
		val, err := DeserializeInt(value)
		if err != nil {
			return err
		}
		*ptr = val
	case *int64:
		val, err := DeserializeInt64(value)
		if err != nil {
			return err
		}
		*ptr = val
	case *int32:
		val, err := DeserializeInt32(value)
		if err != nil {
			return err
		}
		*ptr = val
	case *bool:
		val, err := DeserializeBool(value)
		if err != nil {
			return err
		}
		*ptr = val
	case *string:
		*ptr = DeserializeString(value)
	case *time.Time:
		val, err := DeserializeTime(value)
		if err != nil {
			return err
		}
		*ptr = val
	case **int:
		val, err := DeserializeInt(value)
		if err != nil {
			return err
		}
		*ptr = &val
	case **bool:
		val, err := DeserializeBool(value)
		if err != nil {
			return err
		}
		*ptr = &val
	case **string:
		str := DeserializeString(value)
		*ptr = &str
	case **time.Time:
		val, err := DeserializeTime(value)
		if err != nil {
			return err
		}
		*ptr = &val
	case *[]int:
		arr, err := DeserializeIntArray(value)
		if err != nil {
			return err
		}
		*ptr = arr
	case *[]string:
		arr, err := DeserializeStringArray(value)
		if err != nil {
			return err
		}
		*ptr = arr
	case *map[string]any:
		jsonb, err := DeserializeJSONB(value)
		if err != nil {
			return err
		}
		*ptr = jsonb
	case *map[string]string:
		m, err := DeserializeMap(value)
		if err != nil {
			return err
		}
		*ptr = m
	default:
		// Use reflection for other types
		return deserializeWithReflection(targetValue, targetElem, value)
	}

	return nil
}

// deserializeWithReflection handles complex types using reflection.
func deserializeWithReflection(targetValue reflect.Value, targetElem reflect.Value, value any) error {
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

// DeserializeInt converts a value to int
func DeserializeInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case int16:
		return int(v), nil
	case int8:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint64:
		return int(v), nil
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

// DeserializeInt64 converts a value to int64
func DeserializeInt64(value any) (int64, error) {
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
		return int64(v), nil
	case uint:
		return int64(v), nil
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

// DeserializeInt32 converts a value to int32
func DeserializeInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int32:
		return v, nil
	case int:
		return int32(v), nil
	case int64:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int8:
		return int32(v), nil
	case uint32:
		return int32(v), nil
	case uint:
		return int32(v), nil
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

// DeserializeBool converts a value to bool
func DeserializeBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		if lower == "t" || lower == "true" || lower == "1" {
			return true, nil
		}
		if lower == "f" || lower == "false" || lower == "0" {
			return false, nil
		}
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case int32:
		return v != 0, nil
	default:
		str := fmt.Sprintf("%v", value)
		lower := strings.ToLower(strings.TrimSpace(str))
		if lower == "t" || lower == "true" || lower == "1" {
			return true, nil
		}
		if lower == "f" || lower == "false" || lower == "0" {
			return false, nil
		}
		return strconv.ParseBool(str)
	}
}

// DeserializeString converts a value to string
func DeserializeString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// DeserializeTime converts a value to time.Time
func DeserializeTime(value any) (time.Time, error) {
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

// DeserializeIntArray converts a value to []int
func DeserializeIntArray(value any) ([]int, error) {
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
			val, err := DeserializeInt(item)
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

// DeserializeStringArray converts a value to []string
func DeserializeStringArray(value any) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = DeserializeString(item)
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

// DeserializeJSONB unmarshals a value to map[string]any
func DeserializeJSONB(value any) (map[string]any, error) {
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

// DeserializeMap converts a value to map[string]string
func DeserializeMap(value any) (map[string]string, error) {
	switch v := value.(type) {
	case map[string]string:
		return v, nil
	case map[string]any:
		result := make(map[string]string)
		for key, val := range v {
			result[key] = DeserializeString(val)
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

// SerializeJSONB serializes a Go value to JSONB format (JSON string or bytes).
// Converts map[string]any, map[string]string, or any JSON-marshalable value to JSON.
// Returns the value as-is if it's already a string or []byte.
func SerializeJSONB(value any) (any, error) {
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

// SerializeIntArray serializes a Go slice to PostgreSQL array format.
// Converts []int, []int64, []int32, etc. to PostgreSQL array string "{1,2,3}".
func SerializeIntArray(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	var ints []int
	switch v := value.(type) {
	case []int:
		ints = v
	case []int64:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []int32:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []int16:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []int8:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []uint:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []uint64:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []uint32:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []uint16:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []uint8:
		ints = make([]int, len(v))
		for i, val := range v {
			ints[i] = int(val)
		}
	case []any:
		ints = make([]int, len(v))
		for i, item := range v {
			val, err := DeserializeInt(item)
			if err != nil {
				return "", fmt.Errorf("element %d: %w", i, err)
			}
			ints[i] = val
		}
	default:
		return "", fmt.Errorf("typedb: unsupported type for int array serialization: %T", value)
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

// SerializeStringArray serializes a Go slice to PostgreSQL array format.
// Converts []string or []any to PostgreSQL array string "{a,b,c}".
func SerializeStringArray(value any) (string, error) {
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
			strs[i] = DeserializeString(item)
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

// Serialize converts a Go value to a database-compatible format.
// Handles JSONB, arrays, and other types that need conversion for database operations.
// Returns the value as-is for types that databases handle natively (int, string, bool, time.Time, etc.).
func Serialize(value any) (any, error) {
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
		return SerializeJSONB(value)
	}

	// Handle array types
	switch value.(type) {
	case []int, []int64, []int32, []int16, []int8,
		[]uint, []uint64, []uint32, []uint16, []uint8:
		return SerializeIntArray(value)
	case []string:
		return SerializeStringArray(value)
	}

	// For other types, try JSONB serialization
	return SerializeJSONB(value)
}
