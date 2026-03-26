# Checker Package Documentation

## Overview

The `checker` package is responsible for gathering system facts and verifying installation prerequisites for PentAGI. It performs comprehensive system analysis to determine the current state of the installation and what operations are available.

## Architecture

### Core Design Principles

1. **Delegation Pattern**: Uses a `CheckHandler` interface to delegate information gathering logic, allowing for flexible implementations and testing
2. **Parallel Information Gathering**: Collects information from multiple sources (Docker, filesystem, network) concurrently
3. **Fail-Safe Approach**: Returns sensible defaults when checks cannot be performed, avoiding false negatives
4. **Context-Aware**: All operations support context for cancellation and timeouts

### Key Components

#### CheckResult Structure
Central data structure that holds all system check results:
- Installation status for each component (PentAGI, Langfuse, Observability)
- System resource availability (CPU, memory, disk)
- Docker environment status
- Network connectivity status
- Update availability information
- Computed values for UI display:
  - CPU count
  - Required and available memory in GB
  - Required and available disk space in GB
  - Detailed network failure messages
  - Docker error type (not_installed, not_running, permission, api_error)
  - Write permissions for configuration directory

#### CheckHandler Interface
```go
type CheckHandler interface {
    GatherAllInfo(ctx context.Context, c *CheckResult) error
    GatherDockerInfo(ctx context.Context, c *CheckResult) error
    GatherWorkerInfo(ctx context.Context, c *CheckResult) error
    GatherPentagiInfo(ctx context.Context, c *CheckResult) error
    GatherLangfuseInfo(ctx context.Context, c *CheckResult) error
    GatherObservabilityInfo(ctx context.Context, c *CheckResult) error
    GatherSystemInfo(ctx context.Context, c *CheckResult) error
    GatherUpdatesInfo(ctx context.Context, c *CheckResult) error
}
```

## Check Categories

### 1. Docker Environment Checks
- **Docker API Accessibility**: Verifies connection to Docker daemon
- **Docker Error Detection**: Identifies specific Docker issues (not installed, not running, permission denied)
- **Docker Version**: Ensures Docker version >= 20.0.0
- **Docker Compose Version**: Ensures Docker Compose version >= 1.25.0
- **Worker Environment**: Checks separate Docker environment for pentesting tools (supports remote Docker hosts)

### 2. Component Installation Checks
- **File Existence**: Verifies presence of docker-compose files
- **Container Status**: Checks if containers exist and their running state
- **Script Installation**: Verifies PentAGI CLI script in /usr/local/bin

### 3. System Resource Checks
- **Write Permissions**: Verifies write access to configuration directory
- **CPU**: Minimum 2 CPU cores required
- **Memory**: Dynamic calculation based on components to be installed
  - Base: 0.5GB free
  - PentAGI: +0.5GB
  - Langfuse: +1.5GB
  - Observability: +1.5GB
- **Disk Space**: Context-aware requirements
  - Worker images not present: 25GB (for large pentesting images)
  - Components to install: 10GB + 2GB per component
  - Already installed: 5GB minimum

### 4. Network Connectivity Checks
Three-tier verification process:
1. **DNS Resolution**: Tests ability to resolve docker.io
2. **HTTP Connectivity**: Verifies HTTPS access (proxy-aware)
3. **Docker Pull Test**: Attempts to pull a small test image

### 5. Update Availability Checks
- Communicates with update server to check latest versions
- Sends current component versions and configuration
- Supports proxy configuration
- Checks updates for: Installer, PentAGI, Langfuse, Observability, Worker images

## Public API

### Main Entry Points
```go
// Gather performs all system checks using provided application state
func Gather(ctx context.Context, appState state.State) (CheckResult, error)

// GatherWithHandler allows custom CheckHandler implementation
func GatherWithHandler(ctx context.Context, handler CheckHandler) (CheckResult, error)
```

### Availability Helper Methods
The CheckResult provides helper methods to determine available operations:
```go
func (c *CheckResult) IsReadyToContinue() bool      // Pre-installation checks passed
func (c *CheckResult) CanInstallAll() bool          // Can perform installation
func (c *CheckResult) CanStartAll() bool            // Can start services
func (c *CheckResult) CanStopAll() bool             // Can stop services
func (c *CheckResult) CanUpdateAll() bool           // Updates available
func (c *CheckResult) CanFactoryReset() bool        // Can reset installation
```

## Implementation Details

### OS-Specific Implementations
- **Memory Checks**:
  - Linux: Reads /proc/meminfo for MemAvailable
  - macOS: Parses vm_stat output for free + inactive + purgeable pages
- **Disk Space Checks**:
  - Uses `df` command with appropriate flags per OS

### Docker Integration
- Supports both local and remote Docker environments
- Handles TLS configuration for secure remote connections
- Compatible with Docker contexts and environment variables

### Error Handling Philosophy
- Network failures are treated as "assume OK" to avoid blocking on transient issues
- Missing system information defaults to "sufficient resources"
- Only critical failures (missing env file, Docker API inaccessible) prevent continuation

### Version Parsing
- Flexible regex-based extraction from various version output formats
- Semantic version comparison for compatibility checks
- Handles both docker-compose and docker compose command variants

### Image Information Extraction
- Parses complex Docker image references (registry/namespace/name:tag@hash)
- Handles various edge cases in image naming conventions
- Extracts version information for update comparison

### Helper Functions for Code Reusability
To avoid code duplication, the package provides several shared helper functions:

- **calculateRequiredMemoryGB**: Calculates total memory requirements based on components that need to be started
- **calculateRequiredDiskGB**: Computes disk space requirements considering worker images and local components
- **countLocalComponentsToInstall**: Counts how many components need local installation
- **determineComponentNeeds**: Determines which components need to be started based on their current state
- **getAvailableMemoryGB**: Platform-specific memory availability detection
- **getAvailableDiskGB**: Platform-specific disk space availability detection
- **getNetworkFailures**: Collects detailed network connectivity failure messages
- **getProxyURL**: Centralized proxy URL retrieval from application state
- **getDockerErrorType**: Identifies specific Docker error types (not installed, not running, permission issues)
- **checkDirIsWritable**: Tests write permissions by creating a temporary file

These functions ensure consistent calculations across different parts of the codebase and make maintenance easier.

## Constants and Thresholds

Key configuration values are defined as constants for easy adjustment:
- Container names for each service
- Minimum resource requirements
- Default endpoints for services
- Update server configuration
- Version compatibility thresholds

## Thread Safety

The default implementation uses mutex protection for Docker client management, ensuring safe concurrent access during information gathering operations.
