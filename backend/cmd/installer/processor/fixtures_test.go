package processor

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/loader"

	"github.com/stretchr/testify/require"
)

// mockState implements state.State over a map, with the environment file in a temp dir.
type mockState struct {
	vars    map[string]loader.EnvVar
	envPath string
	stack   []string
	dirty   bool
	// written stands in for the environment file on disk — the one `docker compose
	// --env-file` reads. Only WriteVars puts anything here; SetVars stages.
	written map[string]string
	// onReset runs inside Reset: the real Reset re-reads the file, so when it runs matters.
	onReset func()
}

func newMockState(t *testing.T) *mockState {
	t.Helper()
	envPath := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("PENTAGI_VERSION=1.0.0\n"), 0o644))
	return &mockState{vars: make(map[string]loader.EnvVar), envPath: envPath}
}

func (m *mockState) Reset() error {
	if m.onReset != nil {
		m.onReset()
	}
	m.dirty = false
	return nil
}

func (m *mockState) Exists() bool                             { return true }
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

// WriteVars records into `written` and leaves the state clean: a double that cannot tell
// staging from writing cannot catch a staged image reference compose never sees.
func (m *mockState) WriteVars(vars map[string]string) error {
	if m.written == nil {
		m.written = make(map[string]string, len(vars))
	}
	for name, value := range vars {
		m.written[name] = value
		m.vars[name] = loader.EnvVar{Name: name, Value: value}
	}
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

// mockFiles implements files.Files: `lists` are embedded directories, `content` embedded
// files, `statuses` what Check answers (OK when unset), and every Copy is recorded and
// answers `copyErr`.
type mockFiles struct {
	content  map[string][]byte
	statuses map[string]files.FileStatus
	lists    map[string][]string
	copyErr  error
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
	_, isFile := m.content[name]
	_, isDir := m.lists[name]
	return isFile || isDir
}

func (m *mockFiles) ExistsInFS(name string) bool { return false }

func (m *mockFiles) Stat(name string) (fs.FileInfo, error) {
	if _, exists := m.lists[name]; exists {
		return &mockFileInfo{name: name, isDir: true}, nil
	}
	if _, exists := m.content[name]; exists {
		return &mockFileInfo{name: name, isDir: false}, nil
	}
	return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
}

func (m *mockFiles) Copy(src, dst string, rewrite bool) error {
	m.copies = append(m.copies, struct {
		Src, Dst string
		Rewrite  bool
	}{Src: src, Dst: dst, Rewrite: rewrite})
	return m.copyErr
}

func (m *mockFiles) Check(name string, workingDir string) files.FileStatus {
	if status, exists := m.statuses[name]; exists {
		return status
	}
	return files.FileStatusOK
}

func (m *mockFiles) List(prefix string) ([]string, error) {
	return m.lists[prefix], nil
}

func (m *mockFiles) AddFile(name string, content []byte) {
	m.content[name] = content
}

type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 100 }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0o644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// call is one recorded call on an operations double.
type call struct {
	Method string
	Stack  ProductStack
	Name   string   // the network, for docker network operations
	Args   []string // the compose arguments, for performStackCommand
	Force  bool     // the operation state's force, for the file system operations
	Error  error
}

// processorCallRecorder is what every operations double shares: the calls it saw and the
// errors a test injected, keyed by method.
type processorCallRecorder struct {
	mu    sync.Mutex
	calls []call
	errOn map[string]error
}

func (m *processorCallRecorder) getCalls() []call {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]call(nil), m.calls...)
}

func (m *processorCallRecorder) setError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errOn == nil {
		m.errOn = make(map[string]error)
	}
	m.errOn[method] = err
}

// do records c and answers ctxErr — a done context fails a command or a daemon call
// started under it — or else the injected error. An injected error is spent by its first
// use when oneShot is set.
func (m *processorCallRecorder) do(c call, ctxErr error, oneShot bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.Error = ctxErr
	if c.Error == nil {
		if err, ok := m.errOn[c.Method+"_"+string(c.Stack)]; ok {
			c.Error = err
		} else if err, ok := m.errOn[c.Method]; ok {
			c.Error = err
			if oneShot {
				delete(m.errOn, c.Method)
			}
		}
	}
	m.calls = append(m.calls, c)
	return c.Error
}

