# LLM Agent Testing Report

Generated: Tue, 13 May 2025 21:08:11 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|----------------|
| simple | gpt-4.1-mini | false | 15/15 (100.00%) | 1.004s |
| simple_json | gpt-4.1-mini | false | 5/5 (100.00%) | 1.231s |
| agent | o4-mini | true | 15/15 (100.00%) | 2.293s |
| assistant | o4-mini | true | 15/15 (100.00%) | 2.633s |
| generator | o3 | true | 15/15 (100.00%) | 6.114s |
| refiner | gpt-4.1 | false | 15/15 (100.00%) | 1.234s |
| adviser | o4-mini | true | 15/15 (100.00%) | 2.755s |
| reflector | o4-mini | true | 15/15 (100.00%) | 3.107s |
| searcher | gpt-4.1-mini | false | 15/15 (100.00%) | 1.110s |
| enricher | gpt-4.1-mini | false | 15/15 (100.00%) | 1.193s |
| coder | gpt-4.1 | false | 15/15 (100.00%) | 1.124s |
| installer | gpt-4.1 | false | 15/15 (100.00%) | 1.368s |
| pentester | o4-mini | true | 15/15 (100.00%) | 1.778s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 2.118s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.211s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.587s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.404s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.829s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.602s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.573s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.735s |  |
| Basic echo function | ✅ Pass | 0.820s |  |
| Streaming Basic echo function | ✅ Pass | 0.818s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.686s |  |
| Streaming JSON response function | ✅ Pass | 0.776s |  |
| Search query | ✅ Pass | 0.856s |  |
| Streaming Search query | ✅ Pass | 0.595s |  |
| Ask advice | ✅ Pass | 2.504s |  |
| Streaming Ask advice | ✅ Pass | 2.065s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.004s

---

### simple_json (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.388s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 0.912s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.261s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 1.220s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.374s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.231s

---

### agent (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.057s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.698s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.373s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.551s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.143s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.938s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.072s |  |
| Basic echo function | ✅ Pass | 1.208s |  |
| Streaming Basic echo function | ✅ Pass | 1.003s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.664s |  |
| Streaming JSON response function | ✅ Pass | 4.119s |  |
| Search query | ✅ Pass | 2.038s |  |
| Streaming Search query | ✅ Pass | 0.958s |  |
| Ask advice | ✅ Pass | 9.117s |  |
| Streaming Ask advice | ✅ Pass | 4.451s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.293s

---

### assistant (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.182s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.438s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.140s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.900s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.337s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.247s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.320s |  |
| Basic echo function | ✅ Pass | 2.317s |  |
| Streaming Basic echo function | ✅ Pass | 6.095s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.179s |  |
| Streaming JSON response function | ✅ Pass | 1.457s |  |
| Search query | ✅ Pass | 3.097s |  |
| Streaming Search query | ✅ Pass | 1.689s |  |
| Ask advice | ✅ Pass | 5.774s |  |
| Streaming Ask advice | ✅ Pass | 7.329s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.633s

---

### generator (o3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.494s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.598s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.013s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.978s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.455s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.022s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.870s |  |
| Basic echo function | ✅ Pass | 4.588s |  |
| Streaming Basic echo function | ✅ Pass | 4.369s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.806s |  |
| Streaming JSON response function | ✅ Pass | 2.327s |  |
| Search query | ✅ Pass | 2.460s |  |
| Streaming Search query | ✅ Pass | 3.578s |  |
| Ask advice | ✅ Pass | 26.146s |  |
| Streaming Ask advice | ✅ Pass | 28.000s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.114s

---

### refiner (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.635s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.664s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.712s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.531s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.042s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.692s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.632s |  |
| Basic echo function | ✅ Pass | 0.608s |  |
| Streaming Basic echo function | ✅ Pass | 1.385s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.545s |  |
| Streaming JSON response function | ✅ Pass | 0.598s |  |
| Search query | ✅ Pass | 0.737s |  |
| Streaming Search query | ✅ Pass | 1.087s |  |
| Ask advice | ✅ Pass | 2.079s |  |
| Streaming Ask advice | ✅ Pass | 6.556s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.234s

