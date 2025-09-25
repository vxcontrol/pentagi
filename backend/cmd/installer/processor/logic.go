package processor

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/wizard/logger"
)

const (
	minTerminalWidthForCompose = 56
)

// helper to validate stack applicability for operation
func (p *processor) validateOperation(stack ProductStack, operation ProcessorOperation) error {
	switch operation {
	case ProcessorOperationApplyChanges, ProcessorOperationInstall, ProcessorOperationFactoryReset:
		if stack != ProductStackAll {
			return fmt.Errorf("operation %s not applicable for stack %s", operation, stack)
		}

	case ProcessorOperationUpdate, ProcessorOperationDownload, ProcessorOperationRemove, ProcessorOperationPurge:
		break // can be applied to any stack

	case ProcessorOperationStart, ProcessorOperationStop, ProcessorOperationRestart, ProcessorOperationCheckFiles:
		if stack == ProductStackWorker || stack == ProductStackInstaller {
			return fmt.Errorf("operation %s not applicable for stack %s", operation, stack)
		}

	case ProcessorOperationResetPassword:
		if stack != ProductStackPentagi {
			return fmt.Errorf("operation %s only applicable for PentAGI stack", operation)
		}
	}

	return nil
}

// isEmbeddedDeployment checks if the deployment mode is embedded based on environment variable
func (p *processor) isEmbeddedDeployment(stack ProductStack) bool {
	switch stack {
	case ProductStackObservability:
		envVar, envVarValueEmbedded := "OTEL_HOST", checker.DefaultObservabilityEndpoint
		if envVar, exists := p.state.GetVar(envVar); exists && envVar.Value == envVarValueEmbedded {
			return true
		}

		return false

	case ProductStackLangfuse:
		if !p.checker.LangfuseConnected {
			return false
		}

		envVar, envVarValueEmbedded := "LANGFUSE_BASE_URL", checker.DefaultLangfuseEndpoint
		if envVar, exists := p.state.GetVar(envVar); exists && envVar.Value == envVarValueEmbedded {
			return true
		}

		return false

	case ProductStackPentagi, ProductStackWorker, ProductStackInstaller:
		return true

	default:
		return false
	}
}

func (p *processor) runCommand(cmd *exec.Cmd, stack ProductStack, state *operationState) error {
	if state.terminal != nil {
		// patch env vars for docker compose and small size screen
		if width, _ := state.terminal.GetSize(); width < minTerminalWidthForCompose {
			cmd.Env = append(cmd.Env, "COMPOSE_ANSI=never")
		}

		if err := state.terminal.Execute(cmd); err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		logger.Log("waiting for command: %s", cmd.String())
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("failed to wait for command: %w", err)
		}

		logger.Log("waiting for terminal to finish: %s", cmd.String())
		state.terminal.Wait()
		logger.Log("terminal finished: %s", cmd.String())

		state.sendOutput(state.terminal.View(), false, stack)
	} else {
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to run command: %w\n%s", err, string(output))
		} else {
			state.sendOutput(string(output), true, stack)
		}
	}

	return nil
}

func (p *processor) appendLog(msg string, stack ProductStack, state *operationState) {
	if state.terminal != nil {
		state.terminal.Append(msg)
		state.sendOutput(state.terminal.View(), false, stack)
	} else {
		state.sendOutput(msg, true, stack)
	}
}

func (p *processor) isFileExists(path string) error {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", path)
	} else if err != nil {
		return fmt.Errorf("file %s: %w", path, err)
	} else if info.IsDir() {
		return fmt.Errorf("file %s is a directory", path)
	}

	return nil
}

