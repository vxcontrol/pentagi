package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge/limits"
	"pentagi/pkg/flowfiles"
	"pentagi/pkg/graphiti"
	"pentagi/pkg/providers/embeddings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toolsUnreachableDB refuses every connection at once, so a store opened on it fails as on a database that is down.
const toolsUnreachableDB = "postgres://pentagi@127.0.0.1:1/none?sslmode=disable&connect_timeout=2"

// toolsSharedPool stands in for the one pgxpool every flow's store shares; it never connects until used.
func toolsSharedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), toolsUnreachableDB)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

func toolsSandboxRow(status database.ContainerStatus, localID string) *database.Container {
	return &database.Container{ID: 7, Status: status, LocalID: sql.NullString{String: localID, Valid: localID != ""}}
}

func TestTools_Prepare_KeepsOrReplacesTheSandbox(t *testing.T) {
	errProbe := errors.New("Cannot connect to the Docker daemon")
	errUpdate := errors.New("connection reset by peer")
	errRemove := errors.New("removal of container primary is already in progress")
	errLaunch := errors.New("No such image: vxcontrol/kali-linux")

	cases := []struct {
		name        string
		row         *database.Container
		running     bool
		probeErr    error
		updateErr   error
		removeErr   error
		launchErr   error
		wantErr     error
		wantErrText string
		wantUpdates []database.UpdateContainerStatusParams
		wantRemoved []dockerRemoval
		wantLaunch  bool
		wantPrimary int64
		wantLocalID string
	}{
		{
			name:        "a running sandbox is kept",
			row:         toolsSandboxRow(database.ContainerStatusRunning, "primary"),
			running:     true,
			wantPrimary: 7,
			wantLocalID: "primary",
		},
		{
			name:        "a sandbox marked failed that runs again is kept and marked running",
			row:         toolsSandboxRow(database.ContainerStatusFailed, "primary"),
			running:     true,
			wantUpdates: []database.UpdateContainerStatusParams{{Status: database.ContainerStatusRunning, ID: 7}},
			wantPrimary: 7,
			wantLocalID: "primary",
		},
		{
			// a row without a Docker id cannot be inspected: probing it would fail the flow
			name:        "a sandbox that failed before docker created it is replaced",
			row:         toolsSandboxRow(database.ContainerStatusFailed, ""),
			running:     true,
			wantRemoved: []dockerRemoval{{localID: "", dbID: 7}},
			wantLaunch:  true,
			wantPrimary: 8,
			wantLocalID: "fresh",
		},
		{
			name:        "a sandbox that is gone is replaced",
			row:         toolsSandboxRow(database.ContainerStatusRunning, "primary"),
			running:     false,
			wantRemoved: []dockerRemoval{{localID: "primary", dbID: 7}},
			wantLaunch:  true,
			wantPrimary: 8,
			wantLocalID: "fresh",
		},
		{
			name:        "a stopped sandbox is replaced without a probe",
			row:         toolsSandboxRow(database.ContainerStatusStopped, "primary"),
			running:     true,
			wantRemoved: []dockerRemoval{{localID: "primary", dbID: 7}},
			wantLaunch:  true,
			wantPrimary: 8,
			wantLocalID: "fresh",
		},
		{
			name:        "a flow without a sandbox gets one",
			wantLaunch:  true,
			wantPrimary: 8,
			wantLocalID: "fresh",
		},
		{
			name:        "a sandbox docker cannot inspect is neither removed nor replaced",
			row:         toolsSandboxRow(database.ContainerStatusRunning, "primary"),
			probeErr:    errProbe,
			wantErr:     errProbe,
			wantErrText: "failed to inspect container 'pentagi-terminal-42'",
		},
		{
			name:        "a sandbox that cannot be marked running is not used",
			row:         toolsSandboxRow(database.ContainerStatusFailed, "primary"),
			running:     true,
			updateErr:   errUpdate,
			wantErr:     errUpdate,
			wantErrText: "failed to mark container 'pentagi-terminal-42' running",
			wantUpdates: []database.UpdateContainerStatusParams{{Status: database.ContainerStatusRunning, ID: 7}},
		},
		{
			name:        "a gone sandbox that cannot be removed is still replaced",
			row:         toolsSandboxRow(database.ContainerStatusRunning, "primary"),
			running:     false,
			removeErr:   errRemove,
			wantRemoved: []dockerRemoval{{localID: "primary", dbID: 7}},
			wantLaunch:  true,
			wantPrimary: 8,
			wantLocalID: "fresh",
		},
		{
			name:        "a sandbox that cannot be launched fails the preparation",
			launchErr:   errLaunch,
			wantErr:     errLaunch,
			wantErrText: "failed to launch container 'pentagi-terminal-42'",
			wantLaunch:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := &fakeContainerDB{row: tc.row, updateErr: tc.updateErr}
			docker := &fakeDockerClient{
				isRunning: tc.running,
				probeErr:  tc.probeErr,
				removeErr: tc.removeErr,
				launchErr: tc.launchErr,
			}
			fte := &flowToolsExecutor{db: db, cfg: &config.Config{DataDir: t.TempDir()}, docker: docker, flowID: 42}

			err := fte.Prepare(t.Context())

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Contains(t, err.Error(), tc.wantErrText)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantUpdates, db.updates)
			assert.Equal(t, tc.wantRemoved, docker.removed)
			assert.Equal(t, tc.wantLaunch, len(docker.launches) == 1, "launches: %+v", docker.launches)
			assert.Equal(t, tc.wantPrimary, fte.primaryID)
			assert.Equal(t, tc.wantLocalID, fte.primaryLID, "Release removes the container by this id")
		})
	}
}

