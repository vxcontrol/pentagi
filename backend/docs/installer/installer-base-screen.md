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
    BuildForm() tea.Cmd
    GetFormSummary() string
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

> The two examples above predate two changes to the base: `NewBaseScreen` no longer takes an `args []string` parameter, and `BuildForm()` now returns a `tea.Cmd`.
> The two screens below are written against the current shape.

### **3. Document Screen (viewport, no BaseScreen)**

`wizard/models/update_overview.go` is the shape to copy when a screen shows *text* rather than fields. It does not embed `BaseScreen` at all — it implements `BaseScreenModel` directly, the way `eula.go` does.

Everything `BaseScreen` contributes (fields, focus, save/reset, the form/help split) is dead weight for a document, and the one thing a document needs — a full-width scrolling viewport over rendered markdown — `BaseScreen` does not provide.

```go
type UpdateOverviewModel struct {
    styles     styles.Styles
    window     window.Window
    controller controller.Controller

    viewport viewport.Model

    source     string // the markdown
    content    string // that markdown, rendered at renderedAt columns
    renderedAt int
    ready      bool
    scrolled   bool
}

// only the model interface: there is no form, so there is no handler
var _ BaseScreenModel = (*UpdateOverviewModel)(nil)
```

**Keep the markdown and its rendered form as two separate fields.** A resize re-renders *from the source*. Re-wrapping the already-rendered string would wrap the ANSI escape sequences glamour wrote into it, and the screen fills with fragments of colour codes.

**Re-render only when the wrap width actually changed:**

```go
func (m *UpdateOverviewModel) updateViewport() {
    contentWidth, contentHeight := m.window.GetContentSize()
    if contentWidth <= 0 || contentHeight <= 0 || m.source == "" {
        return
    }

    if m.content == "" || m.renderedAt != contentWidth {
        rendered, err := m.styles.GetRendererForWidth(contentWidth).Render(m.source)
        // ... on error, fall back to the raw markdown plus the error
        m.content = rendered
        m.renderedAt = contentWidth
    }

    if !m.ready {
        m.viewport = viewport.New(contentWidth, contentHeight)
        m.ready = true
    } else {
        m.viewport.Width, m.viewport.Height = contentWidth, contentHeight
    }

    m.viewport.SetContent(m.content)
    m.updateScrollStatus()
}
```

Dragging a window border delivers one `tea.WindowSizeMsg` per column and glamour is not cheap, so `renderedAt` is what stops the whole changelog being re-rendered for a resize that cannot change how it wraps — a height-only drag, or one of the resizes `registry.HandleMsg` forwards to *every* registered screen whether or not the user is looking at it. The comparison is against the raw content width, so widths outside the clamp band below still re-render (into identical output); the renderer itself is memoised in either case.

**Create the viewport on the first real size, never in the constructor.**  
`window.GetContentSize()` is 0×0 until the first `WindowSizeMsg` arrives, and a viewport built at that size shows nothing afterwards. `ready` is the flag that says the viewport exists, and each of its readers does something different: `View()` falls back to the loading text, `GetScrollInfo()` returns zeroes rather than measure an unsized viewport, and every scrolling key sits behind an `if !m.ready { break }` guard.  
`enter` is deliberately *in front of* that guard — continuing needs no rendered document, so the one action the screen offers is never blocked by a size message that has not arrived yet.

**Deliver the content as a message even when nothing is fetched.**  
`loadOverview` only reads the last check result out of the controller and cannot fail, but it is still a `tea.Cmd` returning `UpdateOverviewLoadedMsg` — the same shape as the EULA screen, and it keeps `Init()` the single place that decides what a fresh visit shows.

**`GetFormTitle()` returns the empty string.**  
The document carries its own H1, and the header would otherwise print the heading twice.

**`GetScrollInfo() (scrolled, atEnd bool, percent int)`** is the accessor a footer uses to show scroll position; it returns zeroes while `!ready` rather than reading an unsized viewport.  
`app.renderFooter` consults it only on the EULA screen today, which needs "you have reached the end" for consent.  
This screen instead announces what the keys do, and only once there is something to scroll:

```go
func (m *UpdateOverviewModel) GetFormHotKeys() []string {
    hotkeys := []string{"enter"}
    if m.ready && m.content != "" {
        hotkeys = append(hotkeys, "up|down", "pgup|pgdown", "home|end")
    }
    return hotkeys
}
```

**Build the document in a pure function.**
`overviewDocument(updates []checker.StackUpdate, checkSucceeded bool) string` lives
outside the model and `update_overview_test.go` tests that function. There is no
harness for a bubbletea screen in the wizard, and what is worth testing is what the
document *says*, not how it scrolls.

### **4. Action Screen with a Terminal Panel**

`wizard/models/installer_update.go` is the other new shape: no form fields at all, but
the `BaseScreen` layout and help panel are worth keeping. It embeds `BaseScreen`,
implements `BaseScreenHandler` with empty form plumbing, and puts a
`terminal.Terminal` where the form viewport would be.

```go
type InstallerUpdateModel struct {
    *BaseScreen

    processor processor.ProcessorModel
    terminal  terminal.Terminal

    offered  *processor.InstallerPackage // what the server offers
    offerErr error                       // ...or why there is none

    running    bool
    confirming bool
}

var _ BaseScreenModel = (*InstallerUpdateModel)(nil)
var _ BaseScreenHandler = (*InstallerUpdateModel)(nil)
```

**`BuildForm()` builds the panel, not fields — and sets `initialized` itself:**

```go
func (m *InstallerUpdateModel) BuildForm() tea.Cmd {
    m.SetFormFields([]FormField{}) // action screen: nothing to focus, nothing to save

    contentWidth, contentHeight := m.getViewportFormSize()
    if m.terminal == nil {
        if !m.isVerticalLayout() {
            contentWidth -= 2
        }
        m.terminal = terminal.NewTerminal(contentWidth-2, contentHeight-1,
            terminal.WithAutoScroll(), terminal.WithAutoPoll(), terminal.WithCurrentEnv())
    } else {
        m.terminal.Clear()
    }

    m.terminal.Append(locale.InstallerUpdateAsking)

    if !m.initialized {
        m.initialized = true
        return m.terminal.Init()
    }

    return nil
}
```

The overridden `View()` also calls `BuildForm()` while `!initialized` and discards the command it returns. Without setting the flag here, every frame would clear the panel and re-append the first line, wiping out the output of a running download on each redraw.

**Re-size the terminal by hand.** The terminal is not part of the `BaseScreen` viewports, so `updateViewports()` does not manage it. In the `tea.WindowSizeMsg` branch, the same arithmetic from `BuildForm` is repeated (`getViewportFormSize()`, subtracting two columns in horizontal layout and accounting for the border) before calling `m.terminal.SetSize(...)`. It is important to keep these two pieces of arithmetic in sync.

**Never let key messages fall through to `BaseScreen`.** In the base screen logic, `enter` means save-and-return and `ctrl+r` means reset. On an action screen, this behavior is undesirable, as pressing `enter` could exit mid-download. Therefore, the `tea.KeyMsg` branch for this model handles `enter`, `y`, and `n` itself and passes all other keys to the terminal for scrolling. While `m.terminal.IsRunning()` is true, every key message goes directly to the terminal unfiltered, making the screen's own keys unreachable by construction rather than by explicit checks.

**Load asynchronously, and carry the failure in the same message.**

```go
func (m *InstallerUpdateModel) Init() tea.Cmd {
    m.running, m.confirming = false, false
    m.offered, m.offerErr = nil, nil

    return tea.Batch(m.BaseScreen.Init(), m.loadPackage)
}

type InstallerPackageLoadedMsg struct {
    Package *processor.InstallerPackage
    Err     error
}
```

Registry screens are singletons that are built once at start-up, but `Init()` runs on every navigation to the screen. This means per-visit state has to be cleared here. Otherwise, on a second visit, the screen would display the previous answer, and if `confirming` were still set, pressing `y` could act on an outdated offer. Both the package and error outcomes travel in a single message and are stored, because the logic for allowing a download is based on `canDownload()`—defined as `!running && offered != nil && offerErr == nil`. If the screen fails to load the package, it becomes inactive: pressing `enter` has no effect, and `GetFormHotKeys()` returns no keys to show available actions.

