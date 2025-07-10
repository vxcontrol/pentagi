# LLM Agent Testing Report

Generated: Thu, 10 Jul 2025 18:51:40 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 18/18 (100.00%) | 1.391s |
| simple_json | gemini-2.0-flash-lite | false | 4/4 (100.00%) | 0.709s |
| primary_agent | gemini-2.5-flash-lite-preview-06-17 | false | 18/18 (100.00%) | 0.605s |
| assistant | gemini-2.5-flash | false | 18/18 (100.00%) | 1.098s |
| generator | gemini-2.5-pro | false | 16/18 (88.89%) | 5.342s |
| refiner | gemini-2.0-flash | false | 18/18 (100.00%) | 0.709s |
| adviser | gemini-2.5-flash | false | 18/18 (100.00%) | 1.176s |
| reflector | gemini-2.5-flash | false | 18/18 (100.00%) | 1.176s |
| searcher | gemini-2.0-flash | false | 18/18 (100.00%) | 0.727s |
| enricher | gemini-2.0-flash | false | 18/18 (100.00%) | 0.599s |
| coder | gemini-2.5-pro | false | 17/18 (94.44%) | 5.305s |
| installer | gemini-2.5-flash-lite-preview-06-17 | false | 18/18 (100.00%) | 0.606s |
| pentester | gemini-2.5-flash-lite-preview-06-17 | false | 17/18 (94.44%) | 0.592s |

**Total**: 216/220 (98.18%) successful tests
**Overall average latency**: 1.594s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.430s |  |
| Text Transform Uppercase | ✅ Pass | 0.506s |  |
| Count from 1 to 5 | ✅ Pass | 0.701s |  |
| Math Calculation | ✅ Pass | 0.595s |  |
| Basic Echo Function | ✅ Pass | 0.731s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.530s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.708s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.781s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.866s |  |
| Search Query Function | ✅ Pass | 0.748s |  |
| Ask Advice Function | ✅ Pass | 0.760s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 12.615s |  |
| Penetration Testing Methodology | ✅ Pass | 0.496s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.029s |  |
| SQL Injection Attack Type | ✅ Pass | 0.553s |  |
| Penetration Testing Framework | ✅ Pass | 0.655s |  |
| Web Application Security Scanner | ✅ Pass | 0.593s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.739s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.391s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 0.660s |  |
| Person Information JSON | ✅ Pass | 0.719s |  |
| User Profile JSON | ✅ Pass | 0.750s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.705s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 0.709s

---

### primary_agent (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.380s |  |
| Text Transform Uppercase | ✅ Pass | 0.445s |  |
| Count from 1 to 5 | ✅ Pass | 0.619s |  |
| Math Calculation | ✅ Pass | 0.528s |  |
| Basic Echo Function | ✅ Pass | 0.565s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.547s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.451s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.632s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.470s |  |
| Search Query Function | ✅ Pass | 0.477s |  |
| Ask Advice Function | ✅ Pass | 0.508s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.491s |  |
| Penetration Testing Methodology | ✅ Pass | 0.473s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.444s |  |
| SQL Injection Attack Type | ✅ Pass | 0.430s |  |
| Penetration Testing Framework | ✅ Pass | 0.462s |  |
| Web Application Security Scanner | ✅ Pass | 0.480s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.472s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.605s

---

### assistant (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.554s |  |
| Text Transform Uppercase | ✅ Pass | 0.889s |  |
| Count from 1 to 5 | ✅ Pass | 0.663s |  |
| Math Calculation | ✅ Pass | 0.605s |  |
| Basic Echo Function | ✅ Pass | 0.947s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.767s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.662s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.734s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.874s |  |
| Search Query Function | ✅ Pass | 0.835s |  |
| Ask Advice Function | ✅ Pass | 0.816s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.885s |  |
| Penetration Testing Methodology | ✅ Pass | 0.804s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.118s |  |
| SQL Injection Attack Type | ✅ Pass | 0.892s |  |
| Penetration Testing Framework | ✅ Pass | 0.902s |  |
| Web Application Security Scanner | ✅ Pass | 1.812s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.988s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.098s

