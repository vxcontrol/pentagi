# Charm.sh Debugging & Troubleshooting Guide

> Comprehensive guide to debugging TUI applications and avoiding common pitfalls.

## 🔧 **TUI-Safe Logging System**

### **File-Based Logger Pattern**
**Problem**: fmt.Printf breaks TUI rendering
**Solution**: File-based logger with structured output

```go
// logger.go - TUI-safe logging implementation
package logger

import (
    "encoding/json"
    "os"
    "time"
)

type LogEntry struct {
    Timestamp string `json:"timestamp"`
    Level     string `json:"level"`
    Component string `json:"component"`
    Message   string `json:"message"`
    Data      any    `json:"data,omitempty"`
}

var logFile *os.File

func init() {
    var err error
    logFile, err = os.OpenFile("log.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        panic(err)
    }
}

func Log(format string, args ...any) {
    writeLog("INFO", fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
    writeLog("ERROR", fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...any) {
    writeLog("DEBUG", fmt.Sprintf(format, args...))
}

func LogWithData(message string, data any) {
    entry := LogEntry{
        Timestamp: time.Now().Format(time.RFC3339),
        Level:     "INFO",
        Message:   message,
        Data:      data,
    }

    jsonData, _ := json.Marshal(entry)
    logFile.Write(append(jsonData, '\n'))
}

func writeLog(level, message string) {
    entry := LogEntry{
        Timestamp: time.Now().Format(time.RFC3339),
        Level:     level,
        Message:   message,
    }

    jsonData, _ := json.Marshal(entry)
    logFile.Write(append(jsonData, '\n'))
}
```

### **Development Monitoring**
```bash
# Monitor logs in separate terminal during development
tail -f log.json | jq '.'

# Filter by component
tail -f log.json | jq 'select(.component == "FormModel")'

# Filter by level
tail -f log.json | jq 'select(.level == "ERROR")'

# Real-time pretty printing
tail -f log.json | jq -r '"\(.timestamp) [\(.level)] \(.message)"'
```

### **Safe Debug Output**
```go
// ❌ NEVER: Breaks TUI rendering
fmt.Println("debug")
log.Println("debug")
os.Stdout.WriteString("debug")

// ✅ ALWAYS: File-based logging
logger.Log("[Component] Event: %v", msg)
logger.Log("[Model] UPDATE: key=%s", msg.String())
logger.Log("[Model] VIEWPORT: %dx%d ready=%v", width, height, m.ready)
logger.Errorf("[Model] ERROR: %v", err)

// Development pattern
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    logger.Log("[%s] UPDATE: %T", m.componentName, msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        logger.Log("[%s] KEY: %s", m.componentName, msg.String())
        return m.handleKeyMsg(msg)
    }

    return m, nil
}
```

## 🔧 **Key Debugging Techniques**

### **Dimension Debugging**
```go
func (m *Model) debugDimensions() {
    width, height := m.styles.GetSize()
    contentWidth, contentHeight := m.window.GetContentSize()

    logger.LogWithData("Dimensions Debug", map[string]interface{}{
        "terminal_size":    fmt.Sprintf("%dx%d", width, height),
        "content_size":     fmt.Sprintf("%dx%d", contentWidth, contentHeight),
        "viewport_size":    fmt.Sprintf("%dx%d", m.viewport.Width, m.viewport.Height),
        "viewport_offset":  m.viewport.YOffset,
        "viewport_percent": m.viewport.ScrollPercent(),
        "is_vertical":      m.isVerticalLayout(),
    })
}

func (m *Model) View() string {
    width, height := m.styles.GetSize()
    if width <= 0 || height <= 0 {
        logger.Log("[%s] VIEW: invalid dimensions %dx%d", m.componentName, width, height)
        return "Loading..." // Graceful fallback
    }

    // Debug dimensions on resize
    if m.lastWidth != width || m.lastHeight != height {
        m.debugDimensions()
        m.lastWidth, m.lastHeight = width, height
    }

    return m.viewport.View()
}
```

