# Terminal Integration Guide for Wizard Screens

This guide covers integration of the `terminal` package into wizard configuration screens, providing command execution capabilities with real-time UI updates.

## Core Architecture

The terminal system consists of three layers:
- **Virtual Terminal (VT)**: Low-level ANSI parsing and screen management (`terminal/vt/`)
- **Terminal Interface**: High-level command execution with PTY/pipe support (`terminal/`)
- **Wizard Integration**: Screen-specific integration patterns (`wizard/models/`)

## Terminal Modes

### PTY Mode (Default on Unix)
- Full pseudoterminal emulation via `creack/pty`
- ANSI escape sequence processing through VT layer
- Interactive command support (vim, less, etc.)
- Proper terminal environment variables

### Pipe Mode (Windows/NoPty)
- Standard stdin/stdout/stderr pipes
- Line-by-line output processing
- Simpler but limited interactivity
- Plain text output handling

## Configuration Options

Terminal behavior is controlled via functional options:

```go
// Essential options for wizard integration
terminal.NewTerminal(width, height,
    terminal.WithAutoScroll(),     // Auto-scroll to bottom on updates
    terminal.WithAutoPoll(),       // Continuous update polling
    terminal.WithCurrentEnv(),     // Inherit process environment
)

// Advanced options
terminal.WithNoStyled()            // Disable ANSI styling (PTY only)
terminal.WithNoPty()              // Force pipe mode
terminal.WithStyle(lipgloss.Style) // Custom viewport styling
```

## Integration Patterns

### Complete Integration Template

```go
type YourFormModel struct {
    *BaseScreen
    terminal terminal.Terminal
    // other screen-specific fields
}

func NewYourFormModel(controller *controllers.StateController, styles *styles.Styles, window *window.Window, args []string) *YourFormModel {
    m := &YourFormModel{}
    m.BaseScreen = NewBaseScreen(controller, styles, window, args, m, nil)
    return m
}

// Required BaseScreenHandler implementation
func (m *YourFormModel) BuildForm() tea.Cmd {
    contentWidth, contentHeight := m.getViewportFormSize()

    // Initialize or reset terminal
    if m.terminal == nil {
        m.terminal = terminal.NewTerminal(
            contentWidth-4,  // Account for border + padding
            contentHeight-1, // Account for border
            terminal.WithAutoScroll(),
            terminal.WithAutoPoll(),
            terminal.WithCurrentEnv(),
        )
    } else {
        m.terminal.Clear()
    }

    // Set initial content
    m.terminal.Append("Terminal initialized...")

    // CRITICAL: Return terminal init for update subscription (idempotent)
    // repeated calls to Init() are safe: only a single waiter will receive
    // the next TerminalUpdateMsg; others will return nil quietly
    return m.terminal.Init()
}
```

### Sizing Calculations

The sizing adjustments account for UI elements:
- **Width -4**: Left border (1) + left padding (1) + right padding (1) + right border (1)
- **Height -1**: Top/bottom borders, content area needs space for text
- Use `m.getViewportFormSize()` from BaseScreen for consistent calculations
- Handle dynamic resizing in Update() method

### Event Flow Architecture

The system uses a single-waiter update notifier for real-time updates:

1. **Update Notifier**: Manages single-waiter update notifications (`teacmd.go`)
2. **Update Messages**: `TerminalUpdateMsg` carries terminal ID
3. **Subscription Model**: Commands wait for `release()` signalling the next update
4. **Auto-polling**: Continuous listening when `WithAutoPoll()` enabled
5. **Single-waiter guarantee**: For a given `Terminal`, at most one pending waiter is active at any time. Multiple `Init()` calls are safe; only one will receive the next `TerminalUpdateMsg` after `release()`, others return nil.

### Complete Update Method Implementation

