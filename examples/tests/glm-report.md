# LLM Agent Testing Report

Generated: Thu, 05 Mar 2026 16:50:23 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | glm-4.7-flashx | true | 22/23 (95.65%) | 20.579s |
| simple_json | glm-4.7-flashx | true | 5/5 (100.00%) | 7.107s |
| primary_agent | glm-5 | true | 23/23 (100.00%) | 7.050s |
| assistant | glm-5 | true | 23/23 (100.00%) | 7.197s |
| generator | glm-5 | true | 23/23 (100.00%) | 6.794s |
| refiner | glm-5 | true | 23/23 (100.00%) | 7.235s |
| adviser | glm-5 | true | 23/23 (100.00%) | 7.876s |
| reflector | glm-4.5-air | true | 23/23 (100.00%) | 6.347s |
| searcher | glm-4.5-air | true | 23/23 (100.00%) | 5.492s |
| enricher | glm-4.5-air | true | 23/23 (100.00%) | 6.488s |
| coder | glm-5 | true | 23/23 (100.00%) | 6.128s |
| installer | glm-4.7 | true | 23/23 (100.00%) | 3.903s |
| pentester | glm-4.7 | true | 22/23 (95.65%) | 5.350s |

**Total**: 279/281 (99.29%) successful tests
**Overall average latency**: 7.529s

## Detailed Results

### simple (glm-4.7-flashx)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.350s |  |
| Text Transform Uppercase | ✅ Pass | 3.344s |  |
| Count from 1 to 5 | ✅ Pass | 16.684s |  |
| Math Calculation | ✅ Pass | 21.074s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.029s |  |
| Basic Echo Function | ✅ Pass | 83.870s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.848s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.779s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 32.834s |  |
| Search Query Function | ✅ Pass | 1.665s |  |
| Ask Advice Function | ✅ Pass | 4.198s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 38.178s |  |
| Basic Context Memory Test | ✅ Pass | 4.826s |  |
| Function Argument Memory Test | ✅ Pass | 51.598s |  |
| Function Response Memory Test | ✅ Pass | 59.462s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 18.542s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 2.568s | expected text 'example\.com' not found |
| Penetration Testing Methodology | ✅ Pass | 11.113s |  |
| Vulnerability Assessment Tools | ✅ Pass | 53.941s |  |
| SQL Injection Attack Type | ✅ Pass | 7.778s |  |
| Penetration Testing Framework | ✅ Pass | 32.387s |  |
| Web Application Security Scanner | ✅ Pass | 16.517s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.715s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 20.579s

---

### simple_json (glm-4.7-flashx)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 3.598s |  |
| Project Information JSON | ✅ Pass | 4.589s |  |
| User Profile JSON | ✅ Pass | 2.576s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 2.393s |  |
| Vulnerability Report Memory Test | ✅ Pass | 22.377s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 7.107s

---

### primary_agent (glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.481s |  |
| Text Transform Uppercase | ✅ Pass | 2.408s |  |
| Count from 1 to 5 | ✅ Pass | 5.159s |  |
| Math Calculation | ✅ Pass | 1.981s |  |
| Basic Echo Function | ✅ Pass | 4.224s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.030s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.049s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.907s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.821s |  |
| Search Query Function | ✅ Pass | 2.699s |  |
| Ask Advice Function | ✅ Pass | 4.990s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.747s |  |
| Basic Context Memory Test | ✅ Pass | 1.676s |  |
| Function Argument Memory Test | ✅ Pass | 3.306s |  |
| Function Response Memory Test | ✅ Pass | 1.959s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.072s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.079s |  |
| Penetration Testing Methodology | ✅ Pass | 18.268s |  |
| Vulnerability Assessment Tools | ✅ Pass | 39.499s |  |
| SQL Injection Attack Type | ✅ Pass | 3.957s |  |
| Penetration Testing Framework | ✅ Pass | 22.103s |  |
| Web Application Security Scanner | ✅ Pass | 11.489s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.231s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.050s

---

### assistant (glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.425s |  |
| Text Transform Uppercase | ✅ Pass | 5.846s |  |
| Count from 1 to 5 | ✅ Pass | 3.686s |  |
| Math Calculation | ✅ Pass | 2.497s |  |
| Basic Echo Function | ✅ Pass | 3.883s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.891s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.354s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.570s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.925s |  |
| Search Query Function | ✅ Pass | 2.993s |  |
| Ask Advice Function | ✅ Pass | 4.531s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.684s |  |
| Basic Context Memory Test | ✅ Pass | 2.207s |  |
| Function Argument Memory Test | ✅ Pass | 2.674s |  |
| Function Response Memory Test | ✅ Pass | 2.899s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.218s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.481s |  |
| Penetration Testing Methodology | ✅ Pass | 19.841s |  |
| Vulnerability Assessment Tools | ✅ Pass | 30.548s |  |
| SQL Injection Attack Type | ✅ Pass | 7.352s |  |
| Penetration Testing Framework | ✅ Pass | 26.431s |  |
| Web Application Security Scanner | ✅ Pass | 13.787s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.800s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.197s