### **Navigation Stack Debugging**
```go
func (n *Navigator) debugStack() {
    stackInfo := make([]string, len(n.stack))
    for i, screenID := range n.stack {
        stackInfo[i] = string(screenID)
    }

    logger.LogWithData("Navigation Stack", map[string]interface{}{
        "stack":   stackInfo,
        "current": string(n.Current()),
        "depth":   len(n.stack),
    })
}

func (n *Navigator) Push(screenID ScreenID) {
    logger.Log("[Navigator] PUSH: %s", string(screenID))
    n.stack = append(n.stack, screenID)
    n.debugStack()
    n.persistState()
}

func (n *Navigator) Pop() ScreenID {
    if len(n.stack) <= 1 {
        logger.Log("[Navigator] POP: cannot pop last screen")
        return n.stack[0]
    }

    popped := n.stack[len(n.stack)-1]
    n.stack = n.stack[:len(n.stack)-1]
    logger.Log("[Navigator] POP: %s -> %s", string(popped), string(n.Current()))
    n.debugStack()
    n.persistState()
    return popped
}
```

### **Form State Debugging**
```go
func (m *FormModel) debugFormState() {
    fields := make([]map[string]interface{}, len(m.fields))
    for i, field := range m.fields {
        fields[i] = map[string]interface{}{
            "key":         field.Key,
            "value":       field.Input.Value(),
            "placeholder": field.Input.Placeholder,
            "focused":     i == m.focusedIndex,
            "width":       field.Input.Width,
        }
    }

    logger.LogWithData("Form State", map[string]interface{}{
        "focused_index": m.focusedIndex,
        "has_changes":   m.hasChanges,
        "show_values":   m.showValues,
        "field_count":   len(m.fields),
        "fields":        fields,
    })
}

func (m *FormModel) validateField(index int) {
    logger.Log("[FormModel] VALIDATE: field %d (%s)", index, m.fields[index].Key)

    // ... validation logic ...

    if hasError {
        logger.Log("[FormModel] VALIDATE: field %s failed - %s",
            m.fields[index].Key, errorMsg)
    }
}
```

### **Content Loading Debugging**
```go
func (m *Model) loadContent() tea.Cmd {
    return func() tea.Msg {
        logger.Log("[%s] LOAD: starting content load", m.componentName)

        // Try multiple sources with detailed logging
        sources := []func() (string, error){
            m.loadFromEmbedded,
            m.loadFromFile,
            m.loadFromFallback,
        }

        for i, loadFunc := range sources {
            logger.Log("[%s] LOAD: trying source %d", m.componentName, i+1)

            content, err := loadFunc()
            if err != nil {
                logger.Errorf("[%s] LOAD: source %d failed: %v", m.componentName, i+1, err)
                continue
            }

            logger.Log("[%s] LOAD: source %d success (%d chars)",
                m.componentName, i+1, len(content))
            return ContentLoadedMsg{content}
        }

        logger.Errorf("[%s] LOAD: all sources failed", m.componentName)
        return ErrorMsg{fmt.Errorf("failed to load content")}
    }
}
```

## 🔧 **Common Pitfalls & Solutions**

### **1. Glamour Renderer Freezing**
**Problem**: Creating new renderer instances can freeze
**Solution**: Single shared renderer in styles.New()

```go
// ❌ WRONG: New renderer each time
func (m *Model) renderMarkdown(content string) string {
    renderer, _ := glamour.NewTermRenderer(...)  // Can freeze!
    return renderer.Render(content)
}

// ✅ CORRECT: Shared renderer instance
func (m *Model) renderMarkdown(content string) string {
    rendered, err := m.styles.GetRenderer().Render(content)
    if err != nil {
        logger.Errorf("[%s] RENDER: glamour error: %v", m.componentName, err)
        // Fallback to plain text
        return fmt.Sprintf("# Content\n\n%s\n\n*Render error: %v*", content, err)
    }
    return rendered
}

// Debug renderer creation
func NewStyles() *Styles {
    logger.Log("[Styles] Creating glamour renderer")
    renderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    if err != nil {
        logger.Errorf("[Styles] Failed to create renderer: %v", err)
        panic(err)
    }
    logger.Log("[Styles] Glamour renderer created successfully")
    return &Styles{renderer: renderer}
}
```

### **2. Footer Height Inconsistency**
**Problem**: Border-based footers vary in height
**Solution**: Background approach with padding

