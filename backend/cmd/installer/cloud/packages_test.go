package cloud

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/vxcontrol/cloud/models"
)

// packagesValidInfo is a well-formed package description, so a refusal is not caused by it.
func packagesValidInfo() models.PackageInfoResponse {
	return models.PackageInfoResponse{
		Size:      1024,
		Hash:      strings.Repeat("a", 64),
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
		Signature: models.SignatureValue(strings.Repeat("Q", 86)),
	}
}

// packagesSigned describes payload with a real sha256 and an Ed25519 signature over its SHA-512 digest.
func packagesSigned(t *testing.T, payload []byte) (models.PackageInfoResponse, models.DownloadPackageRequest) {
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

	info := packagesValidInfo()
	info.Size = int64(len(payload))
	info.Hash = hex.EncodeToString(sha256Digest[:])
	info.Signature = signature
	request := models.DownloadPackageRequest{
		Component: models.ComponentTypeInstaller,
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
	}
	return info, request
}

func TestPackages_PackageInfo_UnwrapsTheEnvelopeAndValidates(t *testing.T) {
	malformed := packagesValidInfo()
	malformed.Hash = "not-a-sha256"

	tests := []struct {
		name    string
		answer  models.PackageInfoResponse
		wantErr string
	}{
		{name: "a valid answer is unwrapped", answer: packagesValidInfo()},
		{name: "an answer that fails validation is refused", answer: malformed, wantErr: "package info answer is malformed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enveloped, err := json.Marshal(map[string]any{"status": "success", "data": tt.answer})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			client := &Client{packageInfo: func(ctx context.Context, query map[string]string) ([]byte, error) {
				if query["component"] != "installer" || query["version"] != "2.0.0" {
					t.Errorf("query = %v, want component installer and version 2.0.0", query)
				}
				return enveloped, nil
			}}

			info, err := client.PackageInfo(context.Background(), models.PackageInfoRequest{
				Component: models.ComponentTypeInstaller,
				Version:   "2.0.0",
				OS:        models.OSTypeLinux,
				Arch:      models.ArchTypeAMD64,
			})
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("PackageInfo() = %+v, %v; want an error mentioning %q", info, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("PackageInfo: %v", err)
			}
			if info.Size != 1024 {
				t.Errorf("size = %d, want 1024 - the envelope was not unwrapped", info.Size)
			}
		})
	}
}

func TestPackages_DownloadPackage_VerifiesWhatItStreams(t *testing.T) {
	payload := []byte(strings.Repeat("installer bytes ", 512))
	info, request := packagesSigned(t, payload)

	tests := []struct {
		name     string
		served   []byte
		wantHint string
	}{
		{name: "a truncated package fails the size check", served: payload[:len(payload)-16], wantHint: "bytes, expected"},
		{name: "a padded package fails the size check", served: append(append([]byte{}, payload...), 'x'), wantHint: "bytes, expected"},
		{name: "same length with other content fails the sha256 check", served: append(append([]byte{}, payload[:len(payload)-1]...), 'X'), wantHint: "sha256"},
		// The throwaway key is not the release key, so an intact package gets as far as the signature.
		{name: "an intact package reaches the signature check", served: payload, wantHint: "signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{downloadPackage: func(ctx context.Context, query map[string]string, w io.Writer) error {
				if query["component"] != "installer" {
					t.Errorf("component = %q, want installer", query["component"])
				}
				_, err := w.Write(tt.served)
				return err
			}}

			var sink bytes.Buffer
			err := client.DownloadPackage(context.Background(), request, info, &sink)
			if err == nil || !strings.Contains(err.Error(), tt.wantHint) {
				t.Fatalf("DownloadPackage() error = %v, want it to mention %q", err, tt.wantHint)
			}
			// Whatever the verdict the destination has the bytes, which is why it stays unverified until nil.
			if !bytes.Equal(sink.Bytes(), tt.served) {
				t.Errorf("the destination received %d bytes, want %d", sink.Len(), len(tt.served))
			}
		})
	}
}

func TestPackages_DownloadPackage_RefusesAMalformedRequestBeforeSending(t *testing.T) {
	called := false
	client := &Client{downloadPackage: func(ctx context.Context, query map[string]string, w io.Writer) error {
		called = true
		return nil
	}}

	// neo4j has no package under any circumstances; pentagi would be valid, it ships an executable.
	request := models.DownloadPackageRequest{
		Component: models.ComponentTypeNeo4j,
		Version:   "2.0.0",
		OS:        models.OSTypeLinux,
		Arch:      models.ArchTypeAMD64,
	}

	err := client.DownloadPackage(context.Background(), request, packagesValidInfo(), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "package download request is malformed") {
		t.Fatalf("DownloadPackage() error = %v, want the request refused as malformed", err)
	}
	if called {
		t.Error("the malformed request was sent anyway")
	}
}
