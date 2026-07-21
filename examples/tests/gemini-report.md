# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 17:10:09 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-3.1-flash-lite | true | 23/23 (100.00%) | 0.572s |
| simple_json | gemini-3.1-flash-lite | true | 7/7 (100.00%) | 0.595s |
| primary_agent | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 4.793s |
| assistant | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 9.515s |
| generator | gemini-3.1-pro-preview | true | 19/23 (82.61%) | 6.179s |
| refiner | gemini-3.1-pro-preview | true | 22/23 (95.65%) | 5.690s |
| adviser | gemini-3.1-pro-preview | true | 19/23 (82.61%) | 4.937s |
| reflector | gemini-3.5-flash | true | 23/23 (100.00%) | 1.967s |
| searcher | gemini-3.5-flash | true | 23/23 (100.00%) | 1.885s |
| enricher | gemini-3.5-flash | true | 23/23 (100.00%) | 1.830s |
| coder | gemini-3.1-pro-preview | true | 22/23 (95.65%) | 5.014s |
| installer | gemini-3.5-flash | true | 23/23 (100.00%) | 2.520s |
| pentester | gemini-3.1-pro-preview | true | 22/23 (95.65%) | 4.466s |

**Total**: 268/283 (94.70%) successful tests
**Overall average latency**: 4.027s

## Detailed Results

### simple (gemini-3.1-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.175s |  |
| Text Transform Uppercase | ✅ Pass | 0.496s |  |
| Count from 1 to 5 | ✅ Pass | 0.551s |  |
| Math Calculation | ✅ Pass | 0.504s |  |
| Basic Echo Function | ✅ Pass | 0.434s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.505s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.488s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.483s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.559s |  |
| Search Query Function | ✅ Pass | 0.492s |  |
| Ask Advice Function | ✅ Pass | 0.529s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.430s |  |
| Basic Context Memory Test | ✅ Pass | 0.440s |  |
| Function Argument Memory Test | ✅ Pass | 0.431s |  |
| Function Response Memory Test | ✅ Pass | 0.438s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.691s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.492s |  |
| Penetration Testing Methodology | ✅ Pass | 0.680s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.928s |  |
| SQL Injection Attack Type | ✅ Pass | 0.649s |  |
| Penetration Testing Framework | ✅ Pass | 0.685s |  |
| Web Application Security Scanner | ✅ Pass | 0.498s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.554s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.572s

---

### simple_json (gemini-3.1-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.680s |  |
| Person Information JSON | ✅ Pass | 0.569s |  |
| Project Information JSON | ✅ Pass | 0.567s |  |
| User Profile JSON | ✅ Pass | 0.561s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.678s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.556s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.553s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 0.595s

---

### primary_agent (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.736s |  |
| Text Transform Uppercase | ✅ Pass | 3.911s |  |
| Count from 1 to 5 | ✅ Pass | 4.733s |  |
| Math Calculation | ✅ Pass | 3.294s |  |
| Basic Echo Function | ✅ Pass | 4.043s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.061s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.591s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.864s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.274s |  |
| Search Query Function | ✅ Pass | 5.765s |  |
| Ask Advice Function | ✅ Pass | 4.155s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.039s |  |
| Basic Context Memory Test | ✅ Pass | 4.656s |  |
| Function Argument Memory Test | ✅ Pass | 3.736s |  |
| Function Response Memory Test | ✅ Pass | 4.103s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 8.039s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.921s |  |
| Penetration Testing Methodology | ✅ Pass | 6.699s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.705s |  |
| SQL Injection Attack Type | ✅ Pass | 3.375s |  |
| Penetration Testing Framework | ✅ Pass | 4.664s |  |
| Web Application Security Scanner | ✅ Pass | 5.470s |  |
| Penetration Testing Tool Selection | ❌ Fail | 12.396s | no tool calls found, expected at least 1 |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 4.793s

---

