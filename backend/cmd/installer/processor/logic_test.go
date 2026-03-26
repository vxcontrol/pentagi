package processor

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/pkg/version"
)

// newProcessorForLogicTests creates a processor with recording mocks and mock checker
func newProcessorForLogicTests(t *testing.T) (*processor, *baseMockComposeOperations, *baseMockFileSystemOperations, *baseMockDockerOperations) {
	t.Helper()

	// build fresh processor with mock checker result
	mockState := testState(t)
	checkResult := defaultCheckResult()

	p := createProcessorWithState(mockState, checkResult)

	// return typed mock operations for call verification
	composeOps := p.composeOps.(*baseMockComposeOperations)
	fsOps := p.fsOps.(*baseMockFileSystemOperations)
	dockerOps := p.dockerOps.(*baseMockDockerOperations)

	return p, composeOps, fsOps, dockerOps
}

// newProcessorForLogicTestsWithConfig creates a processor with custom checker configuration
func newProcessorForLogicTestsWithConfig(t *testing.T, configFunc func(*mockCheckConfig)) (*processor, *baseMockComposeOperations, *baseMockFileSystemOperations, *baseMockDockerOperations) {
	t.Helper()

	// build fresh processor with custom checker configuration
	mockState := testState(t)

	handler := newMockCheckHandler()
	if configFunc != nil {
		configFunc(&handler.config)
	}

	checkResult := createCheckResultWithHandler(handler)
	p := createProcessorWithState(mockState, checkResult)

	// return typed mock operations for call verification
	composeOps := p.composeOps.(*baseMockComposeOperations)
	fsOps := p.fsOps.(*baseMockFileSystemOperations)
	dockerOps := p.dockerOps.(*baseMockDockerOperations)

	return p, composeOps, fsOps, dockerOps
}

// injectComposeError injects error into compose operations for testing
func injectComposeError(p *processor, errorMethods map[string]error) {
	baseMock := p.composeOps.(*baseMockComposeOperations)
	for method, err := range errorMethods {
		baseMock.setError(method, err)
	}
}

// injectDockerError injects error into docker operations for testing
func injectDockerError(p *processor, errorMethods map[string]error) {
	baseMock := p.dockerOps.(*baseMockDockerOperations)
	for method, err := range errorMethods {
		baseMock.setError(method, err)
	}
}

// injectFSError injects error into filesystem operations for testing
func injectFSError(p *processor, errorMethods map[string]error) {
	baseMock := p.fsOps.(*baseMockFileSystemOperations)
	for method, err := range errorMethods {
		baseMock.setError(method, err)
	}
}

// testStackOperation is a helper that tests stack operations with standard patterns
func testStackOperation(t *testing.T,
	operation func(*processor, context.Context, ProductStack, *operationState) error,
	expectedMethod string, processorOp ProcessorOperation,
) {
	t.Helper()

	// test successful delegation
	t.Run("delegates_to_compose", func(t *testing.T) {
		p, composeOps, _, _ := newProcessorForLogicTests(t)

		err := operation(p, t.Context(), ProductStackPentagi, testOperationState(t))
		assertNoError(t, err)

		calls := composeOps.getCalls()
		if len(calls) != 1 {
			t.Fatalf("expected 1 compose call, got %d", len(calls))
		}
		if calls[0].Method != expectedMethod || calls[0].Stack != ProductStackPentagi {
			t.Fatalf("unexpected call: %+v", calls[0])
		}
	})

	// test validation errors
	t.Run("validation_errors", func(t *testing.T) {
		p, _, _, _ := newProcessorForLogicTests(t)

		testCases := generateStackTestCases(processorOp)
		for _, tc := range testCases {
			if tc.expectErr {
				t.Run(tc.name, func(t *testing.T) {
					err := operation(p, t.Context(), tc.stack, testOperationState(t))
					assertError(t, err, true, tc.errorMsg)
				})
			}
		}
	})
}

func TestStart(t *testing.T) {
	testStackOperation(t, (*processor).start, "startStack", ProcessorOperationStart)
}

func TestStop(t *testing.T) {
	testStackOperation(t, (*processor).stop, "stopStack", ProcessorOperationStop)
}

func TestRestart(t *testing.T) {
	testStackOperation(t, (*processor).restart, "restartStack", ProcessorOperationRestart)
}

