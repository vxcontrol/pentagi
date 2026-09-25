package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"

	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func statusResponse(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// daemonWith fakes a daemon that answers the ping and hands every other request to handle.
func daemonWith(t *testing.T, handle func(*http.Request) (*http.Response, error)) *client.Client {
	t.Helper()

	cli, err := client.New(
		client.WithHost("tcp://daemon.invalid:2375"),
		client.WithAPIVersion("1.44"),
		client.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if strings.HasSuffix(r.URL.Path, "/_ping") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Api-Version": []string{"1.44"}},
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}
			return handle(r)
		})}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })

	return cli
}

// mockDaemon fakes a daemon running one container per hostname.
func mockDaemon(t *testing.T, hostnames ...string) *client.Client {
	t.Helper()

	items := make([]string, len(hostnames))
	byID := make(map[string]string, len(hostnames))
	for i, hostname := range hostnames {
		id := fmt.Sprintf("c%d", i)
		items[i] = fmt.Sprintf(`{"Id":%q}`, id)
		byID[id] = hostname
	}

	return daemonWith(t, func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/containers/json"):
			return statusResponse(http.StatusOK, "["+strings.Join(items, ",")+"]"), nil
		case strings.HasSuffix(r.URL.Path, "/json"):
			parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/json"), "/")
			id := parts[len(parts)-1]
			return statusResponse(http.StatusOK, fmt.Sprintf(`{"Id":%q,"Config":{"Hostname":%q}}`, id, byID[id])), nil
		}
		return nil, fmt.Errorf("unexpected request: %s", r.URL.Path)
	})
}

func dockerDownDaemon(t *testing.T) *client.Client {
	return daemonWith(t, func(*http.Request) (*http.Response, error) { return nil, errors.New("daemon is down") })
}

func captureLogs(t *testing.T, fn func()) []*logrus.Entry {
	t.Helper()

	hook := new(logtest.Hook)
	previous := logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})
	logrus.AddHook(hook)
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(previous) })

	fn()
	return hook.AllEntries()
}

// dockerLogLine is one expected log entry; worker, when set, is its worker_docker field.
type dockerLogLine struct {
	level  logrus.Level
	text   string
	worker string
}

// dockerRequireLogLines asserts exactly the wanted lines, in order, and none carrying an absent claim.
func dockerRequireLogLines(t *testing.T, entries []*logrus.Entry, want []dockerLogLine, absent ...string) {
	t.Helper()

	var got []string
	for _, entry := range entries {
		got = append(got, entry.Level.String()+": "+entry.Message)
		for _, claim := range absent {
			require.NotContains(t, entry.Message, claim)
		}
	}
	require.Len(t, entries, len(want), "log lines: %q", got)
	for i, line := range want {
		require.Equal(t, line.level, entries[i].Level, got[i])
		require.Contains(t, entries[i].Message, line.text)
		if line.worker != "" {
			require.Equal(t, line.worker, entries[i].Data["worker_docker"], got[i])
		}
	}
}

