# Ollama Provider

The Ollama provider enables PentAGI to use local language models through the [Ollama](https://ollama.ai/) server.

## Installation

1. Install Ollama server on your system following the [official installation guide](https://ollama.ai/download)
2. Start the Ollama server (usually runs on `http://localhost:11434`)
3. Pull required models: `ollama pull gemma3:1b`

## Configuration

Configure the Ollama provider using environment variables:

### Required Variables

```bash
# Ollama server URL (default: http://localhost:11434)
OLLAMA_SERVER_URL=http://localhost:11434
```

### Optional Variables

```bash
# Default model for inference (optional, default: llama3.1:8b-instruct-q8_0)
OLLAMA_SERVER_MODEL=llama3.1:8b-instruct-q8_0

# Path to custom config file (optional)
OLLAMA_SERVER_CONFIG_PATH=/path/to/ollama_config.yml

# Model management settings (optional)
OLLAMA_SERVER_PULL_MODELS_TIMEOUT=600        # Timeout for model downloads in seconds
OLLAMA_SERVER_PULL_MODELS_ENABLED=false      # Auto-download models on startup
OLLAMA_SERVER_LOAD_MODELS_ENABLED=false      # Load model list from server

# Proxy URL if needed
PROXY_URL=http://proxy:8080
```

### Advanced Configuration

Control how PentAGI interacts with your Ollama server:

**Model Management:**

- **Auto-pull Models** (`OLLAMA_SERVER_PULL_MODELS_ENABLED=true`): Automatically downloads models specified in config file on startup
- **Pull Timeout** (`OLLAMA_SERVER_PULL_MODELS_TIMEOUT`): Maximum time to wait for model downloads (default: 600 seconds)
- **Load Models List** (`OLLAMA_SERVER_LOAD_MODELS_ENABLED=true`): Queries Ollama server for available models via API

**Performance Note:** Enabling `OLLAMA_SERVER_LOAD_MODELS_ENABLED` adds startup latency as PentAGI queries the Ollama API. Disable if you only need specific models from config file.

**Recommended Settings:**

```bash
# Fast startup (static config)
OLLAMA_SERVER_MODEL=llama3.1:8b-instruct-q8_0
OLLAMA_SERVER_PULL_MODELS_ENABLED=false
OLLAMA_SERVER_LOAD_MODELS_ENABLED=false

# Auto-discovery (dynamic config)
OLLAMA_SERVER_PULL_MODELS_ENABLED=true
OLLAMA_SERVER_PULL_MODELS_TIMEOUT=900
OLLAMA_SERVER_LOAD_MODELS_ENABLED=true
```

## Supported Models

The provider **dynamically loads models** from your local Ollama server. Available models depend on what you have installed locally.

**Popular model families include:**

- **Gemma models**: `gemma3:1b`, `gemma3:2b`, `gemma3:7b`, `gemma3:27b`
- **Llama models**: `llama3.1:7b`, `llama3.1:8b`, `llama3.1:8b-instruct-q8_0`, `llama3.1:8b-instruct-fp16`, `llama3.1:70b`, `llama3.2:1b`, `llama3.2:3b`, `llama3.2:90b`
- **Qwen models**: `qwen2.5:1.5b`, `qwen2.5:3b`, `qwen2.5:7b`, `qwen2.5:14b`, `qwen2.5:32b`, `qwen2.5:72b`
- **DeepSeek models**: `deepseek-r1:1.5b`, `deepseek-r1:7b`, `deepseek-r1:8b`, `deepseek-r1:14b`, `deepseek-r1:32b`
- **Embedding models**: `nomic-embed-text`

To see available models on your system: `ollama list`
To download new models: `ollama pull <model-name>`

## Features

- **Dynamic model discovery**: Automatically detects models installed on your Ollama server (when enabled)
- **Model caching**: Use only configured models without API calls (when load disabled)
- **Local inference**: No API keys required, models run locally
- **Auto model pulling**: Models are automatically downloaded when needed (when enabled)
- **Agent specialization**: Different agent types (assistant, coder, pentester) with optimized settings
- **Tool support**: Supports function calling for compatible models
- **Streaming**: Real-time response streaming
- **Custom configuration**: Override default settings with YAML config files
- **Zero pricing**: Local models have no usage costs

## Agent Types

The provider supports all PentAGI agent types with optimized configurations:

- `simple`: General purpose chat (temperature: 0.2)
- `assistant`: AI assistant tasks (temperature: 0.2)
- `coder`: Code generation (temperature: 0.1, max tokens: 6000)
- `pentester`: Security testing (temperature: 0.3, max tokens: 8000)
- `generator`: Content generation (temperature: 0.4)
- `refiner`: Content refinement (temperature: 0.3)
- `searcher`: Information searching (temperature: 0.2, max tokens: 3000)
- And more...

## Custom Configuration

Create a custom config file to override default settings:

```yaml
simple:
  model: "llama3.1:8b-instruct-q8_0"
  temperature: 0.2
  top_p: 0.3
  n: 1
  max_tokens: 4000

coder:
  model: "deepseek-r1:8b"
  temperature: 0.1
  top_p: 0.2
  n: 1
  max_tokens: 8000
```

Then set `OLLAMA_SERVER_CONFIG_PATH` to the file path.

## Pricing

Ollama provides free local inference - no usage costs or API limits.

## Example Usage

```bash
# Set environment variables
export OLLAMA_SERVER_URL=http://localhost:11434

# Start PentAGI with Ollama provider
./pentagi
```

## Troubleshooting

1. **Connection errors**: Ensure Ollama server is running and accessible
2. **Model not found**: Pull the model first with `ollama pull <model-name>`
3. **Performance issues**: Use smaller models for faster inference or upgrade hardware
4. **Memory issues**: Monitor system memory usage with larger models