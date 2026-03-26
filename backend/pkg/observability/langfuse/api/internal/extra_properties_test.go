package internal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMarshaler struct {
	Name      string    `json:"name"`
	BirthDate time.Time `json:"birthDate"`
	CreatedAt time.Time `json:"created_at"`
}

func (t *testMarshaler) MarshalJSON() ([]byte, error) {
	type embed testMarshaler
	var marshaler = struct {
		embed
		BirthDate string `json:"birthDate"`
		CreatedAt string `json:"created_at"`
	}{
		embed:     embed(*t),
		BirthDate: t.BirthDate.Format("2006-01-02"),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
	return MarshalJSONWithExtraProperty(marshaler, "type", "test")
}

func TestMarshalJSONWithExtraProperties(t *testing.T) {
	tests := []struct {
		desc                string
		giveMarshaler       interface{}
		giveExtraProperties map[string]interface{}
		wantBytes           []byte
		wantError           string
	}{
		{
			desc:                "invalid type",
			giveMarshaler:       []string{"invalid"},
			giveExtraProperties: map[string]interface{}{"key": "overwrite"},
			wantError:           `cannot extract keys from []string; only structs and maps with string keys are supported`,
		},
		{
			desc:                "invalid key type",
			giveMarshaler:       map[int]interface{}{42: "value"},
			giveExtraProperties: map[string]interface{}{"key": "overwrite"},
			wantError:           `cannot extract keys from map[int]interface {}; only structs and maps with string keys are supported`,
		},
		{
			desc:                "invalid map overwrite",
			giveMarshaler:       map[string]interface{}{"key": "value"},
			giveExtraProperties: map[string]interface{}{"key": "overwrite"},
			wantError:           `cannot add extra property "key" because it is already defined on the type`,
		},
		{
			desc:                "invalid struct overwrite",
			giveMarshaler:       new(testMarshaler),
			giveExtraProperties: map[string]interface{}{"birthDate": "2000-01-01"},
			wantError:           `cannot add extra property "birthDate" because it is already defined on the type`,
		},
		{
			desc:                "invalid struct overwrite embedded type",
			giveMarshaler:       new(testMarshaler),
			giveExtraProperties: map[string]interface{}{"name": "bob"},
			wantError:           `cannot add extra property "name" because it is already defined on the type`,
		},
		{
			desc:                "nil",
			giveMarshaler:       nil,
			giveExtraProperties: nil,
			wantBytes:           []byte(`null`),
		},
		{
			desc:                "empty",
			giveMarshaler:       map[string]interface{}{},
			giveExtraProperties: map[string]interface{}{},
			wantBytes:           []byte(`{}`),
		},
		{
			desc:                "no extra properties",
			giveMarshaler:       map[string]interface{}{"key": "value"},
			giveExtraProperties: map[string]interface{}{},
			wantBytes:           []byte(`{"key":"value"}`),
		},
		{
			desc:                "only extra properties",
			giveMarshaler:       map[string]interface{}{},
			giveExtraProperties: map[string]interface{}{"key": "value"},
			wantBytes:           []byte(`{"key":"value"}`),
		},
		{
			desc:                "single extra property",
			giveMarshaler:       map[string]interface{}{"key": "value"},
			giveExtraProperties: map[string]interface{}{"extra": "property"},
			wantBytes:           []byte(`{"key":"value","extra":"property"}`),
		},
		{
			desc:                "multiple extra properties",
			giveMarshaler:       map[string]interface{}{"key": "value"},
			giveExtraProperties: map[string]interface{}{"one": 1, "two": 2},
			wantBytes:           []byte(`{"key":"value","one":1,"two":2}`),
		},
		{
			desc:          "nested properties",
			giveMarshaler: map[string]interface{}{"key": "value"},
			giveExtraProperties: map[string]interface{}{
				"user": map[string]interface{}{
					"age":  42,
					"name": "alice",
				},
			},
			wantBytes: []byte(`{"key":"value","user":{"age":42,"name":"alice"}}`),
		},
		{
			desc:          "multiple nested properties",
			giveMarshaler: map[string]interface{}{"key": "value"},
			giveExtraProperties: map[string]interface{}{
				"metadata": map[string]interface{}{
					"ip": "127.0.0.1",
				},
				"user": map[string]interface{}{
					"age":  42,
					"name": "alice",
				},
			},
			wantBytes: []byte(`{"key":"value","metadata":{"ip":"127.0.0.1"},"user":{"age":42,"name":"alice"}}`),
		},
		{
			desc: "custom marshaler",
			giveMarshaler: &testMarshaler{
				Name:      "alice",
				BirthDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			giveExtraProperties: map[string]interface{}{
				"extra": "property",
			},
			wantBytes: []byte(`{"name":"alice","birthDate":"2000-01-01","created_at":"2024-01-01T00:00:00Z","type":"test","extra":"property"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			bytes, err := MarshalJSONWithExtraProperties(tt.giveMarshaler, tt.giveExtraProperties)
			if tt.wantError != "" {
				require.EqualError(t, err, tt.wantError)
				assert.Nil(t, tt.wantBytes)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBytes, bytes)

			value := make(map[string]interface{})
			require.NoError(t, json.Unmarshal(bytes, &value))
		})
	}
}

func TestExtractExtraProperties(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		type user struct {
			Name string `json:"name"`
		}
		value := &user{
			Name: "alice",
		}
		extraProperties, err := ExtractExtraProperties([]byte(`{"name": "alice"}`), value)
		require.NoError(t, err)
		assert.Nil(t, extraProperties)
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		type user struct {
			Name string `json:"name"`
		}
		value := &user{
			Name: "alice",
		}
		extraProperties, err := ExtractExtraProperties([]byte(`{"name": "alice", "age": 42}`), value)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"age": float64(42)}, extraProperties)
	})

	t.Run("nil pointer", func(t *testing.T) {
		type user struct {
			Name string `json:"name"`
		}
		var value *user
		_, err := ExtractExtraProperties([]byte(`{"name": "alice", "age": 42}`), value)
		assert.EqualError(t, err, "value must be non-nil to extract extra properties")
	})

	t.Run("non-zero value", func(t *testing.T) {
		type user struct {
			Name string `json:"name"`
		}
		value := user{
			Name: "alice",
		}
		extraProperties, err := ExtractExtraProperties([]byte(`{"name": "alice", "age": 42}`), value)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"age": float64(42)}, extraProperties)
	})

	t.Run("zero value", func(t *testing.T) {
		type user struct {
			Name string `json:"name"`
		}
		var value user
		extraProperties, err := ExtractExtraProperties([]byte(`{"name": "alice", "age": 42}`), value)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"age": float64(42)}, extraProperties)
	})

	t.Run("exclude", func(t *testing.T) {
		type user struct {
			Name string `json:"name"`
		}
		value := &user{
			Name: "alice",
		}
		extraProperties, err := ExtractExtraProperties([]byte(`{"name": "alice", "age": 42}`), value, "age")
		require.NoError(t, err)
		assert.Nil(t, extraProperties)
	})
}
