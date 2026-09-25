package docker

import (
	"context"
	"errors"
	"testing"
	"time"

	"pentagi/pkg/config"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sandboxInfoLine = `{"kind":"info","api":"1.55","body":{"ID":"sandbox-daemon","Plugins":{"Authorization":["opa-docker-authz"]}}}`

func TestSandbox_ParseReport_ReadsTheProbeWithoutGuessing(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		out        string
		wantErr    string
		wantDaemon string
		wantFailed []string
	}{
		{
			name: "a clean run",
			out: sandboxInfoLine + "\n" +
				`{"kind":"check","id":"privileged","group":"create","expect":"deny","result":"pass","status":403}` + "\n" +
				`{"kind":"summary","total":1,"passed":1,"failed":0}`,
			wantDaemon: "sandbox-daemon",
		},
		{
			name: "a failing check is named by group and id",
			out: sandboxInfoLine + "\n" +
				`{"kind":"check","id":"privileged","group":"create","expect":"deny","result":"fail","status":404}` + "\n" +
				`{"kind":"check","id":"image-pull","group":"positive","expect":"allow","result":"fail","status":500}` + "\n" +
				`{"kind":"summary","total":2,"passed":0,"failed":2}`,
			wantDaemon: "sandbox-daemon",
			wantFailed: []string{"create/privileged", "positive/image-pull"},
		},
		{
			name:    "the probe could not reach the daemon",
			out:     `{"kind":"fatal","reason":"version_unreachable"}`,
			wantErr: "version_unreachable",
		},
		// A truncated run must not read as a clean one: the summary is what says
		// the probe finished, and every check could be missing without it.
		{
			name:    "no summary",
			out:     sandboxInfoLine,
			wantErr: "did not finish",
		},
		{
			name:    "no daemon identity",
			out:     `{"kind":"summary","total":0,"passed":0,"failed":0}`,
			wantErr: "which daemon",
		},
		{
			name:    "a line that is not JSON",
			out:     sandboxInfoLine + "\nDocker policy preflight: curl is required\n",
			wantErr: "not JSON",
		},
		{
			name:    "an empty daemon identity",
			out:     `{"kind":"info","body":{"ID":""}}` + "\n" + `{"kind":"summary","total":0,"passed":0,"failed":0}`,
			wantErr: "no identity",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			report, err := parseSandboxReport([]byte(tc.out))
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantDaemon, report.daemonID)
			assert.Equal(t, tc.wantFailed, report.failed)
		})
	}
}

// Not parallel: captureLogs swaps the standard logger's hooks.
func TestSandbox_DecideSandbox_TurnsTheSandboxOffRatherThanRefusingToStart(t *testing.T) {
	enabled := func() *config.Config {
		return &config.Config{
			DockerInside: true, DockerInsideHost: "tcp://10.0.0.1:2376",
			DockerInsidePolicyTests: true, DockerDefaultImageForTest: "vxcontrol/kali-linux:test",
		}
	}

	for _, tc := range []struct {
		name      string
		cfg       *config.Config
		probe     func(context.Context, *config.Config) (sandboxReport, error)
		wantOn    bool
		wantLevel logrus.Level
		wantText  string
	}{
		{
			name: "sandbox off by configuration says nothing",
			cfg:  &config.Config{},
		},
		{
			name:      "no designated daemon disables it instead of failing the boot",
			cfg:       &config.Config{DockerInside: true},
			wantLevel: logrus.ErrorLevel, wantText: "DOCKER_INSIDE_HOST is empty",
		},
		{
			name: "tests not asked for: allowed, but the operator is told it is unmeasured",
			cfg: &config.Config{
				DockerInside: true, DockerInsideHost: "tcp://10.0.0.1:2376",
			},
			wantOn: true, wantLevel: logrus.WarnLevel, wantText: "NOT tested",
		},
		{
			name: "a passing sandbox stays on",
			cfg:  enabled(),
			probe: func(context.Context, *config.Config) (sandboxReport, error) {
				return sandboxReport{daemonID: "sandbox-daemon", total: 35, passed: 35}, nil
			},
			wantOn: true, wantLevel: logrus.InfoLevel, wantText: "isolation verified",
		},
		{
			name: "a probe that could not run disables it",
			cfg:  enabled(),
			probe: func(context.Context, *config.Config) (sandboxReport, error) {
				return sandboxReport{}, errors.New("could not create the check container")
			},
			wantLevel: logrus.ErrorLevel, wantText: "could not be completed",
		},
		{
			name: "the sandbox daemon is the one PentAGI runs on",
			cfg:  enabled(),
			probe: func(context.Context, *config.Config) (sandboxReport, error) {
				return sandboxReport{daemonID: "orchestrator", total: 35, passed: 35}, nil
			},
			wantLevel: logrus.ErrorLevel, wantText: "same daemon PentAGI orchestrates on",
		},
		{
			name: "a failing check disables it and names what failed",
			cfg:  enabled(),
			probe: func(context.Context, *config.Config) (sandboxReport, error) {
				return sandboxReport{daemonID: "sandbox-daemon", total: 35, passed: 34,
					failed: []string{"create/privileged"}}, nil
			},
			wantLevel: logrus.ErrorLevel, wantText: "did not pass the isolation test",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dc := &dockerClient{
				inside: tc.cfg.DockerInside, orchestratorDaemonID: "orchestrator",
				sandboxProbeRunner: tc.probe,
			}

			entries := captureLogs(t, func() { dc.decideSandbox(t.Context(), tc.cfg) })

			if tc.wantText == "" {
				require.Empty(t, entries)
				require.False(t, tc.cfg.DockerInside)

				return
			}
			require.Len(t, entries, 1)
			assert.Equal(t, tc.wantLevel, entries[0].Level)
			assert.Contains(t, entries[0].Message, tc.wantText)

			// Every verdict that asks the operator to do something points at the
			// guide; the one that says it works has nothing to fix.
			if tc.wantLevel != logrus.InfoLevel {
				assert.Contains(t, entries[0].Message, "worker_node.md")
			}
			assert.Equal(t, tc.wantOn, tc.cfg.DockerInside, "cfg.DockerInside")
			assert.Equal(t, tc.wantOn, dc.inside, "dc.inside")
		})
	}
}