```go
func (m *YourFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Helper function to handle terminal delegation
    handleTerminal := func(msg tea.Msg) (tea.Model, tea.Cmd) {
        if m.terminal == nil {
            return m, nil
        }

        updatedModel, cmd := m.terminal.Update(msg)
        if terminalModel := terminal.RestoreModel(updatedModel); terminalModel != nil {
            m.terminal = terminalModel
        }
        return m, cmd
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Update terminal size first
        contentWidth, contentHeight := m.getViewportFormSize()
        if m.terminal != nil {
            m.terminal.SetSize(contentWidth-4, contentHeight-1)
        }

        // Update viewports (BaseScreen functionality)
        m.updateViewports()
        return m, nil

    case terminal.TerminalUpdateMsg:
        // Terminal content updated - delegate and continue listening
        // Only a single update message will be emitted per content change,
        // even if Init() was invoked multiple times
        return handleTerminal(msg)

    case tea.KeyMsg:
        // Route keys based on terminal state
        if m.terminal != nil && m.terminal.IsRunning() {
            // Command is running - all keys go to terminal
            return handleTerminal(msg)
        }

        // Terminal is idle - handle screen-specific hotkeys first
        switch msg.String() {
        case "enter":
            // Your screen-specific action
            if m.terminal != nil && !m.terminal.IsRunning() {
                m.executeCommands()
                return m, nil
            }
        case "ctrl+r":
            // Reset functionality
            m.handleReset()
            return m, nil
        }

        // Pass remaining keys to terminal for scrolling
        return handleTerminal(msg)

    default:
        // Other messages (like custom commands) - delegate to terminal
        return handleTerminal(msg)
    }
}
```

### Event Processing Order

Process events in this strict order to ensure proper functionality:

1. **Window Resize**: Update terminal dimensions before any other processing
2. **Terminal Updates**: Handle `TerminalUpdateMsg` immediately for real-time updates
3. **Key Routing**: Route based on `IsRunning()` state
   - **Running commands**: All keys forwarded to terminal for interaction
   - **Idle terminal**: Screen hotkeys first, then terminal scrolling
4. **Other Messages**: Delegate to terminal for potential internal handling

## Command Execution

### Single Command with Error Handling
```go
func (m *YourFormModel) executeCommand() {
    cmd := exec.Command("echo", "hello")

    err := m.terminal.Execute(cmd)
    if err != nil {
        // Display error to user through terminal
        m.terminal.Append(fmt.Sprintf("❌ Command failed: %v", err))
        m.terminal.Append("Please check the command and try again.")
        return
    }

    // Wait for completion if needed
    go func() {
        for m.terminal.IsRunning() {
            time.Sleep(100 * time.Millisecond)
        }
        m.terminal.Append("✅ Command completed successfully")
    }()
}
```

### Sequential Commands with Robust Error Handling
```go
func (m *YourFormModel) executeCommands() {
    if m.terminal.IsRunning() {
        m.terminal.Append("⚠️ Another command is already running")
        return
    }

    commands := []struct {
        cmd     []string
        desc    string
        canFail bool
    }{
        {[]string{"echo", "Starting process..."}, "Initialize", false},
        {[]string{"docker", "--version"}, "Check Docker", true},
        {[]string{"docker-compose", "up", "-d"}, "Start services", false},
    }

    go func() {
        for i, cmdDef := range commands {
            if i > 0 {
                time.Sleep(500 * time.Millisecond)
            }

            m.terminal.Append(fmt.Sprintf("🔄 Step %d: %s", i+1, cmdDef.desc))

            cmd := exec.Command(cmdDef.cmd[0], cmdDef.cmd[1:]...)
            err := m.terminal.Execute(cmd)

            if err != nil {
                m.terminal.Append(fmt.Sprintf("❌ Failed: %v", err))
                if !cmdDef.canFail {
                    m.terminal.Append("💥 Critical error - stopping execution")
                    return
                }
                m.terminal.Append("⚠️ Non-critical error - continuing...")
                continue
            }

            // Wait for command completion
            timeout := time.After(30 * time.Second)
            ticker := time.NewTicker(100 * time.Millisecond)
            completed := false

            for !completed {
                select {
                case <-timeout:
                    m.terminal.Append("⏰ Command timeout - terminating")
                    return
                case <-ticker.C:
                    if !m.terminal.IsRunning() {
                        completed = true
                    }
                }
            }
            ticker.Stop()
        }

        m.terminal.Append("🎉 All commands completed successfully!")
    }()
}

### Interactive Commands
Terminal automatically handles:
- Stdin forwarding in both PTY and pipe modes
- Key-to-input conversion (`key2uv.go`)
- ANSI escape sequence processing (PTY mode)

## Key Input Handling

### PTY Mode
- Full ANSI key sequence support via Ultraviolet conversion
- Vim-style navigation (arrows, page up/down, home/end)
- Control sequences (Ctrl+C, Ctrl+D, etc.)
- Alt+key combinations

### Pipe Mode
- Basic key mapping to stdin bytes
- Enter, space, tab, backspace support
- Control characters (Ctrl+C → \x03)

### Viewport Scrolling
Keys not consumed by running commands are passed to viewport:
- Page Up/Down, Home/End for navigation
- Preserved when terminal is idle

## Lifecycle Management

### Terminal Creation and Cleanup
```go
// Terminal lifecycle follows screen lifecycle
func (m *YourFormModel) BuildForm() tea.Cmd {
    // Create terminal once per screen instance
    if m.terminal == nil {
        m.terminal = terminal.NewTerminal(...)
    } else {
        // Reset content when re-entering screen
        m.terminal.Clear()
    }
    return m.terminal.Init()
}

