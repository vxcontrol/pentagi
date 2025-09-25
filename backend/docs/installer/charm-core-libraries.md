# Charm.sh Core Libraries Reference

> Comprehensive guide to the core libraries in the Charm ecosystem for building TUI applications.

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

## 🏗️ **Core Integration Patterns**

### Centralized Styles & Dimensions

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

### Component Initialization Pattern
```go
// Standard component initialization
func NewModel(styles *styles.Styles, window *window.Window) *Model {
    viewport := viewport.New(window.GetContentSize())
    viewport.Style = lipgloss.NewStyle() // Clean style

    return &Model{
        styles:   styles,
        window:   window,
        viewport: viewport,
    }
}

func (m *Model) Init() tea.Cmd {
    // ALWAYS reset ALL state
    m.content = ""
    m.ready = false
    m.error = nil
    return m.loadContent
}
```

### Essential Key Handling
```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Update through styles, not directly
        // m.styles.SetSize() called by app.go
        m.updateViewport()

    case tea.KeyMsg:
        switch msg.String() {
        case "up":
            m.viewport.ScrollUp(1)
        case "down":
            m.viewport.ScrollDown(1)
        case "left":
            m.viewport.ScrollLeft(2)
        case "right":
            m.viewport.ScrollRight(2)
        case "pgup":
            m.viewport.ScrollUp(m.viewport.Height)
        case "pgdown":
            m.viewport.ScrollDown(m.viewport.Height)
        case "home":
            m.viewport.GotoTop()
        case "end":
            m.viewport.GotoBottom()
        }
    }
    return m, nil
}
```

## 🔧 **Common Integration Patterns**

### Content Loading
```go
// Load content asynchronously
func (m *Model) loadContent() tea.Cmd {
    return func() tea.Msg {
        content, err := loadFromSource()
        if err != nil {
            return ErrorMsg{err}
        }
        return ContentLoadedMsg{content}
    }
}

// Handle loading in Update
case ContentLoadedMsg:
    m.content = msg.Content
    m.ready = true
    m.viewport.SetContent(m.content)
    return m, nil
```

### Error Handling
```go
// Graceful error handling
func (m *Model) View() string {
    width, height := m.styles.GetSize()
    if width <= 0 || height <= 0 {
        return "Loading..." // Graceful fallback
    }

    if m.error != nil {
        return m.styles.Error.Render("Error: " + m.error.Error())
    }

    if !m.ready {
        return m.styles.Info.Render("Loading content...")
    }

    return m.viewport.View()
}
```

This reference provides the foundation for building robust TUI applications with the Charm ecosystem. Each library serves a specific purpose and when combined correctly, creates powerful terminal interfaces.