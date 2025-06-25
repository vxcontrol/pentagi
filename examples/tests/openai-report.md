# LLM Agent Testing Report

Generated: Wed, 25 Jun 2025 19:15:20 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-4.1-mini | false | 15/15 (100.00%) | 1.099s |
| simple_json | gpt-4.1-mini | false | 5/5 (100.00%) | 1.420s |
| agent | o4-mini | true | 15/15 (100.00%) | 1.926s |
| assistant | o4-mini | true | 15/15 (100.00%) | 2.611s |
| generator | o3 | true | 15/15 (100.00%) | 4.103s |
| refiner | gpt-4.1 | false | 15/15 (100.00%) | 1.275s |
| adviser | o4-mini | true | 15/15 (100.00%) | 4.591s |
| reflector | o4-mini | true | 15/15 (100.00%) | 4.035s |
| searcher | gpt-4.1-mini | false | 15/15 (100.00%) | 1.007s |
| enricher | gpt-4.1-mini | false | 15/15 (100.00%) | 1.073s |
| coder | gpt-4.1 | false | 15/15 (100.00%) | 1.486s |
| installer | gpt-4.1 | false | 15/15 (100.00%) | 1.418s |
| pentester | o4-mini | true | 15/15 (100.00%) | 2.062s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 2.202s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.879s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.588s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.286s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.521s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.677s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.444s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.153s |  |
| Basic echo function | ✅ Pass | 0.668s |  |
| Streaming Basic echo function | ✅ Pass | 0.643s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.896s |  |
| Streaming JSON response function | ✅ Pass | 0.960s |  |
| Search query | ✅ Pass | 0.735s |  |
| Streaming Search query | ✅ Pass | 1.247s |  |
| Ask advice | ✅ Pass | 2.303s |  |
| Streaming Ask advice | ✅ Pass | 2.492s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.099s

---

### simple_json (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.204s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 0.983s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 1.092s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 1.664s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.158s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.420s

---

### agent (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.419s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.267s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.826s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.307s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.249s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.771s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.690s |  |
| Basic echo function | ✅ Pass | 1.406s |  |
| Streaming Basic echo function | ✅ Pass | 1.263s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.105s |  |
| Streaming JSON response function | ✅ Pass | 1.242s |  |
| Search query | ✅ Pass | 1.554s |  |
| Streaming Search query | ✅ Pass | 1.073s |  |
| Ask advice | ✅ Pass | 5.516s |  |
| Streaming Ask advice | ✅ Pass | 3.201s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.926s

---

### assistant (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.533s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.853s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.785s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.953s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.314s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.979s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.601s |  |
| Basic echo function | ✅ Pass | 1.986s |  |
| Streaming Basic echo function | ✅ Pass | 2.845s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.303s |  |
| Streaming JSON response function | ✅ Pass | 5.908s |  |
| Search query | ✅ Pass | 2.288s |  |
| Streaming Search query | ✅ Pass | 2.925s |  |
| Ask advice | ✅ Pass | 3.661s |  |
| Streaming Ask advice | ✅ Pass | 3.238s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.611s

---

### generator (o3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.736s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.307s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.202s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.692s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.597s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.315s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.375s |  |
| Basic echo function | ✅ Pass | 2.486s |  |
| Streaming Basic echo function | ✅ Pass | 2.513s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.910s |  |
| Streaming JSON response function | ✅ Pass | 1.443s |  |
| Search query | ✅ Pass | 2.760s |  |
| Streaming Search query | ✅ Pass | 1.863s |  |
| Ask advice | ✅ Pass | 19.356s |  |
| Streaming Ask advice | ✅ Pass | 15.987s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.103s

---

### refiner (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.539s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.405s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.907s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.886s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.618s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.847s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.807s |  |
| Basic echo function | ✅ Pass | 0.639s |  |
| Streaming Basic echo function | ✅ Pass | 1.124s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.695s |  |
| Streaming JSON response function | ✅ Pass | 0.631s |  |
| Search query | ✅ Pass | 0.575s |  |
| Streaming Search query | ✅ Pass | 1.010s |  |
| Ask advice | ✅ Pass | 5.523s |  |
| Streaming Ask advice | ✅ Pass | 2.918s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.275s

