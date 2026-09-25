package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/vxcontrol/cloud/models"
	"github.com/vxcontrol/cloud/sdk"
)

// updatesValidRequest is the smallest request the contract accepts, so a rejection is for the reason under test.
func updatesValidRequest() models.CheckUpdatesRequest {
	return models.CheckUpdatesRequest{
		InstallerVersion: "2.0.0",
		InstallerOS:      models.OSTypeLinux,
		InstallerArch:    models.ArchTypeAMD64,
		Strategy:         models.UpdateStrategyStable,
		Images: []models.ImageComponentInfo{{
			Component:  models.ComponentTypePentagi,
			Status:     models.ComponentStatusRunning,
			OS:         models.OSTypeLinux,
			Arch:       models.ArchTypeAMD64,
			Repository: "vxcontrol/pentagi",
			Tag:        "latest",
		}},
	}
}

// Decoding the envelope straight into the response type succeeds with every field zero: "no updates", silently.
func TestUpdates_CheckUpdates_SendsTheRequestAndUnwrapsTheAnswer(t *testing.T) {
	latest := "2.1.0"
	enveloped, err := json.Marshal(map[string]any{"status": "success", "data": models.CheckUpdatesResponse{
		Updates: []models.UpdateInfo{{Stack: models.ProductStackPentagi, HasUpdate: true, LatestVersion: &latest}},
	}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var sent models.CheckUpdatesRequest
	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		if err := json.Unmarshal(body, &sent); err != nil {
			return nil, err
		}
		return enveloped, nil
	}}

	request := updatesValidRequest()
	response, err := client.CheckUpdates(context.Background(), request)
	if err != nil {
		t.Fatalf("CheckUpdates: %v", err)
	}

	if !reflect.DeepEqual(sent, request) {
		t.Errorf("sent %+v, want the caller's request unaltered %+v", sent, request)
	}
	if len(response.Updates) != 1 || !response.Updates[0].HasUpdate || *response.Updates[0].LatestVersion != "2.1.0" {
		t.Errorf("updates = %+v, want the one enveloped stack with 2.1.0 available", response.Updates)
	}
}

func TestUpdates_CheckUpdates_RefusesAMalformedRequestBeforeSending(t *testing.T) {
	called := false
	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		called = true
		return nil, nil
	}}

	request := updatesValidRequest()
	request.Strategy = models.UpdateStrategy("bleeding-edge")

	_, err := client.CheckUpdates(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), "update check request is malformed") {
		t.Fatalf("CheckUpdates() error = %v, want the request refused as malformed", err)
	}
	if called {
		t.Error("the malformed request was sent anyway")
	}
}

func TestUpdates_CheckUpdates_ClassifiesCallFailures(t *testing.T) {
	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		return nil, fmt.Errorf("call refused: %w", sdk.ErrForbidden)
	}}

	_, err := client.CheckUpdates(context.Background(), updatesValidRequest())

	var failure *Failure
	if !errors.As(err, &failure) {
		t.Fatalf("error is not classified: %v", err)
	}
	if failure.Reason != FailureForbidden {
		t.Errorf("reason = %q, want %q", failure.Reason, FailureForbidden)
	}
	if err.Error() != "call refused: forbidden" {
		t.Errorf("message = %q, want the call's own %q", err.Error(), "call refused: forbidden")
	}
}
