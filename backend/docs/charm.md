# Charm.sh Ecosystem - Personal Cheat Sheet

> Personal reference for building TUI applications with the Charm stack.

## 📦 **Core Libraries Overview**

### Core Packages
- **`bubbletea`**: Event-driven TUI framework (MVU pattern)
- **`lipgloss`**: Styling and layout engine
- **`bubbles`**: Pre-built components (viewport, textinput, etc.)
- **`huh`**: Advanced form builder
- **`glamour`**: Markdown renderer

## 🫧 **BubbleTea (MVU Pattern)**

### Model-View-Update Lifecycle
```go
// Model holds all state
type Model struct {
    content string
    ready   bool
}

// Update handles events and returns new state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Handle resize
        return m, nil
    case tea.KeyMsg:
        // Handle keyboard input
        return m, nil
    }
    return m, nil
}

// View renders current state
func (m Model) View() string {
    return "content"
}

// Init returns initial command
func (m Model) Init() tea.Cmd {
    return nil
}
```

### Commands and Messages
```go
// Commands return future messages
func loadDataCmd() tea.Msg {
    return DataLoadedMsg{data: "loaded"}
}

// Async operations
return m, tea.Cmd(func() tea.Msg {
    time.Sleep(time.Second)
    return TimerMsg{}
})
```

### Critical Patterns
```go
// Model interface implementation
type Model struct {
    styles *styles.Styles  // ALWAYS use shared styles
}

func (m Model) Init() tea.Cmd {
    // ALWAYS reset state completely
    m.content = ""
    m.ready = false
    return m.loadContent
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // NEVER store dimensions in model - use styles.SetSize()
        // Model gets dimensions via m.styles.GetSize()
    case tea.KeyMsg:
        switch msg.String() {
        case "enter": return m, navigateCmd
        }
    }
}
```

## 🎨 **Lipgloss (Styling & Layout)**

**Purpose**: CSS-like styling for terminal interfaces
**Key Insight**: Height() vs MaxHeight() behavior difference!

### Critical Height Control
```go
// ❌ WRONG: Height() sets MINIMUM height (can expand!)
style := lipgloss.NewStyle().Height(1).Border(lipgloss.NormalBorder())

// ✅ CORRECT: MaxHeight() + Inline() for EXACT height
style := lipgloss.NewStyle().MaxHeight(1).Inline(true)

// ✅ PRODUCTION: Background approach for consistent 1-line footers
footer := lipgloss.NewStyle().
    Width(width).
    Background(borderColor).
    Foreground(textColor).
    Padding(0, 1, 0, 1).  // Only horizontal padding
    Render(text)

// FOOTER APPROACH - PRODUCTION READY (✅ PROVEN SOLUTION)
// ❌ WRONG: Border approach (inconsistent height)
style.BorderTop(true).Height(1)

// ✅ CORRECT: Background approach (always 1 line)
style.Background(color).Foreground(textColor).Padding(0,1,0,1)
```

### Layout Patterns
```go
// LAYOUT COMPOSITION
lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
lipgloss.JoinHorizontal(lipgloss.Top, left, right)
lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, content)

// Horizontal layout
left := lipgloss.NewStyle().Width(leftWidth).Render(leftContent)
right := lipgloss.NewStyle().Width(rightWidth).Render(rightContent)
combined := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

// Vertical layout with consistent spacing
sections := []string{header, content, footer}
combined := lipgloss.JoinVertical(lipgloss.Left, sections...)

// Centering content
centered := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)

// Responsive design
verticalStyle := lipgloss.NewStyle().Width(width).Padding(0, 2, 0, 2)
if width < 80 {
    // Vertical layout for narrow screens
}
```

### Responsive Patterns
```go
// Breakpoint-based layout
width, height := m.styles.GetSize()  // ALWAYS from styles
if width < 80 {
    return lipgloss.JoinVertical(lipgloss.Left, panels...)
} else {
    return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

// Dynamic width allocation
leftWidth := width / 3
rightWidth := width - leftWidth - 4
```

## 📺 **Bubbles (Interactive Components)**

**Purpose**: Pre-built interactive components
**Key Components**: viewport, textinput, list, table

### Viewport - Critical for Scrolling
```go
import "github.com/charmbracelet/bubbles/viewport"

// Setup
viewport := viewport.New(width, height)
viewport.Style = lipgloss.NewStyle() // Clean style prevents conflicts

// Modern scroll methods (use these!)
viewport.ScrollUp(1)     // Replaces LineUp()
viewport.ScrollDown(1)   // Replaces LineDown()
viewport.ScrollLeft(2)   // Horizontal, 2 steps for forms
viewport.ScrollRight(2)

// Deprecated (avoid)
vp.LineUp(lines)       // ❌ Deprecated
vp.LineDown(lines)     // ❌ Deprecated

// Status tracking
viewport.ScrollPercent() // 0.0 to 1.0
viewport.AtBottom()      // bool
viewport.AtTop()         // bool

// State checking
isScrollable := !(vp.AtTop() && vp.AtBottom())
progress := vp.ScrollPercent()

// Content management
viewport.SetContent(content)
viewport.View() // Renders visible portion

// Update in message handling
var cmd tea.Cmd
m.viewport, cmd = m.viewport.Update(msg)
```

### TextInput
```go
import "github.com/charmbracelet/bubbles/textinput"

ti := textinput.New()
ti.Placeholder = "Enter text..."
ti.Focus()
ti.EchoMode = textinput.EchoPassword  // For masked input
ti.CharLimit = 100
```

## 📝 **Huh (Forms)**

**Purpose**: Advanced form builder for complex user input
```go
import "github.com/charmbracelet/huh"

form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Key("api_key").
            Title("API Key").
            Password().  // Masked input
            Validate(func(s string) error {
                if len(s) < 10 {
                    return errors.New("API key too short")
                }
                return nil
            }),

        huh.NewSelect[string]().
            Key("provider").
            Title("Provider").
            Options(
                huh.NewOption("OpenAI", "openai"),
                huh.NewOption("Anthropic", "anthropic"),
            ),
    ),
)

// Integration with bubbletea
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.form, cmd = m.form.Update(msg)

    if m.form.State == huh.StateCompleted {
        // Form submitted - access values
        apiKey := m.form.GetString("api_key")
        provider := m.form.GetString("provider")
    }

    return m, cmd
}
```