func TestWorkerDaemon_RunsPentagi_FindsOurContainerOrSaysWhyItCannot(t *testing.T) {
	self, err := os.Hostname()
	if err != nil {
		t.Skip("no hostname")
	}

	listing := func(r *http.Request, inspect func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/containers/json") {
			return statusResponse(http.StatusOK, `[{"Id":"c0"},{"Id":"c1"}]`), nil
		}
		return inspect(r)
	}

	tests := map[string]struct {
		daemon  func(t *testing.T) *client.Client
		timeout time.Duration
		want    bool
		wantErr string
	}{
		"a daemon running our container": {
			daemon: func(t *testing.T) *client.Client { return mockDaemon(t, "someone-else", self) },
			want:   true,
		},
		"a daemon running only other containers": {
			daemon: func(t *testing.T) *client.Client { return mockDaemon(t, "pentagi-terminal-1", "pentagi-terminal-2") },
		},
		"a container that exited before its inspect does not hide ours": {
			daemon: func(t *testing.T) *client.Client {
				return daemonWith(t, func(r *http.Request) (*http.Response, error) {
					return listing(r, func(r *http.Request) (*http.Response, error) {
						if strings.HasSuffix(r.URL.Path, "/c0/json") {
							return statusResponse(http.StatusNotFound, `{"message":"No such container: c0"}`), nil
						}
						return statusResponse(http.StatusOK, fmt.Sprintf(`{"Id":"c1","Config":{"Hostname":%q}}`, self)), nil
					})
				})
			},
			want: true,
		},
		"a daemon that cannot be listed": {
			daemon:  dockerDownDaemon,
			wantErr: "failed to list containers",
		},
		"an inspect the daemon failed": {
			daemon: func(t *testing.T) *client.Client {
				return daemonWith(t, func(r *http.Request) (*http.Response, error) {
					return listing(r, func(*http.Request) (*http.Response, error) {
						return statusResponse(http.StatusInternalServerError, `{"message":"boom"}`), nil
					})
				})
			},
			wantErr: "failed to inspect container c0",
		},
		"an inspect that ran out of time": {
			daemon: func(t *testing.T) *client.Client {
				return daemonWith(t, func(r *http.Request) (*http.Response, error) {
					return listing(r, func(r *http.Request) (*http.Response, error) {
						<-r.Context().Done()
						return nil, r.Context().Err()
					})
				})
			},
			timeout: 200 * time.Millisecond,
			wantErr: "failed to inspect container c0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			if tc.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.timeout)
				defer cancel()
			}

			got, err := runsPentagi(ctx, tc.daemon(t))

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr, "an unanswered daemon says nothing about what it runs")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestWorkerDaemon_WorkerDaemonClient_TargetsTheConfiguredEndpoint(t *testing.T) {
	cli, err := workerDaemonClient(&config.Config{
		DockerInside:     true,
		DockerInsideHost: "tcp://10.0.0.1:3376",
	})
	require.NoError(t, err)
	defer cli.Close()

	require.Equal(t, "tcp://10.0.0.1:3376", cli.DaemonHost())
}

func TestWorkerDaemon_LogWorkerDaemonIsolation_ReportsWhatAgentsReach(t *testing.T) {
	self, err := os.Hostname()
	if err != nil {
		t.Skip("no hostname")
	}

	tests := map[string]struct {
		cfg    *config.Config
		daemon func(t *testing.T) *client.Client
		want   []dockerLogLine
		absent []string
	}{
		"nothing is logged without a config":      {},
		"nothing is logged without DOCKER_INSIDE": {cfg: &config.Config{DockerInside: false}},
		"a socket mounted beside DOCKER_INSIDE_HOST is the host daemon": {
			cfg: &config.Config{
				DockerInside:     true,
				DockerSocket:     "/run/d.sock",
				DockerInsideHost: "tcp://127.0.0.1:1",
			},
			daemon: func(t *testing.T) *client.Client { return mockDaemon(t, self) },
			want: []dockerLogLine{
				{level: logrus.WarnLevel, text: "DOCKER_SOCKET is set"},
				{level: logrus.WarnLevel, text: "agents are given the host daemon", worker: "/run/d.sock"},
			},
			absent: []string{"sandbox tier"},
		},
		"certificates on the worker node cannot be checked": {
			cfg: &config.Config{
				DockerInside:          true,
				DockerInsideHost:      "tcp://10.0.0.1:3376",
				DockerInsideTLSVerify: "1",
				DockerInsideCertPath:  filepath.Join(t.TempDir(), "on-the-worker-node"),
			},
			want: []dockerLogLine{{level: logrus.InfoLevel, text: "live on the worker node"}},
		},
		"a shared daemon without our container does not rule out this host": {
			cfg:    &config.Config{DockerInside: true},
			daemon: func(t *testing.T) *client.Client { return mockDaemon(t, "someone-else") },
			want: []dockerLogLine{
				{level: logrus.WarnLevel, text: "if it runs directly on this host", worker: "autodetected host socket"},
			},
		},
		"a separate daemon that cannot be listed is not claimed as isolated": {
			cfg:    &config.Config{DockerInside: true, DockerInsideHost: "tcp://127.0.0.1:1"},
			daemon: func(t *testing.T) *client.Client { return mockDaemon(t, "pentagi-terminal-1") },
			want:   []dockerLogLine{{level: logrus.WarnLevel, text: "a separate daemon that could not be listed"}},
			absent: []string{"sandbox tier"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var cli *client.Client
			if tc.daemon != nil {
				cli = tc.daemon(t)
			}

			entries := captureLogs(t, func() { logWorkerDaemonIsolation(t.Context(), tc.cfg, cli) })

			dockerRequireLogLines(t, entries, tc.want, tc.absent...)
		})
	}
}