**A confirmation only where there is a question.** Pressing `enter` starts the download immediately. The confirmation prompt appears only if the file about to be written already exists—only in this case do `y` and `n` have meaning. The hotkeys are managed to match these three states exactly, ensuring that the footer never suggests a key that the screen would ignore.

```go
func (m *InstallerUpdateModel) GetFormHotKeys() []string {
    if m.terminal != nil && m.terminal.IsRunning() {
        return nil
    }
    if m.confirming {
        return []string{"y|n"}
    }
    if !m.canDownload() {
        return nil
    }

    return []string{"enter"}
}
```

Note that this shadows `BaseScreen.GetFormHotKeys()`. The base gates `down|up`, `ctrl+s`, and `ctrl+r` behind having fields or a list, so on a screen with neither it offers none of them — but it appends `enter` unconditionally. `enter` is exactly the key this screen has to be able to withhold: while the download is running, and on a screen whose offer never arrived, pressing it does nothing and the footer must not say otherwise.

The declining branch matters as much as the confirming one: `n` clears `confirming` and rewrites the panel with what the user now needs, rather than returning to a screen that still reads as if nothing happened.

**Text in a pure function, again.** `installerUpdateLines(running, offered, err)` returns the panel contents as `[]string`, so `installer_update_test.go` can assert what the screen *claims* without a terminal.

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

    m.BaseScreen = NewBaseScreen(controller, styles, window, m, m)
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
func (m *FormModel) BuildForm() tea.Cmd {
    // Track initially set fields for cleanup
    m.initiallySetFields = make(map[string]bool)

    for _, fieldConfig := range m.fieldConfigs {
        envVar, _ := m.GetController().GetVar(fieldConfig.EnvVarName)
        m.initiallySetFields[fieldConfig.Key] = envVar.IsPresent()

        field := m.createFieldFromEnvVar(fieldConfig, envVar)
        fields = append(fields, field)
    }

    m.SetFormFields(fields)

    return nil // a form screen has nothing to start; only a panel screen returns a cmd
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

### **Screen Registration — four mandatory steps**

All four are required. Skipping any of them fails *quietly*: nothing panics and nothing is logged. Three of them show a symptom somewhere other than the code you changed; the fourth — step 2 — shows no symptom at all with the screens written so far, which is exactly why it is the one that gets skipped.

**1. The `ScreenID` constant** in `wizard/models/types.go`:

```go
const (
    ExampleFormScreen ScreenID = "example_form"
    // parameterised screens join their arguments with "§"
    UpdatePentagiScreen ScreenID = "processor_operation_form§compose§update"
)
```

The ID is the only key the navigator, the registry and every menu share. Without it there is nothing to push onto the navigation stack and nothing to key the registry by.

**2. A case in `models.RestoreModel`** (`types.go`):

```go
case *ExampleFormModel:
    return m
```

`app.forwardMsgToCurrentModel` calls `Update` on the current model and adopts the result only when `RestoreModel` recognises it — when it returns `nil` there is no fallback and no log: the app keeps the pointer it already holds, and returns the command either way. What a missing branch actually costs is worth stating precisely, because the obvious answer is wrong. Every screen in `wizard/models` has a pointer receiver on `Update` and returns *itself*, so the object the app renders is the object `Update` has just mutated — nothing freezes, and a missing branch shows no symptom at all. The branch is load-bearing for the case the type allows but nobody has written yet: a model that returns a *different* model, dropped silently with nothing anywhere to say why. That is why the contract is "every screen is listed" rather than "list the ones that need it" — the day a screen starts returning something else is not the day anybody remembers this rule. The `nil` is load-bearing in the other direction too: a `MaintenanceModel` delegates `Update` to its embedded `*ListScreen`, which returns *itself*, and `RestoreModel` not knowing `*ListScreen` is exactly what keeps the app pointing at the whole screen instead of one of its parts — so list the concrete screen type and nothing else.

`TestTypes_RestoreModel_RestoresTheUpdateScreens` in `types_test.go` pins membership:

```go
for _, model := range []tea.Model{(*InstallerUpdateModel)(nil), (*UpdateOverviewModel)(nil)} {
    if RestoreModel(model) == nil {
        t.Errorf("%T is missing its branch in RestoreModel", model)
    }
}
```

**3. Construction in `registry.initScreens`** (`wizard/registry/registry.go`):

```go
r.screens[models.ExampleFormScreen] = models.NewExampleFormModel(r.controller, r.styles, r.window)
```

Screens are built once here and reused; `registry.GetScreen` falls back to a `MockFormModel` for an unknown ID **and caches it under that ID**, so the screen you wrote is never constructed at all. The first symptom is not the screen but the menu: `ListScreen` labels each entry by asking the registry for the model behind its ID (`GetFormName`, `GetFormOverview`, `GetCurrentConfiguration`, `IsConfigured`), so the entry reads "Development Screen" and opens a placeholder. An unregistered screen also never receives `tea.WindowSizeMsg`, which `registry.HandleMsg` forwards to registered screens only.

**4. Strings in `wizard/locale/locale.go`, and hotkey captions in `app.initHotkeysLocale`** (`wizard/app.go`). All user-visible text lives in `locale`; the footer resolves each name from `GetFormHotKeys()` through the caption map:

```go
if localeHotKey, ok := app.hotkeys[hotkey]; ok {
    actions = append(actions, localeHotKey)
}
```

An unknown name is skipped without complaint, so the key keeps working and is simply never announced — the quietest of the four failures, and the only one whose symptom is a hint that is missing rather than a screen that is wrong. Neither new screen needed a new caption: `enter`, `up|down`, `pgup|pgdown`, `home|end` and `y|n` are already in the map.

### **Width-aware markdown rendering**

`styles.GetRenderer()` returns a single renderer built once with `glamour.WithWordWrap(DefaultRenderWidth)` — 80 columns, whatever the terminal is doing. That was enough while the only markdown screen was the EULA, which renders once at load and is prose written to be read at a fixed measure. It is not enough for a changelog: at 80 columns in a 200-column window the document uses a third of the screen, and in a terminal narrower than 80 every line breaks two or three times, which is what a list of component names does worst.

`styles.GetRendererForWidth(width)` is the replacement for screens that know how much room they have:

```go
func (s *Styles) GetRendererForWidth(width int) *glamour.TermRenderer {
    width = min(max(width, MinRenderWidth), MaxRenderWidth) // [40, 120]
    // ... memoised per width; on any failure returns the fixed 80-column renderer
}
```

Four properties are worth knowing before you use it:

- **It clamps to `[MinRenderWidth, MaxRenderWidth]` = `[40, 120]`.** The floor keeps a narrow terminal from splitting every line; the ceiling is a readability limit — a line spanning a maximised terminal is measurably harder to read.
- **It never returns `nil`.** A width that cannot be built falls back to the default renderer, so no render path needs a nil check — and a nil check that someone forgets is how a changelog ends up not being shown at all.
- **The cache is behind a pointer** (`renderers *rendererCache`), because `styles.Styles` is copied by value into every screen and a value cache would be cloned per screen, defeating the point.
- **The cache is mutex-guarded**, because renderers are built inside `tea.Cmd` goroutines — off the update loop, where two screens can ask for the same width at once.

`GetRenderer()` still exists and still wraps at 80; prefer `GetRendererForWidth` on anything that renders inside a sized panel.

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

// A screen that embeds BaseScreen asserts both: the registry stores it as a model,
// and BaseScreen calls back into it as a handler
var _ BaseScreenModel = (*InstallerUpdateModel)(nil)
var _ BaseScreenHandler = (*InstallerUpdateModel)(nil)

// A document screen has no form, so it asserts only the model interface
var _ BaseScreenModel = (*UpdateOverviewModel)(nil)
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
4. **Document**: No BaseScreen — implement BaseScreenModel over a viewport (`eula.go`, `update_overview.go`)
5. **Action with terminal panel**: Inherit BaseScreen with zero fields, own the terminal and override `View()` (`installer_update.go`, `apply_changes.go`)

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
