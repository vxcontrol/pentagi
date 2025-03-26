# LLM Agent Testing Report

Generated: Wed, 26 Mar 2025 18:13:04 UTC

## Overall Results

| Agent | Model | Success Rate | Average Latency |
|-------|-------|-------------|----------------|
| simple | deepseek-chat | 8/9 (88.89%) | 9.216s |
| simple_json | deepseek-chat | 5/5 (100.00%) | 7.933s |
| agent | deepseek-chat | 8/9 (88.89%) | 9.584s |
| generator | deepseek-chat | 8/9 (88.89%) | 9.045s |
| refiner | deepseek-chat | 8/9 (88.89%) | 9.489s |
| adviser | deepseek-chat | 8/9 (88.89%) | 9.579s |
| reflector | deepseek-reasoner | 5/9 (55.56%) | 7.022s |
| searcher | deepseek-chat | 8/9 (88.89%) | 9.033s |
| enricher | deepseek-chat | 8/9 (88.89%) | 8.783s |
| coder | deepseek-coder | 8/9 (88.89%) | 9.248s |
| installer | deepseek-coder | 8/9 (88.89%) | 8.823s |
| pentester | deepseek-chat | 8/9 (88.89%) | 9.484s |

**Total**: 90/104 (86.54%) successful tests
**Overall average latency**: 8.975s

## Detailed Results

### simple (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.616s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.566s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 3.717s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.626s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.592s |  |
| Basic echo function | ✅ Pass | 5.748s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.278s |  |
| Search query | ✅ Pass | 5.509s |  |
| Ask advice | ❌ Fail | 32.296s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.216s

---

### simple_json (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 16.570s |  |
| SimpleJSON: Create a JSON object with f... | ✅ Pass | 4.985s |  |
| SimpleJSON: Generate a JSON response fo... | ✅ Pass | 5.939s |  |
| SimpleJSON: Create a JSON array of 3 co... | ✅ Pass | 7.090s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| SimpleJSON: What is 2+2? Explain your a... | ✅ Pass | 5.083s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 7.933s

---

### agent (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 17.164s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 5.192s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.475s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.510s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.797s |  |
| Basic echo function | ✅ Pass | 5.831s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.140s |  |
| Search query | ✅ Pass | 4.861s |  |
| Ask advice | ❌ Fail | 32.290s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.584s

---

### generator (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.669s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.745s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.158s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.489s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.887s |  |
| Basic echo function | ✅ Pass | 5.908s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.344s |  |
| Search query | ✅ Pass | 5.397s |  |
| Ask advice | ❌ Fail | 30.808s | model did not call a function, responded with text: The error \*\*"cannot find package"\*\* in Go devel\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.045s

---

### refiner (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.666s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.517s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.624s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.417s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.014s |  |
| Basic echo function | ✅ Pass | 6.361s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 6.289s |  |
| Search query | ✅ Pass | 4.741s |  |
| Ask advice | ❌ Fail | 32.769s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.489s

---

### adviser (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.734s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 6.303s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 5.046s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 3.998s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 6.319s |  |
| Basic echo function | ✅ Pass | 5.990s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.050s |  |
| Search query | ✅ Pass | 6.025s |  |
| Ask advice | ❌ Fail | 31.745s | model did not call a function, responded with text: The error \`'cannot find package'\` in Go develop\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.579s

---

### reflector (deepseek-reasoner)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 21.024s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 11.234s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 7.805s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 10.717s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 11.055s |  |
| Basic echo function | ❌ Fail | 0.336s | API error: API returned unexpected status code: 400: deepseek\-reasoner does not support Function Calling |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ❌ Fail | 0.363s | API error: API returned unexpected status code: 400: deepseek\-reasoner does not support Function Calling |
| Search query | ❌ Fail | 0.329s | API error: API returned unexpected status code: 400: deepseek\-reasoner does not support Function Calling |
| Ask advice | ❌ Fail | 0.331s | API error: API returned unexpected status code: 400: deepseek\-reasoner does not support Function Calling |

**Summary**: 5/9 (55.56%) successful tests

**Average latency**: 7.022s

---

### searcher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.720s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.717s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.146s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.562s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.755s |  |
| Basic echo function | ✅ Pass | 5.985s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.287s |  |
| Search query | ✅ Pass | 5.298s |  |
| Ask advice | ❌ Fail | 30.823s | model did not call a function, responded with text: The error \*\*"cannot find package"\*\* in Go devel\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.033s

---

### enricher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.733s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.915s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.338s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.482s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.655s |  |
| Basic echo function | ✅ Pass | 5.739s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.298s |  |
| Search query | ✅ Pass | 5.260s |  |
| Ask advice | ❌ Fail | 28.624s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 8.783s

---

### coder (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.730s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.689s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.212s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 6.503s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.119s |  |
| Basic echo function | ✅ Pass | 5.371s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.580s |  |
| Search query | ✅ Pass | 5.026s |  |
| Ask advice | ❌ Fail | 31.005s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.248s

---

### installer (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.718s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.690s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.186s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 4.500s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 4.808s |  |
| Basic echo function | ✅ Pass | 5.946s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.382s |  |
| Search query | ✅ Pass | 5.310s |  |
| Ask advice | ❌ Fail | 28.867s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 8.823s

---

### pentester (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Completion: What is 2+2? Write only the... | ✅ Pass | 15.658s |  |
| Completion: Write 'Hello World' in uppe... | ✅ Pass | 4.760s |  |
| System-User: Count from 1 to 5, separate... | ✅ Pass | 4.157s |  |
| System-User: Calculate 5 * 10 and provid... | ✅ Pass | 6.442s |  |
| SimpleJSON: Return a JSON with a person... | ✅ Pass | 5.318s |  |
| Basic echo function | ✅ Pass | 5.276s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON response function | ✅ Pass | 5.325s |  |
| Search query | ✅ Pass | 5.591s |  |
| Ask advice | ❌ Fail | 32.830s | model did not call a function, responded with text: The error \`cannot find package\` in Go developme\.\.\. |

**Summary**: 8/9 (88.89%) successful tests

**Average latency**: 9.484s

---

