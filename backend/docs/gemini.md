# Google AI (Gemini) Provider

The Google AI provider enables PentAGI to use Google's Gemini language models through the Generative AI API. This provider supports advanced features like function calling, streaming responses, and competitive pricing.

## Features

- **Multi-model Support**: Access to Gemini 2.5 Flash, Gemini 2.5 Pro, and other Google AI models
- **Function Calling**: Full support for tool usage and function calls
- **Streaming Responses**: Real-time response streaming for better user experience
- **Competitive Pricing**: Cost-effective inference with transparent pricing
- **Proxy Support**: HTTP proxy support for enterprise environments
- **Advanced Configuration**: Fine-tuned parameters for different agent types

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GEMINI_API_KEY` | - | Your Google AI API key (required) |
| `GEMINI_SERVER_URL` | `https://generativelanguage.googleapis.com` | Google AI API base URL |

### Getting API Key

1. Visit [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Create a new API key
3. Set it as `GEMINI_API_KEY` environment variable

## Available Models

| Model | Context Window | Max Output | Input Price* | Output Price* | Best For |
|-------|----------------|------------|--------------|---------------|----------|
| gemini-2.5-flash | 1M tokens | 65K tokens | $0.15 | $0.60 | General tasks, fast responses |
| gemini-2.5-pro | 1M tokens | 65K tokens | $2.50 | $10.00 | Complex reasoning, analysis |
| gemini-2.0-flash | 1M tokens | 8K tokens | $0.15 | $0.60 | High-frequency tasks |
| gemini-1.5-flash | 1M tokens | 8K tokens | $0.075 | $0.30 | Legacy model (deprecated) |
| gemini-1.5-pro | 2M tokens | 8K tokens | $1.25 | $5.00 | Legacy model (deprecated) |

*Prices per 1M tokens (USD)

## Agent Configuration

Each agent type is optimized with specific parameters for Google AI models:

### Basic Agents
- **Simple**: General-purpose tasks with balanced settings
- **Simple JSON**: Structured output generation with JSON formatting
- **Primary Agent**: Core reasoning with moderate creativity
- **Assistant (A)**: User interaction with contextual responses

### Specialized Agents
- **Generator**: Creative content with higher temperature
- **Refiner**: Content improvement with focused parameters
- **Adviser**: Strategic guidance with extended context
- **Reflector**: Analysis and evaluation tasks
- **Searcher**: Information retrieval with precise settings
- **Enricher**: Data enhancement and augmentation
- **Coder**: Programming tasks with minimal temperature
- **Installer**: System setup with deterministic responses
- **Pentester**: Security testing with balanced creativity

## Usage Examples

### Basic Setup

```bash
# Set environment variables
export GEMINI_API_KEY="your_api_key_here"
export GEMINI_SERVER_URL="https://generativelanguage.googleapis.com"

# Test the provider
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester -type gemini
```

### Custom Configuration

```yaml
# gemini-custom.yml
simple:
  model: "gemini-2.5-pro"
  temperature: 0.3
  top_p: 0.4
  max_tokens: 8000
  price:
    input: 2.50
    output: 10.00

coder:
  model: "gemini-2.5-flash"
  temperature: 0.05
  top_p: 0.1
  max_tokens: 16000
  price:
    input: 0.15
    output: 0.60
```

### Docker Usage

```bash
# Using pre-configured Gemini provider
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester \
  -config /opt/pentagi/conf/gemini.provider.yml

# Using custom configuration
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  -v $(pwd)/gemini-custom.yml:/opt/pentagi/gemini-custom.yml \
  vxcontrol/pentagi /opt/pentagi/bin/ctester \
  -type gemini \
  -config /opt/pentagi/gemini-custom.yml
```

## Integration with PentAGI

### Environment File (.env)

```bash
# Google AI Configuration
GEMINI_API_KEY=your_api_key_here
GEMINI_SERVER_URL=https://generativelanguage.googleapis.com

# Optional: Proxy settings
PROXY_URL=http://your-proxy:port
```

### Provider Selection

The Google AI provider is automatically available when `GEMINI_API_KEY` is set. You can use it for:

- **Flow Execution**: Autonomous penetration testing workflows
- **Assistant Mode**: Interactive chat and analysis
- **Custom Tasks**: Specialized security assessments
- **API Integration**: Programmatic access to Google AI models

## Best Practices

### Model Selection
- Use **gemini-2.5-flash** for general tasks and fast responses
- Use **gemini-2.5-pro** for complex reasoning and detailed analysis
- Avoid deprecated models (1.5 series) for new projects

### Performance Optimization
- Set appropriate `max_tokens` limits based on your use case
- Use lower `temperature` values for deterministic tasks
- Configure `top_p` to balance creativity and consistency

### Cost Management
- Monitor token usage through PentAGI's cost tracking
- Use cheaper models for simple tasks
- Implement request batching where possible

### Security Considerations
- Store API keys securely (environment variables, secrets management)
- Use HTTPS for all API communications
- Implement rate limiting to prevent abuse
- Monitor API usage and costs regularly

## Troubleshooting

### Common Issues

1. **API Key Issues**
   ```
   Error: failed to create gemini provider: invalid API key
   ```
   - Verify your API key is correct
   - Check API key permissions in Google AI Studio
   - Ensure the key hasn't expired

2. **Model Not Found**
   ```
   Error: model "gemini-x.x-xxx" not found
   ```
   - Use supported model names from the table above
   - Check for typos in model names
   - Verify model availability in your region

3. **Rate Limiting**
   ```
   Error: quota exceeded
   ```
   - Implement exponential backoff
   - Reduce request frequency
   - Check your quota limits in Google AI Studio

4. **Network Issues**
   ```
   Error: connection timeout
   ```
   - Check internet connectivity
   - Verify proxy settings if applicable
   - Check firewall rules for outbound HTTPS

### Testing Provider

```bash
# Test basic functionality
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester \
  -type gemini \
  -agent simple \
  -prompt "Hello, world!"

# Test JSON functionality
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester \
  -type gemini \
  -agent simple_json \
  -prompt "Generate a JSON object with name and age fields"

# Test all agents
docker run --rm \
  -v $(pwd)/.env:/opt/pentagi/.env \
  vxcontrol/pentagi /opt/pentagi/bin/ctester \
  -type gemini
```

## Support and Resources

- [Google AI Documentation](https://ai.google.dev/docs)
- [Gemini API Reference](https://ai.google.dev/api)
- [PentAGI Documentation](https://docs.pentagi.com)
- [Issue Tracker](https://github.com/vxcontrol/pentagi/issues)

For provider-specific issues, include:
- Provider type: `gemini`
- Model name used
- Configuration snippet (without API keys)
- Error messages and logs
- Environment details (Docker, OS, etc.)