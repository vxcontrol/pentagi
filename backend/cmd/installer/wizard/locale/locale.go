package locale

// Common status and UI strings
const (
	// Common status and UI strings
	UIStatistics       = "Statistics"
	UIStatus           = "Status: "
	UIMode             = "Mode: "
	UINoConfigSelected = "No configuration selected"
	UILoading          = "Loading..."
	UINotImplemented   = "Not implemented yet"
	UIUnsavedChanges   = "Unsaved changes"
	UIConfigSaved      = "Configuration saved"

	// Status labels
	StatusEnabled       = "Enabled"
	StatusDisabled      = "Disabled"
	StatusConfigured    = "Configured"
	StatusNotConfigured = "Not configured"
	StatusEmbedded      = "Embedded"
	StatusExternal      = "External"

	// Success/Warning messages
	MessageSearchEnginesNone       = "⚠ No search engines configured"
	MessageSearchEnginesConfigured = "✓ %d search engines configured"
	MessageDockerConfigured        = "✓ Docker environment configured"
	MessageDockerNotConfigured     = "⚠ Docker environment not configured"
)

// Legend constants
const (
	LegendConfigured    = "✓ Configured"
	LegendNotConfigured = "✗ Not configured"
)

// Common Navigation Actions (always available)
const (
	NavBack       = "Esc: Back"
	NavExit       = "Ctrl+Q: Exit"
	NavUpDown     = "↑/↓: Scroll/Select"
	NavLeftRight  = "←/→: Move"
	NavPgUpPgDown = "PgUp/PgDn: Page"
	NavHomeEnd    = "Home/End: Start/End"
	NavEnter      = "Enter: Continue"
	NavYn         = "Y/N: Accept/Reject"
	NavCtrlC      = "Ctrl+C: Cancel"
	NavCtrlS      = "Ctrl+S: Save"
	NavCtrlR      = "Ctrl+R: Reset"
	NavCtrlH      = "Ctrl+H: Show/Hide"
	NavTab        = "Tab: Complete"
	NavSeparator  = " • "
)

// Welcome Screen constants
const (
	// Form interface implementation
	WelcomeFormTitle       = "Welcome to PentAGI"
	WelcomeFormDescription = "PentAGI is an autonomous penetration testing platform that leverages AI technologies to perform comprehensive security assessments."
	WelcomeFormName        = "Welcome"
	WelcomeFormOverview    = `System checks verify:
• Environment configuration file presence
• Docker API accessibility and version compatibility
• Worker environment readiness
• System resources (CPU, memory, disk space)
• Network connectivity for external dependencies

Once all checks pass, proceed through the configuration wizard to set up LLM providers, monitoring, and security tools.

The installer guides you through each component setup with recommendations for different deployment scenarios.`

	// Configuration status messages
	WelcomeConfigurationFailed = "⚠ Failed checks: %s"
	WelcomeConfigurationPassed = "✓ All system checks passed"

	// Workflow steps
	WelcomeWorkflowTitle = "Installation Workflow:"
	WelcomeWorkflowStep1 = "1. Accept End User License Agreement"
	WelcomeWorkflowStep2 = "2. Configure LLM providers (OpenAI, Anthropic, etc.)"
	WelcomeWorkflowStep3 = "3. Set up integrations (Langfuse, Observability)"
	WelcomeWorkflowStep4 = "4. Configure security settings"
	WelcomeWorkflowStep5 = "5. Deploy and start PentAGI services"
	WelcomeSystemReady   = "✓ System ready - Press Enter to continue"
)

// Troubleshooting on welcome screen constants
const (
	TroubleshootTitle = "System Requirements Not Met"

	// Environment file issues
	TroubleshootEnvFileTitle = "Environment Configuration Missing"
	TroubleshootEnvFileDesc  = "The .env file is required for PentAGI configuration but was not found or is not readable."
	TroubleshootEnvFileFix   = `To fix:
1. Copy .env.example to .env in your installation directory
2. Edit .env and configure at least one LLM provider API key
3. Ensure the file has read permissions (chmod 644 .env)

Quick fix:
cp .env.example .env && chmod 644 .env`

	// Write permissions
	TroubleshootWritePermTitle = "Write Permissions Required"
	TroubleshootWritePermDesc  = "The installer needs write access to the configuration directory to save settings and deploy services."
	TroubleshootWritePermFix   = `To fix:
1. Check directory permissions: ls -la
2. Grant write access: chmod 755 .
3. Or run installer from a writable location
4. Ensure sufficient disk space is available`

	// Docker not installed
	TroubleshootDockerNotInstalledTitle = "Docker Not Installed"
	TroubleshootDockerNotInstalledDesc  = "Docker is not installed on this system. PentAGI requires Docker to run containers."
	TroubleshootDockerNotInstalledFix   = `To fix:
1. Install Docker Desktop: https://docs.docker.com/get-docker/
2. For Linux: Follow distribution-specific instructions
3. Verify installation: docker --version
4. Ensure docker command is in your PATH`

	// Docker not running
	TroubleshootDockerNotRunningTitle = "Docker Daemon Not Running"
	TroubleshootDockerNotRunningDesc  = "Docker is installed but the daemon is not running. The Docker service must be active."
	TroubleshootDockerNotRunningFix   = `To fix:
1. Start Docker Desktop (Windows/Mac)
2. Linux: sudo systemctl start docker
3. Check status: docker ps
4. If using DOCKER_HOST, verify the remote daemon is accessible`

	// Docker permission issues
	TroubleshootDockerPermissionTitle = "Docker Permission Denied"
	TroubleshootDockerPermissionDesc  = "Your user account lacks permission to access Docker. This is common on Linux systems."
	TroubleshootDockerPermissionFix   = `To fix:
1. Add user to docker group: sudo usermod -aG docker $USER
2. Log out and back in for changes to take effect
3. Or run with sudo (not recommended for production)
4. Verify: docker ps (should work without sudo)`

	// Generic Docker API issues
	TroubleshootDockerAPITitle = "Docker API Connection Failed"
	TroubleshootDockerAPIDesc  = "Cannot establish connection to Docker API. This may be due to configuration or network issues."
	TroubleshootDockerAPIFix   = `To fix:
1. Check DOCKER_HOST environment variable
2. Verify Docker is running: docker version
3. For remote Docker: ensure network connectivity
4. Check firewall settings if using TCP connection
5. Try: export DOCKER_HOST=unix:///var/run/docker.sock`

	// Docker version issues
	TroubleshootDockerVersionTitle = "Docker Version Too Old"
	TroubleshootDockerVersionDesc  = "Your Docker version is incompatible. PentAGI requires Docker 20.0.0 or newer."
	TroubleshootDockerVersionFix   = `To fix:
1. Update Docker to version 20.0.0 or newer
2. Visit https://docs.docker.com/engine/install/

Current version: %s
Required: 20.0.0+`

	// Docker Compose issues
	TroubleshootComposeTitle = "Docker Compose Not Found"
	TroubleshootComposeDesc  = "Docker Compose is required but not installed or not in PATH."
	TroubleshootComposeFix   = `To fix:
1. Install Docker Desktop (includes Compose) or
2. Install standalone: https://docs.docker.com/compose/install/

Verify installation: docker compose version`

	// Docker Compose version issues
	TroubleshootComposeVersionTitle = "Docker Compose Version Too Old"
	TroubleshootComposeVersionDesc  = "Your Docker Compose version is incompatible. PentAGI requires Docker Compose 1.25.0 or newer."
	TroubleshootComposeVersionFix   = `Current version: %s
Required: 1.25.0+

To fix:
1. Update Docker Desktop to latest version
2. Or install newer Docker Compose:
   https://docs.docker.com/compose/install/`

	// Worker environment issues
	TroubleshootWorkerTitle = "Worker Docker Environment Not Accessible"
	TroubleshootWorkerDesc  = "Cannot connect to the Docker environment for worker containers. This may be a remote or local Docker setup issue."
	TroubleshootWorkerFix   = `To fix:
1. For remote Docker, set env vars before installer:
   export DOCKER_HOST=tcp://remote:2376
   export DOCKER_CERT_PATH=/path/to/certs
   export DOCKER_TLS_VERIFY=1
2. Verify connection: docker -H $DOCKER_HOST ps
3. For local Docker, leave these vars unset
4. Check firewall allows Docker port (2375/2376)
5. Ensure certificates are valid if using TLS`

	// CPU issues
	TroubleshootCPUTitle = "Insufficient CPU Cores"
	TroubleshootCPUDesc  = "PentAGI requires at least 2 CPU cores for proper operation."
	TroubleshootCPUFix   = `Your system has %d CPU core(s), but 2+ are required.

For virtual machines:
1. Increase CPU allocation in VM settings
2. Ensure host has sufficient resources

Docker Desktop users:
Settings → Resources → CPUs: Set to 2 or more`

	// Memory issues
	TroubleshootMemoryTitle = "Insufficient Memory"
	TroubleshootMemoryDesc  = "Not enough free memory for selected components."
	TroubleshootMemoryFix   = `Memory requirements:
• Base system: 0.5 GB
• PentAGI core: +0.5 GB
• Langfuse (if enabled): +1.5 GB
• Observability (if enabled): +1.5 GB

Total needed: %.1f GB
Available: %.1f GB

To fix:
1. Close unnecessary applications
2. Increase Docker memory limit
3. Disable optional components (Langfuse/Observability)`

	// Disk space issues
	TroubleshootDiskTitle = "Insufficient Disk Space"
	TroubleshootDiskDesc  = "Not enough free disk space for installation and operation."
	TroubleshootDiskFix   = `Disk requirements:
• Base installation: 5 GB minimum
• With components: 10 GB + 2 GB per component
• Worker images: 25 GB (includes 6GB+ Kali image)

Required: %.1f GB
Available: %.1f GB

To fix:
1. Free up disk space
2. Use external storage for Docker
3. Prune unused Docker resources:
   docker system prune -a`

	// Network issues
	TroubleshootNetworkTitle = "Network Connectivity Failed"
	TroubleshootNetworkDesc  = "Cannot reach required external services. This prevents downloading Docker images and updates."
	TroubleshootNetworkFix   = `Failed checks:
%s

To fix:
1. Verify internet connection: ping docker.io
2. Check DNS resolution: nslookup docker.io
3. If behind proxy, set before running installer:
   export HTTP_PROXY=http://proxy:port
   export HTTPS_PROXY=http://proxy:port
4. For persistent proxy, add to .env:
   PROXY_URL=http://proxy:port
5. Check firewall allows outbound HTTPS (port 443)
6. Try alternative DNS servers if DNS fails`

	// Generic hint at the bottom
	TroubleshootFixHint = "\nResolve the issues above and run the installer again."

	// Network failure messages (used in checker/helpers.go)
	NetworkFailureDNS        = "• DNS resolution failed for docker.io"
	NetworkFailureHTTPS      = "• Cannot reach external services via HTTPS"
	NetworkFailureDockerPull = "• Cannot pull Docker images from registry"
)

// System Checks constants
const (
	ChecksTitle               = "System Checks"
	ChecksWarningFailed       = "⚠ Some checks failed"
	CheckEnvironmentFile      = "Environment file"
	CheckWritePermissions     = "Write permissions"
	CheckDockerAPI            = "Docker API"
	CheckDockerVersion        = "Docker version"
	CheckDockerCompose        = "Docker Compose"
	CheckDockerComposeVersion = "Docker Compose version"
	CheckWorkerEnvironment    = "Worker environment"
	CheckSystemResources      = "System resources"
	CheckNetworkConnectivity  = "Network connectivity"
)

// EULA Screen constants
const (
	// Form interface implementation
	EULAFormDescription = "Legal terms and conditions for PentAGI usage"
	EULAFormName        = "EULA"
	EULAFormOverview    = `Review and accept the End User License Agreement to proceed with PentAGI installation.

The EULA contains:
• Software license terms and usage rights
• Limitation of liability and warranties
• Data collection and privacy policies
• Compliance requirements and restrictions
• Support and maintenance terms

You must scroll through the entire document and accept the terms to continue with the installation process.

Use arrow keys, page up/down, or home/end keys to navigate through the document.`

	// Error and status messages
	EULAErrorLoadingTitle     = "# Error Loading EULA\n\nFailed to load EULA: %v"
	EULAContentFallback       = "# EULA Content\n\n%s\n\n---\n\n*Note: Markdown rendering failed: %v*"
	EULAConfigurationRead     = "✓ EULA reviewed"
	EULAConfigurationAccepted = "✓ EULA accepted"
	EULAConfigurationPending  = "⚠ EULA not reviewed"
	EULALoading               = "Loading EULA..."
	EULAProgress              = "Progress: %d%%"
	EULAProgressComplete      = " • Complete"
)

// Main Menu Screen constants
const (
	MainMenuTitle       = "PentAGI Configuration"
	MainMenuDescription = "Configure all PentAGI components and settings"
	MainMenuName        = "Main Menu"
	MainMenuOverview    = `Welcome to PentAGI Configuration Center.

Configure essential components:
• LLM Providers - AI language models for autonomous testing
• Monitoring - Observability and analytics platforms
• Tools - Additional capabilities for enhanced testing
• System Settings - Environment and deployment options

Navigate through each section to complete your PentAGI setup.`

	MenuTitle        = "Configuration Menu"
	MenuSystemStatus = "System Status"
)

