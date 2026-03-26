package checker

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"pentagi/cmd/installer/state"
	"pentagi/pkg/version"

	"github.com/docker/docker/client"
)

var (
	InstallerVersion = version.GetBinaryVersion()
	UserAgent        = "PentAGI-Installer/" + InstallerVersion
)

const (
	DockerComposeFile            = "docker-compose.yml"
	GraphitiComposeFile          = "docker-compose-graphiti.yml"
	LangfuseComposeFile          = "docker-compose-langfuse.yml"
	ObservabilityComposeFile     = "docker-compose-observability.yml"
	ExampleCustomConfigLLMFile   = "example.custom.provider.yml"
	ExampleOllamaConfigLLMFile   = "example.ollama.provider.yml"
	PentagiScriptFile            = "/usr/local/bin/pentagi"
	PentagiContainerName         = "pentagi"
	GraphitiContainerName        = "graphiti"
	Neo4jContainerName           = "neo4j"
	LangfuseWorkerContainerName  = "langfuse-worker"
	LangfuseWebContainerName     = "langfuse-web"
	GrafanaContainerName         = "grafana"
	OpenTelemetryContainerName   = "otel"
	DefaultImage                 = "debian:latest"
	DefaultImageForPentest       = "vxcontrol/kali-linux"
	DefaultGraphitiEndpoint      = "http://graphiti:8000"
	DefaultLangfuseEndpoint      = "http://langfuse-web:3000"
	DefaultObservabilityEndpoint = "otelcol:8148"
	DefaultLangfuseOtelEndpoint  = "http://otelcol:4318"
	DefaultUpdateServerEndpoint  = "https://update.pentagi.com"
	UpdatesCheckEndpoint         = "/api/v1/updates/check"
	MinFreeMemGB                 = 0.5
	MinFreeMemGBForPentagi       = 0.5
	MinFreeMemGBForGraphiti      = 2.0
	MinFreeMemGBForLangfuse      = 1.5
	MinFreeMemGBForObservability = 1.5
	MinFreeDiskGB                = 5.0
	MinFreeDiskGBForComponents   = 10.0
	MinFreeDiskGBPerComponents   = 2.0
	MinFreeDiskGBForWorkerImages = 25.0
)

var (
	ErrAppStateNotInitialized = errors.New("appState not initialized")
	ErrHandlerNotInitialized  = errors.New("handler not initialized")
)

