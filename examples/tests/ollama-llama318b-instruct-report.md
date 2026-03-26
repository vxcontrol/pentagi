# LLM Agent Testing Report

Generated: Sat, 17 Jan 2026 16:40:42 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.339s |
| simple_json | llama3.1:8b-instruct-q8_0 | false | 5/5 (100.00%) | 0.834s |
| primary_agent | llama3.1:8b-instruct-q8_0 | false | 21/23 (91.30%) | 1.335s |
| assistant | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.328s |
| generator | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.289s |
| refiner | llama3.1:8b-instruct-q8_0 | false | 21/23 (91.30%) | 1.246s |
| adviser | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.253s |
| reflector | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.305s |
| searcher | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.321s |
| enricher | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.320s |
| coder | llama3.1:8b-instruct-q8_0 | false | 21/23 (91.30%) | 1.321s |
| installer | llama3.1:8b-instruct-q8_0 | false | 21/23 (91.30%) | 1.277s |
| pentester | llama3.1:8b-instruct-q8_0 | false | 22/23 (95.65%) | 1.312s |

**Total**: 265/281 (94.31%) successful tests
**Overall average latency**: 1.295s

## Detailed Results

### simple (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.373s |  |
| Text Transform Uppercase | ✅ Pass | 0.250s |  |
| Count from 1 to 5 | ✅ Pass | 0.297s |  |
| Math Calculation | ✅ Pass | 0.265s |  |
| Basic Echo Function | ✅ Pass | 0.388s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.455s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.293s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.403s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.643s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.806s |  |
| Ask Advice Function | ✅ Pass | 0.686s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.853s |  |
| Basic Context Memory Test | ✅ Pass | 1.018s |  |
| Function Argument Memory Test | ✅ Pass | 1.243s |  |
| Function Response Memory Test | ✅ Pass | 0.258s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.709s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.947s |  |
| Penetration Testing Methodology | ✅ Pass | 1.805s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.452s |  |
| SQL Injection Attack Type | ✅ Pass | 5.091s |  |
| Penetration Testing Framework | ✅ Pass | 0.966s |  |
| Web Application Security Scanner | ✅ Pass | 4.471s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.104s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.339s

---

### simple_json (llama3.1:8b-instruct-q8_0)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.721s |  |
| Person Information JSON | ✅ Pass | 0.867s |  |
| Project Information JSON | ✅ Pass | 0.989s |  |
| User Profile JSON | ✅ Pass | 0.978s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.611s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.834s

---

### primary_agent (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.374s |  |
| Text Transform Uppercase | ✅ Pass | 0.263s |  |
| Count from 1 to 5 | ✅ Pass | 0.287s |  |
| Math Calculation | ✅ Pass | 0.228s |  |
| Basic Echo Function | ✅ Pass | 0.582s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.329s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.314s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.523s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.701s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.724s |  |
| Ask Advice Function | ✅ Pass | 0.772s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.757s |  |
| Basic Context Memory Test | ✅ Pass | 1.178s |  |
| Function Argument Memory Test | ✅ Pass | 0.871s |  |
| Function Response Memory Test | ✅ Pass | 0.239s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.945s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.705s |  |
| Penetration Testing Methodology | ✅ Pass | 2.971s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.949s |  |
| SQL Injection Attack Type | ✅ Pass | 3.361s |  |
| Penetration Testing Framework | ✅ Pass | 1.945s |  |
| Web Application Security Scanner | ✅ Pass | 4.296s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.377s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 1.335s

---

### assistant (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.320s |  |
| Text Transform Uppercase | ✅ Pass | 0.278s |  |
| Count from 1 to 5 | ✅ Pass | 0.351s |  |
| Math Calculation | ✅ Pass | 0.234s |  |
| Basic Echo Function | ✅ Pass | 0.648s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.213s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.363s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.606s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.776s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.663s |  |
| Ask Advice Function | ✅ Pass | 0.854s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.659s |  |
| Basic Context Memory Test | ✅ Pass | 1.354s |  |
| Function Argument Memory Test | ✅ Pass | 0.669s |  |
| Function Response Memory Test | ✅ Pass | 0.233s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.226s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.462s |  |
| Penetration Testing Methodology | ✅ Pass | 3.679s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.123s |  |
| SQL Injection Attack Type | ✅ Pass | 2.013s |  |
| Penetration Testing Framework | ✅ Pass | 2.886s |  |
| Web Application Security Scanner | ✅ Pass | 4.149s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.773s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.328s

---