### assistant (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.922s |  |
| Text Transform Uppercase | ✅ Pass | 4.325s |  |
| Count from 1 to 5 | ✅ Pass | 7.273s |  |
| Math Calculation | ✅ Pass | 2.922s |  |
| Basic Echo Function | ✅ Pass | 5.052s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.366s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.613s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.735s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.776s |  |
| Search Query Function | ✅ Pass | 3.675s |  |
| Ask Advice Function | ✅ Pass | 4.402s |  |
| Basic Context Memory Test | ✅ Pass | 4.029s |  |
| Function Argument Memory Test | ✅ Pass | 3.728s |  |
| Function Response Memory Test | ✅ Pass | 4.033s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 16.108s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.484s |  |
| Penetration Testing Methodology | ✅ Pass | 4.964s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.272s |  |
| SQL Injection Attack Type | ✅ Pass | 4.569s |  |
| Penetration Testing Framework | ✅ Pass | 4.902s |  |
| Web Application Security Scanner | ✅ Pass | 5.659s |  |
| Penetration Testing Tool Selection | ❌ Fail | 16.806s | no tool calls found, expected at least 1 |
| Streaming Search Query Function Streaming | ✅ Pass | 98.207s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 9.515s

---

### generator (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.806s |  |
| Text Transform Uppercase | ✅ Pass | 4.008s |  |
| Count from 1 to 5 | ✅ Pass | 5.580s |  |
| Math Calculation | ✅ Pass | 3.740s |  |
| Basic Echo Function | ✅ Pass | 3.731s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.053s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.228s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 6.705s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.933s |  |
| Search Query Function | ❌ Fail | 6.695s | no tool calls found, expected at least 1 |
| Ask Advice Function | ❌ Fail | 6.017s | no tool calls found, expected at least 1 |
| Streaming Search Query Function Streaming | ✅ Pass | 1.922s |  |
| Basic Context Memory Test | ✅ Pass | 3.532s |  |
| Function Argument Memory Test | ✅ Pass | 3.905s |  |
| Function Response Memory Test | ✅ Pass | 3.967s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.051s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 20.272s | no tool calls found, expected at least 1 |
| Penetration Testing Methodology | ✅ Pass | 7.631s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.752s |  |
| SQL Injection Attack Type | ✅ Pass | 3.358s |  |
| Penetration Testing Framework | ✅ Pass | 6.881s |  |
| Web Application Security Scanner | ✅ Pass | 5.409s |  |
| Penetration Testing Tool Selection | ❌ Fail | 22.927s | no tool calls found, expected at least 1 |

**Summary**: 19/23 (82.61%) successful tests

**Average latency**: 6.179s

---

### refiner (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.555s |  |
| Text Transform Uppercase | ✅ Pass | 3.360s |  |
| Count from 1 to 5 | ✅ Pass | 4.623s |  |
| Math Calculation | ✅ Pass | 2.988s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.184s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.860s |  |
| Basic Echo Function | ✅ Pass | 14.128s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.895s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.466s |  |
| Search Query Function | ✅ Pass | 2.973s |  |
| Ask Advice Function | ✅ Pass | 3.722s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.918s |  |
| Basic Context Memory Test | ✅ Pass | 4.045s |  |
| Function Argument Memory Test | ✅ Pass | 3.300s |  |
| Function Response Memory Test | ✅ Pass | 4.094s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 14.605s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.032s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.394s |  |
| Penetration Testing Methodology | ✅ Pass | 14.148s |  |
| SQL Injection Attack Type | ✅ Pass | 3.666s |  |
| Penetration Testing Framework | ✅ Pass | 4.038s |  |
| Web Application Security Scanner | ✅ Pass | 4.975s |  |
| Penetration Testing Tool Selection | ❌ Fail | 12.892s | no tool calls found, expected at least 1 |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.690s

---