func TestUpdate(t *testing.T) {
	t.Run("compose_stacks", func(t *testing.T) {
		// Test that update respects IsUpToDate flags
		testCases := []struct {
			name         string
			stack        ProductStack
			isUpToDate   bool
			expectUpdate bool
			configSetup  func(*mockCheckConfig)
		}{
			{
				name:         "pentagi_needs_update",
				stack:        ProductStackPentagi,
				isUpToDate:   false,
				expectUpdate: true,
				configSetup: func(config *mockCheckConfig) {
					config.PentagiIsUpToDate = false
				},
			},
			{
				name:         "pentagi_already_updated",
				stack:        ProductStackPentagi,
				isUpToDate:   true,
				expectUpdate: false,
				configSetup: func(config *mockCheckConfig) {
					config.PentagiIsUpToDate = true
				},
			},
			{
				name:         "langfuse_needs_update",
				stack:        ProductStackLangfuse,
				isUpToDate:   false,
				expectUpdate: true,
				configSetup: func(config *mockCheckConfig) {
					config.LangfuseIsUpToDate = false
				},
			},
			{
				name:         "observability_already_updated",
				stack:        ProductStackObservability,
				isUpToDate:   true,
				expectUpdate: false,
				configSetup: func(config *mockCheckConfig) {
					config.ObservabilityIsUpToDate = true
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				p, composeOps, _, _ := newProcessorForLogicTestsWithConfig(t, tc.configSetup)

				err := p.update(t.Context(), tc.stack, testOperationState(t))
				assertNoError(t, err)

				calls := composeOps.getCalls()
				if tc.expectUpdate {
					if len(calls) != 2 || calls[0].Method != "downloadStack" || calls[1].Method != "updateStack" {
						t.Errorf("expected downloadStack and updateStack calls, got: %+v", calls)
					}
				} else {
					if len(calls) != 0 {
						t.Errorf("expected no calls for up-to-date stack, got: %+v", calls)
					}
				}
			})
		}
	})

	t.Run("worker_stack", func(t *testing.T) {
		p, _, _, dockerOps := newProcessorForLogicTests(t)

		err := p.update(t.Context(), ProductStackWorker, testOperationState(t))
		assertNoError(t, err)

		calls := dockerOps.getCalls()
		if len(calls) != 1 || calls[0].Method != "pullWorkerImage" {
			t.Errorf("expected pullWorkerImage call, got: %+v", calls)
		}
	})

	t.Run("installer_stack", func(t *testing.T) {
		p, _, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.InstallerIsUpToDate = false
			config.UpdateServerAccessible = true
		})

		// Mock updateOps to avoid "not implemented" error
		updateOps := p.updateOps.(*baseMockUpdateOperations)
		updateOps.setError("updateInstaller", fmt.Errorf("not implemented"))

		err := p.update(t.Context(), ProductStackInstaller, testOperationState(t))
		assertError(t, err, true, "not implemented")

		calls := updateOps.getCalls()
		if len(calls) != 1 || calls[0].Method != "updateInstaller" {
			t.Errorf("expected updateInstaller call, got: %+v", calls)
		}
	})

	t.Run("compose_stacks", func(t *testing.T) {
		p, composeOps, _, dockerOps := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.PentagiIsUpToDate = true // should skip
			config.LangfuseIsUpToDate = false
			config.ObservabilityIsUpToDate = false
		})

		err := p.update(t.Context(), ProductStackCompose, testOperationState(t))
		assertNoError(t, err)

		// Check compose calls - should update langfuse and observability, skip pentagi
		composeCalls := composeOps.getCalls()
		updateCount := 0
		for _, call := range composeCalls {
			if call.Method == "updateStack" {
				updateCount++
				// Verify we don't update pentagi
				if call.Stack == ProductStackPentagi {
					t.Error("should not update pentagi when it's up to date")
				}
			}
		}
		if updateCount != 2 {
			t.Errorf("expected 2 updateStack calls, got %d", updateCount)
		}

		// Check docker calls for worker
		dockerCalls := dockerOps.getCalls()
		workerPulled := false
		for _, call := range dockerCalls {
			if call.Method == "pullWorkerImage" {
				workerPulled = true
			}
		}
		if workerPulled {
			t.Error("expected no worker image to be pulled")
		}
	})

	t.Run("all_stacks", func(t *testing.T) {
		p, composeOps, _, dockerOps := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.PentagiIsUpToDate = false
			config.LangfuseIsUpToDate = true // should skip
			config.ObservabilityIsUpToDate = false
		})

		err := p.update(t.Context(), ProductStackAll, testOperationState(t))
		assertNoError(t, err)

		// Check compose calls - should update pentagi and observability, skip langfuse
		composeCalls := composeOps.getCalls()
		updateCount := 0
		for _, call := range composeCalls {
			if call.Method == "updateStack" {
				updateCount++
				// Verify we don't update Langfuse
				if call.Stack == ProductStackLangfuse {
					t.Error("should not update Langfuse when it's up to date")
				}
			}
		}
		if updateCount != 2 {
			t.Errorf("expected 2 updateStack calls, got %d", updateCount)
		}

		// Check docker calls for worker
		dockerCalls := dockerOps.getCalls()
		workerPulled := false
		for _, call := range dockerCalls {
			if call.Method == "pullWorkerImage" {
				workerPulled = true
			}
		}
		if !workerPulled {
			t.Error("expected worker image to be pulled")
		}
	})

	// Test validation errors
	t.Run("validation_errors", func(t *testing.T) {
		p, _, _, _ := newProcessorForLogicTests(t)

		testCases := generateStackTestCases(ProcessorOperationUpdate)
		for _, tc := range testCases {
			if tc.expectErr {
				t.Run(tc.name, func(t *testing.T) {
					err := p.update(t.Context(), tc.stack, testOperationState(t))
					assertError(t, err, true, tc.errorMsg)
				})
			}
		}
	})
}

func TestRemove(t *testing.T) {
	testStackOperation(t, (*processor).remove, "removeStack", ProcessorOperationRemove)
}