// baseMockFileSystemOperations ignores its context, as the real file operations do, and
// records force, which decides whether a file the user edited is overwritten. An error is
// injected for one method ("ensureStackIntegrity") or one stack of it
// ("ensureStackIntegrity_langfuse"), the latter taking precedence.
type baseMockFileSystemOperations struct{ processorCallRecorder }

func (m *baseMockFileSystemOperations) ensureStackIntegrity(_ context.Context, stack ProductStack, state *operationState) error {
	return m.do(call{Method: "ensureStackIntegrity", Stack: stack, Force: state.force}, nil, false)
}

func (m *baseMockFileSystemOperations) verifyStackIntegrity(_ context.Context, stack ProductStack, state *operationState) error {
	return m.do(call{Method: "verifyStackIntegrity", Stack: stack, Force: state.force}, nil, false)
}

func (m *baseMockFileSystemOperations) checkStackIntegrity(_ context.Context, stack ProductStack) (FilesCheckResult, error) {
	if err := m.do(call{Method: "checkStackIntegrity", Stack: stack}, nil, false); err != nil {
		return nil, err
	}
	return FilesCheckResult{}, nil
}

func (m *baseMockFileSystemOperations) cleanupStackFiles(_ context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "cleanupStackFiles", Stack: stack}, nil, false)
}

// baseMockDockerOperations records the network name for the network operations.
type baseMockDockerOperations struct{ processorCallRecorder }

func (m *baseMockDockerOperations) pullWorkerImage(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "pullWorkerImage"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) pullDefaultImage(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "pullDefaultImage"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) removeWorkerContainers(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "removeWorkerContainers"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) removeWorkerImages(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "removeWorkerImages"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) purgeWorkerImages(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "purgeWorkerImages"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) ensureMainDockerNetworks(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "ensureMainDockerNetworks"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) removeMainDockerNetwork(ctx context.Context, _ *operationState, name string) error {
	return m.do(call{Method: "removeMainDockerNetwork", Name: name}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) removeMainImages(ctx context.Context, _ *operationState, _ []string) error {
	return m.do(call{Method: "removeMainImages"}, ctx.Err(), false)
}

func (m *baseMockDockerOperations) removeWorkerVolumes(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "removeWorkerVolumes"}, ctx.Err(), false)
}

// baseMockComposeOperations spends an injected error on its first use, so a fan-out fails
// on its first stack and goes on with the rest.
type baseMockComposeOperations struct{ processorCallRecorder }

func (m *baseMockComposeOperations) startStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "startStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) stopStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "stopStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) restartStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "restartStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) updateStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "updateStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) downloadStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "downloadStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) removeStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "removeStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) purgeStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "purgeStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) purgeImagesStack(ctx context.Context, stack ProductStack, _ *operationState) error {
	return m.do(call{Method: "purgeImagesStack", Stack: stack}, ctx.Err(), true)
}

func (m *baseMockComposeOperations) determineComposeFile(stack ProductStack) (string, error) {
	return "test-compose.yml", m.do(call{Method: "determineComposeFile", Stack: stack}, nil, true)
}

// performStackCommand records its arguments: "restart jaeger" and "up -d" are otherwise
// the same call.
func (m *baseMockComposeOperations) performStackCommand(ctx context.Context, stack ProductStack, _ *operationState, args ...string) error {
	return m.do(call{Method: "performStackCommand", Stack: stack, Args: args}, ctx.Err(), true)
}

// baseMockUpdateOperations reports jaegerPluginReplaced from updateJaegerPlugin, which
// decides whether the caller restarts the Jaeger service.
type baseMockUpdateOperations struct {
	processorCallRecorder
	jaegerPluginReplaced bool
}

func (m *baseMockUpdateOperations) checkUpdates(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "checkUpdates"}, ctx.Err(), false)
}

func (m *baseMockUpdateOperations) downloadInstaller(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "downloadInstaller"}, ctx.Err(), false)
}

func (m *baseMockUpdateOperations) updateInstaller(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "updateInstaller"}, ctx.Err(), false)
}

func (m *baseMockUpdateOperations) removeInstaller(ctx context.Context, _ *operationState) error {
	return m.do(call{Method: "removeInstaller"}, ctx.Err(), false)
}

func (m *baseMockUpdateOperations) updateJaegerPlugin(ctx context.Context, _ *operationState) (bool, error) {
	if err := m.do(call{Method: "updateJaegerPlugin"}, ctx.Err(), false); err != nil {
		return false, err
	}
	return m.jaegerPluginReplaced, nil
}

