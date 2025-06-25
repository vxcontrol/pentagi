# LLM Agent Testing Report

Generated: Wed, 25 Jun 2025 19:12:06 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-3-5-haiku-20241022 | false | 15/15 (100.00%) | 2.404s |
| simple_json | claude-3-5-haiku-20241022 | false | 5/5 (100.00%) | 1.943s |
| agent | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.363s |
| assistant | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.392s |
| generator | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.311s |
| refiner | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.704s |
| adviser | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.429s |
| reflector | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.334s |
| searcher | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.451s |
| enricher | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.274s |
| coder | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.466s |
| installer | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.315s |
| pentester | claude-sonnet-4-20250514 | false | 15/15 (100.00%) | 3.615s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 3.300s

## Detailed Results

### simple (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.077s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.182s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.416s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.787s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.942s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.658s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.942s |  |
| Basic echo function | ✅ Pass | 1.471s |  |
| Streaming Basic echo function | ✅ Pass | 1.834s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.927s |  |
| Streaming JSON response function | ✅ Pass | 2.260s |  |
| Search query | ✅ Pass | 1.952s |  |
| Streaming Search query | ✅ Pass | 1.498s |  |
| Ask advice | ✅ Pass | 7.834s |  |
| Streaming Ask advice | ✅ Pass | 7.274s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.404s

---

### simple_json (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.672s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 1.935s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.024s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 2.239s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.847s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.943s

---

### agent (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.446s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.505s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.811s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.919s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.437s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.781s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.578s |  |
| Basic echo function | ✅ Pass | 2.163s |  |
| Streaming Basic echo function | ✅ Pass | 2.716s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.514s |  |
| Streaming JSON response function | ✅ Pass | 2.360s |  |
| Search query | ✅ Pass | 2.376s |  |
| Streaming Search query | ✅ Pass | 2.467s |  |
| Ask advice | ✅ Pass | 11.082s |  |
| Streaming Ask advice | ✅ Pass | 10.283s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.363s

---

### assistant (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.474s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.792s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.819s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.518s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.524s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.650s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.801s |  |
| Basic echo function | ✅ Pass | 2.286s |  |
| Streaming Basic echo function | ✅ Pass | 2.752s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.331s |  |
| Streaming JSON response function | ✅ Pass | 2.393s |  |
| Search query | ✅ Pass | 2.732s |  |
| Streaming Search query | ✅ Pass | 1.995s |  |
| Ask advice | ✅ Pass | 9.755s |  |
| Streaming Ask advice | ✅ Pass | 13.062s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.392s

---

### generator (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.290s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.288s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.041s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.628s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.716s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.703s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.306s |  |
| Basic echo function | ✅ Pass | 2.342s |  |
| Streaming Basic echo function | ✅ Pass | 2.788s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.883s |  |
| Streaming JSON response function | ✅ Pass | 2.604s |  |
| Search query | ✅ Pass | 2.338s |  |
| Streaming Search query | ✅ Pass | 2.357s |  |
| Ask advice | ✅ Pass | 9.968s |  |
| Streaming Ask advice | ✅ Pass | 9.411s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.311s

---

### refiner (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.015s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.895s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.570s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.415s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.115s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.934s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.952s |  |
| Basic echo function | ✅ Pass | 2.216s |  |
| Streaming Basic echo function | ✅ Pass | 2.137s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.512s |  |
| Streaming JSON response function | ✅ Pass | 2.992s |  |
| Search query | ✅ Pass | 2.373s |  |
| Streaming Search query | ✅ Pass | 2.584s |  |
| Ask advice | ✅ Pass | 10.510s |  |
| Streaming Ask advice | ✅ Pass | 15.337s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.704s

---

### adviser (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.565s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.652s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.739s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.668s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.524s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.662s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.882s |  |
| Basic echo function | ✅ Pass | 2.221s |  |
| Streaming Basic echo function | ✅ Pass | 2.474s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.136s |  |
| Streaming JSON response function | ✅ Pass | 3.020s |  |
| Search query | ✅ Pass | 3.064s |  |
| Streaming Search query | ✅ Pass | 3.428s |  |
| Ask advice | ✅ Pass | 11.390s |  |
| Streaming Ask advice | ✅ Pass | 10.004s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.429s

---

### reflector (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.499s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.652s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.771s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.110s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.634s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.728s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.839s |  |
| Basic echo function | ✅ Pass | 3.284s |  |
| Streaming Basic echo function | ✅ Pass | 2.630s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.965s |  |
| Streaming JSON response function | ✅ Pass | 2.761s |  |
| Search query | ✅ Pass | 2.484s |  |
| Streaming Search query | ✅ Pass | 2.682s |  |
| Ask advice | ✅ Pass | 10.148s |  |
| Streaming Ask advice | ✅ Pass | 8.824s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.334s

---

### searcher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.096s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.706s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.537s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.555s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.015s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.085s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.078s |  |
| Basic echo function | ✅ Pass | 2.649s |  |
| Streaming Basic echo function | ✅ Pass | 2.873s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.376s |  |
| Streaming JSON response function | ✅ Pass | 2.465s |  |
| Search query | ✅ Pass | 2.696s |  |
| Streaming Search query | ✅ Pass | 2.167s |  |
| Ask advice | ✅ Pass | 11.414s |  |
| Streaming Ask advice | ✅ Pass | 11.047s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.451s

---

### enricher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.521s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.639s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.090s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.455s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.460s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.718s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.510s |  |
| Basic echo function | ✅ Pass | 1.972s |  |
| Streaming Basic echo function | ✅ Pass | 2.160s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.331s |  |
| Streaming JSON response function | ✅ Pass | 3.460s |  |
| Search query | ✅ Pass | 2.485s |  |
| Streaming Search query | ✅ Pass | 2.858s |  |
| Ask advice | ✅ Pass | 10.026s |  |
| Streaming Ask advice | ✅ Pass | 10.420s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.274s

---

### coder (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.718s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.621s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.584s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.626s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.452s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.816s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.210s |  |
| Basic echo function | ✅ Pass | 2.314s |  |
| Streaming Basic echo function | ✅ Pass | 2.630s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.572s |  |
| Streaming JSON response function | ✅ Pass | 2.504s |  |
| Search query | ✅ Pass | 3.150s |  |
| Streaming Search query | ✅ Pass | 2.258s |  |
| Ask advice | ✅ Pass | 10.898s |  |
| Streaming Ask advice | ✅ Pass | 12.634s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.466s

---

### installer (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.677s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.720s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.854s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.766s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.626s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.402s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.182s |  |
| Basic echo function | ✅ Pass | 2.183s |  |
| Streaming Basic echo function | ✅ Pass | 2.583s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.601s |  |
| Streaming JSON response function | ✅ Pass | 2.671s |  |
| Search query | ✅ Pass | 2.495s |  |
| Streaming Search query | ✅ Pass | 2.541s |  |
| Ask advice | ✅ Pass | 10.043s |  |
| Streaming Ask advice | ✅ Pass | 10.383s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.315s

---

### pentester (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.857s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.513s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.042s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.461s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.793s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.865s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.968s |  |
| Basic echo function | ✅ Pass | 2.412s |  |
| Streaming Basic echo function | ✅ Pass | 2.307s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.442s |  |
| Streaming JSON response function | ✅ Pass | 2.682s |  |
| Search query | ✅ Pass | 2.941s |  |
| Streaming Search query | ✅ Pass | 4.464s |  |
| Ask advice | ✅ Pass | 10.958s |  |
| Streaming Ask advice | ✅ Pass | 12.519s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.615s

---