type CheckResult struct {
	EnvFileExists           bool   `json:"env_file_exists" yaml:"env_file_exists"`
	DockerApiAccessible     bool   `json:"docker_api_accessible" yaml:"docker_api_accessible"`
	WorkerEnvApiAccessible  bool   `json:"worker_env_api_accessible" yaml:"worker_env_api_accessible"`
	WorkerImageExists       bool   `json:"worker_image_exists" yaml:"worker_image_exists"`
	DockerInstalled         bool   `json:"docker_installed" yaml:"docker_installed"`
	DockerComposeInstalled  bool   `json:"docker_compose_installed" yaml:"docker_compose_installed"`
	DockerVersion           string `json:"docker_version" yaml:"docker_version"`
	DockerVersionOK         bool   `json:"docker_version_ok" yaml:"docker_version_ok"`
	DockerComposeVersion    string `json:"docker_compose_version" yaml:"docker_compose_version"`
	DockerComposeVersionOK  bool   `json:"docker_compose_version_ok" yaml:"docker_compose_version_ok"`
	PentagiScriptInstalled  bool   `json:"pentagi_script_installed" yaml:"pentagi_script_installed"`
	PentagiExtracted        bool   `json:"pentagi_extracted" yaml:"pentagi_extracted"`
	PentagiInstalled        bool   `json:"pentagi_installed" yaml:"pentagi_installed"`
	PentagiRunning          bool   `json:"pentagi_running" yaml:"pentagi_running"`
	PentagiVolumesExist     bool   `json:"pentagi_volumes_exist" yaml:"pentagi_volumes_exist"`
	GraphitiConnected       bool   `json:"graphiti_connected" yaml:"graphiti_connected"`
	GraphitiExternal        bool   `json:"graphiti_external" yaml:"graphiti_external"`
	GraphitiExtracted       bool   `json:"graphiti_extracted" yaml:"graphiti_extracted"`
	GraphitiInstalled       bool   `json:"graphiti_installed" yaml:"graphiti_installed"`
	GraphitiRunning         bool   `json:"graphiti_running" yaml:"graphiti_running"`
	GraphitiVolumesExist    bool   `json:"graphiti_volumes_exist" yaml:"graphiti_volumes_exist"`
	LangfuseConnected       bool   `json:"langfuse_connected" yaml:"langfuse_connected"`
	LangfuseExternal        bool   `json:"langfuse_external" yaml:"langfuse_external"`
	LangfuseExtracted       bool   `json:"langfuse_extracted" yaml:"langfuse_extracted"`
	LangfuseInstalled       bool   `json:"langfuse_installed" yaml:"langfuse_installed"`
	LangfuseRunning         bool   `json:"langfuse_running" yaml:"langfuse_running"`
	LangfuseVolumesExist    bool   `json:"langfuse_volumes_exist" yaml:"langfuse_volumes_exist"`
	ObservabilityConnected  bool   `json:"observability_connected" yaml:"observability_connected"`
	ObservabilityExternal   bool   `json:"observability_external" yaml:"observability_external"`
	ObservabilityExtracted  bool   `json:"observability_extracted" yaml:"observability_extracted"`
	ObservabilityInstalled  bool   `json:"observability_installed" yaml:"observability_installed"`
	ObservabilityRunning    bool   `json:"observability_running" yaml:"observability_running"`
	SysNetworkOK            bool   `json:"sys_network_ok" yaml:"sys_network_ok"`
	SysCPUOK                bool   `json:"sys_cpu_ok" yaml:"sys_cpu_ok"`
	SysMemoryOK             bool   `json:"sys_memory_ok" yaml:"sys_memory_ok"`
	SysDiskFreeSpaceOK      bool   `json:"sys_disk_free_space_ok" yaml:"sys_disk_free_space_ok"`
	UpdateServerAccessible  bool   `json:"update_server_accessible" yaml:"update_server_accessible"`
	InstallerIsUpToDate     bool   `json:"installer_is_up_to_date" yaml:"installer_is_up_to_date"`
	PentagiIsUpToDate       bool   `json:"pentagi_is_up_to_date" yaml:"pentagi_is_up_to_date"`
	GraphitiIsUpToDate      bool   `json:"graphiti_is_up_to_date" yaml:"graphiti_is_up_to_date"`
	LangfuseIsUpToDate      bool   `json:"langfuse_is_up_to_date" yaml:"langfuse_is_up_to_date"`
	ObservabilityIsUpToDate bool   `json:"observability_is_up_to_date" yaml:"observability_is_up_to_date"`
	WorkerIsUpToDate        bool   `json:"worker_is_up_to_date" yaml:"worker_is_up_to_date"`

	// System resource details for UI display
	SysCPUCount        int             `json:"sys_cpu_count" yaml:"sys_cpu_count"`
	SysMemoryRequired  float64         `json:"sys_memory_required_gb" yaml:"sys_memory_required_gb"`
	SysMemoryAvailable float64         `json:"sys_memory_available_gb" yaml:"sys_memory_available_gb"`
	SysDiskRequired    float64         `json:"sys_disk_required_gb" yaml:"sys_disk_required_gb"`
	SysDiskAvailable   float64         `json:"sys_disk_available_gb" yaml:"sys_disk_available_gb"`
	SysNetworkFailures []string        `json:"sys_network_failures" yaml:"sys_network_failures"`
	DockerErrorType    DockerErrorType `json:"docker_error_type" yaml:"docker_error_type"`
	EnvDirWritable     bool            `json:"env_dir_writable" yaml:"env_dir_writable"`

	// handler controls how information is gathered. If nil, skip gathering
	handler CheckHandler
}

// CheckHandler defines how to gather information into a CheckResult
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

// Delegating methods that preserve public API
func (c *CheckResult) GatherAllInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherAllInfo(ctx, c)
}

