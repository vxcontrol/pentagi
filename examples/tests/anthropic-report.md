# LLM Agent Testing Report

Generated: Thu, 23 Jul 2026 13:02:11 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-haiku-4-5 | false | 24/24 (100.00%) | 1.855s |
| simple_json | claude-haiku-4-5 | false | 7/7 (100.00%) | 3.020s |
| primary_agent | claude-sonnet-5 | true | 25/25 (100.00%) | 2.759s |
| assistant | claude-sonnet-5 | true | 25/25 (100.00%) | 2.756s |
| generator | claude-opus-4-8 | true | 24/25 (96.00%) | 3.585s |
| refiner | claude-opus-4-8 | true | 25/25 (100.00%) | 3.616s |
| adviser | claude-sonnet-5 | true | 25/25 (100.00%) | 3.526s |
| reflector | claude-haiku-4-5 | true | 24/24 (100.00%) | 2.680s |
| searcher | claude-haiku-4-5 | true | 24/24 (100.00%) | 2.461s |
| enricher | claude-haiku-4-5 | true | 24/24 (100.00%) | 1.837s |
| coder | claude-sonnet-5 | true | 25/25 (100.00%) | 3.079s |
| installer | claude-sonnet-5 | true | 25/25 (100.00%) | 3.176s |
| pentester | claude-sonnet-5 | true | 25/25 (100.00%) | 2.655s |

**Total**: 302/303 (99.67%) successful tests
**Overall average latency**: 2.845s

## Detailed Results

### simple (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.398s |  |
| Text Transform Uppercase | ✅ Pass | 1.879s |  |
| Count from 1 to 5 | ✅ Pass | 1.208s |  |
| Math Calculation | ✅ Pass | 1.107s |  |
| Basic Echo Function | ✅ Pass | 1.796s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.855s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.942s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.173s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.194s |  |
| Search Query Function | ✅ Pass | 2.017s |  |
| Ask Advice Function | ✅ Pass | 2.985s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.315s |  |
| Basic Context Memory Test | ✅ Pass | 1.026s |  |
| Function Argument Memory Test | ✅ Pass | 0.846s |  |
| Function Response Memory Test | ✅ Pass | 0.877s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.913s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.393s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.646s |  |
| Penetration Testing Methodology | ✅ Pass | 3.910s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.134s |  |
| SQL Injection Attack Type | ✅ Pass | 0.975s |  |
| Penetration Testing Framework | ✅ Pass | 2.983s |  |
| Web Application Security Scanner | ✅ Pass | 2.261s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.667s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.855s

---

### simple_json (claude-haiku-4-5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 5.449s |  |
| Person Information JSON | ✅ Pass | 1.652s |  |
| User Profile JSON | ✅ Pass | 1.108s |  |
| Project Information JSON | ✅ Pass | 2.308s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.938s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.357s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 8.323s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 3.020s

---

### primary_agent (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.234s |  |
| Text Transform Uppercase | ✅ Pass | 1.834s |  |
| Count from 1 to 5 | ✅ Pass | 1.676s |  |
| Math Calculation | ✅ Pass | 2.138s |  |
| Basic Echo Function | ✅ Pass | 1.840s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.735s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.986s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.609s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.963s |  |
| Search Query Function | ✅ Pass | 2.208s |  |
| Ask Advice Function | ✅ Pass | 2.037s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.797s |  |
| Basic Context Memory Test | ✅ Pass | 1.733s |  |
| Function Argument Memory Test | ✅ Pass | 1.167s |  |
| Function Response Memory Test | ✅ Pass | 1.552s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.965s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.059s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.190s |  |
| Penetration Testing Methodology | ✅ Pass | 5.149s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.820s |  |
| SQL Injection Attack Type | ✅ Pass | 2.948s |  |
| Penetration Testing Framework | ✅ Pass | 3.259s |  |
| Web Application Security Scanner | ✅ Pass | 3.629s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.398s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.045s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.759s

---

### assistant (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.093s |  |
| Text Transform Uppercase | ✅ Pass | 1.848s |  |
| Count from 1 to 5 | ✅ Pass | 1.720s |  |
| Math Calculation | ✅ Pass | 1.742s |  |
| Basic Echo Function | ✅ Pass | 2.020s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.842s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.933s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.269s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.872s |  |
| Search Query Function | ✅ Pass | 2.514s |  |
| Ask Advice Function | ✅ Pass | 2.134s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.352s |  |
| Basic Context Memory Test | ✅ Pass | 1.616s |  |
| Function Argument Memory Test | ✅ Pass | 1.140s |  |
| Function Response Memory Test | ✅ Pass | 2.854s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.230s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.660s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.565s |  |
| Penetration Testing Methodology | ✅ Pass | 4.917s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.262s |  |
| SQL Injection Attack Type | ✅ Pass | 2.401s |  |
| Penetration Testing Framework | ✅ Pass | 2.604s |  |
| Web Application Security Scanner | ✅ Pass | 4.052s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.192s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.044s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.756s

