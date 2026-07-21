# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 17:06:14 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | qwen3.5-flash | false | 23/23 (100.00%) | 1.089s |
| simple_json | qwen3.5-flash | false | 7/7 (100.00%) | 1.073s |
| primary_agent | qwen3.6-plus | true | 23/23 (100.00%) | 5.687s |
| assistant | qwen3.6-plus | true | 23/23 (100.00%) | 6.443s |
| generator | qwen3.7-max | true | 23/23 (100.00%) | 5.511s |
| refiner | qwen3.7-max | true | 23/23 (100.00%) | 5.634s |
| adviser | qwen3.7-max | true | 23/23 (100.00%) | 5.838s |
| reflector | qwen3.5-flash | true | 23/23 (100.00%) | 1.094s |
| searcher | qwen3.5-flash | true | 23/23 (100.00%) | 1.055s |
| enricher | qwen3.5-flash | true | 22/23 (95.65%) | 0.423s |
| coder | qwen3-coder-plus | true | 23/23 (100.00%) | 1.732s |
| installer | qwen3-coder-flash | true | 21/23 (91.30%) | 1.399s |
| pentester | qwen3.6-plus | true | 23/23 (100.00%) | 4.053s |

**Total**: 280/283 (98.94%) successful tests
**Overall average latency**: 3.274s

## Detailed Results

### simple (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.206s |  |
| Text Transform Uppercase | ✅ Pass | 0.678s |  |
| Count from 1 to 5 | ✅ Pass | 0.704s |  |
| Math Calculation | ✅ Pass | 0.688s |  |
| Basic Echo Function | ✅ Pass | 0.793s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.576s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.800s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.800s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.891s |  |
| Search Query Function | ✅ Pass | 0.779s |  |
| Ask Advice Function | ✅ Pass | 0.904s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.989s |  |
| Basic Context Memory Test | ✅ Pass | 1.027s |  |
| Function Argument Memory Test | ✅ Pass | 1.263s |  |
| Function Response Memory Test | ✅ Pass | 0.818s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.280s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.760s |  |
| Penetration Testing Methodology | ✅ Pass | 2.263s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.873s |  |
| SQL Injection Attack Type | ✅ Pass | 0.970s |  |
| Penetration Testing Framework | ✅ Pass | 1.986s |  |
| Web Application Security Scanner | ✅ Pass | 1.829s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.171s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.089s

---

### simple_json (qwen3.5-flash)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.133s |  |
| Project Information JSON | ✅ Pass | 0.909s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.825s |  |
| User Profile JSON | ✅ Pass | 1.141s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.796s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.753s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.948s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 1.073s

---

### primary_agent (qwen3.6-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.698s |  |
| Text Transform Uppercase | ✅ Pass | 4.584s |  |
| Count from 1 to 5 | ✅ Pass | 5.365s |  |
| Math Calculation | ✅ Pass | 3.524s |  |
| Basic Echo Function | ✅ Pass | 5.089s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.090s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.484s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.239s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.262s |  |
| Search Query Function | ✅ Pass | 5.559s |  |
| Ask Advice Function | ✅ Pass | 3.116s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.530s |  |
| Basic Context Memory Test | ✅ Pass | 5.453s |  |
| Function Argument Memory Test | ✅ Pass | 2.121s |  |
| Function Response Memory Test | ✅ Pass | 5.466s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.208s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.918s |  |
| Penetration Testing Methodology | ✅ Pass | 10.966s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.440s |  |
| SQL Injection Attack Type | ✅ Pass | 5.690s |  |
| Penetration Testing Framework | ✅ Pass | 14.808s |  |
| Web Application Security Scanner | ✅ Pass | 13.232s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.953s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.687s

---

### assistant (qwen3.6-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.941s |  |
| Text Transform Uppercase | ✅ Pass | 5.097s |  |
| Count from 1 to 5 | ✅ Pass | 5.431s |  |
| Math Calculation | ✅ Pass | 3.949s |  |
| Basic Echo Function | ✅ Pass | 3.107s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.569s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.479s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.581s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.656s |  |
| Search Query Function | ✅ Pass | 4.222s |  |
| Ask Advice Function | ✅ Pass | 2.965s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.803s |  |
| Basic Context Memory Test | ✅ Pass | 5.119s |  |
| Function Argument Memory Test | ✅ Pass | 3.266s |  |
| Function Response Memory Test | ✅ Pass | 3.152s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.938s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.738s |  |
| Penetration Testing Methodology | ✅ Pass | 11.137s |  |
| SQL Injection Attack Type | ✅ Pass | 5.517s |  |
| Vulnerability Assessment Tools | ✅ Pass | 30.782s |  |
| Penetration Testing Framework | ✅ Pass | 16.326s |  |
| Web Application Security Scanner | ✅ Pass | 11.357s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.043s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.443s

---

