# LLM Agent Testing Report

Generated: Fri, 24 Jul 2026 23:46:47 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-haiku-4-5 | false | 24/24 (100.00%) | 1.498s |
| simple_json | claude-haiku-4-5 | false | 7/7 (100.00%) | 0.999s |
| primary_agent | claude-sonnet-5 | true | 25/25 (100.00%) | 2.390s |
| assistant | claude-sonnet-5 | true | 25/25 (100.00%) | 2.719s |
| generator | claude-opus-5 | true | 25/25 (100.00%) | 4.536s |
| refiner | claude-opus-5 | true | 24/25 (96.00%) | 4.263s |
| adviser | claude-sonnet-5 | true | 25/25 (100.00%) | 2.570s |
| reflector | claude-haiku-4-5 | true | 24/24 (100.00%) | 2.272s |
| searcher | claude-haiku-4-5 | true | 24/24 (100.00%) | 2.142s |
| enricher | claude-haiku-4-5 | true | 24/24 (100.00%) | 1.463s |
| coder | claude-sonnet-5 | true | 25/25 (100.00%) | 2.202s |
| installer | claude-sonnet-5 | true | 25/25 (100.00%) | 2.262s |
| pentester | claude-sonnet-5 | true | 25/25 (100.00%) | 2.235s |

**Total**: 302/303 (99.67%) successful tests
**Overall average latency**: 2.519s

## Detailed Results

### simple (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.670s |  |
| Text Transform Uppercase | ✅ Pass | 0.866s |  |
| Count from 1 to 5 | ✅ Pass | 0.872s |  |
| Math Calculation | ✅ Pass | 0.685s |  |
| Basic Echo Function | ✅ Pass | 1.009s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.834s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.833s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.958s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.905s |  |
| Search Query Function | ✅ Pass | 0.976s |  |
| Ask Advice Function | ✅ Pass | 1.084s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.997s |  |
| Basic Context Memory Test | ✅ Pass | 1.013s |  |
| Function Argument Memory Test | ✅ Pass | 1.898s |  |
| Function Response Memory Test | ✅ Pass | 0.797s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.457s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.976s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.844s |  |
| Penetration Testing Methodology | ✅ Pass | 2.914s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.607s |  |
| SQL Injection Attack Type | ✅ Pass | 0.880s |  |
| Penetration Testing Framework | ✅ Pass | 2.742s |  |
| Web Application Security Scanner | ✅ Pass | 1.704s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.414s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.498s

---

### simple_json (claude-haiku-4-5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.115s |  |
| Person Information JSON | ✅ Pass | 0.839s |  |
| Project Information JSON | ✅ Pass | 0.928s |  |
| User Profile JSON | ✅ Pass | 0.838s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.026s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.977s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 1.264s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 0.999s

---

### primary_agent (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.076s |  |
| Text Transform Uppercase | ✅ Pass | 1.379s |  |
| Count from 1 to 5 | ✅ Pass | 1.383s |  |
| Basic Echo Function | ✅ Pass | 1.863s |  |
| Math Calculation | ✅ Pass | 4.831s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.334s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.480s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.479s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.449s |  |
| Search Query Function | ✅ Pass | 1.596s |  |
| Ask Advice Function | ✅ Pass | 1.958s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.574s |  |
| Basic Context Memory Test | ✅ Pass | 1.403s |  |
| Function Argument Memory Test | ✅ Pass | 1.190s |  |
| Function Response Memory Test | ✅ Pass | 1.053s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.691s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.967s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.661s |  |
| Penetration Testing Methodology | ✅ Pass | 4.073s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.176s |  |
| SQL Injection Attack Type | ✅ Pass | 2.019s |  |
| Penetration Testing Framework | ✅ Pass | 2.147s |  |
| Web Application Security Scanner | ✅ Pass | 3.792s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.612s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.547s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.390s

---

### assistant (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.997s |  |
| Text Transform Uppercase | ✅ Pass | 1.458s |  |
| Count from 1 to 5 | ✅ Pass | 1.328s |  |
| Basic Echo Function | ✅ Pass | 1.600s |  |
| Math Calculation | ✅ Pass | 5.035s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.878s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.336s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.336s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.123s |  |
| Search Query Function | ✅ Pass | 1.911s |  |
| Ask Advice Function | ✅ Pass | 2.251s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.713s |  |
| Basic Context Memory Test | ✅ Pass | 1.290s |  |
| Function Argument Memory Test | ✅ Pass | 1.026s |  |
| Function Response Memory Test | ✅ Pass | 1.022s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.375s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.461s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.091s |  |
| Penetration Testing Methodology | ✅ Pass | 4.185s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.362s |  |
| SQL Injection Attack Type | ✅ Pass | 1.968s |  |
| Penetration Testing Framework | ✅ Pass | 3.571s |  |
| Web Application Security Scanner | ✅ Pass | 4.008s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.765s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.880s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.719s

