# LLM Agent Testing Report

Generated: Wed, 27 May 2026 23:05:00 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | qwen3.5-flash | true | 23/23 (100.00%) | 3.014s |
| simple_json | qwen3.5-flash | true | 5/5 (100.00%) | 7.088s |
| primary_agent | qwen3.6-plus | true | 23/23 (100.00%) | 5.366s |
| assistant | qwen3.6-plus | true | 23/23 (100.00%) | 5.758s |
| generator | qwen3.7-max | true | 23/23 (100.00%) | 3.473s |
| refiner | qwen3.7-max | true | 23/23 (100.00%) | 3.352s |
| adviser | qwen3.7-max | true | 22/23 (95.65%) | 2.941s |
| reflector | qwen3.5-flash | true | 23/23 (100.00%) | 3.377s |
| searcher | qwen3.5-flash | true | 23/23 (100.00%) | 4.025s |
| enricher | qwen3.5-flash | true | 23/23 (100.00%) | 2.857s |
| coder | qwen3-coder-plus | true | 23/23 (100.00%) | 1.556s |
| installer | qwen3-coder-flash | true | 20/23 (86.96%) | 1.060s |
| pentester | qwen3.6-plus | true | 23/23 (100.00%) | 5.504s |

**Total**: 277/281 (98.58%) successful tests
**Overall average latency**: 3.587s

## Detailed Results

### simple (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.746s |  |
| Text Transform Uppercase | ✅ Pass | 2.043s |  |
| Count from 1 to 5 | ✅ Pass | 3.729s |  |
| Math Calculation | ✅ Pass | 1.686s |  |
| Basic Echo Function | ✅ Pass | 1.282s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.606s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.200s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.928s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.163s |  |
| Search Query Function | ✅ Pass | 1.264s |  |
| Ask Advice Function | ✅ Pass | 1.232s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.884s |  |
| Basic Context Memory Test | ✅ Pass | 2.803s |  |
| Function Argument Memory Test | ✅ Pass | 0.750s |  |
| Function Response Memory Test | ✅ Pass | 0.951s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.953s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.255s |  |
| Penetration Testing Methodology | ✅ Pass | 7.005s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.557s |  |
| SQL Injection Attack Type | ✅ Pass | 2.703s |  |
| Penetration Testing Framework | ✅ Pass | 5.521s |  |
| Web Application Security Scanner | ✅ Pass | 4.520s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.533s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.014s

---

### simple_json (qwen3.5-flash)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 4.643s |  |
| User Profile JSON | ✅ Pass | 5.759s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 6.514s |  |
| Project Information JSON | ✅ Pass | 7.046s |  |
| Vulnerability Report Memory Test | ✅ Pass | 11.475s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 7.088s

---

### primary_agent (qwen3.6-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.300s |  |
| Text Transform Uppercase | ✅ Pass | 3.870s |  |
| Math Calculation | ✅ Pass | 2.999s |  |
| Count from 1 to 5 | ✅ Pass | 12.409s |  |
| Basic Echo Function | ✅ Pass | 3.121s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.737s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.527s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.412s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.506s |  |
| Search Query Function | ✅ Pass | 2.580s |  |
| Ask Advice Function | ✅ Pass | 3.013s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.157s |  |
| Basic Context Memory Test | ✅ Pass | 4.461s |  |
| Function Argument Memory Test | ✅ Pass | 4.159s |  |
| Function Response Memory Test | ✅ Pass | 6.988s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.572s |  |
| Penetration Testing Methodology | ✅ Pass | 11.000s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.996s |  |
| SQL Injection Attack Type | ✅ Pass | 5.396s |  |
| Penetration Testing Framework | ✅ Pass | 11.164s |  |
| Web Application Security Scanner | ✅ Pass | 4.543s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.287s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.366s

---

### assistant (qwen3.6-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.339s |  |
| Text Transform Uppercase | ✅ Pass | 5.104s |  |
| Count from 1 to 5 | ✅ Pass | 5.229s |  |
| Math Calculation | ✅ Pass | 4.509s |  |
| Basic Echo Function | ✅ Pass | 3.068s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.267s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.656s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.860s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.200s |  |
| Search Query Function | ✅ Pass | 1.780s |  |
| Ask Advice Function | ✅ Pass | 2.841s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.020s |  |
| Basic Context Memory Test | ✅ Pass | 5.116s |  |
| Function Argument Memory Test | ✅ Pass | 3.908s |  |
| Function Response Memory Test | ✅ Pass | 4.534s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.970s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 7.041s |  |
| Penetration Testing Methodology | ✅ Pass | 11.914s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.492s |  |
| SQL Injection Attack Type | ✅ Pass | 5.510s |  |
| Penetration Testing Framework | ✅ Pass | 9.591s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.796s |  |
| Web Application Security Scanner | ✅ Pass | 10.669s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.758s

---

