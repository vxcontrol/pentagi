# PentAGI Installer Troubleshooting Guide

> Comprehensive troubleshooting guide including recent fixes, performance optimization, and common issues.

## 🚨 **Development-Specific Issues**

### **TUI Application Constraints**
**Problem**: Running installer breaks terminal session during development
**Solution**: Build-only development workflow

```bash
# ✅ CORRECT: Build and test separately
cd backend/
go build -o ../build/installer ./cmd/installer/main.go

# Test in separate terminal session
cd ../build/
./installer

# ❌ WRONG: Running during development
cd backend/
go run ./cmd/installer/main.go  # Breaks active terminal!
```

**Debug Monitoring**:
```bash
# Monitor debug output during development
tail -f log.json | jq '.'

# Filter by component
tail -f log.json | jq 'select(.component == "FormModel")'

# Pretty print timestamps
tail -f log.json | jq -r '"\(.timestamp) [\(.level)] \(.message)"'
```

## 🔧 **Recent Fixes & Improvements**

### ✅ **Composite ScreenID Navigation System**
**Problem**: Need to preserve selected menu items and provider selections across navigation
**Solution**: Implemented composite ScreenIDs with `§` separator for parameter passing

**Before** (❌ Problematic):
```go
// Lost selection on navigation
func (m *MenuModel) handleSelection() (tea.Model, tea.Cmd) {
    return NavigationMsg{Target: LLMProvidersScreen} // No context preserved
}
```

**After** (✅ Fixed):
```go
// Preserves selection context
func (m *MenuModel) handleSelection() (tea.Model, tea.Cmd) {
    selectedItem := m.getSelectedItem()
    return NavigationMsg{
        Target: CreateScreenID("llm_providers", selectedItem.ID),
    }
}

// Results in: "llm_providers§openai" - selection preserved
```

**Benefits**:
- Type-safe parameter passing via `GetScreen()`, `GetArgs()`, `CreateScreenID()`
- Automatic state restoration - user returns to exact selection after ESC
- Clean navigation stack with full context preservation

### ✅ **Complete Localization Architecture**
**Problem**: Hardcoded strings scattered throughout UI components
**Solution**: Centralized all user-visible text in `locale.go` with structured constants

**Implementation**:
```go
// Multi-line text stored as single constants
const MainMenuLLMProvidersInfo = `Configure AI language model providers for PentAGI.

Supported providers:
• OpenAI (GPT-4, GPT-3.5-turbo)
• Anthropic (Claude-3, Claude-2)
...`

// Usage in components
sections = append(sections, m.styles.Paragraph.Render(locale.MainMenuLLMProvidersInfo))
```

**Coverage**: 100% of user-facing text moved to locale constants
- Menu descriptions and help text
- Form labels and error messages
- Provider-specific documentation
- Keyboard shortcuts and hints

### ✅ **Viewport-Based Form Scrolling**
**Problem**: Forms with many fields don't fit on smaller terminals
**Solution**: Implemented auto-scrolling viewport with focus tracking

**Key Features**:
- **Auto-scroll**: Focused field automatically stays visible
- **Smart positioning**: Calculates field heights for precise scroll positioning
- **Seamless navigation**: Tab/Shift+Tab scroll form as needed
- **No extra hotkeys**: Uses existing navigation keys

**Technical Implementation**:
```go
// Auto-scroll on field focus change
func (m *FormModel) ensureFocusVisible() {
    focusY := m.calculateFieldPosition(m.focusedIndex)
    if focusY < m.viewport.YOffset {
        m.viewport.YOffset = focusY  // Scroll up
    }
    if focusY >= m.viewport.YOffset + m.viewport.Height {
        m.viewport.YOffset = focusY - m.viewport.Height + 1  // Scroll down
    }
}
```

### ✅ **Enhanced Provider Configuration**
**Problem**: Missing configuration fields for several LLM providers
**Solution**: Added complete field sets for all supported providers