// Main Menu Status Labels (not used)
const (
	MainMenuStatusPentagiRunning     = "PentAGI is already running"
	MainMenuStatusPentagiNotRunning  = "Ready to start PentAGI services"
	MainMenuStatusUpToDate           = "PentAGI is up to date"
	MainMenuStatusUpdatesAvailable   = "Updates are available"
	MainMenuStatusReadyToStart       = "Ready to start"
	MainMenuStatusAllServicesRunning = "All services are running"
	MainMenuStatusNoUpdatesAvailable = "No updates available"
)

// LLM Providers Screen constants
const (
	LLMProvidersTitle       = "LLM Providers Configuration"
	LLMProvidersDescription = "Configure Large Language Model providers for AI agents"
	LLMProvidersName        = "LLM Providers"
	LLMProvidersOverview    = `PentAGI uses specialized AI agents (researcher, developer, executor, pentester) that require different LLM capabilities for optimal penetration testing results.

Why multiple providers matter:
• Agent Specialization: Different agents benefit from models optimized for reasoning, coding, or analysis
• Cost Efficiency: Mix expensive reasoning models (o3, grok-4, claude-sonnet-4, gemini-2.5-pro) for complex tasks with cheaper models for simple operations
• Performance Optimization: Each provider excels in different areas - OpenAI for medium tasks, Anthropic for complex tasks, Gemini for saving costs

Provider Selection Guide:
• Cloud Production: OpenAI + Anthropic + Gemini for industry-leading performance and reliability
• Enterprise/Compliance: AWS Bedrock for SOC2, HIPAA, and access to multiple model families
• Privacy/On-premises: Ollama or vLLM with Llama 3.1, Qwen3, or other open models for complete data control

Ready-to-use configurations for OpenRouter, DeepInfra, vLLM, Ollama, and other providers are available in the /opt/pentagi/conf/ directory inside the container`
)

// LLM Provider titles and descriptions
const (
	LLMProviderOpenAI        = "OpenAI"
	LLMProviderAnthropic     = "Anthropic"
	LLMProviderGemini        = "Google Gemini"
	LLMProviderBedrock       = "AWS Bedrock"
	LLMProviderOllama        = "Ollama"
	LLMProviderCustom        = "Custom"
	LLMProviderOpenAIDesc    = "Industry-leading GPT models with excellent general performance"
	LLMProviderAnthropicDesc = "Claude models with superior reasoning and safety features"
	LLMProviderGeminiDesc    = "Google's advanced multimodal models with broad knowledge"
	LLMProviderBedrockDesc   = "Enterprise AWS access to multiple foundation model providers"
	LLMProviderOllamaDesc    = "Local open-source models for privacy and cost control"
	LLMProviderCustomDesc    = "Custom OpenAI-compatible endpoint for maximum flexibility"
)

// Provider-specific help text
const (
	LLMFormOpenAIHelp = `OpenAI delivers industry-leading models with cutting-edge reasoning capabilities perfect for sophisticated penetration testing.

Default PentAGI Models:
• o1, o4-mini: Advanced reasoning models for complex vulnerability analysis and strategic planning
• GPT-4.1, GPT-4.1-mini: Flagship models optimized for exploit development and code generation
• Automatic model selection based on agent type and task complexity

Key Advantages:
• Most advanced reasoning capabilities with step-by-step analysis (o-series models)
• Excellent coding abilities for custom exploit development and payload generation
• Reliable performance with consistent uptime and extensive API documentation
• Proven track record in security research and penetration testing scenarios

Best for: Production environments requiring cutting-edge AI capabilities, teams prioritizing performance over cost
Cost: Premium pricing, but optimized configurations balance cost with quality

Setup: Get your API key from https://platform.openai.com/api-keys`

	LLMFormAnthropicHelp = `Anthropic Claude models excel in safety-conscious penetration testing with superior reasoning and analytical capabilities.

Default PentAGI Models:
• Claude Sonnet-4: Premium reasoning model for complex security analysis and strategic vulnerability assessment
• Claude 3.5 Haiku: High-speed model optimized for rapid information gathering and simple parsing tasks
• Balanced cost-performance ratio across all security testing scenarios

Key Advantages:
• Exceptional safety and ethics focus - reduces harmful output while maintaining security testing effectiveness
• Superior reasoning for methodical vulnerability analysis and systematic penetration testing approaches
• Large context windows ideal for analyzing extensive codebases and configuration files
• Excellent at understanding complex security contexts and regulatory compliance requirements

Best for: Security teams prioritizing responsible testing practices, compliance-focused environments, detailed analysis
Cost: Mid-range pricing with excellent value for reasoning-heavy security workflows

Setup: Get your API key from https://console.anthropic.com/`

	LLMFormGeminiHelp = `Google Gemini combines multimodal capabilities with advanced reasoning, perfect for comprehensive security assessments.

Default PentAGI Models:
• Gemini 2.5 Pro: Advanced reasoning model for deep vulnerability analysis and complex exploit development
• Gemini 2.5 Flash: High-performance model balancing speed and intelligence for most security testing tasks
• Gemini 2.0 Flash Lite: Cost-effective model for rapid scanning and information gathering operations
• Reasoning capabilities with step-by-step analysis for thorough penetration testing

Key Advantages:
• Multimodal support enables analysis of screenshots, network diagrams, and security documentation
• Competitive pricing with generous rate limits for development and testing environments
• Large context windows (up to 2M tokens) for analyzing massive codebases and system configurations
• Strong performance in code analysis and vulnerability identification across multiple programming languages

Best for: Budget-conscious teams, development environments, scenarios requiring image/document analysis
Cost: Most cost-effective option among major cloud providers with excellent performance/price ratio

Setup: Get your API key from https://aistudio.google.com/app/apikey`

	LLMFormBedrockHelp = `AWS Bedrock provides enterprise-grade access to multiple foundation model families with enhanced security and compliance.

Default PentAGI Models:
• Claude Sonnet-4 (via Bedrock): Premium reasoning model with AWS enterprise security controls
• Claude 3.5 Haiku (via Bedrock): High-speed model for rapid security analysis with SOC2/HIPAA compliance
• Access to additional models: Amazon Nova series, Meta Llama, Cohere Command, and more through single interface

Key Advantages:
• Enterprise compliance: SOC2, HIPAA, FedRAMP certifications with data residency controls
• Multi-provider access: Single interface to models from Anthropic, Amazon, Meta, AI21, Cohere, and DeepSeek
• Enhanced security: VPC integration, CloudTrail logging, IAM controls for enterprise governance
• Regional deployment: Choose AWS regions for latency optimization and data sovereignty requirements

Best for: Enterprise environments, regulated industries, teams requiring compliance and governance controls
Cost: Competitive pricing with reserved capacity options, but may have restrictive default rate limits
Warning: New AWS accounts have very low rate limits (2 requests/minute) - request quota increases for production use

Setup: Configure AWS credentials and region. Check rate limits at https://docs.aws.amazon.com/bedrock/`

	LLMFormOllamaHelp = `Ollama enables completely private, on-premises LLM deployment with zero ongoing costs and full data control.

Default PentAGI Models:
• Llama 3.1:8b: Primary model for all agent types providing balanced performance and resource efficiency
• Customizable: Easy model switching between Qwen, Mistral, CodeLlama, Gemma, and 100+ other open models
• Local inference: No internet dependency once models are downloaded

Key Advantages:
• Complete data privacy: All processing happens locally, no data sent to external services
• Zero ongoing costs: Only infrastructure costs, no per-token or per-request charges
• Customizable: Load any compatible model, fine-tune for specific security testing scenarios
• Air-gapped deployment: Perfect for isolated networks and highly secure environments
• Resource efficient: Optimized for local GPUs with quantized models for smaller hardware

Best for: Privacy-focused teams, air-gapped networks, cost-sensitive deployments, organizations with data sovereignty requirements
Cost: Only infrastructure costs - powerful for teams with dedicated hardware
Requirement: Local compute resources (8GB+ RAM recommended, GPU optional but beneficial)

Setup: Install Ollama from https://ollama.ai/ or use container deployments. See examples/configs for optimized configurations`

	LLMFormCustomHelp = `Configure any OpenAI-compatible API endpoint for maximum flexibility and integration with existing infrastructure.

Ready-to-use Configurations:
• vLLM deployments: High-throughput on-premises inference with optimal GPU utilization
• OpenRouter: Access 200+ models from multiple providers through single API with competitive pricing
• DeepInfra: Serverless inference for popular open models with pay-per-use pricing
• Together AI, Groq, Fireworks: Alternative cloud providers with specialized performance optimizations
• LiteLLM Proxy: Universal gateway to 100+ providers with load balancing and unified interface (use LLM_SERVER_PROVIDER for model prefixing)
• Some reasoning models and LLM providers may require preserving reasoning content while using tool calls (LLM_SERVER_PRESERVE_REASONING=true)

Popular On-Premises Options:
• vLLM: Production-grade serving for Qwen, Llama, Mistral models with batching and GPU optimization
• LocalAI: OpenAI-compatible API wrapper for various local models and embedding services
• Text Generation WebUI: Community-favorite interface with extensive model support and fine-tuning capabilities
• Hugging Face TGI: Enterprise text generation inference with auto-scaling and monitoring

Key Advantages:
• Unlimited flexibility: Use any OpenAI-compatible endpoint or service
• Cost optimization: Choose providers with competitive pricing or deploy models on your own infrastructure
• Vendor independence: Avoid lock-in with ability to switch between providers and models seamlessly
• Custom fine-tuning: Deploy specialized models trained on your security testing scenarios

Best for: Teams with specific model requirements, cost optimization needs, or existing LLM infrastructure
LiteLLM Integration: Set LLM_SERVER_PROVIDER to match your provider name (e.g., "openrouter", "moonshot") to use the same config files with both direct API access and LiteLLM proxy
Examples available: Pre-configured setups for major providers in /opt/pentagi/conf/ directory inside the container`
)

// LLM Provider Form field labels and descriptions
const (
	LLMFormFieldBaseURL           = "Base URL"
	LLMFormFieldAPIKey            = "API Key"
	LLMFormFieldAccessKey         = "Access Key ID"
	LLMFormFieldSecretKey         = "Secret Access Key"
	LLMFormFieldSessionToken      = "Session Token"
	LLMFormFieldRegion            = "Region"
	LLMFormFieldModel             = "Model"
	LLMFormFieldConfigPath        = "Config Path"
	LLMFormFieldLegacyReasoning   = "Legacy Reasoning"
	LLMFormFieldPreserveReasoning = "Preserve Reasoning"
	LLMFormFieldProviderName      = "Provider Name"
	LLMFormFieldPullTimeout       = "Model Pull Timeout"
	LLMFormFieldPullEnabled       = "Auto-pull Models"
	LLMFormFieldLoadModelsEnabled = "Load Models from Server"
	LLMFormBaseURLDesc            = "API endpoint URL for the provider"
	LLMFormAPIKeyDesc             = "Your API key for authentication"
	LLMFormAccessKeyDesc          = "AWS Access Key ID for Bedrock access"
	LLMFormSecretKeyDesc          = "AWS Secret Access Key for Bedrock access"
	LLMFormSessionTokenDesc       = "AWS Session Token for Bedrock access"
	LLMFormRegionDesc             = "AWS region for Bedrock service"
	LLMFormModelDesc              = "Default model to use for this provider"
	LLMFormConfigPathDesc         = "Path to configuration file (optional)"
	LLMFormLegacyReasoningDesc    = "Enable legacy reasoning mode (true/false)"
	LLMFormPreserveReasoningDesc  = "Preserve reasoning content in multi-turn conversations (required by some providers)"
	LLMFormProviderNameDesc       = "Provider name prefix for model names (useful for LiteLLM proxy)"
	LLMFormPullTimeoutDesc        = "Timeout in seconds for downloading models (default: 600)"
	LLMFormPullEnabledDesc        = "Automatically download required models on startup"
	LLMFormLoadModelsEnabledDesc  = "Load available models list from Ollama server"
)

// LLM Provider Form status messages
const (
	LLMProviderFormTitle       = "LLM Provider %s Configuration"
	LLMProviderFormDescription = "Configure your Large Language Model provider settings"
	LLMProviderFormName        = "LLM Provider %s"
	LLMProviderFormOverview    = `Agent Role Assignment:
• Primary Agent & Pentester: Use reasoning models (o3, grok-4, claude-sonnet-4, gemini-2.5-pro) for complex vulnerability analysis
• Assistant & Adviser: Advanced models (o4-mini, claude-sonnet-4) for strategic planning and recommendations
• Coder & Installer: Precision models (gpt-4.1, claude-sonnet-4) for exploit development and system configuration
• Searcher & Enricher: Fast models (gpt-4.1-mini, claude-3.5-haiku, gemini-2.0-flash-lite) for information gathering
• Simple tasks: Lightweight models for JSON parsing and basic operations

Performance Considerations:
• Reasoning models provide step-by-step analysis but are slower and more expensive
• Standard models offer faster responses suitable for high-frequency agent interactions
• Each agent type uses provider-specific model configurations optimized for security testing workflows

Your configuration will determine which models each agent uses for different penetration testing scenarios.`
)

