# LLM Agent Testing Report

Generated: Wed, 25 Jun 2025 19:05:19 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | meta-llama/Llama-3.3-70B-Instruct | false | 13/15 (86.67%) | 1.626s |
| simple_json | google/gemma-3-27b-it | false | 5/5 (100.00%) | 2.473s |
| agent | deepseek-ai/DeepSeek-R1 | true | 14/15 (93.33%) | 3.811s |
| assistant | deepseek-ai/DeepSeek-R1 | true | 12/15 (80.00%) | 3.440s |
| generator | deepseek-ai/DeepSeek-R1 | true | 13/15 (86.67%) | 3.246s |
| refiner | deepseek-ai/DeepSeek-R1 | true | 12/15 (80.00%) | 3.338s |
| adviser | deepseek-ai/DeepSeek-R1 | true | 13/15 (86.67%) | 3.344s |
| reflector | Qwen/Qwen3-30B-A3B | true | 15/15 (100.00%) | 2.543s |
| searcher | nvidia/Llama-3.1-Nemotron-70B-Instruct | false | 15/15 (100.00%) | 2.205s |
| enricher | nvidia/Llama-3.1-Nemotron-70B-Instruct | false | 15/15 (100.00%) | 1.546s |
| coder | deepseek-ai/DeepSeek-R1 | true | 13/15 (86.67%) | 3.349s |
| installer | nvidia/Llama-3.1-Nemotron-70B-Instruct | false | 15/15 (100.00%) | 2.003s |
| pentester | deepseek-ai/DeepSeek-R1 | true | 13/15 (86.67%) | 3.416s |

**Total**: 168/185 (90.81%) successful tests
**Overall average latency**: 2.813s

## Detailed Results