func (p *processor) applyChanges(ctx context.Context, state *operationState) (err error) {
	stack := ProductStackAll
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()
	defer func() {
		if err != nil {
			return
		}
		// refresh state for updates
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			err = fmt.Errorf("failed to gather info after update: %w", err)
		}
	}()

	if err := p.validateOperation(stack, ProcessorOperationApplyChanges); err != nil {
		return err
	}

	if !p.state.IsDirty() {
		return nil
	}

	if err = p.state.Commit(); err != nil {
		return fmt.Errorf("failed to commit state: %w", err)
	}

	// refresh checker state after commit to use updated .env values
	if err := p.checker.GatherAllInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather updated info: %w", err)
	}

	// ensure required docker networks exist before bringing up stacks
	if err := p.dockerOps.ensureMainDockerNetworks(ctx, state); err != nil {
		return fmt.Errorf("failed to ensure docker networks: %w", err)
	}

	// phase 1: Observability Stack Management
	if err := p.applyObservabilityChanges(ctx, state); err != nil {
		return fmt.Errorf("failed to apply observability changes: %w", err)
	}

	// phase 2: Langfuse Stack Management
	if err := p.applyLangfuseChanges(ctx, state); err != nil {
		return fmt.Errorf("failed to apply langfuse changes: %w", err)
	}

	// phase 3: PentAGI Stack Management (always embedded, always required)
	if err := p.applyPentagiChanges(ctx, state); err != nil {
		return fmt.Errorf("failed to apply pentagi changes: %w", err)
	}

	return nil
}

func (p *processor) applyObservabilityChanges(ctx context.Context, state *operationState) error {
	if p.isEmbeddedDeployment(ProductStackObservability) {
		// user wants embedded observability
		if !p.checker.ObservabilityExtracted {
			// fresh installation - extract all files
			if err := p.fsOps.ensureStackIntegrity(ctx, ProductStackObservability, state); err != nil {
				return fmt.Errorf("failed to ensure observability integrity: %w", err)
			}
		} else {
			// files exist - verify integrity, update if force=true
			if err := p.fsOps.verifyStackIntegrity(ctx, ProductStackObservability, state); err != nil {
				return fmt.Errorf("failed to verify observability integrity: %w", err)
			}
		}

		// update/start containers
		if err := p.composeOps.updateStack(ctx, ProductStackObservability, state); err != nil {
			return fmt.Errorf("failed to update observability stack: %w", err)
		}
	} else {
		// user wants external/disabled observability
		if p.checker.ObservabilityInstalled {
			// remove containers but keep files (user might re-enable)
			if err := p.composeOps.removeStack(ctx, ProductStackObservability, state); err != nil {
				return fmt.Errorf("failed to remove observability stack: %w", err)
			}
		}
	}

	// refresh state to verify operation success
	if err := p.checker.GatherObservabilityInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather observability info: %w", err)
	}

	return nil
}

func (p *processor) applyLangfuseChanges(ctx context.Context, state *operationState) error {
	if p.isEmbeddedDeployment(ProductStackLangfuse) {
		// user wants embedded langfuse
		if !p.checker.LangfuseExtracted {
			// fresh installation - extract compose file
			if err := p.fsOps.ensureStackIntegrity(ctx, ProductStackLangfuse, state); err != nil {
				return fmt.Errorf("failed to ensure langfuse integrity: %w", err)
			}
		} else {
			// file exists - verify integrity, update if force=true
			if err := p.fsOps.verifyStackIntegrity(ctx, ProductStackLangfuse, state); err != nil {
				return fmt.Errorf("failed to verify langfuse integrity: %w", err)
			}
		}

		// update/start containers
		if err := p.composeOps.updateStack(ctx, ProductStackLangfuse, state); err != nil {
			return fmt.Errorf("failed to update langfuse stack: %w", err)
		}
	} else {
		// user wants external/disabled langfuse
		if p.checker.LangfuseInstalled {
			// remove containers but keep files (user might re-enable)
			if err := p.composeOps.removeStack(ctx, ProductStackLangfuse, state); err != nil {
				return fmt.Errorf("failed to remove langfuse stack: %w", err)
			}
		}
	}

	// refresh state to verify operation success
	if err := p.checker.GatherLangfuseInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather langfuse info: %w", err)
	}

	return nil
}

func (p *processor) applyPentagiChanges(ctx context.Context, state *operationState) error {
	// PentAGI is always embedded, always required
	if !p.checker.PentagiExtracted {
		// fresh installation - extract compose file
		if err := p.fsOps.ensureStackIntegrity(ctx, ProductStackPentagi, state); err != nil {
			return fmt.Errorf("failed to ensure pentagi integrity: %w", err)
		}
	} else {
		// file exists - verify integrity, update if force=true
		if err := p.fsOps.verifyStackIntegrity(ctx, ProductStackPentagi, state); err != nil {
			return fmt.Errorf("failed to verify pentagi integrity: %w", err)
		}
	}

	// update/start containers
	if err := p.composeOps.updateStack(ctx, ProductStackPentagi, state); err != nil {
		return fmt.Errorf("failed to update pentagi stack: %w", err)
	}

	// refresh state to verify operation success
	if err := p.checker.GatherPentagiInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather pentagi info: %w", err)
	}

	return nil
}