**Provider Field Mapping**:
- **OpenAI/Anthropic/Gemini**: Base URL + API Key
- **AWS Bedrock**: Region + Access Key + Secret Key + Base URL (optional)
- **Ollama**: Base URL + Config Path
- **Custom**: Base URL + API Key + Model + Config Path + Legacy Reasoning (boolean)

**Dynamic Form Generation**: Forms adapt based on provider type with appropriate validation and help text.

## 🔧 **Common Issues & Solutions**

### **Navigation Issues**

#### **Navigation Stack Corruption**
**Symptoms**: User gets stuck on screens, ESC doesn't work, back navigation fails
**Cause**: Circular navigation patterns or corrupted navigation stack

**Debug**:
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
```

**Solution**:
```go
// ✅ CORRECT: Use GoBack to prevent loops
func (m *FormModel) saveAndReturn() (tea.Model, tea.Cmd) {
    if err := m.saveConfiguration(); err != nil {
        return m, nil  // Stay on form if save fails
    }
    return m, func() tea.Msg {
        return NavigationMsg{GoBack: true}  // Return to previous screen
    }
}

// ❌ WRONG: Direct navigation creates loops
func (m *FormModel) saveAndReturn() (tea.Model, tea.Cmd) {
    m.saveConfiguration()
    return m, func() tea.Msg {
        return NavigationMsg{Target: ProvidersScreen} // Creates navigation loop!
    }
}
```

#### **Lost Selection State**
**Symptoms**: Menu selections reset, provider choices forgotten, configuration lost
**Cause**: Models not constructed with proper args

**Solution**:
```go
// ✅ CORRECT: Args-based construction
func NewModel(controller *StateController, styles *Styles,
              window *Window, args []string) *Model {
    selectedIndex := 0
    if len(args) > 0 && args[0] != "" {
        // Restore selection from navigation args
        for i, item := range items {
            if item.ID == args[0] {
                selectedIndex = i
                break
            }
        }
    }
    return &Model{selectedIndex: selectedIndex, args: args}
}
```

### **Form Issues**

#### **Form Field Width Problems**
**Symptoms**: Input fields too narrow/wide, don't adapt to terminal size
**Cause**: Fixed width assignments during field creation

**Debug**:
```go
func (m *FormModel) debugFormDimensions() {
    width, height := m.styles.GetSize()
    viewportWidth, viewportHeight := m.getViewportSize()
    inputWidth := m.getInputWidth()

    logger.LogWithData("Form Dimensions", map[string]interface{}{
        "terminal_size":  fmt.Sprintf("%dx%d", width, height),
        "viewport_size":  fmt.Sprintf("%dx%d", viewportWidth, viewportHeight),
        "input_width":    inputWidth,
        "is_vertical":    m.isVerticalLayout(),
        "field_count":    len(m.fields),
    })
}
```

**Solution**:
```go
// ✅ CORRECT: Dynamic width calculation
func (m *FormModel) updateFormContent() {
    inputWidth := m.getInputWidth()

    for i, field := range m.fields {
        // Apply width during rendering, not initialization
        field.Input.Width = inputWidth - 3
        field.Input.SetValue(field.Input.Value()) // Trigger width update
    }
}

