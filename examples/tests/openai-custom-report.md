# LLM Agent Testing Report

Generated: Tue, 01 Jul 2025 17:34:35 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-4.1-mini | false | 15/15 (100.00%) | 1.256s |
| simple_json | gpt-4.1-mini | false | 5/5 (100.00%) | 1.460s |
| agent | o3-mini | true | 15/15 (100.00%) | 2.078s |
| assistant | o3-mini | true | 15/15 (100.00%) | 2.628s |
| generator | o3-mini | true | 15/15 (100.00%) | 2.988s |
| refiner | gpt-4.1 | false | 15/15 (100.00%) | 1.743s |
| adviser | o3-mini | true | 15/15 (100.00%) | 2.499s |
| reflector | o3-mini | true | 15/15 (100.00%) | 3.144s |
| searcher | gpt-4.1-mini | false | 15/15 (100.00%) | 1.195s |
| enricher | gpt-4.1-mini | false | 15/15 (100.00%) | 1.279s |
| coder | gpt-4.1 | false | 15/15 (100.00%) | 1.309s |
| installer | gpt-4.1 | false | 15/15 (100.00%) | 1.413s |
| pentester | o3-mini | true | 15/15 (100.00%) | 2.259s |

**Total**: 185/185 (100.00%) successful tests
**Overall average latency**: 1.968s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.306s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.578s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.697s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.697s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.863s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.549s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.225s |  |
| Basic echo function | ✅ Pass | 0.814s |  |
| Streaming Basic echo function | ✅ Pass | 0.822s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.854s |  |
| Streaming JSON response function | ✅ Pass | 0.828s |  |
| Search query | ✅ Pass | 0.849s |  |
| Streaming Search query | ✅ Pass | 0.816s |  |
| Ask advice | ✅ Pass | 2.210s |  |
| Streaming Ask advice | ✅ Pass | 2.737s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.256s

---

### simple_json (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.013s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 0.872s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 0.909s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 2.224s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 1.284s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.460s

---

### agent (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.484s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 1.547s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.998s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.874s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.201s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.510s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.507s |  |
| Basic echo function | ✅ Pass | 1.908s |  |
| Streaming Basic echo function | ✅ Pass | 1.263s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.488s |  |
| Streaming JSON response function | ✅ Pass | 1.297s |  |
| Search query | ✅ Pass | 2.280s |  |
| Streaming Search query | ✅ Pass | 1.325s |  |
| Ask advice | ✅ Pass | 3.292s |  |
| Streaming Ask advice | ✅ Pass | 3.190s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.078s

---

### assistant (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.405s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.768s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.629s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.718s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.315s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.193s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.186s |  |
| Basic echo function | ✅ Pass | 2.745s |  |
| Streaming Basic echo function | ✅ Pass | 1.325s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.683s |  |
| Streaming JSON response function | ✅ Pass | 1.484s |  |
| Search query | ✅ Pass | 1.997s |  |
| Streaming Search query | ✅ Pass | 1.371s |  |
| Ask advice | ✅ Pass | 6.386s |  |
| Streaming Ask advice | ✅ Pass | 4.221s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.628s

---

### generator (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.948s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.931s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.016s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.702s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.240s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.717s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.459s |  |
| Basic echo function | ✅ Pass | 2.247s |  |
| Streaming Basic echo function | ✅ Pass | 1.426s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.362s |  |
| Streaming JSON response function | ✅ Pass | 1.723s |  |
| Search query | ✅ Pass | 1.928s |  |
| Streaming Search query | ✅ Pass | 2.164s |  |
| Ask advice | ✅ Pass | 5.864s |  |
| Streaming Ask advice | ✅ Pass | 5.099s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.988s

---

### refiner (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.814s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.590s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.594s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.844s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.562s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.635s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.855s |  |
| Basic echo function | ✅ Pass | 1.284s |  |
| Streaming Basic echo function | ✅ Pass | 0.702s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.872s |  |
| Streaming JSON response function | ✅ Pass | 0.637s |  |
| Search query | ✅ Pass | 0.660s |  |
| Streaming Search query | ✅ Pass | 0.632s |  |
| Ask advice | ✅ Pass | 9.359s |  |
| Streaming Ask advice | ✅ Pass | 6.099s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.743s

---

