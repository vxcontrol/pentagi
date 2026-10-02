package processor

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logicOperation func(*processor, context.Context, ProductStack, *operationState) error

var (
	logicApplyChanges logicOperation = func(p *processor, ctx context.Context, _ ProductStack, s *operationState) error {
		return p.applyChanges(ctx, s)
	}
	logicInstall logicOperation = func(p *processor, ctx context.Context, _ ProductStack, s *operationState) error {
		return p.install(ctx, s)
	}
	logicFactoryReset logicOperation = func(p *processor, ctx context.Context, _ ProductStack, s *operationState) error {
		return p.factoryReset(ctx, s)
	}

	// logicEmbedded points every optional stack at its embedded endpoint.
	logicEmbedded = map[string]string{
		"OTEL_HOST":         checker.DefaultObservabilityEndpoint,
		"LANGFUSE_BASE_URL": checker.DefaultLangfuseEndpoint,
		"GRAPHITI_URL":      checker.DefaultGraphitiEndpoint,
	}
)

// logicCase is one operation run against recording doubles, and every call it must make.
type logicCase struct {
	name      string
	run       logicOperation
	stack     ProductStack
	configure func(*mockCheckConfig)
	vars      map[string]string
	// fail reaches every double, each answering its own methods; "method_stack" fails one fs stack.
	fail     map[string]error
	replaced bool                    // what updateJaegerPlugin reports
	prepare  func(*processorHarness) // runs on the harness just before the operation
	wantErr  string

	fs, compose, docker, update []string
}

func logicRun(t *testing.T, tc logicCase) *processorHarness {
	t.Helper()
	h := processorHarnessWith(t, tc.configure)
	h.update.jaegerPluginReplaced = tc.replaced
	for name, value := range tc.vars {
		require.NoError(t, h.p.state.SetVar(name, value))
	}
	for method, err := range tc.fail {
		for _, double := range []*processorCallRecorder{
			&h.fs.processorCallRecorder, &h.docker.processorCallRecorder,
			&h.compose.processorCallRecorder, &h.update.processorCallRecorder,
		} {
			double.setError(method, err)
		}
	}

	if tc.prepare != nil {
		tc.prepare(h)
	}

	err := tc.run(h.p, t.Context(), tc.stack, testOperationState(t))
	if tc.wantErr == "" {
		require.NoError(t, err)
	} else {
		require.EqualError(t, err, tc.wantErr)
	}
	assert.Equal(t, tc.fs, logicTrace(h.fs.getCalls()), "file system calls")
	assert.Equal(t, tc.compose, logicTrace(h.compose.getCalls()), "compose calls")
	assert.Equal(t, tc.docker, logicTrace(h.docker.getCalls()), "docker calls")
	assert.Equal(t, tc.update, logicTrace(h.update.getCalls()), "update calls")
	return h
}

// logicTrace renders calls as "method stack-or-network args [force]" lines, to compare a whole sequence.
func logicTrace(calls []call) []string {
	var trace []string
	for _, c := range calls {
		fields := append([]string{c.Method, string(c.Stack), c.Name}, c.Args...)
		if c.Force {
			fields = append(fields, "force")
		}
		trace = append(trace, strings.Join(strings.Fields(strings.Join(fields, " ")), " "))
	}
	return trace
}

// logicOutdated makes the check call the given stacks outdated.
func logicOutdated(stacks ...ProductStack) func(*mockCheckConfig) {
	return func(c *mockCheckConfig) {
		for _, stack := range stacks {
			switch stack {
			case ProductStackPentagi:
				c.PentagiIsUpToDate = false
			case ProductStackGraphiti:
				c.GraphitiIsUpToDate = false
			case ProductStackLangfuse:
				c.LangfuseIsUpToDate = false
			case ProductStackObservability:
				c.ObservabilityIsUpToDate = false
			case ProductStackInstaller:
				c.InstallerIsUpToDate = false
			}
		}
	}
}

func logicConnected(c *mockCheckConfig) {
	c.LangfuseConnected, c.GraphitiConnected = true, true
}

func countCalls(calls []string, name string) int {
	found := 0
	for _, call := range calls {
		if call == name {
			found++
		}
	}
	return found
}

