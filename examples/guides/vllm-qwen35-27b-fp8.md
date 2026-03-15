# Running PentAGI with vLLM and Qwen3.5-27B-FP8

This guide explains how to deploy PentAGI with a fully local LLM setup using vLLM and Qwen3.5-27B-FP8. This configuration enables complete independence from cloud API providers while maintaining high performance for autonomous penetration testing workflows.

## Table of Contents

- [Model Overview](#model-overview)
- [Hardware Requirements](#hardware-requirements)
- [Prerequisites](#prerequisites)
- [vLLM Installation](#vllm-installation)
- [Server Configuration](#server-configuration)
- [Testing the Deployment](#testing-the-deployment)
- [PentAGI Integration](#pentagi-integration)
- [Performance Benchmarks](#performance-benchmarks)
- [Troubleshooting](#troubleshooting)

---

## Model Overview

**Qwen3.5-27B** is a state-of-the-art dense language model from Alibaba Cloud with 27 billion parameters fully active on every token. It features a hybrid architecture combining:
- **75% Gated DeltaNet layers** (linear attention)
- **25% Gated Attention layers** (traditional attention)
- **Native context window**: 262,144 tokens
- **Extended context**: Up to 1,010,000 tokens via YaRN
- **Quantization**: FP8 W8A8 with block size 128 (performance nearly identical to BF16)

This model is particularly well-suited for PentAGI's multi-agent workflows due to its:
- Strong reasoning capabilities with native thinking mode
- Excellent function calling support
- Large context window for complex security analysis
- Fast inference speed with FP8 quantization

---

## Hardware Requirements

FP8 W8A8 hardware acceleration requires GPUs with **Compute Capability ≥ 8.9** (Ada Lovelace, Hopper, or Blackwell architectures). On older GPUs like Ampere (A100, A6000, RTX 3090), FP8 falls back to W8A16 mode via Marlin kernels with reduced performance.

### Supported GPU Configurations

| Configuration | Total VRAM | Max Context | FP8 Mode | Status |
|---|---|---|---|---|
| 2× RTX 5090 (64 GB) | 64 GB | ≤131k | W8A8 | Good |
| **4× RTX 5090 (128 GB)** | **128 GB** | **262k (native)** | **W8A8** | **Tested (~30 GB/GPU)** |
| 1× H100 SXM (80 GB) | 80 GB | 262k | W8A8 | Single GPU |
| 2× H100 SXM (160 GB) | 160 GB | 262k | W8A8 | Excellent |
| 4× A100 80GB (320 GB) | 320 GB | 262k | W8A16 | Slower fallback |

---

## Prerequisites

### System Requirements

- **OS**: Linux (Ubuntu 22.04+ recommended)
- **CUDA**: 12.1 or higher
- **Python**: 3.9 - 3.12
- **GPU Drivers**: Latest NVIDIA drivers (535+)
- **NCCL**: 2.27.3+ (for multi-GPU setups)

### Required Software

Install CUDA toolkit and verify installation:

```bash
nvidia-smi
nvcc --version
```

Install Python package manager (uv recommended for faster installation):

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

---

## vLLM Installation

### Install vLLM Nightly Build

**IMPORTANT**: The `qwen3_5` architecture is not recognized in stable vLLM releases. You **must** use the nightly build until vLLM v0.17.0 is released.

**Option 1: Using uv (recommended)**

```bash
uv pip install vllm --torch-backend=auto --extra-index-url https://wheels.vllm.ai/nightly
```

**Option 2: Using pip**

```bash
pip install vllm --pre --extra-index-url https://wheels.vllm.ai/nightly
```

**Option 3: Docker (alternative)**

```bash
docker pull vllm/vllm-openai:nightly
```

### Verify Installation

```bash
python -c "import vllm; print(vllm.__version__)"
```

---

## Server Configuration

### Recommended vLLM Parameters

The following configuration has been tested and optimized for 4× RTX 5090 GPUs with ~30 GB VRAM usage per GPU at `--gpu-memory-utilization 0.75`:

| Parameter | Value | Explanation |
|---|---|---|
| `--model` | `Qwen/Qwen3.5-27B-FP8` | HuggingFace model identifier |
| `--tensor-parallel-size` | `4` | Number of GPUs (1 shard per GPU) |
| `--max-model-len` | `262144` | Native context window size |
| `--max-num-batched-tokens` | `4096` | Optimal for low inter-token latency in chat |
| `--block-size` | `128` | Matches FP8 quantization block size |
| `--gpu-memory-utilization` | `0.75` | VRAM allocation ratio (adjust as needed) |
| `--language-model-only` | flag | Skip vision encoder → +2-4 GB KV-cache |
| `--enable-prefix-caching` | flag | Cache repeated system prompts |
| `--reasoning-parser` | `qwen3` | Enable Qwen3.5 reasoning/thinking mode parser |
| `--tool-call-parser` | `qwen3_xml` | Prevents infinite `!!!!` bug with long contexts |
| `--attention-backend` | `FLASHINFER` | Best for Ada/Hopper/Blackwell GPUs |
| `--speculative-config` | `'{"method":"qwen3_next_mtp","num_speculative_tokens":1}'` | Enable Medusa-based speculative decoding (MTP) |
| `-O3` | flag | Maximum optimization via torch.compile |

### Start vLLM Server

**For Single GPU (H200, B200, B300):**

```bash
vllm serve Qwen/Qwen3.5-27B-FP8 \
  --max-model-len 262144 \
  --max-num-batched-tokens 4096 \
  --block-size 128 \
  --gpu-memory-utilization 0.75 \
  --language-model-only \
  --enable-prefix-caching \
  --reasoning-parser qwen3 \
  --tool-call-parser qwen3_xml \
  --attention-backend FLASHINFER \
  --speculative-config '{"method":"qwen3_next_mtp","num_speculative_tokens":1}' \
  -O3 \
  --host 127.0.0.1 \
  --port 8000
```

**For Multi-GPU (4× RTX 5090):**

```bash
NCCL_P2P_DISABLE=1 vllm serve Qwen/Qwen3.5-27B-FP8 \
  --tensor-parallel-size 4 \
  --max-model-len 262144 \
  --max-num-batched-tokens 4096 \
  --block-size 128 \
  --gpu-memory-utilization 0.75 \
  --language-model-only \
  --enable-prefix-caching \
  --reasoning-parser qwen3 \
  --tool-call-parser qwen3_xml \
  --attention-backend FLASHINFER \
  --speculative-config '{"method":"qwen3_next_mtp","num_speculative_tokens":1}' \
  -O3 \
  --host 127.0.0.1 \
  --port 8000
```

**Multi-GPU Note**: The `NCCL_P2P_DISABLE=1` environment variable is **required** for Blackwell GPUs (RTX 5090) with tensor parallelism > 1 to prevent NCCL hangs. Update `nvidia-nccl-cu12` to version 2.27.3+ for additional stability.

### Optional: Disable Thinking Mode by Default

To disable the thinking mode at the server level (can still be enabled per-request):

```bash
vllm serve Qwen/Qwen3.5-27B-FP8 \
  --default-chat-template-kwargs '{"enable_thinking": false}' \
  # ... other parameters
```

### Important: Multi-Turn Conversations

**Best Practice**: In multi-turn conversations, the historical model output should **only include the final output** and **not the thinking content** (`<think>...</think>` tags). This is automatically handled by vLLM's Jinja2 chat template, but if you're implementing custom conversation handling, ensure thinking tags are stripped from message history.

---

## Testing the Deployment

After starting the vLLM server, verify it's working correctly with these test requests.

### Test 1: Thinking Mode Enabled (Default)

```bash
curl "http://127.0.0.1:8000/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen/Qwen3.5-27B-FP8",
    "messages": [{"role": "user", "content": "hey! what is the weather in Moscow?"}],
    "temperature": 1.0,
    "top_k": 20,
    "top_p": 0.95,
    "min_p": 0.0,
    "presence_penalty": 1.5,
    "repetition_penalty": 1.0
  }'
```

**Expected**: Response includes `<think>` tags with reasoning process.

### Test 2: Thinking Mode Disabled (Non-Thinking)

```bash
curl "http://127.0.0.1:8000/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen/Qwen3.5-27B-FP8",
    "messages": [{"role": "user", "content": "hey! what is the weather in Beijing?"}],
    "temperature": 0.7,
    "top_k": 20,
    "top_p": 0.8,
    "min_p": 0.0,
    "presence_penalty": 1.5,
    "repetition_penalty": 1.0,
    "chat_template_kwargs": {"enable_thinking": false}
  }'
```

**Expected**: Direct response without `<think>` tags.

### Test 3: Higher Temperature Reasoning

```bash
curl "http://127.0.0.1:8000/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen/Qwen3.5-27B-FP8",
    "messages": [{"role": "user", "content": "hey! what is the weather in New York?"}],
    "temperature": 1.0,
    "top_k": 40,
    "top_p": 1.0,
    "min_p": 0.0,
    "presence_penalty": 2.0,
    "repetition_penalty": 1.0,
    "chat_template_kwargs": {"enable_thinking": false}
  }'
```

**Expected**: Creative/diverse responses without thinking tags.

If all tests return valid JSON responses with appropriate content, your vLLM server is ready for PentAGI integration.

---

## Recommended Sampling Parameters

The Qwen team provides official recommendations for sampling parameters optimized for different use cases:

| Mode | temp | top_p | top_k | presence_penalty |
|---|---|---|---|---|
| **Thinking, general tasks** | 1.0 | 0.95 | 20 | 1.5 |
| **Thinking, coding (WebDev)** | 0.6 | 0.95 | 20 | 0.0 |
| **Non-thinking (Instruct), general** | 0.7 | 0.8 | 20 | 1.5 |
| **Non-thinking (Instruct), reasoning** | 1.0 | 1.0 | 40 | 2.0 |

**Additional parameters:**
- `repetition_penalty=1.0` for all modes
- `max_tokens=32768` for most tasks
- `max_tokens=81920` for complex math/coding tasks

These parameters are already applied in the PentAGI provider configuration files referenced below.

---

## PentAGI Integration

### Step 1: Configure Custom Provider in PentAGI

PentAGI includes pre-configured provider files for Qwen3.5-27B-FP8 with optimized sampling parameters for different agent roles.

**Two provider configurations are available:**

1. **With Thinking Mode** (default): [`examples/configs/vllm-qwen3.5-27b-fp8.provider.yml`](../configs/vllm-qwen3.5-27b-fp8.provider.yml)
   - Enables `<think>` tags for primary agents (primary_agent, assistant, adviser, refiner, generator)
   - Uses `temp=0.6` for coding agents (coder, installer, pentester)
   - Recommended for maximum reasoning quality

2. **Without Thinking Mode**: [`examples/configs/vllm-qwen3.5-27b-fp8-no-think.provider.yml`](../configs/vllm-qwen3.5-27b-fp8-no-think.provider.yml)
   - Disables thinking for all agents via `chat_template_kwargs`
   - Uses `temp=0.7` for general tasks, `temp=1.0` for reasoning
   - Recommended for faster responses

### Step 2: Add Provider via PentAGI UI

1. Start PentAGI (see [Quick Start](../../README.md#-quick-start))
2. Navigate to **Settings → Providers**
3. Click **Add Provider**
4. Fill in the form:
   - **Name**: `vLLM Qwen3.5-27B-FP8` (or any custom name)
   - **Type**: `Custom`
   - **Base URL**: `http://127.0.0.1:8000/v1` (or your vLLM server address)
   - **API Key**: `dummy` (vLLM doesn't require authentication by default)
   - **Configuration**: Copy contents from one of the YAML files above
5. Click **Save**

### Step 3: Verify Provider Configuration

Test the provider by creating a simple flow:

1. Navigate to **Flows**
2. Click **New Flow**
3. Select your newly created provider
4. Enter a test task: `"Scan localhost port 80"`
5. Monitor execution logs

---

## Performance Benchmarks

Based on internal testing with 4× RTX 5090 GPUs and 10 concurrent requests:

| Metric | Value |
|---|---|
| **Prompt Processing Speed** | ~13,000 tokens/sec |
| **Completion Generation Speed** | ~650 tokens/sec |
| **Concurrent Flows** | 12 flows simultaneously with stable performance |
| **VRAM Usage** | ~30 GB per GPU (at 0.75 utilization) |
| **Context Window** | Full 262K tokens supported |

These benchmarks demonstrate that Qwen3.5-27B-FP8 provides excellent throughput for running multiple PentAGI flows in parallel, making it suitable for production deployments.

---

## Troubleshooting

### Issue: "Unknown architecture 'qwen3_5'"

**Cause**: Using stable vLLM release instead of nightly.

**Solution**: Install vLLM nightly build:

```bash
uv pip install vllm --torch-backend=auto --extra-index-url https://wheels.vllm.ai/nightly
```

### Issue: NCCL Hangs on Multi-GPU Setup

**Cause**: Blackwell GPUs (RTX 5090) require P2P communication to be disabled when using tensor parallelism.

**Solution**: Set environment variable before starting vLLM:

```bash
export NCCL_P2P_DISABLE=1
```

Also update NCCL library:

```bash
pip install --upgrade nvidia-nccl-cu12
```

### Issue: `enable_thinking` Parameter Ignored

**Cause**: Parameter must be passed inside `chat_template_kwargs`, not at root level.

**Solution**: Use correct JSON structure:

```json
{
  "messages": [...],
  "chat_template_kwargs": {"enable_thinking": false}
}
```

### Issue: Infinite `!!!!` Generation on Long Contexts

**Cause**: Using `qwen3_coder` parser with long contexts triggers a known bug.

**Solution**: Switch to XML parser:

```bash
--tool-call-parser qwen3_xml
```

### Issue: Out of Memory (OOM)

**Cause**: Insufficient VRAM for chosen context length.

**Solution**: Reduce `--max-model-len` or `--gpu-memory-utilization`:

```bash
# Reduce context window
--max-model-len 131072

# Or reduce VRAM allocation
--gpu-memory-utilization 0.7
```

### Issue: Speculative Decoding Errors

**Cause**: `num_speculative_tokens > 1` is unstable in current nightly builds.

**Solution**: Use only 1 speculative token:

```bash
--speculative-config '{"method":"qwen3_next_mtp","num_speculative_tokens":1}'
```

---

## Advanced: Extended Context with YaRN

Qwen3.5-27B natively supports 262K tokens. For tasks requiring longer context (up to 1,010,000 tokens), you can enable YaRN (Yet another RoPE extensioN) scaling.

### Enable YaRN via Command Line

```bash
VLLM_ALLOW_LONG_MAX_MODEL_LEN=1 vllm serve Qwen/Qwen3.5-27B-FP8 \
  --hf-overrides '{"text_config": {"rope_parameters": {"mrope_interleaved": true, "mrope_section": [11, 11, 10], "rope_type": "yarn", "rope_theta": 10000000, "partial_rotary_factor": 0.25, "factor": 4.0, "original_max_position_embeddings": 262144}}}' \
  --max-model-len 1010000 \
  # ... other parameters
```

**Important Notes:**
- YaRN uses a **static scaling factor** regardless of input length, which may impact performance on shorter texts
- Only enable YaRN when processing long contexts is required
- Adjust `factor` based on typical context length (e.g., `factor=2.0` for 524K tokens)
- For most PentAGI workflows, the native 262K context is sufficient

---

## Additional Resources

- **Official Qwen3.5 Documentation**: [HuggingFace Model Card](https://huggingface.co/Qwen/Qwen3.5-27B-FP8)
- **vLLM Documentation**: [docs.vllm.ai](https://docs.vllm.ai/)
- **vLLM Qwen3.5 Recipe**: [Official vLLM Guide](https://docs.vllm.ai/en/latest/models/supported_models/)
- **PentAGI Main Documentation**: [README.md](../../README.md)
- **Provider Configuration Reference**: See example configs in [`examples/configs/`](../configs/)
