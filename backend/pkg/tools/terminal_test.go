package tools

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each refusal is a hard error, and none of them reaches Docker.
func TestTerminal_Handle_RejectsACallItCannotDispatch(t *testing.T) {
	tests := []struct {
		name     string
		noDocker bool
		tool     string
		args     string
		want     string
	}{
		{
			name:     "a terminal without a docker client",
			noDocker: true,
			tool:     TerminalToolName,
			args:     `{"input":"id","cwd":"/work","detach":false,"timeout":60,"message":"m"}`,
			want:     "terminal is not available",
		},
		{
			name: "terminal arguments that are not JSON",
			tool: TerminalToolName,
			args: `{"input":`,
			want: "failed to unmarshal terminal action",
		},
		{
			name: "file arguments that are not JSON",
			tool: FileToolName,
			args: `{"path":`,
			want: "failed to unmarshal file action",
		},
		{
			// Inference fills only an absent action; a wrong one the model named stays a hard failure.
			name: "an explicit action it does not know",
			tool: FileToolName,
			args: `{"path":"/work/test.py","action":"delete_file","message":"m"}`,
			want: `unknown file action "delete_file": expected one of read_file, write_file or edit_file`,
		},
		{
			name: "a tool it does not serve",
			tool: BrowserToolName,
			args: `{}`,
			want: "unknown tool: browser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: true}
			term := terminalFor(mock, &recordingTermLog{})
			if tt.noDocker {
				term.dockerClient = nil
			}

			result, err := term.Handle(t.Context(), tt.tool, json.RawMessage(tt.args))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
			assert.Empty(t, result)
			if tt.noDocker {
				// This terminal does not hold the mock, so the mock cannot see a call.
				return
			}
			assert.Empty(t, mock.execCreated.Cmd, "no command may run")
			assert.False(t, mock.copyFromCalled, "no file may be read")
			assert.Zero(t, mock.copies, "no file may be written")
		})
	}
}

func TestTerminal_Handle_InfersTheFileActionFromItsFields(t *testing.T) {
	const current = "line1\nline2\n"
	diff := "@@ -2,1 +2,1 @@\n-line2\n+line2 changed\n"

	tests := []struct {
		name        string
		args        map[string]any
		wantRead    bool
		wantWritten string
		wantResult  string
	}{
		{
			name:        "content is a write",
			args:        map[string]any{"path": "/work/test.py", "content": "print(1)", "message": "m"},
			wantWritten: "print(1)",
			wantResult:  "Successfully wrote 8 bytes to /work/test.py",
		},
		{
			name:        "a diff is an edit",
			args:        map[string]any{"path": "/work/test.py", "diff": diff, "message": "m"},
			wantRead:    true,
			wantWritten: "line1\nline2 changed\n",
			wantResult:  "Applied 1 diff hunk(s) to /work/test.py (12 -> 20 bytes)",
		},
		{
			name:        "a diff beside content is an edit",
			args:        map[string]any{"path": "/work/test.py", "content": "print(1)", "diff": diff, "message": "m"},
			wantRead:    true,
			wantWritten: "line1\nline2 changed\n",
			wantResult:  "Applied 1 diff hunk(s) to /work/test.py (12 -> 20 bytes)",
		},
		{
			name:       "neither is a read",
			args:       map[string]any{"path": "/work/test.py", "message": "m"},
			wantRead:   true,
			wantResult: current,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: true, readFileContent: current}

			result, err := terminalFor(mock, &recordingTermLog{}).
				Handle(t.Context(), FileToolName, mustJSON(tt.args))

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
			assert.Equal(t, tt.wantRead, mock.copyFromCalled, "whether the file was read")
			assert.Equal(t, tt.wantWritten != "", mock.copies > 0, "whether the file was written")
			assert.Equal(t, tt.wantWritten, mock.writtenContent)
		})
	}
}

func TestTerminal_Handle_DispatchesAnExtraQuotedFileAction(t *testing.T) {
	mock := &fakeDockerClient{isRunning: true}

	args := json.RawMessage(`{"action": "\"write_file\"", "path": "\"/home/evidence/inject.py\"", "content": "print(1)", "message": "m"}`)
	result, err := terminalFor(mock, &recordingTermLog{}).Handle(t.Context(), FileToolName, args)

	require.NoError(t, err)
	assert.Equal(t, "Successfully wrote 8 bytes to /home/evidence/inject.py", result)
	assert.Equal(t, 1, mock.copies, "an extra-quoted write_file must still write")
	assert.False(t, mock.copyFromCalled)
	assert.Equal(t, "print(1)", mock.writtenContent)
}

