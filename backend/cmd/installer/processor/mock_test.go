package processor

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/loader"
)

// mockState implements state.State interface for testing
type mockState struct {
	vars    map[string]loader.EnvVar
	envPath string
	stack   []string
	dirty   bool
}

func newMockState() *mockState {
	dir, err := os.MkdirTemp("", "pentagi-test")
	if err != nil {
		panic(err)
	}
	envPath := filepath.Join(dir, ".env")
	return &mockState{
		vars:    make(map[string]loader.EnvVar),
		envPath: envPath,
		stack:   []string{},
	}
}

func (m *mockState) Exists() bool                             { return true }
func (m *mockState) Reset() error                             { m.dirty = false; return nil }
func (m *mockState) Commit() error                            { m.dirty = false; return nil }
func (m *mockState) IsDirty() bool                            { return m.dirty }
func (m *mockState) GetEulaConsent() bool                     { return true }
func (m *mockState) SetEulaConsent() error                    { return nil }
func (m *mockState) SetStack(stack []string) error            { m.stack = stack; return nil }
func (m *mockState) GetStack() []string                       { return m.stack }
func (m *mockState) GetVar(name string) (loader.EnvVar, bool) { v, ok := m.vars[name]; return v, ok }
func (m *mockState) SetVar(name, value string) error {
	m.vars[name] = loader.EnvVar{Name: name, Value: value}
	m.dirty = true
	return nil
}
func (m *mockState) ResetVar(name string) error { delete(m.vars, name); return nil }
func (m *mockState) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
	result := make(map[string]loader.EnvVar)
	present := make(map[string]bool)
	for _, name := range names {
		v, ok := m.vars[name]
		result[name] = v
		present[name] = ok
	}
	return result, present
}
func (m *mockState) SetVars(vars map[string]string) error {
	for name, value := range vars {
		m.vars[name] = loader.EnvVar{Name: name, Value: value}
	}
	m.dirty = true
	return nil
}
func (m *mockState) ResetVars(names []string) error {
	for _, name := range names {
		delete(m.vars, name)
	}
	return nil
}
func (m *mockState) GetAllVars() map[string]loader.EnvVar { return m.vars }
func (m *mockState) GetEnvPath() string                   { return m.envPath }

// mockFiles implements files.Files interface for testing
type mockFiles struct {
	content  map[string][]byte
	statuses map[string]files.FileStatus
	lists    map[string][]string
	copies   []struct {
		Src, Dst string
		Rewrite  bool
	}
}

func newMockFiles() *mockFiles {
	return &mockFiles{
		content:  make(map[string][]byte),
		statuses: make(map[string]files.FileStatus),
		lists:    make(map[string][]string),
	}
}

func (m *mockFiles) GetContent(name string) ([]byte, error) {
	if content, ok := m.content[name]; ok {
		return content, nil
	}
	return nil, &os.PathError{Op: "read", Path: name, Err: os.ErrNotExist}
}

func (m *mockFiles) Exists(name string) bool {
	if _, ok := m.content[name]; ok {
		return true
	}
	// treat presence in lists as directory existence
	if _, ok := m.lists[name]; ok {
		return true
	}
	return false
}

func (m *mockFiles) ExistsInFS(name string) bool {
	return false
}

func (m *mockFiles) Stat(name string) (fs.FileInfo, error) {
	if _, exists := m.lists[name]; exists {
		// directory
		return &mockFileInfo{name: name, isDir: true}, nil
	}
	if _, exists := m.content[name]; exists {
		// file
		return &mockFileInfo{name: name, isDir: false}, nil
	}
	return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
}

func (m *mockFiles) Copy(src, dst string, rewrite bool) error {
	// record copy operation
	m.copies = append(m.copies, struct {
		Src, Dst string
		Rewrite  bool
	}{Src: src, Dst: dst, Rewrite: rewrite})
	return nil
}

func (m *mockFiles) Check(name string, workingDir string) files.FileStatus {
	status, exists := m.statuses[name]
	if !exists {
		return files.FileStatusOK
	}
	return status
}

func (m *mockFiles) List(prefix string) ([]string, error) {
	list, exists := m.lists[prefix]
	if !exists {
		return []string{}, nil
	}
	return list, nil
}

func (m *mockFiles) AddFile(name string, content []byte) {
	m.content[name] = content
}

// mockFileInfo implements fs.FileInfo for testing
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 100 } // arbitrary size
func (m *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// call represents a recorded method call with its parameters and result
type call struct {
	Method string
	Stack  ProductStack
	Name   string      // for network operations
	Args   interface{} // for additional arguments
	Error  error       // returned error
}

// baseMockFileSystemOperations provides base implementation with call logging
type baseMockFileSystemOperations struct {
	mu    sync.Mutex
	calls []call
	errOn map[string]error
}

func newBaseMockFileSystemOperations() *baseMockFileSystemOperations {
	return &baseMockFileSystemOperations{
		calls: make([]call, 0),
		errOn: make(map[string]error),
	}
}

func (m *baseMockFileSystemOperations) record(method string, stack ProductStack, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, call{Method: method, Stack: stack, Error: err})
	return err
}

func (m *baseMockFileSystemOperations) checkError(method string, stack ProductStack) error {
	if m.errOn != nil {
		// check for stack-specific error first
		methodKey := fmt.Sprintf("%s_%s", method, stack)
		if configuredErr, ok := m.errOn[methodKey]; ok {
			return configuredErr
		}
		// check for general method error
		if configuredErr, ok := m.errOn[method]; ok {
			return configuredErr
		}
	}
	return nil
}

func (m *baseMockFileSystemOperations) getCalls() []call {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]call, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *baseMockFileSystemOperations) setError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errOn == nil {
		m.errOn = make(map[string]error)
	}
	m.errOn[method] = err
}

func (m *baseMockFileSystemOperations) ensureStackIntegrity(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("ensureStackIntegrity", stack); err != nil {
		m.record("ensureStackIntegrity", stack, err)
		return err
	}
	return m.record("ensureStackIntegrity", stack, nil)
}

func (m *baseMockFileSystemOperations) verifyStackIntegrity(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("verifyStackIntegrity", stack); err != nil {
		m.record("verifyStackIntegrity", stack, err)
		return err
	}
	return m.record("verifyStackIntegrity", stack, nil)
}

func (m *baseMockFileSystemOperations) checkStackIntegrity(ctx context.Context, stack ProductStack) (FilesCheckResult, error) {
	if err := m.checkError("checkStackIntegrity", stack); err != nil {
		m.record("checkStackIntegrity", stack, err)
		return nil, err
	}
	// return empty map by default for tests; specific tests can stub via errOn if necessary
	_ = m.record("previewStackFilesStatus", stack, nil)
	return make(map[string]files.FileStatus), nil
}

func (m *baseMockFileSystemOperations) cleanupStackFiles(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("cleanupStackFiles", stack); err != nil {
		m.record("cleanupStackFiles", stack, err)
		return err
	}
	return m.record("cleanupStackFiles", stack, nil)
}

// baseMockDockerOperations provides base implementation with call logging
type baseMockDockerOperations struct {
	mu    sync.Mutex
	calls []call
	errOn map[string]error
}

func newBaseMockDockerOperations() *baseMockDockerOperations {
	return &baseMockDockerOperations{
		calls: make([]call, 0),
		errOn: make(map[string]error),
	}
}

