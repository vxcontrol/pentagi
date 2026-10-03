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
- Installation status for each component (PentAGI, Graphiti, Langfuse, Observability)
- System resource availability (CPU, memory, disk)
- Docker environment status
- Network connectivity status
- The update check: the answer per stack (`StackUpdates`), why there is no answer (`UpdateFailure`), and what the check observed on this machine (`InstalledComponents`), plus the per-stack `*IsUpToDate` booleans the menus read
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
    GatherGraphitiInfo(ctx context.Context, c *CheckResult) error
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
  - Graphiti: +2GB
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
3. **Docker Pull Test**: Attempts to pull `debian:latest` — a small probe image, the default of `DOCKER_DEFAULT_IMAGE`, not the pentest worker image — when both Docker clients are available

#### Restricted Network Troubleshooting

The current checker validates Docker Hub reachability by resolving `docker.io`, making an HTTPS connectivity check, and — when both Docker clients are available — attempting a pull of `debian:latest` (the default of `DOCKER_DEFAULT_IMAGE`, used here only as a small reachability probe — the pentest worker image `vxcontrol/kali-linux` is not pulled by this check). This means the installer can fail network validation even when the host has general internet access but Docker itself is not configured for the target network.

Recommended remediation order:

1. Confirm general internet access and DNS resolution for `docker.io`
2. If your environment requires an outbound proxy for installer or PentAGI HTTP traffic, set the `PROXY_URL` environment variable. To route Docker image pulls through a proxy, configure the Docker daemon or Docker Desktop proxy separately — Docker does not use `PROXY_URL` for registry access.
3. If Docker Hub is blocked or rate-limited, configure an organization-approved Docker registry mirror or registry proxy at the Docker daemon / Docker Desktop level
4. Restart Docker and rerun the installer checks

PentAGI variables such as `PENTAGI_IMAGE`, `DOCKER_DEFAULT_IMAGE`, and `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` do not replace Docker daemon registry configuration. They only influence the PentAGI application image or worker image selection after Docker is already able to pull the required images. Note that the main Compose stack already includes a service from `quay.io` (`postgres-exporter`), the optional observability stack includes an image from `gcr.io`, and the optional Langfuse stack includes one from `cgr.dev` (`minio`). A Docker Hub mirror alone is therefore not sufficient for a full deployment — those registries also need to be reachable or individually mirrored.

See Docker's official documentation for [registry mirrors](https://docs.docker.com/docker-hub/image-library/mirror/) and [daemon proxy configuration](https://docs.docker.com/engine/daemon/proxy/).

### 5. Update Availability Checks

`GatherUpdatesInfo` (`checker/updates.go`) asks the PentAGI Cloud API what is available for this installation. It never fails the surrounding gather: a check that cannot be made is a fact *about* the installation, recorded in the result, not a reason to throw away everything else the checker learned.

The call goes to `https://update.pentagi.com` (`UPDATE_SERVER_HOST` overrides the host) through the installer's own `PROXY_URL`, and carries `LICENSE_KEY` when it is present and holds up. A key that does not hold up is deliberately **not** fatal — the check runs without it and the reason is written to the log, because otherwise the only symptom is a tier the user did not expect. `UPDATE_STRATEGY` selects the release channel (`preview` by default, also `stable` and `nightly`); an unrecognised value falls back to `preview` with a warning rather than failing the check, since a typo should not leave an installation unable to learn about security updates. The default is `preview` because `stable` measures the answer against curated releases, and a stack with no published release is then offered nothing at all — which reads as "no updates exist" when it means "no releases are published for this yet".

#### Facts are gathered per component, not per stack

`checker/components.go` turns the running installation into the component list the request carries. Two lists are sent, and they are separate all the way through: an **image** is identified by a registry reference and a digest, a **file** by a version and a hash, and no artefact is ever both.

The list has to be complete. The answer only covers what was asked about, so a component left out comes back as "nothing to report" — which reads exactly like "up to date", and that is the more dangerous of the two mistakes.

For every container image the request carries:

- **Component** — the artefact's own name, taken from an explicit component → container table. Nothing here is derived: for four of the sixteen images the container name is neither the compose service name nor the component name — Langfuse's `postgres`, `clickhouse`, `redis` and `minio` services run in containers called `langfuse-postgres`, `langfuse-clickhouse`, `langfuse-redis` and `langfuse-minio`, while the component sent stays the bare `postgres`, `clickhouse`, `redis`, `minio` — and a wrong container name is indistinguishable from an absent one, so a guess would silently drop the component instead of failing. The container name is what the table holds and what `ContainerInspect` is given; the compose `hostname:` field (`otelcol` for the `otel` service) is never read.
- **Status** — `running` when the container is up, `installed` when it exists but is stopped.
- **OS and Arch** — the **image's**, read from `ImageInspect` (`.Os`, `.Architecture`), never the host's. A container image is `linux` even on macOS and Windows, where Docker runs in a VM, and on Apple silicon an `amd64` image may well be running under emulation. Substituting `runtime.GOOS`/`runtime.GOARCH` is not caught by validation — `darwin` is a valid value, it just matches no published artefact — so the installation would silently be told "no updates" forever. An image on a platform the contract has no name for (say `arm/v7`) is dropped with a log line rather than rounded to the nearest value: rounding asks about one image and gets an answer about another.
- **Repository and Tag** — parsed from the reference the container was configured with, with a Docker Hub host and the implicit `library/` removed and the host of any other registry kept, and the tag `docker pull` would apply filled in when the reference omits one. Both halves are required: an update is the newest image *under a reference*, so a component that cannot name its reference cannot be resolved at all.
- **Image hash** — the image's **config digest**, i.e. what `docker inspect` returns as `Id`.

Note the order: `ContainerInspect` first, then `ImageInspect` **by the config digest that inspection returned**, not by the tag. A tag can have been moved onto a different image since the container started, and inspecting by tag would describe something other than what is actually running.

#### The three digests, and which one is sent

Every image has three sha256 identities: the **index digest** (the whole multi-platform publication, and what a pull by tag records in `RepoDigests`), the per-platform **manifest digest**, and the **config digest** (`docker inspect` → `Id`).

The installer sends the config digest, for three reasons: it is always present, it is per-platform — so it confirms the architecture as well as the build — and it needs no round-trip to the registry. The answer may carry all three (`image_hash`, `config_hash`, `index_hash`), and a match against **any one of them** means the installed artefact is what is being offered.

Digests are sent as **bare lowercase hex, 64 characters, without the `sha256:` prefix** (`normalizeDigest`). The daemon reports the prefixed form, so the conversion is mandatory, and it is not cosmetic: the contract validates against `^[0-9a-f]{64}$`, so the prefixed form and the same 64 characters in uppercase are both rejected — and a rejected request costs the answer for every component in it, not just the offending one. A value that does not normalise is **omitted** rather than sent: a missing digest means "nothing to compare", a malformed one takes the whole request down. Never truncate a digest to the 12 characters the CLI displays.

#### The three components that are not compose containers

- **Worker image.** It has no container and no compose service — it is pulled into the worker Docker environment on demand — so it is reported only when that environment is reachable and the image is actually present. What is reported is what is there, taken from `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` (default `vxcontrol/kali-linux`); substituting the built-in default would claim an image the machine may not have. Status is always `installed`, because the image being present is the whole story. The image is authoritative about its own platform; the host architecture is used only as a fallback for a daemon that will not describe it, and if the result is still not a contract value the component is dropped.
- **Installer binary.** The one component that is a file on *this* host, and therefore the only one whose `os`/`arch` are legitimately the host's own. It is reported with its normalised version and, when the executable path can be read, the sha256 of the running binary. The hash is a nice-to-have — installer builds are identified by version — so failing to read our own path does not drop the component. The request's top-level `installer_version` / `installer_os` / `installer_arch` describe the host for the same reason.
- **Jaeger storage plugin** (`jaeger-clickhouse`). Reported as a **file** component, **once per architecture**, from `observability/jaeger/bin/jaeger-clickhouse-linux-amd64` and `…-arm64` relative to the directory the environment file lives in. Both are reported whatever the host is, because which one runs is decided *inside* the Jaeger container by the architecture of the Docker VM. Each is identified by the **sha256 of the file** and nothing else — the plugin ships without a version — with `os: linux`, because the binary runs in the container. A file that cannot be read is skipped; the two are emitted in a fixed order so two runs on the same machine produce the same request. The directory and file names are exported (`JaegerPluginDir`, `JaegerPluginBinaries`) so that the path the processor downloads into and the path reported here stay one table: if they drift, an update writes a file nothing reads.