func (m *baseMockUpdateOperations) installerPackage(ctx context.Context) (*InstallerPackage, error) {
	if err := m.do(call{Method: "installerPackage"}, ctx.Err(), false); err != nil {
		return nil, err
	}
	return &InstallerPackage{Version: "1.2.3", OS: "linux", Arch: "amd64", Size: 1024, Path: "installer_1.2.3"}, nil
}

// mockCheckHandler implements checker.CheckHandler: every Gather* call is recorded and
// copies its part of config onto the result, so a refresh re-reads the configured state.
type mockCheckHandler struct {
	mu     sync.Mutex
	calls  []string
	config mockCheckConfig
}

type mockCheckConfig struct {
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

	PentagiScriptInstalled bool
	PentagiExtracted       bool
	PentagiInstalled       bool
	PentagiRunning         bool

	GraphitiConnected bool
	GraphitiExternal  bool
	GraphitiExtracted bool
	GraphitiInstalled bool
	GraphitiRunning   bool

	LangfuseConnected bool
	LangfuseExternal  bool
	LangfuseExtracted bool
	LangfuseInstalled bool
	LangfuseRunning   bool

	ObservabilityConnected bool
	ObservabilityExternal  bool
	ObservabilityExtracted bool
	ObservabilityInstalled bool
	ObservabilityRunning   bool

	SysNetworkOK       bool
	SysCPUOK           bool
	SysMemoryOK        bool
	SysDiskFreeSpaceOK bool

	UpdateServerAccessible  bool
	InstallerIsUpToDate     bool
	PentagiIsUpToDate       bool
	GraphitiIsUpToDate      bool
	LangfuseIsUpToDate      bool
	ObservabilityIsUpToDate bool
}

func newMockCheckHandler() *mockCheckHandler {
	return &mockCheckHandler{
		config: mockCheckConfig{
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
			GraphitiIsUpToDate:      true,
			LangfuseIsUpToDate:      true,
			ObservabilityIsUpToDate: true,
		},
	}
}

func (m *mockCheckHandler) getCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.calls...)
}

// apply records the call and copies the configured values under the lock.
func (m *mockCheckHandler) apply(method string, copyTo func(mockCheckConfig)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, method)
	copyTo(m.config)
	return nil
}

func (m *mockCheckHandler) GatherAllInfo(ctx context.Context, c *checker.CheckResult) error {
	_ = m.apply("GatherAllInfo", func(config mockCheckConfig) { c.EnvFileExists = config.EnvFileExists })
	for _, gather := range []func(context.Context, *checker.CheckResult) error{
		m.GatherDockerInfo, m.GatherWorkerInfo, m.GatherPentagiInfo, m.GatherGraphitiInfo,
		m.GatherLangfuseInfo, m.GatherObservabilityInfo, m.GatherSystemInfo, m.GatherUpdatesInfo,
	} {
		if err := gather(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockCheckHandler) GatherDockerInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherDockerInfo", func(config mockCheckConfig) {
		c.DockerApiAccessible = config.DockerApiAccessible
		c.DockerInstalled = config.DockerInstalled
		c.DockerComposeInstalled = config.DockerComposeInstalled
		c.DockerVersion = config.DockerVersion
		c.DockerVersionOK = config.DockerVersionOK
		c.DockerComposeVersion = config.DockerComposeVersion
		c.DockerComposeVersionOK = config.DockerComposeVersionOK
	})
}

func (m *mockCheckHandler) GatherWorkerInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherWorkerInfo", func(config mockCheckConfig) {
		c.WorkerEnvApiAccessible = config.WorkerEnvApiAccessible
		c.WorkerImageExists = config.WorkerImageExists
	})
}

func (m *mockCheckHandler) GatherPentagiInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherPentagiInfo", func(config mockCheckConfig) {
		c.PentagiScriptInstalled = config.PentagiScriptInstalled
		c.PentagiExtracted = config.PentagiExtracted
		c.PentagiInstalled = config.PentagiInstalled
		c.PentagiRunning = config.PentagiRunning
	})
}

func (m *mockCheckHandler) GatherGraphitiInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherGraphitiInfo", func(config mockCheckConfig) {
		c.GraphitiConnected = config.GraphitiConnected
		c.GraphitiExternal = config.GraphitiExternal
		c.GraphitiExtracted = config.GraphitiExtracted
		c.GraphitiInstalled = config.GraphitiInstalled
		c.GraphitiRunning = config.GraphitiRunning
	})
}