func TestTools_Prepare_LaunchesTheSandboxWithTheCapabilityAllowList(t *testing.T) {
	allowList := []string{
		"CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
		"NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
	}

	cases := []struct {
		name     string
		tenant   string
		netAdmin bool
		wantName string
		wantCaps []string
	}{
		{name: "the default allow-list", wantName: "pentagi-terminal-42", wantCaps: allowList},
		{
			name:     "a tenant that allows NET_ADMIN",
			tenant:   "acme",
			netAdmin: true,
			wantName: "acme-pentagi-terminal-42",
			wantCaps: append(slices.Clone(allowList), "NET_ADMIN"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			docker := &fakeDockerClient{}
			cfg := &config.Config{DataDir: t.TempDir(), TenantID: tc.tenant, DockerNetAdmin: tc.netAdmin}
			executor, err := NewFlowToolsExecutor(&fakeContainerDB{}, cfg, docker, nil, 1, 42)
			require.NoError(t, err)
			executor.SetImage("vxcontrol/kali-linux:test")

			require.NoError(t, executor.Prepare(t.Context()))

			require.Len(t, docker.launches, 1)
			launch := docker.launches[0]
			assert.Equal(t, tc.wantName, launch.name)
			assert.Equal(t, database.ContainerTypePrimary, launch.kind)
			assert.Equal(t, int64(42), launch.flowID)
			assert.Equal(t, "vxcontrol/kali-linux:test", launch.config.Image)
			assert.Equal(t, []string{"tail", "-f", "/dev/null"}, launch.config.Entrypoint)
			assert.Equal(t, []string{"ALL"}, launch.hostConfig.CapDrop)
			assert.ElementsMatch(t, tc.wantCaps, launch.hostConfig.CapAdd)
		})
	}
}

// toolsFlowFilePaths are the sandbox paths of the regular files toolsWriteFlowFiles writes, the symlink left out.
var toolsFlowFilePaths = []string{"/work/uploads/targets/ips.txt", "/work/uploads/top.txt", "/work/resources/b.txt"}