// ❌ WRONG: Fixed width at creation
func (m *FormModel) addField() {
    input := textinput.New()
    input.Width = 50 // Breaks responsive design!
}
```

#### **Form Scrolling Issues**
**Symptoms**: Can't reach all fields, focused field goes off-screen
**Cause**: Missing auto-scroll implementation or incorrect field height calculation

**Debug**:
```go
func (m *FormModel) debugScrollState() {
    logger.LogWithData("Scroll State", map[string]interface{}{
        "focused_index":    m.focusedIndex,
        "viewport_offset":  m.viewport.YOffset,
        "viewport_height":  m.viewport.Height,
        "content_height":   lipgloss.Height(m.formContent),
        "field_heights":    m.fieldHeights,
        "total_fields":     len(m.fields),
    })
}
```

**Solution**:
```go
// ✅ CORRECT: Auto-scroll implementation
func (m *FormModel) focusNext() {
    m.fields[m.focusedIndex].Input.Blur()
    m.focusedIndex = (m.focusedIndex + 1) % len(m.fields)
    m.fields[m.focusedIndex].Input.Focus()
    m.updateFormContent()
    m.ensureFocusVisible() // Critical for auto-scroll
}
```

### **Environment Variable Issues**

#### **Configuration Not Persisting**
**Symptoms**: Settings lost between sessions, environment variables not saved
**Cause**: Not calling controller save methods or incorrect cleanup logic

**Debug**:
```go
func (m *FormModel) debugEnvVarState() {
    for _, field := range m.fields {
        envVar, _ := m.controller.GetVar(m.getEnvVarName(getEnvSuffixFromKey(field.Key)))

        logger.LogWithData("Field State", map[string]interface{}{
            "field_key":       field.Key,
            "input_value":     field.Input.Value(),
            "env_var_name":    m.getEnvVarName(getEnvSuffixFromKey(field.Key)),
            "env_var_value":   envVar.Value,
            "env_var_default": envVar.Default,
            "is_present":      envVar.IsPresent(),
            "initially_set":   m.initiallySetFields[field.Key],
        })
    }
}
```

**Solution**:
```go
// ✅ CORRECT: Proper save implementation
func (m *FormModel) saveConfiguration() error {
    // First pass: Remove cleared fields
    for _, field := range m.fields {
        value := strings.TrimSpace(field.Input.Value())

        if value == "" && m.initiallySetFields[field.Key] {
            // Field was set but now empty - remove from environment
            if err := m.controller.SetVar(field.EnvVarName, ""); err != nil {
                return fmt.Errorf("failed to clear %s: %w", field.EnvVarName, err)
            }
            logger.Log("[FormModel] SAVE: cleared %s", field.EnvVarName)
        }
    }

    // Second pass: Save non-empty values
    for _, field := range m.fields {
        value := strings.TrimSpace(field.Input.Value())
        if value != "" {
            if err := m.controller.SetVar(field.EnvVarName, value); err != nil {
                return fmt.Errorf("failed to set %s: %w", field.EnvVarName, err)
            }
            logger.Log("[FormModel] SAVE: set %s=%s", field.EnvVarName, value)
        }
    }

    return nil
}
```

### **Layout Issues**

#### **Content Not Adapting to Terminal Size**
**Symptoms**: Content cut off, panels don't resize, horizontal scrolling
**Cause**: Missing responsive layout logic or incorrect dimension handling

**Debug**:
```go
func (m *Model) debugLayoutState() {
    width, height := m.styles.GetSize()
    contentWidth, contentHeight := m.window.GetContentSize()

    logger.LogWithData("Layout State", map[string]interface{}{
        "terminal_size":    fmt.Sprintf("%dx%d", width, height),
        "content_size":     fmt.Sprintf("%dx%d", contentWidth, contentHeight),
        "is_vertical":      m.isVerticalLayout(),
        "min_terminal":     MinTerminalWidth,
        "min_menu_width":   MinMenuWidth,
        "min_info_width":   MinInfoWidth,
    })
}
```

**Solution**:
```go
// ✅ CORRECT: Responsive layout implementation
func (m *Model) View() string {
    width, height := m.styles.GetSize()

    leftPanel := m.renderContent()
    rightPanel := m.renderInfo()

    if m.isVerticalLayout() {
        return m.renderVerticalLayout(leftPanel, rightPanel, width, height)
    }
    return m.renderHorizontalLayout(leftPanel, rightPanel, width, height)
}

func (m *Model) isVerticalLayout() bool {
    contentWidth := m.window.GetContentWidth()
    return contentWidth < (MinMenuWidth + MinInfoWidth + PaddingWidth)
}
```

#### **Footer Height Inconsistency**
**Symptoms**: Footer takes more/less space than expected, layout calculations wrong
**Cause**: Using border-based footer approach instead of background approach

**Solution**:
```go
// ✅ CORRECT: Background approach (always 1 line)
func (a *App) renderFooter() string {
    actions := a.buildFooterActions()
    footerText := strings.Join(actions, " • ")

    return a.styles.Footer.Render(footerText)
}

