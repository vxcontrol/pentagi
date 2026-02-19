package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"pentagi/pkg/observability/langfuse/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// InternalTestCase represents a single test case.
type InternalTestCase struct {
	description string

	// Server-side assertions.
	givePathSuffix         string
	giveMethod             string
	giveResponseIsOptional bool
	giveHeader             http.Header
	giveErrorDecoder       ErrorDecoder
	giveRequest            *InternalTestRequest
	giveQueryParams        url.Values
	giveBodyProperties     map[string]interface{}

	// Client-side assertions.
	wantResponse *InternalTestResponse
	wantHeaders  http.Header
	wantError    error
}

// InternalTestRequest a simple request body.
type InternalTestRequest struct {
	Id string `json:"id"`
}

// InternalTestResponse a simple response body.
type InternalTestResponse struct {
	Id                  string                 `json:"id"`
	ExtraBodyProperties map[string]interface{} `json:"extraBodyProperties,omitempty"`
	QueryParameters     url.Values             `json:"queryParameters,omitempty"`
}

// InternalTestNotFoundError represents a 404.
type InternalTestNotFoundError struct {
	*core.APIError

	Message string `json:"message"`
}

func TestCall(t *testing.T) {
	tests := []*InternalTestCase{
		{
			description: "GET success",
			giveMethod:  http.MethodGet,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: &InternalTestRequest{
				Id: "123",
			},
			wantResponse: &InternalTestResponse{
				Id: "123",
			},
		},
		{
			description:    "GET success with query",
			givePathSuffix: "?limit=1",
			giveMethod:     http.MethodGet,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: &InternalTestRequest{
				Id: "123",
			},
			wantResponse: &InternalTestResponse{
				Id: "123",
				QueryParameters: url.Values{
					"limit": []string{"1"},
				},
			},
		},
		{
			description: "GET not found",
			giveMethod:  http.MethodGet,
			giveHeader: http.Header{
				"X-API-Status": []string{"fail"},
			},
			giveRequest: &InternalTestRequest{
				Id: strconv.Itoa(http.StatusNotFound),
			},
			giveErrorDecoder: newTestErrorDecoder(t),
			wantError: &InternalTestNotFoundError{
				APIError: core.NewAPIError(
					http.StatusNotFound,
					http.Header{},
					errors.New(`{"message":"ID \"404\" not found"}`),
				),
			},
		},
		{
			description: "POST empty body",
			giveMethod:  http.MethodPost,
			giveHeader: http.Header{
				"X-API-Status": []string{"fail"},
			},
			giveRequest: nil,
			wantError: core.NewAPIError(
				http.StatusBadRequest,
				http.Header{},
				errors.New("invalid request"),
			),
		},
		{
			description: "POST optional response",
			giveMethod:  http.MethodPost,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: &InternalTestRequest{
				Id: "123",
			},
			giveResponseIsOptional: true,
		},
		{
			description: "POST API error",
			giveMethod:  http.MethodPost,
			giveHeader: http.Header{
				"X-API-Status": []string{"fail"},
			},
			giveRequest: &InternalTestRequest{
				Id: strconv.Itoa(http.StatusInternalServerError),
			},
			wantError: core.NewAPIError(
				http.StatusInternalServerError,
				http.Header{},
				errors.New("failed to process request"),
			),
		},
		{
			description: "POST extra properties",
			giveMethod:  http.MethodPost,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: new(InternalTestRequest),
			giveBodyProperties: map[string]interface{}{
				"key": "value",
			},
			wantResponse: &InternalTestResponse{
				ExtraBodyProperties: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			description: "GET extra query parameters",
			giveMethod:  http.MethodGet,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveQueryParams: url.Values{
				"extra": []string{"true"},
			},
			giveRequest: &InternalTestRequest{
				Id: "123",
			},
			wantResponse: &InternalTestResponse{
				Id: "123",
				QueryParameters: url.Values{
					"extra": []string{"true"},
				},
			},
		},
		{
			description:    "GET merge extra query parameters",
			givePathSuffix: "?limit=1",
			giveMethod:     http.MethodGet,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: &InternalTestRequest{
				Id: "123",
			},
			giveQueryParams: url.Values{
				"extra": []string{"true"},
			},
			wantResponse: &InternalTestResponse{
				Id: "123",
				QueryParameters: url.Values{
					"limit": []string{"1"},
					"extra": []string{"true"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var (
				server = newTestServer(t, test)
				client = server.Client()
			)
			caller := NewCaller(
				&CallerParams{
					Client: client,
				},
			)
			var response *InternalTestResponse
			_, err := caller.Call(
				context.Background(),
				&CallParams{
					URL:                server.URL + test.givePathSuffix,
					Method:             test.giveMethod,
					Headers:            test.giveHeader,
					BodyProperties:     test.giveBodyProperties,
					QueryParameters:    test.giveQueryParams,
					Request:            test.giveRequest,
					Response:           &response,
					ResponseIsOptional: test.giveResponseIsOptional,
					ErrorDecoder:       test.giveErrorDecoder,
				},
			)
			if test.wantError != nil {
				assert.EqualError(t, err, test.wantError.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.wantResponse, response)
		})
	}
}

func TestMergeHeaders(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		merged := MergeHeaders(make(http.Header), make(http.Header))
		assert.Empty(t, merged)
	})

	t.Run("empty left", func(t *testing.T) {
		left := make(http.Header)

		right := make(http.Header)
		right.Set("X-API-Version", "0.0.1")

		merged := MergeHeaders(left, right)
		assert.Equal(t, "0.0.1", merged.Get("X-API-Version"))
	})

	t.Run("empty right", func(t *testing.T) {
		left := make(http.Header)
		left.Set("X-API-Version", "0.0.1")

		right := make(http.Header)

		merged := MergeHeaders(left, right)
		assert.Equal(t, "0.0.1", merged.Get("X-API-Version"))
	})

	t.Run("single value override", func(t *testing.T) {
		left := make(http.Header)
		left.Set("X-API-Version", "0.0.0")

		right := make(http.Header)
		right.Set("X-API-Version", "0.0.1")

		merged := MergeHeaders(left, right)
		assert.Equal(t, []string{"0.0.1"}, merged.Values("X-API-Version"))
	})

	t.Run("multiple value override", func(t *testing.T) {
		left := make(http.Header)
		left.Set("X-API-Versions", "0.0.0")

		right := make(http.Header)
		right.Add("X-API-Versions", "0.0.1")
		right.Add("X-API-Versions", "0.0.2")

		merged := MergeHeaders(left, right)
		assert.Equal(t, []string{"0.0.1", "0.0.2"}, merged.Values("X-API-Versions"))
	})

	t.Run("disjoint merge", func(t *testing.T) {
		left := make(http.Header)
		left.Set("X-API-Tenancy", "test")

		right := make(http.Header)
		right.Set("X-API-Version", "0.0.1")

		merged := MergeHeaders(left, right)
		assert.Equal(t, []string{"test"}, merged.Values("X-API-Tenancy"))
		assert.Equal(t, []string{"0.0.1"}, merged.Values("X-API-Version"))
	})
}

// newTestServer returns a new *httptest.Server configured with the
// given test parameters.
func newTestServer(t *testing.T, tc *InternalTestCase) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.giveMethod, r.Method)
				assert.Equal(t, contentType, r.Header.Get(contentTypeHeader))
				for header, value := range tc.giveHeader {
					assert.Equal(t, value, r.Header.Values(header))
				}

				request := new(InternalTestRequest)

				bytes, err := io.ReadAll(r.Body)
				if tc.giveRequest == nil {
					require.Empty(t, bytes)
					w.WriteHeader(http.StatusBadRequest)
					_, err = w.Write([]byte("invalid request"))
					require.NoError(t, err)
					return
				}
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(bytes, request))

				switch request.Id {
				case strconv.Itoa(http.StatusNotFound):
					notFoundError := &InternalTestNotFoundError{
						APIError: &core.APIError{
							StatusCode: http.StatusNotFound,
						},
						Message: fmt.Sprintf("ID %q not found", request.Id),
					}
					bytes, err = json.Marshal(notFoundError)
					require.NoError(t, err)

					w.WriteHeader(http.StatusNotFound)
					_, err = w.Write(bytes)
					require.NoError(t, err)
					return

				case strconv.Itoa(http.StatusInternalServerError):
					w.WriteHeader(http.StatusInternalServerError)
					_, err = w.Write([]byte("failed to process request"))
					require.NoError(t, err)
					return
				}

				if tc.giveResponseIsOptional {
					w.WriteHeader(http.StatusOK)
					return
				}

				extraBodyProperties := make(map[string]interface{})
				require.NoError(t, json.Unmarshal(bytes, &extraBodyProperties))
				delete(extraBodyProperties, "id")

				response := &InternalTestResponse{
					Id:                  request.Id,
					ExtraBodyProperties: extraBodyProperties,
					QueryParameters:     r.URL.Query(),
				}
				bytes, err = json.Marshal(response)
				require.NoError(t, err)

				_, err = w.Write(bytes)
				require.NoError(t, err)
			},
		),
	)
}