### adviser (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.242s |  |
| Text Transform Uppercase | ✅ Pass | 3.971s |  |
| Count from 1 to 5 | ✅ Pass | 4.435s |  |
| Math Calculation | ✅ Pass | 3.617s |  |
| Basic Echo Function | ✅ Pass | 4.109s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.995s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.637s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 4.687s | no tool calls found, expected at least 1 |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.164s |  |
| Search Query Function | ❌ Fail | 3.340s | no tool calls found, expected at least 1 |
| Ask Advice Function | ✅ Pass | 4.271s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.594s |  |
| Basic Context Memory Test | ✅ Pass | 3.187s |  |
| Function Argument Memory Test | ✅ Pass | 2.980s |  |
| Function Response Memory Test | ✅ Pass | 3.858s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 15.658s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.167s |  |
| Penetration Testing Methodology | ✅ Pass | 6.449s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.278s |  |
| SQL Injection Attack Type | ✅ Pass | 3.853s |  |
| Penetration Testing Framework | ✅ Pass | 5.700s |  |
| Web Application Security Scanner | ✅ Pass | 4.230s |  |
| Penetration Testing Tool Selection | ❌ Fail | 14.125s | no tool calls found, expected at least 1 |

**Summary**: 19/23 (82.61%) successful tests

**Average latency**: 4.937s

---

### reflector (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.856s |  |
| Text Transform Uppercase | ✅ Pass | 1.177s |  |
| Count from 1 to 5 | ✅ Pass | 1.626s |  |
| Math Calculation | ✅ Pass | 0.995s |  |
| Basic Echo Function | ✅ Pass | 1.200s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.000s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.423s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.056s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.257s |  |
| Search Query Function | ✅ Pass | 1.050s |  |
| Ask Advice Function | ✅ Pass | 1.228s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.870s |  |
| Basic Context Memory Test | ✅ Pass | 1.409s |  |
| Function Argument Memory Test | ✅ Pass | 1.798s |  |
| Function Response Memory Test | ✅ Pass | 1.306s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.170s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.111s |  |
| Penetration Testing Methodology | ✅ Pass | 2.948s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.927s |  |
| SQL Injection Attack Type | ✅ Pass | 2.370s |  |
| Penetration Testing Framework | ✅ Pass | 4.652s |  |
| Web Application Security Scanner | ✅ Pass | 1.685s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.114s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.967s

---

### searcher (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.171s |  |
| Text Transform Uppercase | ✅ Pass | 1.366s |  |
| Count from 1 to 5 | ✅ Pass | 1.447s |  |
| Math Calculation | ✅ Pass | 1.371s |  |
| Basic Echo Function | ✅ Pass | 1.324s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.439s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.443s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.493s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.152s |  |
| Search Query Function | ✅ Pass | 1.112s |  |
| Ask Advice Function | ✅ Pass | 1.490s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.065s |  |
| Basic Context Memory Test | ✅ Pass | 1.277s |  |
| Function Argument Memory Test | ✅ Pass | 1.805s |  |
| Function Response Memory Test | ✅ Pass | 1.433s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.231s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.482s |  |
| Penetration Testing Methodology | ✅ Pass | 2.539s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.027s |  |
| SQL Injection Attack Type | ✅ Pass | 1.613s |  |
| Penetration Testing Framework | ✅ Pass | 2.977s |  |
| Web Application Security Scanner | ✅ Pass | 2.852s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.242s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.885s

---

### enricher (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.931s |  |
| Text Transform Uppercase | ✅ Pass | 1.375s |  |
| Count from 1 to 5 | ✅ Pass | 1.435s |  |
| Math Calculation | ✅ Pass | 1.295s |  |
| Basic Echo Function | ✅ Pass | 1.140s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.995s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.424s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.549s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.419s |  |
| Search Query Function | ✅ Pass | 1.175s |  |
| Ask Advice Function | ✅ Pass | 1.304s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.055s |  |
| Basic Context Memory Test | ✅ Pass | 1.304s |  |
| Function Argument Memory Test | ✅ Pass | 1.741s |  |
| Function Response Memory Test | ✅ Pass | 1.550s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.407s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.805s |  |
| Penetration Testing Methodology | ✅ Pass | 2.351s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.706s |  |
| SQL Injection Attack Type | ✅ Pass | 1.863s |  |
| Penetration Testing Framework | ✅ Pass | 1.976s |  |
| Web Application Security Scanner | ✅ Pass | 2.727s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.552s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.830s