func toolsWriteFlowFiles(t *testing.T, dataDir string) {
	t.Helper()
	uploads := flowfiles.FlowUploadsDir(dataDir, 42)
	resources := flowfiles.FlowResourcesDir(dataDir, 42)
	require.NoError(t, os.MkdirAll(filepath.Join(uploads, "targets"), 0o755))
	require.NoError(t, os.MkdirAll(resources, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(uploads, "targets", "ips.txt"), []byte("127.0.0.1"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(uploads, "top.txt"), []byte("top"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(resources, "b.txt"), []byte("bee"), 0o644))
	if err := os.Symlink(filepath.Join(uploads, "top.txt"), filepath.Join(uploads, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}
}

// One exec lists what is missing, one archive carries only that, and a symlink never leaves the host.
func TestTools_Prepare_SyncsTheFilesTheSandboxLacks(t *testing.T) {
	errCreate := errors.New("container pentagi-terminal-42 is not running")
	errAttach := errors.New("hijack: connection refused")
	errInspect := errors.New("No such exec instance: file-check")
	errCopy := errors.New("Error response from daemon: no space left on device")
	kept := toolsSandboxRow(database.ContainerStatusRunning, "primary")
	allMissing := strings.Join(toolsFlowFilePaths, "\n") + "\n"
	everyFile := map[string]string{"uploads/targets/ips.txt": "127.0.0.1", "uploads/top.txt": "top", "resources/b.txt": "bee"}

	cases := []struct {
		name        string
		row         *database.Container // nil: the sandbox is launched by this Prepare
		noFiles     bool
		output      string // what the file check prints: the paths it did not find
		exitCode    int
		createErr   error
		attachErr   error
		inspectErr  error
		copyErr     error
		wantErr     error
		wantErrText []string
		wantChecked bool
		wantCopies  int
		wantCopied  map[string]string
	}{
		{
			name:        "only the files the sandbox lacks are copied",
			row:         kept,
			output:      "/work/uploads/targets/ips.txt\n\n/work/resources/b.txt\n/work/unknown.txt\n",
			wantChecked: true,
			wantCopies:  1,
			wantCopied:  map[string]string{"uploads/targets/ips.txt": "127.0.0.1", "resources/b.txt": "bee"},
		},
		{
			name:        "a replaced sandbox receives every file",
			output:      allMissing,
			wantChecked: true,
			wantCopies:  1,
			wantCopied:  everyFile,
		},
		{
			name:        "nothing is copied when the sandbox has every file",
			row:         kept,
			wantChecked: true,
		},
		{
			name:    "a flow without files runs no check",
			row:     kept,
			noFiles: true,
		},
		{
			name:        "a file check that exits non-zero fails the preparation",
			row:         kept,
			output:      "sh: 1: printf: not found\n",
			exitCode:    2,
			wantErrText: []string{"failed to sync missing files to container 'pentagi-terminal-42'", "exit code 2", "printf: not found"},
			wantChecked: true,
		},
		{
			name:        "a replaced sandbox whose file check fails fails the preparation",
			output:      "sh: 1: printf: not found\n",
			exitCode:    2,
			wantErrText: []string{"failed to sync files to container 'pentagi-terminal-42'", "exit code 2"},
			wantChecked: true,
		},
		{
			name:        "a file check that cannot start fails the preparation",
			row:         kept,
			createErr:   errCreate,
			wantErr:     errCreate,
			wantErrText: []string{"failed to create file-check exec"},
			wantChecked: true,
		},
		{
			name:        "a file check that cannot be attached fails the preparation",
			row:         kept,
			attachErr:   errAttach,
			wantErr:     errAttach,
			wantErrText: []string{"failed to attach file-check exec"},
			wantChecked: true,
		},
		{
			name:        "a file check that cannot be inspected fails the preparation",
			row:         kept,
			inspectErr:  errInspect,
			wantErr:     errInspect,
			wantErrText: []string{"failed to inspect file-check exec"},
			wantChecked: true,
		},
		{
			name:        "an archive the sandbox refuses fails the preparation",
			row:         kept,
			output:      allMissing,
			copyErr:     errCopy,
			wantErr:     errCopy,
			wantChecked: true,
			wantCopies:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dataDir := t.TempDir()
			if !tc.noFiles {
				toolsWriteFlowFiles(t, dataDir)
			}
			docker := &fakeDockerClient{
				isRunning:      true,
				execCreateResp: client.ExecCreateResult{ID: "file-check"},
				attachOutput:   []byte(tc.output),
				inspectResp:    client.ExecInspectResult{ExitCode: tc.exitCode},
				execCreateErr:  tc.createErr,
				attachErr:      tc.attachErr,
				inspectErr:     tc.inspectErr,
				copyToErr:      tc.copyErr,
			}
			fte := &flowToolsExecutor{
				db: &fakeContainerDB{row: tc.row}, cfg: &config.Config{DataDir: dataDir}, docker: docker, flowID: 42,
			}

			done := make(chan error, 1)
			go func() { done <- fte.Prepare(t.Context()) }()
			var err error
			select {
			case err = <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("Prepare never returned: the archive writer is still blocked on its pipe")
			}

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			}
			if tc.wantErr != nil || tc.wantErrText != nil {
				require.Error(t, err)
				for _, part := range tc.wantErrText {
					assert.Contains(t, err.Error(), part)
				}
			} else {
				require.NoError(t, err)
			}

			if tc.wantChecked {
				assert.Equal(t, "pentagi-terminal-42", docker.execContainer)
				check := docker.execCreated.Cmd
				require.GreaterOrEqual(t, len(check), 4, "the check: %q", check)
				assert.Equal(t, []string{"sh", "-c"}, check[:2])
				assert.Equal(t, "--", check[3])
				assert.ElementsMatch(t, toolsFlowFilePaths, check[4:], "every regular file is checked, the symlink is not")
			} else {
				assert.Empty(t, docker.execCreated.Cmd, "a flow without files needs no check in the sandbox")
			}

			assert.Equal(t, tc.wantCopies, docker.copies)
			assert.Equal(t, tc.wantCopied, docker.copied)
			if tc.wantCopies > 0 {
				assert.Equal(t, "pentagi-terminal-42", docker.copiedInto)
				assert.Equal(t, "/work", docker.copiedTo)
			}
		})
	}
}

func TestTools_Release_RemovesTheSandboxAndClosesOnlyAStoreItOwns(t *testing.T) {
	errRemove := errors.New("Error response from daemon: container is paused")

	cases := []struct {
		name        string
		sharedPool  bool
		removeErr   error
		wantClosed  int
		wantErrText string
	}{
		{name: "a store it opened itself is closed", wantClosed: 1},
		{name: "a store on the shared pool is left open", sharedPool: true},
		{
			name:        "a sandbox it cannot remove is reported",
			removeErr:   errRemove,
			wantClosed:  1,
			wantErrText: "failed to purge container 'pentagi-terminal-42'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn := &fakeVectorConn{}
			docker := &fakeDockerClient{removeErr: tc.removeErr}
			fte := &flowToolsExecutor{
				cfg: &config.Config{}, docker: docker, flowID: 42,
				primaryID: 7, primaryLID: "primary", store: newFakeVectorStore(t, conn, &fakeEmbedder{}),
			}
			if tc.sharedPool {
				fte.cfg.PgxPool = toolsSharedPool(t)
			}

			err := fte.Release(t.Context())

			if tc.removeErr != nil {
				require.ErrorIs(t, err, tc.removeErr)
				assert.Contains(t, err.Error(), tc.wantErrText)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, []dockerRemoval{{localID: "primary", dbID: 7}}, docker.removed)
			assert.Equal(t, tc.wantClosed, conn.closed)
			assert.Nil(t, fte.store, "a released flow keeps no store")
		})
	}
}