### simple (meta-llama/Llama-3.3-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.650s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.719s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.920s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.631s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.414s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.283s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.033s |  |
| Basic echo function | ✅ Pass | 0.949s |  |
| Streaming Basic echo function | ✅ Pass | 1.448s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.069s |  |
| Streaming JSON response function | ✅ Pass | 1.293s |  |
| Search query | ❌ Fail | 1.810s | model did not call a function, responded with text: <function=search\{"query": "Golang programming l\.\.\. |
| Streaming Search query | ❌ Fail | 1.368s | model did not call a function, responded with text:  |
| Ask advice | ✅ Pass | 2.627s |  |
| Streaming Ask advice | ✅ Pass | 3.175s |  |

**Summary**: 13/15 (86.67%) successful tests

**Average latency**: 1.626s

---

### simple_json (google/gemma-3-27b-it)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.680s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 1.227s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.597s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 5.026s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.834s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 2.473s

---

### agent (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 8.110s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 5.708s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.717s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.936s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.116s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.535s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.737s |  |
| Basic echo function | ✅ Pass | 3.266s |  |
| Streaming Basic echo function | ✅ Pass | 1.630s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.872s |  |
| Streaming JSON response function | ✅ Pass | 2.833s |  |
| Search query | ✅ Pass | 3.552s |  |
| Streaming Search query | ❌ Fail | 2.757s | model did not call a function, responded with text:  |
| Ask advice | ✅ Pass | 3.820s |  |
| Streaming Ask advice | ✅ Pass | 3.578s |  |

**Summary**: 14/15 (93.33%) successful tests

**Average latency**: 3.811s

---

### assistant (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.398s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.757s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.418s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.851s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.938s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.492s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.503s |  |
| Basic echo function | ✅ Pass | 2.071s |  |
| Streaming Basic echo function | ❌ Fail | 2.611s | model did not call a function, responded with text:  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.978s |  |
| Streaming JSON response function | ❌ Fail | 1.739s | model did not call a function, responded with text:  |
| Search query | ✅ Pass | 3.281s |  |
| Streaming Search query | ❌ Fail | 2.067s | model did not call a function, responded with text:  |
| Ask advice | ✅ Pass | 5.486s |  |
| Streaming Ask advice | ✅ Pass | 3.017s |  |

**Summary**: 12/15 (80.00%) successful tests

**Average latency**: 3.440s

---

### generator (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.505s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.631s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.370s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.156s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.608s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.392s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.633s |  |
| Basic echo function | ✅ Pass | 2.892s |  |
| Streaming Basic echo function | ❌ Fail | 2.455s | model did not call a function, responded with text:  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.722s |  |
| Streaming JSON response function | ✅ Pass | 2.493s |  |
| Search query | ✅ Pass | 2.210s |  |
| Streaming Search query | ✅ Pass | 1.529s |  |
| Ask advice | ✅ Pass | 6.756s |  |
| Streaming Ask advice | ❌ Fail | 3.334s | model did not call a function, responded with text:  |

**Summary**: 13/15 (86.67%) successful tests

**Average latency**: 3.246s

---

### refiner (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.411s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.529s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.907s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 10.789s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.251s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.526s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.782s |  |
| Basic echo function | ✅ Pass | 2.276s |  |
| Streaming Basic echo function | ✅ Pass | 2.828s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.579s |  |
| Streaming JSON response function | ❌ Fail | 2.316s | model did not call a function, responded with text:  |
| Search query | ✅ Pass | 3.283s |  |
| Streaming Search query | ❌ Fail | 2.233s | model did not call a function, responded with text:  |
| Ask advice | ✅ Pass | 3.844s |  |
| Streaming Ask advice | ❌ Fail | 3.521s | model did not call a function, responded with text:  |

**Summary**: 12/15 (80.00%) successful tests

**Average latency**: 3.338s

---

### adviser (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.677s |  |
| Completion: Write 'Hello World' in uppe... | ❌ Fail | 1.596s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.913s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.709s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.155s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.911s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.476s |  |
| Basic echo function | ✅ Pass | 2.112s |  |
| Streaming Basic echo function | ❌ Fail | 2.041s | model did not call a function, responded with text:  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.873s |  |
| Streaming JSON response function | ✅ Pass | 3.544s |  |
| Search query | ✅ Pass | 2.221s |  |
| Streaming Search query | ✅ Pass | 5.340s |  |
| Ask advice | ✅ Pass | 6.239s |  |
| Streaming Ask advice | ✅ Pass | 2.359s |  |

**Summary**: 13/15 (86.67%) successful tests

**Average latency**: 3.344s

---

### reflector (Qwen/Qwen3-30B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.867s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.097s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.187s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.535s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.275s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.024s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.422s |  |
| Basic echo function | ✅ Pass | 1.777s |  |
| Streaming Basic echo function | ✅ Pass | 2.907s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.693s |  |
| Streaming JSON response function | ✅ Pass | 2.207s |  |
| Search query | ✅ Pass | 1.805s |  |
| Streaming Search query | ✅ Pass | 1.429s |  |
| Ask advice | ✅ Pass | 3.402s |  |
| Streaming Ask advice | ✅ Pass | 2.523s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.543s

---

### searcher (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.256s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.327s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.664s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.565s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.924s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.411s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.076s |  |
| Basic echo function | ✅ Pass | 0.722s |  |
| Streaming Basic echo function | ✅ Pass | 2.553s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.850s |  |
| Streaming JSON response function | ✅ Pass | 1.210s |  |
| Search query | ✅ Pass | 0.886s |  |
| Streaming Search query | ✅ Pass | 0.932s |  |
| Ask advice | ✅ Pass | 9.887s |  |
| Streaming Ask advice | ✅ Pass | 9.816s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.205s

---

### enricher (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.402s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.333s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.601s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.591s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.898s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.424s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.978s |  |
| Basic echo function | ✅ Pass | 0.698s |  |
| Streaming Basic echo function | ✅ Pass | 0.614s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.140s |  |
| Streaming JSON response function | ✅ Pass | 1.123s |  |
| Search query | ✅ Pass | 1.074s |  |
| Streaming Search query | ✅ Pass | 0.874s |  |
| Ask advice | ✅ Pass | 8.629s |  |
| Streaming Ask advice | ✅ Pass | 2.806s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.546s

---

### coder (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.580s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.038s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.227s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.578s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.790s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.843s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.663s |  |
| Basic echo function | ✅ Pass | 2.374s |  |
| Streaming Basic echo function | ❌ Fail | 1.371s | model did not call a function, responded with text:  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.142s |  |
| Streaming JSON response function | ❌ Fail | 1.901s | model did not call a function, responded with text:  |
| Search query | ✅ Pass | 2.458s |  |
| Streaming Search query | ✅ Pass | 1.386s |  |
| Ask advice | ✅ Pass | 5.244s |  |
| Streaming Ask advice | ✅ Pass | 3.641s |  |

**Summary**: 13/15 (86.67%) successful tests

**Average latency**: 3.349s

---

### installer (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.374s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.327s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.583s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.215s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.428s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.505s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.964s |  |
| Basic echo function | ✅ Pass | 0.673s |  |
| Streaming Basic echo function | ✅ Pass | 0.640s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.996s |  |
| Streaming JSON response function | ✅ Pass | 1.124s |  |
| Search query | ✅ Pass | 1.177s |  |
| Streaming Search query | ✅ Pass | 0.761s |  |
| Ask advice | ✅ Pass | 15.138s |  |
| Streaming Ask advice | ✅ Pass | 3.135s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.003s

---

### pentester (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.928s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.348s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.562s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.357s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.745s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.071s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.409s |  |
| Basic echo function | ✅ Pass | 1.964s |  |
| Streaming Basic echo function | ✅ Pass | 3.534s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.912s |  |
| Streaming JSON response function | ❌ Fail | 1.937s | model did not call a function, responded with text:  |
| Search query | ✅ Pass | 3.197s |  |
| Streaming Search query | ✅ Pass | 2.321s |  |
| Ask advice | ✅ Pass | 3.926s |  |
| Streaming Ask advice | ❌ Fail | 5.023s | model did not call a function, responded with text:  |

**Summary**: 13/15 (86.67%) successful tests

**Average latency**: 3.416s

---

