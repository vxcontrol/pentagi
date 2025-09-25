# Docker Client Package Documentation

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
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
- **File Operations**: Safe file transfer between host and containers
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
const WorkFolderPathInContainer = "/work"              // Standard working directory in containers
const BaseContainerPortsNumber = 28000                // Starting port number for dynamic allocation
const defaultImage = "debian:latest"                  // Fallback image if custom image fails
const containerPortsNumber = 2                        // Number of ports allocated per container
const limitContainerPortsNumber = 2000                // Maximum port range for allocation
```

### Port Allocation Strategy

PentAGI uses a deterministic port allocation algorithm to ensure each flow gets unique, predictable ports:

```go
func GetPrimaryContainerPorts(flowID int64) []int {
    ports := make([]int, containerPortsNumber)
    for i := 0; i < containerPortsNumber; i++ {
        delta := (int(flowID)*containerPortsNumber + i) % limitContainerPortsNumber
        ports[i] = BaseContainerPortsNumber + delta
    }
    return ports
}
```

This ensures that:
- Each flow gets consistent port numbers across restarts
- Port conflicts are avoided between different flows
- Ports are within a controlled range (28000-30000)

## Configuration

### Environment Variables

The Docker client is configured through several environment variables defined in the main configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon connection |
| `DOCKER_INSIDE` | `false` | Whether PentAGI communicates with host Docker daemon from containers |
| `DOCKER_NET_ADMIN` | `false` | Whether PentAGI grants the primary container NET_ADMIN capability for advanced networking. |
| `DOCKER_SOCKET` | `/var/run/docker.sock` | Path to Docker socket on host |
| `DOCKER_NETWORK` | | Docker network for container communication |
| `DOCKER_PUBLIC_IP` | `0.0.0.0` | Public IP for port binding |
| `DOCKER_WORK_DIR` | | Custom work directory path on host |
| `DOCKER_DEFAULT_IMAGE` | `debian:latest` | Fallback image if AI-selected image fails |
| `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` | `vxcontrol/kali-linux` | Default Docker image for penetration testing tasks |
| `DATA_DIR` | `./data` | Local data directory for file operations |

### Configuration Structure

```go
type Config struct {
    // Docker (terminal) settings
    DockerInside                 bool   `env:"DOCKER_INSIDE" envDefault:"false"`
    DockerNetAdmin               bool   `env:"DOCKER_NET_ADMIN" envDefault:"false"`
    DockerSocket                 string `env:"DOCKER_SOCKET"`
    DockerNetwork                string `env:"DOCKER_NETWORK"`
    DockerPublicIP               string `env:"DOCKER_PUBLIC_IP" envDefault:"0.0.0.0"`
    DockerWorkDir                string `env:"DOCKER_WORK_DIR"`
    DockerDefaultImage           string `env:"DOCKER_DEFAULT_IMAGE" envDefault:"debian:latest"`
    DockerDefaultImageForPentest string `env:"DOCKER_DEFAULT_IMAGE_FOR_PENTEST" envDefault:"vxcontrol/kali-linux"`
    DataDir                      string `env:"DATA_DIR" envDefault:"./data"`
}
```

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

The NET_ADMIN capability is applied differently based on container type and configuration:

```go
// Primary containers (when DOCKER_NET_ADMIN=true)
hostConfig := &container.HostConfig{
    CapAdd: []string{"NET_RAW", "NET_ADMIN"},  // Full networking capabilities
    // ... other configurations
}