func (m *baseMockDockerOperations) record(method string, name string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, call{Method: method, Name: name, Error: err})
	return err
}

func (m *baseMockDockerOperations) checkError(method string) error {
	if m.errOn != nil {
		if configuredErr, ok := m.errOn[method]; ok {
			return configuredErr
		}
	}
	return nil
}

func (m *baseMockDockerOperations) getCalls() []call {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]call, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *baseMockDockerOperations) setError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errOn == nil {
		m.errOn = make(map[string]error)
	}
	m.errOn[method] = err
}

func (m *baseMockDockerOperations) pullWorkerImage(ctx context.Context, state *operationState) error {
	if err := m.checkError("pullWorkerImage"); err != nil {
		m.record("pullWorkerImage", "", err)
		return err
	}
	return m.record("pullWorkerImage", "", nil)
}

func (m *baseMockDockerOperations) pullDefaultImage(ctx context.Context, state *operationState) error {
	if err := m.checkError("pullDefaultImage"); err != nil {
		m.record("pullDefaultImage", "", err)
		return err
	}
	return m.record("pullDefaultImage", "", nil)
}

func (m *baseMockDockerOperations) removeWorkerContainers(ctx context.Context, state *operationState) error {
	if err := m.checkError("removeWorkerContainers"); err != nil {
		m.record("removeWorkerContainers", "", err)
		return err
	}
	return m.record("removeWorkerContainers", "", nil)
}

func (m *baseMockDockerOperations) removeWorkerImages(ctx context.Context, state *operationState) error {
	if err := m.checkError("removeWorkerImages"); err != nil {
		m.record("removeWorkerImages", "", err)
		return err
	}
	return m.record("removeWorkerImages", "", nil)
}

func (m *baseMockDockerOperations) purgeWorkerImages(ctx context.Context, state *operationState) error {
	if err := m.checkError("purgeWorkerImages"); err != nil {
		m.record("purgeWorkerImages", "", err)
		return err
	}
	return m.record("purgeWorkerImages", "", nil)
}

func (m *baseMockDockerOperations) ensureMainDockerNetworks(ctx context.Context, state *operationState) error {
	if err := m.checkError("ensureMainDockerNetworks"); err != nil {
		m.record("ensureMainDockerNetworks", "", err)
		return err
	}
	return m.record("ensureMainDockerNetworks", "", nil)
}

func (m *baseMockDockerOperations) removeMainDockerNetwork(ctx context.Context, state *operationState, name string) error {
	if err := m.checkError("removeMainDockerNetwork"); err != nil {
		m.record("removeMainDockerNetwork", name, err)
		return err
	}
	return m.record("removeMainDockerNetwork", name, nil)
}

func (m *baseMockDockerOperations) removeMainImages(ctx context.Context, state *operationState, images []string) error {
	if err := m.checkError("removeMainImages"); err != nil {
		m.record("removeMainImages", "", err)
		return err
	}
	return m.record("removeMainImages", "", nil)
}

func (m *baseMockDockerOperations) removeWorkerVolumes(ctx context.Context, state *operationState) error {
	if err := m.checkError("removeWorkerVolumes"); err != nil {
		m.record("removeWorkerVolumes", "", err)
		return err
	}
	return m.record("removeWorkerVolumes", "", nil)
}

// baseMockComposeOperations provides base implementation with call logging
type baseMockComposeOperations struct {
	mu    sync.Mutex
	calls []call
	errOn map[string]error
}

func newBaseMockComposeOperations() *baseMockComposeOperations {
	return &baseMockComposeOperations{
		calls: make([]call, 0),
		errOn: make(map[string]error),
	}
}

func (m *baseMockComposeOperations) record(method string, stack ProductStack, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, call{Method: method, Stack: stack, Error: err})
	return err
}

func (m *baseMockComposeOperations) checkError(method string) error {
	if m.errOn != nil {
		if configuredErr, ok := m.errOn[method]; ok {
			// one-shot error to avoid leaking into subsequent subtests
			delete(m.errOn, method)
			return configuredErr
		}
	}
	return nil
}

func (m *baseMockComposeOperations) getCalls() []call {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]call, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *baseMockComposeOperations) setError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errOn == nil {
		m.errOn = make(map[string]error)
	}
	if err == nil {
		delete(m.errOn, method)
	} else {
		m.errOn[method] = err
	}
}

func (m *baseMockComposeOperations) startStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("startStack"); err != nil {
		m.record("startStack", stack, err)
		return err
	}
	return m.record("startStack", stack, nil)
}

func (m *baseMockComposeOperations) stopStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("stopStack"); err != nil {
		m.record("stopStack", stack, err)
		return err
	}
	return m.record("stopStack", stack, nil)
}

func (m *baseMockComposeOperations) restartStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("restartStack"); err != nil {
		m.record("restartStack", stack, err)
		return err
	}
	return m.record("restartStack", stack, nil)
}

func (m *baseMockComposeOperations) updateStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("updateStack"); err != nil {
		m.record("updateStack", stack, err)
		return err
	}
	return m.record("updateStack", stack, nil)
}

func (m *baseMockComposeOperations) downloadStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("downloadStack"); err != nil {
		m.record("downloadStack", stack, err)
		return err
	}
	return m.record("downloadStack", stack, nil)
}

func (m *baseMockComposeOperations) removeStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("removeStack"); err != nil {
		m.record("removeStack", stack, err)
		return err
	}
	return m.record("removeStack", stack, nil)
}

func (m *baseMockComposeOperations) purgeStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("purgeStack"); err != nil {
		m.record("purgeStack", stack, err)
		return err
	}
	return m.record("purgeStack", stack, nil)
}

func (m *baseMockComposeOperations) purgeImagesStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := m.checkError("purgeImagesStack"); err != nil {
		m.record("purgeImagesStack", stack, err)
		return err
	}
	return m.record("purgeImagesStack", stack, nil)
}

func (m *baseMockComposeOperations) determineComposeFile(stack ProductStack) (string, error) {
	if err := m.checkError("determineComposeFile"); err != nil {
		m.record("determineComposeFile", stack, err)
		return "", err
	}
	m.record("determineComposeFile", stack, nil)
	return "test-compose.yml", nil
}

func (m *baseMockComposeOperations) performStackCommand(ctx context.Context, stack ProductStack, state *operationState, args ...string) error {
	if err := m.checkError("performStackCommand"); err != nil {
		m.record("performStackCommand", stack, err)
		return err
	}
	return m.record("performStackCommand", stack, nil)
}

// baseMockUpdateOperations provides base implementation with call logging
type baseMockUpdateOperations struct {
	mu    sync.Mutex
	calls []call
	errOn map[string]error
}

func newBaseMockUpdateOperations() *baseMockUpdateOperations {
	return &baseMockUpdateOperations{
		calls: make([]call, 0),
		errOn: make(map[string]error),
	}
}

func (m *baseMockUpdateOperations) record(method string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, call{Method: method, Error: err})
	return err
}

func (m *baseMockUpdateOperations) checkError(method string) error {
	if m.errOn != nil {
		if configuredErr, ok := m.errOn[method]; ok {
			return configuredErr
		}
	}
	return nil
}

func (m *baseMockUpdateOperations) getCalls() []call {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]call, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *baseMockUpdateOperations) setError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errOn == nil {
		m.errOn = make(map[string]error)
	}
	m.errOn[method] = err
}