---

### generator (glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.151s |  |
| Text Transform Uppercase | ✅ Pass | 5.508s |  |
| Count from 1 to 5 | ✅ Pass | 3.995s |  |
| Math Calculation | ✅ Pass | 4.557s |  |
| Basic Echo Function | ✅ Pass | 6.516s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.617s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.486s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.614s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.064s |  |
| Search Query Function | ✅ Pass | 3.644s |  |
| Ask Advice Function | ✅ Pass | 4.685s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.579s |  |
| Basic Context Memory Test | ✅ Pass | 2.774s |  |
| Function Argument Memory Test | ✅ Pass | 3.489s |  |
| Function Response Memory Test | ✅ Pass | 5.024s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.929s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.633s |  |
| Penetration Testing Methodology | ✅ Pass | 14.089s |  |
| Vulnerability Assessment Tools | ✅ Pass | 23.320s |  |
| SQL Injection Attack Type | ✅ Pass | 5.590s |  |
| Penetration Testing Framework | ✅ Pass | 21.081s |  |
| Web Application Security Scanner | ✅ Pass | 16.597s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.296s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.794s

---

### refiner (glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.857s |  |
| Text Transform Uppercase | ✅ Pass | 3.328s |  |
| Count from 1 to 5 | ✅ Pass | 4.175s |  |
| Math Calculation | ✅ Pass | 1.979s |  |
| Basic Echo Function | ✅ Pass | 3.519s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.409s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.773s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.607s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.429s |  |
| Search Query Function | ✅ Pass | 2.801s |  |
| Ask Advice Function | ✅ Pass | 3.807s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.450s |  |
| Basic Context Memory Test | ✅ Pass | 2.899s |  |
| Function Argument Memory Test | ✅ Pass | 2.724s |  |
| Function Response Memory Test | ✅ Pass | 7.180s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 10.559s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.838s |  |
| Penetration Testing Methodology | ✅ Pass | 17.640s |  |
| Vulnerability Assessment Tools | ✅ Pass | 35.316s |  |
| SQL Injection Attack Type | ✅ Pass | 4.652s |  |
| Penetration Testing Framework | ✅ Pass | 19.140s |  |
| Web Application Security Scanner | ✅ Pass | 15.485s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.825s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.235s

---

### adviser (glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.839s |  |
| Text Transform Uppercase | ✅ Pass | 5.472s |  |
| Count from 1 to 5 | ✅ Pass | 4.924s |  |
| Math Calculation | ✅ Pass | 3.169s |  |
| Basic Echo Function | ✅ Pass | 3.077s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.900s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.411s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.469s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.821s |  |
| Search Query Function | ✅ Pass | 3.395s |  |
| Ask Advice Function | ✅ Pass | 6.539s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 6.834s |  |
| Basic Context Memory Test | ✅ Pass | 1.888s |  |
| Function Argument Memory Test | ✅ Pass | 2.962s |  |
| Function Response Memory Test | ✅ Pass | 4.197s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.934s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.397s |  |
| Penetration Testing Methodology | ✅ Pass | 18.101s |  |
| Vulnerability Assessment Tools | ✅ Pass | 46.457s |  |
| SQL Injection Attack Type | ✅ Pass | 9.365s |  |
| Penetration Testing Framework | ✅ Pass | 17.170s |  |
| Web Application Security Scanner | ✅ Pass | 16.017s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.804s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.876s

---

### reflector (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.143s |  |
| Text Transform Uppercase | ✅ Pass | 1.870s |  |
| Count from 1 to 5 | ✅ Pass | 3.652s |  |
| Math Calculation | ✅ Pass | 1.387s |  |
| Basic Echo Function | ✅ Pass | 2.077s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.175s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.212s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.591s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.164s |  |
| Search Query Function | ✅ Pass | 2.576s |  |
| Ask Advice Function | ✅ Pass | 2.395s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.058s |  |
| Basic Context Memory Test | ✅ Pass | 2.424s |  |
| Function Argument Memory Test | ✅ Pass | 1.993s |  |
| Function Response Memory Test | ✅ Pass | 2.025s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.138s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.171s |  |
| Penetration Testing Methodology | ✅ Pass | 24.940s |  |
| Vulnerability Assessment Tools | ✅ Pass | 35.170s |  |
| SQL Injection Attack Type | ✅ Pass | 4.671s |  |
| Penetration Testing Framework | ✅ Pass | 22.360s |  |
| Web Application Security Scanner | ✅ Pass | 9.550s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.231s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.347s

---

