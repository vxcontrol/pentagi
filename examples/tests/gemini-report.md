# LLM Agent Testing Report

Generated: Thu, 29 Jan 2026 17:01:10 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 22/23 (95.65%) | 0.713s |
| simple_json | gemini-2.0-flash-lite | false | 5/5 (100.00%) | 0.677s |
| primary_agent | gemini-3-pro-preview | true | 23/23 (100.00%) | 6.606s |
| assistant | gemini-3-pro-preview | true | 23/23 (100.00%) | 6.532s |
| generator | gemini-3-pro-preview | true | 23/23 (100.00%) | 6.326s |
| refiner | gemini-3-pro-preview | true | 23/23 (100.00%) | 5.930s |
| adviser | gemini-3-pro-preview | true | 23/23 (100.00%) | 6.243s |
| reflector | gemini-3-flash-preview | true | 23/23 (100.00%) | 1.961s |
| searcher | gemini-3-flash-preview | true | 23/23 (100.00%) | 1.936s |
| enricher | gemini-3-flash-preview | true | 23/23 (100.00%) | 1.935s |
| coder | gemini-3-pro-preview | true | 23/23 (100.00%) | 6.741s |
| installer | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.305s |
| pentester | gemini-3-pro-preview | true | 23/23 (100.00%) | 6.837s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 4.437s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.982s |  |
| Text Transform Uppercase | ✅ Pass | 0.534s |  |
| Count from 1 to 5 | ✅ Pass | 0.671s |  |
| Math Calculation | ✅ Pass | 0.601s |  |
| Basic Echo Function | ✅ Pass | 0.533s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.510s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.563s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.532s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.636s |  |
| Search Query Function | ✅ Pass | 0.609s |  |
| Ask Advice Function | ✅ Pass | 0.630s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.603s |  |
| Basic Context Memory Test | ✅ Pass | 0.565s |  |
| Function Argument Memory Test | ✅ Pass | 0.851s |  |
| Function Response Memory Test | ✅ Pass | 0.454s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.791s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.711s | expected text 'example\.com' not found |
| Penetration Testing Methodology | ✅ Pass | 1.254s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.670s |  |
| SQL Injection Attack Type | ✅ Pass | 0.526s |  |
| Penetration Testing Framework | ✅ Pass | 0.921s |  |
| Web Application Security Scanner | ✅ Pass | 0.588s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.665s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.713s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.912s |  |
| Person Information JSON | ✅ Pass | 0.668s |  |
| Project Information JSON | ✅ Pass | 0.556s |  |
| User Profile JSON | ✅ Pass | 0.584s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.662s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.677s

---

### primary_agent (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.612s |  |
| Text Transform Uppercase | ✅ Pass | 5.623s |  |
| Count from 1 to 5 | ✅ Pass | 6.415s |  |
| Math Calculation | ✅ Pass | 4.604s |  |
| Basic Echo Function | ✅ Pass | 4.590s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.268s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.870s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.229s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.373s |  |
| Search Query Function | ✅ Pass | 4.574s |  |
| Ask Advice Function | ✅ Pass | 4.373s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.971s |  |
| Basic Context Memory Test | ✅ Pass | 5.628s |  |
| Function Argument Memory Test | ✅ Pass | 4.949s |  |
| Function Response Memory Test | ✅ Pass | 7.293s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.572s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.343s |  |
| Penetration Testing Methodology | ✅ Pass | 11.848s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.534s |  |
| SQL Injection Attack Type | ✅ Pass | 5.026s |  |
| Penetration Testing Framework | ✅ Pass | 13.254s |  |
| Web Application Security Scanner | ✅ Pass | 11.542s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.435s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.606s

---

### assistant (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.927s |  |
| Text Transform Uppercase | ✅ Pass | 6.719s |  |
| Count from 1 to 5 | ✅ Pass | 5.227s |  |
| Math Calculation | ✅ Pass | 3.832s |  |
| Basic Echo Function | ✅ Pass | 3.657s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.955s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.805s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.937s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.262s |  |
| Search Query Function | ✅ Pass | 3.315s |  |
| Ask Advice Function | ✅ Pass | 3.971s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.101s |  |
| Basic Context Memory Test | ✅ Pass | 4.687s |  |
| Function Argument Memory Test | ✅ Pass | 4.635s |  |
| Function Response Memory Test | ✅ Pass | 4.179s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.757s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.322s |  |
| Penetration Testing Methodology | ✅ Pass | 13.323s |  |
| Vulnerability Assessment Tools | ✅ Pass | 25.251s |  |
| SQL Injection Attack Type | ✅ Pass | 4.422s |  |
| Penetration Testing Framework | ✅ Pass | 12.815s |  |
| Web Application Security Scanner | ✅ Pass | 9.831s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.302s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.532s

---