```go
// ❌ WRONG: Border approach (height varies)
func createFooterWrong(width int, text string) string {
    logger.Log("[Footer] Using border approach - height may vary")
    return lipgloss.NewStyle().
        Height(1).
        Border(lipgloss.Border{Top: true}).
        Render(text)
}

// ✅ CORRECT: Background approach (exactly 1 line)
func createFooter(width int, text string) string {
    logger.Log("[Footer] Using background approach - consistent height")
    return lipgloss.NewStyle().
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("255")).
        Padding(0, 1, 0, 1).
        Render(text)
}

// Debug footer height
func (a *App) View() string {
    header := a.renderHeader()
    footer := a.renderFooter()
    content := a.currentModel.View()

    headerHeight := lipgloss.Height(header)
    footerHeight := lipgloss.Height(footer)

    logger.LogWithData("Layout Heights", map[string]interface{}{
        "header_height": headerHeight,
        "footer_height": footerHeight,
        "total_height":  a.styles.GetHeight(),
    })

    contentHeight := max(a.styles.GetHeight() - headerHeight - footerHeight, 0)
    contentArea := a.styles.Content.Height(contentHeight).Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}
```

### **3. Dimension Synchronization Issues**
**Problem**: Models store their own width/height, get out of sync
**Solution**: Centralize dimensions in styles singleton

```go
// ❌ WRONG: Models managing their own dimensions
type ModelWrong struct {
    width, height int  // Will get out of sync!
}

func (m *ModelWrong) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height  // Inconsistent!
        logger.Log("[Model] Direct dimension update: %dx%d", m.width, m.height)
    }
}

// ✅ CORRECT: Centralized dimension management
type Model struct {
    styles *styles.Styles  // Access via styles.GetSize()
}

func (m *Model) updateViewport() {
    width, height := m.styles.GetSize()
    logger.Log("[%s] Using centralized dimensions: %dx%d", m.componentName, width, height)

    if width <= 0 || height <= 0 {
        logger.Log("[%s] Invalid dimensions, skipping update", m.componentName)
        return
    }

    // ... safe viewport update
}
```

### **4. TUI Rendering Corruption**
**Problem**: Console output breaks rendering
**Solution**: File-based logger, never fmt.Printf

```go
// ❌ NEVER: Use tea.ClearScreen during navigation
func (a *App) handleNavigation() (tea.Model, tea.Cmd) {
    logger.Log("[App] Navigation: using ClearScreen")
    return a, tea.Batch(cmd, tea.ClearScreen)  // Corrupts rendering!
}

// ✅ CORRECT: Let model Init() handle clean state
func (a *App) handleNavigation() (tea.Model, tea.Cmd) {
    logger.Log("[App] Navigation: clean model initialization")
    return a, a.currentModel.Init()
}

// Debug rendering corruption
func (m *Model) View() string {
    view := m.viewport.View()

    // Debug view corruption
    if strings.Contains(view, "\x1b[2J") || strings.Contains(view, "\x1b[H") {
        logger.Errorf("[%s] VIEW: detected ANSI clear sequences", m.componentName)
    }

    logger.Log("[%s] VIEW: rendered %d chars", m.componentName, len(view))
    return view
}
```

### **5. Navigation State Issues**
**Problem**: Models retain state between visits
**Solution**: Complete state reset in Init()

```go
// ❌ WRONG: Partial state reset
func (m *Model) Init() tea.Cmd {
    logger.Log("[%s] INIT: partial reset", m.componentName)
    m.content = ""  // Only resetting some fields!
    return m.loadContent
}

// ✅ CORRECT: Complete state reset
func (m *Model) Init() tea.Cmd {
    logger.Log("[%s] INIT: complete state reset", m.componentName)

    // Reset ALL state fields
    m.content = ""
    m.ready = false
    m.error = nil
    m.initialized = false
    m.scrolled = false
    m.focusedIndex = 0
    m.hasChanges = false

    // Reset component state
    m.viewport.GotoTop()
    m.viewport.SetContent("")

    // Reset form state if applicable
    for i := range m.fields {
        m.fields[i].Input.Blur()
    }

    logger.Log("[%s] INIT: state reset complete", m.componentName)
    return m.loadContent
}
```

## 🔧 **Performance Debugging**

### **Viewport Performance**
```go
func (m *Model) debugViewportPerformance() {
    start := time.Now()

    // Measure viewport operations
    m.viewport.SetContent(m.content)
    setContentDuration := time.Since(start)

    start = time.Now()
    view := m.viewport.View()
    viewDuration := time.Since(start)

    logger.LogWithData("Viewport Performance", map[string]interface{}{
        "content_size":       len(m.content),
        "rendered_size":      len(view),
        "set_content_ms":     setContentDuration.Milliseconds(),
        "view_render_ms":     viewDuration.Milliseconds(),
        "viewport_height":    m.viewport.Height,
        "total_lines":        strings.Count(m.content, "\n"),
        "scroll_percent":     m.viewport.ScrollPercent(),
    })
}
```