func (c *CheckResult) GatherDockerInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherDockerInfo(ctx, c)
}

func (c *CheckResult) GatherWorkerInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherWorkerInfo(ctx, c)
}

func (c *CheckResult) GatherPentagiInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherPentagiInfo(ctx, c)
}

func (c *CheckResult) GatherGraphitiInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherGraphitiInfo(ctx, c)
}

func (c *CheckResult) GatherLangfuseInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherLangfuseInfo(ctx, c)
}

func (c *CheckResult) GatherObservabilityInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherObservabilityInfo(ctx, c)
}

func (c *CheckResult) GatherSystemInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherSystemInfo(ctx, c)
}

func (c *CheckResult) GatherUpdatesInfo(ctx context.Context) error {
	if c.handler == nil {
		return ErrHandlerNotInitialized
	}
	return c.handler.GatherUpdatesInfo(ctx, c)
}

func (c *CheckResult) IsReadyToContinue() bool {
	return c.EnvFileExists &&
		c.EnvDirWritable &&
		c.DockerApiAccessible &&
		c.WorkerEnvApiAccessible &&
		c.DockerComposeInstalled &&
		c.DockerVersionOK &&
		c.DockerComposeVersionOK &&
		c.SysNetworkOK &&
		c.SysCPUOK &&
		c.SysMemoryOK &&
		c.SysDiskFreeSpaceOK
}

// availability helpers for installer operations
// these functions centralize complex visibility/availability logic for UI

// CanStartAll returns true when at least one embedded stack is installed and not running
func (c *CheckResult) CanStartAll() bool {
	if c.PentagiInstalled && !c.PentagiRunning {
		return true
	}
	if c.GraphitiConnected && !c.GraphitiExternal && c.GraphitiInstalled && !c.GraphitiRunning {
		return true
	}
	if c.LangfuseConnected && !c.LangfuseExternal && c.LangfuseInstalled && !c.LangfuseRunning {
		return true
	}
	if c.ObservabilityConnected && !c.ObservabilityExternal && c.ObservabilityInstalled && !c.ObservabilityRunning {
		return true
	}
	return false
}

// CanStopAll returns true when any compose stack is running
func (c *CheckResult) CanStopAll() bool {
	return c.PentagiRunning || c.GraphitiRunning || c.LangfuseRunning || c.ObservabilityRunning
}

// CanRestartAll mirrors stop logic (requires running services)
func (c *CheckResult) CanRestartAll() bool { return c.CanStopAll() }

// CanDownloadWorker returns true when worker image is missing
func (c *CheckResult) CanDownloadWorker() bool { return !c.WorkerImageExists }

// CanUpdateWorker returns true when worker image exists but is not up to date
func (c *CheckResult) CanUpdateWorker() bool { return c.WorkerImageExists && !c.WorkerIsUpToDate }

// CanUpdateAll returns true when any installed stack has updates available
func (c *CheckResult) CanUpdateAll() bool {
	if c.PentagiInstalled && !c.PentagiIsUpToDate {
		return true
	}
	if c.GraphitiInstalled && !c.GraphitiIsUpToDate {
		return true
	}
	if c.LangfuseInstalled && !c.LangfuseIsUpToDate {
		return true
	}
	if c.ObservabilityInstalled && !c.ObservabilityIsUpToDate {
		return true
	}
	return false
}

// CanUpdateInstaller returns true when installer update is available and update server accessible
func (c *CheckResult) CanUpdateInstaller() bool {
	return !c.InstallerIsUpToDate && c.UpdateServerAccessible
}

// CanFactoryReset returns true when any compose stack is installed
func (c *CheckResult) CanFactoryReset() bool {
	return c.PentagiInstalled || c.GraphitiInstalled || c.LangfuseInstalled || c.ObservabilityInstalled
}

// CanRemoveAll returns true when any compose stack is installed
func (c *CheckResult) CanRemoveAll() bool { return c.CanFactoryReset() }

// CanPurgeAll returns true when any compose stack is installed
func (c *CheckResult) CanPurgeAll() bool { return c.CanFactoryReset() }

