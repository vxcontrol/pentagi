# LLM Agent Testing Report

Generated: Tue, 30 Sep 2025 18:43:15 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-3-5-haiku-20241022 | false | 23/23 (100.00%) | 2.104s |
| simple_json | claude-3-5-haiku-20241022 | false | 5/5 (100.00%) | 1.244s |
| primary_agent | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.177s |
| assistant | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.382s |
| generator | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.167s |
| refiner | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.227s |
| adviser | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.096s |
| reflector | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 2.908s |
| searcher | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.085s |
| enricher | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 2.936s |
| coder | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 3.996s |
| installer | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 4.318s |
| pentester | claude-sonnet-4-5-20250929 | true | 23/23 (100.00%) | 3.936s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 3.733s

## Detailed Results

### simple (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.342s |  |
| Text Transform Uppercase | ✅ Pass | 1.067s |  |
| Count from 1 to 5 | ✅ Pass | 0.962s |  |
| Math Calculation | ✅ Pass | 0.884s |  |
| Basic Echo Function | ✅ Pass | 1.857s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.837s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.926s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.724s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.584s |  |
| Search Query Function | ✅ Pass | 1.782s |  |
| Ask Advice Function | ✅ Pass | 1.581s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.805s |  |
| Basic Context Memory Test | ✅ Pass | 1.176s |  |
| Function Argument Memory Test | ✅ Pass | 0.854s |  |
| Function Response Memory Test | ✅ Pass | 0.884s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.651s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.327s |  |
| Penetration Testing Methodology | ✅ Pass | 5.792s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.749s |  |
| SQL Injection Attack Type | ✅ Pass | 1.927s |  |
| Penetration Testing Framework | ✅ Pass | 4.995s |  |
| Web Application Security Scanner | ✅ Pass | 3.148s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.533s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.104s

---

### simple_json (claude-3-5-haiku-20241022)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.645s |  |
| Person Information JSON | ✅ Pass | 1.120s |  |
| Project Information JSON | ✅ Pass | 1.127s |  |
| User Profile JSON | ✅ Pass | 1.168s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.158s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.244s

---

### primary_agent (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.321s |  |
| Text Transform Uppercase | ✅ Pass | 3.083s |  |
| Count from 1 to 5 | ✅ Pass | 2.485s |  |
| Math Calculation | ✅ Pass | 2.267s |  |
| Basic Echo Function | ✅ Pass | 2.851s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.570s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.973s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.669s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.107s |  |
| Search Query Function | ✅ Pass | 2.967s |  |
| Ask Advice Function | ✅ Pass | 3.600s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.403s |  |
| Basic Context Memory Test | ✅ Pass | 3.041s |  |
| Function Argument Memory Test | ✅ Pass | 3.178s |  |
| Function Response Memory Test | ✅ Pass | 3.175s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.107s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.297s |  |
| Penetration Testing Methodology | ✅ Pass | 10.167s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.215s |  |
| SQL Injection Attack Type | ✅ Pass | 4.480s |  |
| Penetration Testing Framework | ✅ Pass | 8.540s |  |
| Web Application Security Scanner | ✅ Pass | 6.254s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.297s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.177s

---

### assistant (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.542s |  |
| Text Transform Uppercase | ✅ Pass | 2.868s |  |
| Count from 1 to 5 | ✅ Pass | 2.849s |  |
| Math Calculation | ✅ Pass | 2.388s |  |
| Basic Echo Function | ✅ Pass | 2.879s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.785s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.945s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.242s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.413s |  |
| Search Query Function | ✅ Pass | 2.607s |  |
| Ask Advice Function | ✅ Pass | 3.400s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.152s |  |
| Basic Context Memory Test | ✅ Pass | 3.961s |  |
| Function Argument Memory Test | ✅ Pass | 3.902s |  |
| Function Response Memory Test | ✅ Pass | 3.773s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.681s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.349s |  |
| Penetration Testing Methodology | ✅ Pass | 9.091s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.360s |  |
| SQL Injection Attack Type | ✅ Pass | 3.971s |  |
| Penetration Testing Framework | ✅ Pass | 8.359s |  |
| Web Application Security Scanner | ✅ Pass | 7.103s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.145s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.382s

---

