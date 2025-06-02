# LLM Agent Testing Report

Generated: Thu, 29 May 2025 14:20:54 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-3-5-haiku-20241022 | false | 15/15 (100.00%) | 2.371s |
| simple_json | claude-3-5-haiku-20241022 | false | 5/5 (100.00%) | 1.555s |
| agent | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 4.102s |
| assistant | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.546s |
| generator | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.352s |
| refiner | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.238s |
| adviser | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.410s |
| reflector | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.676s |
| searcher | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.553s |
| enricher | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.656s |
| coder | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.484s |
| installer | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.363s |
| pentester | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.362s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 3.375s

## Detailed Results

### simple (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.617s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.772s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.853s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.151s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.617s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.691s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.988s |  |
| Basic echo function | ✅ Pass | 1.466s |  |
| Streaming Basic echo function | ✅ Pass | 1.556s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.048s |  |
| Streaming JSON response function | ✅ Pass | 1.908s |  |
| Search query | ✅ Pass | 1.451s |  |
| Streaming Search query | ✅ Pass | 1.454s |  |
| Ask advice | ✅ Pass | 9.427s |  |
| Streaming Ask advice | ✅ Pass | 9.570s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.371s

---

### simple_json (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.700s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 1.686s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 0.945s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 1.903s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.539s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.555s

---

### agent (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 12.515s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.397s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.501s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.870s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.422s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.622s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.880s |  |
| Basic echo function | ✅ Pass | 1.870s |  |
| Streaming Basic echo function | ✅ Pass | 2.144s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.764s |  |
| Streaming JSON response function | ✅ Pass | 2.741s |  |
| Search query | ✅ Pass | 2.430s |  |
| Streaming Search query | ✅ Pass | 2.527s |  |
| Ask advice | ✅ Pass | 12.496s |  |
| Streaming Ask advice | ✅ Pass | 12.346s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.102s

---

### assistant (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.255s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.487s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.549s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.862s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.575s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.288s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.921s |  |
| Basic echo function | ✅ Pass | 2.198s |  |
| Streaming Basic echo function | ✅ Pass | 2.847s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.929s |  |
| Streaming JSON response function | ✅ Pass | 2.595s |  |
| Search query | ✅ Pass | 3.224s |  |
| Streaming Search query | ✅ Pass | 2.681s |  |
| Ask advice | ✅ Pass | 12.546s |  |
| Streaming Ask advice | ✅ Pass | 12.233s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.546s

---

### generator (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.723s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.510s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.765s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.708s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.702s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.860s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.316s |  |
| Basic echo function | ✅ Pass | 1.871s |  |
| Streaming Basic echo function | ✅ Pass | 1.953s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.416s |  |
| Streaming JSON response function | ✅ Pass | 2.757s |  |
| Search query | ✅ Pass | 2.289s |  |
| Streaming Search query | ✅ Pass | 2.727s |  |
| Ask advice | ✅ Pass | 12.493s |  |
| Streaming Ask advice | ✅ Pass | 9.187s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.352s

---

### refiner (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.345s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.914s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.443s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.784s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.711s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.561s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.534s |  |
| Basic echo function | ✅ Pass | 2.139s |  |
| Streaming Basic echo function | ✅ Pass | 2.049s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.435s |  |
| Streaming JSON response function | ✅ Pass | 2.454s |  |
| Search query | ✅ Pass | 2.472s |  |
| Streaming Search query | ✅ Pass | 2.286s |  |
| Ask advice | ✅ Pass | 11.168s |  |
| Streaming Ask advice | ✅ Pass | 10.278s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.238s

---

### adviser (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.953s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.484s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.662s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.758s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.728s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.248s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.863s |  |
| Basic echo function | ✅ Pass | 2.135s |  |
| Streaming Basic echo function | ✅ Pass | 2.077s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.554s |  |
| Streaming JSON response function | ✅ Pass | 2.731s |  |
| Search query | ✅ Pass | 2.691s |  |
| Streaming Search query | ✅ Pass | 2.552s |  |
| Ask advice | ✅ Pass | 13.545s |  |
| Streaming Ask advice | ✅ Pass | 11.165s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.410s

---

### reflector (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.644s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.604s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.660s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.674s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.431s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.241s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.242s |  |
| Basic echo function | ✅ Pass | 2.540s |  |
| Streaming Basic echo function | ✅ Pass | 2.254s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.704s |  |
| Streaming JSON response function | ✅ Pass | 2.801s |  |
| Search query | ✅ Pass | 2.402s |  |
| Streaming Search query | ✅ Pass | 3.050s |  |
| Ask advice | ✅ Pass | 11.931s |  |
| Streaming Ask advice | ✅ Pass | 14.956s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.676s

---

### searcher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.905s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.942s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.520s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.465s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.239s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.352s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.651s |  |
| Basic echo function | ✅ Pass | 2.263s |  |
| Streaming Basic echo function | ✅ Pass | 2.400s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.902s |  |
| Streaming JSON response function | ✅ Pass | 2.816s |  |
| Search query | ✅ Pass | 2.151s |  |
| Streaming Search query | ✅ Pass | 2.435s |  |
| Ask advice | ✅ Pass | 11.784s |  |
| Streaming Ask advice | ✅ Pass | 14.468s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.553s

---

### enricher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.462s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.205s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.546s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.596s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.266s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.759s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.250s |  |
| Basic echo function | ✅ Pass | 2.054s |  |
| Streaming Basic echo function | ✅ Pass | 1.868s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.638s |  |
| Streaming JSON response function | ✅ Pass | 3.378s |  |
| Search query | ✅ Pass | 2.550s |  |
| Streaming Search query | ✅ Pass | 2.442s |  |
| Ask advice | ✅ Pass | 14.031s |  |
| Streaming Ask advice | ✅ Pass | 12.792s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.656s

---

### coder (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.325s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.281s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.387s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.385s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.226s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.569s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.885s |  |
| Basic echo function | ✅ Pass | 2.292s |  |
| Streaming Basic echo function | ✅ Pass | 2.102s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.056s |  |
| Streaming JSON response function | ✅ Pass | 3.173s |  |
| Search query | ✅ Pass | 2.509s |  |
| Streaming Search query | ✅ Pass | 2.403s |  |
| Ask advice | ✅ Pass | 11.993s |  |
| Streaming Ask advice | ✅ Pass | 12.671s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.484s

---

### installer (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.098s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.314s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.622s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.404s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.358s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.807s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.252s |  |
| Basic echo function | ✅ Pass | 2.558s |  |
| Streaming Basic echo function | ✅ Pass | 2.207s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.079s |  |
| Streaming JSON response function | ✅ Pass | 2.685s |  |
| Search query | ✅ Pass | 2.161s |  |
| Streaming Search query | ✅ Pass | 2.360s |  |
| Ask advice | ✅ Pass | 12.899s |  |
| Streaming Ask advice | ✅ Pass | 10.634s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.363s

---

### pentester (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.224s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.331s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.380s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.675s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.555s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.678s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.103s |  |
| Basic echo function | ✅ Pass | 2.266s |  |
| Streaming Basic echo function | ✅ Pass | 2.092s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.843s |  |
| Streaming JSON response function | ✅ Pass | 2.978s |  |
| Search query | ✅ Pass | 2.428s |  |
| Streaming Search query | ✅ Pass | 2.690s |  |
| Ask advice | ✅ Pass | 11.834s |  |
| Streaming Ask advice | ✅ Pass | 11.357s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.362s

---

