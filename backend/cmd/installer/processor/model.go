package processor

import (
	"context"
	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// processorModel implements interface for bubbletea integration
type processorModel struct {
	*processor
}

type ProcessorModel interface {
	ApplyChanges(ctx context.Context, opts ...OperationOption) tea.Cmd
	CheckFiles(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	FactoryReset(ctx context.Context, opts ...OperationOption) tea.Cmd
	Install(ctx context.Context, opts ...OperationOption) tea.Cmd
	Update(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	Download(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	Remove(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	Purge(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	Start(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	Stop(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	Restart(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	ResetPassword(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd
	HandleMsg(msg tea.Msg) tea.Cmd
}

func NewProcessorModel(state state.State, checker *checker.CheckResult, files files.Files) ProcessorModel {
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

	return &processorModel{processor: p}
}

func wrapCommand(
	ctx context.Context, stack ProductStack, mu *sync.Mutex, ch <-chan error,
	fn func(state *operationState), opts ...OperationOption,
) tea.Cmd {
	state := newOperationState(opts)
	go func() {
		mu.Lock()
		fn(state)
		mu.Unlock()
	}()

	teaCmdWaitMsg := func(err error) tea.Cmd {
		return func() tea.Msg {
			return ProcessorWaitMsg{
				ID:        state.id,
				Error:     err,
				Operation: state.operation,
				Stack:     stack,
				state:     state,
			}
		}
	}

	select {
	case <-ctx.Done():
		return teaCmdWaitMsg(ctx.Err())
	case err := <-ch:
		return teaCmdWaitMsg(err)
	case <-time.After(500 * time.Millisecond):
		return teaCmdWaitMsg(nil)
	}
}

func (pm *processorModel) ApplyChanges(ctx context.Context, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, ProductStackAll, pm.mu, ch, func(state *operationState) {
		ch <- pm.applyChanges(ctx, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationApplyChanges))...)
}

func (pm *processorModel) CheckFiles(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		_, err := pm.checkFiles(ctx, stack, state)
		ch <- err
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationCheckFiles))...)
}

func (pm *processorModel) FactoryReset(ctx context.Context, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, ProductStackAll, pm.mu, ch, func(state *operationState) {
		ch <- pm.factoryReset(ctx, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationFactoryReset))...)
}

func (pm *processorModel) Install(ctx context.Context, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, ProductStackAll, pm.mu, ch, func(state *operationState) {
		ch <- pm.install(ctx, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationInstall))...)
}

func (pm *processorModel) Update(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.update(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationUpdate))...)
}

func (pm *processorModel) Download(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.download(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationDownload))...)
}

func (pm *processorModel) Remove(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.remove(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationRemove))...)
}

func (pm *processorModel) Purge(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.purge(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationPurge))...)
}

func (pm *processorModel) Start(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.start(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationStart))...)
}

func (pm *processorModel) Stop(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.stop(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationStop))...)
}

func (pm *processorModel) Restart(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.restart(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationRestart))...)
}

func (pm *processorModel) ResetPassword(ctx context.Context, stack ProductStack, opts ...OperationOption) tea.Cmd {
	ch := make(chan error, 1)
	return wrapCommand(ctx, stack, pm.mu, ch, func(state *operationState) {
		ch <- pm.resetPassword(ctx, stack, state)
	}, append(opts, withContext(ctx), withOperation(ProcessorOperationResetPassword))...)
}

func (pm *processorModel) HandleMsg(msg tea.Msg) tea.Cmd {
	newWaitMsg := func(num int, stack ProductStack, state *operationState, err error) tea.Cmd {
		if state == nil {
			return nil // no state, no poll
		}

		return func() tea.Msg {
			return ProcessorWaitMsg{
				ID:        state.id,
				Error:     err,
				Operation: state.operation,
				Stack:     stack,
				state:     state,
				num:       num,
			}
		}
	}
	pollMsg := func(num int, stack ProductStack, state *operationState) tea.Cmd {
		if state == nil {
			return nil // no state, no poll
		}

		state.mx.Lock()
		ctx := state.ctx
		msgs := state.msgs
		state.mx.Unlock()

		if num < len(msgs) {
			nextMsg := msgs[num]
			return func() tea.Msg {
				return nextMsg
			}
		}

		select {
		case <-ctx.Done():
			return newWaitMsg(num, stack, state, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			return newWaitMsg(num, stack, state, nil)
		}
	}

	switch msg := msg.(type) {
	case ProcessorWaitMsg:
		if msg.Error != nil {
			// stop polling after error
			return nil
		}
		return pollMsg(msg.num, msg.Stack, msg.state)
	case ProcessorOutputMsg:
		return pollMsg(msg.num, msg.Stack, msg.state)
	case ProcessorFilesCheckMsg:
		return pollMsg(msg.num, msg.Stack, msg.state)
	case ProcessorCompletionMsg:
		return nil // final message, no poll
	case ProcessorStartedMsg:
		return pollMsg(msg.num, msg.Stack, msg.state)
	default:
		return nil // unknown message, no poll
	}
}
