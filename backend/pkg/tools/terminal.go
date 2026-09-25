package tools

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

// PrimaryTerminalNamePrefix is the prefix used for all primary terminal container names.
const PrimaryTerminalNamePrefix = "pentagi-terminal-"

var errContainerNotOperational = errors.New("container runtime is not operational")

const (
	maxExplicitExecCommandTimeout = 3 * time.Hour
	defaultExtraExecTimeout       = 5 * time.Second
	defaultQuickCheckTimeout      = 500 * time.Millisecond

	// ANSI terminal color codes (aligned with PentAGI UI palette)
	ansiColorInputCmd  = "\033[96m" // Bright Cyan - matches UI blue accents
	ansiColorSystemMsg = "\033[92m" // Bright Green - universal success/info
	ansiColorReset     = "\033[0m"  // Reset to default
	ansiLineTerminator = "\r\n"     // CRLF for terminal compatibility
)

type execResult struct {
	output string
	err    error
}

type terminal struct {
	flowID             int64
	taskID             *int64
	subtaskID          *int64
	containerID        int64
	containerLID       string
	tenantPrefix       string
	dockerClient       docker.DockerClient
	tlp                TermLogProvider
	defaultExecTimeout time.Duration
}

func NewTerminalTool(
	flowID int64,
	taskID, subtaskID *int64,
	containerID int64, containerLID string,
	tenantPrefix string,
	dockerClient docker.DockerClient,
	tlp TermLogProvider,
	defaultExecTimeout time.Duration,
) Tool {
	return &terminal{
		flowID:             flowID,
		taskID:             taskID,
		subtaskID:          subtaskID,
		containerID:        containerID,
		containerLID:       containerLID,
		tenantPrefix:       tenantPrefix,
		dockerClient:       dockerClient,
		tlp:                tlp,
		defaultExecTimeout: defaultExecTimeout,
	}
}

func (t *terminal) configuredExecTimeout() time.Duration {
	if t.defaultExecTimeout <= 0 || t.defaultExecTimeout > maxExplicitExecCommandTimeout {
		// Zero, negative, or above the operator ceiling: cap to the maximum allowed value.
		// Agents must never execute commands without a time bound.
		return maxExplicitExecCommandTimeout
	}

	return t.defaultExecTimeout
}

func (t *terminal) normalizeExecTimeout(timeout time.Duration) time.Duration {
	switch defaultExecTimeout := t.configuredExecTimeout() + defaultExtraExecTimeout; {
	case timeout > 0 && timeout <= defaultExecTimeout:
		return timeout
	default:
		return defaultExecTimeout
	}
}

func (t *terminal) wrapCommandResult(ctx context.Context, args json.RawMessage, name, result string, err error) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("terminal tool error swallowed"),
			langfuse.WithEventInput(args),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": name,
				"error":     err.Error(),
			}),
		)

		logrus.WithContext(ctx).WithError(err).WithFields(
			enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, logrus.Fields{
				"tool":   name,
				"result": result[:min(len(result), 1000)],
			})).Error("terminal tool failed")
		return fmt.Sprintf("terminal tool '%s' handled with error: %v", name, err), nil
	}
	return result, nil
}

