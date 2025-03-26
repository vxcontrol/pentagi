# LLM Agent Testing Report

Generated: Mon, 24 Mar 2025 11:14:23 UTC

## Overall Results

| Agent | Model | Success Rate | Average Latency |
|-------|-------|-------------|----------------|
| simple | openai/gpt-4o | 9/9 (100.00%) | 0.957s |
| simple_json | openai/gpt-4o-mini | 5/5 (100.00%) | 1.636s |
| agent | openai/gpt-4o | 9/9 (100.00%) | 1.022s |
| generator | anthropic/claude-3.7-sonnet:thinking | 9/9 (100.00%) | 3.167s |
| refiner | anthropic/claude-3.7-sonnet:thinking | 9/9 (100.00%) | 3.305s |
| adviser | anthropic/claude-3.7-sonnet:thinking | 9/9 (100.00%) | 3.196s |
| reflector | mistralai/mistral-large-2411 | 8/9 (88.89%) | 3.359s |
| searcher | openai/gpt-4o-mini | 9/9 (100.00%) | 1.111s |
| enricher | openai/gpt-4o | 9/9 (100.00%) | 1.135s |
| coder | anthropic/claude-3.7-sonnet | 9/9 (100.00%) | 1.696s |
| installer | openai/gpt-4o | 9/9 (100.00%) | 0.965s |
| pentester | openai/gpt-4o | 9/9 (100.00%) | 1.805s |

**Total**: 103/104 (99.04%) successful tests
**Overall average latency**: 1.958s

## Detailed Results

### simple (openai/gpt-4o)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.723s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.776s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.878s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.617s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.943s |  |
| Basic echo function | ✅ Pass | 0.825s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.888s |  |
| Search query | ✅ Pass | 0.826s |  |
| Ask advice | ✅ Pass | 1.138s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 0.957s

---

### simple_json (openai/gpt-4o-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.468s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 1.408s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.359s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 1.756s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.188s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.636s

---

### agent (openai/gpt-4o)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.753s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.201s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.572s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.619s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.920s |  |
| Basic echo function | ✅ Pass | 0.781s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.914s |  |
| Search query | ✅ Pass | 0.964s |  |
| Ask advice | ✅ Pass | 1.474s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 1.022s

---

### generator (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.111s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.572s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.148s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.698s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.174s |  |
| Basic echo function | ✅ Pass | 3.384s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.299s |  |
| Search query | ✅ Pass | 4.040s |  |
| Ask advice | ✅ Pass | 5.081s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 3.167s

---

### refiner (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.769s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.777s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.770s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.353s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.529s |  |
| Basic echo function | ✅ Pass | 3.252s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.417s |  |
| Search query | ✅ Pass | 3.340s |  |
| Ask advice | ✅ Pass | 6.540s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 3.305s

---

### adviser (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.942s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.913s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.726s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.694s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.698s |  |
| Basic echo function | ✅ Pass | 3.306s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.928s |  |
| Search query | ✅ Pass | 3.886s |  |
| Ask advice | ✅ Pass | 5.668s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 3.196s

---

### reflector (mistralai/mistral-large-2411)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.552s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.785s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.294s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.582s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.561s |  |
| Basic echo function | ✅ Pass | 1.050s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.523s |  |
| Search query | ✅ Pass | 1.081s |  |
| Ask advice | ❌ Fail | 20.801s | model did not call a function, responded with text: The error message "cannot find package" in Go d\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 3.359s

---

### searcher (openai/gpt-4o-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.621s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.112s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.172s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.866s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.917s |  |
| Basic echo function | ✅ Pass | 1.146s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.889s |  |
| Search query | ✅ Pass | 1.001s |  |
| Ask advice | ✅ Pass | 1.273s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 1.111s

---

### enricher (openai/gpt-4o)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.033s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.705s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.616s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.724s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.932s |  |
| Basic echo function | ✅ Pass | 0.985s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.083s |  |
| Search query | ✅ Pass | 1.026s |  |
| Ask advice | ✅ Pass | 2.109s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 1.135s

---

### coder (anthropic/claude-3.7-sonnet)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.822s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.984s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.393s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.779s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.884s |  |
| Basic echo function | ✅ Pass | 2.109s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.371s |  |
| Search query | ✅ Pass | 1.724s |  |
| Ask advice | ✅ Pass | 3.197s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 1.696s

---

### installer (openai/gpt-4o)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.722s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.780s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.721s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.568s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.868s |  |
| Basic echo function | ✅ Pass | 0.732s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.925s |  |
| Search query | ✅ Pass | 1.048s |  |
| Ask advice | ✅ Pass | 1.320s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 0.965s

---

### pentester (openai/gpt-4o)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.552s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.152s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.064s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.690s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.874s |  |
| Basic echo function | ✅ Pass | 0.785s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.066s |  |
| Search query | ✅ Pass | 3.488s |  |
| Ask advice | ✅ Pass | 1.571s |  |

**Summary**: 9/9 (100.00%) successful tests

**Average latency**: 1.805s

---

