package processor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"pentagi/cmd/installer/checker"

	"github.com/vxcontrol/cloud/models"
)

// stagingSuffix is what a plugin file is called while it is still arriving.
//
// The staging file sits in the SAME directory as its target, not in the system temp
// directory: os.Rename is only atomic within one filesystem, and /tmp is routinely a
// different one. It is also a directory antivirus and EDR products watch for executables
// being written and then moved, which is exactly the shape of this operation.
const stagingSuffix = ".new"

// jaegerServiceName is the compose service the plugin is mounted into. Restarting it is
// how a replaced plugin takes effect — the entrypoint exec's the file at start.
const jaegerServiceName = "jaeger"

// updateJaegerPlugin brings the Jaeger storage plugin up to what the server offers.
//
// The plugin is not an image: it is a binary mounted into the Jaeger container, which
// loads it over gRPC. Both architectures are kept in step because which one runs is
// decided inside the container by the architecture of the Docker VM — not by this host —
// and the compose entrypoint picks the file by `uname -m` at start.
//
// Runs BEFORE the images are pulled. A plugin that cannot be fetched then aborts the stack
// update with nothing else disturbed; doing it after would leave images pulled and the
// stack restarted around a plugin that never arrived.
//
// Reports whether anything was actually replaced, because the caller has to restart the
// Jaeger service when it was: `docker compose up -d` recreates a container only when its
// SPEC changes, and a different file inside a bind mount is not a spec change. The plugin
// is exec'd by the container's entrypoint at start, so without a restart the container
// goes on running the plugin it loaded when it came up, and the update looks applied while
// nothing about it is.
func (u *updateOperationsImpl) updateJaegerPlugin(
	ctx context.Context, state *operationState,
) (bool, error) {
	if u.processor.checker == nil {
		return false, nil
	}

	base := filepath.Join(u.environmentDir(), checker.JaegerPluginDir)

	// Always, even when there is nothing to fetch: an interrupted attempt leaves a ~28 MB
	// file that nothing else will ever remove. Integrity verification walks the files the
	// installer ships and does not see strangers, so this is the only thing that looks.
	u.removeStagedPlugins(base, state)

	offered := jaegerPluginUpdates(u.processor.checker.StackUpdates)
	if len(offered) == 0 {
		return false, nil
	}

	for _, component := range offered {
		if err := u.fetchJaegerPlugin(ctx, base, component, state); err != nil {
			return false, err
		}
	}

	return true, nil
}

// jaegerPluginUpdates picks the plugin components that actually need fetching.
//
// Membership in the answer is not the test — under a curated release the server names
// every artefact it knows about, whether or not it differs from what is installed. What
// makes a component here is the digest comparison the checker already did.
func jaegerPluginUpdates(updates []checker.StackUpdate) []checker.ComponentUpdate {
	var wanted []checker.ComponentUpdate
	for _, update := range updates {
		for _, component := range update.Components {
			if component.Component != string(models.ComponentTypeJaegerClickhouse) {
				continue
			}
			// An offered file always carries a version — a published binary cannot exist
			// without one — so the empty case is defence, not a scenario: the package
			// request requires a version and has no field for a hash, so asking without
			// one buys a guaranteed rejection instead of a download.
			if !component.Outdated || component.TargetVersion == "" {
				continue
			}
			// The file name says "linux" in it, so an answer for another platform would
			// put non-linux content under a linux name. The server has never sent one;
			// this makes sure a change on that side is skipped rather than written.
			if component.OS != string(models.OSTypeLinux) {
				continue
			}
			wanted = append(wanted, component)
		}
	}

	return wanted
}