func (t *terminal) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !t.IsAvailable() {
		return "", fmt.Errorf("terminal is not available")
	}

	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	switch name {
	case TerminalToolName:
		var action TerminalAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal terminal action")
			return "", fmt.Errorf("failed to unmarshal terminal action: %w", err)
		}
		timeout := t.normalizeExecTimeout(time.Duration(action.Timeout) * time.Second)
		if timeout > 0 {
			timeout += defaultExtraExecTimeout
		}
		result, err := t.ExecCommand(ctx, action.Cwd, action.Input, action.Detach.Bool(), timeout)
		return t.wrapCommandResult(ctx, args, name, result, err)
	case FileToolName:
		var action FileAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal file action")
			return "", fmt.Errorf("failed to unmarshal file action: %w", err)
		}

		if action.Action == "" {
			// The LLM occasionally omits the required 'action' field even though the
			// tool schema marks it required. The intent is almost always unambiguous
			// from the other fields present, so infer it instead of failing the call
			// outright and burning a tool-call-fixer round-trip on something that
			// doesn't need one.
			switch {
			case action.Diff != "":
				action.Action = EditFile
			case action.Content != "":
				action.Action = WriteFile
			default:
				action.Action = ReadFile
			}
		}

		logger = logger.WithFields(logrus.Fields{
			"action": action.Action,
			"path":   action.Path,
		})

		switch action.Action {
		case ReadFile:
			result, err := t.ReadFile(ctx, t.flowID, action.Path.String())
			return t.wrapCommandResult(ctx, args, name, result, err)
		case WriteFile:
			result, err := t.WriteFile(ctx, t.flowID, action.Content, action.Path.String())
			return t.wrapCommandResult(ctx, args, name, result, err)
		case EditFile:
			result, err := t.EditFile(ctx, t.flowID, action.Path.String(), action.Diff.String())
			return t.wrapCommandResult(ctx, args, name, result, err)
		default:
			logger.Error("unknown file action")
			return "", fmt.Errorf("unknown file action %q: expected one of %s, %s or %s",
				action.Action, ReadFile, WriteFile, EditFile)
		}
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (t *terminal) ExecCommand(
	ctx context.Context,
	cwd, command string,
	detach bool,
	timeout time.Duration,
) (string, error) {
	containerName := PrimaryTerminalName(t.tenantPrefix, t.flowID)

	cmd := docker.SandboxCommand(command)
	if t.taskID != nil {
		cmd = docker.FlowCommand(command)
	}

	err := t.requireRunningContainer(ctx)
	if err != nil {
		return "", err
	}

	if cwd == "" {
		cwd = docker.WorkFolderPathInContainer
	}

	// Format command with working directory and ANSI styling
	styledCommand := fmt.Sprintf("%s $ %s%s%s%s", cwd, ansiColorInputCmd, command, ansiColorReset, ansiLineTerminator)
	// Logged before anything runs: a failed write stops the call, so an error means nothing ran and a retry cannot run it twice.
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdin, styledCommand, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (stdin): %w", err)
	}

	timeout = t.normalizeExecTimeout(timeout)

	createResp, err := t.dockerClient.ContainerExecCreate(ctx, containerName, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   cwd,
		TTY:          true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create exec process: %w", err)
	}

	if detach {
		resultChan := make(chan execResult, 1)
		// The caller is often cancelled right after the start, a delegation timing out, and that must not kill the command.
		detachedCtx := context.WithoutCancel(ctx)

		go func() {
			output, err := t.getExecResult(detachedCtx, createResp.ID, timeout)
			resultChan <- execResult{output: output, err: err}
		}()

		select {
		case result := <-resultChan:
			if result.err != nil {
				return "", fmt.Errorf("command failed: %w: %s", result.err, result.output)
			}
			return result.output, nil
		case <-time.After(defaultQuickCheckTimeout):
			return fmt.Sprintf("Command started in background with timeout %s (still running)", timeout), nil
		}
	}

	return t.getExecResult(ctx, createResp.ID, timeout)
}

