package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"pentagi/pkg/observability/langfuse/api/core"
)

const (
	// contentType specifies the JSON Content-Type header value.
	contentType       = "application/json"
	contentTypeHeader = "Content-Type"
)

// Caller calls APIs and deserializes their response, if any.
type Caller struct {
	client  core.HTTPClient
	retrier *Retrier
}

// CallerParams represents the parameters used to constrcut a new *Caller.
type CallerParams struct {
	Client      core.HTTPClient
	MaxAttempts uint
}

// NewCaller returns a new *Caller backed by the given parameters.
func NewCaller(params *CallerParams) *Caller {
	var httpClient core.HTTPClient = http.DefaultClient
	if params.Client != nil {
		httpClient = params.Client
	}
	var retryOptions []RetryOption
	if params.MaxAttempts > 0 {
		retryOptions = append(retryOptions, WithMaxAttempts(params.MaxAttempts))
	}
	return &Caller{
		client:  httpClient,
		retrier: NewRetrier(retryOptions...),
	}
}

// CallParams represents the parameters used to issue an API call.
type CallParams struct {
	URL                string
	Method             string
	MaxAttempts        uint
	Headers            http.Header
	BodyProperties     map[string]interface{}
	QueryParameters    url.Values
	Client             core.HTTPClient
	Request            interface{}
	Response           interface{}
	ResponseIsOptional bool
	ErrorDecoder       ErrorDecoder
}

// Call issues an API call according to the given call parameters.
func (c *Caller) Call(ctx context.Context, params *CallParams) error {
	url := buildURL(params.URL, params.QueryParameters)
	req, err := newRequest(
		ctx,
		url,
		params.Method,
		params.Headers,
		params.Request,
		params.BodyProperties,
	)
	if err != nil {
		return err
	}

	// If the call has been cancelled, don't issue the request.
	if err := ctx.Err(); err != nil {
		return err
	}

	client := c.client
	if params.Client != nil {
		// Use the HTTP client scoped to the request.
		client = params.Client
	}

	var retryOptions []RetryOption
	if params.MaxAttempts > 0 {
		retryOptions = append(retryOptions, WithMaxAttempts(params.MaxAttempts))
	}

	resp, err := c.retrier.Run(
		client.Do,
		req,
		params.ErrorDecoder,
		retryOptions...,
	)
	if err != nil {
		return err
	}

	// Close the response body after we're done.
	defer resp.Body.Close()

	// Check if the call was cancelled before we return the error
	// associated with the call and/or unmarshal the response data.
	if err := ctx.Err(); err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp, params.ErrorDecoder)
	}

	// Mutate the response parameter in-place.
	if params.Response != nil {
		if writer, ok := params.Response.(io.Writer); ok {
			_, err = io.Copy(writer, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(params.Response)
		}
		if err != nil {
			if err == io.EOF {
				if params.ResponseIsOptional {
					// The response is optional, so we should ignore the
					// io.EOF error
					return nil
				}
				return fmt.Errorf("expected a %T response, but the server responded with nothing", params.Response)
			}
			return err
		}
	}

	return nil
}

// buildURL constructs the final URL by appending the given query parameters (if any).
func buildURL(
	url string,
	queryParameters url.Values,
) string {
	if len(queryParameters) == 0 {
		return url
	}
	if strings.ContainsRune(url, '?') {
		url += "&"
	} else {
		url += "?"
	}
	url += queryParameters.Encode()
	return url
}

// newRequest returns a new *http.Request with all of the fields
// required to issue the call.
func newRequest(
	ctx context.Context,
	url string,
	method string,
	endpointHeaders http.Header,
	request interface{},
	bodyProperties map[string]interface{},
) (*http.Request, error) {
	requestBody, err := newRequestBody(request, bodyProperties)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set(contentTypeHeader, contentType)
	for name, values := range endpointHeaders {
		req.Header[name] = values
	}
	return req, nil
}

// newRequestBody returns a new io.Reader that represents the HTTP request body.
func newRequestBody(request interface{}, bodyProperties map[string]interface{}) (io.Reader, error) {
	if isNil(request) {
		if len(bodyProperties) == 0 {
			return nil, nil
		}
		requestBytes, err := json.Marshal(bodyProperties)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(requestBytes), nil
	}
	if body, ok := request.(io.Reader); ok {
		return body, nil
	}
	requestBytes, err := MarshalJSONWithExtraProperties(request, bodyProperties)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(requestBytes), nil
}

// decodeError decodes the error from the given HTTP response. Note that
// it's the caller's responsibility to close the response body.
func decodeError(response *http.Response, errorDecoder ErrorDecoder) error {
	if errorDecoder != nil {
		// This endpoint has custom errors, so we'll
		// attempt to unmarshal the error into a structured
		// type based on the status code.
		return errorDecoder(response.StatusCode, response.Body)
	}
	// This endpoint doesn't have any custom error
	// types, so we just read the body as-is, and
	// put it into a normal error.
	bytes, err := io.ReadAll(response.Body)
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		// The error didn't have a response body,
		// so all we can do is return an error
		// with the status code.
		return core.NewAPIError(response.StatusCode, nil)
	}
	return core.NewAPIError(response.StatusCode, errors.New(string(bytes)))
}

// isNil is used to determine if the request value is equal to nil (i.e. an interface
// value that holds a nil concrete value is itself non-nil).
func isNil(value interface{}) bool {
	return value == nil || reflect.ValueOf(value).IsNil()
}