### generator (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.331s |  |
| Text Transform Uppercase | ✅ Pass | 2.801s |  |
| Count from 1 to 5 | ✅ Pass | 2.718s |  |
| Math Calculation | ✅ Pass | 2.145s |  |
| Basic Echo Function | ✅ Pass | 2.633s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.981s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.889s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.207s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.481s |  |
| Search Query Function | ✅ Pass | 2.913s |  |
| Ask Advice Function | ✅ Pass | 3.554s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.233s |  |
| Basic Context Memory Test | ✅ Pass | 2.935s |  |
| Function Argument Memory Test | ✅ Pass | 3.494s |  |
| Function Response Memory Test | ✅ Pass | 3.110s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.146s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.406s |  |
| Penetration Testing Methodology | ✅ Pass | 10.157s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.651s |  |
| SQL Injection Attack Type | ✅ Pass | 3.755s |  |
| Penetration Testing Framework | ✅ Pass | 6.823s |  |
| Web Application Security Scanner | ✅ Pass | 7.040s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.429s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.167s

---

### refiner (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.746s |  |
| Text Transform Uppercase | ✅ Pass | 2.969s |  |
| Count from 1 to 5 | ✅ Pass | 2.755s |  |
| Math Calculation | ✅ Pass | 2.273s |  |
| Basic Echo Function | ✅ Pass | 3.325s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.426s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.609s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.135s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.468s |  |
| Search Query Function | ✅ Pass | 3.214s |  |
| Ask Advice Function | ✅ Pass | 3.188s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.096s |  |
| Basic Context Memory Test | ✅ Pass | 3.181s |  |
| Function Argument Memory Test | ✅ Pass | 3.648s |  |
| Function Response Memory Test | ✅ Pass | 3.848s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.777s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.207s |  |
| Penetration Testing Methodology | ✅ Pass | 10.975s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.284s |  |
| SQL Injection Attack Type | ✅ Pass | 4.517s |  |
| Penetration Testing Framework | ✅ Pass | 8.192s |  |
| Web Application Security Scanner | ✅ Pass | 6.736s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.650s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.227s

---

### adviser (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.680s |  |
| Text Transform Uppercase | ✅ Pass | 3.380s |  |
| Count from 1 to 5 | ✅ Pass | 2.500s |  |
| Math Calculation | ✅ Pass | 2.176s |  |
| Basic Echo Function | ✅ Pass | 2.803s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.438s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.928s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.740s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.857s |  |
| Search Query Function | ✅ Pass | 2.891s |  |
| Ask Advice Function | ✅ Pass | 3.168s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.775s |  |
| Basic Context Memory Test | ✅ Pass | 3.165s |  |
| Function Argument Memory Test | ✅ Pass | 4.115s |  |
| Function Response Memory Test | ✅ Pass | 3.482s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.830s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.240s |  |
| Penetration Testing Methodology | ✅ Pass | 9.737s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.344s |  |
| SQL Injection Attack Type | ✅ Pass | 4.260s |  |
| Penetration Testing Framework | ✅ Pass | 6.819s |  |
| Web Application Security Scanner | ✅ Pass | 7.147s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.730s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.096s

---

### reflector (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.029s |  |
| Text Transform Uppercase | ✅ Pass | 1.997s |  |
| Count from 1 to 5 | ✅ Pass | 2.659s |  |
| Math Calculation | ✅ Pass | 2.284s |  |
| Basic Echo Function | ✅ Pass | 2.336s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.075s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.070s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.481s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.487s |  |
| Search Query Function | ✅ Pass | 2.302s |  |
| Ask Advice Function | ✅ Pass | 2.844s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.296s |  |
| Basic Context Memory Test | ✅ Pass | 2.022s |  |
| Function Argument Memory Test | ✅ Pass | 4.029s |  |
| Function Response Memory Test | ✅ Pass | 2.100s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.214s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.157s |  |
| Penetration Testing Methodology | ✅ Pass | 5.906s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.945s |  |
| SQL Injection Attack Type | ✅ Pass | 2.673s |  |
| Penetration Testing Framework | ✅ Pass | 4.573s |  |
| Web Application Security Scanner | ✅ Pass | 3.583s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.821s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.908s

---

### searcher (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.493s |  |
| Text Transform Uppercase | ✅ Pass | 2.703s |  |
| Count from 1 to 5 | ✅ Pass | 2.981s |  |
| Math Calculation | ✅ Pass | 2.683s |  |
| Basic Echo Function | ✅ Pass | 2.894s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.566s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.620s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.857s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.266s |  |
| Search Query Function | ✅ Pass | 2.764s |  |
| Ask Advice Function | ✅ Pass | 3.466s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.572s |  |
| Basic Context Memory Test | ✅ Pass | 3.479s |  |
| Function Argument Memory Test | ✅ Pass | 3.080s |  |
| Function Response Memory Test | ✅ Pass | 3.733s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.564s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.052s |  |
| Penetration Testing Methodology | ✅ Pass | 10.896s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.396s |  |
| SQL Injection Attack Type | ✅ Pass | 4.190s |  |
| Penetration Testing Framework | ✅ Pass | 6.489s |  |
| Web Application Security Scanner | ✅ Pass | 5.714s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.478s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.085s

