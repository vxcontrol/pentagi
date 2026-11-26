# LLM Agent Testing Report

Generated: Thu, 26 Nov 2026 04:06:43 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | nodecode-cyber-v1 | false | 26/26 (100.00%) | 0.917s |
| simple_json | nodecode-cyber-v1 | false | 8/8 (100.00%) | 0.604s |
| primary_agent | nodecode-cyber-v1 | true | 25/25 (100.00%) | 0.877s |
| assistant | nodecode-cyber-v1 | true | 25/25 (100.00%) | 1.109s |
| generator | nodecode-cyber-v1 | true | 25/25 (100.00%) | 0.933s |
| refiner | nodecode-cyber-v1 | true | 25/25 (100.00%) | 0.992s |
| adviser | nodecode-cyber-v1 | true | 16/16 (100.00%) | 1.091s |
| reflector | nodecode-cyber-v1 | false | 17/17 (100.00%) | 0.763s |
| searcher | nodecode-cyber-v1 | false | 26/26 (100.00%) | 0.742s |
| enricher | nodecode-cyber-v1 | false | 26/26 (100.00%) | 0.780s |
| coder | nodecode-cyber-v1 | true | 25/25 (100.00%) | 1.126s |
| installer | nodecode-cyber-v1 | true | 25/25 (100.00%) | 1.273s |
| pentester | nodecode-cyber-v1 | true | 25/25 (100.00%) | 1.193s |

**Total**: 294/294 (100.00%) successful tests
**Overall average latency**: 0.974s

## Detailed Results

### simple (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Count from 1 to 5 | ✅ Pass | 0.367s |  |
| Simple Math | ✅ Pass | 0.489s |  |
| Math Calculation | ✅ Pass | 0.217s |  |
| Text Transform Uppercase | ✅ Pass | 0.584s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.311s |  |
| Basic Echo Function | ✅ Pass | 0.415s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.526s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.421s |  |
| Answer Stops At The Output Limit | ✅ Pass | 4.802s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.819s |  |
| Search Query Function | ✅ Pass | 0.783s |  |
| Ask Advice Function | ✅ Pass | 0.685s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.591s |  |
| Basic Context Memory Test | ✅ Pass | 0.414s |  |
| Function Argument Memory Test | ✅ Pass | 0.231s |  |
| Function Response Memory Test | ✅ Pass | 0.399s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.540s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.704s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.502s |  |
| SQL Injection Attack Type | ✅ Pass | 0.336s |  |
| Penetration Testing Framework | ✅ Pass | 0.402s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.996s |  |
| Web Application Security Scanner | ✅ Pass | 0.316s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.531s |  |
| Penetration Testing Methodology | ✅ Pass | 4.963s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Explicit Reasoning Off Suppresses Reasoning | reasoning_off | ✅ Pass | 0.496s |  |

**Summary**: 26/26 (100.00%) successful tests

**Average latency**: 0.917s

---

### simple_json (nodecode-cyber-v1)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.675s |  |
| Person Information JSON | ✅ Pass | 0.803s |  |
| Project Information JSON | ✅ Pass | 0.327s |  |
| User Profile JSON | ✅ Pass | 0.540s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.460s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.460s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output Refuses A Non-Object Schema | structured_output | ✅ Pass | 0.000s |  |
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.562s |  |

**Summary**: 8/8 (100.00%) successful tests

**Average latency**: 0.604s

---

### primary_agent (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.455s |  |
| Text Transform Uppercase | ✅ Pass | 0.567s |  |
| Count from 1 to 5 | ✅ Pass | 0.631s |  |
| Math Calculation | ✅ Pass | 0.453s |  |
| Basic Echo Function | ✅ Pass | 0.451s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.585s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.590s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.677s |  |
| Answer Stops At The Output Limit | ✅ Pass | 5.321s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.697s |  |
| Ask Advice Function | ✅ Pass | 0.502s |  |
| Search Query Function | ✅ Pass | 1.105s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.656s |  |
| Basic Context Memory Test | ✅ Pass | 0.485s |  |
| Function Argument Memory Test | ✅ Pass | 0.353s |  |
| Function Response Memory Test | ✅ Pass | 0.622s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.344s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.041s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.246s |  |
| Penetration Testing Methodology | ✅ Pass | 0.533s |  |
| SQL Injection Attack Type | ✅ Pass | 0.769s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.086s |  |
| Penetration Testing Framework | ✅ Pass | 0.711s |  |
| Web Application Security Scanner | ✅ Pass | 0.454s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.584s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 0.877s

