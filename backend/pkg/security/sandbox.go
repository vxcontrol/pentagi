package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	DefaultTimeout         = 5 * time.Minute
	MaxTimeout             = 20 * time.Minute
	DefaultMemoryLimit   = 512 * 1024 * 1024
	DefaultCPULimit      = 50
	MaxOutputSize        = 1024 * 1024
	SandboxWorkDir       = "/tmp/pentagi-sandbox"
)

type SandboxConfig struct {
	Timeout       time.Duration
	MemoryLimit   int64
	CPULimit      int
	NetworkIsolated bool
	ReadOnlyRoot   bool
	AllowedPaths   []string
	EnvWhitelist   []string
	MaxOutputSize  int64
}

type SandboxResult struct {
	Stdout     string            `json:"stdout"`
	Stderr     string            `json:"stderr"`
	ExitCode   int               `json:"exit_code"`
	Duration   float64           `json:"duration_seconds"`
	TimedOut   bool              `json:"timed_out"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	ExecutedIn string            `json:"executed_in"`
}

type SandboxExecutor struct {
	config  SandboxConfig
	logger  *logrus.Logger
	mu      sync.Mutex
	active  map[string]*exec.Cmd
	workDir string
}

func NewSandboxExecutor(cfg SandboxConfig) *SandboxExecutor {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.Timeout > MaxTimeout {
		cfg.Timeout = MaxTimeout
	}
	if cfg.MemoryLimit <= 0 {
		cfg.MemoryLimit = DefaultMemoryLimit
	}
	if cfg.CPULimit <= 0 {
		cfg.CPULimit = DefaultCPULimit
	}
	if cfg.MaxOutputSize <= 0 {
		cfg.MaxOutputSize = MaxOutputSize
	}

	workDir := SandboxWorkDir
	if wd := os.Getenv("PENTAGI_SANDBOX_DIR"); wd != "" {
		workDir = wd
	}

	return &SandboxExecutor{
		config:  cfg,
		logger:  logrus.StandardLogger(),
		active:  make(map[string]*exec.Cmd),
		workDir: workDir,
	}
}

func (s *SandboxExecutor) ExecuteSubprocess(ctx context.Context, command string, args []string, input string) (*SandboxResult, error) {
	return s.ExecuteSubprocessWithConfig(ctx, command, args, input, s.config)
}

func (s *SandboxExecutor) ExecuteSubprocessWithConfig(ctx context.Context, command string, args []string, input string, cfg SandboxConfig) (*SandboxResult, error) {
	startTime := time.Now()
	result := &SandboxResult{
		Metadata:   make(map[string]string),
		ExecutedIn: "subprocess_sandbox",
	}

	s.logger.WithFields(logrus.Fields{
		"command": command,
		"args":    args,
	}).Debug("Executing command in subprocess sandbox")

	execCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, command, args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS,
	}

	cmd.Dir = s.workDir
	if _, err := os.Stat(s.workDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.workDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create sandbox work directory: %w", err)
		}
	}

	env := []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=/tmp",
		"TMPDIR=/tmp",
	}
	for _, e := range cfg.EnvWhitelist {
		if val, ok := os.LookupEnv(e); ok {
			env = append(env, fmt.Sprintf("%s=%s", e, val))
		}
	}
	cmd.Env = env

	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	s.mu.Lock()
	s.active[execID] = cmd
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.active, execID)
		s.mu.Unlock()
	}()

	err := cmd.Run()
	result.Duration = time.Since(startTime).Seconds()

	if stdout.Len() > int(cfg.MaxOutputSize) {
		result.Stdout = stdout.String()[:cfg.MaxOutputSize] + "\n... [truncated]"
	} else {
		result.Stdout = stdout.String()
	}

	if stderr.Len() > int(cfg.MaxOutputSize) {
		result.Stderr = stderr.String()[:cfg.MaxOutputSize] + "\n... [truncated]"
	} else {
		result.Stderr = stderr.String()
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = -1
		result.Metadata["error"] = "command timed out"
		return result, nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Metadata["error"] = err.Error()
		}
		return result, nil
	}

	result.ExitCode = 0
	return result, nil
}

func (s *SandboxExecutor) ExecuteDocker(ctx context.Context, containerID string, command string, args []string, cfg SandboxConfig) (*SandboxResult, error) {
	startTime := time.Now()
	result := &SandboxResult{
		Metadata:   make(map[string]string),
		ExecutedIn: "docker_sandbox",
	}

	s.logger.WithFields(logrus.Fields{
		"container": containerID,
		"command":   command,
	}).Debug("Executing command in docker sandbox")

	dockerArgs := []string{
		"exec",
		"--interactive",
		"--tty=false",
	}

	if cfg.Timeout > 0 {
		timeoutSecs := int(cfg.Timeout.Seconds())
		dockerArgs = append(dockerArgs, "--timeout", fmt.Sprintf("%d", timeoutSecs))
	}

	if cfg.MemoryLimit > 0 {
		memoryMB := cfg.MemoryLimit / (1024 * 1024)
		dockerArgs = append(dockerArgs, "--memory", fmt.Sprintf("%dm", memoryMB))
	}

	if cfg.NetworkIsolated {
		dockerArgs = append(dockerArgs, "--network", "none")
	}

	if cfg.ReadOnlyRoot {
		dockerArgs = append(dockerArgs, "--read-only")
	}

	dockerArgs = append(dockerArgs, containerID, command)
	dockerArgs = append(dockerArgs, args...)

	execCtx, cancel := context.WithTimeout(ctx, cfg.Timeout+10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "docker", dockerArgs...)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(startTime).Seconds()

	if stdout.Len() > int(cfg.MaxOutputSize) {
		result.Stdout = stdout.String()[:cfg.MaxOutputSize] + "\n... [truncated]"
	} else {
		result.Stdout = stdout.String()
	}

	if stderr.Len() > int(cfg.MaxOutputSize) {
		result.Stderr = stderr.String()[:cfg.MaxOutputSize] + "\n... [truncated]"
	} else {
		result.Stderr = stderr.String()
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = -1
		result.Metadata["error"] = "docker command timed out"
		return result, nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Metadata["error"] = err.Error()
		}
		return result, nil
	}

	result.ExitCode = 0
	return result, nil
}

func (s *SandboxExecutor) KillAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, cmd := range s.active {
		s.logger.WithField("exec_id", id).Debug("Killing active process")
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
	s.active = make(map[string]*exec.Cmd)
}

func (s *SandboxExecutor) Cleanup() error {
	s.KillAll()

	entries, err := os.ReadDir(s.workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read sandbox directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(s.workDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			s.logger.WithError(err).WithField("path", path).Warn("Failed to cleanup sandbox file")
		}
	}

	return nil
}

func (s *SandboxResult) ToJSON() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal result: %s"}`, err)
	}
	return string(data)
}