// Primary containers (when DOCKER_NET_ADMIN=false)
hostConfig := &container.HostConfig{
    CapAdd: []string{"NET_RAW"},  // Basic raw socket access only
    // ... other configurations
}
```

### Docker-in-Docker Support

PentAGI supports running inside Docker containers while still managing other containers. This is controlled by the `DOCKER_INSIDE` setting:

- **`DOCKER_INSIDE=false`**: PentAGI runs on host, manages containers directly
- **`DOCKER_INSIDE=true`**: PentAGI runs in container, mounts Docker socket to manage sibling containers

### Network Configuration

When `DOCKER_NETWORK` is specified, all containers are automatically connected to this network, enabling:
- Isolated communication between PentAGI components
- Controlled access to external networks
- Service discovery within the PentAGI ecosystem

## Core Interfaces

### DockerClient Interface

The main interface defines all Docker operations available to PentAGI components:

```go
type DockerClient interface {
    // Container lifecycle management
    SpawnContainer(ctx context.Context, containerName string, containerType database.ContainerType,
        flowID int64, config *container.Config, hostConfig *container.HostConfig) (database.Container, error)
    StopContainer(ctx context.Context, containerID string, dbID int64) error
    DeleteContainer(ctx context.Context, containerID string, dbID int64) error
    IsContainerRunning(ctx context.Context, containerID string) (bool, error)

    // Command execution
    ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (container.ExecCreateResponse, error)
    ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error)
    ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)

    // File operations
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
    db       database.Querier     // Database for container state management
    logger   *logrus.Logger       // Structured logging
    dataDir  string               // Local data directory
    hostDir  string               // Host-mapped data directory
    client   *client.Client       // Docker SDK client
    inside   bool                 // Running inside Docker
    defImage string               // Default fallback image
    socket   string               // Docker socket path
    network  string               // Docker network name
    publicIP string               // Public IP for port binding
}
```

## Container Lifecycle Management

### Container Creation Process

The `SpawnContainer` method handles the complete container creation workflow:

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
   - Sets up restart policy (`unless-stopped`)
   - Configures logging (JSON driver with rotation)

4. **Storage Setup**:
   - Creates dedicated volume or bind mount
   - Mounts work directory to `/work` in container
   - Optionally mounts Docker socket for Docker-in-Docker

5. **Network and Ports**:
   - Assigns flow-specific ports using deterministic algorithm
   - Connects to specified Docker network if configured
   - Binds ports to public IP

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

hostConfig := &container.HostConfig{
    CapAdd: []string{"NET_RAW"},                 // Required capabilities for network tools
    RestartPolicy: container.RestartPolicy{
        Name: "unless-stopped",                  // Auto-restart unless explicitly stopped
    },
    Binds: []string{
        "/host/data/flow-123:/work",            // Work directory mount
        "/var/run/docker.sock:/var/run/docker.sock", // Docker socket (if inside Docker)
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
func PrimaryTerminalName(flowID int64) string {
    return fmt.Sprintf("pentagi-terminal-%d", flowID)
}
```

This creates names like `pentagi-terminal-123` for flow ID 123, making it easy to:
- Identify containers belonging to specific flows
- Perform flow-based cleanup operations
- Debug container-related issues

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
```go
hostConfig := &container.HostConfig{
    CapAdd: []string{"NET_RAW"},  // Required for network scanning tools
    // Other dangerous capabilities are not granted
}
```

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
containerName := docker.PrimaryTerminalName(flowID)
container, err := dockerClient.SpawnContainer(
    ctx,
    containerName,
    database.ContainerTypePrimary,
    flowID,
    &container.Config{
        Image:      "kali:latest",
        Entrypoint: []string{"tail", "-f", "/dev/null"},
    },
    &container.HostConfig{
        CapAdd: []string{"NET_RAW", "NET_ADMIN"},
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
```

### Cleanup and Resource Management

```go
// Check if container is running
isRunning, err := dockerClient.IsContainerRunning(ctx, containerID)

// Stop container
err = dockerClient.StopContainer(ctx, containerID, dbID)

// Remove container and volumes
err = dockerClient.DeleteContainer(ctx, containerID, dbID)

// Global cleanup (usually called on startup)
err = dockerClient.Cleanup(ctx)
```

### Error Handling

```go
// The client implements comprehensive error handling
container, err := dockerClient.SpawnContainer(ctx, name, containerType, flowID, config, hostConfig)
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
