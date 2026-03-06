package tools

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

// mockDockerClient implements docker.DockerClient for terminal testing.
type mockDockerClient struct {
	isRunning    bool
	isRunningErr error

	execCreateResp container.ExecCreateResponse
	execCreateErr  error

	attachOutput []byte
	attachDelay  time.Duration
	attachErr    error

	inspectResp container.ExecInspect
	inspectErr  error
}

func (m *mockDockerClient) SpawnContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (m *mockDockerClient) StopContainer(_ context.Context, _ string, _ int64) error  { return nil }
func (m *mockDockerClient) DeleteContainer(_ context.Context, _ string, _ int64) error { return nil }
func (m *mockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return m.isRunning, m.isRunningErr
}
func (m *mockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return m.execCreateResp, m.execCreateErr
}
func (m *mockDockerClient) ContainerExecAttach(ctx context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	if m.attachErr != nil {
		return types.HijackedResponse{}, m.attachErr
	}

	// Simulate delay for detach timing tests
	if m.attachDelay > 0 {
		select {
		case <-time.After(m.attachDelay):
		case <-ctx.Done():
			return types.HijackedResponse{}, ctx.Err()
		}
	}

	pr, pw := net.Pipe()
	go func() {
		pw.Write(m.attachOutput)
		pw.Close()
	}()

	return types.HijackedResponse{
		Conn:   pr,
		Reader: bufio.NewReader(pr),
	}, nil
}
func (m *mockDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return m.inspectResp, m.inspectErr
}
func (m *mockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, _ io.Reader, _ container.CopyToContainerOptions) error {
	return nil
}
func (m *mockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(bytes.NewReader(nil)), container.PathStat{}, nil
}
func (m *mockDockerClient) Cleanup(_ context.Context) error { return nil }
func (m *mockDockerClient) GetDefaultImage() string         { return "test-image" }

// Compile-time interface check
var _ docker.DockerClient = (*mockDockerClient)(nil)

// mockTermLogProvider implements TermLogProvider for testing.
type mockTermLogProvider struct{}

func (m *mockTermLogProvider) PutMsg(_ context.Context, _ database.TermlogType, _ string,
	_ int64, _, _ *int64) (int64, error) {
	return 1, nil
}

// Compile-time interface check
var _ TermLogProvider = (*mockTermLogProvider)(nil)

func newTestTerminal(dc docker.DockerClient) *terminal {
	return &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container-id",
		dockerClient: dc,
		tlp:          &mockTermLogProvider{},
	}
}

func TestExecCommandDetachReturnsQuickly(t *testing.T) {
	// When detach=true and command takes long, ExecCommand should return
	// "Command started in background" within defaultQuickCheckTimeout (500ms).
	mock := &mockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-1"},
		attachOutput:   []byte("slow output"),
		attachDelay:    2 * time.Second, // longer than quickCheckTimeout
	}

	term := newTestTerminal(mock)
	start := time.Now()
	output, err := term.ExecCommand(context.Background(), "/work", "sleep 10", true, 5*time.Minute)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Contains(t, output, "Command started in background")
	assert.Less(t, elapsed, 2*time.Second, "detach should return within ~500ms, not wait for command")
}

func TestExecCommandDetachQuickCompletion(t *testing.T) {
	// When detach=true and command completes before quickCheckTimeout,
	// ExecCommand should return the actual output.
	mock := &mockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-2"},
		attachOutput:   []byte("fast output"),
		attachDelay:    0, // no delay, completes immediately
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := newTestTerminal(mock)
	output, err := term.ExecCommand(context.Background(), "/work", "echo hello", true, 5*time.Minute)

	assert.NoError(t, err)
	assert.Contains(t, output, "fast output")
}

func TestExecCommandNonDetachWaitsForCompletion(t *testing.T) {
	// When detach=false, ExecCommand should wait for and return the full output.
	mock := &mockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-3"},
		attachOutput:   []byte("full output from command"),
		attachDelay:    100 * time.Millisecond,
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := newTestTerminal(mock)
	output, err := term.ExecCommand(context.Background(), "/work", "ls -la", false, 5*time.Minute)

	assert.NoError(t, err)
	assert.Contains(t, output, "full output from command")
}

func TestExecCommandContainerNotRunning(t *testing.T) {
	// When the container is not running, ExecCommand should return an error immediately.
	mock := &mockDockerClient{
		isRunning: false,
	}

	term := newTestTerminal(mock)
	_, err := term.ExecCommand(context.Background(), "/work", "ls", false, 5*time.Minute)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container is not running")
}