### generator (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.371s |  |
| Text Transform Uppercase | ✅ Pass | 0.287s |  |
| Count from 1 to 5 | ✅ Pass | 0.373s |  |
| Math Calculation | ✅ Pass | 0.245s |  |
| Basic Echo Function | ✅ Pass | 0.475s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.245s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.284s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.684s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.857s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.585s |  |
| Ask Advice Function | ✅ Pass | 0.947s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.567s |  |
| Basic Context Memory Test | ✅ Pass | 1.460s |  |
| Function Argument Memory Test | ✅ Pass | 0.274s |  |
| Function Response Memory Test | ✅ Pass | 0.245s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.535s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.251s |  |
| Penetration Testing Methodology | ✅ Pass | 4.738s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.499s |  |
| SQL Injection Attack Type | ✅ Pass | 0.240s |  |
| Penetration Testing Framework | ✅ Pass | 3.607s |  |
| Web Application Security Scanner | ✅ Pass | 3.937s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.936s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.289s

---

### refiner (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.240s |  |
| Text Transform Uppercase | ✅ Pass | 0.243s |  |
| Count from 1 to 5 | ✅ Pass | 0.325s |  |
| Math Calculation | ✅ Pass | 0.241s |  |
| Basic Echo Function | ✅ Pass | 0.596s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.274s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.271s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.574s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.847s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.592s |  |
| Ask Advice Function | ✅ Pass | 0.924s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.550s |  |
| Basic Context Memory Test | ✅ Pass | 1.093s |  |
| Function Argument Memory Test | ✅ Pass | 0.286s |  |
| Function Response Memory Test | ✅ Pass | 0.233s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.316s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.257s |  |
| Penetration Testing Methodology | ✅ Pass | 4.687s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.461s |  |
| SQL Injection Attack Type | ✅ Pass | 0.250s |  |
| Penetration Testing Framework | ✅ Pass | 4.012s |  |
| Web Application Security Scanner | ✅ Pass | 3.463s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.901s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 1.246s

---

### adviser (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.275s |  |
| Text Transform Uppercase | ✅ Pass | 0.260s |  |
| Count from 1 to 5 | ✅ Pass | 0.308s |  |
| Math Calculation | ✅ Pass | 0.244s |  |
| Basic Echo Function | ✅ Pass | 0.550s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.258s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.274s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.570s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 1.157s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.573s |  |
| Ask Advice Function | ✅ Pass | 0.925s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.546s |  |
| Basic Context Memory Test | ✅ Pass | 1.180s |  |
| Function Argument Memory Test | ✅ Pass | 0.291s |  |
| Function Response Memory Test | ✅ Pass | 0.245s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.473s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.254s |  |
| Penetration Testing Methodology | ✅ Pass | 4.640s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.311s |  |
| SQL Injection Attack Type | ✅ Pass | 0.255s |  |
| Penetration Testing Framework | ✅ Pass | 4.100s |  |
| Web Application Security Scanner | ✅ Pass | 3.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.900s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.253s

---

### reflector (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.245s |  |
| Text Transform Uppercase | ✅ Pass | 0.264s |  |
| Count from 1 to 5 | ✅ Pass | 0.297s |  |
| Math Calculation | ✅ Pass | 0.238s |  |
| Basic Echo Function | ✅ Pass | 0.564s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.243s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.272s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.575s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 1.144s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.575s |  |
| Ask Advice Function | ✅ Pass | 0.927s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.549s |  |
| Basic Context Memory Test | ✅ Pass | 1.260s |  |
| Function Argument Memory Test | ✅ Pass | 0.278s |  |
| Function Response Memory Test | ✅ Pass | 0.242s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.538s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.255s |  |
| Penetration Testing Methodology | ✅ Pass | 5.074s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.746s |  |
| SQL Injection Attack Type | ✅ Pass | 0.248s |  |
| Penetration Testing Framework | ✅ Pass | 4.304s |  |
| Web Application Security Scanner | ✅ Pass | 3.257s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.902s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.305s

---

### searcher (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.257s |  |
| Text Transform Uppercase | ✅ Pass | 0.266s |  |
| Count from 1 to 5 | ✅ Pass | 0.316s |  |
| Math Calculation | ✅ Pass | 0.260s |  |
| Basic Echo Function | ✅ Pass | 0.573s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.235s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.279s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.574s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 1.142s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.575s |  |
| Ask Advice Function | ✅ Pass | 0.924s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.547s |  |
| Basic Context Memory Test | ✅ Pass | 1.481s |  |
| Function Argument Memory Test | ✅ Pass | 0.288s |  |
| Function Response Memory Test | ✅ Pass | 0.270s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.471s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.256s |  |
| Penetration Testing Methodology | ✅ Pass | 5.094s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.750s |  |
| SQL Injection Attack Type | ✅ Pass | 0.266s |  |
| Penetration Testing Framework | ✅ Pass | 4.493s |  |
| Web Application Security Scanner | ✅ Pass | 3.142s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.900s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.321s

