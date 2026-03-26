package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// MarshalJSONWithExtraProperty marshals the given value to JSON, including the extra property.
func MarshalJSONWithExtraProperty(marshaler interface{}, key string, value interface{}) ([]byte, error) {
	return MarshalJSONWithExtraProperties(marshaler, map[string]interface{}{key: value})
}

// MarshalJSONWithExtraProperties marshals the given value to JSON, including any extra properties.
func MarshalJSONWithExtraProperties(marshaler interface{}, extraProperties map[string]interface{}) ([]byte, error) {
	bytes, err := json.Marshal(marshaler)
	if err != nil {
		return nil, err
	}
	if len(extraProperties) == 0 {
		return bytes, nil
	}
	keys, err := getKeys(marshaler)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if _, ok := extraProperties[key]; ok {
			return nil, fmt.Errorf("cannot add extra property %q because it is already defined on the type", key)
		}
	}
	extraBytes, err := json.Marshal(extraProperties)
	if err != nil {
		return nil, err
	}
	if isEmptyJSON(bytes) {
		if isEmptyJSON(extraBytes) {
			return bytes, nil
		}
		return extraBytes, nil
	}
	result := bytes[:len(bytes)-1]
	result = append(result, ',')
	result = append(result, extraBytes[1:len(extraBytes)-1]...)
	result = append(result, '}')
	return result, nil
}

// ExtractExtraProperties extracts any extra properties from the given value.
func ExtractExtraProperties(bytes []byte, value interface{}, exclude ...string) (map[string]interface{}, error) {
	val := reflect.ValueOf(value)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("value must be non-nil to extract extra properties")
		}
		val = val.Elem()
	}
	if err := json.Unmarshal(bytes, &value); err != nil {
		return nil, err
	}
	var extraProperties map[string]interface{}
	if err := json.Unmarshal(bytes, &extraProperties); err != nil {
		return nil, err
	}
	for i := 0; i < val.Type().NumField(); i++ {
		key := jsonKey(val.Type().Field(i))
		if key == "" || key == "-" {
			continue
		}
		delete(extraProperties, key)
	}
	for _, key := range exclude {
		delete(extraProperties, key)
	}
	if len(extraProperties) == 0 {
		return nil, nil
	}
	return extraProperties, nil
}

// getKeys returns the keys associated with the given value. The value must be a
// a struct or a map with string keys.
func getKeys(value interface{}) ([]string, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if !val.IsValid() {
		return nil, nil
	}
	switch val.Kind() {
	case reflect.Struct:
		return getKeysForStructType(val.Type()), nil
	case reflect.Map:
		var keys []string
		if val.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("cannot extract keys from %T; only structs and maps with string keys are supported", value)
		}
		for _, key := range val.MapKeys() {
			keys = append(keys, key.String())
		}
		return keys, nil
	default:
		return nil, fmt.Errorf("cannot extract keys from %T; only structs and maps with string keys are supported", value)
	}
}

// getKeysForStructType returns all the keys associated with the given struct type,
// visiting embedded fields recursively.
func getKeysForStructType(structType reflect.Type) []string {
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return nil
	}
	var keys []string
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Anonymous {
			keys = append(keys, getKeysForStructType(field.Type)...)
			continue
		}
		keys = append(keys, jsonKey(field))
	}
	return keys
}

// jsonKey returns the JSON key from the struct tag of the given field,
// excluding the omitempty flag (if any).
func jsonKey(field reflect.StructField) string {
	return strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
}

// isEmptyJSON returns true if the given data is empty, the empty JSON object, or
// an explicit null.
func isEmptyJSON(data []byte) bool {
	return len(data) <= 2 || bytes.Equal(data, []byte("null"))
}