// checkFiles computes file statuses for a given stack, honoring the same
// rules as verifyStackIntegrity: active stacks only and excluded files policy.
// It serves as a dry-run for file operations without performing any writes.
func (p *processor) checkFiles(
	ctx context.Context, stack ProductStack, state *operationState,
) (result map[string]files.FileStatus, err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()
	defer func() { state.sendFilesCheck(stack, result, err) }()

	result = make(map[string]files.FileStatus)
	allStacks := []ProductStack{ProductStackPentagi, ProductStackLangfuse, ProductStackObservability}

	if err := p.validateOperation(stack, ProcessorOperationCheckFiles); err != nil {
		return nil, err
	}

	switch stack {
	case ProductStackPentagi, ProductStackLangfuse, ProductStackObservability:
		if !p.isEmbeddedDeployment(stack) {
			return map[string]files.FileStatus{}, nil
		}

		result, err = p.fsOps.checkStackIntegrity(ctx, stack)
		if err != nil {
			return result, fmt.Errorf("failed to check files integrity: %w", err)
		}

	case ProductStackAll, ProductStackCompose:
		for _, s := range allStacks {
			if !p.isEmbeddedDeployment(s) {
				continue
			}

			if r, err := p.fsOps.checkStackIntegrity(ctx, s); err != nil {
				return result, fmt.Errorf("failed to check %s files integrity: %w", s, err)
			} else {
				maps.Copy(result, r)
			}
		}
	}

	return result, nil
}

func (p *processor) factoryReset(ctx context.Context, state *operationState) (err error) {
	stack := ProductStackAll
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	if err := p.validateOperation(stack, ProcessorOperationFactoryReset); err != nil {
		return err
	}

	// step 1: stop and remove stacks with volumes (down -v) with force semantics
	p.appendLog(MsgFactoryResetStarting, ProductStackInstaller, state)
	if err := p.composeOps.purgeStack(ctx, stack, state); err != nil {
		return fmt.Errorf("failed to purge stacks: %w", err)
	}

	// step 2: remove worker containers and volumes in worker env
	if err := p.dockerOps.removeWorkerContainers(ctx, state); err != nil {
		return fmt.Errorf("failed to remove worker containers: %w", err)
	}
	if err := p.dockerOps.removeWorkerVolumes(ctx, state); err != nil {
		return fmt.Errorf("failed to remove worker volumes: %w", err)
	}

	// step 3: remove main networks
	_ = p.dockerOps.removeMainDockerNetwork(ctx, state, string(ProductDockerNetworkPentagi))
	_ = p.dockerOps.removeMainDockerNetwork(ctx, state, string(ProductDockerNetworkObservability))
	_ = p.dockerOps.removeMainDockerNetwork(ctx, state, string(ProductDockerNetworkLangfuse))

	// step 4: restore .env from embedded and reload state
	p.appendLog(MsgRestoringDefaultEnv, ProductStackInstaller, state)
	envDir := filepath.Dir(p.state.GetEnvPath())
	if err := p.files.Copy(".env", envDir, true); err != nil {
		return fmt.Errorf("failed to restore default .env: %w", err)
	}
	p.appendLog(MsgDefaultEnvRestored, ProductStackInstaller, state)
	if err := p.state.Reset(); err != nil {
		return fmt.Errorf("failed to reset state: %w", err)
	}

	// step 5: restore all embedded files to defaults with overwrite
	// observability directory and compose files
	err = p.fsOps.ensureStackIntegrity(ctx, stack, &operationState{force: true, mx: &sync.Mutex{}, ctx: ctx})
	if err != nil {
		return fmt.Errorf("failed to restore embedded files: %w", err)
	}

	p.appendLog(MsgFactoryResetCompleted, ProductStackInstaller, state)

	// refresh checker to reflect clean baseline
	if err := p.checker.GatherAllInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather info after factory reset: %w", err)
	}

	return nil
}

