package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"pentagi/cmd/installer/wizard/logger"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
	"github.com/vxcontrol/cloud/models"
)

// productInfoCommand asks the running product to describe its own state.
const productInfoBinary = "/opt/pentagi/bin/pentagi"

// productInfoTimeout bounds the whole exchange.
//
// This is an optional extra on a gathering the user is waiting for. A product
// that does not answer promptly costs nothing beyond a missing field, and must
// never cost the update check itself.
const productInfoTimeout = 20 * time.Second

// maxProductInfoBytes bounds what is read back.
//
// The contract that carries this field refuses anything larger, so reading more
// would only build a request the service rejects whole — losing the answer for
// every component along with it.
//
// It is the contract's own constant rather than a number that agrees with it
// today. This was written as 64 KiB against a 16 KiB contract: a product whose
// description landed in between passed every check here and cost the installation
// the answer for all of its components, and the local test could not see it
// because it measured against this same constant.
const maxProductInfoBytes = models.MaxProductInfoBytes

// gatherProductInfo asks the running product for its state and returns it exactly
// as it was given.
//
// Nothing here interprets the document. It is read, checked for being JSON at
// all, and carried on; whoever consumes it decides what it means. That is not
// laziness — the installer and the product are released separately, so anything
// the installer understood about the shape would become a second place to update
// whenever the product learns to describe more of itself.
//
// Every failure is silent by design and returns nil: no container, an older build
// without the flag, a product that cannot reach its own database, output that is
// not JSON. The update check proceeds without the field, exactly as it did before
// the field existed.
func (h *defaultCheckHandler) gatherProductInfo(ctx context.Context) json.RawMessage {
	if h.dockerClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, productInfoTimeout)
	defer cancel()

	container, err := h.dockerClient.ContainerInspect(ctx, PentagiContainerName, client.ContainerInspectOptions{})
	if err != nil {
		return nil
	}
	if container.Container.State == nil || !container.Container.State.Running {
		// A stopped product cannot describe itself, and the description of a
		// stopped product would not be the one anybody wants anyway.
		return nil
	}

	exec, err := h.dockerClient.ExecCreate(ctx, PentagiContainerName, client.ExecCreateOptions{
		Cmd:          []string{productInfoBinary, "-info"},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		logger.Debugf("product state: exec create failed: %v", err)
		return nil
	}

	attached, err := h.dockerClient.ExecAttach(ctx, exec.ID, client.ExecAttachOptions{})
	if err != nil {
		logger.Debugf("product state: exec attach failed: %v", err)
		return nil
	}
	defer attached.Close()

	// The stream is MULTIPLEXED. Without a TTY the daemon frames every chunk with
	// an eight-byte header saying which stream it came from and how long it is, so
	// reading the socket directly yields JSON with binary interleaved through it —
	// and the first header lands in front of the opening brace, which makes even a
	// perfectly good document unparsable.
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(
		&stdout,
		&stderr,
		io.LimitReader(attached.Reader, maxProductInfoBytes+1),
	); err != nil {
		logger.Debugf("product state: reading the answer failed: %v", err)
		return nil
	}

	// The exit code decides whether the output is an answer or the remains of one.
	// A truncated document that happens to parse would be worse than no document.
	inspected, err := h.dockerClient.ExecInspect(ctx, exec.ID, client.ExecInspectOptions{})
	if err == nil && inspected.ExitCode != 0 {
		logger.Debugf("product state: the product exited with %d", inspected.ExitCode)
		return nil
	}

	return productInfoFromOutput(stdout.Bytes())
}

// productInfoFromOutput decides whether what came back is a document.
//
// Separate from the exchange above so that the decision can be exercised without
// a container: everything that makes this fail in practice — an older build
// printing usage, a half-written answer, an answer larger than the contract
// accepts — is a property of the bytes, not of the daemon.
func productInfoFromOutput(output []byte) json.RawMessage {
	answer := bytes.TrimSpace(output)
	if len(answer) == 0 || len(answer) > maxProductInfoBytes {
		return nil
	}
	if !json.Valid(answer) {
		// An older build has no such flag and prints its usage, or nothing at all.
		// Either way there is no document here.
		return nil
	}

	return json.RawMessage(answer)
}
