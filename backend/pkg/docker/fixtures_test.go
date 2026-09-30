package docker

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const probeImage = "alpine:3.23.5"

// listCommandsScript prints every live process's cmdline with its NUL separators dropped.
const listCommandsScript = `for p in /proc/[0-9]*; do
  c=
  { while IFS= read -r l; do c="$c$l"; done <"$p/cmdline"; c="$c$l"; } 2>/dev/null
  [ -n "$c" ] && echo "$c"
done`

// listingMarker is in every genuine listing; a refused exec's daemon message lacks it.
const listingMarker = `/proc/[0-9]*`

// newDaemonClient binds a client to the local daemon, skipping the test when none is reachable.
func newDaemonClient(t *testing.T) *dockerClient {
	t.Helper()

	cli, err := client.New(client.FromEnv)
	if err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}

	if _, err := cli.Ping(t.Context(), client.PingOptions{NegotiateAPIVersion: true}); err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	return &dockerClient{client: cli, logger: logger}
}

func startProbeSandbox(t *testing.T, dc *dockerClient, image string) string {
	t.Helper()

	return startSandbox(t, dc, &container.Config{Image: image}, &container.HostConfig{})
}

func startSandbox(t *testing.T, dc *dockerClient, config *container.Config, hostConfig *container.HostConfig) string {
	t.Helper()

	config.Entrypoint = []string{"tail", "-f", "/dev/null"}
	config.Labels = map[string]string{"pentagi.test": t.Name()}

	ctx := t.Context()
	created, err := dc.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     config,
		HostConfig: hostConfig,
	})
	if cerrdefs.IsNotFound(err) {
		t.Skipf("%s is not present locally", config.Image)
	}
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.WithoutCancel(ctx)
		if _, err := dc.client.ContainerRemove(ctx, created.ID, client.ContainerRemoveOptions{Force: true}); err != nil &&
			!cerrdefs.IsNotFound(err) {
			t.Errorf("cleanup: failed to remove container %q: %v", created.ID, err)
		}
	})

	_, err = dc.client.ContainerStart(ctx, created.ID, client.ContainerStartOptions{})
	require.NoError(t, err)

	return created.ID
}

func execOutput(ctx context.Context, dc *dockerClient, containerID, script string) (string, error) {
	created, err := dc.ContainerExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          []string{"sh", "-c", script},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", err
	}

	resp, err := dc.ContainerExecAttach(ctx, created.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	out, err := demuxExecStdout(resp.Reader, maxListStdoutBytes)

	return string(out), err
}

func runInSandbox(t *testing.T, dc *dockerClient, containerID, script string) string {
	t.Helper()

	out, err := execOutput(t.Context(), dc, containerID, script)
	require.NoError(t, err)

	return out
}

// startInSandbox launches cmd as an exec; the channel yields the stream's read error once the exec's output ends.
func startInSandbox(t *testing.T, dc *dockerClient, containerID string, cmd []string, tty bool) <-chan error {
	t.Helper()

	ctx := t.Context()
	created, err := dc.ContainerExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          tty,
	})
	require.NoError(t, err)

	resp, err := dc.ContainerExecAttach(ctx, created.ID, client.ExecAttachOptions{TTY: tty})
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.Discard, resp.Reader)
		resp.Close()
		done <- err
	}()

	return done
}

// liveCommands returns the container's running commands and whether the exec produced a listing at all.
func liveCommands(ctx context.Context, dc *dockerClient, containerID string) (string, bool) {
	out, err := execOutput(ctx, dc, containerID, listCommandsScript)
	if err != nil {
		return "", false
	}

	return out, strings.Contains(out, listingMarker)
}

func readLiveCommands(t *testing.T, dc *dockerClient, containerID string) string {
	t.Helper()

	var commands string
	require.Eventually(t, func() bool {
		var ok bool
		commands, ok = liveCommands(t.Context(), dc, containerID)

		return ok
	}, 15*time.Second, 100*time.Millisecond, "the container never returned a process listing")

	return commands
}
