package docker

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pentagi/pkg/config"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

// daemonProbeTimeout bounds the startup check. Reaching the sandbox daemon is
// diagnostic only, so a slow or unreachable endpoint must delay the boot by a
// few seconds at most and never block it.
const daemonProbeTimeout = 5 * time.Second

// runsPentagi reports whether the daemon behind cli is the one running PentAGI
// itself, found the way the rest of this package finds its own container: a
// running container whose configured hostname matches this process's.
//
// A false answer is only trustworthy when err is nil. A daemon that cannot be
// listed says nothing about what it runs.
func runsPentagi(ctx context.Context, cli *client.Client) (bool, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return false, fmt.Errorf("failed to read hostname: %w", err)
	}

	list, err := cli.ContainerList(ctx, client.ContainerListOptions{
		Filters: make(client.Filters).Add("status", "running"),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, item := range list.Items {
		result, err := cli.ContainerInspect(ctx, item.ID, client.ContainerInspectOptions{})
		if cerrdefs.IsNotFound(err) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("failed to inspect container %s: %w", item.ID, err)
		}
		if result.Container.Config.Hostname == hostname {
			return true, nil
		}
	}

	return false, nil
}

func workerDaemonClient(cfg *config.Config) (*client.Client, error) {
	opts := []client.Opt{client.WithHost(cfg.DockerInsideHost)}
	if cfg.DockerInsideTLSVerify != "" && cfg.DockerInsideCertPath != "" {
		dir := strings.TrimRight(cfg.DockerInsideCertPath, "/")
		opts = append(opts, client.WithTLSClientConfig(
			dir+"/ca.pem", dir+"/cert.pem", dir+"/key.pem"))
	}

	return client.New(opts...)
}

// logWorkerDaemonIsolation reports, once at startup, how much of the host an
// agent reaches through the Docker access it is given.
//
// The question that decides it is whether PentAGI's own container is visible on
// a daemon: if it is, that daemon is the host daemon, and an agent holding it
// can start a privileged container and take the node -- PentAGI, the other
// flows and the certificates on it included. Separate orchestration and sandbox
// daemons are the shape examples/guides/worker_node.md describes, and the only
// one where an escape stays inside the sandbox tier.
func logWorkerDaemonIsolation(ctx context.Context, cfg *config.Config, cli *client.Client) {
	if cfg == nil || !cfg.DockerInside {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, daemonProbeTimeout)
	defer cancel()

	if socket, autodetect := cfg.WorkerDockerSocket(); socket != "" || autodetect {
		if cfg.DockerInsideHost != "" {
			logrus.WithField("docker_inside_host", cfg.DockerInsideHost).Warn("DOCKER_INSIDE=true: " +
				"DOCKER_SOCKET is set, so the host socket is mounted into every worker alongside " +
				"DOCKER_INSIDE_HOST; unset DOCKER_SOCKET to give sandboxes only the daemon of their own")
		}
		logSharedDaemon(ctx, cfg, cli)
		return
	}

	if workerCertsLiveElsewhere(cfg) {
		// The certificates are only unreadable from here, so the check has to
		// happen where they exist -- inside the worker, at creation.
		logrus.WithField("docker_inside_host", cfg.DockerInsideHost).Info("DOCKER_INSIDE=true: " +
			"sandboxes reach DOCKER_INSIDE_HOST with TLS client certificates that live on the worker " +
			"node, so this process cannot open that endpoint itself")
		return
	}

	workerCli, err := workerDaemonClient(cfg)
	if err != nil {
		logrus.WithError(err).Warn("DOCKER_INSIDE=true: could not open a client against " +
			"DOCKER_INSIDE_HOST to check what agents reach through it; verify that endpoint by hand")
		return
	}
	defer workerCli.Close()

	logSeparateDaemon(ctx, cfg, cli, workerCli)
}

func workerCertsLiveElsewhere(cfg *config.Config) bool {
	if cfg.DockerInsideTLSVerify == "" || cfg.DockerInsideCertPath == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(cfg.DockerInsideCertPath, "ca.pem"))
	return errors.Is(err, fs.ErrNotExist)
}