### generator (qwen3.7-max)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.289s |  |
| Text Transform Uppercase | ✅ Pass | 2.103s |  |
| Count from 1 to 5 | ✅ Pass | 4.641s |  |
| Math Calculation | ✅ Pass | 1.319s |  |
| Basic Echo Function | ✅ Pass | 1.852s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.998s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.591s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.928s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.388s |  |
| Search Query Function | ✅ Pass | 2.131s |  |
| Ask Advice Function | ✅ Pass | 1.544s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.977s |  |
| Basic Context Memory Test | ✅ Pass | 4.750s |  |
| Function Argument Memory Test | ✅ Pass | 3.913s |  |
| Function Response Memory Test | ✅ Pass | 4.594s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.141s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.417s |  |
| Penetration Testing Methodology | ✅ Pass | 3.566s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.497s |  |
| SQL Injection Attack Type | ✅ Pass | 4.163s |  |
| Penetration Testing Framework | ✅ Pass | 2.677s |  |
| Web Application Security Scanner | ✅ Pass | 4.448s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.928s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.473s

---

### refiner (qwen3.7-max)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.088s |  |
| Text Transform Uppercase | ✅ Pass | 3.518s |  |
| Count from 1 to 5 | ✅ Pass | 4.139s |  |
| Math Calculation | ✅ Pass | 1.941s |  |
| Basic Echo Function | ✅ Pass | 1.601s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.624s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.881s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.028s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.901s |  |
| Search Query Function | ✅ Pass | 2.251s |  |
| Ask Advice Function | ✅ Pass | 1.898s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.260s |  |
| Basic Context Memory Test | ✅ Pass | 2.202s |  |
| Function Argument Memory Test | ✅ Pass | 2.319s |  |
| Function Response Memory Test | ✅ Pass | 4.750s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.590s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.847s |  |
| Penetration Testing Methodology | ✅ Pass | 4.130s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.315s |  |
| SQL Injection Attack Type | ✅ Pass | 4.225s |  |
| Penetration Testing Framework | ✅ Pass | 2.234s |  |
| Web Application Security Scanner | ✅ Pass | 2.434s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.917s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.352s

---

### adviser (qwen3.7-max)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.094s |  |
| Text Transform Uppercase | ✅ Pass | 3.196s |  |
| Count from 1 to 5 | ✅ Pass | 3.210s |  |
| Math Calculation | ✅ Pass | 2.542s |  |
| Basic Echo Function | ✅ Pass | 1.431s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.476s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.746s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.526s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.530s |  |
| Search Query Function | ✅ Pass | 1.782s |  |
| Ask Advice Function | ✅ Pass | 1.862s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.915s |  |
| Basic Context Memory Test | ✅ Pass | 3.090s |  |
| Function Argument Memory Test | ❌ Fail | 3.063s | expected text 'Go programming language' not found |
| Function Response Memory Test | ✅ Pass | 4.037s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.505s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.618s |  |
| Penetration Testing Methodology | ✅ Pass | 2.493s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.921s |  |
| SQL Injection Attack Type | ✅ Pass | 3.418s |  |
| Penetration Testing Framework | ✅ Pass | 4.994s |  |
| Web Application Security Scanner | ✅ Pass | 1.890s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.285s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.941s

---

### reflector (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.446s |  |
| Text Transform Uppercase | ✅ Pass | 2.472s |  |
| Count from 1 to 5 | ✅ Pass | 3.109s |  |
| Math Calculation | ✅ Pass | 2.244s |  |
| Basic Echo Function | ✅ Pass | 1.191s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.377s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.817s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.952s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.376s |  |
| Search Query Function | ✅ Pass | 1.034s |  |
| Ask Advice Function | ✅ Pass | 1.177s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.399s |  |
| Basic Context Memory Test | ✅ Pass | 2.150s |  |
| Function Argument Memory Test | ✅ Pass | 0.884s |  |
| Function Response Memory Test | ✅ Pass | 0.900s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.927s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.201s |  |
| Penetration Testing Methodology | ✅ Pass | 8.294s |  |
| SQL Injection Attack Type | ✅ Pass | 2.431s |  |
| Vulnerability Assessment Tools | ✅ Pass | 24.680s |  |
| Penetration Testing Framework | ✅ Pass | 4.850s |  |
| Web Application Security Scanner | ✅ Pass | 4.195s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.546s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.377s

---

### searcher (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.959s |  |
| Text Transform Uppercase | ✅ Pass | 1.535s |  |
| Count from 1 to 5 | ✅ Pass | 3.085s |  |
| Math Calculation | ✅ Pass | 1.883s |  |
| Basic Echo Function | ✅ Pass | 1.491s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.364s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.694s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.338s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.277s |  |
| Search Query Function | ✅ Pass | 1.249s |  |
| Ask Advice Function | ✅ Pass | 1.653s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.925s |  |
| Basic Context Memory Test | ✅ Pass | 3.426s |  |
| Function Argument Memory Test | ✅ Pass | 1.552s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.090s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.792s |  |
| Function Response Memory Test | ✅ Pass | 22.405s |  |
| Penetration Testing Methodology | ✅ Pass | 6.872s |  |
| SQL Injection Attack Type | ✅ Pass | 2.512s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.206s |  |
| Penetration Testing Framework | ✅ Pass | 4.480s |  |
| Web Application Security Scanner | ✅ Pass | 4.976s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.793s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.025s

