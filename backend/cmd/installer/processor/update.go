package processor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/cloud"

	"github.com/vxcontrol/cloud/models"
)

// progressStepPercent is how much of the download has to arrive before another line is
// printed. One line per read would be a line per few kilobytes, which buries everything
// around it in a panel that scrolls.
const progressStepPercent = 10

// InstallerPackage is what the update server offers for this host, in the plain form the
// screen showing it can use: the wizard has no business knowing the SDK's types.
//
// Path is where the file will be written, computed here rather than in the interface so
// the name the screen promises and the name the download produces cannot drift apart.
type InstallerPackage struct {
	Version string
	OS      string
	Arch    string
	Size    int64
	Path    string

	// Downloaded says a file already sits at Path. It is the ONLY thing the screen asks
	// the user to confirm: downloading over nothing needs no permission, downloading over
	// a file somebody may have put there on purpose does.
	Downloaded bool

	// CurrentPath is the installer binary running right now, or "" when the operating
	// system would not say. It exists so the screen can print the exact command that puts
	// the download in place — the installer never runs that command itself.
	CurrentPath string
}

// MoveCommand is what the user runs to put the downloaded build in place, or "" when the
// running installer's own path could not be determined.
//
// The installer does not run it, and that is deliberate rather than unfinished: a process
// cannot reliably replace the binary it is executing. Linux refuses the write outright
// (ETXTBSY), Windows holds the file locked, and the failure modes of getting it wrong end
// with the user having no installer at all. Printing the exact command hands that decision
// back with nothing left to guess.
func (p InstallerPackage) MoveCommand() string {
	if p.Path == "" || p.CurrentPath == "" {
		return ""
	}

	// Quoted on both platforms: an installation under "Program Files" or any path with a
	// space silently becomes two arguments otherwise.
	if p.OS == "windows" {
		return fmt.Sprintf(`move /Y "%s" "%s"`, p.Path, p.CurrentPath)
	}

	return fmt.Sprintf(`mv "%s" "%s"`, p.Path, p.CurrentPath)
}

type updateOperationsImpl struct {
	processor *processor

	// described is the last package description the server gave, and the version and
	// platform it describes. It exists because the screen has to say what it is about to
	// download — version and size — before the user agrees to download it, and asking the
	// same question twice costs a second proof-of-work challenge and a second slice of the
	// account's daily allowance.
	//
	// Every path that touches it holds the processor mutex, which is what makes the plain
	// fields safe: installerPackage takes it, and the operations reaching fetchInstaller
	// run under it too.
	described    *models.PackageInfoResponse
	describedKey string
}

func newUpdateOperations(p *processor) updateOperations {
	return &updateOperationsImpl{processor: p}
}

// checkUpdates refreshes what the update server says about this installation.
//
// The answer lands on the shared check result rather than being returned: the menus, the
// update screen and the operations below all read it from there, and a second copy would
// be a second thing to keep current. It is also one call rather than several — every check
// costs a proof-of-work challenge and a slice of the account's daily allowance.
func (u *updateOperationsImpl) checkUpdates(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgCheckingUpdates, ProductStackInstaller, state)

	if err := u.processor.checker.GatherUpdatesInfo(ctx); err != nil {
		return fmt.Errorf("failed to check updates: %w", err)
	}
	if failure := u.processor.checker.UpdateFailure; failure != nil {
		return fmt.Errorf("failed to check updates: %s", failure.Message)
	}

	return nil
}

// downloadInstaller fetches the offered installer build and verifies it, leaving the file
// next to the environment file.
//
// Applying the update is deliberately not part of this: a process cannot reliably replace
// the binary it is executing, and doing it badly leaves the user with no installer at all.
// Downloading and verifying is the whole operation, and it succeeds when the file is on
// disk and proven intact.
func (u *updateOperationsImpl) downloadInstaller(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgDownloadingInstaller, ProductStackInstaller, state)

	path, err := u.fetchInstaller(ctx, state)
	if err != nil {
		return err
	}

	u.processor.appendLog(fmt.Sprintf(MsgInstallerSavedTo, path), ProductStackInstaller, state)
	u.processor.appendLog(MsgInstallerDownloadCompleted, ProductStackInstaller, state)

	return nil
}

// updateInstaller does the same work as downloadInstaller and reports it as an update.
// The two are separate menu entries; the operation behind them is the same, because what
// remains — swapping the running binary — is out of scope.
func (u *updateOperationsImpl) updateInstaller(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgUpdatingInstaller, ProductStackInstaller, state)

	path, err := u.fetchInstaller(ctx, state)
	if err != nil {
		return err
	}

	u.processor.appendLog(fmt.Sprintf(MsgInstallerSavedTo, path), ProductStackInstaller, state)
	u.processor.appendLog(MsgInstallerUpdateCompleted, ProductStackInstaller, state)

	return nil
}