### searcher (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.419s |  |
| Text Transform Uppercase | ✅ Pass | 1.759s |  |
| Count from 1 to 5 | ✅ Pass | 3.229s |  |
| Math Calculation | ✅ Pass | 1.062s |  |
| Basic Echo Function | ✅ Pass | 1.886s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.286s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.933s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.588s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.103s |  |
| Search Query Function | ✅ Pass | 2.430s |  |
| Ask Advice Function | ✅ Pass | 2.791s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.399s |  |
| Basic Context Memory Test | ✅ Pass | 1.652s |  |
| Function Argument Memory Test | ✅ Pass | 1.487s |  |
| Function Response Memory Test | ✅ Pass | 2.492s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.473s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.894s |  |
| Penetration Testing Methodology | ✅ Pass | 11.508s |  |
| Vulnerability Assessment Tools | ✅ Pass | 26.143s |  |
| SQL Injection Attack Type | ✅ Pass | 5.851s |  |
| Penetration Testing Framework | ✅ Pass | 16.044s |  |
| Web Application Security Scanner | ✅ Pass | 20.166s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.711s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.492s

---

### enricher (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.025s |  |
| Text Transform Uppercase | ✅ Pass | 2.244s |  |
| Count from 1 to 5 | ✅ Pass | 2.697s |  |
| Math Calculation | ✅ Pass | 1.304s |  |
| Basic Echo Function | ✅ Pass | 1.865s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.939s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.881s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.618s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.218s |  |
| Search Query Function | ✅ Pass | 2.579s |  |
| Ask Advice Function | ✅ Pass | 2.049s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.947s |  |
| Basic Context Memory Test | ✅ Pass | 1.765s |  |
| Function Argument Memory Test | ✅ Pass | 1.733s |  |
| Function Response Memory Test | ✅ Pass | 1.646s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.578s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.063s |  |
| Penetration Testing Methodology | ✅ Pass | 15.977s |  |
| Vulnerability Assessment Tools | ✅ Pass | 46.410s |  |
| SQL Injection Attack Type | ✅ Pass | 10.036s |  |
| Penetration Testing Framework | ✅ Pass | 20.568s |  |
| Web Application Security Scanner | ✅ Pass | 10.759s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.314s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.488s

---

### coder (glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.028s |  |
| Text Transform Uppercase | ✅ Pass | 2.695s |  |
| Count from 1 to 5 | ✅ Pass | 4.099s |  |
| Math Calculation | ✅ Pass | 2.054s |  |
| Basic Echo Function | ✅ Pass | 4.083s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.808s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.021s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.917s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.616s |  |
| Search Query Function | ✅ Pass | 4.091s |  |
| Ask Advice Function | ✅ Pass | 4.418s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.970s |  |
| Basic Context Memory Test | ✅ Pass | 2.142s |  |
| Function Argument Memory Test | ✅ Pass | 2.669s |  |
| Function Response Memory Test | ✅ Pass | 4.727s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.417s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.213s |  |
| Penetration Testing Methodology | ✅ Pass | 13.789s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.248s |  |
| SQL Injection Attack Type | ✅ Pass | 7.931s |  |
| Penetration Testing Framework | ✅ Pass | 18.277s |  |
| Web Application Security Scanner | ✅ Pass | 15.769s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.949s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.128s

---

### installer (glm-4.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.639s |  |
| Text Transform Uppercase | ✅ Pass | 2.362s |  |
| Count from 1 to 5 | ✅ Pass | 2.142s |  |
| Math Calculation | ✅ Pass | 1.261s |  |
| Basic Echo Function | ✅ Pass | 2.065s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.972s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.666s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 24.519s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.922s |  |
| Search Query Function | ✅ Pass | 1.170s |  |
| Ask Advice Function | ✅ Pass | 1.321s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.223s |  |
| Basic Context Memory Test | ✅ Pass | 2.865s |  |
| Function Argument Memory Test | ✅ Pass | 6.698s |  |
| Function Response Memory Test | ✅ Pass | 1.635s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.691s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.232s |  |
| Penetration Testing Methodology | ✅ Pass | 4.972s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.719s |  |
| SQL Injection Attack Type | ✅ Pass | 3.134s |  |
| Penetration Testing Framework | ✅ Pass | 7.910s |  |
| Web Application Security Scanner | ✅ Pass | 8.168s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.464s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.903s

---

### pentester (glm-4.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.664s |  |
| Text Transform Uppercase | ✅ Pass | 5.869s |  |
| Count from 1 to 5 | ✅ Pass | 2.157s |  |
| Math Calculation | ✅ Pass | 2.004s |  |
| Basic Echo Function | ✅ Pass | 1.038s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.588s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.166s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 20.024s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.388s |  |
| Search Query Function | ✅ Pass | 1.291s |  |
| Ask Advice Function | ✅ Pass | 1.762s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.613s |  |
| Basic Context Memory Test | ✅ Pass | 2.043s |  |
| Function Argument Memory Test | ✅ Pass | 1.674s |  |
| Function Response Memory Test | ✅ Pass | 2.169s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.299s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.199s |  |
| Penetration Testing Methodology | ✅ Pass | 5.637s |  |
| Vulnerability Assessment Tools | ❌ Fail | 36.737s | expected text 'network' not found |
| SQL Injection Attack Type | ✅ Pass | 4.446s |  |
| Penetration Testing Framework | ✅ Pass | 6.258s |  |
| Web Application Security Scanner | ✅ Pass | 15.067s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.940s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.350s

---

