# PentAGI Configuration Guide

This document serves as a comprehensive guide to the configuration system in PentAGI, primarily aimed at developers. It details all available configuration options, their purposes, default values, and how they're used throughout the application.

## Table of Contents

- [PentAGI Configuration Guide](#pentagi-configuration-guide)
  - [Table of Contents](#table-of-contents)
  - [Configuration Basics](#configuration-basics)
  - [General Settings](#general-settings)
    - [Usage Details](#usage-details)
  - [Docker Settings](#docker-settings)
    - [Usage Details](#usage-details-1)
  - [Server Settings](#server-settings)
    - [Usage Details](#usage-details-2)
  - [Frontend Settings](#frontend-settings)
    - [Usage Details](#usage-details-3)
  - [Authentication Settings](#authentication-settings)
    - [Usage Details](#usage-details-4)
  - [Web Scraper Settings](#web-scraper-settings)
    - [Usage Details](#usage-details-5)
  - [LLM Provider Settings](#llm-provider-settings)
    - [OpenAI](#openai)
    - [Anthropic](#anthropic)
    - [Ollama LLM Provider](#ollama-llm-provider)
    - [Google AI (Gemini) LLM Provider](#google-ai-gemini-llm-provider)
    - [AWS Bedrock LLM Provider](#aws-bedrock-llm-provider)
    - [Custom LLM Provider](#custom-llm-provider)
    - [Usage Details](#usage-details-6)
  - [Embedding Settings](#embedding-settings)
    - [Usage Details](#usage-details-7)
  - [Summarizer Settings](#summarizer-settings)
    - [Usage Details and Impact on System Behavior](#usage-details-and-impact-on-system-behavior)
      - [Core Summarization Strategies and Their Parameters](#core-summarization-strategies-and-their-parameters)
      - [Deep Dive: Parameter Impact and Recommendations](#deep-dive-parameter-impact-and-recommendations)
    - [Summarization Effects on Agent Behavior](#summarization-effects-on-agent-behavior)
    - [Implementation Details](#implementation-details)
    - [Recommended Settings for Different Use Cases](#recommended-settings-for-different-use-cases)
  - [Assistant Settings](#assistant-settings)
    - [Usage Details](#usage-details-8)
    - [Recommended Assistant Settings for Different Use Cases](#recommended-assistant-settings-for-different-use-cases)
  - [Search Engine Settings](#search-engine-settings)
    - [DuckDuckGo Search](#duckduckgo-search)
    - [Sploitus Search](#sploitus-search)
    - [Google Search](#google-search)
    - [Traversaal Search](#traversaal-search)
    - [Tavily Search](#tavily-search)
    - [Perplexity Search](#perplexity-search)
    - [Searxng Search](#searxng-search)
    - [Usage Details](#usage-details-9)
  - [Proxy Settings](#proxy-settings)
    - [Usage Details](#usage-details-10)
  - [Graphiti Knowledge Graph Settings](#graphiti-knowledge-graph-settings)
    - [Usage Details](#usage-details-11)
  - [Observability Settings](#observability-settings)
    - [Telemetry](#telemetry)
    - [Langfuse](#langfuse)
    - [Usage Details](#usage-details-12)

## Configuration Basics

PentAGI uses environment variables for configuration, with support for `.env` files through the `godotenv` package. The configuration is defined in the `Config` struct in `pkg/config/config.go` and is loaded using the `NewConfig()` function.

```go
func NewConfig() (*Config, error) {
    godotenv.Load()

    var config Config
    if err := env.ParseWithOptions(&config, env.Options{
        RequiredIfNoDef: false,
        FuncMap: map[reflect.Type]env.ParserFunc{
            reflect.TypeOf(&url.URL{}): func(s string) (interface{}, error) {
                if s == "" {
                    return nil, nil
                }
                return url.Parse(s)
            },
        },
    }); err != nil {
        return nil, err
    }

    return &config, nil
}
```

This function automatically loads environment variables from a `.env` file if present, then parses them into the `Config` struct using the `env` package from `github.com/caarlos0/env/v10`.

## General Settings

These settings control basic application behavior and are foundational for the system's operation.

| Option         | Environment Variable | Default Value                                                                | Description                                                              |
| -------------- | -------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| DatabaseURL    | `DATABASE_URL`       | `postgres://pentagiuser:pentagipass@pgvector:5432/pentagidb?sslmode=disable` | Connection string for the PostgreSQL database with pgvector extension    |
| Debug          | `DEBUG`              | `false`                                                                      | Enables debug mode with additional logging                               |
| DataDir        | `DATA_DIR`           | `./data`                                                                     | Directory for storing persistent data                                    |
| AskUser        | `ASK_USER`           | `false`                                                                      | When enabled, requires explicit user confirmation for certain operations |
| InstallationID | `INSTALLATION_ID`    | *(none)*                                                                     | Unique installation identifier for PentAGI Cloud API communication       |
| LicenseKey     | `LICENSE_KEY`        | *(none)*                                                                     | License key for PentAGI Cloud API authentication and feature activation  |

### Usage Details

- **DatabaseURL**: This is a critical setting used throughout the application for all database connections. It's used to:
  - Initialize the primary SQL database connection in `main.go`
  - Create GORM ORM instances for model operations
  - Configure pgvector connectivity for embedding operations
  - Set up connection pools in various tools and executors

```go
// In main.go for SQL connection
db, err := sql.Open("postgres", cfg.DatabaseURL)

// In main.go for GORM connection
orm, err := database.NewGorm(cfg.DatabaseURL, "postgres")

// In tools for vector database operations
pgvector.WithConnectionURL(fte.cfg.DatabaseURL)
```

- **Debug**: Controls debug mode throughout the application, enabling additional logging and development features:
  - Activates detailed logging in the router setup
  - Can enable development endpoints and tools

```go
// In router.go for enabling debug mode
if cfg.Debug {
    // Enable debug features
}
```

- **DataDir**: Specifies where PentAGI stores persistent data. This is used across multiple components:
  - In `docker/client.go` for container volume mapping
  - For screenshots storage in `services.NewScreenshotService`
  - In tools for file operations and data persistence
  - In Docker container management for mapping volumes

```go
// In docker/client.go
dataDir, err := filepath.Abs(cfg.DataDir)

// In router.go for screenshot service
screenshotService := services.NewScreenshotService(orm, cfg.DataDir)

// In tools.go for various tools
dataDir: fte.cfg.DataDir
```

- **AskUser**: A safety feature that, when enabled, requires explicit user confirmation before executing potentially destructive operations:
  - Used in tools to prompt for confirmation before executing commands
  - Serves as a safeguard for sensitive operations

```go
// In tools.go
if fte.cfg.AskUser {
    // Prompt user for confirmation before executing
}
```

- **InstallationID**: A unique identifier for the PentAGI installation used for cloud API communication:
  - Generated automatically during installation or can be manually set
  - Required for certain cloud-based features and integrations

```go
// Used in cloud SDK initialization
if cfg.InstallationID != "" {
    // Initialize cloud API client with installation ID
}
```

- **LicenseKey**: Authentication key for PentAGI Cloud API and premium feature activation:
  - Validates license and enables licensed features
  - Required for enterprise features and support
  - Used for authentication with PentAGI Cloud services

```go
// Used in cloud SDK initialization
if cfg.LicenseKey != "" {
    // Validate license and activate premium features
}
```

## Docker Settings

These settings control how PentAGI interacts with Docker, which is used for terminal isolation and executing commands in a controlled environment. They're crucial for the security and functionality of tool execution.

| Option                       | Environment Variable               | Default Value          | Description                                                                                |
| ---------------------------- | ---------------------------------- | ---------------------- | ------------------------------------------------------------------------------------------ |
| DockerInside                 | `DOCKER_INSIDE`                    | `false`                | Set to `true` if PentAGI runs inside Docker and needs to access the host Docker daemon.    |
| DockerNetAdmin               | `DOCKER_NET_ADMIN`                 | `false`                | Set to `true` to grant the primary container NET_ADMIN capability for advanced networking. |
| DockerSocket                 | `DOCKER_SOCKET`                    | *(none)*               | Path to Docker socket for container management                                             |
| DockerNetwork                | `DOCKER_NETWORK`                   | *(none)*               | Docker network name for container communication                                            |
| DockerPublicIP               | `DOCKER_PUBLIC_IP`                 | `0.0.0.0`              | Public IP address for Docker containers' port bindings                                     |
| DockerWorkDir                | `DOCKER_WORK_DIR`                  | *(none)*               | Custom working directory inside Docker containers                                          |
| DockerDefaultImage           | `DOCKER_DEFAULT_IMAGE`             | `debian:latest`        | Default Docker image for containers when specific images fail                              |
| DockerDefaultImageForPentest | `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` | `vxcontrol/kali-linux` | Default Docker image for penetration testing tasks                                         |


### Usage Details

The Docker settings are primarily used in `pkg/docker/client.go` which implements the Docker client interface used throughout the application. This client is responsible for creating, managing, and executing commands in Docker containers:

- **DockerInside**: Signals whether PentAGI is running inside a Docker container itself, which affects how volumes and sockets are mounted:
  ```go
  inside := cfg.DockerInside
  ```

- **DockerSocket**: Specifies the path to the Docker socket, which is crucial for container management:
  ```go
  if cfg.DockerSocket != "" {
      socket = cfg.DockerSocket
  }
  ```

- **DockerNetwork**: Sets the network that containers should join, enabling container-to-container communication:
  ```go
  network := cfg.DockerNetwork

  // Used when creating network configuration
  if dc.network != "" {
      networkingConfig = &network.NetworkingConfig{
          EndpointsConfig: map[string]*network.EndpointSettings{
              dc.network: {},
          },
      }
  }
  ```

- **DockerPublicIP**: Defines the IP address to bind container ports to, making services accessible:
  ```go
  publicIP := cfg.DockerPublicIP

  // Used when setting up port bindings
  hostConfig.PortBindings[natPort] = []nat.PortBinding{
      {
          HostIP:   dc.publicIP,
          HostPort: fmt.Sprintf("%d", port),
      },
  }
  ```

- **DockerWorkDir**: Provides a custom working directory path to use inside containers:
  ```go
  hostDir := getHostDataDir(ctx, cli, dataDir, cfg.DockerWorkDir)
  ```

- **DockerDefaultImage**: Specifies the fallback image to use when requested images aren't available:
  ```go
  defImage := strings.ToLower(cfg.DockerDefaultImage)
  if defImage == "" {
      defImage = defaultImage
  }
  ```

This client is used by the tools executor to run commands in isolated containers, providing a secure environment for AI agents to execute terminal commands.

## Server Settings

These settings control the HTTP and GraphQL server that forms the backend API of PentAGI.

| Option       | Environment Variable | Default Value | Description                      |
| ------------ | -------------------- | ------------- | -------------------------------- |
| ServerPort   | `SERVER_PORT`        | `8080`        | Port for the HTTP server         |
| ServerHost   | `SERVER_HOST`        | `0.0.0.0`     | Host address for the HTTP server |
| ServerUseSSL | `SERVER_USE_SSL`     | `false`       | Enable SSL for the HTTP server   |
| ServerSSLKey | `SERVER_SSL_KEY`     | *(none)*      | Path to SSL key file             |
| ServerSSLCrt | `SERVER_SSL_CRT`     | *(none)*      | Path to SSL certificate file     |

### Usage Details

These settings are used in `main.go` to configure and start the HTTP server:

```go
// Build the listen address from host and port
listen := net.JoinHostPort(cfg.ServerHost, strconv.Itoa(cfg.ServerPort))

// Conditionally use TLS based on SSL configuration
if cfg.ServerUseSSL && cfg.ServerSSLCrt != "" && cfg.ServerSSLKey != "" {
    err = r.RunTLS(listen, cfg.ServerSSLCrt, cfg.ServerSSLKey)
} else {
    err = r.Run(listen)
}
```

The settings determine:
- The IP address and port the server listens on
- Whether to use HTTPS (SSL/TLS) for secure connections
- The location of the SSL certificate and key files (when SSL is enabled)

These configurations are crucial for production deployments where proper server binding and secure communication are required.

## Frontend Settings

These settings control how the server serves frontend assets and handles Cross-Origin Resource Sharing (CORS) for API requests from browsers.

| Option      | Environment Variable | Default Value | Description                                                              |
| ----------- | -------------------- | ------------- | ------------------------------------------------------------------------ |
| StaticURL   | `STATIC_URL`         | *(none)*      | URL to serve static frontend assets from (enables reverse proxy mode)    |
| StaticDir   | `STATIC_DIR`         | `./fe`        | Directory containing frontend static files (used when not in proxy mode) |
| CorsOrigins | `CORS_ORIGINS`       | `*`           | Allowed origins for CORS requests (comma-separated)                      |

### Usage Details

The frontend settings are extensively used in `pkg/server/router.go` for configuring how the application serves the frontend:

- **StaticURL**: When set, enables reverse proxy mode where static assets are served from an external URL:
  ```go
  if cfg.StaticURL != nil && cfg.StaticURL.Scheme != "" && cfg.StaticURL.Host != "" {
      // Set up reverse proxy for static assets
      router.NoRoute(func(c *gin.Context) {
          req := c.Request.Clone(c.Request.Context())
          req.URL.Scheme = cfg.StaticURL.Scheme
          req.URL.Host = cfg.StaticURL.Host
          // ...
      })
  }
  ```

- **StaticDir**: When StaticURL is not set, specifies the local directory containing static frontend assets:
  ```go
  // Serve static files from local directory
  router.Use(static.Serve("/", static.LocalFile(cfg.StaticDir, true)))

  // Also used for finding index.html for SPA routes
  indexPath := filepath.Join(cfg.StaticDir, "index.html")
  ```

- **CorsOrigins**: Configures CORS policy for the API, controlling which origins can make requests:
  ```go
  // In GraphQL service initialization
  graphqlService := services.NewGraphqlService(db, baseURL, cfg.CorsOrigins, providers, controller, subscriptions)

  // In CORS middleware configuration
  if !slices.Contains(cfg.CorsOrigins, "*") {
      config.AllowCredentials = true
  }
  config.AllowOrigins = cfg.CorsOrigins
  ```

These settings are essential for:
- Supporting different deployment architectures (single server vs. separate frontend/backend)
- Enabling proper SPA routing for frontend applications
- Configuring security policies for cross-origin requests

## Authentication Settings

These settings control authentication mechanisms, including cookie-based sessions and OAuth providers for user login.

| Option                  | Environment Variable         | Default Value | Description                                            |
| ----------------------- | ---------------------------- | ------------- | ------------------------------------------------------ |
| CookieSigningSalt       | `COOKIE_SIGNING_SALT`        | *(none)*      | Salt for signing and securing cookies used in sessions |
| PublicURL               | `PUBLIC_URL`                 | *(none)*      | Public URL for auth callbacks from OAuth providers     |
| OAuthGoogleClientID     | `OAUTH_GOOGLE_CLIENT_ID`     | *(none)*      | Google OAuth client ID for authentication              |
| OAuthGoogleClientSecret | `OAUTH_GOOGLE_CLIENT_SECRET` | *(none)*      | Google OAuth client secret                             |
| OAuthGithubClientID     | `OAUTH_GITHUB_CLIENT_ID`     | *(none)*      | GitHub OAuth client ID for authentication              |
| OAuthGithubClientSecret | `OAUTH_GITHUB_CLIENT_SECRET` | *(none)*      | GitHub OAuth client secret                             |

### Usage Details

The authentication settings are used in `pkg/server/router.go` to set up authentication middleware and OAuth providers:

- **CookieSigningSalt**: Used to secure cookies for session management:
  ```go
  // Used in auth middleware for authentication checks
  authMiddleware := auth.NewAuthMiddleware(baseURL, cfg.CookieSigningSalt)

  // Used for cookie store creation
  cookieStore := cookie.NewStore(auth.MakeCookieStoreKey(cfg.CookieSigningSalt)...)
  router.Use(sessions.Sessions("auth", cookieStore))
  ```

- **PublicURL**: The base URL for OAuth callback endpoints, crucial for redirects after authentication:
  ```go
  publicURL, err := url.Parse(cfg.PublicURL)
  ```

- **OAuth Provider Settings**: Used to configure authentication with Google and GitHub:
  ```go
  // Google OAuth setup
  if publicURL != nil && cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
      googleOAuth := oauth.NewGoogleOAuthController(
          cfg.OAuthGoogleClientID,
          cfg.OAuthGoogleClientSecret,
          *publicURL,
      )
      // ...
  }

  // GitHub OAuth setup
  if publicURL != nil && cfg.OAuthGithubClientID != "" && cfg.OAuthGithubClientSecret != "" {
      githubOAuth := oauth.NewGithubOAuthController(
          cfg.OAuthGithubClientID,
          cfg.OAuthGithubClientSecret,
          *publicURL,
      )
      // ...
  }
  ```

These settings are essential for:
- Secure user authentication and session management
- Supporting social login through OAuth providers
- Enabling proper redirects in the authentication flow

## Web Scraper Settings

These settings control the web scraper service used for browsing websites and taking screenshots, which allows AI agents to interact with web content.

| Option            | Environment Variable  | Default Value | Description                                               |
| ----------------- | --------------------- | ------------- | --------------------------------------------------------- |
| ScraperPublicURL  | `SCRAPER_PUBLIC_URL`  | *(none)*      | Public URL for accessing the scraper service from clients |
| ScraperPrivateURL | `SCRAPER_PRIVATE_URL` | *(none)*      | Private URL for internal scraper service access           |

### Usage Details

The scraper settings are extensively used in the tools executor to provide web browsing capabilities to AI agents:

```go
// In various tool functions in pkg/tools/tools.go
browseTool = &functions.BrowseFunc{
    scPrvURL: fte.cfg.ScraperPrivateURL,
    scPubURL: fte.cfg.ScraperPublicURL,
    // ...
}

screenshotTool = &functions.ScreenshotFunc{
    scPrvURL: fte.cfg.ScraperPrivateURL,
    scPubURL: fte.cfg.ScraperPublicURL,
    // ...
}
```

These URLs serve different purposes:
- **ScraperPublicURL**: Used when generating URLs that will be accessed by the client (browser)
- **ScraperPrivateURL**: Used for internal communication between the backend and the scraper service

The scraper settings enable critical functionality:
- Web browsing capabilities for AI agents
- Screenshot capturing for web content analysis
- Web information gathering for research tasks

## LLM Provider Settings

These settings control the integration with various Large Language Model (LLM) providers, including OpenAI, Anthropic, and custom providers.

### OpenAI

| Option          | Environment Variable | Default Value               | Description                        |
| --------------- | -------------------- | --------------------------- | ---------------------------------- |
| OpenAIKey       | `OPEN_AI_KEY`        | *(none)*                    | API key for OpenAI services        |
| OpenAIServerURL | `OPEN_AI_SERVER_URL` | `https://api.openai.com/v1` | Server URL for OpenAI API requests |

### Anthropic

| Option             | Environment Variable   | Default Value                  | Description                           |
| ------------------ | ---------------------- | ------------------------------ | ------------------------------------- |
| AnthropicAPIKey    | `ANTHROPIC_API_KEY`    | *(none)*                       | API key for Anthropic Claude services |
| AnthropicServerURL | `ANTHROPIC_SERVER_URL` | `https://api.anthropic.com/v1` | Server URL for Anthropic API requests |

### Ollama LLM Provider

| Option                        | Environment Variable                | Default Value               | Description                                       |
| ----------------------------- | ----------------------------------- | --------------------------- | ------------------------------------------------- |
| OllamaServerURL               | `OLLAMA_SERVER_URL`                 | *(none)*                    | Server URL for Ollama API requests                |
| OllamaServerModel             | `OLLAMA_SERVER_MODEL`               | `llama3.1:8b-instruct-q8_0` | Default model to use for inference                |
| OllamaServerConfig            | `OLLAMA_SERVER_CONFIG_PATH`         | *(none)*                    | Path to config file for Ollama provider options   |
| OllamaServerPullModelsTimeout | `OLLAMA_SERVER_PULL_MODELS_TIMEOUT` | `600`                       | Timeout in seconds for model downloads            |
| OllamaServerPullModelsEnabled | `OLLAMA_SERVER_PULL_MODELS_ENABLED` | `false`                     | Automatically download required models on startup |
| OllamaServerLoadModelsEnabled | `OLLAMA_SERVER_LOAD_MODELS_ENABLED` | `false`                     | Load available models list from server API        |

**Note:** When `OllamaServerLoadModelsEnabled=false`, only the default model is available. Enable this to see all installed models in the UI.

### Google AI (Gemini) LLM Provider

| Option          | Environment Variable | Default Value                               | Description                           |
| --------------- | -------------------- | ------------------------------------------- | ------------------------------------- |
| GeminiAPIKey    | `GEMINI_API_KEY`     | *(none)*                                    | API key for Google AI Gemini services |
| GeminiServerURL | `GEMINI_SERVER_URL`  | `https://generativelanguage.googleapis.com` | Server URL for Gemini API requests    |

### AWS Bedrock LLM Provider

| Option              | Environment Variable        | Default Value | Description                                                                            |
| ------------------- | --------------------------- | ------------- | -------------------------------------------------------------------------------------- |
| BedrockRegion       | `BEDROCK_REGION`            | `us-east-1`   | AWS region for Bedrock service                                                         |
| BedrockAccessKey    | `BEDROCK_ACCESS_KEY_ID`     | *(none)*      | AWS access key ID for Bedrock authentication                                           |
| BedrockSecretKey    | `BEDROCK_SECRET_ACCESS_KEY` | *(none)*      | AWS secret access key for Bedrock authentication                                       |
| BedrockSessionToken | `BEDROCK_SESSION_TOKEN`     | *(none)*      | AWS session token for temporary credentials (optional, required for STS/assumed roles) |
| BedrockServerURL    | `BEDROCK_SERVER_URL`        | *(none)*      | Optional custom endpoint URL for Bedrock service                                       |

### Custom LLM Provider

| Option                     | Environment Variable            | Default Value | Description                                                                  |
| -------------------------- | ------------------------------- | ------------- | ---------------------------------------------------------------------------- |
| LLMServerURL               | `LLM_SERVER_URL`                | *(none)*      | Server URL for custom LLM provider                                           |
| LLMServerKey               | `LLM_SERVER_KEY`                | *(none)*      | API key for custom LLM provider                                              |
| LLMServerModel             | `LLM_SERVER_MODEL`              | *(none)*      | Model name for custom LLM provider                                           |
| LLMServerConfig            | `LLM_SERVER_CONFIG_PATH`        | *(none)*      | Path to config file for custom LLM provider options                          |
| LLMServerProvider          | `LLM_SERVER_PROVIDER`           | *(none)*      | Provider name prefix for model names (useful for LiteLLM proxy)              |
| LLMServerLegacyReasoning   | `LLM_SERVER_LEGACY_REASONING`   | `false`       | Controls reasoning format in API requests                                    |
| LLMServerPreserveReasoning | `LLM_SERVER_PRESERVE_REASONING` | `false`       | Preserve reasoning content in multi-turn conversations (required by some providers) |

### Usage Details

The LLM provider settings are used in `pkg/providers` modules to initialize and configure the appropriate language model providers:

- **OpenAI Settings**: Used in `pkg/providers/openai/openai.go` to create the OpenAI client:
  ```go
  baseURL := cfg.OpenAIServerURL

  client, err := openai.New(
      openai.WithToken(cfg.OpenAIKey),
      openai.WithModel(OpenAIAgentModel),
      openai.WithBaseURL(baseURL),
      // ...
  )
  ```

- **Anthropic Settings**: Used in `pkg/providers/anthropic/anthropic.go` to create the Anthropic client:
  ```go
  baseURL := cfg.AnthropicServerURL

  client, err := anthropic.New(
      anthropic.WithToken(cfg.AnthropicAPIKey),
      anthropic.WithBaseURL(baseURL),
      // ...
  )
  ```

- **Ollama Settings**: Used in `pkg/providers/ollama/ollama.go` to create the Ollama client:
  ```go
  serverURL := cfg.OllamaServerURL

  client, err := ollama.New(
      ollama.WithServerURL(serverURL),
      ollama.WithHTTPClient(httpClient),
      ollama.WithModel(OllamaAgentModel),
      ollama.WithPullModel(),
  )

  // Load provider options from config file if specified
  if cfg.OllamaServerConfig != "" {
      configData, err := os.ReadFile(cfg.OllamaServerConfig)
      providerConfig, err := BuildProviderConfig(cfg, configData)
      // ...
  }
  ```

- **Gemini Settings**: Used in `pkg/providers/gemini/gemini.go` to create the Google AI client:
  ```go
  opts := []googleai.Option{
      googleai.WithRest(),
      googleai.WithAPIKey(cfg.GeminiAPIKey),
      googleai.WithEndpoint(cfg.GeminiServerURL),
      googleai.WithDefaultModel(GeminiAgentModel),
  }

  client, err := googleai.New(context.Background(), opts...)
  ```

- **Bedrock Settings**: Used in `pkg/providers/bedrock/bedrock.go` to create the AWS Bedrock client:
  ```go
  opts := []func(*bconfig.LoadOptions) error{
      bconfig.WithRegion(cfg.BedrockRegion),
      bconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
          cfg.BedrockAccessKey,
          cfg.BedrockSecretKey,
          cfg.BedrockSessionToken,
      )),
  }

  if cfg.BedrockServerURL != "" {
      opts = append(opts, bconfig.WithBaseEndpoint(cfg.BedrockServerURL))
  }

  bcfg, err := bconfig.LoadDefaultConfig(context.Background(), opts...)
  bclient := bedrockruntime.NewFromConfig(bcfg)

  client, err := bedrock.New(
      bedrock.WithClient(bclient),
      bedrock.WithModel(BedrockAgentModel),
      bedrock.WithConverseAPI(),
  )
  ```

  The `BedrockSessionToken` is optional and only required when using temporary AWS credentials (e.g., from STS, assumed roles, or MFA-enabled IAM users). For permanent IAM user credentials, leave this field empty.

- **Custom LLM Settings**: Used in `pkg/providers/custom/custom.go` to create a custom LLM client:
  ```go
  baseKey := cfg.LLMServerKey
  baseURL := cfg.LLMServerURL
  baseModel := cfg.LLMServerModel

  client, err := openai.New(
      openai.WithToken(baseKey),
      openai.WithModel(baseModel),
      openai.WithBaseURL(baseURL),
      // ...
  )

  // Load provider options from config file if specified
  if cfg.LLMServerConfig != "" {
      providerConfig, err := LoadConfig(cfg.LLMServerConfig, simple)
      // ...
  }
  ```

- **LLMServerLegacyReasoning**: Controls the reasoning format used in API requests to custom LLM providers:
  ```go
  // Used in custom provider to determine reasoning format
  if cfg.LLMServerLegacyReasoning {
      // Uses legacy string-based reasoning_effort parameter
  } else {
      // Uses modern structured reasoning object with max_tokens
  }
  ```
  - `false` (default): Uses modern format where reasoning is sent as a structured object with `max_tokens` parameter
  - `true`: Uses legacy format with string-based `reasoning_effort` parameter

This setting is important when working with different LLM providers as they may expect different reasoning formats in their API requests. If you encounter reasoning-related errors with custom providers, try changing this setting.

- **LLMServerPreserveReasoning**: Controls whether reasoning content is preserved and sent back in multi-turn conversations:
  ```go
  // Used in custom provider to preserve reasoning content
  if cfg.LLMServerPreserveReasoning {
      // Preserves and returns reasoning_content in assistant messages
  }
  ```
  - `false` (default): Reasoning content is not preserved in conversation history
  - `true`: Reasoning content is preserved and sent in subsequent API calls

This setting is required by some LLM providers (e.g., Moonshot) that return errors like "thinking is enabled but reasoning_content is missing in assistant tool call message" when reasoning content is not included in multi-turn conversations. Enable this setting if your provider requires reasoning content to be preserved across conversation turns.

The provider registration is managed in `pkg/providers/providers.go`:

```go
// Provider registration based on available credentials
if cfg.OpenAIKey != "" {
    p, err := openai.New(cfg, defaultConfigs[provider.ProviderOpenAI])
    if err != nil {
        return nil, fmt.Errorf("failed to create openai provider: %w", err)
    }
    providers[provider.DefaultProviderNameOpenAI] = p
}

if cfg.AnthropicAPIKey != "" {
    p, err := anthropic.New(cfg, defaultConfigs[provider.ProviderAnthropic])
    if err != nil {
        return nil, fmt.Errorf("failed to create anthropic provider: %w", err)
    }
    providers[provider.DefaultProviderNameAnthropic] = p
}

if cfg.GeminiAPIKey != "" {
    p, err := gemini.New(cfg, defaultConfigs[provider.ProviderGemini])
    if err != nil {
        return nil, fmt.Errorf("failed to create gemini provider: %w", err)
    }
    providers[provider.DefaultProviderNameGemini] = p
}

if cfg.BedrockAccessKey != "" && cfg.BedrockSecretKey != "" {
    p, err := bedrock.New(cfg, defaultConfigs[provider.ProviderBedrock])
    if err != nil {
        return nil, fmt.Errorf("failed to create bedrock provider: %w", err)
    }
    providers[provider.DefaultProviderNameBedrock] = p
}

if cfg.OllamaServerURL != "" {
    p, err := ollama.New(cfg, defaultConfigs[provider.ProviderOllama])
    if err != nil {
        return nil, fmt.Errorf("failed to create ollama provider: %w", err)
    }
    providers[provider.DefaultProviderNameOllama] = p
}

if cfg.LLMServerURL != "" && (cfg.LLMServerModel != "" || cfg.LLMServerConfig != "") {
    p, err := custom.New(cfg, defaultConfigs[provider.ProviderCustom])
    if err != nil {
        return nil, fmt.Errorf("failed to create custom provider: %w", err)
    }
    providers[provider.DefaultProviderNameCustom] = p
}
```

These settings are critical for:
- Connecting to various LLM providers for AI capabilities
- Supporting multiple model options for different tasks
- Enabling custom or self-hosted LLM solutions
- Configuring specific model behaviors and parameters

## Embedding Settings

These settings control the vector embedding service used for semantic search and similarity matching, which is fundamental for PentAGI's intelligent search capabilities.

| Option                 | Environment Variable        | Default Value | Description                                                                |
| ---------------------- | --------------------------- | ------------- | -------------------------------------------------------------------------- |
| EmbeddingURL           | `EMBEDDING_URL`             | *(none)*      | Server URL for embedding provider (overrides provider-specific URLs)       |
| EmbeddingKey           | `EMBEDDING_KEY`             | *(none)*      | API key for embedding provider (overrides provider-specific keys)          |
| EmbeddingModel         | `EMBEDDING_MODEL`           | *(none)*      | Model name for embedding generation                                        |
| EmbeddingStripNewLines | `EMBEDDING_STRIP_NEW_LINES` | `true`        | Whether to strip newlines before embedding (improves quality)              |
| EmbeddingBatchSize     | `EMBEDDING_BATCH_SIZE`      | `512`         | Batch size for embedding operations (affects memory usage and performance) |
| EmbeddingProvider      | `EMBEDDING_PROVIDER`        | `openai`      | Provider for embeddings (openai, ollama, mistral, jina, huggingface)       |

### Usage Details

The embedding settings are extensively used in `pkg/providers/embeddings/embedder.go` to configure the vector embedding service:

- **EmbeddingProvider**: Determines which embedding provider to use:
  ```go
  switch cfg.EmbeddingProvider {
  case "openai":
      return newOpenAIEmbedder(ctx, cfg)
  case "ollama":
      return newOllamaEmbedder(ctx, cfg)
  case "mistral":
      return newMistralEmbedder(ctx, cfg)
  case "jina":
      return newJinaEmbedder(ctx, cfg)
  case "huggingface":
      return newHuggingFaceEmbedder(ctx, cfg)
  default:
      return &embedder{nil}, fmt.Errorf("unsupported embedding provider: %s", cfg.EmbeddingProvider)
  }
  ```

- **Provider-specific configurations**: Used to configure each embedding provider with appropriate options:
  ```go
  // Example for OpenAI embeddings
  if cfg.EmbeddingURL != "" {
      opts = append(opts, openai.WithBaseURL(cfg.EmbeddingURL))
  } else if cfg.OpenAIServerURL != "" {
      opts = append(opts, openai.WithBaseURL(cfg.OpenAIServerURL))
  }

  if cfg.EmbeddingKey != "" {
      opts = append(opts, openai.WithToken(cfg.EmbeddingKey))
  } else if cfg.OpenAIKey != "" {
      opts = append(opts, openai.WithToken(cfg.OpenAIKey))
  }

  if cfg.EmbeddingModel != "" {
      opts = append(opts, openai.WithEmbeddingModel(cfg.EmbeddingModel))
  }
  ```

- **Embedding behavior configuration**: Controls how text is processed for embeddings:
  ```go
  embeddings.WithStripNewLines(cfg.EmbeddingStripNewLines),
  embeddings.WithBatchSize(cfg.EmbeddingBatchSize),
  ```

These settings are essential for:
- Configuring semantic search capabilities
- Determining which embedding model to use
- Optimizing embedding performance and quality
- Supporting multiple embedding providers for flexibility

## Summarizer Settings

These settings control the text summarization behavior used for condensing long conversations and improving context management in AI interactions. The summarization system is a critical component that allows PentAGI to maintain coherent, long-running conversations while managing token usage effectively.

| Option                   | Environment Variable             | Default Value | Description                                                |
| ------------------------ | -------------------------------- | ------------- | ---------------------------------------------------------- |
| SummarizerPreserveLast   | `SUMMARIZER_PRESERVE_LAST`       | `true`        | Preserve the last message in summarization                 |
| SummarizerUseQA          | `SUMMARIZER_USE_QA`              | `true`        | Use question-answer format for summarization               |
| SummarizerSumHumanInQA   | `SUMMARIZER_SUM_MSG_HUMAN_IN_QA` | `false`       | Include human messages in QA summaries                     |
| SummarizerLastSecBytes   | `SUMMARIZER_LAST_SEC_BYTES`      | `51200`       | Bytes to preserve from the last section (50KB)             |
| SummarizerMaxBPBytes     | `SUMMARIZER_MAX_BP_BYTES`        | `16384`       | Maximum bytes for bullet points summarization (16KB)       |
| SummarizerMaxQASections  | `SUMMARIZER_MAX_QA_SECTIONS`     | `10`          | Maximum QA sections to include                             |
| SummarizerMaxQABytes     | `SUMMARIZER_MAX_QA_BYTES`        | `65536`       | Maximum bytes for QA summarization (64KB)                  |
| SummarizerKeepQASections | `SUMMARIZER_KEEP_QA_SECTIONS`    | `1`           | Number of recent QA sections to keep without summarization |

### Usage Details and Impact on System Behavior

The summarizer settings map directly to the `SummarizerConfig` structure that controls the chain summarization algorithm in `pkg/csum`. These settings work together to implement a sophisticated, multi-strategy approach to managing conversation context:

#### Core Summarization Strategies and Their Parameters

1. **Section Summarization** - Always active, ensures all older sections (except the last one) consist of a single summarized pair
   - No specific parameters control this as it's a fundamental part of the algorithm
   - Prevents unbounded growth by consolidating completed conversation sections

2. **Last Section Management** (`SummarizerPreserveLast` and `SummarizerLastSecBytes`)
   - Controls how the current/active conversation section is managed
   - When `SummarizerPreserveLast = true`, older messages within the last section will be summarized when the section exceeds `SummarizerLastSecBytes` bytes
   - A reserve space of 25% is automatically maintained to accommodate new messages without triggering frequent re-summarization
   - Individual oversized pairs are summarized separately if they exceed `SummarizerMaxBPBytes`

3. **QA Pair Summarization** (`SummarizerUseQA`, `SummarizerMaxQASections`, `SummarizerMaxQABytes`, `SummarizerSumHumanInQA`)
   - When `SummarizerUseQA = true`, creates larger summarization units focused on question-answer patterns
   - Preserves the most recent `SummarizerMaxQASections` sections as long as they don't exceed `SummarizerMaxQABytes` total
   - If `SummarizerSumHumanInQA = true`, human messages are also summarized; otherwise, they're preserved verbatim

#### Deep Dive: Parameter Impact and Recommendations

**`SummarizerPreserveLast`** (Default: `true`)
- **Purpose**: Controls whether the last (active) section has size management applied
- **Impact**: When enabled, prevents the active conversation from growing indefinitely
- **When to adjust**:
  - Enable (default) for production systems and long-running conversations
  - Disable only for debugging or when you need to preserve the complete conversation history regardless of size

**`SummarizerLastSecBytes`** (Default: `51200` - 50KB)
- **Purpose**: Maximum byte size for the last (active) section before summarization begins
- **Impact**: Directly controls how much conversation history is preserved verbatim in the active section
- **When to adjust**:
  - Increase for models with larger context windows to maintain more conversation detail
  - Decrease for models with smaller context to prevent token limits from being exceeded
  - Balance with `SummarizerMaxBPBytes` to ensure coherent summarization

**`SummarizerMaxBPBytes`** (Default: `16384` - 16KB)
- **Purpose**: Maximum byte size for individual body pairs (typically AI responses)
- **Impact**: Controls when individual large responses get summarized, even if the overall section is under limit
- **When to adjust**:
  - Increase if your use case involves long but important AI responses that shouldn't be summarized
  - Decrease if you want more aggressive summarization of lengthy responses

**`SummarizerUseQA`** (Default: `true`)
- **Purpose**: Enables question-answer style summarization that creates more cohesive summaries
- **Impact**: When enabled, creates a new first section with a summary of older interactions, preserving recent sections
- **When to adjust**:
  - Enable (default) for more coherent, organized summaries focused on main topics
  - Disable if you prefer simpler, section-by-section summarization without cross-section analysis

**`SummarizerMaxQASections`** (Default: `10`)
- **Purpose**: Maximum number of recent sections to preserve when using QA-style summarization
- **Impact**: Directly controls how many conversation turns remain intact after QA summarization
- **When to adjust**:
  - Increase to preserve more recent conversation context (at the cost of token usage)
  - Decrease to create more compact conversation histories, focusing on only the very recent exchanges

**`SummarizerMaxQABytes`** (Default: `65536` - 64KB)
- **Purpose**: Maximum total byte size for preserved sections in QA-style summarization
- **Impact**: Sets an upper bound on memory used by preserved sections, regardless of section count
- **When to adjust**:
  - Increase for models with larger context windows or when detailed context is essential
  - Decrease for smaller context models or when prioritizing efficiency over context preservation

**`SummarizerSumHumanInQA`** (Default: `false`)
- **Purpose**: Controls whether human messages are summarized in QA-style summarization
- **Impact**: When false, human messages are preserved verbatim; when true, they are also summarized
- **When to adjust**:
  - Keep disabled (default) to preserve the exact wording of user queries
  - Enable only when human messages are very verbose and token efficiency is critical

**`SummarizerKeepQASections`** (Default: `1`)
- **Purpose**: Controls the number of recent QA sections to keep without summarization
- **Impact**: Directly controls how many recent conversation turns are preserved verbatim
- **When to adjust**:
  - Increase to preserve more recent conversation context
  - Decrease to create more compact conversation histories, focusing on only the very recent exchanges

### Summarization Effects on Agent Behavior

The summarization settings have significant effects on agent behavior:

1. **Context Retention vs. Token Efficiency**
   - More aggressive summarization (smaller byte limits) reduces token usage but may lose context details
   - More permissive settings (larger byte limits) preserve more context but increase token consumption

2. **Conversation Coherence**
   - Appropriate summarization helps the agent maintain a coherent understanding of the conversation
   - Over-aggressive summarization may cause the agent to lose important details or previous instructions
   - Under-aggressive summarization may lead to context overflow in longer conversations

3. **Response Quality**
   - QA-style summarization (`SummarizerUseQA = true`) typically improves response quality for complex tasks
   - Preserving human messages (`SummarizerSumHumanInQA = false`) helps maintain alignment with user intent
   - Appropriate `SummarizerMaxBPBytes` prevents loss of detailed information from complex AI responses

### Implementation Details

The summarizer settings are used in `pkg/providers/providers.go` to configure the summarization behavior:

```go
summarizer := provider.SummarizerSettings{
    PreserveLast:  cfg.SummarizerPreserveLast,
    UseQA:         cfg.SummarizerUseQA,
    SummHumanInQA: cfg.SummarizerSumHumanInQA,
    LastSecBytes:  cfg.SummarizerLastSecBytes,
    MaxBPBytes:    cfg.SummarizerMaxBPBytes,
    MaxQASections: cfg.SummarizerMaxQASections,
    MaxQABytes:    cfg.SummarizerMaxQABytes,
}
```

These settings are passed to various components through the chain summarization system:

```go
// In csum/chain_summary.go
func NewSummarizer(config SummarizerConfig) Summarizer {
    if config.PreserveLast {
        if config.LastSecBytes <= 0 {
            config.LastSecBytes = maxLastSectionByteSize
        }
    }

    if config.UseQA {
        if config.MaxQASections <= 0 {
            config.MaxQASections = maxQAPairSections
        }
        if config.MaxQABytes <= 0 {
            config.MaxQABytes = maxQAPairByteSize
        }
    }

    if config.MaxBPBytes <= 0 {
        config.MaxBPBytes = maxSingleBodyPairByteSize
    }

    return &summarizer{config: config}
}
```

### Recommended Settings for Different Use Cases

1. **Long-running Assistant Conversations**
   ```
   SummarizerPreserveLast: true
   SummarizerLastSecBytes: 51200 (50KB)
   SummarizerMaxBPBytes: 16384 (16KB)
   SummarizerUseQA: true
   SummarizerMaxQASections: 10
   SummarizerMaxQABytes: 65536 (64KB)
   SummarizerSumHumanInQA: false
   SummarizerKeepQASections: 1
   ```
   The default settings are optimized for assistant-style conversations. They maintain a good balance between context retention and token efficiency.

2. **Technical Problem-Solving with Large Context Models**
   ```
   SummarizerPreserveLast: true
   SummarizerLastSecBytes: 81920 (80KB)
   SummarizerMaxBPBytes: 32768 (32KB)
   SummarizerUseQA: true
   SummarizerMaxQASections: 15
   SummarizerMaxQABytes: 102400 (100KB)
   SummarizerSumHumanInQA: false
   SummarizerKeepQASections: 1
   ```
   Increased limits to preserve more technical details when using models with large context windows (e.g., GPT-4).

3. **Limited Context Models**
   ```
   SummarizerPreserveLast: true
   SummarizerLastSecBytes: 25600 (25KB)
   SummarizerMaxBPBytes: 8192 (8KB)
   SummarizerUseQA: true
   SummarizerMaxQASections: 5
   SummarizerMaxQABytes: 32768 (32KB)
   SummarizerSumHumanInQA: true
   SummarizerKeepQASections: 1
   ```
   More aggressive summarization for models with smaller context windows (e.g., smaller or older LLMs).

4. **Debugging or Analysis (Maximum Context Preservation)**
   ```
   SummarizerPreserveLast: false
   SummarizerUseQA: false
   SummarizerKeepQASections: 0
   ```
   Disables active summarization to preserve the complete conversation history for debugging purposes. Note that this can lead to context overflow in long conversations.

## Assistant Settings

These settings control the behavior of the AI assistant functionality, including whether to use multi-agent delegation and assistant-specific summarization settings.

| Option                            | Environment Variable                    | Default Value | Description                                                             |
| --------------------------------- | --------------------------------------- | ------------- | ----------------------------------------------------------------------- |
| AssistantUseAgents                | `ASSISTANT_USE_AGENTS`                  | `false`       | Controls the default value for agent usage when creating new assistants |
| AssistantSummarizerPreserveLast   | `ASSISTANT_SUMMARIZER_PRESERVE_LAST`    | `true`        | Whether to preserve all messages in the assistant's last section        |
| AssistantSummarizerLastSecBytes   | `ASSISTANT_SUMMARIZER_LAST_SEC_BYTES`   | `76800`       | Maximum byte size for assistant's last section (75KB)                   |
| AssistantSummarizerMaxBPBytes     | `ASSISTANT_SUMMARIZER_MAX_BP_BYTES`     | `16384`       | Maximum byte size for a single body pair in assistant context (16KB)    |
| AssistantSummarizerMaxQASections  | `ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS`  | `7`           | Maximum QA sections to preserve in assistant context                    |
| AssistantSummarizerMaxQABytes     | `ASSISTANT_SUMMARIZER_MAX_QA_BYTES`     | `76800`       | Maximum byte size for assistant's QA sections (75KB)                    |
| AssistantSummarizerKeepQASections | `ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS` | `3`           | Number of recent QA sections to preserve without summarization          |

### Usage Details

The assistant settings are used to configure the behavior of the AI assistant and its context management:

- **AssistantUseAgents**: Controls the default state of the "Use Agents" toggle when creating new assistants in the UI:
  ```go
  // This setting affects the initial state when creating assistants
  // Users can always override this by toggling the "Use Agents" button in the UI
  ```
  - `false` (default): New assistants are created with agent delegation disabled by default
  - `true`: New assistants are created with agent delegation enabled by default

- **Assistant Summarizer Settings**: These provide dedicated summarization configuration for assistant instances, typically allowing for more memory retention compared to the global settings:
  ```go
  // Assistant summarizer configuration provides more context retention
  // compared to global settings, preserving more recent conversation history
  // while still ensuring efficient token usage
  ```

The assistant summarizer configuration is designed to provide more memory for context retention compared to the global settings, preserving more recent conversation history while still ensuring efficient token usage.

### Recommended Assistant Settings for Different Use Cases

1. **Standard Assistant Conversations**
   ```
   AssistantUseAgents: false
   AssistantSummarizerPreserveLast: true
   AssistantSummarizerLastSecBytes: 76800 (75KB)
   AssistantSummarizerMaxBPBytes: 16384 (16KB)
   AssistantSummarizerMaxQASections: 7
   AssistantSummarizerMaxQABytes: 76800 (75KB)
   AssistantSummarizerKeepQASections: 3
   ```
   The default settings provide a balance between context retention and performance for typical assistant interactions.

2. **Multi-Agent Assistant Workflows**
   ```
   AssistantUseAgents: true
   AssistantSummarizerPreserveLast: true
   AssistantSummarizerLastSecBytes: 102400 (100KB)
   AssistantSummarizerMaxBPBytes: 32768 (32KB)
   AssistantSummarizerMaxQASections: 10
   AssistantSummarizerMaxQABytes: 102400 (100KB)
   AssistantSummarizerKeepQASections: 5
   ```
   Enhanced settings for complex workflows that benefit from agent delegation with increased context preservation.

3. **Resource-Constrained Assistant**
   ```
   AssistantUseAgents: false
   AssistantSummarizerPreserveLast: true
   AssistantSummarizerLastSecBytes: 51200 (50KB)
   AssistantSummarizerMaxBPBytes: 16384 (16KB)
   AssistantSummarizerMaxQASections: 5
   AssistantSummarizerMaxQABytes: 51200 (50KB)
   AssistantSummarizerKeepQASections: 2
   ```
   More conservative settings for environments with limited resources or smaller context models.

## Search Engine Settings

These settings control the integration with various search engines used for web search capabilities, providing AI agents with up-to-date information from the internet.

### DuckDuckGo Search

| Option            | Environment Variable | Default Value | Description                                |
| ----------------- | -------------------- | ------------- | ------------------------------------------ |
| DuckDuckGoEnabled | `DUCKDUCKGO_ENABLED` | `true`        | Enable or disable DuckDuckGo Search engine |

### Sploitus Search

| Option          | Environment Variable | Default Value | Description                                                 |
| --------------- | -------------------- | ------------- | ----------------------------------------------------------- |
| SploitusEnabled | `SPLOITUS_ENABLED`   | `true`        | Enable or disable Sploitus exploit and vulnerability search |

### Google Search

| Option       | Environment Variable | Default Value | Description                                              |
| ------------ | -------------------- | ------------- | -------------------------------------------------------- |
| GoogleAPIKey | `GOOGLE_API_KEY`     | *(none)*      | API key for Google Search                                |
| GoogleCXKey  | `GOOGLE_CX_KEY`      | *(none)*      | Custom Search Engine ID for Google Search                |
| GoogleLRKey  | `GOOGLE_LR_KEY`      | `lang_en`     | Language restriction for Google Search (e.g., `lang_en`) |

### Traversaal Search

| Option           | Environment Variable | Default Value | Description                          |
| ---------------- | -------------------- | ------------- | ------------------------------------ |
| TraversaalAPIKey | `TRAVERSAAL_API_KEY` | *(none)*      | API key for Traversaal search engine |

### Tavily Search

| Option       | Environment Variable | Default Value | Description                      |
| ------------ | -------------------- | ------------- | -------------------------------- |
| TavilyAPIKey | `TAVILY_API_KEY`     | *(none)*      | API key for Tavily search engine |

### Perplexity Search

| Option                | Environment Variable      | Default Value | Description                                                  |
| --------------------- | ------------------------- | ------------- | ------------------------------------------------------------ |
| PerplexityAPIKey      | `PERPLEXITY_API_KEY`      | *(none)*      | API key for Perplexity search engine                         |
| PerplexityModel       | `PERPLEXITY_MODEL`        | `sonar`       | Model to use for Perplexity search                           |
| PerplexityContextSize | `PERPLEXITY_CONTEXT_SIZE` | `low`         | Context size for Perplexity search (`low`, `medium`, `high`) |

### Searxng Search

| Option            | Environment Variable | Default Value | Description                                                         |
| ----------------- | -------------------- | ------------- | ------------------------------------------------------------------- |
| SearxngURL        | `SEARXNG_URL`        | *(none)*      | Base URL for Searxng meta search engine instance                    |
| SearxngCategories | `SEARXNG_CATEGORIES` | `general`     | Search categories to use (e.g., `general`, `news`, `web`)           |
| SearxngLanguage   | `SEARXNG_LANGUAGE`   | *(none)*      | Language filter for search results (e.g., `en`, `ch`)               |
| SearxngSafeSearch | `SEARXNG_SAFESEARCH` | `0`           | Safe search filter level (`0` = none, `1` = moderate, `2` = strict) |
| SearxngTimeRange  | `SEARXNG_TIME_RANGE` | *(none)*      | Time range filter (e.g., `day`, `month`, `year`)                    |

### Usage Details

The search engine settings are used in `pkg/tools/tools.go` to configure various search providers that AI agents can use:

```go
// Google Search configuration
googleSearch: &functions.GoogleSearchFunc{
    apiKey:    fte.cfg.GoogleAPIKey,
    cxKey:     fte.cfg.GoogleCXKey,
    lrKey:     fte.cfg.GoogleLRKey,
    proxyURL:  fte.cfg.ProxyURL,
},

// Traversaal Search configuration
traversaalSearch: &functions.TraversaalSearchFunc{
    apiKey:    fte.cfg.TraversaalAPIKey,
    proxyURL:  fte.cfg.ProxyURL,
},

// Tavily Search configuration
tavilySearch: &functions.TavilySearchFunc{
    apiKey:     fte.cfg.TavilyAPIKey,
    proxyURL:   fte.cfg.ProxyURL,
    summarizer: cfg.Summarizer,
},

// Perplexity Search configuration
perplexitySearch: &functions.PerplexitySearchFunc{
    apiKey:      fte.cfg.PerplexityAPIKey,
    proxyURL:    fte.cfg.ProxyURL,
    model:       fte.cfg.PerplexityModel,
    contextSize: fte.cfg.PerplexityContextSize,
    summarizer:  cfg.Summarizer,
},

// Sploitus Search configuration
sploitus := NewSploitusTool(
    fte.flowID,
    cfg.TaskID,
    cfg.SubtaskID,
    fte.cfg.SploitusEnabled,
    fte.cfg.ProxyURL,
    fte.slp,
)

// Searxng Search configuration
searxng := NewSearxngTool(
    fte.flowID,
    cfg.TaskID,
    cfg.SubtaskID,
    fte.cfg.SearxngURL,
    fte.cfg.SearxngCategories,
    fte.cfg.SearxngLanguage,
    fte.cfg.SearxngSafeSearch,
    fte.cfg.SearxngTimeRange,
    fte.cfg.ProxyURL,
    0, // timeout (will use default 30 seconds)
    fte.slp,
    cfg.Summarizer,
)
```

These settings enable:
- Access to multiple search engines for diverse information sources
- Configuration of search parameters like language, context size, and time range
- Integration of search capabilities into the AI agent's toolset
- Web information gathering with different search strategies
- Security research through Sploitus, providing access to exploit databases and CVE information
- Meta-search capabilities through Searxng, aggregating results from multiple search engines

Having multiple search engine options ensures redundancy and provides different search algorithms for varied information needs. Sploitus is specifically designed for security research, providing comprehensive exploit and vulnerability information essential for penetration testing. Searxng is particularly useful as it provides aggregated results from multiple search engines while offering enhanced privacy and customization options.

## Proxy Settings

These settings control the HTTP proxy used for outbound connections, which is important for network security and access control.

| Option              | Environment Variable    | Default Value | Description                                                      |
| ------------------- | ----------------------- | ------------- | ---------------------------------------------------------------- |
| ProxyURL            | `PROXY_URL`             | *(none)*      | URL for HTTP proxy (e.g., `http://user:pass@proxy:8080`)         |
| ExternalSSLCAPath   | `EXTERNAL_SSL_CA_PATH`  | *(none)*      | Path to trusted CA certificate for external LLM SSL connections  |
| ExternalSSLInsecure | `EXTERNAL_SSL_INSECURE` | `false`       | Skip SSL certificate verification for external connections       |

### Usage Details

The proxy settings are used in various places to configure HTTP clients for external API calls:

```go
// Example from openai.go, anthropic.go, and other provider files
if cfg.ProxyURL != "" {
    httpClient = &http.Client{
        Transport: &http.Transport{
            Proxy: func(req *http.Request) (*url.URL, error) {
                return url.Parse(cfg.ProxyURL)
            },
        },
    }
}
```

The proxy URL is also passed to various tools that make external requests:

```go
// In tools.go for search tools
googleSearch: &functions.GoogleSearchFunc{
    apiKey:    fte.cfg.GoogleAPIKey,
    cxKey:     fte.cfg.GoogleCXKey,
    lrKey:     fte.cfg.GoogleLRKey,
    proxyURL:  fte.cfg.ProxyURL,
},
```

The proxy setting is essential for:
- Routing all outbound API requests through a controlled proxy
- Implementing network-level security policies
- Enabling access to external services from restricted networks
- Monitoring and auditing external API usage

The SSL settings provide additional security configuration:

- **ExternalSSLCAPath**: Specifies a custom CA certificate for validating SSL connections to external services:
  ```go
  // Used in provider initialization to configure custom CA certificates
  if cfg.ExternalSSLCAPath != "" {
      caCert, err := os.ReadFile(cfg.ExternalSSLCAPath)
      // Configure TLS with custom CA
  }
  ```
  This is useful when connecting to LLM providers with self-signed certificates or internal CAs.

- **ExternalSSLInsecure**: Allows skipping SSL certificate verification:
  ```go
  // Used in HTTP client configuration
  if cfg.ExternalSSLInsecure {
      tlsConfig.InsecureSkipVerify = true
  }
  ```
  **Warning**: Only use this in development or trusted environments. Skipping certificate verification exposes connections to man-in-the-middle attacks.

## Graphiti Knowledge Graph Settings

These settings control the integration with Graphiti, a temporal knowledge graph system powered by Neo4j, for advanced semantic understanding and relationship tracking of AI agent operations.

| Option          | Environment Variable | Default Value           | Description                                            |
| --------------- | -------------------- | ----------------------- | ------------------------------------------------------ |
| GraphitiEnabled | `GRAPHITI_ENABLED`   | `false`                 | Enable or disable Graphiti knowledge graph integration |
| GraphitiURL     | `GRAPHITI_URL`       | `http://localhost:8001` | Base URL for Graphiti API service                      |
| GraphitiTimeout | `GRAPHITI_TIMEOUT`   | `30`                    | Timeout in seconds for Graphiti operations             |

### Usage Details

The Graphiti settings are used in `pkg/graphiti/client.go` and integrated throughout the provider system to automatically capture agent interactions and tool executions:

- **GraphitiEnabled**: Controls whether the knowledge graph integration is active:
  ```go
  // Check if Graphiti is enabled
  if !cfg.GraphitiEnabled {
      return &Client{enabled: false}, nil
  }
  ```

- **GraphitiURL**: Specifies the Graphiti API endpoint:
  ```go
  client := graphiti.NewClient(cfg.GraphitiURL, timeout, cfg.GraphitiEnabled)
  ```

- **GraphitiTimeout**: Sets the maximum time for knowledge graph operations:
  ```go
  timeout := time.Duration(cfg.GraphitiTimeout) * time.Second
  storeCtx, cancel := context.WithTimeout(ctx, timeout)
  defer cancel()
  ```

The Graphiti integration captures:
- Agent responses and reasoning for all agent types (pentester, researcher, coder, etc.)
- Tool execution details including function name, arguments, results, and execution status
- Context information including flow, task, and subtask IDs for hierarchical organization
- Temporal relationships between entities, actions, and outcomes

These settings enable:
- Building a comprehensive knowledge base from agent interactions
- Semantic memory across multiple penetration tests
- Advanced querying of relationships between tools, targets, and techniques
- Learning from past successful approaches and strategies

The integration is designed to be non-blocking - if Graphiti operations fail, they are logged but don't interrupt the agent workflow.

## Observability Settings

These settings control the observability and monitoring capabilities, including telemetry and trace collection for system performance and debugging.

### Telemetry

| Option            | Environment Variable | Default Value | Description                                |
| ----------------- | -------------------- | ------------- | ------------------------------------------ |
| TelemetryEndpoint | `OTEL_HOST`          | *(none)*      | Endpoint for OpenTelemetry data collection |

### Langfuse

| Option            | Environment Variable  | Default Value | Description                 |
| ----------------- | --------------------- | ------------- | --------------------------- |
| LangfuseBaseURL   | `LANGFUSE_BASE_URL`   | *(none)*      | Base URL for Langfuse API   |
| LangfuseProjectID | `LANGFUSE_PROJECT_ID` | *(none)*      | Project ID for Langfuse     |
| LangfusePublicKey | `LANGFUSE_PUBLIC_KEY` | *(none)*      | Public key for Langfuse API |
| LangfuseSecretKey | `LANGFUSE_SECRET_KEY` | *(none)*      | Secret key for Langfuse API |

### Usage Details

The observability settings are used in `main.go` and the observability package to initialize monitoring systems:

- **Telemetry Configuration**: Sets up OpenTelemetry for metrics, logs, and traces:
  ```go
  // Check if telemetry is configured
  if cfg.TelemetryEndpoint == "" {
      return nil, ErrNotConfigured
  }

  // Create telemetry client with endpoint
  otelclient, err := obs.NewTelemetryClient(ctx, cfg)
  ```

- **Langfuse Configuration**: Configures Langfuse for LLM operation monitoring:
  ```go
  // Check if Langfuse is configured
  if cfg.LangfuseBaseURL == "" {
      return nil, ErrNotConfigured
  }

  // Configure Langfuse client
  langfuse.WithBaseURL(cfg.LangfuseBaseURL),
  langfuse.WithPublicKey(cfg.LangfusePublicKey),
  langfuse.WithSecretKey(cfg.LangfuseSecretKey),
  langfuse.WithProjectID(cfg.LangfuseProjectID),
  ```

- **Integration in Application**: Used in `main.go` to initialize observability:
  ```go
  lfclient, err := obs.NewLangfuseClient(ctx, cfg)
  if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
      log.Fatalf("Unable to create langfuse client: %v\n", err)
  }

  otelclient, err := obs.NewTelemetryClient(ctx, cfg)
  if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
      log.Fatalf("Unable to create telemetry client: %v\n", err)
  }

  obs.InitObserver(ctx, lfclient, otelclient, []logrus.Level{
      logrus.DebugLevel,
      logrus.InfoLevel,
      logrus.WarnLevel,
      logrus.ErrorLevel,
  })
  ```

These settings enable:
- Comprehensive monitoring of system performance
- LLM-specific metrics collection via Langfuse
- Tracing of requests through the system
- Centralized logging for troubleshooting
- Performance optimization based on collected metrics
