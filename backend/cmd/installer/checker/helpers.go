package checker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"pentagi/cmd/installer/state"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type DockerVersion struct {
	Version string
	Valid   bool
}

type ImageInfo struct {
	Name string
	Tag  string
	Hash string
}

type CheckUpdatesRequest struct {
	InstallerOsType         string  `json:"installer_os_type"`
	InstallerVersion        string  `json:"installer_version"`
	PentagiImageName        *string `json:"pentagi_image_name,omitempty"`
	PentagiImageTag         *string `json:"pentagi_image_tag,omitempty"`
	PentagiImageHash        *string `json:"pentagi_image_hash,omitempty"`
	WorkerImageName         *string `json:"worker_image_name,omitempty"`
	WorkerImageTag          *string `json:"worker_image_tag,omitempty"`
	WorkerImageHash         *string `json:"worker_image_hash,omitempty"`
	GraphitiConnected       bool    `json:"graphiti_connected"`
	GraphitiInstalled       bool    `json:"graphiti_installed"`
	GraphitiExternal        bool    `json:"graphiti_external"`
	GraphitiImageName       *string `json:"graphiti_image_name,omitempty"`
	GraphitiImageTag        *string `json:"graphiti_image_tag,omitempty"`
	GraphitiImageHash       *string `json:"graphiti_image_hash,omitempty"`
	Neo4jImageName          *string `json:"neo4j_image_name,omitempty"`
	Neo4jImageTag           *string `json:"neo4j_image_tag,omitempty"`
	Neo4jImageHash          *string `json:"neo4j_image_hash,omitempty"`
	LangfuseConnected       bool    `json:"langfuse_connected"`
	LangfuseInstalled       bool    `json:"langfuse_installed"`
	LangfuseExternal        bool    `json:"langfuse_external"`
	ObservabilityConnected  bool    `json:"observability_connected"`
	ObservabilityExternal   bool    `json:"observability_external"`
	ObservabilityInstalled  bool    `json:"observability_installed"`
	LangfuseWorkerImageName *string `json:"langfuse_worker_image_name,omitempty"`
	LangfuseWorkerImageTag  *string `json:"langfuse_worker_image_tag,omitempty"`
	LangfuseWorkerImageHash *string `json:"langfuse_worker_image_hash,omitempty"`
	LangfuseWebImageName    *string `json:"langfuse_web_image_name,omitempty"`
	LangfuseWebImageTag     *string `json:"langfuse_web_image_tag,omitempty"`
	LangfuseWebImageHash    *string `json:"langfuse_web_image_hash,omitempty"`
	GrafanaImageName        *string `json:"grafana_image_name,omitempty"`
	GrafanaImageTag         *string `json:"grafana_image_tag,omitempty"`
	GrafanaImageHash        *string `json:"grafana_image_hash,omitempty"`
	OpenTelemetryImageName  *string `json:"otel_image_name,omitempty"`
	OpenTelemetryImageTag   *string `json:"otel_image_tag,omitempty"`
	OpenTelemetryImageHash  *string `json:"otel_image_hash,omitempty"`
}

type CheckUpdatesResponse struct {
	InstallerIsUpToDate     bool `json:"installer_is_up_to_date"`
	PentagiIsUpToDate       bool `json:"pentagi_is_up_to_date"`
	GraphitiIsUpToDate      bool `json:"graphiti_is_up_to_date"`
	LangfuseIsUpToDate      bool `json:"langfuse_is_up_to_date"`
	ObservabilityIsUpToDate bool `json:"observability_is_up_to_date"`
	WorkerIsUpToDate        bool `json:"worker_is_up_to_date"`
}

func checkFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func checkFileIsReadable(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

// checkDirIsWritable checks if we can write to a directory
func checkDirIsWritable(dirPath string) bool {
	// try to create a temporary file in the directory
	tempFile, err := os.CreateTemp(dirPath, ".pentagi_test_*")
	if err != nil {
		return false
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	// clean up the test file
	os.Remove(tempPath)
	return true
}

func getEnvVar(appState state.State, key, defaultValue string) string {
	if appState == nil {
		return defaultValue
	}

	if envVar, exist := appState.GetVar(key); exist && envVar.Value != "" {
		return envVar.Value
	} else if envVar.Default != "" {
		return envVar.Default
	}

	return defaultValue
}

// getProxyURL retrieves the proxy URL from application state if configured
func getProxyURL(appState state.State) string {
	if appState == nil {
		return ""
	}
	return getEnvVar(appState, "PROXY_URL", "")
}

func createDockerClient(host, certPath string, tlsVerify bool) (*client.Client, error) {
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	if host != "" {
		opts = append(opts, client.WithHost(host))
	}

	if tlsVerify && certPath != "" {
		opts = append(opts, client.WithTLSClientConfig(
			filepath.Join(certPath, "ca.pem"),
			filepath.Join(certPath, "cert.pem"),
			filepath.Join(certPath, "key.pem"),
		))
	}

	return client.NewClientWithOpts(opts...)
}

// createDockerClientFromEnv creates a docker client and returns the error type
func createDockerClientFromEnv(ctx context.Context) (*client.Client, DockerErrorType) {
	// first check if docker command exists
	_, err := exec.LookPath("docker")
	if err != nil {
		return nil, DockerErrorNotInstalled
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, DockerErrorAPIError
	}

	// try to ping the daemon
	_, err = cli.Ping(ctx)
	if err != nil {
		cli.Close() // close client on error
		// check if it's a connection error (daemon not running)
		if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") ||
			strings.Contains(err.Error(), "Is the docker daemon running") ||
			strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "no such host") ||
			strings.Contains(err.Error(), "dial unix") {
			return nil, DockerErrorNotRunning
		}
		// check for permission errors
		if strings.Contains(err.Error(), "permission denied") ||
			strings.Contains(err.Error(), "Got permission denied") {
			return nil, DockerErrorPermission
		}
		// other API errors
		return nil, DockerErrorAPIError
	}

	return cli, DockerErrorNone
}

type DockerErrorType string

// DockerErrorType constants
const (
	DockerErrorNone         DockerErrorType = ""
	DockerErrorNotInstalled DockerErrorType = "not_installed"
	DockerErrorNotRunning   DockerErrorType = "not_running"
	DockerErrorAPIError     DockerErrorType = "api_error"
	DockerErrorPermission   DockerErrorType = "permission"
)

func checkDockerVersion(ctx context.Context, cli *client.Client) DockerVersion {
	version, err := cli.ServerVersion(ctx)
	if err != nil {
		return DockerVersion{Version: "", Valid: false}
	}

	versionStr := version.Version
	valid := checkVersionCompatibility(versionStr, "20.0.0")

	return DockerVersion{Version: versionStr, Valid: valid}
}

func checkDockerCliVersion() DockerVersion {
	_, err := exec.LookPath("docker")
	if err != nil {
		return DockerVersion{Version: "", Valid: false}
	}

	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	output, err := cmd.Output()
	if err != nil && len(output) == 0 {
		return DockerVersion{Version: "", Valid: false}
	}

	versionStr := extractVersionFromOutput(string(output))
	valid := checkVersionCompatibility(versionStr, "20.0.0")

	return DockerVersion{Version: versionStr, Valid: valid}
}

func checkDockerComposeVersion() DockerVersion {
	cmd := exec.Command("docker", "compose", "version")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("docker-compose", "--version")
		output, err = cmd.Output()
		if err != nil {
			return DockerVersion{Version: "", Valid: false}
		}
	}

	versionStr := extractVersionFromOutput(string(output))
	valid := checkVersionCompatibility(versionStr, "1.25.0")

	return DockerVersion{Version: versionStr, Valid: valid}
}

func extractVersionFromOutput(output string) string {
	re := regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func checkVersionCompatibility(version, minVersion string) bool {
	if version == "" || minVersion == "" {
		return false
	}

	versionParts := strings.Split(version, ".")
	minVersionParts := strings.Split(minVersion, ".")

	for i := 0; i < len(versionParts) && i < len(minVersionParts); i++ {
		v, err1 := strconv.Atoi(versionParts[i])
		minV, err2 := strconv.Atoi(minVersionParts[i])

		if err1 != nil || err2 != nil {
			return false
		}

		if v > minV {
			return true
		}
		if v < minV {
			return false
		}
	}

	return len(versionParts) >= len(minVersionParts)
}

func checkContainerExists(ctx context.Context, cli *client.Client, name string) (exists, running bool) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, false
	}

	for _, cont := range containers {
		for _, containerName := range cont.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return true, cont.State == "running"
			}
		}
	}

	return false, false
}