#### What is deliberately left out

- A **container that does not exist** — the component is not deployed here, and is not worth a slot in the request.
- Every image of a stack pointing at an **external address**. Such a stack runs somewhere that cannot be inspected, so there is no reference to resolve, and reporting it would claim somebody else's deployment as ours. The Jaeger plugin is skipped for an external observability stack for the same reason.
- Every component of a stack that is **not configured at all**.
- Any component whose platform or image reference cannot be described completely — see above. One dropped component costs one answer; one invalid field costs all of them.

The request may report at most **30 artefacts, counting images and files together**. Today's inventory is 16 stack images + worker + installer + two plugin files = 20, but the list grows with every stack, so the ceiling is enforced during gathering. Components are ordered so the stacks whose updates matter most come first, files are trimmed before images, and whatever is dropped is **named in the log** — a silently truncated list would report the dropped components as complete.

#### Stack statuses

Alongside the components the request says how each product stack is deployed: `installed`, `connected`, `external` or `unused`. Nothing in the answer depends on this — it is reported for its own sake, because a stack hosted elsewhere contributes no components at all and is otherwise indistinguishable from a stack nobody uses. Every stack is reported, including the unused ones: silence cannot be told from a client too old to speak about stacks. `external` takes precedence over `installed`, since it is the more specific statement — it says who *operates* the stack — and an installation pointing at somebody else's Langfuse may well also have the compose file on disk.

#### What the answer contains

`CheckResult.StackUpdates` holds the answer in full; the per-stack `*IsUpToDate` booleans the menus read are the same verdict reduced to one bit. A stack **missing from the answer stays up to date**: the answer covers what was asked about, and silence is not evidence — offering an update nobody has grounds for is the worse failure.

`StackUpdate`:

| Field | Meaning |
|---|---|
| `HasUpdate` | The stack-level verdict, and the only thing the menu needs. |
| `CurrentVersion` / `LatestVersion` | What this installation is attributed to, and what is offered. |
| `CurrentVersionMixed` | The stack's components were attributed to **different** releases, so `CurrentVersion` names the oldest of them rather than a version this installation as a whole ever was. Legitimate — a component nobody rebuilt stays on its old release — but "you are on 2.1.0" and "the oldest thing you have is from 2.1.0" are different sentences, and a screen must not print the first when it means the second. |
| `Releases` | Every curated release this update crosses, **oldest first, the last entry being the target**. Empty for two different reasons: there is genuinely nothing to cross, or the answer predates the field — so an empty list is not by itself evidence of a single-step update. |
| `ReleasesTruncated` | The list was cut, and it is the **oldest** entries that went. |
| `Changelog` / `ReleaseNotes` | The **target** release's text only. When `Releases` is present its last entry carries the same text. |
| `Resolution` | How the answer was arrived at: `release`, `channel`, `ahead_of_release`, `no_artifact_for_tag`, `not_tracked`. **Diagnostic only** — it is logged and nothing more. It exists so that `has_update: false` can be *read* instead of guessed at, most of all in the case that used to be silent: an installation running something newer than anything curated. Branching on it would couple the installer to wording that is free to grow. |
| `Components` | The per-artefact plan, below. |

`ComponentUpdate`. Being listed here does **not** mean the artefact is outdated: under `stable` the answer names every artefact of the release that matches something reported, whether or not it differs, so the installation can be attributed to a version with release notes.