func TestTerminal_Handle_RequiresAFilePath(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{"a read", map[string]any{"action": "read_file", "path": "", "message": "m"}},
		{"a write", map[string]any{"action": "write_file", "path": "", "content": "data", "message": "m"}},
		{"an edit", map[string]any{"action": "edit_file", "path": "", "diff": "@@ -1 +1 @@\n-a\n+b\n", "message": "m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: true}

			result, err := terminalFor(mock, &recordingTermLog{}).
				Handle(t.Context(), FileToolName, mustJSON(tt.args))

			require.NoError(t, err, "a tool failure is a soft (nil-error) response")
			assert.Contains(t, result, "path is required and cannot be empty")
			assert.False(t, mock.copyFromCalled, "an empty path must not be read")
			assert.Zero(t, mock.copies, "an empty path must not be written")
		})
	}
}

func TestTerminal_ReadFile_ReturnsTheWholeContent(t *testing.T) {
	big := terminalLargeFileContent(2000)
	other := terminalLargeFileContent(1500)
	require.Greater(t, len(other), dockerReadChunk, "every file must exceed one read")

	const rule = "--------------------------------------------------\n"
	tests := []struct {
		name    string
		archive []byte
		dir     bool
		want    string
	}{
		{
			name:    "a file larger than one socket read",
			archive: terminalTar(t, terminalTarEntry{name: "big.txt", body: big}),
			want:    big,
		},
		{
			name: "a directory of files larger than one socket read",
			archive: terminalTar(t,
				terminalTarEntry{name: "recon/", dir: true},
				terminalTarEntry{name: "recon/a.txt", body: big},
				terminalTarEntry{name: "recon/sub/", dir: true},
				terminalTarEntry{name: "recon/sub/b.txt", body: other},
			),
			dir: true,
			want: rule + fmt.Sprintf("'recon/a.txt' file content (with size %d bytes) shown below:\n", len(big)) +
				big + "\n\n" +
				rule + fmt.Sprintf("'recon/sub/b.txt' file content (with size %d bytes) shown below:\n", len(other)) +
				other + "\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &fakeDockerClient{isRunning: true, archive: tt.archive}
			if tt.dir {
				dc.stat = container.PathStat{Name: "recon", Mode: os.ModeDir | 0o755}
			}

			result, err := terminalFor(dc, &recordingTermLog{}).Handle(t.Context(),
				FileToolName, json.RawMessage(`{"action":"read_file","path":"/work/recon","message":"m"}`))

			require.NoError(t, err)
			if n := strings.Count(result, "\x00"); n != 0 {
				t.Errorf("read returned %d NUL bytes; the content past the first read was not read at all", n)
			}
			if result != tt.want {
				t.Errorf("read returned %d bytes, want %d:\n%.300q", len(result), len(tt.want), result)
			}
		})
	}
}

func TestTerminal_ReadFile_RefusesWhatItCannotReadWhole(t *testing.T) {
	var huge bytes.Buffer
	require.NoError(t, tar.NewWriter(&huge).WriteHeader(&tar.Header{Name: "huge.bin", Mode: 0o600, Size: 100<<20 + 1}))
	var negative bytes.Buffer
	require.NoError(t, tar.NewWriter(&negative).WriteHeader(&tar.Header{
		Name: "a.txt", Linkname: "target", Mode: 0o777, Typeflag: tar.TypeSymlink, Size: -1,
	}))

	tests := []struct {
		name    string
		copyErr error
		archive []byte
		want    string
	}{
		{
			name:    "a path the sandbox does not have",
			copyErr: errors.New("Could not find the file /work/a.txt in container"),
			want:    "failed to copy file: Could not find the file /work/a.txt in container",
		},
		{
			name:    "a file above the read limit",
			archive: huge.Bytes(),
			want:    "file 'huge.bin' size 104857601 exceeds maximum allowed size 104857600",
		},
		{
			name:    "an archive that ends inside the file",
			archive: terminalShortArchive(t, "a.txt", "12345", 5),
			want:    "failed to read file 'a.txt' content: unexpected EOF",
		},
		{
			name:    "an archive that is not a tar",
			archive: bytes.Repeat([]byte("x"), 512),
			want:    "failed to read tar header: archive/tar: invalid tar header",
		},
		{
			// archive/tar passes a negative size through on a header-only entry, and copying -1 bytes is a no-op.
			name:    "a link entry that declares a negative size",
			archive: negative.Bytes(),
			want:    "file 'a.txt' has invalid size -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &fakeDockerClient{isRunning: true, copyFromErr: tt.copyErr, archive: tt.archive}

			result, err := terminalFor(dc, &recordingTermLog{}).Handle(t.Context(),
				FileToolName, json.RawMessage(`{"action":"read_file","path":"/work/a.txt","message":"m"}`))

			require.NoError(t, err, "a tool failure is a soft (nil-error) response")
			assert.Equal(t, "terminal tool 'file' handled with error: "+tt.want, result)
		})
	}
}

