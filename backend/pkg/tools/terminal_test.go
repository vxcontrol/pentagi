package tools

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/flowfiles"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
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
// when getExecResult runs, proving context.WithoutCancel works. It also
// records which of CopyFromContainer (read_file) / CopyToContainer
// (write_file) was invoked, for tests asserting on the file-tool's
// inferred/validated action.
type contextAwareMockDockerClient struct {
	isRunning      bool
	execCreateResp client.ExecCreateResult
	attachOutput   []byte
	attachDelay    time.Duration
	inspectResp    client.ExecInspectResult

	// Set by ContainerExecAttach to track if ctx was canceled during attach
	ctxWasCanceled bool

	// Set by CopyFromContainer/CopyToContainer to track which file operation ran
	copyFromCalled bool
	copyToCalled   bool

	// readFileContent, when non-empty, is what CopyFromContainer returns as
	// the "current" file content (wrapped in a single-file tar archive) -
	// used by EditFile tests that need to read-then-patch existing content.
	readFileContent string
	// writtenContent captures the last file content CopyToContainer received
	// (unwrapped from its tar archive), so tests can assert on the result of
	// an edit_file/write_file operation.
	writtenContent string
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
func (m *contextAwareMockDockerClient) ContainerExecCreate(_ context.Context, _ string, _ client.ExecCreateOptions) (client.ExecCreateResult, error) {
	return m.execCreateResp, nil
}
func (m *contextAwareMockDockerClient) ContainerExecAttach(ctx context.Context, _ string, _ client.ExecAttachOptions) (client.HijackedResponse, error) {
	// Wait for the configured delay, simulating a long-running command
	if m.attachDelay > 0 {
		select {
		case <-time.After(m.attachDelay):
			// Command completed normally
		case <-ctx.Done():
			// Context was canceled -- this is the bug behavior (without WithoutCancel)
			m.ctxWasCanceled = true
			return client.HijackedResponse{}, ctx.Err()
		}
	}

	// Check if context was already canceled by the time we get here
	select {
	case <-ctx.Done():
		m.ctxWasCanceled = true
		return client.HijackedResponse{}, ctx.Err()
	default:
	}

	pr, pw := net.Pipe()
	go func() {
		pw.Write(m.attachOutput)
		pw.Close()
	}()

	return client.HijackedResponse{
		Conn:   pr,
		Reader: bufio.NewReader(pr),
	}, nil
}
func (m *contextAwareMockDockerClient) ContainerStatPath(_ context.Context, _ string, _ string) (container.PathStat, error) {
	return container.PathStat{}, nil
}
func (m *contextAwareMockDockerClient) ListContainerDir(_ context.Context, _ string, _ string) (docker.ContainerDirListing, error) {
	return docker.ContainerDirListing{}, nil
}
func (m *contextAwareMockDockerClient) ContainerExecInspect(_ context.Context, _ string) (client.ExecInspectResult, error) {
	return m.inspectResp, nil
}
func (m *contextAwareMockDockerClient) CopyToContainer(_ context.Context, _ string, _ string, src io.Reader, _ client.CopyToContainerOptions) error {
	m.copyToCalled = true

	tarReader := tar.NewReader(src)
	if hdr, err := tarReader.Next(); err == nil {
		buf := make([]byte, hdr.Size)
		_, _ = io.ReadFull(tarReader, buf)
		m.writtenContent = string(buf)
	}

	return nil
}
func (m *contextAwareMockDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	m.copyFromCalled = true

	if m.readFileContent == "" {
		return io.NopCloser(bytes.NewReader(nil)), container.PathStat{}, nil
	}

	var tarBuffer bytes.Buffer
	tarWriter := tar.NewWriter(&tarBuffer)
	_ = tarWriter.WriteHeader(&tar.Header{
		Name: "file",
		Mode: 0600,
		Size: int64(len(m.readFileContent)),
	})
	_, _ = tarWriter.Write([]byte(m.readFileContent))
	_ = tarWriter.Close()

	return io.NopCloser(&tarBuffer), container.PathStat{}, nil
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
		execCreateResp: client.ExecCreateResult{ID: "exec-cancel-test"},
		attachOutput:   []byte("background result"),
		attachDelay:    2 * time.Second, // simulates a long-running command
		inspectResp:    client.ExecInspectResult{ExitCode: 0},
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
		execCreateResp: client.ExecCreateResult{ID: "exec-nondetach-cancel"},
		attachOutput:   []byte("should not complete"),
		attachDelay:    5 * time.Second, // longer than cancel delay
		inspectResp:    client.ExecInspectResult{ExitCode: 0},
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

func TestTerminalHandle_FileAction_DefaultsToWriteFile_WhenContentPresent(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	// No "action" field at all - mirrors the malformed tool calls observed in
	// production logs (LLM omits the required 'action' when content+path make
	// the intent unambiguous).
	args := json.RawMessage(`{"path":"/work/test.py","content":"print(1)","message":"m"}`)
	_, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected inferred write_file to succeed, got error: %v", err)
	}
	if !mock.copyToCalled || mock.copyFromCalled {
		t.Fatalf("expected CopyToContainer (write_file) to be called, copyTo=%v copyFrom=%v",
			mock.copyToCalled, mock.copyFromCalled)
	}
}

// TestTerminalHandle_FileAction_ExtraQuotedAction_StillDispatches reproduces
// the exact production failure: the LLM sent action/path wrapped in an extra
// literal pair of quotes (e.g. `"action": "\"write_file\""`), which used to
// fail with "unknown file action" since the corrupted value matched no case.
// The String type now unwraps this at unmarshal time, so dispatch succeeds.
func TestTerminalHandle_FileAction_ExtraQuotedAction_StillDispatches(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	args := json.RawMessage(`{"action": "\"write_file\"", "path": "\"/home/evidence/inject.py\"", "content": "print(1)", "message": "m"}`)
	_, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected extra-quoted write_file to still dispatch, got error: %v", err)
	}
	if !mock.copyToCalled || mock.copyFromCalled {
		t.Fatalf("expected CopyToContainer (write_file) to be called, copyTo=%v copyFrom=%v",
			mock.copyToCalled, mock.copyFromCalled)
	}
}