func (m *baseMockUpdateOperations) checkUpdates(ctx context.Context, state *operationState) (*checker.CheckUpdatesResponse, error) {
	if err := m.checkError("checkUpdates"); err != nil {
		m.record("checkUpdates", err)
		return nil, err
	}
	m.record("checkUpdates", nil)
	return &checker.CheckUpdatesResponse{}, nil
}

func (m *baseMockUpdateOperations) downloadInstaller(ctx context.Context, state *operationState) error {
	if err := m.checkError("downloadInstaller"); err != nil {
		m.record("downloadInstaller", err)
		return err
	}
	return m.record("downloadInstaller", nil)
}

func (m *baseMockUpdateOperations) updateInstaller(ctx context.Context, state *operationState) error {
	if err := m.checkError("updateInstaller"); err != nil {
		m.record("updateInstaller", err)
		return err
	}
	return m.record("updateInstaller", nil)
}

func (m *baseMockUpdateOperations) removeInstaller(ctx context.Context, state *operationState) error {
	if err := m.checkError("removeInstaller"); err != nil {
		m.record("removeInstaller", err)
		return err
	}
	return m.record("removeInstaller", nil)
}

// testState creates a test state with initialized environment
func testState(t *testing.T) *mockState {
	t.Helper()
	mockState := newMockState()
	envPath := mockState.GetEnvPath()
	_ = os.MkdirAll(filepath.Dir(envPath), 0o755)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if err := os.WriteFile(envPath, []byte("PENTAGI_VERSION=1.0.0\n"), 0o644); err != nil {
			t.Fatalf("failed to create env file: %v", err)
		}
	}
	return mockState
}

// defaultCheckResult returns a CheckResult with sensible defaults for testing
func defaultCheckResult() *checker.CheckResult {
	// use mock handler to create CheckResult with defaults
	handler := newMockCheckHandler()

	// Configure the default state we want for tests
	handler.config.PentagiExtracted = false
	handler.config.PentagiInstalled = false
	handler.config.PentagiRunning = false
	handler.config.LangfuseConnected = true
	handler.config.LangfuseExternal = false
	handler.config.LangfuseExtracted = false
	handler.config.LangfuseInstalled = false
	handler.config.LangfuseRunning = false
	handler.config.ObservabilityConnected = true
	handler.config.ObservabilityExternal = false
	handler.config.ObservabilityExtracted = false
	handler.config.ObservabilityInstalled = false
	handler.config.ObservabilityRunning = false
	handler.config.WorkerImageExists = false
	handler.config.PentagiIsUpToDate = true
	handler.config.LangfuseIsUpToDate = true
	handler.config.ObservabilityIsUpToDate = true
	handler.config.InstallerIsUpToDate = true

	// Create CheckResult with noop handler that has default values already set
	result, _ := checker.GatherWithHandler(context.Background(), &defaultStateHandler{
		mockHandler: handler,
	})
	return &result
}

// createCheckResultWithHandler creates a CheckResult that uses the provided mock handler
func createCheckResultWithHandler(handler *mockCheckHandler) *checker.CheckResult {
	// Use the public GatherWithHandler function
	result, _ := checker.GatherWithHandler(context.Background(), handler)
	return &result
}

// defaultStateHandler wraps a mock handler to provide initial state and then act as noop
type defaultStateHandler struct {
	mockHandler *mockCheckHandler
	initialized bool
}

func (h *defaultStateHandler) GatherAllInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		// First time - populate with configured values
		h.initialized = true
		return h.mockHandler.GatherAllInfo(ctx, c)
	}
	// Subsequent calls - act as noop
	return nil
}

func (h *defaultStateHandler) GatherDockerInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherDockerInfo(ctx, c)
	}
	return nil
}

func (h *defaultStateHandler) GatherWorkerInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherWorkerInfo(ctx, c)
	}
	return nil
}

func (h *defaultStateHandler) GatherPentagiInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherPentagiInfo(ctx, c)
	}
	return nil
}

func (h *defaultStateHandler) GatherLangfuseInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherLangfuseInfo(ctx, c)
	}
	return nil
}

func (h *defaultStateHandler) GatherObservabilityInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherObservabilityInfo(ctx, c)
	}
	return nil
}

func (h *defaultStateHandler) GatherSystemInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherSystemInfo(ctx, c)
	}
	return nil
}

func (h *defaultStateHandler) GatherUpdatesInfo(ctx context.Context, c *checker.CheckResult) error {
	if !h.initialized {
		return h.mockHandler.GatherUpdatesInfo(ctx, c)
	}
	return nil
}

// createTestProcessor creates a processor with mocked dependencies using base mock implementations
func createTestProcessor() *processor {
	mockState := newMockState()
	return createProcessorWithState(mockState, defaultCheckResult())
}

// createProcessorWithState creates a processor with specified state and checker
func createProcessorWithState(state *mockState, checkResult *checker.CheckResult) *processor {
	p := &processor{
		state:   state,
		checker: checkResult,
		files:   newMockFiles(),
	}

	// setup base mock operations
	p.fsOps = newBaseMockFileSystemOperations()
	p.dockerOps = newBaseMockDockerOperations()
	p.composeOps = newBaseMockComposeOperations()
	p.updateOps = newBaseMockUpdateOperations()

	return p
}

// common test data
var (
	// standard stacks for testing stack operations
	standardStacks = []ProductStack{
		ProductStackPentagi,
		ProductStackLangfuse,
		ProductStackObservability,
		ProductStackCompose,
		ProductStackAll,
	}

	// unsupported stacks for error testing
	unsupportedStacks = map[ProductStack][]ProcessorOperation{
		ProductStackWorker:    {ProcessorOperationStart, ProcessorOperationStop, ProcessorOperationRestart},
		ProductStackInstaller: {ProcessorOperationRestart},
	}

	// special error cases
	specialErrorCases = map[ProductStack]map[ProcessorOperation]string{
		// Currently no special error cases beyond unsupported operations
	}
)

// stackTestCase represents a test case for stack operations
type stackTestCase struct {
	name      string
	stack     ProductStack
	expectErr bool
	errorMsg  string
}

// generateStackTestCases creates standard test cases for stack operations
func generateStackTestCases(operation ProcessorOperation) []stackTestCase {
	var cases []stackTestCase

	// add successful cases for standard stacks
	for _, stack := range standardStacks {
		cases = append(cases, stackTestCase{
			name:      fmt.Sprintf("%s success", stack),
			stack:     stack,
			expectErr: false,
		})
	}

	// add error cases for unsupported stacks
	for stack, operations := range unsupportedStacks {
		for _, op := range operations {
			if op == operation {
				cases = append(cases, stackTestCase{
					name:      fmt.Sprintf("%s unsupported", stack),
					stack:     stack,
					expectErr: true,
				})
			}
		}
	}

	// add special error cases
	if stackErrors, exists := specialErrorCases[ProductStackInstaller]; exists {
		if expectedMsg, hasError := stackErrors[operation]; hasError {
			cases = append(cases, stackTestCase{
				name:      "installer special error",
				stack:     ProductStackInstaller,
				expectErr: true,
				errorMsg:  expectedMsg,
			})
		}
	}

	return cases
}

// Test helpers for common test patterns

// testOperationState creates a standard operation state for tests
func testOperationState(t *testing.T) *operationState {
	t.Helper()
	return &operationState{mx: &sync.Mutex{}, ctx: t.Context()}
}