// CanResetPassword returns true when PentAGI is running
func (c *CheckResult) CanResetPassword() bool { return c.PentagiRunning }

// CanInstallAll returns true when main stack is not installed yet
func (c *CheckResult) CanInstallAll() bool { return !c.PentagiInstalled }

// defaultCheckHandler provides the existing implementation of gathering logic
type defaultCheckHandler struct {
	mx           *sync.Mutex
	appState     state.State
	dockerClient *client.Client
	workerClient *client.Client
}

func (h *defaultCheckHandler) GatherAllInfo(ctx context.Context, c *CheckResult) error {
	envPath := h.appState.GetEnvPath()
	c.EnvFileExists = checkFileExists(envPath) && checkFileIsReadable(envPath)
	if !c.EnvFileExists {
		return fmt.Errorf("environment file %s does not exist or is not readable", envPath)
	}

	// check write permissions to .env directory
	envDir := filepath.Dir(envPath)
	c.EnvDirWritable = checkDirIsWritable(envDir)

	if err := h.GatherDockerInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherWorkerInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherPentagiInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherGraphitiInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherLangfuseInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherObservabilityInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherSystemInfo(ctx, c); err != nil {
		return err
	}
	if err := h.GatherUpdatesInfo(ctx, c); err != nil {
		return err
	}

	return nil
}

func (h *defaultCheckHandler) GatherDockerInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	var cli *client.Client

	if cli, c.DockerErrorType = createDockerClientFromEnv(ctx); c.DockerErrorType != DockerErrorNone {
		c.DockerApiAccessible = false
		c.DockerInstalled = c.DockerErrorType != DockerErrorNotInstalled
		if c.DockerInstalled {
			version := checkDockerCliVersion()
			c.DockerVersion = version.Version
			c.DockerVersionOK = version.Valid
		}
	} else {
		h.dockerClient = cli
		c.DockerApiAccessible = true
		c.DockerInstalled = true

		version := checkDockerVersion(ctx, cli)
		c.DockerVersion = version.Version
		c.DockerVersionOK = version.Valid
	}

	composeVersion := checkDockerComposeVersion()
	c.DockerComposeInstalled = composeVersion.Version != ""
	c.DockerComposeVersion = composeVersion.Version
	c.DockerComposeVersionOK = composeVersion.Valid

	return nil
}

func (h *defaultCheckHandler) GatherWorkerInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	dockerHost := getEnvVar(h.appState, "DOCKER_HOST", "")
	dockerCertPath := getEnvVar(h.appState, "PENTAGI_DOCKER_CERT_PATH", "")
	dockerTLSVerify := getEnvVar(h.appState, "DOCKER_TLS_VERIFY", "") != ""

	cli, err := createDockerClient(dockerHost, dockerCertPath, dockerTLSVerify)
	if err != nil {
		// fallback to DOCKER_CERT_PATH for backward compatibility
		// this handles cases where migration failed or user manually edited .env
		// note: after migration, DOCKER_CERT_PATH contains container path, not host path
		dockerCertPath = getEnvVar(h.appState, "DOCKER_CERT_PATH", "")
		cli, err = createDockerClient(dockerHost, dockerCertPath, dockerTLSVerify)
		if err != nil {
			c.WorkerEnvApiAccessible = false
			c.WorkerImageExists = false
			return nil
		}
	}

	h.workerClient = cli
	c.WorkerEnvApiAccessible = true

	pentestImage := getEnvVar(h.appState, "DOCKER_DEFAULT_IMAGE_FOR_PENTEST", DefaultImageForPentest)
	c.WorkerImageExists = checkImageExists(ctx, cli, pentestImage)

	return nil
}

