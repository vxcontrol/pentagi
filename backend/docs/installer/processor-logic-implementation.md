# Processor Business Logic Implementation

## Core Concepts

The processor package manages PentAGI stack lifecycle through state-oriented approach. Main idea: maintain consistency between desired user state (state.State) and actual system state (checker.CheckResult).

## Key Principles

1. **State-Driven Operations**: All operations based on comparing current and target state
2. **Stack Independence**: Each stack (observability, langfuse, pentagi) managed independently
3. **Force Mode**: Aggressive state correction ignoring warnings
4. **Idempotency**: Repeated operation calls do not cause side effects
5. **User-Facing Automation**: Installer automates manual Docker/file operations with real-time feedback

## Stack Architecture

### ProductStack Hierarchy
```
ProductStackAll
├── ProductStackObservability (optional, embedded/external/disabled)
├── ProductStackLangfuse (optional, embedded/external/disabled)
└── ProductStackPentagi (mandatory, always embedded)
```

### Deployment Modes
- **Embedded**: Full local stack with Docker Compose files
- **External**: Using external service (configuration only)
- **Disabled**: Functionality turned off

## ApplyChanges Algorithm

### Purpose
Bring system to state matching user configuration. Main installation/update/configuration function.

### Prerequisites
- docker accessibility is validated via `checker.CheckResult` flags
- required compose networks are ensured via `ensureMainDockerNetworks` before stack operations

### Implementation Strategy

#### Pre-phase: interactive file integrity check (wizard)
- before starting ApplyChanges, the wizard runs an integrity scan using processor file checking helpers
- if outdated or missing files are detected, the user is prompted to choose:
  - proceed with updates (force=true) – modified files will be overwritten from embedded content;
  - proceed without updates (force=false) – installer will try to apply changes without touching modified files.
- hotkeys on the screen: Enter (start integrity scan), then Y/N to choose scenario; Ctrl+C cancels the integrity stage and returns to the initial prompt.

#### Phase 1: Observability Stack Management
```go
if p.isEmbeddedDeployment(ProductStackObservability) {
    // user wants embedded observability
    if !p.checker.ObservabilityExtracted {
        // extract files (docker-compose + observability directory)
         p.fsOps.ensureStackIntegrity(ctx, ProductStackObservability, state)
    } else {
        // verify file integrity, update if force=true
         p.fsOps.verifyStackIntegrity(ctx, ProductStackObservability, state)
    }
    // update/start containers
     p.composeOps.updateStack(ctx, ProductStackObservability, state)
} else {
    // user wants external/disabled observability
    if p.checker.ObservabilityInstalled {
        // remove containers but keep files (user might re-enable)
         p.composeOps.removeStack(ctx, ProductStackObservability, state)
    }
}
// refresh state to verify operation success
p.checker.GatherObservabilityInfo(ctx)
```

**Rationale**: Observability processed first as most complex stack (directory + compose file). Force mode used for file conflict resolution. For external/disabled modes containers are removed but files are preserved.

#### Phase 2: Langfuse Stack Management
```go
if p.isEmbeddedDeployment(ProductStackLangfuse) {
    // only docker-compose-langfuse.yml file
     p.fsOps.ensureStackIntegrity(ctx, ProductStackLangfuse, state)
     p.composeOps.updateStack(ctx, ProductStackLangfuse, state)
} else {
    if p.checker.LangfuseInstalled {
         p.composeOps.removeStack(ctx, ProductStackLangfuse, state)
    }
}
p.checker.GatherLangfuseInfo(ctx)
```

**Rationale**: Langfuse simpler than observability (single file only), but follows same logic. As a precondition for local start, configuration must be connected (see checker `LangfuseConnected`).

#### Phase 3: PentAGI Stack Management
```go
// PentAGI always embedded, always required
p.fsOps.ensureStackIntegrity(ctx, ProductStackPentagi, state)
p.composeOps.updateStack(ctx, ProductStackPentagi, state)
p.checker.GatherPentagiInfo(ctx)
```

**Rationale**: PentAGI - main stack, always installed, only file integrity check.

### Critical Implementation Details

#### File System Integrity
- **ensureStackIntegrity**: creates missing files from embed, overwrites with force=true
- **verifyStackIntegrity**: checks existence, updates with force=true
- **Embedded Provider**: uses `files.Files` for embedded content access
- correctness of directory checks: when modified files are detected and force=false, we log skip explicitly and keep files intact; the final directory log reflects whether modified files were present
- excluded files policy: `observability/otel/config.yml`, `observability/grafana/config/grafana.ini`, `example.provider.yml` are ensured to exist but not overwritten if modified