// No manual cleanup needed - handled by finalizers
// Terminal will be cleaned up when screen model is garbage collected
```

### Screen Navigation Considerations
- **Terminal Persistence**: Terminal remains active during screen navigation
- **Content Reset**: Use `Clear()` when re-entering screens to avoid content buildup
- **Resource Cleanup**: Automatic via finalizers when screen model is destroyed
- **State Preservation**: Terminal state (size, options) persists across `BuildForm()` calls

### State Checking and Debugging
```go
// Essential state checks
if m.terminal == nil {
    // Terminal not initialized - call BuildForm()
}

if m.terminal.IsRunning() {
    // Command is executing - avoid new commands
    // Show spinner or disable UI elements
}

// Debugging helpers
terminalID := m.terminal.ID()           // Unique identifier for logging
width, height := m.terminal.GetSize()  // Current dimensions
view := m.terminal.View()               // Current rendered content

// For debugging terminal content
if DEBUG {
    log.Printf("Terminal %s: %dx%d, running=%t",
        terminalID, width, height, m.terminal.IsRunning())
}
```

### Resource Management Details
Resources managed via Go finalizers (`terminal.go:131-142`):
- **PTY file descriptors**: Automatically closed when terminal is garbage collected
- **Process termination**: Running processes killed during cleanup
- **Notifier shutdown**: Wait channel closed and state reset
- **Mutex-protected cleanup**: Thread-safe resource cleanup
- **No manual Close()**: Resources cleaned automatically, no explicit cleanup needed

## Virtual Terminal Capabilities

The VT layer provides advanced features:
- **Screen Buffer**: Main and alternate screen support
- **Scrollback**: Configurable history buffer
- **ANSI Processing**: Full VT100/xterm compatibility
- **Color Support**: 256-color palette + true color
- **Cursor Modes**: Various cursor styles and visibility
- **Character Sets**: GL/GR charset switching

## Testing Strategies

Key test patterns from `terminal_test.go`:
- **Command Output**: Verify content appears in `View()`
- **Interactive Input**: Simulate key sequences via `Update()`
- **Resource Cleanup**: Manual finalizer calls for verification
- **Concurrent Access**: Multiple goroutines with same terminal
- **Error Handling**: Invalid commands and process failures

## Concurrency and Threading

### Thread Safety
```go
// Terminal methods are thread-safe for these operations:
m.terminal.Append("message")       // Safe from any goroutine
m.terminal.IsRunning()            // Safe to check from any goroutine
m.terminal.ID()                   // Safe to call from any goroutine

// UI operations must be on main thread:
m.terminal.Update(msg)            // Only from main BubbleTea thread
m.terminal.View()                 // Only from main rendering thread
m.terminal.SetSize(w, h)          // Only from main thread
```

### Command Execution Patterns
```go
// CORRECT: Run commands in separate goroutine
go func() {
    m.terminal.Append("Starting long operation...")
    cmd := exec.Command("long-running-command")
    err := m.terminal.Execute(cmd)
    // Error handling...
}()

// INCORRECT: Blocking main thread
cmd := exec.Command("long-running-command")
m.terminal.Execute(cmd) // Will block UI updates
```

### AutoPoll vs Manual Updates
- **WithAutoPoll()**: Continuous listening, higher CPU but immediate updates; still single-waiter per terminal ensures no message storm. Updates are triggered internally via `release()` when content changes.
- **Manual polling**: Call `terminal.Init()` only when needed, lower resource usage
- **Use AutoPoll**: For active terminal screens with frequent updates
- **Skip AutoPoll**: For background or rarely updated terminals

## Troubleshooting Guide

### Terminal Not Updating
**Problem**: Terminal content doesn't appear or update

**Solutions**:
1. Ensure `terminal.Init()` is returned from `BuildForm()`
2. Check `TerminalUpdateMsg` handling in `Update()` method — it should return next wait command to continue listening
3. Verify `handleTerminal()` function calls `RestoreModel()`
4. Add debug logging to track message flow

```go
case terminal.TerminalUpdateMsg:
    log.Printf("Received terminal update: %s", msg.ID)
    return handleTerminal(msg)
