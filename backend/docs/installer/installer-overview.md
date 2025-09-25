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
- **LLM Providers**: Support for OpenAI, Anthropic, Gemini, Bedrock, Ollama, Custom endpoints
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
- Shared glamour renderer (prevents freezing)
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
- **Single Renderer**: Shared glamour instance prevents freezing
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