## ✨ **Glamour (Markdown Rendering)**

**Purpose**: Beautiful markdown rendering in terminal
**CRITICAL**: Create renderer ONCE in styles.New(), reuse everywhere

```go
// ✅ CORRECT: Single renderer instance (prevents freezing)
// styles.go
type Styles struct {
    renderer *glamour.TermRenderer
}

func New() *Styles {
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    return &Styles{renderer: renderer}
}

func (s *Styles) GetRenderer() *glamour.TermRenderer {
    return s.renderer
}

// Usage in models
rendered, err := m.styles.GetRenderer().Render(markdown)

// ❌ WRONG: Creating new renderer each time (can freeze!)
renderer, _ := glamour.NewTermRenderer(...)
```

### Safe Rendering with Fallback
```go
// Safe rendering with fallback
rendered, err := renderer.Render(content)
if err != nil {
    // Fallback to plain text
    rendered = fmt.Sprintf("# Content\n\n%s\n\n*Render error: %v*", content, err)
}
```

## 🏗️ **Production Architecture Patterns**

### 1. Centralized Styles & Dimensions

**CRITICAL**: Never store width/height in models - use styles singleton

```go
// ✅ CORRECT: Centralized in styles
type Styles struct {
    width    int
    height   int
    renderer *glamour.TermRenderer
    // ... all styles
}

func (s *Styles) SetSize(width, height int) {
    s.width = width
    s.height = height
    s.updateStyles()  // Recalculate responsive styles
}

func (s *Styles) GetSize() (int, int) {
    return s.width, s.height
}

// Models use styles for dimensions
func (m *Model) updateViewport() {
    width, height := m.styles.GetSize()
    if width <= 0 || height <= 0 {
        return  // Graceful handling
    }
    // ... viewport setup
}
```

### 2. TUI-Safe Logging System

**Problem**: fmt.Printf breaks TUI rendering
**Solution**: File-based logger

```go
// logger.Log() writes to log.json
logger.Log("[Component] ACTION: details %v", value)
logger.Errorf("[Component] ERROR: %v", err)

// Development monitoring (separate terminal)
tail -f log.json

// ❌ WRONG: Console output in TUI
fmt.Printf("Debug: %v\n", value)  // Breaks rendering
```

### 3. Unified Header/Footer Management
```go
// app.go - Central layout control
func (a *App) View() string {
    header := a.renderHeader()
    footer := a.renderFooter()
    content := a.currentModel.View()

    contentHeight := max(height - headerHeight - footerHeight, 0)
    contentArea := a.styles.Content.Height(contentHeight).Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}

func (a *App) renderHeader() string {
    switch a.navigator.Current() {
    case WelcomeScreen:
        return a.styles.RenderASCIILogo()
    default:
        return a.styles.Header.Render(title)
    }
}

func (a *App) renderFooter() string {
    actions := []string{}
    // Dynamic actions based on screen state
    if canContinue {
        actions = append(actions, "Enter: Continue")
    }
    if hasScrollableContent {
        actions = append(actions, "↑↓: Scroll")
    }

    return lipgloss.NewStyle().
        Width(width).
        Background(borderColor).
        Foreground(textColor).
        Padding(0, 1, 0, 1).
        Render(strings.Join(actions, " • "))
}

// locale.go - Helper functions
func BuildCommonActions() []string {
    return []string{NavBack, NavExit}
}

func BuildEULAActions(atEnd bool) []string {
    if !atEnd {
        return []string{EULANavScrollInstructions}
    }
    return []string{EULANavAcceptReject}
}

// Usage
actions := locale.BuildCommonActions()
actions = append(actions, specificActions...)
```

### 4. Type-Safe Navigation with Composite ScreenIDs

**Critical Pattern**: Use typed screen IDs with argument support
```go
type ScreenID string
const (
    WelcomeScreen ScreenID = "welcome"
    EULAScreen   ScreenID = "eula"
    MainMenuScreen ScreenID = "main_menu"
    LLMProviderFormScreen ScreenID = "llm_provider_form"
)

// ScreenID methods for composite support
func (s ScreenID) GetScreen() string {
    parts := strings.Split(string(s), "§")
    return parts[0]
}

func (s ScreenID) GetArgs() []string {
    parts := strings.Split(string(s), "§")
    if len(parts) <= 1 {
        return []string{}
    }
    return parts[1:]
}

func CreateScreenID(screen string, args ...string) ScreenID {
    if len(args) == 0 {
        return ScreenID(screen)
    }
    parts := append([]string{screen}, args...)
    return ScreenID(strings.Join(parts, "§"))
}

type NavigationMsg struct {
    Target ScreenID  // Can be simple or composite!
    GoBack bool
}

// Usage - Simple screen
return m, func() tea.Msg {
    return NavigationMsg{Target: EULAScreen}
}

// Usage - Composite screen with arguments
return m, func() tea.Msg {
    return NavigationMsg{Target: CreateScreenID("llm_provider_form", "openai")}
}
```

### 5. Model State Management

**Pattern**: Complete reset on Init() for predictable behavior
```go
func (m *Model) Init() tea.Cmd {
    logger.Log("[Model] INIT")
    // ALWAYS reset ALL state
    m.content = ""
    m.ready = false
    m.scrolled = false
    m.scrolledToEnd = false
    m.error = nil
    return m.loadContent
}

// Force re-render after async operations
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ContentLoadedMsg:
        m.content = msg.Content
        m.ready = true
        // Force view update
        return m, func() tea.Msg { return nil }
    }
    return m, nil
}
```

## 🐛 **Key Debugging Techniques**

### 1. TUI-Safe Debug Output
```go
// ❌ NEVER: Breaks TUI rendering
fmt.Println("debug")
log.Println("debug")

// ✅ ALWAYS: File-based logging
logger.Log("[Component] Event: %v", msg)
logger.Log("[Model] UPDATE: key=%s", msg.String())
logger.Log("[Model] VIEWPORT: %dx%d ready=%v", width, height, m.ready)

// Monitor in separate terminal
tail -f log.json
```