// assertNoError is a test helper for error assertions
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// assertError is a test helper for error assertions
func assertError(t *testing.T, err error, expectErr bool, expectedMsg string) {
	t.Helper()
	if expectErr {
		if err == nil {
			t.Error("expected error but got none")
		} else if expectedMsg != "" && err.Error() != expectedMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	} else {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// mockCheckHandler implements checker.CheckHandler for testing
type mockCheckHandler struct {
	mu          sync.Mutex
	calls       []string
	config      mockCheckConfig
	gatherError error // if set, all Gather* methods return this error
}

// mockCheckConfig allows configuring what the mock handler returns
type mockCheckConfig struct {
	// File and system states
	EnvFileExists          bool
	DockerApiAccessible    bool
	WorkerEnvApiAccessible bool
	WorkerImageExists      bool
	DockerInstalled        bool
	DockerComposeInstalled bool
	DockerVersionOK        bool
	DockerComposeVersionOK bool
	DockerVersion          string
	DockerComposeVersion   string

	// PentAGI states
	PentagiScriptInstalled bool
	PentagiExtracted       bool
	PentagiInstalled       bool
	PentagiRunning         bool

	// Langfuse states
	LangfuseConnected bool
	LangfuseExternal  bool
	LangfuseExtracted bool
	LangfuseInstalled bool
	LangfuseRunning   bool

	// Observability states
	ObservabilityConnected bool
	ObservabilityExternal  bool
	ObservabilityExtracted bool
	ObservabilityInstalled bool
	ObservabilityRunning   bool

	// System checks
	SysNetworkOK       bool
	SysCPUOK           bool
	SysMemoryOK        bool
	SysDiskFreeSpaceOK bool

	// Update states
	UpdateServerAccessible  bool
	InstallerIsUpToDate     bool
	PentagiIsUpToDate       bool
	LangfuseIsUpToDate      bool
	ObservabilityIsUpToDate bool
}

func newMockCheckHandler() *mockCheckHandler {
	return &mockCheckHandler{
		calls: make([]string, 0),
		config: mockCheckConfig{
			// sensible defaults for most tests
			EnvFileExists:           true,
			DockerApiAccessible:     true,
			WorkerEnvApiAccessible:  true,
			DockerInstalled:         true,
			DockerComposeInstalled:  true,
			DockerVersionOK:         true,
			DockerComposeVersionOK:  true,
			DockerVersion:           "24.0.0",
			DockerComposeVersion:    "2.20.0",
			SysNetworkOK:            true,
			SysCPUOK:                true,
			SysMemoryOK:             true,
			SysDiskFreeSpaceOK:      true,
			UpdateServerAccessible:  true,
			InstallerIsUpToDate:     true,
			PentagiIsUpToDate:       true,
			LangfuseIsUpToDate:      true,
			ObservabilityIsUpToDate: true,
		},
	}
}

func (m *mockCheckHandler) setConfig(config mockCheckConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

func (m *mockCheckHandler) setGatherError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gatherError = err
}

func (m *mockCheckHandler) getCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *mockCheckHandler) recordCall(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, method)
}

func (m *mockCheckHandler) GatherAllInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherAllInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	// Call all gather methods to populate the result
	if err := m.GatherDockerInfo(ctx, c); err != nil {
		return err
	}
	if err := m.GatherWorkerInfo(ctx, c); err != nil {
		return err
	}
	if err := m.GatherPentagiInfo(ctx, c); err != nil {
		return err
	}
	if err := m.GatherLangfuseInfo(ctx, c); err != nil {
		return err
	}
	if err := m.GatherObservabilityInfo(ctx, c); err != nil {
		return err
	}
	if err := m.GatherSystemInfo(ctx, c); err != nil {
		return err
	}
	if err := m.GatherUpdatesInfo(ctx, c); err != nil {
		return err
	}

	c.EnvFileExists = m.config.EnvFileExists
	return nil
}

func (m *mockCheckHandler) GatherDockerInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherDockerInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.DockerApiAccessible = m.config.DockerApiAccessible
	c.DockerInstalled = m.config.DockerInstalled
	c.DockerComposeInstalled = m.config.DockerComposeInstalled
	c.DockerVersion = m.config.DockerVersion
	c.DockerVersionOK = m.config.DockerVersionOK
	c.DockerComposeVersion = m.config.DockerComposeVersion
	c.DockerComposeVersionOK = m.config.DockerComposeVersionOK

	return nil
}

func (m *mockCheckHandler) GatherWorkerInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherWorkerInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.WorkerEnvApiAccessible = m.config.WorkerEnvApiAccessible
	c.WorkerImageExists = m.config.WorkerImageExists

	return nil
}

func (m *mockCheckHandler) GatherPentagiInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherPentagiInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.PentagiScriptInstalled = m.config.PentagiScriptInstalled
	c.PentagiExtracted = m.config.PentagiExtracted
	c.PentagiInstalled = m.config.PentagiInstalled
	c.PentagiRunning = m.config.PentagiRunning

	return nil
}

func (m *mockCheckHandler) GatherLangfuseInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherLangfuseInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.LangfuseConnected = m.config.LangfuseConnected
	c.LangfuseExternal = m.config.LangfuseExternal
	c.LangfuseExtracted = m.config.LangfuseExtracted
	c.LangfuseInstalled = m.config.LangfuseInstalled
	c.LangfuseRunning = m.config.LangfuseRunning

	return nil
}

func (m *mockCheckHandler) GatherObservabilityInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherObservabilityInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.ObservabilityConnected = m.config.ObservabilityConnected
	c.ObservabilityExternal = m.config.ObservabilityExternal
	c.ObservabilityExtracted = m.config.ObservabilityExtracted
	c.ObservabilityInstalled = m.config.ObservabilityInstalled
	c.ObservabilityRunning = m.config.ObservabilityRunning

	return nil
}

func (m *mockCheckHandler) GatherSystemInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherSystemInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.SysNetworkOK = m.config.SysNetworkOK
	c.SysCPUOK = m.config.SysCPUOK
	c.SysMemoryOK = m.config.SysMemoryOK
	c.SysDiskFreeSpaceOK = m.config.SysDiskFreeSpaceOK

	return nil
}