func TestApplyChanges_ErrorPropagation_FromEnsureNetworks(t *testing.T) {
	p, _, _, _ := newProcessorForLogicTests(t)

	// inject error into ensureNetworks
	injectDockerError(p, map[string]error{
		"ensureMainDockerNetworks": fmt.Errorf("network error"),
	})

	_ = p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint)
	_ = p.state.SetVar("LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint)

	// not extracted forces ensure
	p.checker.ObservabilityExtracted = false
	p.checker.LangfuseExtracted = false
	p.checker.PentagiExtracted = false

	err := p.applyChanges(t.Context(), testOperationState(t))
	assertError(t, err, true, "failed to ensure docker networks: network error")
}

func TestPurge_StrictAndDockerCleanup(t *testing.T) {
	t.Run("compose_stack_purge", func(t *testing.T) {
		p, composeOps, _, dockerOps := newProcessorForLogicTests(t)

		err := p.purge(t.Context(), ProductStackPentagi, testOperationState(t))
		assertNoError(t, err)

		// first call must be strict purge (images)
		composeCalls := composeOps.getCalls()
		if len(composeCalls) == 0 || composeCalls[0].Method != "purgeImagesStack" || composeCalls[0].Stack != ProductStackPentagi {
			t.Fatalf("expected purgeImagesStack call for pentagi, got: %+v", composeCalls)
		}

		// docker cleanup operations should NOT be called for individual compose stack
		dockerCalls := dockerOps.getCalls()
		if len(dockerCalls) > 0 {
			t.Errorf("expected no docker calls for individual compose stack purge, got: %+v", dockerCalls)
		}
	})

	t.Run("worker_stack_purge", func(t *testing.T) {
		p, _, _, dockerOps := newProcessorForLogicTests(t)

		err := p.purge(t.Context(), ProductStackWorker, testOperationState(t))
		assertNoError(t, err)

		// worker purge should call purgeWorkerImages (which internally calls removeWorkerContainers)
		dockerCalls := dockerOps.getCalls()
		if len(dockerCalls) == 0 || dockerCalls[0].Method != "purgeWorkerImages" {
			t.Errorf("expected purgeWorkerImages call for worker, got: %+v", dockerCalls)
		}
	})
}

func TestApplyChanges_Embedded_AllStacksUpdated(t *testing.T) {
	// use custom config to set specific states
	p, composeOps, fsOps, dockerOps := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
		// mark as not extracted to force ensure
		config.ObservabilityExtracted = false
		config.LangfuseExtracted = false
		config.GraphitiExtracted = false
		config.PentagiExtracted = false
		// ensure embedded mode conditions
		config.ObservabilityConnected = true
		config.ObservabilityExternal = false
		config.LangfuseConnected = true
		config.LangfuseExternal = false
		config.GraphitiConnected = true
		config.GraphitiExternal = false
	})

	// mark state dirty and set embedded modes
	_ = p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint)
	_ = p.state.SetVar("LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint)
	_ = p.state.SetVar("GRAPHITI_URL", checker.DefaultGraphitiEndpoint)

	err := p.applyChanges(t.Context(), testOperationState(t))
	if err != nil {
		t.Fatalf("applyChanges returned error: %v", err)
	}

	dockerCalls := dockerOps.getCalls()
	if len(dockerCalls) == 0 || dockerCalls[0].Method != "ensureMainDockerNetworks" {
		t.Fatalf("expected ensureMainDockerNetworks first, got: %+v", dockerCalls)
	}

	// ensure/verify for four stacks and update four stacks
	// since all not extracted -> ensure called for obs, langfuse, graphiti, pentagi
	fsCalls := fsOps.getCalls()
	ensureCount := 0
	for _, c := range fsCalls {
		if c.Method == "ensureStackIntegrity" {
			ensureCount++
		}
	}

	composeCalls := composeOps.getCalls()
	updateCount := 0
	for _, c := range composeCalls {
		if c.Method == "updateStack" {
			updateCount++
		}
	}
	if ensureCount != 4 || updateCount != 4 {
		t.Fatalf("expected ensure=4 and update=4, got ensure=%d update=%d", ensureCount, updateCount)
	}
}

func TestApplyChanges_Disabled_RemovesInstalled(t *testing.T) {
	// use custom config to simulate installed observability that should be removed
	p, composeOps, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
		// simulate installed observability
		config.ObservabilityInstalled = true
		config.ObservabilityConnected = true
		config.ObservabilityExternal = true // external means it should be removed if installed
	})

	// mark state dirty and set external for observability
	_ = p.state.SetVar("OTEL_HOST", "http://external-otel:4318")

	err := p.applyChanges(t.Context(), testOperationState(t))
	if err != nil {
		t.Fatalf("applyChanges returned error: %v", err)
	}

	// should include remove for observability
	calls := composeOps.getCalls()
	found := false
	for _, c := range calls {
		if c.Method == "removeStack" && c.Stack == ProductStackObservability {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected removeStack(observability) call not found; calls: %+v", calls)
	}
}

// Additional comprehensive tests for business logic coverage