### 2. Dimension Handling
```go
func (m Model) View() string {
    width, height := m.styles.GetSize()
    if width <= 0 || height <= 0 {
        return "Loading..." // Graceful fallback
    }
    // ... normal rendering
}

// Log dimension changes
logger.Log("[Model] RESIZE: %dx%d", width, height)
```

### 3. Content Loading Debug
```go
func (m *Model) loadContent() tea.Msg {
    logger.Log("[Model] LOAD: start")
    content, err := source.GetContent()
    if err != nil {
        logger.Errorf("[Model] LOAD: error: %v", err)
        return ErrorMsg{err}
    }
    logger.Log("[Model] LOAD: success (%d chars)", len(content))
    return ContentLoadedMsg{content}
}
```

## 🎯 **Advanced Navigation with Composite ScreenIDs**

### Composite ScreenID Pattern
**Problem**: Need to pass parameters to screens (e.g., which provider to configure)
**Solution**: Composite ScreenIDs with `§` separator

```go
// Format: "screen§arg1§arg2§..."
type ScreenID string

// Methods for parsing composite IDs
func (s ScreenID) GetScreen() string {
    parts := strings.Split(string(s), "§")
    return parts[0]
}

func (s ScreenID) GetArgs() []string {
    parts := strings.Split(string(s), "§")
    if len(parts) <= 1 {
        return []string{}
    }
    return parts[1:]
}

// Helper for creating composite IDs
func CreateScreenID(screen string, args ...string) ScreenID {
    if len(args) == 0 {
        return ScreenID(screen)
    }
    parts := append([]string{screen}, args...)
    return ScreenID(strings.Join(parts, "§"))
}
```

### Usage Examples
```go
// Simple screen (no arguments)
welcome := WelcomeScreen  // "welcome"

// Composite screen (with arguments)
providerForm := CreateScreenID("llm_provider_form", "openai")  // "llm_provider_form§openai"

// Navigation with arguments
return m, func() tea.Msg {
    return NavigationMsg{
        Target: CreateScreenID("llm_provider_form", "anthropic"),
        Data:   FormData{ProviderID: "anthropic"},
    }
}

// In createModelForScreen - extract arguments
func (a *App) createModelForScreen(screenID ScreenID, data any) tea.Model {
    baseScreen := screenID.GetScreen()
    args := screenID.GetArgs()

    switch ScreenID(baseScreen) {
    case LLMProviderFormScreen:
        providerID := "openai" // default
        if len(args) > 0 {
            providerID = args[0]
        }
        return NewLLMProviderFormModel(providerID, ...)
    }
}
```

### State Persistence
```go
// Stack automatically preserves composite IDs
navigator.Push(CreateScreenID("llm_provider_form", "gemini"))

// State contains: ["welcome", "main_menu", "llm_providers", "llm_provider_form§gemini"]
// On restore: user returns to Gemini provider form, not default OpenAI
```

## 🎯 **Advanced Form Scrolling with Viewport**

### Auto-Scrolling Forms Pattern
**Problem**: Forms with many fields don't fit on smaller terminals, focused fields go off-screen
**Solution**: Viewport component with automatic scroll-to-focus behavior

```go
import "github.com/charmbracelet/bubbles/viewport"

type FormModel struct {
    fields       []FormField
    focusedIndex int
    viewport     viewport.Model
    formContent  string
    fieldHeights []int // Heights of each field for scroll calculation
}

// Initialize viewport
func New() *FormModel {
    return &FormModel{
        viewport: viewport.New(0, 0),
    }
}

// Update viewport dimensions on resize
func (m *FormModel) updateViewport() {
    contentWidth, contentHeight := m.getContentSize()
    m.viewport.Width = contentWidth - 4  // padding
    m.viewport.Height = contentHeight - 2 // header/footer space
    m.viewport.SetContent(m.formContent)
}

// Render form content and track field positions
func (m *FormModel) updateFormContent() {
    var sections []string
    m.fieldHeights = []int{}

    for i, field := range m.fields {
        fieldHeight := 4 // title + description + input + spacing
        m.fieldHeights = append(m.fieldHeights, fieldHeight)

        sections = append(sections, field.Title)
        sections = append(sections, field.Description)
        sections = append(sections, field.Input.View())
        sections = append(sections, "") // spacing
    }

    m.formContent = strings.Join(sections, "\n")
    m.viewport.SetContent(m.formContent)
}

// Auto-scroll to focused field
func (m *FormModel) ensureFocusVisible() {
    if m.focusedIndex >= len(m.fieldHeights) {
        return
    }

    // Calculate Y position of focused field
    focusY := 0
    for i := 0; i < m.focusedIndex; i++ {
        focusY += m.fieldHeights[i]
    }

    visibleRows := m.viewport.Height
    offset := m.viewport.YOffset

    // Scroll up if field is above visible area
    if focusY < offset {
        m.viewport.YOffset = focusY
    }

    // Scroll down if field is below visible area
    if focusY+m.fieldHeights[m.focusedIndex] >= offset+visibleRows {
        m.viewport.YOffset = focusY + m.fieldHeights[m.focusedIndex] - visibleRows + 1
    }
}

// Navigation with auto-scroll
func (m *FormModel) focusNext() {
    m.fields[m.focusedIndex].Input.Blur()
    m.focusedIndex = (m.focusedIndex + 1) % len(m.fields)
    m.fields[m.focusedIndex].Input.Focus()
    m.updateFormContent()
    m.ensureFocusVisible() // Key addition!
}

// Render scrollable form
func (m *FormModel) View() string {
    return m.viewport.View() // Viewport handles clipping and scrolling
}
```

### Key Benefits of Viewport Forms
- **Automatic Clipping**: Viewport handles content that exceeds available space
- **Smooth Scrolling**: Fields slide into view without jarring jumps
- **Focus Preservation**: Focused field always remains visible
- **No Extra Hotkeys**: Uses standard navigation (Tab, arrows)
- **Terminal Friendly**: Works on any terminal size