func (s *SandboxResult) IsSuccess() bool {
	return s.ExitCode == 0 && !s.TimedOut
}

var ExploitiveToolPatterns = []string{
	"nmap", "metasploit", "sqlmap", "nikto", "hydra", "john",
	"aircrack", "burpsuite", "wireshark", "tcpdump", "netcat",
	"ncat", "nc", "nmap", "masscan", "zmap", "arpspoof",
	"ettercap", "responder", "crackmapexec", "evil-winrm",
	"impacket", "empire", "covenant", "sliver", "havoc",
}

func IsExploitiveTool(command string) bool {
	commandLower := strings.ToLower(command)
	for _, pattern := range ExploitiveToolPatterns {
		if strings.Contains(commandLower, pattern) {
			return true
		}
	}
	return false
}

func GetSandboxConfigForTool(tool string) SandboxConfig {
	cfg := SandboxConfig{
		Timeout:         DefaultTimeout,
		MemoryLimit:     DefaultMemoryLimit,
		CPULimit:        DefaultCPULimit,
		NetworkIsolated: false,
		ReadOnlyRoot:    false,
		MaxOutputSize:   MaxOutputSize,
		EnvWhitelist:    []string{},
		AllowedPaths:    []string{"/tmp", "/work"},
	}

	if IsExploitiveTool(tool) {
		cfg.Timeout = 10 * time.Minute
		cfg.MemoryLimit = 1024 * 1024 * 1024
		cfg.NetworkIsolated = false
	}

	return cfg
}