### generator (qwen3.7-max)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.825s |  |
| Text Transform Uppercase | ✅ Pass | 4.570s |  |
| Math Calculation | ✅ Pass | 2.269s |  |
| Count from 1 to 5 | ✅ Pass | 8.635s |  |
| Basic Echo Function | ✅ Pass | 2.687s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.671s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.068s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 7.738s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.230s |  |
| Search Query Function | ✅ Pass | 2.312s |  |
| Ask Advice Function | ✅ Pass | 2.780s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.880s |  |
| Basic Context Memory Test | ✅ Pass | 5.618s |  |
| Function Argument Memory Test | ✅ Pass | 5.056s |  |
| Function Response Memory Test | ✅ Pass | 6.050s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.941s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.151s |  |
| Penetration Testing Methodology | ✅ Pass | 8.502s |  |
| SQL Injection Attack Type | ✅ Pass | 6.876s |  |
| Penetration Testing Framework | ✅ Pass | 5.269s |  |
| Vulnerability Assessment Tools | ✅ Pass | 25.061s |  |
| Web Application Security Scanner | ✅ Pass | 4.410s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.144s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.511s

---

### refiner (qwen3.7-max)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.902s |  |
| Text Transform Uppercase | ✅ Pass | 4.827s |  |
| Math Calculation | ✅ Pass | 2.847s |  |
| Count from 1 to 5 | ✅ Pass | 8.910s |  |
| Basic Echo Function | ✅ Pass | 2.504s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.818s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.238s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.566s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 2.115s |  |
| JSON Response Function | ✅ Pass | 7.494s |  |
| Ask Advice Function | ✅ Pass | 3.308s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.229s |  |
| Basic Context Memory Test | ✅ Pass | 4.650s |  |
| Function Argument Memory Test | ✅ Pass | 5.215s |  |
| Function Response Memory Test | ✅ Pass | 5.734s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.895s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.718s |  |
| Penetration Testing Methodology | ✅ Pass | 8.927s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.822s |  |
| SQL Injection Attack Type | ✅ Pass | 5.127s |  |
| Penetration Testing Framework | ✅ Pass | 5.584s |  |
| Web Application Security Scanner | ✅ Pass | 7.260s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.882s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.634s

---

### adviser (qwen3.7-max)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.759s |  |
| Text Transform Uppercase | ✅ Pass | 4.685s |  |
| Count from 1 to 5 | ✅ Pass | 7.905s |  |
| Math Calculation | ✅ Pass | 4.080s |  |
| Basic Echo Function | ✅ Pass | 2.272s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.686s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.681s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.691s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.129s |  |
| Search Query Function | ✅ Pass | 2.579s |  |
| Ask Advice Function | ✅ Pass | 2.710s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.434s |  |
| Basic Context Memory Test | ✅ Pass | 6.015s |  |
| Function Argument Memory Test | ✅ Pass | 6.112s |  |
| Function Response Memory Test | ✅ Pass | 5.051s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.913s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 7.863s |  |
| Penetration Testing Methodology | ✅ Pass | 6.513s |  |
| SQL Injection Attack Type | ✅ Pass | 4.288s |  |
| Vulnerability Assessment Tools | ✅ Pass | 23.066s |  |
| Penetration Testing Framework | ✅ Pass | 7.124s |  |
| Web Application Security Scanner | ✅ Pass | 9.098s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.607s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.838s

---

### reflector (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.435s |  |
| Text Transform Uppercase | ✅ Pass | 0.857s |  |
| Count from 1 to 5 | ✅ Pass | 0.930s |  |
| Math Calculation | ✅ Pass | 0.622s |  |
| Basic Echo Function | ✅ Pass | 0.872s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.657s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.622s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.439s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.847s |  |
| Search Query Function | ✅ Pass | 0.758s |  |
| Ask Advice Function | ✅ Pass | 0.927s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.968s |  |
| Basic Context Memory Test | ✅ Pass | 1.015s |  |
| Function Argument Memory Test | ✅ Pass | 0.950s |  |
| Function Response Memory Test | ✅ Pass | 0.617s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.238s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.620s |  |
| Penetration Testing Methodology | ✅ Pass | 1.926s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.342s |  |
| SQL Injection Attack Type | ✅ Pass | 0.764s |  |
| Penetration Testing Framework | ✅ Pass | 1.709s |  |
| Web Application Security Scanner | ✅ Pass | 1.950s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.090s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.094s

---