// Monitoring Screen
const (
	MonitoringTitle       = "Monitoring Configuration"
	MonitoringDescription = "Configure monitoring and observability platforms for comprehensive system insights"
	MonitoringName        = "Monitoring"
	MonitoringOverview    = `Comprehensive monitoring and observability for production-ready deployments.

Why monitoring matters:
• Track performance bottlenecks: Identify slow LLM calls, database queries, and system resources
• Debug issues faster: Detailed traces help diagnose problems across distributed components
• Optimize costs: Monitor token usage patterns and optimize expensive LLM interactions
• Production readiness: Essential for reliable operation in critical environments

Platform Options:
Langfuse: Specialized LLM observability with conversation tracking, prompt engineering insights, and cost analytics
Observability: Full-stack monitoring with metrics, traces, logs, and alerting for infrastructure and application health

Quick Setup:
• Development: Enable Langfuse for LLM insights only
• Production: Enable both platforms for comprehensive monitoring
• Cost-conscious: Use embedded modes to avoid external service fees`
)

// Langfuse Integration constants
const (
	MonitoringLangfuseFormTitle       = "Langfuse Configuration"
	MonitoringLangfuseFormDescription = "Configuration of Langfuse integration for LLM monitoring"
	MonitoringLangfuseFormName        = "Langfuse"
	MonitoringLangfuseFormOverview    = `Langfuse provides:
• Complete conversation tracking
• Model performance metrics
• Cost monitoring and optimization
• User behavior analytics
• Debug traces for AI interactions

Choose between embedded instance or external connection.`

	// Deployment types
	MonitoringLangfuseEmbedded = "Embedded Server"
	MonitoringLangfuseExternal = "External Server"
	MonitoringLangfuseDisabled = "Disabled"

	// Form fields
	MonitoringLangfuseDeploymentType     = "Deployment Type"
	MonitoringLangfuseDeploymentTypeDesc = "Select the deployment type for Langfuse"
	MonitoringLangfuseBaseURL            = "Server URL"
	MonitoringLangfuseBaseURLDesc        = "Address of the Langfuse server (e.g., https://cloud.langfuse.com)"
	MonitoringLangfuseProjectID          = "Project ID"
	MonitoringLangfuseProjectIDDesc      = "Project identifier in Langfuse"
	MonitoringLangfusePublicKey          = "Public Key"
	MonitoringLangfusePublicKeyDesc      = "Public API key for project access"
	MonitoringLangfuseSecretKey          = "Secret Key"
	MonitoringLangfuseSecretKeyDesc      = "Secret API key for project access"
	MonitoringLangfuseListenIP           = "Listen IP"
	MonitoringLangfuseListenIPDesc       = "Bind address used by Docker port mapping (e.g., 0.0.0.0 to expose on all interfaces)"
	MonitoringLangfuseListenPort         = "Listen Port"
	MonitoringLangfuseListenPortDesc     = "External TCP port exposed by Docker for Langfuse web UI"

	// Admin settings for embedded
	MonitoringLangfuseAdminEmail        = "Admin Email"
	MonitoringLangfuseAdminEmailDesc    = "Email for accessing the Langfuse admin panel"
	MonitoringLangfuseAdminPassword     = "Admin Password"
	MonitoringLangfuseAdminPasswordDesc = "Password for accessing the Langfuse admin panel"
	MonitoringLangfuseAdminName         = "Admin Username"
	MonitoringLangfuseAdminNameDesc     = "Administrator username in Langfuse"
	MonitoringLangfuseLicenseKey        = "Enterprise License Key"
	MonitoringLangfuseLicenseKeyDesc    = "Langfuse Enterprise license key (optional)"

	// Help text
	MonitoringLangfuseModeGuide    = "Choose deployment: Embedded (local control), External (cloud/existing), Disabled (no analytics)"
	MonitoringLangfuseEmbeddedHelp = `Embedded deploys complete Langfuse stack:
• PostgreSQL + ClickHouse databases
• MinIO S3 storage + Redis cache
• Full LLM conversation tracking
• Cost analysis and performance metrics
• Private data stays on your server

Resource requirements:
• ~2GB RAM, 5GB disk space minimum
• Additional storage for conversation logs
• Automatic setup and maintenance

Best for: Teams wanting data privacy, custom configurations, or no external dependencies. All analytics data stored locally with full administrative control.

Default admin access:
• Web UI: http://localhost:4000
• Login: admin@pentagi.com
• Password: password (change required)`
	MonitoringLangfuseExternalHelp = `External connects to cloud.langfuse.com or your existing Langfuse server:

• No local infrastructure needed
• Managed updates and maintenance
• Shared analytics across teams
• Enterprise features available
• Data stored on external provider

Setup requirements:
• Langfuse account and API keys
• Internet connectivity required
• Project ID and authentication keys

Best for: Teams using cloud services, wanting managed infrastructure, or integrating with existing Langfuse deployments across organizations.`
	MonitoringLangfuseDisabledHelp = `Langfuse is disabled. Without LLM observability you will not have:

• Conversation history tracking
• Token usage and cost analysis
• Model performance metrics
• Debug traces for AI interactions
• User behavior analytics
• Prompt engineering insights

Consider enabling for production use
to monitor AI agent performance and
optimize costs effectively.`
)

// Graphiti Integration constants
const (
	MonitoringGraphitiFormTitle       = "Graphiti Configuration (beta)"
	MonitoringGraphitiFormDescription = "Configuration of Graphiti knowledge graph integration"
	MonitoringGraphitiFormName        = "Graphiti (beta)"
	MonitoringGraphitiFormOverview    = `⚠️  BETA FEATURE: This functionality is currently under active development. Please monitor updates for improvements and stability fixes.

Graphiti provides temporal knowledge graph capabilities:
• Entity and relationship extraction
• Semantic memory for AI agents
• Temporal context tracking
• Knowledge reuse across flows

⚠️  REQUIREMENT: Graphiti requires configured OpenAI provider (LLM Providers → OpenAI) for entity extraction.

Choose between embedded instance or external connection.`

	// Deployment types
	MonitoringGraphitiEmbedded = "Embedded Stack"
	MonitoringGraphitiExternal = "External Service"
	MonitoringGraphitiDisabled = "Disabled"

	// Form fields
	MonitoringGraphitiDeploymentType     = "Deployment Type"
	MonitoringGraphitiDeploymentTypeDesc = "Select the deployment type for Graphiti"
	MonitoringGraphitiURL                = "Graphiti Server URL"
	MonitoringGraphitiURLDesc            = "Address of the Graphiti API server"
	MonitoringGraphitiTimeout            = "Request Timeout"
	MonitoringGraphitiTimeoutDesc        = "Timeout in seconds for Graphiti operations"
	MonitoringGraphitiModelName          = "Extraction Model"
	MonitoringGraphitiModelNameDesc      = "LLM model for entity extraction (uses OpenAI provider from LLM Providers configuration)"
	MonitoringGraphitiNeo4jUser          = "Neo4j Username"
	MonitoringGraphitiNeo4jUserDesc      = "Username for Neo4j database access"
	MonitoringGraphitiNeo4jPassword      = "Neo4j Password"
	MonitoringGraphitiNeo4jPasswordDesc  = "Password for Neo4j database access"
	MonitoringGraphitiNeo4jDatabase      = "Neo4j Database"
	MonitoringGraphitiNeo4jDatabaseDesc  = "Neo4j database name"

	// Help text
	MonitoringGraphitiModeGuide    = "Choose deployment: Embedded (local Neo4j), External (existing Graphiti), Disabled (no knowledge graph)"
	MonitoringGraphitiEmbeddedHelp = `⚠️  BETA: This feature is under active development. Monitor updates for improvements.

Embedded deploys complete Graphiti stack:
• Neo4j graph database
• Graphiti API service
• Automatic entity extraction from agent interactions
• Temporal relationship tracking
• Private knowledge graph on your server

Prerequisites:
• OpenAI provider must be configured (LLM Providers → OpenAI)
• OpenAI API key is used for entity extraction
• Configured model will be used for knowledge graph operations

Resource requirements:
• ~1.5GB RAM, 3GB disk space minimum
• Neo4j UI: http://localhost:7474
• Graphiti API: http://localhost:8000
• Automatic setup and maintenance

Best for: Teams wanting knowledge graph capabilities with full data control and privacy.`
	MonitoringGraphitiExternalHelp = `⚠️  BETA: This feature is under active development. Monitor updates for improvements.

External connects to your existing Graphiti server:

• No local infrastructure needed
• Managed updates and maintenance
• Shared knowledge graph across teams
• Data stored on external provider

Setup requirements:
• Graphiti server URL and access
• Network connectivity required
• External server must be configured with OpenAI API key
• Model and extraction settings configured on external server

Best for: Teams using existing Graphiti deployments or cloud services.`
	MonitoringGraphitiDisabledHelp = `Graphiti is disabled. You will not have:

• Temporal knowledge graph
• Entity and relationship extraction
• Semantic memory for AI agents
• Knowledge reuse across flows
• Advanced contextual search

Note: Graphiti is currently in beta.
Consider enabling for production use
to build a knowledge base from
penetration testing results.`
)

// Observability Integration constants
const (
	MonitoringObservabilityFormTitle       = "Observability Configuration"
	MonitoringObservabilityFormDescription = "Configuration of monitoring and observability stack"
	MonitoringObservabilityFormName        = "Observability"
	MonitoringObservabilityFormOverview    = `Observability stack includes:
• Grafana dashboards for visualization
• VictoriaMetrics for time-series data
• Jaeger for distributed tracing
• Loki for log aggregation
• OpenTelemetry for data collection

Monitor PentAGI performance and system health.`

	// Deployment types
	MonitoringObservabilityEmbedded = "Embedded Stack"
	MonitoringObservabilityExternal = "External Collector"
	MonitoringObservabilityDisabled = "Disabled"

	// Form fields
	MonitoringObservabilityDeploymentType     = "Deployment Type"
	MonitoringObservabilityDeploymentTypeDesc = "Select the deployment type for monitoring"
	MonitoringObservabilityOTelHost           = "OpenTelemetry Host"
	MonitoringObservabilityOTelHostDesc       = "Address of the external OpenTelemetry collector"

	// embedded listen fields
	MonitoringObservabilityGrafanaListenIP        = "Grafana Listen IP"
	MonitoringObservabilityGrafanaListenIPDesc    = "Bind address used by Docker port mapping (e.g., 0.0.0.0 to expose on all interfaces)"
	MonitoringObservabilityGrafanaListenPort      = "Grafana Listen Port"
	MonitoringObservabilityGrafanaListenPortDesc  = "External TCP port exposed by Docker for Grafana web UI"
	MonitoringObservabilityOTelGrpcListenIP       = "OTel gRPC Listen IP"
	MonitoringObservabilityOTelGrpcListenIPDesc   = "Bind address used by Docker port mapping (e.g., 0.0.0.0 to expose on all interfaces)"
	MonitoringObservabilityOTelGrpcListenPort     = "OTel gRPC Listen Port"
	MonitoringObservabilityOTelGrpcListenPortDesc = "External TCP port exposed by Docker for OTel gRPC receiver"
	MonitoringObservabilityOTelHttpListenIP       = "OTel HTTP Listen IP"
	MonitoringObservabilityOTelHttpListenIPDesc   = "Bind address used by Docker port mapping (e.g., 0.0.0.0 to expose on all interfaces)"
	MonitoringObservabilityOTelHttpListenPort     = "OTel HTTP Listen Port"
	MonitoringObservabilityOTelHttpListenPortDesc = "External TCP port exposed by Docker for OTel HTTP receiver"

	// Help text
	MonitoringObservabilityModeGuide    = "Choose monitoring: Embedded (full stack), External (existing infra), Disabled (no monitoring)"
	MonitoringObservabilityEmbeddedHelp = `Embedded deploys complete monitoring:
• Grafana dashboards and alerting
• VictoriaMetrics time-series database
• Jaeger distributed tracing UI
• Loki log aggregation system
• ClickHouse analytical database
• Node Exporter + cAdvisor metrics
• OpenTelemetry data collection

Auto-instrumented components with
pre-built dashboards for system health,
performance analysis, and debugging.

Resource requirements:
• ~1.5GB RAM, 3GB disk space minimum
• Grafana UI: http://localhost:3000
• Profiling: http://localhost:7777

Best for: Complete system visibility,
troubleshooting, and performance tuning.`
	MonitoringObservabilityExternalHelp = `External sends telemetry to your existing monitoring infrastructure:

• OTLP protocol over HTTP/2 (no TLS)
• Your collector must support:
  - OTLP HTTP receiver (port 4318)
  - OTLP gRPC receiver (port 8148)
  - tls: insecure: true setting
• Sends metrics, traces, and logs
• Compatible with enterprise platforms:
  Datadog, New Relic, Splunk, etc.

OTEL_HOST example:
your-collector:4318

Collector config requirement:
tls: insecure: true

Best for: Organizations with existing
monitoring infrastructure or centralized
observability platforms.`
	MonitoringObservabilityDisabledHelp = `Observability is disabled. You will not have:

• System performance monitoring
• Distributed request tracing
• Structured log aggregation
• Resource usage analytics
• Error tracking and alerting
• Performance bottleneck analysis

Consider enabling for production use
to monitor system health, debug issues,
and optimize performance effectively.`
)