### Critical Implementation Details
1. **Field Height Tracking**: Must calculate actual rendered height of each field
2. **Scroll Timing**: Call `ensureFocusVisible()` after every focus change
3. **Content Updates**: Re-render form content when input values change
4. **Viewport Sizing**: Account for padding, headers, footers in size calculation

This pattern is essential for professional TUI applications with complex forms.

## ⚠️ **Common Pitfalls & Solutions**

### 1. Glamour Renderer Freezing
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
    return m.styles.GetRenderer().Render(content)
}
```

### 2. Footer Height Inconsistency
**Problem**: Border-based footers vary in height
**Solution**: Background approach with padding

```go
// ❌ WRONG: Border approach (height varies)
footer := lipgloss.NewStyle().
    Height(1).
    Border(lipgloss.Border{Top: true}).
    Render(text)

// ✅ CORRECT: Background approach (exactly 1 line)
footer := lipgloss.NewStyle().
    Background(borderColor).
    Foreground(textColor).
    Padding(0, 1, 0, 1).
    Render(text)
```

### 3. Dimension Synchronization
**Problem**: Models store their own width/height, get out of sync
**Solution**: Centralize dimensions in styles singleton

```go
// ❌ WRONG: Models managing their own dimensions
type Model struct {
    width, height int
}

// ✅ CORRECT: Centralized dimension management
type Model struct {
    styles *styles.Styles  // Access via styles.GetSize()
}
```

### 4. TUI Rendering Corruption
**Problem**: Console output breaks rendering
**Solution**: File-based logger, never fmt.Printf

```go
// ❌ NEVER: Use tea.ClearScreen during navigation
return a, tea.Batch(cmd, tea.ClearScreen)

// ✅ CORRECT: Let model Init() handle clean state
return a, a.currentModel.Init()
```

### 5. Navigation State Issues
**Problem**: Models retain state between visits
**Solution**: Complete state reset in Init()

```go
// ❌ WRONG: String-based navigation (typo-prone)
return NavigationMsg{Target: "main_menu"}

// ❌ WRONG: Manual string concatenation for arguments
return NavigationMsg{Target: ScreenID("llm_provider_form/openai")}

// ✅ CORRECT: Type-safe constants
return NavigationMsg{Target: MainMenuScreen}

// ✅ CORRECT: Composite ScreenID with helper
return NavigationMsg{Target: CreateScreenID("llm_provider_form", "openai")}
```

## 🚀 **Performance & Best Practices**

### Proven Patterns
```go
// ✅ DO: Shared renderer
rendered, _ := m.styles.GetRenderer().Render(content)

// ✅ DO: Centralized dimensions
width, height := m.styles.GetSize()

// ✅ DO: File logging
logger.Log("[Component] ACTION: %v", data)

// ✅ DO: Complete state reset
func (m *Model) Init() tea.Cmd {
    m.resetAllState()
    return m.loadContent
}

// ✅ DO: Graceful dimension handling
if width <= 0 || height <= 0 {
    return "Loading..."
}
```

### Anti-Patterns to Avoid
```go
// ❌ DON'T: New renderer instances
renderer, _ := glamour.NewTermRenderer(...)

// ❌ DON'T: Model dimensions
type Model struct {
    width  int  // Store in styles instead
    height int  // Store in styles instead
}

// ❌ DON'T: Console output
fmt.Printf("Debug: %v\n", value)

// ❌ DON'T: Partial state reset
func (m *Model) Init() tea.Cmd {
    // Only resetting some fields - incomplete!
    m.content = ""
    // Missing: m.ready, m.scrolled, etc.
}
```

### Key Best Practices Summary
- **Single glamour renderer**: Prevents freezing, faster rendering
- **Centralized dimensions**: Eliminates sync issues, simplifies models
- **Background footer**: Consistent height, modern appearance
- **Type-safe navigation**: Compile-time error prevention
- **File-based logging**: Debug without breaking TUI
- **Complete state reset**: Predictable model behavior
- **Graceful fallbacks**: Handle edge cases elegantly
- **Resource estimation**: Real-time calculation of token/memory usage
- **Environment integration**: Proper EnvVar handling with cleanup
- **Value formatting**: Consistent human-readable displays (formatBytes, formatNumber)

---

*This cheat sheet contains battle-tested solutions for TUI development in the Charm ecosystem, proven in production use.*

## 🎯 **Advanced Form Field Patterns**

### **Boolean Fields with Tab Completion**
**Innovation**: Auto-completion for boolean values with suggestions

```go
import "github.com/charmbracelet/bubbles/textinput"

func createBooleanField() textinput.Model {
    input := textinput.New()
    input.Prompt = ""
    input.ShowSuggestions = true
    input.SetSuggestions([]string{"true", "false"})  // Enable tab completion

    // Show default value in placeholder
    input.Placeholder = "true (default)"  // Or "false (default)"

    return input
}

// Tab completion handler in Update()
func (m *FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            // Complete boolean suggestion
            if m.focusedField.Input.ShowSuggestions {
                suggestion := m.focusedField.Input.CurrentSuggestion()
                if suggestion != "" {
                    m.focusedField.Input.SetValue(suggestion)
                    m.focusedField.Input.CursorEnd()
                    return m, nil
                }
            }
        }
    }
    return m, nil
}
```

### **Integer Fields with Range Validation**
**Innovation**: Real-time validation with human-readable formatting

```go
type IntegerFieldConfig struct {
    Key         string
    Title       string
    Description string
    Min         int
    Max         int
    Default     int
}

func (m *FormModel) addIntegerField(config IntegerFieldConfig) {
    input := textinput.New()
    input.Prompt = ""
    input.PlaceholderStyle = m.styles.FormPlaceholder

    // Human-readable placeholder with default
    input.Placeholder = fmt.Sprintf("%s (%s default)",
        formatNumber(config.Default), formatBytes(config.Default))

    // Add validation range to description
    fullDescription := fmt.Sprintf("%s (Range: %s - %s)",
        config.Description, formatBytes(config.Min), formatBytes(config.Max))

    field := FormField{
        Key:         config.Key,
        Title:       config.Title,
        Description: fullDescription,
        Input:       input,
        Min:         config.Min,
        Max:         config.Max,
    }

    m.fields = append(m.fields, field)
}

