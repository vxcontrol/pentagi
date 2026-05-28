# LLM Agent Testing Report

Generated: Wed, 27 May 2026 22:33:44 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-3.1-flash-lite | true | 23/23 (100.00%) | 0.768s |
| simple_json | gemini-3.1-flash-lite | true | 5/5 (100.00%) | 0.715s |
| primary_agent | gemini-3.1-pro-preview | true | 22/23 (95.65%) | 4.190s |
| assistant | gemini-3.1-pro-preview | true | 22/23 (95.65%) | 4.239s |
| generator | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 4.434s |
| refiner | gemini-3.1-pro-preview | true | 23/23 (100.00%) | 4.221s |
| adviser | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 4.304s |
| reflector | gemini-3.5-flash | true | 23/23 (100.00%) | 1.692s |
| searcher | gemini-3.5-flash | true | 23/23 (100.00%) | 1.817s |
| enricher | gemini-3.5-flash | true | 23/23 (100.00%) | 1.724s |
| coder | gemini-3.1-pro-preview | true | 23/23 (100.00%) | 4.103s |
| installer | gemini-3.5-flash | true | 22/23 (95.65%) | 2.891s |
| pentester | gemini-3.1-pro-preview | true | 23/23 (100.00%) | 4.197s |

**Total**: 274/281 (97.51%) successful tests
**Overall average latency**: 3.171s

## Detailed Results

### simple (gemini-3.1-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.367s |  |
| Text Transform Uppercase | ✅ Pass | 0.517s |  |
| Count from 1 to 5 | ✅ Pass | 0.704s |  |
| Math Calculation | ✅ Pass | 0.530s |  |
| Basic Echo Function | ✅ Pass | 0.524s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.506s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.626s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.663s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.578s |  |
| Search Query Function | ✅ Pass | 0.544s |  |
| Ask Advice Function | ✅ Pass | 0.576s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.529s |  |
| Basic Context Memory Test | ✅ Pass | 0.551s |  |
| Function Argument Memory Test | ✅ Pass | 0.490s |  |
| Function Response Memory Test | ✅ Pass | 0.688s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.597s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.546s |  |
| Penetration Testing Methodology | ✅ Pass | 1.218s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.050s |  |
| SQL Injection Attack Type | ✅ Pass | 0.661s |  |
| Penetration Testing Framework | ✅ Pass | 0.762s |  |
| Web Application Security Scanner | ✅ Pass | 0.750s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.673s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.768s

---

### simple_json (gemini-3.1-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.664s |  |
| Vulnerability Report Memory Test | ✅ Pass | 0.961s |  |
| User Profile JSON | ✅ Pass | 0.594s |  |
| Project Information JSON | ✅ Pass | 0.766s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.590s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.715s

---

### primary_agent (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.812s |  |
| Text Transform Uppercase | ✅ Pass | 3.313s |  |
| Count from 1 to 5 | ✅ Pass | 4.086s |  |
| Math Calculation | ✅ Pass | 2.632s |  |
| Basic Echo Function | ❌ Fail | 4.651s | no tool calls found, expected at least 1 |
| Streaming Simple Math Streaming | ✅ Pass | 3.243s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.803s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.800s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.060s |  |
| Search Query Function | ✅ Pass | 4.701s |  |
| Ask Advice Function | ✅ Pass | 6.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.818s |  |
| Basic Context Memory Test | ✅ Pass | 3.449s |  |
| Function Argument Memory Test | ✅ Pass | 4.006s |  |
| Function Response Memory Test | ✅ Pass | 4.055s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.735s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.365s |  |
| Penetration Testing Methodology | ✅ Pass | 5.392s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.336s |  |
| SQL Injection Attack Type | ✅ Pass | 3.569s |  |
| Penetration Testing Framework | ✅ Pass | 4.948s |  |
| Web Application Security Scanner | ✅ Pass | 5.728s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.643s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.190s

---

### assistant (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.818s |  |
| Text Transform Uppercase | ✅ Pass | 3.634s |  |
| Count from 1 to 5 | ✅ Pass | 4.193s |  |
| Math Calculation | ✅ Pass | 3.305s |  |
| Basic Echo Function | ❌ Fail | 5.872s | no tool calls found, expected at least 1 |
| Streaming Simple Math Streaming | ✅ Pass | 3.279s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.548s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.377s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.867s |  |
| Search Query Function | ✅ Pass | 3.219s |  |
| Ask Advice Function | ✅ Pass | 5.921s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.671s |  |
| Basic Context Memory Test | ✅ Pass | 3.188s |  |
| Function Argument Memory Test | ✅ Pass | 4.451s |  |
| Function Response Memory Test | ✅ Pass | 3.309s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.425s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.503s |  |
| Penetration Testing Methodology | ✅ Pass | 5.211s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.501s |  |
| SQL Injection Attack Type | ✅ Pass | 4.066s |  |
| Penetration Testing Framework | ✅ Pass | 5.491s |  |
| Web Application Security Scanner | ✅ Pass | 5.037s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.604s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.239s