func TestDownload_ComposeStacks(t *testing.T) {
	p, composeOps, _, dockerOps := newProcessorForLogicTests(t)

	err := p.download(t.Context(), ProductStackCompose, testOperationState(t))
	if err != nil {
		t.Fatalf("download returned error: %v", err)
	}

	// should download all individual stacks
	composeCalls := composeOps.getCalls()
	expectedComposeStacks := []ProductStack{
		ProductStackPentagi, ProductStackGraphiti, ProductStackLangfuse, ProductStackObservability,
	}
	composeCallCount := 0
	for _, call := range composeCalls {
		if call.Method == "downloadStack" {
			composeCallCount++
		}
	}
	if composeCallCount != len(expectedComposeStacks) {
		t.Errorf("expected %d compose download calls, got %d", len(expectedComposeStacks), composeCallCount)
	}

	// should also download worker
	dockerCalls := dockerOps.getCalls()
	workerDownloaded := false
	for _, call := range dockerCalls {
		if call.Method == "pullWorkerImage" {
			workerDownloaded = true
			break
		}
	}
	if workerDownloaded {
		t.Error("expected no worker image download, but found")
	}
}

func TestDownload_AllStacks(t *testing.T) {
	p, composeOps, _, dockerOps := newProcessorForLogicTests(t)

	err := p.download(t.Context(), ProductStackAll, testOperationState(t))
	if err != nil {
		t.Fatalf("download returned error: %v", err)
	}

	// should download all individual stacks
	composeCalls := composeOps.getCalls()
	expectedComposeStacks := []ProductStack{
		ProductStackPentagi, ProductStackGraphiti, ProductStackLangfuse, ProductStackObservability,
	}
	composeCallCount := 0
	for _, call := range composeCalls {
		if call.Method == "downloadStack" {
			composeCallCount++
		}
	}
	if composeCallCount != len(expectedComposeStacks) {
		t.Errorf("expected %d compose download calls, got %d", len(expectedComposeStacks), composeCallCount)
	}

	// should also download worker
	dockerCalls := dockerOps.getCalls()
	workerDownloaded := false
	for _, call := range dockerCalls {
		if call.Method == "pullWorkerImage" {
			workerDownloaded = true
			break
		}
	}
	if !workerDownloaded {
		t.Error("expected worker image download, but not found")
	}
}

func TestDownload_WorkerStack(t *testing.T) {
	p, _, _, dockerOps := newProcessorForLogicTests(t)

	err := p.download(t.Context(), ProductStackWorker, testOperationState(t))
	if err != nil {
		t.Fatalf("download returned error: %v", err)
	}

	calls := dockerOps.getCalls()
	if len(calls) != 1 || calls[0].Method != "pullWorkerImage" {
		t.Fatalf("expected pullWorkerImage call, got: %+v", calls)
	}
}

func TestDownload_InvalidStack(t *testing.T) {
	p, _, _, _ := newProcessorForLogicTests(t)

	err := p.download(t.Context(), ProductStack("invalid"), testOperationState(t))
	if err == nil {
		t.Error("expected error for invalid stack, got nil")
	}
}

