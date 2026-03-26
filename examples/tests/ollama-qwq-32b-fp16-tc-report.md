# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 20:33:51 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 6.716s |
| simple_json | qwq:32b-fp16-tc | true | 5/5 (100.00%) | 6.216s |
| primary_agent | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 9.193s |
| assistant | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 8.104s |
| generator | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 9.544s |
| refiner | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 9.373s |
| adviser | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 8.474s |
| reflector | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 8.746s |
| searcher | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 8.270s |
| enricher | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 10.131s |
| coder | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 8.886s |
| installer | qwq:32b-fp16-tc | true | 22/23 (95.65%) | 8.990s |
| pentester | qwq:32b-fp16-tc | true | 23/23 (100.00%) | 10.520s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 8.864s

## Detailed Results

### simple (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.681s |  |
| Text Transform Uppercase | ✅ Pass | 4.573s |  |
| Count from 1 to 5 | ✅ Pass | 10.128s |  |
| Math Calculation | ✅ Pass | 5.587s |  |
| Basic Echo Function | ✅ Pass | 2.728s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.202s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.603s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.625s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.333s |  |
| Search Query Function | ✅ Pass | 3.209s |  |
| Ask Advice Function | ✅ Pass | 3.321s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.065s |  |
| Basic Context Memory Test | ✅ Pass | 3.660s |  |
| Function Argument Memory Test | ✅ Pass | 5.600s |  |
| Function Response Memory Test | ✅ Pass | 3.156s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.576s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.602s |  |
| Penetration Testing Methodology | ✅ Pass | 15.436s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.553s |  |
| SQL Injection Attack Type | ✅ Pass | 7.660s |  |
| Penetration Testing Framework | ✅ Pass | 15.103s |  |
| Web Application Security Scanner | ✅ Pass | 10.527s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.523s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.716s

---

### simple_json (qwq:32b-fp16-tc)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 11.014s |  |
| Person Information JSON | ✅ Pass | 6.958s |  |
| Project Information JSON | ✅ Pass | 4.410s |  |
| User Profile JSON | ✅ Pass | 3.958s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 4.737s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 6.216s

---

### primary_agent (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.427s |  |
| Text Transform Uppercase | ✅ Pass | 5.094s |  |
| Count from 1 to 5 | ✅ Pass | 6.045s |  |
| Math Calculation | ✅ Pass | 8.572s |  |
| Basic Echo Function | ✅ Pass | 5.126s |  |
| Streaming Simple Math Streaming | ✅ Pass | 7.438s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.221s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.700s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.894s |  |
| Search Query Function | ✅ Pass | 4.260s |  |
| Ask Advice Function | ✅ Pass | 4.531s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.014s |  |
| Basic Context Memory Test | ✅ Pass | 4.387s |  |
| Function Argument Memory Test | ✅ Pass | 5.627s |  |
| Function Response Memory Test | ✅ Pass | 6.668s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.791s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.742s |  |
| Penetration Testing Methodology | ✅ Pass | 23.601s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.807s |  |
| SQL Injection Attack Type | ✅ Pass | 27.442s |  |
| Penetration Testing Framework | ✅ Pass | 23.325s |  |
| Web Application Security Scanner | ✅ Pass | 15.780s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.938s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.193s

---

### assistant (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 10.946s |  |
| Text Transform Uppercase | ✅ Pass | 6.941s |  |
| Count from 1 to 5 | ✅ Pass | 4.256s |  |
| Math Calculation | ✅ Pass | 11.927s |  |
| Basic Echo Function | ✅ Pass | 4.216s |  |
| Streaming Simple Math Streaming | ✅ Pass | 10.500s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.883s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.938s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.806s |  |
| Search Query Function | ✅ Pass | 5.634s |  |
| Ask Advice Function | ✅ Pass | 4.006s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.245s |  |
| Basic Context Memory Test | ✅ Pass | 3.060s |  |
| Function Argument Memory Test | ✅ Pass | 4.733s |  |
| Function Response Memory Test | ✅ Pass | 8.668s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 12.198s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.656s |  |
| Penetration Testing Methodology | ✅ Pass | 12.831s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.861s |  |
| SQL Injection Attack Type | ✅ Pass | 8.588s |  |
| Penetration Testing Framework | ✅ Pass | 17.076s |  |
| Web Application Security Scanner | ✅ Pass | 14.477s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.937s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.104s

---

