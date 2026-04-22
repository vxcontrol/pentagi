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

func TestConfiguredExecTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured time.Duration
		want       time.Duration
	}{
		{
			name:       "typical value is returned as-is",
			configured: 600 * time.Second,
			want:       600 * time.Second,
		},
		{
			name:       "new default (1200 s) is returned as-is",
			configured: 1200 * time.Second,
			want:       1200 * time.Second,
		},
		{
			name:       "exactly at the 3-hour ceiling is returned as-is",
			configured: maxExplicitExecCommandTimeout,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "zero is capped to the 3-hour ceiling",
			configured: 0,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "negative one second is capped to the 3-hour ceiling",
			configured: -1 * time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "large negative is capped to the 3-hour ceiling",
			configured: -9999 * time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "one second above the ceiling is capped",
			configured: maxExplicitExecCommandTimeout + time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
		{
			name:       "very large value (> 3 h) is capped to the 3-hour ceiling",
			configured: 100000 * time.Second,
			want:       maxExplicitExecCommandTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			term := &terminal{defaultExecTimeout: tt.configured}
			assert.Equal(t, tt.want, term.configuredExecTimeout())
		})
	}
}

func TestNormalizeExecTimeout(t *testing.T) {
	t.Parallel()

	// ceilFor computes the effective runtime ceiling for a given configured value:
	// it equals configuredExecTimeout() + defaultExtraExecTimeout.
	ceilFor := func(configured time.Duration) time.Duration {
		term := &terminal{defaultExecTimeout: configured}
		return term.configuredExecTimeout() + defaultExtraExecTimeout
	}

	tests := []struct {
		name       string
		configured time.Duration
		requested  time.Duration
		want       time.Duration
	}{
		// --- Explicit positive values: preserved when within the operator ceiling ---
		{
			name:       "typical explicit value is preserved",
			configured: 10 * time.Minute,
			requested:  45 * time.Second,
			want:       45 * time.Second,
		},
		{
			name:       "explicit value exactly at the ceiling is preserved",
			configured: 10 * time.Minute,
			requested:  ceilFor(10 * time.Minute), // 600s + 5s = 605s
			want:       ceilFor(10 * time.Minute),
		},
		{
			name:       "explicit value one second above the ceiling falls back to ceiling",
			configured: 10 * time.Minute,
			requested:  ceilFor(10*time.Minute) + time.Second,
			want:       ceilFor(10 * time.Minute),
		},
		{
			name:       "explicit value at the default configured (1200 s) is preserved",
			configured: 1200 * time.Second,
			requested:  1200 * time.Second,
			want:       1200 * time.Second,
		},
		{
			name:       "explicit value above the 1200-s ceiling falls back to that ceiling",
			configured: 1200 * time.Second,
			requested:  ceilFor(1200*time.Second) + time.Second, // 1205s + 1s → fallback
			want:       ceilFor(1200 * time.Second),             // 1205s
		},
		{
			name:       "explicit value at the 3-hour ceiling is preserved when configured=0",
			configured: 0,
			requested:  ceilFor(0), // 3h + 5s
			want:       ceilFor(0),
		},
		{
			name:       "explicit value above the 3-hour ceiling falls back to 3-hour ceiling",
			configured: 0,
			requested:  ceilFor(0) + time.Second,
			want:       ceilFor(0),
		},

		// --- Zero requested: falls back to the operator ceiling ---
		{
			name:       "zero requested with typical configured falls back to ceiling",
			configured: 10 * time.Minute,
			requested:  0,
			want:       ceilFor(10 * time.Minute), // 605s
		},
		{
			name:       "zero requested with default configured (1200 s) falls back to ceiling",
			configured: 1200 * time.Second,
			requested:  0,
			want:       ceilFor(1200 * time.Second), // 1205s
		},
		{
			name:       "zero requested with configured=0 falls back to 3-hour ceiling",
			configured: 0,
			requested:  0,
			want:       ceilFor(0), // 3h + 5s
		},
		{
			name:       "zero requested with oversized configured (> 3 h) falls back to 3-hour ceiling",
			configured: 100000 * time.Second,
			requested:  0,
			want:       ceilFor(0), // capped to 3h + 5s
		},

		// --- Negative requested: treated identically to zero ---
		{
			name:       "negative requested falls back to configured ceiling",
			configured: 10 * time.Minute,
			requested:  -5 * time.Second,
			want:       ceilFor(10 * time.Minute),
		},
		{
			name:       "negative requested with configured=0 falls back to 3-hour ceiling",
			configured: 0,
			requested:  -1 * time.Second,
			want:       ceilFor(0),
		},
		{
			name:       "negative requested with negative configured falls back to 3-hour ceiling",
			configured: -5 * time.Second,
			requested:  -1 * time.Second,
			want:       ceilFor(0), // both negative → absolute 3-hour max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			term := &terminal{defaultExecTimeout: tt.configured}
			assert.Equal(t, tt.want, term.normalizeExecTimeout(tt.requested))
		})
	}
}