// The daemon itself refuses a container joined to the host network as an endpoint, so this runs against a real one.
func TestSandbox_ProbeSandbox_StartsTheCheckContainerOnTheConfiguredNetwork(t *testing.T) {
	for _, tc := range []struct {
		name    string
		network string
	}{
		{name: "the host network", network: "host"},
		{name: "the default network"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dc := newDaemonClient(t)
			if err := dc.pullImage(t.Context(), probeImage); err != nil {
				t.Skipf("probe image %s unavailable: %v", probeImage, err)
			}
			dc.network = tc.network

			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
			defer cancel()
			_, err := dc.probeSandbox(ctx, &config.Config{DockerDefaultImageForTest: probeImage})

			// Nothing here answers as a sandbox daemon, so the probe itself fails; what
			// must not fail is starting the container it runs in.
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "the check container")
			assert.Contains(t, err.Error(), "probe")
		})
	}
}

// The probe measures the container agents get, so its spec must come from the
// same place theirs does.
func TestSandbox_WorkerSpec_IsTheSameContainerAgentsGet(t *testing.T) {
	t.Parallel()

	plain, plainHost := WorkerSpec(&config.Config{}, "some/image:tag")
	assert.Equal(t, "some/image:tag", plain.Image)
	assert.Equal(t, []string{"ALL"}, plainHost.CapDrop)
	assert.NotContains(t, plainHost.CapAdd, "NET_ADMIN")
	for _, forbidden := range []string{"SYS_ADMIN", "SYS_MODULE", "SYS_RAWIO", "SYS_BOOT", "MKNOD"} {
		assert.NotContains(t, plainHost.CapAdd, forbidden,
			"%s is a host-escape primitive and must never be in the bounding set", forbidden)
	}

	_, admin := WorkerSpec(&config.Config{DockerNetAdmin: true}, "some/image:tag")
	assert.Contains(t, admin.CapAdd, "NET_ADMIN")
}

// The probe is shipped in the binary, so a change that breaks its contract with
// the parser has to fail here rather than on a customer's daemon.
func TestSandbox_ProbeScript_EmitsTheContractTheParserReads(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		`\"kind\":\"info\"`, `\"kind\":\"check\"`, `\"kind\":\"summary\"`, `\"kind\":\"fatal\"`,
		"PENTAGI_TEST_IMAGE", "DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH",
	} {
		assert.Contains(t, sandboxProbeScript, want)
	}
	assert.NotContains(t, sandboxProbeScript, "\r\n", "CRLF would break the shell")
}