// Opening a store needs a live PostgreSQL, so no row here ends with one.
func TestTools_SetEmbedder_ClosesOnlyAStoreItOwnsAndReportsOneItCannotOpen(t *testing.T) {
	cases := []struct {
		name         string
		embedder     embeddings.Embedder
		previous     bool // a store is already open
		sharedPool   bool
		wantClosed   int
		wantReported bool
	}{
		{name: "an unavailable embedder opens no store", embedder: &fakeEmbedder{unavailable: true}},
		{name: "a store that cannot be opened is reported", embedder: &fakeEmbedder{}, wantReported: true},
		{
			name:         "a store it opened itself is closed before the next one opens",
			embedder:     &fakeEmbedder{},
			previous:     true,
			wantClosed:   1,
			wantReported: true,
		},
		{
			name:         "a store on the shared pool is left open",
			embedder:     &fakeEmbedder{},
			previous:     true,
			sharedPool:   true,
			wantReported: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hook := captureLogrus(t)
			conn := &fakeVectorConn{}
			fte := &flowToolsExecutor{flowID: 7, cfg: &config.Config{DatabaseURL: toolsUnreachableDB}}
			if tc.previous {
				fte.store = newFakeVectorStore(t, conn, &fakeEmbedder{})
			}
			if tc.sharedPool {
				fte.cfg.PgxPool = toolsSharedPool(t)
			}

			fte.SetEmbedder(tc.embedder)

			assert.Equal(t, tc.embedder, fte.embedder)
			assert.Nil(t, fte.store, "a store was kept although none could be opened")
			assert.Equal(t, tc.wantClosed, conn.closed)
			var reported bool
			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "vector store") {
					reported = true
				}
			}
			assert.Equal(t, tc.wantReported, reported, "a flow that lost its memory tools must say so in the log: %v", hook.AllEntries())
		})
	}
}

func TestTools_NewFlowToolsExecutor_MasksTheConfiguredSecrets(t *testing.T) {
	// The replacer is built once per process from the first caller's config, so this test resets it.
	sharedReplacerOnce, sharedReplacer, sharedReplacerErr = sync.Once{}, nil, nil
	t.Cleanup(func() { sharedReplacerOnce, sharedReplacer, sharedReplacerErr = sync.Once{}, nil, nil })

	const licenseKey = "PLUM-WALRUS-7731-QUIVER"
	executor, err := NewFlowToolsExecutor(nil, &config.Config{LicenseKey: licenseKey}, nil, nil, 1, 42)
	require.NoError(t, err)

	masked := executor.(*flowToolsExecutor).replacer.ReplaceString("applied license " + licenseKey + " to the node")

	assert.NotContains(t, masked, licenseKey)
	assert.True(t, strings.HasPrefix(masked, "applied license "), "only the secret is replaced: %q", masked)
}

// toolsAgentExecutorCase is one agent's executor factory, built from a config whose handlers answer with their field name.
type toolsAgentExecutorCase struct {
	factory string
	variant string
	build   func(fte *flowToolsExecutor, everyService bool, without string) (ContextToolsExecutor, error)
	// tools maps each tool offered with every optional service to the config handler it reaches ("" for the executor's own).
	tools map[string]string
	// optional lists the tools, barriers included, that exist only with the optional services.
	optional []string
	barriers []string
	// required are the handler fields the factory refuses to build without.
	required []string
	// sandbox: the factory cannot build without the flow's primary container.
	sandbox       bool
	task, subtask int64 // 0: tool calls are not attributed to one
}

func toolsWithout[C any](cfg C, field string) C {
	if field != "" {
		reflect.ValueOf(&cfg).Elem().FieldByName(field).SetZero()
	}
	return cfg
}