// In styles.go
func (s *Styles) updateStyles() {
    s.Footer = lipgloss.NewStyle().
        Width(s.width).
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("255")).
        Padding(0, 1, 0, 1)
}

// ❌ WRONG: Border approach (height varies)
footer := lipgloss.NewStyle().
    Height(1).
    Border(lipgloss.Border{Top: true}).
    Render(text)
```

## 🔧 **Performance Issues**

### **Slow Rendering**
**Symptoms**: Laggy UI, delayed responses to keystrokes
**Cause**: Multiple glamour renderers, excessive content updates

**Debug**:
```go
func (m *Model) debugRenderPerformance() {
    start := time.Now()
    content := m.buildContent()
    buildDuration := time.Since(start)

    start = time.Now()
    m.viewport.SetContent(content)
    setContentDuration := time.Since(start)

    start = time.Now()
    view := m.viewport.View()
    viewDuration := time.Since(start)

    logger.LogWithData("Render Performance", map[string]interface{}{
        "content_size":       len(content),
        "rendered_size":      len(view),
        "build_ms":           buildDuration.Milliseconds(),
        "set_content_ms":     setContentDuration.Milliseconds(),
        "view_render_ms":     viewDuration.Milliseconds(),
    })
}
```

**Solution**:
```go
// ✅ CORRECT: Single shared renderer
// In styles.go
func New() *Styles {
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    return &Styles{renderer: renderer}
}

// Usage
rendered, err := m.styles.GetRenderer().Render(content)

// ❌ WRONG: Multiple renderers
func (m *Model) renderMarkdown(content string) string {
    renderer, _ := glamour.NewTermRenderer(...) // Performance killer!
    return renderer.Render(content)
}
```

### **Memory Leaks**
**Symptoms**: Increasing memory usage, application becomes sluggish over time
**Cause**: Not properly cleaning up resources, creating multiple renderer instances

**Solution**:
```go
// ✅ CORRECT: Complete state reset
func (m *Model) Init() tea.Cmd {
    // Reset ALL state completely
    m.content = ""
    m.ready = false
    m.error = nil
    m.initialized = false

    // Reset component state
    m.viewport.GotoTop()
    m.viewport.SetContent("")

    // Reset form state
    m.focusedIndex = 0
    m.hasChanges = false
    for i := range m.fields {
        m.fields[i].Input.Blur()
    }

    return m.loadContent
}
```

## 🔧 **Error Recovery Patterns**

### **Graceful State Recovery**
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

### **Safe Async Operations**
```go
func (m *Model) loadContent() tea.Cmd {
    return func() tea.Msg {
        defer func() {
            if r := recover(); r != nil {
                logger.Errorf("[%s] PANIC: recovered from panic: %v", m.componentName, r)
                return ErrorMsg{fmt.Errorf("panic in loadContent: %v", r)}
            }
        }()

        content, err := m.loadFromSource()
        if err != nil {
            return ErrorMsg{err}
        }

        return ContentLoadedMsg{content}
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

func testDimensions() {
    testSizes := []struct{ width, height int }{
        {80, 24},   // Standard
        {40, 12},   // Small
        {120, 40},  // Large
        {20, 10},   // Tiny
    }

    for _, size := range testSizes {
        logger.LogWithData("Dimension Test", map[string]interface{}{
            "test_size":   fmt.Sprintf("%dx%d", size.width, size.height),
            "layout_mode": getLayoutMode(size.width, size.height),
        })
    }
}
```

### **Navigation Flow Testing**
```go
func testNavigationFlow() {
    testSteps := []struct {
        action   string
        expected string
    }{
        {"start", "welcome"},
        {"continue", "main_menu"},
        {"select_providers", "llm_providers"},
        {"select_openai", "llm_provider_form§openai"},
        {"go_back", "llm_providers§openai"},
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

This troubleshooting guide provides comprehensive solutions for:
- **Development Workflow**: TUI-safe development patterns
- **Navigation Issues**: Stack management and state preservation
- **Form Problems**: Responsive design and scrolling
- **Configuration**: Environment variable management
- **Performance**: Optimization and resource management
- **Recovery**: Graceful error handling and state restoration