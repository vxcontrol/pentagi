package processor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type dockerOperationsImpl struct {
	processor *processor
}

func newDockerOperations(p *processor) dockerOperations {
	return &dockerOperationsImpl{processor: p}
}

func (d *dockerOperationsImpl) pullWorkerImage(ctx context.Context, state *operationState) error {
	return d.pullImage(ctx, state, d.getWorkerImageName())
}

func (d *dockerOperationsImpl) pullDefaultImage(ctx context.Context, state *operationState) error {
	return d.pullImage(ctx, state, d.getDefaultImageName())
}

func (d *dockerOperationsImpl) pullImage(ctx context.Context, state *operationState, name string) error {
	d.processor.appendLog(fmt.Sprintf(MsgPullingImage, name), ProductStackWorker, state)

	cmd := exec.CommandContext(ctx, "docker", "pull", name)
	cmd.Env = d.getWorkerDockerEnv()

	if err := d.processor.runCommand(cmd, ProductStackWorker, state); err != nil {
		d.processor.appendLog(fmt.Sprintf(MsgImagePullFailed, name, err), ProductStackWorker, state)
		return fmt.Errorf("failed to pull image %s: %w", name, err)
	}

	d.processor.appendLog(fmt.Sprintf(MsgImagePullCompleted, name), ProductStackWorker, state)
	return nil
}

func (d *dockerOperationsImpl) removeWorkerContainers(ctx context.Context, state *operationState) error {
	cli, err := d.createWorkerDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	d.processor.appendLog(MsgRemovingWorkerContainers, ProductStackWorker, state)

	allContainers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list worker containers: %w", err)
	}

	var containers []container.Summary
	for _, c := range allContainers {
		for _, name := range c.Names {
			if strings.HasPrefix(name, "pentagi-") {
				containers = append(containers, c)
				break
			}
		}
	}

	if len(containers) == 0 {
		d.processor.appendLog(MsgNoWorkerContainersFound, ProductStackWorker, state)
		return nil
	}

	totalContainers := len(containers)
	for _, cont := range containers {
		if cont.State == "running" {
			d.processor.appendLog(fmt.Sprintf(MsgStoppingContainer, cont.ID[:12]), ProductStackWorker, state)
			if err := cli.ContainerStop(ctx, cont.ID, container.StopOptions{}); err != nil {
				return fmt.Errorf("failed to stop worker container %s: %w", cont.ID, err)
			}
		}

		d.processor.appendLog(fmt.Sprintf(MsgRemovingContainer, cont.ID[:12]), ProductStackWorker, state)
		if err := cli.ContainerRemove(ctx, cont.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			return fmt.Errorf("failed to remove worker container %s: %w", cont.ID, err)
		}

		d.processor.appendLog(fmt.Sprintf(MsgContainerRemoved, cont.ID[:12]), ProductStackWorker, state)
	}

	d.processor.appendLog(fmt.Sprintf(MsgWorkerContainersRemoved, totalContainers), ProductStackWorker, state)
	return nil
}

func (d *dockerOperationsImpl) removeWorkerImages(ctx context.Context, state *operationState) error {
	return d.removeImages(ctx, state, image.RemoveOptions{
		Force:         false,
		PruneChildren: false,
	})
}

func (d *dockerOperationsImpl) purgeWorkerImages(ctx context.Context, state *operationState) error {
	return d.removeImages(ctx, state, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
}

func (d *dockerOperationsImpl) removeImages(
	ctx context.Context, state *operationState, options image.RemoveOptions,
) error {
	if err := d.removeWorkerContainers(ctx, state); err != nil {
		return err
	}

	cli, err := d.createWorkerDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	for _, imageName := range []string{d.getWorkerImageName(), d.getDefaultImageName()} {
		d.processor.appendLog(fmt.Sprintf(MsgRemovingImage, imageName), ProductStackWorker, state)

		if _, err := cli.ImageRemove(ctx, imageName, options); err != nil {
			if !cerrdefs.IsNotFound(err) {
				return fmt.Errorf("failed to remove image %s: %w", imageName, err)
			} else {
				d.processor.appendLog(fmt.Sprintf(MsgImageNotFound, imageName), ProductStackWorker, state)
			}
		} else {
			d.processor.appendLog(fmt.Sprintf(MsgImageRemoved, imageName), ProductStackWorker, state)
		}
	}

	d.processor.appendLog(MsgWorkerImagesRemoveCompleted, ProductStackWorker, state)
	return nil
}

// createMainDockerClient creates docker client for the main stack (non-worker) using current process env
func (d *dockerOperationsImpl) createMainDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
}