// validateFilePath lets both through, so the archive refuses them before anything is copied.
func TestTerminal_WriteFile_RefusesANameTheArchiveCannotCarry(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "a name holding a NUL byte",
			path: "/work/a\x00b.txt",
			want: "tar archive header generation failed: archive/tar: cannot encode header",
		},
		{
			// The entry for "/" is a directory, which holds no content.
			name: "the root directory",
			path: "/",
			want: "tar archive content serialization failed: archive/tar: write too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: true}

			result, err := terminalFor(mock, &recordingTermLog{}).Handle(t.Context(), FileToolName,
				mustJSON(map[string]any{"action": "write_file", "path": tt.path, "content": "print(1)", "message": "m"}))

			require.NoError(t, err, "a tool failure is a soft (nil-error) response")
			assert.Contains(t, result, "terminal tool 'file' handled with error: "+tt.want)
			assert.Zero(t, mock.copies, "nothing may be copied into the sandbox")
		})
	}
}

func TestTerminal_EditFile_AppliesTheDiff(t *testing.T) {
	big := terminalLargeFileContent(2000)
	const firstLine = "line 00000: the quick brown fox jumps over the lazy dog"

	tests := []struct {
		name    string
		current string
		diff    string
		want    string
	}{
		{
			name:    "a small file",
			current: "line1\nline2\nline3\n",
			diff:    "@@ -1,2 +1,2 @@\n line1\n-line2\n+line2 changed\n",
			want:    "line1\nline2 changed\nline3\n",
		},
		{
			name:    "a file larger than one socket read",
			current: big,
			diff: "@@ -1,2 +1,2 @@\n-" + firstLine + "\n+line 00000: patched\n" +
				" line 00001: the quick brown fox jumps over the lazy dog\n",
			want: strings.Replace(big, firstLine, "line 00000: patched", 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: true, readFileContent: tt.current}

			result, err := terminalFor(mock, &recordingTermLog{}).Handle(t.Context(), FileToolName,
				mustJSON(map[string]any{"action": "edit_file", "path": "/work/test.py", "diff": tt.diff, "message": "m"}))

			require.NoError(t, err)
			assert.True(t, mock.copyFromCalled, "edit_file must read the current content")
			assert.Equal(t, 1, mock.copies, "edit_file must write the result back")
			if n := strings.Count(mock.writtenContent, "\x00"); n != 0 {
				t.Errorf("edit_file wrote %d NUL bytes back to the file, destroying what it never read", n)
			}
			if mock.writtenContent != tt.want {
				t.Errorf("edit_file wrote %d bytes, want %d; only the patched line may differ",
					len(mock.writtenContent), len(tt.want))
			}
			assert.Contains(t, result, "Applied 1 diff hunk(s)", "the result must say how many hunks applied")
		})
	}
}

func TestTerminal_EditFile_LeavesTheFileUntouchedOnFailure(t *testing.T) {
	const applies = "@@ -2,1 +2,1 @@\n-line2\n+line2 changed\n"

	tests := []struct {
		name     string
		diff     string
		prepare  func(*fakeDockerClient)
		wantRead bool
		want     string
	}{
		{
			name:     "a diff whose context is not in the file",
			diff:     "@@ -2,1 +2,1 @@\n-this text is not in the file\n+replacement\n",
			wantRead: true,
			want:     "could not be applied",
		},
		{
			name: "an empty diff",
			diff: "",
			want: "diff is required",
		},
		{
			name: "a read that ends inside the file",
			diff: applies,
			prepare: func(d *fakeDockerClient) {
				// What arrives holds the diff's context, so only the read error prevents a truncated write-back.
				d.archive = terminalShortArchive(t, "test.py", "line1\nline2\n", len("line3\n"))
			},
			wantRead: true,
			want:     "failed to read current content of /work/test.py before editing",
		},
		{
			name:     "a sandbox that stops between the read and the write",
			diff:     applies,
			prepare:  func(d *fakeDockerClient) { d.running = terminalRunningOnlyAtFirst },
			wantRead: true,
			want:     "failed to write edited content of /work/test.py: container runtime is not operational",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: true, readFileContent: "line1\nline2\nline3\n"}
			if tt.prepare != nil {
				tt.prepare(mock)
			}

			result, err := terminalFor(mock, &recordingTermLog{}).Handle(t.Context(), FileToolName,
				mustJSON(map[string]any{"action": "edit_file", "path": "/work/test.py", "diff": tt.diff, "message": "m"}))

			require.NoError(t, err, "a tool failure is a soft (nil-error) response")
			assert.Contains(t, result, tt.want)
			assert.Equal(t, tt.wantRead, mock.copyFromCalled, "whether the file was read")
			assert.Zero(t, mock.copies, "the file must be left untouched")
		})
	}
}

