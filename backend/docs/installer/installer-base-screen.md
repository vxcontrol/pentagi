# BaseScreen Architecture Guide

> Practical guide for implementing new installer screens and migrating existing ones using the BaseScreen architecture.

## 🏗️ **Architecture Overview**

The `BaseScreen` provides a unified foundation for all installer form screens, encapsulating:

- **State Management**: `initialized`, `hasChanges`, `focusedIndex`, `showValues`
- **Form Handling**: `fields []FormField`, viewport management, auto-scrolling
- **Navigation**: Composite ScreenID support, GoBack patterns
- **Layout**: Responsive horizontal/vertical layouts
- **Lists**: Optional dropdown lists with delegates

### **Core Components**

```go
type BaseScreen struct {
    // Dependencies (injected)
    controller *controllers.StateController
    styles     *styles.Styles
    window     *window.Window

    // State
    args         []string
    initialized  bool
    hasChanges   bool
    focusedIndex int
    showValues   bool

    // Form data
    fields       []FormField
    fieldHeights []int

    // UI components
    viewport    viewport.Model
    formContent string

    // Handlers (must be implemented)
    handler     BaseScreenHandler
    listHandler BaseListHandler // optional
}
```

### **Required Interfaces**

```go
type BaseScreenHandler interface {
    BuildForm()
    GetFormTitle() string
    GetHelpContent() string
    HandleSave() error
    HandleReset()
    OnFieldChanged(fieldIndex int, oldValue, newValue string)
    GetFormFields() []FormField
    SetFormFields(fields []FormField)
}

type BaseListHandler interface { // Optional
    GetList() *list.Model
    OnListSelectionChanged(oldSelection, newSelection string)
    GetListHeight() int
}
```

## 🚀 **Creating New Screens**

### **1. Basic Form Screen**

```go
// example_form.go
type ExampleFormModel struct {
    *BaseScreen
    config *controllers.ExampleConfig
}

func NewExampleFormModel(
    controller *controllers.StateController,
    styles *styles.Styles,
    window *window.Window,
    args []string,
) *ExampleFormModel {
    m := &ExampleFormModel{
        config: controller.GetExampleConfig(),
    }

    m.BaseScreen = NewBaseScreen(controller, styles, window, args, m, nil)
    return m
}

// Required interface implementations
func (m *ExampleFormModel) BuildForm() {
    fields := []FormField{}

    // Text field
    apiKeyInput := textinput.New()
    apiKeyInput.Placeholder = "Enter API key"
    apiKeyInput.EchoMode = textinput.EchoPassword
    apiKeyInput.SetValue(m.config.APIKey)

    fields = append(fields, FormField{
        Key:         "api_key",
        Title:       "API Key",
        Description: "Your service API key",
        Required:    true,
        Masked:      true,
        Input:       apiKeyInput,
        Value:       apiKeyInput.Value(),
    })

    // Boolean field
    enabledInput := textinput.New()
    enabledInput.Placeholder = "true/false"
    enabledInput.ShowSuggestions = true
    enabledInput.SetSuggestions([]string{"true", "false"})
    enabledInput.SetValue(fmt.Sprintf("%t", m.config.Enabled))

    fields = append(fields, FormField{
        Key:         "enabled",
        Title:       "Enabled",
        Description: "Enable or disable service",
        Required:    false,
        Masked:      false,
        Input:       enabledInput,
        Value:       enabledInput.Value(),
    })

    m.SetFormFields(fields)
}

func (m *ExampleFormModel) GetFormTitle() string {
    return "Example Service Configuration"
}

func (m *ExampleFormModel) GetHelpContent() string {
    return "Configure your Example service settings here."
}

func (m *ExampleFormModel) HandleSave() error {
    fields := m.GetFormFields()
    for _, field := range fields {
        switch field.Key {
        case "api_key":
            m.config.APIKey = field.Input.Value()
        case "enabled":
            m.config.Enabled = field.Input.Value() == "true"
        }
    }

    if m.config.APIKey == "" {
        return fmt.Errorf("API key is required")
    }

    return m.GetController().UpdateExampleConfig(m.config)
}

func (m *ExampleFormModel) HandleReset() {
    m.config = m.GetController().GetExampleConfig()
    m.BuildForm()
}

func (m *ExampleFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
    // Additional validation logic if needed
}

func (m *ExampleFormModel) GetFormFields() []FormField {
    return m.BaseScreen.fields
}

func (m *ExampleFormModel) SetFormFields(fields []FormField) {
    m.BaseScreen.fields = fields
}

// Update method with field input handling
func (m *ExampleFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            // Handle tab completion for boolean fields
            return m.handleTabCompletion()
        default:
            // Handle field input
            if cmd := m.HandleFieldInput(msg); cmd != nil {
                return m, cmd
            }
        }
    }

    // Base screen handling
    return m.BaseScreen.Update(msg)
}
```

### **2. Screen with List Selection**

```go
// list_form.go
type ListFormModel struct {
    *BaseScreen
    config         *controllers.ListConfig
    selectionList  list.Model
    delegate       *ExampleDelegate
}

func NewListFormModel(...) *ListFormModel {
    m := &ListFormModel{
        config: controller.GetListConfig(),
    }

    m.initializeList()
    m.BaseScreen = NewBaseScreen(controller, styles, window, args, m, m) // Both handlers
    return m
}

func (m *ListFormModel) initializeList() {
    items := []list.Item{
        ExampleOption("Option 1"),
        ExampleOption("Option 2"),
    }

    m.delegate = &ExampleDelegate{
        style: m.GetStyles().FormLabel,
        width: MinMenuWidth - 6,
    }

    m.selectionList = list.New(items, m.delegate, MinMenuWidth-6, 3)
    m.selectionList.SetShowStatusBar(false)
    m.selectionList.SetFilteringEnabled(false)
    m.selectionList.SetShowHelp(false)
    m.selectionList.SetShowTitle(false)
}

// BaseListHandler implementation
func (m *ListFormModel) GetList() *list.Model {
    return &m.selectionList
}

func (m *ListFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
    m.config.SelectedOption = newSelection
    m.BuildForm() // Rebuild form based on selection
}

func (m *ListFormModel) GetListHeight() int {
    return 5
}

func (m *ListFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle list input first
        if cmd := m.HandleListInput(msg); cmd != nil {
            return m, cmd
        }

        // Then field input
        if cmd := m.HandleFieldInput(msg); cmd != nil {
            return m, cmd
        }
    }

    return m.BaseScreen.Update(msg)
}
```