```

### Commands Not Executing
**Problem**: `Execute()` returns nil but nothing happens

**Solutions**:
1. Check if previous command is still running: `m.terminal.IsRunning()`
2. Verify command path and arguments
3. Check terminal initialization
4. Add error logging and terminal output

```go
if m.terminal.IsRunning() {
    m.terminal.Append("⚠️ Previous command still running")
    return
}
```

### UI Freezing During Commands
**Problem**: Interface becomes unresponsive

**Solutions**:
1. Always run `Execute()` in goroutines for long commands
2. Use `WithAutoPoll()` for real-time updates
3. Implement proper key routing in `Update()`

### Resource Leaks
**Problem**: Memory or file descriptor leaks

**Solutions**:
1. Avoid creating multiple terminals unnecessarily
2. Let finalizers handle cleanup (don't try manual cleanup)
3. Check for goroutine leaks in command execution

### Size and Layout Issues
**Problem**: Terminal appears cut off or incorrectly sized

**Solutions**:
1. Use proper sizing calculations (width-4, height-1)
2. Handle `tea.WindowSizeMsg` correctly
3. Call `m.updateViewports()` after size changes

## Best Practices

### Initialization Checklist
- ✅ Return `terminal.Init()` from `BuildForm()` (idempotent, single-waiter)
- ✅ Use `WithAutoPoll()` for active terminals
- ✅ Set appropriate dimensions with border adjustments
- ✅ Initialize once per screen, clear content on re-entry

### Event Processing Checklist
- ✅ Handle `TerminalUpdateMsg` first in `Update()` and return next wait command
- ✅ Properly restore terminal models after updates
- ✅ Route keys based on `IsRunning()` state
- ✅ Update terminal size on window resize

### Command Management Checklist
- ✅ Run long operations in goroutines
- ✅ Check `IsRunning()` before new commands
- ✅ Use `Append()` for progress and error messages
- ✅ Implement timeouts for long-running commands
- ✅ Handle both critical and non-critical errors

### Performance Optimization
- **VT Layer**: Automatically caches rendered lines for efficiency
- **Notifier**: Single-waiter, release-based notifications to prevent message storms and deadlocks
- **Resource Cleanup**: Deferred via finalizers to avoid blocking
- **AutoPoll Usage**: Enable only for active terminals requiring real-time updates

## Integration Examples

### Progress Display (`apply_changes.go`)
Shows terminal integration in configuration screen with:
- Dynamic content based on configuration state
- Command execution with progress feedback
- Proper event routing and error handling

### Test Scenarios (`terminal_test.go`)
Demonstrates various usage patterns:
- Simple command output verification
- Interactive input simulation
- Concurrent command execution prevention
- Resource lifecycle management

## Platform Considerations

### Unix Systems
- PTY mode provides full terminal emulation
- ANSI sequences processed through VT layer
- Interactive commands work naturally

### Windows
- Pipe mode used automatically
- Limited interactivity compared to PTY
- Plain text output processing

### Environment Variables
- `TERM=xterm-256color` set automatically in PTY mode
- Current process environment inherited with `WithCurrentEnv()`
- Custom environment via `exec.Cmd.Env`

## Quick Reference

### Essential Methods
```go
// Creation and lifecycle
terminal.NewTerminal(width, height, options...)
m.terminal.Init()                    // Subscribe to updates
m.terminal.Clear()                   // Reset content

// Command execution
m.terminal.Execute(cmd)              // Run command
m.terminal.IsRunning()               // Check execution status
m.terminal.Append(text)              // Add content

// UI integration
m.terminal.Update(msg)               // Handle messages
m.terminal.View()                    // Render content
m.terminal.SetSize(width, height)    // Update dimensions
```

### Common Error Patterns to Avoid
- ❌ Creating multiple terminals per screen
- ❌ Running `Execute()` on main thread for long commands
- ❌ Forgetting to return `terminal.Init()` from `BuildForm()`
- ❌ Not handling `TerminalUpdateMsg` in `Update()`
- ❌ Calling UI methods from background goroutines
- ❌ Manual resource cleanup (use finalizers instead)

### Integration Checklist for New Screens
1. ✅ Add `terminal terminal.Terminal` to model struct
2. ✅ Initialize in `BuildForm()` with proper sizing
3. ✅ Return `terminal.Init()` from `BuildForm()`
4. ✅ Handle `TerminalUpdateMsg` first in `Update()`
5. ✅ Implement proper key routing based on `IsRunning()`
6. ✅ Handle window resize events
7. ✅ Run commands in goroutines with error handling
8. ✅ Add progress feedback via `Append()`

This comprehensive architecture provides robust terminal integration with wizard screens while maintaining proper resource management, real-time UI updates, and cross-platform compatibility.