### searcher (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.237s |  |
| Text Transform Uppercase | ✅ Pass | 0.869s |  |
| Count from 1 to 5 | ✅ Pass | 0.895s |  |
| Math Calculation | ✅ Pass | 0.217s |  |
| Basic Echo Function | ✅ Pass | 0.802s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.225s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.991s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.370s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.451s |  |
| Search Query Function | ✅ Pass | 0.221s |  |
| Ask Advice Function | ✅ Pass | 0.919s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.839s |  |
| Basic Context Memory Test | ✅ Pass | 1.333s |  |
| Function Argument Memory Test | ✅ Pass | 0.685s |  |
| Function Response Memory Test | ✅ Pass | 0.755s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.358s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.639s |  |
| Penetration Testing Methodology | ✅ Pass | 2.105s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.267s |  |
| SQL Injection Attack Type | ✅ Pass | 0.217s |  |
| Penetration Testing Framework | ✅ Pass | 1.696s |  |
| Web Application Security Scanner | ✅ Pass | 2.174s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.987s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.055s

---

### enricher (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.216s |  |
| Text Transform Uppercase | ✅ Pass | 0.642s |  |
| Count from 1 to 5 | ✅ Pass | 0.210s |  |
| Math Calculation | ✅ Pass | 0.220s |  |
| Basic Echo Function | ✅ Pass | 0.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.216s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.204s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.417s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.179s |  |
| Search Query Function | ✅ Pass | 0.216s |  |
| Ask Advice Function | ✅ Pass | 0.211s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.214s |  |
| Basic Context Memory Test | ✅ Pass | 0.279s |  |
| Function Argument Memory Test | ✅ Pass | 0.213s |  |
| Function Response Memory Test | ✅ Pass | 0.228s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.207s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.217s |  |
| Penetration Testing Methodology | ✅ Pass | 0.211s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.240s |  |
| SQL Injection Attack Type | ✅ Pass | 0.213s |  |
| Penetration Testing Framework | ✅ Pass | 1.309s |  |
| Web Application Security Scanner | ✅ Pass | 0.217s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.213s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.423s

---

### coder (qwen3-coder-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.002s |  |
| Text Transform Uppercase | ✅ Pass | 0.987s |  |
| Count from 1 to 5 | ✅ Pass | 1.064s |  |
| Math Calculation | ✅ Pass | 1.040s |  |
| Basic Echo Function | ✅ Pass | 1.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.038s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.502s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.672s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.275s |  |
| Search Query Function | ✅ Pass | 1.327s |  |
| Ask Advice Function | ✅ Pass | 1.423s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.873s |  |
| Basic Context Memory Test | ✅ Pass | 1.080s |  |
| Function Argument Memory Test | ✅ Pass | 1.052s |  |
| Function Response Memory Test | ✅ Pass | 1.020s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.004s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.048s |  |
| Penetration Testing Methodology | ✅ Pass | 3.898s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.180s |  |
| SQL Injection Attack Type | ✅ Pass | 1.730s |  |
| Penetration Testing Framework | ✅ Pass | 4.312s |  |
| Web Application Security Scanner | ✅ Pass | 2.572s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.520s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.732s

---

### installer (qwen3-coder-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.925s |  |
| Text Transform Uppercase | ✅ Pass | 0.924s |  |
| Count from 1 to 5 | ✅ Pass | 1.112s |  |
| Math Calculation | ✅ Pass | 0.917s |  |
| Basic Echo Function | ✅ Pass | 1.052s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.991s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.480s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.999s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.082s |  |
| Search Query Function | ❌ Fail | 1.016s | no tool calls found, expected at least 1 |
| Ask Advice Function | ✅ Pass | 1.419s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 1.374s | no tool calls found, expected at least 1 |
| Basic Context Memory Test | ✅ Pass | 1.308s |  |
| Function Argument Memory Test | ✅ Pass | 0.864s |  |
| Function Response Memory Test | ✅ Pass | 1.073s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.342s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.194s |  |
| Penetration Testing Methodology | ✅ Pass | 3.036s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.095s |  |
| SQL Injection Attack Type | ✅ Pass | 1.001s |  |
| Penetration Testing Framework | ✅ Pass | 1.386s |  |
| Web Application Security Scanner | ✅ Pass | 1.717s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.854s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 1.399s

---

### pentester (qwen3.6-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.094s |  |
| Text Transform Uppercase | ✅ Pass | 4.964s |  |
| Math Calculation | ✅ Pass | 0.212s |  |
| Count from 1 to 5 | ✅ Pass | 5.474s |  |
| Basic Echo Function | ✅ Pass | 3.864s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.113s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.273s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.204s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.828s |  |
| Search Query Function | ✅ Pass | 2.215s |  |
| Ask Advice Function | ✅ Pass | 3.054s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.363s |  |
| Function Argument Memory Test | ✅ Pass | 0.216s |  |
| Basic Context Memory Test | ✅ Pass | 4.736s |  |
| Function Response Memory Test | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.218s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.134s |  |
| Penetration Testing Methodology | ✅ Pass | 11.298s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.796s |  |
| SQL Injection Attack Type | ✅ Pass | 0.220s |  |
| Penetration Testing Framework | ✅ Pass | 13.884s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.216s |  |
| Web Application Security Scanner | ✅ Pass | 8.616s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.053s

---