---

### enricher (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.239s |  |
| Text Transform Uppercase | ✅ Pass | 0.247s |  |
| Count from 1 to 5 | ✅ Pass | 0.315s |  |
| Math Calculation | ✅ Pass | 0.258s |  |
| Basic Echo Function | ✅ Pass | 0.575s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.243s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.298s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.575s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 1.164s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.556s |  |
| Ask Advice Function | ✅ Pass | 0.923s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.542s |  |
| Basic Context Memory Test | ✅ Pass | 1.482s |  |
| Function Argument Memory Test | ✅ Pass | 0.273s |  |
| Function Response Memory Test | ✅ Pass | 0.267s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.561s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.257s |  |
| Penetration Testing Methodology | ✅ Pass | 4.954s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.721s |  |
| SQL Injection Attack Type | ✅ Pass | 0.268s |  |
| Penetration Testing Framework | ✅ Pass | 4.513s |  |
| Web Application Security Scanner | ✅ Pass | 3.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.900s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.320s

---

### coder (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.243s |  |
| Text Transform Uppercase | ✅ Pass | 0.260s |  |
| Count from 1 to 5 | ✅ Pass | 0.311s |  |
| Math Calculation | ✅ Pass | 0.234s |  |
| Basic Echo Function | ✅ Pass | 0.576s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.268s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.277s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.577s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.878s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.577s |  |
| Ask Advice Function | ✅ Pass | 0.923s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.547s |  |
| Basic Context Memory Test | ✅ Pass | 1.531s |  |
| Function Argument Memory Test | ✅ Pass | 0.276s |  |
| Function Response Memory Test | ✅ Pass | 0.237s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.404s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.257s |  |
| Penetration Testing Methodology | ✅ Pass | 4.907s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.909s |  |
| SQL Injection Attack Type | ✅ Pass | 0.261s |  |
| Penetration Testing Framework | ✅ Pass | 4.538s |  |
| Web Application Security Scanner | ✅ Pass | 3.478s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.900s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 1.321s

---

### installer (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.247s |  |
| Text Transform Uppercase | ✅ Pass | 0.256s |  |
| Count from 1 to 5 | ✅ Pass | 0.302s |  |
| Math Calculation | ✅ Pass | 0.225s |  |
| Basic Echo Function | ✅ Pass | 0.573s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.267s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.276s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.574s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.877s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.580s |  |
| Ask Advice Function | ✅ Pass | 0.925s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.544s |  |
| Basic Context Memory Test | ✅ Pass | 1.422s |  |
| Function Argument Memory Test | ✅ Pass | 0.281s |  |
| Function Response Memory Test | ✅ Pass | 0.250s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.250s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.258s |  |
| Penetration Testing Methodology | ✅ Pass | 4.888s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.426s |  |
| SQL Injection Attack Type | ✅ Pass | 0.258s |  |
| Penetration Testing Framework | ✅ Pass | 4.368s |  |
| Web Application Security Scanner | ✅ Pass | 3.404s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.899s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 1.277s

---

### pentester (llama3.1:8b-instruct-q8_0)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.251s |  |
| Text Transform Uppercase | ✅ Pass | 0.258s |  |
| Count from 1 to 5 | ✅ Pass | 0.312s |  |
| Math Calculation | ✅ Pass | 0.434s |  |
| Basic Echo Function | ✅ Pass | 0.575s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.242s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.270s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.575s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 0.876s | expected function 'respond\_with\_json' not found in tool calls: expected function respond\_with\_json not found in tool calls |
| Search Query Function | ✅ Pass | 0.574s |  |
| Ask Advice Function | ✅ Pass | 0.926s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.545s |  |
| Basic Context Memory Test | ✅ Pass | 1.406s |  |
| Function Argument Memory Test | ✅ Pass | 0.293s |  |
| Function Response Memory Test | ✅ Pass | 0.250s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.249s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.258s |  |
| Penetration Testing Methodology | ✅ Pass | 5.022s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.589s |  |
| SQL Injection Attack Type | ✅ Pass | 0.241s |  |
| Penetration Testing Framework | ✅ Pass | 4.368s |  |
| Web Application Security Scanner | ✅ Pass | 3.743s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.900s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.312s

---

