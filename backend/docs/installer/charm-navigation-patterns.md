# Charm.sh Navigation Patterns

> Comprehensive guide to implementing robust navigation systems in TUI applications.

## 🎯 **Type-Safe Navigation with Composite ScreenIDs**

### **Composite ScreenID Pattern**
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

### **Usage Examples**
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

### **State Persistence**
```go
// Stack automatically preserves composite IDs
navigator.Push(CreateScreenID("llm_provider_form", "gemini"))

// State contains: ["welcome", "main_menu", "llm_providers", "llm_provider_form§gemini"]
// On restore: user returns to Gemini provider form, not default OpenAI
```

## 🎯 **Navigation Message Pattern**

### **NavigationMsg Structure**
```go
type NavigationMsg struct {
    Target ScreenID  // Can be simple or composite
    GoBack bool      // Return to previous screen
    Data   any       // Optional data to pass
}

// Type-safe constants
type ScreenID string
const (
    WelcomeScreen         ScreenID = "welcome"
    EULAScreen           ScreenID = "eula"
    MainMenuScreen       ScreenID = "main_menu"
    LLMProviderFormScreen ScreenID = "llm_provider_form"
)
```

### **Navigation Commands**
```go
// Simple navigation
return m, func() tea.Msg {
    return NavigationMsg{Target: EULAScreen}
}

// Navigation with parameters
return m, func() tea.Msg {
    return NavigationMsg{Target: CreateScreenID("llm_provider_form", "openai")}
}

// Go back to previous screen
return m, func() tea.Msg {
    return NavigationMsg{GoBack: true}
}

// Navigation with data passing
return m, func() tea.Msg {
    return NavigationMsg{
        Target: CreateScreenID("config_form", "database"),
        Data:   ConfigData{Type: "database", Settings: currentSettings},
    }
}
```

## 🎯 **Navigator Implementation**

### **Navigation Stack Management**
```go
type Navigator struct {
    stack        []ScreenID
    stateManager StateManager
}

func NewNavigator(stateManager StateManager) *Navigator {
    return &Navigator{
        stack:        []ScreenID{WelcomeScreen},
        stateManager: stateManager,
    }
}

func (n *Navigator) Push(screenID ScreenID) {
    n.stack = append(n.stack, screenID)
    n.persistState()
}

func (n *Navigator) Pop() ScreenID {
    if len(n.stack) <= 1 {
        return n.stack[0] // Can't pop last screen
    }

    popped := n.stack[len(n.stack)-1]
    n.stack = n.stack[:len(n.stack)-1]
    n.persistState()
    return popped
}

func (n *Navigator) Current() ScreenID {
    if len(n.stack) == 0 {
        return WelcomeScreen
    }
    return n.stack[len(n.stack)-1]
}

func (n *Navigator) Replace(screenID ScreenID) {
    if len(n.stack) == 0 {
        n.stack = []ScreenID{screenID}
    } else {
        n.stack[len(n.stack)-1] = screenID
    }
    n.persistState()
}

func (n *Navigator) persistState() {
    stringStack := make([]string, len(n.stack))
    for i, screenID := range n.stack {
        stringStack[i] = string(screenID)
    }
    n.stateManager.SetStack(stringStack)
}

func (n *Navigator) RestoreState() {
    stringStack := n.stateManager.GetStack()
    if len(stringStack) == 0 {
        n.stack = []ScreenID{WelcomeScreen}
        return
    }

    n.stack = make([]ScreenID, len(stringStack))
    for i, s := range stringStack {
        n.stack[i] = ScreenID(s)
    }
}
```

## 🎯 **Universal ESC Behavior**

### **Global Navigation Handling**
```go
func (a *App) handleGlobalNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "esc":
        // Universal ESC: ALWAYS returns to Welcome screen
        if a.navigator.Current().GetScreen() != string(WelcomeScreen) {
            a.navigator.stack = []ScreenID{WelcomeScreen}
            a.navigator.persistState()
            a.currentModel = a.createModelForScreen(WelcomeScreen, nil)
            return a, a.currentModel.Init()
        }

    case "ctrl+c":
        // Global quit
        return a, tea.Quit
    }
    return a, nil
}

// In main Update loop
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle global navigation first
        if newModel, cmd := a.handleGlobalNavigation(msg); cmd != nil {
            return newModel, cmd
        }

        // Then pass to current model
        var cmd tea.Cmd
        a.currentModel, cmd = a.currentModel.Update(msg)
        return a, cmd

    case NavigationMsg:
        return a.handleNavigationMsg(msg)
    }

    // Delegate to current model
    var cmd tea.Cmd
    a.currentModel, cmd = a.currentModel.Update(msg)
    return a, cmd
}
```

## 🎯 **Navigation Message Handling**

