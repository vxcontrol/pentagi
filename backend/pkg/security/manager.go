package security

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type SecurityManager struct {
	sandbox    *SandboxExecutor
	sanitizer  *ResponseSanitizer
	mu         sync.RWMutex
	auditLog   []AuditEntry
}

type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Actor       string    `json:"actor"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	Result      string    `json:"result"`
	Details     string    `json:"details,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
}

type SecurityContext struct {
	FlowID        int64
	TaskID        *int64
	SubtaskID     *int64
	TargetHost    string
	AuthorizedBy  string
	SessionID     string
	Roles         []string
	Permissions   []string
}

func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		sandbox:   NewSandboxExecutor(DefaultSandboxConfig()),
		sanitizer: NewResponseSanitizer(DefaultSanitizerConfig()),
		auditLog:  []AuditEntry{},
	}
}

func NewSecurityManagerWithConfig(sandboxCfg SandboxConfig, sanitizerCfg SanitizerConfig) *SecurityManager {
	return &SecurityManager{
		sandbox:   NewSandboxExecutor(sandboxCfg),
		sanitizer: NewResponseSanitizer(sanitizerCfg),
		auditLog:  []AuditEntry{},
	}
}

func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Timeout:         DefaultTimeout,
		MemoryLimit:     DefaultMemoryLimit,
		CPULimit:        DefaultCPULimit,
		NetworkIsolated: false,
		ReadOnlyRoot:    false,
		MaxOutputSize:   MaxOutputSize,
		EnvWhitelist:    []string{},
		AllowedPaths:    []string{"/tmp", "/work"},
	}
}

func (sm *SecurityManager) ExecuteSandboxed(ctx context.Context, command string, args []string) (*SandboxResult, error) {
	sm.audit("security_manager", "sandbox_execute", command, "started", "")

	cfg := GetSandboxConfigForTool(command)
	result, err := sm.sandbox.ExecuteSubprocessWithConfig(ctx, command, args, "", cfg)
	if err != nil {
		sm.audit("security_manager", "sandbox_execute", command, "error", err.Error())
		return nil, err
	}

	sm.audit("security_manager", "sandbox_execute", command, "completed", fmt.Sprintf("exit_code=%d", result.ExitCode))
	return result, nil
}

func (sm *SecurityManager) ExecuteDockerSandboxed(ctx context.Context, containerID, command string, args []string) (*SandboxResult, error) {
	sm.audit("security_manager", "docker_execute", fmt.Sprintf("%s:%s", containerID, command), "started", "")

	cfg := GetSandboxConfigForTool(command)
	result, err := sm.sandbox.ExecuteDocker(ctx, containerID, command, args, cfg)
	if err != nil {
		sm.audit("security_manager", "docker_execute", containerID, "error", err.Error())
		return nil, err
	}

	sm.audit("security_manager", "docker_execute", containerID, "completed", fmt.Sprintf("exit_code=%d", result.ExitCode))
	return result, nil
}

func (sm *SecurityManager) SanitizeResponse(response string, targetHost string) *SanitizationResult {
	return sm.sanitizer.SanitizeTargetResponse(response, targetHost)
}

func (sm *SecurityManager) SanitizeThinking(thinking string) *SanitizationResult {
	return sm.sanitizer.SanitizeThinking(thinking)
}

func (sm *SecurityManager) SanitizeCommand(command string) string {
	return SanitizeCommand(command)
}

func (sm *SecurityManager) ValidateTarget(host string) error {
	return ValidateTargetHost(host)
}

func (sm *SecurityManager) CreateReport(flowID int64) *ReportBuilder {
	sm.audit("security_manager", "create_report", fmt.Sprintf("flow:%d", flowID), "started", "")
	return NewReportBuilder(flowID)
}

func (sm *SecurityManager) audit(actor, action, resource, result, details string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	entry := AuditEntry{
		Timestamp: time.Now(),
		EventType: "security",
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
	}

	sm.auditLog = append(sm.auditLog, entry)
}