func TestTerminal_ExecCommand_DetachedSurvivesParentCancel(t *testing.T) {
	mock := &fakeDockerClient{
		isRunning:      true,
		execCreateResp: client.ExecCreateResult{ID: "exec-cancel-test"},
		attachOutput:   []byte("background result"),
		// Outlasts the quick check, so the call returns while the command runs.
		attachDelay: 2 * time.Second,
		inspectResp: client.ExecInspectResult{ExitCode: 0},
	}
	term := terminalFor(mock, &recordingTermLog{})
	parentCtx, cancel := context.WithCancel(t.Context())

	output, err := term.ExecCommand(parentCtx, "/work", "long-running-scan", true, 5*time.Minute)
	require.NoError(t, err)
	assert.Contains(t, output, "Command started in background")

	cancel()
	// Outlasts attachDelay, so by the check the watcher has finished or seen the cancel.
	time.Sleep(3 * time.Second)

	assert.False(t, mock.ctxWasCanceled,
		"the detached command saw its caller's cancellation and was stopped")
}

func TestTerminal_ExecCommand_NonDetachedStopsOnParentCancel(t *testing.T) {
	mock := &fakeDockerClient{
		isRunning:      true,
		execCreateResp: client.ExecCreateResult{ID: "exec-nondetach-cancel"},
		attachOutput:   []byte("should not complete"),
		// Far longer than the cancel below, so only the cancel can end it early.
		attachDelay: 5 * time.Second,
		inspectResp: client.ExecInspectResult{ExitCode: 0},
	}
	term := terminalFor(mock, &recordingTermLog{})
	parentCtx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(200*time.Millisecond, cancel)

	_, err := term.ExecCommand(parentCtx, "/work", "long-command", false, 5*time.Minute)

	require.ErrorIs(t, err, context.Canceled, "the call must end because its caller cancelled it")
	assert.True(t, mock.ctxWasCanceled, "the foreground command ignored its caller's cancellation")
}

func TestTerminal_ExecCommand_LaunchesOnlyAFlowTaskCommandForStopToEnd(t *testing.T) {
	taskID := int64(7)

	tests := map[string]struct {
		taskID  *int64
		wantCmd []string
	}{
		"a flow task's command":  {taskID: &taskID, wantCmd: docker.FlowCommand("nmap -sV target")},
		"an assistant's command": {wantCmd: docker.SandboxCommand("nmap -sV target")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mock := &fakeDockerClient{
				isRunning:      true,
				execCreateResp: client.ExecCreateResult{ID: "exec-owner-test"},
				inspectResp:    client.ExecInspectResult{ExitCode: 0},
			}
			term := terminalFor(mock, &recordingTermLog{})
			term.taskID = tc.taskID

			_, err := term.ExecCommand(t.Context(), "/work", "nmap -sV target", false, time.Minute)
			require.NoError(t, err)

			assert.Equal(t, tc.wantCmd, mock.execCreated.Cmd)
		})
	}
}

func TestTerminal_ExecCommand_ReportsTheExitCodeItEndedWith(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		exitCode int
		want     string
	}{
		{"a command that fails without output", "", 1, "[no output]\n[exit code: 1]"},
		{"a failing command keeps its output", "grep: /x: No such file or directory\r\n", 2,
			"grep: /x: No such file or directory\r\n\n[exit code: 2]"},
		{"a command that succeeds", "uid=0(root)\r\n", 0, "uid=0(root)\r\n\n[exit code: 0]"},
		{"a command that succeeds silently", "", 0, "[no output]\n[exit code: 0]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlp := &recordingTermLog{}
			term := terminalFor(&fakeDockerClient{
				isRunning:      true,
				execCreateResp: client.ExecCreateResult{ID: "exec-status"},
				attachOutput:   []byte(tt.output),
				inspectResp:    client.ExecInspectResult{ExitCode: tt.exitCode},
			}, tlp)

			result, err := term.Handle(t.Context(), TerminalToolName,
				json.RawMessage(`{"input":"cmd","cwd":"/work","detach":false,"timeout":60,"message":"m"}`))

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
			assert.Empty(t, tlp.states(termLogNotRunning), "a sandbox that is still running is not reported")
		})
	}
}

