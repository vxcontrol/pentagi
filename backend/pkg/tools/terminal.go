package tools

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

const (
	defaultExecCommandTimeout = 5 * time.Minute
	defaultExtraExecTimeout   = 5 * time.Second
	defaultQuickCheckTimeout  = 500 * time.Millisecond
)

type execResult struct {
	output string
	err    error
}

type terminal struct {
	flowID       int64
	taskID       *int64
	subtaskID    *int64
	containerID  int64
	containerLID string
	dockerClient docker.DockerClient
	tlp          TermLogProvider
}

func NewTerminalTool(flowID int64, taskID, subtaskID *int64,
	containerID int64, containerLID string,
	dockerClient docker.DockerClient, tlp TermLogProvider,
) Tool {
	return &terminal{
		flowID:       flowID,
		taskID:       taskID,
		subtaskID:    subtaskID,
		containerID:  containerID,
		containerLID: containerLID,
		dockerClient: dockerClient,
		tlp:          tlp,
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

		logrus.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"tool":   name,
			"result": result[:min(len(result), 1000)],
		}).Error("terminal tool failed")
		return fmt.Sprintf("terminal tool '%s' handled with error: %v", name, err), nil
	}
	return result, nil
}

func (t *terminal) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

	switch name {
	case TerminalToolName:
		var action TerminalAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal terminal action")
			return "", fmt.Errorf("failed to unmarshal terminal action: %w", err)
		}
		timeout := time.Duration(action.Timeout)*time.Second + defaultExtraExecTimeout
		result, err := t.ExecCommand(ctx, action.Cwd, action.Input, action.Detach.Bool(), timeout)
		return t.wrapCommandResult(ctx, args, name, result, err)
	case FileToolName:
		var action FileAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal file action")
			return "", fmt.Errorf("failed to unmarshal file action: %w", err)
		}

		logger = logger.WithFields(logrus.Fields{
			"action": action.Action,
			"path":   action.Path,
		})

		switch action.Action {
		case ReadFile:
			result, err := t.ReadFile(ctx, t.flowID, action.Path)
			return t.wrapCommandResult(ctx, args, name, result, err)
		case UpdateFile:
			result, err := t.WriteFile(ctx, t.flowID, action.Content, action.Path)
			return t.wrapCommandResult(ctx, args, name, result, err)
		default:
			logger.Error("unknown file action")
			return "", fmt.Errorf("unknown file action: %s", action.Action)
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
	containerName := PrimaryTerminalName(t.flowID)

	// create options for starting the exec process
	cmd := []string{
		"sh",
		"-c",
		command,
	}

	// check if container is running
	isRunning, err := t.dockerClient.IsContainerRunning(ctx, t.containerLID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}
	if !isRunning {
		return "", fmt.Errorf("container is not running")
	}

	if cwd == "" {
		cwd = docker.WorkFolderPathInContainer
	}

	formattedCommand := FormatTerminalInput(cwd, command)
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdin, formattedCommand, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (stdin): %w", err)
	}

	if timeout <= 0 || timeout > 20*time.Minute {
		timeout = defaultExecCommandTimeout
	}

	createResp, err := t.dockerClient.ContainerExecCreate(ctx, containerName, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   cwd,
		Tty:          true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create exec process: %w", err)
	}

	if detach {
		resultChan := make(chan execResult, 1)

		go func() {
			output, err := t.getExecResult(ctx, createResp.ID, timeout)
			resultChan <- execResult{output: output, err: err}
		}()

		select {
		case result := <-resultChan:
			if result.err != nil {
				return "", fmt.Errorf("command failed: %w: %s", result.err, result.output)
			}
			if result.output == "" {
				return "Command completed in background with exit code 0", nil
			}
			return result.output, nil
		case <-time.After(defaultQuickCheckTimeout):
			return fmt.Sprintf("Command started in background with timeout %s (still running)", timeout), nil
		}
	}

	return t.getExecResult(ctx, createResp.ID, timeout)
}

func (t *terminal) getExecResult(ctx context.Context, id string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// attach to the exec process
	resp, err := t.dockerClient.ContainerExecAttach(ctx, id, container.ExecAttachOptions{
		Tty: true,
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

		result := fmt.Sprintf("temporary output: %s", dst.String())
		return "", fmt.Errorf("timeout value is too low, use greater value if you need so: %w: %s", ctx.Err(), result)
	}

	// wait for the exec process to finish
	_, err = t.dockerClient.ContainerExecInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to inspect exec process: %w", err)
	}

	results := dst.String()
	formattedResults := FormatTerminalSystemOutput(results)
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdout, formattedResults, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (stdout): %w", err)
	}

	if results == "" {
		results = "Command completed successfully with exit code 0. No output produced (silent success)"
	}

	return results, nil
}

