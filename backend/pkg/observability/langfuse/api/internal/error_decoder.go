package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"pentagi/pkg/observability/langfuse/api/core"
)

// ErrorDecoder decodes *http.Response errors and returns a
// typed API error (e.g. *core.APIError).
type ErrorDecoder func(statusCode int, body io.Reader) error

// ErrorCodes maps HTTP status codes to error constructors.
type ErrorCodes map[int]func(*core.APIError) error

// NewErrorDecoder returns a new ErrorDecoder backed by the given error codes.
func NewErrorDecoder(errorCodes ErrorCodes) ErrorDecoder {
	return func(statusCode int, body io.Reader) error {
		raw, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("failed to read error from response body: %w", err)
		}
		apiError := core.NewAPIError(
			statusCode,
			errors.New(string(raw)),
		)
		newErrorFunc, ok := errorCodes[statusCode]
		if !ok {
			// This status code isn't recognized, so we return
			// the API error as-is.
			return apiError
		}
		customError := newErrorFunc(apiError)
		if err := json.NewDecoder(bytes.NewReader(raw)).Decode(customError); err != nil {
			// If we fail to decode the error, we return the
			// API error as-is.
			return apiError
		}
		return customError
	}
}