// Real-time validation
func (m *FormModel) validateIntegerField(field *FormField) {
    value := field.Input.Value()

    if value == "" {
        field.Input.Placeholder = fmt.Sprintf("%s (default)", formatNumber(field.Default))
        return
    }

    if intVal, err := strconv.Atoi(value); err != nil {
        field.Input.Placeholder = "Enter a valid number or leave empty for default"
    } else {
        if intVal < field.Min || intVal > field.Max {
            field.Input.Placeholder = fmt.Sprintf("Range: %s - %s",
                formatBytes(field.Min), formatBytes(field.Max))
        } else {
            field.Input.Placeholder = "" // Clear error
        }
    }
}
```

### **Value Formatting Utilities**
**Critical**: Consistent formatting across all forms

```go
// Universal byte formatting for configuration values
func formatBytes(bytes int) string {
    if bytes >= 1048576 {
        return fmt.Sprintf("%.1fMB", float64(bytes)/1048576)
    } else if bytes >= 1024 {
        return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
    }
    return fmt.Sprintf("%d bytes", bytes)
}

// Universal number formatting for display
func formatNumber(num int) string {
    if num >= 1000000 {
        return fmt.Sprintf("%.1fM", float64(num)/1000000)
    } else if num >= 1000 {
        return fmt.Sprintf("%.1fK", float64(num)/1000)
    }
    return strconv.Itoa(num)
}

// Usage in forms and info panels
sections = append(sections, fmt.Sprintf("• Memory Limit: %s", formatBytes(memoryLimit)))
sections = append(sections, fmt.Sprintf("• Estimated tokens: ~%s", formatNumber(tokenCount)))
```

### **Environment Variable Integration Pattern**
**Innovation**: Direct EnvVar integration with presence detection

```go
// EnvVar wrapper (from loader package)
type EnvVar struct {
    Key     string
    Value   string  // Current value in environment
    Default string  // Default value from config
}

func (e EnvVar) IsPresent() bool {
    return e.Value != "" // Check if actually set in environment
}

// Form field creation from EnvVar
func (m *FormModel) addFieldFromEnvVar(envVarName, fieldKey, title, description string) {
    envVar, _ := m.controller.GetVar(envVarName)

    // Track initially set fields for cleanup logic
    m.initiallySetFields[fieldKey] = envVar.IsPresent()

    input := textinput.New()
    input.Prompt = ""

    // Show default in placeholder if not set
    if !envVar.IsPresent() {
        input.Placeholder = fmt.Sprintf("%s (default)", envVar.Default)
    } else {
        input.SetValue(envVar.Value) // Set current value
    }

    field := FormField{
        Key:         fieldKey,
        Title:       title,
        Description: description,
        Input:       input,
        EnvVarName:  envVarName,
    }

    m.fields = append(m.fields, field)
}
```

### **Smart Field Cleanup Pattern**
**Innovation**: Environment variable cleanup for empty values

```go
func (m *FormModel) saveConfiguration() error {
    // First pass: Remove cleared fields from environment
    for _, field := range m.fields {
        value := strings.TrimSpace(field.Input.Value())

        // If field was initially set but now empty, remove it
        if value == "" && m.initiallySetFields[field.Key] {
            if err := m.controller.SetVar(field.EnvVarName, ""); err != nil {
                return fmt.Errorf("failed to clear %s: %w", field.EnvVarName, err)
            }
            logger.Log("[FormModel] SAVE: cleared %s", field.EnvVarName)
        }
    }

    // Second pass: Save only non-empty values
    for _, field := range m.fields {
        value := strings.TrimSpace(field.Input.Value())
        if value == "" {
            continue // Skip empty - use defaults
        }

        // Validate before saving
        if err := m.validateFieldValue(field, value); err != nil {
            return fmt.Errorf("validation failed for %s: %w", field.Key, err)
        }

        // Save validated value
        if err := m.controller.SetVar(field.EnvVarName, value); err != nil {
            return fmt.Errorf("failed to set %s: %w", field.EnvVarName, err)
        }
        logger.Log("[FormModel] SAVE: set %s=%s", field.EnvVarName, value)
    }

    return nil
}
```

### **Resource Estimation Pattern**
**Innovation**: Real-time calculation of resource usage

```go
func (m *ConfigFormModel) calculateResourceEstimate() string {
    // Get current form values or defaults
    maxMemory := m.getIntValueOrDefault("max_memory")
    maxConnections := m.getIntValueOrDefault("max_connections")
    cacheSize := m.getIntValueOrDefault("cache_size")

    // Algorithm-specific calculations
    var estimatedMemory int
    switch m.configType {
    case "database":
        estimatedMemory = maxMemory + (maxConnections * 1024) + cacheSize
    case "worker":
        estimatedMemory = maxMemory * maxConnections
    default:
        estimatedMemory = maxMemory
    }

    // Convert to human-readable format
    return fmt.Sprintf("~%s RAM", formatBytes(estimatedMemory))
}

// Helper to get form value or default
func (m *FormModel) getIntValueOrDefault(fieldKey string) int {
    // First check current form input
    for _, field := range m.fields {
        if field.Key == fieldKey {
            if value := strings.TrimSpace(field.Input.Value()); value != "" {
                if intVal, err := strconv.Atoi(value); err == nil {
                    return intVal
                }
            }
        }
    }

    // Fall back to environment default
    envVar, _ := m.controller.GetVar(m.getEnvVarName(fieldKey))
    if defaultVal, err := strconv.Atoi(envVar.Default); err == nil {
        return defaultVal
    }

    return 0
}