// Summarizer Screen
const (
	SummarizerTitle       = "Summarizer Configuration"
	SummarizerDescription = "Enable conversation summarization to reduce LLM costs and improve context management"
	SummarizerName        = "Summarizer"
	SummarizerOverview    = `Optimize context usage, reduce LLM costs, and match your model capabilities.

When to adjust summarization:
• High token costs: Reduce context size (4K-12K vs 22K+ tokens)
• "Context too long" errors: Configure for your model's limits
• Poor conversation flow: Increase context retention for quality
• Different model types: Short-context vs long-context model tuning

General Summarization: Maximum cost control and precision tuning for research/analysis tasks
Assistant Summarization: Optimal conversation quality with intelligent context management for interactive sessions

Quick wins:
• Cost reduction: Use General, reduce Recent Sections to 1-2
• Context errors: Match limits to your model (8K/32K/128K)
• Quality priority: Use Assistant with increased limits`

	SummarizerTypeGeneralName = "General Summarization"
	SummarizerTypeGeneralDesc = "Global summarization settings for conversation context management"

	SummarizerTypeGeneralInfo = `Choose this for maximum cost control and short-context model compatibility.

Perfect when you need:
• Aggressive cost reduction: Fine-tune every parameter for minimal token usage
• Short-context models (8K-32K): Precise limits to avoid overflow errors
• Research/analysis tasks: Controlled compression without losing key data
• Custom QA handling: Full control over question-answer pair processing

Typical results:
• 40-70% cost reduction vs default settings
• 4K-12K token contexts (vs 22K+ in Assistant mode)
• Better performance on GPT-3.5, Claude Instant, smaller models
• Precise control over conversation memory vs fresh context balance

Best practices:
• Start with 1-2 Recent Sections for maximum savings
• Enable Size Management for automatic overflow protection
• Disable QA compression only for critical reasoning tasks`

	SummarizerTypeAssistantName = "Assistant Summarization"
	SummarizerTypeAssistantDesc = "Specialized summarization settings for AI assistant contexts"

	SummarizerTypeAssistantInfo = `Choose this for optimal conversation quality and dialogue continuity.

Perfect when you need:
• Extended reasoning chains: Maintain context for complex multi-step thinking
• High-quality conversations: Preserve dialogue flow and assistant personality
• Long-context models (64K+): Leverage full model capabilities efficiently
• Interactive sessions: Better memory of user preferences and conversation history

Typical results:
• 8K-40K token contexts with intelligent compression
• Superior conversation continuity vs manual settings
• Automatic context optimization for reasoning tasks
• Balanced cost vs quality (3x more context than General mode)

Best practices:
• Use default settings for most scenarios - they're pre-optimized
• Increase Recent Sections only for very complex tasks
• Monitor context usage - costs scale with token count
• Perfect for GPT-4, Claude, and other large context models`
)

// Summarizer Form Screen
const (
	SummarizerFormGeneralTitle   = "General Summarizer Configuration"
	SummarizerFormAssistantTitle = "Assistant Summarizer Configuration"
	SummarizerFormDescription    = "Configure %s Settings"

	// Field Labels and Descriptions
	SummarizerFormPreserveLast     = "Size Management"
	SummarizerFormPreserveLastDesc = "Controls last section compression. Enabled: sections fit LastSecBytes (smaller context). Disabled: sections grow freely (larger context)"

	SummarizerFormUseQA     = "QA Summarization"
	SummarizerFormUseQADesc = "Enables question-answer pair compression when total QA content exceeds MaxQABytes or MaxQASections limits"

	SummarizerFormSumHumanInQA     = "Compress User Messages"
	SummarizerFormSumHumanInQADesc = "Include user messages in QA compression. Disabled: preserves original user text (recommended for most cases)"

	SummarizerFormLastSecBytes     = "Section Size Limit"
	SummarizerFormLastSecBytesDesc = "Maximum bytes per recent section when Size Management enabled. Larger: more detail per section, higher token usage"

	SummarizerFormMaxBPBytes     = "Response Size Limit"
	SummarizerFormMaxBPBytesDesc = "Maximum bytes for individual AI responses before compression. Prevents single large responses from dominating context"

	SummarizerFormMaxQASections     = "QA Section Limit"
	SummarizerFormMaxQASectionsDesc = "Maximum question-answer sections before QA compression triggers. Works with MaxQABytes to control total QA memory"

	SummarizerFormMaxQABytes     = "Total QA Memory"
	SummarizerFormMaxQABytesDesc = "Maximum bytes for all QA sections combined. When exceeded (with MaxQASections), triggers QA compression to fit limit"

	SummarizerFormKeepQASections     = "Recent Sections"
	SummarizerFormKeepQASectionsDesc = "Number of most recent conversation sections preserved without compression. PRIMARY parameter affecting context size"

	// Enhanced Help Text - General (common principles)
	SummarizerFormGeneralHelp = `Context estimation: 4K-22K tokens (typical), up to 94K (maximum settings).

Key relationships:
• Recent Sections: Most critical - each +1 adds ~1.5-9K tokens
• Size Management OFF: 2-3x larger context (less compression)
• Section/Response Limits: Control individual component sizes
• QA Memory: Manages total conversation history when limits exceeded

Parameter interactions:
• QA compression activates when BOTH MaxQABytes AND MaxQASections exceeded
• Size Management disabled → sections can grow 2x larger than limits
• Response Limit prevents single large outputs from dominating context
• User message compression (SummHumanInQA) saves 5% but loses original phrasing

Reduce for smaller models:
• Recent Sections: 1-2 (vs 3+ default)
• Section Limit: 25-35KB (vs 50KB+)
• Disable Size Management for simple conversations

Common mistakes:
• Setting Recent Sections too high (main cause of context overflow)
• Enabling Size Management with very low Section Limits (over-compression)
• Mismatched QA limits (high bytes + low sections = ineffective)

Current algorithm compresses older content while preserving recent context quality.`

	// Enhanced Help Text - Assistant specific (interactive conversations)
	SummarizerFormAssistantHelp = `Optimized for interactive conversations requiring context continuity.

Default tuning (3 Recent Sections, 75KB limits):
• Typical range: 8K-40K tokens
• Good for: Extended dialogues, reasoning chains, context-dependent tasks
• Models: Works well with 32K+ context models

Adjustments by model type:
• Short context (≤16K): Recent Sections=1-2, Section Limit=45KB
• Long context (128K+): Can increase Recent Sections=5-7
• High-frequency chat: Reduce Recent Sections=2 for faster responses

Advanced tuning:
• QA Memory 200KB+ for document analysis conversations
• Response Limit 24-32KB for detailed technical responses
• Keep User Messages uncompressed (SummHumanInQA=false) for better context

Performance optimization:
• Each Recent Section ≈ 9-18KB in assistant mode
• Size Management reduces growth by ~20% but may lose detail
• QA compression triggers less often due to larger default limits

Size Management enabled by default - maintains conversation flow while preventing context overflow.
Monitor actual token usage and adjust Recent Sections first, then limits.`

	// Context size estimation
	SummarizerContextEstimatedSize    = "Estimated context size: %s\n%s"
	SummarizerContextTokenRange       = "~%s tokens"
	SummarizerContextTokenRangeMinMax = "~%s-%s tokens"
	SummarizerContextRequires256K     = "Requires 256K+ context model"
	SummarizerContextRequires128K     = "Requires 128K+ context model"
	SummarizerContextRequires64K      = "Requires 64K+ context model"
	SummarizerContextRequires32K      = "Requires 32K+ context model"
	SummarizerContextRequires16K      = "Requires 16K+ context model"
	SummarizerContextFitsIn8K         = "Fits in 8K+ context model"
)

// Tools screen strings
const (
	ToolsTitle       = "Tools Configuration"
	ToolsDescription = "Enhance agent capabilities with additional tools and options"
	ToolsName        = "Tools"
	ToolsOverview    = `Configure additional tools and capabilities for AI agents.
Each tool can be enabled and configured according to your requirements.

Available settings:
• Human-in-the-loop - Enable user interaction during testing
• AI Agents Settings - Configure global behavior for AI agents
• Search Engines - Configure external search providers
• Scraper - Web content extraction and analysis
• Graphiti (beta) - Temporal knowledge graph for semantic memory
• Docker - Container environment configuration`
)

// Server Settings screen strings
const (
	ServerSettingsFormTitle       = "Server Settings"
	ServerSettingsFormDescription = "Configure PentAGI server network access and public routing"
	ServerSettingsFormName        = "Server Settings"
	ServerSettingsFormOverview    = `• Network binding - control which interface and port PentAGI listens on
• Public URL - external address and optional base path used in redirects
• CORS - allowed origins for browser access
• Proxy - HTTP/HTTPS proxy for outbound traffic to LLM/search providers
• SSL directory - custom certificates directory containing server.crt and server.key (PEM)
• Data directory - persistent storage for agent artifacts and flow workspaces`

	// Field labels and descriptions
	ServerSettingsLicenseKey     = "License Key"
	ServerSettingsLicenseKeyDesc = "PentAGI License Key in format of XXXX-XXXX-XXXX-XXXX"

	ServerSettingsHost     = "Server Host (Listen IP)"
	ServerSettingsHostDesc = "Bind address used by Docker port mapping (e.g., 0.0.0.0 to expose on all interfaces)"

	ServerSettingsPort     = "Server Port (Listen Port)"
	ServerSettingsPortDesc = "External TCP port exposed by Docker for PentAGI web UI"

	ServerSettingsPublicURL     = "Public URL"
	ServerSettingsPublicURLDesc = "Base public URL for redirects and links (supports base path, e.g., https://example.com/pentagi/)"

	ServerSettingsCORSOrigins     = "CORS Origins"
	ServerSettingsCORSOriginsDesc = "Comma-separated list of allowed origins (e.g., https://localhost:8443,https://localhost)"

	ServerSettingsProxyURL     = "HTTP/HTTPS Proxy"
	ServerSettingsProxyURLDesc = "Proxy for outbound requests to LLMs and external tools (not used for Docker API access)"

	ServerSettingsProxyUsername     = "Proxy Username"
	ServerSettingsProxyUsernameDesc = "Username for proxy authentication (optional)"
	ServerSettingsProxyPassword     = "Proxy Password"
	ServerSettingsProxyPasswordDesc = "Password for proxy authentication (optional)"

	ServerSettingsExternalSSLCAPath     = "Custom CA Certificate Path"
	ServerSettingsExternalSSLCAPathDesc = "Path inside container to custom root CA cert (e.g., /opt/pentagi/ssl/ca-bundle.pem)"

	ServerSettingsExternalSSLInsecure     = "Skip SSL Verification"
	ServerSettingsExternalSSLInsecureDesc = "Disable SSL/TLS certificate validation (use only for testing with self-signed certs)"

	ServerSettingsSSLDir     = "SSL Directory"
	ServerSettingsSSLDirDesc = "Directory containing server.crt and server.key in PEM format (server.crt may include fullchain)"

	ServerSettingsDataDir     = "Data Directory"
	ServerSettingsDataDirDesc = "Directory for all agent-generated files; contains flow-N subdirectories used as /work in worker containers"

	ServerSettingsCookieSigningSalt     = "Cookie Signing Salt"
	ServerSettingsCookieSigningSaltDesc = "Secret used to sign cookies (keep private)"

	// Hints for fields overview
	ServerSettingsLicenseKeyHint          = "License Key"
	ServerSettingsHostHint                = "Listen IP"
	ServerSettingsPortHint                = "Listen Port"
	ServerSettingsPublicURLHint           = "Public URL"
	ServerSettingsCORSOriginsHint         = "CORS Origins"
	ServerSettingsProxyURLHint            = "Proxy URL"
	ServerSettingsProxyUsernameHint       = "Proxy Username"
	ServerSettingsProxyPasswordHint       = "Proxy Password"
	ServerSettingsExternalSSLCAPathHint   = "Custom CA Path"
	ServerSettingsExternalSSLInsecureHint = "Skip SSL Verification"
	ServerSettingsSSLDirHint              = "SSL Directory"
	ServerSettingsDataDirHint             = "Data Directory"

	// Help texts per-field
	ServerSettingsGeneralHelp = `PentAGI exposes its web UI via Docker with configurable host and port.

Public URL must reflect how users reach the server. If using a subpath (e.g., /pentagi/), include it here. CORS controls browser access from specified origins. Proxy affects outbound traffic to LLM/search providers and other external services used by Tools.

SSL directory allows providing custom certificates. When set, server will use server.crt and server.key from that directory. Data directory stores artifacts and working files for flows.`

	ServerSettingsLicenseKeyHelp = `PentAGI License Key in format of XXXX-XXXX-XXXX-XXXX. It's used to communicate with PentAGI Cloud API.`

	ServerSettingsHostHelp = `Bind address for published port in docker-compose mapping.

Examples:
• 127.0.0.1 — local-only access
• 0.0.0.0 — expose on all interfaces`

	ServerSettingsPortHelp = `External port for PentAGI UI. Must be available on the host. Example: 8443.`

	ServerSettingsPublicURLHelp = `Set the public base URL used in redirects and links.

Examples:
• http://localhost:8443
• https://example.com/
• https://example.com/pentagi/ (with base path)`

	ServerSettingsCORSOriginsHelp = `Comma-separated allowed origins for browser access.`

	ServerSettingsProxyURLHelp = `HTTP or HTTPS proxy for outbound requests to LLM providers and external tools. Not used for Docker API communication.`

	ServerSettingsExternalSSLCAPathHelp = `Path to custom CA certificate file (PEM format) inside the container.

Must point to /opt/pentagi/ssl/ directory, which is mounted from pentagi-ssl volume on the host.

Examples:
• /opt/pentagi/ssl/ca-bundle.pem
• /opt/pentagi/ssl/corporate-ca.pem

File can contain multiple root and intermediate certificates.`

	ServerSettingsExternalSSLInsecureHelp = `Disable SSL/TLS certificate validation for connections to LLM providers and external services.

⚠ WARNING: Use only for testing with self-signed certificates. Never enable in production.

When enabled, all certificate validation is bypassed, making connections vulnerable to man-in-the-middle attacks.`

	ServerSettingsSSLDirHelp = `Path to directory with server.crt and server.key in PEM format. server.crt may include fullchain. Overrides default generated certificate behavior.`

	ServerSettingsDataDirHelp = `Host directory for persistent data. PentAGI stores agent artifacts under flow-N subdirectories, which map to /work inside worker containers.`

	ServerSettingsCookieSigningSaltHelp = `Secret salt used to sign cookies. Keep it private.`
)