### generator (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.366s |  |
| Text Transform Uppercase | ✅ Pass | 4.111s |  |
| Count from 1 to 5 | ✅ Pass | 4.739s |  |
| Math Calculation | ✅ Pass | 12.855s |  |
| Basic Echo Function | ✅ Pass | 4.534s |  |
| Streaming Simple Math Streaming | ✅ Pass | 10.861s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.929s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.671s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.297s |  |
| Search Query Function | ✅ Pass | 7.993s |  |
| Ask Advice Function | ✅ Pass | 3.878s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.270s |  |
| Basic Context Memory Test | ✅ Pass | 3.761s |  |
| Function Argument Memory Test | ✅ Pass | 4.728s |  |
| Function Response Memory Test | ✅ Pass | 4.591s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 29.808s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.493s |  |
| Penetration Testing Methodology | ✅ Pass | 18.866s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.203s |  |
| SQL Injection Attack Type | ✅ Pass | 20.241s |  |
| Penetration Testing Framework | ✅ Pass | 19.454s |  |
| Web Application Security Scanner | ✅ Pass | 13.553s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.303s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.544s

---

### refiner (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.994s |  |
| Text Transform Uppercase | ✅ Pass | 6.657s |  |
| Count from 1 to 5 | ✅ Pass | 4.197s |  |
| Math Calculation | ✅ Pass | 12.493s |  |
| Basic Echo Function | ✅ Pass | 4.838s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.617s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.921s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.528s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.596s |  |
| Search Query Function | ✅ Pass | 8.016s |  |
| Ask Advice Function | ✅ Pass | 4.720s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.481s |  |
| Basic Context Memory Test | ✅ Pass | 3.840s |  |
| Function Argument Memory Test | ✅ Pass | 8.249s |  |
| Function Response Memory Test | ✅ Pass | 24.309s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.445s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.230s |  |
| Penetration Testing Methodology | ✅ Pass | 16.988s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.847s |  |
| SQL Injection Attack Type | ✅ Pass | 22.903s |  |
| Penetration Testing Framework | ✅ Pass | 18.108s |  |
| Web Application Security Scanner | ✅ Pass | 12.641s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.945s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.373s

---

### adviser (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 8.448s |  |
| Text Transform Uppercase | ✅ Pass | 5.223s |  |
| Count from 1 to 5 | ✅ Pass | 4.137s |  |
| Math Calculation | ✅ Pass | 29.630s |  |
| Basic Echo Function | ✅ Pass | 3.791s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.284s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.324s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.104s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.474s |  |
| Search Query Function | ✅ Pass | 5.012s |  |
| Ask Advice Function | ✅ Pass | 3.713s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.286s |  |
| Basic Context Memory Test | ✅ Pass | 4.592s |  |
| Function Argument Memory Test | ✅ Pass | 9.007s |  |
| Function Response Memory Test | ✅ Pass | 4.417s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.419s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.139s |  |
| Penetration Testing Methodology | ✅ Pass | 13.577s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.854s |  |
| SQL Injection Attack Type | ✅ Pass | 9.491s |  |
| Penetration Testing Framework | ✅ Pass | 14.146s |  |
| Web Application Security Scanner | ✅ Pass | 11.518s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.300s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.474s

---

### reflector (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.056s |  |
| Text Transform Uppercase | ✅ Pass | 4.968s |  |
| Count from 1 to 5 | ✅ Pass | 4.893s |  |
| Math Calculation | ✅ Pass | 9.789s |  |
| Basic Echo Function | ✅ Pass | 4.689s |  |
| Streaming Simple Math Streaming | ✅ Pass | 17.710s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.866s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 6.350s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.813s |  |
| Search Query Function | ✅ Pass | 6.374s |  |
| Ask Advice Function | ✅ Pass | 3.841s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.710s |  |
| Basic Context Memory Test | ✅ Pass | 4.339s |  |
| Function Argument Memory Test | ✅ Pass | 6.259s |  |
| Function Response Memory Test | ✅ Pass | 13.187s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.633s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.170s |  |
| Penetration Testing Methodology | ✅ Pass | 17.012s |  |
| Vulnerability Assessment Tools | ✅ Pass | 20.805s |  |
| SQL Injection Attack Type | ✅ Pass | 9.169s |  |
| Penetration Testing Framework | ✅ Pass | 17.306s |  |
| Web Application Security Scanner | ✅ Pass | 16.287s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.913s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.746s

---

### searcher (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.501s |  |
| Text Transform Uppercase | ✅ Pass | 5.733s |  |
| Count from 1 to 5 | ✅ Pass | 4.384s |  |
| Math Calculation | ✅ Pass | 19.789s |  |
| Basic Echo Function | ✅ Pass | 3.466s |  |
| Streaming Simple Math Streaming | ✅ Pass | 11.112s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.044s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.030s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.001s |  |
| Search Query Function | ✅ Pass | 7.560s |  |
| Ask Advice Function | ✅ Pass | 4.992s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.649s |  |
| Basic Context Memory Test | ✅ Pass | 4.280s |  |
| Function Argument Memory Test | ✅ Pass | 11.166s |  |
| Function Response Memory Test | ✅ Pass | 4.679s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.225s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.642s |  |
| Penetration Testing Methodology | ✅ Pass | 18.262s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.810s |  |
| SQL Injection Attack Type | ✅ Pass | 10.062s |  |
| Penetration Testing Framework | ✅ Pass | 17.466s |  |
| Web Application Security Scanner | ✅ Pass | 13.754s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.590s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.270s