---

### generator (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.066s |  |
| Text Transform Uppercase | ✅ Pass | 4.138s |  |
| Count from 1 to 5 | ✅ Pass | 3.613s |  |
| Math Calculation | ✅ Pass | 3.089s |  |
| Basic Echo Function | ❌ Fail | 4.109s | no tool calls found, expected at least 1 |
| Streaming Simple Math Streaming | ✅ Pass | 3.430s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.411s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.347s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.487s |  |
| Search Query Function | ✅ Pass | 3.946s |  |
| Ask Advice Function | ✅ Pass | 4.281s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 3.204s | no tool calls found, expected at least 1 |
| Basic Context Memory Test | ✅ Pass | 4.219s |  |
| Function Argument Memory Test | ✅ Pass | 4.124s |  |
| Function Response Memory Test | ✅ Pass | 3.339s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.760s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.648s |  |
| Penetration Testing Methodology | ✅ Pass | 7.310s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.153s |  |
| SQL Injection Attack Type | ✅ Pass | 3.768s |  |
| Penetration Testing Framework | ✅ Pass | 3.997s |  |
| Web Application Security Scanner | ✅ Pass | 4.903s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.634s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 4.434s

---

### refiner (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.559s |  |
| Text Transform Uppercase | ✅ Pass | 4.160s |  |
| Count from 1 to 5 | ✅ Pass | 3.893s |  |
| Math Calculation | ✅ Pass | 2.809s |  |
| Basic Echo Function | ✅ Pass | 4.024s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.314s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.412s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.594s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.043s |  |
| Search Query Function | ✅ Pass | 3.295s |  |
| Ask Advice Function | ✅ Pass | 4.504s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.448s |  |
| Basic Context Memory Test | ✅ Pass | 3.640s |  |
| Function Argument Memory Test | ✅ Pass | 3.824s |  |
| Function Response Memory Test | ✅ Pass | 4.427s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.450s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.576s |  |
| Penetration Testing Methodology | ✅ Pass | 5.822s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.900s |  |
| SQL Injection Attack Type | ✅ Pass | 3.912s |  |
| Penetration Testing Framework | ✅ Pass | 4.930s |  |
| Web Application Security Scanner | ✅ Pass | 5.459s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.079s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.221s

---

### adviser (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.811s |  |
| Text Transform Uppercase | ✅ Pass | 3.618s |  |
| Count from 1 to 5 | ✅ Pass | 5.524s |  |
| Math Calculation | ✅ Pass | 3.096s |  |
| Basic Echo Function | ✅ Pass | 5.575s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.725s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.862s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 3.410s | no tool calls found, expected at least 1 |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.542s |  |
| Search Query Function | ❌ Fail | 4.056s | no tool calls found, expected at least 1 |
| Ask Advice Function | ✅ Pass | 3.966s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.561s |  |
| Basic Context Memory Test | ✅ Pass | 3.647s |  |
| Function Argument Memory Test | ✅ Pass | 4.782s |  |
| Function Response Memory Test | ✅ Pass | 4.070s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.209s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.835s |  |
| Penetration Testing Methodology | ✅ Pass | 6.464s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.427s |  |
| SQL Injection Attack Type | ✅ Pass | 3.244s |  |
| Penetration Testing Framework | ✅ Pass | 5.763s |  |
| Web Application Security Scanner | ✅ Pass | 5.714s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.082s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 4.304s

---

### reflector (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.016s |  |
| Text Transform Uppercase | ✅ Pass | 1.398s |  |
| Count from 1 to 5 | ✅ Pass | 1.549s |  |
| Math Calculation | ✅ Pass | 1.023s |  |
| Basic Echo Function | ✅ Pass | 1.100s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.953s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.477s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.172s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.330s |  |
| Search Query Function | ✅ Pass | 0.972s |  |
| Ask Advice Function | ✅ Pass | 1.386s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.149s |  |
| Basic Context Memory Test | ✅ Pass | 1.303s |  |
| Function Argument Memory Test | ✅ Pass | 1.155s |  |
| Function Response Memory Test | ✅ Pass | 1.319s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.053s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.048s |  |
| Penetration Testing Methodology | ✅ Pass | 2.124s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.923s |  |
| SQL Injection Attack Type | ✅ Pass | 1.678s |  |
| Penetration Testing Framework | ✅ Pass | 2.756s |  |
| Web Application Security Scanner | ✅ Pass | 1.608s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.421s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.692s

---