// TestTerminalHandle_FileAction_EditFile_AppliesDiff exercises edit_file end
// to end through Handle(): it reads the mock's current content, applies the
// diff in memory, and writes the result back - all via CopyFromContainer /
// CopyToContainer, exactly like write_file, but without resending unchanged
// content.
func TestTerminalHandle_FileAction_EditFile_AppliesDiff(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		isRunning:       true,
		readFileContent: "line1\nline2\nline3\n",
	}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	diff := "@@ -1,2 +1,2 @@\n line1\n-line2\n+line2 changed\n"
	args := json.RawMessage(fmt.Sprintf(
		`{"action":"edit_file","path":"/work/test.py","diff":%s,"message":"m"}`,
		mustJSONString(t, diff),
	))
	result, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected edit_file to succeed, got error: %v", err)
	}
	if !mock.copyFromCalled || !mock.copyToCalled {
		t.Fatalf("expected both CopyFromContainer (read) and CopyToContainer (write), got from=%v to=%v",
			mock.copyFromCalled, mock.copyToCalled)
	}
	want := "line1\nline2 changed\nline3\n"
	if mock.writtenContent != want {
		t.Errorf("written content = %q, want %q", mock.writtenContent, want)
	}
	if !strings.Contains(result, "1 diff hunk") {
		t.Errorf("result = %q, want it to mention the number of hunks applied", result)
	}
}

// TestTerminalHandle_FileAction_DefaultsToEditFile_WhenDiffPresent mirrors
// the write_file/read_file action-inference tests: when 'action' is omitted
// but 'diff' is present, the intent is unambiguous.
func TestTerminalHandle_FileAction_DefaultsToEditFile_WhenDiffPresent(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		isRunning:       true,
		readFileContent: "line1\nline2\n",
	}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	diff := "@@ -2,1 +2,1 @@\n-line2\n+line2 changed\n"
	args := json.RawMessage(fmt.Sprintf(
		`{"path":"/work/test.py","diff":%s,"message":"m"}`,
		mustJSONString(t, diff),
	))
	_, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected inferred edit_file to succeed, got error: %v", err)
	}
	if !mock.copyFromCalled || !mock.copyToCalled {
		t.Fatalf("expected inferred edit_file to read then write, got from=%v to=%v",
			mock.copyFromCalled, mock.copyToCalled)
	}
}

// TestTerminalHandle_FileAction_EditFile_NoMatch_LeavesFileUntouched checks
// that a diff whose context doesn't match the current content fails clearly
// and never reaches CopyToContainer - a bad edit must not corrupt the file.
func TestTerminalHandle_FileAction_EditFile_NoMatch_LeavesFileUntouched(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		isRunning:       true,
		readFileContent: "line1\nline2\nline3\n",
	}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	diff := "@@ -2,1 +2,1 @@\n-this text is not in the file\n+replacement\n"
	args := json.RawMessage(fmt.Sprintf(
		`{"action":"edit_file","path":"/work/test.py","diff":%s,"message":"m"}`,
		mustJSONString(t, diff),
	))
	result, err := term.Handle(t.Context(), FileToolName, args)

	// Handle() wraps tool errors into a successful-looking result string (see
	// wrapCommandResult) rather than a Go error, so assert on the message.
	if err != nil {
		t.Fatalf("Handle() should swallow tool errors via wrapCommandResult, got error: %v", err)
	}
	if !strings.Contains(result, "could not be applied") {
		t.Errorf("result = %q, want it to mention the hunk could not be applied", result)
	}
	if mock.copyToCalled {
		t.Fatal("expected CopyToContainer to NOT be called when the diff doesn't apply")
	}
}

