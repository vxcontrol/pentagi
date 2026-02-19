package internal

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

    "pentagi/pkg/observability/langfuse/api/core"
	"github.com/stretchr/testify/assert"
)

func TestErrorDecoder(t *testing.T) {
	decoder := NewErrorDecoder(
		ErrorCodes{
			http.StatusNotFound: func(apiError *core.APIError) error {
				return &InternalTestNotFoundError{APIError: apiError}
			},
		})

	tests := []struct {
		description    string
		giveStatusCode int
		giveHeader     http.Header
		giveBody       string
		wantError      error
	}{
		{
			description:    "unrecognized status code",
			giveStatusCode: http.StatusInternalServerError,
			giveHeader:     http.Header{},
			giveBody:       "Internal Server Error",
			wantError:      core.NewAPIError(http.StatusInternalServerError, http.Header{}, errors.New("Internal Server Error")),
		},
		{
			description:    "not found with valid JSON",
			giveStatusCode: http.StatusNotFound,
			giveHeader:     http.Header{},
			giveBody:       `{"message": "Resource not found"}`,
			wantError: &InternalTestNotFoundError{
				APIError: core.NewAPIError(http.StatusNotFound, http.Header{}, errors.New(`{"message": "Resource not found"}`)),
				Message:  "Resource not found",
			},
		},
		{
			description:    "not found with invalid JSON",
			giveStatusCode: http.StatusNotFound,
			giveHeader:     http.Header{},
			giveBody:       `Resource not found`,
			wantError:      core.NewAPIError(http.StatusNotFound, http.Header{}, errors.New("Resource not found")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.wantError, decoder(tt.giveStatusCode, tt.giveHeader, bytes.NewReader([]byte(tt.giveBody))))
		})
	}
}