// toolsAgentExecutorCases records every code pointer h hands out, so isConfigHandler holds whether or not closures share one.
func toolsAgentExecutorCases() ([]toolsAgentExecutorCase, func(ExecutorHandler) bool) {
	configHandlers := map[uintptr]bool{}
	h := func(field string) ExecutorHandler {
		handler := ExecutorHandler(func(context.Context, string, json.RawMessage) (string, error) { return field, nil })
		configHandlers[reflect.ValueOf(handler).Pointer()] = true
		return handler
	}
	isConfigHandler := func(handler ExecutorHandler) bool {
		return configHandlers[reflect.ValueOf(handler).Pointer()]
	}
	task, subtask := int64(3), int64(5)
	flowManager := FlowManagerHandlers{
		StopFlow:      func(context.Context, string) error { return nil },
		SendFlowInput: func(context.Context, string) error { return nil },
		PatchSubtasks: func(context.Context, int64, SubtaskPatch) error { return nil },
		WaitFlow:      func(context.Context) error { return nil },
	}
	assistant := func(useAgents bool) func(*flowToolsExecutor, bool, string) (ContextToolsExecutor, error) {
		return func(fte *flowToolsExecutor, everyService bool, without string) (ContextToolsExecutor, error) {
			cfg := AssistantExecutorConfig{
				UseAgents: useAgents,
				Adviser:   h("Adviser"), Coder: h("Coder"), Installer: h("Installer"),
				Memorist: h("Memorist"), Pentester: h("Pentester"), Searcher: h("Searcher"),
			}
			if everyService {
				cfg.FlowManager = flowManager
			}
			return fte.GetAssistantExecutor(toolsWithout(cfg, without))
		}
	}

	return []toolsAgentExecutorCase{
		{
			factory: "GetAssistantExecutor",
			build:   assistant(false),
			tools: map[string]string{
				"terminal": "", "file": "", "browser": "", "web_search": "",
				"search_in_memory": "", "search_guide": "", "search_answer": "", "search_code": "",
				"get_flow_status": "", "stop_flow": "", "submit_flow_input": "", "patch_flow_subtasks": "", "wait_flow_completion": "",
			},
			optional: []string{
				"browser", "web_search", "search_in_memory", "search_guide", "search_answer", "search_code",
				"stop_flow", "submit_flow_input", "patch_flow_subtasks", "wait_flow_completion",
			},
			required: []string{"Adviser", "Coder", "Installer", "Memorist", "Pentester", "Searcher"},
			sandbox:  true,
		},
		{
			factory: "GetAssistantExecutor",
			variant: "using agents",
			build:   assistant(true),
			tools: map[string]string{
				"terminal": "", "file": "", "browser": "", "web_search": "",
				"advice": "Adviser", "coder": "Coder", "maintenance": "Installer",
				"memorist": "Memorist", "pentester": "Pentester", "search": "Searcher",
				"get_flow_status": "", "stop_flow": "", "submit_flow_input": "", "patch_flow_subtasks": "", "wait_flow_completion": "",
			},
			optional: []string{
				"browser", "web_search", "stop_flow", "submit_flow_input", "patch_flow_subtasks", "wait_flow_completion",
			},
			sandbox: true,
		},
		{
			factory: "GetPrimaryExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetPrimaryExecutor(toolsWithout(PrimaryExecutorConfig{
					TaskID: task, SubtaskID: subtask,
					Barrier: h("Barrier"), Adviser: h("Adviser"), Coder: h("Coder"), Installer: h("Installer"),
					Memorist: h("Memorist"), Pentester: h("Pentester"), Searcher: h("Searcher"),
				}, without))
			},
			tools: map[string]string{
				"done": "Barrier", "ask": "Barrier", "advice": "Adviser", "coder": "Coder",
				"maintenance": "Installer", "memorist": "Memorist", "pentester": "Pentester", "search": "Searcher",
			},
			optional: []string{"ask"},
			barriers: []string{"done", "ask"},
			required: []string{"Barrier", "Adviser", "Coder", "Installer", "Memorist", "Pentester", "Searcher"},
			task:     task, subtask: subtask,
		},
		{
			factory: "GetInstallerExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetInstallerExecutor(toolsWithout(InstallerExecutorConfig{
					TaskID: &task, SubtaskID: &subtask,
					MaintenanceResult: h("MaintenanceResult"), Adviser: h("Adviser"),
					Memorist: h("Memorist"), Searcher: h("Searcher"),
				}, without))
			},
			tools: map[string]string{
				"maintenance_result": "MaintenanceResult", "advice": "Adviser", "memorist": "Memorist", "search": "Searcher",
				"terminal": "", "file": "", "browser": "", "store_guide": "", "search_guide": "",
			},
			optional: []string{"browser", "store_guide", "search_guide"},
			barriers: []string{"maintenance_result"},
			required: []string{"MaintenanceResult", "Adviser", "Memorist", "Searcher"},
			sandbox:  true,
			task:     task, subtask: subtask,
		},
		{
			factory: "GetCoderExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetCoderExecutor(toolsWithout(CoderExecutorConfig{
					TaskID: &task, SubtaskID: &subtask,
					CodeResult: h("CodeResult"), Adviser: h("Adviser"), Installer: h("Installer"),
					Memorist: h("Memorist"), Searcher: h("Searcher"),
				}, without))
			},
			tools: map[string]string{
				"code_result": "CodeResult", "advice": "Adviser", "maintenance": "Installer",
				"memorist": "Memorist", "search": "Searcher",
				"terminal": "", "file": "", "browser": "", "search_code": "", "store_code": "", "graphiti_search": "",
			},
			optional: []string{"browser", "search_code", "store_code", "graphiti_search"},
			barriers: []string{"code_result"},
			required: []string{"CodeResult", "Adviser", "Installer", "Memorist", "Searcher"},
			sandbox:  true,
			task:     task, subtask: subtask,
		},
		{
			factory: "GetPentesterExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetPentesterExecutor(toolsWithout(PentesterExecutorConfig{
					TaskID: &task, SubtaskID: &subtask,
					HackResult: h("HackResult"), Adviser: h("Adviser"), Coder: h("Coder"),
					Installer: h("Installer"), Memorist: h("Memorist"), Searcher: h("Searcher"),
				}, without))
			},
			tools: map[string]string{
				"hack_result": "HackResult", "advice": "Adviser", "coder": "Coder", "maintenance": "Installer",
				"memorist": "Memorist", "search": "Searcher",
				"terminal": "", "file": "", "browser": "", "store_guide": "", "search_guide": "",
				"graphiti_search": "", "web_search": "",
			},
			optional: []string{"browser", "store_guide", "search_guide", "graphiti_search", "web_search"},
			barriers: []string{"hack_result"},
			required: []string{"HackResult", "Adviser", "Coder", "Installer", "Memorist", "Searcher"},
			sandbox:  true,
			task:     task, subtask: subtask,
		},
		{
			factory: "GetSearcherExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetSearcherExecutor(toolsWithout(SearcherExecutorConfig{
					TaskID: &task, SubtaskID: &subtask,
					SearchResult: h("SearchResult"), Memorist: h("Memorist"),
				}, without))
			},
			tools: map[string]string{
				"search_result": "SearchResult", "memorist": "Memorist",
				"browser": "", "web_search": "", "search_answer": "", "store_answer": "",
			},
			optional: []string{"browser", "web_search", "search_answer", "store_answer"},
			barriers: []string{"search_result"},
			required: []string{"SearchResult", "Memorist"},
			task:     task, subtask: subtask,
		},
		{
			factory: "GetGeneratorExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetGeneratorExecutor(toolsWithout(GeneratorExecutorConfig{
					TaskID:      task,
					SubtaskList: h("SubtaskList"), Memorist: h("Memorist"), Searcher: h("Searcher"),
				}, without))
			},
			tools: map[string]string{
				"subtask_list": "SubtaskList", "memorist": "Memorist", "search": "Searcher",
				"terminal": "", "file": "", "browser": "",
			},
			optional: []string{"browser"},
			barriers: []string{"subtask_list"},
			required: []string{"SubtaskList", "Memorist"},
			sandbox:  true,
			task:     task,
		},
		{
			factory: "GetRefinerExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetRefinerExecutor(toolsWithout(RefinerExecutorConfig{
					TaskID:       task,
					SubtaskPatch: h("SubtaskPatch"), Memorist: h("Memorist"), Searcher: h("Searcher"),
				}, without))
			},
			tools: map[string]string{
				"subtask_patch": "SubtaskPatch", "memorist": "Memorist", "search": "Searcher",
				"terminal": "", "file": "", "browser": "",
			},
			optional: []string{"browser"},
			barriers: []string{"subtask_patch"},
			required: []string{"SubtaskPatch", "Memorist"},
			sandbox:  true,
			task:     task,
		},
		{
			factory: "GetMemoristExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetMemoristExecutor(toolsWithout(MemoristExecutorConfig{
					TaskID: &task, SubtaskID: &subtask, SearchResult: h("SearchResult"),
				}, without))
			},
			tools: map[string]string{
				"memorist_result": "SearchResult", "terminal": "", "file": "", "search_in_memory": "", "graphiti_search": "",
			},
			optional: []string{"search_in_memory", "graphiti_search"},
			barriers: []string{"memorist_result"},
			required: []string{"SearchResult"},
			sandbox:  true,
			task:     task, subtask: subtask,
		},
		{
			factory: "GetEnricherExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetEnricherExecutor(toolsWithout(EnricherExecutorConfig{
					TaskID: &task, SubtaskID: &subtask, EnricherResult: h("EnricherResult"),
				}, without))
			},
			tools: map[string]string{
				"enricher_result": "EnricherResult", "terminal": "", "file": "",
				"search_in_memory": "", "graphiti_search": "", "browser": "",
			},
			optional: []string{"search_in_memory", "graphiti_search", "browser"},
			barriers: []string{"enricher_result"},
			required: []string{"EnricherResult"},
			sandbox:  true,
			task:     task, subtask: subtask,
		},
		{
			factory: "GetReporterExecutor",
			build: func(fte *flowToolsExecutor, _ bool, without string) (ContextToolsExecutor, error) {
				return fte.GetReporterExecutor(toolsWithout(ReporterExecutorConfig{
					TaskID: &task, SubtaskID: &subtask, ReportResult: h("ReportResult"),
				}, without))
			},
			tools:    map[string]string{"report_result": "ReportResult"},
			barriers: []string{"report_result"},
			required: []string{"ReportResult"},
			task:     task, subtask: subtask,
		},
	}, isConfigHandler
}

