# LLM Agent Testing Report

Generated: Wed, 25 Jun 2025 18:14:55 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | openai/gpt-4.1-mini | false | 15/15 (100.00%) | 1.301s |
| simple_json | openai/gpt-4.1-mini | false | 5/5 (100.00%) | 1.490s |
| agent | openai/o4-mini | true | 15/15 (100.00%) | 5.685s |
| assistant | x-ai/grok-3-beta | false | 15/15 (100.00%) | 1.512s |
| generator | anthropic/claude-3.7-sonnet:thinking | true | 15/15 (100.00%) | 5.204s |
| refiner | google/gemini-2.5-flash-preview:thinking | true | 15/15 (100.00%) | 3.033s |
| adviser | google/gemini-2.5-pro-preview | true | 15/15 (100.00%) | 4.847s |
| reflector | openai/gpt-4.1-mini | false | 15/15 (100.00%) | 1.530s |
| searcher | x-ai/grok-3-mini-beta | true | 15/15 (100.00%) | 3.473s |
| enricher | openai/gpt-4.1-mini | false | 15/15 (100.00%) | 1.408s |
| coder | anthropic/claude-3.7-sonnet:thinking | true | 15/15 (100.00%) | 5.068s |
| installer | google/gemini-2.5-flash-preview:thinking | true | 15/15 (100.00%) | 2.625s |
| pentester | x-ai/grok-3-mini-beta | true | 15/15 (100.00%) | 3.668s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 3.231s

## Detailed Results

### simple (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.714s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.683s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.172s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.647s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.452s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.436s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.919s |  |
| Basic echo function | ✅ Pass | 0.979s |  |
| Streaming Basic echo function | ✅ Pass | 1.099s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.874s |  |
| Streaming JSON response function | ✅ Pass | 1.085s |  |
| Search query | ✅ Pass | 0.833s |  |
| Streaming Search query | ✅ Pass | 0.706s |  |
| Ask advice | ✅ Pass | 3.371s |  |
| Streaming Ask advice | ✅ Pass | 2.545s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.301s

---

### simple_json (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.519s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 1.056s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 0.875s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 1.711s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.291s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.490s

---

### agent (openai/o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.387s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.243s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.060s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 9.683s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.613s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.372s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.231s |  |
| Basic echo function | ✅ Pass | 3.264s |  |
| Streaming Basic echo function | ✅ Pass | 4.116s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 7.350s |  |
| Streaming JSON response function | ✅ Pass | 5.979s |  |
| Search query | ✅ Pass | 3.995s |  |
| Streaming Search query | ✅ Pass | 6.429s |  |
| Ask advice | ✅ Pass | 14.578s |  |
| Streaming Ask advice | ✅ Pass | 11.974s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 5.685s

---

### assistant (x-ai/grok-3-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.915s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.680s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.711s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.567s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.426s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.646s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.686s |  |
| Basic echo function | ✅ Pass | 1.409s |  |
| Streaming Basic echo function | ✅ Pass | 1.259s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.611s |  |
| Streaming JSON response function | ✅ Pass | 0.815s |  |
| Search query | ✅ Pass | 1.699s |  |
| Streaming Search query | ✅ Pass | 1.343s |  |
| Ask advice | ✅ Pass | 3.412s |  |
| Streaming Ask advice | ✅ Pass | 4.506s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.512s

---

### generator (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.612s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.909s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.603s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.233s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.283s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.167s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.947s |  |
| Basic echo function | ✅ Pass | 3.497s |  |
| Streaming Basic echo function | ✅ Pass | 4.374s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.689s |  |
| Streaming JSON response function | ✅ Pass | 4.093s |  |
| Search query | ✅ Pass | 3.229s |  |
| Streaming Search query | ✅ Pass | 4.333s |  |
| Ask advice | ✅ Pass | 16.036s |  |
| Streaming Ask advice | ✅ Pass | 13.052s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 5.204s

---

### refiner (google/gemini-2.5-flash-preview:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.675s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.678s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.662s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.904s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.567s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.695s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.766s |  |
| Basic echo function | ✅ Pass | 1.767s |  |
| Streaming Basic echo function | ✅ Pass | 1.787s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.038s |  |
| Streaming JSON response function | ✅ Pass | 2.201s |  |
| Search query | ✅ Pass | 1.708s |  |
| Streaming Search query | ✅ Pass | 2.203s |  |
| Ask advice | ✅ Pass | 10.467s |  |
| Streaming Ask advice | ✅ Pass | 11.383s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.033s

