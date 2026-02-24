package docker

import (
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"pentagi/pkg/config"
	"pentagi/pkg/database"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

const WorkFolderPathInContainer = "/work"

const BaseContainerPortsNumber = 28000

const (
	defaultImage                = "debian:latest"
	defaultDockerSocketPath     = "/var/run/docker.sock"
	containerPrimaryTypePattern = "-terminal-"
	containerLocalCwdTemplate   = "flow-%d"
	containerPortsNumber        = 2
	limitContainerPortsNumber   = 2000
)

type dockerClient struct {
	db       database.Querier
	logger   *logrus.Logger
	dataDir  string
	hostDir  string
	client   *client.Client
	inside   bool
	defImage string
	socket   string
	network  string
	publicIP string
}

type DockerClient interface {
	SpawnContainer(ctx context.Context, containerName string, containerType database.ContainerType,
		flowID int64, config *container.Config, hostConfig *container.HostConfig) (database.Container, error)
	StopContainer(ctx context.Context, containerID string, dbID int64) error
	DeleteContainer(ctx context.Context, containerID string, dbID int64) error
	IsContainerRunning(ctx context.Context, containerID string) (bool, error)
	ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (container.ExecCreateResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
	CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error
	CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error)
	Cleanup(ctx context.Context) error
	GetDefaultImage() string
}

func GetPrimaryContainerPorts(flowID int64) []int {
	ports := make([]int, containerPortsNumber)
	for i := 0; i < containerPortsNumber; i++ {
		delta := (int(flowID)*containerPortsNumber + i) % limitContainerPortsNumber
		ports[i] = BaseContainerPortsNumber + delta
	}
	return ports
}

func NewDockerClient(ctx context.Context, db database.Querier, cfg *config.Config) (DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}
	cli.NegotiateAPIVersion(ctx)

	info, err := cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker info: %w", err)
	}

	var socket string
	if cfg.DockerSocket != "" {
		socket = cfg.DockerSocket
	} else {
		socket = getHostDockerSocket(ctx, cli)
	}
	inside := cfg.DockerInside
	netName := cfg.DockerNetwork
	publicIP := cfg.DockerPublicIP
	defImage := strings.ToLower(cfg.DockerDefaultImage)
	if defImage == "" {
		defImage = defaultImage
	}

	// TODO: if this process running in a docker container, we need to use the host machine's data directory
	// maybe there need to resolve the data directory path from volume list
	// or maybe need to sync files from container to host machine
	// or disable passing data directory to the container
	// or create temporary volume for each container

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
		"docker_name":    info.Name,
		"docker_arch":    info.Architecture,
		"docker_version": info.ServerVersion,
		"client_version": cli.ClientVersion(),
		"data_dir":       dataDir,
		"host_dir":       hostDir,
		"docker_inside":  inside,
		"docker_socket":  socket,
		"public_ip":      publicIP,
	}).Debug("Docker client initialized")

	return &dockerClient{
		db:       db,
		client:   cli,
		dataDir:  dataDir,
		hostDir:  hostDir,
		logger:   logger,
		inside:   inside,
		defImage: defImage,
		socket:   socket,
		network:  netName,
		publicIP: publicIP,
	}, nil
}

func (dc *dockerClient) SpawnContainer(
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
	logger.Info("spawning container")

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
		volumeName, err := dc.client.VolumeCreate(ctx, volume.CreateOptions{
			Name:   fmt.Sprintf("%s-data", containerName),
			Driver: "local",
		})
		if err != nil {
			return database.Container{}, fmt.Errorf("failed to create volume: %w", err)
		}
		hostDir = volumeName.Name
	}
	hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", hostDir, WorkFolderPathInContainer))

	if dc.inside {
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", dc.socket, defaultDockerSocketPath))
	}

	hostConfig.LogConfig = container.LogConfig{
		Type: "json-file",
		Config: map[string]string{
			"max-size": "10m",
			"max-file": "5",
		},
	}

	if hostConfig.PortBindings == nil {
		hostConfig.PortBindings = nat.PortMap{}
	}
	if config.ExposedPorts == nil {
		config.ExposedPorts = nat.PortSet{}
	}
	for _, port := range GetPrimaryContainerPorts(flowID) {
		natPort := nat.Port(fmt.Sprintf("%d/tcp", port))
		hostConfig.PortBindings[natPort] = []nat.PortBinding{
			{
				HostIP:   dc.publicIP,
				HostPort: fmt.Sprintf("%d", port),
			},
		}
		config.ExposedPorts[natPort] = struct{}{}
	}

	var networkingConfig *network.NetworkingConfig
	if dc.network != "" {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				dc.network: {},
			},
		}
	}

	resp, err := dc.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
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
		containers, err := dc.client.ContainerList(ctx, container.ListOptions{})
		if err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, fmt.Errorf("failed to list containers: %w", err)
		}
		options := container.RemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}
		for _, container := range containers {
			// containerName is unique for PentAGI environment, so we can use it to find the container
			if len(container.Names) > 0 && container.Names[0] == containerName {
				_ = dc.client.ContainerRemove(ctx, container.ID, options)
			}
		}

		// try to create container again with default image
		resp, err = dc.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
		if err != nil {
			defer updateContainerInfo(database.ContainerStatusFailed, "")
			return database.Container{}, fmt.Errorf("failed to create container '%s': %w", config.Image, err)
		}
	}

	containerID := resp.ID
	logger = logger.WithField("local_id", containerID)
	logger.Info("container created")

	err = dc.client.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		defer updateContainerInfo(database.ContainerStatusFailed, containerID)
		return database.Container{}, fmt.Errorf("failed to start container: %w", err)
	}

	logger.Info("container started")
	updateContainerInfo(database.ContainerStatusRunning, containerID)

	return dbContainer, nil
}