func (sm *SecurityManager) GetAuditLog() []AuditEntry {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]AuditEntry, len(sm.auditLog))
	copy(result, sm.auditLog)
	return result
}

func (sm *SecurityManager) GetAuditLogJSON() (string, error) {
	log := sm.GetAuditLog()
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal audit log: %w", err)
	}
	return string(data), nil
}

func (sm *SecurityManager) Cleanup() error {
	return sm.sandbox.Cleanup()
}

func (sm *SecurityManager) GetSandbox() *SandboxExecutor {
	return sm.sandbox
}

func (sm *SecurityManager) GetSanitizer() *ResponseSanitizer {
	return sm.sanitizer
}

type Middleware func(ctx context.Context, secCtx *SecurityContext, next func() error) error

func RequireAuthorization(secMgr *SecurityManager) Middleware {
	return func(ctx context.Context, secCtx *SecurityContext, next func() error) error {
		if secCtx.AuthorizedBy == "" {
			secMgr.audit("middleware", "authorization_check", "access", "denied", "no authorization")
			return fmt.Errorf("authorization required")
		}
		secMgr.audit("middleware", "authorization_check", "access", "granted", secCtx.AuthorizedBy)
		return next()
	}
}

func ValidateTarget(secMgr *SecurityManager) Middleware {
	return func(ctx context.Context, secCtx *SecurityContext, next func() error) error {
		if err := secMgr.ValidateTarget(secCtx.TargetHost); err != nil {
			secMgr.audit("middleware", "target_validation", secCtx.TargetHost, "rejected", err.Error())
			return err
		}
		secMgr.audit("middleware", "target_validation", secCtx.TargetHost, "approved", "")
		return next()
	}
}

func SanitizeOutput(secMgr *SecurityManager) Middleware {
	return func(ctx context.Context, secCtx *SecurityContext, next func() error) error {
		if err := next(); err != nil {
			return err
		}
		return nil
	}
}

type ToolExecutionWrapper struct {
	secMgr   *SecurityManager
	secCtx   *SecurityContext
}

func NewToolExecutionWrapper(secMgr *SecurityManager, secCtx *SecurityContext) *ToolExecutionWrapper {
	return &ToolExecutionWrapper{
		secMgr: secMgr,
		secCtx: secCtx,
	}
}

func (w *ToolExecutionWrapper) ExecuteCommand(ctx context.Context, command string, args []string) (*SandboxResult, *SanitizationResult, error) {
	sanitizedCommand := w.secMgr.SanitizeCommand(command)
	if sanitizedCommand != command {
		w.secMgr.audit("tool_wrapper", "command_sanitized", command, "modified", "")
	}

	result, err := w.secMgr.ExecuteSandboxed(ctx, sanitizedCommand, args)
	if err != nil {
		return nil, nil, err
	}

	sanitizedOutput := w.secMgr.SanitizeResponse(result.Stdout, w.secCtx.TargetHost)
	sanitizedStderr := w.secMgr.SanitizeResponse(result.Stderr, w.secCtx.TargetHost)

	if sanitizedOutput.WasRedacted || sanitizedStderr.WasRedacted {
		w.secMgr.audit("tool_wrapper", "output_sanitized", command, "redacted", "")
	}

	result.Stdout = sanitizedOutput.Sanitized
	result.Stderr = sanitizedStderr.Sanitized

	return result, sanitizedOutput, nil
}

func (w *ToolExecutionWrapper) ExecuteDockerCommand(ctx context.Context, containerID, command string, args []string) (*SandboxResult, *SanitizationResult, error) {
	sanitizedCommand := w.secMgr.SanitizeCommand(command)

	result, err := w.secMgr.ExecuteDockerSandboxed(ctx, containerID, sanitizedCommand, args)
	if err != nil {
		return nil, nil, err
	}

	sanitizedOutput := w.secMgr.SanitizeResponse(result.Stdout, w.secCtx.TargetHost)

	result.Stdout = sanitizedOutput.Sanitized

	return result, sanitizedOutput, nil
}
