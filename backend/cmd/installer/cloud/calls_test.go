package cloud

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/vxcontrol/cloud/models"
	"github.com/vxcontrol/cloud/sdk"
)

// The tests below drive the client through substituted call functions. That is deliberate:
// the transport always negotiates TLS and solves a proof-of-work challenge, so there is no
// local server to point it at, and what needs covering here is what the client does around
// the call — validate, envelope, verify — not the transport itself.

func stringPtr(s string) *string { return &s }

// validCheckRequest is the smallest request the contract accepts, so a test that expects
// rejection is rejected for the reason it is testing and not by accident.
func validCheckRequest() models.CheckUpdatesRequest {
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

// TestCheckUpdatesUnwrapsTheEnvelope is the regression that matters most here. The answer
// arrives wrapped in a status envelope, and decoding it straight into the response type
// SUCCEEDS — every field simply stays zero. Skipping the unwrapping therefore reports "no
// updates available" for every answer, with no error anywhere to explain it.
func TestCheckUpdatesUnwrapsTheEnvelope(t *testing.T) {
	latest := "2.1.0"
	payload := models.CheckUpdatesResponse{
		Updates: []models.UpdateInfo{{
			Stack:         models.ProductStackPentagi,
			HasUpdate:     true,
			LatestVersion: &latest,
		}},
	}
	enveloped, err := json.Marshal(map[string]any{"status": "success", "data": payload})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		return enveloped, nil
	}}

	response, err := client.CheckUpdates(context.Background(), validCheckRequest())
	if err != nil {
		t.Fatalf("CheckUpdates: %v", err)
	}

	if len(response.Updates) != 1 {
		t.Fatalf("got %d stacks, want 1 — the envelope was not unwrapped", len(response.Updates))
	}
	if !response.Updates[0].HasUpdate {
		t.Error("the answer says an update is available, the client says it is not")
	}
}

// TestCheckUpdatesSendsWhatItWasGiven: the request reaching the wire is the caller's,
// unaltered. The server answers only about the components it is told about, so anything
// dropped here comes back as "nothing to report" rather than "not asked".
func TestCheckUpdatesSendsWhatItWasGiven(t *testing.T) {
	var sent models.CheckUpdatesRequest

	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		if err := json.Unmarshal(body, &sent); err != nil {
			return nil, err
		}
		return []byte(`{"status":"success","data":{"updates":[]}}`), nil
	}}

	request := validCheckRequest()
	if _, err := client.CheckUpdates(context.Background(), request); err != nil {
		t.Fatalf("CheckUpdates: %v", err)
	}

	if len(sent.Images) != len(request.Images) || len(sent.Files) != len(request.Files) {
		t.Fatalf("sent %d images and %d files, want %d and %d",
			len(sent.Images), len(sent.Files), len(request.Images), len(request.Files))
	}
	if sent.Strategy != request.Strategy {
		t.Errorf("strategy = %q, want %q", sent.Strategy, request.Strategy)
	}
	if sent.InstallerVersion != request.InstallerVersion {
		t.Errorf("installer version = %q, want %q", sent.InstallerVersion, request.InstallerVersion)
	}
}

// TestCheckUpdatesRefusesAMalformedRequestBeforeSending: a request the contract rejects
// comes back as a server-side failure, which is indistinguishable from an outage and just
// as pointless to retry. Catching it here names the field instead — and costs no call.
func TestCheckUpdatesRefusesAMalformedRequestBeforeSending(t *testing.T) {
	called := false
	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		called = true
		return nil, nil
	}}

	request := validCheckRequest()
	request.Strategy = models.UpdateStrategy("bleeding-edge")

	if _, err := client.CheckUpdates(context.Background(), request); err == nil {
		t.Fatal("a request with an unknown strategy was accepted")
	}
	if called {
		t.Error("the malformed request was sent anyway")
	}
}

func TestCheckUpdatesClassifiesCallFailures(t *testing.T) {
	client := &Client{checkUpdates: func(ctx context.Context, body []byte) ([]byte, error) {
		return nil, sdk.ErrForbidden
	}}

	_, err := client.CheckUpdates(context.Background(), validCheckRequest())
	if err == nil {
		t.Fatal("expected an error")
	}

	var failure *Failure
	if !errors.As(err, &failure) {
		t.Fatalf("error is not classified: %v", err)
	}
	if failure.Reason != FailureForbidden {
		t.Errorf("reason = %q, want %q", failure.Reason, FailureForbidden)
	}
}

func TestPackageInfoUnwrapsTheEnvelopeAndValidates(t *testing.T) {
	payload := models.PackageInfoResponse{
		Size:      1024,
		Hash:      strings.Repeat("a", 64),
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
		Signature: models.SignatureValue(strings.Repeat("Q", 86)),
	}
	enveloped, err := json.Marshal(map[string]any{"status": "success", "data": payload})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	client := &Client{packageInfo: func(ctx context.Context, query map[string]string) ([]byte, error) {
		// The query is how this endpoint carries its arguments; an empty one would ask
		// about nothing in particular.
		if query["component"] != "installer" {
			t.Errorf("component = %q, want installer", query["component"])
		}
		if query["version"] != "2.0.0" {
			t.Errorf("version = %q, want 2.0.0", query["version"])
		}
		return enveloped, nil
	}}

	info, err := client.PackageInfo(context.Background(), models.PackageInfoRequest{
		Component: models.ComponentTypeInstaller,
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
	})
	if err != nil {
		t.Fatalf("PackageInfo: %v", err)
	}
	if info.Size != 1024 {
		t.Errorf("size = %d, want 1024 — the envelope was not unwrapped", info.Size)
	}
}

