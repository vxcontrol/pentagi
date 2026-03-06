# PentAGI

<div align="center" style="font-size: 1.5em; margin: 20px 0;">
    <strong>P</strong>enetration testing <strong>A</strong>rtificial <strong>G</strong>eneral <strong>I</strong>ntelligence
</div>
<br>
<div align="center">

> 🚀 **Join the Community!** Connect with security researchers, AI enthusiasts, and fellow ethical hackers. Get support, share insights, and stay updated with the latest PentAGI developments.

[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/2xrMh7qX6m)⠀[![Telegram](https://img.shields.io/badge/Telegram-2CA5E0?logo=telegram&logoColor=white)](https://t.me/+Ka9i6CNwe71hMWQy)

<a href="https://trendshift.io/repositories/15161" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15161" alt="vxcontrol%2Fpentagi | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

</div>

## 📖 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Quick Start](#-quick-start)
- [API Access](#-api-access)
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
- 📚 Knowledge Graph Integration. Graphiti-powered knowledge graph using Neo4j for semantic relationship tracking and advanced context understanding.
- 🔍 Web Intelligence. Built-in browser via [scraper](https://hub.docker.com/r/vxcontrol/scraper) for gathering latest information from web sources.
- 🔎 External Search Systems. Integration with advanced search APIs including [Tavily](https://tavily.com), [Traversaal](https://traversaal.ai), [Perplexity](https://www.perplexity.ai), [DuckDuckGo](https://duckduckgo.com/), [Google Custom Search](https://programmablesearchengine.google.com/), [Sploitus Search](https://sploitus.com) and [Searxng](https://searxng.org) for comprehensive information gathering.
- 👥 Team of Specialists. Delegation system with specialized AI agents for research, development, and infrastructure tasks.
- 📊 Comprehensive Monitoring. Detailed logging and integration with Grafana/Prometheus for real-time system observation.
- 📝 Detailed Reporting. Generation of thorough vulnerability reports with exploitation guides.
- 📦 Smart Container Management. Automatic Docker image selection based on specific task requirements.
- 📱 Modern Interface. Clean and intuitive web UI for system management and monitoring.
- 🔌 Comprehensive APIs. Full-featured REST and GraphQL APIs with Bearer token authentication for automation and integration.
- 💾 Persistent Storage. All commands and outputs are stored in PostgreSQL with [pgvector](https://hub.docker.com/r/vxcontrol/pgvector) extension.
- 🎯 Scalable Architecture. Microservices-based design supporting horizontal scaling.
- 🏠 Self-Hosted Solution. Complete control over your deployment and data.
- 🔑 Flexible Authentication. Support for various LLM providers ([OpenAI](https://platform.openai.com/), [Anthropic](https://www.anthropic.com/), [Ollama](https://ollama.com/), [AWS Bedrock](https://aws.amazon.com/bedrock/), [Google AI/Gemini](https://ai.google.dev/), [Deep Infra](https://deepinfra.com/), [OpenRouter](https://openrouter.ai/), [DeepSeek](https://www.deepseek.com/en)), [Moonshot](https://platform.moonshot.ai/) and custom configurations.
- 🔐 API Token Authentication. Secure Bearer token system for programmatic access to REST and GraphQL APIs.
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
    (OpenAI/Anthropic/Ollama/Bedrock/Gemini/Custom)"]
    search["🔍 search-systems
    (Google/DuckDuckGo/Tavily/Traversaal/Perplexity/Sploitus/Searxng)"]
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

    subgraph Knowledge Graph
        Graphiti[Graphiti<br/>Knowledge Graph API]
        Neo4j[(Neo4j<br/>Graph Database)]
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
    Agent --> |Commands| PenTest
    Agent --> |Queries| DB
    Agent --> |Knowledge| Graphiti
    Graphiti --> |Graph| Neo4j

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
    classDef knowledge fill:#ffa,stroke:#333,stroke-width:2px,color:#000
    classDef monitoring fill:#bbf,stroke:#333,stroke-width:2px,color:#000
    classDef analytics fill:#bfb,stroke:#333,stroke-width:2px,color:#000
    classDef tools fill:#fbb,stroke:#333,stroke-width:2px,color:#000

    class UI,API,DB,MQ,Agent core
    class Graphiti,Neo4j knowledge
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

| Parameter             | Environment Variable             | Default | Description                                                |
| --------------------- | -------------------------------- | ------- | ---------------------------------------------------------- |
| Preserve Last         | `SUMMARIZER_PRESERVE_LAST`       | `true`  | Whether to keep all messages in the last section intact    |
| Use QA Pairs          | `SUMMARIZER_USE_QA`              | `true`  | Whether to use QA pair summarization strategy              |
| Summarize Human in QA | `SUMMARIZER_SUM_MSG_HUMAN_IN_QA` | `false` | Whether to summarize human messages in QA pairs            |
| Last Section Size     | `SUMMARIZER_LAST_SEC_BYTES`      | `51200` | Maximum byte size for last section (50KB)                  |
| Max Body Pair Size    | `SUMMARIZER_MAX_BP_BYTES`        | `16384` | Maximum byte size for a single body pair (16KB)            |
| Max QA Sections       | `SUMMARIZER_MAX_QA_SECTIONS`     | `10`    | Maximum QA pair sections to preserve                       |
| Max QA Size           | `SUMMARIZER_MAX_QA_BYTES`        | `65536` | Maximum byte size for QA pair sections (64KB)              |
| Keep QA Sections      | `SUMMARIZER_KEEP_QA_SECTIONS`    | `1`     | Number of recent QA sections to keep without summarization |

### Assistant Summarizer Configuration Options

Assistant instances can use customized summarization settings to fine-tune context management behavior:

| Parameter          | Environment Variable                    | Default | Description                                                          |
| ------------------ | --------------------------------------- | ------- | -------------------------------------------------------------------- |
| Preserve Last      | `ASSISTANT_SUMMARIZER_PRESERVE_LAST`    | `true`  | Whether to preserve all messages in the assistant's last section     |
| Last Section Size  | `ASSISTANT_SUMMARIZER_LAST_SEC_BYTES`   | `76800` | Maximum byte size for assistant's last section (75KB)                |
| Max Body Pair Size | `ASSISTANT_SUMMARIZER_MAX_BP_BYTES`     | `16384` | Maximum byte size for a single body pair in assistant context (16KB) |
| Max QA Sections    | `ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS`  | `7`     | Maximum QA sections to preserve in assistant context                 |
| Max QA Size        | `ASSISTANT_SUMMARIZER_MAX_QA_BYTES`     | `76800` | Maximum byte size for assistant's QA sections (75KB)                 |
| Keep QA Sections   | `ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS` | `3`     | Number of recent QA sections to preserve without summarization       |

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
   - Backend API: Go-based REST and GraphQL APIs with Bearer token authentication for programmatic access
   - Vector Store: PostgreSQL with pgvector for semantic search and memory storage
   - Task Queue: Async task processing system for reliable operation
   - AI Agent: Multi-agent system with specialized roles for efficient testing

2. **Knowledge Graph**
   - Graphiti: Knowledge graph API for semantic relationship tracking and contextual understanding
   - Neo4j: Graph database for storing and querying relationships between entities, actions, and outcomes
   - Automatic capturing of agent responses and tool executions for building comprehensive knowledge base

3. **Monitoring Stack**
   - OpenTelemetry: Unified observability data collection and correlation
   - Grafana: Real-time visualization and alerting dashboards
   - VictoriaMetrics: High-performance time-series metrics storage
   - Jaeger: End-to-end distributed tracing for debugging
   - Loki: Scalable log aggregation and analysis

4. **Analytics Platform**
   - Langfuse: Advanced LLM observability and performance analytics
   - ClickHouse: Column-oriented analytics data warehouse
   - Redis: High-speed caching and rate limiting
   - MinIO: S3-compatible object storage for artifacts

5. **Security Tools**
   - Web Scraper: Isolated browser environment for safe web interaction
   - Pentesting Tools: Comprehensive suite of 20+ professional security tools
   - Sandboxed Execution: All operations run in isolated containers

6. **Memory Systems**
   - Long-term Memory: Persistent storage of knowledge and experiences
   - Working Memory: Active context and goals for current operations
   - Episodic Memory: Historical actions and success patterns
   - Knowledge Base: Structured domain expertise and tool capabilities
   - Context Management: Intelligently manages growing LLM context windows using chain summarization

The system uses Docker containers for isolation and easy deployment, with separate networks for core services, monitoring, and analytics to ensure proper security boundaries. Each component is designed to scale horizontally and can be configured for high availability in production environments.

## 🚀 Quick Start

### System Requirements

- Docker and Docker Compose (or Podman - see [Podman configuration](#running-pentagi-with-podman))
- Minimum 2 vCPU
- Minimum 4GB RAM
- 20GB free disk space
- Internet access for downloading images and updates

### Using Installer (Recommended)

PentAGI provides an interactive installer with a terminal-based UI for streamlined configuration and deployment. The installer guides you through system checks, LLM provider setup, search engine configuration, and security hardening.

**Supported Platforms:**
- **Linux**: amd64 [download](https://pentagi.com/downloads/linux/amd64/installer-latest.zip) | arm64 [download](https://pentagi.com/downloads/linux/arm64/installer-latest.zip)
- **Windows**: amd64 [download](https://pentagi.com/downloads/windows/amd64/installer-latest.zip)
- **macOS**: amd64 (Intel) [download](https://pentagi.com/downloads/darwin/amd64/installer-latest.zip) | arm64 (M-series) [download](https://pentagi.com/downloads/darwin/arm64/installer-latest.zip)

**Quick Installation (Linux amd64):**

```bash
# Create installation directory
mkdir -p pentagi && cd pentagi

# Download installer
wget -O installer.zip https://pentagi.com/downloads/linux/amd64/installer-latest.zip

# Extract
unzip installer.zip

# Run interactive installer
./installer
```

**Prerequisites & Permissions:**

The installer requires appropriate privileges to interact with the Docker API for proper operation. By default, it uses the Docker socket (`/var/run/docker.sock`) which requires either:

- **Option 1 (Recommended for production):** Run the installer as root:
  ```bash
  sudo ./installer
  ```

- **Option 2 (Development environments):** Grant your user access to the Docker socket by adding them to the `docker` group:
  ```bash
  # Add your user to the docker group
  sudo usermod -aG docker $USER
  
  # Log out and log back in, or activate the group immediately
  newgrp docker
  
  # Verify Docker access (should run without sudo)
  docker ps
  ```

  ⚠️ **Security Note:** Adding a user to the `docker` group grants root-equivalent privileges. Only do this for trusted users in controlled environments. For production deployments, consider using rootless Docker mode or running the installer with sudo.

The installer will:
1. **System Checks**: Verify Docker, network connectivity, and system requirements
2. **Environment Setup**: Create and configure `.env` file with optimal defaults
3. **Provider Configuration**: Set up LLM providers (OpenAI, Anthropic, Gemini, Bedrock, Ollama, Custom)
4. **Search Engines**: Configure DuckDuckGo, Google, Tavily, Traversaal, Perplexity, Sploitus, Searxng
5. **Security Hardening**: Generate secure credentials and configure SSL certificates
6. **Deployment**: Start PentAGI with docker-compose

**For Production & Enhanced Security:**

For production deployments or security-sensitive environments, we **strongly recommend** using a distributed two-node architecture where worker operations are isolated on a separate server. This prevents untrusted code execution and network access issues on your main system.

👉 **See detailed guide**: [Worker Node Setup](examples/guides/worker_node.md)

The two-node setup provides:
- **Isolated Execution**: Worker containers run on dedicated hardware
- **Network Isolation**: Separate network boundaries for penetration testing
- **Security Boundaries**: Docker-in-Docker with TLS authentication
- **OOB Attack Support**: Dedicated port ranges for out-of-band techniques

### Manual Installation

1. Create a working directory or clone the repository:

```bash
mkdir pentagi && cd pentagi
```

2. Copy `.env.example` to `.env` or download it:

```bash
curl -o .env https://raw.githubusercontent.com/vxcontrol/pentagi/master/.env.example
```

3. Touch examples files (`example.custom.provider.yml`, `example.ollama.provider.yml`) or download it:

```bash
curl -o example.custom.provider.yml https://raw.githubusercontent.com/vxcontrol/pentagi/master/examples/configs/custom-openai.provider.yml
curl -o example.ollama.provider.yml https://raw.githubusercontent.com/vxcontrol/pentagi/master/examples/configs/ollama-llama318b.provider.yml
```

4. Fill in the required API keys in `.env` file.

```bash
# Required: At least one of these LLM providers
OPEN_AI_KEY=your_openai_key
ANTHROPIC_API_KEY=your_anthropic_key
GEMINI_API_KEY=your_gemini_key

# Optional: AWS Bedrock provider (enterprise-grade models)
BEDROCK_REGION=us-east-1
BEDROCK_ACCESS_KEY_ID=your_aws_access_key
BEDROCK_SECRET_ACCESS_KEY=your_aws_secret_key

# Optional: Local LLM provider (zero-cost inference)
OLLAMA_SERVER_URL=http://localhost:11434
OLLAMA_SERVER_MODEL=your_model_name

# Optional: Additional search capabilities
DUCKDUCKGO_ENABLED=true
SPLOITUS_ENABLED=true
GOOGLE_API_KEY=your_google_key
GOOGLE_CX_KEY=your_google_cx
TAVILY_API_KEY=your_tavily_key
TRAVERSAAL_API_KEY=your_traversaal_key
PERPLEXITY_API_KEY=your_perplexity_key
PERPLEXITY_MODEL=sonar-pro
PERPLEXITY_CONTEXT_SIZE=medium

# Searxng meta search engine (aggregates results from multiple sources)
SEARXNG_URL=http://your-searxng-instance:8080
SEARXNG_CATEGORIES=general
SEARXNG_LANGUAGE=
SEARXNG_SAFESEARCH=0
SEARXNG_TIME_RANGE=

## Graphiti knowledge graph settings
GRAPHITI_ENABLED=true
GRAPHITI_TIMEOUT=30
GRAPHITI_URL=http://graphiti:8000
GRAPHITI_MODEL_NAME=gpt-5-mini

# Neo4j settings (used by Graphiti stack)
NEO4J_USER=neo4j
NEO4J_DATABASE=neo4j
NEO4J_PASSWORD=devpassword
NEO4J_URI=bolt://neo4j:7687

# Assistant configuration
ASSISTANT_USE_AGENTS=false         # Default value for agent usage when creating new assistants
```

5. Change all security related environment variables in `.env` file to improve security.

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
- `NEO4J_USER` and `NEO4J_PASSWORD` - Neo4j credentials (for Graphiti knowledge graph)

</details>

6. Remove all inline comments from `.env` file if you want to use it in VSCode or other IDEs as a envFile option:

```bash
perl -i -pe 's/\s+#.*$//' .env
```

7. Run the PentAGI stack:

```bash
curl -O https://raw.githubusercontent.com/vxcontrol/pentagi/master/docker-compose.yml
docker compose up -d
```

Visit [localhost:8443](https://localhost:8443) to access PentAGI Web UI (default is `admin@pentagi.com` / `admin`)

> [!NOTE]
> If you caught an error about `pentagi-network` or `observability-network` or `langfuse-network` you need to run `docker-compose.yml` firstly to create these networks and after that run `docker-compose-langfuse.yml`, `docker-compose-graphiti.yml`, and `docker-compose-observability.yml` to use Langfuse, Graphiti, and Observability services.
>
> You have to set at least one Language Model provider (OpenAI, Anthropic, Gemini, AWS Bedrock, or Ollama) to use PentAGI. AWS Bedrock provides enterprise-grade access to multiple foundation models from leading AI companies, while Ollama provides zero-cost local inference if you have sufficient computational resources. Additional API keys for search engines are optional but recommended for better results.
>
> `LLM_SERVER_*` environment variables are experimental feature and will be changed in the future. Right now you can use them to specify custom LLM server URL and one model for all agent types.
>
> `PROXY_URL` is a global proxy URL for all LLM providers and external search systems. You can use it for isolation from external networks.
>
> The `docker-compose.yml` file runs the PentAGI service as root user because it needs access to docker.sock for container management. If you're using TCP/IP network connection to Docker instead of socket file, you can remove root privileges and use the default `pentagi` user for better security.

### Accessing PentAGI from External Networks

By default, PentAGI binds to `127.0.0.1` (localhost only) for security. To access PentAGI from other machines on your network, you need to configure external access.

#### Configuration Steps

1. **Update `.env` file** with your server's IP address:

```bash
# Network binding - allow external connections
PENTAGI_LISTEN_IP=0.0.0.0
PENTAGI_LISTEN_PORT=8443

# Public URL - use your actual server IP or hostname
# Replace 192.168.1.100 with your server's IP address
PUBLIC_URL=https://192.168.1.100:8443

# CORS origins - list all URLs that will access PentAGI
# Include localhost for local access AND your server IP for external access
CORS_ORIGINS=https://localhost:8443,https://192.168.1.100:8443
```

> [!IMPORTANT]
> - Replace `192.168.1.100` with your actual server's IP address
> - Do NOT use `0.0.0.0` in `PUBLIC_URL` or `CORS_ORIGINS` - use the actual IP address
> - Include both localhost and your server IP in `CORS_ORIGINS` for flexibility

2. **Recreate containers** to apply the changes:

```bash
docker compose down
docker compose up -d --force-recreate
```

3. **Verify port binding:**

```bash
docker ps | grep pentagi
```

You should see `0.0.0.0:8443->8443/tcp` or `:::8443->8443/tcp`.

If you see `127.0.0.1:8443->8443/tcp`, the environment variable wasn't picked up. In this case, directly edit `docker-compose.yml` line 31:

```yaml
ports:
  - "0.0.0.0:8443:8443"
```

Then recreate containers again.

4. **Configure firewall** to allow incoming connections on port 8443:

```bash
# Ubuntu/Debian with UFW
sudo ufw allow 8443/tcp
sudo ufw reload

# CentOS/RHEL with firewalld
sudo firewall-cmd --permanent --add-port=8443/tcp
sudo firewall-cmd --reload
```

5. **Access PentAGI:**

- **Local access:** `https://localhost:8443`
- **Network access:** `https://your-server-ip:8443`

> [!NOTE]
> You'll need to accept the self-signed SSL certificate warning in your browser when accessing via IP address.

---

### Running PentAGI with Podman

PentAGI fully supports Podman as a Docker alternative. However, when using **Podman in rootless mode**, the scraper service requires special configuration because rootless containers cannot bind privileged ports (ports below 1024).

#### Podman Rootless Configuration

The default scraper configuration uses port 443 (HTTPS), which is a privileged port. For Podman rootless, reconfigure the scraper to use a non-privileged port:

**1. Edit `docker-compose.yml`** - modify the `scraper` service (around line 199):

```yaml
scraper:
  image: vxcontrol/scraper:latest
  restart: unless-stopped
  container_name: scraper
  hostname: scraper
  expose:
    - 3000/tcp  # Changed from 443 to 3000
  ports:
    - "${SCRAPER_LISTEN_IP:-127.0.0.1}:${SCRAPER_LISTEN_PORT:-9443}:3000"  # Map to port 3000
  environment:
    - MAX_CONCURRENT_SESSIONS=${LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS:-10}
    - USERNAME=${LOCAL_SCRAPER_USERNAME:-someuser}
    - PASSWORD=${LOCAL_SCRAPER_PASSWORD:-somepass}
  logging:
    options:
      max-size: 50m
      max-file: "7"
  volumes:
    - scraper-ssl:/usr/src/app/ssl
  networks:
    - pentagi-network
  shm_size: 2g
```

**2. Update `.env` file** - change the scraper URL to use HTTP and port 3000:

```bash
# Scraper configuration for Podman rootless
SCRAPER_PRIVATE_URL=http://someuser:somepass@scraper:3000/
LOCAL_SCRAPER_USERNAME=someuser
LOCAL_SCRAPER_PASSWORD=somepass
```

> [!IMPORTANT]
> Key changes for Podman:
> - Use **HTTP** instead of HTTPS for `SCRAPER_PRIVATE_URL`
> - Use port **3000** instead of 443
> - Change internal `expose` to `3000/tcp`
> - Update port mapping to target `3000` instead of `443`

**3. Recreate containers:**

```bash
podman-compose down
podman-compose up -d --force-recreate
```

**4. Test scraper connectivity:**

```bash
# Test from within the pentagi container
podman exec -it pentagi wget -O- "http://someuser:somepass@scraper:3000/html?url=http://example.com"
```

If you see HTML output, the scraper is working correctly.

#### Podman Rootful Mode

If you're running Podman in rootful mode (with sudo), you can use the default configuration without modifications. The scraper will work on port 443 as intended.

#### Docker Compatibility

All Podman configurations remain fully compatible with Docker. The non-privileged port approach works identically on both container runtimes.

### Assistant Configuration

PentAGI allows you to configure default behavior for assistants:

| Variable               | Default | Description                                                             |
| ---------------------- | ------- | ----------------------------------------------------------------------- |
| `ASSISTANT_USE_AGENTS` | `false` | Controls the default value for agent usage when creating new assistants |

The `ASSISTANT_USE_AGENTS` setting affects the initial state of the "Use Agents" toggle when creating a new assistant in the UI:
- `false` (default): New assistants are created with agent delegation disabled by default
- `true`: New assistants are created with agent delegation enabled by default

Note that users can always override this setting by toggling the "Use Agents" button in the UI when creating or editing an assistant. This environment variable only controls the initial default state.

## 🔌 API Access

PentAGI provides comprehensive programmatic access through both REST and GraphQL APIs, allowing you to integrate penetration testing workflows into your automation pipelines, CI/CD processes, and custom applications.

### Generating API Tokens

API tokens are managed through the PentAGI web interface:

1. Navigate to **Settings** → **API Tokens** in the web UI
2. Click **Create Token** to generate a new API token
3. Configure token properties:
   - **Name** (optional): A descriptive name for the token
   - **Expiration Date**: When the token will expire (minimum 1 minute, maximum 3 years)
4. Click **Create** and **copy the token immediately** - it will only be shown once for security reasons
5. Use the token as a Bearer token in your API requests

Each token is associated with your user account and inherits your role's permissions.

### Using API Tokens

Include the API token in the `Authorization` header of your HTTP requests:

```bash
# GraphQL API example
curl -X POST https://your-pentagi-instance:8443/api/v1/graphql \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "{ flows { id title status } }"}'

# REST API example
curl https://your-pentagi-instance:8443/api/v1/flows \
  -H "Authorization: Bearer YOUR_API_TOKEN"
```

### API Exploration and Testing

PentAGI provides interactive documentation for exploring and testing API endpoints:

#### GraphQL Playground

Access the GraphQL Playground at `https://your-pentagi-instance:8443/api/v1/graphql/playground`

1. Click the **HTTP Headers** tab at the bottom
2. Add your authorization header:
   ```json
   {
     "Authorization": "Bearer YOUR_API_TOKEN"
   }
   ```
3. Explore the schema, run queries, and test mutations interactively

#### Swagger UI

Access the REST API documentation at `https://your-pentagi-instance:8443/api/v1/swagger/index.html`

1. Click the **Authorize** button
2. Enter your token in the format: `Bearer YOUR_API_TOKEN`
3. Click **Authorize** to apply
4. Test endpoints directly from the Swagger UI

### Generating API Clients

You can generate type-safe API clients for your preferred programming language using the schema files included with PentAGI:

#### GraphQL Clients

The GraphQL schema is available at:
- **Web UI**: Navigate to Settings to download `schema.graphqls`
- **Direct file**: `backend/pkg/graph/schema.graphqls` in the repository

Generate clients using tools like:
- **GraphQL Code Generator** (JavaScript/TypeScript): [https://the-guild.dev/graphql/codegen](https://the-guild.dev/graphql/codegen)
- **genqlient** (Go): [https://github.com/Khan/genqlient](https://github.com/Khan/genqlient)
- **Apollo iOS** (Swift): [https://www.apollographql.com/docs/ios](https://www.apollographql.com/docs/ios)

#### REST API Clients

The OpenAPI specification is available at:
- **Swagger JSON**: `https://your-pentagi-instance:8443/api/v1/swagger/doc.json`
- **Swagger YAML**: Available in `backend/pkg/server/docs/swagger.yaml`

Generate clients using:
- **OpenAPI Generator**: [https://openapi-generator.tech](https://openapi-generator.tech)
  ```bash
  openapi-generator-cli generate \
    -i https://your-pentagi-instance:8443/api/v1/swagger/doc.json \
    -g python \
    -o ./pentagi-client
  ```

- **Swagger Codegen**: [https://github.com/swagger-api/swagger-codegen](https://github.com/swagger-api/swagger-codegen)
  ```bash
  swagger-codegen generate \
    -i https://your-pentagi-instance:8443/api/v1/swagger/doc.json \
    -l typescript-axios \
    -o ./pentagi-client
  ```

- **swagger-typescript-api** (TypeScript): [https://github.com/acacode/swagger-typescript-api](https://github.com/acacode/swagger-typescript-api)
  ```bash
  npx swagger-typescript-api \
    -p https://your-pentagi-instance:8443/api/v1/swagger/doc.json \
    -o ./src/api \
    -n pentagi-api.ts
  ```

### API Usage Examples

<details>
<summary><b>Creating a New Flow (GraphQL)</b></summary>

```graphql
mutation CreateFlow {
  createFlow(
    modelProvider: "openai"
    input: "Test the security of https://example.com"
  ) {
    id
    title
    status
    createdAt
  }
}
```

</details>

<details>
<summary><b>Listing Flows (REST API)</b></summary>

```bash
curl https://your-pentagi-instance:8443/api/v1/flows \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  | jq '.flows[] | {id, title, status}'
```

</details>

<details>
<summary><b>Python Client Example</b></summary>

```python
import requests

class PentAGIClient:
    def __init__(self, base_url, api_token):
        self.base_url = base_url
        self.headers = {
            "Authorization": f"Bearer {api_token}",
            "Content-Type": "application/json"
        }
    
    def create_flow(self, provider, target):
        query = """
        mutation CreateFlow($provider: String!, $input: String!) {
          createFlow(modelProvider: $provider, input: $input) {
            id
            title
            status
          }
        }
        """
        response = requests.post(
            f"{self.base_url}/api/v1/graphql",
            json={
                "query": query,
                "variables": {
                    "provider": provider,
                    "input": target
                }
            },
            headers=self.headers
        )
        return response.json()
    
    def get_flows(self):
        response = requests.get(
            f"{self.base_url}/api/v1/flows",
            headers=self.headers
        )
        return response.json()

# Usage
client = PentAGIClient(
    "https://your-pentagi-instance:8443",
    "your_api_token_here"
)

# Create a new flow
flow = client.create_flow("openai", "Scan https://example.com for vulnerabilities")
print(f"Created flow: {flow}")

# List all flows
flows = client.get_flows()
print(f"Total flows: {len(flows['flows'])}")
```

</details>

<details>
<summary><b>TypeScript Client Example</b></summary>

```typescript
import axios, { AxiosInstance } from 'axios';

interface Flow {
  id: string;
  title: string;
  status: string;
  createdAt: string;
}

class PentAGIClient {
  private client: AxiosInstance;

  constructor(baseURL: string, apiToken: string) {
    this.client = axios.create({
      baseURL: `${baseURL}/api/v1`,
      headers: {
        'Authorization': `Bearer ${apiToken}`,
        'Content-Type': 'application/json',
      },
    });
  }

  async createFlow(provider: string, input: string): Promise<Flow> {
    const query = `
      mutation CreateFlow($provider: String!, $input: String!) {
        createFlow(modelProvider: $provider, input: $input) {
          id
          title
          status
          createdAt
        }
      }
    `;

    const response = await this.client.post('/graphql', {
      query,
      variables: { provider, input },
    });

    return response.data.data.createFlow;
  }

  async getFlows(): Promise<Flow[]> {
    const response = await this.client.get('/flows');
    return response.data.flows;
  }

  async getFlow(flowId: string): Promise<Flow> {
    const response = await this.client.get(`/flows/${flowId}`);
    return response.data;
  }
}

// Usage
const client = new PentAGIClient(
  'https://your-pentagi-instance:8443',
  'your_api_token_here'
);

// Create a new flow
const flow = await client.createFlow(
  'openai',
  'Perform penetration test on https://example.com'
);
console.log('Created flow:', flow);

// List all flows
const flows = await client.getFlows();
console.log(`Total flows: ${flows.length}`);
```

</details>

### Security Best Practices

When working with API tokens:

- **Never commit tokens to version control** - use environment variables or secrets management
- **Rotate tokens regularly** - set appropriate expiration dates and create new tokens periodically
- **Use separate tokens for different applications** - makes it easier to revoke access if needed
- **Monitor token usage** - review API token activity in the Settings page
- **Revoke unused tokens** - disable or delete tokens that are no longer needed
- **Use HTTPS only** - never send API tokens over unencrypted connections

### Token Management

- **View tokens**: See all your active tokens in Settings → API Tokens
- **Edit tokens**: Update token names or revoke tokens
- **Delete tokens**: Permanently remove tokens (this action cannot be undone)
- **Token ID**: Each token has a unique ID that can be copied for reference

The token list shows:
- Token name (if provided)
- Token ID (unique identifier)
- Status (active/revoked/expired)
- Creation date
- Expiration date

### Custom LLM Provider Configuration

When using custom LLM providers with the `LLM_SERVER_*` variables, you can fine-tune the reasoning format used in requests:

| Variable                        | Default | Description                                                                             |
| ------------------------------- | ------- | --------------------------------------------------------------------------------------- |
| `LLM_SERVER_URL`                |         | Base URL for the custom LLM API endpoint                                                |
| `LLM_SERVER_KEY`                |         | API key for the custom LLM provider                                                     |
| `LLM_SERVER_MODEL`              |         | Default model to use (can be overridden in provider config)                             |
| `LLM_SERVER_CONFIG_PATH`        |         | Path to the YAML configuration file for agent-specific models                           |
| `LLM_SERVER_PROVIDER`           |         | Provider name prefix for model names (e.g., `openrouter`, `deepseek` for LiteLLM proxy) |
| `LLM_SERVER_LEGACY_REASONING`   | `false` | Controls reasoning format in API requests                                               |
| `LLM_SERVER_PRESERVE_REASONING` | `false` | Preserve reasoning content in multi-turn conversations (required by some providers)     |

The `LLM_SERVER_PROVIDER` setting is particularly useful when using **LiteLLM proxy**, which adds a provider prefix to model names. For example, when connecting to Moonshot API through LiteLLM, models like `kimi-2.5` become `moonshot/kimi-2.5`. By setting `LLM_SERVER_PROVIDER=moonshot`, you can use the same provider configuration file for both direct API access and LiteLLM proxy access without modifications.

The `LLM_SERVER_LEGACY_REASONING` setting affects how reasoning parameters are sent to the LLM:
- `false` (default): Uses modern format where reasoning is sent as a structured object with `max_tokens` parameter
- `true`: Uses legacy format with string-based `reasoning_effort` parameter

This setting is important when working with different LLM providers as they may expect different reasoning formats in their API requests. If you encounter reasoning-related errors with custom providers, try changing this setting.

The `LLM_SERVER_PRESERVE_REASONING` setting controls whether reasoning content is preserved in multi-turn conversations:
- `false` (default): Reasoning content is not preserved in conversation history
- `true`: Reasoning content is preserved and sent in subsequent API calls

This setting is required by some LLM providers (e.g., Moonshot) that return errors like "thinking is enabled but reasoning_content is missing in assistant tool call message" when reasoning content is not included in multi-turn conversations. Enable this setting if your provider requires reasoning content to be preserved.

### Local LLM Provider Configuration

PentAGI supports Ollama for local LLM inference, providing zero-cost operation and enhanced privacy:

| Variable                            | Default                     | Description                             |
| ----------------------------------- | --------------------------- | --------------------------------------- |
| `OLLAMA_SERVER_URL`                 |                             | URL of your Ollama server               |
| `OLLAMA_SERVER_MODEL`               | `llama3.1:8b-instruct-q8_0` | Default model for inference             |
| `OLLAMA_SERVER_CONFIG_PATH`         |                             | Path to custom agent configuration file |
| `OLLAMA_SERVER_PULL_MODELS_TIMEOUT` | `600`                       | Timeout for model downloads (seconds)   |
| `OLLAMA_SERVER_PULL_MODELS_ENABLED` | `false`                     | Auto-download models on startup         |
| `OLLAMA_SERVER_LOAD_MODELS_ENABLED` | `false`                     | Query server for available models       |

Configuration examples:

```bash
# Basic Ollama setup with default model
OLLAMA_SERVER_URL=http://localhost:11434
OLLAMA_SERVER_MODEL=llama3.1:8b-instruct-q8_0

# Production setup with auto-pull and model discovery
OLLAMA_SERVER_URL=http://ollama-server:11434
OLLAMA_SERVER_PULL_MODELS_ENABLED=true
OLLAMA_SERVER_PULL_MODELS_TIMEOUT=900
OLLAMA_SERVER_LOAD_MODELS_ENABLED=true

# Custom configuration with agent-specific models
OLLAMA_SERVER_CONFIG_PATH=/path/to/ollama-config.yml

# Default configuration file inside docker container
OLLAMA_SERVER_CONFIG_PATH=/opt/pentagi/conf/ollama-llama318b.provider.yml
```

**Performance Considerations:**

- **Model Discovery** (`OLLAMA_SERVER_LOAD_MODELS_ENABLED=true`): Adds 1-2s startup latency querying Ollama API
- **Auto-pull** (`OLLAMA_SERVER_PULL_MODELS_ENABLED=true`): First startup may take several minutes downloading models
- **Pull timeout** (`OLLAMA_SERVER_PULL_MODELS_TIMEOUT=900`): 15 minutes in seconds
- **Static Config**: Disable both flags and specify models in config file for fastest startup

#### Creating Custom Ollama Models with Extended Context

PentAGI requires models with larger context windows than the default Ollama configurations. You need to create custom models with increased `num_ctx` parameter through Modelfiles. While typical agent workflows consume around 64K tokens, PentAGI uses 110K context size for safety margin and handling complex penetration testing scenarios.

**Important**: The `num_ctx` parameter can only be set during model creation via Modelfile - it cannot be changed after model creation or overridden at runtime.

##### Example: Qwen3 32B FP16 with Extended Context

Create a Modelfile named `Modelfile_qwen3_32b_fp16_tc`:

```dockerfile
FROM qwen3:32b-fp16
PARAMETER num_ctx 110000
PARAMETER temperature 0.3
PARAMETER top_p 0.8
PARAMETER min_p 0.0
PARAMETER top_k 20
PARAMETER repeat_penalty 1.1
```

Build the custom model:

```bash
ollama create qwen3:32b-fp16-tc -f Modelfile_qwen3_32b_fp16_tc
```

##### Example: QwQ 32B FP16 with Extended Context

Create a Modelfile named `Modelfile_qwq_32b_fp16_tc`:

```dockerfile
FROM qwq:32b-fp16
PARAMETER num_ctx 110000
PARAMETER temperature 0.2
PARAMETER top_p 0.7
PARAMETER min_p 0.0
PARAMETER top_k 40
PARAMETER repeat_penalty 1.2
```

Build the custom model:

```bash
ollama create qwq:32b-fp16-tc -f Modelfile_qwq_32b_fp16_tc
```

> **Note**: The QwQ 32B FP16 model requires approximately **71.3 GB VRAM** for inference. Ensure your system has sufficient GPU memory before attempting to use this model.

These custom models are referenced in the pre-built provider configuration files (`ollama-qwen332b-fp16-tc.provider.yml` and `ollama-qwq32b-fp16-tc.provider.yml`) that are included in the Docker image at `/opt/pentagi/conf/`.

### OpenAI Provider Configuration

PentAGI supports OpenAI's advanced language models, including the latest reasoning-capable o-series models designed for complex analytical tasks:

| Variable             | Default                     | Description                 |
| -------------------- | --------------------------- | --------------------------- |
| `OPEN_AI_KEY`        |                             | API key for OpenAI services |
| `OPEN_AI_SERVER_URL` | `https://api.openai.com/v1` | OpenAI API endpoint         |

Configuration examples:

```bash
# Basic OpenAI setup
OPEN_AI_KEY=your_openai_api_key
OPEN_AI_SERVER_URL=https://api.openai.com/v1

# Using with proxy for enhanced security
OPEN_AI_KEY=your_openai_api_key
PROXY_URL=http://your-proxy:8080
```

The OpenAI provider offers cutting-edge capabilities including:

- **Reasoning Models**: Advanced o-series models (o1, o3, o4-mini) with step-by-step analytical thinking
- **Latest GPT-4.1 Series**: Flagship models optimized for complex security research and exploit development
- **Cost-Effective Options**: From nano models for high-volume scanning to powerful reasoning models for deep analysis
- **Versatile Performance**: Fast, intelligent models perfect for multi-step security analysis and penetration testing
- **Proven Reliability**: Industry-leading models with consistent performance across diverse security scenarios

The system automatically selects appropriate OpenAI models based on task complexity, optimizing for both performance and cost-effectiveness.

### Anthropic Provider Configuration

PentAGI integrates with Anthropic's Claude models, known for their exceptional safety, reasoning capabilities, and sophisticated understanding of complex security contexts:

| Variable               | Default                        | Description                    |
| ---------------------- | ------------------------------ | ------------------------------ |
| `ANTHROPIC_API_KEY`    |                                | API key for Anthropic services |
| `ANTHROPIC_SERVER_URL` | `https://api.anthropic.com/v1` | Anthropic API endpoint         |

Configuration examples:

```bash
# Basic Anthropic setup
ANTHROPIC_API_KEY=your_anthropic_api_key
ANTHROPIC_SERVER_URL=https://api.anthropic.com/v1

# Using with proxy for secure environments
ANTHROPIC_API_KEY=your_anthropic_api_key
PROXY_URL=http://your-proxy:8080
```

The Anthropic provider delivers superior capabilities including:

- **Advanced Reasoning**: Claude 4 series with exceptional reasoning for sophisticated penetration testing
- **Extended Thinking**: Claude 3.7 with step-by-step thinking capabilities for methodical security research
- **High-Speed Performance**: Claude 3.5 Haiku for blazing-fast vulnerability scans and real-time monitoring
- **Comprehensive Analysis**: Claude Sonnet models for complex security analysis and threat hunting
- **Safety-First Design**: Built-in safety mechanisms ensuring responsible security testing practices

The system leverages Claude's advanced understanding of security contexts to provide thorough and responsible penetration testing guidance.

### Google AI (Gemini) Provider Configuration

PentAGI supports Google's Gemini models through the Google AI API, offering state-of-the-art reasoning capabilities and multimodal features:

| Variable            | Default                                     | Description                    |
| ------------------- | ------------------------------------------- | ------------------------------ |
| `GEMINI_API_KEY`    |                                             | API key for Google AI services |
| `GEMINI_SERVER_URL` | `https://generativelanguage.googleapis.com` | Google AI API endpoint         |

Configuration examples:

```bash
# Basic Gemini setup
GEMINI_API_KEY=your_gemini_api_key
GEMINI_SERVER_URL=https://generativelanguage.googleapis.com

# Using with proxy
GEMINI_API_KEY=your_gemini_api_key
PROXY_URL=http://your-proxy:8080
```

The Gemini provider offers advanced features including:

- **Thinking Capabilities**: Advanced reasoning models (Gemini 2.5 series) with step-by-step analysis
- **Multimodal Support**: Text and image processing for comprehensive security assessments
- **Large Context Windows**: Up to 2M tokens for analyzing extensive codebases and documentation
- **Cost-Effective Options**: From high-performance pro models to economical flash variants
- **Security-Focused Models**: Specialized configurations optimized for penetration testing workflows

The system automatically selects appropriate Gemini models based on agent requirements, balancing performance, capabilities, and cost-effectiveness.

### AWS Bedrock Provider Configuration

PentAGI integrates with Amazon Bedrock, offering access to a wide range of foundation models from leading AI companies including Anthropic, AI21, Cohere, Meta, and Amazon's own models:

| Variable                    | Default     | Description                                             |
| --------------------------- | ----------- | ------------------------------------------------------- |
| `BEDROCK_REGION`            | `us-east-1` | AWS region for Bedrock service                          |
| `BEDROCK_ACCESS_KEY_ID`     |             | AWS access key ID for authentication                    |
| `BEDROCK_SECRET_ACCESS_KEY` |             | AWS secret access key for authentication                |
| `BEDROCK_SESSION_TOKEN`     |             | AWS session token as alternative way for authentication |
| `BEDROCK_SERVER_URL`        |             | Optional custom Bedrock endpoint URL                    |

Configuration examples:

```bash
# Basic AWS Bedrock setup with credentials
BEDROCK_REGION=us-east-1
BEDROCK_ACCESS_KEY_ID=your_aws_access_key
BEDROCK_SECRET_ACCESS_KEY=your_aws_secret_key

# Using with proxy for enhanced security
BEDROCK_REGION=us-east-1
BEDROCK_ACCESS_KEY_ID=your_aws_access_key
BEDROCK_SECRET_ACCESS_KEY=your_aws_secret_key
PROXY_URL=http://your-proxy:8080

# Using custom endpoint (for VPC endpoints or testing)
BEDROCK_REGION=us-east-1
BEDROCK_ACCESS_KEY_ID=your_aws_access_key
BEDROCK_SECRET_ACCESS_KEY=your_aws_secret_key
BEDROCK_SERVER_URL=https://bedrock-runtime.us-east-1.amazonaws.com
```

> [!IMPORTANT]
> **AWS Bedrock Rate Limits Warning**
>
> The default PentAGI configuration for AWS Bedrock uses two primary models:
> - `us.anthropic.claude-sonnet-4-20250514-v1:0` (for most agents) - **2 requests per minute** for new AWS accounts
> - `us.anthropic.claude-3-5-haiku-20241022-v1:0` (for simple tasks) - **20 requests per minute** for new AWS accounts
>
> These default rate limits are **extremely restrictive** for comfortable penetration testing scenarios and will significantly impact your workflow. We **strongly recommend**:
>
> 1. **Request quota increases** for your AWS Bedrock models through the AWS Service Quotas console
> 2. **Use provisioned throughput models** with hourly billing for higher throughput requirements
> 3. **Switch to alternative models** with higher default quotas (e.g., Amazon Nova series, Meta Llama models)
> 4. **Consider using a different LLM provider** (OpenAI, Anthropic, Gemini) if you need immediate high-throughput access
>
> Without adequate rate limits, you may experience frequent delays, timeouts, and degraded testing performance.

The AWS Bedrock provider delivers comprehensive capabilities including:

- **Multi-Provider Access**: Access to models from Anthropic (Claude), AI21 (Jamba), Cohere (Command), Meta (Llama), Amazon (Nova, Titan), and DeepSeek (R1) through a single interface
- **Advanced Reasoning**: Support for Claude 4 and other reasoning-capable models with step-by-step thinking
- **Multimodal Models**: Amazon Nova series supporting text, image, and video processing for comprehensive security analysis
- **Enterprise Security**: AWS-native security controls, VPC integration, and compliance certifications
- **Cost Optimization**: Wide range of model sizes and capabilities for cost-effective penetration testing
- **Regional Availability**: Deploy models in your preferred AWS region for data residency and performance
- **High Performance**: Low-latency inference through AWS's global infrastructure

The system automatically selects appropriate Bedrock models based on task complexity and requirements, leveraging the full spectrum of available foundation models for optimal security testing results.

> [!WARNING]
> **Converse API Requirements**
>
> PentAGI uses the **Amazon Bedrock Converse API** for model interactions, which requires models to support the following features:
> - ✅ **Converse** - Basic conversation API support
> - ✅ **ConverseStream** - Streaming response support
> - ✅ **Tool use** - Function calling capabilities for penetration testing tools
> - ✅ **Streaming tool use** - Real-time tool execution feedback
>
> **Before selecting models**, verify their feature support at: [Supported models and model features](https://docs.aws.amazon.com/bedrock/latest/userguide/conversation-inference-supported-models-features.html)
>
> ⚠️ **Important**: Some models like AI21 Jurassic-2 and Cohere Command (Text) have **limited chat support** and may not work properly with PentAGI's multi-turn conversation workflows.

> **Note**: AWS credentials can also be provided through IAM roles, environment variables, or AWS credential files following standard AWS SDK authentication patterns. Ensure your AWS account has appropriate permissions for Amazon Bedrock service access.

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
> To run all available stacks together (Langfuse, Graphiti, and Observability):
>
> ```bash
> docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-graphiti.yml -f docker-compose-observability.yml up -d
> ```
>
> You can also register aliases for these commands in your shell to run it faster:
>
> ```bash
> alias pentagi="docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-graphiti.yml -f docker-compose-observability.yml"
> alias pentagi-up="docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-graphiti.yml -f docker-compose-observability.yml up -d"
> alias pentagi-down="docker compose -f docker-compose.yml -f docker-compose-langfuse.yml -f docker-compose-graphiti.yml -f docker-compose-observability.yml down"
> ```

### Knowledge Graph Integration (Graphiti)

PentAGI integrates with [Graphiti](https://github.com/vxcontrol/pentagi-graphiti), a temporal knowledge graph system powered by Neo4j, to provide advanced semantic understanding and relationship tracking for AI agent operations. The vxcontrol fork provides custom entity and edge types that are specific to pentesting purposes.

#### What is Graphiti?

Graphiti automatically extracts and stores structured knowledge from agent interactions, building a graph of entities, relationships, and temporal context. This enables:

- **Semantic Memory**: Store and recall relationships between tools, targets, vulnerabilities, and techniques
- **Contextual Understanding**: Track how different pentesting actions relate to each other over time
- **Knowledge Reuse**: Learn from past penetration tests and apply insights to new assessments
- **Advanced Querying**: Search for complex patterns like "What tools were effective against similar targets?"

#### Enabling Graphiti

The Graphiti knowledge graph is **optional** and disabled by default. To enable it:

1. Configure Graphiti environment variables in `.env` file:

```bash
## Graphiti knowledge graph settings
GRAPHITI_ENABLED=true
GRAPHITI_TIMEOUT=30
GRAPHITI_URL=http://graphiti:8000
GRAPHITI_MODEL_NAME=gpt-5-mini

# Neo4j settings (used by Graphiti stack)
NEO4J_USER=neo4j
NEO4J_DATABASE=neo4j
NEO4J_PASSWORD=devpassword
NEO4J_URI=bolt://neo4j:7687

# OpenAI API key (required by Graphiti for entity extraction)
OPEN_AI_KEY=your_openai_api_key
```

2. Run the Graphiti stack along with the main PentAGI services:

```bash
# Download the Graphiti compose file if needed
curl -O https://raw.githubusercontent.com/vxcontrol/pentagi/master/docker-compose-graphiti.yml

# Start PentAGI with Graphiti
docker compose -f docker-compose.yml -f docker-compose-graphiti.yml up -d
```

3. Verify Graphiti is running:

```bash
# Check service health
docker compose -f docker-compose.yml -f docker-compose-graphiti.yml ps graphiti neo4j

# View Graphiti logs
docker compose -f docker-compose.yml -f docker-compose-graphiti.yml logs -f graphiti

# Access Neo4j Browser (optional)
# Visit http://localhost:7474 and login with NEO4J_USER/NEO4J_PASSWORD

# Access Graphiti API (optional, for debugging)
# Visit http://localhost:8000/docs for Swagger API documentation
```

> [!NOTE]
> The Graphiti service is defined in `docker-compose-graphiti.yml` as a separate stack. You must run both compose files together to enable the knowledge graph functionality. The pre-built Docker image `vxcontrol/graphiti:latest` is used by default.

#### What Gets Stored

When enabled, PentAGI automatically captures:

- **Agent Responses**: All agent reasoning, analysis, and decisions
- **Tool Executions**: Commands executed, tools used, and their results
- **Context Information**: Flow, task, and subtask hierarchy

### GitHub and Google OAuth Integration

OAuth integration with GitHub and Google allows users to authenticate using their existing accounts on these platforms. This provides several benefits:

- Simplified login process without need to create separate credentials
- Enhanced security through trusted identity providers
- Access to user profile information from GitHub/Google accounts
- Seamless integration with existing development workflows

For using GitHub OAuth you need to create a new OAuth application in your GitHub account and set the `OAUTH_GITHUB_CLIENT_ID` and `OAUTH_GITHUB_CLIENT_SECRET` in `.env` file.

For using Google OAuth you need to create a new OAuth application in your Google account and set the `OAUTH_GOOGLE_CLIENT_ID` and `OAUTH_GOOGLE_CLIENT_SECRET` in `.env` file.

### Docker Image Configuration

PentAGI allows you to configure Docker image selection for executing various tasks. The system automatically chooses the most appropriate image based on the task type, but you can constrain this selection by specifying your preferred images:

| Variable                           | Default                | Description                                                 |
| ---------------------------------- | ---------------------- | ----------------------------------------------------------- |
| `DOCKER_DEFAULT_IMAGE`             | `debian:latest`        | Default Docker image for general tasks and ambiguous cases  |
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
- **Comprehensive Test Suite**: Evaluates basic completion, JSON responses, function calling, and penetration testing knowledge
- **Detailed Reporting**: Generates markdown reports with success rates and performance metrics
- **Flexible Configuration**: Test specific agents or test groups as needed
- **Specialized Test Groups**: Includes domain-specific tests for cybersecurity and penetration testing scenarios

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
go run cmd/ctester/*.go -agents simple,simple_json,primary_agent -verbose

# Test specific test groups only
go run cmd/ctester/*.go -groups basic,advanced -verbose
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
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/config.yml -agents simple,primary_agent,coder -verbose

# Generate a detailed report
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  -v $(pwd):/opt/pentagi/output \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -report /opt/pentagi/output/report.md
```

#### Using Pre-configured Providers

The Docker image comes with built-in support for major providers (OpenAI, Anthropic, Gemini, Ollama) and pre-configured provider files for additional services (OpenRouter, DeepInfra, DeepSeek, Moonshot):

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

# Test with Moonshot configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/moonshot.provider.yml

# Test with OpenAI configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -type openai

# Test with Anthropic configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -type anthropic

# Test with Gemini configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -type gemini

# Test with AWS Bedrock configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -type bedrock

# Test with Custom OpenAI configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/custom-openai.provider.yml

# Test with Ollama configuration (local inference)
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/ollama-llama318b.provider.yml

# Test with Ollama Qwen3 32B configuration (requires custom model creation)
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/ollama-qwen332b-fp16-tc.provider.yml

# Test with Ollama QwQ 32B configuration (requires custom model creation and 71.3GB VRAM)
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -config /opt/pentagi/conf/ollama-qwq32b-fp16-tc.provider.yml
```

To use these configurations, your `.env` file only needs to contain:

```
LLM_SERVER_URL=https://openrouter.ai/api/v1      # or https://api.deepinfra.com/v1/openai or https://api.deepseek.com or https://api.openai.com/v1 or https://api.moonshot.ai/v1
LLM_SERVER_KEY=your_api_key
LLM_SERVER_MODEL=                                # Leave empty, as models are specified in the config
LLM_SERVER_CONFIG_PATH=/opt/pentagi/conf/openrouter.provider.yml  # or deepinfra.provider.yml or deepseek.provider.yml or custom-openai.provider.yml or moonshot.provider.yml
LLM_SERVER_PROVIDER=                             # Provider name for LiteLLM proxy (e.g., openrouter, deepseek, moonshot)
LLM_SERVER_LEGACY_REASONING=false                # Controls reasoning format, for OpenAI must be true (default: false)
LLM_SERVER_PRESERVE_REASONING=false              # Preserve reasoning content in multi-turn conversations (required by Moonshot, default: false)

# For OpenAI (official API)
OPEN_AI_KEY=your_openai_api_key                  # Your OpenAI API key
OPEN_AI_SERVER_URL=https://api.openai.com/v1     # OpenAI API endpoint

# For Anthropic (Claude models)
ANTHROPIC_API_KEY=your_anthropic_api_key         # Your Anthropic API key
ANTHROPIC_SERVER_URL=https://api.anthropic.com/v1  # Anthropic API endpoint

# For Gemini (Google AI)
GEMINI_API_KEY=your_gemini_api_key               # Your Google AI API key
GEMINI_SERVER_URL=https://generativelanguage.googleapis.com  # Google AI API endpoint

# For AWS Bedrock (enterprise foundation models)
BEDROCK_REGION=us-east-1                         # AWS region for Bedrock service
BEDROCK_ACCESS_KEY_ID=your_aws_access_key        # AWS access key ID
BEDROCK_SECRET_ACCESS_KEY=your_aws_secret_key    # AWS secret access key
BEDROCK_SESSION_TOKEN=your_aws_session_token     # AWS session token (alternative auth method)
BEDROCK_SERVER_URL=                              # Optional custom Bedrock endpoint

# For Ollama (local inference)
OLLAMA_SERVER_URL=http://localhost:11434
OLLAMA_SERVER_MODEL=llama3.1:8b-instruct-q8_0
OLLAMA_SERVER_CONFIG_PATH=/opt/pentagi/conf/ollama-llama318b.provider.yml
OLLAMA_SERVER_PULL_MODELS_ENABLED=false
OLLAMA_SERVER_LOAD_MODELS_ENABLED=false
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

#### Using LiteLLM Proxy

When using LiteLLM proxy to access various LLM providers, model names are prefixed with the provider name (e.g., `moonshot/kimi-2.5` instead of `kimi-2.5`). To use the same provider configuration files with both direct API access and LiteLLM proxy, set the `LLM_SERVER_PROVIDER` variable:

```bash
# Direct access to Moonshot API
LLM_SERVER_URL=https://api.moonshot.ai/v1
LLM_SERVER_KEY=your_moonshot_api_key
LLM_SERVER_CONFIG_PATH=/opt/pentagi/conf/moonshot.provider.yml
LLM_SERVER_PROVIDER=                             # Empty for direct access

# Access via LiteLLM proxy
LLM_SERVER_URL=http://litellm-proxy:4000
LLM_SERVER_KEY=your_litellm_api_key
LLM_SERVER_CONFIG_PATH=/opt/pentagi/conf/moonshot.provider.yml
LLM_SERVER_PROVIDER=moonshot                     # Provider prefix for LiteLLM
```

With `LLM_SERVER_PROVIDER=moonshot`, the system automatically prefixes all model names from the configuration file with `moonshot/`, making them compatible with LiteLLM's model naming convention.

**Supported provider names for LiteLLM:**
- `openai` - for OpenAI models via LiteLLM
- `anthropic` - for Anthropic/Claude models via LiteLLM
- `gemini` - for Google Gemini models via LiteLLM
- `openrouter` - for OpenRouter aggregator
- `deepseek` - for DeepSeek models
- `deepinfra` - for DeepInfra hosting
- `moonshot` - for Moonshot AI (Kimi)
- Any other provider name configured in your LiteLLM instance

This approach allows you to:
- Use the same configuration files for both direct and proxied access
- Switch between providers without modifying configuration files
- Easily test different routing strategies with LiteLLM

#### Running Tests in a Production Environment

If you already have a running PentAGI container and want to test the current configuration:

```bash
# Run ctester in an existing container using current environment variables
docker exec -it pentagi /opt/pentagi/bin/ctester -verbose

# Test specific agent types with deterministic ordering
docker exec -it pentagi /opt/pentagi/bin/ctester -agents simple,primary_agent,pentester -groups basic,knowledge -verbose

# Generate a report file inside the container
docker exec -it pentagi /opt/pentagi/bin/ctester -report /opt/pentagi/data/agent-test-report.md

# Access the report from the host
docker cp pentagi:/opt/pentagi/data/agent-test-report.md ./
```

### Command-line Options

The utility accepts several options:

- `-env <path>` - Path to environment file (default: `.env`)
- `-type <provider>` - Provider type: `custom`, `openai`, `anthropic`, `ollama`, `bedrock`, `gemini` (default: `custom`)
- `-config <path>` - Path to custom provider config (default: from `LLM_SERVER_CONFIG_PATH` env variable)
- `-tests <path>` - Path to custom tests YAML file (optional)
- `-report <path>` - Path to write the report file (optional)
- `-agents <list>` - Comma-separated list of agent types to test (default: `all`)
- `-groups <list>` - Comma-separated list of test groups to run (default: `all`)
- `-verbose` - Enable verbose output with detailed test results for each agent

### Available Agent Types

Agents are tested in the following deterministic order:

1. **simple** - Basic completion tasks
2. **simple_json** - JSON-structured responses
3. **primary_agent** - Main reasoning agent
4. **assistant** - Interactive assistant mode
5. **generator** - Content generation
6. **refiner** - Content refinement and improvement
7. **adviser** - Expert advice and consultation
8. **reflector** - Self-reflection and analysis
9. **searcher** - Information gathering and search
10. **enricher** - Data enrichment and expansion
11. **coder** - Code generation and analysis
12. **installer** - Installation and setup tasks
13. **pentester** - Penetration testing and security assessment

### Available Test Groups

- **basic** - Fundamental completion and prompt response tests
- **advanced** - Complex reasoning and function calling tests
- **json** - JSON format validation and structure tests (specifically designed for `simple_json` agent)
- **knowledge** - Domain-specific cybersecurity and penetration testing knowledge tests

> **Note**: The `json` test group is specifically designed for the `simple_json` agent type, while all other agents are tested with `basic`, `advanced`, and `knowledge` groups. This specialization ensures optimal testing coverage for each agent's intended purpose.

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

1. **Create a baseline**: Run tests with default configuration to establish benchmark performance
2. **Analyze agent-specific performance**: Review the deterministic agent ordering to identify underperforming agents
3. **Test specialized configurations**: Experiment with different models for each agent type using provider-specific configs
4. **Focus on domain knowledge**: Pay special attention to knowledge group tests for cybersecurity expertise
5. **Validate function calling**: Ensure tool-based tests pass consistently for critical agent types
6. **Compare results**: Look for the best success rate and performance across all test groups
7. **Deploy optimal configuration**: Use in production with your optimized setup

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

# SSL/TLS Certificate Configuration (for external communication with LLM backends and tool servers)
EXTERNAL_SSL_CA_PATH=           # Path to custom CA certificate file (PEM format) inside the container
                                # Must point to /opt/pentagi/ssl/ directory (e.g., /opt/pentagi/ssl/ca-bundle.pem)
EXTERNAL_SSL_INSECURE=false     # Skip certificate verification (use only for testing)
```

<details>
<summary><b>How to Add Custom CA Certificates</b> (click to expand)</summary>

If you see this error: `tls: failed to verify certificate: x509: certificate signed by unknown authority`

**Step 1:** Get your CA certificate bundle in PEM format (can contain multiple certificates)

**Step 2:** Place the file in the SSL directory on your host machine:
```bash
# Default location (if PENTAGI_SSL_DIR is not set)
cp ca-bundle.pem ./pentagi-ssl/

# Or custom location (if using PENTAGI_SSL_DIR in docker-compose.yml)
cp ca-bundle.pem /path/to/your/ssl/dir/
```

**Step 3:** Set the path in `.env` file (path must be inside the container):
```bash
# The volume pentagi-ssl is mounted to /opt/pentagi/ssl inside the container
EXTERNAL_SSL_CA_PATH=/opt/pentagi/ssl/ca-bundle.pem
EXTERNAL_SSL_INSECURE=false
```

**Step 4:** Restart PentAGI:
```bash
docker compose restart pentagi
```

**Notes:**
- The `pentagi-ssl` volume is mounted to `/opt/pentagi/ssl` inside the container
- You can change host directory using `PENTAGI_SSL_DIR` variable in docker-compose.yml
- File supports multiple certificates and intermediate CAs in one PEM file
- Use `EXTERNAL_SSL_INSECURE=true` only for testing (not recommended for production)

</details>

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
- **sploitus**: Search for security exploits, vulnerabilities (CVEs), and pentesting tools
- **searxng**: Search using Searxng meta search engine (aggregates results from multiple engines)

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
- `-provider <type>` - Provider type to use (default: `custom`, options: `openai`, `anthropic`, `ollama`, `bedrock`, `gemini`, `custom`)
- `-flow <id>` - Flow ID for testing (0 means using mocks, default: `0`)
- `-task <id>` - Task ID for agent context (optional)
- `-subtask <id>` - Subtask ID for agent context (optional)

Function-specific arguments are passed after the function name using `-name value` format.

## 🏗️ Building

### Building Docker Image

The Docker build process automatically embeds version information from git tags. To properly version your build, use the provided scripts:

#### Linux/macOS

```bash
# Load version variables
source ./scripts/version.sh

# Standard build
docker build \
  --build-arg PACKAGE_VER=$PACKAGE_VER \
  --build-arg PACKAGE_REV=$PACKAGE_REV \
  -t pentagi:$PACKAGE_VER .

# Multi-platform build
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg PACKAGE_VER=$PACKAGE_VER \
  --build-arg PACKAGE_REV=$PACKAGE_REV \
  -t pentagi:$PACKAGE_VER .

# Build and push
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg PACKAGE_VER=$PACKAGE_VER \
  --build-arg PACKAGE_REV=$PACKAGE_REV \
  -t myregistry/pentagi:$PACKAGE_VER \
  --push .
```

#### Windows (PowerShell)

```powershell
# Load version variables
. .\scripts\version.ps1

# Standard build
docker build `
  --build-arg PACKAGE_VER=$env:PACKAGE_VER `
  --build-arg PACKAGE_REV=$env:PACKAGE_REV `
  -t pentagi:$env:PACKAGE_VER .

# Multi-platform build
docker buildx build `
  --platform linux/amd64,linux/arm64 `
  --build-arg PACKAGE_VER=$env:PACKAGE_VER `
  --build-arg PACKAGE_REV=$env:PACKAGE_REV `
  -t pentagi:$env:PACKAGE_VER .
```

#### Quick build without version

For development builds without version tracking:

```bash
docker build -t pentagi:dev .
```

> [!NOTE]
> - The build scripts automatically determine version from git tags
> - Release builds (on tag commit) have no revision suffix
> - Development builds (after tag) include commit hash as revision (e.g., `1.1.0-bc6e800`)
> - To use the built image locally, update the image name in `docker-compose.yml` or use the `build` option

## 👏 Credits

This project is made possible thanks to the following research and developments:
- [Emerging Architectures for LLM Applications](https://lilianweng.github.io/posts/2023-06-23-agent)
- [A Survey of Autonomous LLM Agents](https://arxiv.org/abs/2403.08299)

## 📄 License

### PentAGI Core License

**PentAGI Core**: Licensed under [MIT License](LICENSE)  
Copyright (c) 2025 PentAGI Development Team

### VXControl Cloud SDK Integration

**VXControl Cloud SDK Integration**: This repository integrates [VXControl Cloud SDK](https://github.com/vxcontrol/cloud) under a **special licensing exception** that applies **ONLY** to the official PentAGI project.

#### ✅ Official PentAGI Project
- This official repository: `https://github.com/vxcontrol/pentagi`
- Official releases distributed by VXControl LLC
- Code used under direct authorization from VXControl LLC

#### ⚠️ Important for Forks and Third-Party Use

If you fork this project or create derivative works, the VXControl SDK components are subject to **AGPL-3.0** license terms. You must either:

1. **Remove VXControl SDK integration**
2. **Open source your entire application** (comply with AGPL-3.0 copyleft terms)
3. **Obtain a commercial license** from VXControl LLC

#### Commercial Licensing

For commercial use of VXControl Cloud SDK in proprietary applications, contact:
- **Email**: info@vxcontrol.com  
- **Subject**: "VXControl Cloud SDK Commercial License"
