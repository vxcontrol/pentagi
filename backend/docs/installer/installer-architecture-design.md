# PentAGI Installer Architecture & Design Patterns

> Architecture patterns, design decisions, and implementation strategies specific to the PentAGI installer.

## 🏗️ **Unified App Architecture**

### **Central Orchestrator Pattern**
The installer implements a centralized app controller that manages all global concerns:

```go
// File: wizard/app.go
type App struct {
    // Navigation state
    navigator *Navigator
    currentModel tea.Model

    // Shared resources (injected into all models)
    controller *controllers.StateController
    styles     *styles.Styles
    window     *window.Window

    // Global state
    eulaAccepted bool
    systemReady  bool
}

func (a *App) View() string {
    header := a.renderHeader()    // Screen-specific header
    footer := a.renderFooter()    // Dynamic footer with actions
    content := a.currentModel.View()  // Content only from model

    // App.go enforces layout constraints
    contentWidth, contentHeight := a.window.GetContentSize()
    contentArea := a.styles.Content.
        Width(contentWidth).
        Height(contentHeight).
        Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}
```

### **Responsibilities Separation**
- **App Layer**: Navigation, layout, global state, resource management
- **Model Layer**: Screen-specific logic, user interaction, content rendering
- **Controller Layer**: Business logic, environment variables, configuration
- **Styles Layer**: Presentation, theming, responsive calculations
- **Window Layer**: Terminal size management, dimension coordination

## 🏗️ **Navigation Architecture**

### **Composite ScreenID System**
**Innovation**: Parameters embedded in screen identifiers for type-safe navigation

```go
// Screen ID structure: "screen§arg1§arg2§..."
type ScreenID string

// Helper methods for parsing composite IDs
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

// Type-safe creation
func CreateScreenID(screen string, args ...string) ScreenID {
    if len(args) == 0 {
        return ScreenID(screen)
    }
    return ScreenID(screen + "§" + strings.Join(args, "§"))
}
```

### **Navigator Implementation**
```go
type Navigator struct {
    stack        []ScreenID
    stateManager StateManager // Persists stack across sessions
}

func (n *Navigator) Push(screenID ScreenID) {
    n.stack = append(n.stack, screenID)
    n.persistState()
}

func (n *Navigator) Pop() ScreenID {
    if len(n.stack) <= 1 {
        return n.stack[0] // Can't pop welcome screen
    }
    popped := n.stack[len(n.stack)-1]
    n.stack = n.stack[:len(n.stack)-1]
    n.persistState()
    return popped
}

// Universal ESC behavior
func (a *App) handleGlobalNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "esc":
        if a.navigator.Current().GetScreen() != string(WelcomeScreen) {
            a.navigator.stack = []ScreenID{WelcomeScreen}
            a.navigator.persistState()
            a.currentModel = a.createModelForScreen(WelcomeScreen, nil)
            return a, a.currentModel.Init()
        }
    }
    return a, nil
}
```

### **Args-Based Model Construction**
```go
func (a *App) createModelForScreen(screenID ScreenID, data any) tea.Model {
    baseScreen := screenID.GetScreen()
    args := screenID.GetArgs()

    switch ScreenID(baseScreen) {
    case LLMProviderFormScreen:
        providerID := "openai" // default
        if len(args) > 0 {
            providerID = args[0]
        }
        return NewLLMProviderFormModel(a.controller, a.styles, a.window, []string{providerID})

    case SummarizerFormScreen:
        summarizerType := "general" // default
        if len(args) > 0 {
            summarizerType = args[0]
        }
        return NewSummarizerFormModel(a.controller, a.styles, a.window, []string{summarizerType})
    }
}
```

## 🏗️ **Adaptive Layout Strategy**

### **Responsive Design Pattern**
The installer implements a sophisticated responsive design that adapts to terminal capabilities:

