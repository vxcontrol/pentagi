# LLM Agent Testing Report

Generated: Wed, 27 May 2026 22:59:39 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek-v4-flash | false | 23/23 (100.00%) | 1.039s |
| simple_json | deepseek-v4-flash | false | 5/5 (100.00%) | 0.989s |
| primary_agent | deepseek-v4-pro | true | 23/23 (100.00%) | 2.896s |
| assistant | deepseek-v4-pro | true | 22/23 (95.65%) | 2.929s |
| generator | deepseek-v4-pro | true | 23/23 (100.00%) | 3.046s |
| refiner | deepseek-v4-pro | true | 23/23 (100.00%) | 2.932s |
| adviser | deepseek-v4-pro | true | 23/23 (100.00%) | 2.667s |
| reflector | deepseek-v4-flash | true | 23/23 (100.00%) | 1.288s |
| searcher | deepseek-v4-flash | true | 23/23 (100.00%) | 0.991s |
| enricher | deepseek-v4-flash | true | 23/23 (100.00%) | 0.981s |
| coder | deepseek-v4-pro | true | 23/23 (100.00%) | 2.817s |
| installer | deepseek-v4-flash | true | 23/23 (100.00%) | 1.246s |
| pentester | deepseek-v4-pro | true | 23/23 (100.00%) | 3.086s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 2.139s

## Detailed Results

### simple (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.973s |  |
| Text Transform Uppercase | ✅ Pass | 0.709s |  |
| Count from 1 to 5 | ✅ Pass | 0.903s |  |
| Math Calculation | ✅ Pass | 0.698s |  |
| Basic Echo Function | ✅ Pass | 1.075s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.775s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.871s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.081s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.254s |  |
| Search Query Function | ✅ Pass | 0.963s |  |
| Ask Advice Function | ✅ Pass | 1.142s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.976s |  |
| Basic Context Memory Test | ✅ Pass | 0.768s |  |
| Function Argument Memory Test | ✅ Pass | 0.865s |  |
| Function Response Memory Test | ✅ Pass | 0.782s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.414s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.847s |  |
| Penetration Testing Methodology | ✅ Pass | 1.116s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.258s |  |
| SQL Injection Attack Type | ✅ Pass | 0.876s |  |
| Penetration Testing Framework | ✅ Pass | 1.179s |  |
| Web Application Security Scanner | ✅ Pass | 1.190s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.180s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.039s

---

### simple_json (deepseek-v4-flash)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.872s |  |
| Project Information JSON | ✅ Pass | 0.944s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.059s |  |
| User Profile JSON | ✅ Pass | 0.976s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.093s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.989s

---

### primary_agent (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.178s |  |
| Text Transform Uppercase | ✅ Pass | 1.700s |  |
| Count from 1 to 5 | ✅ Pass | 4.825s |  |
| Math Calculation | ✅ Pass | 1.798s |  |
| Basic Echo Function | ✅ Pass | 2.080s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.931s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.996s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.133s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.657s |  |
| Search Query Function | ✅ Pass | 2.248s |  |
| Ask Advice Function | ✅ Pass | 2.200s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.060s |  |
| Basic Context Memory Test | ✅ Pass | 3.547s |  |
| Function Argument Memory Test | ✅ Pass | 2.111s |  |
| Function Response Memory Test | ✅ Pass | 2.269s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.453s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.267s |  |
| Penetration Testing Methodology | ✅ Pass | 4.575s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.988s |  |
| SQL Injection Attack Type | ✅ Pass | 2.190s |  |
| Penetration Testing Framework | ✅ Pass | 5.697s |  |
| Web Application Security Scanner | ✅ Pass | 2.219s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.468s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.896s

---

### assistant (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.123s |  |
| Text Transform Uppercase | ✅ Pass | 1.972s |  |
| Count from 1 to 5 | ✅ Pass | 6.415s |  |
| Math Calculation | ✅ Pass | 1.814s |  |
| Basic Echo Function | ✅ Pass | 2.143s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.970s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.243s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.283s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.509s |  |
| Search Query Function | ✅ Pass | 2.421s |  |
| Ask Advice Function | ✅ Pass | 2.316s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.957s |  |
| Basic Context Memory Test | ✅ Pass | 2.203s |  |
| Function Argument Memory Test | ✅ Pass | 2.293s |  |
| Function Response Memory Test | ✅ Pass | 2.042s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 3.318s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.571s |  |
| Penetration Testing Methodology | ✅ Pass | 3.431s |  |
| SQL Injection Attack Type | ✅ Pass | 2.289s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.808s |  |
| Penetration Testing Framework | ✅ Pass | 3.233s |  |
| Web Application Security Scanner | ✅ Pass | 3.258s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.752s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.929s