### adviser (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 2.650s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.589s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 2.167s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 3.109s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.726s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.028s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.879s |  |
| Basic echo function | ✅ Pass | 2.255s |  |
| Streaming Basic echo function | ✅ Pass | 1.800s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.366s |  |
| Streaming JSON response function | ✅ Pass | 1.424s |  |
| Search query | ✅ Pass | 2.239s |  |
| Streaming Search query | ✅ Pass | 1.994s |  |
| Ask advice | ✅ Pass | 5.183s |  |
| Streaming Ask advice | ✅ Pass | 3.075s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.499s

---

### reflector (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.602s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.869s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.881s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.552s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 2.121s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.416s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 3.032s |  |
| Basic echo function | ✅ Pass | 1.813s |  |
| Streaming Basic echo function | ✅ Pass | 2.198s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.519s |  |
| Streaming JSON response function | ✅ Pass | 3.245s |  |
| Search query | ✅ Pass | 1.430s |  |
| Streaming Search query | ✅ Pass | 2.139s |  |
| Ask advice | ✅ Pass | 6.274s |  |
| Streaming Ask advice | ✅ Pass | 5.062s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 3.144s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.805s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.697s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.829s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.767s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.727s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.683s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.982s |  |
| Basic echo function | ✅ Pass | 1.200s |  |
| Streaming Basic echo function | ✅ Pass | 0.781s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.788s |  |
| Streaming JSON response function | ✅ Pass | 0.900s |  |
| Search query | ✅ Pass | 0.891s |  |
| Streaming Search query | ✅ Pass | 0.735s |  |
| Ask advice | ✅ Pass | 3.237s |  |
| Streaming Ask advice | ✅ Pass | 2.896s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.195s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.112s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.659s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.254s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 2.683s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.368s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.831s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 1.024s |  |
| Basic echo function | ✅ Pass | 0.797s |  |
| Streaming Basic echo function | ✅ Pass | 0.883s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 1.075s |  |
| Streaming JSON response function | ✅ Pass | 0.867s |  |
| Search query | ✅ Pass | 0.700s |  |
| Streaming Search query | ✅ Pass | 0.785s |  |
| Ask advice | ✅ Pass | 2.513s |  |
| Streaming Ask advice | ✅ Pass | 2.641s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.279s

---

### coder (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.295s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.514s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 0.689s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.564s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.613s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.728s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.651s |  |
| Basic echo function | ✅ Pass | 0.766s |  |
| Streaming Basic echo function | ✅ Pass | 0.655s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.809s |  |
| Streaming JSON response function | ✅ Pass | 0.839s |  |
| Search query | ✅ Pass | 0.829s |  |
| Streaming Search query | ✅ Pass | 1.205s |  |
| Ask advice | ✅ Pass | 3.895s |  |
| Streaming Ask advice | ✅ Pass | 5.587s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.309s

---

### installer (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 1.195s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 0.600s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 1.228s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 0.743s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.508s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 0.576s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 0.706s |  |
| Basic echo function | ✅ Pass | 0.782s |  |
| Streaming Basic echo function | ✅ Pass | 0.801s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 0.969s |  |
| Streaming JSON response function | ✅ Pass | 1.619s |  |
| Search query | ✅ Pass | 0.701s |  |
| Streaming Search query | ✅ Pass | 0.918s |  |
| Ask advice | ✅ Pass | 5.309s |  |
| Streaming Ask advice | ✅ Pass | 4.543s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 1.413s

---

### pentester (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 3.838s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 2.388s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.741s |  |
| Streaming System-User: Count from 1 to 5, separate... | ✅ Pass | 1.713s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.302s |  |
| Streaming System-User: Calculate 5 * 10 and provid... | ✅ Pass | 1.433s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 2.120s |  |
| Basic echo function | ✅ Pass | 1.914s |  |
| Streaming Basic echo function | ✅ Pass | 1.239s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 2.852s |  |
| Streaming JSON response function | ✅ Pass | 1.637s |  |
| Search query | ✅ Pass | 1.522s |  |
| Streaming Search query | ✅ Pass | 1.654s |  |
| Ask advice | ✅ Pass | 4.168s |  |
| Streaming Ask advice | ✅ Pass | 2.358s |  |

**Summary**: 15/15 (100.00%) successful tests

**Average latency**: 2.259s

---

