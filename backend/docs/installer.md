# PentAGI Installer Documentation

## Overview

The PentAGI installer provides a robust Terminal User Interface (TUI) for configuring the application. Built using the [Charm](https://charm.sh/) tech stack (bubbletea, lipgloss, bubbles), it implements modern responsive design patterns optimized for terminal environments.

## ⚠️ Development Constraints & TUI Workflow

### Core Workflow Principles

1. **Build-Only Development**: NEVER run TUI apps during development - breaks terminal session
2. **Test Cycle**: Build → Run separately → Return to development session
3. **Debug Output**: All debug MUST go to `logger.Log()` (writes to `log.json`) - never `fmt.Printf`
4. **Development Monitoring**: Use `tail -f log.json` in separate terminal

## Advanced Form Patterns

### **Boolean Field Pattern with Suggestions**

```go
func (m *FormModel) addBooleanField(key, title, description string, envVar loader.EnvVar) {
    input := textinput.New()
    input.Prompt = ""
    input.PlaceholderStyle = m.styles.FormPlaceholder
    input.ShowSuggestions = true
    input.SetSuggestions([]string{"true", "false"})  // Tab completion

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
            m.hasChanges = true
            m.updateFormContent()
        }
    }
}
```

### **Integer Field with Range Validation**

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

    // Add validation range to description
    fullDescription := fmt.Sprintf("%s (Range: %s - %s)",
        description, m.formatBytes(min), m.formatBytes(max))
}

// Real-time validation
func (m *FormModel) validateField(index int) {
    field := &m.fields[index]
    value := field.Input.Value()

    if intVal, err := strconv.Atoi(value); err != nil {
        field.Input.Placeholder = "Enter a valid number or leave empty for default"
    } else {
        // Check ranges with human-readable feedback
        if intVal < min || intVal > max {
            field.Input.Placeholder = fmt.Sprintf("Range: %s - %s",
                m.formatBytes(min), m.formatBytes(max))
        } else {
            field.Input.Placeholder = "" // Clear error
        }
    }
}
```

### **Value Formatting Helpers**

```go
// Universal byte formatting
func (m *FormModel) formatBytes(bytes int) string {
    if bytes >= 1048576 {
        return fmt.Sprintf("%.1fMB", float64(bytes)/1048576)
    } else if bytes >= 1024 {
        return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
    }
    return fmt.Sprintf("%d bytes", bytes)
}