---

### adviser (google/gemini-2.5-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.061s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.541s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.932s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.113s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.068s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.743s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.252s |  |
| Basic echo function | ✅ Pass | 5.022s |  |
| Streaming Basic echo function | ✅ Pass | 2.883s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.794s |  |
| Streaming JSON response function | ✅ Pass | 4.169s |  |
| Search query | ✅ Pass | 2.639s |  |
| Streaming Search query | ✅ Pass | 2.815s |  |
| Ask advice | ✅ Pass | 12.666s |  |
| Streaming Ask advice | ✅ Pass | 14.007s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.847s

---

### reflector (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.942s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.520s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.067s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.749s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.473s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.256s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.855s |  |
| Basic echo function | ✅ Pass | 1.455s |  |
| Streaming Basic echo function | ✅ Pass | 0.831s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.986s |  |
| Streaming JSON response function | ✅ Pass | 0.776s |  |
| Search query | ✅ Pass | 0.753s |  |
| Streaming Search query | ✅ Pass | 0.733s |  |
| Ask advice | ✅ Pass | 2.883s |  |
| Streaming Ask advice | ✅ Pass | 2.666s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.530s

---

### searcher (x-ai/grok-3-mini-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.316s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.376s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.824s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.076s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.557s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.503s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.674s |  |
| Basic echo function | ✅ Pass | 3.261s |  |
| Streaming Basic echo function | ✅ Pass | 3.272s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.203s |  |
| Streaming JSON response function | ✅ Pass | 5.214s |  |
| Search query | ✅ Pass | 3.767s |  |
| Streaming Search query | ✅ Pass | 3.231s |  |
| Ask advice | ✅ Pass | 6.789s |  |
| Streaming Ask advice | ✅ Pass | 5.030s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.473s

---

### enricher (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.859s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.565s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.181s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.003s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.609s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.579s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.008s |  |
| Basic echo function | ✅ Pass | 1.789s |  |
| Streaming Basic echo function | ✅ Pass | 1.035s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.821s |  |
| Streaming JSON response function | ✅ Pass | 0.671s |  |
| Search query | ✅ Pass | 0.774s |  |
| Streaming Search query | ✅ Pass | 0.743s |  |
| Ask advice | ✅ Pass | 2.727s |  |
| Streaming Ask advice | ✅ Pass | 2.756s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.408s

---

### coder (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.735s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.702s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.962s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.301s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.322s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.809s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.307s |  |
| Basic echo function | ✅ Pass | 3.843s |  |
| Streaming Basic echo function | ✅ Pass | 4.634s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.058s |  |
| Streaming JSON response function | ✅ Pass | 3.345s |  |
| Search query | ✅ Pass | 6.903s |  |
| Streaming Search query | ✅ Pass | 3.638s |  |
| Ask advice | ✅ Pass | 12.688s |  |
| Streaming Ask advice | ✅ Pass | 15.767s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 5.068s

---

### installer (google/gemini-2.5-flash-preview:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.849s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.423s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.969s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.800s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.317s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.250s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.100s |  |
| Basic echo function | ✅ Pass | 1.588s |  |
| Streaming Basic echo function | ✅ Pass | 1.701s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.613s |  |
| Streaming JSON response function | ✅ Pass | 2.157s |  |
| Search query | ✅ Pass | 2.032s |  |
| Streaming Search query | ✅ Pass | 1.934s |  |
| Ask advice | ✅ Pass | 8.892s |  |
| Streaming Ask advice | ✅ Pass | 8.753s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.625s

---

### pentester (x-ai/grok-3-mini-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.964s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.909s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.342s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.109s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.308s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.365s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.270s |  |
| Basic echo function | ✅ Pass | 3.229s |  |
| Streaming Basic echo function | ✅ Pass | 3.219s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.493s |  |
| Streaming JSON response function | ✅ Pass | 3.351s |  |
| Search query | ✅ Pass | 3.671s |  |
| Streaming Search query | ✅ Pass | 3.537s |  |
| Ask advice | ✅ Pass | 6.142s |  |
| Streaming Ask advice | ✅ Pass | 12.117s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.668s

---