func TestWorkerDaemon_LogSeparateDaemon_ClaimsIsolationOnlyWhenItHolds(t *testing.T) {
	self, err := os.Hostname()
	if err != nil {
		t.Skip("no hostname")
	}

	nodeDaemon := func(t *testing.T) *client.Client {
		return daemonWith(t, func(r *http.Request) (*http.Response, error) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/info"):
				return statusResponse(http.StatusOK, `{"ID":"node-daemon"}`), nil
			case strings.HasSuffix(r.URL.Path, "/containers/json"):
				return statusResponse(http.StatusOK, `[{"Id":"c0"}]`), nil
			case strings.HasSuffix(r.URL.Path, "/c0/json"):
				return statusResponse(http.StatusOK, `{"Id":"c0","Config":{"Hostname":"pentagi-terminal-1"}}`), nil
			}
			return nil, fmt.Errorf("unexpected request: %s", r.URL.Path)
		})
	}
	running := func(hostname string) func(t *testing.T) *client.Client {
		return func(t *testing.T) *client.Client { return mockDaemon(t, hostname) }
	}

	tests := map[string]struct {
		orchestrator, worker func(t *testing.T) *client.Client
		want                 []dockerLogLine
		absent               []string
	}{
		"an orchestration daemon that could not be listed is not reported as clear": {
			orchestrator: dockerDownDaemon,
			worker:       running("pentagi-terminal-1"),
			want:         []dockerLogLine{{level: logrus.InfoLevel, text: "could not be determined"}},
			absent:       []string{"on neither"},
		},
		"one daemon reached through both variables is reported as shared": {
			orchestrator: nodeDaemon,
			worker:       nodeDaemon,
			want: []dockerLogLine{
				{level: logrus.WarnLevel, text: "names the daemon PentAGI orchestrates on"},
				{level: logrus.WarnLevel, text: "if it runs directly on this host"},
			},
			absent: []string{"sandbox tier", "on neither"},
		},
		"a worker daemon that runs PentAGI is not claimed as isolated": {
			orchestrator: running("pentagi-terminal-1"),
			worker:       running(self),
			want:         []dockerLogLine{{level: logrus.WarnLevel, text: "the daemon designated for sandboxes is the one running PentAGI"}},
			absent:       []string{"sandbox tier"},
		},
		"a worker daemon apart from the one running PentAGI keeps an escape in the sandbox tier": {
			orchestrator: running(self),
			worker:       running("pentagi-terminal-1"),
			want:         []dockerLogLine{{level: logrus.InfoLevel, text: "an escape stays inside the sandbox tier"}},
			absent:       []string{"could not be determined"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			orchestrator, worker := tc.orchestrator(t), tc.worker(t)

			entries := captureLogs(t, func() {
				logSeparateDaemon(t.Context(), &config.Config{DockerInsideHost: "tcp://node:2376"}, orchestrator, worker)
			})

			dockerRequireLogLines(t, entries, tc.want, tc.absent...)
		})
	}
}