func (h *defaultCheckHandler) GatherPentagiInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	envDir := filepath.Dir(h.appState.GetEnvPath())
	dockerComposeFile := filepath.Join(envDir, DockerComposeFile)
	c.PentagiExtracted = checkFileExists(dockerComposeFile) &&
		checkFileExists(ExampleCustomConfigLLMFile) &&
		checkFileExists(ExampleOllamaConfigLLMFile)
	c.PentagiScriptInstalled = checkFileExists(PentagiScriptFile)

	if h.dockerClient != nil {
		exists, running := checkContainerExists(ctx, h.dockerClient, PentagiContainerName)
		c.PentagiInstalled = exists
		c.PentagiRunning = running

		// check if pentagi-related volumes exist (indicates previous installation)
		pentagiVolumes := []string{"pentagi-postgres-data", "pentagi-data", "pentagi-ssl", "scraper-ssl"}
		c.PentagiVolumesExist = checkVolumesExist(ctx, h.dockerClient, pentagiVolumes)
	}

	return nil
}

func (h *defaultCheckHandler) GatherGraphitiInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	graphitiEnabled := getEnvVar(h.appState, "GRAPHITI_ENABLED", "")
	graphitiURL := getEnvVar(h.appState, "GRAPHITI_URL", "")

	c.GraphitiConnected = graphitiEnabled == "true" && graphitiURL != ""
	c.GraphitiExternal = graphitiURL != DefaultGraphitiEndpoint

	envDir := filepath.Dir(h.appState.GetEnvPath())
	graphitiComposeFile := filepath.Join(envDir, GraphitiComposeFile)
	c.GraphitiExtracted = checkFileExists(graphitiComposeFile)

	if h.dockerClient != nil {
		graphitiExists, graphitiRunning := checkContainerExists(ctx, h.dockerClient, GraphitiContainerName)
		neo4jExists, neo4jRunning := checkContainerExists(ctx, h.dockerClient, Neo4jContainerName)

		c.GraphitiInstalled = graphitiExists && neo4jExists
		c.GraphitiRunning = graphitiRunning && neo4jRunning

		// check if graphiti-related volumes exist (indicates previous installation)
		graphitiVolumes := []string{"neo4j_data"}
		c.GraphitiVolumesExist = checkVolumesExist(ctx, h.dockerClient, graphitiVolumes)
	}

	return nil
}

func (h *defaultCheckHandler) GatherLangfuseInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	baseURL := getEnvVar(h.appState, "LANGFUSE_BASE_URL", "")
	projectID := getEnvVar(h.appState, "LANGFUSE_PROJECT_ID", "")
	publicKey := getEnvVar(h.appState, "LANGFUSE_PUBLIC_KEY", "")
	secretKey := getEnvVar(h.appState, "LANGFUSE_SECRET_KEY", "")

	c.LangfuseConnected = baseURL != "" && projectID != "" && publicKey != "" && secretKey != ""
	c.LangfuseExternal = baseURL != DefaultLangfuseEndpoint

	envDir := filepath.Dir(h.appState.GetEnvPath())
	langfuseFile := filepath.Join(envDir, LangfuseComposeFile)
	c.LangfuseExtracted = checkFileExists(langfuseFile)

	if h.dockerClient != nil {
		workerExists, workerRunning := checkContainerExists(ctx, h.dockerClient, LangfuseWorkerContainerName)
		webExists, webRunning := checkContainerExists(ctx, h.dockerClient, LangfuseWebContainerName)

		c.LangfuseInstalled = workerExists && webExists
		c.LangfuseRunning = workerRunning && webRunning

		// check if langfuse-related volumes exist (indicates previous installation)
		langfuseVolumes := []string{"langfuse-postgres-data", "langfuse-clickhouse-data", "langfuse-minio-data"}
		c.LangfuseVolumesExist = checkVolumesExist(ctx, h.dockerClient, langfuseVolumes)
	}

	return nil
}

func (h *defaultCheckHandler) GatherObservabilityInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	otelHost := getEnvVar(h.appState, "OTEL_HOST", "")
	c.ObservabilityConnected = otelHost != ""
	c.ObservabilityExternal = otelHost != DefaultObservabilityEndpoint

	envDir := filepath.Dir(h.appState.GetEnvPath())
	obsFile := filepath.Join(envDir, ObservabilityComposeFile)
	c.ObservabilityExtracted = checkFileExists(obsFile)

	if h.dockerClient != nil {
		exists, running := checkContainerExists(ctx, h.dockerClient, OpenTelemetryContainerName)
		c.ObservabilityInstalled = exists
		c.ObservabilityRunning = running
	}

	return nil
}