```go
// Layout constants define breakpoints
const (
    MinTerminalWidth = 80        // Minimum for horizontal layout
    MinMenuWidth     = 38        // Minimum left panel width
    MaxMenuWidth     = 66        // Maximum left panel width (prevents too wide forms)
    MinInfoWidth     = 34        // Minimum right panel width
    PaddingWidth     = 8         // Total horizontal padding
)

// Layout decision logic
func (m *Model) isVerticalLayout() bool {
    contentWidth := m.window.GetContentWidth()
    return contentWidth < (MinMenuWidth + MinInfoWidth + PaddingWidth)
}

// Dynamic width allocation
func (m *Model) renderHorizontalLayout(leftPanel, rightPanel string, width, height int) string {
    leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
    extraWidth := width - leftWidth - rightWidth - PaddingWidth

    // Distribute extra space intelligently, but cap left panel
    if extraWidth > 0 {
        leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)
        rightWidth = width - leftWidth - PaddingWidth/2
    }

    leftStyled := lipgloss.NewStyle().Width(leftWidth).Padding(0, 2, 0, 2).Render(leftPanel)
    rightStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(rightPanel)

    return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)
}
```

### **Content Hiding Strategy**
```go
func (m *Model) renderVerticalLayout(leftPanel, rightPanel string, width, height int) string {
    verticalStyle := lipgloss.NewStyle().Width(width).Padding(0, 4, 0, 2)

    leftStyled := verticalStyle.Render(leftPanel)
    rightStyled := verticalStyle.Render(rightPanel)

    // Show both panels if they fit
    if lipgloss.Height(leftStyled)+lipgloss.Height(rightStyled)+2 < height {
        return lipgloss.JoinVertical(lipgloss.Left,
            leftStyled,
            verticalStyle.Height(1).Render(""),
            rightStyled,
        )
    }

    // Hide right panel if insufficient space - show only essential content
    return leftStyled
}
```

## 🏗️ **Form Architecture Patterns**

### **Production Form Model Structure**
```go
type FormModel struct {
    // Standard dependencies (injected)
    controller *controllers.StateController
    styles     *styles.Styles
    window     *window.Window

    // Form state
    fields       []FormField
    focusedIndex int
    showValues   bool
    hasChanges   bool

    // Environment integration
    configType         string
    typeName          string
    initiallySetFields map[string]bool // Track for cleanup

    // Navigation state
    args []string // From composite ScreenID

    // Viewport as permanent property (preserves scroll state)
    viewport     viewport.Model
    formContent  string
    fieldHeights []int
}
```

### **Dynamic Field Generation Pattern**
```go
func (m *FormModel) buildForm() {
    m.fields = []FormField{}
    m.initiallySetFields = make(map[string]bool)

    // Helper function for consistent field creation
    addFieldFromEnvVar := func(suffix, key, title, description string) {
        envVar, _ := m.controller.GetVar(m.getEnvVarName(suffix))

        // Track initial state for cleanup
        m.initiallySetFields[key] = envVar.IsPresent()

        if key == "preserve_last" || key == "use_qa" {
            m.addBooleanField(key, title, description, envVar)
        } else {
            // Determine validation ranges
            var min, max int
            switch key {
            case "last_sec_bytes", "max_qa_bytes":
                min, max = 1024, 1048576 // 1KB to 1MB
            case "max_bp_bytes":
                min, max = 1024, 524288 // 1KB to 512KB
            default:
                min, max = 0, 999999
            }
            m.addIntegerField(key, title, description, envVar, min, max)
        }
    }

    // Type-specific field generation
    switch m.configType {
    case "general":
        addFieldFromEnvVar("USE_QA", "use_qa", locale.SummarizerFormUseQA, locale.SummarizerFormUseQADesc)
        addFieldFromEnvVar("SUM_MSG_HUMAN_IN_QA", "sum_human_in_qa", locale.SummarizerFormSumHumanInQA, locale.SummarizerFormSumHumanInQADesc)
    case "assistant":
        // Assistant-specific fields
    }

    // Common fields for all types
    addFieldFromEnvVar("PRESERVE_LAST", "preserve_last", locale.SummarizerFormPreserveLast, locale.SummarizerFormPreserveLastDesc)
    addFieldFromEnvVar("LAST_SEC_BYTES", "last_sec_bytes", locale.SummarizerFormLastSecBytes, locale.SummarizerFormLastSecBytesDesc)
}
```

