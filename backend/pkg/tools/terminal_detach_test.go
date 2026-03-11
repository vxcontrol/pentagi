package tools

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// mockDockerClientBlocking is a docker client that blocks ContainerExecAttach
// until the context is done, useful for testing cancellation propagation.
type mockDockerClientBlocking struct {
	docker.DockerClient

	mu        sync.Mutex
	attachCtx context.Context
}

func (m *mockDockerClientBlocking) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func (m *mockDockerClientBlocking) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{ID: "test-exec-id"}, nil
}

func (m *mockDockerClientBlocking) ContainerExecAttach(ctx context.Context, _ string, _ container.ExecAttachOptions) (dtypes.HijackedResponse, error) {
	m.mu.Lock()
	m.attachCtx = ctx
	m.mu.Unlock()

	// Create a connection that stays open — it will be closed by context cancellation/timeout.
	server, client := net.Pipe()

	// Close server side after context is done to unblock reads.
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return dtypes.NewHijackedResponse(client, ""), nil
}

func (m *mockDockerClientBlocking) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return container.ExecInspect{ExitCode: 0}, nil
}

func (m *mockDockerClientBlocking) getAttachCtx() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.attachCtx
}

// mockTermLogProvider implements TermLogProvider for testing.
type mockTermLogProvider struct{}

func (m *mockTermLogProvider) PutMsg(_ context.Context, _ database.TermlogType, _ string, _ int64, _, _ *int64) (int64, error) {
	return 1, nil
}

// TestExecCommandDetachSurvivesParentCancel verifies that when detach=true,
// the background goroutine uses a context that is NOT canceled when the parent
// context is canceled. This is the fix for issue #176.
func TestExecCommandDetachSurvivesParentCancel(t *testing.T) {
	t.Parallel()

	dockerMock := &mockDockerClientBlocking{}
	tlpMock := &mockTermLogProvider{}

	term := &terminal{
		flowID:       1,
		containerLID: "test-container",
		dockerClient: dockerMock,
		tlp:          tlpMock,
	}

	// Create a parent context with a very short timeout (simulating the ~2.5 min parent timeout).
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer parentCancel()

	// Start ExecCommand with detach=true and a long timeout for the command itself.
	result, err := term.ExecCommand(parentCtx, "/tmp", "long-running-scan", true, 5*time.Minute)
	if err != nil {
		t.Fatalf("ExecCommand returned error: %v", err)
	}

	// The detached command should return quickly with "started in background" message.
	if result == "" {
		t.Fatal("expected non-empty result for detached command")
	}

	// Wait for parent context to expire.
	<-parentCtx.Done()
	time.Sleep(200 * time.Millisecond)

	// Verify the goroutine's context (used inside getExecResult) is NOT canceled
	// even though the parent context is done.
	attachCtx := dockerMock.getAttachCtx()
	if attachCtx == nil {
		t.Fatal("attach context was never set — goroutine may not have started")
	}
	if attachCtx.Err() != nil {
		t.Errorf("detached goroutine context should NOT be canceled when parent is canceled, but got: %v", attachCtx.Err())
	}
}

// TestExecCommandDetachGoroutineUsesOwnTimeout verifies that the detached
// goroutine's context has its own timeout independent of the parent.
func TestExecCommandDetachGoroutineUsesOwnTimeout(t *testing.T) {
	t.Parallel()

	dockerMock := &mockDockerClientBlocking{}
	tlpMock := &mockTermLogProvider{}

	term := &terminal{
		flowID:       2,
		containerLID: "test-container-2",
		dockerClient: dockerMock,
		tlp:          tlpMock,
	}

	// Parent context with very short timeout.
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer parentCancel()

	_, err := term.ExecCommand(parentCtx, "/tmp", "long-running-cmd", true, 2*time.Second)
	if err != nil {
		t.Fatalf("ExecCommand returned error: %v", err)
	}

	// Wait for parent to expire.
	<-parentCtx.Done()
	time.Sleep(100 * time.Millisecond)

	// The goroutine's context should NOT be canceled yet (it has its own 2s timeout).
	attachCtx := dockerMock.getAttachCtx()
	if attachCtx == nil {
		t.Fatal("goroutine never called ContainerExecAttach")
	}
	if attachCtx.Err() != nil {
		t.Errorf("detached goroutine context canceled prematurely: %v", attachCtx.Err())
	}

	// Wait for the goroutine's own timeout to fire.
	time.Sleep(3 * time.Second)
	if attachCtx.Err() == nil {
		t.Error("detached goroutine context should have timed out by now")
	}
}