// checkMainDockerNetwork returns true if a docker network with given name exists
func (d *dockerOperationsImpl) checkMainDockerNetwork(ctx context.Context, cli *client.Client, name string) (bool, error) {
	nets, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list docker networks: %w", err)
	}
	for _, n := range nets {
		if n.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// createMainDockerNetwork creates a docker network with given name if it does not exist
func (d *dockerOperationsImpl) createMainDockerNetwork(ctx context.Context, cli *client.Client, state *operationState, name string) error {
	exists, err := d.checkMainDockerNetwork(ctx, cli, name)
	if err != nil {
		return err
	}
	if exists {
		// inspect to validate labels
		nw, err := cli.NetworkInspect(ctx, name, network.InspectOptions{})
		if err == nil {
			wantProject := ""
			if envPath := d.processor.state.GetEnvPath(); envPath != "" {
				wantProject = filepath.Base(filepath.Dir(envPath))
			}
			hasComposeLabel := nw.Labels["com.docker.compose.network"] == name
			projectMatches := wantProject == "" || nw.Labels["com.docker.compose.project"] == wantProject
			if hasComposeLabel && projectMatches {
				d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkExists, name), ProductStackInstaller, state)
				return nil
			}
			// if labels incorrect and network has no containers attached, recreate with correct labels
			if len(nw.Containers) > 0 {
				d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkInUse, name), ProductStackInstaller, state)
				return nil
			}
			d.processor.appendLog(fmt.Sprintf(MsgRecreatingDockerNetwork, name), ProductStackInstaller, state)
			if err := cli.NetworkRemove(ctx, nw.ID); err != nil {
				d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkRemoveFailed, name, err), ProductStackInstaller, state)
				return fmt.Errorf("failed to remove network %s: %w", name, err)
			}
			d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkRemoved, name), ProductStackInstaller, state)
		}
	}
	d.processor.appendLog(fmt.Sprintf(MsgCreatingDockerNetwork, name), ProductStackInstaller, state)
	// mimic docker compose-created network by setting compose labels
	// project name: derived from working directory of env file (same as compose default)
	projectName := ""
	if envPath := d.processor.state.GetEnvPath(); envPath != "" {
		projectName = filepath.Base(filepath.Dir(envPath))
	}
	labels := map[string]string{
		"com.docker.compose.network": name,
	}
	if projectName != "" {
		labels["com.docker.compose.project"] = projectName
	}
	// driver: bridge (compose default for local networks)
	_, err = cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Labels: labels,
	})
	if err != nil {
		d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkCreateFailed, name, err), ProductStackInstaller, state)
		return fmt.Errorf("failed to create network %s: %w", name, err)
	}
	d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkCreated, name), ProductStackInstaller, state)
	return nil
}

// ensureMainDockerNetworks ensures all required networks for main stacks exist
func (d *dockerOperationsImpl) ensureMainDockerNetworks(ctx context.Context, state *operationState) error {
	d.processor.appendLog(MsgEnsuringDockerNetworks, ProductStackInstaller, state)
	defer d.processor.appendLog("", ProductStackInstaller, state)

	if !d.processor.checker.DockerApiAccessible {
		return fmt.Errorf("docker api is not accessible")
	}

	cli, err := d.createMainDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	required := []string{
		string(ProductDockerNetworkPentagi),
		string(ProductDockerNetworkObservability),
		string(ProductDockerNetworkLangfuse),
	}

	for _, net := range required {
		if err := d.createMainDockerNetwork(ctx, cli, state, net); err != nil {
			return err
		}
	}
	return nil
}

// removeMainDockerNetwork removes a docker network by name, detaching containers if possible
func (d *dockerOperationsImpl) removeMainDockerNetwork(ctx context.Context, state *operationState, name string) error {
	if !d.processor.checker.DockerApiAccessible {
		return fmt.Errorf("docker api is not accessible")
	}

	cli, err := d.createMainDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	// try inspect; if not found just return
	nw, err := cli.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		return nil
	}

	// attempt to disconnect all containers first (best effort)
	for id := range nw.Containers {
		_ = cli.NetworkDisconnect(ctx, nw.ID, id, true)
	}

	if err := cli.NetworkRemove(ctx, nw.ID); err != nil {
		return err
	}
	d.processor.appendLog(fmt.Sprintf(MsgDockerNetworkRemoved, name), ProductStackInstaller, state)
	return nil
}

