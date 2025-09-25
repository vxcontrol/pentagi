# Charm.sh Production Architecture Patterns

> Essential patterns for building robust, scalable TUI applications with the Charm ecosystem.

## 🏗️ **Centralized Styles & Dimensions**

### **Styles Singleton Pattern**
**CRITICAL**: Never store width/height in models - use styles singleton

```go
// ✅ CORRECT: Centralized in styles
type Styles struct {
    width    int
    height   int
    renderer *glamour.TermRenderer

    // Core styles
    Header      lipgloss.Style
    Footer      lipgloss.Style
    Content     lipgloss.Style
    Title       lipgloss.Style
    Subtitle    lipgloss.Style
    Paragraph   lipgloss.Style

    // Form styles
    FormLabel       lipgloss.Style
    FormInput       lipgloss.Style
    FormPlaceholder lipgloss.Style
    FormHelp        lipgloss.Style

    // Status styles
    Success lipgloss.Style
    Error   lipgloss.Style
    Warning lipgloss.Style
    Info    lipgloss.Style
}

func New() *Styles {
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )

    styles := &Styles{
        renderer: renderer,
        width:    80,
        height:   24,
    }

    styles.updateStyles()
    return styles
}

func (s *Styles) SetSize(width, height int) {
    s.width = width
    s.height = height
    s.updateStyles()  // Recalculate responsive styles
}

func (s *Styles) GetSize() (int, int) {
    return s.width, s.height
}

func (s *Styles) GetWidth() int {
    return s.width
}

func (s *Styles) GetHeight() int {
    return s.height
}

func (s *Styles) GetRenderer() *glamour.TermRenderer {
    return s.renderer
}

// Update styles based on current dimensions
func (s *Styles) updateStyles() {
    // Base styles
    s.Header = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205"))

    s.Footer = lipgloss.NewStyle().
        Width(s.width).
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("255")).
        Padding(0, 1, 0, 1)

    s.Content = lipgloss.NewStyle().
        Width(s.width).
        Padding(1, 2, 1, 2)

    // Form styles with responsive sizing
    s.FormInput = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240")).
        Padding(0, 1, 0, 1)

    // Status styles
    s.Success = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    s.Error = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
    s.Warning = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
    s.Info = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
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

## 🏗️ **Unified Header/Footer Management**

### **App-Level Layout Control**
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
        title := a.getScreenTitle(a.navigator.Current())
        return a.styles.Header.Render(title)
    }
}

func (a *App) renderFooter() string {
    actions := a.buildFooterActions()
    footerText := strings.Join(actions, " • ")

    return a.styles.Footer.Render(footerText)
}

func (a *App) buildFooterActions() []string {
    actions := []string{"Esc: Back", "Ctrl+C: Exit"}

    // Screen-specific actions
    switch a.navigator.Current().GetScreen() {
    case "eula":
        if a.isEULAScrolledToEnd() {
            actions = append(actions, "Y: Accept", "N: Reject")
        } else {
            actions = append(actions, "↑↓: Scroll", "PgUp/PgDn: Page")
        }
    case "form":
        actions = append(actions, "Tab: Complete", "Ctrl+S: Save", "Enter: Save & Return")
    case "menu":
        actions = append(actions, "Enter: Select")
    }

    return actions
}
```

### **Background Footer Approach (Production)**
```go
// ✅ PRODUCTION READY: Background approach (always 1 line)
func (s *Styles) createFooter(width int, text string) string {
    return lipgloss.NewStyle().
        Width(width).
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("255")).
        Padding(0, 1, 0, 1).
        Render(text)
}

// ❌ WRONG: Border approach (height inconsistency)
func createFooterWrong(width int, text string) string {
    return lipgloss.NewStyle().
        Width(width).
        Height(1).
        Border(lipgloss.Border{Top: true}).
        Render(text)
}
```

## 🏗️ **Component Initialization Pattern**

### **Standard Component Architecture**
```go
// Standard component initialization
func NewModel(styles *styles.Styles, window *window.Window, controller *controllers.StateController) *Model {
    viewport := viewport.New(window.GetContentSize())
    viewport.Style = lipgloss.NewStyle() // Clean style prevents conflicts

    return &Model{
        styles:     styles,
        window:     window,
        controller: controller,
        viewport:   viewport,
        initialized: false,
    }
}

func (m *Model) Init() tea.Cmd {
    // ALWAYS reset ALL state
    m.content = ""
    m.ready = false
    m.error = nil
    m.initialized = false

    logger.Log("[Model] INIT: starting initialization")
    return m.loadContent
}

// Window size handling
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Update through styles, not directly
        // m.styles.SetSize() called by app.go
        m.updateViewport()

    case ContentLoadedMsg:
        m.content = msg.Content
        m.ready = true
        m.initialized = true
        m.viewport.SetContent(m.content)
        return m, nil

    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
    }

    return m, nil
}
```