func (m *mockCheckHandler) GatherUpdatesInfo(ctx context.Context, c *checker.CheckResult) error {
	m.recordCall("GatherUpdatesInfo")
	if m.gatherError != nil {
		return m.gatherError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	c.UpdateServerAccessible = m.config.UpdateServerAccessible
	c.InstallerIsUpToDate = m.config.InstallerIsUpToDate
	c.PentagiIsUpToDate = m.config.PentagiIsUpToDate
	c.LangfuseIsUpToDate = m.config.LangfuseIsUpToDate
	c.ObservabilityIsUpToDate = m.config.ObservabilityIsUpToDate

	return nil
}

// Test functions to verify mock CheckHandler works correctly

func TestMockCheckHandler_BasicFunctionality(t *testing.T) {
	handler := newMockCheckHandler()
	result := &checker.CheckResult{}

	// Test that all methods can be called and record their calls
	ctx := context.Background()

	err := handler.GatherDockerInfo(ctx, result)
	assertNoError(t, err)

	err = handler.GatherWorkerInfo(ctx, result)
	assertNoError(t, err)

	err = handler.GatherPentagiInfo(ctx, result)
	assertNoError(t, err)

	// Verify calls were recorded
	calls := handler.getCalls()
	expectedCalls := []string{"GatherDockerInfo", "GatherWorkerInfo", "GatherPentagiInfo"}

	if len(calls) != len(expectedCalls) {
		t.Fatalf("expected %d calls, got %d", len(expectedCalls), len(calls))
	}

	for i, expected := range expectedCalls {
		if calls[i] != expected {
			t.Errorf("call %d: expected %s, got %s", i, expected, calls[i])
		}
	}

	// Verify default values were set
	if !result.DockerApiAccessible {
		t.Error("expected DockerApiAccessible to be true by default")
	}
	if !result.WorkerEnvApiAccessible {
		t.Error("expected WorkerEnvApiAccessible to be true by default")
	}
}

func TestMockCheckHandler_CustomConfiguration(t *testing.T) {
	handler := newMockCheckHandler()
	result := &checker.CheckResult{}

	// Set custom configuration
	customConfig := mockCheckConfig{
		PentagiExtracted:       false,
		PentagiInstalled:       true,
		PentagiRunning:         false,
		LangfuseConnected:      true,
		LangfuseExternal:       true,
		ObservabilityConnected: false,
	}
	handler.setConfig(customConfig)

	// Gather info
	ctx := context.Background()
	_ = handler.GatherPentagiInfo(ctx, result)
	_ = handler.GatherLangfuseInfo(ctx, result)
	_ = handler.GatherObservabilityInfo(ctx, result)

	// Verify custom values were applied
	if result.PentagiExtracted != false {
		t.Error("expected PentagiExtracted to be false")
	}
	if result.PentagiInstalled != true {
		t.Error("expected PentagiInstalled to be true")
	}
	if result.LangfuseExternal != true {
		t.Error("expected LangfuseExternal to be true")
	}
	if result.ObservabilityConnected != false {
		t.Error("expected ObservabilityConnected to be false")
	}
}

func TestMockCheckHandler_ErrorInjection(t *testing.T) {
	handler := newMockCheckHandler()
	result := &checker.CheckResult{}

	// Set error to be returned
	expectedErr := fmt.Errorf("mock gather error")
	handler.setGatherError(expectedErr)

	// All gather methods should return the error
	ctx := context.Background()

	err := handler.GatherDockerInfo(ctx, result)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	err = handler.GatherAllInfo(ctx, result)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	// Verify calls were still recorded
	calls := handler.getCalls()
	if len(calls) != 2 {
		t.Errorf("expected 2 calls recorded even with errors, got %d", len(calls))
	}
}

func TestMockCheckHandler_GatherAllInfo(t *testing.T) {
	handler := newMockCheckHandler()
	result := &checker.CheckResult{}

	// Set specific configuration
	handler.config.PentagiExtracted = false
	handler.config.LangfuseConnected = true
	handler.config.ObservabilityExternal = true

	// Call GatherAllInfo
	ctx := context.Background()
	err := handler.GatherAllInfo(ctx, result)
	assertNoError(t, err)

	// Verify all gather methods were called
	calls := handler.getCalls()
	expectedCalls := []string{
		"GatherAllInfo",
		"GatherDockerInfo",
		"GatherWorkerInfo",
		"GatherPentagiInfo",
		"GatherLangfuseInfo",
		"GatherObservabilityInfo",
		"GatherSystemInfo",
		"GatherUpdatesInfo",
	}

	if len(calls) != len(expectedCalls) {
		t.Fatalf("expected %d calls, got %d: %v", len(expectedCalls), len(calls), calls)
	}

	for i, expected := range expectedCalls {
		if calls[i] != expected {
			t.Errorf("call %d: expected %s, got %s", i, expected, calls[i])
		}
	}

	// Verify configuration was applied
	if result.PentagiExtracted != false {
		t.Error("expected PentagiExtracted to be false")
	}
	if result.LangfuseConnected != true {
		t.Error("expected LangfuseConnected to be true")
	}
	if result.ObservabilityExternal != true {
		t.Error("expected ObservabilityExternal to be true")
	}
}

// Test for complex scenarios with multiple mock operations

func TestMockOperations_CallAccumulation(t *testing.T) {
	t.Run("FileSystemOperations", func(t *testing.T) {
		mock := newBaseMockFileSystemOperations()
		state := testOperationState(t)

		// make multiple calls
		_ = mock.ensureStackIntegrity(t.Context(), ProductStackPentagi, state)
		_ = mock.verifyStackIntegrity(t.Context(), ProductStackLangfuse, state)
		_ = mock.cleanupStackFiles(t.Context(), ProductStackObservability, state)
		_ = mock.ensureStackIntegrity(t.Context(), ProductStackCompose, state)
		_ = mock.ensureStackIntegrity(t.Context(), ProductStackAll, state)

		calls := mock.getCalls()
		if len(calls) != 5 {
			t.Fatalf("expected 5 calls, got %d", len(calls))
		}

		// verify specific calls
		expectedCalls := []struct {
			method string
			stack  ProductStack
		}{
			{"ensureStackIntegrity", ProductStackPentagi},
			{"verifyStackIntegrity", ProductStackLangfuse},
			{"cleanupStackFiles", ProductStackObservability},
			{"ensureStackIntegrity", ProductStackCompose},
			{"ensureStackIntegrity", ProductStackAll},
		}

		for i, expected := range expectedCalls {
			if calls[i].Method != expected.method || calls[i].Stack != expected.stack {
				t.Errorf("call %d: expected %s(%s), got %s(%s)",
					i, expected.method, expected.stack, calls[i].Method, calls[i].Stack)
			}
		}
	})

	t.Run("DockerOperations", func(t *testing.T) {
		mock := newBaseMockDockerOperations()
		state := testOperationState(t)

		// make mixed calls
		_ = mock.pullWorkerImage(t.Context(), state)
		_ = mock.removeMainDockerNetwork(t.Context(), state, "network1")
		_ = mock.pullDefaultImage(t.Context(), state)
		_ = mock.removeMainDockerNetwork(t.Context(), state, "network2")

		calls := mock.getCalls()
		if len(calls) != 4 {
			t.Fatalf("expected 4 calls, got %d", len(calls))
		}

		// verify network names are captured
		if calls[1].Name != "network1" || calls[3].Name != "network2" {
			t.Error("network names not captured correctly")
		}
	})
}

func TestMockOperations_ErrorIsolation(t *testing.T) {
	t.Run("StackSpecificErrors", func(t *testing.T) {
		mock := newBaseMockFileSystemOperations()
		state := testOperationState(t)

		// set error for specific stack+method combination
		mock.errOn["ensureStackIntegrity_"+string(ProductStackPentagi)] = fmt.Errorf("pentagi-specific error")

		// pentagi should fail
		err := mock.ensureStackIntegrity(t.Context(), ProductStackPentagi, state)
		if err == nil || err.Error() != "pentagi-specific error" {
			t.Errorf("expected pentagi-specific error, got %v", err)
		}

		// langfuse should succeed
		err = mock.ensureStackIntegrity(t.Context(), ProductStackLangfuse, state)
		if err != nil {
			t.Errorf("unexpected error for langfuse: %v", err)
		}
	})

	t.Run("MethodLevelErrors", func(t *testing.T) {
		mock := newBaseMockDockerOperations()
		state := testOperationState(t)

		// set error for all calls to specific method
		mock.setError("pullWorkerImage", fmt.Errorf("pull error"))

		// pullWorkerImage should fail
		err := mock.pullWorkerImage(t.Context(), state)
		if err == nil || err.Error() != "pull error" {
			t.Errorf("expected pull error, got %v", err)
		}

		// other methods should succeed
		err = mock.pullDefaultImage(t.Context(), state)
		if err != nil {
			t.Errorf("unexpected error for pullDefaultImage: %v", err)
		}
	})
}

func TestMockState_ComplexOperations(t *testing.T) {
	state := newMockState()

	t.Run("MultipleVariableOperations", func(t *testing.T) {
		// set multiple variables
		vars := map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
			"VAR3": "value3",
		}
		err := state.SetVars(vars)
		assertNoError(t, err)

		if !state.IsDirty() {
			t.Error("expected state to be dirty after SetVars")
		}

		// get specific variables
		names := []string{"VAR1", "VAR3", "VAR_MISSING"}
		result, present := state.GetVars(names)

		if !present["VAR1"] || !present["VAR3"] {
			t.Error("expected VAR1 and VAR3 to be present")
		}
		if present["VAR_MISSING"] {
			t.Error("expected VAR_MISSING to not be present")
		}

		if result["VAR1"].Value != "value1" || result["VAR3"].Value != "value3" {
			t.Error("unexpected variable values")
		}

		// reset specific variables
		err = state.ResetVars([]string{"VAR1", "VAR3"})
		assertNoError(t, err)

		// verify VAR2 still exists
		v, ok := state.GetVar("VAR2")
		if !ok || v.Value != "value2" {
			t.Error("VAR2 should still exist")
		}

		// verify VAR1 and VAR3 are gone
		_, ok = state.GetVar("VAR1")
		if ok {
			t.Error("VAR1 should be removed")
		}
	})

	t.Run("StackManagement", func(t *testing.T) {
		stack := []string{"pentagi", "langfuse", "observability"}
		err := state.SetStack(stack)
		assertNoError(t, err)

		retrievedStack := state.GetStack()
		if len(retrievedStack) != len(stack) {
			t.Fatalf("expected stack length %d, got %d", len(stack), len(retrievedStack))
		}

		for i, s := range stack {
			if retrievedStack[i] != s {
				t.Errorf("stack[%d]: expected %s, got %s", i, s, retrievedStack[i])
			}
		}
	})

	t.Run("EnvPathVerification", func(t *testing.T) {
		envPath := state.GetEnvPath()
		if envPath == "" {
			t.Error("expected non-empty env path")
		}

		// verify path contains expected components
		if !filepath.IsAbs(envPath) {
			t.Error("expected absolute path")
		}
		if !strings.Contains(envPath, ".env") {
			t.Error("expected path to contain .env")
		}
	})
}

