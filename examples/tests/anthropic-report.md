# LLM Agent Testing Report

Generated: Tue, 08 Jul 2025 21:35:47 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-3-5-haiku-20241022 | false | 18/18 (100.00%) | 3.526s |
| simple_json | claude-3-5-haiku-20241022 | false | 4/4 (100.00%) | 1.645s |
| primary_agent | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 3.123s |
| assistant | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 3.083s |
| generator | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 3.015s |
| refiner | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 3.258s |
| adviser | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 2.901s |
| reflector | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 3.054s |
| searcher | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 2.754s |
| enricher | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 3.125s |
| coder | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 2.801s |
| installer | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 2.805s |
| pentester | claude-sonnet-4-20250514 | false | 18/18 (100.00%) | 2.636s |

**Total**: 220/220 (100.00%) successful tests
**Overall average latency**: 2.982s

## Detailed Results

### simple (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.809s |  |
| Text Transform Uppercase | ✅ Pass | 1.355s |  |
| Count from 1 to 5 | ✅ Pass | 2.051s |  |
| Math Calculation | ✅ Pass | 1.752s |  |
| Basic Echo Function | ✅ Pass | 1.966s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.870s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.349s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.973s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.981s |  |
| Ask Advice Function | ✅ Pass | 2.920s |  |
| Search Query Function | ✅ Pass | 13.106s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.498s |  |
| Penetration Testing Methodology | ✅ Pass | 5.128s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.526s |  |
| SQL Injection Attack Type | ✅ Pass | 1.763s |  |
| Penetration Testing Framework | ✅ Pass | 6.313s |  |
| Web Application Security Scanner | ✅ Pass | 4.197s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.903s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.526s

---

### simple_json (claude-3-5-haiku-20241022)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.523s |  |
| Project Information JSON | ✅ Pass | 1.400s |  |
| User Profile JSON | ✅ Pass | 1.372s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 2.282s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 1.645s

---

### primary_agent (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.840s |  |
| Text Transform Uppercase | ✅ Pass | 2.240s |  |
| Count from 1 to 5 | ✅ Pass | 1.642s |  |
| Math Calculation | ✅ Pass | 1.995s |  |
| Basic Echo Function | ✅ Pass | 2.092s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.466s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.337s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.040s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.489s |  |
| Search Query Function | ✅ Pass | 2.305s |  |
| Ask Advice Function | ✅ Pass | 2.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.946s |  |
| Penetration Testing Methodology | ✅ Pass | 4.019s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.258s |  |
| SQL Injection Attack Type | ✅ Pass | 1.460s |  |
| Penetration Testing Framework | ✅ Pass | 6.952s |  |
| Web Application Security Scanner | ✅ Pass | 4.724s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.203s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.123s

---

### assistant (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.859s |  |
| Text Transform Uppercase | ✅ Pass | 2.402s |  |
| Count from 1 to 5 | ✅ Pass | 6.359s |  |
| Math Calculation | ✅ Pass | 2.073s |  |
| Basic Echo Function | ✅ Pass | 2.211s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.737s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.999s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.082s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.999s |  |
| Search Query Function | ✅ Pass | 2.803s |  |
| Ask Advice Function | ✅ Pass | 2.310s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.179s |  |
| Penetration Testing Methodology | ✅ Pass | 5.120s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.986s |  |
| SQL Injection Attack Type | ✅ Pass | 1.578s |  |
| Penetration Testing Framework | ✅ Pass | 4.496s |  |
| Web Application Security Scanner | ✅ Pass | 5.019s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.266s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.083s

---

### generator (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.284s |  |
| Text Transform Uppercase | ✅ Pass | 1.539s |  |
| Count from 1 to 5 | ✅ Pass | 1.873s |  |
| Math Calculation | ✅ Pass | 2.077s |  |
| Basic Echo Function | ✅ Pass | 2.382s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.592s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.383s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.080s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.823s |  |
| Search Query Function | ✅ Pass | 2.383s |  |
| Ask Advice Function | ✅ Pass | 2.319s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.590s |  |
| Penetration Testing Methodology | ✅ Pass | 4.117s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.945s |  |
| SQL Injection Attack Type | ✅ Pass | 2.536s |  |
| Penetration Testing Framework | ✅ Pass | 5.067s |  |
| Web Application Security Scanner | ✅ Pass | 5.005s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.258s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.015s

---

### refiner (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.761s |  |
| Text Transform Uppercase | ✅ Pass | 2.254s |  |
| Count from 1 to 5 | ✅ Pass | 6.548s |  |
| Math Calculation | ✅ Pass | 2.120s |  |
| Basic Echo Function | ✅ Pass | 2.504s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.006s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.680s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.119s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.090s |  |
| Search Query Function | ✅ Pass | 3.247s |  |
| Ask Advice Function | ✅ Pass | 2.372s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.175s |  |
| Penetration Testing Methodology | ✅ Pass | 4.501s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.222s |  |
| SQL Injection Attack Type | ✅ Pass | 2.697s |  |
| Penetration Testing Framework | ✅ Pass | 6.013s |  |
| Web Application Security Scanner | ✅ Pass | 4.046s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.286s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.258s

---

