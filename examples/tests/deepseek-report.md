# LLM Agent Testing Report

Generated: Tue, 13 May 2025 22:09:02 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|----------------|
| simple | deepseek-chat | false | 15/15 (100.00%) | 8.492s |
| simple_json | deepseek-chat | false | 5/5 (100.00%) | 7.679s |
| agent | deepseek-chat | false | 15/15 (100.00%) | 7.776s |
| assistant | deepseek-chat | false | 15/15 (100.00%) | 7.701s |
| generator | deepseek-chat | false | 15/15 (100.00%) | 9.046s |
| refiner | deepseek-chat | false | 15/15 (100.00%) | 8.232s |
| adviser | deepseek-chat | false | 15/15 (100.00%) | 7.268s |
| reflector | deepseek-chat | false | 15/15 (100.00%) | 7.412s |
| searcher | deepseek-chat | false | 15/15 (100.00%) | 6.842s |
| enricher | deepseek-chat | false | 15/15 (100.00%) | 7.158s |
| coder | deepseek-coder | false | 15/15 (100.00%) | 7.057s |
| installer | deepseek-coder | false | 15/15 (100.00%) | 7.237s |
| pentester | deepseek-chat | false | 15/15 (100.00%) | 6.858s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 7.592s

## Detailed Results

### simple (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.357s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 8.614s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.678s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 10.025s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.339s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.331s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 8.825s |  |
| Basic echo function | ✅ Pass | 4.560s |  |
| Streaming Basic echo function | ✅ Pass | 5.117s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.356s |  |
| Streaming JSON response function | ✅ Pass | 5.869s |  |
| Search query | ✅ Pass | 6.178s |  |
| Streaming Search query | ✅ Pass | 5.113s |  |
| Ask advice | ✅ Pass | 24.290s |  |
| Streaming Ask advice | ✅ Pass | 23.732s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 8.492s

---

### simple_json (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.023s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 12.408s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 5.795s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 8.478s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 5.692s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 7.679s

---

### agent (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.880s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 7.228s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.212s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.238s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 6.395s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.209s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 7.348s |  |
| Basic echo function | ✅ Pass | 6.422s |  |
| Streaming Basic echo function | ✅ Pass | 6.752s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.589s |  |
| Streaming JSON response function | ✅ Pass | 4.798s |  |
| Search query | ✅ Pass | 5.759s |  |
| Streaming Search query | ✅ Pass | 5.463s |  |
| Ask advice | ✅ Pass | 21.087s |  |
| Streaming Ask advice | ✅ Pass | 20.266s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.776s

---

### assistant (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 9.402s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 5.880s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.812s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.339s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.820s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.945s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.818s |  |
| Basic echo function | ✅ Pass | 4.575s |  |
| Streaming Basic echo function | ✅ Pass | 4.971s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 7.793s |  |
| Streaming JSON response function | ✅ Pass | 5.294s |  |
| Search query | ✅ Pass | 6.497s |  |
| Streaming Search query | ✅ Pass | 5.331s |  |
| Ask advice | ✅ Pass | 20.006s |  |
| Streaming Ask advice | ✅ Pass | 21.034s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.701s

---

### generator (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.931s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.136s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 8.960s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 8.736s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 20.097s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 9.871s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.964s |  |
| Basic echo function | ✅ Pass | 5.398s |  |
| Streaming Basic echo function | ✅ Pass | 5.347s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.416s |  |
| Streaming JSON response function | ✅ Pass | 4.498s |  |
| Search query | ✅ Pass | 5.168s |  |
| Streaming Search query | ✅ Pass | 4.293s |  |
| Ask advice | ✅ Pass | 23.295s |  |
| Streaming Ask advice | ✅ Pass | 19.577s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 9.046s

---

### refiner (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 10.692s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 7.111s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.493s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.377s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.600s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.419s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.855s |  |
| Basic echo function | ✅ Pass | 4.525s |  |
| Streaming Basic echo function | ✅ Pass | 6.239s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 6.237s |  |
| Streaming JSON response function | ✅ Pass | 5.047s |  |
| Search query | ✅ Pass | 5.023s |  |
| Streaming Search query | ✅ Pass | 6.765s |  |
| Ask advice | ✅ Pass | 21.618s |  |
| Streaming Ask advice | ✅ Pass | 24.481s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 8.232s