// removeInstaller reports that the binary stays, and succeeds.
//
// A process cannot delete the file it is executing — on Windows it is locked, and on Unix
// the unlink would leave the user with no installer and a half-finished removal. So this
// says where the file is and lets the user decide, exactly as the self-update screen does
// with its move command.
//
// Succeeding matters as much as the message. `remove all` and `purge all` walk every stack
// with the installer LAST, so returning an error here failed the whole operation after all
// the real work had already been done — and in purge it also skipped the docker networks,
// which are removed after that loop.
func (u *updateOperationsImpl) removeInstaller(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgRemovingInstaller, ProductStackInstaller, state)

	if running, err := os.Executable(); err == nil {
		u.processor.appendLog(fmt.Sprintf(MsgInstallerNotRemoved, running), ProductStackInstaller, state)
	} else {
		u.processor.appendLog(MsgInstallerNotRemovedUnknownPath, ProductStackInstaller, state)
	}

	return nil
}

// fetchInstaller downloads the offered build into the environment directory and returns
// where it landed.
//
// The file is written under its final name, in the directory the installation already
// lives in — not through a temporary copy in the system temp directory. Writing an
// executable to %TEMP% and then moving it is a pattern antivirus and EDR products score
// heavily, and a false positive there costs the user their update.
func (u *updateOperationsImpl) fetchInstaller(ctx context.Context, state *operationState) (string, error) {
	version, err := u.installerTargetVersion()
	if err != nil {
		return "", err
	}

	client, err := u.cloudClient()
	if err != nil {
		return "", err
	}

	// The description comes first because everything the download is checked against comes
	// from it: the length, the sha256 and the signature.
	info, err := u.packageInfo(ctx, version)
	if err != nil {
		return "", err
	}

	path := filepath.Join(u.environmentDir(), installerFileName(version))

	// Refusing to write over ourselves. The versioned name normally makes this impossible,
	// but an installer renamed to installer_<its own version> lands exactly here, and the
	// O_TRUNC below would empty the binary the user is running. Linux answers ETXTBSY and
	// macOS does not, so the guard cannot be left to the operating system.
	if running, err := os.Executable(); err == nil && isSameFile(running, path) {
		return "", fmt.Errorf("refusing to overwrite the running installer at %s", path)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", path, err)
	}

	progress := &progressWriter{total: info.Size, report: func(percent int, written, total int64) {
		u.processor.appendLog(
			fmt.Sprintf(MsgInstallerDownloadProgress, percent, FormatSize(written), FormatSize(total)),
			ProductStackInstaller, state,
		)
	}}

	downloadErr := client.DownloadPackage(ctx, models.DownloadPackageRequest{
		Component: models.ComponentTypeInstaller,
		Version:   version,
		OS:        models.OSType(runtime.GOOS),
		Arch:      models.ArchType(runtime.GOARCH),
	}, *info, io.MultiWriter(file, progress))

	if closeErr := file.Close(); closeErr != nil && downloadErr == nil {
		downloadErr = fmt.Errorf("failed to finish writing %s: %w", path, closeErr)
	}

	if downloadErr != nil {
		// Verification only completes once the last byte has arrived, so anything left
		// here is unverified. Leaving it would put a file named like a release next to the
		// installation with nothing marking it as suspect.
		_ = os.Remove(path)
		// The description is the yardstick every one of those checks used, so a build
		// republished under the same version since it was fetched fails them all. Dropping
		// it makes the next attempt ask again instead of failing identically forever.
		u.described, u.describedKey = nil, ""
		return "", fmt.Errorf("failed to download installer: %w", downloadErr)
	}

	u.processor.appendLog(MsgVerifyingBinaryChecksum, ProductStackInstaller, state)

	return path, nil
}

// installerPackage describes the build on offer without downloading it.
//
// The screen behind "Update Installer" needs the version and the size to say what it is
// about to fetch, and it must say that BEFORE the user agrees to fetch it — the old screen
// asked for consent to an operation it described wrongly.
func (u *updateOperationsImpl) installerPackage(ctx context.Context) (*InstallerPackage, error) {
	version, err := u.installerTargetVersion()
	if err != nil {
		return nil, err
	}

	info, err := u.packageInfo(ctx, version)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(u.environmentDir(), installerFileName(version))

	// os.Executable failing is not an error worth stopping for — everything else on the
	// screen is still true, and the only thing lost is the ready-made move command.
	current, _ := os.Executable()

	_, statErr := os.Stat(path)

	return &InstallerPackage{
		Version:     version,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Size:        info.Size,
		Path:        path,
		Downloaded:  statErr == nil,
		CurrentPath: current,
	}, nil
}

