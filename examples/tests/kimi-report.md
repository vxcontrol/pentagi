# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 16:45:17 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | kimi-k2.5 | true | 23/23 (100.00%) | 1.748s |
| simple_json | kimi-k2.5 | true | 7/7 (100.00%) | 2.417s |
| primary_agent | kimi-k2.7-code-highspeed | true | 23/23 (100.00%) | 0.918s |
| assistant | kimi-k2.7-code-highspeed | true | 23/23 (100.00%) | 0.859s |
| generator | kimi-k3 | true | 22/23 (95.65%) | 6.887s |
| refiner | kimi-k3 | true | 22/23 (95.65%) | 5.991s |
| adviser | kimi-k3 | true | 22/23 (95.65%) | 6.501s |
| reflector | kimi-k2.5 | true | 23/23 (100.00%) | 3.374s |
| searcher | kimi-k2.5 | true | 23/23 (100.00%) | 3.541s |
| enricher | kimi-k2.5 | true | 23/23 (100.00%) | 2.676s |
| coder | kimi-k2.7-code | true | 23/23 (100.00%) | 1.857s |
| installer | kimi-k2.5 | true | 23/23 (100.00%) | 4.023s |
| pentester | kimi-k2.7-code | true | 23/23 (100.00%) | 1.658s |

**Total**: 280/283 (98.94%) successful tests
**Overall average latency**: 3.313s

## Detailed Results

### simple (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.168s |  |
| Text Transform Uppercase | ✅ Pass | 0.847s |  |
| Count from 1 to 5 | ✅ Pass | 1.035s |  |
| Math Calculation | ✅ Pass | 0.850s |  |
| Basic Echo Function | ✅ Pass | 1.267s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.791s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.979s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.140s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.398s |  |
| Search Query Function | ✅ Pass | 1.141s |  |
| Ask Advice Function | ✅ Pass | 3.642s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.218s |  |
| Basic Context Memory Test | ✅ Pass | 1.036s |  |
| Function Argument Memory Test | ✅ Pass | 0.924s |  |
| Function Response Memory Test | ✅ Pass | 0.819s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.141s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.905s |  |
| Penetration Testing Methodology | ✅ Pass | 6.074s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.232s |  |
| SQL Injection Attack Type | ✅ Pass | 1.155s |  |
| Penetration Testing Framework | ✅ Pass | 4.287s |  |
| Web Application Security Scanner | ✅ Pass | 2.473s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.677s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.748s

---

### simple_json (kimi-k2.5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 1.108s |  |
| User Profile JSON | ✅ Pass | 1.267s |  |
| Person Information JSON | ✅ Pass | 3.657s |  |
| Vulnerability Report Memory Test | ✅ Pass | 3.801s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 3.193s |  |
| JSON Array Response Without Schema | ✅ Pass | 3.663s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.225s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 2.417s

---

### primary_agent (kimi-k2.7-code-highspeed)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.475s |  |
| Text Transform Uppercase | ✅ Pass | 0.213s |  |
| Count from 1 to 5 | ✅ Pass | 0.217s |  |
| Math Calculation | ✅ Pass | 0.218s |  |
| Basic Echo Function | ✅ Pass | 0.234s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.215s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.243s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.008s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.081s |  |
| Search Query Function | ✅ Pass | 0.896s |  |
| Ask Advice Function | ✅ Pass | 1.040s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.019s |  |
| Basic Context Memory Test | ✅ Pass | 1.074s |  |
| Function Argument Memory Test | ✅ Pass | 1.040s |  |
| Function Response Memory Test | ✅ Pass | 0.944s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.317s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.104s |  |
| Penetration Testing Methodology | ✅ Pass | 1.063s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.716s |  |
| SQL Injection Attack Type | ✅ Pass | 1.002s |  |
| Penetration Testing Framework | ✅ Pass | 1.607s |  |
| Web Application Security Scanner | ✅ Pass | 1.204s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.160s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.918s

---

### assistant (kimi-k2.7-code-highspeed)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.473s |  |
| Text Transform Uppercase | ✅ Pass | 0.213s |  |
| Count from 1 to 5 | ✅ Pass | 0.208s |  |
| Math Calculation | ✅ Pass | 0.212s |  |
| Basic Echo Function | ✅ Pass | 0.216s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.217s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.166s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.120s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.086s |  |
| Search Query Function | ✅ Pass | 1.038s |  |
| Ask Advice Function | ✅ Pass | 2.175s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.014s |  |
| Basic Context Memory Test | ✅ Pass | 1.012s |  |
| Function Argument Memory Test | ✅ Pass | 1.053s |  |
| Function Response Memory Test | ✅ Pass | 1.089s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.377s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.055s |  |
| Penetration Testing Methodology | ✅ Pass | 1.270s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.209s |  |
| SQL Injection Attack Type | ✅ Pass | 0.224s |  |
| Penetration Testing Framework | ✅ Pass | 1.119s |  |
| Web Application Security Scanner | ✅ Pass | 0.997s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.199s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.859s

