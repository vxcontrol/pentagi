# LLM Agent Testing Report

Generated: Thu, 05 Mar 2026 17:08:56 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-3.1-flash-lite-preview | true | 23/23 (100.00%) | 1.105s |
| simple_json | gemini-3.1-flash-lite-preview | true | 5/5 (100.00%) | 1.603s |
| primary_agent | gemini-3.1-pro-preview | true | 23/23 (100.00%) | 5.646s |
| assistant | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 6.289s |
| generator | gemini-3.1-pro-preview | true | 23/23 (100.00%) | 7.440s |
| refiner | gemini-3.1-pro-preview | true | 22/23 (95.65%) | 12.764s |
| adviser | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 6.169s |
| reflector | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.045s |
| searcher | gemini-3-flash-preview | true | 23/23 (100.00%) | 1.992s |
| enricher | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.107s |
| coder | gemini-3.1-pro-preview | true | 23/23 (100.00%) | 5.779s |
| installer | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.763s |
| pentester | gemini-3.1-pro-preview | true | 21/23 (91.30%) | 5.733s |

**Total**: 274/281 (97.51%) successful tests
**Overall average latency**: 4.926s

## Detailed Results

### simple (gemini-3.1-flash-lite-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.997s |  |
| Text Transform Uppercase | ✅ Pass | 0.678s |  |
| Count from 1 to 5 | ✅ Pass | 1.306s |  |
| Math Calculation | ✅ Pass | 0.788s |  |
| Basic Echo Function | ✅ Pass | 1.675s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.154s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.903s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.944s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.733s |  |
| Search Query Function | ✅ Pass | 1.855s |  |
| Ask Advice Function | ✅ Pass | 0.980s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.825s |  |
| Basic Context Memory Test | ✅ Pass | 0.683s |  |
| Function Argument Memory Test | ✅ Pass | 0.889s |  |
| Function Response Memory Test | ✅ Pass | 2.236s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.009s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.596s |  |
| Penetration Testing Methodology | ✅ Pass | 0.980s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.341s |  |
| SQL Injection Attack Type | ✅ Pass | 0.655s |  |
| Penetration Testing Framework | ✅ Pass | 1.067s |  |
| Web Application Security Scanner | ✅ Pass | 0.735s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.376s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.105s

---

### simple_json (gemini-3.1-flash-lite-preview)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.034s |  |
| Person Information JSON | ✅ Pass | 0.715s |  |
| User Profile JSON | ✅ Pass | 0.761s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.657s |  |
| Project Information JSON | ✅ Pass | 4.845s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.603s

---

### primary_agent (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.034s |  |
| Text Transform Uppercase | ✅ Pass | 6.435s |  |
| Count from 1 to 5 | ✅ Pass | 5.477s |  |
| Math Calculation | ✅ Pass | 3.502s |  |
| Basic Echo Function | ✅ Pass | 5.140s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.687s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.316s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 6.680s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.011s |  |
| Search Query Function | ✅ Pass | 4.770s |  |
| Ask Advice Function | ✅ Pass | 7.735s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 7.586s |  |
| Basic Context Memory Test | ✅ Pass | 4.296s |  |
| Function Argument Memory Test | ✅ Pass | 5.851s |  |
| Function Response Memory Test | ✅ Pass | 3.933s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.244s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.651s |  |
| Penetration Testing Methodology | ✅ Pass | 6.143s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.006s |  |
| SQL Injection Attack Type | ✅ Pass | 3.981s |  |
| Penetration Testing Framework | ✅ Pass | 5.996s |  |
| Web Application Security Scanner | ✅ Pass | 4.363s |  |
| Penetration Testing Tool Selection | ✅ Pass | 9.017s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.646s

---

### assistant (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.311s |  |
| Text Transform Uppercase | ✅ Pass | 10.801s |  |
| Count from 1 to 5 | ✅ Pass | 10.225s |  |
| Math Calculation | ✅ Pass | 3.895s |  |
| Basic Echo Function | ✅ Pass | 8.776s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.328s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.836s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 11.157s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.437s |  |
| Search Query Function | ✅ Pass | 4.580s |  |
| Ask Advice Function | ✅ Pass | 4.888s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 11.694s | no tool calls found, expected at least 1 |
| Basic Context Memory Test | ✅ Pass | 4.081s |  |
| Function Argument Memory Test | ✅ Pass | 4.616s |  |
| Function Response Memory Test | ✅ Pass | 4.995s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.145s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.072s |  |
| Penetration Testing Methodology | ✅ Pass | 7.007s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.281s |  |
| SQL Injection Attack Type | ✅ Pass | 4.479s |  |
| Penetration Testing Framework | ✅ Pass | 6.102s |  |
| Web Application Security Scanner | ✅ Pass | 6.151s |  |
| Penetration Testing Tool Selection | ❌ Fail | 6.783s | no tool calls found, expected at least 1 |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 6.289s

---