---

### enricher (qwen3.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.626s |  |
| Text Transform Uppercase | ✅ Pass | 2.623s |  |
| Count from 1 to 5 | ✅ Pass | 2.794s |  |
| Math Calculation | ✅ Pass | 1.657s |  |
| Basic Echo Function | ✅ Pass | 1.406s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.086s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.390s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.279s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.314s |  |
| Search Query Function | ✅ Pass | 1.160s |  |
| Ask Advice Function | ✅ Pass | 1.645s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.948s |  |
| Basic Context Memory Test | ✅ Pass | 3.003s |  |
| Function Argument Memory Test | ✅ Pass | 2.254s |  |
| Function Response Memory Test | ✅ Pass | 4.521s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.225s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.494s |  |
| Penetration Testing Methodology | ✅ Pass | 5.008s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.717s |  |
| SQL Injection Attack Type | ✅ Pass | 4.587s |  |
| Penetration Testing Framework | ✅ Pass | 4.199s |  |
| Web Application Security Scanner | ✅ Pass | 4.499s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.261s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.857s

---

### coder (qwen3-coder-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.873s |  |
| Text Transform Uppercase | ✅ Pass | 0.824s |  |
| Count from 1 to 5 | ✅ Pass | 0.953s |  |
| Math Calculation | ✅ Pass | 0.842s |  |
| Basic Echo Function | ✅ Pass | 1.514s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.386s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.346s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.009s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.687s |  |
| Search Query Function | ✅ Pass | 0.994s |  |
| Ask Advice Function | ✅ Pass | 1.286s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.041s |  |
| Basic Context Memory Test | ✅ Pass | 0.905s |  |
| Function Argument Memory Test | ✅ Pass | 0.845s |  |
| Function Response Memory Test | ✅ Pass | 0.806s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.240s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.862s |  |
| Penetration Testing Methodology | ✅ Pass | 4.261s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.398s |  |
| SQL Injection Attack Type | ✅ Pass | 0.869s |  |
| Penetration Testing Framework | ✅ Pass | 3.691s |  |
| Web Application Security Scanner | ✅ Pass | 2.796s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.349s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.556s

---

### installer (qwen3-coder-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.007s |  |
| Text Transform Uppercase | ✅ Pass | 0.566s |  |
| Count from 1 to 5 | ✅ Pass | 1.100s |  |
| Math Calculation | ✅ Pass | 1.011s |  |
| Basic Echo Function | ❌ Fail | 0.727s | no tool calls found, expected at least 1 |
| Streaming Simple Math Streaming | ✅ Pass | 0.542s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.560s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.774s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.830s |  |
| Search Query Function | ❌ Fail | 1.222s | no tool calls found, expected at least 1 |
| Ask Advice Function | ✅ Pass | 0.877s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 1.146s | no tool calls found, expected at least 1 |
| Basic Context Memory Test | ✅ Pass | 0.973s |  |
| Function Argument Memory Test | ✅ Pass | 0.613s |  |
| Function Response Memory Test | ✅ Pass | 0.583s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.826s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.625s |  |
| Penetration Testing Methodology | ✅ Pass | 1.940s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.151s |  |
| SQL Injection Attack Type | ✅ Pass | 0.588s |  |
| Penetration Testing Framework | ✅ Pass | 1.175s |  |
| Web Application Security Scanner | ✅ Pass | 1.550s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.973s |  |

**Summary**: 20/23 (86.96%) successful tests

**Average latency**: 1.060s

---

### pentester (qwen3.6-plus)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.527s |  |
| Text Transform Uppercase | ✅ Pass | 4.599s |  |
| Count from 1 to 5 | ✅ Pass | 5.743s |  |
| Math Calculation | ✅ Pass | 3.509s |  |
| Basic Echo Function | ✅ Pass | 4.851s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.707s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.781s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.300s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.456s |  |
| Search Query Function | ✅ Pass | 3.989s |  |
| Ask Advice Function | ✅ Pass | 3.255s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.430s |  |
| Basic Context Memory Test | ✅ Pass | 4.780s |  |
| Function Argument Memory Test | ✅ Pass | 5.338s |  |
| Function Response Memory Test | ✅ Pass | 2.911s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.094s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.163s |  |
| Penetration Testing Methodology | ✅ Pass | 12.475s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.406s |  |
| SQL Injection Attack Type | ✅ Pass | 6.216s |  |
| Penetration Testing Framework | ✅ Pass | 11.490s |  |
| Web Application Security Scanner | ✅ Pass | 8.229s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.336s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.504s

---