func TestIsNil(t *testing.T) {
	t.Run("nil interface", func(t *testing.T) {
		assert.True(t, isNil(nil))
	})

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *string
		assert.True(t, isNil(ptr))
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		s := "test"
		assert.False(t, isNil(&s))
	})

	t.Run("nil slice", func(t *testing.T) {
		var slice []string
		assert.True(t, isNil(slice))
	})

	t.Run("non-nil slice", func(t *testing.T) {
		slice := []string{}
		assert.False(t, isNil(slice))
	})

	t.Run("nil map", func(t *testing.T) {
		var m map[string]string
		assert.True(t, isNil(m))
	})

	t.Run("non-nil map", func(t *testing.T) {
		m := make(map[string]string)
		assert.False(t, isNil(m))
	})

	t.Run("string value", func(t *testing.T) {
		assert.False(t, isNil("test"))
	})

	t.Run("empty string value", func(t *testing.T) {
		assert.False(t, isNil(""))
	})

	t.Run("int value", func(t *testing.T) {
		assert.False(t, isNil(42))
	})

	t.Run("zero int value", func(t *testing.T) {
		assert.False(t, isNil(0))
	})

	t.Run("bool value", func(t *testing.T) {
		assert.False(t, isNil(true))
	})

	t.Run("false bool value", func(t *testing.T) {
		assert.False(t, isNil(false))
	})

	t.Run("struct value", func(t *testing.T) {
		type testStruct struct {
			Field string
		}
		assert.False(t, isNil(testStruct{Field: "test"}))
	})

	t.Run("empty struct value", func(t *testing.T) {
		type testStruct struct {
			Field string
		}
		assert.False(t, isNil(testStruct{}))
	})
}