### searcher (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.992s |  |
| Text Transform Uppercase | ✅ Pass | 1.363s |  |
| Count from 1 to 5 | ✅ Pass | 1.590s |  |
| Math Calculation | ✅ Pass | 0.971s |  |
| Basic Echo Function | ✅ Pass | 1.336s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.916s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.542s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.124s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.079s |  |
| Search Query Function | ✅ Pass | 1.121s |  |
| Ask Advice Function | ✅ Pass | 1.387s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.260s |  |
| Basic Context Memory Test | ✅ Pass | 1.399s |  |
| Function Argument Memory Test | ✅ Pass | 1.557s |  |
| Function Response Memory Test | ✅ Pass | 1.474s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.472s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.838s |  |
| Penetration Testing Methodology | ✅ Pass | 2.772s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.313s |  |
| SQL Injection Attack Type | ✅ Pass | 1.202s |  |
| Penetration Testing Framework | ✅ Pass | 2.830s |  |
| Web Application Security Scanner | ✅ Pass | 2.120s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.116s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.817s

---

### enricher (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.981s |  |
| Text Transform Uppercase | ✅ Pass | 1.215s |  |
| Count from 1 to 5 | ✅ Pass | 1.425s |  |
| Math Calculation | ✅ Pass | 0.932s |  |
| Basic Echo Function | ✅ Pass | 1.084s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.958s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.604s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.237s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.128s |  |
| Search Query Function | ✅ Pass | 0.993s |  |
| Ask Advice Function | ✅ Pass | 1.090s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.009s |  |
| Basic Context Memory Test | ✅ Pass | 1.461s |  |
| Function Argument Memory Test | ✅ Pass | 0.955s |  |
| Function Response Memory Test | ✅ Pass | 1.856s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.277s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.771s |  |
| Penetration Testing Methodology | ✅ Pass | 2.691s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.802s |  |
| SQL Injection Attack Type | ✅ Pass | 1.611s |  |
| Penetration Testing Framework | ✅ Pass | 2.182s |  |
| Web Application Security Scanner | ✅ Pass | 2.253s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.135s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.724s

---

### coder (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.599s |  |
| Text Transform Uppercase | ✅ Pass | 3.774s |  |
| Count from 1 to 5 | ✅ Pass | 3.470s |  |
| Math Calculation | ✅ Pass | 3.170s |  |
| Basic Echo Function | ✅ Pass | 4.147s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.895s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.859s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.385s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.047s |  |
| Search Query Function | ✅ Pass | 3.097s |  |
| Ask Advice Function | ✅ Pass | 3.666s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.431s |  |
| Basic Context Memory Test | ✅ Pass | 3.474s |  |
| Function Argument Memory Test | ✅ Pass | 3.706s |  |
| Function Response Memory Test | ✅ Pass | 3.504s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.181s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.421s |  |
| Penetration Testing Methodology | ✅ Pass | 5.411s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.503s |  |
| SQL Injection Attack Type | ✅ Pass | 3.996s |  |
| Penetration Testing Framework | ✅ Pass | 4.476s |  |
| Web Application Security Scanner | ✅ Pass | 4.807s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.343s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.103s

---

### installer (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.355s |  |
| Text Transform Uppercase | ✅ Pass | 3.024s |  |
| Count from 1 to 5 | ✅ Pass | 2.930s |  |
| Math Calculation | ✅ Pass | 1.667s |  |
| Basic Echo Function | ✅ Pass | 2.377s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.014s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.497s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.267s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.773s |  |
| Search Query Function | ✅ Pass | 2.193s |  |
| Ask Advice Function | ✅ Pass | 1.839s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.064s |  |
| Basic Context Memory Test | ✅ Pass | 2.661s |  |
| Function Argument Memory Test | ✅ Pass | 2.603s |  |
| Function Response Memory Test | ✅ Pass | 2.563s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 5.230s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.456s |  |
| Penetration Testing Methodology | ✅ Pass | 4.745s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.977s |  |
| SQL Injection Attack Type | ✅ Pass | 2.499s |  |
| Penetration Testing Framework | ✅ Pass | 4.620s |  |
| Web Application Security Scanner | ✅ Pass | 4.105s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.022s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.891s

---

### pentester (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.995s |  |
| Text Transform Uppercase | ✅ Pass | 3.980s |  |
| Count from 1 to 5 | ✅ Pass | 3.836s |  |
| Math Calculation | ✅ Pass | 3.256s |  |
| Basic Echo Function | ✅ Pass | 4.078s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.645s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.757s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.314s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.023s |  |
| Search Query Function | ✅ Pass | 3.725s |  |
| Ask Advice Function | ✅ Pass | 4.173s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.589s |  |
| Basic Context Memory Test | ✅ Pass | 3.997s |  |
| Function Argument Memory Test | ✅ Pass | 4.753s |  |
| Function Response Memory Test | ✅ Pass | 4.861s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.704s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.880s |  |
| Penetration Testing Methodology | ✅ Pass | 6.182s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.410s |  |
| SQL Injection Attack Type | ✅ Pass | 3.714s |  |
| Penetration Testing Framework | ✅ Pass | 4.977s |  |
| Web Application Security Scanner | ✅ Pass | 5.953s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.708s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.197s

---

