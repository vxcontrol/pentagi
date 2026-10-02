# Processor Implementation Summary

## Overview
Processor package implements the operational engine for PentAGI installer operations per [processor.md](processor.md). Core lifecycle flows, file integrity logic, Docker/Compose orchestration, and Bubble Tea integration are implemented. Cloud-facing work — the update check, the installer download and the Jaeger storage plugin — is implemented too and talks to the PentAGI Cloud API through `cmd/installer/cloud`. Post-update verification is implemented as well, but it makes no request of its own: it compares the components observed after the update against targets captured locally from the answer that was already in hand.

## Implementation Notes

### Architecture Decisions
- **Interface-based design**: Internal interfaces per operation type (`fileSystemOperations`, `dockerOperations`, `composeOperations`, `updateOperations`) enable separation of concerns and testability
- **Two-track execution**: Docker API SDK for worker environment; Compose stacks via console commands with live output streaming
- **OperationOption pattern**: Functional options applied to an internal `operationState` support force mode, embedded terminal integration and the reset-password value via `WithForce`, `WithTerminal` and `WithPasswordValue`
- **State machine logic**: `ApplyChanges` implements four-phase stack management (Observability → Langfuse → Graphiti → PentAGI) with integrity validation; the wizard performs a pre-phase interactive integrity check with Y/N decision for force mode
- **Single-responsibility operations**: business logic delegates to compose layer; strict purge of images is implemented as `purgeImagesStack` alongside other compose operations
- **One lock for everything**: every entry point on `Processor` takes the processor mutex, including `InstallerPackage`, which changes nothing. It takes it because it *remembers* something the download that follows will read — see [Installer package description](#installer-package-description)

### Key Features Implemented
1. **File System Operations** (`fs.go`):
   - Ensure/verify stack file integrity with force mode support
   - Handle embedded directory trees (observability) and compose files
   - YAML validation and automatic file recovery
   - Support deployment modes (embedded/external/disabled) for applicable stacks
   - Excluded files policy for integrity verification (`filesToExcludeFromVerification`): presence ensured, content changes tolerated — service configs (`observability/otel/config.yml`, `observability/grafana/config/grafana.ini`), user-editable presets (`example.custom.provider.yml`, `example.ollama.provider.yml`, `example.bedrock.provider.yml`, `neo4j/conf/neo4j.conf`, `neo4j/conf/apoc.conf`, `graphiti/{custom,gemini,litellm,openai}.yaml`) and both Jaeger storage plugin binaries

2. **Docker Operations** (`docker.go`):
   - Worker and default image management with progress reporting
   - Worker container lifecycle management (removal, purging)
   - Support for custom Docker configuration via environment variables

3. **Compose Operations** (`compose.go`):
   - Stack lifecycle management with dependency ordering
   - Rolling updates with health checks
   - Live output streaming to TUI callbacks
   - Environment variable injection for compose commands
    - `purgeStack` (down -v) and `purgeImagesStack` (down --rmi all -v) placed together for clarity

4. **Update Operations** (`update.go`, `jaeger_plugin.go`, `verify.go`):
   - `checkUpdates` refreshes the shared `checker.CheckResult` through `GatherUpdatesInfo` and returns nothing: the menus, the update screens and the operations all read the answer from there, and a second copy would be a second thing to keep current. Nothing calls it today — the answer is refreshed by `GatherAllInfo` at startup and by the re-gather every operation does when it finishes — so it is an entry point on the interface rather than a step on any current code path
   - `downloadInstaller` / `updateInstaller` fetch the offered build over `/api/v1/proxy/packages/info` → `/api/v1/proxy/packages/download` and leave it beside the environment file as `installer_<version>[.exe]`; both are the same operation reported under two names, because what would separate them — swapping the running binary — is out of scope
   - `installerPackage` describes the offered build without downloading it, for the screen that asks the user to agree to the download
   - `updateJaegerPlugin` fetches the Jaeger storage plugin, which is a mounted file rather than an image and so is invisible to `docker compose pull`
   - `removeInstaller` is the one member of the interface still returning "not implemented". It is reached by `Remove`/`Purge` on the installer stack, and the `all` fan-out visits that stack last, so both maintenance screens report a failure from their final step. Under "Remove" that is the whole cost: everything ahead of it has already run and nothing follows the loop. Under "Purge" it costs more — the error aborts the fan-out, and what sits after the loop never happens: the three custom Docker networks are not removed, and the deferred refresh of the update answer is skipped because it returns early on a failed operation

5. **Remove/Purge Operations**:
   - Soft removal (preserve data) vs purge (complete cleanup); strict image purge via compose in `purgeImagesStack`
   - Proper cleanup ordering and external/existing deployment handling

### Critical Implementation Details
- **Four-phase execution**: Observability → Langfuse → Graphiti → PentAGI with state validation after each phase
- **Force mode behavior**: Aggressive file overwriting and state correction when explicitly requested
- **File integrity logic**: `ensureStackIntegrity` for missing files, `verifyStackIntegrity` for existing files; modified files are explicitly skipped and logged when `force=false`; excluded files are ensured to exist but not overwritten when modified
- **State consistency**: `Gather*Info` calls after each phase validate operation success
- **Error isolation**: Phase failures don't affect other stacks, partial state preserved
- **Compose environment tweaks**: `COMPOSE_IGNORE_ORPHANS=1`, `PYTHONUNBUFFERED=1`; ANSI disabled on narrow terminals via `COMPOSE_ANSI=never`
- **Ordering inside a stack update**: the Jaeger plugin first — for observability only — then `downloadStack`, then `updateStack`, then a `restart` of the jaeger service when the plugin was actually replaced, and last the refresh of the answer. Writing the pull references is not a step in this sequence: `applyPullReferences` runs at the top of `downloadStack` and `updateStack` themselves, the only two compose operations that fetch anything, rather than at each of the five call sites that reach them — a call site that forgot would not fail to compile, because `${<COMPONENT>_IMAGE:-<default>}` quietly resolves to the default the compose file ships with and the operation would report success having fetched something the update server never offered. The plugin goes first because it is a mounted file no pull can bring up to date, and one that cannot be fetched must abort the update with nothing else disturbed, rather than leave images pulled and the stack restarted around a plugin that never arrived. The restart goes last because `up -d` recreates a container only when its spec changed, and a different file inside a bind mount is not a spec change — without it Jaeger goes on serving through the plugin it exec'd when it started, and the update reports success anyway
- **The Jaeger plugin is exempt from integrity verification**: both plugin paths are in `filesToExcludeFromVerification`, because the embedded copy and the update server are two sources for one path. Without the exemption a downloaded plugin reads as "modified", and a stack update — which runs with force — copies the embedded version back over it, undoing itself

### Operation Interfaces
The four internal interfaces in `processor.go` are how the business logic in `logic.go` reaches anything external. A method takes the `*operationState` when it produces output, because that is what carries the terminal panel and the message chain; the three that produce a value instead of output — `checkStackIntegrity`, `determineComposeFile` and `installerPackage` — do not:

```go
type updateOperations interface {
    checkUpdates(ctx context.Context, state *operationState) error
    downloadInstaller(ctx context.Context, state *operationState) error
    updateInstaller(ctx context.Context, state *operationState) error
    removeInstaller(ctx context.Context, state *operationState) error
    installerPackage(ctx context.Context) (*InstallerPackage, error)
    updateJaegerPlugin(ctx context.Context, state *operationState) (bool, error)
}
```

`installerPackage` is the odd one out here: it produces a value for a screen to render rather than output for a panel to show. `updateJaegerPlugin` is both at once: it takes a `state` because it downloads tens of megabytes and has to say so while it does, and it returns a `bool` beside the error — whether it actually replaced a file — because that is what decides the jaeger restart afterwards.

`updateOperationsImpl` is the only implementation and holds two mutable fields — `described` and `describedKey`, the remembered package description. They are plain fields rather than guarded ones because every path that touches them runs under the processor mutex: `installerPackage` takes it in `processor.InstallerPackage`, and the operations that reach `fetchInstaller` run under it too.

### Streaming Output Into the Terminal Panel
There are two producers of panel content, and they differ in whether the output is a whole view or a single line.

- `runCommand` hands the `*exec.Cmd` to `state.terminal.Execute`, waits for both the process and the terminal, and then sends `terminal.View()` as a **full** replacement (`isPartial=false`). Compose redraws its own progress in place, so the panel has to be the terminal's rendering rather than an append-only log.
- `appendLog` writes one line through `terminal.Append` and again sends the whole `View()`. This is what every cloud-facing operation uses: download progress, the plugin lines, the verification verdicts.

Both end at `operationState.sendOutput`, which appends a `ProcessorOutputMsg` to `state.msgs` under `state.mx`, stamped with its own 1-based position. Nothing is pushed to Bubble Tea — `HandleMsg` polls: handed a message numbered *n*, it returns `msgs[n]` (the next one) when it has arrived, and otherwise a `ProcessorWaitMsg` carrying the same *n* after 100 ms, which brings it back. The operation itself runs in a goroutine started by `wrapCommand`, which returns the first `ProcessorWaitMsg` as soon as the operation finishes or after 500 ms, whichever comes first. `ProcessorCompletionMsg` ends the chain by returning no command, and a `ProcessorWaitMsg` carrying an error stops the polling too. Walking an accumulated slice rather than draining a channel is what keeps the chain ordered and replayable when the screen is slower than the operation.

When no terminal is supplied, `newOperationState` builds a headless one (80×24, `WithNoPty`), so an operation invoked outside the wizard still produces the same message chain.

### Installer Package Description
`InstallerPackage` is the plain form of what the update server offers for this host — version, OS, arch, size — plus three fields the screen cannot compute for itself:

- `Path`: where the download will land, `installer_<version>[.exe]` in the directory holding `.env`. It is computed here rather than by the caller so that the name the screen promises and the name the download produces cannot drift apart.
- `Downloaded`: whether a file already sits at `Path`. It is the only thing the screen asks the user to confirm — downloading over nothing needs no permission, downloading over a file somebody may have put there on purpose does.
- `CurrentPath`: the running binary, or `""` when the OS would not say. `MoveCommand()` turns the pair into the exact `mv` (`move /Y` on Windows) the user runs afterwards, with both paths quoted so an installation under a directory with a space does not become two arguments. The installer never runs it: a process cannot reliably replace the binary it is executing — Linux answers `ETXTBSY`, Windows holds the file locked — so the operation honestly ends at "downloaded and verified".

**Why the description is reused by the download.** The screen has to state the size *before* the user agrees, which means `packages/info` is called on entering the screen and not only inside the download. Asking twice would cost a second proof-of-work challenge and a second slice of the account's daily allowance for one keypress, so `packageInfo` caches the answer in `described` and `fetchInstaller` reuses it.

The cache key is `version/GOOS/GOARCH`, not the version alone: the same version is published per platform with its own size, sha256 and signature, so a description of another platform's build would fail all three checks with a message about a corrupt download. On a failed download the description is dropped along with the partial file — it was the yardstick every check used, so a build republished under the same version would otherwise make every retry fail identically forever.

### Where Downloaded Files Are Written
Neither download goes through the system temp directory. Writing an executable to `%TEMP%` and then moving it is a pattern antivirus and EDR products score heavily, and a false positive there costs the user the update.

- **The installer** is written straight under its final name in the directory the installation lives in. Before opening it, `fetchInstaller` refuses to write to a path that is the running binary, comparing by inode rather than by string — a symlink or a different spelling of the same directory compares unequal as text while naming one file. The versioned name normally makes the collision impossible, but an installer renamed to `installer_<its own version>` lands exactly there, and `O_TRUNC` would empty the binary the user is running: Linux answers `ETXTBSY`, macOS does not.
- **The Jaeger plugin** is staged as `<target>.new` in the *same* directory as its target and moved with `os.Rename`, which is atomic only within one filesystem. Staging matters because the file being replaced is the one a running Jaeger has loaded; an in-place write cut short would leave a truncated plugin and the container would come back up with an opaque gRPC error. Mode `0755` is set at creation and not afterwards — `os.Rename` does not preserve the destination's mode, so a plugin created `0644` and renamed over an executable one ends up not executable, and Jaeger again fails with a gRPC error rather than a permission one.

On failure both paths delete what they wrote: verification only completes on the last byte, so anything left behind is unverified. Leftover `.new` files from an interrupted run are removed at the start of `updateJaegerPlugin` — always, even when there is nothing to fetch, because integrity verification only walks the files the installer ships and would never see a ~28 MB stranger. They are deleted by exact name rather than by globbing `*.new`: the directory is mounted into a container and an operator may keep their own files in it.

Both architectures of the plugin are kept in step regardless of the host, because which one runs is decided inside the container by the Docker VM's architecture — the compose entrypoint picks the file by `uname -m` at start. Only components the check marked outdated are fetched: membership in the answer is not the test, since a curated release names every artefact it knows about, and fetching everything listed would mean ~28 MB per architecture on every observability update. A component whose OS is not linux is skipped — the file name says `linux`, and content for another platform would be written under a linux name.

### Download Progress
`progressWriter` sits in an `io.MultiWriter` beside the destination file, so the bytes are counted on their way past without a second pass.

Its percentage is computed against `info.Size` from the package description, **not** a `Content-Length`: the response body is encrypted, which removes that header, so the description is the only size the client ever learns. This is also why `installerPackage` and `fetchInstaller` both need a description in hand before anything can start.

Two details are deliberate. A line is printed once per `progressStepPercent` (10%) rather than per read — a line every few kilobytes would bury the file path and the verification verdict in a panel that scrolls. And the percentage is clamped to 100: a stream longer than advertised is caught by the length check in `cloud.DownloadPackage`, and it is not the indicator's job to report it. With `total == 0` the writer stays silent instead of dividing by zero.

`FormatSize` is exported so the screen and the progress lines it prints into use one renderer; they end up in the same panel, and two formatters would eventually disagree about the same number in front of the user.

`cloud.DownloadPackage` checks three things in the same single pass the progress writer counts in — the byte count against `info.Size`, the sha256 against `info.Hash`, and an Ed25519 signature over the file's SHA-512 digest. The verdict therefore already exists when the call returns, which is why the line printed after it is past tense: `Checksum and signature verified`.

### Post-Update Verification
`verify.go` answers one question after a stack update: did the artefacts now installed become the ones that were offered? Targets are captured **before** the update (`captureStackTargets`), because the deferred `GatherUpdatesInfo` replaces the answer — comparing against the refreshed one would compare the result with itself and always agree.

The comparison runs inside the same `defer`, after the refresh, because the fresh answer is what decides whether anything is still pending. That distinction is made on evidence rather than heuristics: when a component differs from its captured target *and* the fresh answer no longer offers an update for the stack, the registry moved ahead between the answer and the pull, and the line says so; when the offer is still there, it is a mismatch. A failed refresh has no appeal — `stackHasNoUpdate` returns false — and a difference counts as a mismatch.

Verification never fails the operation. The update either worked or already reported an error; turning "the registry moved ahead" into a failed update would be worse than saying nothing. The installer stack is excluded from `verifiableUpdateStacks` on purpose: its operation downloads a build without applying it, so comparing against the running binary would report the intended outcome as a mismatch.

### Testing Strategy
Comprehensive tests include:
- Mock implementations for external dependencies (state, checker, files)
- Unit tests for file system integrity operations (ensure/verify/cleanup, excluded files policy, YAML validation)
- Validation tests for operation applicability
- Factory reset, lifecycle, and ordering behavior at logic level
- `update_test.go` — progress stepping and clamping, silence without a size, the description key, cleanup of file and description after a failed download, the versioned file name, the inode comparison guarding the running binary, `MoveCommand` quoting, `FormatSize`
- `jaeger_plugin_test.go` — only outdated linux components are fetched, a failed download leaves the running plugin alone, the installed plugin is executable, staging leftovers are removed by exact name
- `verify_test.go` — targets captured before the answer is replaced, an already current component's target is what it runs, an unverifiable component carries no target, "newer" is distinguished from "wrong", the installer is not verified against the running binary
- `pull_reference_test.go` — the compose variable is derived from the component name, only resolved components get a reference, and every `image:` in the shipped compose files is parameterised

### Integration Points
- **State management**: Integrates with `state.State` for configuration and environment variables
- **System assessment**: Uses `checker.CheckResult` for current system state analysis
- **File handling**: Integrates with `files.Files` for embedded content extraction
- **TUI integration**: Bubble Tea integration via `ProcessorModel` with message polling; wizard performs pre-phase integrity scan (Enter → scan; Y/N → overwrite decision; Ctrl+C → cancel integrity stage)

## Files Created/Modified

### Core Implementation
- `processor.go` - Processor interface, options, operation interfaces, and synchronous operations entry points
- `model.go` - Bubble Tea `ProcessorModel` with `HandleMsg` polling
- `logic.go` - Business logic (ApplyChanges, lifecycle operations, factory reset)
- `state.go` - `operationState`, message types and the `sendOutput`/`sendCompletion` chain
- `fs.go` - File system operations and integrity verification
- `docker.go` - Docker API/CLI operations and worker image/volumes management
- `compose.go` - Docker Compose stack lifecycle management (including `purgeImagesStack`)
- `update.go` - Update check, installer package description and download, progress reporting
- `jaeger_plugin.go` - Jaeger storage plugin download with staged install
- `verify.go` - Post-update comparison of installed artefacts against what was offered
- `pull_reference.go` - Writing the server-chosen image reference into `.env` before a pull
- `reference_drift.go` - Reporting compose variables that differ from what this build ships
- `locale.go` - Operation message strings

### Testing
- `fixtures_test.go` - Mocks for interfaces with call tracking
- `logic_test.go` - Business logic tests (state machine and sequencing)
- `fs_test.go` - File system operations tests (including excluded files policy, the Jaeger plugin among them)
- `update_test.go`, `jaeger_plugin_test.go`, `verify_test.go`, `pull_reference_test.go`, `reference_drift_test.go`, `compose_test.go`, `pg_test.go`, `state_test.go`, `model_test.go`

## Status
✅ **COMPLETE** - Processor functionality implemented and tested, with one known gap
- Lifecycle, file integrity, Docker/Compose orchestration are production-ready
- Bubble Tea integration via `ProcessorModel` is complete
- Installer download/update, Jaeger plugin download and post-update verification are implemented and covered by tests
- `removeInstaller` remains unimplemented; because `Remove`/`Purge` on `all` end on the installer stack, both maintenance screens finish with a "not implemented" error. Under "Remove" the real work is already done by then; under "Purge" the failure also skips everything after the fan-out — the three custom Docker networks stay in place and the update answer is not refreshed

## Current Architecture

### ProcessorModel Integration
The processor integrates with Bubble Tea through `ProcessorModel` that wraps operations as `tea.Cmd` and provides a polling handler:

```go
// ProcessorModel provides tea.Cmd wrappers for all operations and a polling handler
type ProcessorModel interface {
    ApplyChanges(ctx context.Context, opts ...OperationOption) tea.Cmd
    CheckFiles(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    FactoryReset(ctx context.Context, opts ...OperationOption) tea.Cmd
    Install(ctx context.Context, opts ...OperationOption) tea.Cmd
    Update(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    Download(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    Remove(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    Purge(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    Start(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    Stop(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    Restart(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    ResetPassword(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
    HandleMsg(msg tea.Msg) tea.Cmd

    // InstallerPackage describes the installer build on offer without downloading it.
    InstallerPackage(ctx context.Context) (*InstallerPackage, error)
}
```

`InstallerPackage` is blocking rather than a `tea.Cmd`, unlike everything above it: it is one short request with one answer, not an operation that streams output into a terminal panel, and the screen that needs it already wraps it in a command of its own.

### Message Types
Processor operations communicate via messages:
- `ProcessorStartedMsg` - operation started
- `ProcessorOutputMsg` - command output (partial or full)
- `ProcessorFilesCheckMsg` - file statuses computed during `CheckFiles`
- `ProcessorCompletionMsg` - operation completed
- `ProcessorWaitMsg` - polling tick

### Terminal Integration
Operations support real-time terminal output through `WithTerminal(term terminal.Terminal)`; ANSI is auto-disabled for narrow terminals (< 56 columns) and on Windows. Compose commands inherit current env plus `COMPOSE_IGNORE_ORPHANS=1` and `PYTHONUNBUFFERED=1`. Without the option an operation still runs, against a headless 80×24 terminal built by `newOperationState`. See [Streaming Output Into the Terminal Panel](#streaming-output-into-the-terminal-panel) for how a line becomes a message.