// Subtests are keyed by unit; a refused stack must not reach compose.
func TestLogic_StartStopAndRestartDelegateToComposeOrRefuseTheStack(t *testing.T) {
	start, stop, restart := (*processor).start, (*processor).stop, (*processor).restart
	for _, tc := range []logicCase{
		{name: "start hands pentagi to compose", run: start, stack: ProductStackPentagi,
			compose: []string{"startStack pentagi"}},
		{name: "start refuses the worker", run: start, stack: ProductStackWorker,
			wantErr: "operation start not applicable for stack worker"},
		{name: "stop hands pentagi to compose", run: stop, stack: ProductStackPentagi,
			compose: []string{"stopStack pentagi"}},
		{name: "stop refuses the worker", run: stop, stack: ProductStackWorker,
			wantErr: "operation stop not applicable for stack worker"},
		{name: "restart hands pentagi to compose", run: restart, stack: ProductStackPentagi,
			compose: []string{"restartStack pentagi"}},
		{name: "restart refuses the worker", run: restart, stack: ProductStackWorker,
			wantErr: "operation restart not applicable for stack worker"},
		{name: "restart refuses the installer", run: restart, stack: ProductStackInstaller,
			wantErr: "operation restart not applicable for stack installer"},
	} {
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_UpdateStacks_UpdatesOnlyTheOutdatedStacks(t *testing.T) {
	for _, tc := range []logicCase{
		{name: "an outdated pentagi is pulled then brought up", stack: ProductStackPentagi,
			configure: logicOutdated(ProductStackPentagi),
			compose:   []string{"downloadStack pentagi", "updateStack pentagi"}},
		{name: "a current pentagi is left alone", stack: ProductStackPentagi},
		{name: "an outdated langfuse is pulled then brought up", stack: ProductStackLangfuse,
			configure: logicOutdated(ProductStackLangfuse),
			compose:   []string{"downloadStack langfuse", "updateStack langfuse"}},
		{name: "a current observability is left alone", stack: ProductStackObservability},
		{name: "the worker pulls its image", stack: ProductStackWorker,
			docker: []string{"pullWorkerImage"}},
		{name: "an installer update failure is returned", stack: ProductStackInstaller,
			configure: logicOutdated(ProductStackInstaller),
			fail:      map[string]error{"updateInstaller": errors.New("download refused")},
			wantErr:   "download refused", update: []string{"updateInstaller"}},
		{name: "compose skips the current stacks", stack: ProductStackCompose,
			configure: logicOutdated(ProductStackLangfuse, ProductStackObservability),
			compose: []string{
				"downloadStack observability", "updateStack observability",
				"downloadStack langfuse", "updateStack langfuse",
			},
			update: []string{"updateJaegerPlugin"}},
		{name: "all adds the worker and skips a current stack", stack: ProductStackAll,
			configure: logicOutdated(ProductStackPentagi, ProductStackObservability),
			compose: []string{
				"downloadStack observability", "updateStack observability",
				"downloadStack pentagi", "updateStack pentagi",
			},
			docker: []string{"pullWorkerImage"}, update: []string{"updateJaegerPlugin"}},
	} {
		tc.run = (*processor).update
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_UpdateStacks_RestartsJaegerOnlyWhenThePluginWasReplaced(t *testing.T) {
	for _, tc := range []logicCase{
		{name: "a replaced plugin restarts jaeger", replaced: true, compose: []string{
			"downloadStack observability", "updateStack observability",
			"performStackCommand observability restart jaeger",
		}},
		{name: "an unchanged plugin costs no downtime", compose: []string{
			"downloadStack observability", "updateStack observability",
		}},
	} {
		tc.run, tc.stack = (*processor).update, ProductStackObservability
		tc.configure, tc.update = logicOutdated(ProductStackObservability), []string{"updateJaegerPlugin"}
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_Update_AsksTheCloudOnceWhateverItFansOutTo(t *testing.T) {
	everything := logicOutdated(ProductStackPentagi, ProductStackLangfuse, ProductStackGraphiti, ProductStackObservability)
	for _, tc := range []struct {
		name      string
		stack     ProductStack
		configure func(*mockCheckConfig)
	}{
		{"an update of every compose stack", ProductStackCompose, everything},
		{"an update of a single stack", ProductStackPentagi, everything},
		{"an update of everything", ProductStackAll, everything},
		{"an update of compose where three stacks have nothing to do", ProductStackCompose, logicOutdated(ProductStackPentagi)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := processorHarnessWith(t, tc.configure)
			before := countCalls(h.checks.getCalls(), "GatherUpdatesInfo")
			require.NoError(t, h.p.update(t.Context(), tc.stack, testOperationState(t)))
			asked := countCalls(h.checks.getCalls(), "GatherUpdatesInfo") - before
			assert.Equal(t, 1, asked, "one press must cost one check out of a daily budget")
		})
	}
}

func TestLogic_StacksTouchedByUpdate_ExpandsComposeAndAll(t *testing.T) {
	assert.Equal(t, []ProductStack{
		ProductStackObservability, ProductStackLangfuse, ProductStackGraphiti, ProductStackPentagi,
	}, stacksTouchedByUpdate(ProductStackCompose))
	assert.Equal(t, []ProductStack{ProductStackPentagi}, stacksTouchedByUpdate(ProductStackPentagi))
	assert.Equal(t, []ProductStack{
		ProductStackObservability, ProductStackLangfuse, ProductStackGraphiti, ProductStackPentagi,
		ProductStackWorker, ProductStackInstaller,
	}, stacksTouchedByUpdate(ProductStackAll))
}

func TestLogic_ApplyChanges_BringsEachStackToItsConfiguredMode(t *testing.T) {
	networks := []string{"ensureMainDockerNetworks"}
	for _, tc := range []logicCase{
		{name: "a clean state changes nothing"},
		{name: "a network failure stops before any stack",
			configure: logicConnected, vars: logicEmbedded,
			fail:    map[string]error{"ensureMainDockerNetworks": errors.New("network error")},
			wantErr: "failed to ensure docker networks: network error", docker: networks},
		{name: "every embedded stack not yet extracted is extracted and brought up",
			configure: logicConnected, vars: logicEmbedded,
			fs: []string{
				"ensureStackIntegrity observability", "ensureStackIntegrity langfuse",
				"ensureStackIntegrity graphiti", "ensureStackIntegrity pentagi",
			},
			compose: []string{
				"updateStack observability", "updateStack langfuse", "updateStack graphiti", "updateStack pentagi",
			},
			docker: networks},
		{name: "an external observability that is installed is taken down",
			configure: func(c *mockCheckConfig) {
				c.ObservabilityExternal, c.ObservabilityInstalled = true, true
				c.LangfuseConnected, c.LangfuseExtracted = true, true
			},
			vars: map[string]string{
				"OTEL_HOST":         "http://external:4318",
				"LANGFUSE_BASE_URL": checker.DefaultLangfuseEndpoint,
			},
			fs:      []string{"verifyStackIntegrity langfuse", "ensureStackIntegrity pentagi"},
			compose: []string{"removeStack observability", "updateStack langfuse", "updateStack pentagi"},
			docker:  networks},
		{name: "an external graphiti that is installed is taken down",
			configure: func(c *mockCheckConfig) {
				c.GraphitiConnected, c.GraphitiExternal, c.GraphitiInstalled = true, true, true
			},
			vars:    map[string]string{"GRAPHITI_URL": "http://external:8000"},
			fs:      []string{"ensureStackIntegrity pentagi"},
			compose: []string{"removeStack graphiti", "updateStack pentagi"},
			docker:  networks},
		{name: "an embedded graphiti already extracted is verified instead",
			configure: func(c *mockCheckConfig) { c.GraphitiConnected, c.GraphitiExtracted = true, true },
			vars:      map[string]string{"GRAPHITI_URL": checker.DefaultGraphitiEndpoint},
			fs:        []string{"verifyStackIntegrity graphiti", "ensureStackIntegrity pentagi"},
			compose:   []string{"updateStack graphiti", "updateStack pentagi"},
			docker:    networks},
		{name: "a pentagi already extracted is verified instead",
			configure: func(c *mockCheckConfig) { c.PentagiExtracted = true },
			vars:      map[string]string{"PENTAGI_VERSION": "1.0.1"},
			fs:        []string{"verifyStackIntegrity pentagi"},
			compose:   []string{"updateStack pentagi"},
			docker:    networks},
		{name: "an observability failure names its phase",
			vars:    map[string]string{"OTEL_HOST": checker.DefaultObservabilityEndpoint},
			fail:    map[string]error{"ensureStackIntegrity": errors.New("fs error")},
			wantErr: "failed to apply observability changes: failed to ensure observability integrity: fs error",
			fs:      []string{"ensureStackIntegrity observability"}, docker: networks},
		{name: "a langfuse failure names its phase after observability went through",
			configure: logicConnected,
			vars: map[string]string{
				"OTEL_HOST":         checker.DefaultObservabilityEndpoint,
				"LANGFUSE_BASE_URL": checker.DefaultLangfuseEndpoint,
			},
			fail:    map[string]error{"ensureStackIntegrity_langfuse": errors.New("langfuse error")},
			wantErr: "failed to apply langfuse changes: failed to ensure langfuse integrity: langfuse error",
			fs:      []string{"ensureStackIntegrity observability", "ensureStackIntegrity langfuse"},
			compose: []string{"updateStack observability"}, docker: networks},
		{name: "a graphiti failure names its phase",
			configure: logicConnected,
			vars:      map[string]string{"GRAPHITI_URL": checker.DefaultGraphitiEndpoint},
			fail:      map[string]error{"ensureStackIntegrity": errors.New("graphiti error")},
			wantErr:   "failed to apply graphiti changes: failed to ensure graphiti integrity: graphiti error",
			fs:        []string{"ensureStackIntegrity graphiti"}, docker: networks},
		{name: "a pentagi failure names its phase",
			vars:    map[string]string{"PENTAGI_VERSION": "1.0.1"},
			fail:    map[string]error{"ensureStackIntegrity_pentagi": errors.New("pentagi error")},
			wantErr: "failed to apply pentagi changes: failed to ensure pentagi integrity: pentagi error",
			fs:      []string{"ensureStackIntegrity pentagi"}, docker: networks},
	} {
		tc.run = logicApplyChanges
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_Install_InstallsOnlyWhatIsNotInstalled(t *testing.T) {
	networks := []string{"ensureMainDockerNetworks"}
	for _, tc := range []logicCase{
		{name: "a fresh installation installs every embedded stack",
			configure: logicConnected, vars: logicEmbedded,
			fs: []string{
				"ensureStackIntegrity observability", "ensureStackIntegrity langfuse",
				"ensureStackIntegrity graphiti", "ensureStackIntegrity pentagi",
			},
			compose: []string{
				"updateStack observability", "updateStack langfuse", "updateStack graphiti", "updateStack pentagi",
			},
			docker: networks},
		{name: "an installed stack is skipped while the others are installed",
			configure: func(c *mockCheckConfig) { c.PentagiInstalled = true },
			vars:      map[string]string{"OTEL_HOST": checker.DefaultObservabilityEndpoint},
			fs:        []string{"ensureStackIntegrity observability"},
			compose:   []string{"updateStack observability"},
			docker:    networks},
	} {
		tc.run = logicInstall
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_FactoryReset_PurgesEverythingAndRestoresTheEmbeddedFiles(t *testing.T) {
	var copiesAtReset []int
	h := logicRun(t, logicCase{
		run:  logicFactoryReset,
		vars: map[string]string{"PENTAGI_VERSION": "1.0.1"},
		prepare: func(h *processorHarness) {
			h.p.state.(*mockState).onReset = func() {
				copiesAtReset = append(copiesAtReset, len(h.p.files.(*mockFiles).copies))
			}
		},
		compose: []string{"purgeStack all"},
		docker: []string{
			"removeWorkerContainers", "removeWorkerVolumes",
			"removeMainDockerNetwork pentagi-network",
			"removeMainDockerNetwork observability-network",
			"removeMainDockerNetwork langfuse-network",
		},
		fs: []string{"ensureStackIntegrity all force"},
	})
	assert.Equal(t, []struct {
		Src, Dst string
		Rewrite  bool
	}{{".env", filepath.Dir(h.p.state.GetEnvPath()), true}}, h.p.files.(*mockFiles).copies,
		"the embedded .env must overwrite the user's")
	assert.Equal(t, []int{1}, copiesAtReset, "the state must be reloaded once, from the restored .env")
	assert.False(t, h.p.state.IsDirty(), "a staged edit would be written back over the restored .env")
}

func TestLogic_Download_FetchesEveryStackItCovers(t *testing.T) {
	everyComposeStack := []string{
		"downloadStack observability", "downloadStack langfuse", "downloadStack graphiti", "downloadStack pentagi",
	}
	for _, tc := range []logicCase{
		{name: "compose fetches every compose stack and not the worker", stack: ProductStackCompose,
			compose: everyComposeStack},
		{name: "all fetches the worker too", stack: ProductStackAll,
			compose: everyComposeStack, docker: []string{"pullWorkerImage"}},
		{name: "the worker pulls its image", stack: ProductStackWorker, docker: []string{"pullWorkerImage"}},
		{name: "a current installer is not fetched", stack: ProductStackInstaller},
		{name: "an unreachable update server is reported", stack: ProductStackInstaller,
			configure: func(c *mockCheckConfig) { c.InstallerIsUpToDate, c.UpdateServerAccessible = false, false },
			wantErr:   "update server is not accessible"},
		{name: "an unknown stack is refused", stack: ProductStack("invalid"),
			wantErr: "operation download not applicable for stack invalid"},
	} {
		tc.run = (*processor).download
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_Remove_KeepsDataAndAttemptsEveryStack(t *testing.T) {
	for _, tc := range []logicCase{
		{name: "pentagi is taken down and keeps its volumes", stack: ProductStackPentagi,
			compose: []string{"removeStack pentagi"}},
		{name: "langfuse is taken down and keeps its volumes", stack: ProductStackLangfuse,
			compose: []string{"removeStack langfuse"}},
		{name: "observability is taken down and keeps its volumes", stack: ProductStackObservability,
			compose: []string{"removeStack observability"}},
		{name: "the worker loses its images", stack: ProductStackWorker,
			docker: []string{"removeWorkerImages"}},
		{name: "the installer is left in place without failing", stack: ProductStackInstaller,
			update: []string{"removeInstaller"}},
		{name: "every stack is attempted when the first one fails", stack: ProductStackAll,
			fail:    map[string]error{"removeStack": errors.New("compose exploded")},
			wantErr: "failed to remove stack: compose exploded",
			compose: []string{
				"removeStack observability", "removeStack langfuse", "removeStack graphiti", "removeStack pentagi",
			},
			docker: []string{"removeWorkerImages"}, update: []string{"removeInstaller"}},
	} {
		tc.run = (*processor).remove
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_Purge_RemovesImagesAndNetworksEvenWhenAStackFails(t *testing.T) {
	everyComposeStack := []string{
		"purgeImagesStack observability", "purgeImagesStack langfuse",
		"purgeImagesStack graphiti", "purgeImagesStack pentagi",
	}
	workerAndNetworks := []string{
		"purgeWorkerImages",
		"removeMainDockerNetwork pentagi-network",
		"removeMainDockerNetwork observability-network",
		"removeMainDockerNetwork langfuse-network",
	}
	for _, tc := range []logicCase{
		{name: "pentagi is purged with its images and nothing else", stack: ProductStackPentagi,
			compose: []string{"purgeImagesStack pentagi"}},
		{name: "the worker is purged", stack: ProductStackWorker, docker: []string{"purgeWorkerImages"}},
		{name: "everything is purged and the networks removed", stack: ProductStackAll,
			compose: everyComposeStack, docker: workerAndNetworks, update: []string{"removeInstaller"}},
		{name: "a failing stack still leaves no network behind", stack: ProductStackAll,
			fail:    map[string]error{"purgeImagesStack": errors.New("compose exploded")},
			wantErr: "failed to purge with images stack: compose exploded",
			compose: everyComposeStack, docker: workerAndNetworks, update: []string{"removeInstaller"}},
	} {
		tc.run = (*processor).purge
		t.Run(tc.name, func(t *testing.T) { logicRun(t, tc) })
	}
}

func TestLogic_IsEmbeddedDeployment_FollowsTheConfiguredEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name              string
		stack             ProductStack
		envVar, envValue  string
		langfuseConnected bool
		graphitiConnected bool
		want              bool
	}{
		{"observability at its embedded endpoint", ProductStackObservability, "OTEL_HOST", checker.DefaultObservabilityEndpoint, false, false, true},
		{"observability at an external endpoint", ProductStackObservability, "OTEL_HOST", "http://external:4318", false, false, false},
		{"langfuse at its embedded endpoint", ProductStackLangfuse, "LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint, true, false, true},
		{"langfuse at an external endpoint", ProductStackLangfuse, "LANGFUSE_BASE_URL", "http://external:3000", true, false, false},
		{"langfuse at its embedded endpoint but not connected", ProductStackLangfuse, "LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint, false, false, false},
		{"graphiti at its embedded endpoint", ProductStackGraphiti, "GRAPHITI_URL", checker.DefaultGraphitiEndpoint, false, true, true},
		{"graphiti at an external endpoint", ProductStackGraphiti, "GRAPHITI_URL", "http://external:8000", false, true, false},
		{"graphiti at its embedded endpoint but not connected", ProductStackGraphiti, "GRAPHITI_URL", checker.DefaultGraphitiEndpoint, false, false, false},
		{"pentagi is always embedded", ProductStackPentagi, "", "", false, false, true},
		{"the worker is always embedded", ProductStackWorker, "", "", false, false, true},
		{"the installer is always embedded", ProductStackInstaller, "", "", false, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := processorHarnessWith(t, func(c *mockCheckConfig) {
				c.LangfuseConnected, c.GraphitiConnected = tc.langfuseConnected, tc.graphitiConnected
			})
			if tc.envVar != "" {
				require.NoError(t, h.p.state.SetVar(tc.envVar, tc.envValue))
			}
			assert.Equal(t, tc.want, h.p.isEmbeddedDeployment(tc.stack))
		})
	}
}

func TestLogic_CheckFiles_ReportsEmbeddedStacksOnlyAndHidesEditedExcludedFiles(t *testing.T) {
	h := processorHarnessWith(t, func(c *mockCheckConfig) { c.LangfuseConnected, c.LangfuseExternal = true, true })
	h.p.fsOps = newFileSystemOperations(h.p)
	require.NoError(t, h.p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint))

	regular := "observability/subdir/config.yml"
	excludedMissing, excludedOK := "observability/otel/config.yml", "observability/grafana/config/grafana.ini"
	embedded := h.p.files.(*mockFiles)
	embedded.statuses = map[string]files.FileStatus{
		"docker-compose.yml":               files.FileStatusModified,
		"docker-compose-langfuse.yml":      files.FileStatusOK,
		"docker-compose-observability.yml": files.FileStatusMissing,
		regular:                            files.FileStatusModified,
		excludedMissing:                    files.FileStatusMissing,
		excludedOK:                         files.FileStatusOK,
	}
	for _, excluded := range filesToExcludeFromVerification {
		if _, set := embedded.statuses[excluded]; !set {
			embedded.statuses[excluded] = files.FileStatusModified
		}
	}
	embedded.lists[observabilityDirectory] = append([]string{regular}, filesToExcludeFromVerification...)

	result, err := h.p.checkFiles(t.Context(), ProductStackAll, testOperationState(t))
	require.NoError(t, err)
	assert.Equal(t, map[string]files.FileStatus{
		"docker-compose.yml":               files.FileStatusModified,
		"docker-compose-observability.yml": files.FileStatusMissing,
		regular:                            files.FileStatusModified,
		excludedMissing:                    files.FileStatusMissing,
		excludedOK:                         files.FileStatusOK,
	}, result, "langfuse is external, graphiti disabled, and an edited excluded file is nobody's business")
}