// Display in form content
func (m *FormModel) updateFormContent() {
    // ... form fields ...

    // Resource estimation section
    sections = append(sections, "")
    sections = append(sections, m.styles.Subtitle.Render("Resource Estimation"))
    sections = append(sections, m.styles.Paragraph.Render("Estimated usage: "+m.calculateResourceEstimate()))

    m.formContent = strings.Join(sections, "\n")
    m.viewport.SetContent(m.formContent)
}
```

### **Current Configuration Preview Pattern**
**Innovation**: Live display of current settings in info panel

```go
func (m *TypeSelectionModel) renderConfigurationPreview() string {
    selectedType := m.types[m.selectedIndex]
    var sections []string

    // Helper to get current environment values
    getValue := func(suffix string) string {
        envVar, _ := m.controller.GetVar(m.getEnvVarName(selectedType.ID, suffix))
        if envVar.Value != "" {
            return envVar.Value
        }
        return envVar.Default + " (default)"
    }

    getIntValue := func(suffix string) int {
        envVar, _ := m.controller.GetVar(m.getEnvVarName(selectedType.ID, suffix))
        if envVar.Value != "" {
            if val, err := strconv.Atoi(envVar.Value); err == nil {
                return val
            }
        }
        if val, err := strconv.Atoi(envVar.Default); err == nil {
            return val
        }
        return 0
    }

    // Display current configuration
    sections = append(sections, m.styles.Subtitle.Render("Current Configuration"))
    sections = append(sections, "")

    maxMemory := getIntValue("MAX_MEMORY")
    timeout := getIntValue("TIMEOUT")
    enabled := getValue("ENABLED")

    sections = append(sections, fmt.Sprintf("• Max Memory: %s", formatBytes(maxMemory)))
    sections = append(sections, fmt.Sprintf("• Timeout: %d seconds", timeout))
    sections = append(sections, fmt.Sprintf("• Enabled: %s", enabled))

    // Type-specific configuration
    if selectedType.ID == "advanced" {
        retries := getIntValue("MAX_RETRIES")
        sections = append(sections, fmt.Sprintf("• Max Retries: %d", retries))
    }

    return strings.Join(sections, "\n")
}
```

### **Type-Based Dynamic Forms**
**Innovation**: Conditional field generation based on selection

```go
func (m *FormModel) buildDynamicForm() {
    m.fields = []FormField{} // Reset

    // Common fields for all types
    m.addFieldFromEnvVar("ENABLED", "enabled", "Enable Service", "Enable or disable this service")
    m.addFieldFromEnvVar("MAX_MEMORY", "max_memory", "Memory Limit", "Maximum memory usage in bytes")

    // Type-specific fields
    switch m.configType {
    case "database":
        m.addFieldFromEnvVar("MAX_CONNECTIONS", "max_connections", "Max Connections", "Maximum database connections")
        m.addFieldFromEnvVar("CACHE_SIZE", "cache_size", "Cache Size", "Database cache size in bytes")

    case "worker":
        m.addFieldFromEnvVar("WORKER_COUNT", "worker_count", "Worker Count", "Number of worker processes")
        m.addFieldFromEnvVar("QUEUE_SIZE", "queue_size", "Queue Size", "Maximum queue size")

    case "api":
        m.addFieldFromEnvVar("RATE_LIMIT", "rate_limit", "Rate Limit", "API requests per minute")
        m.addFieldFromEnvVar("TIMEOUT", "timeout", "Request Timeout", "Request timeout in seconds")
    }

    // Set focus on first field
    if len(m.fields) > 0 {
        m.fields[0].Input.Focus()
    }
}

// Environment variable naming helper
func (m *FormModel) getEnvVarName(configType, suffix string) string {
    prefix := strings.ToUpper(configType) + "_"
    return prefix + suffix
}
```

These advanced patterns enable:
- **Smart Validation**: Real-time feedback with user-friendly error messages
- **Resource Awareness**: Live estimation of memory, CPU, or token usage
- **Environment Integration**: Proper handling of defaults, presence detection, and cleanup
- **Type Safety**: Compile-time validation and runtime error handling
- **User Experience**: Auto-completion, formatting, and intuitive navigation

## 🎯 **Production Form Architecture Patterns**

### Form Model Structure (Latest Pattern)
**Based on successful llm_provider_form.go and summarizer_form.go implementations**

```go
type FormModel struct {
    controller *controllers.StateController
    styles     *styles.Styles
    window     *window.Window

    // Core form state
    fields       []FormField
    focusedIndex int
    showValues   bool
    hasChanges   bool
    args         []string // Arguments from composite ScreenID

    // Enhanced state tracking (from summarizer implementation)
    initialized        bool
    configType        string
    typeName          string
    initiallySetFields map[string]bool // Track fields for cleanup

    // Viewport as permanent property for forms
    viewport     viewport.Model
    formContent  string
    fieldHeights []int
}

// Constructor pattern - args from composite ScreenID
func NewFormModel(
    controller *controllers.StateController, styles *styles.Styles,
    window *window.Window, args []string,
) *FormModel {
    // Extract primary argument (e.g., provider ID)
    primaryArg := "default"
    if len(args) > 0 && args[0] != "" {
        primaryArg = args[0]
    }

    return &FormModel{
        controller: controller,
        styles:     styles,
        window:     window,
        args:       args,
        viewport:   viewport.New(window.GetContentSize()), // Permanent viewport
    }
}
```

### Key Form Implementation Patterns

#### 1. Proper Navigation Hotkeys
```go
// Modern form navigation (Production Pattern - from summarizer_form.go)
switch msg.String() {
case "down":           // ↓: Next field
    m.focusNext()
    m.ensureFocusVisible()

case "up":             // ↑: Previous field
    m.focusPrev()
    m.ensureFocusVisible()

case "tab":            // Tab: Complete suggestion (boolean auto-complete)
    m.completeSuggestion()

case "ctrl+h":         // Ctrl+H: Toggle show/hide masked values
    m.toggleShowValues()

case "ctrl+s":         // Ctrl+S: Save configuration only
    return m.saveConfiguration()

case "ctrl+r":         // Ctrl+R: Reset form to defaults
    m.resetForm()
    return m, nil

case "enter":          // Enter: Save and return (GoBack navigation)
    return m.saveAndReturn()
}