func TestMockFiles_ComplexOperations(t *testing.T) {
	filesMock := newMockFiles()

	t.Run("DirectoryOperations", func(t *testing.T) {
		// setup directory structure
		filesMock.lists["/app"] = []string{"file1.go", "file2.go", "subdir/"}
		filesMock.lists["/app/subdir"] = []string{"file3.go", "file4.go"}

		// test directory existence
		if !filesMock.Exists("/app") {
			t.Error("expected /app to exist")
		}

		// test stat for directory
		info, err := filesMock.Stat("/app")
		assertNoError(t, err)
		if !info.IsDir() {
			t.Error("expected /app to be a directory")
		}

		// test list operation
		list, err := filesMock.List("/app")
		assertNoError(t, err)
		if len(list) != 3 {
			t.Fatalf("expected 3 items in /app, got %d", len(list))
		}
	})

	t.Run("FileOperations", func(t *testing.T) {
		content := []byte("package main\n\nfunc main() {}")
		filesMock.AddFile("/app/main.go", content)

		// test file existence
		if !filesMock.Exists("/app/main.go") {
			t.Error("expected /app/main.go to exist")
		}

		// test stat for file
		info, err := filesMock.Stat("/app/main.go")
		assertNoError(t, err)
		if info.IsDir() {
			t.Error("expected /app/main.go to be a file")
		}

		// test content retrieval
		retrieved, err := filesMock.GetContent("/app/main.go")
		assertNoError(t, err)
		if !bytes.Equal(retrieved, content) {
			t.Error("unexpected file content")
		}

		// test non-existent file
		_, err = filesMock.GetContent("/app/missing.go")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("CopyOperations", func(t *testing.T) {
		// perform multiple copy operations
		_ = filesMock.Copy("/src/file1.go", "/dst/file1.go", false)
		_ = filesMock.Copy("/src/file2.go", "/dst/file2.go", true)
		_ = filesMock.Copy("/src/file3.go", "/dst/file3.go", false)

		if len(filesMock.copies) != 3 {
			t.Fatalf("expected 3 copy operations, got %d", len(filesMock.copies))
		}

		// verify copy details
		if filesMock.copies[1].Rewrite != true {
			t.Error("expected second copy to have rewrite=true")
		}
		if filesMock.copies[0].Src != "/src/file1.go" || filesMock.copies[0].Dst != "/dst/file1.go" {
			t.Error("unexpected copy source/destination")
		}
	})

	t.Run("FileStatusOperations", func(t *testing.T) {
		// set different statuses
		filesMock.statuses["/app/file1.go"] = files.FileStatusModified
		filesMock.statuses["/app/file2.go"] = files.FileStatusMissing

		// check statuses
		status := filesMock.Check("/app/file1.go", "/workspace")
		if status != files.FileStatusModified {
			t.Errorf("expected FileStatusModified, got %v", status)
		}

		// check default status for unset file
		status = filesMock.Check("/app/file3.go", "/workspace")
		if status != files.FileStatusOK {
			t.Errorf("expected FileStatusOK for unset file, got %v", status)
		}
	})
}

// Test concurrent access to mock objects

func TestMockOperations_ConcurrentAccess(t *testing.T) {
	t.Run("FileSystemOperations", func(t *testing.T) {
		mock := newBaseMockFileSystemOperations()
		state := testOperationState(t)

		// concurrent access test
		var wg sync.WaitGroup
		numGoroutines := 10
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				stack := ProductStack(fmt.Sprintf("stack-%d", id))
				_ = mock.ensureStackIntegrity(t.Context(), stack, state)
			}(i)
		}

		wg.Wait()

		calls := mock.getCalls()
		if len(calls) != numGoroutines {
			t.Errorf("expected %d calls, got %d", numGoroutines, len(calls))
		}
	})

	t.Run("DockerOperations", func(t *testing.T) {
		mock := newBaseMockDockerOperations()
		state := testOperationState(t)

		// concurrent error setting and method calls
		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				mock.setError("pullWorkerImage", fmt.Errorf("error-%d", i))
				time.Sleep(time.Millisecond)
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				_ = mock.pullWorkerImage(t.Context(), state)
				time.Sleep(time.Millisecond)
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				_ = mock.pullDefaultImage(t.Context(), state)
				time.Sleep(time.Millisecond)
			}
		}()

		wg.Wait()

		calls := mock.getCalls()
		if len(calls) != 10 {
			t.Errorf("expected 10 calls, got %d", len(calls))
		}
	})
}

// Test edge cases and boundary conditions