func (m *mockCheckHandler) GatherLangfuseInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherLangfuseInfo", func(config mockCheckConfig) {
		c.LangfuseConnected = config.LangfuseConnected
		c.LangfuseExternal = config.LangfuseExternal
		c.LangfuseExtracted = config.LangfuseExtracted
		c.LangfuseInstalled = config.LangfuseInstalled
		c.LangfuseRunning = config.LangfuseRunning
	})
}

func (m *mockCheckHandler) GatherObservabilityInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherObservabilityInfo", func(config mockCheckConfig) {
		c.ObservabilityConnected = config.ObservabilityConnected
		c.ObservabilityExternal = config.ObservabilityExternal
		c.ObservabilityExtracted = config.ObservabilityExtracted
		c.ObservabilityInstalled = config.ObservabilityInstalled
		c.ObservabilityRunning = config.ObservabilityRunning
	})
}

func (m *mockCheckHandler) GatherSystemInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherSystemInfo", func(config mockCheckConfig) {
		c.SysNetworkOK = config.SysNetworkOK
		c.SysCPUOK = config.SysCPUOK
		c.SysMemoryOK = config.SysMemoryOK
		c.SysDiskFreeSpaceOK = config.SysDiskFreeSpaceOK
	})
}

func (m *mockCheckHandler) GatherUpdatesInfo(_ context.Context, c *checker.CheckResult) error {
	return m.apply("GatherUpdatesInfo", func(config mockCheckConfig) {
		c.UpdateServerAccessible = config.UpdateServerAccessible
		c.InstallerIsUpToDate = config.InstallerIsUpToDate
		c.PentagiIsUpToDate = config.PentagiIsUpToDate
		c.GraphitiIsUpToDate = config.GraphitiIsUpToDate
		c.LangfuseIsUpToDate = config.LangfuseIsUpToDate
		c.ObservabilityIsUpToDate = config.ObservabilityIsUpToDate
	})
}

// processorHarness is a processor wired to recording doubles, with the check handler that
// answers its refreshes so a test can count what an operation asked the cloud.
type processorHarness struct {
	p       *processor
	fs      *baseMockFileSystemOperations
	docker  *baseMockDockerOperations
	compose *baseMockComposeOperations
	update  *baseMockUpdateOperations
	checks  *mockCheckHandler
}

func processorHarnessWith(t *testing.T, configure func(*mockCheckConfig)) *processorHarness {
	t.Helper()
	h := &processorHarness{
		fs:      &baseMockFileSystemOperations{},
		docker:  &baseMockDockerOperations{},
		compose: &baseMockComposeOperations{},
		update:  &baseMockUpdateOperations{},
		checks:  newMockCheckHandler(),
	}
	if configure != nil {
		configure(&h.checks.config)
	}
	result, err := checker.GatherWithHandler(context.Background(), h.checks)
	require.NoError(t, err)

	h.p = &processor{
		mu:         &sync.Mutex{},
		state:      newMockState(t),
		checker:    &result,
		files:      newMockFiles(),
		fsOps:      h.fs,
		dockerOps:  h.docker,
		composeOps: h.compose,
		updateOps:  h.update,
	}
	return h
}

func testOperationState(t *testing.T) *operationState {
	t.Helper()
	return &operationState{mx: &sync.Mutex{}, ctx: t.Context()}
}

// processorWithPullReference is a processor whose update answer names
// vxcontrol/pentagi:2.3.4 for the pentagi stack, with real compose operations.
func processorWithPullReference(t *testing.T) (*processor, *mockState) {
	t.Helper()
	state := newMockState(t)
	p := &processor{
		state: state,
		checker: &checker.CheckResult{
			StackUpdates: []checker.StackUpdate{{
				Stack: "pentagi",
				Components: []checker.ComponentUpdate{
					{Component: "pentagi", Action: "upgrade", PullReference: "vxcontrol/pentagi:2.3.4"},
				},
			}},
		},
	}
	p.composeOps = newComposeOperations(p)
	return p, state
}

// repositoryRoot walks up to the directory holding the compose files.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for range 8 {
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "no docker-compose.yml above the test directory")
		dir = parent
	}
	t.Fatal("could not find the repository root")
	return ""
}
