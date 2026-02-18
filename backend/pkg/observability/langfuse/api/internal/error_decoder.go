package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"pentagi/pkg/observability/langfuse/api/core"
)

// ErrorCodes maps HTTP status codes to error constructors.
type ErrorCodes map[int]func(*core.APIError) error

// ErrorDecoder decodes *http.Response errors and returns a
// typed API error (e.g. *core.APIError).
type ErrorDecoder func(statusCode int, header http.Header, body io.Reader) error

// NewErrorDecoder returns a new ErrorDecoder backed by the given error codes.
// errorCodesOverrides is optional and will be merged with the default error codes,
// with overrides taking precedence.
func NewErrorDecoder(errorCodes ErrorCodes, errorCodesOverrides ...ErrorCodes) ErrorDecoder {
	// Merge default error codes with overrides
	mergedErrorCodes := make(ErrorCodes)

	// Start with default error codes
	for statusCode, errorFunc := range errorCodes {
		mergedErrorCodes[statusCode] = errorFunc
	}

	// Apply overrides if provided
	if len(errorCodesOverrides) > 0 && errorCodesOverrides[0] != nil {
		for statusCode, errorFunc := range errorCodesOverrides[0] {
			mergedErrorCodes[statusCode] = errorFunc
		}
	}

	return func(statusCode int, header http.Header, body io.Reader) error {
		raw, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("failed to read error from response body: %w", err)
		}
		apiError := core.NewAPIError(
			statusCode,
			header,
			errors.New(string(raw)),
		)
		newErrorFunc, ok := mergedErrorCodes[statusCode]
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
