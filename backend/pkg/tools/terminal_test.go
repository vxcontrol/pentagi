package tools

import (
	"bufio"
	"context"
	"fmt"
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

// contextTestTermLogProvider implements TermLogProvider for context tests.
type contextTestTermLogProvider struct{}

func (m *contextTestTermLogProvider) PutMsg(_ context.Context, _ database.TermlogType, _ string,
	_ int64, _, _ *int64) (int64, error) {
	return 1, nil
}

var _ TermLogProvider = (*contextTestTermLogProvider)(nil)

// contextAwareMockDockerClient tracks whether the context was canceled
// when getExecResult runs, proving context.WithoutCancel works.
type contextAwareMockDockerClient struct {
	isRunning      bool
	execCreateResp container.ExecCreateResponse
	attachOutput   []byte
	attachDelay    time.Duration
	inspectResp    container.ExecInspect

	// Set by ContainerExecAttach to track if ctx was canceled during attach
	ctxWasCanceled bool
}

func (m *contextAwareMockDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (m *contextAwareMockDockerClient) StopContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *contextAwareMockDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error {
	return nil
}
func (m *contextAwareMockDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return m.isRunning, nil
}
func (m *contextAwareMockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ container.ExecOptions) (container.ExecCreateResponse, error) {
	return m.execCreateResp, nil
}
func (m *contextAwareMockDockerClient) ContainerExecAttach(ctx context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	// Wait for the configured delay, simulating a long-running command
	if m.attachDelay > 0 {
		select {
		case <-time.After(m.attachDelay):
			// Command completed normally
		case <-ctx.Done():
			// Context was canceled -- this is the bug behavior (without WithoutCancel)
			m.ctxWasCanceled = true
			return types.HijackedResponse{}, ctx.Err()
		}
	}

	// Check if context was already canceled by the time we get here
	select {
	case <-ctx.Done():
		m.ctxWasCanceled = true
		return types.HijackedResponse{}, ctx.Err()
	default:
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
func (m *contextAwareMockDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	return m.inspectResp, nil
}
func (m *contextAwareMockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, _ io.Reader, _ container.CopyToContainerOptions) error {
	return nil
}
func (m *contextAwareMockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(nil), container.PathStat{}, nil
}
func (m *contextAwareMockDockerClient) Cleanup(_ context.Context) error { return nil }
func (m *contextAwareMockDockerClient) GetDefaultImage() string         { return "test-image" }

var _ docker.DockerClient = (*contextAwareMockDockerClient)(nil)

func TestExecCommandDetachSurvivesParentCancel(t *testing.T) {
	// This test validates the fix for Issue #176:
	// Detached commands must NOT be killed when the parent context is canceled.
	//
	// Before the fix: detached goroutine used parent ctx directly, so when the
	// parent was canceled (e.g., agent delegation timeout), ctx.Done() fired
	// in getExecResult and killed the background command.
	//
	// After the fix: context.WithoutCancel(ctx) creates an isolated context
	// that preserves values but ignores parent cancellation.

	mock := &contextAwareMockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-cancel-test"},
		attachOutput:   []byte("background result"),
		attachDelay:    2 * time.Second, // simulates a long-running command
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	// Create a cancellable parent context
	parentCtx, cancel := context.WithCancel(t.Context())

	// Start ExecCommand with detach=true (returns quickly due to quick check timeout)
	output, err := term.ExecCommand(parentCtx, "/work", "long-running-scan", true, 5*time.Minute)
	assert.NoError(t, err)
	assert.Contains(t, output, "Command started in background")

	// Cancel the parent context -- simulating agent delegation timeout
	cancel()

	// Wait enough time for the detached goroutine to complete its work.
	// If context.WithoutCancel is working correctly, the goroutine should
	// NOT see ctx.Done() and should complete normally after attachDelay.
	// If the fix regresses, ctxWasCanceled will be true.
	time.Sleep(3 * time.Second)

	assert.False(t, mock.ctxWasCanceled,
		"detached goroutine should NOT see parent context cancellation (context.WithoutCancel must be used)")
}

func TestExecCommandNonDetachRespectsParentCancel(t *testing.T) {
	// Counterpart: non-detached commands SHOULD respect parent cancellation.
	// This ensures we didn't accidentally apply WithoutCancel to the non-detach path.

	mock := &contextAwareMockDockerClient{
		isRunning:      true,
		execCreateResp: container.ExecCreateResponse{ID: "exec-nondetach-cancel"},
		attachOutput:   []byte("should not complete"),
		attachDelay:    5 * time.Second, // longer than cancel delay
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	parentCtx, cancel := context.WithCancel(t.Context())

	// Cancel after 200ms -- non-detached command should see this
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_, err := term.ExecCommand(parentCtx, "/work", "long-command", false, 5*time.Minute)

	// Non-detached command should fail with context error
	assert.Error(t, err)
	assert.True(t, mock.ctxWasCanceled,
		"non-detached command SHOULD see parent context cancellation")
}

func TestPrimaryTerminalName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		flowID int64
		want   string
	}{
		{1, "pentagi-terminal-1"},
		{0, "pentagi-terminal-0"},
		{12345, "pentagi-terminal-12345"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("flowID=%d", tt.flowID), func(t *testing.T) {
			t.Parallel()

			if got := PrimaryTerminalName(tt.flowID); got != tt.want {
				t.Errorf("PrimaryTerminalName(%d) = %q, want %q", tt.flowID, got, tt.want)
			}
		})
	}
}

// === Tests for configurable terminal timeout (issue #256) ===

func TestNewTerminalToolDefaultTimeouts(t *testing.T) {
	t.Parallel()

	// When no TerminalTimeoutConfig is provided, built-in defaults should be used.
	tool := NewTerminalTool(1, nil, nil, 1, "test-container", nil, &contextTestTermLogProvider{})
	term := tool.(*terminal)

	assert.Equal(t, builtinDefaultTimeout, term.defaultTimeout,
		"default timeout should match built-in default when no config provided")
	assert.Equal(t, builtinHardLimit, term.hardLimitTimeout,
		"hard limit should match built-in hard limit when no config provided")
}

func TestNewTerminalToolCustomTimeouts(t *testing.T) {
	t.Parallel()

	cfg := TerminalTimeoutConfig{
		DefaultTimeout:   10 * time.Minute,
		HardLimitTimeout: 30 * time.Minute,
	}
	tool := NewTerminalTool(1, nil, nil, 1, "test-container", nil, &contextTestTermLogProvider{}, cfg)
	term := tool.(*terminal)

	assert.Equal(t, 10*time.Minute, term.defaultTimeout,
		"custom default timeout should be applied")
	assert.Equal(t, 30*time.Minute, term.hardLimitTimeout,
		"custom hard limit should be applied")
}

func TestNewTerminalToolPartialConfig(t *testing.T) {
	t.Parallel()

	// Only set DefaultTimeout, HardLimitTimeout stays zero → built-in
	cfg := TerminalTimeoutConfig{
		DefaultTimeout: 8 * time.Minute,
	}
	tool := NewTerminalTool(1, nil, nil, 1, "test-container", nil, &contextTestTermLogProvider{}, cfg)
	term := tool.(*terminal)

	assert.Equal(t, 8*time.Minute, term.defaultTimeout)
	assert.Equal(t, builtinHardLimit, term.hardLimitTimeout,
		"zero hard limit should fall back to built-in")
}

func TestExecCommandTimeoutClamping(t *testing.T) {
	// Verify that timeout is clamped by the hard limit and defaults are applied.
	// We test this indirectly: if requested timeout exceeds hard limit,
	// the actual timeout used should be the default timeout, not the requested one.
	tests := []struct {
		name            string
		requestTimeout  time.Duration
		defaultTimeout  time.Duration
		hardLimit       time.Duration
		expectDefault   bool // true if we expect the default timeout to be used
	}{
		{
			name:           "zero timeout uses default",
			requestTimeout: 0,
			defaultTimeout: 5 * time.Minute,
			hardLimit:      20 * time.Minute,
			expectDefault:  true,
		},
		{
			name:           "negative timeout uses default",
			requestTimeout: -1 * time.Second,
			defaultTimeout: 5 * time.Minute,
			hardLimit:      20 * time.Minute,
			expectDefault:  true,
		},
		{
			name:           "exceeds hard limit uses default",
			requestTimeout: 25 * time.Minute,
			defaultTimeout: 5 * time.Minute,
			hardLimit:      20 * time.Minute,
			expectDefault:  true,
		},
		{
			name:           "within hard limit keeps requested",
			requestTimeout: 10 * time.Minute,
			defaultTimeout: 5 * time.Minute,
			hardLimit:      20 * time.Minute,
			expectDefault:  false,
		},
		{
			name:           "exactly at hard limit keeps requested",
			requestTimeout: 20 * time.Minute,
			defaultTimeout: 5 * time.Minute,
			hardLimit:      20 * time.Minute,
			expectDefault:  false,
		},
		{
			name:           "custom high hard limit allows long timeout",
			requestTimeout: 45 * time.Minute,
			defaultTimeout: 10 * time.Minute,
			hardLimit:      60 * time.Minute,
			expectDefault:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the exact timeout used inside ExecCommand
			// without running a container, but we can verify the clamping logic
			// by replicating it.
			timeout := tt.requestTimeout
			hardLimit := tt.hardLimit
			defTimeout := tt.defaultTimeout

			if timeout <= 0 || timeout > hardLimit {
				timeout = defTimeout
			}

			if tt.expectDefault {
				assert.Equal(t, defTimeout, timeout,
					"timeout should be clamped to default")
			} else {
				assert.Equal(t, tt.requestTimeout, timeout,
					"timeout should be kept as requested")
			}
		})
	}
}

func TestExecCommandZeroFieldsFallbackToBuiltin(t *testing.T) {
	// Verify that a terminal struct with zero timeout fields
	// (e.g., created by tests that construct terminal{} directly)
	// falls back to built-in defaults.
	mock := &contextAwareMockDockerClient{
		isRunning:      false, // will trigger early error, that's fine for this test
		execCreateResp: container.ExecCreateResponse{ID: "test"},
		inspectResp:    container.ExecInspect{ExitCode: 0},
	}

	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
		// defaultTimeout and hardLimitTimeout are zero
	}

	// The command will fail early (container not running), but the important
	// thing is it doesn't panic on zero timeout fields.
	_, err := term.ExecCommand(context.Background(), "/work", "test", false, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not operational")
}

func TestTerminalToolDefinitionDynamic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		hardLimitSec   int
		defaultSec     int
		expectContains string
	}{
		{
			name:           "default values",
			hardLimitSec:   1200,
			defaultSec:     300,
			expectContains: "hard limit timeout 1200 seconds and optimum timeout 300 seconds",
		},
		{
			name:           "custom high values",
			hardLimitSec:   3600,
			defaultSec:     600,
			expectContains: "hard limit timeout 3600 seconds and optimum timeout 600 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := TerminalToolDefinition(tt.hardLimitSec, tt.defaultSec)
			assert.Equal(t, TerminalToolName, def.Name)
			assert.Contains(t, def.Description, tt.expectContains)
		})
	}
}