## 🏗️ **Model State Management**

### **Complete State Reset Pattern**
```go
// Model interface implementation
type Model struct {
    styles *styles.Styles  // ALWAYS use shared styles
    window *window.Window

    // Core state
    content     string
    ready       bool
    error       error
    initialized bool

    // Component state
    viewport viewport.Model

    // Navigation state
    args []string
}

func (m *Model) Init() tea.Cmd {
    logger.Log("[Model] INIT")

    // ALWAYS reset ALL state completely
    m.content = ""
    m.ready = false
    m.error = nil
    m.initialized = false

    // Reset component state
    m.viewport.GotoTop()
    m.viewport.SetContent("")

    return m.loadContent
}

// Force re-render after async operations
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ContentLoadedMsg:
        m.content = msg.Content
        m.ready = true
        m.initialized = true
        m.viewport.SetContent(m.content)

        // Force view update with no-op command
        return m, func() tea.Msg { return nil }
    }
    return m, nil
}
```

### **Graceful Error Handling**
```go
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

## 🏗️ **Responsive Layout Architecture**

### **Breakpoint-Based Design**
```go
// Layout Constants
const (
    SmallScreenThreshold = 30    // Height threshold for viewport mode
    MinTerminalWidth = 80        // Minimum width for horizontal layout
    MinPanelWidth = 25          // Panel width constraints
    WelcomeHeaderHeight = 8     // Fixed by ASCII Art Logo (8 lines)
    EULAHeaderHeight = 3        // Title + subtitle + spacing
    FooterHeight = 1           // Always 1 line with background approach
)

// Responsive layout detection
func (m *Model) isVerticalLayout() bool {
    width := m.styles.GetWidth()
    return width < MinTerminalWidth
}

func (m *Model) isSmallScreen() bool {
    height := m.styles.GetHeight()
    return height < SmallScreenThreshold
}

// Adaptive content rendering
func (m *Model) View() string {
    width, height := m.styles.GetSize()

    leftPanel := m.renderContent()
    rightPanel := m.renderInfo()

    if m.isVerticalLayout() {
        return m.renderVerticalLayout(leftPanel, rightPanel, width, height)
    }

    return m.renderHorizontalLayout(leftPanel, rightPanel, width, height)
}

func (m *Model) renderVerticalLayout(leftPanel, rightPanel string, width, height int) string {
    verticalStyle := lipgloss.NewStyle().Width(width).Padding(0, 2, 0, 2)

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

func (m *Model) renderHorizontalLayout(leftPanel, rightPanel string, width, height int) string {
    leftWidth := width * 2 / 3
    rightWidth := width - leftWidth - 4

    if leftWidth < MinPanelWidth {
        leftWidth = MinPanelWidth
        rightWidth = width - leftWidth - 4
    }

    leftStyled := lipgloss.NewStyle().Width(leftWidth).Padding(0, 2, 0, 2).Render(leftPanel)
    rightStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(rightPanel)

    return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)
}
```

## 🏗️ **Content Loading Architecture**

### **Async Content Loading**
```go
// Content loading pattern
func (m *Model) loadContent() tea.Cmd {
    return func() tea.Msg {
        logger.Log("[Model] LOAD: start")

        content, err := m.loadFromSource()
        if err != nil {
            logger.Errorf("[Model] LOAD: error: %v", err)
            return ErrorMsg{err}
        }

        logger.Log("[Model] LOAD: success (%d chars)", len(content))
        return ContentLoadedMsg{content}
    }
}

// Safe content loading with fallbacks
func (m *Model) loadFromSource() (string, error) {
    // Try embedded filesystem first
    if content, err := embedded.GetContent("file.md"); err == nil {
        return content, nil
    }

    // Fallback to direct file access
    if content, err := os.ReadFile("file.md"); err == nil {
        return string(content), nil
    }

    // Final fallback
    return "Content not available", fmt.Errorf("could not load content")
}

// Handle loading results
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ContentLoadedMsg:
        m.content = msg.Content
        m.ready = true

        // Render with glamour (safe pattern)
        rendered, err := m.styles.GetRenderer().Render(m.content)
        if err != nil {
            // Fallback to plain text
            rendered = fmt.Sprintf("# Content\n\n%s\n\n*Render error: %v*", m.content, err)
        }

        m.viewport.SetContent(rendered)
        return m, func() tea.Msg { return nil } // Force re-render

    case ErrorMsg:
        m.error = msg.Error
        m.ready = true
        return m, nil
    }

    return m, nil
}
```

## 🏗️ **Window Management Pattern**

### **Window Integration**
```go
type Window struct {
    width  int
    height int
}