func (p *processor) install(ctx context.Context, state *operationState) (err error) {
	stack := ProductStackAll
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()
	defer func() {
		if err != nil {
			return
		}
		// refresh state for updates
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			err = fmt.Errorf("failed to gather info after update: %w", err)
		}
	}()

	if err := p.validateOperation(stack, ProcessorOperationInstall); err != nil {
		return err
	}

	// ensure required docker networks exist before bringing up stacks
	if err := p.dockerOps.ensureMainDockerNetworks(ctx, state); err != nil {
		return fmt.Errorf("failed to ensure docker networks: %w", err)
	}

	// phase 1: Observability Stack Management
	if !p.checker.ObservabilityInstalled {
		if err := p.applyObservabilityChanges(ctx, state); err != nil {
			return fmt.Errorf("failed to apply observability changes: %w", err)
		}
	}

	// phase 2: Langfuse Stack Management
	if !p.checker.LangfuseInstalled {
		if err := p.applyLangfuseChanges(ctx, state); err != nil {
			return fmt.Errorf("failed to apply langfuse changes: %w", err)
		}
	}

	// phase 3: PentAGI Stack Management (always embedded, always required)
	if !p.checker.PentagiInstalled {
		if err := p.applyPentagiChanges(ctx, state); err != nil {
			return fmt.Errorf("failed to apply pentagi changes: %w", err)
		}
	}

	return nil
}

func (p *processor) update(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()
	defer func() {
		if err != nil {
			return
		}
		// refresh state for updates
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			err = fmt.Errorf("failed to gather info after update: %w", err)
		}
	}()

	allStacks := append(composeOperationAllStacksOrder[ProcessorOperationUpdate],
		ProductStackWorker,
		ProductStackInstaller,
	)

	composeStacksUpToDate := map[ProductStack]bool{
		ProductStackPentagi:       p.checker.PentagiIsUpToDate,
		ProductStackLangfuse:      p.checker.LangfuseIsUpToDate,
		ProductStackObservability: p.checker.ObservabilityIsUpToDate,
	}
	composeStacksGatherInfo := map[ProductStack]func(ctx context.Context) error{
		ProductStackPentagi:       p.checker.GatherPentagiInfo,
		ProductStackLangfuse:      p.checker.GatherLangfuseInfo,
		ProductStackObservability: p.checker.GatherObservabilityInfo,
	}

	if err := p.validateOperation(stack, ProcessorOperationUpdate); err != nil {
		return err
	}

	switch stack {
	case ProductStackPentagi, ProductStackLangfuse, ProductStackObservability:
		if composeStacksUpToDate[stack] {
			return nil
		}

		if err := p.composeOps.downloadStack(ctx, stack, state); err != nil {
			return fmt.Errorf("failed to download stack: %w", err)
		}

		// docker compose update equivalent for all images
		if err := p.composeOps.updateStack(ctx, stack, state); err != nil {
			return fmt.Errorf("failed to update stack: %w", err)
		}

		if err := composeStacksGatherInfo[stack](ctx); err != nil {
			return fmt.Errorf("failed to gather info after update: %w", err)
		}

	case ProductStackWorker:
		// pull worker images
		if err := p.dockerOps.pullWorkerImage(ctx, state); err != nil {
			return fmt.Errorf("failed to pull worker images: %w", err)
		}

		if err := p.checker.GatherWorkerInfo(ctx); err != nil {
			return fmt.Errorf("failed to gather worker info after download: %w", err)
		}

	case ProductStackInstaller:
		if p.checker.InstallerIsUpToDate {
			return nil
		}
		if !p.checker.UpdateServerAccessible {
			return fmt.Errorf("update server is not accessible")
		}

		// HTTP GET from update server
		return p.updateOps.updateInstaller(ctx, state)

	case ProductStackCompose:
		for _, s := range composeOperationAllStacksOrder[ProcessorOperationUpdate] {
			if err := p.update(ctx, s, state); err != nil {
				return err
			}
		}

	case ProductStackAll:
		// update all applicable stacks
		for _, s := range allStacks {
			if err := p.update(ctx, s, state); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("operation update not applicable for stack %s", stack)
	}

	return nil
}