---

### adviser (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.189s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.521s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.760s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.291s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.138s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.501s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.298s |  |
| Basic echo function | ✅ Pass | 8.408s |  |
| Streaming Basic echo function | ✅ Pass | 3.147s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 8.511s |  |
| Streaming JSON response function | ✅ Pass | 5.311s |  |
| Search query | ✅ Pass | 7.769s |  |
| Streaming Search query | ✅ Pass | 1.947s |  |
| Ask advice | ✅ Pass | 7.097s |  |
| Streaming Ask advice | ✅ Pass | 12.973s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.591s

---

### reflector (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.948s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 3.733s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.386s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.971s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.573s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.127s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.366s |  |
| Basic echo function | ✅ Pass | 2.228s |  |
| Streaming Basic echo function | ✅ Pass | 2.774s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 3.395s |  |
| Streaming JSON response function | ✅ Pass | 5.751s |  |
| Search query | ✅ Pass | 5.058s |  |
| Streaming Search query | ✅ Pass | 2.560s |  |
| Ask advice | ✅ Pass | 9.365s |  |
| Streaming Ask advice | ✅ Pass | 12.283s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 4.035s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.738s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.399s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.014s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.601s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.515s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.517s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.166s |  |
| Basic echo function | ✅ Pass | 0.639s |  |
| Streaming Basic echo function | ✅ Pass | 0.739s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.713s |  |
| Streaming JSON response function | ✅ Pass | 0.714s |  |
| Search query | ✅ Pass | 0.785s |  |
| Streaming Search query | ✅ Pass | 0.616s |  |
| Ask advice | ✅ Pass | 2.559s |  |
| Streaming Ask advice | ✅ Pass | 2.397s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.007s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.440s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.146s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.652s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.869s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.723s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.627s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.993s |  |
| Basic echo function | ✅ Pass | 0.878s |  |
| Streaming Basic echo function | ✅ Pass | 0.616s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.823s |  |
| Streaming JSON response function | ✅ Pass | 0.784s |  |
| Search query | ✅ Pass | 0.572s |  |
| Streaming Search query | ✅ Pass | 0.731s |  |
| Ask advice | ✅ Pass | 2.903s |  |
| Streaming Ask advice | ✅ Pass | 2.340s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.073s

---

### coder (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.768s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.475s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.720s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.112s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.616s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.671s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.292s |  |
| Basic echo function | ✅ Pass | 0.519s |  |
| Streaming Basic echo function | ✅ Pass | 0.994s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.558s |  |
| Streaming JSON response function | ✅ Pass | 0.764s |  |
| Search query | ✅ Pass | 0.537s |  |
| Streaming Search query | ✅ Pass | 0.702s |  |
| Ask advice | ✅ Pass | 5.163s |  |
| Streaming Ask advice | ✅ Pass | 5.403s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.486s

---

### installer (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.580s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.788s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.745s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.554s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.394s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.395s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.731s |  |
| Basic echo function | ✅ Pass | 1.211s |  |
| Streaming Basic echo function | ✅ Pass | 0.642s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.527s |  |
| Streaming JSON response function | ✅ Pass | 0.643s |  |
| Search query | ✅ Pass | 0.664s |  |
| Streaming Search query | ✅ Pass | 0.868s |  |
| Ask advice | ✅ Pass | 5.199s |  |
| Streaming Ask advice | ✅ Pass | 6.323s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.418s

---

### pentester (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.764s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.865s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.523s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.276s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.103s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.695s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.473s |  |
| Basic echo function | ✅ Pass | 1.619s |  |
| Streaming Basic echo function | ✅ Pass | 2.652s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.285s |  |
| Streaming JSON response function | ✅ Pass | 3.101s |  |
| Search query | ✅ Pass | 1.850s |  |
| Streaming Search query | ✅ Pass | 0.986s |  |
| Ask advice | ✅ Pass | 5.011s |  |
| Streaming Ask advice | ✅ Pass | 2.724s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.062s

---