func TestTerminal_ExecCommand_ADetachedCommandThatEndsAtOnceReportsHowItEnded(t *testing.T) {
	tests := []struct {
		name      string
		attachErr error
		want      string
		wantErr   string
	}{
		{
			name: "a command that exits at once",
			want: "[no output]\n[exit code: 1]",
		},
		{
			name:      "a command Docker cannot attach to",
			attachErr: errors.New("exec 1a2b is not running"),
			wantErr:   "command failed: failed to attach to exec process: exec 1a2b is not running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := terminalFor(&fakeDockerClient{
				isRunning:      true,
				execCreateResp: client.ExecCreateResult{ID: "exec-detached"},
				inspectResp:    client.ExecInspectResult{ExitCode: 1},
				attachErr:      tt.attachErr,
			}, &recordingTermLog{})

			result, err := term.ExecCommand(t.Context(), "/work", "nc -lvnp 4444", true, time.Minute)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Only an exit code above 128, a death by signal, makes the terminal probe the sandbox a second time.
func TestTerminal_ExecCommand_BlamesTheSandboxOnlyWhenItStopped(t *testing.T) {
	tests := []struct {
		name         string
		exitCode     int
		running      func(probe int) (bool, error)
		want         string
		wantProbes   int
		wantReported []int64
	}{
		{
			name:         "a signal death while the sandbox stopped",
			exitCode:     137,
			running:      terminalRunningOnlyAtFirst,
			want:         "[no output]\n[exit code: 137]\n[interrupted: the sandbox container stopped]",
			wantProbes:   2,
			wantReported: []int64{7},
		},
		{
			name:       "a signal death inside a running sandbox",
			exitCode:   137,
			running:    terminalAlwaysRunning,
			want:       "[no output]\n[exit code: 137]",
			wantProbes: 2,
		},
		{
			name:       "an ordinary failure",
			exitCode:   1,
			running:    terminalAlwaysRunning,
			want:       "[no output]\n[exit code: 1]",
			wantProbes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &fakeDockerClient{
				execCreateResp: client.ExecCreateResult{ID: "exec-killed"},
				inspectResp:    client.ExecInspectResult{ExitCode: tt.exitCode},
				running:        tt.running,
			}
			tlp := &recordingTermLog{}

			result, err := terminalFor(dc, tlp).ExecCommand(t.Context(), "/work", "sleep 90 && echo done", false, time.Minute)

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantProbes, len(dc.probed), "only a signal death asks Docker a second time")
			assert.Equal(t, tt.wantReported, terminalRecorded(tlp, termLogNotRunning))
		})
	}
}

// The log write runs after the call's context fired, and recordingTermLog refuses a done context.
func TestTerminal_ExecCommand_TimeoutKeepsWhatTheCommandPrinted(t *testing.T) {
	partial := strings.Repeat("scanning host 10.0.0.1 ... open\n", 100)
	require.Greater(t, len(partial), 500, "the fixture must be long enough for a 500-byte cut to show")

	tests := []struct {
		name        string
		timeout     time.Duration
		cancelAfter time.Duration
		refuseLog   bool
		wantCause   string
	}{
		{
			name:      "the deadline passes",
			timeout:   time.Second,
			wantCause: "context deadline exceeded",
		},
		{
			name:        "the parent is cancelled",
			timeout:     time.Minute,
			cancelAfter: 200 * time.Millisecond,
			wantCause:   "context canceled",
		},
		{
			name:      "the terminal log refuses the write",
			timeout:   time.Second,
			refuseLog: true,
			wantCause: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{
				isRunning:      true,
				execCreateResp: client.ExecCreateResult{ID: "exec-timeout"},
				attachOutput:   []byte(partial),
				attachHold:     5 * time.Second,
				inspectResp:    client.ExecInspectResult{ExitCode: 0},
			}
			tlp := &recordingTermLog{}
			if tt.refuseLog {
				tlp.failOn = database.TermlogTypeStdout
			}
			ctx := t.Context()
			if tt.cancelAfter > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				time.AfterFunc(tt.cancelAfter, cancel)
			}

			_, err := terminalFor(mock, tlp).ExecCommand(ctx, "/work", "nmap -sV target", false, tt.timeout)

			require.Error(t, err, "a command that outlives its budget must report the timeout")
			assert.Contains(t, err.Error(), "command execution timeout ("+tt.wantCause+")")
			assert.Contains(t, err.Error(), partial,
				"the whole partial output must reach the agent, not the first 500 bytes")
			if tt.refuseLog {
				assert.NotContains(t, tlp.written(), partial, "the terminal log refused the output")
			} else {
				assert.Contains(t, tlp.written(), partial,
					"the partial output must also reach the terminal log the user is watching")
			}
		})
	}
}

func TestTerminal_RequireRunningContainer_ReportsASandboxFoundNotRunning(t *testing.T) {
	calls := map[string]func(context.Context, *terminal) error{
		"a command": func(ctx context.Context, term *terminal) error {
			_, err := term.ExecCommand(ctx, "/work", "id", false, time.Minute)
			return err
		},
		"a file read": func(ctx context.Context, term *terminal) error {
			_, err := term.ReadFile(ctx, 1, "/work/a")
			return err
		},
		"a file write": func(ctx context.Context, term *terminal) error {
			_, err := term.WriteFile(ctx, 1, "x", "/work/a")
			return err
		},
		"a file edit": func(ctx context.Context, term *terminal) error {
			_, err := term.EditFile(ctx, 1, "/work/a", "@@ -1 +1 @@\n-a\n+b\n")
			return err
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			mock := &fakeDockerClient{isRunning: false}
			tlp := &recordingTermLog{}
			subtaskID := int64(12)
			term := terminalFor(mock, tlp)
			term.subtaskID = &subtaskID
			term.containerLID = "dead"

			err := call(t.Context(), term)

			require.ErrorIs(t, err, errContainerNotOperational)
			assert.Equal(t, []termLogEntry{{state: termLogNotRunning, containerID: 7, subtaskID: &subtaskID}},
				tlp.states(termLogNotRunning), "recorded once, against the sandbox's row and the subtask that found it")
			assert.Empty(t, mock.execCreated.Cmd)
			assert.False(t, mock.copyFromCalled)
			assert.Zero(t, mock.copies)
		})
	}
}

