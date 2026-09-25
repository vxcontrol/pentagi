package docker

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"pentagi/pkg/config"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

//go:embed sandbox_probe.sh
var sandboxProbeScript string

const (
	// sandboxProbeTimeout covers the whole check. The pull inside it is the slow
	// part and only on the first run; everything else is a few seconds.
	sandboxProbeTimeout = 10 * time.Minute

	// maxSandboxReportBytes bounds the report. The daemon document it carries is
	// a few kilobytes; anything near this is not a report.
	maxSandboxReportBytes = 256 * 1024

	sandboxGuideURL = "https://raw.githubusercontent.com/vxcontrol/pentagi/refs/heads/main/examples/guides/worker_node.md"
)

// sandboxLine is one line of the probe's report. The probe emits JSON so this
// side never has to read prose: which checks ran and which failed is data.
type sandboxLine struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Group  string `json:"group"`
	Expect string `json:"expect"`
	Result string `json:"result"`
	Status int    `json:"status"`
	Reason string `json:"reason"`
	Total  int    `json:"total"`
	Passed int    `json:"passed"`
	Failed int    `json:"failed"`
	Body   struct {
		ID      string `json:"ID"`
		Plugins struct {
			Authorization []string `json:"Authorization"`
		} `json:"Plugins"`
	} `json:"body"`
}

// sandboxReport is what the probe found: which daemon answered, and how each
// check went.
type sandboxReport struct {
	daemonID string
	authz    []string
	total    int
	passed   int
	failed   []string
}

// parseSandboxReport reads the probe's output. A line that is not JSON is a
// defect in the probe rather than a verdict about the sandbox, so it fails the
// parse instead of being skipped: a silently dropped line could hide a failure.
func parseSandboxReport(out []byte) (sandboxReport, error) {
	var (
		report    sandboxReport
		sawInfo   bool
		sawTotals bool
	)

	for _, raw := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		var line sandboxLine
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			return report, fmt.Errorf("probe emitted a line that is not JSON: %w", err)
		}

		switch line.Kind {
		case "fatal":
			return report, fmt.Errorf("probe could not reach the sandbox daemon: %s", line.Reason)
		case "info":
			report.daemonID, report.authz, sawInfo = line.Body.ID, line.Body.Plugins.Authorization, true
		case "check":
			if line.Result != "pass" {
				report.failed = append(report.failed, line.Group+"/"+line.ID)
			}
		case "summary":
			report.total, report.passed, sawTotals = line.Total, line.Passed, true
		}
	}

	switch {
	case !sawInfo:
		return report, fmt.Errorf("probe did not report which daemon it reached")
	case !sawTotals:
		return report, fmt.Errorf("probe did not finish: no summary")
	case report.daemonID == "":
		return report, fmt.Errorf("the sandbox daemon reported no identity")
	}

	return report, nil
}

// probeSandbox starts a worker the way a flow starts one -- same spec, same
// Docker access, same network -- and runs the checks inside it. Only the image
// differs, and the container is removed whatever happens.
func (dc *dockerClient) probeSandbox(ctx context.Context, cfg *config.Config) (sandboxReport, error) {
	ctx, cancel := context.WithTimeout(ctx, sandboxProbeTimeout)
	defer cancel()

	image := cfg.DockerDefaultImageForTest
	if image == "" {
		return sandboxReport{}, fmt.Errorf("DOCKER_DEFAULT_IMAGE_FOR_TEST is empty")
	}

	// The work volume is the one thing a flow's worker has and this does not:
	// it holds that flow's files and has no bearing on what the daemon allows.
	containerConfig, hostConfig := WorkerSpec(cfg, image)
	containerConfig.Env = append(containerConfig.Env, "PENTAGI_TEST_IMAGE="+image)
	dc.applyWorkerDockerAccess(containerConfig, hostConfig)
	if hostConfig.PidsLimit == nil {
		pidsLimit := int64(2048)
		hostConfig.PidsLimit = &pidsLimit
	}

	var networkingConfig *network.NetworkingConfig
	switch dc.network {
	case "":
	case "host":
		// A mode the container is created in, as RunContainer creates a flow's
		// worker: joined as an endpoint instead, the daemon refuses to start it.
		hostConfig.NetworkMode = container.NetworkMode("host")
	default:
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{dc.network: {}},
		}
	}

	name := fmt.Sprintf("pentagi-sandbox-check-%d", time.Now().UnixNano())
	created, err := dc.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: containerConfig, HostConfig: hostConfig, NetworkingConfig: networkingConfig, Name: name,
	})
	if err != nil {
		return sandboxReport{}, fmt.Errorf("could not create the check container from %q: %w", image, err)
	}
	defer func() {
		removeCtx, cancelRemove := context.WithTimeout(context.WithoutCancel(ctx), containerDiscardTimeout)
		defer cancelRemove()
		if _, err := dc.client.ContainerRemove(removeCtx, created.ID,
			client.ContainerRemoveOptions{Force: true}); err != nil {
			logrus.WithContext(ctx).WithError(err).Warnf("failed to remove the check container %s", name)
		}
	}()

	if _, err := dc.client.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
		return sandboxReport{}, fmt.Errorf("could not start the check container from %q: %w", image, err)
	}

	out, err := dc.runSandboxProbe(ctx, created.ID)
	if err != nil {
		return sandboxReport{}, err
	}

	return parseSandboxReport(out)
}

