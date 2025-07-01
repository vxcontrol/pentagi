# PentAGI

<div align="center" style="font-size: 1.5em; margin: 20px 0;">
    <strong>P</strong>enetration testing <strong>A</strong>rtificial <strong>G</strong>eneral <strong>I</strong>ntelligence
</div>
<br>
<div align="center">

> 🚀 **Join the Community!** Connect with security researchers, AI enthusiasts, and fellow ethical hackers. Get support, share insights, and stay updated with the latest PentAGI developments.

[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/2xrMh7qX6m)⠀[![Telegram](https://img.shields.io/badge/Telegram-2CA5E0?logo=telegram&logoColor=white)](https://t.me/+Ka9i6CNwe71hMWQy)

</div>

## 📖 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Quick Start](#-quick-start)
- [Advanced Setup](#-advanced-setup)
- [Development](#-development)
- [Testing LLM Agents](#-testing-llm-agents)
- [Embedding Configuration and Testing](#-embedding-configuration-and-testing)
- [Function Testing with ftester](#-function-testing-with-ftester)
- [Building](#%EF%B8%8F-building)
- [Credits](#-credits)
- [License](#-license)

## 🎯 Overview

PentAGI is an innovative tool for automated security testing that leverages cutting-edge artificial intelligence technologies. The project is designed for information security professionals, researchers, and enthusiasts who need a powerful and flexible solution for conducting penetration tests.

You can watch the video **PentAGI overview**:
[![PentAGI Overview Video](https://github.com/user-attachments/assets/0828dc3e-15f1-4a1d-858e-9696a146e478)](https://youtu.be/R70x5Ddzs1o)

## ✨ Features

- 🛡️ Secure & Isolated. All operations are performed in a sandboxed Docker environment with complete isolation.
- 🤖 Fully Autonomous. AI-powered agent that automatically determines and executes penetration testing steps.
- 🔬 Professional Pentesting Tools. Built-in suite of 20+ professional security tools including nmap, metasploit, sqlmap, and more.
- 🧠 Smart Memory System. Long-term storage of research results and successful approaches for future use.
- 🔍 Web Intelligence. Built-in browser via [scraper](https://hub.docker.com/r/vxcontrol/scraper) for gathering latest information from web sources.
- 🔎 External Search Systems. Integration with advanced search APIs including [Tavily](https://tavily.com), [Traversaal](https://traversaal.ai), [Perplexity](https://www.perplexity.ai), [DuckDuckGo](https://duckduckgo.com/) and [Google Custom Search](https://programmablesearchengine.google.com/) for comprehensive information gathering.
- 👥 Team of Specialists. Delegation system with specialized AI agents for research, development, and infrastructure tasks.
- 📊 Comprehensive Monitoring. Detailed logging and integration with Grafana/Prometheus for real-time system observation.
- 📝 Detailed Reporting. Generation of thorough vulnerability reports with exploitation guides.
- 📦 Smart Container Management. Automatic Docker image selection based on specific task requirements.
- 📱 Modern Interface. Clean and intuitive web UI for system management and monitoring.
- 🔌 API Integration. Support for REST and GraphQL APIs for seamless external system integration.
- 💾 Persistent Storage. All commands and outputs are stored in PostgreSQL with [pgvector](https://hub.docker.com/r/vxcontrol/pgvector) extension.
- 🎯 Scalable Architecture. Microservices-based design supporting horizontal scaling.
- 🏠 Self-Hosted Solution. Complete control over your deployment and data.
- 🔑 Flexible Authentication. Support for various LLM providers ([OpenAI](https://platform.openai.com/), [Anthropic](https://www.anthropic.com/), [Deep Infra](https://deepinfra.com/), [OpenRouter](https://openrouter.ai/), [DeepSeek](https://www.deepseek.com/)) and custom configurations.
- ⚡ Quick Deployment. Easy setup through [Docker Compose](https://docs.docker.com/compose/) with comprehensive environment configuration.

## 🏗️ Architecture

### System Context

```mermaid
flowchart TB
    classDef person fill:#08427B,stroke:#073B6F,color:#fff
    classDef system fill:#1168BD,stroke:#0B4884,color:#fff
    classDef external fill:#666666,stroke:#0B4884,color:#fff
    
    pentester["👤 Security Engineer
    (User of the system)"]
    
    pentagi["✨ PentAGI
    (Autonomous penetration testing system)"]
    
    target["🎯 target-system
    (System under test)"]
    llm["🧠 llm-provider
    (OpenAI/Anthropic/Custom)"]
    search["🔍 search-systems
    (Google/DuckDuckGo/Tavily/Traversaal/Perplexity)"]
    langfuse["📊 langfuse-ui
    (LLM Observability Dashboard)"]
    grafana["📈 grafana
    (System Monitoring Dashboard)"]
    
    pentester --> |Uses HTTPS| pentagi
    pentester --> |Monitors AI HTTPS| langfuse
    pentester --> |Monitors System HTTPS| grafana
    pentagi --> |Tests Various protocols| target
    pentagi --> |Queries HTTPS| llm
    pentagi --> |Searches HTTPS| search
    pentagi --> |Reports HTTPS| langfuse
    pentagi --> |Reports HTTPS| grafana
    
    class pentester person
    class pentagi system
    class target,llm,search,langfuse,grafana external
    
    linkStyle default stroke:#ffffff,color:#ffffff
```

<details>
<summary><b>🔄 Container Architecture</b> (click to expand)</summary>

```mermaid
graph TB
    subgraph Core Services
        UI[Frontend UI<br/>React + TypeScript]
        API[Backend API<br/>Go + GraphQL]
        DB[(Vector Store<br/>PostgreSQL + pgvector)]
        MQ[Task Queue<br/>Async Processing]
        Agent[AI Agents<br/>Multi-Agent System]
    end

    subgraph Monitoring
        Grafana[Grafana<br/>Dashboards]
        VictoriaMetrics[VictoriaMetrics<br/>Time-series DB]
        Jaeger[Jaeger<br/>Distributed Tracing]
        Loki[Loki<br/>Log Aggregation]
        OTEL[OpenTelemetry<br/>Data Collection]
    end

    subgraph Analytics
        Langfuse[Langfuse<br/>LLM Analytics]
        ClickHouse[ClickHouse<br/>Analytics DB]
        Redis[Redis<br/>Cache + Rate Limiter]
        MinIO[MinIO<br/>S3 Storage]
    end

    subgraph Security Tools
        Scraper[Web Scraper<br/>Isolated Browser]
        PenTest[Security Tools<br/>20+ Pro Tools<br/>Sandboxed Execution]
    end

    UI --> |HTTP/WS| API
    API --> |SQL| DB
    API --> |Events| MQ
    MQ --> |Tasks| Agent
    Agent --> |Commands| Tools
    Agent --> |Queries| DB
    
    API --> |Telemetry| OTEL
    OTEL --> |Metrics| VictoriaMetrics
    OTEL --> |Traces| Jaeger
    OTEL --> |Logs| Loki
    
    Grafana --> |Query| VictoriaMetrics
    Grafana --> |Query| Jaeger
    Grafana --> |Query| Loki
    
    API --> |Analytics| Langfuse
    Langfuse --> |Store| ClickHouse
    Langfuse --> |Cache| Redis
    Langfuse --> |Files| MinIO

    classDef core fill:#f9f,stroke:#333,stroke-width:2px,color:#000
    classDef monitoring fill:#bbf,stroke:#333,stroke-width:2px,color:#000
    classDef analytics fill:#bfb,stroke:#333,stroke-width:2px,color:#000
    classDef tools fill:#fbb,stroke:#333,stroke-width:2px,color:#000
    
    class UI,API,DB,MQ,Agent core
    class Grafana,VictoriaMetrics,Jaeger,Loki,OTEL monitoring
    class Langfuse,ClickHouse,Redis,MinIO analytics
    class Scraper,PenTest tools
```

</details>

<details>
<summary><b>📊 Entity Relationship</b> (click to expand)</summary>

```mermaid
erDiagram
    Flow ||--o{ Task : contains
    Task ||--o{ SubTask : contains
    SubTask ||--o{ Action : contains
    Action ||--o{ Artifact : produces
    Action ||--o{ Memory : stores
    
    Flow {
        string id PK
        string name "Flow name"
        string description "Flow description"
        string status "active/completed/failed"
        json parameters "Flow parameters"
        timestamp created_at
        timestamp updated_at
    }
    
    Task {
        string id PK
        string flow_id FK
        string name "Task name"
        string description "Task description"
        string status "pending/running/done/failed"
        json result "Task results"
        timestamp created_at
        timestamp updated_at
    }
    
    SubTask {
        string id PK
        string task_id FK
        string name "Subtask name"
        string description "Subtask description"
        string status "queued/running/completed/failed"
        string agent_type "researcher/developer/executor"
        json context "Agent context"
        timestamp created_at
        timestamp updated_at
    }
    
    Action {
        string id PK
        string subtask_id FK
        string type "command/search/analyze/etc"
        string status "success/failure"
        json parameters "Action parameters"
        json result "Action results"
        timestamp created_at
    }

    Artifact {
        string id PK
        string action_id FK
        string type "file/report/log"
        string path "Storage path"
        json metadata "Additional info"
        timestamp created_at
    }

    Memory {
        string id PK
        string action_id FK
        string type "observation/conclusion"
        vector embedding "Vector representation"
        text content "Memory content"
        timestamp created_at
    }
```

</details>

<details>
<summary><b>🤖 Agent Interaction</b> (click to expand)</summary>

```mermaid
sequenceDiagram
    participant O as Orchestrator
    participant R as Researcher
    participant D as Developer
    participant E as Executor
    participant VS as Vector Store
    participant KB as Knowledge Base
    
    Note over O,KB: Flow Initialization
    O->>VS: Query similar tasks
    VS-->>O: Return experiences
    O->>KB: Load relevant knowledge
    KB-->>O: Return context
    
    Note over O,R: Research Phase
    O->>R: Analyze target
    R->>VS: Search similar cases
    VS-->>R: Return patterns
    R->>KB: Query vulnerabilities
    KB-->>R: Return known issues
    R->>VS: Store findings
    R-->>O: Research results
    
    Note over O,D: Planning Phase
    O->>D: Plan attack
    D->>VS: Query exploits
    VS-->>D: Return techniques
    D->>KB: Load tools info
    KB-->>D: Return capabilities
    D-->>O: Attack plan
    
    Note over O,E: Execution Phase
    O->>E: Execute plan
    E->>KB: Load tool guides
    KB-->>E: Return procedures
    E->>VS: Store results
    E-->>O: Execution status
```

</details>

<details>
<summary><b>🧠 Memory System</b> (click to expand)</summary>

```mermaid
graph TB
    subgraph "Long-term Memory"
        VS[(Vector Store<br/>Embeddings DB)]
        KB[Knowledge Base<br/>Domain Expertise]
        Tools[Tools Knowledge<br/>Usage Patterns]
    end
    
    subgraph "Working Memory"
        Context[Current Context<br/>Task State]
        Goals[Active Goals<br/>Objectives]
        State[System State<br/>Resources]
    end
    
    subgraph "Episodic Memory"
        Actions[Past Actions<br/>Commands History]
        Results[Action Results<br/>Outcomes]
        Patterns[Success Patterns<br/>Best Practices]
    end
    
    Context --> |Query| VS
    VS --> |Retrieve| Context
    
    Goals --> |Consult| KB
    KB --> |Guide| Goals
    
    State --> |Record| Actions
    Actions --> |Learn| Patterns
    Patterns --> |Store| VS
    
    Tools --> |Inform| State
    Results --> |Update| Tools
    
    VS --> |Enhance| KB
    KB --> |Index| VS

    classDef ltm fill:#f9f,stroke:#333,stroke-width:2px,color:#000
    classDef wm fill:#bbf,stroke:#333,stroke-width:2px,color:#000
    classDef em fill:#bfb,stroke:#333,stroke-width:2px,color:#000
    
    class VS,KB,Tools ltm
    class Context,Goals,State wm
    class Actions,Results,Patterns em
```

</details>

<details>
<summary><b>🔄 Chain Summarization</b> (click to expand)</summary>

The chain summarization system manages conversation context growth by selectively summarizing older messages. This is critical for preventing token limits from being exceeded while maintaining conversation coherence.

```mermaid
flowchart TD
    A[Input Chain] --> B{Needs Summarization?}
    B -->|No| C[Return Original Chain]
    B -->|Yes| D[Convert to ChainAST]
    D --> E[Apply Section Summarization]
    E --> F[Process Oversized Pairs]
    F --> G[Manage Last Section Size]
    G --> H[Apply QA Summarization]
    H --> I[Rebuild Chain with Summaries]
    I --> J{Is New Chain Smaller?}
    J -->|Yes| K[Return Optimized Chain]
    J -->|No| C
    
    classDef process fill:#bbf,stroke:#333,stroke-width:2px,color:#000
    classDef decision fill:#bfb,stroke:#333,stroke-width:2px,color:#000
    classDef output fill:#fbb,stroke:#333,stroke-width:2px,color:#000
    
    class A,D,E,F,G,H,I process
    class B,J decision
    class C,K output
```

The algorithm operates on a structured representation of conversation chains (ChainAST) that preserves message types including tool calls and their responses. All summarization operations maintain critical conversation flow while reducing context size.

### Global Summarizer Configuration Options

| Parameter | Environment Variable | Default | Description |
|-----------|----------------------|---------|-------------|
| Preserve Last | `SUMMARIZER_PRESERVE_LAST` | `true` | Whether to keep all messages in the last section intact |
| Use QA Pairs | `SUMMARIZER_USE_QA` | `true` | Whether to use QA pair summarization strategy |
| Summarize Human in QA | `SUMMARIZER_SUM_MSG_HUMAN_IN_QA` | `false` | Whether to summarize human messages in QA pairs |
| Last Section Size | `SUMMARIZER_LAST_SEC_BYTES` | `51200` | Maximum byte size for last section (50KB) |
| Max Body Pair Size | `SUMMARIZER_MAX_BP_BYTES` | `16384` | Maximum byte size for a single body pair (16KB) |
| Max QA Sections | `SUMMARIZER_MAX_QA_SECTIONS` | `10` | Maximum QA pair sections to preserve |
| Max QA Size | `SUMMARIZER_MAX_QA_BYTES` | `65536` | Maximum byte size for QA pair sections (64KB) |
| Keep QA Sections | `SUMMARIZER_KEEP_QA_SECTIONS` | `1` | Number of recent QA sections to keep without summarization |

### Assistant Summarizer Configuration Options

Assistant instances can use customized summarization settings to fine-tune context management behavior:

| Parameter | Environment Variable | Default | Description |
|-----------|----------------------|---------|-------------|
| Preserve Last | `ASSISTANT_SUMMARIZER_PRESERVE_LAST` | `true` | Whether to preserve all messages in the assistant's last section |
| Last Section Size | `ASSISTANT_SUMMARIZER_LAST_SEC_BYTES` | `76800` | Maximum byte size for assistant's last section (75KB) |
| Max Body Pair Size | `ASSISTANT_SUMMARIZER_MAX_BP_BYTES` | `16384` | Maximum byte size for a single body pair in assistant context (16KB) |
| Max QA Sections | `ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS` | `7` | Maximum QA sections to preserve in assistant context |
| Max QA Size | `ASSISTANT_SUMMARIZER_MAX_QA_BYTES` | `76800` | Maximum byte size for assistant's QA sections (75KB) |
| Keep QA Sections | `ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS` | `3` | Number of recent QA sections to preserve without summarization |

The assistant summarizer configuration provides more memory for context retention compared to the global settings, preserving more recent conversation history while still ensuring efficient token usage.

### Summarizer Environment Configuration

```bash
# Default values for global summarizer logic
SUMMARIZER_PRESERVE_LAST=true
SUMMARIZER_USE_QA=true
SUMMARIZER_SUM_MSG_HUMAN_IN_QA=false
SUMMARIZER_LAST_SEC_BYTES=51200
SUMMARIZER_MAX_BP_BYTES=16384
SUMMARIZER_MAX_QA_SECTIONS=10
SUMMARIZER_MAX_QA_BYTES=65536
SUMMARIZER_KEEP_QA_SECTIONS=1

# Default values for assistant summarizer logic
ASSISTANT_SUMMARIZER_PRESERVE_LAST=true
ASSISTANT_SUMMARIZER_LAST_SEC_BYTES=76800
ASSISTANT_SUMMARIZER_MAX_BP_BYTES=16384
ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS=7
ASSISTANT_SUMMARIZER_MAX_QA_BYTES=76800
ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS=3
```

</details>

The architecture of PentAGI is designed to be modular, scalable, and secure. Here are the key components:

1. **Core Services**
   - Frontend UI: React-based web interface with TypeScript for type safety
   - Backend API: Go-based REST and GraphQL APIs for flexible integration
   - Vector Store: PostgreSQL with pgvector for semantic search and memory storage
   - Task Queue: Async task processing system for reliable operation
   - AI Agent: Multi-agent system with specialized roles for efficient testing

2. **Monitoring Stack**
   - OpenTelemetry: Unified observability data collection and correlation
   - Grafana: Real-time visualization and alerting dashboards
   - VictoriaMetrics: High-performance time-series metrics storage
   - Jaeger: End-to-end distributed tracing for debugging
   - Loki: Scalable log aggregation and analysis

3. **Analytics Platform**
   - Langfuse: Advanced LLM observability and performance analytics
   - ClickHouse: Column-oriented analytics data warehouse
   - Redis: High-speed caching and rate limiting
   - MinIO: S3-compatible object storage for artifacts

4. **Security Tools**
   - Web Scraper: Isolated browser environment for safe web interaction
   - Pentesting Tools: Comprehensive suite of 20+ professional security tools
   - Sandboxed Execution: All operations run in isolated containers

5. **Memory Systems**
   - Long-term Memory: Persistent storage of knowledge and experiences
   - Working Memory: Active context and goals for current operations
   - Episodic Memory: Historical actions and success patterns
   - Knowledge Base: Structured domain expertise and tool capabilities
   - Context Management: Intelligently manages growing LLM context windows using chain summarization

The system uses Docker containers for isolation and easy deployment, with separate networks for core services, monitoring, and analytics to ensure proper security boundaries. Each component is designed to scale horizontally and can be configured for high availability in production environments.

## 🚀 Quick Start

### System Requirements

- Docker and Docker Compose
- Minimum 4GB RAM
- 10GB free disk space
- Internet access for downloading images and updates

### Basic Installation

1. Create a working directory or clone the repository:

```bash
mkdir pentagi && cd pentagi
```

2. Copy `.env.example` to `.env` or download it:

```bash
curl -o .env https://raw.githubusercontent.com/vxcontrol/pentagi/master/.env.example
```

3. Fill in the required API keys in `.env` file.

```bash
# Required: At least one of these LLM providers
OPEN_AI_KEY=your_openai_key
ANTHROPIC_API_KEY=your_anthropic_key

# Optional: Additional search capabilities
DUCKDUCKGO_ENABLED=true
GOOGLE_API_KEY=your_google_key
GOOGLE_CX_KEY=your_google_cx
TAVILY_API_KEY=your_tavily_key
TRAVERSAAL_API_KEY=your_traversaal_key
PERPLEXITY_API_KEY=your_perplexity_key
PERPLEXITY_MODEL=sonar-pro
PERPLEXITY_CONTEXT_SIZE=medium

# Assistant configuration
ASSISTANT_USE_AGENTS=false         # Default value for agent usage when creating new assistants
```

4. Change all security related environment variables in `.env` file to improve security.

<details>
    <summary>Security related environment variables</summary>

### Main Security Settings
- `COOKIE_SIGNING_SALT` - Salt for cookie signing, change to random value
- `PUBLIC_URL` - Public URL of your server (eg. `https://pentagi.example.com`)
- `SERVER_SSL_CRT` and `SERVER_SSL_KEY` - Custom paths to your existing SSL certificate and key for HTTPS (these paths should be used in the docker-compose.yml file to mount as volumes)

### Scraper Access
- `SCRAPER_PUBLIC_URL` - Public URL for scraper if you want to use different scraper server for public URLs
- `SCRAPER_PRIVATE_URL` - Private URL for scraper (local scraper server in docker-compose.yml file to access it to local URLs)

### Access Credentials
- `PENTAGI_POSTGRES_USER` and `PENTAGI_POSTGRES_PASSWORD` - PostgreSQL credentials

</details>

5. Remove all inline comments from `.env` file if you want to use it in VSCode or other IDEs as a envFile option:

```bash
perl -i -pe 's/\s+#.*$//' .env
```

6. Run the PentAGI stack:

```bash
curl -O https://raw.githubusercontent.com/vxcontrol/pentagi/master/docker-compose.yml
docker compose up -d
```

Visit [localhost:8443](https://localhost:8443) to access PentAGI Web UI (default is `admin@pentagi.com` / `admin`)

> [!NOTE]
> If you caught an error about `pentagi-network` or `observability-network` or `langfuse-network` you need to run `docker-compose.yml` firstly to create these networks and after that run `docker-compose-langfuse.yml` and `docker-compose-observability.yml` to use Langfuse and Observability services.
> 
> You have to set at least one Language Model provider (OpenAI or Anthropic) to use PentAGI. Additional API keys for search engines are optional but recommended for better results.
> 
> `LLM_SERVER_*` environment variables are experimental feature and will be changed in the future. Right now you can use them to specify custom LLM server URL and one model for all agent types.
> 
> `PROXY_URL` is a global proxy URL for all LLM providers and external search systems. You can use it for isolation from external networks.
>
> The `docker-compose.yml` file runs the PentAGI service as root user because it needs access to docker.sock for container management. If you're using TCP/IP network connection to Docker instead of socket file, you can remove root privileges and use the default `pentagi` user for better security.

### Assistant Configuration

PentAGI allows you to configure default behavior for assistants:

| Variable | Default | Description |
|----------|---------|-------------|
| `ASSISTANT_USE_AGENTS` | `false` | Controls the default value for agent usage when creating new assistants |

The `ASSISTANT_USE_AGENTS` setting affects the initial state of the "Use Agents" toggle when creating a new assistant in the UI:
- `false` (default): New assistants are created with agent delegation disabled by default
- `true`: New assistants are created with agent delegation enabled by default

Note that users can always override this setting by toggling the "Use Agents" button in the UI when creating or editing an assistant. This environment variable only controls the initial default state.

### Custom LLM Provider Configuration

When using custom LLM providers with the `LLM_SERVER_*` variables, you can fine-tune the reasoning format used in requests:

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_SERVER_URL` | | Base URL for the custom LLM API endpoint |
| `LLM_SERVER_KEY` | | API key for the custom LLM provider |
| `LLM_SERVER_MODEL` | | Default model to use (can be overridden in provider config) |
| `LLM_SERVER_CONFIG_PATH` | | Path to the YAML configuration file for agent-specific models |
| `LLM_SERVER_LEGACY_REASONING` | `false` | Controls reasoning format in API requests |

The `LLM_SERVER_LEGACY_REASONING` setting affects how reasoning parameters are sent to the LLM:
- `false` (default): Uses modern format where reasoning is sent as a structured object with `max_tokens` parameter
- `true`: Uses legacy format with string-based `reasoning_effort` parameter

This setting is important when working with different LLM providers as they may expect different reasoning formats in their API requests. If you encounter reasoning-related errors with custom providers, try changing this setting.

For advanced configuration options and detailed setup instructions, please visit our [documentation](https://docs.pentagi.com).

## 🔧 Advanced Setup

### Langfuse Integration

Langfuse provides advanced capabilities for monitoring and analyzing AI agent operations.

1. Configure Langfuse environment variables in existing `.env` file.

<details>
    <summary>Langfuse valuable environment variables</summary>

### Database Credentials
- `LANGFUSE_POSTGRES_USER` and `LANGFUSE_POSTGRES_PASSWORD` - Langfuse PostgreSQL credentials
- `LANGFUSE_CLICKHOUSE_USER` and `LANGFUSE_CLICKHOUSE_PASSWORD` - ClickHouse credentials
- `LANGFUSE_REDIS_AUTH` - Redis password

### Encryption and Security Keys
- `LANGFUSE_SALT` - Salt for hashing in Langfuse Web UI
- `LANGFUSE_ENCRYPTION_KEY` - Encryption key (32 bytes in hex)
- `LANGFUSE_NEXTAUTH_SECRET` - Secret key for NextAuth

### Admin Credentials
- `LANGFUSE_INIT_USER_EMAIL` - Admin email
- `LANGFUSE_INIT_USER_PASSWORD` - Admin password
- `LANGFUSE_INIT_USER_NAME` - Admin username

### API Keys and Tokens
- `LANGFUSE_INIT_PROJECT_PUBLIC_KEY` - Project public key (used from PentAGI side too)
- `LANGFUSE_INIT_PROJECT_SECRET_KEY` - Project secret key (used from PentAGI side too)

### S3 Storage
- `LANGFUSE_S3_ACCESS_KEY_ID` - S3 access key ID
- `LANGFUSE_S3_SECRET_ACCESS_KEY` - S3 secret access key

</details>

2. Enable integration with Langfuse for PentAGI service in `.env` file.

```bash
LANGFUSE_BASE_URL=http://langfuse-web:3000
LANGFUSE_PROJECT_ID= # default: value from ${LANGFUSE_INIT_PROJECT_ID}
LANGFUSE_PUBLIC_KEY= # default: value from ${LANGFUSE_INIT_PROJECT_PUBLIC_KEY}
LANGFUSE_SECRET_KEY= # default: value from ${LANGFUSE_INIT_PROJECT_SECRET_KEY}
```

3. Run the Langfuse stack:

```bash
curl -O https://raw.githubusercontent.com/vxcontrol/pentagi/master/docker-compose-langfuse.yml
docker compose -f docker-compose.yml -f docker-compose-langfuse.yml up -d
```

Visit [localhost:4000](http://localhost:4000) to access Langfuse Web UI with credentials from `.env` file:

- `LANGFUSE_INIT_USER_EMAIL` - Admin email
- `LANGFUSE_INIT_USER_PASSWORD` - Admin password

### Monitoring and Observability

For detailed system operation tracking, integration with monitoring tools is available.

1. Enable integration with OpenTelemetry and all observability services for PentAGI in `.env` file.

```bash
OTEL_HOST=otelcol:8148
```

2. Run the observability stack:

```bash
curl -O https://raw.githubusercontent.com/vxcontrol/pentagi/master/docker-compose-observability.yml
docker compose -f docker-compose.yml -f docker-compose-observability.yml up -d
```

Visit [localhost:3000](http://localhost:3000) to access Grafana Web UI.

> [!NOTE]
> If you want to use Observability stack with Langfuse, you need to enable integration in `.env` file to set `LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT` to `http://otelcol:4318`.
> 
> And you need to run both stacks `docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-observability.yml up -d` to have all services running.
> 
> Also you can register aliases for these commands in your shell to run it faster:
>
> ```bash
> alias pentagi="docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-observability.yml"
> alias pentagi-up="docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-observability.yml up -d"
> alias pentagi-down="docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-observability.yml down"
> ```

### GitHub and Google OAuth Integration

OAuth integration with GitHub and Google allows users to authenticate using their existing accounts on these platforms. This provides several benefits:

- Simplified login process without need to create separate credentials
- Enhanced security through trusted identity providers
- Access to user profile information from GitHub/Google accounts
- Seamless integration with existing development workflows

For using GitHub OAuth you need to create a new OAuth application in your GitHub account and set the `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` in `.env` file.

For using Google OAuth you need to create a new OAuth application in your Google account and set the `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in `.env` file.

### Docker Image Configuration

PentAGI allows you to configure Docker image selection for executing various tasks. The system automatically chooses the most appropriate image based on the task type, but you can constrain this selection by specifying your preferred images:

| Variable | Default | Description |
|----------|---------|-------------|
| `DOCKER_DEFAULT_IMAGE` | `debian:latest` | Default Docker image for general tasks and ambiguous cases |
| `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` | `vxcontrol/kali-linux` | Default Docker image for security/penetration testing tasks |

When these environment variables are set, AI agents will be limited to the image choices you specify. This is particularly useful for:

- **Security Enforcement**: Restricting usage to only verified and trusted images
- **Environment Standardization**: Using corporate or customized images across all operations
- **Performance Optimization**: Utilizing pre-built images with necessary tools already installed

Configuration examples:

```bash
# Using a custom image for general tasks
DOCKER_DEFAULT_IMAGE=mycompany/custom-debian:latest

# Using a specialized image for penetration testing
DOCKER_DEFAULT_IMAGE_FOR_PENTEST=mycompany/pentest-tools:v2.0
```

> [!NOTE]
> If a user explicitly specifies a particular Docker image in their task, the system will try to use that exact image, ignoring these settings. These variables only affect the system's automatic image selection process.

## 💻 Development

### Development Requirements

- golang
- nodejs
- docker
- postgres
- commitlint

### Environment Setup

#### Backend Setup

Run once `cd backend && go mod download` to install needed packages.

For generating swagger files have to run 

```bash
swag init -g ../../pkg/server/router.go -o pkg/server/docs/ --parseDependency --parseInternal --parseDepth 2 -d cmd/pentagi
```

before installing `swag` package via 

```bash
go install github.com/swaggo/swag/cmd/swag@v1.8.7
```

For generating graphql resolver files have to run 

```bash
go run github.com/99designs/gqlgen --config ./gqlgen/gqlgen.yml
```

after that you can see the generated files in `pkg/graph` folder.

For generating ORM methods (database package) from sqlc configuration

```bash
docker run --rm -v $(pwd):/src -w /src --network pentagi-network -e DATABASE_URL="{URL}" sqlc/sqlc generate -f sqlc/sqlc.yml
```

For generating Langfuse SDK from OpenAPI specification

```bash
fern generate --local
```

and to install fern-cli

```bash
npm install -g fern-api
```

#### Testing

For running tests `cd backend && go test -v ./...`

#### Frontend Setup

Run once `cd frontend && npm install` to install needed packages.

For generating graphql files have to run `npm run graphql:generate` which using `graphql-codegen.ts` file.

Be sure that you have `graphql-codegen` installed globally:

```bash
npm install -g graphql-codegen
```

After that you can run:
* `npm run prettier` to check if your code is formatted correctly
* `npm run prettier:fix` to fix it
* `npm run lint` to check if your code is linted correctly
* `npm run lint:fix` to fix it

For generating SSL certificates you need to run `npm run ssl:generate` which using `generate-ssl.ts` file or it will be generated automatically when you run `npm run dev`.

#### Backend Configuration

Edit the configuration for `backend` in `.vscode/launch.json` file:
- `DATABASE_URL` - PostgreSQL database URL (eg. `postgres://postgres:postgres@localhost:5432/pentagidb?sslmode=disable`)
- `DOCKER_HOST` - Docker SDK API (eg. for macOS `DOCKER_HOST=unix:///Users/<my-user>/Library/Containers/com.docker.docker/Data/docker.raw.sock`) [more info](https://stackoverflow.com/a/62757128/5922857)

Optional:
- `SERVER_PORT` - Port to run the server (default: `8443`)
- `SERVER_USE_SSL` - Enable SSL for the server (default: `false`)

#### Frontend Configuration

Edit the configuration for `frontend` in `.vscode/launch.json` file:
- `VITE_API_URL` - Backend API URL. *Omit* the URL scheme (e.g., `localhost:8080` *NOT* `http://localhost:8080`)
- `VITE_USE_HTTPS` - Enable SSL for the server (default: `false`)
- `VITE_PORT` - Port to run the server (default: `8000`)
- `VITE_HOST` - Host to run the server (default: `0.0.0.0`)

### Running the Application

#### Backend

Run the command(s) in `backend` folder:
- Use `.env` file to set environment variables like a `source .env`
- Run `go run cmd/pentagi/main.go` to start the server

> [!NOTE]
> The first run can take a while as dependencies and docker images need to be downloaded to setup the backend environment.

#### Frontend

Run the command(s) in `frontend` folder:
- Run `npm install` to install the dependencies
- Run `npm run dev` to run the web app
- Run `npm run build` to build the web app

Open your browser and visit the web app URL.

## 🧪 Testing LLM Agents

PentAGI includes a powerful utility called `ctester` for testing and validating LLM agent capabilities. This tool helps ensure your LLM provider configurations work correctly with different agent types, allowing you to optimize model selection for each specific agent role.

The utility features parallel testing of multiple agents, detailed reporting, and flexible configuration options.

### Key Features

- **Parallel Testing**: Tests multiple agents simultaneously for faster results
- **Comprehensive Test Suite**: Evaluates basic completion, JSON responses, function calling, and more
- **Detailed Reporting**: Generates markdown reports with success rates and performance metrics
- **Flexible Configuration**: Test specific agents or test groups as needed

### Usage Scenarios

#### For Developers (with local Go environment)

If you've cloned the repository and have Go installed:

```bash
# Default configuration with .env file
cd backend
go run cmd/ctester/*.go -verbose

# Custom provider configuration
go run cmd/ctester/*.go -config ../examples/configs/openrouter.provider.yml -verbose

# Generate a report file
go run cmd/ctester/*.go -config ../examples/configs/deepinfra.provider.yml -report ../test-report.md

# Test specific agent types only 
go run cmd/ctester/*.go -agents simple,simple_json,agent -verbose

# Test specific test groups only
go run cmd/ctester/*.go -tests "Simple Completion,System User Prompts" -verbose
```

#### For Users (using Docker image)

If you prefer to use the pre-built Docker image without setting up a development environment:

```bash
# Using Docker to test with default environment
docker run --rm -v $(pwd)/.env:/opt/pentagi/.env vxcontrol/pentagi /opt/pentagi/bin/ctester -verbose

# Test with your custom provider configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  -v $(pwd)/my-config.yml:/opt/pentagi/config.yml \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/config.yml -verbose

# Generate a detailed report
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  -v $(pwd):/opt/pentagi/output \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -report /opt/pentagi/output/report.md
```

#### Using Pre-configured Providers

The Docker image comes with pre-configured provider files for OpenRouter or DeepInfra or DeepSeek:

```bash
# Test with OpenRouter configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/openrouter.provider.yml

# Test with DeepInfra configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/deepinfra.provider.yml

# Test with DeepSeek configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/deepseek.provider.yml

# Test with Custom OpenAI configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/custom-openai.provider.yml
```

To use these configurations, your `.env` file only needs to contain:

```
LLM_SERVER_URL=https://openrouter.ai/api/v1      # or https://api.deepinfra.com/v1/openai or https://api.deepseek.com or https://api.openai.com/v1
LLM_SERVER_KEY=your_api_key
LLM_SERVER_MODEL=                                # Leave empty, as models are specified in the config
LLM_SERVER_CONFIG_PATH=/opt/pentagi/conf/openrouter.provider.yml  # or deepinfra.provider.yml or deepseek.provider.yml or custom-openai.provider.yml
LLM_SERVER_LEGACY_REASONING=false                # Controls reasoning format, for OpenAI must be true (default: false)
```

#### Using OpenAI with Unverified Organizations

For OpenAI accounts with unverified organizations that don't have access to the latest reasoning models (o1, o3, o4-mini), you need to use a custom configuration.

To use OpenAI with unverified organization accounts, configure your `.env` file as follows:

```bash
LLM_SERVER_URL=https://api.openai.com/v1
LLM_SERVER_KEY=your_openai_api_key
LLM_SERVER_MODEL=                                # Leave empty, models are specified in config
LLM_SERVER_CONFIG_PATH=/opt/pentagi/conf/custom-openai.provider.yml
LLM_SERVER_LEGACY_REASONING=true                 # Required for OpenAI reasoning format
```

This configuration uses the pre-built `custom-openai.provider.yml` file that maps all agent types to models available for unverified organizations, using `o3-mini` instead of models like `o1`, `o3`, and `o4-mini`.

You can test this configuration using:

```bash
# Test with custom OpenAI configuration for unverified accounts
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/custom-openai.provider.yml
```

> [!NOTE]
> The `LLM_SERVER_LEGACY_REASONING=true` setting is crucial for OpenAI compatibility as it ensures reasoning parameters are sent in the format expected by OpenAI's API.

#### Running Tests in a Production Environment

If you already have a running PentAGI container and want to test the current configuration:

```bash
# Run ctester in an existing container using current environment variables
docker exec -it pentagi /opt/pentagi/bin/ctester -verbose

# Generate a report file inside the container
docker exec -it pentagi /opt/pentagi/bin/ctester -report /opt/pentagi/data/agent-test-report.md

# Access the report from the host
docker cp pentagi:/opt/pentagi/data/agent-test-report.md ./
```

### Command-line Options

The utility accepts several options:

- `-env <path>` - Path to environment file (default: `.env`)
- `-config <path>` - Path to custom provider config (default: from `LLM_SERVER_CONFIG_PATH` env variable)
- `-report <path>` - Path to write the report file (optional)
- `-agents <list>` - Comma-separated list of agent types to test (default: `all`)
- `-tests <list>` - Comma-separated list of test groups to run (default: `all`)
- `-verbose` - Enable verbose output with detailed test results for each agent

### Example Provider Configuration

Provider configuration defines which models to use for different agent types:

```yaml
simple:
  model: "provider/model-name"
  temperature: 0.7
  top_p: 0.95
  n: 1
  max_tokens: 4000

simple_json:
  model: "provider/model-name"
  temperature: 0.7
  top_p: 1.0
  n: 1
  max_tokens: 4000
  json: true

# ... other agent types ...
```

### Optimization Workflow

1. **Create a baseline**: Run tests with default configuration
2. **Experiment**: Try different models for each agent type
3. **Compare results**: Look for the best success rate and performance
4. **Deploy optimal configuration**: Use in production with your optimized setup

This tool helps ensure your AI agents are using the most effective models for their specific tasks, improving reliability while optimizing costs.

## 🧮 Embedding Configuration and Testing

PentAGI uses vector embeddings for semantic search, knowledge storage, and memory management. The system supports multiple embedding providers that can be configured according to your needs and preferences.

### Supported Embedding Providers

PentAGI supports the following embedding providers:

- **OpenAI** (default): Uses OpenAI's text embedding models
- **Ollama**: Local embedding model through Ollama
- **Mistral**: Mistral AI's embedding models
- **Jina**: Jina AI's embedding service
- **HuggingFace**: Models from HuggingFace
- **GoogleAI**: Google's embedding models
- **VoyageAI**: VoyageAI's embedding models

<details>
<summary><b>Embedding Provider Configuration</b> (click to expand)</summary>

### Environment Variables

To configure the embedding provider, set the following environment variables in your `.env` file:

```bash
# Primary embedding configuration
EMBEDDING_PROVIDER=openai       # Provider type (openai, ollama, mistral, jina, huggingface, googleai, voyageai)
EMBEDDING_MODEL=text-embedding-3-small  # Model name to use
EMBEDDING_URL=                  # Optional custom API endpoint
EMBEDDING_KEY=                  # API key for the provider (if required)
EMBEDDING_BATCH_SIZE=100        # Number of documents to process in a batch
EMBEDDING_STRIP_NEW_LINES=true  # Whether to remove new lines from text before embedding

# Advanced settings
PROXY_URL=                      # Optional proxy for all API calls
```

### Provider-Specific Limitations

Each provider has specific limitations and supported features:

- **OpenAI**: Supports all configuration options
- **Ollama**: Does not support `EMBEDDING_KEY` as it uses local models
- **Mistral**: Does not support `EMBEDDING_MODEL` or custom HTTP client
- **Jina**: Does not support custom HTTP client
- **HuggingFace**: Requires `EMBEDDING_KEY` and supports all other options
- **GoogleAI**: Does not support `EMBEDDING_URL`, requires `EMBEDDING_KEY`
- **VoyageAI**: Supports all configuration options

If `EMBEDDING_URL` and `EMBEDDING_KEY` are not specified, the system will attempt to use the corresponding LLM provider settings (e.g., `OPEN_AI_KEY` when `EMBEDDING_PROVIDER=openai`).

### Why Consistent Embedding Providers Matter

It's crucial to use the same embedding provider consistently because:

1. **Vector Compatibility**: Different providers produce vectors with different dimensions and mathematical properties
2. **Semantic Consistency**: Changing providers can break semantic similarity between previously embedded documents
3. **Memory Corruption**: Mixed embeddings can lead to poor search results and broken knowledge base functionality

If you change your embedding provider, you should flush and reindex your entire knowledge base (see `etester` utility below).

</details>

### Embedding Tester Utility (etester)

PentAGI includes a specialized `etester` utility for testing, managing, and debugging embedding functionality. This tool is essential for diagnosing and resolving issues related to vector embeddings and knowledge storage.

<details>
<summary><b>Etester Commands</b> (click to expand)</summary>

```bash
# Test embedding provider and database connection
cd backend
go run cmd/etester/main.go test -verbose

# Show statistics about the embedding database
go run cmd/etester/main.go info

# Delete all documents from the embedding database (use with caution!)
go run cmd/etester/main.go flush

# Recalculate embeddings for all documents (after changing provider)
go run cmd/etester/main.go reindex

# Search for documents in the embedding database
go run cmd/etester/main.go search -query "How to install PostgreSQL" -limit 5
```

### Using Docker

If you're running PentAGI in Docker, you can use etester from within the container:

```bash
# Test embedding provider
docker exec -it pentagi /opt/pentagi/bin/etester test

# Show detailed database information
docker exec -it pentagi /opt/pentagi/bin/etester info -verbose
```

### Advanced Search Options

The `search` command supports various filters to narrow down results:

```bash
# Filter by document type
docker exec -it pentagi /opt/pentagi/bin/etester search -query "Security vulnerability" -doc_type guide -threshold 0.8

# Filter by flow ID
docker exec -it pentagi /opt/pentagi/bin/etester search -query "Code examples" -doc_type code -flow_id 42

# All available search options
docker exec -it pentagi /opt/pentagi/bin/etester search -help
```

Available search parameters:
- `-query STRING`: Search query text (required)
- `-doc_type STRING`: Filter by document type (answer, memory, guide, code)
- `-flow_id NUMBER`: Filter by flow ID (positive number)
- `-answer_type STRING`: Filter by answer type (guide, vulnerability, code, tool, other) 
- `-guide_type STRING`: Filter by guide type (install, configure, use, pentest, development, other)
- `-limit NUMBER`: Maximum number of results (default: 3)
- `-threshold NUMBER`: Similarity threshold (0.0-1.0, default: 0.7)

### Common Troubleshooting Scenarios

1. **After changing embedding provider**: Always run `flush` or `reindex` to ensure consistency
2. **Poor search results**: Try adjusting the similarity threshold or check if embeddings are correctly generated
3. **Database connection issues**: Verify PostgreSQL is running with pgvector extension installed
4. **Missing API keys**: Check environment variables for your chosen embedding provider

</details>

## 🔍 Function Testing with ftester

PentAGI includes a versatile utility called `ftester` for debugging, testing, and developing specific functions and AI agent behaviors. While `ctester` focuses on testing LLM model capabilities, `ftester` allows you to directly invoke individual system functions and AI agent components with precise control over execution context.

### Key Features

- **Direct Function Access**: Test individual functions without running the entire system
- **Mock Mode**: Test functions without a live PentAGI deployment using built-in mocks
- **Interactive Input**: Fill function arguments interactively for exploratory testing
- **Detailed Output**: Color-coded terminal output with formatted responses and errors
- **Context-Aware Testing**: Debug AI agents within the context of specific flows, tasks, and subtasks
- **Observability Integration**: All function calls are logged to Langfuse and Observability stack

### Usage Modes

#### Command Line Arguments

Run ftester with specific function and arguments directly from the command line:

```bash
# Basic usage with mock mode
cd backend
go run cmd/ftester/main.go [function_name] -[arg1] [value1] -[arg2] [value2]

# Example: Test terminal command in mock mode
go run cmd/ftester/main.go terminal -command "ls -la" -message "List files"

# Using a real flow context
go run cmd/ftester/main.go -flow 123 terminal -command "whoami" -message "Check user"

# Testing AI agent in specific task/subtask context
go run cmd/ftester/main.go -flow 123 -task 456 -subtask 789 pentester -message "Find vulnerabilities"
```

#### Interactive Mode

Run ftester without arguments for a guided interactive experience:

```bash
# Start interactive mode
go run cmd/ftester/main.go [function_name]

# For example, to interactively fill browser tool arguments
go run cmd/ftester/main.go browser
```

<details>
<summary><b>Available Functions</b> (click to expand)</summary>

### Environment Functions
- **terminal**: Execute commands in a container and return the output
- **file**: Perform file operations (read, write, list) in a container

### Search Functions
- **browser**: Access websites and capture screenshots
- **google**: Search the web using Google Custom Search
- **duckduckgo**: Search the web using DuckDuckGo
- **tavily**: Search using Tavily AI search engine
- **traversaal**: Search using Traversaal AI search engine
- **perplexity**: Search using Perplexity AI

### Vector Database Functions
- **search_in_memory**: Search for information in vector database
- **search_guide**: Find guidance documents in vector database
- **search_answer**: Find answers to questions in vector database
- **search_code**: Find code examples in vector database

### AI Agent Functions
- **advice**: Get expert advice from an AI agent
- **coder**: Request code generation or modification
- **maintenance**: Run system maintenance tasks
- **memorist**: Store and organize information in vector database
- **pentester**: Perform security tests and vulnerability analysis
- **search**: Complex search across multiple sources

### Utility Functions
- **describe**: Show information about flows, tasks, and subtasks

</details>

<details>
<summary><b>Debugging Flow Context</b> (click to expand)</summary>

The `describe` function provides detailed information about tasks and subtasks within a flow. This is particularly useful for diagnosing issues when PentAGI encounters problems or gets stuck.

```bash
# List all flows in the system
go run cmd/ftester/main.go describe

# Show all tasks and subtasks for a specific flow
go run cmd/ftester/main.go -flow 123 describe

# Show detailed information for a specific task
go run cmd/ftester/main.go -flow 123 -task 456 describe

# Show detailed information for a specific subtask
go run cmd/ftester/main.go -flow 123 -task 456 -subtask 789 describe

# Show verbose output with full descriptions and results
go run cmd/ftester/main.go -flow 123 describe -verbose
```

This function allows you to identify the exact point where a flow might be stuck and resume processing by directly invoking the appropriate agent function.

</details>

<details>
<summary><b>Function Help and Discovery</b> (click to expand)</summary>

Each function has a help mode that shows available parameters:

```bash
# Get help for a specific function
go run cmd/ftester/main.go [function_name] -help

# Examples:
go run cmd/ftester/main.go terminal -help
go run cmd/ftester/main.go browser -help
go run cmd/ftester/main.go describe -help
```

You can also run ftester without arguments to see a list of all available functions:

```bash
go run cmd/ftester/main.go
```

</details>

<details>
<summary><b>Output Format</b> (click to expand)</summary>

The `ftester` utility uses color-coded output to make interpretation easier:

- **Blue headers**: Section titles and key names
- **Cyan [INFO]**: General information messages
- **Green [SUCCESS]**: Successful operations
- **Red [ERROR]**: Error messages
- **Yellow [WARNING]**: Warning messages
- **Yellow [MOCK]**: Indicates mock mode operation
- **Magenta values**: Function arguments and results

JSON and Markdown responses are automatically formatted for readability.

</details>

<details>
<summary><b>Advanced Usage Scenarios</b> (click to expand)</summary>

### Debugging Stuck AI Flows

When PentAGI gets stuck in a flow:

1. Pause the flow through the UI
2. Use `describe` to identify the current task and subtask
3. Directly invoke the agent function with the same task/subtask IDs
4. Examine the detailed output to identify the issue
5. Resume the flow or manually intervene as needed

### Testing Environment Variables

Verify that API keys and external services are configured correctly:

```bash
# Test Google search API configuration
go run cmd/ftester/main.go google -query "pentesting tools"

# Test browser access to external websites
go run cmd/ftester/main.go browser -url "https://example.com"
```

### Developing New AI Agent Behaviors

When developing new prompt templates or agent behaviors:

1. Create a test flow in the UI
2. Use ftester to directly invoke the agent with different prompts
3. Observe responses and adjust prompts accordingly
4. Check Langfuse for detailed traces of all function calls

### Verifying Docker Container Setup

Ensure containers are properly configured:

```bash
go run cmd/ftester/main.go -flow 123 terminal -command "env | grep -i proxy" -message "Check proxy settings"
```

</details>

<details>
<summary><b>Docker Container Usage</b> (click to expand)</summary>

If you have PentAGI running in Docker, you can use ftester from within the container:

```bash
# Run ftester inside the running PentAGI container
docker exec -it pentagi /opt/pentagi/bin/ftester [arguments]

# Examples:
docker exec -it pentagi /opt/pentagi/bin/ftester -flow 123 describe
docker exec -it pentagi /opt/pentagi/bin/ftester -flow 123 terminal -command "ps aux" -message "List processes"
```

This is particularly useful for production deployments where you don't have a local development environment.

</details>

<details>
<summary><b>Integration with Observability Tools</b> (click to expand)</summary>

All function calls made through ftester are logged to:

1. **Langfuse**: Captures the entire AI agent interaction chain, including prompts, responses, and function calls
2. **OpenTelemetry**: Records metrics, traces, and logs for system performance analysis
3. **Terminal Output**: Provides immediate feedback on function execution

To access detailed logs:

- Check Langfuse UI for AI agent traces (typically at `http://localhost:4000`)
- Use Grafana dashboards for system metrics (typically at `http://localhost:3000`)
- Examine terminal output for immediate function results and errors

</details>

### Command-line Options

The main utility accepts several options:

- `-env <path>` - Path to environment file (optional, default: `.env`)
- `-provider <type>` - Provider type to use (default: `custom`, options: `openai`, `anthropic`, `custom`)
- `-flow <id>` - Flow ID for testing (0 means using mocks, default: `0`)
- `-task <id>` - Task ID for agent context (optional)
- `-subtask <id>` - Subtask ID for agent context (optional)

Function-specific arguments are passed after the function name using `-name value` format.

## 🏗️ Building

### Building Docker Image

```bash
docker build -t local/pentagi:latest .
```

> [!NOTE]
> You can use `docker buildx` to build the image for different platforms like a `docker buildx build --platform linux/amd64 -t local/pentagi:latest .`
>
> You need to change image name in docker-compose.yml file to `local/pentagi:latest` and run `docker compose up -d` to start the server or use `build` key option in [docker-compose.yml](docker-compose.yml) file.

## 👏 Credits

This project is made possible thanks to the following research and developments:
- [Emerging Architectures for LLM Applications](https://lilianweng.github.io/posts/2023-06-23-agent)
- [A Survey of Autonomous LLM Agents](https://arxiv.org/abs/2403.08299)

## 📄 License

Copyright (c) PentAGI Development Team. [MIT License](https://github.com/vxcontrol/pentagi/blob/master/LICENSE)