---

### generator (claude-opus-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.084s |  |
| Text Transform Uppercase | ✅ Pass | 2.240s |  |
| Count from 1 to 5 | ✅ Pass | 2.427s |  |
| Math Calculation | ✅ Pass | 2.359s |  |
| Basic Echo Function | ✅ Pass | 2.638s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.284s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.295s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.563s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.638s |  |
| Search Query Function | ✅ Pass | 3.425s |  |
| Ask Advice Function | ✅ Pass | 2.907s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.381s |  |
| Function Argument Memory Test | ✅ Pass | 2.146s |  |
| Function Response Memory Test | ✅ Pass | 1.409s |  |
| Basic Context Memory Test | ✅ Pass | 6.291s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.189s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.801s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 8.585s |  |
| Penetration Testing Methodology | ✅ Pass | 8.149s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.748s |  |
| SQL Injection Attack Type | ✅ Pass | 9.585s |  |
| Penetration Testing Framework | ✅ Pass | 10.837s |  |
| Web Application Security Scanner | ✅ Pass | 8.504s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.451s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.447s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 4.536s

---

### refiner (claude-opus-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.608s |  |
| Text Transform Uppercase | ✅ Pass | 2.551s |  |
| Count from 1 to 5 | ✅ Pass | 2.286s |  |
| Math Calculation | ✅ Pass | 2.362s |  |
| Basic Echo Function | ✅ Pass | 3.507s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.228s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.692s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.511s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.949s |  |
| Search Query Function | ✅ Pass | 3.934s |  |
| Ask Advice Function | ✅ Pass | 2.935s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.757s |  |
| Basic Context Memory Test | ✅ Pass | 2.124s |  |
| Function Argument Memory Test | ✅ Pass | 1.640s |  |
| Function Response Memory Test | ✅ Pass | 1.611s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 3.711s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.274s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 6.303s |  |
| Penetration Testing Methodology | ✅ Pass | 7.976s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.994s |  |
| SQL Injection Attack Type | ✅ Pass | 7.579s |  |
| Penetration Testing Framework | ✅ Pass | 8.911s |  |
| Web Application Security Scanner | ✅ Pass | 7.870s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.696s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.550s |  |

**Summary**: 24/25 (96.00%) successful tests

**Average latency**: 4.263s

---

### adviser (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.224s |  |
| Text Transform Uppercase | ✅ Pass | 1.400s |  |
| Count from 1 to 5 | ✅ Pass | 2.183s |  |
| Math Calculation | ✅ Pass | 2.158s |  |
| Basic Echo Function | ✅ Pass | 1.957s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.292s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.239s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.491s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.402s |  |
| Search Query Function | ✅ Pass | 2.010s |  |
| Ask Advice Function | ✅ Pass | 1.863s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.358s |  |
| Basic Context Memory Test | ✅ Pass | 1.897s |  |
| Function Argument Memory Test | ✅ Pass | 0.977s |  |
| Function Response Memory Test | ✅ Pass | 0.951s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.593s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.715s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.305s |  |
| Penetration Testing Methodology | ✅ Pass | 5.507s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.458s |  |
| SQL Injection Attack Type | ✅ Pass | 2.115s |  |
| Penetration Testing Framework | ✅ Pass | 3.715s |  |
| Web Application Security Scanner | ✅ Pass | 3.801s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.578s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.047s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.570s

---

### reflector (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.280s |  |
| Text Transform Uppercase | ✅ Pass | 1.407s |  |
| Count from 1 to 5 | ✅ Pass | 1.184s |  |
| Math Calculation | ✅ Pass | 0.886s |  |
| Basic Echo Function | ✅ Pass | 1.644s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.988s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.454s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.403s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.157s |  |
| Search Query Function | ✅ Pass | 1.251s |  |
| Ask Advice Function | ✅ Pass | 3.550s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.370s |  |
| Basic Context Memory Test | ✅ Pass | 1.345s |  |
| Function Argument Memory Test | ✅ Pass | 1.228s |  |
| Function Response Memory Test | ✅ Pass | 1.481s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.061s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.439s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.653s |  |
| Penetration Testing Methodology | ✅ Pass | 4.865s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.008s |  |
| SQL Injection Attack Type | ✅ Pass | 1.822s |  |
| Penetration Testing Framework | ✅ Pass | 3.449s |  |
| Web Application Security Scanner | ✅ Pass | 3.100s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.496s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.272s

---

