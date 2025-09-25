# Processor Implementation Summary

## Overview
Processor package implements the operational engine for PentAGI installer operations per [processor.md](processor.md). Core lifecycle flows, file integrity logic, Docker/Compose orchestration, and Bubble Tea integration are implemented. Installer self-update flows are stubbed and intentionally not finalized yet.

## Implementation Notes

### Architecture Decisions
- **Interface-based design**: Internal interfaces per operation type (`fileSystemOperations`, `dockerOperations`, `composeOperations`, `updateOperations`) enable separation of concerns and testability
- **Two-track execution**: Docker API SDK for worker environment; Compose stacks via console commands with live output streaming
- **OperationOption pattern**: Functional options applied to an internal `operationState` support force mode and embedded terminal integration via `WithForce` and `WithTerminal`
- **State machine logic**: `ApplyChanges` implements three-phase stack management (Observability → Langfuse → PentAGI) with integrity validation; the wizard performs a pre-phase interactive integrity check with Y/N decision for force mode
- **Single-responsibility operations**: business logic delegates to compose layer; strict purge of images is implemented as `purgeImagesStack` alongside other compose operations

### Key Features Implemented
1. **File System Operations** (`fs.go`):
   - Ensure/verify stack file integrity with force mode support
   - Handle embedded directory trees (observability) and compose files
   - YAML validation and automatic file recovery
   - Support deployment modes (embedded/external/disabled) for applicable stacks
   - Excluded files policy for integrity verification: `observability/otel/config.yml`, `observability/grafana/config/grafana.ini`, `example.provider.yml` (presence ensured, content changes tolerated)

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

4. **Update Operations** (`update.go`):
   - Update server communication and binary replacement helpers are scaffolded (checksum, atomic replace/backup)
   - Installer update/download/remove operations are currently stubs and return "not implemented"; network calls use placeholder logic for now

5. **Remove/Purge Operations**:
   - Soft removal (preserve data) vs purge (complete cleanup); strict image purge via compose in `purgeImagesStack`
   - Proper cleanup ordering and external/existing deployment handling

### Critical Implementation Details
- **Three-phase execution**: Observability → Langfuse → PentAGI with state validation after each phase
- **Force mode behavior**: Aggressive file overwriting and state correction when explicitly requested
- **File integrity logic**: `ensureStackIntegrity` for missing files, `verifyStackIntegrity` for existing files; modified files are explicitly skipped and logged when `force=false`; excluded files are ensured to exist but not overwritten when modified
- **State consistency**: `Gather*Info` calls after each phase validate operation success
- **Error isolation**: Phase failures don't affect other stacks, partial state preserved
- **Compose environment tweaks**: `COMPOSE_IGNORE_ORPHANS=1`, `PYTHONUNBUFFERED=1`; ANSI disabled on narrow terminals via `COMPOSE_ANSI=never`

### Testing Strategy
Comprehensive tests include:
- Mock implementations for external dependencies (state, checker, files)
- Unit tests for file system integrity operations (ensure/verify/cleanup, excluded files policy, YAML validation)
- Validation tests for operation applicability
- Factory reset, lifecycle, and ordering behavior at logic level

### Integration Points
- **State management**: Integrates with `state.State` for configuration and environment variables
- **System assessment**: Uses `checker.CheckResult` for current system state analysis
- **File handling**: Integrates with `files.Files` for embedded content extraction
- **TUI integration**: Bubble Tea integration via `ProcessorModel` with message polling; wizard performs pre-phase integrity scan (Enter → scan; Y/N → overwrite decision; Ctrl+C → cancel integrity stage)

## Files Created/Modified

### Core Implementation
- `processor.go` - Processor interface, options, and synchronous operations entry points
- `model.go` - Bubble Tea `ProcessorModel` with `HandleMsg` polling
- `logic.go` - Business logic (ApplyChanges, lifecycle operations, factory reset)
- `fs.go` - File system operations and integrity verification
- `docker.go` - Docker API/CLI operations and worker image/volumes management
- `compose.go` - Docker Compose stack lifecycle management (including `purgeImagesStack`)
- `update.go` - Scaffolding for update mechanisms (stubs for installer update flows)

### Testing
- `mock_test.go` - Mocks for interfaces with call tracking
- `logic_test.go` - Business logic tests (state machine and sequencing)
- `fs_test.go` - File system operations tests (including excluded files policy)

## Status
✅ **MOSTLY COMPLETE** - Core processor functionality implemented and tested
- Lifecycle, file integrity, Docker/Compose orchestration are production-ready
- Bubble Tea integration via `ProcessorModel` is complete
- Installer self-update flows (download/update/remove) are stubbed and not enabled yet
- All current tests pass; additional tests will be added once update flows are finalized

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
    HandleMsg(msg tea.Msg) tea.Cmd
}
```

### Message Types
Processor operations communicate via messages:
- `ProcessorStartedMsg` - operation started
- `ProcessorOutputMsg` - command output (partial or full)
- `ProcessorFilesCheckMsg` - file statuses computed during `CheckFiles`
- `ProcessorCompletionMsg` - operation completed
- `ProcessorWaitMsg` - polling tick

### Terminal Integration
Operations support real-time terminal output through `WithTerminal(term terminal.Terminal)`; ANSI is auto-disabled for narrow terminals. Compose commands inherit current env plus `COMPOSE_IGNORE_ORPHANS=1` and `PYTHONUNBUFFERED=1`.