func (dc *dockerClient) runSandboxProbe(ctx context.Context, containerID string) ([]byte, error) {
	created, err := dc.client.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          []string{"sh", "-c", sandboxProbeScript},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not start the probe: %w", err)
	}

	attached, err := dc.client.ExecAttach(ctx, created.ID, client.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not attach to the probe: %w", err)
	}
	raw, readErr := io.ReadAll(io.LimitReader(attached.Reader, maxSandboxReportBytes+1))
	attached.Close()
	if readErr != nil {
		return nil, fmt.Errorf("could not read the probe's report: %w", readErr)
	}
	if int64(len(raw)) > maxSandboxReportBytes {
		return nil, fmt.Errorf("the probe's report exceeded %d bytes", maxSandboxReportBytes)
	}

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, bytes.NewReader(raw)); err != nil {
		return nil, fmt.Errorf("could not decode the probe's report: %w", err)
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("the probe produced no report: %s",
			strings.TrimSpace(truncate(stderr.String(), 300)))
	}

	return stdout.Bytes(), nil
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	return value[:limit]
}

// decideSandbox settles, once at startup, whether agents get Docker at all.
//
// It never refuses to start: an arrangement that cannot be verified turns the
// sandbox OFF and says why, because a deployment that keeps running without
// agent Docker access is recoverable and one that will not boot is not.
func (dc *dockerClient) decideSandbox(ctx context.Context, cfg *config.Config) {
	if cfg == nil || !cfg.DockerInside {
		return
	}

	logger := logrus.WithContext(ctx).WithField("docker_inside_host", cfg.DockerInsideHost)

	if cfg.DockerInsideHost == "" {
		dc.disableSandbox(cfg, logger,
			"DOCKER_INSIDE_HOST is empty, so a worker would be given whichever daemon this process "+
				"could autodetect -- possibly the one running PentAGI itself")

		return
	}

	if !cfg.DockerInsidePolicyTests {
		logger.Warn("sandbox is ENABLED but its isolation was NOT tested: agents get Docker access " +
			"nobody has measured. Set DOCKER_INSIDE_POLICY_TESTS=true to have PentAGI check at " +
			"startup that the sandbox daemon is separate from its own and refuses host-escape " +
			"requests. Setup guide: " + sandboxGuideURL)

		return
	}

	probe := dc.probeSandbox
	if dc.sandboxProbeRunner != nil {
		probe = dc.sandboxProbeRunner
	}

	report, err := probe(ctx, cfg)
	if err != nil {
		dc.disableSandbox(cfg, logger.WithError(err), "the sandbox isolation test could not be completed")

		return
	}

	if report.daemonID == dc.orchestratorDaemonID {
		dc.disableSandbox(cfg, logger.WithField("daemon", report.daemonID),
			"DOCKER_INSIDE_HOST is the same daemon PentAGI orchestrates on, so an agent that escapes "+
				"its container reaches PentAGI, the other flows and their credentials")

		return
	}

	if len(report.failed) > 0 {
		dc.disableSandbox(cfg, logger.WithFields(logrus.Fields{
			"daemon": report.daemonID, "failed_checks": report.failed,
			"passed": report.passed, "total": report.total,
		}), "the sandbox daemon did not pass the isolation test")

		return
	}

	logger.WithFields(logrus.Fields{
		"daemon": report.daemonID, "authorization": report.authz, "checks": report.total,
	}).Info("sandbox isolation verified: agents reach a separate Docker daemon that refused every " +
		"host-escape request and served every legitimate one")
}

// disableSandbox turns agent Docker access off for this deployment and says so
// once, with what to fix. Flipping the config rather than a field of this
// client is deliberate: everything that asks whether the sandbox exists --
// including the settings the UI shows -- reads the config.
func (dc *dockerClient) disableSandbox(cfg *config.Config, logger *logrus.Entry, reason string) {
	cfg.DockerInside = false
	dc.inside = false

	logger.Error("sandbox DISABLED for this deployment: " + reason +
		". Agents will run without Docker access until this is fixed; PentAGI keeps running. " +
		"Setup guide: " + sandboxGuideURL)
}
