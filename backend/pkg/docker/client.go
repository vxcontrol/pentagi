package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	WorkFolderPathInContainer = "/work"
	WorkerVolumeNameSuffix    = "-data"
	BaseContainerPortsNumber  = 28000
)

const (
	defaultImage                = "debian:latest"
	defaultDockerSocketPath     = "/var/run/docker.sock"
	containerPrimaryTypePattern = "-terminal-"
	containerLocalCwdTemplate   = "flow-%d"
	containerPortsNumber        = 2
	limitContainerPortsNumber   = 2000
	containerListWorkers        = 20
	// maxListEntries caps how many directory children a single listing will stat,
	// so a hostile /work with hundreds of thousands of files can't fan out into
	// that many Docker API calls per request.
	maxListEntries = 10000
	// maxListStdoutBytes bounds the exec stdout buffered before parsing so a
	// compromised sandbox can't stream unbounded output into memory; sized as the
	// entry cap × a generous per-path length (PATH_MAX).
	maxListStdoutBytes = maxListEntries * 4096
	// containerStartupGrace is how long a freshly started container has to stay
	// alive before it is accepted as started. Starting is asynchronous: the
	// daemon acknowledges the request as soon as the process is spawned, so an
	// entrypoint that dies milliseconds later still looks like a successful
	// start. Everything downstream assumes a live sandbox, so it is verified
	// here instead.
	containerStartupGrace = 1500 * time.Millisecond
	// maxStartupLogBytes bounds the output tail attached to a startup failure so
	// a container that floods its log before dying can't inflate the error.
	maxStartupLogBytes = 4096
	// startupLogTailLines is how many trailing log lines are quoted in a startup
	// failure; enough for a stack trace or a usage message.
	startupLogTailLines = "20"
	// containerDiscardTimeout bounds the cleanup of a container that failed to
	// start, so an unresponsive daemon delays the error instead of withholding
	// it indefinitely.
	containerDiscardTimeout = 30 * time.Second
)

type dockerClient struct {
	db             database.Querier
	logger         *logrus.Logger
	dataDir        string
	hostDir        string
	client         *client.Client
	inside         bool
	defImage       string
	socket         string
	network        string
	publicIP       string
	portsBase      int
	labels         map[string]string
	insideEnv      []string
	insideCertPath string
}

type DockerClient interface {
	RunContainer(ctx context.Context, containerName string, containerType database.ContainerType,
		flowID int64, config *container.Config, hostConfig *container.HostConfig) (database.Container, error)
	StopContainer(ctx context.Context, containerID string, dbID int64) error
	RemoveContainer(ctx context.Context, containerID string, dbID int64) error
	IsContainerRunning(ctx context.Context, containerID string) (bool, error)
	ContainerExecCreate(ctx context.Context, container string, config client.ExecCreateOptions) (client.ExecCreateResult, error)
	ContainerExecAttach(ctx context.Context, execID string, config client.ExecAttachOptions) (client.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (client.ExecInspectResult, error)
	ContainerStatPath(ctx context.Context, containerID string, path string) (container.PathStat, error)
	ListContainerDir(ctx context.Context, containerID string, dirPath string) (ContainerDirListing, error)
	CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options client.CopyToContainerOptions) error
	CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error)
	Cleanup(ctx context.Context) error
	GetDefaultImage() string
}

// GetPrimaryContainerPorts returns the host ports for a flow relative to
// portsBase. A portsBase of 0 falls back to BaseContainerPortsNumber.
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

func NewDockerClient(ctx context.Context, db database.Querier, cfg *config.Config) (DockerClient, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	infoResult, err := cli.Info(ctx, client.InfoOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get docker info: %w", err)
	}
	info := infoResult.Info

	// Resolve which host socket (if any) gets bind-mounted into worker containers.
	// Autodetection is skipped when DOCKER_INSIDE_HOST designates a daemon
	// endpoint for sandboxes: mounting the host socket alongside it would grant an
	// agent control of the daemon running PentAGI itself.
	socket, autodetectSocket := cfg.WorkerDockerSocket()
	if autodetectSocket {
		socket = getHostDockerSocket(ctx, cli)
	}
	inside := cfg.DockerInside
	if inside {
		switch {
		case cfg.DockerSocket != "":
			logrus.Infof("DOCKER_INSIDE=true: worker containers will be given Docker access "+
				"via the configured socket %q.", cfg.DockerSocket)
		case cfg.DockerInsideHost != "":
			logrus.Infof("DOCKER_INSIDE=true: worker containers will be given Docker access "+
				"to the configured external daemon at %q.", cfg.DockerInsideHost)
		default:
			logrus.Warn("DOCKER_INSIDE=true with neither DOCKER_SOCKET nor DOCKER_INSIDE_HOST set: " +
				"the host Docker socket will be autodetected and bind-mounted into every worker " +
				"container, so any process inside it gets control of the same daemon that runs " +
				"PentAGI. Set DOCKER_SOCKET or DOCKER_INSIDE_HOST explicitly, or front the socket " +
				"with a least-privilege proxy (e.g. Tecnativa/docker-socket-proxy), if that is not intended.")
		}
	}
	netName := cfg.DockerNetwork
	publicIP := cfg.DockerPublicIP
	defImage := strings.ToLower(cfg.DockerDefaultImage)
	if defImage == "" {
		defImage = defaultImage
	}

	dataDir, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tmp directory: %w", err)
	}

	hostDir := getHostDataDir(ctx, cli, dataDir, cfg.DockerWorkDir)

	// ensure network exists if configured
	if err := ensureDockerNetwork(ctx, cli, netName); err != nil {
		return nil, fmt.Errorf("failed to ensure docker network %s: %w", netName, err)
	}

	logger := logrus.StandardLogger()
	logger.WithFields(logrus.Fields{
		"docker_name":        info.Name,
		"docker_arch":        info.Architecture,
		"docker_version":     info.ServerVersion,
		"client_version":     cli.ClientVersion(),
		"data_dir":           dataDir,
		"host_dir":           hostDir,
		"docker_inside":      inside,
		"docker_socket":      socket,
		"docker_inside_host": cfg.DockerInsideHost,
		"public_ip":          publicIP,
	}).Debug("Docker client initialized")

	return &dockerClient{
		db:             db,
		client:         cli,
		dataDir:        dataDir,
		hostDir:        hostDir,
		logger:         logger,
		inside:         inside,
		defImage:       defImage,
		socket:         socket,
		network:        netName,
		publicIP:       publicIP,
		portsBase:      cfg.DockerPortsBase,
		labels:         cfg.TenantLabels(),
		insideEnv:      cfg.WorkerDockerEnv(),
		insideCertPath: cfg.WorkerDockerCertPath(),
	}, nil
}