// removeStagedPlugins deletes leftovers from an interrupted attempt.
//
// By exact name rather than by globbing "*.new": this directory is mounted into a
// container and an operator may keep their own files in it, and a delete that guesses is
// how a tool destroys something it was never asked to touch.
func (u *updateOperationsImpl) removeStagedPlugins(base string, state *operationState) {
	for _, name := range checker.JaegerPluginBinaries {
		staged := filepath.Join(base, name+stagingSuffix)
		if _, err := os.Stat(staged); err != nil {
			continue
		}
		if err := os.Remove(staged); err != nil {
			u.processor.appendLog(
				fmt.Sprintf(MsgJaegerPluginStaleRemoveFailed, staged, err),
				ProductStackObservability, state)
			continue
		}
		u.processor.appendLog(fmt.Sprintf(MsgJaegerPluginStaleRemoved, staged),
			ProductStackObservability, state)
	}
}

// fetchJaegerPlugin downloads one architecture's plugin and puts it in place.
func (u *updateOperationsImpl) fetchJaegerPlugin(
	ctx context.Context, base string, component checker.ComponentUpdate, state *operationState,
) error {
	arch := models.ArchType(component.Arch)
	name, known := checker.JaegerPluginBinaries[arch]
	if !known {
		return fmt.Errorf("no jaeger plugin file is defined for architecture %s", component.Arch)
	}

	u.processor.appendLog(fmt.Sprintf(MsgJaegerPluginDownloading, name, component.TargetVersion),
		ProductStackObservability, state)

	client, err := u.cloudClient()
	if err != nil {
		return err
	}

	request := models.PackageInfoRequest{
		Component: models.ComponentTypeJaegerClickhouse,
		Version:   component.TargetVersion,
		OS:        models.OSTypeLinux,
		Arch:      arch,
	}
	info, err := client.PackageInfo(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to get jaeger plugin package info for %s: %w", arch, err)
	}

	progress := &progressWriter{total: info.Size, report: func(percent int, written, total int64) {
		u.processor.appendLog(
			fmt.Sprintf(MsgJaegerPluginProgress, name, percent, FormatSize(written), FormatSize(total)),
			ProductStackObservability, state)
	}}

	target := filepath.Join(base, name)
	err = installPluginFile(target, func(w io.Writer) error {
		return client.DownloadPackage(ctx, models.DownloadPackageRequest{
			Component: models.ComponentTypeJaegerClickhouse,
			Version:   component.TargetVersion,
			OS:        models.OSTypeLinux,
			Arch:      arch,
		}, *info, io.MultiWriter(w, progress))
	})
	if err != nil {
		return fmt.Errorf("failed to install jaeger plugin for %s: %w", arch, err)
	}

	u.processor.appendLog(fmt.Sprintf(MsgJaegerPluginInstalled, target),
		ProductStackObservability, state)

	return nil
}

// installPluginFile streams into <target>.new and moves it over target only once write
// has returned without complaint.
//
// Staged rather than written in place, because the file being replaced is the one a
// running Jaeger loads: a download that dies halfway would otherwise leave a truncated
// plugin, and the container would come back up unable to start with an opaque gRPC error.
// The staging file is a sibling of the target so the rename stays within one filesystem,
// where it is atomic.
func installPluginFile(target string, write func(io.Writer) error) error {
	// The stack may never have been installed, in which case the directory the mount
	// expects does not exist yet.
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(target), err)
	}

	staged := target + stagingSuffix

	// 0755 at CREATION and not afterwards: os.Rename does not preserve the destination's
	// mode — the result keeps the staging file's — so a plugin created 0644 and renamed
	// over an executable one ends up not executable, and Jaeger fails to start with an
	// opaque gRPC plugin error rather than a permission one.
	file, err := os.OpenFile(staged, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", staged, err)
	}

	writeErr := write(file)

	if closeErr := file.Close(); closeErr != nil && writeErr == nil {
		writeErr = fmt.Errorf("failed to finish writing %s: %w", staged, closeErr)
	}

	if writeErr != nil {
		// The length, the sha256 and the signature are only decided once the last byte has
		// arrived, so what is on disk here is unverified. The target is untouched — that is
		// the entire reason for staging — and the installation keeps the plugin it had.
		_ = os.Remove(staged)
		return writeErr
	}

	if err := os.Rename(staged, target); err != nil {
		_ = os.Remove(staged)
		return fmt.Errorf("failed to put %s in place: %w", target, err)
	}

	return nil
}
