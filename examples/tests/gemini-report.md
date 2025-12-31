# LLM Agent Testing Report

Generated: Wed, 31 Dec 2025 07:28:56 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 22/23 (95.65%) | 0.806s |
| simple_json | gemini-2.0-flash-lite | false | 5/5 (100.00%) | 0.751s |
| primary_agent | gemini-2.5-flash | true | 23/23 (100.00%) | 2.702s |
| assistant | gemini-2.5-flash | true | 23/23 (100.00%) | 2.600s |
| generator | gemini-2.5-pro | true | 23/23 (100.00%) | 5.529s |
| refiner | gemini-2.5-flash | true | 23/23 (100.00%) | 2.683s |
| adviser | gemini-2.5-flash | true | 23/23 (100.00%) | 2.749s |
| reflector | gemini-2.0-flash | false | 21/23 (91.30%) | 0.746s |
| searcher | gemini-2.0-flash | false | 22/23 (95.65%) | 0.726s |
| enricher | gemini-2.0-flash | false | 22/23 (95.65%) | 0.739s |
| coder | gemini-2.5-pro | true | 23/23 (100.00%) | 5.272s |
| installer | gemini-2.5-flash-lite | true | 23/23 (100.00%) | 2.687s |
| pentester | gemini-2.5-flash | true | 23/23 (100.00%) | 2.579s |

**Total**: 276/281 (98.22%) successful tests
**Overall average latency**: 2.454s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.364s |  |
| Text Transform Uppercase | ✅ Pass | 0.618s |  |
| Count from 1 to 5 | ✅ Pass | 0.679s |  |
| Math Calculation | ✅ Pass | 0.895s |  |
| Basic Echo Function | ✅ Pass | 0.791s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.619s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.638s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.807s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.907s |  |
| Search Query Function | ✅ Pass | 0.772s |  |
| Ask Advice Function | ✅ Pass | 0.953s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.703s |  |
| Basic Context Memory Test | ✅ Pass | 0.617s |  |
| Function Argument Memory Test | ✅ Pass | 0.672s |  |
| Function Response Memory Test | ✅ Pass | 0.725s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.968s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.581s | expected text 'example\.com' not found |
| Penetration Testing Methodology | ✅ Pass | 1.434s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.628s |  |
| SQL Injection Attack Type | ✅ Pass | 0.621s |  |
| Penetration Testing Framework | ✅ Pass | 1.049s |  |
| Web Application Security Scanner | ✅ Pass | 0.790s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.704s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.806s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.781s |  |
| Vulnerability Report Memory Test | ✅ Pass | 0.955s |  |
| Project Information JSON | ✅ Pass | 0.727s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.617s |  |
| User Profile JSON | ✅ Pass | 0.671s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.751s

---

### primary_agent (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.963s |  |
| Text Transform Uppercase | ✅ Pass | 1.564s |  |
| Count from 1 to 5 | ✅ Pass | 0.679s |  |
| Math Calculation | ✅ Pass | 1.260s |  |
| Basic Echo Function | ✅ Pass | 1.725s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.284s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.689s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.974s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.481s |  |
| Search Query Function | ✅ Pass | 2.028s |  |
| Ask Advice Function | ✅ Pass | 1.654s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.493s |  |
| Basic Context Memory Test | ✅ Pass | 1.975s |  |
| Function Argument Memory Test | ✅ Pass | 1.662s |  |
| Function Response Memory Test | ✅ Pass | 1.940s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.326s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.324s |  |
| Penetration Testing Methodology | ✅ Pass | 4.371s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.897s |  |
| SQL Injection Attack Type | ✅ Pass | 2.345s |  |
| Penetration Testing Framework | ✅ Pass | 7.770s |  |
| Web Application Security Scanner | ✅ Pass | 8.355s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.383s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.702s

---

### assistant (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.218s |  |
| Text Transform Uppercase | ✅ Pass | 1.434s |  |
| Count from 1 to 5 | ✅ Pass | 0.595s |  |
| Math Calculation | ✅ Pass | 0.965s |  |
| Basic Echo Function | ✅ Pass | 1.758s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.589s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.465s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.163s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.370s |  |
| Search Query Function | ✅ Pass | 2.005s |  |
| Ask Advice Function | ✅ Pass | 1.856s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.281s |  |
| Basic Context Memory Test | ✅ Pass | 2.303s |  |
| Function Argument Memory Test | ✅ Pass | 1.539s |  |
| Function Response Memory Test | ✅ Pass | 1.877s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.173s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.751s |  |
| Penetration Testing Methodology | ✅ Pass | 4.326s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.249s |  |
| SQL Injection Attack Type | ✅ Pass | 4.327s |  |
| Penetration Testing Framework | ✅ Pass | 6.789s |  |
| Web Application Security Scanner | ✅ Pass | 6.907s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.857s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.600s