// Human-in-the-loop screen strings
const (
	// AI Agents Settings screen strings
	ToolsAIAgentsSettingsFormTitle       = "AI Agents Settings"
	ToolsAIAgentsSettingsFormDescription = "Configure global behavior for AI agents"
	ToolsAIAgentsSettingsFormName        = "AI Agents Settings"
	ToolsAIAgentsSettingsFormOverview    = `This section configures global behavior of AI agents across PentAGI.

Current options:
• Enable User Interaction (Human-in-the-loop): allow agents to request user input when needed
• Use Multi-Agent Mode: enable assistant to orchestrate multiple specialized agents when appropriate

These settings affect how AI agents operate system-wide.`

	// field labels and descriptions
	ToolsAIAgentsSettingHumanInTheLoop     = "Enable User Interaction"
	ToolsAIAgentsSettingHumanInTheLoopDesc = "Allow agents to ask for user input when needed"
	ToolsAIAgentsSettingUseAgents          = "Use Multi-Agent Mode"
	ToolsAIAgentsSettingUseAgentsDesc      = "Enable assistant to orchestrate multiple specialized agents"

	// help content
	ToolsAIAgentsSettingsHelp = `AI Agents Settings define how agents collaborate and interact with users.

• Enable User Interaction to allow safe prompts for approval and clarifications.
• Use Multi-Agent Mode to let the assistant spawn domain-specific agents (researcher, coder, pentester) for better results on complex tasks.

Note: These are global settings for control AI agents behavior.`
)

// Search Engines screen strings
const (
	ToolsSearchEnginesFormTitle       = "Search Engines Configuration"
	ToolsSearchEnginesFormDescription = "Configure search engines for AI agents to gather intelligence during testing"
	ToolsSearchEnginesFormName        = "Search Engines"
	ToolsSearchEnginesFormOverview    = `Available search engines:
• DuckDuckGo - Free search engine (no API key required)
• Sploitus - Security exploits and vulnerabilities database (no API key required)
• Perplexity - AI-powered search with reasoning
• Tavily - Search API for AI applications
• Traversaal - Web scraping and search
• Google Search - Requires API key and Custom Search Engine ID
• Searxng - Internet metasearch engine

Get API keys from:
• Perplexity: https://www.perplexity.ai/
• Tavily: https://tavily.com/
• Traversaal: https://traversaal.ai/
• Google: https://developers.google.com/custom-search/v1/introduction`

	ToolsSearchEnginesDuckDuckGo            = "DuckDuckGo Search"
	ToolsSearchEnginesDuckDuckGoDesc        = "Enable DuckDuckGo search (no API key required)"
	ToolsSearchEnginesSploitus              = "Sploitus Search"
	ToolsSearchEnginesSploitusDesc          = "Enable Sploitus search for exploits and vulnerabilities (no API key required)"
	ToolsSearchEnginesPerplexityKey         = "Perplexity API Key"
	ToolsSearchEnginesPerplexityKeyDesc     = "API key for Perplexity AI search"
	ToolsSearchEnginesTavilyKey             = "Tavily API Key"
	ToolsSearchEnginesTavilyKeyDesc         = "API key for Tavily search service"
	ToolsSearchEnginesTraversaalKey         = "Traversaal API Key"
	ToolsSearchEnginesTraversaalKeyDesc     = "API key for Traversaal web scraping"
	ToolsSearchEnginesGoogleKey             = "Google Search API Key"
	ToolsSearchEnginesGoogleKeyDesc         = "Google Custom Search API key"
	ToolsSearchEnginesGoogleCX              = "Google Search Engine ID"
	ToolsSearchEnginesGoogleCXDesc          = "Google Custom Search Engine ID"
	ToolsSearchEnginesGoogleLR              = "Google Language Restriction"
	ToolsSearchEnginesGoogleLRDesc          = "Google Search Engine	language restriction (e.g., lang_en, lang_cn, etc.)"
	ToolsSearchEnginesSearxngURL            = "Searxng Search URL"
	ToolsSearchEnginesSearxngURLDesc        = "Searxng search engine URL"
	ToolsSearchEnginesSearxngCategories     = "Searxng Search Categories"
	ToolsSearchEnginesSearxngCategoriesDesc = "Searxng search engine categories (e.g., general, it, web, news, technology, science, health, other)"
	ToolsSearchEnginesSearxngLanguage       = "Searxng Search Language"
	ToolsSearchEnginesSearxngLanguageDesc   = "Searxng search engine language (en, ch, fr, de, it, es, pt, ru, zh, empty for all languages)"
	ToolsSearchEnginesSearxngSafeSearch     = "Searxng Safe Search"
	ToolsSearchEnginesSearxngSafeSearchDesc = "Searxng search engine safe search (0: off, 1: moderate, 2: strict)"
	ToolsSearchEnginesSearxngTimeRange      = "Searxng Time Range"
	ToolsSearchEnginesSearxngTimeRangeDesc  = "Searxng search engine time range (day, month, year)"
)

// Scraper screen strings
const (
	ToolsScraperFormTitle       = "Scraper Configuration"
	ToolsScraperFormDescription = "Configure web scraping service"
	ToolsScraperFormName        = "Scraper"
	ToolsScraperFormOverview    = `Web scraper service for content extraction and analysis using vxcontrol/scraper Docker image.

Modes:
• Embedded - Run local scraper container (recommended)
• External - Use external scraper services
• Disabled - No web scraping capabilities

Docker image: https://hub.docker.com/r/vxcontrol/scraper

The scraper supports:
• Public URL access for external links
• Private URL access for internal/local links
• Content extraction and analysis
• Multiple output formats`

	ToolsScraperModeTitle                 = "Scraper Mode"
	ToolsScraperModeDesc                  = "Select how the scraper service should operate"
	ToolsScraperEmbedded                  = "Embedded Container"
	ToolsScraperExternal                  = "External Service"
	ToolsScraperDisabled                  = "Disabled"
	ToolsScraperPublicURL                 = "Public Scraper URL"
	ToolsScraperPublicURLDesc             = "URL for scraping public/external websites. If empty, the same value as private URL will be used."
	ToolsScraperPublicURLEmbeddedDesc     = "URL for embedded scraper (optional override). If empty, the same value as private URL will be used."
	ToolsScraperPrivateURL                = "Private Scraper URL"
	ToolsScraperPrivateURLDesc            = "URL for scraping private/internal websites"
	ToolsScraperPublicUsername            = "Public URL Username"
	ToolsScraperPublicUsernameDesc        = "Username for public scraper access"
	ToolsScraperPublicPassword            = "Public URL Password"
	ToolsScraperPublicPasswordDesc        = "Password for public scraper access"
	ToolsScraperPrivateUsername           = "Private URL Username"
	ToolsScraperPrivateUsernameDesc       = "Username for private scraper access"
	ToolsScraperPrivatePassword           = "Private URL Password"
	ToolsScraperPrivatePasswordDesc       = "Password for private scraper access"
	ToolsScraperLocalUsername             = "Local URL Username"
	ToolsScraperLocalUsernameDesc         = "Username for embedded scraper service"
	ToolsScraperLocalPassword             = "Local URL Password"
	ToolsScraperLocalPasswordDesc         = "Password for embedded scraper service"
	ToolsScraperMaxConcurrentSessions     = "Max Concurrent Sessions"
	ToolsScraperMaxConcurrentSessionsDesc = "Maximum number of concurrent scraping sessions"
	ToolsScraperEmbeddedHelp              = "Embedded mode runs a local scraper container that can access both public and private resources. The default configuration uses https://someuser:somepass@scraper/."
	ToolsScraperExternalHelp              = "External mode uses separate scraper services. Configure different URLs for public and private access as needed."
	ToolsScraperDisabledHelp              = "Scraper is disabled. Web content extraction and analysis capabilities will not be available."
)

// Docker Environment screen strings
const (
	ToolsDockerFormTitle       = "Docker Environment Configuration"
	ToolsDockerFormDescription = "Configure Docker environment for worker containers"
	ToolsDockerFormName        = "Docker Environment"
	ToolsDockerFormOverview    = `• Worker Isolation - Containers provide security boundaries for tasks
• Network Capabilities - Enable privileged network operations for pentesting
• Container Management - Control how workers access Docker daemon
• Storage Configuration - Define workspace and artifact storage
• Image Selection - Set default images for different task types

Critical for penetration testing workflows requiring network scanning, custom tools, and secure task isolation.`

	// General help text
	ToolsDockerGeneralHelp = `Each AI agent task runs in an isolated Docker container with two ports (28000-32000 range) automatically allocated per flow. Worker containers are created on-demand from default images or agent-selected ones.

Basic setup requires enabling capabilities: Docker Access allows spawning additional containers for specialized tools, while Network Admin grants low-level network permissions essential for scanning tools like nmap.

Storage operates via Docker volumes by default, or host directories when Work Directory is specified. Connection settings control the Docker daemon location - local socket for standard setups, or remote TCP with TLS for distributed environments.

Default images serve as fallbacks: general tasks use standard images, while security testing defaults to pentesting-focused containers. Public IP enables reverse shell attacks by providing workers with a reachable address for target callbacks. Usually it's a local interface address of the host machine with Docker daemon running for the workers containers.

Configuration combines based on scenario: enable both capabilities for full pentesting, use Work Directory for persistent artifacts, or configure remote connection for isolated Docker environments.`

	// Container capabilities
	ToolsDockerInside       = "Docker Access"
	ToolsDockerInsideDesc   = "Allow workers to manage Docker containers"
	ToolsDockerNetAdmin     = "Network Admin"
	ToolsDockerNetAdminDesc = "Grant NET_ADMIN capability for network scanning tools like nmap"

	// Connection settings
	ToolsDockerSocket       = "Docker Socket"
	ToolsDockerSocketDesc   = "Path to Docker socket on host filesystem"
	ToolsDockerNetwork      = "Docker Network"
	ToolsDockerNetworkDesc  = "Custom network name for worker containers"
	ToolsDockerPublicIP     = "Public IP Address"
	ToolsDockerPublicIPDesc = "Public IP for reverse connections in OOB attacks"

	// Storage configuration
	ToolsDockerWorkDir     = "Work Directory"
	ToolsDockerWorkDirDesc = "Host directory for worker filesystems (default: Docker volumes)"

	// Default images
	ToolsDockerDefaultImage               = "Default Image"
	ToolsDockerDefaultImageDesc           = "Default Docker image for general tasks"
	ToolsDockerDefaultImageForPentest     = "Pentesting Image"
	ToolsDockerDefaultImageForPentestDesc = "Default Docker image for security testing tasks"

	// TLS connection settings (optional)
	ToolsDockerHost          = "Docker Host"
	ToolsDockerHostDesc      = "Docker daemon connection (unix:// or tcp://)"
	ToolsDockerTLSVerify     = "TLS Verification"
	ToolsDockerTLSVerifyDesc = "Enable TLS verification for Docker connection"
	ToolsDockerCertPath      = "TLS Certificates"
	ToolsDockerCertPathDesc  = "Directory containing ca.pem, cert.pem, key.pem files"

	// Help content for specific configurations
	ToolsDockerInsideHelp = `Docker Access enables workers to spawn additional containers for specialized tools and environments. Required when tasks need custom software not available in default images.

When enabled, workers can pull and run any Docker image, providing maximum flexibility for complex testing scenarios.`

	ToolsDockerNetAdminHelp = `Network Admin capability allows workers to perform low-level network operations essential for penetration testing.

Required for:
• Network scanning with nmap, masscan
• Custom packet crafting
• Network interface manipulation
• Raw socket operations

Critical for comprehensive security assessments.`

	ToolsDockerSocketHelp = `Docker Socket path defines how workers access the Docker daemon. Use only file path to the socket file. Used with Docker Access to enable container management.

For enhanced security, consider using docker-in-docker (DinD) instead of exposing the main Docker daemon directly to workers.
When using DinD, use the path to the Docker socket file of the DinD container which binded to the host filesystem.

Example: /var/run/docker.sock`

	ToolsDockerNetworkHelp = `Custom Docker Network provides isolation for worker containers. Allows fine-grained firewall rules and network policies.

Useful for:
• Isolating worker traffic
• Custom network configurations
• Enhanced security boundaries
• Network-based monitoring`

	ToolsDockerPublicIPHelp = `Public IP Address enables out-of-band (OOB) attack techniques by providing workers with a reachable address for reverse connections.

Workers automatically receive two random ports (28000-32000 range) mapped to this IP for receiving callbacks from exploited targets.

By default agents will try to get public address from the services api.ipify.org, ipinfo.io/ip or ifconfig.me.`

	ToolsDockerWorkDirHelp = `Work Directory specifies host filesystem location for worker storage. When set, replaces default Docker volumes with host directory mounts.

Benefits:
• Persistent storage across restarts
• Direct file system access
• Easier artifact management
• Custom backup strategies

By default uses Docker dedicated volume per worker container.

Example: /path/to/workdir/`

	ToolsDockerDefaultImageHelp = `Default Image provides fallback for workers when task requirements don't specify a particular container image.

Should contain basic utilities and tools for general-purpose tasks. Default: debian:latest`

	ToolsDockerDefaultImageForPentestHelp = `Pentesting Image serves as default for security testing tasks. Should include comprehensive security tools and utilities.

Recommended images include Kali Linux, Parrot Security, or custom security-focused containers. Default: vxcontrol/kali-linux`

	ToolsDockerHostHelp = `Docker Host uses for start primary worker containers and overrides default Docker daemon connection. Supports Unix sockets and TCP connections.

Examples:
• unix:///var/run/docker.sock (local)
• tcp://docker-host:2376 (remote)

Enable TLS for remote connections.`

	ToolsDockerTLSVerifyHelp = `TLS Verification secures Docker daemon connections over TCP. Strongly recommended for remote Docker hosts.

Requires valid certificates in the specified certificate directory.`

	ToolsDockerCertPathHelp = `TLS Certificates directory must contain:
• ca.pem - Certificate Authority
• cert.pem - Client certificate
• key.pem - Private key

Required for secure remote Docker connections when using TLS to manage worker containers.

Example: /path/to/certs`
)

