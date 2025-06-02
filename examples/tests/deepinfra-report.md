# LLM Agent Testing Report

Generated: Tue, 13 May 2025 21:54:21 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|----------------|
| simple | meta-llama/Llama-3.3-70B-Instruct | false | 15/15 (100.00%) | 4.696s |
| simple_json | google/gemma-3-27b-it | false | 5/5 (100.00%) | 1.763s |
| agent | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.901s |
| assistant | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.868s |
| generator | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.872s |
| refiner | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.977s |
| adviser | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.744s |
| reflector | Qwen/Qwen3-30B-A3B | true | 15/15 (100.00%) | 3.675s |
| searcher | nvidia/Llama-3.1-Nemotron-70B-Instruct | false | 15/15 (100.00%) | 2.511s |
| enricher | nvidia/Llama-3.1-Nemotron-70B-Instruct | false | 15/15 (100.00%) | 2.233s |
| coder | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.178s |
| installer | nvidia/Llama-3.1-Nemotron-70B-Instruct | false | 15/15 (100.00%) | 1.811s |
| pentester | deepseek-ai/DeepSeek-R1 | true | 15/15 (100.00%) | 4.078s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 3.903s

## Detailed Results

### simple (meta-llama/Llama-3.3-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.207s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 5.013s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.575s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.537s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.434s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.597s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.750s |  |
| Basic echo function | ✅ Pass | 3.060s |  |
| Streaming Basic echo function | ✅ Pass | 1.836s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.377s |  |
| Streaming JSON response function | ✅ Pass | 1.108s |  |
| Search query | ✅ Pass | 5.002s |  |
| Streaming Search query | ✅ Pass | 2.077s |  |
| Ask advice | ✅ Pass | 7.682s |  |
| Streaming Ask advice | ✅ Pass | 18.191s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.696s

---

### simple_json (google/gemma-3-27b-it)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.839s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 0.984s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.126s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 3.517s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.350s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.763s

---

### agent (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.883s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.278s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.285s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.711s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.149s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.573s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.691s |  |
| Basic echo function | ✅ Pass | 3.564s |  |
| Streaming Basic echo function | ✅ Pass | 5.810s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.886s |  |
| Streaming JSON response function | ✅ Pass | 4.291s |  |
| Search query | ✅ Pass | 2.969s |  |
| Streaming Search query | ✅ Pass | 3.069s |  |
| Ask advice | ✅ Pass | 10.211s |  |
| Streaming Ask advice | ✅ Pass | 8.149s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.901s

---

### assistant (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 6.733s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.284s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.380s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.986s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.801s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.331s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 7.797s |  |
| Basic echo function | ✅ Pass | 3.554s |  |
| Streaming Basic echo function | ✅ Pass | 2.941s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.572s |  |
| Streaming JSON response function | ✅ Pass | 3.306s |  |
| Search query | ✅ Pass | 5.595s |  |
| Streaming Search query | ✅ Pass | 2.817s |  |
| Ask advice | ✅ Pass | 8.059s |  |
| Streaming Ask advice | ✅ Pass | 5.858s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.868s

---

### generator (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 7.603s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.531s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.492s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.127s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.103s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.210s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.567s |  |
| Basic echo function | ✅ Pass | 3.267s |  |
| Streaming Basic echo function | ✅ Pass | 3.036s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.321s |  |
| Streaming JSON response function | ✅ Pass | 3.334s |  |
| Search query | ✅ Pass | 4.450s |  |
| Streaming Search query | ✅ Pass | 2.882s |  |
| Ask advice | ✅ Pass | 7.805s |  |
| Streaming Ask advice | ✅ Pass | 8.349s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.872s

---

### refiner (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 7.581s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.761s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.617s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.287s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.103s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.742s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 7.255s |  |
| Basic echo function | ✅ Pass | 2.937s |  |
| Streaming Basic echo function | ✅ Pass | 3.324s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.972s |  |
| Streaming JSON response function | ✅ Pass | 3.933s |  |
| Search query | ✅ Pass | 5.614s |  |
| Streaming Search query | ✅ Pass | 3.374s |  |
| Ask advice | ✅ Pass | 6.770s |  |
| Streaming Ask advice | ✅ Pass | 8.382s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.977s

