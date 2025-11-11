package internal

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"pentagi/pkg/observability/langfuse/api/core"
)

func TestErrorDecoder(t *testing.T) {
	decoder := NewErrorDecoder(
		ErrorCodes{
			http.StatusNotFound: func(apiError *core.APIError) error {
				return &NotFoundError{APIError: apiError}
			},
		})

	tests := []struct {
		description    string
		giveStatusCode int
		giveBody       string
		wantError      error
	}{
		{
			description:    "unrecognized status code",
			giveStatusCode: http.StatusInternalServerError,
			giveBody:       "Internal Server Error",
			wantError:      core.NewAPIError(http.StatusInternalServerError, errors.New("Internal Server Error")),
		},
		{
			description:    "not found with valid JSON",
			giveStatusCode: http.StatusNotFound,
			giveBody:       `{"message": "Resource not found"}`,
			wantError: &NotFoundError{
				APIError: core.NewAPIError(http.StatusNotFound, errors.New(`{"message": "Resource not found"}`)),
				Message:  "Resource not found",
			},
		},
		{
			description:    "not found with invalid JSON",
			giveStatusCode: http.StatusNotFound,
			giveBody:       `Resource not found`,
			wantError:      core.NewAPIError(http.StatusNotFound, errors.New("Resource not found")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.wantError, decoder(tt.giveStatusCode, bytes.NewReader([]byte(tt.giveBody))))
		})
	}
}