---

### adviser (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.061s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.874s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.699s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.524s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.295s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.499s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.938s |  |
| Basic echo function | ✅ Pass | 1.769s |  |
| Streaming Basic echo function | ✅ Pass | 2.254s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.579s |  |
| Streaming JSON response function | ✅ Pass | 2.966s |  |
| Search query | ✅ Pass | 2.677s |  |
| Streaming Search query | ✅ Pass | 2.274s |  |
| Ask advice | ✅ Pass | 7.057s |  |
| Streaming Ask advice | ✅ Pass | 4.864s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.755s

---

### reflector (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.367s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.894s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.990s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.797s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.351s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.893s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.958s |  |
| Basic echo function | ✅ Pass | 3.056s |  |
| Streaming Basic echo function | ✅ Pass | 2.508s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.676s |  |
| Streaming JSON response function | ✅ Pass | 1.651s |  |
| Search query | ✅ Pass | 2.213s |  |
| Streaming Search query | ✅ Pass | 2.225s |  |
| Ask advice | ✅ Pass | 9.214s |  |
| Streaming Ask advice | ✅ Pass | 8.809s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.107s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.568s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.508s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.572s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.304s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.602s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.592s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.097s |  |
| Basic echo function | ✅ Pass | 0.951s |  |
| Streaming Basic echo function | ✅ Pass | 0.876s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.674s |  |
| Streaming JSON response function | ✅ Pass | 1.397s |  |
| Search query | ✅ Pass | 0.972s |  |
| Streaming Search query | ✅ Pass | 0.656s |  |
| Ask advice | ✅ Pass | 2.860s |  |
| Streaming Ask advice | ✅ Pass | 2.016s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.110s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.221s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.664s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.064s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.231s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.920s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.254s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.211s |  |
| Basic echo function | ✅ Pass | 0.778s |  |
| Streaming Basic echo function | ✅ Pass | 0.795s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.907s |  |
| Streaming JSON response function | ✅ Pass | 0.636s |  |
| Search query | ✅ Pass | 0.903s |  |
| Streaming Search query | ✅ Pass | 0.922s |  |
| Ask advice | ✅ Pass | 2.183s |  |
| Streaming Ask advice | ✅ Pass | 3.203s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.193s

---

### coder (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.612s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.603s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.489s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.525s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.684s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.853s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.593s |  |
| Basic echo function | ✅ Pass | 0.598s |  |
| Streaming Basic echo function | ✅ Pass | 0.811s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.523s |  |
| Streaming JSON response function | ✅ Pass | 0.552s |  |
| Search query | ✅ Pass | 0.679s |  |
| Streaming Search query | ✅ Pass | 0.719s |  |
| Ask advice | ✅ Pass | 3.391s |  |
| Streaming Ask advice | ✅ Pass | 5.221s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.124s

---

### installer (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.065s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.618s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.455s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.483s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.667s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.662s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.644s |  |
| Basic echo function | ✅ Pass | 0.582s |  |
| Streaming Basic echo function | ✅ Pass | 0.677s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.580s |  |
| Streaming JSON response function | ✅ Pass | 0.788s |  |
| Search query | ✅ Pass | 1.133s |  |
| Streaming Search query | ✅ Pass | 0.634s |  |
| Ask advice | ✅ Pass | 6.363s |  |
| Streaming Ask advice | ✅ Pass | 5.170s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.368s

---

### pentester (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.955s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.429s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.831s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.423s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.150s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.771s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.319s |  |
| Basic echo function | ✅ Pass | 1.412s |  |
| Streaming Basic echo function | ✅ Pass | 2.037s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.388s |  |
| Streaming JSON response function | ✅ Pass | 1.465s |  |
| Search query | ✅ Pass | 1.442s |  |
| Streaming Search query | ✅ Pass | 1.296s |  |
| Ask advice | ✅ Pass | 5.329s |  |
| Streaming Ask advice | ✅ Pass | 3.420s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.778s

---

