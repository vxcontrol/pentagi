package internal

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	bytesType        = reflect.TypeOf([]byte{})
	queryEncoderType = reflect.TypeOf(new(QueryEncoder)).Elem()
	timeType         = reflect.TypeOf(time.Time{})
	uuidType         = reflect.TypeOf(uuid.UUID{})
)

// QueryEncoder is an interface implemented by any type that wishes to encode
// itself into URL values in a non-standard way.
type QueryEncoder interface {
	EncodeQueryValues(key string, v *url.Values) error
}

// prepareValue handles common validation and unwrapping logic for both functions
func prepareValue(v interface{}) (reflect.Value, url.Values, error) {
	values := make(url.Values)
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return reflect.Value{}, values, nil
		}
		val = val.Elem()
	}

	if v == nil {
		return reflect.Value{}, values, nil
	}

	if val.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("query: Values() expects struct input. Got %v", val.Kind())
	}

	err := reflectValue(values, val, "")
	if err != nil {
		return reflect.Value{}, nil, err
	}

	return val, values, nil
}

// QueryValues encodes url.Values from request objects.
//
// Note: This type is inspired by Google's query encoding library, but
// supports far less customization and is tailored to fit this SDK's use case.
//
// Ref: https://github.com/google/go-querystring
func QueryValues(v interface{}) (url.Values, error) {
	_, values, err := prepareValue(v)
	return values, err
}

// QueryValuesWithDefaults encodes url.Values from request objects
// and default values, merging the defaults into the request.
// It's expected that the values of defaults are wire names.
func QueryValuesWithDefaults(v interface{}, defaults map[string]interface{}) (url.Values, error) {
	val, values, err := prepareValue(v)
	if err != nil {
		return values, err
	}
	if !val.IsValid() {
		return values, nil
	}

	// apply defaults to zero-value fields directly on the original struct
	valType := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := valType.Field(i)
		fieldName := fieldType.Name

		if fieldType.PkgPath != "" && !fieldType.Anonymous {
			// Skip unexported fields.
			continue
		}

		// check if field is zero value and we have a default for it
		if field.CanSet() && field.IsZero() {
			tag := fieldType.Tag.Get("url")
			if tag == "" || tag == "-" {
				continue
			}
			wireName, _ := parseTag(tag)
			if wireName == "" {
				wireName = fieldName
			}
			if defaultVal, exists := defaults[wireName]; exists {
				values.Set(wireName, valueString(reflect.ValueOf(defaultVal), tagOptions{}, reflect.StructField{}))
			}
		}
	}

	return values, err
}

// reflectValue populates the values parameter from the struct fields in val.
// Embedded structs are followed recursively (using the rules defined in the
// Values function documentation) breadth-first.
func reflectValue(values url.Values, val reflect.Value, scope string) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			// Skip unexported fields.
			continue
		}

		sv := val.Field(i)
		tag := sf.Tag.Get("url")
		if tag == "" || tag == "-" {
			continue
		}

		name, opts := parseTag(tag)
		if name == "" {
			name = sf.Name
		}

		if scope != "" {
			name = scope + "[" + name + "]"
		}

		if opts.Contains("omitempty") && isEmptyValue(sv) {
			continue
		}

		if sv.Type().Implements(queryEncoderType) {
			// If sv is a nil pointer and the custom encoder is defined on a non-pointer
			// method receiver, set sv to the zero value of the underlying type
			if !reflect.Indirect(sv).IsValid() && sv.Type().Elem().Implements(queryEncoderType) {
				sv = reflect.New(sv.Type().Elem())
			}

			m := sv.Interface().(QueryEncoder)
			if err := m.EncodeQueryValues(name, &values); err != nil {
				return err
			}
			continue
		}

		// Recursively dereference pointers, but stop at nil pointers.
		for sv.Kind() == reflect.Ptr {
			if sv.IsNil() {
				break
			}
			sv = sv.Elem()
		}

		if sv.Type() == uuidType || sv.Type() == bytesType || sv.Type() == timeType {
			values.Add(name, valueString(sv, opts, sf))
			continue
		}

		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
			if sv.Len() == 0 {
				// Skip if slice or array is empty.
				continue
			}
			for i := 0; i < sv.Len(); i++ {
				value := sv.Index(i)
				if isStructPointer(value) && !value.IsNil() {
					if err := reflectValue(values, value.Elem(), name); err != nil {
						return err
					}
				} else {
					values.Add(name, valueString(value, opts, sf))
				}
			}
			continue
		}

		if sv.Kind() == reflect.Map {
			if err := reflectMap(values, sv, name); err != nil {
				return err
			}
			continue
		}

		if sv.Kind() == reflect.Struct {
			if err := reflectValue(values, sv, name); err != nil {
				return err
			}
			continue
		}

		values.Add(name, valueString(sv, opts, sf))
	}

	return nil
}

// reflectMap handles map types specifically, generating query parameters in the format key[mapkey]=value
func reflectMap(values url.Values, val reflect.Value, scope string) error {
	if val.IsNil() {
		return nil
	}

	iter := val.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		key := fmt.Sprint(k.Interface())
		paramName := scope + "[" + key + "]"

		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				break
			}
			v = v.Elem()
		}

		for v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		if v.Kind() == reflect.Map {
			if err := reflectMap(values, v, paramName); err != nil {
				return err
			}
			continue
		}

		if v.Kind() == reflect.Struct {
			if err := reflectValue(values, v, paramName); err != nil {
				return err
			}
			continue
		}

		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			if v.Len() == 0 {
				continue
			}
			for i := 0; i < v.Len(); i++ {
				value := v.Index(i)
				if isStructPointer(value) && !value.IsNil() {
					if err := reflectValue(values, value.Elem(), paramName); err != nil {
						return err
					}
				} else {
					values.Add(paramName, valueString(value, tagOptions{}, reflect.StructField{}))
				}
			}
			continue
		}

		values.Add(paramName, valueString(v, tagOptions{}, reflect.StructField{}))
	}

	return nil
}

// valueString returns the string representation of a value.
func valueString(v reflect.Value, opts tagOptions, sf reflect.StructField) string {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Type() == timeType {
		t := v.Interface().(time.Time)
		if format := sf.Tag.Get("format"); format == "date" {
			return t.Format("2006-01-02")
		}
		return t.Format(time.RFC3339)
	}

	if v.Type() == uuidType {
		u := v.Interface().(uuid.UUID)
		return u.String()
	}

	if v.Type() == bytesType {
		b := v.Interface().([]byte)
		return base64.StdEncoding.EncodeToString(b)
	}

	return fmt.Sprint(v.Interface())
}

// isEmptyValue checks if a value should be considered empty for the purposes
// of omitting fields with the "omitempty" option.
func isEmptyValue(v reflect.Value) bool {
	type zeroable interface {
		IsZero() bool
	}

	if !v.IsZero() {
		if z, ok := v.Interface().(zeroable); ok {
			return z.IsZero()
		}
	}

	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.Struct, reflect.UnsafePointer:
		return false
	}

	return false
}

// isStructPointer returns true if the given reflect.Value is a pointer to a struct.
func isStructPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct
}

// tagOptions is the string following a comma in a struct field's "url" tag, or
// the empty string. It does not include the leading comma.
type tagOptions []string

// parseTag splits a struct field's url tag into its name and comma-separated
// options.
func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

// Contains checks whether the tagOptions contains the specified option.
func (o tagOptions) Contains(option string) bool {
	for _, s := range o {
		if s == option {
			return true
		}
	}
	return false
}