---

### generator (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.071s |  |
| Text Transform Uppercase | ✅ Pass | 2.231s |  |
| Count from 1 to 5 | ✅ Pass | 6.367s |  |
| Math Calculation | ✅ Pass | 1.514s |  |
| Basic Echo Function | ✅ Pass | 1.995s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.534s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.143s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.014s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.431s |  |
| Search Query Function | ✅ Pass | 2.231s |  |
| Ask Advice Function | ✅ Pass | 2.042s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.115s |  |
| Basic Context Memory Test | ✅ Pass | 2.382s |  |
| Function Argument Memory Test | ✅ Pass | 2.525s |  |
| Function Response Memory Test | ✅ Pass | 2.067s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.271s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.383s |  |
| Penetration Testing Methodology | ✅ Pass | 4.247s |  |
| SQL Injection Attack Type | ✅ Pass | 2.293s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.485s |  |
| Penetration Testing Framework | ✅ Pass | 4.695s |  |
| Web Application Security Scanner | ✅ Pass | 2.208s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.794s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.046s

---

### refiner (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.115s |  |
| Text Transform Uppercase | ✅ Pass | 1.996s |  |
| Count from 1 to 5 | ✅ Pass | 4.321s |  |
| Math Calculation | ✅ Pass | 1.722s |  |
| Basic Echo Function | ✅ Pass | 2.028s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.742s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.044s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.170s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.374s |  |
| Search Query Function | ✅ Pass | 2.454s |  |
| Ask Advice Function | ✅ Pass | 2.166s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.974s |  |
| Basic Context Memory Test | ✅ Pass | 3.341s |  |
| Function Argument Memory Test | ✅ Pass | 2.086s |  |
| Function Response Memory Test | ✅ Pass | 2.400s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.446s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.015s |  |
| Penetration Testing Methodology | ✅ Pass | 3.273s |  |
| SQL Injection Attack Type | ✅ Pass | 3.515s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.480s |  |
| Penetration Testing Framework | ✅ Pass | 4.129s |  |
| Web Application Security Scanner | ✅ Pass | 2.298s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.323s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.932s

---

### adviser (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.074s |  |
| Text Transform Uppercase | ✅ Pass | 1.958s |  |
| Count from 1 to 5 | ✅ Pass | 4.874s |  |
| Math Calculation | ✅ Pass | 1.681s |  |
| Basic Echo Function | ✅ Pass | 2.290s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.667s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.040s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.949s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.484s |  |
| Search Query Function | ✅ Pass | 2.248s |  |
| Ask Advice Function | ✅ Pass | 1.988s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.986s |  |
| Basic Context Memory Test | ✅ Pass | 2.873s |  |
| Function Argument Memory Test | ✅ Pass | 2.211s |  |
| Function Response Memory Test | ✅ Pass | 1.874s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.502s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.064s |  |
| Penetration Testing Methodology | ✅ Pass | 3.537s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.184s |  |
| SQL Injection Attack Type | ✅ Pass | 4.424s |  |
| Penetration Testing Framework | ✅ Pass | 3.939s |  |
| Web Application Security Scanner | ✅ Pass | 2.206s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.267s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.667s

---

### reflector (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.798s |  |
| Text Transform Uppercase | ✅ Pass | 0.888s |  |
| Count from 1 to 5 | ✅ Pass | 7.083s |  |
| Math Calculation | ✅ Pass | 0.887s |  |
| Basic Echo Function | ✅ Pass | 1.023s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.759s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.762s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.272s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.026s |  |
| Search Query Function | ✅ Pass | 0.935s |  |
| Ask Advice Function | ✅ Pass | 1.016s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.064s |  |
| Basic Context Memory Test | ✅ Pass | 0.733s |  |
| Function Argument Memory Test | ✅ Pass | 0.869s |  |
| Function Response Memory Test | ✅ Pass | 0.791s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.358s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.853s |  |
| Penetration Testing Methodology | ✅ Pass | 1.117s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.876s |  |
| SQL Injection Attack Type | ✅ Pass | 0.844s |  |
| Penetration Testing Framework | ✅ Pass | 1.286s |  |
| Web Application Security Scanner | ✅ Pass | 1.159s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.207s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.288s

---