func (dc *dockerClient) RunContainer(
	ctx context.Context,
	containerName string,
	containerType database.ContainerType,
	flowID int64,
	config *container.Config,
	hostConfig *container.HostConfig,
) (database.Container, error) {
	if config == nil {
		return database.Container{}, fmt.Errorf("no config found for container %s", containerName)
	}

	workDir := filepath.Join(dc.dataDir, fmt.Sprintf(containerLocalCwdTemplate, flowID))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return database.Container{}, fmt.Errorf("failed to create tmp directory: %w", err)
	}

	hostDir := dc.hostDir
	if hostDir != "" {
		hostDir = filepath.Join(hostDir, fmt.Sprintf(containerLocalCwdTemplate, flowID))
	}

	logger := dc.logger.WithContext(ctx).WithFields(logrus.Fields{
		"image":    config.Image,
		"name":     containerName,
		"type":     containerType,
		"flow_id":  flowID,
		"work_dir": workDir,
		"host_dir": hostDir,
	})
	logger.Info("running container")

	dbContainer, err := dc.db.CreateContainer(ctx, database.CreateContainerParams{
		Type:     containerType,
		Name:     containerName,
		Image:    config.Image,
		Status:   database.ContainerStatusStarting,
		FlowID:   flowID,
		LocalID:  database.StringToNullString(fmt.Sprintf("tmp-id-%d", flowID)),
		LocalDir: database.StringToNullString(hostDir),
	})
	if err != nil {
		return database.Container{}, fmt.Errorf("failed to create container in database: %w", err)
	}

	updateContainerInfo := func(status database.ContainerStatus, localID string) {
		dbContainer, err = dc.db.UpdateContainerStatusLocalID(ctx, database.UpdateContainerStatusLocalIDParams{
			Status:  status,
			LocalID: database.StringToNullString(localID),
			ID:      dbContainer.ID,
		})
		if err != nil {
			logger.WithError(err).Error("failed to update container info in database")
		}
	}

	fallbackDockerImage := func() error {
		logger = logger.WithField("image", dc.defImage)
		logger.Warn("try to use default image")
		config.Image = dc.defImage

		dbContainer, err = dc.db.UpdateContainerImage(ctx, database.UpdateContainerImageParams{
			Image: config.Image,
			ID:    dbContainer.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to update container image in database: %w", err)
		}

		if err := dc.pullImage(ctx, config.Image); err != nil {
			return fmt.Errorf("failed to pull default image '%s': %w", config.Image, err)
		}

		return nil
	}

	if err := dc.pullImage(ctx, config.Image); err != nil {
		logger.WithError(err).Warnf("failed to pull image '%s' and using default image", config.Image)
		if err := fallbackDockerImage(); err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, err
		}
	}

	logger.Info("creating container")

	config.Hostname = fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(containerName)))
	config.WorkingDir = WorkFolderPathInContainer

	// Tag the container with its owning tenant so daemon-wide sweeps can filter on
	// ownership rather than matching name prefixes. Absent when no tenant is
	// configured, keeping the created container byte-identical to before.
	for k, v := range dc.labels {
		if config.Labels == nil {
			config.Labels = make(map[string]string, len(dc.labels))
		}
		config.Labels[k] = v
	}

	if hostConfig == nil {
		hostConfig = &container.HostConfig{}
	}

	// prevent containers from auto-starting after OS or docker daemon restart
	// because on startup they create docker.sock directory for DinD if it's enabled
	hostConfig.RestartPolicy = container.RestartPolicy{
		Name:              container.RestartPolicyOnFailure,
		MaximumRetryCount: 5,
	}

	if hostDir == "" {
		volumeName, err := dc.client.VolumeCreate(ctx, client.VolumeCreateOptions{
			Name:   fmt.Sprintf("%s%s", containerName, WorkerVolumeNameSuffix),
			Driver: "local",
			Labels: dc.labels,
		})
		if err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, fmt.Errorf("failed to create volume: %w", err)
		}
		hostDir = volumeName.Volume.Name
	}
	hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", hostDir, WorkFolderPathInContainer))

	if dc.inside {
		// The socket is empty when DOCKER_INSIDE_HOST designates a daemon endpoint
		// instead; binding "" would produce a malformed mount spec.
		if dc.socket != "" {
			hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", dc.socket, defaultDockerSocketPath))
		}

		// Point the sandbox's Docker CLI at the designated daemon.
		config.Env = append(config.Env, dc.insideEnv...)

		// TLS material is mounted read-only at the same path on both sides so the
		// injected DOCKER_CERT_PATH resolves unchanged inside the container.
		if dc.insideCertPath != "" {
			hostConfig.Binds = append(hostConfig.Binds,
				fmt.Sprintf("%s:%s:ro", dc.insideCertPath, dc.insideCertPath))
		}
	}

	// no-new-privileges was evaluated and deliberately not applied: the capability
	// bounding set in tools.go Prepare already caps what any process can gain, so
	// it added no protection beyond that (and none against issue #337) while
	// breaking SUID/SGID privesc testing and sudo/su. See "Capability Management"
	// in docker.md for the full rationale.

	// Cap fork-bomb / resource-exhaustion risk at a bounded limit.
	// 2048 pids is generous for most pentest workloads (nmap, hydra, parallel scans).
	// Callers can override by setting hostConfig.PidsLimit before calling RunContainer.
	if hostConfig.PidsLimit == nil {
		pidsLimit := int64(2048)
		hostConfig.PidsLimit = &pidsLimit
	}

	hostConfig.LogConfig = container.LogConfig{
		Type: "json-file",
		Config: map[string]string{
			"max-size": "10m",
			"max-file": "5",
		},
	}

	// Configure network mode and port bindings
	var networkingConfig *network.NetworkingConfig
	if dc.network == "host" {
		// Host network mode: container uses host network stack directly
		// No port bindings needed as container has direct access to host interfaces
		hostConfig.NetworkMode = container.NetworkMode("host")
		logger.Debug("using host network mode - container will have direct access to host network interfaces")
	} else {
		// Bridge network mode: configure port bindings and custom network
		if hostConfig.PortBindings == nil {
			hostConfig.PortBindings = network.PortMap{}
		}
		if config.ExposedPorts == nil {
			config.ExposedPorts = network.PortSet{}
		}
		var hostIP netip.Addr
		if dc.publicIP != "" {
			var err error
			hostIP, err = netip.ParseAddr(dc.publicIP)
			if err != nil {
				defer updateContainerInfo(database.ContainerStatusFailed, "")
				return database.Container{}, fmt.Errorf("invalid Docker public IP %q: %w", dc.publicIP, err)
			}
		}
		for _, port := range GetPrimaryContainerPorts(dc.portsBase, flowID) {
			containerPort, ok := network.PortFrom(uint16(port), network.TCP)
			if !ok {
				return database.Container{}, fmt.Errorf("invalid container port %d", port)
			}
			hostConfig.PortBindings[containerPort] = []network.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: fmt.Sprintf("%d", port),
				},
			}
			config.ExposedPorts[containerPort] = struct{}{}
		}

		if dc.network != "" {
			networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					dc.network: {},
				},
			}
		}
	}

	// Config is referenced by pointer, so a later image fallback is picked up by
	// the retries below without rebuilding the options.
	createOptions := client.ContainerCreateOptions{
		Config:           config,
		HostConfig:       hostConfig,
		NetworkingConfig: networkingConfig,
		Name:             containerName,
	}

	resp, err := dc.client.ContainerCreate(ctx, createOptions)
	if err != nil && cerrdefs.IsConflict(err) {
		// The name is still held by an earlier container, typically one left
		// behind when PentAGI died mid-start. Nothing can be created under that
		// name until it is gone, and the image is not at fault, so clear it and
		// retry before blaming the image below.
		logger.WithError(err).Warn("container name is already taken, removing the container holding it")
		if staleErr := dc.removeContainerByName(ctx, containerName); staleErr != nil {
			logger.WithError(staleErr).Warn("failed to remove the container holding the name")
		} else {
			resp, err = dc.client.ContainerCreate(ctx, createOptions)
		}
	}
	if err != nil {
		if config.Image == dc.defImage {
			logger.WithError(err).Warn("failed to create container with default image")
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, fmt.Errorf("failed to create container: %w", err)
		}

		logger.WithError(err).Warn("failed to create container, try to use default image")
		if err := fallbackDockerImage(); err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, err
		}

		// try to cleanup previous container
		if err := dc.removeContainerByName(ctx, containerName); err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, err
		}

		// try to create container again with default image
		resp, err = dc.client.ContainerCreate(ctx, createOptions)
		if err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, fmt.Errorf("failed to create container '%s': %w", config.Image, err)
		}
	}

	containerID := resp.ID
	logger = logger.WithField("local_id", containerID)
	logger.Info("container created")

	// Arm the exit watch before starting. The restart policy replaces a container
	// that dies with a fresh attempt within milliseconds, and from then on the
	// daemon reports the state of that new attempt — the status code that
	// actually killed the sandbox is no longer reachable. Watching from before
	// the start is what makes it observable at all.
	waitCtx, cancelWait := context.WithCancel(ctx)
	defer cancelWait()
	exitWatch := dc.client.ContainerWait(waitCtx, containerID, client.ContainerWaitOptions{
		Condition: container.WaitConditionNextExit,
	})

	_, err = dc.client.ContainerStart(ctx, containerID, client.ContainerStartOptions{})
	if err != nil {
		defer updateContainerInfo(database.ContainerStatusFailed, containerID)
		dc.discardContainer(ctx, containerID, logger)
		return database.Container{}, fmt.Errorf("failed to start container: %w", err)
	}

	// The start call above only proves the daemon accepted the request. Callers
	// hand this container straight to the agents, so a sandbox that is already
	// dead has to be reported here rather than degrade into unrelated exec
	// failures against a container that never ran.
	if err := dc.ensureContainerStarted(ctx, containerName, containerID, exitWatch); err != nil {
		defer updateContainerInfo(database.ContainerStatusFailed, containerID)
		logger.WithError(err).Error("container did not stay running after start")
		dc.discardContainer(ctx, containerID, logger)
		return database.Container{}, err
	}

	logger.Info("container started")
	updateContainerInfo(database.ContainerStatusRunning, containerID)

	return dbContainer, nil
}