---

### generator (kimi-k3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.409s |  |
| Text Transform Uppercase | ✅ Pass | 3.522s |  |
| Count from 1 to 5 | ✅ Pass | 3.907s |  |
| Math Calculation | ✅ Pass | 2.995s |  |
| Basic Echo Function | ✅ Pass | 4.075s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.937s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.215s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.155s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.522s |  |
| Search Query Function | ✅ Pass | 5.932s |  |
| Ask Advice Function | ✅ Pass | 3.868s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.012s |  |
| Basic Context Memory Test | ✅ Pass | 5.962s |  |
| Function Argument Memory Test | ✅ Pass | 4.789s |  |
| Function Response Memory Test | ✅ Pass | 4.603s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 9.806s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.822s |  |
| Penetration Testing Methodology | ✅ Pass | 20.204s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.992s |  |
| SQL Injection Attack Type | ✅ Pass | 5.997s |  |
| Penetration Testing Framework | ✅ Pass | 19.052s |  |
| Web Application Security Scanner | ✅ Pass | 11.354s |  |
| Penetration Testing Tool Selection | ✅ Pass | 9.255s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 6.887s

---

### refiner (kimi-k3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.383s |  |
| Text Transform Uppercase | ✅ Pass | 3.330s |  |
| Count from 1 to 5 | ✅ Pass | 4.068s |  |
| Math Calculation | ✅ Pass | 5.388s |  |
| Basic Echo Function | ✅ Pass | 4.278s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.022s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.832s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.464s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.040s |  |
| Search Query Function | ✅ Pass | 3.786s |  |
| Ask Advice Function | ✅ Pass | 4.272s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.479s |  |
| Basic Context Memory Test | ✅ Pass | 4.717s |  |
| Function Argument Memory Test | ✅ Pass | 6.360s |  |
| Function Response Memory Test | ✅ Pass | 3.190s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 7.865s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.861s |  |
| Penetration Testing Methodology | ✅ Pass | 19.021s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.633s |  |
| SQL Injection Attack Type | ✅ Pass | 5.304s |  |
| Penetration Testing Framework | ✅ Pass | 11.172s |  |
| Web Application Security Scanner | ✅ Pass | 12.213s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.113s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.991s

---

### adviser (kimi-k3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.838s |  |
| Text Transform Uppercase | ✅ Pass | 3.910s |  |
| Count from 1 to 5 | ✅ Pass | 4.158s |  |
| Math Calculation | ✅ Pass | 2.513s |  |
| Basic Echo Function | ✅ Pass | 4.136s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.902s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.926s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.287s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 4.621s | no tool calls found, expected at least 1 |
| Search Query Function | ✅ Pass | 5.229s |  |
| Ask Advice Function | ✅ Pass | 4.019s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.333s |  |
| Basic Context Memory Test | ✅ Pass | 7.273s |  |
| Function Argument Memory Test | ✅ Pass | 4.714s |  |
| Function Response Memory Test | ✅ Pass | 3.870s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.639s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.006s |  |
| Penetration Testing Methodology | ✅ Pass | 18.671s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.315s |  |
| SQL Injection Attack Type | ✅ Pass | 5.862s |  |
| Penetration Testing Framework | ✅ Pass | 14.244s |  |
| Web Application Security Scanner | ✅ Pass | 11.567s |  |
| Penetration Testing Tool Selection | ✅ Pass | 7.487s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 6.501s

---

### reflector (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.784s |  |
| Text Transform Uppercase | ✅ Pass | 0.986s |  |
| Count from 1 to 5 | ✅ Pass | 1.347s |  |
| Math Calculation | ✅ Pass | 2.671s |  |
| Basic Echo Function | ✅ Pass | 4.613s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.516s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.414s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.371s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.791s |  |
| Search Query Function | ✅ Pass | 1.454s |  |
| Ask Advice Function | ✅ Pass | 5.769s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.161s |  |
| Basic Context Memory Test | ✅ Pass | 2.177s |  |
| Function Argument Memory Test | ✅ Pass | 0.825s |  |
| Function Response Memory Test | ✅ Pass | 0.815s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.192s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.839s |  |
| Penetration Testing Methodology | ✅ Pass | 9.492s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.875s |  |
| SQL Injection Attack Type | ✅ Pass | 1.054s |  |
| Penetration Testing Framework | ✅ Pass | 8.044s |  |
| Web Application Security Scanner | ✅ Pass | 2.056s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.347s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.374s

---