#### Container State Management
- `updateStack`: executes `docker compose up -d` for rolling update
- `removeStack`: executes `docker compose down` without removing volumes
- `purgeStack`: executes `docker compose down -v`
- `purgeImagesStack`: executes `docker compose down --rmi all -v`
- dependency ordering: observability → langfuse → pentagi
- environment: `COMPOSE_IGNORE_ORPHANS=1`, `PYTHONUNBUFFERED=1`; ANSI disabled on narrow terminals via `COMPOSE_ANSI=never`

#### State Consistency
- After each phase corresponding `Gather*Info()` method called
- CheckResult updated for next decisions
- On errors state remains partially updated (no rollback)
 - optimization with `state.IsDirty()`: optional early exit can be used by the UI to avoid unnecessary work; installer remains consistent without it.

## Force Mode Behavior

### Normal Mode (force=false)
- Does not overwrite existing files
- Stops on filesystem conflicts
- Conservative approach, minimal changes

### Force Mode (force=true)
- Overwrites any files without warnings
- Ignores validation errors
- Maximum effort to reach target state
- Used on explicit user request

### Disabled branches and YAML validation
- when a stack is configured as disabled, compose operations are skipped and file system changes are not required (except prior installation remains preserved);
- YAML validation is performed on compose files during integrity ensuring to fail fast in case of syntax errors.

## Error Handling Strategy

### Fail-Fast Principle
- Each phase can interrupt execution
- Partial state preserved (no rollback)
- Errors bubbled up with context

### Recovery Scenarios
- User can repeat operation with force=true
- Partial installation can be completed
- Remove/Purge operations for complete cleanup

## Operation Algorithms

### Update Operation
- checks `checker.*IsUpToDate` flags for compose stacks
- compose stacks: download then `docker compose up -d`, gather info
- worker: pull images only, gather info
- installer: stubbed (download/update/remove return not implemented), checksum and replace helpers exist
- refreshes updates info at the end of successful flows

### FactoryReset Operation
Correct cleanup sequence ensuring no dangling resources:
1. `purgeStack(all)` - removes all compose containers, networks, volumes
2. Remove worker containers/volumes (Docker API managed)
3. Remove residual networks (fallback cleanup)
4. Restore default .env from embedded
5. Restore all stack files with force=true
6. Refresh checker state

### Install Operation
- ensures docker networks exist before any stack operations
- checks `*Installed` flags to skip already installed components
- follows the same three-phase approach as ApplyChanges
- designed for fresh system setup

### Remove vs Purge
- **Remove**: Soft operation preserving user data (`docker compose down`)
- **Purge**: Complete cleanup including volumes (`docker compose down -v`)
- Files preserved in both cases for potential re-enablement

## Code Organization

### Clean Architecture
- Each operation delegates to specialized handlers (compose/docker/fs/update)
- No duplicate logic - single responsibility principle
- Force mode propagated through operationState
- Global mutex prevents concurrent modifications

## Integration Points

### State Management
- `state.State.IsDirty()` determines need for operations
- `state.State.Commit()` commits changes
- Environment variables control deployment modes

### Checker Integration
- `checker.CheckResult` contains current state
- Gather methods update state after operations
- Boolean flags optimize decision logic

### Files Integration
- `files.Files` provides embedded content access
- Fallback to filesystem when embedded missing
- Copy operations with rewrite flag for force mode

## Testing Strategy

### Unit Tests Focus
- State transition logic (current → target)
- Force mode behavior verification
- Error handling and partial state recovery
- Integration between components

### Mock Requirements
- files.Files for embedded content control
- checker.CheckResult with mockCheckHandler for state simulation
- baseMockFileSystemOperations, baseMockDockerOperations, baseMockComposeOperations with call tracking
- mockCheckHandler configured via mockCheckConfig for various scenarios

### Test Implementation
- Base mocks provide call logging and positive scenarios
- Error injection through setError() method on base mocks
- CheckResult states controlled via mockCheckHandler configuration
- Helper functions: testState(), testOperationState(), assertNoError(), assertError()

## Performance Characteristics

### Time Complexity
- O(1) for each stack (processed in parallel)
- File operations: O(n) where n = number of files in stack
- Container operations: depend on Docker/network latency

### Memory Usage
- Minimal: only state metadata
- Files read streaming without full memory loading
- CheckResult caches results until explicit refresh

### Network Impact
- Docker image pulls only when necessary
- Container updates use efficient rolling strategy
- External service connections minimal (checks only)