// checkVolumesExist checks if any of the specified volumes exist
// it matches both exact names and volumes with compose project prefix (e.g., "pentagi_pentagi-data")
func checkVolumesExist(ctx context.Context, cli *client.Client, volumeNames []string) bool {
	if cli == nil || len(volumeNames) == 0 {
		return false
	}

	volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return false
	}

	// collect all volume names from Docker
	existingVolumes := make([]string, 0, len(volumes.Volumes))
	for _, vol := range volumes.Volumes {
		existingVolumes = append(existingVolumes, vol.Name)
	}

	// check if any of the requested volumes exist
	// matches both exact names and volumes with compose prefix (project_volume-name)
	for _, volumeName := range volumeNames {
		for _, existingVolume := range existingVolumes {
			// exact match or suffix match with underscore separator
			if existingVolume == volumeName || strings.HasSuffix(existingVolume, "_"+volumeName) {
				return true
			}
		}
	}

	return false
}

func checkCPUResources() bool {
	return runtime.NumCPU() >= 2
}

// determineComponentNeeds checks which components need to be started based on their status
func determineComponentNeeds(c *CheckResult) (needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability bool) {
	needsForPentagi = !c.PentagiRunning
	needsForGraphiti = c.GraphitiConnected && !c.GraphitiExternal && !c.GraphitiRunning
	needsForLangfuse = c.LangfuseConnected && !c.LangfuseExternal && !c.LangfuseRunning
	needsForObservability = c.ObservabilityConnected && !c.ObservabilityExternal && !c.ObservabilityRunning
	return
}

// calculateRequiredMemoryGB calculates the total memory required based on which components need to be started
func calculateRequiredMemoryGB(needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability bool) float64 {
	requiredGB := MinFreeMemGB
	if needsForPentagi {
		requiredGB += MinFreeMemGBForPentagi
	}
	if needsForGraphiti {
		requiredGB += MinFreeMemGBForGraphiti
	}
	if needsForLangfuse {
		requiredGB += MinFreeMemGBForLangfuse
	}
	if needsForObservability {
		requiredGB += MinFreeMemGBForObservability
	}
	return requiredGB
}

func checkMemoryResources(needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability bool) bool {
	if !needsForPentagi && !needsForGraphiti && !needsForLangfuse && !needsForObservability {
		return true
	}

	requiredGB := calculateRequiredMemoryGB(needsForPentagi, needsForGraphiti, needsForLangfuse, needsForObservability)

	// check available memory using different methods depending on OS
	switch runtime.GOOS {
	case "linux":
		return checkLinuxMemory(requiredGB)
	case "darwin":
		return checkDarwinMemory(requiredGB)
	default:
		return true // assume OK for other systems
	}
}

// getAvailableMemoryGB returns the available memory in GB for the current OS
func getAvailableMemoryGB() float64 {
	switch runtime.GOOS {
	case "linux":
		return getLinuxAvailableMemoryGB()
	case "darwin":
		return getDarwinAvailableMemoryGB()
	default:
		return 0.0 // unknown for other systems
	}
}

// getLinuxAvailableMemoryGB reads available memory from /proc/meminfo on Linux
func getLinuxAvailableMemoryGB() float64 {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0.0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var memFree, memAvailable int64

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					memAvailable = val * 1024 // Convert KB to bytes
				}
			}
			break
		}
		if strings.HasPrefix(line, "MemFree:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					memFree = val * 1024 // Convert KB to bytes
				}
			}
		}
	}

	availableMemGB := float64(memAvailable) / (1024 * 1024 * 1024)
	if availableMemGB > 0 {
		return availableMemGB
	}

	return float64(memFree) / (1024 * 1024 * 1024)
}