func TestValidateOperation_ErrorCases(t *testing.T) {
	p, _, _, _ := newProcessorForLogicTests(t)

	tests := []struct {
		name      string
		stack     ProductStack
		operation ProcessorOperation
		expectErr bool
		errMsg    string
	}{
		{"start worker", ProductStackWorker, ProcessorOperationStart, true, "operation start not applicable for stack worker"},
		{"stop worker", ProductStackWorker, ProcessorOperationStop, true, "operation stop not applicable for stack worker"},
		{"restart installer", ProductStackInstaller, ProcessorOperationRestart, true, "operation restart not applicable for stack installer"},
		{"remove installer", ProductStackInstaller, ProcessorOperationRemove, false, ""}, // remove is allowed for installer
		{"valid start pentagi", ProductStackPentagi, ProcessorOperationStart, false, ""},
		{"valid remove worker", ProductStackWorker, ProcessorOperationRemove, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateOperation(tt.stack, tt.operation)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIsEmbeddedDeployment(t *testing.T) {
	tests := []struct {
		name              string
		stack             ProductStack
		envVar            string
		envValue          string
		langfuseConnected bool
		graphitiConnected bool
		expected          bool
	}{
		{"observability embedded", ProductStackObservability, "OTEL_HOST", checker.DefaultObservabilityEndpoint, false, false, true},
		{"observability external", ProductStackObservability, "OTEL_HOST", "http://external:4318", false, false, false},
		{"langfuse embedded", ProductStackLangfuse, "LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint, true, false, true},
		{"langfuse external", ProductStackLangfuse, "LANGFUSE_BASE_URL", "http://external:3000", true, false, false},
		{"langfuse disabled", ProductStackLangfuse, "", "", false, false, false},
		{"graphiti embedded", ProductStackGraphiti, "GRAPHITI_URL", checker.DefaultGraphitiEndpoint, false, true, true},
		{"graphiti external", ProductStackGraphiti, "GRAPHITI_URL", "http://external:8000", false, true, false},
		{"graphiti disabled", ProductStackGraphiti, "", "", false, false, false},
		{"pentagi always embedded", ProductStackPentagi, "", "", false, false, true},     // pentagi is always embedded
		{"worker always embedded", ProductStackWorker, "", "", false, false, true},       // worker is always embedded
		{"installer always embedded", ProductStackInstaller, "", "", false, false, true}, // installer is always embedded
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
				config.LangfuseConnected = tt.langfuseConnected
				config.GraphitiConnected = tt.graphitiConnected
			})

			if tt.envVar != "" {
				_ = p.state.SetVar(tt.envVar, tt.envValue)
			}

			result := p.isEmbeddedDeployment(tt.stack)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFactoryReset_FullSequence(t *testing.T) {
	p, composeOps, _, dockerOps := newProcessorForLogicTests(t)

	err := p.factoryReset(t.Context(), testOperationState(t))
	if err != nil {
		t.Fatalf("factoryReset returned error: %v", err)
	}

	// verify sequence: purge stacks, remove worker containers/volumes, remove networks
	composeCalls := composeOps.getCalls()
	if len(composeCalls) == 0 || composeCalls[0].Method != "purgeStack" || composeCalls[0].Stack != ProductStackAll {
		t.Errorf("expected purgeStack(all) as first call, got: %+v", composeCalls)
	}

	dockerCalls := dockerOps.getCalls()
	expectedMethods := []string{"removeWorkerContainers", "removeWorkerVolumes", "removeMainDockerNetwork", "removeMainDockerNetwork", "removeMainDockerNetwork"}
	if len(dockerCalls) < len(expectedMethods) {
		t.Errorf("expected at least %d docker calls, got %d", len(expectedMethods), len(dockerCalls))
	}

	for i, expectedMethod := range expectedMethods {
		if i < len(dockerCalls) && dockerCalls[i].Method != expectedMethod {
			t.Errorf("docker call %d: expected %s, got %s", i, expectedMethod, dockerCalls[i].Method)
		}
	}
}

func TestApplyChanges_StateMachine_PhaseErrors(t *testing.T) {
	tests := []struct {
		name          string
		configSetup   func(*mockCheckConfig)
		setupError    func(*processor)
		expectedError string
	}{
		{
			name: "observability phase error",
			configSetup: func(config *mockCheckConfig) {
				config.ObservabilityExtracted = false
				config.ObservabilityConnected = true
				config.ObservabilityExternal = false
			},
			setupError: func(p *processor) {
				_ = p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint)
				_ = p.state.SetVar("PENTAGI_VERSION", version.GetBinaryVersion()) // make state dirty
				injectFSError(p, map[string]error{
					"ensureStackIntegrity": fmt.Errorf("fs error"),
				})
			},
			expectedError: "failed to apply observability changes: failed to ensure observability integrity: fs error",
		},
		{
			name: "langfuse phase error",
			configSetup: func(config *mockCheckConfig) {
				config.LangfuseExtracted = false
				config.LangfuseConnected = true
				config.LangfuseExternal = false
			},
			setupError: func(p *processor) {
				_ = p.state.SetVar("LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint)
				_ = p.state.SetVar("PENTAGI_VERSION", version.GetBinaryVersion()) // make state dirty
				injectFSError(p, map[string]error{
					"ensureStackIntegrity": fmt.Errorf("langfuse error"),
				})
			},
			expectedError: "failed to apply langfuse changes: failed to ensure langfuse integrity: langfuse error",
		},
		{
			name: "graphiti phase error",
			configSetup: func(config *mockCheckConfig) {
				config.GraphitiExtracted = false
				config.GraphitiConnected = true
				config.GraphitiExternal = false
			},
			setupError: func(p *processor) {
				_ = p.state.SetVar("GRAPHITI_URL", checker.DefaultGraphitiEndpoint)
				_ = p.state.SetVar("PENTAGI_VERSION", version.GetBinaryVersion()) // make state dirty
				injectFSError(p, map[string]error{
					"ensureStackIntegrity": fmt.Errorf("graphiti error"),
				})
			},
			expectedError: "failed to apply graphiti changes: failed to ensure graphiti integrity: graphiti error",
		},
		{
			name: "pentagi phase error",
			configSetup: func(config *mockCheckConfig) {
				config.PentagiExtracted = false
			},
			setupError: func(p *processor) {
				// make state dirty so applyChanges proceeds
				_ = p.state.SetVar("PENTAGI_VERSION", version.GetBinaryVersion())
				injectFSError(p, map[string]error{
					"ensureStackIntegrity_pentagi": fmt.Errorf("pentagi error"),
					"ensureStackIntegrity":         fmt.Errorf("general error"), // fallback to catch any call
				})
			},
			expectedError: "failed to apply pentagi changes: failed to ensure pentagi integrity: pentagi error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _, _, _ := newProcessorForLogicTestsWithConfig(t, tt.configSetup)
			tt.setupError(p)

			err := p.applyChanges(t.Context(), testOperationState(t))
			assertError(t, err, true, tt.expectedError)
		})
	}
}

func TestApplyChanges_CleanState_NoOp(t *testing.T) {
	p, composeOps, fsOps, dockerOps := newProcessorForLogicTests(t)

	// clean state should result in no operations
	p.state.Reset() // ensure state is not dirty

	err := p.applyChanges(t.Context(), testOperationState(t))
	if err != nil {
		t.Fatalf("applyChanges returned error: %v", err)
	}

	// no operations should be called
	if len(composeOps.getCalls()) > 0 {
		t.Errorf("expected no compose calls, got: %+v", composeOps.getCalls())
	}
	if len(fsOps.getCalls()) > 0 {
		t.Errorf("expected no fs calls, got: %+v", fsOps.getCalls())
	}
	if len(dockerOps.getCalls()) > 0 {
		t.Errorf("expected no docker calls, got: %+v", dockerOps.getCalls())
	}
}

func TestInstall_FullScenario(t *testing.T) {
	t.Run("all_stacks_fresh_install", func(t *testing.T) {
		p, composeOps, fsOps, dockerOps := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			// simulate fresh install - nothing installed
			config.ObservabilityInstalled = false
			config.LangfuseInstalled = false
			config.GraphitiInstalled = false
			config.PentagiInstalled = false
			config.ObservabilityExtracted = false
			config.LangfuseExtracted = false
			config.GraphitiExtracted = false
			config.PentagiExtracted = false
			// mark as embedded
			config.ObservabilityConnected = true
			config.ObservabilityExternal = false
			config.LangfuseConnected = true
			config.LangfuseExternal = false
			config.GraphitiConnected = true
			config.GraphitiExternal = false
		})

		// set embedded mode for all
		_ = p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint)
		_ = p.state.SetVar("LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint)
		_ = p.state.SetVar("GRAPHITI_URL", checker.DefaultGraphitiEndpoint)

		err := p.install(t.Context(), testOperationState(t))
		assertNoError(t, err)

		// verify docker networks created first
		dockerCalls := dockerOps.getCalls()
		if len(dockerCalls) == 0 || dockerCalls[0].Method != "ensureMainDockerNetworks" {
			t.Errorf("expected ensureMainDockerNetworks as first docker call, got: %+v", dockerCalls)
		}

		// verify file system operations
		fsCalls := fsOps.getCalls()
		ensureCount := 0
		for _, call := range fsCalls {
			if call.Method == "ensureStackIntegrity" {
				ensureCount++
			}
		}
		// should be 3 (observability, langfuse, graphiti) since pentagi might be handled differently
		if ensureCount < 3 {
			t.Errorf("expected at least 3 ensureStackIntegrity calls, got %d", ensureCount)
		}

		// verify compose update operations
		composeCalls := composeOps.getCalls()
		updateCount := 0
		for _, call := range composeCalls {
			if call.Method == "updateStack" {
				updateCount++
			}
		}
		// all 4 stacks should be updated (observability, langfuse, graphiti, pentagi)
		if updateCount != 4 {
			t.Errorf("expected 4 updateStack calls, got %d", updateCount)
		}
	})

	t.Run("partial_install_skip_installed", func(t *testing.T) {
		p, composeOps, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			// pentagi already installed
			config.PentagiInstalled = true
			config.LangfuseInstalled = false
			config.ObservabilityInstalled = false
		})

		err := p.install(t.Context(), testOperationState(t))
		assertNoError(t, err)

		// should not update pentagi since it's already installed
		composeCalls := composeOps.getCalls()
		for _, call := range composeCalls {
			if call.Method == "updateStack" && call.Stack == ProductStackPentagi {
				t.Error("should not update pentagi when already installed")
			}
		}
	})
}