---

### generator (claude-opus-4-8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.076s |  |
| Text Transform Uppercase | ✅ Pass | 1.825s |  |
| Count from 1 to 5 | ✅ Pass | 1.902s |  |
| Math Calculation | ✅ Pass | 1.923s |  |
| Basic Echo Function | ✅ Pass | 1.835s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.536s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.587s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.343s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.416s |  |
| Search Query Function | ✅ Pass | 2.193s |  |
| Ask Advice Function | ✅ Pass | 3.601s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.942s |  |
| Basic Context Memory Test | ✅ Pass | 1.566s |  |
| Function Argument Memory Test | ✅ Pass | 2.205s |  |
| Function Response Memory Test | ✅ Pass | 1.949s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 4.818s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.702s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 6.894s |  |
| Penetration Testing Methodology | ✅ Pass | 7.353s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.864s |  |
| SQL Injection Attack Type | ✅ Pass | 4.188s |  |
| Penetration Testing Framework | ✅ Pass | 4.760s |  |
| Web Application Security Scanner | ✅ Pass | 4.269s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.857s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 5.000s |  |

**Summary**: 24/25 (96.00%) successful tests

**Average latency**: 3.585s

---

### refiner (claude-opus-4-8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.877s |  |
| Text Transform Uppercase | ✅ Pass | 2.332s |  |
| Count from 1 to 5 | ✅ Pass | 1.613s |  |
| Math Calculation | ✅ Pass | 1.848s |  |
| Basic Echo Function | ✅ Pass | 3.425s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.034s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.435s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.421s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.115s |  |
| Search Query Function | ✅ Pass | 3.505s |  |
| Ask Advice Function | ✅ Pass | 3.498s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.722s |  |
| Basic Context Memory Test | ✅ Pass | 1.496s |  |
| Function Argument Memory Test | ✅ Pass | 1.986s |  |
| Function Response Memory Test | ✅ Pass | 1.969s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.300s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.694s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 8.963s |  |
| Penetration Testing Methodology | ✅ Pass | 5.209s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.425s |  |
| SQL Injection Attack Type | ✅ Pass | 4.031s |  |
| Penetration Testing Framework | ✅ Pass | 4.437s |  |
| Web Application Security Scanner | ✅ Pass | 4.789s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.824s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 6.451s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.616s

---

### adviser (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.777s |  |
| Text Transform Uppercase | ✅ Pass | 2.771s |  |
| Count from 1 to 5 | ✅ Pass | 2.376s |  |
| Math Calculation | ✅ Pass | 2.776s |  |
| Basic Echo Function | ✅ Pass | 1.942s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.596s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.659s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.265s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.713s |  |
| Search Query Function | ✅ Pass | 3.245s |  |
| Ask Advice Function | ✅ Pass | 3.364s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.931s |  |
| Basic Context Memory Test | ✅ Pass | 2.392s |  |
| Function Argument Memory Test | ✅ Pass | 1.375s |  |
| Function Response Memory Test | ✅ Pass | 1.804s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.027s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.391s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.342s |  |
| Penetration Testing Methodology | ✅ Pass | 5.276s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.074s |  |
| SQL Injection Attack Type | ✅ Pass | 6.938s |  |
| Penetration Testing Framework | ✅ Pass | 3.638s |  |
| Web Application Security Scanner | ✅ Pass | 5.040s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.408s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.030s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.526s

---

### reflector (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.790s |  |
| Text Transform Uppercase | ✅ Pass | 2.143s |  |
| Count from 1 to 5 | ✅ Pass | 1.463s |  |
| Math Calculation | ✅ Pass | 1.056s |  |
| Basic Echo Function | ✅ Pass | 1.201s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.702s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.498s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.472s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.195s |  |
| Search Query Function | ✅ Pass | 1.676s |  |
| Ask Advice Function | ✅ Pass | 2.923s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.380s |  |
| Basic Context Memory Test | ✅ Pass | 1.696s |  |
| Function Argument Memory Test | ✅ Pass | 2.497s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.696s |  |
| Function Response Memory Test | ✅ Pass | 7.649s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.961s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.645s |  |
| Penetration Testing Methodology | ✅ Pass | 5.969s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.284s |  |
| SQL Injection Attack Type | ✅ Pass | 1.739s |  |
| Penetration Testing Framework | ✅ Pass | 5.243s |  |
| Web Application Security Scanner | ✅ Pass | 2.501s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.929s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.680s

---