### generator (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.409s |  |
| Text Transform Uppercase | ✅ Pass | 5.281s |  |
| Count from 1 to 5 | ✅ Pass | 5.887s |  |
| Math Calculation | ✅ Pass | 4.106s |  |
| Basic Echo Function | ✅ Pass | 12.134s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.296s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.472s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.108s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 17.754s |  |
| Search Query Function | ✅ Pass | 22.188s |  |
| Ask Advice Function | ✅ Pass | 7.185s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 6.614s |  |
| Basic Context Memory Test | ✅ Pass | 4.104s |  |
| Function Argument Memory Test | ✅ Pass | 5.608s |  |
| Function Response Memory Test | ✅ Pass | 4.502s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.979s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.700s |  |
| Penetration Testing Methodology | ✅ Pass | 8.708s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.459s |  |
| SQL Injection Attack Type | ✅ Pass | 3.890s |  |
| Penetration Testing Framework | ✅ Pass | 10.137s |  |
| Web Application Security Scanner | ✅ Pass | 7.074s |  |
| Penetration Testing Tool Selection | ✅ Pass | 7.520s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.440s

---

### refiner (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.836s |  |
| Text Transform Uppercase | ✅ Pass | 4.510s |  |
| Count from 1 to 5 | ✅ Pass | 4.798s |  |
| Math Calculation | ✅ Pass | 3.319s |  |
| Basic Echo Function | ✅ Pass | 8.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.405s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.426s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.710s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 12.893s |  |
| Search Query Function | ✅ Pass | 5.456s |  |
| Ask Advice Function | ✅ Pass | 14.030s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.218s |  |
| Basic Context Memory Test | ✅ Pass | 4.220s |  |
| Function Argument Memory Test | ✅ Pass | 4.692s |  |
| Function Response Memory Test | ✅ Pass | 4.569s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.465s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.908s |  |
| Penetration Testing Methodology | ✅ Pass | 6.765s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.448s |  |
| Penetration Testing Framework | ✅ Pass | 5.388s |  |
| Web Application Security Scanner | ✅ Pass | 8.114s |  |
| SQL Injection Attack Type | ✅ Pass | 163.281s |  |
| Penetration Testing Tool Selection | ❌ Fail | 3.896s | no tool calls found, expected at least 1 |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 12.764s

---

### adviser (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.233s |  |
| Text Transform Uppercase | ✅ Pass | 5.863s |  |
| Count from 1 to 5 | ✅ Pass | 5.006s |  |
| Math Calculation | ✅ Pass | 3.472s |  |
| Basic Echo Function | ✅ Pass | 9.962s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.602s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 9.473s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.990s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 10.251s |  |
| Search Query Function | ❌ Fail | 5.857s | no tool calls found, expected at least 1 |
| Ask Advice Function | ✅ Pass | 4.049s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 5.435s | no tool calls found, expected at least 1 |
| Basic Context Memory Test | ✅ Pass | 4.114s |  |
| Function Argument Memory Test | ✅ Pass | 4.434s |  |
| Function Response Memory Test | ✅ Pass | 4.202s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.379s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.014s |  |
| Penetration Testing Methodology | ✅ Pass | 9.402s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.589s |  |
| SQL Injection Attack Type | ✅ Pass | 6.598s |  |
| Penetration Testing Framework | ✅ Pass | 7.364s |  |
| Web Application Security Scanner | ✅ Pass | 5.184s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.397s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 6.169s

---

### reflector (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.522s |  |
| Text Transform Uppercase | ✅ Pass | 1.702s |  |
| Count from 1 to 5 | ✅ Pass | 2.115s |  |
| Math Calculation | ✅ Pass | 1.125s |  |
| Basic Echo Function | ✅ Pass | 1.679s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.487s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.506s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.450s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.182s |  |
| Search Query Function | ✅ Pass | 1.515s |  |
| Ask Advice Function | ✅ Pass | 1.298s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.354s |  |
| Basic Context Memory Test | ✅ Pass | 1.174s |  |
| Function Argument Memory Test | ✅ Pass | 1.423s |  |
| Function Response Memory Test | ✅ Pass | 1.403s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.036s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.681s |  |
| Penetration Testing Methodology | ✅ Pass | 3.639s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.827s |  |
| SQL Injection Attack Type | ✅ Pass | 1.681s |  |
| Penetration Testing Framework | ✅ Pass | 2.840s |  |
| Web Application Security Scanner | ✅ Pass | 2.972s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.421s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.045s

---