func TestPreviewFilesStatus_Behavior(t *testing.T) {
	if len(filesToExcludeFromVerification) < 3 {
		t.Skip("not enough excluded files configured; skipping excluded files tests")
	}

	p, _, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
		config.ObservabilityConnected = true
		config.LangfuseConnected = true
		config.LangfuseExternal = true // external langfuse should not be present
	})
	// use real fs implementation for preview to exercise real logic
	p.fsOps = newFileSystemOperations(p)

	// prepare files in mock
	tmpDir := t.TempDir()
	mockState := p.state.(*mockState)
	mockState.envPath = filepath.Join(tmpDir, ".env")
	mockFiles := p.files.(*mockFiles)
	mockFiles.statuses[composeFilePentagi] = files.FileStatusModified
	mockFiles.statuses[composeFileLangfuse] = files.FileStatusOK
	mockFiles.statuses[composeFileObservability] = files.FileStatusMissing
	mockFiles.statuses["observability/subdir/config.yml"] = files.FileStatusModified
	mockFiles.statuses[filesToExcludeFromVerification[0]] = files.FileStatusMissing
	mockFiles.statuses[filesToExcludeFromVerification[1]] = files.FileStatusOK
	for i := 2; i < len(filesToExcludeFromVerification); i++ {
		mockFiles.statuses[filesToExcludeFromVerification[i]] = files.FileStatusModified
	}
	mockFiles.lists[observabilityDirectory] = append([]string{
		"observability/subdir/config.yml", // normal
	}, filesToExcludeFromVerification...)

	// ensure embedded mode via state env for observability
	_ = p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint)

	statuses, err := p.checkFiles(t.Context(), ProductStackAll, testOperationState(t))
	assertNoError(t, err)

	// pentagi, observability compose must be present and reflect modified
	for _, k := range []string{composeFilePentagi, composeFileObservability} {
		if statuses[k] != mockFiles.statuses[k] {
			t.Errorf("expected %s to be %s, got %s", k, mockFiles.statuses[k], statuses[k])
		}
	}

	// langfuse compose should not be present because it's not embedded
	if _, ok := statuses[composeFileLangfuse]; ok {
		t.Errorf("expected langfuse compose to be missing, got %s", statuses[composeFileLangfuse])
	}

	// all non-modified excluded files must be present and reflect modified
	for i := range 2 {
		if k := filesToExcludeFromVerification[i]; statuses[k] != mockFiles.statuses[k] {
			t.Errorf("expected %s to be %s, got %s", k, mockFiles.statuses[k], statuses[k])
		}
	}

	// all non-excluded modified files should not be present
	for i := 2; i < len(filesToExcludeFromVerification); i++ {
		k, empty := filesToExcludeFromVerification[i], files.FileStatus("")
		if status, ok := statuses[k]; ok || status != empty {
			t.Errorf("expected %s to be missing, got %s", k, status)
		}
	}
}

