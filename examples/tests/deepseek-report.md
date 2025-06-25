# LLM Agent Testing Report

Generated: Wed, 25 Jun 2025 18:44:18 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek-chat | false | 15/15 (100.00%) | 7.121s |
| simple_json | deepseek-chat | false | 5/5 (100.00%) | 5.988s |
| agent | deepseek-chat | false | 15/15 (100.00%) | 6.544s |
| assistant | deepseek-chat | false | 15/15 (100.00%) | 7.621s |
| generator | deepseek-chat | false | 15/15 (100.00%) | 6.856s |
| refiner | deepseek-chat | false | 15/15 (100.00%) | 8.249s |
| adviser | deepseek-chat | false | 15/15 (100.00%) | 6.978s |
| reflector | deepseek-chat | false | 15/15 (100.00%) | 6.784s |
| searcher | deepseek-chat | false | 15/15 (100.00%) | 6.995s |
| enricher | deepseek-chat | false | 15/15 (100.00%) | 7.240s |
| coder | deepseek-coder | false | 15/15 (100.00%) | 8.083s |
| installer | deepseek-coder | false | 15/15 (100.00%) | 6.007s |
| pentester | deepseek-chat | false | 15/15 (100.00%) | 6.756s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 7.073s

## Detailed Results

### simple (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.857s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.133s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.797s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.376s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.403s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.541s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.647s |  |
| Basic echo function | ✅ Pass | 4.817s |  |
| Streaming Basic echo function | ✅ Pass | 4.871s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.807s |  |
| Streaming JSON response function | ✅ Pass | 5.578s |  |
| Search query | ✅ Pass | 4.941s |  |
| Streaming Search query | ✅ Pass | 4.981s |  |
| Ask advice | ✅ Pass | 19.863s |  |
| Streaming Ask advice | ✅ Pass | 23.196s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.121s

---

### simple_json (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.125s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 4.867s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 4.557s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 7.543s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 6.849s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 5.988s

---

### agent (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.586s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.694s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.106s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.113s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.972s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.795s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.374s |  |
| Basic echo function | ✅ Pass | 4.504s |  |
| Streaming Basic echo function | ✅ Pass | 4.805s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.035s |  |
| Streaming JSON response function | ✅ Pass | 4.364s |  |
| Search query | ✅ Pass | 4.885s |  |
| Streaming Search query | ✅ Pass | 5.360s |  |
| Ask advice | ✅ Pass | 17.217s |  |
| Streaming Ask advice | ✅ Pass | 20.346s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.544s

---

### assistant (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.694s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.748s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.145s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.509s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 18.962s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.112s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.496s |  |
| Basic echo function | ✅ Pass | 4.424s |  |
| Streaming Basic echo function | ✅ Pass | 4.455s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.729s |  |
| Streaming JSON response function | ✅ Pass | 5.114s |  |
| Search query | ✅ Pass | 4.224s |  |
| Streaming Search query | ✅ Pass | 4.710s |  |
| Ask advice | ✅ Pass | 18.864s |  |
| Streaming Ask advice | ✅ Pass | 20.131s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.621s

---

### generator (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.882s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.740s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.289s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.913s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.825s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.131s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.554s |  |
| Basic echo function | ✅ Pass | 4.857s |  |
| Streaming Basic echo function | ✅ Pass | 4.563s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.085s |  |
| Streaming JSON response function | ✅ Pass | 4.902s |  |
| Search query | ✅ Pass | 5.616s |  |
| Streaming Search query | ✅ Pass | 4.530s |  |
| Ask advice | ✅ Pass | 24.792s |  |
| Streaming Ask advice | ✅ Pass | 17.160s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.856s

---

### refiner (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.190s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.258s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.150s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.286s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 21.175s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.139s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.804s |  |
| Basic echo function | ✅ Pass | 4.835s |  |
| Streaming Basic echo function | ✅ Pass | 5.526s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.517s |  |
| Streaming JSON response function | ✅ Pass | 5.579s |  |
| Search query | ✅ Pass | 5.080s |  |
| Streaming Search query | ✅ Pass | 4.866s |  |
| Ask advice | ✅ Pass | 23.834s |  |
| Streaming Ask advice | ✅ Pass | 19.503s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 8.249s