// ContainerStartupError reports a container that the daemon started but that did
// not stay up. It carries the daemon's own diagnostics together with a tail of
// the container output, so the reason is visible where the failure surfaces
// instead of only on the host the sandbox ran on.
type ContainerStartupError struct {
	ContainerName string
	ContainerID   string
	Status        string
	ExitCode      int
	RestartCount  int
	OOMKilled     bool
	DaemonError   string
	LogTail       string
}

func (e *ContainerStartupError) Error() string {
	message := fmt.Sprintf("container '%s' did not stay running after start (state '%s', exit code %d)",
		e.ContainerName, e.Status, e.ExitCode)
	if e.RestartCount > 0 {
		message += fmt.Sprintf(", restarted %d time(s)", e.RestartCount)
	}
	if e.OOMKilled {
		message += ", terminated by the OOM killer"
	}
	if e.DaemonError != "" {
		message += fmt.Sprintf(", daemon reported: %s", e.DaemonError)
	}
	if e.LogTail != "" {
		message += fmt.Sprintf(", last output: %s", e.LogTail)
	}
	return message
}

// ensureContainerStarted confirms the container is really running. It reports
// the exit as soon as it happens, and otherwise waits out containerStartupGrace:
// an entrypoint that dies milliseconds after start would still be acknowledged
// as a successful start, so surviving the window is the only evidence that the
// sandbox is usable.
func (dc *dockerClient) ensureContainerStarted(
	ctx context.Context,
	containerName, containerID string,
	exitWatch client.ContainerWaitResult,
) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("interrupted while starting container '%s': %w", containerName, ctx.Err())

	case exit := <-exitWatch.Result:
		startupErr := dc.newStartupError(ctx, containerName, containerID, int(exit.StatusCode))
		if exit.Error != nil {
			startupErr.DaemonError = exit.Error.Message
		}
		return startupErr

	case err := <-exitWatch.Error:
		// The watch itself broke, which says nothing about the container. Ask
		// the daemon directly rather than assuming either outcome.
		return dc.verifyContainerRunning(ctx, containerName, containerID, err)

	case <-time.After(containerStartupGrace):
		// Still alive. Drain the watch so the reader goroutine behind it can
		// finish once the request is cancelled on the way out.
		go func() {
			select {
			case <-exitWatch.Result:
			case <-exitWatch.Error:
			}
		}()
		return nil
	}
}