func toolsAgentFlow(t *testing.T, everyService bool) *flowToolsExecutor {
	t.Helper()
	fte := &flowToolsExecutor{
		userID:   1,
		flowID:   42,
		cfg:      &config.Config{},
		db:       &fakeContainerDB{row: toolsSandboxRow(database.ContainerStatusRunning, "primary")},
		docker:   &fakeDockerClient{},
		replacer: identityReplacer{},
	}
	if !everyService {
		return fte
	}

	graphitiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(graphitiServer.Close)
	graphitiClient, err := graphiti.NewClient(graphitiServer.URL, 5*time.Second, true)
	require.NoError(t, err)

	fte.cfg = &config.Config{AskUser: true, ScraperPrivateURL: "http://scraper.invalid", DuckDuckGoEnabled: true}
	fte.store = newFakeVectorStore(t, &fakeVectorConn{}, &fakeEmbedder{})
	fte.embedder = &fakeEmbedder{}
	fte.SetGraphitiClient(graphitiClient)
	return fte
}

// Covers the Get*Executor factories; subtests are keyed by factory.
func TestTools_GetExecutors_GiveEachAgentItsTools(t *testing.T) {
	cases, isConfigHandler := toolsAgentExecutorCases()
	for _, tc := range cases {
		for _, everyService := range []bool{false, true} {
			name := strings.TrimSpace(tc.factory + " " + tc.variant)
			if everyService {
				name += "/with every optional service"
			} else {
				name += "/without optional services"
			}
			t.Run(name, func(t *testing.T) {
				fte := toolsAgentFlow(t, everyService)

				executor, err := tc.build(fte, everyService, "")
				require.NoError(t, err)
				ce := executor.(*customExecutor)

				var wantTools, wantBarriers []string
				for tool := range tc.tools {
					if everyService || !slices.Contains(tc.optional, tool) {
						wantTools = append(wantTools, tool)
					}
				}
				for _, barrier := range tc.barriers {
					if everyService || !slices.Contains(tc.optional, barrier) {
						wantBarriers = append(wantBarriers, barrier)
					}
				}

				var advertised []string
				for _, def := range ce.definitions {
					advertised = append(advertised, def.Name)
				}
				assert.ElementsMatch(t, wantTools, advertised, "the tools the model is offered")
				assert.ElementsMatch(t, advertised, slices.Collect(maps.Keys(ce.handlers)),
					"every tool offered has a handler, and no handler serves a tool that is not offered")
				for _, tool := range wantTools {
					handler := ce.handlers[tool]
					if !assert.NotNil(t, handler, "the %s tool is offered with no handler", tool) {
						continue
					}
					field := tc.tools[tool]
					if field == "" {
						assert.False(t, isConfigHandler(handler), "the agent's own %s tool is served by a handler from its config", tool)
						continue
					}
					got, err := handler(t.Context(), tool, nil)
					require.NoError(t, err)
					assert.Equal(t, field, got, "the %s tool reaches the wrong handler", tool)
				}
				assert.ElementsMatch(t, wantBarriers, slices.Collect(maps.Keys(ce.barriers)), "the tools that end the agent's run")

				assert.Equal(t, int64(42), ce.flowID)
				assert.Equal(t, int64(1), ce.userID)
				assert.Same(t, fte.store, ce.store, "tool results reach the flow's memory")
				assert.Equal(t, fte.replacer, ce.replacer)
				if tc.task == 0 {
					assert.Nil(t, ce.taskID)
				} else if assert.NotNil(t, ce.taskID) {
					assert.Equal(t, tc.task, *ce.taskID)
				}
				if tc.subtask == 0 {
					assert.Nil(t, ce.subtaskID)
				} else if assert.NotNil(t, ce.subtaskID) {
					assert.Equal(t, tc.subtask, *ce.subtaskID)
				}
			})
		}
	}
}

