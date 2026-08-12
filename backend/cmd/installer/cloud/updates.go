package cloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vxcontrol/cloud/models"
)

// CheckUpdates asks the update server what is available for the components described in
// req, and returns the answer per product stack.
//
// The request is validated before it leaves: a malformed one is refused by the server as an
// internal error, which is indistinguishable from an outage and equally pointless to retry.
// Catching it here names the offending field instead.
//
// The component list must describe the whole installation, not just the parts that look
// suspicious: the server answers only about what it was told, so anything left out comes
// back as "nothing to report" rather than "not asked".
func (c *Client) CheckUpdates(
	ctx context.Context,
	req models.CheckUpdatesRequest,
) (*models.CheckUpdatesResponse, error) {
	if c == nil || c.checkUpdates == nil {
		return nil, fmt.Errorf("update client is not initialized")
	}

	if err := req.Valid(); err != nil {
		return nil, fmt.Errorf("update check request is malformed: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode update check request: %w", err)
	}

	answer, err := c.checkUpdates(ctx, body)
	if err != nil {
		return nil, Classify(err)
	}

	// Answers are wrapped in a status envelope. Decoding straight into the response type
	// succeeds against the wrapper too — every field simply stays at its zero value — so
	// the unwrapping is not optional: skipping it reports "no updates" for every answer.
	response, err := models.ParseEnvelope[models.CheckUpdatesResponse](answer)
	if err != nil {
		return nil, fmt.Errorf("failed to read update check answer: %w", err)
	}

	return &response, nil
}