### **App-Level Navigation**
```go
func (a *App) handleNavigationMsg(msg NavigationMsg) (tea.Model, tea.Cmd) {
    if msg.GoBack {
        if len(a.navigator.stack) > 1 {
            a.navigator.Pop()
            currentScreen := a.navigator.Current()
            a.currentModel = a.createModelForScreen(currentScreen, msg.Data)
            return a, a.currentModel.Init()
        }
        // Can't go back further, stay on current screen
        return a, nil
    }

    // Forward navigation
    a.navigator.Push(msg.Target)
    a.currentModel = a.createModelForScreen(msg.Target, msg.Data)
    return a, a.currentModel.Init()
}

func (a *App) createModelForScreen(screenID ScreenID, data any) tea.Model {
    baseScreen := screenID.GetScreen()
    args := screenID.GetArgs()

    switch ScreenID(baseScreen) {
    case WelcomeScreen:
        return NewWelcomeModel(a.controller, a.styles, a.window)

    case EULAScreen:
        return NewEULAModel(a.controller, a.styles, a.window)

    case MainMenuScreen:
        selectedItem := ""
        if len(args) > 0 {
            selectedItem = args[0]
        }
        return NewMainMenuModel(a.controller, a.styles, a.window, []string{selectedItem})

    case LLMProviderFormScreen:
        providerID := "openai"
        if len(args) > 0 {
            providerID = args[0]
        }
        return NewLLMProviderFormModel(a.controller, a.styles, a.window, []string{providerID})

    default:
        // Fallback to welcome screen
        return NewWelcomeModel(a.controller, a.styles, a.window)
    }
}
```

## 🎯 **Args-Based Model Construction**

### **Model Constructor Pattern**
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

### **Selection Preservation Pattern**
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

## 🎯 **Data Passing Pattern**

### **Structured Data Transfer**
```go
// Define data structures for navigation
type FormData struct {
    ProviderID string
    Settings   map[string]string
}

type ConfigData struct {
    Type     string
    Settings map[string]interface{}
}

// Pass data through navigation
func (m *MenuModel) openConfiguration() (tea.Model, tea.Cmd) {
    return m, func() tea.Msg {
        return NavigationMsg{
            Target: CreateScreenID("config_form", "database"),
            Data: ConfigData{
                Type: "database",
                Settings: m.getCurrentSettings(),
            },
        }
    }
}

// Receive data in target model
func NewConfigFormModel(
    controller *controllers.StateController, styles *styles.Styles,
    window *window.Window, args []string, data any,
) *ConfigFormModel {
    configType := "default"
    if len(args) > 0 {
        configType = args[0]
    }

    var settings map[string]interface{}
    if configData, ok := data.(ConfigData); ok {
        settings = configData.Settings
    }

    return &ConfigFormModel{
        configType: configType,
        settings:   settings,
        // ...
    }
}
```

## 🎯 **Navigation Anti-Patterns & Solutions**

### **❌ Common Mistakes**
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

// ❌ WRONG: String-based navigation (typo-prone)
return NavigationMsg{Target: "main_menu"}

// ❌ WRONG: Manual string concatenation for arguments
return NavigationMsg{Target: ScreenID("llm_provider_form/openai")}
```

### **✅ Correct Patterns**
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

// ✅ CORRECT: Type-safe constants
return NavigationMsg{Target: MainMenuScreen}

// ✅ CORRECT: Composite ScreenID with helper
return NavigationMsg{Target: CreateScreenID("llm_provider_form", "openai")}
```

## 🎯 **Navigation Stack Examples**

### **Typical Navigation Flow**
```go
// Stack progression example:
// 1. Start: ["welcome"]
// 2. Continue: ["welcome", "main_menu"]
// 3. LLM Providers: ["welcome", "main_menu§llm_providers", "llm_providers"]
// 4. OpenAI Form: ["welcome", "main_menu§llm_providers", "llm_providers§openai", "llm_provider_form§openai"]
// 5. GoBack: ["welcome", "main_menu§llm_providers", "llm_providers§openai"]
// 6. ESC: ["welcome"]

func demonstrateNavigation() {
    nav := NewNavigator(stateManager)

    // Initial state
    current := nav.Current() // "welcome"

    // Navigate to main menu
    nav.Push(CreateScreenID("main_menu", "llm_providers"))
    current = nav.Current() // "main_menu§llm_providers"

    // Navigate to providers list
    nav.Push(CreateScreenID("llm_providers", "openai"))
    current = nav.Current() // "llm_providers§openai"

    // Navigate to form
    nav.Push(CreateScreenID("llm_provider_form", "openai"))
    current = nav.Current() // "llm_provider_form§openai"

    // Go back
    nav.Pop()
    current = nav.Current() // "llm_providers§openai"

    // ESC to home (clear stack)
    nav.stack = []ScreenID{WelcomeScreen}
    current = nav.Current() // "welcome"
}
```

### **State Restoration**
```go
// On app restart, navigation stack is restored with all parameters
func (a *App) initializeNavigation() {
    a.navigator.RestoreState()

    // User returns to exact screen with preserved selection
    // e.g., "llm_provider_form§anthropic" restores Anthropic form
    currentScreen := a.navigator.Current()
    a.currentModel = a.createModelForScreen(currentScreen, nil)
}
```

This navigation system provides:
- **Type Safety**: Compile-time validation of screen IDs
- **Parameter Preservation**: Arguments maintained across navigation
- **Stack Management**: Proper back navigation without loops
- **State Persistence**: Complete navigation state restoration
- **Universal Behavior**: Consistent ESC and global navigation