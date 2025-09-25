# Charm.sh Advanced Form Patterns

> Comprehensive guide to building sophisticated forms using Charm ecosystem libraries.

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

## 🎯 **Environment Variable Integration Pattern**

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

## 🎯 **Resource Estimation Pattern**

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

## 🎯 **Current Configuration Preview Pattern**

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

## 🎯 **Type-Based Dynamic Forms**

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

## 🏗️ **Form Architecture Best Practices**

### Viewport Usage Patterns

#### **Forms: Permanent Viewport Property**
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

#### **Layout: Temporary Viewport Creation**
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

### Form Field State Management

```go
type FormField struct {
    Key         string
    Title       string
    Description string
    Input       textinput.Model
    Value       string
    Required    bool
    Masked      bool
    Min         int // For integer validation
    Max         int // For integer validation
    EnvVarName  string
}

// Dynamic width application
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

These advanced patterns enable:
- **Smart Validation**: Real-time feedback with user-friendly error messages
- **Resource Awareness**: Live estimation of memory, CPU, or token usage
- **Environment Integration**: Proper handling of defaults, presence detection, and cleanup
- **Type Safety**: Compile-time validation and runtime error handling
- **User Experience**: Auto-completion, formatting, and intuitive navigation