// packageInfo asks the update server to describe one installer build, reusing the answer
// already held for the same version and platform.
func (u *updateOperationsImpl) packageInfo(
	ctx context.Context, version string,
) (*models.PackageInfoResponse, error) {
	key := installerPackageKey(version)
	if held := u.cachedPackage(key); held != nil {
		return held, nil
	}

	client, err := u.cloudClient()
	if err != nil {
		return nil, err
	}

	info, err := client.PackageInfo(ctx, models.PackageInfoRequest{
		Component: models.ComponentTypeInstaller,
		Version:   version,
		OS:        models.OSType(runtime.GOOS),
		Arch:      models.ArchType(runtime.GOARCH),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get installer package info: %w", err)
	}

	u.described, u.describedKey = info, key

	return info, nil
}

// installerPackageKey names one downloadable build.
//
// The platform is part of it and not only the version: the same version is published per
// os/arch, and each build has its own size, sha256 and signature — the three values the
// download is checked against. A description matched on version alone would fail every one
// of them on a machine that asked twice from different platforms, which is what a shared
// home directory or a cross-built binary looks like.
func installerPackageKey(version string) string {
	return version + "/" + runtime.GOOS + "/" + runtime.GOARCH
}

// cachedPackage returns the held description when it describes exactly this build.
func (u *updateOperationsImpl) cachedPackage(key string) *models.PackageInfoResponse {
	if u.described == nil || u.describedKey != key {
		return nil
	}

	return u.described
}

// progressWriter turns the bytes going past into an occasional line of progress.
//
// The percentage is against the size from packages/info and not against a Content-Length:
// the response body is encrypted, which removes that header, so this size is the only one
// the client ever learns.
type progressWriter struct {
	total    int64
	written  int64
	reported int
	report   func(percent int, written, total int64)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	w.written += int64(len(p))

	// A stream longer than advertised is the download's business — it fails the length
	// check afterwards. Here it only must not print percentages above a hundred.
	if w.total > 0 && w.report != nil {
		percent := min(int(w.written*100/w.total), 100)
		if step := percent / progressStepPercent; step > w.reported {
			w.reported = step
			w.report(step*progressStepPercent, w.written, w.total)
		}
	}

	return len(p), nil
}

// FormatSize renders a byte count the way a download would be described out loud.
//
// Exported so the screen and the progress line it prints into use the same renderer: they
// end up in the same terminal panel, and two formatters would sooner or later disagree
// about the same number in front of the user.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value, exp := float64(bytes)/unit, 0
	for value >= unit && exp < 3 {
		value /= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", value, [...]string{"KiB", "MiB", "GiB", "TiB"}[exp])
}

// installerTargetVersion reads which build the server offered, from the answer the check
// already obtained.
func (u *updateOperationsImpl) installerTargetVersion() (string, error) {
	for _, stack := range u.processor.checker.StackUpdates {
		if stack.Stack != string(models.ProductStackInstaller) {
			continue
		}
		if version := stack.LatestVersion; version != "" {
			return version, nil
		}
		// The stack entry carries no version of its own on some answers; the artefact for
		// this platform does.
		for _, component := range stack.Components {
			if component.TargetVersion != "" {
				return component.TargetVersion, nil
			}
		}
	}

	return "", fmt.Errorf("update server offered no installer version")
}

// isSameFile reports whether two paths name the same file on disk.
//
// By inode rather than by string: a symlink, a relative path and a different spelling of
// the same directory all compare unequal as text while naming one file. A path that does
// not exist is not the same as anything — which is the common case here, and the reason
// this returns false rather than an error.
func isSameFile(a, b string) bool {
	infoA, err := os.Stat(a)
	if err != nil {
		return false
	}
	infoB, err := os.Stat(b)
	if err != nil {
		return false
	}

	return os.SameFile(infoA, infoB)
}

// installerFileName is what the downloaded build is called. The version is part of the
// name so a download never overwrites the installer currently in use.
func installerFileName(version string) string {
	name := "installer_" + version
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// cloudClient builds a client from this installation's configuration.
func (u *updateOperationsImpl) cloudClient() (*cloud.Client, error) {
	defaultUpdateServerHost := strings.TrimPrefix(strings.TrimSuffix(checker.DefaultUpdateServerEndpoint, "/"), "https://")
	return cloud.New(cloud.Config{
		Host:             u.stateVar("UPDATE_SERVER_HOST", defaultUpdateServerHost),
		ProxyURL:         u.stateVar("PROXY_URL", ""),
		LicenseKey:       u.stateVar("LICENSE_KEY", ""),
		InstallerVersion: checker.InstallerVersion,
	})
}

// environmentDir is where the installation lives — the directory holding the environment
// file, which is also where compose files and stack data sit.
func (u *updateOperationsImpl) environmentDir() string {
	if u.processor.state == nil {
		return "."
	}
	return filepath.Dir(u.processor.state.GetEnvPath())
}

// stateVar reads a configured value, falling back to its default and then to the supplied
// one. PROXY_URL is read through here like everything else: it is the single proxy setting
// the installer keeps, and every call to the update server has to honour it.
func (u *updateOperationsImpl) stateVar(name, fallback string) string {
	if u.processor.state == nil {
		return fallback
	}
	if value, exists := u.processor.state.GetVar(name); exists {
		if value.Value != "" {
			return value.Value
		}
		if value.Default != "" {
			return value.Default
		}
	}
	return fallback
}