### searcher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.620s |  |
| Text Transform Uppercase | ✅ Pass | 1.068s |  |
| Count from 1 to 5 | ✅ Pass | 1.576s |  |
| Math Calculation | ✅ Pass | 0.973s |  |
| Basic Echo Function | ✅ Pass | 2.160s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.072s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.153s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.047s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.187s |  |
| Search Query Function | ✅ Pass | 1.153s |  |
| Ask Advice Function | ✅ Pass | 1.627s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.276s |  |
| Basic Context Memory Test | ✅ Pass | 1.234s |  |
| Function Argument Memory Test | ✅ Pass | 1.090s |  |
| Function Response Memory Test | ✅ Pass | 1.431s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.940s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.420s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.394s |  |
| Penetration Testing Methodology | ✅ Pass | 4.558s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.243s |  |
| SQL Injection Attack Type | ✅ Pass | 1.937s |  |
| Penetration Testing Framework | ✅ Pass | 5.154s |  |
| Web Application Security Scanner | ✅ Pass | 3.623s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.454s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.142s

---

### enricher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.800s |  |
| Text Transform Uppercase | ✅ Pass | 0.790s |  |
| Count from 1 to 5 | ✅ Pass | 0.795s |  |
| Math Calculation | ✅ Pass | 0.761s |  |
| Basic Echo Function | ✅ Pass | 1.136s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.809s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.733s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.065s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.107s |  |
| Search Query Function | ✅ Pass | 0.966s |  |
| Ask Advice Function | ✅ Pass | 1.276s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.667s |  |
| Basic Context Memory Test | ✅ Pass | 1.014s |  |
| Function Argument Memory Test | ✅ Pass | 0.797s |  |
| Function Response Memory Test | ✅ Pass | 0.980s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.422s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.858s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.384s |  |
| Penetration Testing Methodology | ✅ Pass | 3.624s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.636s |  |
| SQL Injection Attack Type | ✅ Pass | 0.815s |  |
| Penetration Testing Framework | ✅ Pass | 2.886s |  |
| Web Application Security Scanner | ✅ Pass | 2.159s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.612s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.463s

---

### coder (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.393s |  |
| Text Transform Uppercase | ✅ Pass | 1.398s |  |
| Count from 1 to 5 | ✅ Pass | 1.332s |  |
| Math Calculation | ✅ Pass | 1.512s |  |
| Basic Echo Function | ✅ Pass | 1.481s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.319s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.308s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.431s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.850s |  |
| Search Query Function | ✅ Pass | 1.632s |  |
| Ask Advice Function | ✅ Pass | 1.642s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.929s |  |
| Basic Context Memory Test | ✅ Pass | 1.591s |  |
| Function Argument Memory Test | ✅ Pass | 0.965s |  |
| Function Response Memory Test | ✅ Pass | 1.279s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.722s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.653s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.885s |  |
| Penetration Testing Methodology | ✅ Pass | 4.644s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.805s |  |
| SQL Injection Attack Type | ✅ Pass | 2.013s |  |
| Penetration Testing Framework | ✅ Pass | 2.157s |  |
| Web Application Security Scanner | ✅ Pass | 3.637s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.882s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.578s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.202s

---

### installer (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.345s |  |
| Text Transform Uppercase | ✅ Pass | 1.543s |  |
| Count from 1 to 5 | ✅ Pass | 1.277s |  |
| Math Calculation | ✅ Pass | 1.448s |  |
| Basic Echo Function | ✅ Pass | 1.583s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.604s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.313s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.353s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.583s |  |
| Search Query Function | ✅ Pass | 2.719s |  |
| Ask Advice Function | ✅ Pass | 3.472s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.484s |  |
| Basic Context Memory Test | ✅ Pass | 1.446s |  |
| Function Argument Memory Test | ✅ Pass | 0.890s |  |
| Function Response Memory Test | ✅ Pass | 0.991s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.601s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.259s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.577s |  |
| Penetration Testing Methodology | ✅ Pass | 4.132s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.232s |  |
| SQL Injection Attack Type | ✅ Pass | 1.887s |  |
| Penetration Testing Framework | ✅ Pass | 4.077s |  |
| Web Application Security Scanner | ✅ Pass | 3.495s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.649s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.569s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.262s

---

### pentester (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.415s |  |
| Text Transform Uppercase | ✅ Pass | 1.242s |  |
| Count from 1 to 5 | ✅ Pass | 1.300s |  |
| Math Calculation | ✅ Pass | 1.326s |  |
| Basic Echo Function | ✅ Pass | 1.560s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.787s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.271s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.406s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.642s |  |
| Search Query Function | ✅ Pass | 1.500s |  |
| Ask Advice Function | ✅ Pass | 1.649s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.479s |  |
| Basic Context Memory Test | ✅ Pass | 1.452s |  |
| Function Argument Memory Test | ✅ Pass | 0.970s |  |
| Function Response Memory Test | ✅ Pass | 1.053s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.122s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.730s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.055s |  |
| Penetration Testing Methodology | ✅ Pass | 5.257s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.053s |  |
| SQL Injection Attack Type | ✅ Pass | 2.083s |  |
| Penetration Testing Framework | ✅ Pass | 3.286s |  |
| Web Application Security Scanner | ✅ Pass | 3.034s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.035s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.158s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.235s

---