func NewWindow() *Window {
    return &Window{width: 80, height: 24}
}

func (w *Window) SetSize(width, height int) {
    w.width = width
    w.height = height
}

func (w *Window) GetSize() (int, int) {
    return w.width, w.height
}

func (w *Window) GetContentSize() (int, int) {
    // Account for padding and borders
    contentWidth := max(w.width-4, 0)
    contentHeight := max(w.height-6, 0) // Header + footer + padding
    return contentWidth, contentHeight
}

func (w *Window) GetContentWidth() int {
    width, _ := w.GetContentSize()
    return width
}

func (w *Window) GetContentHeight() int {
    _, height := w.GetContentSize()
    return height
}

// Integration with models
func (m *Model) updateDimensions() {
    width, height := m.window.GetContentSize()
    if width <= 0 || height <= 0 {
        return
    }

    m.viewport.Width = width
    m.viewport.Height = height
}
```

## 🏗️ **Controller Integration Pattern**

### **StateController Bridge**
```go
type StateController struct {
    state *state.State
}

func NewStateController(state *state.State) *StateController {
    return &StateController{state: state}
}

// Environment variable management
func (c *StateController) GetVar(name string) (loader.EnvVar, error) {
    return c.state.GetVar(name)
}

func (c *StateController) SetVar(name, value string) error {
    return c.state.SetVar(name, value)
}

// Configuration management
func (c *StateController) GetLLMProviders() map[string]ProviderConfig {
    // Implementation specific to controller
    return c.state.GetLLMProviders()
}

func (c *StateController) SaveConfiguration() error {
    return c.state.Save()
}

// Model integration
type Model struct {
    controller *StateController
    // ... other fields
}

func (m *Model) loadConfiguration() {
    configs := m.controller.GetLLMProviders()
    // Use configs to populate model state
}
```

## 🏗️ **Essential Key Handling**

### **Universal Key Patterns**
```go
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    // Universal scroll controls
    case "up":
        m.viewport.ScrollUp(1)
    case "down":
        m.viewport.ScrollDown(1)
    case "left", "h":
        m.viewport.ScrollLeft(2)  // 2 steps for faster horizontal
    case "right", "l":
        m.viewport.ScrollRight(2)

    // Page controls
    case "pgup":
        m.viewport.ScrollUp(m.viewport.Height)
    case "pgdown":
        m.viewport.ScrollDown(m.viewport.Height)
    case "home":
        m.viewport.GotoTop()
    case "end":
        m.viewport.GotoBottom()

    // Universal actions (handled at app level)
    case "esc":
        // Handled by app.go - returns to welcome
        return m, nil
    case "ctrl+c":
        // Handled by app.go - quits application
        return m, nil

    default:
        // Screen-specific handling
        return m.handleScreenSpecificKeys(msg)
    }

    return m, nil
}

func (m *Model) handleScreenSpecificKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Override in specific models
    return m, nil
}
```

## 🎯 **Architecture Best Practices**

### **Separation of Concerns**
```go
// ✅ CORRECT: Clean separation
// app.go        - Navigation, layout, global state
// models/       - Screen-specific logic, local state
// styles/       - Styling, dimensions, rendering
// controller/   - Business logic, state management
// window/       - Terminal size management

// Model responsibilities
type Model struct {
    // ONLY screen-specific state
    content string
    ready   bool

    // Dependencies (injected)
    styles     *styles.Styles    // Shared styling
    window     *window.Window    // Size management
    controller *Controller       // Business logic
}

// App responsibilities
type App struct {
    // Global state
    navigator    *Navigator
    currentModel tea.Model

    // Shared resources
    styles     *styles.Styles
    window     *window.Window
    controller *Controller
}
```

### **Resource Management**
```go
// ✅ CORRECT: Shared resources
func NewApp() *App {
    styles := styles.New()
    window := window.New()
    controller := controller.New()

    return &App{
        styles:     styles,     // Single instance
        window:     window,     // Single instance
        controller: controller, // Single instance
        navigator:  navigator.New(),
    }
}

// ❌ WRONG: Resource duplication
func NewModelWrong() *Model {
    styles := styles.New()       // Multiple instances!
    renderer, _ := glamour.New() // Multiple renderers!
    return &Model{styles: styles}
}
```

This architecture provides:
- **Scalability**: Clean separation enables easy feature additions
- **Maintainability**: Centralized resources reduce coupling
- **Performance**: Shared instances prevent resource waste
- **Consistency**: Unified patterns across all components
- **Reliability**: Proper error handling and state management