---

### assistant (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.493s |  |
| Text Transform Uppercase | ✅ Pass | 0.670s |  |
| Count from 1 to 5 | ✅ Pass | 0.684s |  |
| Math Calculation | ✅ Pass | 0.444s |  |
| Basic Echo Function | ✅ Pass | 0.555s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.611s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.429s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.513s |  |
| Answer Stops At The Output Limit | ✅ Pass | 5.275s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.772s |  |
| Search Query Function | ✅ Pass | 0.549s |  |
| Ask Advice Function | ✅ Pass | 1.938s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.640s |  |
| Basic Context Memory Test | ✅ Pass | 1.740s |  |
| Function Argument Memory Test | ✅ Pass | 1.906s |  |
| Function Response Memory Test | ✅ Pass | 0.602s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.931s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.082s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.044s |  |
| Penetration Testing Methodology | ✅ Pass | 0.452s |  |
| Penetration Testing Framework | ✅ Pass | 0.473s |  |
| SQL Injection Attack Type | ✅ Pass | 0.651s |  |
| Web Application Security Scanner | ✅ Pass | 0.426s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.229s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.598s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.109s

---

### generator (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.441s |  |
| Text Transform Uppercase | ✅ Pass | 0.536s |  |
| Count from 1 to 5 | ✅ Pass | 0.577s |  |
| Math Calculation | ✅ Pass | 0.434s |  |
| Basic Echo Function | ✅ Pass | 0.604s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.611s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.448s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.417s |  |
| Answer Stops At The Output Limit | ✅ Pass | 5.544s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.588s |  |
| Ask Advice Function | ✅ Pass | 0.528s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.613s |  |
| Basic Context Memory Test | ✅ Pass | 0.590s |  |
| Function Response Memory Test | ✅ Pass | 0.525s |  |
| Function Argument Memory Test | ✅ Pass | 0.702s |  |
| Search Query Function | ✅ Pass | 1.782s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.624s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.970s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.105s |  |
| Penetration Testing Methodology | ✅ Pass | 0.459s |  |
| SQL Injection Attack Type | ✅ Pass | 0.854s |  |
| Penetration Testing Framework | ✅ Pass | 0.789s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.327s |  |
| Web Application Security Scanner | ✅ Pass | 0.690s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.558s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 0.933s

---

### refiner (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.418s |  |
| Text Transform Uppercase | ✅ Pass | 0.624s |  |
| Count from 1 to 5 | ✅ Pass | 0.746s |  |
| Math Calculation | ✅ Pass | 0.545s |  |
| Basic Echo Function | ✅ Pass | 0.505s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.691s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.614s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.644s |  |
| Answer Stops At The Output Limit | ✅ Pass | 5.575s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.562s |  |
| Search Query Function | ✅ Pass | 0.906s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.494s |  |
| Ask Advice Function | ✅ Pass | 0.558s |  |
| Function Argument Memory Test | ✅ Pass | 0.345s |  |
| Basic Context Memory Test | ✅ Pass | 0.833s |  |
| Function Response Memory Test | ✅ Pass | 0.535s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.159s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.195s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 1.552s |  |
| Penetration Testing Methodology | ✅ Pass | 0.759s |  |
| Penetration Testing Framework | ✅ Pass | 0.647s |  |
| SQL Injection Attack Type | ✅ Pass | 0.765s |  |
| Web Application Security Scanner | ✅ Pass | 0.898s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.992s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.226s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 0.992s

---