---

### adviser (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.906s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.394s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.012s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.496s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.496s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.592s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.985s |  |
| Basic echo function | ✅ Pass | 5.791s |  |
| Streaming Basic echo function | ✅ Pass | 5.101s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.952s |  |
| Streaming JSON response function | ✅ Pass | 4.447s |  |
| Search query | ✅ Pass | 6.985s |  |
| Streaming Search query | ✅ Pass | 4.975s |  |
| Ask advice | ✅ Pass | 20.933s |  |
| Streaming Ask advice | ✅ Pass | 19.612s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.978s

---

### reflector (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.124s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.175s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.306s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.851s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.683s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.742s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.948s |  |
| Basic echo function | ✅ Pass | 4.425s |  |
| Streaming Basic echo function | ✅ Pass | 4.772s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.017s |  |
| Streaming JSON response function | ✅ Pass | 4.587s |  |
| Search query | ✅ Pass | 4.425s |  |
| Streaming Search query | ✅ Pass | 5.572s |  |
| Ask advice | ✅ Pass | 20.255s |  |
| Streaming Ask advice | ✅ Pass | 18.876s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.784s

---

### searcher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.602s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.198s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.046s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.195s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.499s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.181s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.719s |  |
| Basic echo function | ✅ Pass | 4.593s |  |
| Streaming Basic echo function | ✅ Pass | 5.097s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.936s |  |
| Streaming JSON response function | ✅ Pass | 5.204s |  |
| Search query | ✅ Pass | 5.100s |  |
| Streaming Search query | ✅ Pass | 4.907s |  |
| Ask advice | ✅ Pass | 20.505s |  |
| Streaming Ask advice | ✅ Pass | 23.139s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.995s

---

### enricher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.576s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.349s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.432s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.807s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 7.246s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.331s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.065s |  |
| Basic echo function | ✅ Pass | 4.445s |  |
| Streaming Basic echo function | ✅ Pass | 4.834s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.691s |  |
| Streaming JSON response function | ✅ Pass | 5.008s |  |
| Search query | ✅ Pass | 4.529s |  |
| Streaming Search query | ✅ Pass | 4.964s |  |
| Ask advice | ✅ Pass | 21.274s |  |
| Streaming Ask advice | ✅ Pass | 23.045s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 7.240s

---

### coder (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.994s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.413s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.209s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 5.175s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.633s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.040s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 20.881s |  |
| Basic echo function | ✅ Pass | 6.351s |  |
| Streaming Basic echo function | ✅ Pass | 4.521s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.808s |  |
| Streaming JSON response function | ✅ Pass | 4.876s |  |
| Search query | ✅ Pass | 5.230s |  |
| Streaming Search query | ✅ Pass | 6.334s |  |
| Ask advice | ✅ Pass | 20.241s |  |
| Streaming Ask advice | ✅ Pass | 20.544s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 8.083s

---

### installer (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 5.234s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.204s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.397s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.531s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.184s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.625s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.718s |  |
| Basic echo function | ✅ Pass | 5.048s |  |
| Streaming Basic echo function | ✅ Pass | 4.879s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.761s |  |
| Streaming JSON response function | ✅ Pass | 4.667s |  |
| Search query | ✅ Pass | 4.477s |  |
| Streaming Search query | ✅ Pass | 4.976s |  |
| Ask advice | ✅ Pass | 18.771s |  |
| Streaming Ask advice | ✅ Pass | 9.640s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.007s

---

### pentester (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 4.607s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.118s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.842s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 4.454s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 5.320s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.419s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.936s |  |
| Basic echo function | ✅ Pass | 5.000s |  |
| Streaming Basic echo function | ✅ Pass | 6.064s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 4.942s |  |
| Streaming JSON response function | ✅ Pass | 4.720s |  |
| Search query | ✅ Pass | 4.801s |  |
| Streaming Search query | ✅ Pass | 4.670s |  |
| Ask advice | ✅ Pass | 22.337s |  |
| Streaming Ask advice | ✅ Pass | 15.103s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 6.756s

---

