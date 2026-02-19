package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryValues(t *testing.T) {
	t.Run("empty optional", func(t *testing.T) {
		type nested struct {
			Value *string `json:"value,omitempty" url:"value,omitempty"`
		}
		type example struct {
			Nested *nested `json:"nested,omitempty" url:"nested,omitempty"`
		}

		values, err := QueryValues(&example{})
		require.NoError(t, err)
		assert.Empty(t, values)
	})

	t.Run("empty required", func(t *testing.T) {
		type nested struct {
			Value *string `json:"value,omitempty" url:"value,omitempty"`
		}
		type example struct {
			Required string  `json:"required" url:"required"`
			Nested   *nested `json:"nested,omitempty" url:"nested,omitempty"`
		}

		values, err := QueryValues(&example{})
		require.NoError(t, err)
		assert.Equal(t, "required=", values.Encode())
	})

	t.Run("allow multiple", func(t *testing.T) {
		type example struct {
			Values []string `json:"values" url:"values"`
		}

		values, err := QueryValues(
			&example{
				Values: []string{"foo", "bar", "baz"},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "values=foo&values=bar&values=baz", values.Encode())
	})

	t.Run("nested object", func(t *testing.T) {
		type nested struct {
			Value *string `json:"value,omitempty" url:"value,omitempty"`
		}
		type example struct {
			Required string  `json:"required" url:"required"`
			Nested   *nested `json:"nested,omitempty" url:"nested,omitempty"`
		}

		nestedValue := "nestedValue"
		values, err := QueryValues(
			&example{
				Required: "requiredValue",
				Nested: &nested{
					Value: &nestedValue,
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "nested%5Bvalue%5D=nestedValue&required=requiredValue", values.Encode())
	})

	t.Run("url unspecified", func(t *testing.T) {
		type example struct {
			Required string `json:"required" url:"required"`
			NotFound string `json:"notFound"`
		}

		values, err := QueryValues(
			&example{
				Required: "requiredValue",
				NotFound: "notFound",
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "required=requiredValue", values.Encode())
	})

	t.Run("url ignored", func(t *testing.T) {
		type example struct {
			Required string `json:"required" url:"required"`
			NotFound string `json:"notFound" url:"-"`
		}

		values, err := QueryValues(
			&example{
				Required: "requiredValue",
				NotFound: "notFound",
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "required=requiredValue", values.Encode())
	})

	t.Run("datetime", func(t *testing.T) {
		type example struct {
			DateTime time.Time `json:"dateTime" url:"dateTime"`
		}

		values, err := QueryValues(
			&example{
				DateTime: time.Date(1994, 3, 16, 12, 34, 56, 0, time.UTC),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "dateTime=1994-03-16T12%3A34%3A56Z", values.Encode())
	})

	t.Run("date", func(t *testing.T) {
		type example struct {
			Date time.Time `json:"date" url:"date" format:"date"`
		}

		values, err := QueryValues(
			&example{
				Date: time.Date(1994, 3, 16, 12, 34, 56, 0, time.UTC),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "date=1994-03-16", values.Encode())
	})

	t.Run("optional time", func(t *testing.T) {
		type example struct {
			Date *time.Time `json:"date,omitempty" url:"date,omitempty" format:"date"`
		}

		values, err := QueryValues(
			&example{},
		)
		require.NoError(t, err)
		assert.Empty(t, values.Encode())
	})

	t.Run("omitempty with non-pointer zero value", func(t *testing.T) {
		type enum string

		type example struct {
			Enum enum `json:"enum,omitempty" url:"enum,omitempty"`
		}

		values, err := QueryValues(
			&example{},
		)
		require.NoError(t, err)
		assert.Empty(t, values.Encode())
	})

	t.Run("object array", func(t *testing.T) {
		type object struct {
			Key   string `json:"key" url:"key"`
			Value string `json:"value" url:"value"`
		}
		type example struct {
			Objects []*object `json:"objects,omitempty" url:"objects,omitempty"`
		}

		values, err := QueryValues(
			&example{
				Objects: []*object{
					{
						Key:   "hello",
						Value: "world",
					},
					{
						Key:   "foo",
						Value: "bar",
					},
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "objects%5Bkey%5D=hello&objects%5Bkey%5D=foo&objects%5Bvalue%5D=world&objects%5Bvalue%5D=bar", values.Encode())
	})

	t.Run("map", func(t *testing.T) {
		type request struct {
			Metadata map[string]interface{} `json:"metadata" url:"metadata"`
		}
		values, err := QueryValues(
			&request{
				Metadata: map[string]interface{}{
					"foo": "bar",
					"baz": "qux",
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "metadata%5Bbaz%5D=qux&metadata%5Bfoo%5D=bar", values.Encode())
	})

	t.Run("nested map", func(t *testing.T) {
		type request struct {
			Metadata map[string]interface{} `json:"metadata" url:"metadata"`
		}
		values, err := QueryValues(
			&request{
				Metadata: map[string]interface{}{
					"inner": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "metadata%5Binner%5D%5Bfoo%5D=bar", values.Encode())
	})

	t.Run("nested map array", func(t *testing.T) {
		type request struct {
			Metadata map[string]interface{} `json:"metadata" url:"metadata"`
		}
		values, err := QueryValues(
			&request{
				Metadata: map[string]interface{}{
					"inner": []string{
						"one",
						"two",
						"three",
					},
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "metadata%5Binner%5D=one&metadata%5Binner%5D=two&metadata%5Binner%5D=three", values.Encode())
	})
}

func TestQueryValuesWithDefaults(t *testing.T) {
	t.Run("apply defaults to zero values", func(t *testing.T) {
		type example struct {
			Name    string `json:"name" url:"name"`
			Age     int    `json:"age" url:"age"`
			Enabled bool   `json:"enabled" url:"enabled"`
		}

		defaults := map[string]interface{}{
			"name":    "default-name",
			"age":     25,
			"enabled": true,
		}

		values, err := QueryValuesWithDefaults(&example{}, defaults)
		require.NoError(t, err)
		assert.Equal(t, "age=25&enabled=true&name=default-name", values.Encode())
	})

	t.Run("preserve non-zero values over defaults", func(t *testing.T) {
		type example struct {
			Name    string `json:"name" url:"name"`
			Age     int    `json:"age" url:"age"`
			Enabled bool   `json:"enabled" url:"enabled"`
		}

		defaults := map[string]interface{}{
			"name":    "default-name",
			"age":     25,
			"enabled": true,
		}

		values, err := QueryValuesWithDefaults(&example{
			Name: "actual-name",
			Age:  30,
			// Enabled remains false (zero value), should get default
		}, defaults)
		require.NoError(t, err)
		assert.Equal(t, "age=30&enabled=true&name=actual-name", values.Encode())
	})

	t.Run("ignore defaults for fields not in struct", func(t *testing.T) {
		type example struct {
			Name string `json:"name" url:"name"`
			Age  int    `json:"age" url:"age"`
		}

		defaults := map[string]interface{}{
			"name":        "default-name",
			"age":         25,
			"nonexistent": "should-be-ignored",
		}

		values, err := QueryValuesWithDefaults(&example{}, defaults)
		require.NoError(t, err)
		assert.Equal(t, "age=25&name=default-name", values.Encode())
	})

	t.Run("type conversion for compatible defaults", func(t *testing.T) {
		type example struct {
			Count   int64   `json:"count" url:"count"`
			Rate    float64 `json:"rate" url:"rate"`
			Message string  `json:"message" url:"message"`
		}

		defaults := map[string]interface{}{
			"count":   int(100),     // int -> int64 conversion
			"rate":    float32(2.5), // float32 -> float64 conversion
			"message": "hello",      // string -> string (no conversion needed)
		}

		values, err := QueryValuesWithDefaults(&example{}, defaults)
		require.NoError(t, err)
		assert.Equal(t, "count=100&message=hello&rate=2.5", values.Encode())
	})

	t.Run("mixed with pointer fields and omitempty", func(t *testing.T) {
		type example struct {
			Required string  `json:"required" url:"required"`
			Optional *string `json:"optional,omitempty" url:"optional,omitempty"`
			Count    int     `json:"count,omitempty" url:"count,omitempty"`
		}

		defaultOptional := "default-optional"
		defaults := map[string]interface{}{
			"required": "default-required",
			"optional": &defaultOptional, // pointer type
			"count":    42,
		}

		values, err := QueryValuesWithDefaults(&example{
			Required: "custom-required", // should override default
			// Optional is nil, should get default
			// Count is 0, should get default
		}, defaults)
		require.NoError(t, err)
		assert.Equal(t, "count=42&optional=default-optional&required=custom-required", values.Encode())
	})

	t.Run("override non-zero defaults with explicit zero values", func(t *testing.T) {
		type example struct {
			Name    *string `json:"name" url:"name"`
			Age     *int    `json:"age" url:"age"`
			Enabled *bool   `json:"enabled" url:"enabled"`
		}

		defaults := map[string]interface{}{
			"name":    "default-name",
			"age":     25,
			"enabled": true,
		}

		// first, test that a properly empty request is overridden:
		{
			values, err := QueryValuesWithDefaults(&example{}, defaults)
			require.NoError(t, err)
			assert.Equal(t, "age=25&enabled=true&name=default-name", values.Encode())
		}

		// second, test that a request that contains zeros is not overridden:
		var (
			name    = ""
			age     = 0
			enabled = false
		)
		values, err := QueryValuesWithDefaults(&example{
			Name:    &name,    // explicit empty string should override default
			Age:     &age,     // explicit zero should override default
			Enabled: &enabled, // explicit false should override default
		}, defaults)
		require.NoError(t, err)
		assert.Equal(t, "age=0&enabled=false&name=", values.Encode())
	})

	t.Run("nil input returns empty values", func(t *testing.T) {
		defaults := map[string]any{
			"name": "default-name",
			"age":  25,
		}

		// Test with nil
		values, err := QueryValuesWithDefaults(nil, defaults)
		require.NoError(t, err)
		assert.Empty(t, values)

		// Test with nil pointer
		type example struct {
			Name string `json:"name" url:"name"`
		}
		var nilPtr *example
		values, err = QueryValuesWithDefaults(nilPtr, defaults)
		require.NoError(t, err)
		assert.Empty(t, values)
	})
}
