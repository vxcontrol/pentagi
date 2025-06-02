# LLM Agent Testing Report

Generated: Tue, 13 May 2025 20:28:47 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|----------------|
| simple | openai/gpt-4.1-mini | false | 15/15 (100.00%) | 1.283s |
| simple_json | openai/gpt-4.1-mini | false | 5/5 (100.00%) | 1.274s |
| agent | openai/o4-mini | true | 15/15 (100.00%) | 2.749s |
| assistant | x-ai/grok-3-beta | false | 15/15 (100.00%) | 1.935s |
| generator | anthropic/claude-3.7-sonnet:thinking | true | 15/15 (100.00%) | 4.570s |
| refiner | google/gemini-2.5-flash-preview:thinking | true | 15/15 (100.00%) | 3.187s |
| adviser | google/gemini-2.5-pro-preview | true | 15/15 (100.00%) | 7.121s |
| reflector | openai/gpt-4.1-mini | false | 15/15 (100.00%) | 1.092s |
| searcher | x-ai/grok-3-mini-beta | true | 15/15 (100.00%) | 4.636s |
| enricher | openai/gpt-4.1-mini | false | 15/15 (100.00%) | 1.080s |
| coder | anthropic/claude-3.7-sonnet:thinking | true | 15/15 (100.00%) | 4.548s |
| installer | google/gemini-2.5-flash-preview:thinking | true | 15/15 (100.00%) | 3.686s |
| pentester | x-ai/grok-3-mini-beta | true | 15/15 (100.00%) | 4.410s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 3.302s

## Detailed Results

### simple (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.041s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.742s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.808s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.162s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.467s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.829s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.726s |  |
| Basic echo function | ✅ Pass | 1.045s |  |
| Streaming Basic echo function | ✅ Pass | 0.896s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.658s |  |
| Streaming JSON response function | ✅ Pass | 0.726s |  |
| Search query | ✅ Pass | 0.848s |  |
| Streaming Search query | ✅ Pass | 1.255s |  |
| Ask advice | ✅ Pass | 2.197s |  |
| Streaming Ask advice | ✅ Pass | 3.846s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.283s

---

### simple_json (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.702s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 0.814s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.016s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 1.495s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.342s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.274s

---

### agent (openai/o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.068s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.268s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.447s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.358s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.896s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.124s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.546s |  |
| Basic echo function | ✅ Pass | 2.108s |  |
| Streaming Basic echo function | ✅ Pass | 2.451s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.919s |  |
| Streaming JSON response function | ✅ Pass | 2.961s |  |
| Search query | ✅ Pass | 1.580s |  |
| Streaming Search query | ✅ Pass | 2.187s |  |
| Ask advice | ✅ Pass | 7.227s |  |
| Streaming Ask advice | ✅ Pass | 9.089s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.749s

---

### assistant (x-ai/grok-3-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.815s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.596s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.034s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.737s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.626s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.657s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.782s |  |
| Basic echo function | ✅ Pass | 1.388s |  |
| Streaming Basic echo function | ✅ Pass | 1.358s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.483s |  |
| Streaming JSON response function | ✅ Pass | 1.367s |  |
| Search query | ✅ Pass | 1.315s |  |
| Streaming Search query | ✅ Pass | 1.414s |  |
| Ask advice | ✅ Pass | 6.346s |  |
| Streaming Ask advice | ✅ Pass | 9.110s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.935s

---

### generator (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.742s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.790s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.101s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.853s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.334s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.281s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.874s |  |
| Basic echo function | ✅ Pass | 7.674s |  |
| Streaming Basic echo function | ✅ Pass | 3.526s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.315s |  |
| Streaming JSON response function | ✅ Pass | 3.765s |  |
| Search query | ✅ Pass | 2.085s |  |
| Streaming Search query | ✅ Pass | 3.392s |  |
| Ask advice | ✅ Pass | 17.240s |  |
| Streaming Ask advice | ✅ Pass | 14.585s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.570s

---

### refiner (google/gemini-2.5-flash-preview:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.760s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.825s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.731s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.912s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.756s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.782s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.919s |  |
| Basic echo function | ✅ Pass | 2.029s |  |
| Streaming Basic echo function | ✅ Pass | 1.974s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.117s |  |
| Streaming JSON response function | ✅ Pass | 2.748s |  |
| Search query | ✅ Pass | 2.703s |  |
| Streaming Search query | ✅ Pass | 2.150s |  |
| Ask advice | ✅ Pass | 7.263s |  |
| Streaming Ask advice | ✅ Pass | 13.140s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.187s