---

### generator (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.198s |  |
| Text Transform Uppercase | ✅ Pass | 3.882s |  |
| Count from 1 to 5 | ✅ Pass | 3.170s |  |
| Math Calculation | ✅ Pass | 2.403s |  |
| Basic Echo Function | ✅ Pass | 2.424s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.644s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.657s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.993s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.829s |  |
| Search Query Function | ✅ Pass | 3.143s |  |
| Ask Advice Function | ✅ Pass | 14.705s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.304s |  |
| Penetration Testing Methodology | ❌ Fail | 7.392s | expected text 'reconnaissance' not found |
| Vulnerability Assessment Tools | ✅ Pass | 14.085s |  |
| SQL Injection Attack Type | ✅ Pass | 4.069s |  |
| Penetration Testing Framework | ❌ Fail | 9.297s | expected text 'exploitation' not found |
| Web Application Security Scanner | ✅ Pass | 11.678s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.276s |  |

**Summary**: 16/18 (88.89%) successful tests

**Average latency**: 5.342s

---

### refiner (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.549s |  |
| Text Transform Uppercase | ✅ Pass | 0.540s |  |
| Count from 1 to 5 | ✅ Pass | 0.478s |  |
| Math Calculation | ✅ Pass | 0.464s |  |
| Basic Echo Function | ✅ Pass | 0.648s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.528s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.564s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.758s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.756s |  |
| Search Query Function | ✅ Pass | 0.644s |  |
| Ask Advice Function | ✅ Pass | 0.583s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.855s |  |
| Penetration Testing Methodology | ✅ Pass | 0.442s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.124s |  |
| SQL Injection Attack Type | ✅ Pass | 0.508s |  |
| Penetration Testing Framework | ✅ Pass | 0.992s |  |
| Web Application Security Scanner | ✅ Pass | 0.667s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.649s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.709s

---

### adviser (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.768s |  |
| Text Transform Uppercase | ✅ Pass | 0.614s |  |
| Count from 1 to 5 | ✅ Pass | 0.732s |  |
| Math Calculation | ✅ Pass | 0.685s |  |
| Basic Echo Function | ✅ Pass | 0.742s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.752s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.680s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.975s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.830s |  |
| Search Query Function | ✅ Pass | 0.876s |  |
| Ask Advice Function | ✅ Pass | 0.807s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.747s |  |
| Penetration Testing Methodology | ✅ Pass | 0.674s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.148s |  |
| SQL Injection Attack Type | ✅ Pass | 0.972s |  |
| Penetration Testing Framework | ✅ Pass | 1.776s |  |
| Web Application Security Scanner | ✅ Pass | 3.345s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.029s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.176s

---

### reflector (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.809s |  |
| Text Transform Uppercase | ✅ Pass | 0.737s |  |
| Count from 1 to 5 | ✅ Pass | 0.991s |  |
| Math Calculation | ✅ Pass | 0.782s |  |
| Basic Echo Function | ✅ Pass | 0.956s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.730s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.763s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.839s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.781s |  |
| Search Query Function | ✅ Pass | 0.818s |  |
| Ask Advice Function | ✅ Pass | 0.738s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.029s |  |
| Penetration Testing Methodology | ✅ Pass | 0.712s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.478s |  |
| SQL Injection Attack Type | ✅ Pass | 0.859s |  |
| Penetration Testing Framework | ✅ Pass | 3.154s |  |
| Web Application Security Scanner | ✅ Pass | 0.989s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.983s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.176s

---

