package processor

import (
	"context"
	"sync"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/terminal"
)

type ProductStack string

const (
	ProductStackPentagi       ProductStack = "pentagi"
	ProductStackGraphiti      ProductStack = "graphiti"
	ProductStackLangfuse      ProductStack = "langfuse"
	ProductStackObservability ProductStack = "observability"
	ProductStackCompose       ProductStack = "compose"
	ProductStackWorker        ProductStack = "worker"
	ProductStackInstaller     ProductStack = "installer"
	ProductStackAll           ProductStack = "all"
)

type ProcessorOperation string

const (
	ProcessorOperationApplyChanges  ProcessorOperation = "apply_changes"
	ProcessorOperationCheckFiles    ProcessorOperation = "check_files"
	ProcessorOperationFactoryReset  ProcessorOperation = "factory_reset"
	ProcessorOperationInstall       ProcessorOperation = "install"
	ProcessorOperationUpdate        ProcessorOperation = "update"
	ProcessorOperationDownload      ProcessorOperation = "download"
	ProcessorOperationRemove        ProcessorOperation = "remove"
	ProcessorOperationPurge         ProcessorOperation = "purge"
	ProcessorOperationStart         ProcessorOperation = "start"
	ProcessorOperationStop          ProcessorOperation = "stop"
	ProcessorOperationRestart       ProcessorOperation = "restart"
	ProcessorOperationResetPassword ProcessorOperation = "reset_password"
)

type ProductDockerNetwork string

const (
	ProductDockerNetworkPentagi       ProductDockerNetwork = "pentagi-network"
	ProductDockerNetworkObservability ProductDockerNetwork = "observability-network"
	ProductDockerNetworkLangfuse      ProductDockerNetwork = "langfuse-network"
)

type FilesCheckResult map[string]files.FileStatus

type Processor interface {
	ApplyChanges(ctx context.Context, opts ...OperationOption) error
	CheckFiles(ctx context.Context, stack ProductStack, opts ...OperationOption) (FilesCheckResult, error)
	FactoryReset(ctx context.Context, opts ...OperationOption) error
	Install(ctx context.Context, opts ...OperationOption) error
	Update(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	Download(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	Remove(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	Purge(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	Start(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	Stop(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	Restart(ctx context.Context, stack ProductStack, opts ...OperationOption) error
	ResetPassword(ctx context.Context, stack ProductStack, opts ...OperationOption) error
}

// WithForce skips validation checks and attempts maximum operations
func WithForce() OperationOption {
	return func(c *operationState) { c.force = true }
}

// WithTerminalModel enables embedded terminal model integration
func WithTerminal(term terminal.Terminal) OperationOption {
	return func(c *operationState) {
		if term != nil {
			c.terminal = term
		}
	}
}

// WithPasswordValue sets password value for reset password operation
func WithPasswordValue(password string) OperationOption {
	return func(c *operationState) {
		c.passwordValue = password
	}
}

// internal interfaces for specialized operations
type fileSystemOperations interface {
	ensureStackIntegrity(ctx context.Context, stack ProductStack, state *operationState) error
	verifyStackIntegrity(ctx context.Context, stack ProductStack, state *operationState) error
	cleanupStackFiles(ctx context.Context, stack ProductStack, state *operationState) error
	checkStackIntegrity(ctx context.Context, stack ProductStack) (FilesCheckResult, error)
}

type dockerOperations interface {
	pullWorkerImage(ctx context.Context, state *operationState) error
	pullDefaultImage(ctx context.Context, state *operationState) error
	removeWorkerContainers(ctx context.Context, state *operationState) error
	removeWorkerImages(ctx context.Context, state *operationState) error
	purgeWorkerImages(ctx context.Context, state *operationState) error
	ensureMainDockerNetworks(ctx context.Context, state *operationState) error
	removeMainDockerNetwork(ctx context.Context, state *operationState, name string) error
	removeMainImages(ctx context.Context, state *operationState, images []string) error
	removeWorkerVolumes(ctx context.Context, state *operationState) error
}

type composeOperations interface {
	startStack(ctx context.Context, stack ProductStack, state *operationState) error
	stopStack(ctx context.Context, stack ProductStack, state *operationState) error
	restartStack(ctx context.Context, stack ProductStack, state *operationState) error
	updateStack(ctx context.Context, stack ProductStack, state *operationState) error
	downloadStack(ctx context.Context, stack ProductStack, state *operationState) error
	removeStack(ctx context.Context, stack ProductStack, state *operationState) error
	purgeStack(ctx context.Context, stack ProductStack, state *operationState) error
	purgeImagesStack(ctx context.Context, stack ProductStack, state *operationState) error
	performStackCommand(ctx context.Context, stack ProductStack, state *operationState, args ...string) error
	determineComposeFile(stack ProductStack) (string, error)
}

type updateOperations interface {
	checkUpdates(ctx context.Context, state *operationState) (*checker.CheckUpdatesResponse, error)
	downloadInstaller(ctx context.Context, state *operationState) error
	updateInstaller(ctx context.Context, state *operationState) error
	removeInstaller(ctx context.Context, state *operationState) error
}

type processor struct {
	mu      *sync.Mutex
	state   state.State
	checker *checker.CheckResult
	files   files.Files

	// internal operation handlers
	fsOps      fileSystemOperations
	dockerOps  dockerOperations
	composeOps composeOperations
	updateOps  updateOperations
}

func NewProcessor(state state.State, checker *checker.CheckResult, files files.Files) Processor {
	p := &processor{
		mu:      &sync.Mutex{},
		state:   state,
		checker: checker,
		files:   files,
	}

	// initialize operation handlers with processor instance
	p.fsOps = newFileSystemOperations(p)
	p.dockerOps = newDockerOperations(p)
	p.composeOps = newComposeOperations(p)
	p.updateOps = newUpdateOperations(p)

	return p
}

func (p *processor) ApplyChanges(ctx context.Context, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationApplyChanges))
	return p.applyChanges(ctx, newOperationState(opts))
}

func (p *processor) CheckFiles(ctx context.Context, stack ProductStack, opts ...OperationOption) (FilesCheckResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationCheckFiles))
	return p.checkFiles(ctx, stack, newOperationState(opts))
}

func (p *processor) FactoryReset(ctx context.Context, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationFactoryReset))
	return p.factoryReset(ctx, newOperationState(opts))
}

func (p *processor) Install(ctx context.Context, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationInstall))
	return p.install(ctx, newOperationState(opts))
}

func (p *processor) Update(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationUpdate))
	return p.update(ctx, stack, newOperationState(opts))
}

func (p *processor) Download(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationDownload))
	return p.download(ctx, stack, newOperationState(opts))
}

func (p *processor) Remove(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationRemove))
	return p.remove(ctx, stack, newOperationState(opts))
}

func (p *processor) Purge(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationPurge))
	return p.purge(ctx, stack, newOperationState(opts))
}

func (p *processor) Start(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationStart))
	return p.start(ctx, stack, newOperationState(opts))
}

func (p *processor) Stop(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationStop))
	return p.stop(ctx, stack, newOperationState(opts))
}

func (p *processor) Restart(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationRestart))
	return p.restart(ctx, stack, newOperationState(opts))
}

func (p *processor) ResetPassword(ctx context.Context, stack ProductStack, opts ...OperationOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	opts = append(opts, withContext(ctx), withOperation(ProcessorOperationResetPassword))
	return p.resetPassword(ctx, stack, newOperationState(opts))
}