// getDarwinAvailableMemoryGB parses vm_stat output to get available memory on macOS
func getDarwinAvailableMemoryGB() float64 {
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(output), "\n")
	var pageSize, freePages, inactivePages, purgeablePages int64 = 4096, 0, 0, 0 // default page size

	for _, line := range lines {
		if strings.Contains(line, "page size of") {
			re := regexp.MustCompile(`(\d+) bytes`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					pageSize = val
				}
			}
		}
		if strings.HasPrefix(line, "Pages free:") {
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					freePages = val
				}
			}
		}
		if strings.HasPrefix(line, "Pages inactive:") {
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					inactivePages = val
				}
			}
		}
		if strings.HasPrefix(line, "Pages purgeable:") {
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					purgeablePages = val
				}
			}
		}
	}

	// Available memory = free + inactive + purgeable (can be reclaimed)
	availablePages := freePages + inactivePages + purgeablePages
	return float64(availablePages*pageSize) / (1024 * 1024 * 1024)
}

func checkLinuxMemory(requiredGB float64) bool {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return true // assume OK if can't check
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var memFree, memAvailable int64

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					memAvailable = val * 1024 // Convert KB to bytes
				}
			}
			break
		}
		if strings.HasPrefix(line, "MemFree:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					memFree = val * 1024 // Convert KB to bytes
				}
			}
		}
	}

	availableMemGB := float64(memAvailable) / (1024 * 1024 * 1024)
	if availableMemGB > 0 {
		return availableMemGB >= requiredGB
	}

	freeMemGB := float64(memFree) / (1024 * 1024 * 1024)
	return freeMemGB >= requiredGB
}

func checkDarwinMemory(requiredGB float64) bool {
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return true // assume OK if can't check
	}

	lines := strings.Split(string(output), "\n")
	var pageSize, freePages, inactivePages, purgeablePages int64 = 4096, 0, 0, 0 // default page size

	for _, line := range lines {
		if strings.Contains(line, "page size of") {
			re := regexp.MustCompile(`(\d+) bytes`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					pageSize = val
				}
			}
		}
		if strings.HasPrefix(line, "Pages free:") {
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					freePages = val
				}
			}
		}
		if strings.HasPrefix(line, "Pages inactive:") {
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					inactivePages = val
				}
			}
		}
		if strings.HasPrefix(line, "Pages purgeable:") {
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					purgeablePages = val
				}
			}
		}
	}

	// Available memory = free + inactive + purgeable (can be reclaimed)
	availablePages := freePages + inactivePages + purgeablePages
	availableMemGB := float64(availablePages*pageSize) / (1024 * 1024 * 1024)
	return availableMemGB >= requiredGB
}

// calculateRequiredDiskGB calculates the disk space required based on worker images and local components
func calculateRequiredDiskGB(workerImageExists bool, localComponents int) float64 {
	// adjust required space based on components and worker images
	if !workerImageExists {
		// need to download worker images (can be large)
		return MinFreeDiskGBForWorkerImages
	} else if localComponents > 0 {
		// have local components that need space for containers/volumes
		return MinFreeDiskGBForComponents + float64(localComponents)*MinFreeDiskGBPerComponents
	}
	// default minimum disk space required
	return MinFreeDiskGB
}

// countLocalComponentsToInstall counts how many components need to be installed locally
func countLocalComponentsToInstall(
	pentagiInstalled,
	graphitiConnected, graphitiExternal, graphitiInstalled,
	langfuseConnected, langfuseExternal, langfuseInstalled,
	obsConnected, obsExternal, obsInstalled bool,
) int {
	localComponents := 0
	if !pentagiInstalled {
		localComponents++
	}
	if graphitiConnected && !graphitiExternal && !graphitiInstalled {
		localComponents++
	}
	if langfuseConnected && !langfuseExternal && !langfuseInstalled {
		localComponents++
	}
	if obsConnected && !obsExternal && !obsInstalled {
		localComponents++
	}
	return localComponents
}