// Covers the Get*Executor factories; subtests are keyed by factory. The terminal log carries the agent's task and subtask.
func TestTools_GetExecutors_BindTheTerminalToTheFlowSandbox(t *testing.T) {
	cases, _ := toolsAgentExecutorCases()
	for _, tc := range cases {
		if _, ok := tc.tools["terminal"]; !ok {
			continue
		}
		t.Run(strings.TrimSpace(tc.factory+" "+tc.variant), func(t *testing.T) {
			fte := toolsAgentFlow(t, false)
			fte.cfg.TenantID = "acme"
			docker := &fakeDockerClient{
				isRunning:       true,
				attachOutput:    []byte("uid=0(root)"),
				readFileContent: "10.0.0.7 target",
			}
			termLog := &recordingTermLog{}
			fte.docker, fte.tlp = docker, termLog

			executor, err := tc.build(fte, false, "")
			require.NoError(t, err)
			handlers := executor.(*customExecutor).handlers
			// a hung call fails at this deadline instead of at the package timeout
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			ran, err := handlers["terminal"](ctx, "terminal", json.RawMessage(`{"input":"id","cwd":"/work"}`))
			require.NoError(t, err)
			read, err := handlers["file"](ctx, "file", json.RawMessage(`{"action":"read_file","path":"/work/notes.txt"}`))
			require.NoError(t, err)

			assert.Contains(t, ran, "uid=0(root)")
			assert.Equal(t, "10.0.0.7 target", read)
			assert.Equal(t, []string{"primary", "primary"}, docker.probed, "each tool checks the flow's own container")
			assert.Equal(t, "acme-pentagi-terminal-42", docker.execContainer)
			// 0: the write is not attributed to a task or subtask
			type filedUnder struct{ container, task, subtask int64 }
			entries := termLog.all()
			require.NotEmpty(t, entries)
			for _, entry := range entries {
				assert.Equal(t, filedUnder{container: 7, task: tc.task, subtask: tc.subtask},
					filedUnder{container: entry.containerID, task: deref(entry.taskID), subtask: deref(entry.subtaskID)},
					"the terminal log is filed under the wrong container, task or subtask")
			}
		})
	}
}

// Covers the Get*Executor factories; subtests are keyed by factory.
func TestTools_GetExecutors_RequireWhatTheyUse(t *testing.T) {
	messages := map[string]string{
		"Barrier":           "barrier (done) handler is required",
		"Adviser":           "adviser handler is required",
		"Coder":             "coder handler is required",
		"Installer":         "installer handler is required",
		"Memorist":          "memorist handler is required",
		"Pentester":         "pentester handler is required",
		"Searcher":          "searcher handler is required",
		"MaintenanceResult": "maintenance result handler is required",
		"CodeResult":        "code result handler is required",
		"HackResult":        "hack result handler is required",
		"SearchResult":      "search result handler is required",
		"SubtaskList":       "subtask list handler is required",
		"SubtaskPatch":      "subtask patch handler is required",
		"EnricherResult":    "enricher result handler is required",
		"ReportResult":      "report result handler is required",
	}
	errLookup := errors.New("sql: connection is already closed")

	cases, _ := toolsAgentExecutorCases()
	seen := map[string]bool{}
	for _, tc := range cases {
		if seen[tc.factory] {
			continue
		}
		seen[tc.factory] = true

		for _, field := range tc.required {
			t.Run(tc.factory+"/without its "+field+" handler", func(t *testing.T) {
				_, err := tc.build(toolsAgentFlow(t, false), false, field)

				assert.EqualError(t, err, messages[field])
			})
		}

		t.Run(tc.factory+"/without a sandbox", func(t *testing.T) {
			fte := toolsAgentFlow(t, false)
			fte.db = &fakeContainerDB{lookupErr: errLookup}

			_, err := tc.build(fte, false, "")

			if !tc.sandbox {
				assert.NoError(t, err, "this agent runs no command, so it needs no sandbox")
				return
			}
			require.ErrorIs(t, err, errLookup)
			assert.Contains(t, err.Error(), "failed to get container 42")
		})
	}
}