func (t *terminal) getExecResult(ctx context.Context, id string, timeout time.Duration) (string, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	resp, err := t.dockerClient.ContainerExecAttach(ctx, id, client.ExecAttachOptions{
		TTY: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec process: %w", err)
	}
	defer resp.Close()

	dst := bytes.Buffer{}
	errChan := make(chan error, 1)

	go func() {
		_, copyErr := io.Copy(&dst, resp.Reader)
		errChan <- copyErr
	}()

	select {
	case err := <-errChan:
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to copy output: %w", err)
		}
	case <-ctx.Done():
		// Close the response to unblock io.Copy
		resp.Close()

		// Wait for the copy goroutine to finish
		<-errChan

		partial := dst.String()

		// What the command managed to print is the only record of it. The
		// success path below writes its output to the terminal log; without the
		// same write here the UI shows the command and nothing under it. The
		// call's own context has already fired, so the write needs one that has
		// not, the way the detached path does.
		if partial != "" {
			styled := fmt.Sprintf("%s%s%s%s", ansiColorSystemMsg, partial, ansiColorReset, ansiLineTerminator)
			if _, err := t.tlp.PutMsg(context.WithoutCancel(ctx), database.TermlogTypeStdout,
				styled, t.containerID, t.taskID, t.subtaskID); err != nil {
				logrus.WithContext(ctx).WithError(err).Error("failed to put terminal log (timeout partial)")
			}
		}

		suggestedTimeout := max(int(timeout.Seconds())-10, 10)

		// The output stays inside the error text because that is the only
		// channel wrapCommandResult forwards on the error branch; its `result`
		// argument is used as a log field and then dropped. The cap is the
		// pipeline's own, not the 500 bytes this used to keep -- the executor
		// already summarizes a terminal result above DefaultResultSizeLimit.
		return "", fmt.Errorf(
			"command execution timeout (%v). Partial output: %s. "+
				"HINT: If this is an interactive command (shell/REPL/listener), use detach=true. "+
				"For long batch commands, wrap with shell timeout utility: 'timeout %d <command>' to ensure clean completion",
			ctx.Err(),
			truncateString(partial, DefaultResultSizeLimit),
			suggestedTimeout,
		)
	}

	inspect, err := t.dockerClient.ContainerExecInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to inspect exec process: %w", err)
	}

	results := dst.String()
	// Style system output with color coding
	styledOutput := fmt.Sprintf("%s%s%s%s", ansiColorSystemMsg, results, ansiColorReset, ansiLineTerminator)
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdout, styledOutput, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to put terminal log (stdout)")
	}

	if results == "" {
		results = "[no output]"
	}
	results = fmt.Sprintf("%s\n[exit code: %d]", results, inspect.ExitCode)

	if inspect.ExitCode > 128 && errors.Is(t.requireRunningContainer(ctx), errContainerNotOperational) {
		results += "\n[interrupted: the sandbox container stopped]"
	}

	return results, nil
}

func (t *terminal) requireRunningContainer(ctx context.Context) error {
	isRunning, err := t.dockerClient.IsContainerRunning(ctx, t.containerLID)
	if err != nil {
		return fmt.Errorf("runtime verification failed: %w", err)
	}
	if isRunning {
		if err := t.tlp.ContainerRunning(ctx, t.containerID, t.taskID, t.subtaskID); err != nil {
			t.logReconcileFailure(ctx, err, "failed to record the sandbox as running again")
		}
		return nil
	}

	if err := t.tlp.ContainerNotRunning(ctx, t.containerID, t.taskID, t.subtaskID); err != nil {
		t.logReconcileFailure(ctx, err, "failed to record the sandbox as not running")
	}

	return errContainerNotOperational
}

func (t *terminal) logReconcileFailure(ctx context.Context, err error, msg string) {
	logrus.WithContext(ctx).
		WithError(err).
		WithFields(enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, nil)).
		Warn(msg)
}

var shellExpansion = regexp.MustCompile("\\$\\(|\\$\\{|`|\\$[A-Za-z_][A-Za-z0-9_]*")

func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required and cannot be empty: give an absolute path such as /work/notes.txt")
	}

	if token := shellExpansion.FindString(path); token != "" {
		return fmt.Errorf(
			"path must be a literal filename, but it contains %q: this field is not run through a shell, "+
				"so the expression is taken as part of the name. Use the terminal tool to resolve it "+
				"first, then pass the result here",
			token)
	}

	return nil
}

