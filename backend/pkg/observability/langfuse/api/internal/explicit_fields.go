package internal

import (
	"math/big"
	"reflect"
	"strings"
)

// HandleExplicitFields processes a struct to remove `omitempty` from
// fields that have been explicitly set (as indicated by their corresponding bit in explicitFields).
// Note that `marshaler` should be an embedded struct to avoid infinite recursion.
// Returns an interface{} that can be passed to json.Marshal.
func HandleExplicitFields(marshaler interface{}, explicitFields *big.Int) interface{} {
	val := reflect.ValueOf(marshaler)
	typ := reflect.TypeOf(marshaler)

	// Handle pointer types
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	// Only handle struct types
	if val.Kind() != reflect.Struct {
		return marshaler
	}

	// Handle embedded struct pattern
	var sourceVal reflect.Value
	var sourceType reflect.Type

	// Check if this is an embedded struct pattern
	if typ.NumField() == 1 && typ.Field(0).Anonymous {
		// This is likely an embedded struct, get the embedded value
		embeddedField := val.Field(0)
		sourceVal = embeddedField
		sourceType = embeddedField.Type()
	} else {
		// Regular struct
		sourceVal = val
		sourceType = typ
	}

	// If no explicit fields set, use standard marshaling
	if explicitFields == nil || explicitFields.Sign() == 0 {
		return marshaler
	}

	// Create a new struct type with modified tags
	fields := make([]reflect.StructField, 0, sourceType.NumField())

	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)

		// Skip unexported fields and the explicitFields field itself
		if !field.IsExported() || field.Name == "explicitFields" {
			continue
		}

		// Check if this field has been explicitly set
		fieldBit := big.NewInt(1)
		fieldBit.Lsh(fieldBit, uint(i))
		if big.NewInt(0).And(explicitFields, fieldBit).Sign() != 0 {
			// Remove omitempty from the json tag
			tag := field.Tag.Get("json")
			if tag != "" && tag != "-" {
				// Parse the json tag, remove omitempty from options
				parts := strings.Split(tag, ",")
				if len(parts) > 1 {
					var newParts []string
					newParts = append(newParts, parts[0]) // Keep the field name
					for _, part := range parts[1:] {
						if strings.TrimSpace(part) != "omitempty" {
							newParts = append(newParts, part)
						}
					}
					tag = strings.Join(newParts, ",")
				}

				// Reconstruct the struct tag
				newTag := `json:"` + tag + `"`
				if urlTag := field.Tag.Get("url"); urlTag != "" {
					newTag += ` url:"` + urlTag + `"`
				}

				field.Tag = reflect.StructTag(newTag)
			}
		}

		fields = append(fields, field)
	}

	// Create new struct type with modified tags
	newType := reflect.StructOf(fields)
	newVal := reflect.New(newType).Elem()

	// Copy field values from original struct to new struct
	fieldIndex := 0
	for i := 0; i < sourceType.NumField(); i++ {
		originalField := sourceType.Field(i)

		// Skip unexported fields and the explicitFields field itself
		if !originalField.IsExported() || originalField.Name == "explicitFields" {
			continue
		}

		originalValue := sourceVal.Field(i)
		newVal.Field(fieldIndex).Set(originalValue)
		fieldIndex++
	}

	return newVal.Interface()
}