---

### adviser (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.734s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.725s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.260s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.191s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.869s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.379s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.916s |  |
| Basic echo function | ✅ Pass | 2.861s |  |
| Streaming Basic echo function | ✅ Pass | 3.083s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.709s |  |
| Streaming JSON response function | ✅ Pass | 3.079s |  |
| Search query | ✅ Pass | 3.190s |  |
| Streaming Search query | ✅ Pass | 3.992s |  |
| Ask advice | ✅ Pass | 10.170s |  |
| Streaming Ask advice | ✅ Pass | 7.000s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.744s

---

### reflector (Qwen/Qwen3-30B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.270s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.078s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.575s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.610s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.475s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.482s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.734s |  |
| Basic echo function | ✅ Pass | 2.405s |  |
| Streaming Basic echo function | ✅ Pass | 2.338s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.804s |  |
| Streaming JSON response function | ✅ Pass | 2.940s |  |
| Search query | ✅ Pass | 3.291s |  |
| Streaming Search query | ✅ Pass | 1.801s |  |
| Ask advice | ✅ Pass | 9.303s |  |
| Streaming Ask advice | ✅ Pass | 5.021s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.675s

---

### searcher (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 0.978s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.351s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.570s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.569s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.957s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.308s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.699s |  |
| Basic echo function | ✅ Pass | 0.892s |  |
| Streaming Basic echo function | ✅ Pass | 0.648s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.481s |  |
| Streaming JSON response function | ✅ Pass | 0.708s |  |
| Search query | ✅ Pass | 1.284s |  |
| Streaming Search query | ✅ Pass | 0.627s |  |
| Ask advice | ✅ Pass | 12.329s |  |
| Streaming Ask advice | ✅ Pass | 14.270s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.511s

---

### enricher (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.061s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.341s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.540s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.534s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.048s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.309s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.605s |  |
| Basic echo function | ✅ Pass | 0.803s |  |
| Streaming Basic echo function | ✅ Pass | 2.939s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.923s |  |
| Streaming JSON response function | ✅ Pass | 0.780s |  |
| Search query | ✅ Pass | 0.637s |  |
| Streaming Search query | ✅ Pass | 0.631s |  |
| Ask advice | ✅ Pass | 14.236s |  |
| Streaming Ask advice | ✅ Pass | 7.110s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.233s

---

### coder (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.221s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.597s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.668s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.561s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.838s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.446s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.576s |  |
| Basic echo function | ✅ Pass | 4.222s |  |
| Streaming Basic echo function | ✅ Pass | 3.081s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.393s |  |
| Streaming JSON response function | ✅ Pass | 2.919s |  |
| Search query | ✅ Pass | 3.856s |  |
| Streaming Search query | ✅ Pass | 2.591s |  |
| Ask advice | ✅ Pass | 6.366s |  |
| Streaming Ask advice | ✅ Pass | 9.330s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.178s

---

### installer (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.051s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.336s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.570s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.531s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.966s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.314s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.580s |  |
| Basic echo function | ✅ Pass | 0.662s |  |
| Streaming Basic echo function | ✅ Pass | 1.289s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.003s |  |
| Streaming JSON response function | ✅ Pass | 0.980s |  |
| Search query | ✅ Pass | 0.656s |  |
| Streaming Search query | ✅ Pass | 0.665s |  |
| Ask advice | ✅ Pass | 9.815s |  |
| Streaming Ask advice | ✅ Pass | 5.741s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.811s

---

### pentester (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.737s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.109s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.939s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.602s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.938s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.850s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.396s |  |
| Basic echo function | ✅ Pass | 2.951s |  |
| Streaming Basic echo function | ✅ Pass | 3.104s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.728s |  |
| Streaming JSON response function | ✅ Pass | 2.776s |  |
| Search query | ✅ Pass | 3.291s |  |
| Streaming Search query | ✅ Pass | 2.951s |  |
| Ask advice | ✅ Pass | 7.132s |  |
| Streaming Ask advice | ✅ Pass | 7.671s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.078s

---