// verifyContainerRunning is the fallback for when the exit watch cannot be used:
// the container state is read directly, so a broken watch degrades into a slower
// check instead of into a sandbox that is wrongly assumed to be alive.
func (dc *dockerClient) verifyContainerRunning(
	ctx context.Context, containerName, containerID string, watchErr error,
) error {
	inspectResult, err := dc.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to check container '%s' after start (exit watch failed: %v): %w",
			containerName, watchErr, err)
	}

	inspection := inspectResult.Container
	state := inspection.State
	if state == nil {
		// the daemon always reports state for a container it just started, so
		// its absence means the container is not in a usable state
		return fmt.Errorf("no state reported for container '%s' after start", containerName)
	}

	// A recorded restart is a failure of its own: under the on-failure policy a
	// crash-looping sandbox keeps coming back up without ever being usable.
	if state.Running && !state.Restarting && inspection.RestartCount == 0 {
		return nil
	}

	startupErr := dc.newStartupError(ctx, containerName, containerID, state.ExitCode)
	startupErr.DaemonError = state.Error
	return startupErr
}

// newStartupError assembles the report for a container that did not stay up,
// enriching the authoritative exit code with whatever the daemon still knows and
// with the container's own last words.
func (dc *dockerClient) newStartupError(
	ctx context.Context, containerName, containerID string, exitCode int,
) *ContainerStartupError {
	startupErr := &ContainerStartupError{
		ContainerName: containerName,
		ContainerID:   containerID,
		ExitCode:      exitCode,
		Status:        "exited",
	}

	inspectResult, err := dc.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return startupErr
	}

	inspection := inspectResult.Container
	if state := inspection.State; state != nil {
		startupErr.Status = string(state.Status)
		startupErr.OOMKilled = state.OOMKilled
	}
	startupErr.RestartCount = inspection.RestartCount
	startupErr.LogTail = dc.containerLogTail(ctx, containerID, inspection.Config)

	return startupErr
}