func (t *terminal) ReadFile(ctx context.Context, flowID int64, path string) (string, error) {
	containerName := PrimaryTerminalName(flowID)

	isRunning, err := t.dockerClient.IsContainerRunning(ctx, t.containerLID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}
	if !isRunning {
		return "", fmt.Errorf("container is not running")
	}

	cwd := docker.WorkFolderPathInContainer
	escapedPath := strings.ReplaceAll(path, "'", "'\"'\"'")
	formattedCommand := FormatTerminalInput(cwd, fmt.Sprintf("cat '%s'", escapedPath))
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdin, formattedCommand, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (read file cmd): %w", err)
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
			buffer.WriteString(
				fmt.Sprintf("'%s' file content (with size %d bytes) keeps bellow:\n",
					tarHeader.Name, tarHeader.Size,
				),
			)
		}

		const maxReadFileSize int64 = 100 * 1024 * 1024 // 100 MB limit
		if tarHeader.Size > maxReadFileSize {
			return "", fmt.Errorf("file '%s' size %d exceeds maximum allowed size %d", tarHeader.Name, tarHeader.Size, maxReadFileSize)
		}
		if tarHeader.Size < 0 {
			return "", fmt.Errorf("file '%s' has invalid size %d", tarHeader.Name, tarHeader.Size)
		}

		var fileContent = make([]byte, tarHeader.Size)
		_, err = tarReader.Read(fileContent)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file '%s' content: %w", tarHeader.Name, err)
		}
		buffer.Write(fileContent)

		if stats.Mode.IsDir() {
			buffer.WriteString("\n\n")
		}
	}

	content := buffer.String()
	formattedContent := FormatTerminalSystemOutput(content)
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdout, formattedContent, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (read file content): %w", err)
	}

	return content, nil
}

func (t *terminal) WriteFile(ctx context.Context, flowID int64, content string, path string) (string, error) {
	containerName := PrimaryTerminalName(flowID)

	isRunning, err := t.dockerClient.IsContainerRunning(ctx, t.containerLID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}
	if !isRunning {
		return "", fmt.Errorf("container is not running")
	}

	// put content into a tar archive
	archive := &bytes.Buffer{}
	tarWriter := tar.NewWriter(archive)
	defer tarWriter.Close()

	filename := filepath.Base(path)
	tarHeader := &tar.Header{
		Name: filename,
		Mode: 0600,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(tarHeader)
	if err != nil {
		return "", fmt.Errorf("failed to write tar header: %w", err)
	}

	_, err = tarWriter.Write([]byte(content))
	if err != nil {
		return "", fmt.Errorf("failed to write tar content: %w", err)
	}

	err = tarWriter.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close tar writer: %w", err)
	}

	dir := filepath.Dir(path)
	err = t.dockerClient.CopyToContainer(ctx, containerName, dir, archive, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	formattedCommand := FormatTerminalSystemOutput(fmt.Sprintf("Wrote to %s", path))
	_, err = t.tlp.PutMsg(ctx, database.TermlogTypeStdin, formattedCommand, t.containerID, t.taskID, t.subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to put terminal log (write file cmd): %w", err)
	}

	return fmt.Sprintf("file %s written successfully", path), nil
}

func PrimaryTerminalName(flowID int64) string {
	return fmt.Sprintf("pentagi-terminal-%d", flowID)
}

func FormatTerminalInput(cwd, text string) string {
	yellow := "\033[33m" // ANSI escape code for yellow color
	reset := "\033[0m"   // ANSI escape code to reset color
	return fmt.Sprintf("%s $ %s%s%s\r\n", cwd, yellow, text, reset)
}

func FormatTerminalSystemOutput(text string) string {
	blue := "\033[34m" // ANSI escape code for blue color
	reset := "\033[0m" // ANSI escape code to reset color
	return fmt.Sprintf("%s%s%s\r\n", blue, text, reset)
}

func (t *terminal) IsAvailable() bool {
	return t.dockerClient != nil
}