---

### enricher (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.205s |  |
| Text Transform Uppercase | ✅ Pass | 2.111s |  |
| Count from 1 to 5 | ✅ Pass | 2.530s |  |
| Math Calculation | ✅ Pass | 2.086s |  |
| Basic Echo Function | ✅ Pass | 2.449s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.038s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.955s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.454s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.747s |  |
| Search Query Function | ✅ Pass | 2.468s |  |
| Ask Advice Function | ✅ Pass | 2.754s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.281s |  |
| Basic Context Memory Test | ✅ Pass | 2.564s |  |
| Function Argument Memory Test | ✅ Pass | 2.071s |  |
| Function Response Memory Test | ✅ Pass | 2.007s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.589s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.348s |  |
| Penetration Testing Methodology | ✅ Pass | 6.553s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.401s |  |
| SQL Injection Attack Type | ✅ Pass | 2.337s |  |
| Penetration Testing Framework | ✅ Pass | 4.532s |  |
| Web Application Security Scanner | ✅ Pass | 3.918s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.114s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.936s

---

### coder (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.685s |  |
| Text Transform Uppercase | ✅ Pass | 2.832s |  |
| Count from 1 to 5 | ✅ Pass | 2.675s |  |
| Math Calculation | ✅ Pass | 2.334s |  |
| Basic Echo Function | ✅ Pass | 2.907s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.720s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.285s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.862s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.917s |  |
| Search Query Function | ✅ Pass | 2.919s |  |
| Ask Advice Function | ✅ Pass | 3.222s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.683s |  |
| Basic Context Memory Test | ✅ Pass | 3.702s |  |
| Function Argument Memory Test | ✅ Pass | 3.559s |  |
| Function Response Memory Test | ✅ Pass | 3.141s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.236s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.343s |  |
| Penetration Testing Methodology | ✅ Pass | 8.124s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.131s |  |
| SQL Injection Attack Type | ✅ Pass | 3.982s |  |
| Penetration Testing Framework | ✅ Pass | 8.877s |  |
| Web Application Security Scanner | ✅ Pass | 5.495s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.263s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.996s

---

### installer (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.704s |  |
| Text Transform Uppercase | ✅ Pass | 2.771s |  |
| Count from 1 to 5 | ✅ Pass | 3.838s |  |
| Math Calculation | ✅ Pass | 2.479s |  |
| Basic Echo Function | ✅ Pass | 3.186s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.786s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.481s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.082s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.988s |  |
| Search Query Function | ✅ Pass | 3.367s |  |
| Ask Advice Function | ✅ Pass | 3.296s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.595s |  |
| Basic Context Memory Test | ✅ Pass | 3.357s |  |
| Function Argument Memory Test | ✅ Pass | 3.017s |  |
| Function Response Memory Test | ✅ Pass | 3.952s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.144s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.279s |  |
| Penetration Testing Methodology | ✅ Pass | 10.138s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.926s |  |
| SQL Injection Attack Type | ✅ Pass | 6.682s |  |
| Penetration Testing Framework | ✅ Pass | 8.627s |  |
| Web Application Security Scanner | ✅ Pass | 6.695s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.916s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.318s

---

### pentester (claude-sonnet-4-5-20250929)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.077s |  |
| Text Transform Uppercase | ✅ Pass | 2.711s |  |
| Count from 1 to 5 | ✅ Pass | 2.561s |  |
| Math Calculation | ✅ Pass | 2.228s |  |
| Basic Echo Function | ✅ Pass | 2.846s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.301s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.961s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.601s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.922s |  |
| Search Query Function | ✅ Pass | 2.944s |  |
| Ask Advice Function | ✅ Pass | 3.221s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.898s |  |
| Basic Context Memory Test | ✅ Pass | 3.706s |  |
| Function Argument Memory Test | ✅ Pass | 2.692s |  |
| Function Response Memory Test | ✅ Pass | 3.924s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.041s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.868s |  |
| Penetration Testing Methodology | ✅ Pass | 9.312s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.902s |  |
| SQL Injection Attack Type | ✅ Pass | 3.950s |  |
| Penetration Testing Framework | ✅ Pass | 6.823s |  |
| Web Application Security Scanner | ✅ Pass | 6.797s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.233s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.936s

---