## 🔄 **Migrating Existing Screens**

### **Step 1: Analyze Current Screen**

From existing screen, identify:
- Form fields and their types
- List components (if any)
- Special keyboard handling
- Save/reset logic

### **Step 2: Refactor Structure**

```go
// Before:
type OldFormModel struct {
    controller   *controllers.StateController
    styles       *styles.Styles
    window       *window.Window
    args         []string
    initialized  bool
    hasChanges   bool
    focusedIndex int
    showValues   bool
    fields       []FormField
    viewport     viewport.Model
    formContent  string
    fieldHeights []int
    // ... config-specific fields
}

// After:
type NewFormModel struct {
    *BaseScreen                    // Embedded base screen
    // Only config-specific fields
    config *controllers.Config
    list   list.Model             // If needed
}
```

### **Step 3: Update Constructor**

```go
// Before:
func NewOldFormModel(...) *OldFormModel {
    return &OldFormModel{
        controller: controller,
        styles:     styles,
        // ... lots of boilerplate
    }
}

// After:
func NewNewFormModel(...) *NewFormModel {
    m := &NewFormModel{
        config: controller.GetConfig(),
    }

    // Initialize list if needed
    m.initializeList()

    m.BaseScreen = NewBaseScreen(controller, styles, window, args, m, m)
    return m
}
```

### **Step 4: Implement Required Interfaces**

Move existing methods to interface implementations:

```go
// Move: buildForm() → BuildForm()
// Move: save logic → HandleSave()
// Move: reset logic → HandleReset()
// Move: help content → GetHelpContent()
```

### **Step 5: Remove Redundant Methods**

Delete these methods from migrated screens:
- `getInputWidth()`, `getViewportSize()`, `updateViewport()`
- `focusNext()`, `focusPrev()`, `toggleShowValues()`
- `renderVerticalLayout()`, `renderHorizontalLayout()`
- `ensureFocusVisible()`

## 📋 **Environment Variable Integration**

Follow existing patterns for environment variable handling:

```go
func (m *FormModel) BuildForm() {
    // Track initially set fields for cleanup
    m.initiallySetFields = make(map[string]bool)

    for _, fieldConfig := range m.fieldConfigs {
        envVar, _ := m.GetController().GetVar(fieldConfig.EnvVarName)
        m.initiallySetFields[fieldConfig.Key] = envVar.IsPresent()

        field := m.createFieldFromEnvVar(fieldConfig, envVar)
        fields = append(fields, field)
    }

    m.SetFormFields(fields)
}

func (m *FormModel) HandleSave() error {
    // First pass: Remove cleared fields
    for _, field := range m.GetFormFields() {
        value := strings.TrimSpace(field.Input.Value())
        if value == "" && m.initiallySetFields[field.Key] {
            m.GetController().SetVar(field.EnvVarName, "")
        }
    }

    // Second pass: Save non-empty values
    for _, field := range m.GetFormFields() {
        value := strings.TrimSpace(field.Input.Value())
        if value != "" {
            m.GetController().SetVar(field.EnvVarName, value)
        }
    }

    return nil
}
```

## 🎯 **Navigation Integration**

### **Screen Registration**

Add new screen to navigation system:

```go
// In types.go
const (
    ExampleFormScreen ScreenID = "example_form"
)

// In app.go createModelForScreen()
case ExampleFormScreen:
    return NewExampleFormModel(a.controller, a.styles, a.window, args)
```

### **Navigation Usage**

```go
// Navigate to screen with parameters
return m, func() tea.Msg {
    return NavigationMsg{
        Target: CreateScreenID("example_form", "config_type"),
    }
}

// Return to previous screen
return m, func() tea.Msg {
    return NavigationMsg{GoBack: true}
}
```

## 🔧 **Interface Validation**

Add compile-time interface checks:

```go
// Ensure interfaces are implemented
var _ BaseScreenHandler = (*ExampleFormModel)(nil)
var _ BaseListHandler = (*ListFormModel)(nil)
```

## 📊 **Benefits Summary**

- **Code Reduction**: 50-60% less boilerplate per screen
- **Consistency**: Unified behavior across all forms
- **Maintainability**: Centralized bug fixes and improvements
- **Development Speed**: Faster new screen implementation

## 🎯 **Quick Reference**

### **Screen Types**

1. **Simple Form**: Inherit BaseScreen, implement BaseScreenHandler
2. **Form with List**: Inherit BaseScreen, implement both handlers
3. **Menu Screen**: Use existing patterns without BaseScreen

### **Required Methods**

- `BuildForm()` - Create form fields
- `HandleSave()` - Save configuration with validation
- `HandleReset()` - Reset to defaults
- `GetFormTitle()` - Screen title
- `GetHelpContent()` - Right panel content

### **Optional Methods**

- `OnFieldChanged()` - Real-time validation
- List handler methods (if using lists)

This architecture enables rapid development of new installer screens while maintaining consistency and reducing code duplication.
