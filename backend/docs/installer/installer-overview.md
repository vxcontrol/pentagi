# PentAGI Installer Overview

> Comprehensive guide to the PentAGI installer - a robust Terminal User Interface (TUI) for configuring and deploying PentAGI services.

## 🎯 **Project Overview**

The PentAGI installer provides a modern, interactive Terminal User Interface for configuring and deploying the PentAGI autonomous penetration testing platform. Built using the [Charm](https://charm.sh/) tech stack, it implements responsive design patterns optimized for terminal environments.

### **Core Purpose**
- **Configuration Management**: Interactive setup of LLM providers, monitoring, and security settings
- **Environment Setup**: Automated configuration of Docker services and environment variables
- **User Experience**: Professional TUI with intuitive navigation and real-time validation
- **Production Ready**: Robust error handling, state persistence, and graceful degradation

### **Build Command**
```bash
# From backend/ directory
go build -o ../build/installer ./cmd/installer/main.go

# Monitor debug output
tail -f log.json | jq '.'
```

## 🏗️ **Technology Stack**

### **Core Technologies**
- **TUI Framework**: BubbleTea (Model-View-Update pattern)
- **Styling**: Lipgloss (CSS-like styling for terminals)
- **Components**: Bubbles (viewport, textinput, etc.)
- **Markdown**: Glamour (markdown rendering)
- **Language**: Go 1.21+

### **Architecture Components**
- **Navigation**: Type-safe screen routing with parameter passing
- **State Management**: Persistent configuration with environment variable integration
- **Layout System**: Responsive design with breakpoint-based layouts
- **Form System**: Dynamic forms with validation and auto-completion
- **Controller Layer**: Business logic abstraction from UI components

## 🎯 **Key Features**

### **Responsive Design**
- **Adaptive Layout**: Automatically adjusts to terminal size
- **Breakpoint System**: Horizontal/vertical layouts based on terminal width
- **Content Hiding**: Graceful degradation when space is insufficient
- **Dynamic Sizing**: Form fields and panels resize automatically

### **Interactive Configuration**
- **LLM Providers**: Support for OpenAI, Anthropic, Gemini, Bedrock, DeepSeek, GLM, Kimi, Qwen, Ollama, Custom endpoints
- **Monitoring Setup**: Langfuse integration for LLM observability
- **Observability**: Complete monitoring stack with Grafana, VictoriaMetrics, Jaeger
- **Summarization**: Advanced context management for LLM interactions

### **Professional UX**
- **Auto-Scrolling Forms**: Fields automatically scroll into view when focused
- **Tab Completion**: Boolean fields offer `true`/`false` suggestions
- **Real-time Validation**: Immediate feedback with human-readable error messages
- **Resource Estimation**: Live calculation of token usage and memory requirements
- **State Persistence**: Navigation and form state preserved across sessions

## 🏗️ **Architecture Overview**

### **Directory Structure**
```
backend/cmd/installer/
├── main.go                 # Application entry point
├── wizard/
│   ├── app.go             # Main application controller
│   ├── controller/        # Business logic layer
│   │   └── controller.go
│   ├── locale/            # Localization constants
│   │   └── locale.go
│   ├── logger/            # TUI-safe logging
│   │   └── logger.go
│   ├── models/            # Screen implementations
│   │   ├── welcome.go     # Welcome screen
│   │   ├── eula.go        # EULA acceptance
│   │   ├── main_menu.go   # Main navigation
│   │   ├── llm_providers.go
│   │   ├── llm_provider_form.go
│   │   ├── summarizer.go
│   │   ├── summarizer_form.go
│   │   ├── update_overview.go   # What an update changes (informational)
│   │   ├── installer_update.go  # Installer build download
│   │   └── types.go       # Shared types
│   ├── styles/            # Styling and layout
│   │   └── styles.go
│   └── window/            # Terminal size management
│       └── window.go
```

### **Component Responsibilities**

#### **App Layer** (`app.go`)
- Global navigation management
- Screen lifecycle (creation, initialization, cleanup)
- Unified header and footer rendering
- Window size distribution to models
- Global event handling (ESC, Ctrl+C, resize)

#### **Models Layer** (`models/`)
- Screen-specific logic and state
- User interaction handling
- Content rendering (content area only)
- Local state management

#### **Controller Layer** (`controller/`)
- Business logic abstraction
- Environment variable management
- Configuration persistence
- State validation

#### **Styles Layer** (`styles/`)
- Centralized styling and theming
- Dimension management (singleton pattern)
- Memoised glamour renderers: one per wrap width, plus the default 80-column one
- Responsive style calculations

#### **Window Layer** (`window/`)
- Terminal size management
- Content area size calculations
- Dimension change coordination

## 🎯 **Navigation System**

### **Composite ScreenID Architecture**
The installer implements a sophisticated navigation system using composite screen IDs:

```go
// Format: "screen§arg1§arg2§..."
type ScreenID string

// Examples:
"welcome"                    // Simple screen
"main_menu§llm_providers"    // Menu with selection
"llm_provider_form§openai"   // Form with provider type
"summarizer_form§general"    // Form with configuration type
```

### **Navigation Features**
- **Parameter Preservation**: Arguments maintained across navigation
- **Stack Management**: Proper back navigation without loops
- **State Persistence**: Complete navigation state restoration
- **Universal ESC**: Always returns to welcome screen
- **Type Safety**: Compile-time validation of screen IDs

### **Navigation Flow Example**
```
1. Start: ["welcome"]
2. Continue: ["welcome", "main_menu"]
3. LLM Providers: ["welcome", "main_menu§llm_providers", "llm_providers"]
4. OpenAI Form: [..., "llm_provider_form§openai"]
5. GoBack: [..., "llm_providers§openai"]
6. ESC: ["welcome"]
```

## 🎯 **Form System Architecture**

### **Advanced Form Patterns**
- **Boolean Fields**: Tab completion with `true`/`false` suggestions
- **Integer Fields**: Range validation with human-readable formatting
- **Environment Integration**: Direct EnvVar integration with presence detection
- **Smart Cleanup**: Automatic removal of cleared environment variables
- **Resource Estimation**: Real-time calculation of token/memory usage

### **Dynamic Field Generation**
Forms adapt based on configuration type:
```go
// Type-specific field generation
switch m.configType {
case "general":
    m.addBooleanField("use_qa", "Use QA Pairs", envVar)
    m.addIntegerField("max_sections", "Max Sections", envVar, 1, 50)
case "assistant":
    m.addIntegerField("keep_sections", "Keep Sections", envVar, 1, 10)
}
```

### **Viewport-Based Scrolling**
Forms automatically scroll to keep focused fields visible:
- **Auto-scroll**: Focused field automatically stays visible
- **Smart positioning**: Calculates field heights for precise scroll positioning
- **No extra hotkeys**: Uses existing navigation keys

## 🎯 **Configuration Management**

### **Supported Configurations**

#### **LLM Providers**
- **OpenAI**: GPT-4, GPT-3.5-turbo with API key configuration
- **Anthropic**: Claude-3, Claude-2 with API key configuration
- **Google Gemini**: Gemini Pro, Ultra with API key configuration
- **AWS Bedrock**: Multi-model support with AWS credentials
- **DeepSeek/GLM/Kimi/Qwen**: Base URL + API Key + Provider Name (optional, for LiteLLM)
- **Ollama**: Local model server integration
- **Custom**: OpenAI-compatible endpoint configuration

#### **Monitoring & Observability**
- **Langfuse**: LLM observability (embedded or external)
- **Observability Stack**: Grafana, VictoriaMetrics, Jaeger, Loki
- **Performance Monitoring**: System metrics and health checks

#### **Summarization Settings**
- **General**: Global conversation context management
- **Assistant**: Specialized settings for AI assistant contexts
- **Token Estimation**: Real-time calculation of context size

## 🎯 **Update Screens**

Two screens under **Maintenance** stand outside the generic operation form: neither edits configuration, and neither is an operation the form framework describes for itself. They sit at different distances from the form system — the overview is a plain viewport over a rendered document, while the installer update is still a `BaseScreen`, built with an empty field list, which is what gives it the help panel and the vertical/horizontal layout.

The two entries appear under different conditions. **Update Installer** shows only when a newer build is on offer *and* the update server answered. **Update PentAGI** shows whenever an installed stack is not known to be up to date — which includes a check that failed, because a failed check resets every per-stack verdict instead of leaving a stale one in place. So the overview is reachable with nothing yet known to need doing, and that is exactly why it has to be able to say the check did not complete.

### **Update Overview** (`models/update_overview.go`)

Sits **between** the "Update PentAGI" menu entry and the apply form, which the entry used to point at directly — so the first thing the user saw was "are you sure?" about something the interface had never described. The overview describes it, then continues to the same form, confirmation and all.

- **Informational only**: the document is built from the update check already in memory, so the screen makes no network call and cannot fail
- **Per stack**: the version move (`current → target`), a note when components come from different releases and the version shown is therefore the oldest of them, and the per-component plan
- **Three component verdicts, spelled out in words**: `will be updated`, `unchanged`, and `cannot verify — nothing to compare against`. The third is not a synonym for the second — a component with nothing to compare against is not a component that is current, and collapsing the two is how an installation gets told it is up to date on no evidence at all
- **The release notes of every release the update crosses**, oldest first, each with its date and a `(preview)` marker when it is not stable. An answer that carries no release list falls back to the target release's changelog and release notes
- **Stacks with nothing to do** are gathered in one list at the bottom rather than omitted: silence about a stack reads as "not installed"
- **A check that did not complete** says so instead of saying "up to date" — the screen is reachable from a menu drawn by an earlier check, and "up to date" there would be the opposite claim
- **Keys**: `enter` continues to the apply form; `↑`/`↓`, `pgup`/`pgdown`, `home`/`end` scroll. Deliberately *not* the EULA's `enter`-pages-down: this screen asks for no consent, so the one unambiguous key belongs to the one action it offers
- **Width-aware markdown**: rendered through `styles.GetRendererForWidth()` (clamped to a readable band, memoised) instead of the fixed 80-column renderer, and re-rendered only when the wrap width actually changes — a drag-resize delivers a `WindowSizeMsg` per column and glamour is not cheap

### **Installer Update** (`models/installer_update.go`)

Its own screen rather than an operation form, because the operation is a single verified download and the generic form described it as replacing the running binary and exiting.

- **Says what it will do before it does it**: running version, offered version, platform (`os/arch`), download size and the full path the file will be written to. The package description is fetched on entry so the size is on screen *before* the user agrees to the download; the answer is reused by the download itself rather than asked for twice
- **Download and verification are the whole operation.** Length, `sha256` and signature are checked in a single pass as the bytes arrive; anything that fails is deleted rather than left behind under a release-looking name
- **The installer never replaces itself**, and that is deliberate rather than unfinished: a process cannot reliably overwrite the binary it is executing — Linux refuses the write, Windows holds the file locked — and failing halfway leaves the user with no installer at all. Instead the screen prints the exact command: `mv "<downloaded file>" "<running installer>"`, or `move /Y` on Windows plus a note to close the installer first. Both paths are quoted, so an installation under a path with a space does not become two arguments; when the running executable's path cannot be determined, no command is printed at all, because a command with a guessed destination is worse than none
- **One confirmation, for one thing**: a file already sitting under that exact name — a previous download, or one interrupted partway. Writing over empty space is not a question, so it is not asked. Declining still prints the move command, since the file the user chose to keep is the build they came for
- **Where the file lands**: beside the environment file, named `installer_<version>` (`.exe` on Windows), written directly under its final name — staging an executable through the system temp directory and moving it is a pattern antivirus and EDR products score heavily. A download that would land on the running binary is refused outright
- **Progress** is reported against the size from the package description rather than a `Content-Length`, which the encrypted response body does not carry, and printed once per 10% rather than once per read — a line every few kilobytes would bury everything around it in a panel that scrolls, and the path and the verification result are printed after the last of them

## 🎯 **Localization Architecture**

### **Centralized Constants**
All user-visible text stored in `locale/locale.go`:
```go
// Screen-specific constants
const (
    WelcomeTitle = "PentAGI Installer"
    WelcomeGreeting = "Welcome to PentAGI!"

    // Form help text with practical guidance
    LLMFormOpenAIHelp = `OpenAI provides access to GPT models...

Get your API key from:
https://platform.openai.com/api-keys`
)
```

### **Multi-line Help Text**
Detailed guidance integrated into forms:
- Provider-specific setup instructions
- Configuration recommendations
- Troubleshooting tips
- Best practices

## 🎯 **Error Handling & Recovery**

### **Graceful Degradation**
- **Dimension Fallbacks**: Handles invalid terminal sizes
- **Content Fallbacks**: Shows loading states and error messages
- **Network Resilience**: Offline operation support
- **State Recovery**: Automatic restoration from corrupted state

### **User-Friendly Error Messages**
- **Validation Errors**: Real-time feedback with clear guidance
- **System Errors**: Plain English explanations with suggested fixes
- **Network Errors**: Offline alternatives and retry mechanisms

## 🎯 **Performance Considerations**

### **Optimizations**
- **Lazy Loading**: Content loaded on-demand when screens accessed
- **Memoised Renderers**: one cached glamour instance per wrap width, built once instead of per render
- **Efficient Scrolling**: Viewport-based rendering for large content
- **Memory Management**: Proper cleanup and resource sharing

### **Responsive Performance**
- **Breakpoint-Based**: Layout decisions based on terminal capabilities
- **Content Adaptation**: Hide non-essential content on small screens
- **Progressive Enhancement**: Full features on capable terminals

## 🎯 **Development Workflow**

### **File Organization**
- **One Model Per File**: Clear separation of screen logic
- **Shared Constants**: Type definitions in `types.go`
- **Centralized Locale**: All text in `locale.go`
- **Clean Dependencies**: Business logic isolated in controllers

### **Code Style**
- **Compact Syntax**: Where appropriate for readability
- **Expanded Logic**: For complex business rules
- **Comments**: Explain "why" and "how", not "what"
- **Error Handling**: Graceful degradation with user guidance

### **Testing Strategy**
- **Build Testing**: Successful compilation verification
- **Manual Testing**: Interactive validation on various terminal sizes
- **Dimension Testing**: Minimum (80x24) to large terminal support
- **Navigation Testing**: Complete flow validation

This overview provides the foundation for understanding the PentAGI installer's architecture, features, and development approach. The system prioritizes user experience, maintainability, and production reliability.