func TestMockState_EdgeCases(t *testing.T) {
	state := newMockState()

	t.Run("EmptyOperations", func(t *testing.T) {
		// test empty variable operations
		result, present := state.GetVars([]string{})
		if len(result) != 0 || len(present) != 0 {
			t.Error("expected empty results for empty input")
		}

		// test resetting non-existent variables
		err := state.ResetVars([]string{"NON_EXISTENT"})
		assertNoError(t, err)

		// test setting empty stack
		err = state.SetStack([]string{})
		assertNoError(t, err)
		if len(state.GetStack()) != 0 {
			t.Error("expected empty stack")
		}
	})

	t.Run("StateTransitions", func(t *testing.T) {
		// test dirty state transitions
		state.dirty = false

		err := state.SetVar("TEST", "value")
		assertNoError(t, err)
		if !state.IsDirty() {
			t.Error("expected dirty state after SetVar")
		}

		err = state.Commit()
		assertNoError(t, err)
		if state.IsDirty() {
			t.Error("expected clean state after Commit")
		}

		err = state.Reset()
		assertNoError(t, err)
		if state.IsDirty() {
			t.Error("expected clean state after Reset")
		}
	})
}

func TestMockFiles_EdgeCases(t *testing.T) {
	filesMock := newMockFiles()

	t.Run("EmptyPaths", func(t *testing.T) {
		// test operations with empty paths
		exists := filesMock.Exists("")
		if exists {
			t.Error("empty path should not exist")
		}

		_, err := filesMock.GetContent("")
		if err == nil {
			t.Error("expected error for empty path")
		}

		list, err := filesMock.List("")
		assertNoError(t, err)
		if len(list) != 0 {
			t.Error("expected empty list for unset path")
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// test paths with special characters
		specialPaths := []string{
			"/path with spaces/file.txt",
			"/path/with/unicode/файл.txt",
			"/path/with/special!@#$%^&*()chars.txt",
		}

		for _, path := range specialPaths {
			filesMock.AddFile(path, []byte("content"))
			if !filesMock.Exists(path) {
				t.Errorf("file with special path should exist: %s", path)
			}
		}
	})
}

func TestMockCheckHandler_CompleteScenarios(t *testing.T) {
	t.Run("AllSystemsHealthy", func(t *testing.T) {
		handler := newMockCheckHandler()
		result := &checker.CheckResult{}

		// configure all systems as healthy
		handler.config.DockerApiAccessible = true
		handler.config.WorkerImageExists = true
		handler.config.PentagiInstalled = true
		handler.config.PentagiRunning = true
		handler.config.LangfuseInstalled = true
		handler.config.LangfuseRunning = true
		handler.config.ObservabilityInstalled = true
		handler.config.ObservabilityRunning = true
		handler.config.SysNetworkOK = true
		handler.config.SysCPUOK = true
		handler.config.SysMemoryOK = true
		handler.config.SysDiskFreeSpaceOK = true
		handler.config.InstallerIsUpToDate = true
		handler.config.PentagiIsUpToDate = true
		handler.config.LangfuseIsUpToDate = true
		handler.config.ObservabilityIsUpToDate = true

		err := handler.GatherAllInfo(context.Background(), result)
		assertNoError(t, err)

		// verify all systems report as healthy
		if !result.DockerApiAccessible || !result.WorkerImageExists ||
			!result.PentagiRunning || !result.LangfuseRunning ||
			!result.ObservabilityRunning || !result.SysNetworkOK ||
			!result.InstallerIsUpToDate {
			t.Error("expected all systems to be healthy")
		}
	})

	t.Run("PartialFailures", func(t *testing.T) {
		handler := newMockCheckHandler()
		result := &checker.CheckResult{}

		// configure partial failures
		handler.config.DockerApiAccessible = true
		handler.config.PentagiInstalled = false
		handler.config.LangfuseRunning = true
		handler.config.ObservabilityRunning = false
		handler.config.SysMemoryOK = false
		handler.config.UpdateServerAccessible = false

		err := handler.GatherAllInfo(context.Background(), result)
		assertNoError(t, err)

		// verify mixed states
		if !result.DockerApiAccessible {
			t.Error("expected Docker API to be accessible")
		}
		if result.PentagiInstalled {
			t.Error("expected PentAGI not to be installed")
		}
		if !result.LangfuseRunning {
			t.Error("expected Langfuse to be running")
		}
		if result.ObservabilityRunning {
			t.Error("expected Observability not to be running")
		}
		if result.SysMemoryOK {
			t.Error("expected memory check to fail")
		}
		if result.UpdateServerAccessible {
			t.Error("expected update server to be inaccessible")
		}
	})
}

// Test functions to verify mock implementations work correctly

type mockCtxStateFunc func(context.Context, *operationState) error
type mockCtxStackStateFunc func(context.Context, ProductStack, *operationState) error

type mockBaseTest[F mockCtxStateFunc | mockCtxStackStateFunc] struct {
	handler  F
	funcName string
	stack    ProductStack
}

func TestBaseMockFileSystemOperations(t *testing.T) {
	mock := newBaseMockFileSystemOperations()
	state := testOperationState(t)

	tests := []mockBaseTest[mockCtxStackStateFunc]{
		{
			handler:  mock.ensureStackIntegrity,
			funcName: "ensureStackIntegrity",
			stack:    ProductStackPentagi,
		},
		{
			handler:  mock.verifyStackIntegrity,
			funcName: "verifyStackIntegrity",
			stack:    ProductStackPentagi,
		},
		{
			handler:  mock.cleanupStackFiles,
			funcName: "cleanupStackFiles",
			stack:    ProductStackPentagi,
		},
	}

	for tid, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			// ensure clean state for error injections between subtests
			mock.setError(tt.funcName, nil)
			err := tt.handler(t.Context(), tt.stack, state)
			assertNoError(t, err)

			expCallId := tid * 2
			calls := mock.getCalls()
			if len(calls) != expCallId+1 || calls[expCallId].Method != tt.funcName || calls[expCallId].Stack != tt.stack {
				t.Fatalf("unexpected calls: %+v", calls)
			}

			// test error injection
			testErr := fmt.Errorf("test error")
			mock.setError(tt.funcName, testErr)
			err = tt.handler(t.Context(), tt.stack, state)
			if err != testErr {
				t.Errorf("expected error %v, got %v", testErr, err)
			}
			// clear injected error for next iterations using same method name
			mock.setError(tt.funcName, nil)
		})
	}
}

func TestBaseMockDockerOperations(t *testing.T) {
	mock := newBaseMockDockerOperations()
	state := testOperationState(t)

	// test methods without extra parameters
	tests := []mockBaseTest[mockCtxStateFunc]{
		{
			handler:  mock.pullWorkerImage,
			funcName: "pullWorkerImage",
		},
		{
			handler:  mock.pullDefaultImage,
			funcName: "pullDefaultImage",
		},
		{
			handler:  mock.removeWorkerContainers,
			funcName: "removeWorkerContainers",
		},
		{
			handler:  mock.removeWorkerImages,
			funcName: "removeWorkerImages",
		},
		{
			handler:  mock.purgeWorkerImages,
			funcName: "purgeWorkerImages",
		},
		{
			handler:  mock.ensureMainDockerNetworks,
			funcName: "ensureMainDockerNetworks",
		},
		{
			handler:  mock.removeWorkerVolumes,
			funcName: "removeWorkerVolumes",
		},
	}

	for tid, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			err := tt.handler(t.Context(), state)
			assertNoError(t, err)

			expCallId := tid * 2
			calls := mock.getCalls()
			if len(calls) != expCallId+1 || calls[expCallId].Method != tt.funcName {
				t.Fatalf("unexpected calls: %+v", calls)
			}

			// test error injection
			testErr := fmt.Errorf("docker error for %s", tt.funcName)
			mock.setError(tt.funcName, testErr)
			err = tt.handler(t.Context(), state)
			if err != testErr {
				t.Errorf("expected error %v, got %v", testErr, err)
			}
		})
	}
}