func (p *processor) download(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	allStacks := append(composeOperationAllStacksOrder[ProcessorOperationDownload],
		ProductStackWorker,
		ProductStackInstaller,
	)

	if err := p.validateOperation(stack, ProcessorOperationDownload); err != nil {
		return err
	}

	switch stack {
	case ProductStackPentagi, ProductStackLangfuse, ProductStackObservability:
		// docker compose pull equivalent for all images
		if err := p.composeOps.downloadStack(ctx, stack, state); err != nil {
			return fmt.Errorf("failed to download stack: %w", err)
		}

	case ProductStackWorker:
		// pull worker images
		if err := p.dockerOps.pullWorkerImage(ctx, state); err != nil {
			return err
		}

		if err := p.checker.GatherWorkerInfo(ctx); err != nil {
			return fmt.Errorf("failed to gather worker info after download: %w", err)
		}
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			return fmt.Errorf("failed to gather worker info after download: %w", err)
		}

	case ProductStackInstaller:
		if p.checker.InstallerIsUpToDate {
			return nil
		}
		if !p.checker.UpdateServerAccessible {
			return fmt.Errorf("update server is not accessible")
		}

		// HTTP GET from update server
		return p.updateOps.downloadInstaller(ctx, state)

	case ProductStackCompose:
		for _, s := range composeOperationAllStacksOrder[ProcessorOperationDownload] {
			if err := p.download(ctx, s, state); err != nil {
				return err
			}
		}

	case ProductStackAll:
		// download all applicable stacks
		for _, s := range allStacks {
			if err := p.download(ctx, s, state); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("operation download not applicable for stack %s", stack)
	}

	return nil
}

func (p *processor) remove(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	allStacks := append(composeOperationAllStacksOrder[ProcessorOperationRemove],
		ProductStackWorker,
		ProductStackInstaller,
	)

	composeStacksGatherInfo := map[ProductStack]func(ctx context.Context) error{
		ProductStackPentagi:       p.checker.GatherPentagiInfo,
		ProductStackLangfuse:      p.checker.GatherLangfuseInfo,
		ProductStackObservability: p.checker.GatherObservabilityInfo,
	}

	if err := p.validateOperation(stack, ProcessorOperationRemove); err != nil {
		return err
	}

	switch stack {
	case ProductStackPentagi, ProductStackLangfuse, ProductStackObservability:
		if err := p.composeOps.removeStack(ctx, stack, state); err != nil {
			return fmt.Errorf("failed to remove stack: %w", err)
		}

		if err := composeStacksGatherInfo[stack](ctx); err != nil {
			return fmt.Errorf("failed to gather info after remove: %w", err)
		}

	case ProductStackWorker:
		// remove worker images and containers
		if err := p.dockerOps.removeWorkerImages(ctx, state); err != nil {
			return fmt.Errorf("failed to remove worker images: %w", err)
		}

		if err := p.checker.GatherWorkerInfo(ctx); err != nil {
			return fmt.Errorf("failed to gather worker info after remove: %w", err)
		}
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			err = fmt.Errorf("failed to gather info after update: %w", err)
		}

	case ProductStackInstaller:
		// remove installer binary
		if err := p.updateOps.removeInstaller(ctx, state); err != nil {
			return fmt.Errorf("failed to remove installer: %w", err)
		}

	case ProductStackCompose:
		for _, s := range composeOperationAllStacksOrder[ProcessorOperationRemove] {
			if err := p.remove(ctx, s, state); err != nil {
				return err
			}
		}

	case ProductStackAll:
		// remove all stacks
		for _, s := range allStacks {
			if err := p.remove(ctx, s, state); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("operation remove not applicable for stack %s", stack)
	}

	return nil
}