// containerLogTail returns the tail of a container's combined output. It is
// best effort by design: this only ever enriches a failure that has already been
// decided, so a log that cannot be read must not replace the real error.
func (dc *dockerClient) containerLogTail(ctx context.Context, containerID string, config *container.Config) string {
	logs, err := dc.client.ContainerLogs(ctx, containerID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       startupLogTailLines,
	})
	if err != nil {
		return ""
	}
	defer logs.Close()

	var output bytes.Buffer
	reader := io.LimitReader(logs, maxStartupLogBytes)
	if config != nil && config.Tty {
		// a TTY stream is the raw container output, without stream framing
		_, _ = io.Copy(&output, reader)
	} else {
		// stdout and stderr are interleaved as framed chunks; merge both, since
		// the reason a container died is usually on stderr
		_, _ = stdcopy.StdCopy(&output, &output, reader)
	}

	return strings.TrimSpace(output.String())
}

// removeContainerByName drops whichever container currently holds the name, so a
// leftover from an earlier run cannot block every future create with a name
// conflict. The daemon reports names with a leading slash and omits non-running
// containers unless all of them are requested, and a stale container is usually
// exited rather than running.
func (dc *dockerClient) removeContainerByName(ctx context.Context, containerName string) error {
	containerList, err := dc.client.ContainerList(ctx, client.ContainerListOptions{
		All: true,
		// the name filter matches substrings, so the exact comparison below
		// still decides; the filter only keeps the listing small
		Filters: make(client.Filters).Add("name", containerName),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	options := client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	for _, existing := range containerList.Items {
		if !slices.ContainsFunc(existing.Names, func(name string) bool {
			return strings.TrimPrefix(name, "/") == containerName
		}) {
			continue
		}

		if _, err := dc.client.ContainerRemove(ctx, existing.ID, options); err != nil && !cerrdefs.IsNotFound(err) {
			return fmt.Errorf("failed to remove container '%s' holding name '%s': %w", existing.ID, containerName, err)
		}
	}

	return nil
}

// discardContainer removes a container that never became usable, so the failure
// leaves nothing behind: no dead container accumulating on the host and no name
// blocking the retry. Removal is detached from the caller's context because a
// cancelled flow is exactly when the leftover would otherwise survive, and any
// failure is logged rather than returned so it cannot mask the startup error.
func (dc *dockerClient) discardContainer(ctx context.Context, containerID string, logger *logrus.Entry) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), containerDiscardTimeout)
	defer cancel()

	_, err := dc.client.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil && !cerrdefs.IsNotFound(err) {
		logger.WithError(err).Error("failed to remove the container that did not start")
	}
}

func (dc *dockerClient) StopContainer(ctx context.Context, containerID string, dbID int64) error {
	logger := dc.logger.WithContext(ctx).WithField("local_id", containerID)
	logger.Info("initiating container shutdown sequence")

	_, stopErr := dc.client.ContainerStop(ctx, containerID, client.ContainerStopOptions{})
	if stopErr != nil {
		if cerrdefs.IsNotFound(stopErr) {
			logger.Warn("target container already removed or never existed")
		} else {
			return fmt.Errorf("container shutdown failed: %w", stopErr)
		}
	}

	_, err := dc.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		Status: database.ContainerStatusStopped,
		ID:     dbID,
	})
	if err != nil {
		return fmt.Errorf("database status update failed during container stop: %w", err)
	}

	logger.Info("container shutdown completed successfully")

	return nil
}

func (dc *dockerClient) RemoveContainer(ctx context.Context, containerID string, dbID int64) error {
	logger := dc.logger.WithContext(ctx).WithField("local_id", containerID)
	logger.Info("removing container and associated resources")

	if err := dc.StopContainer(ctx, containerID, dbID); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	options := client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if _, err := dc.client.ContainerRemove(ctx, containerID, options); err != nil {
		if !cerrdefs.IsNotFound(err) {
			return fmt.Errorf("failed to remove container: %w", err)
		}
		// already gone (removed manually, or a prior call already succeeded);
		// still mark it deleted below so the database row does not go stale.
		logger.WithError(err).Warn("container not found")
	}

	_, err := dc.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		Status: database.ContainerStatusDeleted,
		ID:     dbID,
	})
	if err != nil {
		return fmt.Errorf("failed to update container status to deleted: %w", err)
	}

	logger.Info("container removed")

	return nil
}