### searcher (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.813s |  |
| Text Transform Uppercase | ✅ Pass | 0.770s |  |
| Count from 1 to 5 | ✅ Pass | 0.893s |  |
| Math Calculation | ✅ Pass | 0.742s |  |
| Basic Echo Function | ✅ Pass | 0.971s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.695s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.929s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.123s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.010s |  |
| Search Query Function | ✅ Pass | 0.922s |  |
| Ask Advice Function | ✅ Pass | 1.081s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.909s |  |
| Basic Context Memory Test | ✅ Pass | 0.755s |  |
| Function Argument Memory Test | ✅ Pass | 0.789s |  |
| Function Response Memory Test | ✅ Pass | 0.742s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.220s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.870s |  |
| Penetration Testing Methodology | ✅ Pass | 0.981s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.004s |  |
| SQL Injection Attack Type | ✅ Pass | 0.899s |  |
| Penetration Testing Framework | ✅ Pass | 1.223s |  |
| Web Application Security Scanner | ✅ Pass | 1.287s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.151s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.991s

---

### enricher (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.759s |  |
| Text Transform Uppercase | ✅ Pass | 0.897s |  |
| Count from 1 to 5 | ✅ Pass | 0.872s |  |
| Math Calculation | ✅ Pass | 0.710s |  |
| Basic Echo Function | ✅ Pass | 0.948s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.870s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.746s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.899s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.144s |  |
| Search Query Function | ✅ Pass | 0.972s |  |
| Ask Advice Function | ✅ Pass | 0.989s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.015s |  |
| Basic Context Memory Test | ✅ Pass | 0.858s |  |
| Function Argument Memory Test | ✅ Pass | 0.726s |  |
| Function Response Memory Test | ✅ Pass | 0.754s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.414s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.851s |  |
| Penetration Testing Methodology | ✅ Pass | 0.991s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.597s |  |
| SQL Injection Attack Type | ✅ Pass | 0.857s |  |
| Penetration Testing Framework | ✅ Pass | 1.269s |  |
| Web Application Security Scanner | ✅ Pass | 1.036s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.377s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.981s

---

### coder (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.791s |  |
| Text Transform Uppercase | ✅ Pass | 6.028s |  |
| Count from 1 to 5 | ✅ Pass | 1.873s |  |
| Math Calculation | ✅ Pass | 1.707s |  |
| Basic Echo Function | ✅ Pass | 2.051s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.951s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.989s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.446s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.494s |  |
| Search Query Function | ✅ Pass | 2.010s |  |
| Ask Advice Function | ✅ Pass | 2.242s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.133s |  |
| Basic Context Memory Test | ✅ Pass | 3.207s |  |
| Function Argument Memory Test | ✅ Pass | 2.075s |  |
| Function Response Memory Test | ✅ Pass | 1.921s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.561s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.345s |  |
| Penetration Testing Methodology | ✅ Pass | 3.598s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.780s |  |
| SQL Injection Attack Type | ✅ Pass | 2.581s |  |
| Penetration Testing Framework | ✅ Pass | 2.960s |  |
| Web Application Security Scanner | ✅ Pass | 2.493s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.548s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.817s

---

### installer (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.185s |  |
| Text Transform Uppercase | ✅ Pass | 1.103s |  |
| Count from 1 to 5 | ✅ Pass | 1.222s |  |
| Math Calculation | ✅ Pass | 0.962s |  |
| Basic Echo Function | ✅ Pass | 0.990s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.908s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.216s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.106s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.264s |  |
| Search Query Function | ✅ Pass | 1.140s |  |
| Ask Advice Function | ✅ Pass | 1.158s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.171s |  |
| Basic Context Memory Test | ✅ Pass | 1.188s |  |
| Function Argument Memory Test | ✅ Pass | 0.905s |  |
| Function Response Memory Test | ✅ Pass | 1.120s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.758s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.140s |  |
| Penetration Testing Methodology | ✅ Pass | 1.382s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.060s |  |
| SQL Injection Attack Type | ✅ Pass | 1.298s |  |
| Penetration Testing Framework | ✅ Pass | 1.423s |  |
| Web Application Security Scanner | ✅ Pass | 1.586s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.352s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.246s

---

### pentester (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.845s |  |
| Text Transform Uppercase | ✅ Pass | 6.297s |  |
| Count from 1 to 5 | ✅ Pass | 1.982s |  |
| Math Calculation | ✅ Pass | 1.653s |  |
| Basic Echo Function | ✅ Pass | 1.838s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.884s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.004s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.209s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.702s |  |
| Search Query Function | ✅ Pass | 1.898s |  |
| Ask Advice Function | ✅ Pass | 2.049s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.145s |  |
| Basic Context Memory Test | ✅ Pass | 3.227s |  |
| Function Argument Memory Test | ✅ Pass | 1.892s |  |
| Function Response Memory Test | ✅ Pass | 2.334s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.631s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.276s |  |
| Penetration Testing Methodology | ✅ Pass | 3.126s |  |
| SQL Injection Attack Type | ✅ Pass | 2.551s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.848s |  |
| Penetration Testing Framework | ✅ Pass | 3.944s |  |
| Web Application Security Scanner | ✅ Pass | 1.612s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.012s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.086s

---