func TestTerminal_RequireRunningContainer_ReportsASandboxFoundRunning(t *testing.T) {
	tlp := &recordingTermLog{}
	term := terminalFor(&fakeDockerClient{isRunning: true}, tlp)

	_, err := term.WriteFile(t.Context(), 1, "x", "/work/a")

	require.NoError(t, err)
	assert.Equal(t, []int64{7}, terminalRecorded(tlp, termLogRunning))
	assert.Empty(t, tlp.states(termLogNotRunning))
}

func TestTerminal_RequireRunningContainer_FailsWhenDockerCannotBeAsked(t *testing.T) {
	daemonDown := errors.New("Cannot connect to the Docker daemon at unix:///var/run/docker.sock")
	mock := &fakeDockerClient{isRunning: true, probeErr: daemonDown}
	tlp := &recordingTermLog{}
	term := terminalFor(mock, tlp)

	_, err := term.ExecCommand(t.Context(), "/work", "id", false, time.Minute)

	require.ErrorIs(t, err, daemonDown)
	assert.Contains(t, err.Error(), "runtime verification failed")
	assert.NotErrorIs(t, err, errContainerNotOperational)
	assert.Empty(t, tlp.states(termLogNotRunning), "an unanswered probe must not mark the sandbox stopped")
	assert.Empty(t, tlp.states(termLogRunning), "an unanswered probe must not mark the sandbox running")
	assert.Empty(t, mock.execCreated.Cmd, "no command may run")
}

// A failed recording leaves the call's answer unchanged.
func TestTerminal_RequireRunningContainer_WarnsOfAStateItCannotRecord(t *testing.T) {
	hook := captureLogrus(t)

	recordErr := errors.New("connection reset by peer")
	tests := []struct {
		name        string
		running     bool
		wantErr     error
		wantWarning string
	}{
		{
			name:        "a sandbox found running",
			running:     true,
			wantWarning: "failed to record the sandbox as running again",
		},
		{
			name:        "a sandbox found stopped",
			running:     false,
			wantErr:     errContainerNotOperational,
			wantWarning: "failed to record the sandbox as not running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()
			term := terminalFor(&fakeDockerClient{
				isRunning:      tt.running,
				execCreateResp: client.ExecCreateResult{ID: "exec-state"},
				inspectResp:    client.ExecInspectResult{ExitCode: 0},
			}, &recordingTermLog{stateErr: recordErr})

			_, err := term.ExecCommand(t.Context(), "/work", "id", false, time.Minute)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
				assert.NotErrorIs(t, err, recordErr)
			}
			var warning *logrus.Entry
			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel && entry.Message == tt.wantWarning {
					warning = entry
				}
			}
			require.NotNil(t, warning, "the failed recording went unreported: %v", hook.AllEntries())
			assert.Equal(t, recordErr, warning.Data[logrus.ErrorKey])
			assert.Equal(t, int64(1), warning.Data["flow_id"])
		})
	}
}

func TestTerminal_AnActionThatRanIsNotReportedFailedWhenItsLogIsNot(t *testing.T) {
	tests := []struct {
		name        string
		failOn      database.TermlogType
		call        func(context.Context, *terminal) (string, error)
		want        string
		wantWritten string
	}{
		{
			name:   "a command whose output is not logged",
			failOn: database.TermlogTypeStdout,
			call: func(ctx context.Context, term *terminal) (string, error) {
				return term.ExecCommand(ctx, "/work", "id", false, time.Minute)
			},
			want: "uid=0(root)",
		},
		{
			name:   "a written file whose write is not logged",
			failOn: database.TermlogTypeStdin,
			call: func(ctx context.Context, term *terminal) (string, error) {
				return term.WriteFile(ctx, 1, "print(1)", "/work/test.py")
			},
			want:        "Successfully wrote 8 bytes to /work/test.py",
			wantWritten: "print(1)",
		},
		{
			name:   "an edited file whose edit is not logged",
			failOn: database.TermlogTypeStdin,
			call: func(ctx context.Context, term *terminal) (string, error) {
				return term.EditFile(ctx, 1, "/work/test.py", "@@ -2,1 +2,1 @@\n-line2\n+line2 changed\n")
			},
			want:        "Applied 1 diff hunk(s)",
			wantWritten: "line1\nline2 changed\nline3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{
				isRunning:       true,
				execCreateResp:  client.ExecCreateResult{ID: "exec-unlogged"},
				attachOutput:    []byte("uid=0(root)"),
				inspectResp:     client.ExecInspectResult{ExitCode: 0},
				readFileContent: "line1\nline2\nline3\n",
			}

			result, err := tt.call(t.Context(), terminalFor(mock, &recordingTermLog{failOn: tt.failOn}))

			require.NoError(t, err, "an action that ran must not come back as a failed call")
			assert.Contains(t, result, tt.want)
			assert.Equal(t, tt.wantWritten, mock.writtenContent)
		})
	}
}