func (dc *dockerClient) Cleanup(ctx context.Context) error {
	logger := dc.logger.WithContext(ctx).WithField("docker", "cleanup")
	logger.Info("cleaning up containers and making all flows finished...")

	flows, err := dc.db.GetFlows(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all flows: %w", err)
	}

	containers, err := dc.db.GetContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all containers: %w", err)
	}

	flowsStatusMap := make(map[int64]database.FlowStatus)
	for _, flow := range flows {
		flowsStatusMap[flow.ID] = flow.Status
	}
	flowContainersMap := make(map[int64][]database.Container)
	for _, container := range containers {
		flowContainersMap[container.FlowID] = append(flowContainersMap[container.FlowID], container)
	}

	var wg sync.WaitGroup
	removeContainer := func(containerID string, dbID int64) {
		defer wg.Done()
		logger := logger.WithField("local_id", containerID)

		if err := dc.RemoveContainer(ctx, containerID, dbID); err != nil {
			logger.WithError(err).Errorf("failed to remove container")
		}

		_, err := dc.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
			Status: database.ContainerStatusDeleted,
			ID:     dbID,
		})
		if err != nil {
			logger.WithError(err).Errorf("failed to update container status to deleted")
		}
	}
	isAllContainersRunning := func(flowID int64) bool {
		containers, ok := flowContainersMap[flowID]
		if !ok || len(containers) == 0 {
			return false
		}
		for _, container := range containers {
			switch container.Status {
			case database.ContainerStatusStarting, database.ContainerStatusRunning:
				return false
			}
		}
		return true
	}
	markFlowAsFailed := func(flowID int64) {
		logger := logger.WithField("flow_id", flowID)
		_, err := dc.db.UpdateFlowStatus(ctx, database.UpdateFlowStatusParams{
			Status: database.FlowStatusFailed,
			ID:     flowID,
		})
		if err != nil {
			logger.WithError(err).Errorf("failed to update flow status to failed")
		}
	}

	for _, flow := range flows {
		switch flowsStatusMap[flow.ID] {
		case database.FlowStatusRunning, database.FlowStatusWaiting:
			if isAllContainersRunning(flow.ID) {
				continue
			}
			fallthrough
		case database.FlowStatusCreated:
			markFlowAsFailed(flow.ID)
			fallthrough
		default: // FlowStatusFinished, FlowStatusFailed
			for _, container := range flowContainersMap[flow.ID] {
				switch container.Status {
				case database.ContainerStatusStarting, database.ContainerStatusRunning:
					wg.Add(1)
					go removeContainer(container.LocalID.String, container.ID)
				}
			}
		}
	}

	wg.Wait()
	logger.Info("cleanup finished")

	return nil
}

func (dc *dockerClient) IsContainerRunning(ctx context.Context, containerID string) (bool, error) {
	inspectResult, err := dc.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		if !cerrdefs.IsNotFound(err) {
			return false, fmt.Errorf("container inspection failed: %w", err)
		}
		// a removed container is missing, not an inspection failure
		return false, nil
	}

	inspection := inspectResult.Container
	if inspection.State == nil {
		// the daemon always populates State for a successfully inspected
		// container; treat the unexpected absence as "not running" rather
		// than panicking on the field access below.
		return false, nil
	}

	// Check both Running state and health status if available
	isOperational := inspection.State.Running
	if inspection.State.Health != nil && inspection.State.Health.Status != "" {
		isOperational = isOperational && inspection.State.Health.Status != "unhealthy"
	}

	return isOperational, nil
}

func (dc *dockerClient) GetDefaultImage() string {
	return dc.defImage
}

func (dc *dockerClient) ContainerExecCreate(
	ctx context.Context,
	container string,
	config client.ExecCreateOptions,
) (client.ExecCreateResult, error) {
	return dc.client.ExecCreate(ctx, container, config)
}

func (dc *dockerClient) ContainerExecAttach(
	ctx context.Context,
	execID string,
	config client.ExecAttachOptions,
) (client.HijackedResponse, error) {
	result, err := dc.client.ExecAttach(ctx, execID, config)
	return result.HijackedResponse, err
}

func (dc *dockerClient) ContainerExecInspect(
	ctx context.Context,
	execID string,
) (client.ExecInspectResult, error) {
	return dc.client.ExecInspect(ctx, execID, client.ExecInspectOptions{})
}

func (dc *dockerClient) ContainerStatPath(
	ctx context.Context,
	containerID string,
	path string,
) (container.PathStat, error) {
	result, err := dc.client.ContainerStatPath(ctx, containerID, client.ContainerStatPathOptions{Path: path})
	return result.Stat, err
}

// ContainerEntryError is a directory entry that could not be stat'd during a
// listing; the rest of the listing is still returned.
type ContainerEntryError struct {
	Name string
	Path string
	Err  error
}

// ContainerDirListing is the result of ListContainerDir: the entries that
// stat'd successfully, plus per-entry failures for entries that individually
// couldn't be read. A directory-level fault — path not a directory, the list
// command failed, the request was cancelled, or every entry failed (which
// signals the container is gone) — is returned as an error instead, so a
// partial listing never silently hides a systemic failure.
type ContainerDirListing struct {
	Files    []container.PathStat
	Failures []ContainerEntryError
	// Truncated is set when the directory held more than maxListEntries children
	// and only the first maxListEntries were listed.
	Truncated bool
}