func (h *defaultCheckHandler) GatherSystemInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	// CPU check and count
	c.SysCPUCount = runtime.NumCPU()
	c.SysCPUOK = checkCPUResources()

	// memory check and calculations
	needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability := determineComponentNeeds(c)

	// calculate required memory using shared function
	c.SysMemoryRequired = calculateRequiredMemoryGB(needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability)

	// get available memory and check if sufficient
	c.SysMemoryAvailable = getAvailableMemoryGB()
	c.SysMemoryOK = checkMemoryResources(needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability)

	// disk check and calculations
	localComponents := countLocalComponentsToInstall(
		c.PentagiInstalled,
		c.GraphitiConnected, c.GraphitiExternal, c.GraphitiInstalled,
		c.LangfuseConnected, c.LangfuseExternal, c.LangfuseInstalled,
		c.ObservabilityConnected, c.ObservabilityExternal, c.ObservabilityInstalled,
	)

	// calculate required disk space using shared function
	c.SysDiskRequired = calculateRequiredDiskGB(c.WorkerImageExists, localComponents)

	// get available disk space and check if sufficient
	c.SysDiskAvailable = getAvailableDiskGB(ctx)
	c.SysDiskFreeSpaceOK = checkDiskSpaceWithContext(
		ctx,
		c.WorkerImageExists,
		c.PentagiInstalled,
		c.GraphitiConnected,
		c.GraphitiExternal,
		c.GraphitiInstalled,
		c.LangfuseConnected,
		c.LangfuseExternal,
		c.LangfuseInstalled,
		c.ObservabilityConnected,
		c.ObservabilityExternal,
		c.ObservabilityInstalled,
	)

	// network check with proxy and docker clients
	proxyURL := getProxyURL(h.appState)
	c.SysNetworkFailures = getNetworkFailures(ctx, proxyURL, h.dockerClient, h.workerClient)
	c.SysNetworkOK = len(c.SysNetworkFailures) == 0

	return nil
}