func TestDownload_EdgeCases(t *testing.T) {
	t.Run("installer_up_to_date", func(t *testing.T) {
		p, _, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.InstallerIsUpToDate = true
		})

		updateOps := p.updateOps.(*baseMockUpdateOperations)

		err := p.download(t.Context(), ProductStackInstaller, testOperationState(t))
		assertNoError(t, err)

		// should not call downloadInstaller when up to date
		calls := updateOps.getCalls()
		if len(calls) > 0 {
			t.Errorf("expected no update calls when installer is up to date, got: %+v", calls)
		}
	})

	t.Run("update_server_inaccessible", func(t *testing.T) {
		p, _, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.InstallerIsUpToDate = false
			config.UpdateServerAccessible = false
		})

		err := p.download(t.Context(), ProductStackInstaller, testOperationState(t))
		assertError(t, err, true, "update server is not accessible")
	})
}

func TestPurge_AllStacks_Detailed(t *testing.T) {
	p, composeOps, _, dockerOps := newProcessorForLogicTests(t)

	err := p.purge(t.Context(), ProductStackAll, testOperationState(t))
	assertNoError(t, err)

	// verify compose operations for all stacks
	composeCalls := composeOps.getCalls()

	// should have purgeImagesStack for all four compose stacks in order
	expectedOrder := []ProductStack{
		ProductStackObservability, ProductStackLangfuse, ProductStackGraphiti, ProductStackPentagi,
	}
	purgeImagesCalls := 0
	for _, call := range composeCalls {
		if call.Method == "purgeImagesStack" {
			if purgeImagesCalls < len(expectedOrder) && call.Stack != expectedOrder[purgeImagesCalls] {
				t.Errorf("expected purgeImagesStack call %d for %s, got %s",
					purgeImagesCalls, expectedOrder[purgeImagesCalls], call.Stack)
			}
			purgeImagesCalls++
		}
	}
	if purgeImagesCalls != 4 {
		t.Errorf("expected 4 purgeImagesStack calls, got %d", purgeImagesCalls)
	}

	// verify docker cleanup operations
	dockerCalls := dockerOps.getCalls()

	// For ProductStackAll, we expect:
	// 1. purgeWorkerImages (from purge worker)
	// 2-4. removeMainDockerNetwork x3 (cleanup networks)
	expectedDockerMethods := []string{
		"purgeWorkerImages",
		"removeMainDockerNetwork",
		"removeMainDockerNetwork",
		"removeMainDockerNetwork",
	}

	if len(dockerCalls) < len(expectedDockerMethods) {
		t.Fatalf("expected at least %d docker calls, got %d", len(expectedDockerMethods), len(dockerCalls))
	}

	for i, expected := range expectedDockerMethods {
		if i < len(dockerCalls) && dockerCalls[i].Method != expected {
			t.Errorf("docker call %d: expected %s, got %s", i, expected, dockerCalls[i].Method)
		}
	}
}

func TestRemove_PreservesData(t *testing.T) {
	t.Run("compose_stacks_preserve_volumes", func(t *testing.T) {
		p, composeOps, _, _ := newProcessorForLogicTests(t)

		stacks := []ProductStack{ProductStackPentagi, ProductStackLangfuse, ProductStackObservability}

		for _, stack := range stacks {
			err := p.remove(t.Context(), stack, testOperationState(t))
			assertNoError(t, err)
		}

		// verify removeStack (not purgeStack) was called
		calls := composeOps.getCalls()
		for _, call := range calls {
			if call.Method != "removeStack" {
				t.Errorf("expected removeStack, got %s", call.Method)
			}
		}
	})

	t.Run("worker_removes_images_and_containers", func(t *testing.T) {
		p, _, _, dockerOps := newProcessorForLogicTests(t)

		err := p.remove(t.Context(), ProductStackWorker, testOperationState(t))
		assertNoError(t, err)

		calls := dockerOps.getCalls()
		// remove for worker calls removeWorkerImages (which internally removes containers first)
		hasRemoveImages := false
		for _, call := range calls {
			if call.Method == "removeWorkerImages" {
				hasRemoveImages = true
			}
		}

		if !hasRemoveImages {
			t.Error("expected removeWorkerImages to be called")
		}
	})
}