func TestBaseMockDockerOperations_WithParameters(t *testing.T) {
	mock := newBaseMockDockerOperations()
	state := testOperationState(t)

	t.Run("removeMainDockerNetwork", func(t *testing.T) {
		networkName := "test-network"
		err := mock.removeMainDockerNetwork(t.Context(), state, networkName)
		assertNoError(t, err)

		calls := mock.getCalls()
		if len(calls) != 1 || calls[0].Method != "removeMainDockerNetwork" || calls[0].Name != networkName {
			t.Fatalf("unexpected calls: %+v", calls)
		}

		// test error injection
		testErr := fmt.Errorf("network removal error")
		mock.setError("removeMainDockerNetwork", testErr)
		err = mock.removeMainDockerNetwork(t.Context(), state, networkName)
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
	})

	t.Run("removeMainImages", func(t *testing.T) {
		images := []string{"image1:tag", "image2:tag", "image3:tag"}
		err := mock.removeMainImages(t.Context(), state, images)
		assertNoError(t, err)

		calls := mock.getCalls()
		// offset by 2 due to previous test
		if len(calls) != 3 || calls[2].Method != "removeMainImages" {
			t.Fatalf("unexpected calls: %+v", calls)
		}

		// test error injection
		testErr := fmt.Errorf("image removal error")
		mock.setError("removeMainImages", testErr)
		err = mock.removeMainImages(t.Context(), state, images)
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
	})
}

func TestBaseMockComposeOperations(t *testing.T) {
	mock := newBaseMockComposeOperations()
	state := testOperationState(t)

	// test stack-based operations
	tests := []mockBaseTest[mockCtxStackStateFunc]{
		{
			handler:  mock.startStack,
			funcName: "startStack",
			stack:    ProductStackPentagi,
		},
		{
			handler:  mock.stopStack,
			funcName: "stopStack",
			stack:    ProductStackLangfuse,
		},
		{
			handler:  mock.restartStack,
			funcName: "restartStack",
			stack:    ProductStackObservability,
		},
		{
			handler:  mock.updateStack,
			funcName: "updateStack",
			stack:    ProductStackPentagi,
		},
		{
			handler:  mock.downloadStack,
			funcName: "downloadStack",
			stack:    ProductStackLangfuse,
		},
		{
			handler:  mock.removeStack,
			funcName: "removeStack",
			stack:    ProductStackObservability,
		},
		{
			handler:  mock.purgeStack,
			funcName: "purgeStack",
			stack:    ProductStackPentagi,
		},
		{
			handler:  mock.purgeImagesStack,
			funcName: "purgeImagesStack",
			stack:    ProductStackCompose,
		},
		{
			handler:  mock.purgeImagesStack,
			funcName: "purgeImagesStack",
			stack:    ProductStackAll,
		},
	}

	for tid, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			err := tt.handler(t.Context(), tt.stack, state)
			assertNoError(t, err)

			expCallId := tid * 2
			calls := mock.getCalls()
			if len(calls) != expCallId+1 || calls[expCallId].Method != tt.funcName || calls[expCallId].Stack != tt.stack {
				t.Fatalf("unexpected calls: %+v", calls)
			}

			// test error injection
			testErr := fmt.Errorf("compose error for %s", tt.funcName)
			mock.setError(tt.funcName, testErr)
			err = tt.handler(t.Context(), tt.stack, state)
			if err != testErr {
				t.Errorf("expected error %v, got %v", testErr, err)
			}
		})
	}
}

func TestBaseMockComposeOperations_SpecialMethods(t *testing.T) {
	mock := newBaseMockComposeOperations()
	state := testOperationState(t)

	t.Run("determineComposeFile", func(t *testing.T) {
		// test successful case
		file, err := mock.determineComposeFile(ProductStackPentagi)
		assertNoError(t, err)
		if file != "test-compose.yml" {
			t.Errorf("expected test-compose.yml, got %s", file)
		}

		calls := mock.getCalls()
		if len(calls) != 1 || calls[0].Method != "determineComposeFile" || calls[0].Stack != ProductStackPentagi {
			t.Fatalf("unexpected calls: %+v", calls)
		}

		// test error injection
		testErr := fmt.Errorf("compose file error")
		mock.setError("determineComposeFile", testErr)
		file, err = mock.determineComposeFile(ProductStackPentagi)
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
		if file != "" {
			t.Errorf("expected empty file on error, got %s", file)
		}
	})

	t.Run("performStackCommand", func(t *testing.T) {
		args := []string{"up", "-d", "--remove-orphans"}
		err := mock.performStackCommand(t.Context(), ProductStackPentagi, state, args...)
		assertNoError(t, err)

		calls := mock.getCalls()
		// offset by 2 due to previous test
		if len(calls) != 3 || calls[2].Method != "performStackCommand" || calls[2].Stack != ProductStackPentagi {
			t.Fatalf("unexpected calls: %+v", calls)
		}

		// test error injection
		testErr := fmt.Errorf("command execution error")
		mock.setError("performStackCommand", testErr)
		err = mock.performStackCommand(t.Context(), ProductStackPentagi, state, args...)
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
	})
}

func TestBaseMockUpdateOperations(t *testing.T) {
	mock := newBaseMockUpdateOperations()
	state := testOperationState(t)

	// test methods that return only error
	tests := []mockBaseTest[mockCtxStateFunc]{
		{
			handler:  mock.downloadInstaller,
			funcName: "downloadInstaller",
		},
		{
			handler:  mock.updateInstaller,
			funcName: "updateInstaller",
		},
		{
			handler:  mock.removeInstaller,
			funcName: "removeInstaller",
		},
	}

	for tid, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			err := tt.handler(t.Context(), state)
			assertNoError(t, err)

			expCallId := tid * 2
			calls := mock.getCalls()
			if len(calls) != expCallId+1 || calls[expCallId].Method != tt.funcName {
				t.Fatalf("unexpected calls: %+v", calls)
			}

			// test error injection
			testErr := fmt.Errorf("update error for %s", tt.funcName)
			mock.setError(tt.funcName, testErr)
			err = tt.handler(t.Context(), state)
			if err != testErr {
				t.Errorf("expected error %v, got %v", testErr, err)
			}
		})
	}

	// test checkUpdates separately as it returns a response
	t.Run("checkUpdates", func(t *testing.T) {
		resp, err := mock.checkUpdates(t.Context(), state)
		assertNoError(t, err)
		if resp == nil {
			t.Error("expected response, got nil")
		}

		calls := mock.getCalls()
		// offset by 6 due to previous tests (3 tests * 2 calls each)
		if len(calls) != 7 || calls[6].Method != "checkUpdates" {
			t.Fatalf("unexpected calls: %+v", calls)
		}

		// test error injection
		testErr := fmt.Errorf("check updates error")
		mock.setError("checkUpdates", testErr)
		resp, err = mock.checkUpdates(t.Context(), state)
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
		if resp != nil {
			t.Error("expected nil response on error")
		}
	})
}
