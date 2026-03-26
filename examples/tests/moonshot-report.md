# LLM Agent Testing Report

Generated: Thu, 29 Jan 2026 17:23:17 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | kimi-k2-0905-preview | true | 23/23 (100.00%) | 2.527s |
| simple_json | kimi-k2-0905-preview | false | 5/5 (100.00%) | 3.434s |
| primary_agent | kimi-k2.5 | true | 23/23 (100.00%) | 4.712s |
| assistant | kimi-k2.5 | true | 23/23 (100.00%) | 4.800s |
| generator | kimi-k2.5 | true | 23/23 (100.00%) | 4.819s |
| refiner | kimi-k2.5 | true | 23/23 (100.00%) | 5.105s |
| adviser | kimi-k2.5 | true | 23/23 (100.00%) | 4.209s |
| reflector | kimi-k2-0905-preview | true | 23/23 (100.00%) | 2.616s |
| searcher | kimi-k2-0905-preview | true | 23/23 (100.00%) | 2.564s |
| enricher | kimi-k2-0905-preview | true | 23/23 (100.00%) | 2.497s |
| coder | kimi-k2.5 | true | 23/23 (100.00%) | 5.042s |
| installer | kimi-k2-turbo-preview | true | 23/23 (100.00%) | 1.057s |
| pentester | kimi-k2-turbo-preview | true | 23/23 (100.00%) | 1.050s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 3.417s

## Detailed Results

### simple (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.969s |  |
| Text Transform Uppercase | ✅ Pass | 1.409s |  |
| Count from 1 to 5 | ✅ Pass | 2.185s |  |
| Math Calculation | ✅ Pass | 1.264s |  |
| Basic Echo Function | ✅ Pass | 3.142s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.245s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.640s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.451s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.030s |  |
| Search Query Function | ✅ Pass | 3.010s |  |
| Ask Advice Function | ✅ Pass | 4.312s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.796s |  |
| Basic Context Memory Test | ✅ Pass | 2.255s |  |
| Function Argument Memory Test | ✅ Pass | 1.492s |  |
| Function Response Memory Test | ✅ Pass | 1.159s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.136s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.327s |  |
| Penetration Testing Methodology | ✅ Pass | 1.652s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.114s |  |
| SQL Injection Attack Type | ✅ Pass | 2.276s |  |
| Penetration Testing Framework | ✅ Pass | 1.798s |  |
| Web Application Security Scanner | ✅ Pass | 1.280s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.168s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.527s

---

### simple_json (kimi-k2-0905-preview)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 2.695s |  |
| Project Information JSON | ✅ Pass | 2.609s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 2.682s |  |
| User Profile JSON | ✅ Pass | 3.029s |  |
| Vulnerability Report Memory Test | ✅ Pass | 6.151s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 3.434s

---

### primary_agent (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.156s |  |
| Text Transform Uppercase | ✅ Pass | 2.189s |  |
| Count from 1 to 5 | ✅ Pass | 4.424s |  |
| Math Calculation | ✅ Pass | 1.450s |  |
| Basic Echo Function | ✅ Pass | 1.899s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.228s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.028s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.610s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.485s |  |
| Search Query Function | ✅ Pass | 1.712s |  |
| Ask Advice Function | ✅ Pass | 2.803s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.520s |  |
| Basic Context Memory Test | ✅ Pass | 2.999s |  |
| Function Argument Memory Test | ✅ Pass | 2.119s |  |
| Function Response Memory Test | ✅ Pass | 2.689s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.852s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.575s |  |
| Penetration Testing Methodology | ✅ Pass | 15.432s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.437s |  |
| SQL Injection Attack Type | ✅ Pass | 6.549s |  |
| Penetration Testing Framework | ✅ Pass | 9.912s |  |
| Web Application Security Scanner | ✅ Pass | 13.061s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.227s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.712s

---

### assistant (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.533s |  |
| Text Transform Uppercase | ✅ Pass | 2.086s |  |
| Count from 1 to 5 | ✅ Pass | 4.770s |  |
| Math Calculation | ✅ Pass | 2.369s |  |
| Basic Echo Function | ✅ Pass | 2.024s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.364s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.988s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.295s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.751s |  |
| Search Query Function | ✅ Pass | 1.407s |  |
| Ask Advice Function | ✅ Pass | 2.156s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.947s |  |
| Basic Context Memory Test | ✅ Pass | 3.082s |  |
| Function Argument Memory Test | ✅ Pass | 2.787s |  |
| Function Response Memory Test | ✅ Pass | 2.670s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.721s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.700s |  |
| Penetration Testing Methodology | ✅ Pass | 15.716s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.163s |  |
| SQL Injection Attack Type | ✅ Pass | 4.081s |  |
| Penetration Testing Framework | ✅ Pass | 10.991s |  |
| Web Application Security Scanner | ✅ Pass | 14.581s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.207s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.800s