---

### generator (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.320s |  |
| Text Transform Uppercase | ✅ Pass | 5.009s |  |
| Count from 1 to 5 | ✅ Pass | 4.054s |  |
| Math Calculation | ✅ Pass | 2.570s |  |
| Basic Echo Function | ✅ Pass | 3.028s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.074s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.783s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.937s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.935s |  |
| Search Query Function | ✅ Pass | 3.255s |  |
| Ask Advice Function | ✅ Pass | 3.455s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.888s |  |
| Basic Context Memory Test | ✅ Pass | 4.211s |  |
| Function Argument Memory Test | ✅ Pass | 7.922s |  |
| Function Response Memory Test | ✅ Pass | 5.385s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.217s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.559s |  |
| Penetration Testing Methodology | ✅ Pass | 10.242s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.766s |  |
| SQL Injection Attack Type | ✅ Pass | 5.932s |  |
| Penetration Testing Framework | ✅ Pass | 13.203s |  |
| Web Application Security Scanner | ✅ Pass | 10.831s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.573s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.529s

---

### refiner (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.524s |  |
| Text Transform Uppercase | ✅ Pass | 1.866s |  |
| Count from 1 to 5 | ✅ Pass | 1.501s |  |
| Math Calculation | ✅ Pass | 1.016s |  |
| Basic Echo Function | ✅ Pass | 1.808s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.395s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.147s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.450s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.649s |  |
| Search Query Function | ✅ Pass | 1.835s |  |
| Ask Advice Function | ✅ Pass | 1.701s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.119s |  |
| Basic Context Memory Test | ✅ Pass | 2.488s |  |
| Function Argument Memory Test | ✅ Pass | 1.572s |  |
| Function Response Memory Test | ✅ Pass | 2.434s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.923s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.598s |  |
| Penetration Testing Methodology | ✅ Pass | 4.115s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.157s |  |
| SQL Injection Attack Type | ✅ Pass | 2.371s |  |
| Penetration Testing Framework | ✅ Pass | 7.413s |  |
| Web Application Security Scanner | ✅ Pass | 7.120s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.496s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.683s

---

### adviser (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.411s |  |
| Text Transform Uppercase | ✅ Pass | 1.564s |  |
| Count from 1 to 5 | ✅ Pass | 1.319s |  |
| Math Calculation | ✅ Pass | 1.374s |  |
| Basic Echo Function | ✅ Pass | 1.701s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.202s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.014s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.194s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.612s |  |
| Search Query Function | ✅ Pass | 1.379s |  |
| Ask Advice Function | ✅ Pass | 1.960s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.247s |  |
| Basic Context Memory Test | ✅ Pass | 2.192s |  |
| Function Argument Memory Test | ✅ Pass | 1.647s |  |
| Function Response Memory Test | ✅ Pass | 2.557s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.344s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.090s |  |
| Penetration Testing Methodology | ✅ Pass | 5.052s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.814s |  |
| SQL Injection Attack Type | ✅ Pass | 2.699s |  |
| Penetration Testing Framework | ✅ Pass | 6.294s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.347s |  |
| Web Application Security Scanner | ✅ Pass | 8.212s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.749s

---

### reflector (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.168s |  |
| Text Transform Uppercase | ✅ Pass | 0.585s |  |
| Count from 1 to 5 | ✅ Pass | 0.644s |  |
| Math Calculation | ✅ Pass | 0.476s |  |
| Basic Echo Function | ✅ Pass | 0.611s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.539s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.513s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.558s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.746s |  |
| Search Query Function | ✅ Pass | 0.559s |  |
| Ask Advice Function | ✅ Pass | 0.582s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.661s |  |
| Basic Context Memory Test | ✅ Pass | 0.675s |  |
| Function Argument Memory Test | ❌ Fail | 0.799s | expected text 'Go programming language' not found |
| Function Response Memory Test | ✅ Pass | 0.781s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.968s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.445s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 2.027s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.389s |  |
| SQL Injection Attack Type | ✅ Pass | 0.580s |  |
| Penetration Testing Framework | ✅ Pass | 0.569s |  |
| Web Application Security Scanner | ✅ Pass | 0.549s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.728s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 0.746s

---