// newTestErrorDecoder returns an error decoder suitable for tests.
func newTestErrorDecoder(t *testing.T) func(int, http.Header, io.Reader) error {
	return func(statusCode int, header http.Header, body io.Reader) error {
		raw, err := io.ReadAll(body)
		require.NoError(t, err)

		var (
			apiError = core.NewAPIError(statusCode, header, errors.New(string(raw)))
			decoder  = json.NewDecoder(bytes.NewReader(raw))
		)
		if statusCode == http.StatusNotFound {
			value := new(InternalTestNotFoundError)
			value.APIError = apiError
			require.NoError(t, decoder.Decode(value))

			return value
		}
		return apiError
	}
}

// FormURLEncodedTestRequest is a test struct for form URL encoding tests.
type FormURLEncodedTestRequest struct {
	ClientID     string  `json:"client_id"`
	ClientSecret string  `json:"client_secret"`
	GrantType    string  `json:"grant_type,omitempty"`
	Scope        *string `json:"scope,omitempty"`
	NilPointer   *string `json:"nil_pointer,omitempty"`
}

func TestNewFormURLEncodedBody(t *testing.T) {
	t.Run("simple key-value pairs", func(t *testing.T) {
		bodyProperties := map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"grant_type":    "client_credentials",
		}
		reader := newFormURLEncodedBody(bodyProperties)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		// Parse the body and verify values
		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", values.Get("client_id"))
		assert.Equal(t, "test_client_secret", values.Get("client_secret"))
		assert.Equal(t, "client_credentials", values.Get("grant_type"))

		// Verify it's not JSON
		bodyStr := string(body)
		assert.False(t, strings.HasPrefix(strings.TrimSpace(bodyStr), "{"),
			"Body should not be JSON, got: %s", bodyStr)
	})

	t.Run("special characters requiring URL encoding", func(t *testing.T) {
		bodyProperties := map[string]interface{}{
			"value_with_space":     "hello world",
			"value_with_ampersand": "a&b",
			"value_with_equals":    "a=b",
			"value_with_plus":      "a+b",
		}
		reader := newFormURLEncodedBody(bodyProperties)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		// Parse the body and verify values are correctly decoded
		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "hello world", values.Get("value_with_space"))
		assert.Equal(t, "a&b", values.Get("value_with_ampersand"))
		assert.Equal(t, "a=b", values.Get("value_with_equals"))
		assert.Equal(t, "a+b", values.Get("value_with_plus"))
	})

	t.Run("empty map", func(t *testing.T) {
		bodyProperties := map[string]interface{}{}
		reader := newFormURLEncodedBody(bodyProperties)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Empty(t, string(body))
	})
}