### **Environment Variable Integration**
```go
// Environment variable naming pattern
func (m *FormModel) getEnvVarName(suffix string) string {
    var prefix string
    switch m.configType {
    case "assistant":
        prefix = "ASSISTANT_SUMMARIZER_"
    default:
        prefix = "SUMMARIZER_"
    }
    return prefix + suffix
}

// Smart cleanup pattern
func (m *FormModel) saveConfiguration() (tea.Model, tea.Cmd) {
    // First pass: Handle fields that were cleared (remove from environment)
    for _, field := range m.fields {
        value := strings.TrimSpace(field.Input.Value())

        // If field was initially set but now empty, remove it
        if value == "" && m.initiallySetFields[field.Key] {
            envVarName := m.getEnvVarName(getEnvSuffixFromKey(field.Key))

            if err := m.controller.SetVar(envVarName, ""); err != nil {
                logger.Errorf("[FormModel] SAVE: error clearing %s: %v", envVarName, err)
                return m, nil
            }
            logger.Log("[FormModel] SAVE: cleared %s", envVarName)
        }
    }

    // Second pass: Save only non-empty values
    for _, field := range m.fields {
        value := strings.TrimSpace(field.Input.Value())
        if value == "" {
            continue // Skip empty values - use defaults
        }

        envVarName := m.getEnvVarName(getEnvSuffixFromKey(field.Key))
        if err := m.controller.SetVar(envVarName, value); err != nil {
            logger.Errorf("[FormModel] SAVE: error setting %s: %v", envVarName, err)
            return m, nil
        }
    }

    return m, func() tea.Msg {
        return NavigationMsg{GoBack: true}
    }
}
```

## 🏗️ **Advanced Form Field Patterns**

### **Boolean Field with Auto-completion**
```go
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

    field := FormField{
        Key:         key,
        Title:       title,
        Description: description,
        Input:       input,
        Type:        "boolean",
    }

    m.fields = append(m.fields, field)
}
```

### **Integer Field with Validation**
```go
func (m *FormModel) addIntegerField(key, title, description string, envVar loader.EnvVar, min, max int) {
    input := textinput.New()
    input.Prompt = ""
    input.PlaceholderStyle = m.styles.FormPlaceholder

    // Parse and format default value
    defaultValue := 0
    if envVar.Default != "" {
        if val, err := strconv.Atoi(envVar.Default); err == nil {
            defaultValue = val
        }
    }

    // Human-readable placeholder with default
    input.Placeholder = fmt.Sprintf("%s (%s default)",
        m.formatNumber(defaultValue), m.formatBytes(defaultValue))

    // Set value only if present
    if envVar.Value != "" && envVar.IsPresent() {
        input.SetValue(envVar.Value)
    }

    // Add validation range to description
    fullDescription := fmt.Sprintf("%s (Range: %s - %s)",
        description, m.formatBytes(min), m.formatBytes(max))

    field := FormField{
        Key:         key,
        Title:       title,
        Description: fullDescription,
        Input:       input,
        Type:        "integer",
        Min:         min,
        Max:         max,
    }

    m.fields = append(m.fields, field)
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

// Higher-level configuration management
func (c *StateController) GetLLMProviders() map[string]ProviderConfig {
    // Aggregate multiple environment variables into structured config
    providers := make(map[string]ProviderConfig)

    for _, providerID := range []string{"openai", "anthropic", "gemini", "bedrock", "ollama", "custom"} {
        config := c.loadProviderConfig(providerID)
        providers[providerID] = config
    }

    return providers
}

func (c *StateController) loadProviderConfig(providerID string) ProviderConfig {
    prefix := strings.ToUpper(providerID) + "_"

    apiKey, _ := c.GetVar(prefix + "API_KEY")
    baseURL, _ := c.GetVar(prefix + "BASE_URL")

    return ProviderConfig{
        ID:         providerID,
        Configured: apiKey.IsPresent() && baseURL.IsPresent(),
        APIKey:     apiKey.Value,
        BaseURL:    baseURL.Value,
    }
}
```

