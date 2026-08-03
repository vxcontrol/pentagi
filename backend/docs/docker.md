# Docker Client Package Documentation

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
  - [Worker Docker Access](#worker-docker-access)
- [Core Interfaces](#core-interfaces)
- [Container Lifecycle Management](#container-lifecycle-management)
- [Security and Isolation](#security-and-isolation)
- [Integration with PentAGI](#integration-with-pentagi)
- [Usage Examples](#usage-examples)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

The Docker client package (`backend/pkg/docker`) provides a secure and isolated containerized environment for PentAGI's AI agents to execute penetration testing operations. This package serves as a wrapper around the official Docker SDK, offering specialized functionality for managing containers that AI agents use to perform security testing tasks.

### Key Features

- **Secure Isolation**: All operations are performed in sandboxed Docker containers with complete isolation
- **AI Agent Integration**: Specifically designed to support AI agent workflows and terminal operations
- **Container Lifecycle Management**: Comprehensive container creation, execution, and cleanup
- **Port Management**: Automatic port allocation for flow-specific containers
- **File Operations**: Safe file transfer, path metadata lookup, and non-recursive directory listing between host and containers
- **Network Isolation**: Configurable network policies for security
- **Resource Management**: Memory and CPU limits for controlled execution
- **Volume Management**: Persistent and temporary storage solutions

### Role in PentAGI Ecosystem

The Docker client is a critical component that enables PentAGI's core promise of secure, isolated penetration testing. It provides the foundation for:

- **Terminal Access**: AI agents execute commands in isolated environments
- **Tool Execution**: Professional pentesting tools run in dedicated containers
- **File Management**: Secure file operations and artifact storage
- **Environment Preparation**: Dynamic container setup based on task requirements
- **Resource Cleanup**: Automatic cleanup of completed or failed operations

## Architecture

### Core Components

The Docker client package consists of several key components:

```
backend/pkg/docker/
├── client.go          # Main Docker client implementation
└── (future files)     # Additional Docker utilities
```

### Key Constants and Configuration

```go
const WorkFolderPathInContainer = "/work"   // Standard working directory in containers
const WorkerVolumeNameSuffix   = "-data"    // Suffix of the per-flow named volume
const BaseContainerPortsNumber = 28000      // Default base for dynamic port allocation

const defaultImage              = "debian:latest"          // Fallback image if custom image fails
const defaultDockerSocketPath   = "/var/run/docker.sock"   // Mount point of a bound socket
const containerPortsNumber      = 2                        // Number of ports allocated per container
const limitContainerPortsNumber = 2000                     // Port window size per instance
const containerListWorkers      = 20                       // Parallel stat workers for directory listing
```

### Port Allocation Strategy

PentAGI uses a deterministic port allocation algorithm to ensure each flow gets unique, predictable ports:

```go
func GetPrimaryContainerPorts(portsBase int, flowID int64) []int {
    if portsBase <= 0 || portsBase > (65535-limitContainerPortsNumber) {
        portsBase = BaseContainerPortsNumber
    }
    ports := make([]int, containerPortsNumber)
    for i := range containerPortsNumber {
        delta := (int(flowID)*containerPortsNumber + i) % limitContainerPortsNumber
        ports[i] = portsBase + delta
    }
    return ports
}
```

This ensures that:
- Each flow gets consistent port numbers across restarts
- Port conflicts are avoided between different flows
- Ports stay inside a 2000-wide window starting at `DOCKER_PORTS_BASE` (default `28000`, so `28000-30000`)

The base is configurable because flow ids restart at `1` in every PentAGI instance: two instances sharing one worker node would otherwise request identical host ports for their respective flow `1`. Give each instance a disjoint window (`28000`, `30000`, …). An out-of-range base silently falls back to `28000` rather than producing unbindable ports.

## Configuration

### Environment Variables

The Docker client is configured through several environment variables defined in the main configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon connection |
| `DOCKER_INSIDE` | `false` | Whether PentAGI communicates with host Docker daemon from containers |
| `DOCKER_NET_ADMIN` | `false` | Whether PentAGI grants the primary container NET_ADMIN capability for advanced networking. |
| `DOCKER_SOCKET` | | Explicit socket path on the worker node to bind into worker containers. Empty enables autodetection — see [Worker Docker Access](#worker-docker-access) |
| `DOCKER_INSIDE_HOST` | | Docker daemon endpoint given **to worker containers**; also disables socket autodetection |
| `DOCKER_INSIDE_TLS_VERIFY` | | TLS verification for the worker container's Docker connection |
| `DOCKER_INSIDE_CERT_PATH` | | TLS certificate directory **on the worker node**, mounted read-only into worker containers |
| `DOCKER_PORTS_BASE` | `28000` | First host port of this instance's per-flow allocation window |
| `DOCKER_NETWORK` | | Docker network for container communication (bridge mode) or `host` for host network mode |
| `DOCKER_PUBLIC_IP` | `0.0.0.0` | Public IP for port binding (bridge mode only) |
| `DOCKER_WORK_DIR` | | Custom work directory path on host |
| `DOCKER_DEFAULT_IMAGE` | `debian:latest` | Fallback image if AI-selected image fails |
| `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` | `vxcontrol/kali-linux` | Default Docker image for penetration testing tasks |
| `DATA_DIR` | `./data` | Local data directory for file operations |

### Configuration Structure

```go
type Config struct {
    // Docker (terminal) settings
    DockerInside   bool   `env:"DOCKER_INSIDE" envDefault:"false"`
    DockerNetAdmin bool   `env:"DOCKER_NET_ADMIN" envDefault:"false"`
    DockerSocket   string `env:"DOCKER_SOCKET"`

    // How a worker container reaches a Docker daemon (see Worker Docker Access)
    DockerInsideHost      string `env:"DOCKER_INSIDE_HOST"`
    DockerInsideTLSVerify string `env:"DOCKER_INSIDE_TLS_VERIFY"`
    DockerInsideCertPath  string `env:"DOCKER_INSIDE_CERT_PATH"`

    DockerNetwork                string `env:"DOCKER_NETWORK"`
    DockerPublicIP               string `env:"DOCKER_PUBLIC_IP" envDefault:"0.0.0.0"`
    DockerWorkDir                string `env:"DOCKER_WORK_DIR"`
    DockerPortsBase              int    `env:"DOCKER_PORTS_BASE" envDefault:"28000"`
    DockerDefaultImage           string `env:"DOCKER_DEFAULT_IMAGE" envDefault:"debian:latest"`
    DockerDefaultImageForPentest string `env:"DOCKER_DEFAULT_IMAGE_FOR_PENTEST" envDefault:"vxcontrol/kali-linux"`
    DataDir                      string `env:"DATA_DIR" envDefault:"./data"`
}
```

The three `DOCKER_INSIDE_*` values are consumed through helpers on `*config.Config` (`WorkerDockerSocket`, `WorkerDockerEnv`, `WorkerDockerCertPath`) so the decision logic lives in one place and the Docker client stays a thin consumer.

### NET_ADMIN Capability Configuration

The `DOCKER_NET_ADMIN` option controls whether PentAGI containers are granted the `NET_ADMIN` Linux capability, which provides advanced networking permissions essential for many penetration testing operations.

#### Network Administration Capabilities

When `DOCKER_NET_ADMIN=true`, containers receive the following networking capabilities:

- **Network Interface Management**: Create, modify, and delete network interfaces
- **Routing Control**: Manipulate routing tables and network routes
- **Firewall Rules**: Configure iptables, netfilter, and other firewall systems
- **Traffic Shaping**: Implement QoS (Quality of Service) and bandwidth controls
- **Bridge Operations**: Create and manage network bridges
- **VLAN Configuration**: Set up and modify VLAN configurations
- **Packet Capture**: Enhanced access to raw sockets and packet capture mechanisms

#### Security Implications

**Enabling NET_ADMIN (`DOCKER_NET_ADMIN=true`)**:
- **Benefits**: Enables full-featured network penetration testing tools
- **Risks**: Containers can potentially modify host network configuration
- **Use Cases**: Network scanning, traffic interception, custom routing setups
- **Tools Enabled**: Advanced nmap features, tcpdump, wireshark, custom networking tools

**Disabling NET_ADMIN (`DOCKER_NET_ADMIN=false`)**:
- **Benefits**: Enhanced security isolation from host networking
- **Limitations**: Some advanced networking tools may not function fully (nmap)
- **Use Cases**: Application-level testing, web security assessment
- **Recommended**: For environments where network-level testing is not required

#### Container Capability Assignment

The primary container does not rely on Docker's implicit default capability set. `tools.go` (`flowToolsExecutor.Prepare`) sets `CapDrop: ["ALL"]` and then adds back an explicit allow-list — Docker's own default 14-capability set minus `MKNOD`, plus `NET_ADMIN` when `DOCKER_NET_ADMIN=true`:

```go
// Primary containers (when DOCKER_NET_ADMIN=false)
hostConfig := &container.HostConfig{
    CapDrop: []string{"ALL"},
    CapAdd: []string{
        "CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER",
        "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
        "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
    },
}

// Primary containers (when DOCKER_NET_ADMIN=true) — same list, plus NET_ADMIN
hostConfig := &container.HostConfig{
    CapDrop: []string{"ALL"},
    CapAdd: []string{
        "CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER",
        "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
        "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
        "NET_ADMIN",
    },
}
```

Why the full default set (minus one) instead of just `NET_RAW`/`NET_ADMIN`: pentest workflows routinely install new tools at runtime via `apt`/`dpkg` (the Installer Agent's core job), and several common network tools' `postinst` maintainer scripts call `setcap` on their binaries instead of relying on setuid (`ping`, `traceroute`, `nmap`, `dumpcap`, `hping3`, …). That needs `SETFCAP`; `SETPCAP`/`FSETID`/`AUDIT_WRITE` round out the rest of Docker's default set that ordinary package management and privilege-dropping daemons expect. See [Capability Management](#capability-management) below for the full rationale, including the one deliberate omission (`MKNOD`).

### Worker Docker Access

Two independent questions are often confused, and PentAGI answers them with two separate sets of variables:

| Question | Variables | Consumer |
|---|---|---|
| Which daemon does **PentAGI** create worker containers on? | `DOCKER_HOST`, `DOCKER_TLS_VERIFY`, `DOCKER_CERT_PATH` | Docker SDK inside the PentAGI process |
| Which daemon may a **worker container** talk to? | `DOCKER_INSIDE`, `DOCKER_SOCKET`, `DOCKER_INSIDE_HOST`, `DOCKER_INSIDE_TLS_VERIFY`, `DOCKER_INSIDE_CERT_PATH` | The agent's sandbox |

`DOCKER_INSIDE` is the master switch for the second row. With it disabled a sandbox gets no Docker access and no Docker configuration at all.

#### Socket selection algorithm

With `DOCKER_INSIDE=true`, `NewDockerClient` resolves the socket once at startup:

```go
socket, autodetectSocket := cfg.WorkerDockerSocket()
if autodetectSocket {
    socket = getHostDockerSocket(ctx, cli)
}
```

| `DOCKER_SOCKET` | `DOCKER_INSIDE_HOST` | Result |
|---|---|---|
| set | any | That socket is bind-mounted at `/var/run/docker.sock` inside every worker |
| empty | set | **Nothing is mounted**; autodetection is skipped entirely |
| empty | empty | Host socket is autodetected and mounted (historical behaviour) |

#### Environment injection

Every non-empty `DOCKER_INSIDE_*` value is injected into the worker container with the `_INSIDE_` segment removed, so the Docker CLI inside picks it up unmodified:

| Configured on PentAGI | Seen inside the worker container |
|---|---|
| `DOCKER_INSIDE_HOST=tcp://10.0.0.5:3376` | `DOCKER_HOST=tcp://10.0.0.5:3376` |
| `DOCKER_INSIDE_TLS_VERIFY=1` | `DOCKER_TLS_VERIFY=1` |
| `DOCKER_INSIDE_CERT_PATH=/etc/docker/dind-client` | `DOCKER_CERT_PATH=/etc/docker/dind-client` |

Empty values are omitted rather than injected blank. When `DOCKER_INSIDE_CERT_PATH` is set, that directory is additionally bind-mounted **read-only at the identical path**, so the injected `DOCKER_CERT_PATH` resolves unchanged inside the container. The path is resolved on the **worker node** — the machine whose daemon creates sandboxes — which is frequently not the machine running PentAGI.

```go
if dc.inside {
    if dc.socket != "" {
        hostConfig.Binds = append(hostConfig.Binds, dc.socket+":"+defaultDockerSocketPath)
    }
    config.Env = append(config.Env, dc.insideEnv...)
    if dc.insideCertPath != "" {
        hostConfig.Binds = append(hostConfig.Binds, dc.insideCertPath+":"+dc.insideCertPath+":ro")
    }
}
```

#### Why the TCP endpoint is preferred over a bound socket

Bind-mounting a socket into worker containers has two distinct problems.

**Ordering fragility.** A bind-mount source that does not exist yet is created by Docker as a **directory**. If the worker node reboots and a worker container with `restart: on-failure` starts before the dind daemon has recreated its socket, Docker materialises a directory at `/var/run/docker-dind/docker.sock` — and dind then cannot bind its own socket at that path. The sandbox gets a useless mount and dind fails to start until the directory is removed by hand.

**Blast radius.** A socket bind-mount only avoids that race reliably when it is the **host** daemon's socket, since that one exists before anything else starts. But handing an autonomous agent the host daemon means handing it the host: it can start a privileged container, mount `/`, and take over the node — including PentAGI itself and every other flow's containers.

Pointing sandboxes at a hardened dind daemon over TLS avoids both. There is no mount to race on, and the authorization policy on that daemon constrains what the agent may create. See [Worker Node Setup](../../examples/guides/worker_node.md) for a complete configuration.

### Network Configuration

PentAGI supports two network modes for container isolation:

#### Bridge Network Mode (Default)

When `DOCKER_NETWORK` is set to a custom network name (e.g., `pentagi-network`), containers are connected to an isolated bridge network:
- **Isolated Communication**: Containers communicate only within the defined network
- **Port Mapping**: Container ports are mapped to host ports for external access
- **Service Discovery**: Enables internal DNS-based service discovery
- **Enhanced Security**: Network-level isolation from other containers

#### Host Network Mode

When `DOCKER_NETWORK` is set to the special value `host`, containers use the host's network stack directly:
- **Direct Network Access**: Container shares the host's network interfaces
- **No Port Mapping**: Ports are directly accessible on host interfaces (no NAT)
- **Performance**: Eliminates network virtualization overhead
- **Use Cases**: Advanced network testing, raw packet manipulation, network monitoring

**Security Consideration**: Host network mode reduces isolation. Use only when necessary for penetration testing tasks requiring direct host network access.

## Core Interfaces

### DockerClient Interface

The main interface defines all Docker operations available to PentAGI components:

```go
type DockerClient interface {
    // Container lifecycle management
    RunContainer(ctx context.Context, containerName string, containerType database.ContainerType,
        flowID int64, config *container.Config, hostConfig *container.HostConfig) (database.Container, error)
    StopContainer(ctx context.Context, containerID string, dbID int64) error
    RemoveContainer(ctx context.Context, containerID string, dbID int64) error
    IsContainerRunning(ctx context.Context, containerID string) (bool, error)

    // Command execution
    ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (container.ExecCreateResponse, error)
    ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error)
    ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)

    // File operations
    ContainerStatPath(ctx context.Context, containerID string, path string) (container.PathStat, error)
    ListContainerDir(ctx context.Context, containerID string, dirPath string) (ContainerDirListing, error)
    CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error
    CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error)

    // Utility methods
    Cleanup(ctx context.Context) error
    GetDefaultImage() string
}
```

### Implementation Structure

```go
type dockerClient struct {
    db        database.Querier    // Database for container state management
    logger    *logrus.Logger      // Structured logging
    dataDir   string              // Local data directory
    hostDir   string              // Host-mapped data directory
    client    *client.Client      // Docker SDK client
    inside    bool                // Worker containers may reach a Docker daemon
    defImage  string              // Default fallback image
    socket    string              // Socket to bind into workers ("" = none)
    network   string              // Docker network name
    publicIP  string              // Public IP for port binding
    portsBase int                 // First host port of this instance's window
    labels    map[string]string   // Tenant ownership labels (nil without a tenant)

    insideEnv      []string       // DOCKER_* injected into worker containers
    insideCertPath string         // TLS dir mounted read-only into workers
}
```

## Container Lifecycle Management

### Container Creation Process

The `RunContainer` method handles the complete container creation workflow:

1. **Preparation**:
   - Creates flow-specific work directory
   - Generates unique container name
   - Records container in database with "starting" status

2. **Image Management**:
   - Attempts to pull requested image
   - Falls back to default image if pull fails
   - Updates database with actual image used

3. **Container Configuration**:
   - Sets hostname based on container name hash
   - Configures working directory to `/work`
   - Sets up restart policy (`on-failure`, maximum 5 retries)
   - Configures logging (JSON driver with rotation)

4. **Storage Setup**:
   - Creates dedicated volume or bind mount
   - Mounts work directory to `/work` in container
   - When `DOCKER_INSIDE=true`: optionally mounts a Docker socket, injects the `DOCKER_*` environment derived from `DOCKER_INSIDE_*`, and mounts the TLS certificate directory read-only — see [Worker Docker Access](#worker-docker-access)

5. **Network and Ports**:
   - **Bridge Mode**: Assigns flow-specific ports using deterministic algorithm, binds to public IP
   - **Host Mode** (`DOCKER_NETWORK=host`): Uses host network stack, skips port bindings
   - Connects to specified Docker network (unless host mode)

6. **Container Startup**:
   - Creates container with all configurations
   - Starts container
   - Updates database status to "running"

### Example Container Configuration

```go
containerConfig := &container.Config{
    Image:      "kali:latest",                    // AI-selected or default image
    Hostname:   "a1b2c3d4",                      // Generated from container name
    WorkingDir: "/work",                         // Standard working directory
    Entrypoint: []string{"tail", "-f", "/dev/null"}, // Keep container running
    ExposedPorts: nat.PortSet{
        "28000/tcp": {},                         // Flow-specific ports
        "28001/tcp": {},
    },
}

pidsLimit := int64(2048) // fork-bomb guard, default when the caller does not set one

hostConfig := &container.HostConfig{
    CapDrop: []string{"ALL"},                    // Explicit allow-list below, see Capability Management
    CapAdd: []string{
        "CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER",
        "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
        "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
    },
    PidsLimit: &pidsLimit,
    RestartPolicy: container.RestartPolicy{
        Name:              "on-failure",         // Restart failed containers only
        MaximumRetryCount: 5,
    },
    Binds: []string{
        "/host/data/flow-123:/work",            // Work directory mount
        // Only when DOCKER_INSIDE=true and a socket was selected; with
        // DOCKER_INSIDE_HOST set instead, no socket is mounted at all.
        "/var/run/docker.sock:/var/run/docker.sock",
    },
    PortBindings: nat.PortMap{
        "28000/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "28000"}},
        "28001/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "28001"}},
    },
}
```

### Container States and Transitions

PentAGI tracks container states in the database:

- **`Starting`**: Container creation in progress
- **`Running`**: Container is active and available
- **`Stopped`**: Container has been stopped but not removed
- **`Failed`**: Container creation or startup failed
- **`Deleted`**: Container has been removed

### Container Naming Convention

Containers follow a specific naming pattern for easy identification:

```go
func PrimaryTerminalName(tenantPrefix string, flowID int64) string {
    return fmt.Sprintf("%s%s%d", tenantPrefix, PrimaryTerminalNamePrefix, flowID)
}
```

This creates names like `pentagi-terminal-123` for flow ID 123, making it easy to:
- Identify containers belonging to specific flows
- Perform flow-based cleanup operations
- Debug container-related issues

When `TENANT_ID` is set the tenant leads the name — `acme-pentagi-terminal-123` — so that several PentAGI instances can share one worker node. The prefix goes in front deliberately: sweeps that match `pentagi-terminal-*` then reach only their own instance's containers. The per-flow volume (`<container name>-data`) and the container hostname derive from this name, so both inherit the scoping. Containers and volumes additionally carry a `pentagi.tenant` label for filter-based cleanup.

### Cleanup Operations

The `Cleanup` method performs comprehensive cleanup:

1. **Flow State Assessment**:
   - Identifies flows that should be terminated
   - Marks incomplete flows as failed
   - Preserves running flows that should continue

2. **Container Cleanup**:
   - Stops all containers for terminated flows
   - Removes stopped containers and their volumes
   - Updates database to reflect current state

3. **Parallel Processing**:
   - Uses goroutines for concurrent container deletion
   - Ensures cleanup doesn't block system operation

## Security and Isolation

### Container Security Model

PentAGI implements a multi-layered security approach for container isolation:

#### Network Isolation
- **Custom Networks**: Containers run in dedicated Docker networks
- **Port Control**: Only specific ports are exposed to the host
- **Host Protection**: Container cannot access host network by default

#### File System Isolation
- **Read-Only Root**: Base container filesystem is immutable
- **Controlled Mounts**: Only specific directories are writable
- **Volume Separation**: Each flow gets isolated storage space

#### Capability Management

The primary container uses an explicit allow-list instead of Docker's implicit defaults: `CapDrop: ["ALL"]`, then `CapAdd` back Docker's own default 14-capability set minus `MKNOD`, plus `NET_ADMIN` when `DOCKER_NET_ADMIN=true` and `SYS_PTRACE` (one deliberate addition beyond Docker's defaults, see below):

```go
hostConfig := &container.HostConfig{
    CapDrop: []string{"ALL"},
    CapAdd: []string{
        "CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER",
        "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
        "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
        // + "NET_ADMIN" when DOCKER_NET_ADMIN=true
    },
}
```

Why this exact set, rather than a minimal `NET_RAW`-only list:

| Capability | Why it is needed |
|---|---|
| `NET_RAW` | Raw sockets — nmap, ping, packet crafting |
| `NET_BIND_SERVICE` | Bind ports below 1024 — reverse shells, Responder, rogue DNS |
| `SETUID` / `SETGID` | Daemons and tools that drop privileges after starting as root |
| `SETFCAP` / `SETPCAP` | `apt`/`dpkg` `postinst` scripts that `setcap` network tools instead of relying on setuid (`ping`, `traceroute`, `nmap`, `dumpcap`, `hping3`, …) — without these, on-the-fly package installs (the Installer Agent's core job) fail |
| `FSETID` | Preserves set-id bits when dpkg installs/modifies files as a non-owner |
| `CHOWN` / `DAC_OVERRIDE` / `FOWNER` | Root file-permission overrides needed during package installs and builds |
| `KILL` | Signal other processes inside the container |
| `SYS_CHROOT` | chroot-based isolation within the sandbox |
| `AUDIT_WRITE` | Lets `sudo`/`sshd` write audit-log entries instead of warning |
| `SYS_PTRACE` | Not a Docker default — added so `gdb`/`strace`/`ltrace`/dynamic binary analysis (`pwndbg`, `radare2`) work for the Coder Agent's exploit-development role. Without it, `ptrace()` and friends (`process_vm_readv`/`writev`, `kcmp`) stay blocked by Docker's *default seccomp profile*, which independently gates them behind `CAP_SYS_PTRACE` — no custom seccomp profile is needed to unblock them, since moby/containerd auto-extend the default profile's syscall allow-list to match added capabilities. Scope stays contained to the sandbox: `ptrace` only works within the container's own PID namespace, never against host or sibling-container processes. |
| `NET_ADMIN` (opt-in via `DOCKER_NET_ADMIN`) | Interface/routing/firewall control for advanced network pentesting |

**`MKNOD` is the one deliberate omission** from Docker's default set: creating device nodes has no legitimate use for pentest tooling or package management, and this repository's own dind-hardening research (see [Worker Node Setup](../../examples/guides/worker_node.md) and its [`authz.rego`](../../examples/guides/worker_node/authz.rego)) identifies block-device `mknod` combined with `debugfs` as "the primary confirmed escape vector" for a hostile-code container. `SYS_ADMIN`, `SYS_MODULE`, `SYS_RAWIO`, and `SYS_BOOT` are never granted — none are part of Docker's default set and none are required by any supported workflow.

Docker's default set minus `MKNOD`, plus `NET_ADMIN`, is exactly the `allowed_caps` whitelist already vetted in `authz.rego` for nested dind containers running the same kind of pentest workload — the two are kept intentionally consistent at that shared baseline. `SYS_PTRACE` is the one place the primary container's allow-list goes further than `authz.rego`'s: it is not offered to nested dind containers (an agent there can request arbitrary `containers/create` calls, and the dind threat model does not special-case debugging), but the primary worker container is created solely by PentAGI itself with a fixed capability list, so granting it here does not expand what an agent can ask for.

An earlier revision of this container also forced `no-new-privileges:true` via `SecurityOpt`, intended as defense-in-depth against setuid/file-capability escalation. It was removed: the capability bounding set above already caps what any process can ever gain regardless of setuid, so the flag added no protection beyond the allow-list while unconditionally breaking SUID/SGID privilege-escalation testing and `sudo`/`su` from a non-root shell — both routine penetration-testing workflows.

#### Resource Limits
- **PidsLimit**: Defaults to 2048 when the caller does not set one — a cheap fork-bomb / resource-exhaustion guard, generous enough for parallel scans (nmap, hydra). Mirrors the same default used for the dind daemon in the [Worker Node Setup](../../examples/guides/worker_node.md) guide.
- **Memory/CPU**: Controlled via standard `HostConfig` resource fields when set by the caller.

#### Process Isolation
- **User Namespaces**: Containers run with isolated user space
- **PID Isolation**: Container processes are isolated from host
- **Resource Limits**: Memory and CPU usage are controlled

### Security Best Practices Implemented

1. **Image Validation**: All images are pulled and verified before use
2. **Fallback Strategy**: Safe default image used if custom image fails
3. **State Tracking**: All container operations are logged and monitored
4. **Automatic Cleanup**: Failed or abandoned containers are automatically removed
5. **Socket Security**: Docker socket is only mounted when explicitly required

## Integration with PentAGI

### Tool Integration

The Docker client integrates with PentAGI's tool system to provide terminal access:

```go
type terminal struct {
    flowID       int64
    containerID  int64
    containerLID string
    dockerClient docker.DockerClient
    tlp          TermLogProvider
}
```

The terminal tool uses the Docker client for:
- **Command Execution**: Running shell commands in isolated containers
- **File Operations**: Reading and writing files safely
- **Result Capture**: Collecting command output and artifacts

### Flow File Integration

Flow files are managed by the REST API in `pkg/server/services/flow_files.go` and use Docker client file APIs for synchronization with the running primary container.

PentAGI keeps two different storage areas for flow files:

- **Local cache**: `{DATA_DIR}/flow-{id}-data/uploads` and `{DATA_DIR}/flow-{id}-data/container`
- **Container workspace**: `/work` inside the primary container

This separation is intentional. It supports both single-node deployments and remote worker-node deployments where the backend host filesystem is not the same filesystem used by Docker workers.

The current behavior is:

- User uploads are saved to the local cache under `uploads/`.
- If the primary container is running, uploaded files are pushed best-effort to `/work/uploads`.
- When the primary container starts or is reused, cached uploads are synchronized into `/work/uploads`; the cache is the source of truth.
- Files pulled from the container are stored under `container/` using their normalized full container path, for example:
  - `/etc/nginx/nginx.conf` -> `container/etc/nginx/nginx.conf`
  - `/work/test.md` -> `container/work/test.md`
- Deleting cached upload files is allowed even when the container is not running. The next container start will resynchronize `/work/uploads` from cache.

The flow files API also exposes a non-recursive live container directory listing endpoint. It uses `ContainerStatPath` to determine whether the requested path is a file or directory:

- If the path is a file, it returns that file metadata directly.
- If the path is a directory, it calls `ListContainerDir`.
- If the path is omitted, it defaults to `/work`.

### Provider Integration

The provider system uses Docker client for environment preparation:

```go
// In providers.go
type flowProvider struct {
    // ... other fields
    docker    docker.DockerClient
    publicIP  string
}
```

Providers use the Docker client to:
- **Image Selection**: AI agents choose appropriate container images
- **Environment Setup**: Prepare containers for specific tasks
- **Resource Management**: Allocate and deallocate containers as needed

### Database Integration

Container states are persisted in the PostgreSQL database:

```sql
-- Container state tracking
CREATE TABLE containers (
    id SERIAL PRIMARY KEY,
    flow_id INTEGER REFERENCES flows(id),
    name VARCHAR NOT NULL,
    image VARCHAR NOT NULL,
    status container_status NOT NULL,
    local_id VARCHAR,
    local_dir VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Observability Integration

All Docker operations are instrumented with:
- **Structured Logging**: JSON logs with context and metadata
- **Error Tracking**: Comprehensive error capture and reporting
- **Performance Metrics**: Container creation and execution timing
- **Resource Monitoring**: CPU, memory, and network usage tracking

## Usage Examples

### Basic Container Creation

```go
// Initialize Docker client
dockerClient, err := docker.NewDockerClient(ctx, db, cfg)
if err != nil {
    return fmt.Errorf("failed to create docker client: %w", err)
}

// Create container for a flow
containerName := tools.PrimaryTerminalName(cfg.TenantPrefix(), flowID)
container, err := dockerClient.RunContainer(
    ctx,
    containerName,
    database.ContainerTypePrimary,
    flowID,
    &container.Config{
        Image:      "kali:latest",
        Entrypoint: []string{"tail", "-f", "/dev/null"},
    },
    &container.HostConfig{
        // See Container Capability Assignment above for the full allow-list
        // that flowToolsExecutor.Prepare actually passes in production.
        CapDrop: []string{"ALL"},
        CapAdd: []string{
            "CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER",
            "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
            "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
            "NET_ADMIN",
        },
    },
)
```

### Command Execution

```go
// Execute command in container
createResp, err := dockerClient.ContainerExecCreate(ctx, containerName, container.ExecOptions{
    Cmd:          []string{"sh", "-c", "nmap -sS 192.168.1.1"},
    AttachStdout: true,
    AttachStderr: true,
    WorkingDir:   "/work",
    Tty:          true,
})

// Attach to execution
resp, err := dockerClient.ContainerExecAttach(ctx, createResp.ID, container.ExecAttachOptions{
    Tty: true,
})

// Read output
output, err := io.ReadAll(resp.Reader)
```

### File Operations

```go
// Write file to container
content := "#!/bin/bash\necho 'Hello from container'"
archive := createTarArchive("script.sh", content)
err := dockerClient.CopyToContainer(ctx, containerID, "/work", archive, container.CopyToContainerOptions{})

// Read file from container
reader, stats, err := dockerClient.CopyFromContainer(ctx, containerID, "/work/results.txt")
defer reader.Close()

// Extract content from tar
content := extractFromTar(reader)

// Stat a file or directory in the container
stat, err := dockerClient.ContainerStatPath(ctx, containerID, "/work/results.txt")

// List direct entries in a container directory
entries, err := dockerClient.ListContainerDir(ctx, containerID, "/work")
```

### Container Directory Listing

`ListContainerDir` performs a non-recursive directory listing inside a running container:

1. Uses `ContainerStatPath` to verify that `dirPath` exists and is a directory.
2. Executes `find <dirPath> -maxdepth 1 -mindepth 1 ! -name '.*' -print0` (no TTY) inside the container. `find -print0` emits literal, NUL-delimited entry paths, so names with spaces, newlines, or non-UTF8 bytes survive intact — the old `ls -1` parse mangled them.
3. Stats every entry concurrently through an `errgroup` bounded to `containerListWorkers = 20`, preserving input order.

The method returns a `ContainerDirListing`: `Files` are the readable entries' `container.PathStat` metadata, `Failures` carries any per-entry stat errors so a live directory degrades to a partial listing instead of failing outright, and `Truncated` is set when the directory held more than the entry cap and only the first page was listed.

If `dirPath` is empty, it defaults to `WorkFolderPathInContainer` (`/work`).

### Cleanup and Resource Management

```go
// Check if container is running
isRunning, err := dockerClient.IsContainerRunning(ctx, containerID)

// Stop container
err = dockerClient.StopContainer(ctx, containerID, dbID)

// Remove container and volumes
err = dockerClient.RemoveContainer(ctx, containerID, dbID)

// Global cleanup (usually called on startup)
err = dockerClient.Cleanup(ctx)
```

### Error Handling

```go
// The client implements comprehensive error handling
container, err := dockerClient.RunContainer(ctx, name, containerType, flowID, config, hostConfig)
if err != nil {
    // Errors include:
    // - Image pull failures (handled with fallback)
    // - Container creation failures
    // - Network configuration issues
    // - Database update failures

    // The client automatically:
    // - Updates database with failure status
    // - Cleans up partially created resources
    // - Logs detailed error information

    return fmt.Errorf("container creation failed: %w", err)
}
```

## Error Handling

### Error Categories

The Docker client handles several categories of errors:

1. **Docker Daemon Errors**:
   - Connection failures to Docker daemon
   - API version mismatches
   - Permission issues

2. **Image-Related Errors**:
   - Image pull failures (network, authentication)
   - Invalid image names or tags
   - Image compatibility issues

3. **Container Runtime Errors**:
   - Container creation failures
   - Container startup issues
   - Resource allocation problems

4. **Network and Storage Errors**:
   - Port binding conflicts
   - Volume mount failures
   - Network configuration issues

### Error Recovery Strategies

1. **Image Fallback**:
   ```go
   if err := dc.pullImage(ctx, config.Image); err != nil {
       logger.WithError(err).Warnf("failed to pull image '%s', using default", config.Image)
       config.Image = dc.defImage
       // Retry with default image
   }
   ```

2. **Container Cleanup**:
   ```go
   if containerCreationFails {
       defer updateContainerInfo(database.ContainerStatusFailed, containerID)
       // Clean up any partially created resources
   }
   ```

3. **State Synchronization**:
   - Database state always reflects actual container state
   - Failed operations are marked appropriately
   - Orphaned resources are cleaned up automatically

## Best Practices

### Resource Management
- Always use the `Cleanup()` method on application startup
- Monitor container resource usage through observability tools
- Set appropriate timeouts for long-running operations
- Use deterministic port allocation to avoid conflicts

### Security Considerations
- Regularly update base images used for containers
- Minimize capabilities granted to containers
- Use dedicated networks for container communication
- Monitor and audit all container operations

### Development and Debugging
- Use structured logging for all Docker operations
- Implement comprehensive error handling with context
- Test container operations in isolated environments
- Use the ftester utility for debugging specific operations

### Performance Optimization
- Reuse containers when possible instead of creating new ones
- Implement efficient cleanup to prevent resource leaks
- Use appropriate container restart policies
- Monitor container startup times and optimize configurations

### Integration Guidelines
- Always use the DockerClient interface instead of direct Docker SDK calls
- Integrate with PentAGI's database for state management
- Use the provided logging and observability infrastructure
- Follow the established naming conventions for containers