func TestNewFormURLEncodedRequestBody(t *testing.T) {
	t.Run("struct with json tags", func(t *testing.T) {
		scope := "read write"
		request := &FormURLEncodedTestRequest{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			GrantType:    "client_credentials",
			Scope:        &scope,
			NilPointer:   nil,
		}
		reader, err := newFormURLEncodedRequestBody(request, nil)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		// Parse the body and verify values
		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", values.Get("client_id"))
		assert.Equal(t, "test_client_secret", values.Get("client_secret"))
		assert.Equal(t, "client_credentials", values.Get("grant_type"))
		assert.Equal(t, "read write", values.Get("scope"))
		// nil_pointer should not be present (nil pointer with omitempty)
		assert.Empty(t, values.Get("nil_pointer"))

		// Verify it's not JSON
		bodyStr := string(body)
		assert.False(t, strings.HasPrefix(strings.TrimSpace(bodyStr), "{"),
			"Body should not be JSON, got: %s", bodyStr)
	})

	t.Run("struct with omitempty and zero values", func(t *testing.T) {
		request := &FormURLEncodedTestRequest{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			GrantType:    "", // empty string with omitempty should be omitted
			Scope:        nil,
			NilPointer:   nil,
		}
		reader, err := newFormURLEncodedRequestBody(request, nil)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", values.Get("client_id"))
		assert.Equal(t, "test_client_secret", values.Get("client_secret"))
		// grant_type should not be present (empty string with omitempty)
		assert.Empty(t, values.Get("grant_type"))
		assert.Empty(t, values.Get("scope"))
	})

	t.Run("struct with extra body properties", func(t *testing.T) {
		request := &FormURLEncodedTestRequest{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		}
		bodyProperties := map[string]interface{}{
			"extra_param": "extra_value",
		}
		reader, err := newFormURLEncodedRequestBody(request, bodyProperties)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", values.Get("client_id"))
		assert.Equal(t, "test_client_secret", values.Get("client_secret"))
		assert.Equal(t, "extra_value", values.Get("extra_param"))
	})

	t.Run("special characters in struct fields", func(t *testing.T) {
		scope := "read&write=all+permissions"
		request := &FormURLEncodedTestRequest{
			ClientID:     "client with spaces",
			ClientSecret: "secret&with=special+chars",
			Scope:        &scope,
		}
		reader, err := newFormURLEncodedRequestBody(request, nil)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "client with spaces", values.Get("client_id"))
		assert.Equal(t, "secret&with=special+chars", values.Get("client_secret"))
		assert.Equal(t, "read&write=all+permissions", values.Get("scope"))
	})
}

func TestNewRequestBodyFormURLEncoded(t *testing.T) {
	t.Run("selects form encoding when content-type is form-urlencoded", func(t *testing.T) {
		request := &FormURLEncodedTestRequest{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			GrantType:    "client_credentials",
		}
		reader, err := newRequestBody(request, nil, contentTypeFormURLEncoded)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		// Verify it's form-urlencoded, not JSON
		bodyStr := string(body)
		assert.False(t, strings.HasPrefix(strings.TrimSpace(bodyStr), "{"),
			"Body should not be JSON when Content-Type is form-urlencoded, got: %s", bodyStr)

		// Parse and verify values
		values, err := url.ParseQuery(bodyStr)
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", values.Get("client_id"))
		assert.Equal(t, "test_client_secret", values.Get("client_secret"))
		assert.Equal(t, "client_credentials", values.Get("grant_type"))
	})

	t.Run("selects JSON encoding when content-type is application/json", func(t *testing.T) {
		request := &FormURLEncodedTestRequest{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		}
		reader, err := newRequestBody(request, nil, contentType)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		// Verify it's JSON
		bodyStr := string(body)
		assert.True(t, strings.HasPrefix(strings.TrimSpace(bodyStr), "{"),
			"Body should be JSON when Content-Type is application/json, got: %s", bodyStr)

		// Parse and verify it's valid JSON
		var parsed map[string]interface{}
		err = json.Unmarshal(body, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", parsed["client_id"])
		assert.Equal(t, "test_client_secret", parsed["client_secret"])
	})

	t.Run("form encoding with body properties only (nil request)", func(t *testing.T) {
		bodyProperties := map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
		}
		reader, err := newRequestBody(nil, bodyProperties, contentTypeFormURLEncoded)
		require.NoError(t, err)

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, "test_client_id", values.Get("client_id"))
		assert.Equal(t, "test_client_secret", values.Get("client_secret"))
	})
}