// signedPackage builds a package together with a description of it that verifies, so the
// download test exercises the real Ed25519 and sha256 paths rather than stubbed checks.
// The signature covers the file's SHA-512 digest.
func signedPackage(t *testing.T, payload []byte) (models.PackageInfoResponse, models.DownloadPackageRequest) {
	t.Helper()

	_, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	sha512Digest := sha512.Sum512(payload)
	sha256Digest := sha256.Sum256(payload)

	var signature models.SignatureValue
	if _, err := signature.FromBytes(ed25519.Sign(private, sha512Digest[:])); err != nil {
		t.Fatalf("encode signature: %v", err)
	}

	info := models.PackageInfoResponse{
		Size:      int64(len(payload)),
		Hash:      hex.EncodeToString(sha256Digest[:]),
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
		Signature: signature,
	}
	request := models.DownloadPackageRequest{
		Component: models.ComponentTypeInstaller,
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
	}

	return info, request
}

// TestDownloadPackageRejectsATamperedPayload covers the three checks that make a downloaded
// binary safe to run. The signature is not verifiable in a unit test — it is made with a
// throwaway key, not the release key — so this asserts what the size and digest checks
// catch, which is every alteration of the bytes.
func TestDownloadPackageRejectsATamperedPayload(t *testing.T) {
	payload := []byte(strings.Repeat("installer bytes ", 512))
	info, request := signedPackage(t, payload)

	tests := []struct {
		name     string
		served   []byte
		wantHint string
	}{
		{name: "truncated", served: payload[:len(payload)-16], wantHint: "bytes, expected"},
		{name: "padded", served: append(append([]byte{}, payload...), 'x'), wantHint: "bytes, expected"},
		{
			name:     "same length, different content",
			served:   append(append([]byte{}, payload[:len(payload)-1]...), 'X'),
			wantHint: "sha256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				downloadPackage: func(ctx context.Context, query map[string]string, w io.Writer) error {
					_, err := w.Write(tt.served)
					return err
				},
			}

			var sink bytes.Buffer
			err := client.DownloadPackage(context.Background(), request, info, &sink)
			if err == nil {
				t.Fatalf("a %s package was accepted", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHint) {
				t.Errorf("error = %v, want it to mention %q", err, tt.wantHint)
			}
		})
	}
}

// TestDownloadPackageWritesEverythingItReceives: whatever the verdict, the caller's writer
// gets the bytes — which is why the caller has to treat the destination as unverified until
// this returns nil.
func TestDownloadPackageWritesEverythingItReceives(t *testing.T) {
	payload := []byte(strings.Repeat("installer bytes ", 512))
	info, request := signedPackage(t, payload)

	client := &Client{
		downloadPackage: func(ctx context.Context, query map[string]string, w io.Writer) error {
			if query["component"] != "installer" {
				t.Errorf("component = %q, want installer", query["component"])
			}
			_, err := w.Write(payload)
			return err
		},
	}

	var sink bytes.Buffer
	// The signature check fails: the test key is not the release key. Everything before it
	// must already have passed, which is what the assertions below establish.
	err := client.DownloadPackage(context.Background(), request, info, &sink)
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("expected the signature check to be the one that fails, got: %v", err)
	}
	if !bytes.Equal(sink.Bytes(), payload) {
		t.Errorf("the destination received %d bytes, want %d", sink.Len(), len(payload))
	}
}

func TestDownloadPackageRefusesAMalformedRequestBeforeSending(t *testing.T) {
	called := false
	client := &Client{
		downloadPackage: func(ctx context.Context, query map[string]string, w io.Writer) error {
			called = true
			return nil
		},
	}

	// Only components an executable is published for can be downloaded, and neo4j is
	// somebody else's container image — there is no file behind it under any
	// circumstances.
	//
	// Deliberately NOT pentagi, which is the one component where "delivered as an
	// image" and "has no package" came apart: it ships an executable a running
	// product can replace itself with, so it is a valid download and would make this
	// test pass for the wrong reason.
	request := models.DownloadPackageRequest{
		Component: models.ComponentTypeNeo4j,
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
	}
	// The description has to be a valid one, or it — not the request — would be what stops
	// the call, and this test would pass with the request check removed.
	info := models.PackageInfoResponse{
		Size:      1024,
		Hash:      strings.Repeat("a", 64),
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
		Signature: models.SignatureValue(strings.Repeat("Q", 86)),
	}

	err := client.DownloadPackage(context.Background(), request, info, io.Discard)
	if err == nil {
		t.Fatal("a download was accepted for a component that has no package")
	}
	if called {
		t.Error("the malformed request was sent anyway")
	}
}

// TestCallsOnAnUninitializedClientFailCleanly: the zero Client is what a failed New leaves
// behind, and calling into it must report that rather than panic.
func TestCallsOnAnUninitializedClientFailCleanly(t *testing.T) {
	var client *Client

	if _, err := client.CheckUpdates(context.Background(), validCheckRequest()); err == nil {
		t.Error("CheckUpdates on a nil client did not fail")
	}
	if _, err := client.PackageInfo(context.Background(), models.PackageInfoRequest{}); err == nil {
		t.Error("PackageInfo on a nil client did not fail")
	}
	if err := client.DownloadPackage(
		context.Background(), models.DownloadPackageRequest{}, models.PackageInfoResponse{}, io.Discard,
	); err == nil {
		t.Error("DownloadPackage on a nil client did not fail")
	}
}
