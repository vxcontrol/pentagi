package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schemaBadRef carries a $ref with a control character, which gojsonschema refuses to compile.
var schemaBadRef = Schema{Type: Type{Ref: "not-a-valid-ref-uri\x00"}}

const schemaBadRefError = "invalid control character in URL"

func TestSchema_Valid_RejectsASchemaThatDoesNotCompile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schema  Schema
		wantErr string
	}{
		{name: "a string schema", schema: Schema{Type: Type{Type: "string"}}},
		{name: "an object with properties and required", schema: Schema{Type: Type{
			Type:       "object",
			Properties: map[string]*Type{"name": {Type: "string"}},
			Required:   []string{"name"},
		}}},
		{name: "a ref that is not a uri", schema: schemaBadRef, wantErr: schemaBadRefError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.schema.Valid()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSchema_GetValidator_CompilesTheSchema(t *testing.T) {
	t.Parallel()

	v, err := Schema{Type: Type{Type: "string"}}.GetValidator()
	require.NoError(t, err)
	assert.NotNil(t, v)

	_, err = schemaBadRef.GetValidator()
	assert.ErrorContains(t, err, schemaBadRefError)
}

// Each keyword reaches the compiler only through MarshalJSON, so a row per keyword pins both.
func TestSchema_ValidateString_EnforcesStringAndArrayKeywords(t *testing.T) {
	t.Parallel()

	length := Schema{Type: Type{Type: "string", MinLength: 1, MaxLength: 10}}
	enum := Schema{Type: Type{Type: "string", Enum: []interface{}{"red", "green", "blue"}}}
	pattern := Schema{Type: Type{Type: "string", Pattern: "^[a-z]+$"}}
	array := Schema{Type: Type{Type: "array", Items: &Type{Type: "integer"}, MinItems: 1, MaxItems: 3}}

	tests := []struct {
		name   string
		schema Schema
		doc    string
		valid  bool
	}{
		{"a string within its length", length, `"hello"`, true},
		{"a string over maxLength", length, `"this string is way too long"`, false},
		{"an empty string under minLength", length, `""`, false},
		{"an enum member", enum, `"red"`, true},
		{"a value outside the enum", enum, `"yellow"`, false},
		{"a string matching the pattern", pattern, `"hello"`, true},
		{"a string off the pattern", pattern, `"Hello123"`, false},
		{"an array within its bounds", array, `[1, 2, 3]`, true},
		{"an empty array under minItems", array, `[]`, false},
		{"an array over maxItems", array, `[1, 2, 3, 4]`, false},
		{"an array with a wrong item type", array, `["a", "b"]`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := tt.schema.ValidateString(tt.doc)
			require.NoError(t, err)
			assert.Equal(t, tt.valid, result.Valid())
		})
	}

	t.Run("a document that is not json is an error", func(t *testing.T) {
		t.Parallel()
		_, err := length.ValidateString("not json")
		assert.ErrorContains(t, err, "invalid character")
	})

	t.Run("a schema that does not compile is an error", func(t *testing.T) {
		t.Parallel()
		_, err := schemaBadRef.ValidateString(`"x"`)
		assert.ErrorContains(t, err, schemaBadRefError)
	})
}

func TestSchema_ValidateBytes_EnforcesTypeAndBounds(t *testing.T) {
	t.Parallel()

	intSchema := Schema{Type: Type{Type: "integer", Minimum: 0, Maximum: 100}}

	for doc, valid := range map[string]bool{"42": true, "200": false, `"not a number"`: false} {
		result, err := intSchema.ValidateBytes([]byte(doc))
		require.NoError(t, err, doc)
		assert.Equal(t, valid, result.Valid(), doc)
	}

	_, err := intSchema.ValidateBytes([]byte("{invalid json"))
	assert.ErrorContains(t, err, "invalid character")
}

func TestSchema_ValidateGo_EnforcesRequiredAndFieldTypes(t *testing.T) {
	t.Parallel()

	objectSchema := Schema{Type: Type{
		Type: "object",
		Properties: map[string]*Type{
			"name": {Type: "string"},
			"age":  {Type: "integer", Minimum: 0},
		},
		Required: []string{"name"},
	}}

	tests := []struct {
		name  string
		doc   map[string]interface{}
		valid bool
	}{
		{"a valid object", map[string]interface{}{"name": "Alice", "age": 30}, true},
		{"a missing required field", map[string]interface{}{"age": 25}, false},
		{"a field of the wrong type", map[string]interface{}{"name": 12345}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := objectSchema.ValidateGo(tt.doc)
			require.NoError(t, err)
			assert.Equal(t, tt.valid, result.Valid())
		})
	}
}