// Enhanced field navigation with auto-scroll
func (m *FormModel) focusNext() {
    if len(m.fields) == 0 {
        return
    }
    m.fields[m.focusedIndex].Input.Blur()
    m.focusedIndex = (m.focusedIndex + 1) % len(m.fields)
    m.fields[m.focusedIndex].Input.Focus()
    m.updateFormContent()
}
```

#### 2. Suggestions and Auto-completion
```go
// Boolean field with suggestions (from summarizer_form.go)
func (m *FormModel) addBooleanField(key, title, description string, envVar loader.EnvVar) {
    input := textinput.New()
    input.Prompt = ""
    input.PlaceholderStyle = m.styles.FormPlaceholder
    input.ShowSuggestions = true
    input.SetSuggestions([]string{"true", "false"})

    // Show default in placeholder
    if envVar.Default == "true" {
        input.Placeholder = "true (default)"
    } else {
        input.Placeholder = "false (default)"
    }

    // Set value only if actually present in environment
    if envVar.Value != "" && envVar.IsPresent() {
        input.SetValue(envVar.Value)
    }
}

// Tab completion handler
func (m *FormModel) completeSuggestion() {
    if m.focusedIndex < len(m.fields) {
        suggestion := m.fields[m.focusedIndex].Input.CurrentSuggestion()
        if suggestion != "" {
            m.fields[m.focusedIndex].Input.SetValue(suggestion)
            m.fields[m.focusedIndex].Input.CursorEnd()
            m.fields[m.focusedIndex].Value = suggestion
            m.hasChanges = true
            m.updateFormContent()
        }
    }
}
```

#### 3. Dynamic Input Width Calculation
```go
// Adaptive input sizing
func (m *FormModel) getInputWidth() int {
    viewportWidth, _ := m.getViewportSize()
    inputWidth := viewportWidth - 6  // Account for padding
    if m.isVerticalLayout() {
        inputWidth = viewportWidth - 4  // Less padding in vertical
    }
    return inputWidth
}

func (m *FormModel) getViewportSize() (int, int) {
    contentWidth, contentHeight := m.window.GetContentSize()
    if contentWidth <= 0 || contentHeight <= 0 {
        return 0, 0
    }

    if m.isVerticalLayout() {
        return contentWidth - PaddingWidth/2, contentHeight - PaddingHeight
    } else {
        leftWidth := MinMenuWidth
        extraWidth := contentWidth - leftWidth - MinInfoWidth - PaddingWidth
        if extraWidth > 0 {
            leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
        }
        return leftWidth, contentHeight - PaddingHeight
    }
}
```

#### 4. Viewport as Permanent Form Property
```go
// ✅ CORRECT: Viewport as permanent property for forms
type FormModel struct {
    viewport viewport.Model  // Permanent - preserves scroll position
}

// Update viewport dimensions on resize
func (m *FormModel) updateViewport() {
    formContentHeight := lipgloss.Height(m.formContent) + 2
    viewportWidth, viewportHeight := m.getViewportSize()
    m.viewport.Width = viewportWidth
    m.viewport.Height = min(viewportHeight, formContentHeight)
    m.viewport.SetContent(m.formContent)
}

// ❌ WRONG: Creating viewport in View() - loses scroll state
func (m *FormModel) View() string {
    vp := viewport.New(width, height) // State lost on re-render!
    return vp.View()
}
```

### Layout Architecture (Two-Column Pattern)

#### 1. Layout Constants (Production Values)
```go
const (
    MinMenuWidth  = 38  // Minimum left panel width
    MaxMenuWidth  = 66  // Maximum left panel width (prevents too wide)
    MinInfoWidth  = 34  // Minimum right panel width
    PaddingWidth  = 8   // Total horizontal padding
    PaddingHeight = 2   // Vertical padding
)
```

#### 2. Adaptive Layout Logic
```go
func (m *Model) isVerticalLayout() bool {
    contentWidth := m.window.GetContentWidth()
    return contentWidth < (MinMenuWidth + MinInfoWidth + PaddingWidth)
}

// Horizontal layout with dynamic width allocation
func (m *Model) renderHorizontalLayout(leftPanel, rightPanel string, width, height int) string {
    leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
    extraWidth := width - leftWidth - rightWidth - PaddingWidth

    // Distribute extra space, but cap left panel at MaxMenuWidth
    if extraWidth > 0 {
        leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
        rightWidth = width - leftWidth - PaddingWidth/2
    }

    leftStyled := lipgloss.NewStyle().Width(leftWidth).Padding(0, 2, 0, 2).Render(leftPanel)
    rightStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(rightPanel)

    // Use viewport for final layout rendering
    viewport := viewport.New(width, height-PaddingHeight)
    viewport.SetContent(lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled))
    return viewport.View()
}
```

#### 3. Content Hiding When Space Insufficient
```go
// Vertical layout with conditional content hiding
func (m *Model) renderVerticalLayout(leftPanel, rightPanel string, width, height int) string {
    verticalStyle := lipgloss.NewStyle().Width(width).Padding(0, 4, 0, 2)

    leftStyled := verticalStyle.Render(leftPanel)
    rightStyled := verticalStyle.Render(rightPanel)

    // Hide right panel if both don't fit
    if lipgloss.Height(leftStyled)+lipgloss.Height(rightStyled)+2 < height {
        return lipgloss.JoinVertical(lipgloss.Left,
            leftStyled,
            verticalStyle.Height(1).Render(""),
            rightStyled,
        )
    }

    // Show only essential left panel
    return leftStyled
}
```

### Composite ScreenID Navigation (Production Pattern)

#### 1. Proper ScreenID Creation for Navigation
```go
// Navigation from menu with argument preservation
func (m *MenuModel) handleSelection() (tea.Model, tea.Cmd) {
    selectedItem := m.getSelectedItem()

    // Create composite ScreenID with current selection for stack preservation
    return m, func() tea.Msg {
        return NavigationMsg{
            Target: CreateScreenID(string(targetScreen), selectedItem.ID),
        }
    }
}

// Form navigation back - use GoBack to avoid stack loops
func (m *FormModel) saveAndReturn() (tea.Model, tea.Cmd) {
    model, cmd := m.saveConfiguration()
    if cmd != nil {
        return model, cmd
    }

    // ✅ CORRECT: Use GoBack to return to previous screen
    return m, func() tea.Msg {
        return NavigationMsg{GoBack: true}
    }
}