func (t *terminal) ReadFile(ctx context.Context, flowID int64, path string) (string, error) {
	if err := validateFilePath(path); err != nil {
		return "", err
	}

	cwd := docker.WorkFolderPathInContainer
	escapedPath := strings.ReplaceAll(path, "'", "'\"'\"'")
	catCommand := fmt.Sprintf("cat '%s'", escapedPath)
	// Format read file command with styling
	styledCommand := fmt.Sprintf("%s $ %s%s%s%s", cwd, ansiColorInputCmd, catCommand, ansiColorReset, ansiLineTerminator)
	_, err := t.tlp.PutMsg(ctx, database.TermlogTypeStdin, styledCommand, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (read file cmd): %w", err)
	}

	content, err := t.readFileFromContainer(ctx, flowID, path)
	if err != nil {
		return "", err
	}

	// Style file content output
	styledContent := fmt.Sprintf("%s%s%s%s", ansiColorSystemMsg, content, ansiColorReset, ansiLineTerminator)
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdout, styledContent, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (read file content): %w", err)
	}

	return content, nil
}

// readFileFromContainer copies path out of the flow's container and returns
// its content. It performs no terminal-log writes, so callers that need the
// content only as an intermediate step (e.g. EditFile, before reapplying a
// diff and writing back) don't echo a spurious "cat" transcript entry.
func (t *terminal) readFileFromContainer(ctx context.Context, flowID int64, path string) (string, error) {
	containerName := PrimaryTerminalName(t.tenantPrefix, flowID)

	err := t.requireRunningContainer(ctx)
	if err != nil {
		return "", err
	}

	reader, stats, err := t.dockerClient.CopyFromContainer(ctx, containerName, path)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	defer reader.Close()

	var buffer strings.Builder
	tarReader := tar.NewReader(reader)
	for {
		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		if tarHeader.FileInfo().IsDir() {
			continue
		}

		if stats.Mode.IsDir() {
			buffer.WriteString("--------------------------------------------------\n")
			fmt.Fprintf(&buffer, "'%s' file content (with size %d bytes) shown below:\n",
				tarHeader.Name, tarHeader.Size)
		}

		const maxReadFileSize int64 = 100 * 1024 * 1024 // 100 MB limit
		if tarHeader.Size > maxReadFileSize {
			return "", fmt.Errorf("file '%s' size %d exceeds maximum allowed size %d", tarHeader.Name, tarHeader.Size, maxReadFileSize)
		}
		if tarHeader.Size < 0 {
			return "", fmt.Errorf("file '%s' has invalid size %d", tarHeader.Name, tarHeader.Size)
		}

		// Copied rather than read into a full-size buffer. tar.Reader.Read
		// returns whatever the underlying HTTP body has to hand -- typically far
		// less than the whole entry -- and the previous code discarded that count
		// and appended the entire buffer, so everything past the first chunk
		// arrived as NUL bytes. EditFile reads through here and writes the
		// result back, which committed that padding to the file.
		if _, err := io.CopyN(&buffer, tarReader, tarHeader.Size); err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file '%s' content: %w", tarHeader.Name, err)
		}

		if stats.Mode.IsDir() {
			buffer.WriteString("\n\n")
		}
	}

	return buffer.String(), nil
}