---

### adviser (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.201s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 5.703s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.969s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.169s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.593s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.094s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.819s |  |
| Basic echo function | ✅ Pass | 7.061s |  |
| Streaming Basic echo function | ✅ Pass | 4.648s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.611s |  |
| Streaming JSON response function | ✅ Pass | 6.368s |  |
| Search query | ✅ Pass | 3.829s |  |
| Streaming Search query | ✅ Pass | 4.613s |  |
| Ask advice | ✅ Pass | 23.527s |  |
| Streaming Ask advice | ✅ Pass | 23.817s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.268s

---

### reflector (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 9.212s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.175s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.001s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.266s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.905s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.784s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.638s |  |
| Basic echo function | ✅ Pass | 3.946s |  |
| Streaming Basic echo function | ✅ Pass | 5.202s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.479s |  |
| Streaming JSON response function | ✅ Pass | 4.549s |  |
| Search query | ✅ Pass | 4.538s |  |
| Streaming Search query | ✅ Pass | 6.594s |  |
| Ask advice | ✅ Pass | 22.585s |  |
| Streaming Ask advice | ✅ Pass | 22.308s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.412s

---

### searcher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.672s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.743s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.852s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.893s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.177s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.376s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.628s |  |
| Basic echo function | ✅ Pass | 4.099s |  |
| Streaming Basic echo function | ✅ Pass | 4.520s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.838s |  |
| Streaming JSON response function | ✅ Pass | 4.512s |  |
| Search query | ✅ Pass | 4.464s |  |
| Streaming Search query | ✅ Pass | 6.105s |  |
| Ask advice | ✅ Pass | 22.923s |  |
| Streaming Ask advice | ✅ Pass | 20.828s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.842s

---

### enricher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 6.575s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 5.599s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.032s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.821s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.742s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.761s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.896s |  |
| Basic echo function | ✅ Pass | 4.455s |  |
| Streaming Basic echo function | ✅ Pass | 4.894s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.072s |  |
| Streaming JSON response function | ✅ Pass | 6.049s |  |
| Search query | ✅ Pass | 4.653s |  |
| Streaming Search query | ✅ Pass | 4.432s |  |
| Ask advice | ✅ Pass | 20.657s |  |
| Streaming Ask advice | ✅ Pass | 21.728s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.158s

---

### coder (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.544s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.537s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.938s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.441s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.057s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.113s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.038s |  |
| Basic echo function | ✅ Pass | 4.679s |  |
| Streaming Basic echo function | ✅ Pass | 4.412s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.504s |  |
| Streaming JSON response function | ✅ Pass | 4.951s |  |
| Search query | ✅ Pass | 5.662s |  |
| Streaming Search query | ✅ Pass | 4.953s |  |
| Ask advice | ✅ Pass | 21.208s |  |
| Streaming Ask advice | ✅ Pass | 23.823s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.057s

---

### installer (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.529s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.046s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 6.532s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.803s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.622s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.109s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.801s |  |
| Basic echo function | ✅ Pass | 4.170s |  |
| Streaming Basic echo function | ✅ Pass | 4.442s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.127s |  |
| Streaming JSON response function | ✅ Pass | 6.177s |  |
| Search query | ✅ Pass | 4.462s |  |
| Streaming Search query | ✅ Pass | 4.369s |  |
| Ask advice | ✅ Pass | 26.696s |  |
| Streaming Ask advice | ✅ Pass | 20.669s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.237s

---

### pentester (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.076s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.367s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.276s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.097s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.005s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.726s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.338s |  |
| Basic echo function | ✅ Pass | 4.961s |  |
| Streaming Basic echo function | ✅ Pass | 4.311s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.260s |  |
| Streaming JSON response function | ✅ Pass | 4.467s |  |
| Search query | ✅ Pass | 5.253s |  |
| Streaming Search query | ✅ Pass | 4.274s |  |
| Ask advice | ✅ Pass | 22.227s |  |
| Streaming Ask advice | ✅ Pass | 21.233s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.858s

---