### generator (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.123s |  |
| Text Transform Uppercase | ✅ Pass | 5.756s |  |
| Count from 1 to 5 | ✅ Pass | 6.097s |  |
| Math Calculation | ✅ Pass | 3.933s |  |
| Basic Echo Function | ✅ Pass | 3.708s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.918s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.009s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.217s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.098s |  |
| Search Query Function | ✅ Pass | 3.187s |  |
| Ask Advice Function | ✅ Pass | 4.557s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.988s |  |
| Basic Context Memory Test | ✅ Pass | 4.866s |  |
| Function Argument Memory Test | ✅ Pass | 3.708s |  |
| Function Response Memory Test | ✅ Pass | 5.403s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.927s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.820s |  |
| Penetration Testing Methodology | ✅ Pass | 11.769s |  |
| Vulnerability Assessment Tools | ✅ Pass | 20.617s |  |
| SQL Injection Attack Type | ✅ Pass | 4.766s |  |
| Penetration Testing Framework | ✅ Pass | 10.828s |  |
| Web Application Security Scanner | ✅ Pass | 14.632s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.563s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.326s

---

### refiner (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.643s |  |
| Text Transform Uppercase | ✅ Pass | 5.124s |  |
| Count from 1 to 5 | ✅ Pass | 5.336s |  |
| Math Calculation | ✅ Pass | 3.943s |  |
| Basic Echo Function | ✅ Pass | 3.792s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.471s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.543s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.480s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.476s |  |
| Search Query Function | ✅ Pass | 3.798s |  |
| Ask Advice Function | ✅ Pass | 3.685s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.702s |  |
| Basic Context Memory Test | ✅ Pass | 5.346s |  |
| Function Argument Memory Test | ✅ Pass | 5.051s |  |
| Function Response Memory Test | ✅ Pass | 3.915s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.498s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.082s |  |
| Penetration Testing Methodology | ✅ Pass | 11.024s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.169s |  |
| SQL Injection Attack Type | ✅ Pass | 4.849s |  |
| Penetration Testing Framework | ✅ Pass | 13.487s |  |
| Web Application Security Scanner | ✅ Pass | 10.750s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.204s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.930s

---

### adviser (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.454s |  |
| Text Transform Uppercase | ✅ Pass | 4.399s |  |
| Count from 1 to 5 | ✅ Pass | 5.610s |  |
| Math Calculation | ✅ Pass | 3.650s |  |
| Basic Echo Function | ✅ Pass | 4.427s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.981s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.551s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.587s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.590s |  |
| Search Query Function | ✅ Pass | 3.214s |  |
| Ask Advice Function | ✅ Pass | 5.376s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.864s |  |
| Basic Context Memory Test | ✅ Pass | 5.745s |  |
| Function Argument Memory Test | ✅ Pass | 4.750s |  |
| Function Response Memory Test | ✅ Pass | 5.471s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.728s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.989s |  |
| Penetration Testing Methodology | ✅ Pass | 9.843s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.216s |  |
| SQL Injection Attack Type | ✅ Pass | 4.195s |  |
| Penetration Testing Framework | ✅ Pass | 12.598s |  |
| Web Application Security Scanner | ✅ Pass | 11.163s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.179s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.243s

---

### reflector (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.187s |  |
| Text Transform Uppercase | ✅ Pass | 1.418s |  |
| Count from 1 to 5 | ✅ Pass | 1.682s |  |
| Math Calculation | ✅ Pass | 1.030s |  |
| Basic Echo Function | ✅ Pass | 1.429s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.809s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.269s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.105s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.808s |  |
| Search Query Function | ✅ Pass | 0.968s |  |
| Ask Advice Function | ✅ Pass | 1.478s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.035s |  |
| Basic Context Memory Test | ✅ Pass | 1.395s |  |
| Function Argument Memory Test | ✅ Pass | 1.157s |  |
| Function Response Memory Test | ✅ Pass | 1.237s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.589s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.101s |  |
| Penetration Testing Methodology | ✅ Pass | 3.671s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.632s |  |
| SQL Injection Attack Type | ✅ Pass | 1.432s |  |
| Penetration Testing Framework | ✅ Pass | 3.604s |  |
| Web Application Security Scanner | ✅ Pass | 3.648s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.409s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.961s

---

### searcher (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.326s |  |
| Text Transform Uppercase | ✅ Pass | 1.271s |  |
| Count from 1 to 5 | ✅ Pass | 1.938s |  |
| Math Calculation | ✅ Pass | 1.461s |  |
| Basic Echo Function | ✅ Pass | 1.399s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.879s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.432s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.878s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.348s |  |
| Search Query Function | ✅ Pass | 1.331s |  |
| Ask Advice Function | ✅ Pass | 1.525s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.048s |  |
| Basic Context Memory Test | ✅ Pass | 1.345s |  |
| Function Argument Memory Test | ✅ Pass | 1.205s |  |
| Function Response Memory Test | ✅ Pass | 1.376s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.032s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.438s |  |
| Penetration Testing Methodology | ✅ Pass | 3.120s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.494s |  |
| SQL Injection Attack Type | ✅ Pass | 2.380s |  |
| Penetration Testing Framework | ✅ Pass | 3.624s |  |
| Web Application Security Scanner | ✅ Pass | 2.490s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.171s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.936s