func TestTools_Functions_ScansWhatTheDatabaseDriverReturns(t *testing.T) {
	const stored = `{"token":"t-1","disabled":[{"name":"terminal","context":["agent"]}],` +
		`"functions":[{"name":"lookup","url":"https://example.com/api","timeout":30}]}`

	cases := []struct {
		name    string
		input   any
		wantErr string
	}{
		{name: "a string", input: stored},
		{name: "bytes", input: []byte(stored)},
		{name: "a raw JSON message", input: json.RawMessage(stored)},
		{name: "an integer", input: int64(1), wantErr: "unsupported type of input value to scan"},
		{name: "malformed JSON", input: `{"token":`, wantErr: "unexpected end of JSON input"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var functions Functions

			err := functions.Scan(tc.input)

			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, functions.Token)
			assert.Equal(t, "t-1", *functions.Token)
			assert.Equal(t, []DisableFunction{{Name: "terminal", Context: []string{"agent"}}}, functions.Disabled)
			require.Len(t, functions.Function, 1)
			assert.Equal(t, "lookup", functions.Function[0].Name)
			assert.Equal(t, "https://example.com/api", functions.Function[0].URL)
			if assert.NotNil(t, functions.Function[0].Timeout) {
				assert.Equal(t, int64(30), *functions.Function[0].Timeout)
			}
		})
	}
}

func TestTools_EnrichLogrusFields_TagsAnEntryWithTheIDsItHas(t *testing.T) {
	task, subtask := int64(3), int64(5)

	cases := []struct {
		name      string
		taskID    *int64
		subtaskID *int64
		fields    logrus.Fields
		want      logrus.Fields
	}{
		{
			name:      "every id on an entry without fields",
			taskID:    &task,
			subtaskID: &subtask,
			want:      logrus.Fields{"flow_id": int64(42), "task_id": int64(3), "subtask_id": int64(5)},
		},
		{
			name:   "the flow alone beside the entry's own fields",
			fields: logrus.Fields{"tool": "terminal"},
			want:   logrus.Fields{"tool": "terminal", "flow_id": int64(42)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, enrichLogrusFields(42, tc.taskID, tc.subtaskID, tc.fields))
		})
	}
}

type toolsExpandingReplacer struct {
	from string
	to   string
}

func (r toolsExpandingReplacer) ReplaceString(s string) string {
	return strings.ReplaceAll(s, r.from, r.to)
}
func (r toolsExpandingReplacer) ReplaceBytes(b []byte) []byte {
	return []byte(r.ReplaceString(string(b)))
}
func (r toolsExpandingReplacer) WrapReader(reader io.Reader) io.Reader {
	return reader
}

func TestTools_BoundedContent_BoundsWhatAnonymizationProduced(t *testing.T) {
	replacer := toolsExpandingReplacer{from: "ip", to: "<REDACTED-ADDRESS-PLACEHOLDER>"}
	raw := strings.Repeat("ip", limits.MaxContentLen)

	bounded := boundedContent(replacer, raw)

	if got := len([]rune(bounded)); got != limits.MaxContentLen {
		t.Errorf("anonymization grew the text and the bound was applied to the wrong form: %d runes stored", got)
	}
	if strings.Contains(bounded, "ip") {
		t.Error("the stored form carries text the replacer was supposed to remove")
	}
}

// Scans the source, so a door added later is caught before a runtime test covers it.
func TestTools_VectorStoreLogProvider_NoDoorLogsARawArgument(t *testing.T) {
	fset := token.NewFileSet()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}

	calls := 0

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}

		redacted := map[string]bool{}

		ast.Inspect(file, func(node ast.Node) bool {
			switch declaration := node.(type) {
			case *ast.AssignStmt:
				if len(declaration.Lhs) != 1 || len(declaration.Rhs) != 1 {
					return true
				}

				if name, ok := declaration.Lhs[0].(*ast.Ident); ok && anonymizes(declaration.Rhs[0]) {
					redacted[name.Name] = true
				}
			case *ast.ValueSpec:
				for at, name := range declaration.Names {
					if at < len(declaration.Values) && anonymizes(declaration.Values[at]) {
						redacted[name.Name] = true
					}
				}
			}

			return true
		})

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "PutLog" || len(call.Args) < 5 {
				return true
			}

			receiver, ok := selector.X.(*ast.SelectorExpr)
			if !ok || receiver.Sel.Name != "vslp" {
				return true
			}

			calls++
			query := call.Args[4]

			if identifier, ok := query.(*ast.Ident); ok && redacted[identifier.Name] {
				return true
			}

			if anonymizes(query) {
				return true
			}

			t.Errorf(
				"%s: PutLog is handed %s as the query — the vector store tab renders it verbatim, so it must be the anonymized form",
				fset.Position(call.Pos()), types.ExprString(query),
			)

			return true
		})
	}

	if calls < 5 {
		t.Errorf("found only %d PutLog calls on the vector store log; the sweep is not seeing every door", calls)
	}
}

func anonymizes(expression ast.Expr) bool {
	found := false

	ast.Inspect(expression, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if fun.Sel.Name == "ReplaceString" {
				found = true
			}
		case *ast.Ident:
			if fun.Name == "boundedContent" {
				found = true
			}
		}

		return !found
	})

	return found
}