func (dc *dockerClient) ListContainerDir(
	ctx context.Context,
	containerID string,
	dirPath string,
) (ContainerDirListing, error) {
	if strings.TrimSpace(dirPath) == "" {
		dirPath = WorkFolderPathInContainer
	}

	dirStat, err := dc.ContainerStatPath(ctx, containerID, dirPath)
	if err != nil {
		return ContainerDirListing{}, fmt.Errorf("failed to stat container path '%s': %w", dirPath, err)
	}
	if !dirStat.Mode.IsDir() {
		return ContainerDirListing{}, fmt.Errorf("container path '%s' is not a directory", dirPath)
	}

	// List direct children NUL-delimited. Parsing `ls` output is unsafe: under a
	// TTY GNU coreutils shell-quotes names and busybox wraps them in ANSI escapes,
	// so a readable file with a space / quote / non-ASCII byte would be mis-stat'd
	// and reported unreadable. `find -print0` emits literal bytes and is portable
	// (GNU + busybox); the NUL delimiter also survives names containing newlines.
	// No TTY: a TTY's onlcr would rewrite every \n in the stream to \r\n —
	// including a \n that is part of a filename — corrupting the name. Without a
	// TTY the exec stream is multiplexed and demuxed below.
	createResp, err := dc.ContainerExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          []string{"find", dirPath, "-maxdepth", "1", "-mindepth", "1", "!", "-name", ".*", "-print0"},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return ContainerDirListing{}, fmt.Errorf("failed to create list exec for '%s': %w", dirPath, err)
	}

	resp, err := dc.ContainerExecAttach(ctx, createResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return ContainerDirListing{}, fmt.Errorf("failed to attach list exec for '%s': %w", dirPath, err)
	}
	output, readErr := demuxExecStdout(resp.Reader, maxListStdoutBytes)
	resp.Close()
	if readErr != nil {
		return ContainerDirListing{}, fmt.Errorf("failed to read list output for '%s': %w", dirPath, readErr)
	}

	inspect, err := dc.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return ContainerDirListing{}, fmt.Errorf("failed to inspect list exec for '%s': %w", dirPath, err)
	}
	if inspect.ExitCode != 0 {
		return ContainerDirListing{}, fmt.Errorf("list command failed for '%s' with exit code %d: %s", dirPath, inspect.ExitCode, string(output))
	}

	entryPaths, truncated := parseFindEntries(output)

	stats, failures := statContainerEntries(ctx, entryPaths, containerListWorkers, func(ctx context.Context, entryPath string) (container.PathStat, error) {
		return dc.ContainerStatPath(ctx, containerID, entryPath)
	})

	// A cancelled context is a directory-level fault (the client is gone), not a
	// partial listing, so surface it as an error. Entries that individually failed
	// to stat are carried in Failures — the find exec already proved the container
	// alive, so a live directory degrades rather than 500s even if every entry failed.
	if err := ctx.Err(); err != nil {
		return ContainerDirListing{}, fmt.Errorf("listing container directory '%s': %w", dirPath, err)
	}

	listing := ContainerDirListing{Files: stats, Truncated: truncated}
	for _, f := range failures {
		listing.Failures = append(listing.Failures, ContainerEntryError{
			Name: path.Base(f.name),
			Path: f.name,
			Err:  f.err,
		})
	}

	return listing, nil
}

// parseFindEntries splits `find -print0` output (NUL-delimited absolute paths),
// dropping empties, and caps the result at maxListEntries — returning truncated=true
// so the caller reports the partiality instead of failing or silently dropping.
func parseFindEntries(output []byte) (entries []string, truncated bool) {
	for _, tok := range strings.Split(string(output), "\x00") {
		if tok == "" {
			continue
		}
		entries = append(entries, tok)
	}
	if len(entries) > maxListEntries {
		return entries[:maxListEntries], true
	}
	return entries, false
}

// demuxExecStdout reads a non-TTY Docker exec stream — stdout and stderr
// interleaved as frames with an 8-byte header (stream id + big-endian size) —
// and returns only the stdout bytes, erroring if stdout exceeds maxStdout so a
// compromised sandbox can't stream unbounded output into memory.
func demuxExecStdout(r io.Reader, maxStdout int) ([]byte, error) {
	var stdout bytes.Buffer
	header := make([]byte, 8)
	for {
		if _, err := io.ReadFull(r, header); err != nil {
			if err == io.EOF {
				break // clean end at a frame boundary
			}
			// A header cut short (ErrUnexpectedEOF) means the stream was truncated
			// mid-frame — the listing is incomplete, so fail rather than silently
			// dropping the tail.
			return nil, fmt.Errorf("truncated exec stream: %w", err)
		}
		size := int64(binary.BigEndian.Uint32(header[4:8]))
		if size == 0 {
			continue
		}
		switch header[0] {
		case 1: // stdout
			if _, err := io.CopyN(&stdout, r, size); err != nil {
				return nil, err
			}
			if stdout.Len() > maxStdout {
				return nil, fmt.Errorf("listing output exceeded %d bytes", maxStdout)
			}
		case 3: // systemerr — a daemon-level error injected mid-stream; surface it
			var msg bytes.Buffer
			_, _ = io.CopyN(&msg, r, size)
			return nil, fmt.Errorf("docker exec systemerr: %s", strings.TrimSpace(msg.String()))
		default: // stderr and anything else — discard
			if _, err := io.CopyN(io.Discard, r, size); err != nil {
				return nil, err
			}
		}
	}
	return stdout.Bytes(), nil
}

type statFailure struct {
	name string
	err  error
}

// statContainerEntries stats every name concurrently, bounded to `workers`
// in-flight calls. A per-entry stat error never aborts the batch: the successful
// stats and the failures are both returned (in input order) so the caller can
// serve a partial listing rather than discarding everything on one bad entry.
func statContainerEntries(
	ctx context.Context,
	names []string,
	workers int,
	statFn func(context.Context, string) (container.PathStat, error),
) ([]container.PathStat, []statFailure) {
	// errgroup.SetLimit(0) blocks the first Go() forever and SetLimit(<0) runs
	// unbounded; clamp a non-positive count to the standard bound so the pool
	// never deadlocks or floods the Docker daemon.
	if workers <= 0 {
		workers = containerListWorkers
	}

	stats := make([]container.PathStat, len(names))
	errs := make([]error, len(names))

	var g errgroup.Group
	g.SetLimit(workers)
	for i, name := range names {
		g.Go(func() error {
			if stat, err := statFn(ctx, name); err != nil {
				errs[i] = err
			} else {
				stats[i] = stat
			}
			return nil
		})
	}
	_ = g.Wait()

	oks := make([]container.PathStat, 0, len(names))
	var failures []statFailure
	for i, name := range names {
		if errs[i] != nil {
			failures = append(failures, statFailure{name: name, err: errs[i]})
		} else {
			oks = append(oks, stats[i])
		}
	}

	return oks, failures
}