| Field | Meaning |
|---|---|
| `Action` | What to do about this artefact: `current`, `install`, `upgrade`, `downgrade`, `unknown`. Exactly one comes back per reported component, including the ones nothing could be resolved for. |
| `Reason` | Set only for `unknown`: `tag_not_published` (the component is published, but nothing under the reference this installation follows), `repository_not_tracked` (an image nobody publishes — usually a private rebuild), `component_not_tracked` (nothing published for this component on this platform at all), `no_release_artifact_for_tag` (the component exists, but the release the stack is measured against links no artefact of it). This is the difference between "nothing to do" and "we publish nothing for what you are running" — opposite situations that used to arrive identically. |
| `PullReference` | What to write into the compose variable before pulling, ready to use. It is **not** always `repository:tag`: under `stable` it pins the most specific immutable tag so the pull is reproducible, under `preview` and `nightly` it names the moving tag. Write it as given — composing a reference locally undoes that. |
| `Outdated` | Whether this artefact would actually change. |
| `Verifiable` | Whether **this machine** had something to compare — a digest for an image, a version or hash for a file. It describes the installed side only; it says nothing about what the answer carried. |
| `CurrentVersion` | What is installed, in whatever identity the comparison used: the config digest for an image, the version — or, for the plugin, the file hash — for a file. |
| `TargetVersion` | For an image, the offered **config digest**: the same identity the installed side is measured in, and therefore the value post-update verification compares a pulled image against. For a file, the offered version. Empty when it is not known, and verification then degrades honestly to "cannot verify". |

#### How `Outdated` and `Verifiable` are decided

Images are compared **by digest, not by tag** — a moving tag such as `latest` points at different bytes over time, so the tag alone says nothing — and a match against any of the three offered identities counts. Files are compared by version, and by hash when no version was reported, which is the Jaeger plugin's case.

Where the comparison has nothing to work with, **the server's `Action` is the authority**:

- `install`, `upgrade`, `downgrade` set `Outdated` unconditionally. `install` is the case that matters: nothing has been pulled, so there is no digest to compare and the comparison alone leaves `Outdated` false — which reads as "nothing to do" for exactly the component that most needs doing.
- `unknown` clears both `Outdated` and `Verifiable`: there is no artefact at all, so whatever a local comparison produced means nothing.
- An artefact with no installed counterpart is reported **unverifiable, not mismatched** — "this will change" and "we cannot tell" are different statements, and conflating them offers an update on no evidence. The guard is on the **installed** side only, though: `Verifiable` is set whenever this machine had a digest to compare, whatever the answer carried. An offer that names the image only by `image_hash` — the per-platform manifest digest, which no Docker API reports and which therefore can never equal the config digest the installer sends — comes back verifiable *and* outdated, not "cannot tell". Where that case does degrade honestly is one layer later: `TargetVersion` stays empty, so the post-update verification reports "cannot verify" instead of a mismatch.

`CheckResult.InstalledComponents` keeps what this check observed on the machine, recorded **before** the call is made, since it is a local fact whether or not the server answers. Post-update verification reads it instead of interrogating Docker for a second time.

#### When there is no answer

A failed check sets `UpdateServerAccessible = false`, clears `StackUpdates`, and resets **all six** per-stack flags. Resetting all of them matters: a flag left behind keeps whatever verdict an earlier successful check put there, and the interface then presents a stale answer as a current one — the worker flag used to be the one left behind.

`UpdateFailure` says why, in terms the interface can act on, because a spent quota and a broken proxy call for opposite reactions. `Retryable` separates "come back later" from "this needs a different license", and `RetryAfter` carries the advertised cooldown when there is one.

| Reason | What it means | Retryable |
|---|---|---|
| `unreachable` | Nothing reached a verdict: no network, no DNS, a proxy that refused, a connection that died mid-flight. | yes |
| `timeout` | A deadline — the caller's, or the proof-of-work challenge exceeding its budget on slow hardware. | yes |
| `rate_limited` | Too many requests in a rolling window; `RetryAfter` says when it reopens. | yes |
| `quota_exceeded` | The allowance for the period is spent; `RetryAfter` says when it resets. | yes |
| `server_error` | The server failed on its side. | yes |
| `forbidden` | This client is not allowed to make the call at all — a license tier without it, or a blocked client. Waiting does not help; a different license does. | no |
| `not_found` | No such package or version. That is an answer, not an outage. | no |
| `rejected` | The request was refused as malformed, including replay and signature refusals. That is our bug, not the user's, and resending an identical request cannot succeed. | no |

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
