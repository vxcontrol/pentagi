# LLM Agent Testing Report

Generated: Wed, 31 Dec 2025 08:43:01 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 21/23 (91.30%) | 0.830s |
| simple_json | gemini-2.0-flash-lite | false | 5/5 (100.00%) | 0.794s |
| primary_agent | gemini-2.5-flash | true | 23/23 (100.00%) | 2.504s |
| assistant | gemini-2.5-pro | true | 23/23 (100.00%) | 5.656s |
| generator | gemini-2.5-pro | true | 23/23 (100.00%) | 5.449s |
| refiner | gemini-2.5-pro | true | 23/23 (100.00%) | 5.408s |
| adviser | gemini-2.5-pro | true | 23/23 (100.00%) | 5.608s |
| reflector | gemini-2.0-flash | false | 21/23 (91.30%) | 0.780s |
| searcher | gemini-2.0-flash | false | 21/23 (91.30%) | 0.712s |
| enricher | gemini-2.0-flash | false | 22/23 (95.65%) | 0.666s |
| coder | gemini-2.5-pro | true | 23/23 (100.00%) | 5.527s |
| installer | gemini-2.5-flash | true | 23/23 (100.00%) | 2.646s |
| pentester | gemini-2.5-pro | true | 23/23 (100.00%) | 4.911s |

**Total**: 274/281 (97.51%) successful tests
**Overall average latency**: 3.345s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.268s |  |
| Text Transform Uppercase | ✅ Pass | 0.662s |  |
| Count from 1 to 5 | ✅ Pass | 0.792s |  |
| Math Calculation | ✅ Pass | 0.596s |  |
| Basic Echo Function | ✅ Pass | 0.722s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.598s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.580s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.691s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.021s |  |
| Search Query Function | ✅ Pass | 0.859s |  |
| Ask Advice Function | ✅ Pass | 0.741s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.776s |  |
| Basic Context Memory Test | ✅ Pass | 0.563s |  |
| Function Argument Memory Test | ❌ Fail | 0.755s | expected text 'Go programming language' not found |
| Function Response Memory Test | ✅ Pass | 0.543s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.941s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.850s | expected text 'example\.com' not found |
| Penetration Testing Methodology | ✅ Pass | 1.229s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.800s |  |
| SQL Injection Attack Type | ✅ Pass | 0.649s |  |
| Penetration Testing Framework | ✅ Pass | 0.926s |  |
| Web Application Security Scanner | ✅ Pass | 0.718s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.798s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 0.830s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.065s |  |
| Person Information JSON | ✅ Pass | 0.686s |  |
| Project Information JSON | ✅ Pass | 0.685s |  |
| User Profile JSON | ✅ Pass | 0.752s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.781s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.794s

---

### primary_agent (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.299s |  |
| Text Transform Uppercase | ✅ Pass | 1.847s |  |
| Count from 1 to 5 | ✅ Pass | 0.431s |  |
| Math Calculation | ✅ Pass | 1.034s |  |
| Basic Echo Function | ✅ Pass | 1.677s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.527s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.389s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.596s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.710s |  |
| Search Query Function | ✅ Pass | 1.802s |  |
| Ask Advice Function | ✅ Pass | 2.217s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.307s |  |
| Basic Context Memory Test | ✅ Pass | 2.192s |  |
| Function Argument Memory Test | ✅ Pass | 2.186s |  |
| Function Response Memory Test | ✅ Pass | 2.140s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.648s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.512s |  |
| Penetration Testing Methodology | ✅ Pass | 3.101s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.848s |  |
| SQL Injection Attack Type | ✅ Pass | 2.219s |  |
| Penetration Testing Framework | ✅ Pass | 7.616s |  |
| Web Application Security Scanner | ✅ Pass | 5.465s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.825s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.504s

---

### assistant (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.932s |  |
| Text Transform Uppercase | ✅ Pass | 5.746s |  |
| Count from 1 to 5 | ✅ Pass | 5.350s |  |
| Math Calculation | ✅ Pass | 2.950s |  |
| Basic Echo Function | ✅ Pass | 2.509s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.059s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.268s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.625s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.290s |  |
| Search Query Function | ✅ Pass | 3.211s |  |
| Ask Advice Function | ✅ Pass | 3.336s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.657s |  |
| Basic Context Memory Test | ✅ Pass | 6.433s |  |
| Function Argument Memory Test | ✅ Pass | 6.351s |  |
| Function Response Memory Test | ✅ Pass | 3.604s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.441s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.471s |  |
| Penetration Testing Methodology | ✅ Pass | 11.383s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.715s |  |
| SQL Injection Attack Type | ✅ Pass | 7.418s |  |
| Penetration Testing Framework | ✅ Pass | 12.961s |  |
| Web Application Security Scanner | ✅ Pass | 11.338s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.040s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.656s