---

### generator (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.573s |  |
| Text Transform Uppercase | ✅ Pass | 3.412s |  |
| Count from 1 to 5 | ✅ Pass | 4.818s |  |
| Math Calculation | ✅ Pass | 1.616s |  |
| Basic Echo Function | ✅ Pass | 1.914s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.830s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.596s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.960s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.316s |  |
| Search Query Function | ✅ Pass | 1.623s |  |
| Ask Advice Function | ✅ Pass | 1.644s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.135s |  |
| Basic Context Memory Test | ✅ Pass | 3.843s |  |
| Function Argument Memory Test | ✅ Pass | 1.947s |  |
| Function Response Memory Test | ✅ Pass | 4.476s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.291s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.153s |  |
| Penetration Testing Methodology | ✅ Pass | 11.999s |  |
| SQL Injection Attack Type | ✅ Pass | 4.011s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.057s |  |
| Penetration Testing Framework | ✅ Pass | 12.720s |  |
| Web Application Security Scanner | ✅ Pass | 11.890s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.008s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.819s

---

### refiner (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.575s |  |
| Text Transform Uppercase | ✅ Pass | 1.467s |  |
| Count from 1 to 5 | ✅ Pass | 2.710s |  |
| Math Calculation | ✅ Pass | 2.063s |  |
| Basic Echo Function | ✅ Pass | 2.177s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.719s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.366s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.311s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.040s |  |
| Search Query Function | ✅ Pass | 1.610s |  |
| Ask Advice Function | ✅ Pass | 2.109s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.672s |  |
| Basic Context Memory Test | ✅ Pass | 3.430s |  |
| Function Argument Memory Test | ✅ Pass | 2.247s |  |
| Function Response Memory Test | ✅ Pass | 3.155s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.300s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.164s |  |
| Penetration Testing Methodology | ✅ Pass | 11.814s |  |
| SQL Injection Attack Type | ✅ Pass | 4.782s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.726s |  |
| Penetration Testing Framework | ✅ Pass | 17.103s |  |
| Web Application Security Scanner | ✅ Pass | 14.709s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.152s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.105s

---

### adviser (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.087s |  |
| Text Transform Uppercase | ✅ Pass | 2.282s |  |
| Count from 1 to 5 | ✅ Pass | 1.617s |  |
| Math Calculation | ✅ Pass | 2.105s |  |
| Basic Echo Function | ✅ Pass | 2.211s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.083s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.229s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.895s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.561s |  |
| Search Query Function | ✅ Pass | 1.690s |  |
| Ask Advice Function | ✅ Pass | 2.217s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.371s |  |
| Basic Context Memory Test | ✅ Pass | 2.929s |  |
| Function Argument Memory Test | ✅ Pass | 1.864s |  |
| Function Response Memory Test | ✅ Pass | 3.357s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.003s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.202s |  |
| Penetration Testing Methodology | ✅ Pass | 13.754s |  |
| SQL Injection Attack Type | ✅ Pass | 2.961s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.462s |  |
| Penetration Testing Framework | ✅ Pass | 6.220s |  |
| Web Application Security Scanner | ✅ Pass | 12.801s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.884s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.209s

---

### reflector (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.968s |  |
| Text Transform Uppercase | ✅ Pass | 1.353s |  |
| Count from 1 to 5 | ✅ Pass | 2.131s |  |
| Math Calculation | ✅ Pass | 1.275s |  |
| Basic Echo Function | ✅ Pass | 2.887s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.255s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.752s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.544s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.087s |  |
| Search Query Function | ✅ Pass | 3.010s |  |
| Ask Advice Function | ✅ Pass | 4.280s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.324s |  |
| Basic Context Memory Test | ✅ Pass | 2.626s |  |
| Function Argument Memory Test | ✅ Pass | 1.436s |  |
| Function Response Memory Test | ✅ Pass | 1.267s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.401s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.327s |  |
| Penetration Testing Methodology | ✅ Pass | 2.447s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.007s |  |
| SQL Injection Attack Type | ✅ Pass | 2.523s |  |
| Penetration Testing Framework | ✅ Pass | 1.611s |  |
| Web Application Security Scanner | ✅ Pass | 2.019s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.628s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.616s

---