### searcher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.404s |  |
| Text Transform Uppercase | ✅ Pass | 0.613s |  |
| Count from 1 to 5 | ✅ Pass | 0.439s |  |
| Math Calculation | ✅ Pass | 0.419s |  |
| Basic Echo Function | ✅ Pass | 0.560s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.454s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.482s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.718s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.716s |  |
| Search Query Function | ✅ Pass | 0.545s |  |
| Ask Advice Function | ✅ Pass | 0.602s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.676s |  |
| Penetration Testing Methodology | ✅ Pass | 0.460s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.685s |  |
| SQL Injection Attack Type | ✅ Pass | 0.435s |  |
| Penetration Testing Framework | ✅ Pass | 0.544s |  |
| Web Application Security Scanner | ✅ Pass | 0.698s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.627s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.727s

---

### enricher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.393s |  |
| Text Transform Uppercase | ✅ Pass | 0.654s |  |
| Count from 1 to 5 | ✅ Pass | 0.505s |  |
| Math Calculation | ✅ Pass | 0.440s |  |
| Basic Echo Function | ✅ Pass | 0.511s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.465s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.486s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.572s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.674s |  |
| Search Query Function | ✅ Pass | 0.494s |  |
| Ask Advice Function | ✅ Pass | 0.558s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.707s |  |
| Penetration Testing Methodology | ✅ Pass | 0.452s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.471s |  |
| SQL Injection Attack Type | ✅ Pass | 0.496s |  |
| Penetration Testing Framework | ✅ Pass | 0.532s |  |
| Web Application Security Scanner | ✅ Pass | 0.691s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.667s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.599s

---

### coder (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.807s |  |
| Text Transform Uppercase | ✅ Pass | 3.393s |  |
| Count from 1 to 5 | ✅ Pass | 2.383s |  |
| Math Calculation | ✅ Pass | 2.549s |  |
| Basic Echo Function | ✅ Pass | 2.741s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.192s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.095s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.954s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.883s |  |
| Search Query Function | ✅ Pass | 2.389s |  |
| Ask Advice Function | ✅ Pass | 13.172s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.013s |  |
| Penetration Testing Methodology | ❌ Fail | 7.377s | expected text 'reconnaissance' not found |
| SQL Injection Attack Type | ✅ Pass | 5.136s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.127s |  |
| Penetration Testing Framework | ✅ Pass | 11.197s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.360s |  |
| Web Application Security Scanner | ✅ Pass | 11.714s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 5.305s

---

### installer (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.062s |  |
| Text Transform Uppercase | ✅ Pass | 0.541s |  |
| Count from 1 to 5 | ✅ Pass | 0.597s |  |
| Math Calculation | ✅ Pass | 0.569s |  |
| Basic Echo Function | ✅ Pass | 0.619s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.456s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.584s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.593s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.496s |  |
| Search Query Function | ✅ Pass | 0.498s |  |
| Ask Advice Function | ✅ Pass | 0.499s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.460s |  |
| Penetration Testing Methodology | ✅ Pass | 0.533s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.459s |  |
| SQL Injection Attack Type | ✅ Pass | 0.480s |  |
| Penetration Testing Framework | ✅ Pass | 0.426s |  |
| Web Application Security Scanner | ✅ Pass | 0.566s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.466s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.606s

---

### pentester (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.844s |  |
| Text Transform Uppercase | ✅ Pass | 0.501s |  |
| Count from 1 to 5 | ✅ Pass | 0.677s |  |
| Math Calculation | ✅ Pass | 0.450s |  |
| Basic Echo Function | ✅ Pass | 0.680s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.653s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.529s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.553s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.511s |  |
| Search Query Function | ✅ Pass | 0.476s |  |
| Ask Advice Function | ❌ Fail | 0.692s | expected 1 tool calls, got 0 |
| Streaming Search Query Function Streaming | ✅ Pass | 0.505s |  |
| Penetration Testing Methodology | ✅ Pass | 0.519s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.031s |  |
| SQL Injection Attack Type | ✅ Pass | 0.457s |  |
| Penetration Testing Framework | ✅ Pass | 0.509s |  |
| Web Application Security Scanner | ✅ Pass | 0.523s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.537s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 0.592s

---