---

### coder (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.937s |  |
| Text Transform Uppercase | ✅ Pass | 3.855s |  |
| Count from 1 to 5 | ✅ Pass | 3.427s |  |
| Math Calculation | ✅ Pass | 3.287s |  |
| Basic Echo Function | ✅ Pass | 3.881s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.231s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.236s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.602s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.478s |  |
| Search Query Function | ✅ Pass | 3.723s |  |
| Ask Advice Function | ✅ Pass | 4.353s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.942s |  |
| Basic Context Memory Test | ✅ Pass | 3.981s |  |
| Function Argument Memory Test | ✅ Pass | 2.978s |  |
| Function Response Memory Test | ✅ Pass | 3.865s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 11.925s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.649s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.950s |  |
| Penetration Testing Methodology | ✅ Pass | 19.930s |  |
| SQL Injection Attack Type | ✅ Pass | 4.029s |  |
| Penetration Testing Framework | ✅ Pass | 5.204s |  |
| Web Application Security Scanner | ✅ Pass | 4.520s |  |
| Penetration Testing Tool Selection | ❌ Fail | 5.328s | no tool calls found, expected at least 1 |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.014s

---

### installer (gemini-3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.758s |  |
| Text Transform Uppercase | ✅ Pass | 3.107s |  |
| Count from 1 to 5 | ✅ Pass | 2.562s |  |
| Math Calculation | ✅ Pass | 2.153s |  |
| Basic Echo Function | ✅ Pass | 1.689s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.054s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.487s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.991s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.981s |  |
| Search Query Function | ✅ Pass | 2.041s |  |
| Ask Advice Function | ✅ Pass | 0.928s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.933s |  |
| Basic Context Memory Test | ✅ Pass | 2.175s |  |
| Function Argument Memory Test | ✅ Pass | 3.162s |  |
| Function Response Memory Test | ✅ Pass | 2.878s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.772s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.602s |  |
| Penetration Testing Methodology | ✅ Pass | 3.909s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.209s |  |
| SQL Injection Attack Type | ✅ Pass | 2.117s |  |
| Penetration Testing Framework | ✅ Pass | 3.217s |  |
| Web Application Security Scanner | ✅ Pass | 2.360s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.856s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.520s

---

### pentester (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.191s |  |
| Text Transform Uppercase | ✅ Pass | 3.663s |  |
| Count from 1 to 5 | ✅ Pass | 4.359s |  |
| Math Calculation | ✅ Pass | 2.914s |  |
| Basic Echo Function | ✅ Pass | 3.995s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.860s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.531s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.597s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.768s |  |
| Search Query Function | ✅ Pass | 3.387s |  |
| Ask Advice Function | ✅ Pass | 4.440s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.496s |  |
| Basic Context Memory Test | ✅ Pass | 3.669s |  |
| Function Argument Memory Test | ✅ Pass | 3.852s |  |
| Function Response Memory Test | ✅ Pass | 3.675s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.701s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.324s |  |
| Penetration Testing Methodology | ✅ Pass | 5.095s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.883s |  |
| SQL Injection Attack Type | ✅ Pass | 3.547s |  |
| Penetration Testing Framework | ✅ Pass | 4.535s |  |
| Web Application Security Scanner | ✅ Pass | 5.704s |  |
| Penetration Testing Tool Selection | ❌ Fail | 12.517s | no tool calls found, expected at least 1 |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.466s

---