// Embedder form strings
const (
	EmbedderFormTitle       = "Embedder Configuration"
	EmbedderFormDescription = "Configure text vectorization for semantic search and knowledge storage"
	EmbedderFormName        = "Embedder"
	EmbedderFormOverview    = `Text embeddings convert documents into vectors for semantic search and knowledge storage.
Different providers offer various models with different capabilities and pricing.

Choose carefully as changing providers requires reindexing all stored data.`

	EmbedderFormProvider     = "Embedding Provider"
	EmbedderFormProviderDesc = "Select the provider for text vectorization. Embeddings are used for semantic search and knowledge storage."

	EmbedderFormURL     = "API Endpoint URL"
	EmbedderFormURLDesc = "Custom API endpoint (leave empty to use default)"

	EmbedderFormAPIKey     = "API Key"
	EmbedderFormAPIKeyDesc = "Authentication key for the provider (not required for Ollama)"

	EmbedderFormModel     = "Model Name"
	EmbedderFormModelDesc = "Specific embedding model to use (leave empty for provider default)"

	EmbedderFormBatchSize     = "Batch Size"
	EmbedderFormBatchSizeDesc = "Number of documents to process in a single batch (1-1000)"

	EmbedderFormStripNewLines     = "Strip New Lines"
	EmbedderFormStripNewLinesDesc = "Remove line breaks from text before embedding (true/false)"

	EmbedderFormHelpTitle   = "Embedding Configuration"
	EmbedderFormHelpContent = `Configure text vectorization for semantic search and knowledge storage.

If no specific embedding settings are configured, the system will use OpenAI embeddings with the API key from LLM Providers.

Change providers carefully - different embedders produce incompatible vectors requiring database reindexing.`

	EmbedderFormHelpOpenAI      = "OpenAI: Most reliable option with excellent quality. Requires API key from LLM Providers if not set here."
	EmbedderFormHelpOllama      = "Ollama: Local embeddings, no API key needed. Requires Ollama server running."
	EmbedderFormHelpHuggingFace = "HuggingFace: Open source models with API key required."
	EmbedderFormHelpGoogleAI    = "Google AI: Quality embeddings, requires API key."

	// Provider names and descriptions
	EmbedderProviderDefault         = "Default (OpenAI)"
	EmbedderProviderDefaultDesc     = "Use OpenAI embeddings with API key from LLM Providers configuration"
	EmbedderProviderOpenAI          = "OpenAI"
	EmbedderProviderOpenAIDesc      = "OpenAI text embeddings API (text-embedding-3-small, ada-002)"
	EmbedderProviderOllama          = "Ollama"
	EmbedderProviderOllamaDesc      = "Local Ollama server for open-source embedding models"
	EmbedderProviderMistral         = "Mistral"
	EmbedderProviderMistralDesc     = "Mistral AI embedding models"
	EmbedderProviderJina            = "Jina"
	EmbedderProviderJinaDesc        = "Jina AI embedding API"
	EmbedderProviderHuggingFace     = "HuggingFace"
	EmbedderProviderHuggingFaceDesc = "HuggingFace inference API for embedding models"
	EmbedderProviderGoogleAI        = "Google AI"
	EmbedderProviderGoogleAIDesc    = "Google AI embedding models (embedding-001)"
	EmbedderProviderVoyageAI        = "VoyageAI"
	EmbedderProviderVoyageAIDesc    = "VoyageAI embedding API"
	EmbedderProviderDisabled        = "Disabled"
	EmbedderProviderDisabledDesc    = "Disable embeddings functionality completely"

	// Provider-specific placeholders and help
	EmbedderURLPlaceholderOpenAI      = "https://api.openai.com/v1"
	EmbedderURLPlaceholderOllama      = "http://localhost:11434"
	EmbedderURLPlaceholderMistral     = "https://api.mistral.ai/v1"
	EmbedderURLPlaceholderJina        = "https://api.jina.ai/v1"
	EmbedderURLPlaceholderHuggingFace = "https://api-inference.huggingface.co"
	EmbedderURLPlaceholderGoogleAI    = "Not supported - uses default endpoint"
	EmbedderURLPlaceholderVoyageAI    = "Not supported - uses default endpoint"

	EmbedderAPIKeyPlaceholderOllama      = "Not required for local models"
	EmbedderAPIKeyPlaceholderMistral     = "Mistral API key"
	EmbedderAPIKeyPlaceholderJina        = "Jina API key"
	EmbedderAPIKeyPlaceholderHuggingFace = "HuggingFace API key"
	EmbedderAPIKeyPlaceholderGoogleAI    = "Google AI API key"
	EmbedderAPIKeyPlaceholderVoyageAI    = "VoyageAI API key"
	EmbedderAPIKeyPlaceholderDefault     = "API key for the provider"

	EmbedderModelPlaceholderOpenAI      = "text-embedding-3-small"
	EmbedderModelPlaceholderOllama      = "nomic-embed-text"
	EmbedderModelPlaceholderMistral     = "mistral-embed"
	EmbedderModelPlaceholderJina        = "jina-embeddings-v2-base-en"
	EmbedderModelPlaceholderHuggingFace = "sentence-transformers/all-MiniLM-L6-v2"
	EmbedderModelPlaceholderGoogleAI    = "gemini-embedding-001"
	EmbedderModelPlaceholderVoyageAI    = "voyage-2"
	EmbedderModelPlaceholderDefault     = "Model name"

	// Provider IDs for internal use
	EmbedderProviderIDDefault     = "default"
	EmbedderProviderIDOpenAI      = "openai"
	EmbedderProviderIDOllama      = "ollama"
	EmbedderProviderIDMistral     = "mistral"
	EmbedderProviderIDJina        = "jina"
	EmbedderProviderIDHuggingFace = "huggingface"
	EmbedderProviderIDGoogleAI    = "googleai"
	EmbedderProviderIDVoyageAI    = "voyageai"
	EmbedderProviderIDDisabled    = "none"

	EmbedderHelpGeneral = `Embeddings convert text into vectors for semantic search and knowledge storage. This enables PentAGI to understand meaning rather than just keywords, making search results more relevant and intelligent.

Key benefits:
• Find documents by meaning, not exact words
• Build a smart knowledge base from pentesting results
• Enable AI agents to locate relevant information quickly
• Support advanced reasoning with contextual data

Choose Ollama for completely local processing - your data never leaves your infrastructure. Other providers offer cloud-based processing with different model capabilities and pricing.

Configure carefully as changing providers requires rebuilding the entire knowledge base.`

	EmbedderHelpAttentionPrefix = "Important:"
	EmbedderHelpAttention       = `Different embedding providers create incompatible vectors. Changing providers or models will break existing semantic search.

You must flush or reindex your entire knowledge base using the etester utility:
• Run 'etester flush' to clear old embeddings
• Run 'etester reindex' to rebuild with new provider
• This process can take significant time for large datasets`

	EmbedderHelpAttentionSuffix = `Only change providers if absolutely necessary.`

	// Provider help texts
	EmbedderHelpDefault = `Default mode uses OpenAI embeddings with the API key configured in LLM Providers.

This is the recommended option for most users as it requires no additional configuration if you already have OpenAI set up.`

	EmbedderHelpOpenAI = `Direct OpenAI API access for embedding generation.

Get your API key from:
https://platform.openai.com/api-keys

Recommended models:
• text-embedding-3-small (cost-effective, 1536 dimensions)
• text-embedding-3-large (highest quality, 3072 dimensions)
• text-embedding-ada-002 (legacy, still supported)`

	EmbedderHelpOllama = `Local Ollama server for open-source embedding models.

Popular embedding models:
• nomic-embed-text (recommended, 768 dimensions)
• mxbai-embed-large (large model, 1024 dimensions)
• snowflake-arctic-embed (multilingual support)

Install Ollama from:
https://ollama.com/

Start with: ollama pull nomic-embed-text`

	EmbedderHelpMistral = `Mistral AI embedding models via API.

Get your API key from:
https://console.mistral.ai/

Uses Mistral's embedding model with fixed configuration.
No model selection required - uses the default embedding model.`

	EmbedderHelpJina = `Jina AI embedding API with specialized models.

Get your API key from:
https://jina.ai/

Recommended models:
• jina-embeddings-v2-base-en (general purpose, 768 dimensions)
• jina-embeddings-v2-small-en (lightweight, 512 dimensions)
• jina-embeddings-v2-base-code (code-specific embeddings)`

	EmbedderHelpHuggingFace = `HuggingFace Inference API for open-source models.

Get your API key from:
https://huggingface.co/settings/tokens

Popular models:
• sentence-transformers/all-MiniLM-L6-v2 (384 dimensions)
• sentence-transformers/all-mpnet-base-v2 (768 dimensions)
• intfloat/e5-large-v2 (1024 dimensions)`

	EmbedderHelpGoogleAI = `Google AI embedding models (Gemini).

Get your API key from:
https://aistudio.google.com/app/apikey

Available models:
• gemini-embedding-001 (latest model, 768 dimensions)
• text-embedding-004 (legacy Vertex AI model)

Uses Google's fixed endpoint - URL configuration not supported.`

	EmbedderHelpVoyageAI = `VoyageAI embedding API optimized for retrieval.

Get your API key from:
https://www.voyageai.com/

Recommended models:
• voyage-2 (general purpose, 1024 dimensions)
• voyage-large-2 (highest quality, 1536 dimensions)
• voyage-code-2 (code embeddings, 1536 dimensions)`

	EmbedderHelpDisabled = `Disables all embedding functionality.

This will:
• Disable semantic search capabilities
• Turn off knowledge storage vectorization
• Reduce memory and computational requirements

Only recommended if embeddings are not needed for your use case.`
)

// Development and Mock Screen constants
const (
	MockScreenTitle       = "Development Screen"
	MockScreenDescription = "This screen is under development"
)