---

### adviser (google/gemini-2.5-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.409s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.936s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.263s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.116s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.497s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.306s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.502s |  |
| Basic echo function | ✅ Pass | 2.942s |  |
| Streaming Basic echo function | ✅ Pass | 1.989s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.429s |  |
| Streaming JSON response function | ✅ Pass | 4.249s |  |
| Search query | ✅ Pass | 3.088s |  |
| Streaming Search query | ✅ Pass | 3.009s |  |
| Ask advice | ✅ Pass | 37.403s |  |
| Streaming Ask advice | ✅ Pass | 30.672s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.121s

---

### reflector (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.841s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.698s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.782s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.241s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.819s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.665s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.938s |  |
| Basic echo function | ✅ Pass | 1.106s |  |
| Streaming Basic echo function | ✅ Pass | 0.780s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.703s |  |
| Streaming JSON response function | ✅ Pass | 0.961s |  |
| Search query | ✅ Pass | 0.852s |  |
| Streaming Search query | ✅ Pass | 0.949s |  |
| Ask advice | ✅ Pass | 2.458s |  |
| Streaming Ask advice | ✅ Pass | 2.585s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.092s

---

### searcher (x-ai/grok-3-mini-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.135s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.463s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.940s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.202s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.269s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.545s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.447s |  |
| Basic echo function | ✅ Pass | 4.651s |  |
| Streaming Basic echo function | ✅ Pass | 4.334s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.938s |  |
| Streaming JSON response function | ✅ Pass | 4.808s |  |
| Search query | ✅ Pass | 5.180s |  |
| Streaming Search query | ✅ Pass | 4.175s |  |
| Ask advice | ✅ Pass | 6.516s |  |
| Streaming Ask advice | ✅ Pass | 8.942s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.636s

---

### enricher (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.216s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.649s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.705s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.710s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.873s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.798s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.054s |  |
| Basic echo function | ✅ Pass | 1.274s |  |
| Streaming Basic echo function | ✅ Pass | 0.846s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.740s |  |
| Streaming JSON response function | ✅ Pass | 1.032s |  |
| Search query | ✅ Pass | 0.793s |  |
| Streaming Search query | ✅ Pass | 0.862s |  |
| Ask advice | ✅ Pass | 2.119s |  |
| Streaming Ask advice | ✅ Pass | 2.527s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.080s

---

### coder (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.133s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.158s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.740s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.651s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.325s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.128s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.337s |  |
| Basic echo function | ✅ Pass | 3.849s |  |
| Streaming Basic echo function | ✅ Pass | 8.092s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.365s |  |
| Streaming JSON response function | ✅ Pass | 3.957s |  |
| Search query | ✅ Pass | 4.005s |  |
| Streaming Search query | ✅ Pass | 3.102s |  |
| Ask advice | ✅ Pass | 14.507s |  |
| Streaming Ask advice | ✅ Pass | 14.869s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.548s

---

### installer (google/gemini-2.5-flash-preview:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.589s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.538s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.358s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.752s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.835s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.012s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.935s |  |
| Basic echo function | ✅ Pass | 2.061s |  |
| Streaming Basic echo function | ✅ Pass | 1.824s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.999s |  |
| Streaming JSON response function | ✅ Pass | 2.129s |  |
| Search query | ✅ Pass | 2.097s |  |
| Streaming Search query | ✅ Pass | 2.206s |  |
| Ask advice | ✅ Pass | 14.223s |  |
| Streaming Ask advice | ✅ Pass | 12.739s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.686s

---

### pentester (x-ai/grok-3-mini-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.810s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.009s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.108s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.633s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.223s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.321s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.473s |  |
| Basic echo function | ✅ Pass | 3.682s |  |
| Streaming Basic echo function | ✅ Pass | 3.898s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.719s |  |
| Streaming JSON response function | ✅ Pass | 4.657s |  |
| Search query | ✅ Pass | 4.354s |  |
| Streaming Search query | ✅ Pass | 3.676s |  |
| Ask advice | ✅ Pass | 7.859s |  |
| Streaming Ask advice | ✅ Pass | 8.722s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.410s

---