func TestTerminal_DoesNotActOnACommandItCouldNotLog(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *terminal) error
		want string
	}{
		{
			name: "a command",
			call: func(ctx context.Context, term *terminal) error {
				_, err := term.ExecCommand(ctx, "/work", "rm -rf /work/loot", false, time.Minute)
				return err
			},
			want: "failed to put terminal log (stdin): connection reset by peer",
		},
		{
			name: "a file read",
			call: func(ctx context.Context, term *terminal) error {
				_, err := term.ReadFile(ctx, 1, "/work/a")
				return err
			},
			want: "failed to put terminal log (read file cmd): connection reset by peer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fakeDockerClient{
				isRunning:       true,
				execCreateResp:  client.ExecCreateResult{ID: "exec-unlogged-cmd"},
				readFileContent: "secret",
			}

			err := tt.call(t.Context(), terminalFor(mock, &recordingTermLog{failOn: database.TermlogTypeStdin}))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
			assert.Empty(t, mock.execCreated.Cmd, "no command may run")
			assert.False(t, mock.copyFromCalled, "no file may be read")
		})
	}
}

func TestTerminal_ValidateFilePath_AcceptsOnlyALiteralPath(t *testing.T) {
	t.Parallel()

	t.Run("accepts a literal path", func(t *testing.T) {
		t.Parallel()

		for _, path := range []string{
			"/work/notes.txt",
			"/work/recon/pages/index.html",
			"/tmp/a b/c-d_e.2026.log",
			"/work/O'Brien.txt",
			"/work/отчёт.md",
			// Legal in a Linux filename, and the copy API takes them literally.
			"/work/*.log",
			"/work/report?.txt",
			"/work/a;b.txt",
			"/work/out>1.txt",
			"/work/a|b.txt",
			"/work/cost$5.txt",
			"/mnt/C$/share.txt",
		} {
			if err := validateFilePath(path); err != nil {
				t.Errorf("%q: want no error, got %v", path, err)
			}
		}
	})

	t.Run("rejects a shell expression", func(t *testing.T) {
		t.Parallel()

		for _, tt := range []struct{ path, token string }{
			{"/work/recon/pages/$(echo -n '/page' | md5sum | cut -c1-12).html", "$("},
			{"/work/${TARGET}/notes.txt", "${"},
			{"/work/`hostname`.txt", "`"},
			{"/work/$TARGET/notes.txt", "$TARGET"},
			{"$HOME/.ssh/config", "$HOME"},
		} {
			err := validateFilePath(tt.path)
			if err == nil {
				t.Errorf("%q: want an error naming %q", tt.path, tt.token)
				continue
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("contains %q", tt.token)) {
				t.Errorf("%q: the message must name %q, got %q", tt.path, tt.token, err)
			}
			if !strings.Contains(err.Error(), "literal filename") {
				t.Errorf("%q: the message must say the field is literal, got %q", tt.path, err)
			}
			if !strings.Contains(err.Error(), "terminal tool") {
				t.Errorf("%q: the message must point at the way to resolve the name, got %q", tt.path, err)
			}
		}
	})

	t.Run("rejects an absent path", func(t *testing.T) {
		t.Parallel()

		err := validateFilePath("")
		if err == nil {
			t.Fatal("want an error")
		}
		if !strings.Contains(err.Error(), "/work/") {
			t.Errorf("the message must show the shape expected, got %q", err)
		}
	})
}