// Apply Changes screen constants
const (
	ApplyChangesFormTitle       = "Apply Configuration Changes"
	ApplyChangesFormName        = "Apply Changes"
	ApplyChangesFormDescription = "Review and apply your configuration changes"

	// Apply Changes overview and help
	ApplyChangesFormOverview = `This screen allows you to review all pending configuration changes and apply them to your PentAGI installation.

When you apply changes, the system will:
• Save all modified environment variables to the .env file
• Restart affected services with the new configuration
• Install additional components if needed`

	// Apply Changes status messages
	ApplyChangesNotStarted     = "Configuration changes are ready to be applied"
	ApplyChangesInProgress     = "Applying configuration changes...\n"
	ApplyChangesCompleted      = "Configuration changes have been successfully applied\n"
	ApplyChangesFailed         = "Failed to perform configuration changes"
	ApplyChangesResetCompleted = "Configuration changes have been successfully reset\n"

	ApplyChangesTerminalIsNotInitialized = "Terminal is not initialized"

	// Apply Changes instructions
	ApplyChangesInstructions = `Press Enter to begin applying the configuration changes.`

	ApplyChangesNoChanges = "No configuration changes are pending"

	// Apply Changes installation status
	ApplyChangesInstallNotFound = `PentAGI is not currently installed on this system.

The following actions will be performed:
• Docker environment setup and validation
• Creation of docker-compose.yml file
• Installation and startup of PentAGI core services`

	ApplyChangesInstallFoundLangfuse      = `• Installation of Langfuse observability stack (docker-compose-langfuse.yml)`
	ApplyChangesInstallFoundObservability = `• Installation of comprehensive observability stack with Grafana, VictoriaMetrics, and Jaeger (docker-compose-observability.yml)`

	ApplyChangesUpdateFound = `PentAGI is currently installed on this system.

The following actions will be performed:
• Update environment variables in .env file
• Recreate and restart affected Docker containers
• Apply new configuration to running services`

	// Apply Changes warnings and notes
	ApplyChangesWarningCritical = "⚠️  Critical changes detected - services will be restarted"
	ApplyChangesWarningSecrets  = "🔒 Secret values detected - they will be securely stored"
	ApplyChangesNoteBackup      = "💾 Current configuration will be backed up before changes"
	ApplyChangesNoteTime        = "⏱️  This process may take less than a minute depending on selected components"

	// Apply Changes progress messages
	ApplyChangesStageValidation = "Validating environment and dependencies..."
	ApplyChangesStageBackup     = "Creating configuration backup..."
	ApplyChangesStageEnvFile    = "Updating environment file..."
	ApplyChangesStageCompose    = "Generating Docker Compose files..."
	ApplyChangesStageDocker     = "Managing Docker containers..."
	ApplyChangesStageServices   = "Starting services..."
	ApplyChangesStageComplete   = "Configuration changes applied successfully"

	// Apply Changes change list headers
	ApplyChangesChangesTitle  = "Pending Configuration Changes"
	ApplyChangesChangesCount  = "Total changes: %d"
	ApplyChangesChangesMasked = "(hidden for security)"
	ApplyChangesChangesEmpty  = "No changes to apply"

	// Apply Changes help content
	ApplyChangesHelpTitle   = "Applying Configuration Changes"
	ApplyChangesHelpContent = `Be sure to check the current configuration before applying changes.`
)

// apply changes integrity prompt
const (
	ApplyChangesIntegrityPromptTitle   = "File integrity check"
	ApplyChangesIntegrityPromptMessage = "Out-of-date files were detected.\nDo you want to update them to the latest version?"
	ApplyChangesIntegrityOutdatedList  = "Out-of-date files:\n%s\nConfirm update? (y/n)"
	ApplyChangesIntegrityChecking      = "Collecting file integrity information..."
	ApplyChangesIntegrityNoOutdated    = "No out-of-date files found. Proceeding with apply."
)

// Maintenance Screen constants
const (
	MaintenanceTitle       = "System Maintenance"
	MaintenanceDescription = "Manage PentAGI services and perform maintenance operations"
	MaintenanceName        = "Maintenance"
	MaintenanceOverview    = `Perform system maintenance operations for PentAGI.

Available operations depend on the current system state and will only be shown when applicable.

Operations include:
• Service lifecycle management (Start/Stop/Restart)
• Component updates and downloads
• System reset and cleanup
• Container and image management

Each operation will provide real-time status updates and confirmation when required.`

	// Maintenance menu items
	MaintenanceStartPentagi            = "Start PentAGI"
	MaintenanceStartPentagiDesc        = "Start all configured PentAGI services"
	MaintenanceStopPentagi             = "Stop PentAGI"
	MaintenanceStopPentagiDesc         = "Stop all running PentAGI services"
	MaintenanceRestartPentagi          = "Restart PentAGI"
	MaintenanceRestartPentagiDesc      = "Restart all PentAGI services"
	MaintenanceDownloadWorkerImage     = "Download Worker Image"
	MaintenanceDownloadWorkerImageDesc = "Download pentesting container image for worker tasks"
	MaintenanceUpdateWorkerImage       = "Update Worker Image"
	MaintenanceUpdateWorkerImageDesc   = "Update pentesting container image to latest version"
	MaintenanceUpdatePentagi           = "Update PentAGI"
	MaintenanceUpdatePentagiDesc       = "Update PentAGI to the latest version"
	MaintenanceUpdateInstaller         = "Update Installer"
	MaintenanceUpdateInstallerDesc     = "Update this installer to the latest version"
	MaintenanceFactoryReset            = "Factory Reset"
	MaintenanceFactoryResetDesc        = "Reset PentAGI to factory defaults"
	MaintenanceRemovePentagi           = "Remove PentAGI"
	MaintenanceRemovePentagiDesc       = "Remove PentAGI containers but keep data"
	MaintenancePurgePentagi            = "Purge PentAGI"
	MaintenancePurgePentagiDesc        = "Completely remove PentAGI including all data"
	MaintenanceResetPassword           = "Reset Admin Password"
	MaintenanceResetPasswordDesc       = "Reset the administrator password for PentAGI"
)

// Reset Password Screen constants
const (
	ResetPasswordFormTitle       = "Reset Admin Password"
	ResetPasswordFormDescription = "Reset the administrator password for PentAGI"
	ResetPasswordFormName        = "Reset Password"
	ResetPasswordFormOverview    = `Reset the password for the default administrator account (admin@pentagi.com).

This operation requires PentAGI to be running and will update the password in the PostgreSQL database.

Enter your new password twice to confirm and press Enter to apply the change.

Password requirements:
• Minimum 5 characters
• Both password fields must match`

	// Form fields
	ResetPasswordNewPassword         = "New Password"
	ResetPasswordNewPasswordDesc     = "Enter the new administrator password"
	ResetPasswordConfirmPassword     = "Confirm Password"
	ResetPasswordConfirmPasswordDesc = "Re-enter the new password to confirm"

	// Status messages
	ResetPasswordNotAvailable = "PentAGI must be running to reset password"
	ResetPasswordAvailable    = "Password reset is available"
	ResetPasswordInProgress   = "Resetting password..."
	ResetPasswordSuccess      = "Password has been successfully reset"
	ResetPasswordErrorPrefix  = "Error: "

	// Validation errors
	ResetPasswordErrorEmptyPassword = "Password cannot be empty"
	ResetPasswordErrorShortPassword = "Password must be at least 5 characters long"
	ResetPasswordErrorMismatch      = "Passwords do not match"

	// Help content
	ResetPasswordHelpContent = `Reset the administrator password for accessing PentAGI.

This operation:
• Updates the password for admin@pentagi.com account
• Sets the user status to 'active'
• Requires PentAGI database to be accessible
• Does not affect other user accounts

The password change takes effect immediately after successful completion.

Enter the same password in both fields and press Enter to confirm the change.`
)

// Processor Operation Form constants
const (
	// Dynamic title templates
	ProcessorOperationFormTitle       = "%s"
	ProcessorOperationFormDescription = "Execute %s operation"
	ProcessorOperationFormName        = "%s"

	// Common status messages
	ProcessorOperationNotStarted = "Ready to execute %s operation"
	ProcessorOperationInProgress = "Executing %s operation...\n"
	ProcessorOperationCompleted  = "%s operation completed successfully\n"
	ProcessorOperationFailed     = "Failed to execute %s operation"

	// Confirmation messages
	ProcessorOperationConfirmation = "Are you sure you want to %s?"
	ProcessorOperationPressEnter   = "Press Enter to %s"
	ProcessorOperationPressYN      = "Press Y to confirm, N to cancel"
	// Short notice without hotkeys (for static help panel)
	ProcessorOperationRequiresConfirmationShort = "This operation requires confirmation"
	// Additional terminal messages
	ProcessorOperationCancelled = "Operation cancelled"
	ProcessorOperationUnknown   = "Unknown operation: %s"

	// Operation specific messages
	ProcessorOperationStarting    = "Starting services..."
	ProcessorOperationStopping    = "Stopping services..."
	ProcessorOperationRestarting  = "Restarting services..."
	ProcessorOperationDownloading = "Downloading images..."
	ProcessorOperationUpdating    = "Updating components..."
	ProcessorOperationResetting   = "Resetting to factory defaults..."
	ProcessorOperationRemoving    = "Removing containers..."
	ProcessorOperationPurging     = "Purging all data..."
	ProcessorOperationInstalling  = "Installing PentAGI services..."

	// Help text templates
	ProcessorOperationHelpTitle           = "%s Operation"
	ProcessorOperationHelpContent         = "This operation will %s."
	ProcessorOperationHelpContentDownload = "This operation will download %s components."
	ProcessorOperationHelpContentUpdate   = "This operation will update %s components."
	// Generic title/description/builders for dynamic operations
	OperationTitleInstallPentagi    = "Install PentAGI"
	OperationDescInstallPentagi     = "Install and configure PentAGI services"
	OperationTitleDownload          = "Download %s"
	OperationDescDownloadComponents = "Download %s components"
	OperationTitleUpdate            = "Update %s"
	OperationDescUpdateToLatest     = "Update %s to latest version"
	OperationTitleExecute           = "Execute %s"
	OperationDescExecuteOn          = "Execute %s on %s"
	OperationProgressExecuting      = "Executing %s..."

	// Terminal not initialized
	ProcessorOperationTerminalNotInitialized = "Terminal is not initialized"
)

// Operation-specific help texts
const (
	ProcessorHelpInstallPentagi = `This will:
• Deploy Docker containers for selected services
• Configure networking and volumes
• Start all enabled services
• Set up monitoring if configured

Installation will use your current configuration settings.`

	ProcessorHelpStartPentagi = `This will:
• Core PentAGI API and web interface
• Configured Langfuse analytics (if enabled)
• Observability stack (if enabled)

Services will be started in the correct dependency order.`

	ProcessorHelpStopPentagi = `This will:
• Gracefully shutdown containers
• Preserve all data and configurations
• Network connections will be closed

You can restart services later without losing any data.`

	ProcessorHelpRestartPentagi = `This will:
• Stop running containers
• Apply any configuration changes
• Start services with fresh state

Useful after configuration updates or to resolve issues.`

	ProcessorHelpDownloadWorkerImage = `This large image (6GB+) contains:
• Kali Linux tools and utilities
• Security testing frameworks
• Network analysis software

Required for pentesting operations.`

	ProcessorHelpUpdateWorkerImage = `This will:
• Pull the latest pentesting image
• Update security tools and frameworks
• Preserve existing worker containers

Note: This is a large download (6GB+).`

	ProcessorHelpUpdatePentagi = `This will:
• Download latest container images
• Perform rolling update of services
• Preserve all data and configurations

Services will be briefly unavailable during update.`

	ProcessorHelpUpdateInstaller = `This will:
• Download the latest installer binary
• Replace the current installer
• Exit for manual restart

You'll need to restart the installer after update.`

	ProcessorHelpFactoryReset = `⚠️  WARNING: This operation will:
• Remove all containers and networks
• Delete all configuration files
• Clear stored data and volumes
• Restore default settings

This action cannot be undone!`

	ProcessorHelpRemovePentagi = `This will:
• Stop and remove all containers
• Remove Docker networks
• Preserve volumes and data
• Keep configuration files

You can reinstall later without losing data.`

	ProcessorHelpPurgePentagi = `⚠️  WARNING: This will permanently delete:
• All containers and images
• All data volumes
• All configuration files
• All stored results

This action cannot be undone!`
)