func (h *defaultCheckHandler) GatherUpdatesInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	proxyURL := getProxyURL(h.appState)
	updateServerURL := getEnvVar(h.appState, "UPDATE_SERVER_URL", DefaultUpdateServerEndpoint)

	request := CheckUpdatesRequest{
		InstallerOsType:        runtime.GOOS,
		InstallerVersion:       InstallerVersion,
		GraphitiConnected:      c.GraphitiConnected,
		GraphitiExternal:       c.GraphitiExternal,
		GraphitiInstalled:      c.GraphitiInstalled,
		LangfuseConnected:      c.LangfuseConnected,
		LangfuseExternal:       c.LangfuseExternal,
		LangfuseInstalled:      c.LangfuseInstalled,
		ObservabilityConnected: c.ObservabilityConnected,
		ObservabilityExternal:  c.ObservabilityExternal,
		ObservabilityInstalled: c.ObservabilityInstalled,
	}

	// get PentAGI container image info
	if h.dockerClient != nil && c.PentagiInstalled {
		if imageInfo := getContainerImageInfo(ctx, h.dockerClient, PentagiContainerName); imageInfo != nil {
			request.PentagiImageName = &imageInfo.Name
			request.PentagiImageTag = &imageInfo.Tag
			request.PentagiImageHash = &imageInfo.Hash
		}
	}

	// get Worker image info from environment
	if h.workerClient != nil {
		defaultImage := getEnvVar(h.appState, "DOCKER_DEFAULT_IMAGE_FOR_PENTEST", DefaultImageForPentest)
		if imageInfo := getImageInfo(ctx, h.workerClient, defaultImage); imageInfo != nil {
			request.WorkerImageName = &imageInfo.Name
			request.WorkerImageTag = &imageInfo.Tag
			request.WorkerImageHash = &imageInfo.Hash
		}
	}

	// get Graphiti image info if installed locally
	if h.dockerClient != nil && c.GraphitiConnected && !c.GraphitiExternal && c.GraphitiInstalled {
		if graphitiInfo := getContainerImageInfo(ctx, h.dockerClient, GraphitiContainerName); graphitiInfo != nil {
			request.GraphitiImageName = &graphitiInfo.Name
			request.GraphitiImageTag = &graphitiInfo.Tag
			request.GraphitiImageHash = &graphitiInfo.Hash
		}
		if neo4jInfo := getContainerImageInfo(ctx, h.dockerClient, Neo4jContainerName); neo4jInfo != nil {
			request.Neo4jImageName = &neo4jInfo.Name
			request.Neo4jImageTag = &neo4jInfo.Tag
			request.Neo4jImageHash = &neo4jInfo.Hash
		}
	}

	// get Langfuse image info if installed locally
	if h.dockerClient != nil && c.LangfuseConnected && !c.LangfuseExternal && c.LangfuseInstalled {
		if workerInfo := getContainerImageInfo(ctx, h.dockerClient, LangfuseWorkerContainerName); workerInfo != nil {
			request.LangfuseWorkerImageName = &workerInfo.Name
			request.LangfuseWorkerImageTag = &workerInfo.Tag
			request.LangfuseWorkerImageHash = &workerInfo.Hash
		}
		if webInfo := getContainerImageInfo(ctx, h.dockerClient, LangfuseWebContainerName); webInfo != nil {
			request.LangfuseWebImageName = &webInfo.Name
			request.LangfuseWebImageTag = &webInfo.Tag
			request.LangfuseWebImageHash = &webInfo.Hash
		}
	}

	// get Grafana and OpenTelemetry image info if observability installed locally
	if h.dockerClient != nil && c.ObservabilityConnected && !c.ObservabilityExternal && c.ObservabilityInstalled {
		if grafanaInfo := getContainerImageInfo(ctx, h.dockerClient, GrafanaContainerName); grafanaInfo != nil {
			request.GrafanaImageName = &grafanaInfo.Name
			request.GrafanaImageTag = &grafanaInfo.Tag
			request.GrafanaImageHash = &grafanaInfo.Hash
		}
		if otelInfo := getContainerImageInfo(ctx, h.dockerClient, OpenTelemetryContainerName); otelInfo != nil {
			request.OpenTelemetryImageName = &otelInfo.Name
			request.OpenTelemetryImageTag = &otelInfo.Tag
			request.OpenTelemetryImageHash = &otelInfo.Hash
		}
	}

	response := checkUpdatesServer(ctx, updateServerURL, proxyURL, request)
	if response != nil {
		c.UpdateServerAccessible = true
		c.InstallerIsUpToDate = response.InstallerIsUpToDate
		c.PentagiIsUpToDate = response.PentagiIsUpToDate
		c.GraphitiIsUpToDate = response.GraphitiIsUpToDate
		c.LangfuseIsUpToDate = response.LangfuseIsUpToDate
		c.ObservabilityIsUpToDate = response.ObservabilityIsUpToDate
		c.WorkerIsUpToDate = response.WorkerIsUpToDate
	} else {
		c.UpdateServerAccessible = false
		c.InstallerIsUpToDate = false
		c.PentagiIsUpToDate = false
		c.GraphitiIsUpToDate = false
		c.LangfuseIsUpToDate = false
		c.ObservabilityIsUpToDate = false
	}

	return nil
}

func Gather(ctx context.Context, appState state.State) (CheckResult, error) {
	if appState == nil {
		return CheckResult{}, ErrAppStateNotInitialized
	}

	c := CheckResult{
		// default to the built-in handler
		handler: &defaultCheckHandler{
			mx:       &sync.Mutex{},
			appState: appState,
		},
	}

	if err := c.GatherAllInfo(ctx); err != nil {
		return c, err
	}

	return c, nil
}

func GatherWithHandler(ctx context.Context, handler CheckHandler) (CheckResult, error) {
	if handler == nil {
		return CheckResult{}, ErrHandlerNotInitialized
	}

	c := CheckResult{
		handler: handler,
	}

	if err := handler.GatherAllInfo(ctx, &c); err != nil {
		return c, err
	}

	return c, nil
}