// logSharedDaemon covers DOCKER_SOCKET and the autodetected socket: agents are
// given the same daemon PentAGI orchestrates on.
func logSharedDaemon(ctx context.Context, cfg *config.Config, cli *client.Client) {
	logger := logrus.WithField("worker_docker", workerDockerDescription(cfg))

	isHost, err := runsPentagi(ctx, cli)
	switch {
	case err != nil:
		logger.WithError(err).Warn("DOCKER_INSIDE=true: agents share the daemon PentAGI " +
			"orchestrates on, and whether that daemon also runs PentAGI could not be determined; " +
			"check the daemon's hardening and whether an authorization plugin is installed")
	case isHost:
		logger.Warn("DOCKER_INSIDE=true: agents are given the host daemon, the one running " +
			"PentAGI itself. An agent that escapes its sandbox can start a privileged container " +
			"and take the node, PentAGI and the other flows' containers with it. Check the " +
			"daemon's hardening and whether an authorization plugin is installed, or move " +
			"sandboxes onto their own daemon with DOCKER_INSIDE_HOST " +
			"(see examples/guides/worker_node.md)")
	default:
		logger.Warn("DOCKER_INSIDE=true: agents and the orchestrator share one daemon, and PentAGI's " +
			"own container is not among the containers it runs. If PentAGI runs in a container " +
			"elsewhere, an escape does not reach this process; if it runs directly on this host, it " +
			"does. Check the daemon's hardening and whether an authorization plugin is installed")
	}
}

// logSeparateDaemon covers DOCKER_INSIDE_HOST: agents are pointed at an endpoint
// of their own, which is the arrangement worth confirming rather than assuming.
func logSeparateDaemon(ctx context.Context, cfg *config.Config, cli, workerCli *client.Client) {
	logger := logrus.WithField("docker_inside_host", cfg.DockerInsideHost)

	if sameDaemon(ctx, cli, workerCli) {
		logger.Warn("DOCKER_INSIDE=true: DOCKER_INSIDE_HOST names the daemon PentAGI orchestrates on, " +
			"so sandboxes do not get a daemon of their own")
		logSharedDaemon(ctx, cfg, cli)
		return
	}

	workerRunsPentagi, workerErr := runsPentagi(ctx, workerCli)
	if workerErr != nil {
		logger.WithError(workerErr).Warn("DOCKER_INSIDE=true: agents are pointed at a separate " +
			"daemon that could not be listed, so whether it also runs PentAGI is unknown; " +
			"verify it by hand")
		return
	}

	if workerRunsPentagi {
		logger.Warn("DOCKER_INSIDE=true: the daemon designated for sandboxes is the one running " +
			"PentAGI itself, so pointing agents at it grants what a separate daemon exists to " +
			"withhold. Point DOCKER_INSIDE_HOST at a daemon that does not run PentAGI " +
			"(see examples/guides/worker_node.md)")
		return
	}

	orchestratorIsHost, err := runsPentagi(ctx, cli)
	switch {
	case err != nil:
		logger.WithError(err).Info("DOCKER_INSIDE=true: sandboxes use a daemon of their own and " +
			"PentAGI does not run on it, so an escape stays inside the sandbox tier; whether PentAGI " +
			"runs on the daemon it orchestrates on could not be determined")
	case orchestratorIsHost:
		logger.Info("DOCKER_INSIDE=true: sandboxes use a daemon of their own and PentAGI does " +
			"not run on it, so an escape stays inside the sandbox tier")
	default:
		logger.Info("DOCKER_INSIDE=true: sandboxes use a daemon of their own, separate from the one " +
			"PentAGI orchestrates on, and PentAGI's own container is on neither")
	}
}

func workerDockerDescription(cfg *config.Config) string {
	if cfg.DockerSocket != "" {
		return cfg.DockerSocket
	}
	return "autodetected host socket"
}

func sameDaemon(ctx context.Context, a, b *client.Client) bool {
	first, err := a.Info(ctx, client.InfoOptions{})
	if err != nil {
		return false
	}
	second, err := b.Info(ctx, client.InfoOptions{})
	if err != nil {
		return false
	}
	return first.Info.ID != "" && first.Info.ID == second.Info.ID
}