---

### enricher (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.481s |  |
| Text Transform Uppercase | ✅ Pass | 1.466s |  |
| Count from 1 to 5 | ✅ Pass | 2.485s |  |
| Math Calculation | ✅ Pass | 1.165s |  |
| Basic Echo Function | ✅ Pass | 1.479s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.183s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.613s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.348s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.741s |  |
| Search Query Function | ✅ Pass | 1.160s |  |
| Ask Advice Function | ✅ Pass | 1.044s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.402s |  |
| Basic Context Memory Test | ✅ Pass | 1.250s |  |
| Function Argument Memory Test | ✅ Pass | 1.171s |  |
| Function Response Memory Test | ✅ Pass | 1.648s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.583s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.126s |  |
| Penetration Testing Methodology | ✅ Pass | 3.182s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.711s |  |
| SQL Injection Attack Type | ✅ Pass | 1.292s |  |
| Penetration Testing Framework | ✅ Pass | 4.116s |  |
| Web Application Security Scanner | ✅ Pass | 2.294s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.549s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.935s

---

### coder (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.032s |  |
| Text Transform Uppercase | ✅ Pass | 4.882s |  |
| Count from 1 to 5 | ✅ Pass | 6.599s |  |
| Math Calculation | ✅ Pass | 5.324s |  |
| Basic Echo Function | ✅ Pass | 3.620s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.759s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.643s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.176s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 11.434s |  |
| Search Query Function | ✅ Pass | 3.553s |  |
| Ask Advice Function | ✅ Pass | 4.218s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.584s |  |
| Basic Context Memory Test | ✅ Pass | 4.882s |  |
| Function Argument Memory Test | ✅ Pass | 3.452s |  |
| Function Response Memory Test | ✅ Pass | 4.953s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.948s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.386s |  |
| Penetration Testing Methodology | ✅ Pass | 11.711s |  |
| Vulnerability Assessment Tools | ✅ Pass | 24.087s |  |
| SQL Injection Attack Type | ✅ Pass | 4.633s |  |
| Penetration Testing Framework | ✅ Pass | 12.447s |  |
| Web Application Security Scanner | ✅ Pass | 11.867s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.848s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.741s

---

### installer (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.578s |  |
| Text Transform Uppercase | ✅ Pass | 2.568s |  |
| Count from 1 to 5 | ✅ Pass | 2.810s |  |
| Math Calculation | ✅ Pass | 1.872s |  |
| Basic Echo Function | ✅ Pass | 0.945s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.861s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.869s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.003s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.845s |  |
| Search Query Function | ✅ Pass | 0.989s |  |
| Ask Advice Function | ✅ Pass | 1.119s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.903s |  |
| Basic Context Memory Test | ✅ Pass | 0.768s |  |
| Function Argument Memory Test | ✅ Pass | 0.825s |  |
| Function Response Memory Test | ✅ Pass | 0.838s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.351s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.234s |  |
| Penetration Testing Methodology | ✅ Pass | 5.309s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.538s |  |
| SQL Injection Attack Type | ✅ Pass | 2.749s |  |
| Penetration Testing Framework | ✅ Pass | 5.176s |  |
| Web Application Security Scanner | ✅ Pass | 3.745s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.115s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.305s

---

### pentester (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.420s |  |
| Text Transform Uppercase | ✅ Pass | 5.388s |  |
| Count from 1 to 5 | ✅ Pass | 5.135s |  |
| Math Calculation | ✅ Pass | 4.095s |  |
| Basic Echo Function | ✅ Pass | 4.286s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.421s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.814s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.391s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.334s |  |
| Search Query Function | ✅ Pass | 3.590s |  |
| Ask Advice Function | ✅ Pass | 4.237s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.751s |  |
| Basic Context Memory Test | ✅ Pass | 6.268s |  |
| Function Argument Memory Test | ✅ Pass | 5.445s |  |
| Function Response Memory Test | ✅ Pass | 5.529s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.511s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.786s |  |
| Penetration Testing Methodology | ✅ Pass | 12.314s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.233s |  |
| SQL Injection Attack Type | ✅ Pass | 6.576s |  |
| Penetration Testing Framework | ✅ Pass | 10.759s |  |
| Web Application Security Scanner | ✅ Pass | 21.503s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.443s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.837s

---