---

### generator (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.432s |  |
| Text Transform Uppercase | ✅ Pass | 3.929s |  |
| Count from 1 to 5 | ✅ Pass | 4.748s |  |
| Math Calculation | ✅ Pass | 3.204s |  |
| Basic Echo Function | ✅ Pass | 2.563s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.390s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.575s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.096s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.746s |  |
| Search Query Function | ✅ Pass | 3.149s |  |
| Ask Advice Function | ✅ Pass | 3.405s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.331s |  |
| Basic Context Memory Test | ✅ Pass | 5.823s |  |
| Function Argument Memory Test | ✅ Pass | 3.395s |  |
| Function Response Memory Test | ✅ Pass | 8.071s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.907s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.464s |  |
| Penetration Testing Methodology | ✅ Pass | 10.379s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.231s |  |
| SQL Injection Attack Type | ✅ Pass | 5.881s |  |
| Penetration Testing Framework | ✅ Pass | 12.434s |  |
| Web Application Security Scanner | ✅ Pass | 9.648s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.515s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.449s

---

### refiner (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.287s |  |
| Text Transform Uppercase | ✅ Pass | 4.345s |  |
| Count from 1 to 5 | ✅ Pass | 3.294s |  |
| Math Calculation | ✅ Pass | 2.564s |  |
| Basic Echo Function | ✅ Pass | 2.965s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.086s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.093s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.879s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.188s |  |
| Search Query Function | ✅ Pass | 3.309s |  |
| Ask Advice Function | ✅ Pass | 2.675s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.208s |  |
| Basic Context Memory Test | ✅ Pass | 4.221s |  |
| Function Argument Memory Test | ✅ Pass | 3.363s |  |
| Function Response Memory Test | ✅ Pass | 6.784s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.967s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.710s |  |
| Penetration Testing Methodology | ✅ Pass | 13.833s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.572s |  |
| SQL Injection Attack Type | ✅ Pass | 7.084s |  |
| Penetration Testing Framework | ✅ Pass | 10.719s |  |
| Web Application Security Scanner | ✅ Pass | 11.415s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.820s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.408s

---

### adviser (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.976s |  |
| Text Transform Uppercase | ✅ Pass | 4.146s |  |
| Count from 1 to 5 | ✅ Pass | 5.262s |  |
| Math Calculation | ✅ Pass | 2.597s |  |
| Basic Echo Function | ✅ Pass | 3.280s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.734s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.490s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.017s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.639s |  |
| Search Query Function | ✅ Pass | 2.921s |  |
| Ask Advice Function | ✅ Pass | 2.482s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.620s |  |
| Basic Context Memory Test | ✅ Pass | 4.823s |  |
| Function Argument Memory Test | ✅ Pass | 3.166s |  |
| Function Response Memory Test | ✅ Pass | 4.711s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.770s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.007s |  |
| Penetration Testing Methodology | ✅ Pass | 13.040s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.538s |  |
| SQL Injection Attack Type | ✅ Pass | 5.348s |  |
| Penetration Testing Framework | ✅ Pass | 15.773s |  |
| Web Application Security Scanner | ✅ Pass | 11.902s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.723s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.608s

---

### reflector (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.071s |  |
| Text Transform Uppercase | ✅ Pass | 0.687s |  |
| Count from 1 to 5 | ✅ Pass | 0.445s |  |
| Math Calculation | ✅ Pass | 0.485s |  |
| Basic Echo Function | ✅ Pass | 0.528s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.476s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.425s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.649s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.867s |  |
| Search Query Function | ✅ Pass | 0.580s |  |
| Ask Advice Function | ✅ Pass | 0.659s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.643s |  |
| Basic Context Memory Test | ✅ Pass | 0.680s |  |
| Function Argument Memory Test | ❌ Fail | 0.690s | expected text 'Go programming language' not found |
| Function Response Memory Test | ✅ Pass | 0.527s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.844s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.431s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.629s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.241s |  |
| SQL Injection Attack Type | ✅ Pass | 0.408s |  |
| Penetration Testing Framework | ✅ Pass | 0.471s |  |
| Web Application Security Scanner | ✅ Pass | 0.752s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.740s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 0.780s

---