func checkDiskSpaceWithContext(
	ctx context.Context,
	workerImageExists, pentagiInstalled,
	graphitiConnected, graphitiExternal, graphitiInstalled,
	langfuseConnected, langfuseExternal, langfuseInstalled,
	obsConnected, obsExternal, obsInstalled bool,
) bool {
	// determine required disk space based on what needs to be installed locally
	localComponents := countLocalComponentsToInstall(
		pentagiInstalled,
		graphitiConnected, graphitiExternal, graphitiInstalled,
		langfuseConnected, langfuseExternal, langfuseInstalled,
		obsConnected, obsExternal, obsInstalled,
	)

	requiredGB := calculateRequiredDiskGB(workerImageExists, localComponents)

	// check disk space using different methods depending on OS
	switch runtime.GOOS {
	case "linux":
		return checkLinuxDiskSpace(ctx, requiredGB)
	case "darwin":
		return checkDarwinDiskSpace(ctx, requiredGB)
	default:
		return true // assume OK for other systems
	}
}

// getAvailableDiskGB returns the available disk space in GB for the current OS
func getAvailableDiskGB(ctx context.Context) float64 {
	switch runtime.GOOS {
	case "linux":
		return getLinuxAvailableDiskGB(ctx)
	case "darwin":
		return getDarwinAvailableDiskGB(ctx)
	default:
		return 0.0 // unknown for other systems
	}
}

// getLinuxAvailableDiskGB uses df command to get available disk space on Linux
func getLinuxAvailableDiskGB(ctx context.Context) float64 {
	cmd := exec.CommandContext(ctx, "df", "-BG", ".")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0.0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0.0
	}

	availableStr := strings.TrimSuffix(fields[3], "G")
	if available, err := strconv.ParseFloat(availableStr, 64); err == nil {
		return available
	}

	return 0.0
}

// getDarwinAvailableDiskGB uses df command to get available disk space on macOS
func getDarwinAvailableDiskGB(ctx context.Context) float64 {
	cmd := exec.CommandContext(ctx, "df", "-g", ".")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0.0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0.0
	}

	if available, err := strconv.ParseFloat(fields[3], 64); err == nil {
		return available
	}

	return 0.0
}

func checkLinuxDiskSpace(ctx context.Context, requiredGB float64) bool {
	cmd := exec.CommandContext(ctx, "df", "-BG", ".")
	output, err := cmd.Output()
	if err != nil {
		return true // assume OK if can't check
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return true
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return true
	}

	availableStr := strings.TrimSuffix(fields[3], "G")
	if available, err := strconv.ParseFloat(availableStr, 64); err == nil {
		return available >= requiredGB
	}

	return true
}

func checkDarwinDiskSpace(ctx context.Context, requiredGB float64) bool {
	cmd := exec.CommandContext(ctx, "df", "-g", ".")
	output, err := cmd.Output()
	if err != nil {
		return true // assume OK if can't check
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return true
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return true
	}

	if available, err := strconv.ParseFloat(fields[3], 64); err == nil {
		return available >= requiredGB
	}

	return true
}

func getNetworkFailures(ctx context.Context, proxyURL string, dockerClient, workerClient *client.Client) []string {
	var failures []string

	// 1. DNS resolution test
	if !checkDNSResolution("docker.io") {
		// Using hardcoded string here to avoid circular dependency with locale package
		failures = append(failures, "• DNS resolution failed for docker.io")
	}

	// 2. HTTP connectivity test
	if !checkHTTPConnectivity(ctx, proxyURL) {
		// Using hardcoded string here to avoid circular dependency with locale package
		failures = append(failures, "• Cannot reach external services via HTTPS")
	}

	// 3. Docker pull test (only if both clients are available)
	if dockerClient != nil && workerClient != nil && !checkDockerPullConnectivity(ctx, dockerClient, workerClient) {
		// Using hardcoded string here to avoid circular dependency with locale package
		failures = append(failures, "• Cannot pull Docker images from registry")
	}

	return failures
}

func getContainerImageInfo(ctx context.Context, cli *client.Client, containerName string) *ImageInfo {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil
	}

	for _, cont := range containers {
		for _, name := range cont.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				return parseImageRef(cont.Image, cont.ImageID)
			}
		}
	}

	return nil
}