func TestApplyChanges_ComplexScenarios(t *testing.T) {
	t.Run("mixed_deployment_modes", func(t *testing.T) {
		p, composeOps, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			// observability external, langfuse embedded, graphiti disabled, pentagi always embedded
			config.ObservabilityExternal = true
			config.ObservabilityInstalled = true // should be removed
			config.LangfuseExternal = false
			config.LangfuseExtracted = true // mark as extracted so it goes to update path
			config.LangfuseConnected = true // required for isEmbeddedDeployment to return true
			config.GraphitiConnected = false
			config.PentagiExtracted = false
		})

		_ = p.state.SetVar("OTEL_HOST", "http://external:4318")
		_ = p.state.SetVar("LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint)

		err := p.applyChanges(t.Context(), testOperationState(t))
		assertNoError(t, err)

		// verify observability removed
		composeCalls := composeOps.getCalls()
		obsRemoved := false
		for _, call := range composeCalls {
			if call.Method == "removeStack" && call.Stack == ProductStackObservability {
				obsRemoved = true
			}
		}
		if !obsRemoved {
			t.Error("expected observability to be removed when external")
		}

		// verify langfuse installed - check for update operation
		langfuseUpdated := false
		for _, call := range composeCalls {
			if call.Method == "updateStack" && call.Stack == ProductStackLangfuse {
				langfuseUpdated = true
			}
		}
		if !langfuseUpdated {
			t.Error("expected langfuse to be updated")
		}
	})

	t.Run("graphiti_external_removes_installed", func(t *testing.T) {
		p, composeOps, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			// graphiti external but installed locally - should be removed
			config.GraphitiConnected = true
			config.GraphitiExternal = true
			config.GraphitiInstalled = true
		})

		_ = p.state.SetVar("GRAPHITI_URL", "http://external:8000")

		err := p.applyChanges(t.Context(), testOperationState(t))
		assertNoError(t, err)

		// verify graphiti removed
		composeCalls := composeOps.getCalls()
		graphitiRemoved := false
		for _, call := range composeCalls {
			if call.Method == "removeStack" && call.Stack == ProductStackGraphiti {
				graphitiRemoved = true
			}
		}
		if !graphitiRemoved {
			t.Error("expected graphiti to be removed when external")
		}
	})

	t.Run("graphiti_embedded_installs", func(t *testing.T) {
		p, composeOps, fsOps, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			// graphiti embedded but not installed yet
			config.GraphitiConnected = true
			config.GraphitiExternal = false
			config.GraphitiExtracted = false
			config.GraphitiInstalled = false
		})

		_ = p.state.SetVar("GRAPHITI_URL", checker.DefaultGraphitiEndpoint)

		err := p.applyChanges(t.Context(), testOperationState(t))
		assertNoError(t, err)

		// verify graphiti files ensured
		fsCalls := fsOps.getCalls()
		graphitiEnsured := false
		for _, call := range fsCalls {
			if call.Method == "ensureStackIntegrity" && call.Stack == ProductStackGraphiti {
				graphitiEnsured = true
			}
		}
		if !graphitiEnsured {
			t.Error("expected graphiti files to be ensured")
		}

		// verify graphiti updated
		composeCalls := composeOps.getCalls()
		graphitiUpdated := false
		for _, call := range composeCalls {
			if call.Method == "updateStack" && call.Stack == ProductStackGraphiti {
				graphitiUpdated = true
			}
		}
		if !graphitiUpdated {
			t.Error("expected graphiti to be updated")
		}
	})

	t.Run("graphiti_embedded_already_extracted", func(t *testing.T) {
		p, composeOps, fsOps, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			// graphiti embedded and already extracted - should verify integrity
			config.GraphitiConnected = true
			config.GraphitiExternal = false
			config.GraphitiExtracted = true
			config.GraphitiInstalled = false
		})

		_ = p.state.SetVar("GRAPHITI_URL", checker.DefaultGraphitiEndpoint)

		err := p.applyChanges(t.Context(), testOperationState(t))
		assertNoError(t, err)

		// verify graphiti files verified (not ensured)
		fsCalls := fsOps.getCalls()
		graphitiVerified := false
		for _, call := range fsCalls {
			if call.Method == "verifyStackIntegrity" && call.Stack == ProductStackGraphiti {
				graphitiVerified = true
			}
		}
		if !graphitiVerified {
			t.Error("expected graphiti files to be verified")
		}

		// verify graphiti updated
		composeCalls := composeOps.getCalls()
		graphitiUpdated := false
		for _, call := range composeCalls {
			if call.Method == "updateStack" && call.Stack == ProductStackGraphiti {
				graphitiUpdated = true
			}
		}
		if !graphitiUpdated {
			t.Error("expected graphiti to be updated")
		}
	})

	t.Run("error_recovery_partial_state", func(t *testing.T) {
		p, _, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.ObservabilityExtracted = false
			config.LangfuseExtracted = false
			config.LangfuseConnected = true // required for isEmbeddedDeployment to return true
			config.PentagiExtracted = false
		})

		_ = p.state.SetVar("OTEL_HOST", checker.DefaultObservabilityEndpoint)
		_ = p.state.SetVar("LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint)
		_ = p.state.SetVar("DIRTY_FLAG", "true") // ensure state is dirty

		// inject error in langfuse phase
		injectFSError(p, map[string]error{
			"ensureStackIntegrity_langfuse": fmt.Errorf("langfuse error"),
		})

		err := p.applyChanges(t.Context(), testOperationState(t))
		assertError(t, err, true, "failed to apply langfuse changes: failed to ensure langfuse integrity: langfuse error")

		// verify observability was processed before langfuse error
		fsCalls := p.fsOps.(*baseMockFileSystemOperations).getCalls()
		obsProcessed := false
		langfuseAttempted := false
		for _, call := range fsCalls {
			if call.Method == "ensureStackIntegrity" {
				if call.Stack == ProductStackObservability && call.Error == nil {
					obsProcessed = true
				}
				if call.Stack == ProductStackLangfuse && call.Error != nil {
					langfuseAttempted = true
				}
			}
		}

		if !obsProcessed {
			t.Error("expected observability to be processed before error")
		}
		if !langfuseAttempted {
			t.Error("expected langfuse processing to be attempted")
		}
	})
}