### searcher (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.055s |  |
| Text Transform Uppercase | ✅ Pass | 1.362s |  |
| Count from 1 to 5 | ✅ Pass | 1.617s |  |
| Math Calculation | ✅ Pass | 1.431s |  |
| Basic Echo Function | ✅ Pass | 1.369s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.326s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.519s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.820s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.993s |  |
| Search Query Function | ✅ Pass | 1.155s |  |
| Ask Advice Function | ✅ Pass | 1.018s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.403s |  |
| Basic Context Memory Test | ✅ Pass | 2.049s |  |
| Function Argument Memory Test | ✅ Pass | 1.272s |  |
| Function Response Memory Test | ✅ Pass | 1.256s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.351s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.467s |  |
| Penetration Testing Methodology | ✅ Pass | 3.546s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.066s |  |
| SQL Injection Attack Type | ✅ Pass | 1.849s |  |
| Penetration Testing Framework | ✅ Pass | 3.731s |  |
| Web Application Security Scanner | ✅ Pass | 3.150s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.988s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.992s

---

### enricher (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.531s |  |
| Text Transform Uppercase | ✅ Pass | 1.052s |  |
| Count from 1 to 5 | ✅ Pass | 1.923s |  |
| Math Calculation | ✅ Pass | 1.989s |  |
| Basic Echo Function | ✅ Pass | 1.358s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.571s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.678s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.817s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.947s |  |
| Search Query Function | ✅ Pass | 1.491s |  |
| Ask Advice Function | ✅ Pass | 1.126s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.128s |  |
| Basic Context Memory Test | ✅ Pass | 1.206s |  |
| Function Argument Memory Test | ✅ Pass | 1.426s |  |
| Function Response Memory Test | ✅ Pass | 1.258s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.798s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.380s |  |
| Penetration Testing Methodology | ✅ Pass | 3.086s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.220s |  |
| SQL Injection Attack Type | ✅ Pass | 1.592s |  |
| Penetration Testing Framework | ✅ Pass | 3.472s |  |
| Web Application Security Scanner | ✅ Pass | 3.306s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.093s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.107s

---

### coder (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.564s |  |
| Text Transform Uppercase | ✅ Pass | 9.168s |  |
| Count from 1 to 5 | ✅ Pass | 4.297s |  |
| Math Calculation | ✅ Pass | 12.848s |  |
| Basic Echo Function | ✅ Pass | 4.367s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.170s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.534s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.830s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 9.196s |  |
| Search Query Function | ✅ Pass | 4.121s |  |
| Ask Advice Function | ✅ Pass | 5.221s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.875s |  |
| Basic Context Memory Test | ✅ Pass | 4.935s |  |
| Function Argument Memory Test | ✅ Pass | 4.348s |  |
| Function Response Memory Test | ✅ Pass | 4.011s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.053s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.931s |  |
| Penetration Testing Methodology | ✅ Pass | 8.298s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.146s |  |
| SQL Injection Attack Type | ✅ Pass | 4.431s |  |
| Penetration Testing Framework | ✅ Pass | 5.921s |  |
| Web Application Security Scanner | ✅ Pass | 4.735s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.905s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.779s

---

### installer (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.176s |  |
| Text Transform Uppercase | ✅ Pass | 2.361s |  |
| Count from 1 to 5 | ✅ Pass | 3.194s |  |
| Math Calculation | ✅ Pass | 2.707s |  |
| Basic Echo Function | ✅ Pass | 2.371s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.318s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.116s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.306s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.202s |  |
| Search Query Function | ✅ Pass | 2.480s |  |
| Ask Advice Function | ✅ Pass | 1.455s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.719s |  |
| Basic Context Memory Test | ✅ Pass | 2.621s |  |
| Function Argument Memory Test | ✅ Pass | 2.249s |  |
| Function Response Memory Test | ✅ Pass | 2.472s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.575s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.560s |  |
| Penetration Testing Methodology | ✅ Pass | 5.055s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.685s |  |
| SQL Injection Attack Type | ✅ Pass | 2.319s |  |
| Penetration Testing Framework | ✅ Pass | 5.229s |  |
| Web Application Security Scanner | ✅ Pass | 4.249s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.111s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.763s

---

### pentester (gemini-3.1-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.938s |  |
| Text Transform Uppercase | ✅ Pass | 5.110s |  |
| Count from 1 to 5 | ✅ Pass | 4.386s |  |
| Math Calculation | ✅ Pass | 4.925s |  |
| Basic Echo Function | ❌ Fail | 10.105s | no tool calls found, expected at least 1 |
| Streaming Simple Math Streaming | ✅ Pass | 7.901s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.401s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 3.443s | no tool calls found, expected at least 1 |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.250s |  |
| Search Query Function | ✅ Pass | 8.325s |  |
| Ask Advice Function | ✅ Pass | 4.344s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.315s |  |
| Basic Context Memory Test | ✅ Pass | 5.149s |  |
| Function Argument Memory Test | ✅ Pass | 3.930s |  |
| Function Response Memory Test | ✅ Pass | 4.254s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.142s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.946s |  |
| Penetration Testing Methodology | ✅ Pass | 6.441s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.660s |  |
| SQL Injection Attack Type | ✅ Pass | 5.839s |  |
| Penetration Testing Framework | ✅ Pass | 6.380s |  |
| Web Application Security Scanner | ✅ Pass | 8.225s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.434s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 5.733s

---