### adviser (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.015s |  |
| Text Transform Uppercase | ✅ Pass | 0.631s |  |
| Count from 1 to 5 | ✅ Pass | 0.635s |  |
| Math Calculation | ✅ Pass | 0.411s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.381s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.567s |  |
| Answer Stops At The Output Limit | ✅ Pass | 5.036s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Basic Context Memory Test | ✅ Pass | 0.506s |  |
| Function Argument Memory Test | ✅ Pass | 0.654s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.440s |  |
| Function Response Memory Test | ✅ Pass | 0.803s |  |
| Penetration Testing Methodology | ✅ Pass | 0.899s |  |
| Penetration Testing Framework | ✅ Pass | 1.042s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.860s |  |
| Web Application Security Scanner | ✅ Pass | 1.059s |  |
| SQL Injection Attack Type | ✅ Pass | 1.513s |  |

**Summary**: 16/16 (100.00%) successful tests

**Average latency**: 1.091s

---

### reflector (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.624s |  |
| Text Transform Uppercase | ✅ Pass | 0.413s |  |
| Math Calculation | ✅ Pass | 0.189s |  |
| Count from 1 to 5 | ✅ Pass | 0.399s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.281s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.492s |  |
| Answer Stops At The Output Limit | ✅ Pass | 4.009s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Basic Context Memory Test | ✅ Pass | 0.237s |  |
| Function Argument Memory Test | ✅ Pass | 0.197s |  |
| Function Response Memory Test | ✅ Pass | 0.417s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.666s |  |
| SQL Injection Attack Type | ✅ Pass | 0.307s |  |
| Penetration Testing Methodology | ✅ Pass | 0.831s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.590s |  |
| Web Application Security Scanner | ✅ Pass | 0.324s |  |
| Penetration Testing Framework | ✅ Pass | 1.632s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Explicit Reasoning Off Suppresses Reasoning | reasoning_off | ✅ Pass | 0.361s |  |

**Summary**: 17/17 (100.00%) successful tests

**Average latency**: 0.763s

---

### searcher (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.073s |  |
| Text Transform Uppercase | ✅ Pass | 0.412s |  |
| Count from 1 to 5 | ✅ Pass | 0.422s |  |
| Math Calculation | ✅ Pass | 0.272s |  |
| Basic Echo Function | ✅ Pass | 0.461s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.610s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.439s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.419s |  |
| Answer Stops At The Output Limit | ✅ Pass | 4.204s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.631s |  |
| Search Query Function | ✅ Pass | 0.535s |  |
| Ask Advice Function | ✅ Pass | 0.462s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.339s |  |
| Basic Context Memory Test | ✅ Pass | 0.245s |  |
| Function Argument Memory Test | ✅ Pass | 0.200s |  |
| Function Response Memory Test | ✅ Pass | 0.404s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.440s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.016s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.173s |  |
| SQL Injection Attack Type | ✅ Pass | 0.328s |  |
| Penetration Testing Methodology | ✅ Pass | 0.655s |  |
| Penetration Testing Framework | ✅ Pass | 0.637s |  |
| Web Application Security Scanner | ✅ Pass | 0.689s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.256s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.591s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Explicit Reasoning Off Suppresses Reasoning | reasoning_off | ✅ Pass | 0.358s |  |

**Summary**: 26/26 (100.00%) successful tests

**Average latency**: 0.742s

---

### enricher (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.299s |  |
| Text Transform Uppercase | ✅ Pass | 0.337s |  |
| Count from 1 to 5 | ✅ Pass | 0.466s |  |
| Math Calculation | ✅ Pass | 0.298s |  |
| Basic Echo Function | ✅ Pass | 0.400s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.544s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.425s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.502s |  |
| Answer Stops At The Output Limit | ✅ Pass | 4.528s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.626s |  |
| Ask Advice Function | ✅ Pass | 0.579s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.348s |  |
| Basic Context Memory Test | ✅ Pass | 0.445s |  |
| Function Argument Memory Test | ✅ Pass | 0.194s |  |
| Function Response Memory Test | ✅ Pass | 0.379s |  |
| Search Query Function | ✅ Pass | 1.183s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.425s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.758s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.238s |  |
| SQL Injection Attack Type | ✅ Pass | 0.360s |  |
| Penetration Testing Methodology | ✅ Pass | 0.890s |  |
| Web Application Security Scanner | ✅ Pass | 0.628s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.248s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.644s |  |
| Penetration Testing Framework | ✅ Pass | 1.222s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Explicit Reasoning Off Suppresses Reasoning | reasoning_off | ✅ Pass | 0.313s |  |