func (p *processor) purge(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()
	defer func() {
		if err != nil {
			return
		}
		// refresh state for updates
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			err = fmt.Errorf("failed to gather info after update: %w", err)
		}
	}()

	allStacks := append(composeOperationAllStacksOrder[ProcessorOperationPurge],
		ProductStackWorker,
		ProductStackInstaller,
	)

	composeStacksGatherInfo := map[ProductStack]func(ctx context.Context) error{
		ProductStackPentagi:       p.checker.GatherPentagiInfo,
		ProductStackLangfuse:      p.checker.GatherLangfuseInfo,
		ProductStackObservability: p.checker.GatherObservabilityInfo,
	}

	if err := p.validateOperation(stack, ProcessorOperationPurge); err != nil {
		return err
	}

	switch stack {
	case ProductStackPentagi, ProductStackLangfuse, ProductStackObservability:
		if err := p.composeOps.purgeImagesStack(ctx, stack, state); err != nil {
			return fmt.Errorf("failed to purge with images stack: %w", err)
		}

		if err := composeStacksGatherInfo[stack](ctx); err != nil {
			return fmt.Errorf("failed to gather info after purge: %w", err)
		}

	case ProductStackWorker:
		// purge worker images and containers and volumes
		if err := p.dockerOps.purgeWorkerImages(ctx, state); err != nil {
			return fmt.Errorf("failed to purge worker images: %w", err)
		}

		if err := p.checker.GatherWorkerInfo(ctx); err != nil {
			return fmt.Errorf("failed to gather worker info after purge: %w", err)
		}
		if err := p.checker.GatherUpdatesInfo(ctx); err != nil {
			err = fmt.Errorf("failed to gather info after update: %w", err)
		}

	case ProductStackInstaller:
		// remove installer binary
		if err := p.updateOps.removeInstaller(ctx, state); err != nil {
			return fmt.Errorf("failed to remove installer: %w", err)
		}

	case ProductStackCompose:
		for _, s := range composeOperationAllStacksOrder[ProcessorOperationPurge] {
			if err := p.purge(ctx, s, state); err != nil {
				return err
			}
		}

	case ProductStackAll:
		// purge all stacks
		for _, s := range allStacks {
			if err := p.purge(ctx, s, state); err != nil {
				return err
			}
		}

		// remove custom networks
		_ = p.dockerOps.removeMainDockerNetwork(ctx, state, string(ProductDockerNetworkPentagi))
		_ = p.dockerOps.removeMainDockerNetwork(ctx, state, string(ProductDockerNetworkObservability))
		_ = p.dockerOps.removeMainDockerNetwork(ctx, state, string(ProductDockerNetworkLangfuse))

	default:
		return fmt.Errorf("operation purge not applicable for stack %s", stack)
	}

	return nil
}

func (p *processor) start(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	if err := p.validateOperation(stack, ProcessorOperationStart); err != nil {
		return err
	}

	if err := p.composeOps.startStack(ctx, stack, state); err != nil {
		return fmt.Errorf("failed to start stack: %w", err)
	}

	if err := p.checker.GatherAllInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather info after start: %w", err)
	}

	return nil
}

func (p *processor) stop(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	if err := p.validateOperation(stack, ProcessorOperationStop); err != nil {
		return err
	}

	if err := p.composeOps.stopStack(ctx, stack, state); err != nil {
		return fmt.Errorf("failed to stop stack: %w", err)
	}

	if err := p.checker.GatherAllInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather info after stop: %w", err)
	}

	return nil
}

func (p *processor) restart(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	if err := p.validateOperation(stack, ProcessorOperationRestart); err != nil {
		return err
	}

	if err := p.composeOps.restartStack(ctx, stack, state); err != nil {
		return fmt.Errorf("failed to restart stack: %w", err)
	}

	if err := p.checker.GatherAllInfo(ctx); err != nil {
		return fmt.Errorf("failed to gather info after restart: %w", err)
	}

	return nil
}

func (p *processor) resetPassword(ctx context.Context, stack ProductStack, state *operationState) (err error) {
	state.sendStarted(stack)
	defer func() { state.sendCompletion(stack, err) }()

	if err := p.validateOperation(stack, ProcessorOperationResetPassword); err != nil {
		return err
	}

	if stack != ProductStackPentagi {
		return fmt.Errorf("reset password operation only supported for PentAGI stack")
	}

	if !p.checker.PentagiRunning {
		return fmt.Errorf("PentAGI must be running to reset password")
	}

	if state.passwordValue == "" {
		return fmt.Errorf("password value is required")
	}

	p.appendLog("Resetting admin password...", stack, state)

	// perform password reset using PostgreSQL operations
	if err := p.performPasswordReset(ctx, state.passwordValue, state); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	p.appendLog("Password reset completed successfully", stack, state)

	return nil
}