func (t *terminal) WriteFile(ctx context.Context, flowID int64, content string, path string) (string, error) {
	if err := validateFilePath(path); err != nil {
		return "", err
	}

	if err := t.writeFileToContainer(ctx, flowID, path, content); err != nil {
		return "", err
	}

	// Format success message with styling
	successMsg := fmt.Sprintf("File successfully saved to %s", path)
	styledMsg := fmt.Sprintf("%s%s%s%s", ansiColorSystemMsg, successMsg, ansiColorReset, ansiLineTerminator)
	_, err := t.tlp.PutMsg(ctx, database.TermlogTypeStdin, styledMsg, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to put terminal log (write file cmd)")
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// writeFileToContainer copies content into the flow's container at path,
// overwriting it. It performs no terminal-log writes; WriteFile and EditFile
// each log their own, differently-worded, success message.
func (t *terminal) writeFileToContainer(ctx context.Context, flowID int64, path, content string) error {
	containerName := PrimaryTerminalName(t.tenantPrefix, flowID)

	err := t.requireRunningContainer(ctx)
	if err != nil {
		return err
	}

	// Docker SDK requires TAR format for file transfer
	tarBuffer := &bytes.Buffer{}
	archiveWriter := tar.NewWriter(tarBuffer)
	defer archiveWriter.Close()

	filename := filepath.Base(path)
	fileDescriptor := &tar.Header{
		Name: filename,
		Mode: 0600,
		Size: int64(len(content)),
	}
	err = archiveWriter.WriteHeader(fileDescriptor)
	if err != nil {
		return fmt.Errorf("tar archive header generation failed: %w", err)
	}

	_, err = archiveWriter.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("tar archive content serialization failed: %w", err)
	}

	err = archiveWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	dir := filepath.Dir(path)
	err = t.dockerClient.CopyToContainer(ctx, containerName, dir, tarBuffer, client.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
	if err != nil {
		return fmt.Errorf("container file transfer failed: %w", err)
	}

	return nil
}

// EditFile applies a unified diff to the file at path: it reads the current
// content, applies the diff to it entirely in memory (see ApplyUnifiedDiff),
// and only if every hunk applied cleanly writes the result back - a diff
// that doesn't fully apply leaves the file untouched.
func (t *terminal) EditFile(ctx context.Context, flowID int64, path, diffText string) (string, error) {
	if err := validateFilePath(path); err != nil {
		return "", err
	}
	if strings.TrimSpace(diffText) == "" {
		return "", fmt.Errorf("diff is required and cannot be empty")
	}

	current, err := t.readFileFromContainer(ctx, flowID, path)
	if err != nil {
		return "", fmt.Errorf("failed to read current content of %s before editing: %w", path, err)
	}

	newContent, hunksApplied, err := ApplyUnifiedDiff(current, diffText)
	if err != nil {
		return "", fmt.Errorf("failed to apply diff to %s: %w", path, err)
	}

	if err := t.writeFileToContainer(ctx, flowID, path, newContent); err != nil {
		return "", fmt.Errorf("failed to write edited content of %s: %w", path, err)
	}

	successMsg := fmt.Sprintf("Applied %d diff hunk(s) to %s (%d -> %d bytes)", hunksApplied, path, len(current), len(newContent))
	styledMsg := fmt.Sprintf("%s%s%s%s", ansiColorSystemMsg, successMsg, ansiColorReset, ansiLineTerminator)
	if _, err := t.tlp.PutMsg(ctx, database.TermlogTypeStdin, styledMsg, t.containerID, t.taskID, t.subtaskID); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to put terminal log (edit file cmd)")
	}

	return successMsg, nil
}

// PrimaryTerminalName returns the docker container name for a flow's primary
// terminal, namespaced by the configured tenant.
//
//	"pentagi-terminal-1"       (single instance)
//	"acme-pentagi-terminal-1"  (TENANT_ID=acme)
//
// The tenant goes in FRONT of the well-known prefix on purpose: the installer's
// volume garbage collector force-removes anything matching
// HasPrefix("pentagi-terminal-") && HasSuffix("-data"), so a leading tenant
// segment keeps one tenant's objects outside another tenant's sweep. A tenant
// segment placed after the prefix would stay inside it and be destroyed.
func PrimaryTerminalName(tenantPrefix string, flowID int64) string {
	return fmt.Sprintf("%s%s%d", tenantPrefix, PrimaryTerminalNamePrefix, flowID)
}

func (t *terminal) IsAvailable() bool {
	return t.dockerClient != nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... [truncated full size is " + strconv.Itoa(len(s)) + " bytes]"
}