### searcher (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.660s |  |
| Text Transform Uppercase | ✅ Pass | 1.039s |  |
| Count from 1 to 5 | ✅ Pass | 1.295s |  |
| Math Calculation | ✅ Pass | 1.552s |  |
| Basic Echo Function | ✅ Pass | 4.212s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.500s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.993s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.287s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.724s |  |
| Search Query Function | ✅ Pass | 5.527s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.268s |  |
| Ask Advice Function | ✅ Pass | 5.755s |  |
| Basic Context Memory Test | ✅ Pass | 2.214s |  |
| Function Argument Memory Test | ✅ Pass | 0.813s |  |
| Function Response Memory Test | ✅ Pass | 0.886s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.317s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.290s |  |
| Penetration Testing Methodology | ✅ Pass | 9.712s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.599s |  |
| SQL Injection Attack Type | ✅ Pass | 1.166s |  |
| Penetration Testing Framework | ✅ Pass | 12.429s |  |
| Web Application Security Scanner | ✅ Pass | 6.200s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.004s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.541s

---

### enricher (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.194s |  |
| Text Transform Uppercase | ✅ Pass | 0.984s |  |
| Count from 1 to 5 | ✅ Pass | 1.359s |  |
| Math Calculation | ✅ Pass | 1.392s |  |
| Basic Echo Function | ✅ Pass | 2.918s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.439s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.296s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.038s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.582s |  |
| Search Query Function | ✅ Pass | 5.275s |  |
| Ask Advice Function | ✅ Pass | 1.614s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.364s |  |
| Basic Context Memory Test | ✅ Pass | 1.956s |  |
| Function Argument Memory Test | ✅ Pass | 0.816s |  |
| Function Response Memory Test | ✅ Pass | 0.827s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.101s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.220s |  |
| Penetration Testing Methodology | ✅ Pass | 5.329s |  |
| SQL Injection Attack Type | ✅ Pass | 1.034s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.682s |  |
| Penetration Testing Framework | ✅ Pass | 6.363s |  |
| Web Application Security Scanner | ✅ Pass | 0.217s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.527s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.676s

---

### coder (kimi-k2.7-code)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.222s |  |
| Text Transform Uppercase | ✅ Pass | 0.210s |  |
| Count from 1 to 5 | ✅ Pass | 0.208s |  |
| Math Calculation | ✅ Pass | 0.214s |  |
| Basic Echo Function | ✅ Pass | 0.210s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.214s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.737s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.779s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.980s |  |
| Search Query Function | ✅ Pass | 1.959s |  |
| Ask Advice Function | ✅ Pass | 1.882s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.525s |  |
| Basic Context Memory Test | ✅ Pass | 1.567s |  |
| Function Argument Memory Test | ✅ Pass | 1.594s |  |
| Function Response Memory Test | ✅ Pass | 1.481s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.682s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.334s |  |
| Penetration Testing Methodology | ✅ Pass | 2.137s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.048s |  |
| SQL Injection Attack Type | ✅ Pass | 1.848s |  |
| Penetration Testing Framework | ✅ Pass | 3.963s |  |
| Web Application Security Scanner | ✅ Pass | 2.175s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.730s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.857s

---

### installer (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.195s |  |
| Text Transform Uppercase | ✅ Pass | 1.955s |  |
| Count from 1 to 5 | ✅ Pass | 6.687s |  |
| Math Calculation | ✅ Pass | 3.855s |  |
| Basic Echo Function | ✅ Pass | 1.849s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.962s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.923s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.689s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.772s |  |
| Search Query Function | ✅ Pass | 1.840s |  |
| Ask Advice Function | ✅ Pass | 1.972s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.081s |  |
| Basic Context Memory Test | ✅ Pass | 3.618s |  |
| Function Argument Memory Test | ✅ Pass | 2.526s |  |
| Function Response Memory Test | ✅ Pass | 2.659s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.096s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.482s |  |
| Penetration Testing Methodology | ✅ Pass | 12.085s |  |
| SQL Injection Attack Type | ✅ Pass | 2.480s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.315s |  |
| Penetration Testing Framework | ✅ Pass | 9.074s |  |
| Web Application Security Scanner | ✅ Pass | 11.971s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.424s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.023s

---

### pentester (kimi-k2.7-code)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.210s |  |
| Text Transform Uppercase | ✅ Pass | 0.208s |  |
| Count from 1 to 5 | ✅ Pass | 0.214s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.209s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.220s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.391s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.889s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.085s |  |
| Search Query Function | ✅ Pass | 2.094s |  |
| Ask Advice Function | ✅ Pass | 1.860s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.555s |  |
| Basic Context Memory Test | ✅ Pass | 1.725s |  |
| Function Argument Memory Test | ✅ Pass | 1.524s |  |
| Function Response Memory Test | ✅ Pass | 1.178s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.426s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.003s |  |
| Penetration Testing Methodology | ✅ Pass | 1.798s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.907s |  |
| SQL Injection Attack Type | ✅ Pass | 2.046s |  |
| Penetration Testing Framework | ✅ Pass | 2.358s |  |
| Web Application Security Scanner | ✅ Pass | 2.833s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.177s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.658s

---