// environment variable descriptions (centralized)
const (
	EnvDesc_OPEN_AI_KEY                       = "OpenAI API Key"
	EnvDesc_OPEN_AI_SERVER_URL                = "OpenAI Server URL"
	EnvDesc_ANTHROPIC_API_KEY                 = "Anthropic API Key"
	EnvDesc_ANTHROPIC_SERVER_URL              = "Anthropic Server URL"
	EnvDesc_GEMINI_API_KEY                    = "Google Gemini API Key"
	EnvDesc_GEMINI_SERVER_URL                 = "Gemini Server URL"
	EnvDesc_BEDROCK_ACCESS_KEY_ID             = "AWS Bedrock Access Key"
	EnvDesc_BEDROCK_SECRET_ACCESS_KEY         = "AWS Bedrock Secret Key"
	EnvDesc_BEDROCK_SESSION_TOKEN             = "AWS Bedrock Session Token"
	EnvDesc_BEDROCK_REGION                    = "AWS Bedrock Region"
	EnvDesc_BEDROCK_SERVER_URL                = "AWS Bedrock Server URL"
	EnvDesc_OLLAMA_SERVER_URL                 = "Ollama Server URL"
	EnvDesc_OLLAMA_SERVER_MODEL               = "Ollama Default Model"
	EnvDesc_OLLAMA_SERVER_CONFIG_PATH         = "Ollama Container Config Path"
	EnvDesc_OLLAMA_SERVER_PULL_MODELS_TIMEOUT = "Ollama Model Pull Timeout"
	EnvDesc_OLLAMA_SERVER_PULL_MODELS_ENABLED = "Ollama Auto-pull Models"
	EnvDesc_OLLAMA_SERVER_LOAD_MODELS_ENABLED = "Ollama Load Models List"
	EnvDesc_LLM_SERVER_URL                    = "Custom LLM Server URL"
	EnvDesc_LLM_SERVER_KEY                    = "Custom LLM API Key"
	EnvDesc_LLM_SERVER_MODEL                  = "Custom LLM Model"
	EnvDesc_LLM_SERVER_CONFIG_PATH            = "Custom LLM Container Config Path"
	EnvDesc_LLM_SERVER_LEGACY_REASONING       = "Custom LLM Legacy Reasoning"
	EnvDesc_LLM_SERVER_PRESERVE_REASONING     = "Custom LLM Preserve Reasoning Content"
	EnvDesc_LLM_SERVER_PROVIDER               = "Custom LLM Provider Name"

	EnvDesc_LANGFUSE_LISTEN_IP   = "Langfuse Listen IP"
	EnvDesc_LANGFUSE_LISTEN_PORT = "Langfuse Listen Port"
	EnvDesc_LANGFUSE_BASE_URL    = "Langfuse Base URL"
	EnvDesc_LANGFUSE_PROJECT_ID  = "Langfuse Project ID"
	EnvDesc_LANGFUSE_PUBLIC_KEY  = "Langfuse Public Key"
	EnvDesc_LANGFUSE_SECRET_KEY  = "Langfuse Secret Key"

	// langfuse init variables
	EnvDesc_LANGFUSE_INIT_PROJECT_ID         = "Langfuse Init Project ID"
	EnvDesc_LANGFUSE_INIT_PROJECT_PUBLIC_KEY = "Langfuse Init Project Public Key"
	EnvDesc_LANGFUSE_INIT_PROJECT_SECRET_KEY = "Langfuse Init Project Secret Key"
	EnvDesc_LANGFUSE_INIT_USER_EMAIL         = "Langfuse Init User Email"
	EnvDesc_LANGFUSE_INIT_USER_NAME          = "Langfuse Init User Name"
	EnvDesc_LANGFUSE_INIT_USER_PASSWORD      = "Langfuse Init User Password"

	EnvDesc_LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT = "Langfuse OTLP endpoint for OpenTelemetry exporter"

	EnvDesc_GRAFANA_LISTEN_IP     = "Grafana Listen IP"
	EnvDesc_GRAFANA_LISTEN_PORT   = "Grafana Listen Port"
	EnvDesc_OTEL_GRPC_LISTEN_IP   = "OTel gRPC Listen IP"
	EnvDesc_OTEL_GRPC_LISTEN_PORT = "OTel gRPC Listen Port"
	EnvDesc_OTEL_HTTP_LISTEN_IP   = "OTel HTTP Listen IP"
	EnvDesc_OTEL_HTTP_LISTEN_PORT = "OTel HTTP Listen Port"
	EnvDesc_OTEL_HOST             = "OpenTelemetry Host"

	EnvDesc_SUMMARIZER_PRESERVE_LAST       = "Summarizer Preserve Last"
	EnvDesc_SUMMARIZER_USE_QA              = "Summarizer Use QA"
	EnvDesc_SUMMARIZER_SUM_MSG_HUMAN_IN_QA = "Summarizer Human in QA"
	EnvDesc_SUMMARIZER_LAST_SEC_BYTES      = "Summarizer Last Section Bytes"
	EnvDesc_SUMMARIZER_MAX_BP_BYTES        = "Summarizer Max BP Bytes"
	EnvDesc_SUMMARIZER_MAX_QA_BYTES        = "Summarizer Max QA Bytes"
	EnvDesc_SUMMARIZER_MAX_QA_SECTIONS     = "Summarizer Max QA Sections"
	EnvDesc_SUMMARIZER_KEEP_QA_SECTIONS    = "Summarizer Keep QA Sections"

	EnvDesc_ASSISTANT_SUMMARIZER_PRESERVE_LAST    = "Assistant Summarizer Preserve Last"
	EnvDesc_ASSISTANT_SUMMARIZER_LAST_SEC_BYTES   = "Assistant Summarizer Last Section Bytes"
	EnvDesc_ASSISTANT_SUMMARIZER_MAX_BP_BYTES     = "Assistant Summarizer Max BP Bytes"
	EnvDesc_ASSISTANT_SUMMARIZER_MAX_QA_BYTES     = "Assistant Summarizer Max QA Bytes"
	EnvDesc_ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS  = "Assistant Summarizer Max QA Sections"
	EnvDesc_ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS = "Assistant Summarizer Keep QA Sections"

	EnvDesc_EMBEDDING_PROVIDER        = "Embedding Provider"
	EnvDesc_EMBEDDING_URL             = "Embedding URL"
	EnvDesc_EMBEDDING_KEY             = "Embedding API Key"
	EnvDesc_EMBEDDING_MODEL           = "Embedding Model"
	EnvDesc_EMBEDDING_BATCH_SIZE      = "Embedding Batch Size"
	EnvDesc_EMBEDDING_STRIP_NEW_LINES = "Embedding Strip New Lines"

	EnvDesc_ASK_USER = "Human-in-the-loop"

	EnvDesc_ASSISTANT_USE_AGENTS = "Enable multi-agent mode for assistant"

	EnvDesc_SCRAPER_PUBLIC_URL                    = "Scraper Public URL"
	EnvDesc_SCRAPER_PRIVATE_URL                   = "Scraper Private URL"
	EnvDesc_LOCAL_SCRAPER_USERNAME                = "Local Scraper Username"
	EnvDesc_LOCAL_SCRAPER_PASSWORD                = "Local Scraper Password"
	EnvDesc_LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS = "Scraper Max Concurrent Sessions"

	EnvDesc_DUCKDUCKGO_ENABLED = "DuckDuckGo Search"
	EnvDesc_SPLOITUS_ENABLED   = "Sploitus Search"
	EnvDesc_PERPLEXITY_API_KEY = "Perplexity API Key"
	EnvDesc_TAVILY_API_KEY     = "Tavily API Key"
	EnvDesc_TRAVERSAAL_API_KEY = "Traversaal API Key"
	EnvDesc_GOOGLE_API_KEY     = "Google Search API Key"
	EnvDesc_GOOGLE_CX_KEY      = "Google Search CX Key"
	EnvDesc_GOOGLE_LR_KEY      = "Google Search LR Key"

	EnvDesc_DOCKER_INSIDE                    = "Docker Inside Container"
	EnvDesc_DOCKER_NET_ADMIN                 = "Docker Network Admin"
	EnvDesc_DOCKER_SOCKET                    = "Docker Socket Path"
	EnvDesc_DOCKER_NETWORK                   = "Docker Network"
	EnvDesc_DOCKER_PUBLIC_IP                 = "Docker Public IP"
	EnvDesc_DOCKER_WORK_DIR                  = "Docker Work Directory"
	EnvDesc_DOCKER_DEFAULT_IMAGE             = "Docker Default Image"
	EnvDesc_DOCKER_DEFAULT_IMAGE_FOR_PENTEST = "Docker Pentest Image"
	EnvDesc_DOCKER_HOST                      = "Docker Host"
	EnvDesc_DOCKER_TLS_VERIFY                = "Docker TLS Verify"
	EnvDesc_DOCKER_CERT_PATH                 = "Docker Certificate Path"

	EnvDesc_LICENSE_KEY                       = "PentAGI License Key"
	EnvDesc_PENTAGI_LISTEN_IP                 = "PentAGI Server Host"
	EnvDesc_PENTAGI_LISTEN_PORT               = "PentAGI Server Port"
	EnvDesc_PUBLIC_URL                        = "PentAGI Public URL"
	EnvDesc_CORS_ORIGINS                      = "PentAGI CORS Origins"
	EnvDesc_COOKIE_SIGNING_SALT               = "PentAGI Cookie Signing Salt"
	EnvDesc_PROXY_URL                         = "HTTP/HTTPS Proxy URL"
	EnvDesc_EXTERNAL_SSL_CA_PATH              = "Custom CA Certificate Path"
	EnvDesc_EXTERNAL_SSL_INSECURE             = "Skip SSL Verification"
	EnvDesc_PENTAGI_SSL_DIR                   = "PentAGI SSL Directory"
	EnvDesc_PENTAGI_DATA_DIR                  = "PentAGI Data Directory"
	EnvDesc_PENTAGI_DOCKER_SOCKET             = "Mount Docker Socket Path"
	EnvDesc_PENTAGI_DOCKER_CERT_PATH          = "Mount Docker Certificate Path"
	EnvDesc_PENTAGI_LLM_SERVER_CONFIG_PATH    = "Custom LLM Host Config Path"
	EnvDesc_PENTAGI_OLLAMA_SERVER_CONFIG_PATH = "Ollama Host Config Path"

	EnvDesc_STATIC_DIR     = "Frontend Static Directory"
	EnvDesc_STATIC_URL     = "Frontend Static URL"
	EnvDesc_SERVER_PORT    = "Backend Server Port"
	EnvDesc_SERVER_HOST    = "Backend Server Host"
	EnvDesc_SERVER_SSL_CRT = "Backend Server SSL Certificate Path"
	EnvDesc_SERVER_SSL_KEY = "Backend Server SSL Key Path"
	EnvDesc_SERVER_USE_SSL = "Backend Server Use SSL"

	EnvDesc_PERPLEXITY_MODEL        = "Perplexity Model"
	EnvDesc_PERPLEXITY_CONTEXT_SIZE = "Perplexity Context Size"

	EnvDesc_SEARXNG_URL        = "Searxng Search URL"
	EnvDesc_SEARXNG_CATEGORIES = "Searxng Search Categories"
	EnvDesc_SEARXNG_LANGUAGE   = "Searxng Search Language"
	EnvDesc_SEARXNG_SAFESEARCH = "Searxng Safe Search"
	EnvDesc_SEARXNG_TIME_RANGE = "Searxng Time Range"

	EnvDesc_OAUTH_GOOGLE_CLIENT_ID     = "OAuth Google Client ID"
	EnvDesc_OAUTH_GOOGLE_CLIENT_SECRET = "OAuth Google Client Secret"
	EnvDesc_OAUTH_GITHUB_CLIENT_ID     = "OAuth GitHub Client ID"
	EnvDesc_OAUTH_GITHUB_CLIENT_SECRET = "OAuth GitHub Client Secret"

	EnvDesc_LANGFUSE_EE_LICENSE_KEY   = "Langfuse Enterprise License Key"
	EnvDesc_PENTAGI_POSTGRES_PASSWORD = "PentAGI PostgreSQL Password"

	EnvDesc_GRAPHITI_URL        = "Graphiti Server URL"
	EnvDesc_GRAPHITI_TIMEOUT    = "Graphiti Request Timeout"
	EnvDesc_GRAPHITI_MODEL_NAME = "Graphiti Extraction Model"
	EnvDesc_NEO4J_USER          = "Neo4j Username"
	EnvDesc_NEO4J_DATABASE      = "Neo4j Database Name"
	EnvDesc_NEO4J_PASSWORD      = "Neo4j Database Password"
)

// dynamic, contextual sections used in processor operation forms
const (
	// section headers
	ProcessorSectionCurrentState = "Current state"
	ProcessorSectionPlanned      = "Planned actions"
	ProcessorSectionEffects      = "Effects"

	// component labels
	ProcessorComponentPentagi       = "PentAGI"
	ProcessorComponentLangfuse      = "Langfuse"
	ProcessorComponentObservability = "Observability"

	ProcessorComponentWorkerImage           = "worker image"
	ProcessorComponentComposeStacks         = "compose stacks"
	ProcessorComponentDefaultFiles          = "default files"
	ProcessorItemComposeFiles               = "compose files"
	ProcessorItemComposeStacksImagesVolumes = "compose stacks, images, volumes"

	// common states
	ProcessorStateInstalled = "installed"
	ProcessorStateMissing   = "not installed"
	ProcessorStateRunning   = "running"
	ProcessorStateStopped   = "stopped"
	ProcessorStateEmbedded  = "embedded"
	ProcessorStateExternal  = "external"
	ProcessorStateConnected = "connected"
	ProcessorStateDisabled  = "disabled"
	ProcessorStateUnknown   = "unknown"

	// planned action bullet prefixes
	PlannedWillStart    = "will start:"
	PlannedWillStop     = "will stop:"
	PlannedWillRestart  = "will restart:"
	PlannedWillUpdate   = "will update:"
	PlannedWillSkip     = "will skip:"
	PlannedWillRemove   = "will remove:"
	PlannedWillPurge    = "will purge:"
	PlannedWillDownload = "will download:"
	PlannedWillRestore  = "will restore:"

	// effect notes per operation (concise and practical)
	EffectsStart           = "PentAGI web UI becomes available. Background services are brought online in the required order."
	EffectsStop            = "Web UI becomes unavailable. In-progress flows pause safely. When you start PentAGI again, flows resume automatically. A small portion of the current agent step may be lost."
	EffectsRestart         = "Services stop and start again with a clean state. Brief downtime is expected. Flows resume automatically afterwards."
	EffectsUpdateAll       = "Images are pulled and services are recreated where needed. External or disabled components are skipped. Temporary downtime is expected."
	EffectsDownloadWorker  = "Running worker containers are not touched. New flows will use the downloaded image. To switch an existing flow to the new image, finish the flow and start a new task or create a new assistant."
	EffectsUpdateWorker    = "Pulls latest worker image. Running worker containers keep using the old image; new containers will use the updated one."
	EffectsUpdateInstaller = "The installer binary will be updated and the app will exit. Start the installer again to continue."
	EffectsFactoryReset    = "Removes containers, volumes and networks, restores default .env and embedded files. Produces a clean baseline. This action cannot be undone."
	EffectsRemove          = "Stops and removes containers but keeps volumes and images. Data is preserved. Web UI becomes unavailable until you start again."
	EffectsPurge           = "Complete cleanup: containers, images, volumes and configuration files are deleted. Irreversible."
	EffectsInstall         = "Required files are created and services are started. External components are detected and skipped."
)
