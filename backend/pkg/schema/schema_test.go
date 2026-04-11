package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name:    "valid string schema",
			schema:  Schema{Type: Type{Type: "string"}},
			wantErr: false,
		},
		{
			name: "valid object schema with properties",
			schema: Schema{
				Type: Type{
					Type:       "object",
					Properties: map[string]*Type{"name": {Type: "string"}},
					Required:   []string{"name"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid array schema",
			schema: Schema{
				Type: Type{Type: "array", Items: &Type{Type: "string"}},
			},
			wantErr: false,
		},
		{
			name: "valid integer with bounds",
			schema: Schema{
				Type: Type{Type: "integer", Minimum: 0, Maximum: 100},
			},
			wantErr: false,
		},
		{
			name:    "valid boolean",
			schema:  Schema{Type: Type{Type: "boolean"}},
			wantErr: false,
		},
		{
			name:    "valid number",
			schema:  Schema{Type: Type{Type: "number"}},
			wantErr: false,
		},
		{
			name: "invalid schema with bad ref containing null byte",
			// Null byte (\x00) in URI should be rejected by gojsonschema validator
			schema:  Schema{Type: Type{Ref: "not-a-valid-ref-uri\x00"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.schema.Valid()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchemaGetValidator(t *testing.T) {
	t.Parallel()

	t.Run("valid schema returns validator", func(t *testing.T) {
		t.Parallel()
		s := Schema{Type: Type{Type: "string"}}
		v, err := s.GetValidator()
		require.NoError(t, err)
		assert.NotNil(t, v)
	})

	t.Run("object schema returns validator", func(t *testing.T) {
		t.Parallel()
		s := Schema{
			Type: Type{
				Type:       "object",
				Properties: map[string]*Type{"name": {Type: "string"}},
			},
		}
		v, err := s.GetValidator()
		require.NoError(t, err)
		assert.NotNil(t, v)
	})

	t.Run("invalid schema with bad ref (null byte) fails", func(t *testing.T) {
		t.Parallel()
		// Null byte in URI should be rejected by the validator
		s := Schema{
			Type: Type{Ref: "not-a-valid-ref-uri\x00"},
		}
		_, err := s.GetValidator()
		assert.Error(t, err)
	})
}

func TestSchemaValidateString(t *testing.T) {
	t.Parallel()

	stringSchema := Schema{
		Type: Type{Type: "string", MinLength: 1, MaxLength: 10},
	}

	t.Run("valid string", func(t *testing.T) {
		t.Parallel()
		result, err := stringSchema.ValidateString(`"hello"`)
		require.NoError(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("string too long", func(t *testing.T) {
		t.Parallel()
		result, err := stringSchema.ValidateString(`"this string is way too long"`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("empty string fails min length", func(t *testing.T) {
		t.Parallel()
		result, err := stringSchema.ValidateString(`""`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		_, err := stringSchema.ValidateString("not json")
		assert.Error(t, err)
	})
}

func TestSchemaValidateBytes(t *testing.T) {
	t.Parallel()

	intSchema := Schema{
		Type: Type{Type: "integer", Minimum: 0, Maximum: 100},
	}

	t.Run("valid integer", func(t *testing.T) {
		t.Parallel()
		result, err := intSchema.ValidateBytes([]byte("42"))
		require.NoError(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("integer out of range", func(t *testing.T) {
		t.Parallel()
		result, err := intSchema.ValidateBytes([]byte("200"))
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("wrong type", func(t *testing.T) {
		t.Parallel()
		result, err := intSchema.ValidateBytes([]byte(`"not a number"`))
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("malformed json", func(t *testing.T) {
		t.Parallel()
		_, err := intSchema.ValidateBytes([]byte("{invalid json"))
		assert.Error(t, err)
	})
}

func TestSchemaValidateGo(t *testing.T) {
	t.Parallel()

	objectSchema := Schema{
		Type: Type{
			Type: "object",
			Properties: map[string]*Type{
				"name": {Type: "string"},
				"age":  {Type: "integer", Minimum: 0},
			},
			Required: []string{"name"},
		},
	}

	t.Run("valid object", func(t *testing.T) {
		t.Parallel()
		doc := map[string]interface{}{"name": "Alice", "age": 30}
		result, err := objectSchema.ValidateGo(doc)
		require.NoError(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("missing required field", func(t *testing.T) {
		t.Parallel()
		doc := map[string]interface{}{"age": 25}
		result, err := objectSchema.ValidateGo(doc)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("wrong field type", func(t *testing.T) {
		t.Parallel()
		doc := map[string]interface{}{"name": 12345}
		result, err := objectSchema.ValidateGo(doc)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})
}

func TestSchemaValueScan(t *testing.T) {
	t.Parallel()

	t.Run("value and scan round trip", func(t *testing.T) {
		t.Parallel()
		original := Schema{
			Type: Type{
				Type:       "object",
				Properties: map[string]*Type{"name": {Type: "string"}},
				Required:   []string{"name"},
			},
		}
		val, err := original.Value()
		require.NoError(t, err)

		var scanned Schema
		switch v := val.(type) {
		case string:
			err = scanned.Scan(v)
		case []byte:
			err = scanned.Scan(v)
		default:
			t.Fatalf("unexpected Value() type: %T", val)
		}
		require.NoError(t, err)
		assert.Equal(t, original.Type.Type, scanned.Type.Type)
		assert.Contains(t, scanned.Type.Required, "name")
	})

	t.Run("scan from string", func(t *testing.T) {
		t.Parallel()
		var s Schema
		err := s.Scan(`{"type":"string"}`)
		require.NoError(t, err)
		assert.Equal(t, "string", s.Type.Type)
	})

	t.Run("scan from bytes", func(t *testing.T) {
		t.Parallel()
		var s Schema
		err := s.Scan([]byte(`{"type":"integer"}`))
		require.NoError(t, err)
		assert.Equal(t, "integer", s.Type.Type)
	})

	t.Run("scan unsupported type", func(t *testing.T) {
		t.Parallel()
		var s Schema
		err := s.Scan(12345)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("scan invalid json", func(t *testing.T) {
		t.Parallel()
		var s Schema
		err := s.Scan("not valid json")
		assert.Error(t, err)
	})
}

func TestTypeMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("basic type marshaling", func(t *testing.T) {
		t.Parallel()
		tp := Type{Type: "string", MinLength: 1}
		data, err := json.Marshal(tp)
		require.NoError(t, err)

		var raw map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &raw))
		assert.Equal(t, "string", raw["type"])
		assert.Equal(t, float64(1), raw["minLength"])
	})

	t.Run("object type adds empty properties and required", func(t *testing.T) {
		t.Parallel()
		tp := Type{Type: "object"}
		data, err := json.Marshal(tp)
		require.NoError(t, err)

		var raw map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &raw))
		assert.Equal(t, "object", raw["type"])
		assert.NotNil(t, raw["properties"])
		assert.NotNil(t, raw["required"])
	})

	t.Run("object type with existing properties preserves them", func(t *testing.T) {
		t.Parallel()
		tp := Type{
			Type:       "object",
			Properties: map[string]*Type{"name": {Type: "string"}},
			Required:   []string{"name"},
		}
		data, err := json.Marshal(tp)
		require.NoError(t, err)

		var raw map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &raw))
		props := raw["properties"].(map[string]interface{})
		assert.Contains(t, props, "name")
	})

	t.Run("extended properties included", func(t *testing.T) {
		t.Parallel()
		tp := Type{
			Type:     "string",
			ExtProps: map[string]interface{}{"x-custom": "value"},
		}
		data, err := json.Marshal(tp)
		require.NoError(t, err)

		var raw map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &raw))
		assert.Equal(t, "value", raw["x-custom"])
	})
}

func TestTypeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("basic type unmarshaling", func(t *testing.T) {
		t.Parallel()
		var tp Type
		err := json.Unmarshal([]byte(`{"type":"string","minLength":5}`), &tp)
		require.NoError(t, err)
		assert.Equal(t, "string", tp.Type)
		assert.Equal(t, 5, tp.MinLength)
	})

	t.Run("extended properties extracted", func(t *testing.T) {
		t.Parallel()
		var tp Type
		err := json.Unmarshal([]byte(`{"type":"string","x-custom":"hello","x-other":42}`), &tp)
		require.NoError(t, err)
		assert.Equal(t, "string", tp.Type)
		assert.Equal(t, "hello", tp.ExtProps["x-custom"])
		assert.Equal(t, float64(42), tp.ExtProps["x-other"])
	})

	t.Run("no extended properties", func(t *testing.T) {
		t.Parallel()
		var tp Type
		err := json.Unmarshal([]byte(`{"type":"integer","minimum":0}`), &tp)
		require.NoError(t, err)
		assert.Empty(t, tp.ExtProps)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		var tp Type
		err := json.Unmarshal([]byte("not json"), &tp)
		assert.Error(t, err)
	})

	t.Run("marshal unmarshal round trip preserves extprops", func(t *testing.T) {
		t.Parallel()
		original := Type{
			Type:     "string",
			ExtProps: map[string]interface{}{"x-example": "test"},
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var roundTripped Type
		require.NoError(t, json.Unmarshal(data, &roundTripped))
		assert.Equal(t, "string", roundTripped.Type)
		assert.Equal(t, "test", roundTripped.ExtProps["x-example"])
	})
}

func TestSchemaValidateObjectWithEnum(t *testing.T) {
	t.Parallel()

	enumSchema := Schema{
		Type: Type{
			Type: "string",
			Enum: []interface{}{"red", "green", "blue"},
		},
	}

	t.Run("valid enum value", func(t *testing.T) {
		t.Parallel()
		result, err := enumSchema.ValidateString(`"red"`)
		require.NoError(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("invalid enum value", func(t *testing.T) {
		t.Parallel()
		result, err := enumSchema.ValidateString(`"yellow"`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})
}

func TestSchemaValidateWithPattern(t *testing.T) {
	t.Parallel()

	patternSchema := Schema{
		Type: Type{Type: "string", Pattern: "^[a-z]+$"},
	}

	t.Run("matching pattern", func(t *testing.T) {
		t.Parallel()
		result, err := patternSchema.ValidateString(`"hello"`)
		require.NoError(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("non matching pattern", func(t *testing.T) {
		t.Parallel()
		result, err := patternSchema.ValidateString(`"Hello123"`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})
}

func TestSchemaValidateArray(t *testing.T) {
	t.Parallel()

	arraySchema := Schema{
		Type: Type{
			Type:     "array",
			Items:    &Type{Type: "integer"},
			MinItems: 1,
			MaxItems: 3,
		},
	}

	t.Run("valid array", func(t *testing.T) {
		t.Parallel()
		result, err := arraySchema.ValidateString(`[1, 2, 3]`)
		require.NoError(t, err)
		assert.True(t, result.Valid())
	})

	t.Run("empty array fails min items", func(t *testing.T) {
		t.Parallel()
		result, err := arraySchema.ValidateString(`[]`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("too many items", func(t *testing.T) {
		t.Parallel()
		result, err := arraySchema.ValidateString(`[1, 2, 3, 4]`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})

	t.Run("wrong item type", func(t *testing.T) {
		t.Parallel()
		result, err := arraySchema.ValidateString(`["a", "b"]`)
		require.NoError(t, err)
		assert.False(t, result.Valid())
	})
}

func TestScanFromJSON(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name string `json:"name"`
	}

	t.Run("scan from string", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON(`{"name":"test"}`, &out)
		require.NoError(t, err)
		assert.Equal(t, "test", out.Name)
	})

	t.Run("scan from bytes", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON([]byte(`{"name":"hello"}`), &out)
		require.NoError(t, err)
		assert.Equal(t, "hello", out.Name)
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON(12345, &out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("invalid json string", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON("not json", &out)
		assert.Error(t, err)
	})

	t.Run("invalid json bytes", func(t *testing.T) {
		t.Parallel()
		var out testStruct
		err := scanFromJSON([]byte("{invalid}"), &out)
		assert.Error(t, err)
	})
}