### searcher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.114s |  |
| Text Transform Uppercase | ✅ Pass | 0.523s |  |
| Count from 1 to 5 | ✅ Pass | 0.497s |  |
| Math Calculation | ✅ Pass | 0.469s |  |
| Basic Echo Function | ✅ Pass | 0.559s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.439s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.510s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.542s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.663s |  |
| Search Query Function | ✅ Pass | 0.541s |  |
| Ask Advice Function | ✅ Pass | 0.549s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.643s |  |
| Basic Context Memory Test | ✅ Pass | 0.704s |  |
| Function Argument Memory Test | ❌ Fail | 0.737s | expected text 'Go programming language' not found |
| Function Response Memory Test | ✅ Pass | 0.881s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.844s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.451s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.632s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.757s |  |
| SQL Injection Attack Type | ✅ Pass | 0.454s |  |
| Penetration Testing Framework | ✅ Pass | 0.515s |  |
| Web Application Security Scanner | ✅ Pass | 0.590s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.743s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 0.712s

---

### enricher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.433s |  |
| Text Transform Uppercase | ✅ Pass | 0.573s |  |
| Count from 1 to 5 | ✅ Pass | 0.543s |  |
| Math Calculation | ✅ Pass | 0.562s |  |
| Basic Echo Function | ✅ Pass | 0.581s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.444s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.496s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.595s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.747s |  |
| Search Query Function | ✅ Pass | 0.551s |  |
| Ask Advice Function | ✅ Pass | 0.588s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.709s |  |
| Basic Context Memory Test | ✅ Pass | 0.739s |  |
| Function Argument Memory Test | ✅ Pass | 0.542s |  |
| Function Response Memory Test | ✅ Pass | 0.682s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.844s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.491s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.710s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.281s |  |
| SQL Injection Attack Type | ✅ Pass | 0.406s |  |
| Penetration Testing Framework | ✅ Pass | 0.489s |  |
| Web Application Security Scanner | ✅ Pass | 0.505s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.795s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.666s

---

### coder (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.313s |  |
| Text Transform Uppercase | ✅ Pass | 4.114s |  |
| Count from 1 to 5 | ✅ Pass | 5.380s |  |
| Math Calculation | ✅ Pass | 2.424s |  |
| Basic Echo Function | ✅ Pass | 4.567s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.506s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.195s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.409s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.641s |  |
| Search Query Function | ✅ Pass | 5.703s |  |
| Ask Advice Function | ✅ Pass | 5.153s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.988s |  |
| Basic Context Memory Test | ✅ Pass | 5.252s |  |
| Function Argument Memory Test | ✅ Pass | 4.784s |  |
| Function Response Memory Test | ✅ Pass | 5.819s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.309s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.613s |  |
| Penetration Testing Methodology | ✅ Pass | 10.237s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.610s |  |
| SQL Injection Attack Type | ✅ Pass | 5.741s |  |
| Penetration Testing Framework | ✅ Pass | 11.556s |  |
| Web Application Security Scanner | ✅ Pass | 11.622s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.174s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.527s

---

### installer (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.761s |  |
| Text Transform Uppercase | ✅ Pass | 1.464s |  |
| Count from 1 to 5 | ✅ Pass | 0.892s |  |
| Math Calculation | ✅ Pass | 1.581s |  |
| Basic Echo Function | ✅ Pass | 2.126s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.854s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.732s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.490s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.714s |  |
| Search Query Function | ✅ Pass | 1.759s |  |
| Ask Advice Function | ✅ Pass | 1.893s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.661s |  |
| Basic Context Memory Test | ✅ Pass | 1.859s |  |
| Function Argument Memory Test | ✅ Pass | 1.663s |  |
| Function Response Memory Test | ✅ Pass | 2.079s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.505s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.579s |  |
| Penetration Testing Methodology | ✅ Pass | 4.164s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.094s |  |
| SQL Injection Attack Type | ✅ Pass | 7.180s |  |
| Penetration Testing Framework | ✅ Pass | 6.264s |  |
| Web Application Security Scanner | ✅ Pass | 7.140s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.403s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.646s

---

### pentester (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.111s |  |
| Text Transform Uppercase | ✅ Pass | 3.677s |  |
| Count from 1 to 5 | ✅ Pass | 3.134s |  |
| Math Calculation | ✅ Pass | 2.381s |  |
| Basic Echo Function | ✅ Pass | 3.460s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.165s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.758s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.125s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.293s |  |
| Search Query Function | ✅ Pass | 4.133s |  |
| Ask Advice Function | ✅ Pass | 3.697s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.152s |  |
| Basic Context Memory Test | ✅ Pass | 4.960s |  |
| Function Argument Memory Test | ✅ Pass | 4.346s |  |
| Function Response Memory Test | ✅ Pass | 5.360s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.672s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.162s |  |
| Penetration Testing Methodology | ✅ Pass | 10.637s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.567s |  |
| SQL Injection Attack Type | ✅ Pass | 5.285s |  |
| Penetration Testing Framework | ✅ Pass | 9.983s |  |
| Web Application Security Scanner | ✅ Pass | 9.963s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.915s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.911s

---

