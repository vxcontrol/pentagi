package processor

import (
	"context"
	"strings"
	"sync"

	"pentagi/cmd/installer/wizard/terminal"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// operationState holds execution options and state for processor operations
type operationState struct {
	id            string            // unique identifier for the command
	force         bool              // attempt maximum operations
	terminal      terminal.Terminal // embedded terminal model for interactive display
	operation     ProcessorOperation
	passwordValue string // password value for reset password operation

	// message chain for reply
	mx     *sync.Mutex
	ctx    context.Context
	output strings.Builder
	msgs   []tea.Msg
}

// ProcessorOutputMsg contains command output line
type ProcessorOutputMsg struct {
	ID        string
	Output    string
	Operation ProcessorOperation
	Stack     ProductStack

	// keeps for continuing the message chain
	state *operationState
	num   int
}

// ProcessorCompletionMsg signals operation completion
type ProcessorCompletionMsg struct {
	ID        string
	Error     error
	Operation ProcessorOperation
	Stack     ProductStack

	// keeps for continuing the message chain
	state *operationState
	num   int
}

// ProcessorStartedMsg signals operation start
type ProcessorStartedMsg struct {
	ID        string
	Operation ProcessorOperation
	Stack     ProductStack

	// keeps for continuing the message chain
	state *operationState
	num   int
}

// ProcessorWaitMsg signals operation wait for
type ProcessorWaitMsg struct {
	ID        string
	Error     error
	Operation ProcessorOperation
	Stack     ProductStack

	// keeps for continuing the message chain
	state *operationState
	num   int
}

// ProcessorFilesCheckMsg carries file statuses computed in check
type ProcessorFilesCheckMsg struct {
	ID     string
	Stack  ProductStack
	Result FilesCheckResult
	Error  error

	// keeps for continuing the message chain
	state *operationState
	num   int
}

type OperationOption func(c *operationState)

func withID(id string) OperationOption {
	return func(c *operationState) { c.id = id }
}

func withOperation(operation ProcessorOperation) OperationOption {
	return func(c *operationState) { c.operation = operation }
}

func withContext(ctx context.Context) OperationOption {
	return func(c *operationState) { c.ctx = ctx }
}

// helper to build operation state with defaults
func newOperationState(opts []OperationOption) *operationState {
	state := &operationState{
		id:   uuid.New().String(),
		mx:   &sync.Mutex{},
		ctx:  context.Background(),
		msgs: []tea.Msg{},
	}

	for _, opt := range opts {
		opt(state)
	}

	if state.terminal == nil {
		state.terminal = terminal.NewTerminal(
			80, 24,
			terminal.WithAutoScroll(),
			terminal.WithAutoPoll(),
			terminal.WithCurrentEnv(),
			terminal.WithNoPty(),
		)
	}

	return state
}

// helper to send output message
func (state *operationState) sendOutput(output string, isPartial bool, stack ProductStack) {
	state.mx.Lock()
	defer state.mx.Unlock()

	if isPartial {
		state.output.WriteString(output)
		state.output.WriteString("\n")
	} else {
		state.output.Reset()
		state.output.WriteString(output)
	}

	state.msgs = append(state.msgs, ProcessorOutputMsg{
		ID:        state.id,
		Output:    state.output.String(),
		Operation: state.operation,
		Stack:     stack,
		state:     state,
		num:       len(state.msgs) + 1,
	})
}

// helper to send completion message
func (state *operationState) sendCompletion(stack ProductStack, err error) {
	state.mx.Lock()
	defer state.mx.Unlock()

	state.msgs = append(state.msgs, ProcessorCompletionMsg{
		ID:        state.id,
		Error:     err,
		Operation: state.operation,
		Stack:     stack,
		state:     state,
		num:       len(state.msgs) + 1,
	})
}

// helper to send started message
func (state *operationState) sendStarted(stack ProductStack) {
	state.mx.Lock()
	defer state.mx.Unlock()

	state.msgs = append(state.msgs, ProcessorStartedMsg{
		ID:        state.id,
		Operation: state.operation,
		Stack:     stack,
		state:     state,
		num:       len(state.msgs) + 1,
	})
}

// helper to send files check message
func (state *operationState) sendFilesCheck(stack ProductStack, result FilesCheckResult, err error) {
	state.mx.Lock()
	defer state.mx.Unlock()

	state.msgs = append(state.msgs, ProcessorFilesCheckMsg{
		ID:     state.id,
		Stack:  stack,
		Result: result,
		Error:  err,
		state:  state,
		num:    len(state.msgs) + 1,
	})
}
