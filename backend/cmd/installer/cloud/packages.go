package cloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/vxcontrol/cloud/models"
)

// PackageInfo asks the update server about one downloadable package: its size, its
// sha256 and the signature over it.
//
// The answer is validated before it is returned, because everything DownloadPackage checks
// is taken from it — an unverified description of a file is not a basis for trusting the
// file.
func (c *Client) PackageInfo(
	ctx context.Context,
	req models.PackageInfoRequest,
) (*models.PackageInfoResponse, error) {
	if c == nil || c.packageInfo == nil {
		return nil, fmt.Errorf("update client is not initialized")
	}

	if err := req.Valid(); err != nil {
		return nil, fmt.Errorf("package info request is malformed: %w", err)
	}

	answer, err := c.packageInfo(ctx, req.Query())
	if err != nil {
		return nil, Classify(err)
	}

	response, err := models.ParseEnvelope[models.PackageInfoResponse](answer)
	if err != nil {
		return nil, fmt.Errorf("failed to read package info answer: %w", err)
	}

	if err := response.Valid(); err != nil {
		return nil, fmt.Errorf("package info answer is malformed: %w", err)
	}

	return &response, nil
}

// DownloadPackage streams a package into dst, verifying it against info as the bytes
// arrive.
//
// Three things are checked, and all three describe the file itself: the transport
// encryption is transparent and changes none of them. The length must match info.Size —
// the encrypted response carries no Content-Length, so this is the only length the caller
// ever learns. The sha256 must match info.Hash. The Ed25519 signature must verify against
// the file's SHA-512 digest, which is what makes the package ours rather than merely
// intact.
//
// Verification can only complete once the last byte has been written, so on any error dst
// holds a partial or unverified file. It is the caller's job to discard it — write to a
// file that is only put in place after this returns nil.
func (c *Client) DownloadPackage(
	ctx context.Context,
	req models.DownloadPackageRequest,
	info models.PackageInfoResponse,
	dst io.Writer,
) error {
	if c == nil || c.downloadPackage == nil {
		return fmt.Errorf("update client is not initialized")
	}

	if err := req.Valid(); err != nil {
		return fmt.Errorf("package download request is malformed: %w", err)
	}
	if err := info.Valid(); err != nil {
		return fmt.Errorf("package description is malformed: %w", err)
	}

	digest := sha256.New()
	counter := &countingWriter{}
	// The signature wrapper hashes with SHA-512 on the way through; the tee adds the
	// sha256 and the byte count, so a single pass over the stream feeds all three checks.
	signed := info.Signature.ValidateWrapWriter(io.MultiWriter(dst, digest, counter))

	if err := c.downloadPackage(ctx, req.Query(), signed); err != nil {
		return Classify(err)
	}

	if counter.written != info.Size {
		return fmt.Errorf("package is %d bytes, expected %d", counter.written, info.Size)
	}

	actual := hex.EncodeToString(digest.Sum(nil))
	if !strings.EqualFold(actual, info.Hash) {
		return fmt.Errorf("package sha256 is %s, expected %s", actual, info.Hash)
	}

	if err := signed.Valid(); err != nil {
		return fmt.Errorf("package signature does not verify: %w", err)
	}

	return nil
}

// countingWriter counts the bytes passing through it and discards them; the real
// destination sits alongside it in the MultiWriter.
type countingWriter struct {
	written int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.written += int64(len(p))
	return len(p), nil
}