### adviser (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.144s |  |
| Text Transform Uppercase | ✅ Pass | 2.189s |  |
| Count from 1 to 5 | ✅ Pass | 2.250s |  |
| Math Calculation | ✅ Pass | 1.737s |  |
| Basic Echo Function | ✅ Pass | 1.956s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.855s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.192s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.285s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.045s |  |
| Search Query Function | ✅ Pass | 2.315s |  |
| Ask Advice Function | ✅ Pass | 2.395s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.306s |  |
| Penetration Testing Methodology | ✅ Pass | 6.272s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.580s |  |
| SQL Injection Attack Type | ✅ Pass | 1.629s |  |
| Penetration Testing Framework | ✅ Pass | 5.180s |  |
| Web Application Security Scanner | ✅ Pass | 3.948s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.932s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.901s

---

### reflector (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.786s |  |
| Text Transform Uppercase | ✅ Pass | 2.385s |  |
| Count from 1 to 5 | ✅ Pass | 2.348s |  |
| Math Calculation | ✅ Pass | 3.790s |  |
| Basic Echo Function | ✅ Pass | 2.252s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.740s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.851s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.361s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.200s |  |
| Search Query Function | ✅ Pass | 3.251s |  |
| Ask Advice Function | ✅ Pass | 2.456s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.388s |  |
| Penetration Testing Methodology | ✅ Pass | 4.488s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.247s |  |
| SQL Injection Attack Type | ✅ Pass | 2.038s |  |
| Penetration Testing Framework | ✅ Pass | 5.000s |  |
| Web Application Security Scanner | ✅ Pass | 4.222s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.156s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.054s

---

### searcher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.724s |  |
| Text Transform Uppercase | ✅ Pass | 1.625s |  |
| Count from 1 to 5 | ✅ Pass | 2.571s |  |
| Math Calculation | ✅ Pass | 1.738s |  |
| Basic Echo Function | ✅ Pass | 2.129s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.791s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.584s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.303s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.531s |  |
| Search Query Function | ✅ Pass | 2.985s |  |
| Ask Advice Function | ✅ Pass | 2.615s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.054s |  |
| Penetration Testing Methodology | ✅ Pass | 3.740s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.222s |  |
| SQL Injection Attack Type | ✅ Pass | 2.129s |  |
| Penetration Testing Framework | ✅ Pass | 6.153s |  |
| Web Application Security Scanner | ✅ Pass | 3.595s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.076s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.754s

---

### enricher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.097s |  |
| Text Transform Uppercase | ✅ Pass | 2.588s |  |
| Count from 1 to 5 | ✅ Pass | 1.988s |  |
| Math Calculation | ✅ Pass | 1.959s |  |
| Basic Echo Function | ✅ Pass | 2.494s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.639s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.009s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.077s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.433s |  |
| Search Query Function | ✅ Pass | 2.518s |  |
| Ask Advice Function | ✅ Pass | 2.182s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.641s |  |
| Penetration Testing Methodology | ✅ Pass | 4.323s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.595s |  |
| SQL Injection Attack Type | ✅ Pass | 1.972s |  |
| Penetration Testing Framework | ✅ Pass | 7.677s |  |
| Web Application Security Scanner | ✅ Pass | 4.585s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.473s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.125s

---

### coder (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.205s |  |
| Text Transform Uppercase | ✅ Pass | 1.880s |  |
| Count from 1 to 5 | ✅ Pass | 2.332s |  |
| Math Calculation | ✅ Pass | 1.758s |  |
| Basic Echo Function | ✅ Pass | 2.311s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.558s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.770s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.978s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.398s |  |
| Search Query Function | ✅ Pass | 2.632s |  |
| Ask Advice Function | ✅ Pass | 2.960s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.345s |  |
| Penetration Testing Methodology | ✅ Pass | 4.352s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.097s |  |
| SQL Injection Attack Type | ✅ Pass | 2.402s |  |
| Penetration Testing Framework | ✅ Pass | 4.698s |  |
| Web Application Security Scanner | ✅ Pass | 4.126s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.606s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.801s

---

### installer (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.443s |  |
| Text Transform Uppercase | ✅ Pass | 2.094s |  |
| Count from 1 to 5 | ✅ Pass | 1.699s |  |
| Math Calculation | ✅ Pass | 1.679s |  |
| Basic Echo Function | ✅ Pass | 2.019s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.157s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.611s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.992s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.521s |  |
| Search Query Function | ✅ Pass | 1.871s |  |
| Ask Advice Function | ✅ Pass | 2.911s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.641s |  |
| Penetration Testing Methodology | ✅ Pass | 4.667s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.004s |  |
| SQL Injection Attack Type | ✅ Pass | 2.037s |  |
| Penetration Testing Framework | ✅ Pass | 5.193s |  |
| Web Application Security Scanner | ✅ Pass | 3.999s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.934s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.805s

---

### pentester (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.994s |  |
| Text Transform Uppercase | ✅ Pass | 1.585s |  |
| Count from 1 to 5 | ✅ Pass | 1.496s |  |
| Math Calculation | ✅ Pass | 2.149s |  |
| Basic Echo Function | ✅ Pass | 2.427s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.696s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.245s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.269s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.589s |  |
| Search Query Function | ✅ Pass | 2.282s |  |
| Ask Advice Function | ✅ Pass | 2.583s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.577s |  |
| Penetration Testing Methodology | ✅ Pass | 2.987s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.187s |  |
| SQL Injection Attack Type | ✅ Pass | 2.695s |  |
| Penetration Testing Framework | ✅ Pass | 4.714s |  |
| Web Application Security Scanner | ✅ Pass | 3.759s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.197s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.636s

---