// Universal number formatting
func (m *FormModel) formatNumber(num int) string {
    if num >= 1000000 {
        return fmt.Sprintf("%.1fM", float64(num)/1000000)
    } else if num >= 1000 {
        return fmt.Sprintf("%.1fK", float64(num)/1000)
    }
    return strconv.Itoa(num)
}
```

### **EnvVar Integration Pattern**

```go
// Helper to create field from EnvVar
addFieldFromEnvVar := func(suffix, key, title, description string) {
    envVar, _ := m.controller.GetVar(m.getEnvVarName(suffix))

    // Track if field was initially set for cleanup logic
    m.initiallySetFields[key] = envVar.Value != ""

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
```

### **Field Cleanup Pattern**

```go
func (m *FormModel) saveConfiguration() (tea.Model, tea.Cmd) {
    // First pass: Handle fields that were cleared (remove from environment)
    for _, field := range m.fields {
        value := field.Input.Value()

        // If field was initially set but now empty, remove it
        if value == "" && m.initiallySetFields[field.Key] {
            envVarName := m.getEnvVarName(getEnvSuffixFromKey(field.Key))

            // Remove the environment variable
            if err := m.controller.SetVar(envVarName, ""); err != nil {
                logger.Errorf("[FormModel] SAVE: error clearing %s: %v", envVarName, err)
                return m, nil
            }
            logger.Log("[FormModel] SAVE: cleared %s", envVarName)
        }
    }

    // Second pass: Save only non-empty values
    for _, field := range m.fields {
        value := field.Input.Value()
        if value == "" {
            continue // Skip empty values - use defaults
        }

        // Validate and save
        envVarName := m.getEnvVarName(getEnvSuffixFromKey(field.Key))
        if err := m.controller.SetVar(envVarName, value); err != nil {
            logger.Errorf("[FormModel] SAVE: error setting %s: %v", envVarName, err)
            return m, nil
        }
    }

    m.hasChanges = false
    return m, nil
}
```

### **Token Estimation Pattern**

```go
func (m *FormModel) calculateTokenEstimate() string {
    // Get current form values or defaults
    useQAVal := m.getBoolValueOrDefault("use_qa")
    lastSecBytesVal := m.getIntValueOrDefault("last_sec_bytes")
    maxQABytesVal := m.getIntValueOrDefault("max_qa_bytes")
    keepQASectionsVal := m.getIntValueOrDefault("keep_qa_sections")

    var estimatedBytes int

    // Algorithm-specific calculations
    if m.summarizerType == "assistant" {
        estimatedBytes = keepQASectionsVal * lastSecBytesVal
    } else {
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

// Helper methods to get form values or defaults
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

### **Current Configuration Display Pattern**

```go
func (m *TypesModel) renderTypeInfo() string {
    selectedType := m.types[m.selectedIndex]

    // Helper functions for value retrieval
    getIntValue := func(varName string) int {
        var prefix string
        if selectedType.ID == "assistant" {
            prefix = "ASSISTANT_SUMMARIZER_"
        } else {
            prefix = "SUMMARIZER_"
        }

        envVar, _ := m.controller.GetVar(prefix + varName)
        if envVar.Value != "" {
            if val, err := strconv.Atoi(envVar.Value); err == nil {
                return val
            }
        }
        // Use default if value is empty or invalid
        if val, err := strconv.Atoi(envVar.Default); err == nil {
            return val
        }
        return 0
    }

    // Display current configuration
    sections = append(sections, m.styles.Subtitle.Render("Current Configuration"))
    sections = append(sections, "")

    lastSecBytes := getIntValue("LAST_SEC_BYTES")
    maxBPBytes := getIntValue("MAX_BP_BYTES")
    preserveLast := getBoolValue("PRESERVE_LAST")

    sections = append(sections, fmt.Sprintf("• Last Section Size: %s", formatBytes(lastSecBytes)))
    sections = append(sections, fmt.Sprintf("• Max Body Pair Size: %s", formatBytes(maxBPBytes)))
    sections = append(sections, fmt.Sprintf("• Preserve Last: %t", preserveLast))

    // Type-specific fields
    if selectedType.ID == "general" {
        useQA := getBoolValue("USE_QA")
        sections = append(sections, fmt.Sprintf("• Use QA Pairs: %t", useQA))
    }

    // Token estimation in info panel
    sections = append(sections, "")
    sections = append(sections, m.styles.Subtitle.Render("Token Estimation"))
    sections = append(sections, fmt.Sprintf("• Estimated context size: ~%s tokens", formatNumber(estimatedTokens)))

    return strings.Join(sections, "\n")
}
```

### **Enhanced Localization Pattern**

```go
// Field-specific descriptions with validation hints
const (
    SummarizerFormLastSecBytes     = "Last Section Size (bytes)"
    SummarizerFormLastSecBytesDesc = "Maximum byte size for each preserved conversation section"

    // Enhanced help with practical guidance
    SummarizerFormGeneralHelp = `Balance information depth vs model performance.

Reduce these settings if:
• Using models with ≤64K context (Open Source Reasoning Models)
• Getting "context too long" errors
• Responses become vague or unfocused with long conversations

Key Settings Impact:
• Last Section Size: Larger = more detail, but uses more tokens
• Keep QA Sections: More sections = better continuity, higher token usage

Recommended Adjustments:
• Open Source Reasoning Models: Reduce Last Section to 25-35KB, Keep QA to 1
• OpenAI/Anthropic/Google: Default settings work well`
)
```

### **Type-Based Dynamic Field Generation**

```go
func (m *FormModel) buildForm() {
    // Set type-specific name
    switch m.summarizerType {
    case "general":
        m.typeName = locale.SummarizerTypeGeneralName
    case "assistant":
        m.typeName = locale.SummarizerTypeAssistantName
    }

    // Common fields for all types
    addFieldFromEnvVar("PRESERVE_LAST", "preserve_last", locale.SummarizerFormPreserveLast, locale.SummarizerFormPreserveLastDesc)

    // Type-specific fields
    if m.summarizerType == "general" {
        addFieldFromEnvVar("USE_QA", "use_qa", locale.SummarizerFormUseQA, locale.SummarizerFormUseQADesc)
        addFieldFromEnvVar("SUM_MSG_HUMAN_IN_QA", "sum_human_in_qa", locale.SummarizerFormSumHumanInQA, locale.SummarizerFormSumHumanInQADesc)
    }

    // Common configuration fields
    addFieldFromEnvVar("LAST_SEC_BYTES", "last_sec_bytes", locale.SummarizerFormLastSecBytes, locale.SummarizerFormLastSecBytesDesc)
    // ... additional fields
}
```

These patterns provide a robust foundation for implementing advanced configuration forms with:
- **Type Safety**: Validation at input time
- **User Experience**: Auto-completion, formatting, real-time feedback
- **Resource Awareness**: Token estimation and optimization guidance
- **Environment Integration**: Proper handling of defaults and cleanup
- **Maintainability**: Centralized helpers and consistent patterns

### **Implementation Guidelines for Future Screens**

#### **Langfuse Integration Forms**

**Ready Patterns**: Based on locale constants, implement:
- **Deployment Type Selection**: Embedded/External/Disabled pattern (similar to summarizer types)
- **Conditional Fields**: Show admin fields only for embedded deployment
- **Connection Testing**: Validate external server connectivity
- **Environment Variables**: `LANGFUSE_*` prefix pattern with cleanup

```go
// Implementation pattern
func (m *LangfuseFormModel) buildForm() {
    // Deployment type field (radio-style selection)
    m.addDeploymentTypeField("deployment_type", locale.LangfuseDeploymentType, locale.LangfuseDeploymentTypeDesc)

    // Conditional fields based on deployment type
    if m.deploymentType == "external" {
        m.addFieldFromEnvVar("LANGFUSE_BASE_URL", "base_url", locale.LangfuseBaseURL, locale.LangfuseBaseURLDesc)
        m.addFieldFromEnvVar("LANGFUSE_PROJECT_ID", "project_id", locale.LangfuseProjectID, locale.LangfuseProjectIDDesc)
        m.addFieldFromEnvVar("LANGFUSE_PUBLIC_KEY", "public_key", locale.LangfusePublicKey, locale.LangfusePublicKeyDesc)
        m.addMaskedFieldFromEnvVar("LANGFUSE_SECRET_KEY", "secret_key", locale.LangfuseSecretKey, locale.LangfuseSecretKeyDesc)
    } else if m.deploymentType == "embedded" {
        // Admin configuration for embedded instance
        m.addFieldFromEnvVar("LANGFUSE_ADMIN_EMAIL", "admin_email", locale.LangfuseAdminEmail, locale.LangfuseAdminEmailDesc)
        m.addMaskedFieldFromEnvVar("LANGFUSE_ADMIN_PASSWORD", "admin_password", locale.LangfuseAdminPassword, locale.LangfuseAdminPasswordDesc)
        m.addFieldFromEnvVar("LANGFUSE_ADMIN_NAME", "admin_name", locale.LangfuseAdminName, locale.LangfuseAdminNameDesc)
    }
}
```

#### **Observability Integration Forms**

**Ready Patterns**: Monitoring stack configuration with similar architecture:
- **Deployment Selection**: Embedded/External/Disabled (reuse pattern)
- **External Collector**: OpenTelemetry endpoint configuration
- **Service Selection**: Enable/disable individual monitoring components
- **Resource Estimation**: Calculate monitoring overhead

```go
// Environment variables pattern
func (m *ObservabilityFormModel) getEnvVarName(suffix string) string {
    if m.deploymentType == "external" {
        return "OTEL_" + suffix
    }
    return "OBSERVABILITY_" + suffix
}
```

#### **Security Configuration**
**Potential Patterns**: Based on established architecture:
- **Certificate Management**: File path inputs with validation
- **Access Control**: Boolean enable/disable with role configuration
- **Network Security**: Port ranges, IP allowlists with validation
- **Encryption Settings**: Key generation, algorithm selection

#### **Enhanced Troubleshooting**
**AI-Powered Diagnostics**:
- **System Analysis**: Real-time health checks with recommendations
- **Log Analysis**: Parse error logs and suggest solutions
- **Configuration Validation**: Cross-check settings for conflicts
- **Interactive Fixes**: Guided repair workflows

### **Screen Development Template**

#### **Type Selection Screen Pattern**

```go
type TypesModel struct {
    controller *controllers.StateController
    types      []TypeInfo
    selectedIndex int
    args       []string
}

// Universal type info structure
type TypeInfo struct {
    ID          string
    Name        string
    Description string
}

// Current configuration display (reusable pattern)
func (m *TypesModel) renderCurrentConfiguration(selectedType TypeInfo) string {
    sections = append(sections, m.styles.Subtitle.Render("Current Configuration"))

    // Type-specific value retrieval
    getValue := func(suffix string) string {
        envVar, _ := m.controller.GetVar(m.getEnvVarName(selectedType.ID, suffix))
        if envVar.Value != "" {
            return envVar.Value
        }
        return envVar.Default + " (default)"
    }

    // Display current settings with formatting
    sections = append(sections, fmt.Sprintf("• Setting 1: %s", getValue("SETTING_1")))
    sections = append(sections, fmt.Sprintf("• Setting 2: %s", formatBytes(getIntValue("SETTING_2"))))

    return strings.Join(sections, "\n")
}
```

#### **Form Screen Pattern**

```go
type FormModel struct {
    // Standard form architecture
    controller         *controllers.StateController
    configType        string
    fields            []FormField
    initiallySetFields map[string]bool
    viewport          viewport.Model

    // Pattern-specific additions
    resourceEstimation string
    validationErrors   map[string]string
}

// Universal form building
func (m *FormModel) buildForm() {
    m.fields = []FormField{}
    m.initiallySetFields = make(map[string]bool)

    // Type-specific field generation
    switch m.configType {
    case "type1":
        m.addCommonFields()
        m.addType1SpecificFields()
    case "type2":
        m.addCommonFields()
        m.addType2SpecificFields()
    }

    // Focus and content update
    if len(m.fields) > 0 {
        m.fields[0].Input.Focus()
    }
    m.updateFormContent()
}

// Resource calculation (reusable pattern)
func (m *FormModel) calculateResourceEstimate() string {
    // Get current values from form or defaults
    setting1 := m.getIntValueOrDefault("setting1")
    setting2 := m.getBoolValueOrDefault("setting2")

    // Algorithm-specific calculation
    var estimate int
    if setting2 {
        estimate = setting1 * 2
    } else {
        estimate = setting1
    }

    return fmt.Sprintf("~%s", m.formatNumber(estimate))
}
```

These templates ensure consistency across all future configuration screens while leveraging the proven patterns from summarizer and LLM provider implementations.

## New Implementation Architecture

### **Controller Layer Design**

- **StateController**: Central bridge between TUI forms and state persistence
- **Purpose**: Abstracts environment variable management from UI components
- **Benefits**: Type-safe configuration, automatic validation, dirty state tracking
- **Integration**: All form screens use controller instead of direct state access

### **Adaptive Layout Strategy**

- **Right Panel Hiding**: Main innovation for responsive design
- **Breakpoint Logic**: `contentWidth < (MinMenuWidth + MinInfoWidth + 8)`
- **Graceful Degradation**: Information still accessible, just condensed
- **Performance**: No complex re-rendering, simple layout switching

### **Form Architecture with Bubbles**

- **textinput.Model**: Used for all form inputs with consistent styling
- **Masked Input Toggle**: Ctrl+H to show/hide sensitive values
- **Field Navigation**: Tab/Shift+Tab for keyboard-only navigation
- **Real-time Validation**: Immediate feedback and dirty state tracking
- **Provider-Specific Forms**: Dynamic field generation based on provider type
- **State Persistence**: Composite ScreenIDs with `§` separator (`llm_provider_form§openai`)

### **Composite ScreenID Architecture**

- **Format**: `"screen"` or `"screen§arg1§arg2§..."` for parameterized screens
- **Helper Methods**: `GetScreen()`, `GetArgs()`, `CreateScreenID()`
- **State Persistence**: Complete navigation stack with arguments preserved
- **Benefits**: Type-safe parameter passing, automatic state restoration

```go
// Example: LLM Provider Form with specific provider
targetScreen := CreateScreenID("llm_provider_form", "gemini")
// Results in: "llm_provider_form§gemini"

// Navigation preserves arguments
return NavigationMsg{Target: targetScreen}

// On app restart, user returns to Gemini form, not default OpenAI
```

### **File Organization Pattern**

- **One Model Per File**: `welcome.go`, `eula.go`, `main_menu.go`, etc.
- **Shared Constants**: All in `types.go` for type safety
- **Locale Centralization**: All user-visible text in `locale/locale.go`
- **Controller Separation**: Business logic isolated from presentation

## Architecture & Design Patterns

### 1. Unified App Architecture

**Central Orchestrator (`app.go`)**:
- **Navigation Management**: Stack-based navigation with step persistence
- **Screen Lifecycle**: Model creation, initialization, and cleanup
- **Unified Layout**: Header and footer rendering for all screens
- **Global Event Handling**: ESC, Ctrl+C, window resize
- **Dimension Management**: Terminal size distribution to models

```go
// UNIFIED RENDERING - All screens follow this pattern:
func (a *App) View() string {
    header := a.renderHeader()  // Screen-specific header
    footer := a.renderFooter()  // Dynamic footer with actions
    content := a.currentModel.View()  // Model provides content only

    // App.go calculates and enforces layout constraints
    contentHeight := max(height - headerHeight - footerHeight, 0)
    contentArea := a.styles.Content.Height(contentHeight).Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}
```

### 2. Navigation & State Management

#### Navigation Rules (Universal)

- **ESC Behavior**: ALWAYS returns to Welcome screen from any screen (never nested back navigation)
- **Type Safety**: Use `ScreenID` with `CreateScreenID()` for parameterized screens
- **Composite Support**: Screens can carry arguments via `§` separator
- **State Persistence**: Complete navigation stack with arguments preserved
- **EULA Consent**: Check `GetEulaConsent()` on Welcome→EULA transition, call `SetEulaConsent()` on acceptance

```go
// Type-safe navigation structure with composite support
type NavigationMsg struct {
    Target ScreenID  // Can be simple or composite
    GoBack bool
}

type ScreenID string
const (
    WelcomeScreen         ScreenID = "welcome"
    EULAScreen            ScreenID = "eula"
    MainMenuScreen        ScreenID = "main_menu"
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

// Navigation with parameters
targetScreen := CreateScreenID("llm_provider_form", "anthropic")
return NavigationMsg{Target: targetScreen}

// Universal ESC implementation
case "esc":
    if a.navigator.Current().GetScreen() != string(models.WelcomeScreen) {
        a.navigator.stack = []models.ScreenID{models.WelcomeScreen}
        a.navigator.stateManager.SetStack([]string{"welcome"})
        a.currentModel = a.createModelForScreen(models.WelcomeScreen, nil)
        return a, a.currentModel.Init()
    }
```

#### State Integration

- `state.State` remains authoritative for env variables
- Controllers translate between TUI models and state operations
- Complete state reset in `Init()` for predictable behavior

### 3. Layout & Responsive Design

#### Constants & Breakpoints

```go
// Layout Constants
const (
    SmallScreenThreshold = 30    // Height threshold for viewport mode
    MinTerminalWidth = 80        // Minimum width for horizontal layout
    MinPanelWidth = 25           // Panel width constraints
    WelcomeHeaderHeight = 8      // Fixed by ASCII Art Logo (8 lines)
    EULAHeaderHeight = 3         // Title + subtitle + spacing
    FooterHeight = 1             // Always 1 line with background approach
)
```

#### Responsive Breakpoints

- **Small screens**: < 30 rows → viewport mode for scrolling
- **Large screens**: ≥ 30 rows → normal layout mode
- **Narrow terminals**: < 80 cols → vertical stacking
- **Wide terminals**: ≥ 80 cols → horizontal panel layout

#### Height Control (CRITICAL)

```go
// ❌ WRONG - Height() sets MINIMUM height, can expand
style.Height(1).Border(lipgloss.Border{Top: true})

// ✅ CORRECT - Background approach ensures exactly 1 line
style.Background(borderColor).Foreground(textColor).Padding(0,1,0,1)
```

### 4. Footer & Header Systems

#### Unified Footer Strategy

**Background Approach (Production-Ready)**:
- Always exactly 1 line regardless of terminal size
- Modern appearance with background color
- Reliable height calculations
- Dynamic actions based on screen state

```go
// Footer pattern implementation
actions := locale.BuildCommonActions()
if specificCondition {
    actions = append(actions, locale.SpecificAction)
}
footerText := strings.Join(actions, locale.NavSeparator)

return lipgloss.NewStyle().
    Width(width).
    Background(styles.Border).
    Foreground(styles.Foreground).
    Padding(0, 1, 0, 1).
    Render(footerText)
```

#### Header Strategy

- **Welcome Screen**: ASCII Art Logo (8 lines height)
- **Other Screens**: Text title with consistent styling
- **Responsive**: Always present, managed by `app.go`

### 5. Scrolling & Input Handling

#### Modern Scroll Methods

```go
viewport.ScrollUp(1)     // Replaces deprecated LineUp()
viewport.ScrollDown(1)   // Replaces deprecated LineDown()
viewport.ScrollLeft(2)   // Horizontal scroll (2 steps for faster navigation)
viewport.ScrollRight(2)  // Horizontal scroll (2 steps for faster navigation)
```

#### Essential Key Handling

- **↑/↓**: Vertical scrolling (1 line per press)
- **←/→**: Horizontal scrolling (2 steps per press for faster navigation)
- **PgUp/PgDn**: Page-level scrolling
- **Home/End**: Jump to beginning/end

### 6. Content & Resource Management

#### Shared Renderer (Prevents Freezing)
```go
// Single renderer instance in styles.New()
type Styles struct {
    renderer *glamour.TermRenderer
    width    int
    height   int
}

// Usage pattern
rendered, err := m.styles.GetRenderer().Render(markdown)
if err != nil {
    // Fallback to plain text
    rendered = fmt.Sprintf("# Content\n\n%s\n\n*Render error: %v*", content, err)
}
```

#### Content Loading Strategy

- Single renderer instance prevents glamour freezing
- Reset model state completely in `Init()` for clean transitions
- Force view update after content loading with no-op command
- Use embedded files via `files.GetContent()` - handles working directory variations

### 7. Component Architecture

#### Component Types

1. **Welcome Screen**: ASCII art, system checks, info display
2. **EULA Screen**: Markdown viewer with scroll-to-accept
3. **Menu Screen**: Main navigation with dynamic availability
4. **Form Screens**: Configuration input with validation
5. **Status Screens**: Progress and result display

#### Key Components

- **StatusIndicator**: System check results with green checkmarks/red X's
- **MarkdownViewer**: EULA and help text display with scroll support
- **FormController**: Bridges huh forms with state package
- **MenuList**: Dynamic menu with availability checking

### 8. Localization & Styling

#### Localization Structure

```
wizard/locale/
└── locale.go    # All user-visible text constants
```

**Naming Convention**:
- `Welcome*`, `EULA*`, `Menu*`, `LLM*`, `Checks*` - Screen-specific
- `Nav*`, `Status*`, `Error*`, `UI*` - Functional prefixes

#### Styles Centralization

- Single styles instance with shared renderer and dimensions
- Prevents glamour freezing, centralizes terminal size management
- All models access dimensions via `m.styles.GetSize()`

## Development Guidelines

### Screen Model Requirements

#### Required Implementation Pattern

```go
// REQUIRED: State reset in Init()
func (m *Model) Init() tea.Cmd {
    logger.Log("[Model] INIT")
    m.content = ""
    m.ready = false
    // ... reset ALL state
    return m.loadContent
}

// REQUIRED: Dimension handling via styles
func (m *Model) updateViewport() {
    width, height := m.styles.GetSize()
    if width <= 0 || height <= 0 {
        return
    }
    // ... viewport logic
}

// REQUIRED: Adaptive layout methods
func (m *Model) isVerticalLayout() bool {
    return m.styles.GetWidth() < MinTerminalWidth
}
```

### Screen Development Checklist

**For each new screen:**
- [ ] Type-safe `ScreenID` defined in `types.go`
- [ ] State reset in `Init()` method with logger
- [ ] Dimension handling via `m.styles.GetSize()`
- [ ] Modern `Scroll*` methods for navigation
- [ ] Arrow key handling (↑/↓/←/→) with 2-step horizontal
- [ ] Background footer approach using locale helpers
- [ ] Shared renderer from `styles.GetRenderer()`
- [ ] ESC navigation to Welcome screen
- [ ] Logger integration for debug output

### Code Style Guidelines

#### Compact vs Expanded Style

```go
// ✅ Compact where appropriate:
leftWidth = max(leftWidth, MinPanelWidth)
return lipgloss.NewStyle().Width(width).Padding(0, 2, 0, 2).Render(content)

// ✅ Expanded where needed:
coreChecks := []struct {
    label string
    value bool
}{
    {locale.CheckEnvironmentFile, m.checker.EnvFileExists},
    {locale.CheckDockerAPI, m.checker.DockerApiAccessible},
}
```

#### Comment Guidelines

- Comments explain **why** and **how**, not **what**
- Place comments where code might raise questions about business logic
- Avoid redundant comments that repeat obvious code behavior

## Recent Fixes & Improvements

### ✅ **Composite ScreenID Navigation System**
**Problem**: Need to preserve selected menu items and provider selections across navigation
**Solution**: Implemented composite ScreenIDs with `§` separator for parameter passing

**Features**:
```go
// Composite ScreenID examples
"main_menu§llm_providers"           // Main menu with "llm_providers" selected
"llm_providers§gemini"              // Providers list with "gemini" selected
"llm_provider_form§anthropic"       // Form for "anthropic" provider
```

**Benefits**:
- Type-safe parameter passing via `GetScreen()`, `GetArgs()`, `CreateScreenID()`
- Automatic state restoration - user returns to exact selection after ESC
- Clean navigation stack with full context preservation
- Extensible for multiple arguments per screen

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

**Based on research**: [BubbleTea viewport best practices](https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport) and [Perplexity guidance on form scrolling](https://www.inngest.com/blog/interactive-clis-with-bubbletea)

**Key Features**:
- **Auto-scroll**: Focused field automatically stays visible
- **Smart positioning**: Calculates field heights for precise scroll positioning
- **Seamless navigation**: Tab/Shift+Tab scroll form as needed
- **No extra hotkeys**: Uses existing navigation keys

**Technical Implementation**:
```go
// Auto-scroll on field focus change
func (m *Model) ensureFocusVisible() {
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

## Error Handling & Performance

### Error Handling Strategy

**Graceful degradation with user-friendly messages**:
1. System check failures: Show specific resolution steps
2. Form validation: Real-time feedback with clear messaging
3. State persistence errors: Allow retry with explanation
4. Network issues: Offer offline/manual alternatives

### Performance Considerations

**Lazy loading approach**:
- System checks run asynchronously after welcome screen loads
- Markdown content loaded on-demand when screens are accessed
- Form validation debounced to avoid excessive state updates

## Common Pitfalls & Solutions

### Content Loading Issues

**Problem**: "Loading EULA" state persists, content doesn't appear
**Solutions**:
1. **Multiple Path Fallback**: Try embedded FS first, then direct file access
2. **State Reset**: Always reset model state in `Init()` for clean loading
3. **No ClearScreen**: Avoid `tea.ClearScreen` during navigation
4. **Force View Update**: Return no-op command after content loading

### Layout Consistency Issues

**Problem**: Layout breaks on terminal resize
**Solution**: Always account for actual footer height (1 line)

```go
// Consistent height calculation across all screens
headerHeight := 3 // Fixed based on content
footerHeight := 1 // Background approach always 1 line
contentHeight := m.height - headerHeight - footerHeight
```

### Common Mistakes to Avoid
- Using `tea.ClearScreen` in navigation
- Border-based footer (height inconsistency)
- String-based navigation messages
- Creating new glamour renderer instances
- Forgetting state reset in `Init()`
- Using `fmt.Printf` for debug output
- Deprecated `Line*` scroll methods

## Technology Stack

- **bubbletea**: Core TUI framework using Model-View-Update pattern
- **lipgloss**: Styling and layout engine for visual presentation
- **bubbles**: Component library for interactive elements (list, textinput, viewport)
- **huh**: Form builder for structured input collection (future screens)
- **glamour**: Markdown rendering with single shared instance
- **logger**: Custom file-based logging for TUI-safe development

## ✅ **Production Architecture Implementation**

### **Completed Form System Architecture**

#### **Form Model Pattern (llm_provider_form.go)**

```go
type LLMProviderFormModel struct {
    controller *controllers.StateController
    styles     *styles.Styles
    window     *window.Window

    // Form state
    providerID   string
    fields       []FormField
    focusedIndex int
    showValues   bool
    hasChanges   bool
    args         []string // From composite ScreenID

    // Permanent viewport for scroll state
    viewport     viewport.Model
    formContent  string
    fieldHeights []int
}
```

**Key Implementation Decisions**:
- **Args-based Construction**: `NewLLMProviderFormModel(controller, styles, window, args)`
- **Permanent Viewport**: Form viewport as struct property to preserve scroll state
- **Auto-completion**: Tab key triggers suggestion completion for boolean fields
- **GoBack Navigation**: `return NavigationMsg{GoBack: true}` prevents navigation loops

#### **Navigation Hotkeys (Production Pattern)**

```go
// Modern form navigation
case "down":    // ↓: Next field + auto-scroll
case "up":      // ↑: Previous field + auto-scroll
case "tab":     // Tab: Complete suggestion (true/false for booleans)
case "ctrl+h":  // Ctrl+H: Toggle show/hide masked values
case "ctrl+s":  // Ctrl+S: Save configuration
case "enter":   // Enter: Save and return via GoBack
```

**Important**: Tab navigation replaced with suggestion completion. Field navigation uses ↑/↓ only.

### **Adaptive Layout System**

#### **Layout Constants (Production Values)**

```go
const (
    MinMenuWidth  = 38  // Minimum left panel width
    MaxMenuWidth  = 66  // Maximum left panel width (prevents too wide forms)
    MinInfoWidth  = 34  // Minimum right panel width
    PaddingWidth  = 8   // Total horizontal padding
    PaddingHeight = 2   // Vertical padding
)
```

#### **Two-Column Layout Implementation**

```go
func (m *Model) renderHorizontalLayout(leftPanel, rightPanel string, width, height int) string {
    leftWidth, rightWidth := MinMenuWidth, MinInfoWidth
    extraWidth := width - leftWidth - rightWidth - PaddingWidth

    // Distribute extra space intelligently
    if extraWidth > 0 {
        leftWidth = min(leftWidth+extraWidth/2, MaxMenuWidth)  // Cap at MaxMenuWidth
        rightWidth = width - leftWidth - PaddingWidth/2
    }

    leftStyled := lipgloss.NewStyle().Width(leftWidth).Padding(0, 2, 0, 2).Render(leftPanel)
    rightStyled := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(2).Render(rightPanel)

    // Final layout viewport (temporary)
    viewport := viewport.New(width, height-PaddingHeight)
    viewport.SetContent(lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled))
    return viewport.View()
}
```

#### **Content Hiding Strategy**

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

### **Composite ScreenID Navigation System**

#### **ScreenID Argument Packaging**

```go
// Navigation with selection preservation
func (m *MainMenuModel) handleMenuSelection() (tea.Model, tea.Cmd) {
    selectedItem := m.getSelectedItem()

    return m, func() tea.Msg {
        return NavigationMsg{
            Target: CreateScreenID(string(targetScreen), selectedItem.ID),
        }
    }
}

// Result: "llm_providers§openai" -> llm_providers screen with "openai" pre-selected
```

#### **Args-Based Model Construction**

```go
// No SetSelected* methods needed - selection from constructor
func NewLLMProvidersModel(
    controller *controllers.StateController, styles *styles.Styles,
    window *window.Window, args []string,
) *LLMProvidersModel {
    return &LLMProvidersModel{
        controller: controller,
        args:       args,  // Selection restored from args in Init()
    }
}

func (m *LLMProvidersModel) Init() tea.Cmd {
    // Automatic selection restoration from args[1]
    if len(m.args) > 1 && m.args[1] != "" {
        for i, provider := range m.providers {
            if provider.ID == m.args[1] {
                m.selectedIndex = i
                break
            }
        }
    }
    return nil
}
```

#### **Navigation Stack Management**

**Stack Example**: `["main_menu§llm_providers", "llm_providers§openai", "llm_provider_form§openai"]`

- **Forward Navigation**: Pushes composite ScreenID with arguments
- **Back Navigation**: `GoBack: true` pops current screen, returns to previous with preserved selection
- **No Navigation Loops**: GoBack pattern prevents infinite stack growth

### **Viewport Usage Patterns**

#### **Forms: Permanent Viewport Property**

```go
// ✅ CORRECT: Form viewport as struct property
type FormModel struct {
    viewport viewport.Model  // Preserves scroll position across updates
}

func (m *FormModel) ensureFocusVisible() {
    // Auto-scroll to focused field
    focusY := m.calculateFieldPosition(m.focusedIndex)
    if focusY < m.viewport.YOffset {
        m.viewport.YOffset = focusY
    }
    if focusY+m.fieldHeights[m.focusedIndex] >= offset+visibleRows {
        m.viewport.YOffset = focusY + m.fieldHeights[m.focusedIndex] - visibleRows + 1
    }
}
```

#### **Layout: Temporary Viewport Creation**

```go
// ✅ CORRECT: Layout viewport created for rendering only
func (m *Model) renderHorizontalLayout(left, right string, width, height int) string {
    content := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)

    vp := viewport.New(width, height-PaddingHeight)  // Temporary
    vp.SetContent(content)
    return vp.View()
}
```

### **Dynamic Form Field Architecture**

#### **Field Configuration Pattern**

```go
// Clean input setup without fixed width
func (m *FormModel) addInputField(fieldType string) {
    input := textinput.New()
    input.Prompt = ""  // Clean appearance
    input.PlaceholderStyle = m.styles.FormPlaceholder

    // Width set dynamically during updateFormContent()
    // NOT set here: input.Width = 50

    if fieldType == "boolean" {
        input.ShowSuggestions = true
        input.SetSuggestions([]string{"true", "false"})
    }
}
```

#### **Dynamic Width Calculation**

```go
func (m *FormModel) getInputWidth() int {
    viewportWidth, _ := m.getViewportSize()
    inputWidth := viewportWidth - 6  // Standard padding
    if m.isVerticalLayout() {
        inputWidth = viewportWidth - 4  // Tighter in vertical mode
    }
    return inputWidth
}

// Applied during form content update
func (m *FormModel) updateFormContent() {
    inputWidth := m.getInputWidth()

    for i, field := range m.fields {
        field.Input.Width = inputWidth - 3  // Account for border/cursor
        field.Input.SetValue(field.Input.Value())  // Trigger width update

        inputStyle := m.styles.FormInput.Width(inputWidth)
        if i == m.focusedIndex {
            inputStyle = inputStyle.BorderForeground(styles.Primary)
        }

        renderedInput := inputStyle.Render(field.Input.View())
        sections = append(sections, renderedInput)
    }
}
```

### **Provider Configuration Architecture**

#### **Simplified Status Model**

```go
// ✅ PRODUCTION: Single status field
type ProviderInfo struct {
    ID          string
    Name        string
    Description string
    Configured  bool    // Single status - has required fields
}

// Status check via controller
configs := m.controller.GetLLMProviders()
provider := ProviderInfo{
    Configured: configs["openai"].Configured,  // Controller determines status
}
```

**Removed**: Dual `Configured`/`Enabled` status - controller handles enable/disable logic internally.

#### **Provider-Specific Field Sets**
- **OpenAI/Anthropic/Gemini**: Base URL + API Key
- **AWS Bedrock**: Region + Access Key + Secret Key + Base URL (optional)
- **Ollama**: Base URL + Config Path
- **Custom**: Base URL + API Key + Model + Config Path + Legacy Reasoning (boolean with suggestions)

### **Screen Architecture (App.go Integration)**

#### **Content Area Responsibility**

```go
// ✅ Screen models handle ONLY content area
func (m *Model) View() string {
    leftPanel := m.renderForm()
    rightPanel := m.renderHelp()

    // Adaptive layout decision
    if m.isVerticalLayout() {
        return m.renderVerticalLayout(leftPanel, rightPanel, width, height)
    }
    return m.renderHorizontalLayout(leftPanel, rightPanel, width, height)
}
```

#### **App.go Layout Management**

```go
// App.go handles complete layout structure
func (a *App) View() string {
    header := a.renderHeader()    // Screen-specific (logo or title)
    footer := a.renderFooter()    // Dynamic actions based on screen
    content := a.currentModel.View()  // Content only from model

    contentWidth, contentHeight := a.window.GetContentSize()
    contentArea := a.styles.Content.
        Width(contentWidth).
        Height(contentHeight).
        Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}
```

### **Navigation Anti-Patterns & Solutions**

#### **❌ Common Mistakes**

```go
// ❌ WRONG: Direct navigation creates loops
func (m *FormModel) saveAndReturn() (tea.Model, tea.Cmd) {
    m.saveConfiguration()
    return m, func() tea.Msg {
        return NavigationMsg{Target: LLMProvidersScreen}  // Loop!
    }
}

// ❌ WRONG: Separate SetSelected methods
func (m *Model) SetSelectedProvider(providerID string) {
    // Complexity - removed in favor of args-based construction
}

// ❌ WRONG: Fixed input widths
input.Width = 50  // Breaks responsive design
```

#### **✅ Correct Patterns**

```go
// ✅ CORRECT: GoBack navigation
func (m *FormModel) saveAndReturn() (tea.Model, tea.Cmd) {
    if err := m.saveConfiguration(); err != nil {
        return m, nil  // Stay on form if save fails
    }
    return m, func() tea.Msg {
        return NavigationMsg{GoBack: true}  // Return to previous screen
    }
}

// ✅ CORRECT: Args-based selection
func NewModel(..., args []string) *Model {
    selectedIndex := 0
    if len(args) > 1 && args[1] != "" {
        // Set selection from args during construction
        for i, item := range items {
            if item.ID == args[1] {
                selectedIndex = i
                break
            }
        }
    }
    return &Model{selectedIndex: selectedIndex, args: args}
}

// ✅ CORRECT: Dynamic input sizing
func (m *FormModel) updateFormContent() {
    inputWidth := m.getInputWidth()  // Calculate based on available space
    field.Input.Width = inputWidth - 3
}
```