---

### enricher (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.136s |  |
| Text Transform Uppercase | ✅ Pass | 6.673s |  |
| Count from 1 to 5 | ✅ Pass | 4.038s |  |
| Math Calculation | ✅ Pass | 18.707s |  |
| Basic Echo Function | ✅ Pass | 4.421s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.519s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.789s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.283s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.865s |  |
| Search Query Function | ✅ Pass | 10.054s |  |
| Ask Advice Function | ✅ Pass | 3.730s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.154s |  |
| Basic Context Memory Test | ✅ Pass | 4.669s |  |
| Function Argument Memory Test | ✅ Pass | 3.649s |  |
| Function Response Memory Test | ✅ Pass | 16.702s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.791s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.484s |  |
| Penetration Testing Methodology | ✅ Pass | 18.141s |  |
| Vulnerability Assessment Tools | ✅ Pass | 22.787s |  |
| SQL Injection Attack Type | ✅ Pass | 39.473s |  |
| Penetration Testing Framework | ✅ Pass | 18.883s |  |
| Web Application Security Scanner | ✅ Pass | 12.108s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.941s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 10.131s

---

### coder (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.236s |  |
| Text Transform Uppercase | ✅ Pass | 5.392s |  |
| Count from 1 to 5 | ✅ Pass | 5.107s |  |
| Math Calculation | ✅ Pass | 8.484s |  |
| Basic Echo Function | ✅ Pass | 4.541s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.311s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.351s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.162s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.865s |  |
| Search Query Function | ✅ Pass | 15.405s |  |
| Ask Advice Function | ✅ Pass | 4.197s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.541s |  |
| Basic Context Memory Test | ✅ Pass | 3.293s |  |
| Function Argument Memory Test | ✅ Pass | 5.456s |  |
| Function Response Memory Test | ✅ Pass | 11.370s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 15.621s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.115s |  |
| Penetration Testing Methodology | ✅ Pass | 22.034s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.513s |  |
| SQL Injection Attack Type | ✅ Pass | 18.884s |  |
| Penetration Testing Framework | ✅ Pass | 12.967s |  |
| Web Application Security Scanner | ✅ Pass | 9.560s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.956s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.886s

---

### installer (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.613s |  |
| Text Transform Uppercase | ✅ Pass | 5.875s |  |
| Count from 1 to 5 | ✅ Pass | 3.987s |  |
| Math Calculation | ✅ Pass | 23.690s |  |
| Basic Echo Function | ✅ Pass | 3.616s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.350s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.202s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.302s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.813s |  |
| Search Query Function | ❌ Fail | 6.340s | expected function 'search' not found in tool calls: expected function search not found in tool calls |
| Ask Advice Function | ✅ Pass | 3.913s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.425s |  |
| Basic Context Memory Test | ✅ Pass | 3.478s |  |
| Function Argument Memory Test | ✅ Pass | 6.654s |  |
| Function Response Memory Test | ✅ Pass | 5.056s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.050s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.954s |  |
| Penetration Testing Methodology | ✅ Pass | 15.131s |  |
| Vulnerability Assessment Tools | ✅ Pass | 20.484s |  |
| SQL Injection Attack Type | ✅ Pass | 27.444s |  |
| Penetration Testing Framework | ✅ Pass | 12.985s |  |
| Web Application Security Scanner | ✅ Pass | 15.344s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.053s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 8.990s

---

### pentester (qwq:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.272s |  |
| Text Transform Uppercase | ✅ Pass | 5.369s |  |
| Count from 1 to 5 | ✅ Pass | 3.969s |  |
| Math Calculation | ✅ Pass | 20.641s |  |
| Basic Echo Function | ✅ Pass | 3.630s |  |
| Streaming Simple Math Streaming | ✅ Pass | 8.335s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.560s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.832s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.319s |  |
| Search Query Function | ✅ Pass | 7.127s |  |
| Ask Advice Function | ✅ Pass | 4.739s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 6.342s |  |
| Basic Context Memory Test | ✅ Pass | 4.692s |  |
| Function Argument Memory Test | ✅ Pass | 12.869s |  |
| Function Response Memory Test | ✅ Pass | 26.694s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.736s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.734s |  |
| Penetration Testing Methodology | ✅ Pass | 18.070s |  |
| Vulnerability Assessment Tools | ✅ Pass | 25.093s |  |
| SQL Injection Attack Type | ✅ Pass | 34.538s |  |
| Penetration Testing Framework | ✅ Pass | 9.951s |  |
| Web Application Security Scanner | ✅ Pass | 11.550s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.882s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 10.520s

---