### searcher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.431s |  |
| Text Transform Uppercase | ✅ Pass | 1.154s |  |
| Count from 1 to 5 | ✅ Pass | 2.054s |  |
| Math Calculation | ✅ Pass | 3.548s |  |
| Basic Echo Function | ✅ Pass | 1.683s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.444s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.506s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.796s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.811s |  |
| Search Query Function | ✅ Pass | 1.847s |  |
| Ask Advice Function | ✅ Pass | 1.502s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.090s |  |
| Basic Context Memory Test | ✅ Pass | 1.586s |  |
| Function Argument Memory Test | ✅ Pass | 1.900s |  |
| Function Response Memory Test | ✅ Pass | 2.537s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.362s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.734s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.827s |  |
| Penetration Testing Methodology | ✅ Pass | 5.147s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.173s |  |
| SQL Injection Attack Type | ✅ Pass | 2.149s |  |
| Penetration Testing Framework | ✅ Pass | 4.290s |  |
| Web Application Security Scanner | ✅ Pass | 3.945s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.547s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.461s

---

### enricher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.245s |  |
| Text Transform Uppercase | ✅ Pass | 0.852s |  |
| Count from 1 to 5 | ✅ Pass | 1.479s |  |
| Math Calculation | ✅ Pass | 1.016s |  |
| Basic Echo Function | ✅ Pass | 2.044s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.868s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.960s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.977s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.703s |  |
| Search Query Function | ✅ Pass | 2.571s |  |
| Ask Advice Function | ✅ Pass | 1.128s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.473s |  |
| Basic Context Memory Test | ✅ Pass | 1.040s |  |
| Function Argument Memory Test | ✅ Pass | 1.611s |  |
| Function Response Memory Test | ✅ Pass | 2.478s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.718s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.125s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.180s |  |
| Penetration Testing Methodology | ✅ Pass | 3.150s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.297s |  |
| SQL Injection Attack Type | ✅ Pass | 1.498s |  |
| Penetration Testing Framework | ✅ Pass | 2.390s |  |
| Web Application Security Scanner | ✅ Pass | 1.942s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.323s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.837s

---

### coder (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.118s |  |
| Text Transform Uppercase | ✅ Pass | 1.652s |  |
| Count from 1 to 5 | ✅ Pass | 2.969s |  |
| Math Calculation | ✅ Pass | 1.779s |  |
| Basic Echo Function | ✅ Pass | 2.285s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.591s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.736s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.716s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.990s |  |
| Search Query Function | ✅ Pass | 3.434s |  |
| Ask Advice Function | ✅ Pass | 2.233s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.917s |  |
| Basic Context Memory Test | ✅ Pass | 2.235s |  |
| Function Argument Memory Test | ✅ Pass | 1.149s |  |
| Function Response Memory Test | ✅ Pass | 1.222s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.756s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.355s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.077s |  |
| Penetration Testing Methodology | ✅ Pass | 7.451s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.827s |  |
| SQL Injection Attack Type | ✅ Pass | 2.380s |  |
| Penetration Testing Framework | ✅ Pass | 3.099s |  |
| Web Application Security Scanner | ✅ Pass | 3.172s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.942s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.880s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.079s

---

### installer (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.737s |  |
| Text Transform Uppercase | ✅ Pass | 5.016s |  |
| Count from 1 to 5 | ✅ Pass | 3.376s |  |
| Math Calculation | ✅ Pass | 2.094s |  |
| Basic Echo Function | ✅ Pass | 3.254s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.049s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.538s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.074s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.987s |  |
| Search Query Function | ✅ Pass | 1.803s |  |
| Ask Advice Function | ✅ Pass | 2.054s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.411s |  |
| Basic Context Memory Test | ✅ Pass | 1.712s |  |
| Function Argument Memory Test | ✅ Pass | 1.201s |  |
| Function Response Memory Test | ✅ Pass | 1.384s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.776s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.598s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.793s |  |
| Penetration Testing Methodology | ✅ Pass | 9.094s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.085s |  |
| SQL Injection Attack Type | ✅ Pass | 3.290s |  |
| Penetration Testing Framework | ✅ Pass | 3.626s |  |
| Web Application Security Scanner | ✅ Pass | 3.966s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.891s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.565s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.176s

---

### pentester (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.683s |  |
| Text Transform Uppercase | ✅ Pass | 1.718s |  |
| Count from 1 to 5 | ✅ Pass | 1.858s |  |
| Math Calculation | ✅ Pass | 1.838s |  |
| Basic Echo Function | ✅ Pass | 1.989s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.024s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.588s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.812s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.852s |  |
| Search Query Function | ✅ Pass | 2.001s |  |
| Ask Advice Function | ✅ Pass | 3.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.657s |  |
| Basic Context Memory Test | ✅ Pass | 2.018s |  |
| Function Argument Memory Test | ✅ Pass | 2.264s |  |
| Function Response Memory Test | ✅ Pass | 1.478s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.958s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.116s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.258s |  |
| Penetration Testing Methodology | ✅ Pass | 5.471s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.851s |  |
| SQL Injection Attack Type | ✅ Pass | 2.248s |  |
| Penetration Testing Framework | ✅ Pass | 2.710s |  |
| Web Application Security Scanner | ✅ Pass | 3.531s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.584s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.650s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.655s

---