// removeMainImages removes a list of images from main docker daemon
func (d *dockerOperationsImpl) removeMainImages(ctx context.Context, state *operationState, images []string) error {
	if !d.processor.checker.DockerApiAccessible {
		return fmt.Errorf("docker api is not accessible")
	}

	cli, err := d.createMainDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	opts := image.RemoveOptions{Force: state.force, PruneChildren: state.force}
	for _, img := range images {
		if img == "" {
			continue
		}
		d.processor.appendLog(fmt.Sprintf(MsgRemovingImage, img), ProductStackInstaller, state)
		if _, err := cli.ImageRemove(ctx, img, opts); err != nil {
			if !cerrdefs.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

// removeWorkerVolumes removes worker volumes (pentagi-terminal-*-data) in worker environment
func (d *dockerOperationsImpl) removeWorkerVolumes(ctx context.Context, state *operationState) error {
	cli, err := d.createWorkerDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	vols, err := cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return err
	}
	for _, v := range vols.Volumes {
		if strings.HasPrefix(v.Name, "pentagi-terminal-") && strings.HasSuffix(v.Name, "-data") {
			_ = cli.VolumeRemove(ctx, v.Name, true)
		}
	}
	return nil
}

func (d *dockerOperationsImpl) createWorkerDockerClient() (*client.Client, error) {
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	envVar, exists := d.processor.state.GetVar(client.EnvOverrideHost)
	if exists && (envVar.Value != "" || envVar.IsChanged) {
		opts = append(opts, client.WithHost(envVar.Value))
	} else if envVar.Default != "" {
		opts = append(opts, client.WithHost(envVar.Default))
	} else if envVar := os.Getenv(client.EnvOverrideHost); envVar != "" {
		opts = append(opts, client.WithHost(envVar))
	} else {
		opts = append(opts, client.WithHostFromEnv())
	}

	type tlsConfig struct {
		certPath string
		keyPath  string
		caPath   string
	}
	getTLSConfig := func(path string) tlsConfig {
		return tlsConfig{
			certPath: filepath.Join(path, "cert.pem"),
			keyPath:  filepath.Join(path, "key.pem"),
			caPath:   filepath.Join(path, "ca.pem"),
		}
	}

	envVar, exists = d.processor.state.GetVar(client.EnvOverrideCertPath)
	if exists && (envVar.Value != "" || envVar.IsChanged) {
		cfg := getTLSConfig(envVar.Value)
		opts = append(opts, client.WithTLSClientConfig(cfg.certPath, cfg.keyPath, cfg.caPath))
	} else if envVar.Default != "" {
		cfg := getTLSConfig(envVar.Default)
		opts = append(opts, client.WithTLSClientConfig(cfg.certPath, cfg.keyPath, cfg.caPath))
	} else {
		opts = append(opts, client.WithTLSClientConfigFromEnv())
	}

	return client.NewClientWithOpts(opts...)
}

func (d *dockerOperationsImpl) getWorkerDockerEnv() []string {
	var env []string

	envVar, exists := d.processor.state.GetVar(client.EnvOverrideHost)
	if exists && (envVar.Value != "" || envVar.IsChanged) {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvOverrideHost, envVar.Value))
	} else if envVar.Default != "" {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvOverrideHost, envVar.Default))
	} else if envVar := os.Getenv(client.EnvOverrideHost); envVar != "" {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvOverrideHost, envVar))
	}

	envVar, exists = d.processor.state.GetVar(client.EnvOverrideCertPath)
	if exists && (envVar.Value != "" || envVar.IsChanged) {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvOverrideCertPath, envVar.Value))
	} else if envVar.Default != "" {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvOverrideCertPath, envVar.Default))
	} else if envVar := os.Getenv(client.EnvOverrideCertPath); envVar != "" {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvOverrideCertPath, envVar))
	}

	envVar, exists = d.processor.state.GetVar(client.EnvTLSVerify)
	if exists && (envVar.Value != "" || envVar.IsChanged) {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvTLSVerify, envVar.Value))
	} else if envVar.Default != "" {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvTLSVerify, envVar.Default))
	} else if envVar := os.Getenv(client.EnvTLSVerify); envVar != "" {
		env = append(env, fmt.Sprintf("%s=%s", client.EnvTLSVerify, envVar))
	}

	return env
}

func (d *dockerOperationsImpl) getWorkerImageName() string {
	envVar, exists := d.processor.state.GetVar("DOCKER_DEFAULT_IMAGE_FOR_PENTEST")
	if exists && envVar.Value != "" {
		return envVar.Value
	}
	if envVar.Default != "" {
		return envVar.Default
	}
	return "vxcontrol/kali-linux:latest"
}

func (d *dockerOperationsImpl) getDefaultImageName() string {
	envVar, exists := d.processor.state.GetVar("DOCKER_DEFAULT_IMAGE")
	if exists && envVar.Value != "" {
		return envVar.Value
	}
	if envVar.Default != "" {
		return envVar.Default
	}
	return "debian:latest"
}