// ❌ WRONG: Direct navigation creates stack loops
return m, func() tea.Msg {
    return NavigationMsg{Target: LLMProvidersScreen} // Creates loop!
}
```

#### 2. Constructor with Args Pattern
```go
// Model constructor receives args from composite ScreenID
func NewModel(
    controller *controllers.StateController, styles *styles.Styles,
    window *window.Window, args []string,
) *Model {
    // Initialize with selection from args
    selectedIndex := 0
    if len(args) > 1 && args[1] != "" {
        // Find matching item and set selectedIndex
        for i, item := range items {
            if item.ID == args[1] {
                selectedIndex = i
                break
            }
        }
    }

    return &Model{
        controller:    controller,
        selectedIndex: selectedIndex,
        args:          args,
    }
}

// No separate SetSelected* methods needed
func (m *Model) Init() tea.Cmd {
    logger.Log("[Model] INIT: args=%s", strings.Join(m.args, " § "))

    // Selection already set in constructor from args
    m.loadData()
    return nil
}
```

### Viewport Usage Patterns

#### 1. Forms: Permanent Viewport Property
```go
// ✅ For forms with user interaction and scroll state
type FormModel struct {
    viewport viewport.Model  // Permanent - preserves scroll position
}

func (m *FormModel) ensureFocusVisible() {
    // Auto-scroll to focused field
    focusY := m.calculateFieldPosition(m.focusedIndex)
    if focusY < m.viewport.YOffset {
        m.viewport.YOffset = focusY
    }
    // ... scroll logic
}
```

#### 2. Layout: Temporary Viewport Creation
```go
// ✅ For final layout rendering only
func (m *Model) renderHorizontalLayout(left, right string, width, height int) string {
    content := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)

    // Create viewport just for layout rendering
    vp := viewport.New(width, height-PaddingHeight)
    vp.SetContent(content)
    return vp.View()
}
```

### Screen Architecture (App.go Integration)

#### 1. Content Area Only Pattern
```go
// Screen models ONLY handle content area
func (m *Model) View() string {
    // ✅ CORRECT: Only content, no header/footer
    leftPanel := m.renderForm()
    rightPanel := m.renderHelp()

    if m.isVerticalLayout() {
        return m.renderVerticalLayout(leftPanel, rightPanel, width, height)
    }
    return m.renderHorizontalLayout(leftPanel, rightPanel, width, height)
}

// ❌ WRONG: Handling header/footer in screen
func (m *Model) View() string {
    header := m.renderHeader()    // App.go handles this!
    footer := m.renderFooter()    // App.go handles this!
    // ...
}
```

#### 2. App.go Layout Management
```go
// App.go manages complete layout structure
func (a *App) View() string {
    header := a.renderHeader()    // Screen-specific header
    footer := a.renderFooter()    // Dynamic footer with actions
    content := a.currentModel.View()  // Content from model

    // Calculate content area size
    contentWidth, contentHeight := a.window.GetContentSize()
    contentArea := a.styles.Content.
        Width(contentWidth).
        Height(contentHeight).
        Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}
```

### Field Configuration Best Practices

#### 1. Clean Input Setup
```go
// Modern input field setup
func (m *FormModel) addInputField(config *Config, fieldType string) {
    input := textinput.New()
    input.Prompt = ""  // Clean appearance
    input.PlaceholderStyle = m.styles.FormPlaceholder

    // Dynamic width - set during rendering
    // input.Width NOT set here - calculated in updateFormContent()

    if fieldType == "password" {
        input.EchoMode = textinput.EchoPassword
    }

    if fieldType == "boolean" {
        input.ShowSuggestions = true
        input.SetSuggestions([]string{"true", "false"})
    }

    // Set value from config
    if config != nil {
        input.SetValue(config.GetValue(fieldType))
    }
}
```

#### 2. Dynamic Width Application
```go
// Apply width during content update
func (m *FormModel) updateFormContent() {
    inputWidth := m.getInputWidth()

    for i, field := range m.fields {
        // Apply dynamic width to input
        field.Input.Width = inputWidth - 3  // Account for borders
        field.Input.SetValue(field.Input.Value())  // Trigger width update

        // Render with consistent styling
        inputStyle := m.styles.FormInput.Width(inputWidth)
        if i == m.focusedIndex {
            inputStyle = inputStyle.BorderForeground(styles.Primary)
        }

        renderedInput := inputStyle.Render(field.Input.View())
        sections = append(sections, renderedInput)
    }
}
```

### State Management Best Practices

#### 1. Configuration vs Status Separation
```go
// ✅ SIMPLIFIED: Single status field
type ProviderInfo struct {
    ID          string
    Name        string
    Description string
    Configured  bool  // Single status - provider has required fields
}

// Load status logic
func (m *Model) loadProviders() {
    configs := m.controller.GetLLMProviders()

    provider := ProviderInfo{
        ID:         "openai",
        Name:       locale.LLMProviderOpenAI,
        Configured: configs["openai"].Configured,  // From controller
    }
}

// ❌ COMPLEX: Multiple status fields (removed)
type ProviderInfo struct {
    Configured bool
    Enabled    bool  // Removed - controller handles this
}
```

#### 2. GoBack Navigation Pattern
```go
// ✅ CORRECT: GoBack prevents navigation loops
func (m *FormModel) saveAndReturn() (tea.Model, tea.Cmd) {
    if err := m.saveConfiguration(); err != nil {
        return m, nil  // Stay on form if save fails
    }

    // Return to previous screen (from navigation stack)
    return m, func() tea.Msg {
        return NavigationMsg{GoBack: true}
    }
}

// Navigation stack automatically maintained:
// ["main_menu§llm_providers", "llm_providers§openai", "llm_provider_form§openai"]
// GoBack removes current and returns to: "llm_providers§openai"
```

This production architecture ensures:
- **Clean separation**: Forms handle content, app.go handles layout
- **Persistent state**: Viewport scroll positions maintained
- **Adaptive design**: Content hides gracefully when space insufficient
- **Type-safe navigation**: Arguments preserved in composite ScreenIDs
- **No navigation loops**: GoBack pattern prevents stack corruption
