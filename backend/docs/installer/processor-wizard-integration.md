# Processor-Wizard Integration Guide

> Technical documentation for embedded terminal integration in PentAGI installer wizard using Bubble Tea and pseudoterminals.

## Architecture Decision: Pseudoterminal vs Built-in Bubble Tea

After extensive research of `tea.ExecProcess` and alternative approaches, **pseudoterminal solution was chosen** for the following reasons:

**Why Not `tea.ExecProcess`:**
- ❌ Fullscreen takeover - cannot be embedded in viewport regions
- ❌ Blocking execution - pauses entire Bubble Tea program
- ❌ No real-time output streaming to specific UI components
- ❌ Cannot handle interactive Docker commands within constrained areas

**Built-in Approach Limitations:**
- Limited to `os/exec` with pipes (no true terminal semantics)
- Manual ANSI escape sequence handling required
- Reduced interactivity (no Ctrl+C, terminal properties)
- Complex input/output coordination

**Pseudoterminal Advantages:**
- ✅ Embedded in specific viewport regions
- ✅ Real-time command output with ANSI colors/formatting
- ✅ Full interactivity (stdin/stdout/stderr, Ctrl+C)
- ✅ Professional terminal experience within TUI
- ✅ Docker compatibility (`docker exec -it`, progress bars)

## Core Architecture

### Key Components

**`Processor`** interface (defined in [`processor.go`](../../cmd/installer/processor/processor.go)):
```go
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
```

**Bubble Tea integration** in [`model.go`](../../cmd/installer/processor/model.go), [`state.go`](../../cmd/installer/processor/state.go), and [`logic.go`](../../cmd/installer/processor/logic.go):
- Wraps processor operations as `tea.Cmd` values through `ProcessorModel`
- Streams operation updates through `ProcessorStartedMsg`, `ProcessorOutputMsg`, `ProcessorFilesCheckMsg`, and `ProcessorCompletionMsg`
- Executes commands through the shared `terminal.Terminal` abstraction when available
- Falls back to non-ANSI compose output on small terminals and Windows by setting `COMPOSE_ANSI=never`

### Integration Pattern

**Wizard Screen Integration** (see [`apply_changes.go`](../../cmd/installer/wizard/models/apply_changes.go)):
```go
type ApplyChangesFormModel struct {
    processor processor.ProcessorModel
    running   bool
    terminal  terminal.Terminal
}

// Create terminal model for the screen
m.terminal = terminal.NewTerminal(width, height, terminal.WithAutoScroll(), terminal.WithAutoPoll())

// Start operation through the processor model
return m.processor.ApplyChanges(context.Background(), processor.WithTerminal(m.terminal))
```

## Message Flow Architecture

The integration uses a **buffered channel-based approach** for real-time terminal output streaming:

```mermaid
graph TD
    A[User Action - Enter] --> B[ProcessorModel Command]
    B --> C[Operation State]
    C --> D[Terminal Abstraction]
    D --> E[Execute Command - docker compose]
    E --> F[Terminal Update Loop]
    F --> G[Processor Messages]
    G --> H[HandleMsg Polling]
    H --> I[ProcessorOutputMsg]
    I --> J[appendOutput + Viewport Update]
    J --> K[UI Refresh - Real-time Display]

    L[User Input] --> M[handleTerminalInput]
    M --> N{Key Type?}
    N -->|Scroll Keys| O[Viewport Handling]
    N -->|Other Keys| P[keyToTerminalSequence]
    P --> Q[Direct PTY Write]

    subgraph "Real-time Flow"
        F
        G
        H
        I
        J
    end

    subgraph "Input Handling"
        M
        N
        O
        P
        Q
    end

    subgraph "Terminal Display"
        C
        R[Blue Border Only]
        S[Maximized Content]
        T[Auto-scroll Bottom]
    end
```

## Apply Changes integrity pre-check (Wizard)

Before invoking `processor.ApplyChanges()`, the Apply Changes screen performs an embedded files integrity scan:

- Enter: start async scan using `GetStackFilesStatus(files, ProductStackAll, workingDir)`
- If outdated/missing files found: prompt user to update (Y) or proceed without updates (N)
- Ctrl+C: cancel the integrity stage and return to initial instruction screen

Hotkeys on this screen:
- Initial: Enter
- During scan: Ctrl+C
- When prompt is shown: Y/N, Ctrl+C

Depending on choice, `processor.ApplyChanges()` is called with/without `WithForce()`. This keeps user in control of overwriting modified files while still allowing a smooth path when no updates are required.

Note: the integrity prompt lists only modified files; missing files are considered normal on a fresh installation and are not shown to the user.

**Key Improvements:**
- **50ms polling** via `waitForOutput()` for responsive UI updates
- **Buffered channel** (100 messages) prevents blocking
- **Direct key mapping** - no intermediate input buffers
- **Viewport delegation** for scrolling (PageUp/PageDown, mouse wheel)

## Implementation Guide

### 1. Screen Model Setup

Add terminal integration to any wizard screen following this pattern:

```go
type YourFormModel struct {
    *BaseScreen
    processor processor.ProcessorModel
    terminal  terminal.Terminal
}

func (m *YourFormModel) BuildForm() tea.Cmd {
    contentWidth, contentHeight := m.getViewportFormSize()

    if m.terminal == nil {
        m.terminal = terminal.NewTerminal(
            contentWidth-2,
            contentHeight-1,
            terminal.WithAutoScroll(),
            terminal.WithAutoPoll(),
            terminal.WithCurrentEnv(),
        )
    }

    return m.terminal.Init()
}
```

### 2. Update Method Integration

Handle terminal model updates with proper type assertion and input delegation:

```go
func (m *YourFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // Update terminal model first (handles real-time output)
    if m.terminal != nil {
        updatedModel, terminalCmd := m.terminal.Update(msg)
        if terminalModel := terminal.RestoreModel(updatedModel); terminalModel != nil {
            m.terminal = terminalModel
        }
        if terminalCmd != nil {
            cmds = append(cmds, terminalCmd)
        }
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Update terminal size dynamically
        if m.terminal != nil {
            contentWidth, contentHeight := m.getViewportFormSize()
            m.terminal.SetSize(contentWidth-2, contentHeight-1)
        }

    case tea.KeyMsg:
        // Terminal takes priority when operation is running
        if m.terminal != nil && m.terminal.IsRunning() {
            return m, tea.Batch(cmds...)  // Terminal already processed input
        }
        // Your screen-specific hotkeys here...
    }

    return m, tea.Batch(cmds...)
}
```

### 3. Operation Triggers

Start processor operations through terminal model:

```go
// In your action handler
func (m *YourFormModel) startProcess() tea.Cmd {
    return m.processor.Install(
        context.Background(),
        processor.WithTerminal(m.terminal),
    )
}
```

### 4. Display Integration

Render terminal within your screen layout - simplified approach:

```go
func (m *YourFormModel) renderMainPanel() string {
    if m.terminal != nil {
        // Terminal model returns complete styled content with blue border
        return m.terminal.View()
    }

    return m.GetStyles().Error.Render(locale.ApplyChangesTerminalIsNotInitialized)
}
```

**Terminal View Structure:**
- **No header/footer** - maximized content space
- **Blue border only** - clean visual boundaries
- **Auto-scrolling viewport** - always shows latest output
- **Responsive sizing** - adapts to window changes via `SetSize()`

## Features & Capabilities

### Interactive Terminal
- **Real-time output**: 50ms polling via buffered channel (100 messages)
- **Native input handling**: Direct key-to-terminal-sequence conversion
- **ANSI support**: Colors, formatting, progress bars rendered correctly
- **Smart scrolling**: PageUp/PageDown and mouse wheel for viewport, all other keys to PTY

### Docker Integration
- **Interactive commands**: `docker exec -it container bash`
- **Progress visualization**: Docker pull progress bars with colors
- **Color output**: Docker's colored status messages preserved
- **Signal handling**: Ctrl+C, Ctrl+D, Ctrl+Z properly handled

### UI Features
- **Maximized space**: No headers, input lines, or inner borders
- **Blue border styling**: Clean visual boundaries with `lipgloss.Color("62")`
- **Dynamic resizing**: `SetSize()` method for window changes
- **Toggle mode**: Ctrl+T switches between embedded terminal and message logs
- **Auto-scroll**: Always shows latest output, bottom-aligned

## Command Configuration

Processor operations support flexible configuration via [`processor.go`](../../cmd/installer/processor/processor.go):

```go
// Available options
processor.WithForce()                   // Skip validation checks
processor.WithTerminal(terminal)        // Terminal-backed integration
processor.WithPasswordValue(password)   // Reset password operation input
```

Choose integration method based on screen requirements:
- **Embedded terminal**: Interactive operations, real-time output, full PTY support
- **Processor model**: Bubble Tea command wrappers and message polling for wizard screens (see [`model.go`](../../cmd/installer/processor/model.go))

**Alternative Integration:** For simpler use cases or Windows compatibility, the current processor model uses `ProcessorOutputMsg` events from [`state.go`](../../cmd/installer/processor/state.go) and disables Docker Compose ANSI output when needed in [`logic.go`](../../cmd/installer/processor/logic.go).

## Limitations & Considerations

### Performance
- **Memory usage**: Output buffer auto-managed, channel limited to 100 messages
- **Goroutine management**: Automatic cleanup via `defer close(outputChan)`
- **Resource overhead**: Minimal - one goroutine per operation, 1ms throttling
- **Update frequency**: 50ms polling prevents UI blocking

### Platform Compatibility
- **Unix/Linux/macOS**: Full pseudoterminal support via `github.com/creack/pty v1.1.21`
- **Windows**: Compose ANSI output is disabled in the command runner when the terminal is small or the host is Windows

### UI Constraints
- **Minimum size**: `width-2, height-2` for border space
- **Input delegation**: Terminal captures all input except scroll keys when running
- **Layout integration**: Single terminal per screen, full area utilization

## Testing & Debugging

### Processor Model Testing
Processor model behavior can be tested independently of wizard integration (see [`logic_test.go`](../../cmd/installer/processor/logic_test.go) and [`mock_test.go`](../../cmd/installer/processor/mock_test.go) for current mock patterns).

### Debug Features
- **Ctrl+T toggle**: Switch to message-based mode for debugging
- **Output logging**: All terminal output available for inspection
- **Error reporting**: Terminal errors bubble up to UI state

## Future Enhancements

### Potential Improvements
- **Session recording**: Capture terminal sessions for replay/debugging
- **Multiple terminals**: Support for concurrent operations with tabs
- **Terminal themes**: Customizable color schemes via lipgloss
- **Buffer persistence**: Save terminal output between screen switches
- **Copy/paste**: Terminal text selection and clipboard integration

### Integration Opportunities
- **Service management**: Real-time `docker ps` monitoring in terminal
- **Log streaming**: Live log viewing with `docker logs -f`
- **Interactive debugging**: Terminal-based troubleshooting tools
- **Configuration editing**: Embedded editors for compose files

### Architecture Extensions
- **Message broadcasting**: Share terminal output across multiple UI components
- **Operation queuing**: Sequential command execution with progress tracking
- **Error recovery**: Automatic retry mechanisms with user confirmation

This clean, channel-based architecture provides a maintainable foundation for embedding professional terminal functionality within Bubble Tea applications while maximizing screen real estate and user experience.