func TestTerminal_ConfiguredExecTimeout_KeepsAValueUpToThreeHoursAndCapsTheRest(t *testing.T) {
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
			name:       "the shipped default (1200 s) is returned as-is",
			configured: 1200 * time.Second,
			want:       1200 * time.Second,
		},
		{
			name:       "exactly at the 3-hour ceiling is returned as-is",
			configured: 3 * time.Hour,
			want:       3 * time.Hour,
		},
		{
			name:       "zero is capped to the 3-hour ceiling",
			configured: 0,
			want:       3 * time.Hour,
		},
		{
			name:       "negative one second is capped to the 3-hour ceiling",
			configured: -1 * time.Second,
			want:       3 * time.Hour,
		},
		{
			name:       "large negative is capped to the 3-hour ceiling",
			configured: -9999 * time.Second,
			want:       3 * time.Hour,
		},
		{
			name:       "one second above the ceiling is capped",
			configured: 3*time.Hour + time.Second,
			want:       3 * time.Hour,
		},
		{
			name:       "very large value (> 3 h) is capped to the 3-hour ceiling",
			configured: 100000 * time.Second,
			want:       3 * time.Hour,
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

// The ceiling is the configured value plus a five-second grace for the exec round-trip.
func TestTerminal_NormalizeExecTimeout_KeepsOnlyARequestWithinTheCeiling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured time.Duration
		requested  time.Duration
		want       time.Duration
	}{
		{
			name:       "a request within the ceiling is kept",
			configured: 10 * time.Minute,
			requested:  45 * time.Second,
			want:       45 * time.Second,
		},
		{
			name:       "a request at the ceiling is kept",
			configured: 10 * time.Minute,
			requested:  10*time.Minute + 5*time.Second,
			want:       10*time.Minute + 5*time.Second,
		},
		{
			name:       "a request one second above the ceiling falls back to it",
			configured: 10 * time.Minute,
			requested:  10*time.Minute + 6*time.Second,
			want:       10*time.Minute + 5*time.Second,
		},
		{
			name:       "no request falls back to the ceiling",
			configured: 10 * time.Minute,
			requested:  0,
			want:       10*time.Minute + 5*time.Second,
		},
		{
			name:       "a negative request falls back to the ceiling",
			configured: 10 * time.Minute,
			requested:  -5 * time.Second,
			want:       10*time.Minute + 5*time.Second,
		},
		{
			name:       "no request on an unconfigured server gets the 3-hour ceiling",
			configured: 0,
			requested:  0,
			want:       3*time.Hour + 5*time.Second,
		},
		{
			name:       "a request above 3 hours on an unconfigured server falls back to that ceiling",
			configured: 0,
			requested:  3*time.Hour + 6*time.Second,
			want:       3*time.Hour + 5*time.Second,
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

func TestTerminal_PrimaryTerminalName_PutsTheTenantFirstAndTheFlowLast(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tenantPrefix string
		flowID       int64
		want         string
	}{
		{"a single instance", "", 1, "pentagi-terminal-1"},
		{"flow zero", "", 0, "pentagi-terminal-0"},
		{"a multi-digit flow", "", 12345, "pentagi-terminal-12345"},
		{"a tenant", "acme-", 1, "acme-pentagi-terminal-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := PrimaryTerminalName(tt.tenantPrefix, tt.flowID); got != tt.want {
				t.Errorf("PrimaryTerminalName(%q, %d) = %q, want %q", tt.tenantPrefix, tt.flowID, got, tt.want)
			}
		})
	}
}

func TestTerminal_TruncateString_CutsPastTheLimitAndSaysTheFullSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     string
		maxLen int
		want   string
	}{
		{"shorter than the limit", "abc", 5, "abc"},
		{"exactly at the limit", "abcde", 5, "abcde"},
		{"longer than the limit", "abcdefgh", 5, "abcde... [truncated full size is 8 bytes]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, truncateString(tt.in, tt.maxLen))
		})
	}
}

// terminalFor is the terminal of flow 1 whose sandbox is container row 7.
func terminalFor(dc docker.DockerClient, tlp TermLogProvider) *terminal {
	return &terminal{
		flowID:       1,
		containerID:  7,
		containerLID: "test-container",
		dockerClient: dc,
		tlp:          tlp,
	}
}

func terminalRunningOnlyAtFirst(probe int) (bool, error) { return probe == 1, nil }

func terminalAlwaysRunning(int) (bool, error) { return true, nil }

type terminalTarEntry struct {
	name string
	body string
	dir  bool
}

func terminalTar(t *testing.T, entries ...terminalTarEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o600, Size: int64(len(e.body)), Typeflag: tar.TypeReg}
		if e.dir {
			hdr = &tar.Header{Name: e.name, Mode: 0o755, Typeflag: tar.TypeDir}
		}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write([]byte(e.body))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())

	return buf.Bytes()
}

// terminalShortArchive declares a size `missing` bytes past body and ends after body, like a copy cut off mid-stream.
func terminalShortArchive(t *testing.T, name, body string, missing int) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(body) + missing)}))
	_, err := tw.Write([]byte(body))
	require.NoError(t, err)

	return buf.Bytes()
}

// terminalLargeFileContent numbers every line and holds no NUL, so a lost or zeroed chunk shows.
func terminalLargeFileContent(lines int) string {
	var b strings.Builder
	for i := range lines {
		fmt.Fprintf(&b, "line %05d: the quick brown fox jumps over the lazy dog\n", i)
	}

	return b.String()
}

// terminalRecorded is the container id of every write of that state, nil when there is none.
func terminalRecorded(tlp *recordingTermLog, state termLogState) []int64 {
	var containers []int64
	for _, entry := range tlp.states(state) {
		containers = append(containers, entry.containerID)
	}
	return containers
}