func TestSchema_Value_RoundTripsThroughScan(t *testing.T) {
	t.Parallel()

	original := Schema{Type: Type{
		Type:       "object",
		Properties: map[string]*Type{"name": {Type: "string"}},
		Required:   []string{"name"},
	}}
	val, err := original.Value()
	require.NoError(t, err)
	require.IsType(t, "", val)

	var scanned Schema
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, "object", scanned.Type.Type)
	assert.Equal(t, []string{"name"}, scanned.Required)
}

func TestSchema_Scan_ReadsStringsAndBytesOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		wantType string
		wantErr  string
	}{
		{name: "a json string", input: `{"type":"string"}`, wantType: "string"},
		{name: "json bytes", input: []byte(`{"type":"integer"}`), wantType: "integer"},
		{name: "an unsupported type", input: 12345, wantErr: "unsupported type"},
		{name: "a string that is not json", input: "not valid json", wantErr: "invalid character"},
		{name: "bytes that are not json", input: []byte("{invalid}"), wantErr: "invalid character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			err := s.Scan(tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, s.Type.Type)
		})
	}
}

func TestSchema_MarshalJSON_MergesExtPropsAndFillsObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tp   Type
		want map[string]interface{}
	}{
		{
			name: "declared fields",
			tp:   Type{Type: "string", MinLength: 1},
			want: map[string]interface{}{"type": "string", "minLength": float64(1)},
		},
		{
			name: "an object gets empty properties and required",
			tp:   Type{Type: "object"},
			want: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []interface{}{}},
		},
		{
			name: "an object keeps its own properties",
			tp:   Type{Type: "object", Properties: map[string]*Type{"name": {Type: "string"}}, Required: []string{"name"}},
			want: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
				"required":   []interface{}{"name"},
			},
		},
		{
			name: "extended properties",
			tp:   Type{Type: "string", ExtProps: map[string]interface{}{"x-custom": "value"}},
			want: map[string]interface{}{"type": "string", "x-custom": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.tp)
			require.NoError(t, err)

			var raw map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &raw))
			assert.Equal(t, tt.want, raw)
		})
	}
}

func TestSchema_UnmarshalJSON_KeepsUnknownKeysAsExtProps(t *testing.T) {
	t.Parallel()

	t.Run("declared fields", func(t *testing.T) {
		t.Parallel()
		var tp Type
		require.NoError(t, json.Unmarshal([]byte(`{"type":"string","minLength":5}`), &tp))
		assert.Equal(t, "string", tp.Type)
		assert.Equal(t, 5, tp.MinLength)
		assert.Empty(t, tp.ExtProps)
	})

	t.Run("unknown keys", func(t *testing.T) {
		t.Parallel()
		var tp Type
		require.NoError(t, json.Unmarshal([]byte(`{"type":"string","x-custom":"hello","x-other":42}`), &tp))
		assert.Equal(t, "string", tp.Type)
		assert.Equal(t, map[string]interface{}{"x-custom": "hello", "x-other": float64(42)}, tp.ExtProps)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		var tp Type
		assert.ErrorContains(t, json.Unmarshal([]byte("not json"), &tp), "invalid character")
	})

	t.Run("a marshal round trip keeps extended properties", func(t *testing.T) {
		t.Parallel()
		data, err := json.Marshal(Type{Type: "string", ExtProps: map[string]interface{}{"x-example": "test"}})
		require.NoError(t, err)

		var roundTripped Type
		require.NoError(t, json.Unmarshal(data, &roundTripped))
		assert.Equal(t, "string", roundTripped.Type)
		assert.Equal(t, "test", roundTripped.ExtProps["x-example"])
	})
}