// TestTerminalHandle_FileAction_EmptyDiff_ReturnsClearError checks edit_file
// with an empty diff fails with a clear message instead of a Docker-level error.
func TestTerminalHandle_FileAction_EmptyDiff_ReturnsClearError(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	args := json.RawMessage(`{"action":"edit_file","path":"/work/test.py","diff":"","message":"m"}`)
	result, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("Handle() should swallow tool errors via wrapCommandResult, got error: %v", err)
	}
	if !strings.Contains(result, "diff is required") {
		t.Errorf("result = %q, want it to mention that diff is required", result)
	}
	if mock.copyFromCalled || mock.copyToCalled {
		t.Fatal("expected neither CopyFromContainer nor CopyToContainer to be called for an empty diff")
	}
}

// mustJSONString marshals s as a JSON string literal, for embedding
// multi-line diff text into a hand-written JSON args payload in tests.
func mustJSONString(t *testing.T, s string) string {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal(%q) error: %v", s, err)
	}
	return string(b)
}

func TestTerminalHandle_FileAction_DefaultsToReadFile_WhenContentAbsent(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	args := json.RawMessage(`{"path":"/work/test.py","message":"m"}`)
	_, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected inferred read_file to succeed, got error: %v", err)
	}
	if !mock.copyFromCalled || mock.copyToCalled {
		t.Fatalf("expected CopyFromContainer (read_file) to be called, copyFrom=%v copyTo=%v",
			mock.copyFromCalled, mock.copyToCalled)
	}
}

func TestTerminalHandle_FileAction_EmptyPath_ReadFile_ReturnsClearError(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	args := json.RawMessage(`{"action":"read_file","path":"","message":"m"}`)
	// Handle() wraps ReadFile/WriteFile errors into a soft (nil-error) response
	// via wrapCommandResult, same as every other terminal-tool failure - so the
	// error text is asserted on the returned string, not a returned Go error.
	result, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected soft-failed (nil error) response, got error: %v", err)
	}
	if !strings.Contains(result, "path is required and cannot be empty") {
		t.Fatalf("expected result to mention the empty-path error, got: %q", result)
	}
	if mock.copyFromCalled {
		t.Fatal("expected CopyFromContainer to not be called for an empty path")
	}
}

func TestTerminalHandle_FileAction_EmptyPath_WriteFile_ReturnsClearError(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	args := json.RawMessage(`{"action":"write_file","path":"","content":"data","message":"m"}`)
	result, err := term.Handle(t.Context(), FileToolName, args)

	if err != nil {
		t.Fatalf("expected soft-failed (nil error) response, got error: %v", err)
	}
	if !strings.Contains(result, "path is required and cannot be empty") {
		t.Fatalf("expected result to mention the empty-path error, got: %q", result)
	}
	if mock.copyToCalled {
		t.Fatal("expected CopyToContainer to not be called for an empty path")
	}
}

func TestTerminalHandle_FileAction_ExplicitInvalidAction_StillFails(t *testing.T) {
	mock := &contextAwareMockDockerClient{isRunning: true}
	term := &terminal{
		flowID:       1,
		containerID:  1,
		containerLID: "test-container",
		dockerClient: mock,
		tlp:          &contextTestTermLogProvider{},
	}

	// An explicit but invalid action must still be a hard failure - inference
	// only kicks in when the field is empty.
	args := json.RawMessage(`{"path":"/work/test.py","action":"delete_file","message":"m"}`)
	_, err := term.Handle(t.Context(), FileToolName, args)

	if err == nil || !strings.Contains(err.Error(), "unknown file action") {
		t.Fatalf("expected unknown file action error, got: %v", err)
	}
	if mock.copyFromCalled || mock.copyToCalled {
		t.Fatalf("expected no docker calls for an invalid explicit action")
	}
}

func TestPrimaryTerminalName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		flowID int64
		want   string
	}{
		{1, PrimaryTerminalNamePrefix + "1"},
		{0, PrimaryTerminalNamePrefix + "0"},
		{12345, PrimaryTerminalNamePrefix + "12345"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("flowID=%d", tt.flowID), func(t *testing.T) {
			t.Parallel()

			if got := PrimaryTerminalName("", tt.flowID); got != tt.want {
				t.Errorf("PrimaryTerminalName(%d) = %q, want %q", tt.flowID, got, tt.want)
			}
		})
	}
}

