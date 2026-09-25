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
- **Cloud Layer**: The only route out to the PentAGI Cloud API — request construction, envelope handling, error classification

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

## 🏗️ **Screen Registration Architecture**

### **Where the Update Screens Sit**
Two screens were inserted into the maintenance branch of the navigation graph, and both took over a destination that used to belong to the generic processor operation form:

```
MaintenanceScreen (list)
  ├── UpdateOverviewScreen ──enter──> UpdatePentagiScreen   ("processor_operation_form§compose§update")
  └── InstallerUpdateScreen                                  (replaced "processor_operation_form§installer§update")
```

- **`UpdateOverviewScreen`** stands *between* the menu entry and the operation form.
The entry used to lead straight into "are you sure?", asking for consent to something the interface had never described.
The overview describes it, then continues to the same form — the confirmation there is untouched.
The screen makes no network call: everything it renders comes from the last check result held by the controller, so it cannot fail and has no error state.
`Enter` is deliberately *not* bound to page-down as it is on the EULA screen: this screen asks for no consent, so its one unambiguous key belongs to its one action.

- **`InstallerUpdateScreen`** is a screen rather than an operation form because what it does is describe a build and download one verified file, and the generic form's help text claimed the running binary would be replaced and the app would exit — something no version of the installer has ever done.
It names version, platform, size and target path *first*, then offers a single action, and confirms only the one thing worth confirming: a file already occupying the target name.

### **The Four Registration Steps**
A new screen is not a single file. It is four edits, and only the first of them is enforced by the compiler:

| # | Where | What |
|---|---|---|
| 1 | `wizard/models/types.go` | the `ScreenID` constant |
| 2 | `wizard/models/types.go` — `RestoreModel` | a `case` for the concrete model type |
| 3 | `wizard/registry/registry.go` — `initScreens` | construct the screen and store it under its `ScreenID` |
| 4 | `wizard/locale/locale.go` + `app.initHotkeysLocale` | the screen's strings, and a caption for every hotkey it returns |

```go
// Step 2 — the branch without which the screen is inert
func RestoreModel(model tea.Model) BaseScreenModel {
    switch m := model.(type) {
    // ...
    case *UpdateOverviewModel:
        return m
    case *InstallerUpdateModel:
        return m
    default:
        return nil
    }
}
```

### **Why Step 2 Is the Dangerous One**
`RestoreModel` is the *only* path by which a model returned from `Update` gets back into the
app:

```go
func (app *App) forwardMsgToCurrentModel(msg tea.Msg) tea.Cmd {
    model, cmd := app.currentModel.Update(msg)
    if newModel := models.RestoreModel(model); newModel != nil {
        app.currentModel = newModel
    }
    return cmd
}
```

A type the switch does not know yields `nil` and the app keeps the model it already had.

What that costs is worth stating precisely, because the obvious answer is wrong: every screen here has a pointer receiver on `Update` and returns itself, so the mutation has already happened in place and the pointer the app is holding *is* the updated model — nothing freezes.

The branch is load-bearing for the case the signature allows but no screen has used yet: a model whose `Update` returns a *different* model. That one is dropped silently, with no panic, no log line and no error anywhere.

The compile-time assertion each screen carries (`var _ BaseScreenModel = (*InstallerUpdateModel)(nil)`) does **not** catch it either, because `RestoreModel` is a runtime type switch and a missing `case` is not a type error.

`TestTypes_RestoreModel_RestoresTheUpdateScreens` (`models/types_test.go`) therefore enumerates both new screens, and that test is the only guard there is. The contract it encodes is "every screen is listed" rather than "list the ones that need it", because the day a screen starts returning something else is not the day anybody remembers this rule.

Steps 3 and 4 fail just as quietly, which is why they are listed rather than assumed:

- A screen missing from `initScreens` is not a crash — `GetScreen` falls back to `initMockScreen` and the user gets a placeholder form where the feature should be.
- The footer renders a hotkey only if its key is present in `app.hotkeys` (`if localeHotKey, ok := app.hotkeys[hotkey]; ok`). A screen returning a key with no caption simply shows nothing in the footer.

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

    for _, providerID := range []string{"openai", "anthropic", "gemini", "bedrock", "deepseek", "glm", "kimi", "qwen", "ollama", "custom"} {
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

## 🏗️ **Cloud Client Architecture**

### **Single Egress Pattern**
Everything the installer asks of the PentAGI Cloud API goes through one package: `cmd/installer/cloud`.

Both `checker` and `processor` build a `cloud.Config` from the installer state and call methods on the client. Neither constructs requests, parses responses, or opens connections to `update.pentagi.com` directly.

The `cloud` package is independent from the rest of the installer — it does not depend on `checker`, `processor`, or `wizard`. The dependency arrow points only one way. In tests, the client can be exercised by substituting the call functions, with no actual server involved.

```go
// cloud/client.go — three routes bound in one build
const (
    routeUpdatesCheck     = "updates_check"      // POST /api/v1/proxy/updates/check
    routePackagesInfo     = "packages_info"      // GET  /api/v1/proxy/packages/info
    routePackagesDownload = "packages_download"  // GET  /api/v1/proxy/packages/download
)

type Client struct {
    checkUpdates    sdk.CallReqBytesRespBytes
    packageInfo     sdk.CallReqQueryRespBytes
    downloadPackage sdk.CallReqQueryRespWriter
    // ...
}
```

Route names are part of the contract, not internal labels: the server dispatches on the name, and one it does not recognise is refused outright rather than reported as a missing page.

### **One Proxy Setting Governs Every Call**
The SDK's default transport reads the proxy from the process environment. The installer has its own `PROXY_URL` setting — often carrying credentials the environment does not — and that setting has to win, because it is the one the user actually configured.

```go
func newTransport(proxyURL string) (*http.Transport, error) {
    transport := sdk.DefaultTransport()
    value := strings.TrimSpace(proxyURL)
    if value == "" {
        return transport, nil // nothing configured: leave the environment behaviour alone
    }
    parsed, err := url.Parse(value)
    // ... reject a URL with no host ...
    transport.Proxy = http.ProxyURL(parsed)
    return transport, nil
}
```

The transport is built once in `New` and shared by all three calls, so there is no route that can quietly bypass it. Both construction sites — the checker's update check and the processor's package download — read the same `PROXY_URL`, which is what makes "one setting" true in practice rather than only by intent.

Two neighbouring normalisations belong to the same idea of a single, predictable egress:

- **`NormalizeHost`** reduces the configured address to `host[:port]`. A scheme and trailing path are tolerated because the setting historically held a full URL, but an explicit `http://` is an **error** rather than a silent upgrade: the connection is always encrypted, and quietly doing the opposite of what the user wrote would hide a misconfiguration.

- **`NormalizeVersion`** turns the build version into one the contract accepts, falling back to `0.0.0` for anything that is not a version number — a branch name, a bare commit hash, an empty string. The original is not lost: it travels in the `User-Agent`, where it identifies the exact build without having to satisfy the version grammar.

### **The License Key Is Validated Before It Is Handed Over**
```go
if key := strings.TrimSpace(cfg.LicenseKey); key != "" {
    if _, err := sdk.IntrospectLicenseKey(key); err != nil {
        client.licenseWarning = fmt.Errorf("license key is not usable, continuing without it: %w", err)
    } else {
        options = append(options, sdk.WithLicenseKey(key))
    }
}
```

The SDK drops a key it cannot decode without saying so, and the call then runs anonymously — which from the outside looks exactly like the license not being honoured. Validating first turns that into a stated reason, available from `LicenseWarning()`.

An unusable key is deliberately **not** fatal. The update check is the one thing that still works without a license, so refusing to build the client would leave the user with no way to discover that their key is the problem. `New` fails only on configuration that cannot produce working calls at all: an unusable host or an unusable proxy.

### **Answers Arrive Inside an Envelope**
Every non-streaming answer is wrapped as `{"status": …, "data": …}`, and unwrapping it is not optional:

```go
answer, err := c.checkUpdates(ctx, body)
if err != nil {
    return nil, Classify(err)
}
// Decoding straight into the response type SUCCEEDS against the wrapper — every field
// simply stays at its zero value. Skipping this reports "no updates" for every answer.
response, err := models.ParseEnvelope[models.CheckUpdatesResponse](answer)
```

This is the sharpest edge in the whole client. `json.Unmarshal` of an envelope into the payload type produces no error and no populated fields, so the installer concludes there is nothing to install and never learns it was wrong.

`ParseEnvelope` also turns a `2xx` body whose status is not `success` into an `*APIError` describing what the server reported. Binary downloads are the one exception — they carry the file itself, with no envelope around it.

### **Failure Classification**
`Classify` maps SDK errors onto reasons the interface can act on, wrapping the original so `errors.Is`/`errors.As` keep working through it:

| Reason | What the user can do about it |
|---|---|
| `unreachable` | check the network — nothing ever reached a verdict |
| `timeout` | a deadline, including the proof-of-work budget on slow hardware |
| `rate_limited` / `quota_exceeded` | wait; `RetryAfter` says how long |
| `forbidden` | retrying never helps; the license does |
| `not_found` | an answer, not an outage — no such package or version |
| `rejected` | the request was malformed: our bug, and resending it cannot succeed |
| `server_error` | worth retrying later |

`Retryable()` collapses the table back down for callers that only need the one bit. Ordering inside `Classify` matters: the typed errors are matched before the sentinels they wrap, because that is where the server-advertised cooldown lives. Collapsing all of this into a single "update server unavailable" flag — as the installer used to — tells the user nothing they can act on.

### **Verification Belongs to the Client, Not the Caller**

`PackageInfo` validates the answer before returning it, because everything `DownloadPackage` checks is taken from that answer: an unverified description of a file is not a basis for trusting the file.

The download then feeds one pass over the stream into three checks at once:

```go
digest := sha256.New()
counter := &countingWriter{}
signed := info.Signature.ValidateWrapWriter(io.MultiWriter(dst, digest, counter))
```

The expected length has to come from `info.Size` because the encrypted response carries no `Content-Length` — there is no header to read it from. It catches nothing the sha256 and the signature would miss; what it buys is a legible error ("package is N bytes, expected M") ahead of a bare hash mismatch.

Verification can only complete once the last byte is written, so on any error `dst` holds a partial or unverified file. Discarding it is the caller's job, and the two callers discharge that obligation differently on purpose.

`installPluginFile` streams into `<target>.new` and renames only once the write has returned clean, because the file it replaces is the one a running Jaeger has mounted — a download that dies halfway would otherwise leave a truncated plugin and the container would come back up unable to start.

`fetchInstaller` does the opposite: it writes the binary under its final name in the environment directory and removes it on error. Staging an executable through the system temp directory and moving it into place is a pattern antivirus and EDR products score heavily, and a false positive there costs the user their update.

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
- **Single Egress**: One package, one proxy setting, one place where answers are unwrapped and failures are classified