func (dc *dockerClient) StopContainer(ctx context.Context, containerID string, dbID int64) error {
	logger := dc.logger.WithContext(ctx).WithField("local_id", containerID)
	logger.Info("stopping container")

	if err := dc.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		if client.IsErrNotFound(err) {
			logger.Warn("container not found")
		} else {
			return fmt.Errorf("failed to stop container: %w", err)
		}
	}

	_, err := dc.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		Status: database.ContainerStatusStopped,
		ID:     dbID,
	})
	if err != nil {
		return fmt.Errorf("failed to update container status to stopped: %w", err)
	}

	logger.Info("container stopped")

	return nil
}

func (dc *dockerClient) DeleteContainer(ctx context.Context, containerID string, dbID int64) error {
	logger := dc.logger.WithContext(ctx).WithField("local_id", containerID)
	logger.Info("deleting container")

	if err := dc.StopContainer(ctx, containerID, dbID); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	options := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := dc.client.ContainerRemove(ctx, containerID, options); err != nil {
		if !client.IsErrNotFound(err) {
			return fmt.Errorf("failed to remove container: %w", err)
		}
		// TODO: fix this case
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
	deleteContainer := func(containerID string, dbID int64) {
		defer wg.Done()
		logger := logger.WithField("local_id", containerID)

		if err := dc.DeleteContainer(ctx, containerID, dbID); err != nil {
			logger.WithError(err).Errorf("failed to delete container")
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
					go deleteContainer(container.LocalID.String, container.ID)
				}
			}
		}
	}

	wg.Wait()
	logger.Info("cleanup finished")

	return nil
}

func (dc *dockerClient) IsContainerRunning(ctx context.Context, containerID string) (bool, error) {
	containerInfo, err := dc.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container: %w", err)
	}

	return containerInfo.State.Running, err
}

func (dc *dockerClient) GetDefaultImage() string {
	return dc.defImage
}

func (dc *dockerClient) ContainerExecCreate(
	ctx context.Context,
	container string,
	config container.ExecOptions,
) (container.ExecCreateResponse, error) {
	return dc.client.ContainerExecCreate(ctx, container, config)
}

func (dc *dockerClient) ContainerExecAttach(
	ctx context.Context,
	execID string,
	config container.ExecAttachOptions,
) (types.HijackedResponse, error) {
	return dc.client.ContainerExecAttach(ctx, execID, config)
}

func (dc *dockerClient) ContainerExecInspect(
	ctx context.Context,
	execID string,
) (container.ExecInspect, error) {
	return dc.client.ContainerExecInspect(ctx, execID)
}

func (dc *dockerClient) CopyToContainer(
	ctx context.Context,
	containerID string,
	dstPath string,
	content io.Reader,
	options container.CopyToContainerOptions,
) error {
	return dc.client.CopyToContainer(ctx, containerID, dstPath, content, options)
}

func (dc *dockerClient) CopyFromContainer(
	ctx context.Context,
	containerID string,
	srcPath string,
) (io.ReadCloser, container.PathStat, error) {
	return dc.client.CopyFromContainer(ctx, containerID, srcPath)
}

func (dc *dockerClient) pullImage(ctx context.Context, imageName string) error {
	filters := filters.NewArgs()
	filters.Add("reference", imageName)
	images, err := dc.client.ImageList(ctx, image.ListOptions{
		Filters: filters,
	})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if imageExistsLocally := len(images) > 0; imageExistsLocally {
		return nil
	}

	dc.logger.WithContext(ctx).WithField("image", imageName).Info("pulling image...")

	readCloser, err := dc.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer readCloser.Close()

	// wait for the pull to finish
	_, err = io.Copy(io.Discard, readCloser)
	if err != nil {
		return fmt.Errorf("failed to wait for image pull: %w", err)
	}

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

	filterArgs := filters.NewArgs()
	filterArgs.Add("status", "running")

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return daemonHost
	}

	for _, container := range containers {
		inspect, err := cli.ContainerInspect(ctx, container.ID)
		if err != nil {
			continue
		}

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

	filterArgs := filters.NewArgs()
	filterArgs.Add("status", "running")

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return "" // unexpected error
	}

	mounts := []types.MountPoint{}
	for _, container := range containers {
		inspect, err := cli.ContainerInspect(ctx, container.ID)
		if err != nil {
			continue
		}

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
	slices.SortFunc(mounts, func(a, b types.MountPoint) int {
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
func ensureDockerNetwork(ctx context.Context, cli *client.Client, name string) error {
	if name == "" {
		return nil
	}

	if _, err := cli.NetworkInspect(ctx, name, network.InspectOptions{}); err == nil {
		return nil
	}

	_, err := cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return fmt.Errorf("failed to create network %s: %w", name, err)
	}

	return nil
}