### searcher (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.976s |  |
| Text Transform Uppercase | ✅ Pass | 1.436s |  |
| Count from 1 to 5 | ✅ Pass | 2.454s |  |
| Math Calculation | ✅ Pass | 1.724s |  |
| Basic Echo Function | ✅ Pass | 2.867s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.358s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.621s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.471s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.997s |  |
| Search Query Function | ✅ Pass | 2.949s |  |
| Ask Advice Function | ✅ Pass | 4.267s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.800s |  |
| Basic Context Memory Test | ✅ Pass | 2.723s |  |
| Function Argument Memory Test | ✅ Pass | 1.476s |  |
| Function Response Memory Test | ✅ Pass | 1.480s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.167s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.430s |  |
| Penetration Testing Methodology | ✅ Pass | 2.091s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.125s |  |
| SQL Injection Attack Type | ✅ Pass | 2.275s |  |
| Penetration Testing Framework | ✅ Pass | 1.584s |  |
| Web Application Security Scanner | ✅ Pass | 1.475s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.217s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.564s

---

### enricher (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.380s |  |
| Text Transform Uppercase | ✅ Pass | 1.480s |  |
| Count from 1 to 5 | ✅ Pass | 2.090s |  |
| Math Calculation | ✅ Pass | 1.219s |  |
| Basic Echo Function | ✅ Pass | 2.964s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.157s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.274s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.458s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.441s |  |
| Search Query Function | ✅ Pass | 2.873s |  |
| Ask Advice Function | ✅ Pass | 4.240s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.152s |  |
| Basic Context Memory Test | ✅ Pass | 2.943s |  |
| Function Argument Memory Test | ✅ Pass | 1.419s |  |
| Function Response Memory Test | ✅ Pass | 1.301s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.490s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.617s |  |
| Penetration Testing Methodology | ✅ Pass | 2.430s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.518s |  |
| SQL Injection Attack Type | ✅ Pass | 1.380s |  |
| Penetration Testing Framework | ✅ Pass | 1.676s |  |
| Web Application Security Scanner | ✅ Pass | 1.834s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.091s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.497s

---

### coder (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.846s |  |
| Text Transform Uppercase | ✅ Pass | 1.852s |  |
| Count from 1 to 5 | ✅ Pass | 5.834s |  |
| Math Calculation | ✅ Pass | 2.142s |  |
| Basic Echo Function | ✅ Pass | 2.158s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.971s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.713s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.089s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.249s |  |
| Search Query Function | ✅ Pass | 2.192s |  |
| Ask Advice Function | ✅ Pass | 3.064s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.303s |  |
| Basic Context Memory Test | ✅ Pass | 3.178s |  |
| Function Argument Memory Test | ✅ Pass | 1.985s |  |
| Function Response Memory Test | ✅ Pass | 3.064s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.572s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.607s |  |
| Penetration Testing Methodology | ✅ Pass | 17.249s |  |
| SQL Injection Attack Type | ✅ Pass | 2.996s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.738s |  |
| Penetration Testing Framework | ✅ Pass | 14.365s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.291s |  |
| Web Application Security Scanner | ✅ Pass | 14.502s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.042s

---

### installer (kimi-k2-turbo-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.799s |  |
| Text Transform Uppercase | ✅ Pass | 0.831s |  |
| Count from 1 to 5 | ✅ Pass | 0.949s |  |
| Math Calculation | ✅ Pass | 0.844s |  |
| Basic Echo Function | ✅ Pass | 1.011s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.789s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.455s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.465s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.008s |  |
| Search Query Function | ✅ Pass | 1.117s |  |
| Ask Advice Function | ✅ Pass | 1.262s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.063s |  |
| Basic Context Memory Test | ✅ Pass | 0.915s |  |
| Function Argument Memory Test | ✅ Pass | 0.787s |  |
| Function Response Memory Test | ✅ Pass | 0.798s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.997s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.827s |  |
| Penetration Testing Methodology | ✅ Pass | 0.989s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.864s |  |
| SQL Injection Attack Type | ✅ Pass | 0.902s |  |
| Penetration Testing Framework | ✅ Pass | 1.020s |  |
| Web Application Security Scanner | ✅ Pass | 1.263s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.345s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.057s

---

### pentester (kimi-k2-turbo-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.856s |  |
| Text Transform Uppercase | ✅ Pass | 0.816s |  |
| Count from 1 to 5 | ✅ Pass | 0.893s |  |
| Math Calculation | ✅ Pass | 0.829s |  |
| Basic Echo Function | ✅ Pass | 0.923s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.820s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.375s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.587s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.088s |  |
| Search Query Function | ✅ Pass | 0.976s |  |
| Ask Advice Function | ✅ Pass | 1.218s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.231s |  |
| Basic Context Memory Test | ✅ Pass | 0.975s |  |
| Function Argument Memory Test | ✅ Pass | 0.927s |  |
| Function Response Memory Test | ✅ Pass | 0.830s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.959s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.795s |  |
| Penetration Testing Methodology | ✅ Pass | 1.035s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.987s |  |
| SQL Injection Attack Type | ✅ Pass | 0.972s |  |
| Penetration Testing Framework | ✅ Pass | 1.222s |  |
| Web Application Security Scanner | ✅ Pass | 0.834s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.989s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.050s

---

