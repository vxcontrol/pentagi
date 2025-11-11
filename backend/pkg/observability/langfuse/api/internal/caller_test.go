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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pentagi/pkg/observability/langfuse/api/core"
)

// TestCase represents a single test case.
type TestCase struct {
	description string

	// Server-side assertions.
	givePathSuffix         string
	giveMethod             string
	giveResponseIsOptional bool
	giveHeader             http.Header
	giveErrorDecoder       ErrorDecoder
	giveRequest            *Request
	giveQueryParams        url.Values
	giveBodyProperties     map[string]interface{}

	// Client-side assertions.
	wantResponse *Response
	wantError    error
}

// Request a simple request body.
type Request struct {
	Id string `json:"id"`
}

// Response a simple response body.
type Response struct {
	Id                  string                 `json:"id"`
	ExtraBodyProperties map[string]interface{} `json:"extraBodyProperties,omitempty"`
	QueryParameters     url.Values             `json:"queryParameters,omitempty"`
}

// NotFoundError represents a 404.
type NotFoundError struct {
	*core.APIError

	Message string `json:"message"`
}

func TestCall(t *testing.T) {
	tests := []*TestCase{
		{
			description: "GET success",
			giveMethod:  http.MethodGet,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: &Request{
				Id: "123",
			},
			wantResponse: &Response{
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
			giveRequest: &Request{
				Id: "123",
			},
			wantResponse: &Response{
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
			giveRequest: &Request{
				Id: strconv.Itoa(http.StatusNotFound),
			},
			giveErrorDecoder: newTestErrorDecoder(t),
			wantError: &NotFoundError{
				APIError: core.NewAPIError(
					http.StatusNotFound,
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
				errors.New("invalid request"),
			),
		},
		{
			description: "POST optional response",
			giveMethod:  http.MethodPost,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: &Request{
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
			giveRequest: &Request{
				Id: strconv.Itoa(http.StatusInternalServerError),
			},
			wantError: core.NewAPIError(
				http.StatusInternalServerError,
				errors.New("failed to process request"),
			),
		},
		{
			description: "POST extra properties",
			giveMethod:  http.MethodPost,
			giveHeader: http.Header{
				"X-API-Status": []string{"success"},
			},
			giveRequest: new(Request),
			giveBodyProperties: map[string]interface{}{
				"key": "value",
			},
			wantResponse: &Response{
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
			giveRequest: &Request{
				Id: "123",
			},
			wantResponse: &Response{
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
			giveRequest: &Request{
				Id: "123",
			},
			giveQueryParams: url.Values{
				"extra": []string{"true"},
			},
			wantResponse: &Response{
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
			var response *Response
			err := caller.Call(
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
func newTestServer(t *testing.T, tc *TestCase) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.giveMethod, r.Method)
				assert.Equal(t, contentType, r.Header.Get(contentTypeHeader))
				for header, value := range tc.giveHeader {
					assert.Equal(t, value, r.Header.Values(header))
				}

				request := new(Request)

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
					notFoundError := &NotFoundError{
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

				response := &Response{
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

// newTestErrorDecoder returns an error decoder suitable for tests.
func newTestErrorDecoder(t *testing.T) func(int, io.Reader) error {
	return func(statusCode int, body io.Reader) error {
		raw, err := io.ReadAll(body)
		require.NoError(t, err)

		var (
			apiError = core.NewAPIError(statusCode, errors.New(string(raw)))
			decoder  = json.NewDecoder(bytes.NewReader(raw))
		)
		if statusCode == http.StatusNotFound {
			value := new(NotFoundError)
			value.APIError = apiError
			require.NoError(t, decoder.Decode(value))

			return value
		}
		return apiError
	}
}