func (dc *dockerClient) CopyToContainer(
	ctx context.Context,
	containerID string,
	dstPath string,
	content io.Reader,
	options client.CopyToContainerOptions,
) error {
	options.DestinationPath = dstPath
	options.Content = content
	_, err := dc.client.CopyToContainer(ctx, containerID, options)
	return err
}

func (dc *dockerClient) CopyFromContainer(
	ctx context.Context,
	containerID string,
	srcPath string,
) (io.ReadCloser, container.PathStat, error) {
	result, err := dc.client.CopyFromContainer(ctx, containerID, client.CopyFromContainerOptions{SourcePath: srcPath})
	return result.Content, result.Stat, err
}

func (dc *dockerClient) pullImage(ctx context.Context, imageName string) error {
	filterArgs := make(client.Filters).Add("reference", imageName)
	images, err := dc.client.ImageList(ctx, client.ImageListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if imageExistsLocally := len(images.Items) > 0; imageExistsLocally {
		return nil
	}

	dc.logger.WithContext(ctx).WithField("image", imageName).Info("initiating image download from registry...")

	pullStream, err := dc.client.ImagePull(ctx, imageName, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer pullStream.Close()

	// drain pull stream to completion
	if _, err := io.Copy(io.Discard, pullStream); err != nil {
		return fmt.Errorf("image download stream processing failed: %w", err)
	}

	dc.logger.WithContext(ctx).WithField("image", imageName).Debug("image pull completed")

	return nil
}

func getHostDockerSocket(ctx context.Context, cli *client.Client) string {
	daemonHost := strings.TrimPrefix(cli.DaemonHost(), "unix://")
	if info, err := os.Stat(daemonHost); err != nil || info.IsDir() {
		return defaultDockerSocketPath
	}

	hostname, err := os.Hostname()
	if err != nil {
		return daemonHost
	}

	filterArgs := make(client.Filters).Add("status", "running")

	containerList, err := cli.ContainerList(ctx, client.ContainerListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return daemonHost
	}

	for _, container := range containerList.Items {
		result, err := cli.ContainerInspect(ctx, container.ID, client.ContainerInspectOptions{})
		if err != nil {
			continue
		}

		inspect := result.Container
		if inspect.Config.Hostname != hostname {
			continue
		}

		for _, mount := range inspect.Mounts {
			if mount.Destination == daemonHost {
				return mount.Source
			}
		}
	}

	return daemonHost
}

// return empty string if dataDir should be unique dedicated volume
// otherwise return the path to the host's file system data directory or custom workDir
func getHostDataDir(ctx context.Context, cli *client.Client, dataDir, workDir string) string {
	if workDir != "" {
		return workDir
	}

	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	filterArgs := make(client.Filters).Add("status", "running")

	containerList, err := cli.ContainerList(ctx, client.ContainerListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return "" // unexpected error
	}

	mounts := []container.MountPoint{}
	for _, container := range containerList.Items {
		result, err := cli.ContainerInspect(ctx, container.ID, client.ContainerInspectOptions{})
		if err != nil {
			continue
		}

		inspect := result.Container
		if inspect.Config.Hostname != hostname {
			continue
		}

		for _, mount := range inspect.Mounts {
			if strings.HasPrefix(dataDir, mount.Destination) {
				mounts = append(mounts, mount)
			}
		}
	}

	if len(mounts) == 0 {
		// it's for the following cases:
		// * docker socket hosted on the different machine
		// * data directory is not mounted
		// * pentagi is not running as a docker container
		return ""
	}

	// sort mounts by destination length to get the most accurate mount point
	slices.SortFunc(mounts, func(a, b container.MountPoint) int {
		return len(b.Destination) - len(a.Destination)
	})

	// get more accurate path to the data directory
	mountPoint := mounts[0]
	switch mountPoint.Type {
	case mount.TypeBind:
		deltaPath := strings.TrimPrefix(dataDir, mountPoint.Destination)
		return filepath.Join(mountPoint.Source, deltaPath)
	default:
		// skip volume mount type because it leads to unexpected behavior
		// e.g. macOS or Windows usually mounts directory from the docker VM
		// and it's not the same as the host machine's directory
		return ""
	}
}

// ensureDockerNetwork verifies that a docker network with the given name exists;
// if it does not, it attempts to create it.
// Special case: "host" network mode is built-in and doesn't need creation.
func ensureDockerNetwork(ctx context.Context, cli *client.Client, name string) error {
	if name == "" || name == "host" {
		return nil
	}

	if _, err := cli.NetworkInspect(ctx, name, client.NetworkInspectOptions{}); err == nil {
		return nil
	}

	_, err := cli.NetworkCreate(ctx, name, client.NetworkCreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return fmt.Errorf("failed to create network %s: %w", name, err)
	}

	return nil
}