func checkImageExists(ctx context.Context, cli *client.Client, imageName string) bool {
	imageInfo := getImageInfo(ctx, cli, imageName)
	return imageInfo != nil && imageInfo.Hash != ""
}

func getImageInfo(ctx context.Context, cli *client.Client, imageName string) *ImageInfo {
	if cli == nil {
		return nil
	}

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil
	}

	imageInfo := parseImageRef(imageName, "")
	if imageInfo == nil {
		return nil
	}

	fullImageName := imageInfo.Name + ":" + imageInfo.Tag
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName || tag == fullImageName {
				imageInfo.Hash = img.ID
				break
			}
		}
	}

	return imageInfo
}

func parseImageRef(imageRef, imageID string) *ImageInfo {
	if imageRef == "" {
		return nil
	}

	info := &ImageInfo{
		Hash: imageID,
	}

	// normalize the image reference
	originalRef := imageRef

	// parse hash if present (image@sha256:...)
	if strings.Contains(imageRef, "@") {
		parts := strings.SplitN(imageRef, "@", 2)
		imageRef = parts[0]
		if len(parts) > 1 {
			info.Hash = parts[1]
		}
	}

	// handle registry/namespace/repository parsing
	var name string
	if strings.Contains(imageRef, "/") {
		// has registry or namespace
		nameParts := strings.Split(imageRef, "/")
		if len(nameParts) >= 2 {
			// check if first part looks like a registry (contains . or :)
			if strings.Contains(nameParts[0], ".") || strings.Contains(nameParts[0], ":") {
				// first part is registry, skip it for name extraction
				if len(nameParts) > 2 {
					name = strings.Join(nameParts[1:], "/")
				} else {
					name = nameParts[1]
				}
			} else {
				// no registry, combine all parts as name
				name = imageRef
			}
		} else {
			name = imageRef
		}
	} else {
		name = imageRef
	}

	// parse name and tag
	if strings.Contains(name, ":") {
		parts := strings.SplitN(name, ":", 2)
		info.Name = parts[0]
		if len(parts) > 1 && parts[1] != "" {
			// validate that it's not a port number (for registry detection edge case)
			if !strings.Contains(parts[1], ".") {
				info.Tag = parts[1]
			} else {
				info.Name = name
				info.Tag = "latest"
			}
		}
	} else {
		info.Name = name
		info.Tag = "latest"
	}

	// if we still don't have a tag, default to latest
	if info.Tag == "" {
		info.Tag = "latest"
	}

	// validate that name is not empty
	if info.Name == "" {
		info.Name = originalRef
		info.Tag = "latest"
	}

	return info
}

func checkUpdatesServer(
	ctx context.Context,
	serverURL, proxyURL string,
	request CheckUpdatesRequest,
) *CheckUpdatesResponse {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if proxyURL != "" {
		if proxyURLParsed, err := url.Parse(proxyURL); err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
		}
	}

	fullURL := serverURL + UpdatesCheckEndpoint
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var response CheckUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil
	}

	return &response
}

func checkDNSResolution(hostname string) bool {
	_, err := net.LookupHost(hostname)
	return err == nil
}

func checkHTTPConnectivity(ctx context.Context, proxyURL string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if proxyURL != "" {
		if proxyURLParsed, err := url.Parse(proxyURL); err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://docker.io", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500 // accept any non-server-error response
}

func checkDockerPullConnectivity(ctx context.Context, dockerClient, workerClient *client.Client) bool {
	// test with main docker client
	if dockerClient != nil {
		if checkSingleDockerPull(ctx, dockerClient, DefaultImage) {
			return true
		}
	}

	// test with worker client
	if workerClient != nil {
		if checkSingleDockerPull(ctx, workerClient, DefaultImage) {
			return true
		}
	}

	return false
}

func checkSingleDockerPull(ctx context.Context, cli *client.Client, imageName string) bool {
	// try to pull the image
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return false
	}
	defer reader.Close()

	// read a bit from the response to ensure it's working
	buf := make([]byte, 1024)
	_, err = reader.Read(buf)
	return err == nil || errors.Is(err, io.EOF) || errors.Is(err, syscall.EIO)
}