### **Memory Usage Tracking**
```go
import "runtime"

func (m *Model) debugMemoryUsage(operation string) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    logger.LogWithData("Memory Usage", map[string]interface{}{
        "operation":      operation,
        "alloc_mb":       memStats.Alloc / 1024 / 1024,
        "total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
        "sys_mb":         memStats.Sys / 1024 / 1024,
        "num_gc":         memStats.NumGC,
    })
}

// Usage in critical operations
func (m *Model) updateFormContent() {
    m.debugMemoryUsage("form_update_start")

    // ... form update logic ...

    m.debugMemoryUsage("form_update_end")
}
```

## 🔧 **Error Recovery Patterns**

### **Graceful Degradation**
```go
func (m *Model) View() string {
    defer func() {
        if r := recover(); r != nil {
            logger.Errorf("[%s] VIEW: panic recovered: %v", m.componentName, r)
        }
    }()

    // Multi-level fallbacks
    width, height := m.styles.GetSize()
    if width <= 0 || height <= 0 {
        logger.Log("[%s] VIEW: invalid dimensions, using fallback", m.componentName)
        return "Loading..."
    }

    if m.error != nil {
        logger.Log("[%s] VIEW: error state, showing error message", m.componentName)
        return m.styles.Error.Render("Error: " + m.error.Error())
    }

    if !m.ready {
        logger.Log("[%s] VIEW: not ready, showing loading", m.componentName)
        return m.styles.Info.Render("Loading content...")
    }

    return m.viewport.View()
}
```

### **State Recovery**
```go
func (m *Model) recoverFromError(err error) tea.Cmd {
    logger.Errorf("[%s] ERROR: %v", m.componentName, err)

    // Try to recover state
    m.error = err
    m.ready = true

    // Attempt graceful recovery
    return func() tea.Msg {
        logger.Log("[%s] RECOVERY: attempting state recovery", m.componentName)

        // Try to reload content
        if content, loadErr := m.loadFallbackContent(); loadErr == nil {
            logger.Log("[%s] RECOVERY: fallback content loaded", m.componentName)
            return ContentLoadedMsg{content}
        }

        logger.Log("[%s] RECOVERY: using minimal content", m.componentName)
        return ContentLoadedMsg{"# Error\n\nContent temporarily unavailable."}
    }
}
```

## 🔧 **Testing Strategies**

### **Manual Testing Checklist**
```go
// Test dimensions
// 1. Resize terminal to various sizes
// 2. Test minimum dimensions (80x24)
// 3. Test very narrow terminals (< 80 cols)
// 4. Test very short terminals (< 24 rows)

func (m *Model) testDimensions() {
    testSizes := []struct{ width, height int }{
        {80, 24},   // Standard
        {40, 12},   // Small
        {120, 40},  // Large
        {20, 10},   // Tiny
    }

    for _, size := range testSizes {
        m.styles.SetSize(size.width, size.height)
        view := m.View()

        logger.LogWithData("Dimension Test", map[string]interface{}{
            "test_size":   fmt.Sprintf("%dx%d", size.width, size.height),
            "view_length": len(view),
            "has_ansi":    strings.Contains(view, "\x1b["),
            "line_count":  strings.Count(view, "\n"),
        })
    }
}
```

### **Navigation Testing**
```go
func testNavigationFlow() {
    // Test complete navigation flow
    testSteps := []struct {
        action   string
        expected string
    }{
        {"start", "welcome"},
        {"continue", "main_menu"},
        {"select_providers", "llm_providers"},
        {"select_openai", "llm_provider_form§openai"},
        {"go_back", "llm_providers"},
        {"esc", "welcome"},
    }

    for _, step := range testSteps {
        logger.LogWithData("Navigation Test", map[string]interface{}{
            "action":   step.action,
            "expected": step.expected,
            "actual":   string(navigator.Current()),
        })
    }
}
```

This debugging guide provides comprehensive tools for:
- **Safe Development**: TUI-compatible logging without rendering corruption
- **State Inspection**: Real-time monitoring of component state
- **Performance Analysis**: Memory and viewport performance tracking
- **Error Recovery**: Graceful degradation and state recovery patterns
- **Testing Strategies**: Systematic approaches to manual testing