## 🏗️ **Resource Estimation Architecture**

### **Token Calculation Pattern**
```go
func (m *FormModel) calculateTokenEstimate() string {
    // Get current form values or defaults
    useQAVal := m.getBoolValueOrDefault("use_qa")
    lastSecBytesVal := m.getIntValueOrDefault("last_sec_bytes")
    maxQABytesVal := m.getIntValueOrDefault("max_qa_bytes")
    keepQASectionsVal := m.getIntValueOrDefault("keep_qa_sections")

    var estimatedBytes int

    // Algorithm-specific calculations
    switch m.configType {
    case "assistant":
        estimatedBytes = keepQASectionsVal * lastSecBytesVal
    default: // general
        if useQAVal {
            basicSize := keepQASectionsVal * lastSecBytesVal
            if basicSize > maxQABytesVal {
                estimatedBytes = maxQABytesVal
            } else {
                estimatedBytes = basicSize
            }
        } else {
            estimatedBytes = keepQASectionsVal * lastSecBytesVal
        }
    }

    // Convert to tokens with overhead
    estimatedTokens := int(float64(estimatedBytes) * 1.1 / 4) // 4 bytes per token + 10% overhead

    return fmt.Sprintf("~%s tokens", m.formatNumber(estimatedTokens))
}

// Helper methods to get form values or environment defaults
func (m *FormModel) getBoolValueOrDefault(key string) bool {
    // First check form field value
    for _, field := range m.fields {
        if field.Key == key && field.Input.Value() != "" {
            return field.Input.Value() == "true"
        }
    }

    // Return default value from EnvVar
    envVar, _ := m.controller.GetVar(m.getEnvVarName(getEnvSuffixFromKey(key)))
    return envVar.Default == "true"
}
```

## 🏗️ **Auto-Scrolling Form Architecture**

### **Viewport-Based Scrolling**
```go
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

// Enhanced field navigation with auto-scroll
func (m *FormModel) focusNext() {
    if len(m.fields) == 0 {
        return
    }
    m.fields[m.focusedIndex].Input.Blur()
    m.focusedIndex = (m.focusedIndex + 1) % len(m.fields)
    m.fields[m.focusedIndex].Input.Focus()
    m.updateFormContent()
    m.ensureFocusVisible() // Key addition for auto-scroll
}
```

## 🏗️ **Layout Integration Architecture**

### **Content Area Management**
```go
// Models handle ONLY content area
func (m *Model) View() string {
    leftPanel := m.renderForm()
    rightPanel := m.renderHelp()

    // Adaptive layout decision
    if m.isVerticalLayout() {
        return m.renderVerticalLayout(leftPanel, rightPanel, width, height)
    }
    return m.renderHorizontalLayout(leftPanel, rightPanel, width, height)
}

// App.go handles complete layout structure
func (a *App) View() string {
    header := a.renderHeader()    // Screen-specific header (logo or title)
    footer := a.renderFooter()    // Dynamic actions based on screen
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

This architecture provides:
- **Clean Separation**: Each layer has clear responsibilities
- **Type Safety**: Compile-time navigation validation
- **State Persistence**: Complete session restoration
- **Responsive Design**: Adaptive to terminal capabilities
- **Resource Awareness**: Real-time estimation and optimization
- **User Experience**: Professional interaction patterns