**Summary**: 26/26 (100.00%) successful tests

**Average latency**: 0.780s

---

### coder (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.535s |  |
| Text Transform Uppercase | ✅ Pass | 0.578s |  |
| Count from 1 to 5 | ✅ Pass | 0.411s |  |
| Math Calculation | ✅ Pass | 0.977s |  |
| Basic Echo Function | ✅ Pass | 1.193s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.698s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.030s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.811s |  |
| Answer Stops At The Output Limit | ✅ Pass | 7.286s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.622s |  |
| Search Query Function | ✅ Pass | 0.551s |  |
| Ask Advice Function | ✅ Pass | 0.692s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.556s |  |
| Basic Context Memory Test | ✅ Pass | 0.515s |  |
| Function Argument Memory Test | ✅ Pass | 1.095s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.097s |  |
| Function Response Memory Test | ✅ Pass | 1.304s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.485s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 1.569s |  |
| Penetration Testing Methodology | ✅ Pass | 0.651s |  |
| Penetration Testing Framework | ✅ Pass | 0.545s |  |
| SQL Injection Attack Type | ✅ Pass | 1.012s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.267s |  |
| Web Application Security Scanner | ✅ Pass | 0.911s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.747s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.126s

---

### installer (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.452s |  |
| Text Transform Uppercase | ✅ Pass | 0.809s |  |
| Count from 1 to 5 | ✅ Pass | 1.005s |  |
| Math Calculation | ✅ Pass | 0.612s |  |
| Basic Echo Function | ✅ Pass | 0.984s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.356s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.721s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.125s |  |
| Answer Stops At The Output Limit | ✅ Pass | 9.580s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.933s |  |
| Search Query Function | ✅ Pass | 0.669s |  |
| Ask Advice Function | ✅ Pass | 0.746s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.585s |  |
| Basic Context Memory Test | ✅ Pass | 0.446s |  |
| Function Argument Memory Test | ✅ Pass | 0.404s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.498s |  |
| Function Response Memory Test | ✅ Pass | 1.050s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.062s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.183s |  |
| Penetration Testing Methodology | ✅ Pass | 0.816s |  |
| Penetration Testing Framework | ✅ Pass | 0.954s |  |
| SQL Injection Attack Type | ✅ Pass | 1.315s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.383s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.688s |  |
| Web Application Security Scanner | ✅ Pass | 1.429s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.273s

---

### pentester (nodecode-cyber-v1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.449s |  |
| Text Transform Uppercase | ✅ Pass | 0.939s |  |
| Count from 1 to 5 | ✅ Pass | 0.939s |  |
| Math Calculation | ✅ Pass | 0.621s |  |
| Basic Echo Function | ✅ Pass | 0.610s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.738s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.178s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.245s |  |
| Answer Stops At The Output Limit | ✅ Pass | 7.980s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.855s |  |
| Search Query Function | ✅ Pass | 0.773s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.654s |  |
| Ask Advice Function | ✅ Pass | 0.795s |  |
| Basic Context Memory Test | ✅ Pass | 0.725s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.319s |  |
| Function Response Memory Test | ✅ Pass | 0.621s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.920s |  |
| Function Argument Memory Test | ✅ Pass | 2.715s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 1.608s |  |
| Penetration Testing Methodology | ✅ Pass | 0.637s |  |
| Penetration Testing Framework | ✅ Pass | 0.585s |  |
| SQL Injection Attack Type | ✅ Pass | 0.689s |  |
| Web Application Security Scanner | ✅ Pass | 0.849s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.698s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.680s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.193s

---