### searcher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.121s |  |
| Text Transform Uppercase | ✅ Pass | 0.638s |  |
| Count from 1 to 5 | ✅ Pass | 0.530s |  |
| Math Calculation | ✅ Pass | 0.545s |  |
| Basic Echo Function | ✅ Pass | 0.514s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.439s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.506s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.708s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.862s |  |
| Search Query Function | ✅ Pass | 0.541s |  |
| Ask Advice Function | ✅ Pass | 0.575s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.683s |  |
| Basic Context Memory Test | ✅ Pass | 0.705s |  |
| Function Argument Memory Test | ✅ Pass | 0.640s |  |
| Function Response Memory Test | ✅ Pass | 0.682s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.918s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.516s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.873s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.085s |  |
| SQL Injection Attack Type | ✅ Pass | 0.428s |  |
| Penetration Testing Framework | ✅ Pass | 0.500s |  |
| Web Application Security Scanner | ✅ Pass | 0.752s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.923s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.726s

---

### enricher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.421s |  |
| Text Transform Uppercase | ✅ Pass | 0.552s |  |
| Count from 1 to 5 | ✅ Pass | 0.434s |  |
| Math Calculation | ✅ Pass | 0.474s |  |
| Basic Echo Function | ✅ Pass | 0.557s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.607s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.644s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.615s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.738s |  |
| Search Query Function | ✅ Pass | 0.559s |  |
| Ask Advice Function | ✅ Pass | 0.573s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.681s |  |
| Basic Context Memory Test | ✅ Pass | 0.621s |  |
| Function Argument Memory Test | ✅ Pass | 0.584s |  |
| Function Response Memory Test | ✅ Pass | 0.639s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.944s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.417s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.601s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.732s |  |
| SQL Injection Attack Type | ✅ Pass | 0.431s |  |
| Penetration Testing Framework | ✅ Pass | 0.498s |  |
| Web Application Security Scanner | ✅ Pass | 0.796s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.862s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.739s

---

### coder (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.166s |  |
| Text Transform Uppercase | ✅ Pass | 3.982s |  |
| Count from 1 to 5 | ✅ Pass | 3.276s |  |
| Math Calculation | ✅ Pass | 2.823s |  |
| Basic Echo Function | ✅ Pass | 3.123s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.194s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.819s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.459s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.627s |  |
| Search Query Function | ✅ Pass | 3.347s |  |
| Ask Advice Function | ✅ Pass | 2.283s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.822s |  |
| Basic Context Memory Test | ✅ Pass | 5.314s |  |
| Function Argument Memory Test | ✅ Pass | 3.368s |  |
| Function Response Memory Test | ✅ Pass | 4.220s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.660s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.366s |  |
| Penetration Testing Methodology | ✅ Pass | 12.234s |  |
| SQL Injection Attack Type | ✅ Pass | 5.984s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.107s |  |
| Penetration Testing Framework | ✅ Pass | 12.276s |  |
| Web Application Security Scanner | ✅ Pass | 9.753s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.044s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.272s

---

### installer (gemini-2.5-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.625s |  |
| Text Transform Uppercase | ✅ Pass | 1.301s |  |
| Count from 1 to 5 | ✅ Pass | 1.913s |  |
| Math Calculation | ✅ Pass | 1.191s |  |
| Basic Echo Function | ✅ Pass | 1.372s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.142s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.665s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.059s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.348s |  |
| Search Query Function | ✅ Pass | 1.250s |  |
| Ask Advice Function | ✅ Pass | 1.587s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.514s |  |
| Basic Context Memory Test | ✅ Pass | 1.421s |  |
| Function Argument Memory Test | ✅ Pass | 1.412s |  |
| Function Response Memory Test | ✅ Pass | 1.760s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.628s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.564s |  |
| Penetration Testing Methodology | ✅ Pass | 4.774s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.836s |  |
| SQL Injection Attack Type | ✅ Pass | 6.520s |  |
| Penetration Testing Framework | ✅ Pass | 7.052s |  |
| Web Application Security Scanner | ✅ Pass | 5.313s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.547s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.687s

---

### pentester (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.952s |  |
| Text Transform Uppercase | ✅ Pass | 1.406s |  |
| Count from 1 to 5 | ✅ Pass | 1.674s |  |
| Math Calculation | ✅ Pass | 1.314s |  |
| Basic Echo Function | ✅ Pass | 1.638s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.545s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.924s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.179s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.394s |  |
| Search Query Function | ✅ Pass | 1.658s |  |
| Ask Advice Function | ✅ Pass | 1.969s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.704s |  |
| Basic Context Memory Test | ✅ Pass | 1.914s |  |
| Function Argument Memory Test | ✅ Pass | 2.214s |  |
| Function Response Memory Test | ✅ Pass | 3.196s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.613s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.711s |  |
| Penetration Testing Methodology | ✅ Pass | 5.567s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.892s |  |
| SQL Injection Attack Type | ✅ Pass | 3.134s |  |
| Penetration Testing Framework | ✅ Pass | 7.647s |  |
| Web Application Security Scanner | ✅ Pass | 5.268s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.798s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.579s

---