func TestWriteUploadsTar(t *testing.T) {
	uploadDir := t.TempDir()
	requireNoError := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	requireNoError(os.WriteFile(filepath.Join(uploadDir, "a.txt"), []byte("alpha"), 0644))
	requireNoError(os.Mkdir(filepath.Join(uploadDir, "sub"), 0755))
	requireNoError(os.WriteFile(filepath.Join(uploadDir, "sub", "b.txt"), []byte("bravo"), 0644))
	if err := os.Symlink(filepath.Join(uploadDir, "a.txt"), filepath.Join(uploadDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- flowfiles.WriteUploadsTar(pw, uploadDir)
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, pr)
	assert.NoError(t, err)
	assert.NoError(t, <-errCh)

	tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
	contents := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		assert.NoError(t, err)
		contents[hdr.Name] = string(data)
	}

	assert.Equal(t, "alpha", contents["uploads/a.txt"])
	assert.Equal(t, "bravo", contents["uploads/sub/b.txt"])
	assert.NotContains(t, contents, "uploads/link.txt")
}

func TestCollectFileSyncEntries(t *testing.T) {
	localDir := t.TempDir()
	requireNoError := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	requireNoError(os.MkdirAll(filepath.Join(localDir, "targets"), 0755))
	requireNoError(os.WriteFile(filepath.Join(localDir, "targets", "ips.txt"), []byte("127.0.0.1"), 0644))
	requireNoError(os.WriteFile(filepath.Join(localDir, "top.txt"), []byte("top"), 0644))
	if err := os.Symlink(filepath.Join(localDir, "top.txt"), filepath.Join(localDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	entries, err := collectFileSyncEntries(localDir, flowfiles.UploadsDirName)
	assert.NoError(t, err)

	byTarPath := map[string]fileSyncEntry{}
	for _, entry := range entries {
		byTarPath[entry.tarPath] = entry
	}

	assert.Contains(t, byTarPath, "uploads/top.txt")
	assert.Contains(t, byTarPath, "uploads/targets/ips.txt")
	assert.NotContains(t, byTarPath, "uploads/link.txt")
	assert.Equal(t, docker.WorkFolderPathInContainer+"/uploads/top.txt", byTarPath["uploads/top.txt"].containerPath)
	assert.Equal(t, filepath.Join(localDir, "top.txt"), byTarPath["uploads/top.txt"].localPath)
}

func TestCollectFileSyncEntriesMissingDir(t *testing.T) {
	entries, err := collectFileSyncEntries(filepath.Join(t.TempDir(), "missing"), flowfiles.ResourcesDirName)
	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestConvertSyncEntriesToTarEntries(t *testing.T) {
	entries := []fileSyncEntry{
		{localPath: "/tmp/a.txt", tarPath: "uploads/a.txt"},
		{localPath: "/tmp/b.txt", tarPath: "resources/b.txt"},
	}

	tarEntries := convertSyncEntriesToTarEntries(entries)

	assert.Equal(t, []flowfiles.TarEntry{
		{LocalPath: "/tmp/a.txt", TarPath: "uploads/a.txt"},
		{LocalPath: "/tmp/b.txt", TarPath: "resources/b.txt"},
	}, tarEntries)
}

func TestFindMissingInContainer(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		execCreateResp: client.ExecCreateResult{ID: "exec-file-check"},
		attachOutput: []byte(
			docker.WorkFolderPathInContainer + "/uploads/a.txt\n" +
				docker.WorkFolderPathInContainer + "/resources/b.txt\n" +
				docker.WorkFolderPathInContainer + "/unknown.txt\n",
		),
		inspectResp: client.ExecInspectResult{ExitCode: 0},
	}
	fte := &flowToolsExecutor{flowID: 7, docker: mock}
	entries := []fileSyncEntry{
		{localPath: "/tmp/a.txt", containerPath: docker.WorkFolderPathInContainer + "/uploads/a.txt", tarPath: "uploads/a.txt"},
		{localPath: "/tmp/b.txt", containerPath: docker.WorkFolderPathInContainer + "/resources/b.txt", tarPath: "resources/b.txt"},
	}

	missing, err := fte.findMissingInContainer(t.Context(), entries)

	assert.NoError(t, err)
	assert.Equal(t, entries, missing)
}

func TestFindMissingInContainerChecksExitCode(t *testing.T) {
	mock := &contextAwareMockDockerClient{
		execCreateResp: client.ExecCreateResult{ID: "exec-file-check"},
		attachOutput:   []byte("shell failed"),
		inspectResp:    client.ExecInspectResult{ExitCode: 2},
	}
	fte := &flowToolsExecutor{flowID: 7, docker: mock}

	_, err := fte.findMissingInContainer(t.Context(), []fileSyncEntry{
		{containerPath: docker.WorkFolderPathInContainer + "/uploads/a.txt"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit code 2")
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
