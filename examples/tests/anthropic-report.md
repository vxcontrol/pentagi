# LLM Agent Testing Report

Generated: Wed, 29 Jul 2026 14:00:41 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-haiku-4-5 | false | 24/24 (100.00%) | 1.708s |
| simple_json | claude-haiku-4-5 | false | 7/7 (100.00%) | 1.528s |
| primary_agent | claude-sonnet-5 | true | 25/25 (100.00%) | 2.673s |
| assistant | claude-sonnet-5 | true | 25/25 (100.00%) | 2.812s |
| generator | claude-opus-4-8 | true | 25/25 (100.00%) | 3.058s |
| refiner | claude-opus-4-8 | true | 24/25 (96.00%) | 3.235s |
| adviser | claude-sonnet-5 | true | 25/25 (100.00%) | 3.067s |
| reflector | claude-haiku-4-5 | true | 24/24 (100.00%) | 2.253s |
| searcher | claude-haiku-4-5 | true | 24/24 (100.00%) | 2.188s |
| enricher | claude-haiku-4-5 | true | 24/24 (100.00%) | 1.718s |
| coder | claude-sonnet-5 | true | 25/25 (100.00%) | 3.067s |
| installer | claude-sonnet-5 | true | 25/25 (100.00%) | 3.496s |
| pentester | claude-sonnet-5 | true | 25/25 (100.00%) | 2.430s |

**Total**: 302/303 (99.67%) successful tests
**Overall average latency**: 2.625s

## Detailed Results

### simple (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.608s |  |
| Text Transform Uppercase | ✅ Pass | 1.003s |  |
| Count from 1 to 5 | ✅ Pass | 0.928s |  |
| Math Calculation | ✅ Pass | 1.267s |  |
| Basic Echo Function | ✅ Pass | 1.090s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.412s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.998s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.989s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.020s |  |
| Search Query Function | ✅ Pass | 1.148s |  |
| Ask Advice Function | ✅ Pass | 1.257s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.356s |  |
| Basic Context Memory Test | ✅ Pass | 1.513s |  |
| Function Argument Memory Test | ✅ Pass | 1.061s |  |
| Function Response Memory Test | ✅ Pass | 1.098s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.457s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.474s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.397s |  |
| Penetration Testing Methodology | ✅ Pass | 3.010s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.530s |  |
| SQL Injection Attack Type | ✅ Pass | 1.120s |  |
| Penetration Testing Framework | ✅ Pass | 3.330s |  |
| Web Application Security Scanner | ✅ Pass | 2.469s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.440s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.708s

---

### simple_json (claude-haiku-4-5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.631s |  |
| Person Information JSON | ✅ Pass | 1.031s |  |
| Project Information JSON | ✅ Pass | 0.928s |  |
| User Profile JSON | ✅ Pass | 0.934s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.330s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.592s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 3.248s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 1.528s

---

### primary_agent (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.567s |  |
| Text Transform Uppercase | ✅ Pass | 1.687s |  |
| Count from 1 to 5 | ✅ Pass | 1.642s |  |
| Math Calculation | ✅ Pass | 2.230s |  |
| Basic Echo Function | ✅ Pass | 2.197s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.370s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.590s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.226s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.771s |  |
| Search Query Function | ✅ Pass | 2.064s |  |
| Ask Advice Function | ✅ Pass | 2.092s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.684s |  |
| Basic Context Memory Test | ✅ Pass | 1.543s |  |
| Function Argument Memory Test | ✅ Pass | 1.192s |  |
| Function Response Memory Test | ✅ Pass | 1.807s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.637s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.325s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.663s |  |
| Penetration Testing Methodology | ✅ Pass | 3.865s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.838s |  |
| SQL Injection Attack Type | ✅ Pass | 2.388s |  |
| Penetration Testing Framework | ✅ Pass | 3.208s |  |
| Web Application Security Scanner | ✅ Pass | 3.788s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.807s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.630s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.673s

---

### assistant (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.793s |  |
| Text Transform Uppercase | ✅ Pass | 1.541s |  |
| Count from 1 to 5 | ✅ Pass | 1.651s |  |
| Math Calculation | ✅ Pass | 1.648s |  |
| Basic Echo Function | ✅ Pass | 1.716s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.867s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.659s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.925s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.277s |  |
| Search Query Function | ✅ Pass | 1.665s |  |
| Ask Advice Function | ✅ Pass | 2.003s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.371s |  |
| Basic Context Memory Test | ✅ Pass | 1.536s |  |
| Function Argument Memory Test | ✅ Pass | 1.312s |  |
| Function Response Memory Test | ✅ Pass | 1.041s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.453s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.828s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.482s |  |
| Penetration Testing Methodology | ✅ Pass | 6.465s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.883s |  |
| SQL Injection Attack Type | ✅ Pass | 2.771s |  |
| Penetration Testing Framework | ✅ Pass | 2.675s |  |
| Web Application Security Scanner | ✅ Pass | 4.184s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.722s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.823s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.812s

---

### generator (claude-opus-4-8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.910s |  |
| Text Transform Uppercase | ✅ Pass | 2.127s |  |
| Count from 1 to 5 | ✅ Pass | 1.503s |  |
| Math Calculation | ✅ Pass | 1.815s |  |
| Basic Echo Function | ✅ Pass | 2.475s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.242s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.196s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.997s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.811s |  |
| Search Query Function | ✅ Pass | 2.041s |  |
| Ask Advice Function | ✅ Pass | 3.837s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.749s |  |
| Basic Context Memory Test | ✅ Pass | 1.645s |  |
| Function Argument Memory Test | ✅ Pass | 1.502s |  |
| Function Response Memory Test | ✅ Pass | 1.433s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.961s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.287s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.659s |  |
| Penetration Testing Methodology | ✅ Pass | 6.504s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.918s |  |
| SQL Injection Attack Type | ✅ Pass | 2.417s |  |
| Penetration Testing Framework | ✅ Pass | 4.413s |  |
| Web Application Security Scanner | ✅ Pass | 3.852s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.304s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.835s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.058s

---

### refiner (claude-opus-4-8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.830s |  |
| Text Transform Uppercase | ✅ Pass | 1.603s |  |
| Count from 1 to 5 | ✅ Pass | 1.467s |  |
| Math Calculation | ✅ Pass | 1.599s |  |
| Basic Echo Function | ✅ Pass | 1.658s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.658s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.287s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.628s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 6.073s |  |
| Search Query Function | ✅ Pass | 4.042s |  |
| Ask Advice Function | ✅ Pass | 3.766s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.665s |  |
| Basic Context Memory Test | ✅ Pass | 1.894s |  |
| Function Argument Memory Test | ✅ Pass | 1.678s |  |
| Function Response Memory Test | ✅ Pass | 1.198s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 3.520s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.903s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.112s |  |
| Penetration Testing Methodology | ✅ Pass | 6.030s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.960s |  |
| SQL Injection Attack Type | ✅ Pass | 3.440s |  |
| Penetration Testing Framework | ✅ Pass | 3.542s |  |
| Web Application Security Scanner | ✅ Pass | 2.965s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.259s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 4.077s |  |

**Summary**: 24/25 (96.00%) successful tests

**Average latency**: 3.235s

---

### adviser (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.897s |  |
| Text Transform Uppercase | ✅ Pass | 2.011s |  |
| Count from 1 to 5 | ✅ Pass | 2.705s |  |
| Math Calculation | ✅ Pass | 2.534s |  |
| Basic Echo Function | ✅ Pass | 1.696s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.699s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.254s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.686s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.852s |  |
| Search Query Function | ✅ Pass | 2.132s |  |
| Ask Advice Function | ✅ Pass | 2.942s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.694s |  |
| Basic Context Memory Test | ✅ Pass | 1.605s |  |
| Function Argument Memory Test | ✅ Pass | 1.717s |  |
| Function Response Memory Test | ✅ Pass | 0.979s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.845s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.779s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 6.310s |  |
| Penetration Testing Methodology | ✅ Pass | 5.578s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.506s |  |
| SQL Injection Attack Type | ✅ Pass | 2.548s |  |
| Penetration Testing Framework | ✅ Pass | 4.756s |  |
| Web Application Security Scanner | ✅ Pass | 4.049s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.857s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 4.037s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.067s

---

### reflector (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.919s |  |
| Text Transform Uppercase | ✅ Pass | 1.397s |  |
| Count from 1 to 5 | ✅ Pass | 1.177s |  |
| Math Calculation | ✅ Pass | 1.751s |  |
| Basic Echo Function | ✅ Pass | 1.306s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.320s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.229s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.525s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.878s |  |
| Search Query Function | ✅ Pass | 1.282s |  |
| Ask Advice Function | ✅ Pass | 1.473s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.361s |  |
| Basic Context Memory Test | ✅ Pass | 1.223s |  |
| Function Argument Memory Test | ✅ Pass | 1.466s |  |
| Function Response Memory Test | ✅ Pass | 1.630s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.174s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.834s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.444s |  |
| Penetration Testing Methodology | ✅ Pass | 5.340s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.531s |  |
| SQL Injection Attack Type | ✅ Pass | 1.868s |  |
| Penetration Testing Framework | ✅ Pass | 4.969s |  |
| Web Application Security Scanner | ✅ Pass | 3.206s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.763s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.253s

---

### searcher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.774s |  |
| Text Transform Uppercase | ✅ Pass | 1.140s |  |
| Count from 1 to 5 | ✅ Pass | 1.315s |  |
| Math Calculation | ✅ Pass | 0.980s |  |
| Basic Echo Function | ✅ Pass | 1.426s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.246s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.481s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.849s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.304s |  |
| Search Query Function | ✅ Pass | 1.269s |  |
| Ask Advice Function | ✅ Pass | 1.551s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.261s |  |
| Basic Context Memory Test | ✅ Pass | 1.460s |  |
| Function Argument Memory Test | ✅ Pass | 1.449s |  |
| Function Response Memory Test | ✅ Pass | 1.714s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.895s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.605s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.194s |  |
| Penetration Testing Methodology | ✅ Pass | 5.508s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.114s |  |
| SQL Injection Attack Type | ✅ Pass | 1.793s |  |
| Penetration Testing Framework | ✅ Pass | 3.345s |  |
| Web Application Security Scanner | ✅ Pass | 4.184s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.640s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.188s

---

### enricher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.102s |  |
| Text Transform Uppercase | ✅ Pass | 0.933s |  |
| Count from 1 to 5 | ✅ Pass | 0.931s |  |
| Math Calculation | ✅ Pass | 1.224s |  |
| Basic Echo Function | ✅ Pass | 2.447s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.928s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.551s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.160s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.229s |  |
| Search Query Function | ✅ Pass | 1.153s |  |
| Ask Advice Function | ✅ Pass | 1.211s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.123s |  |
| Basic Context Memory Test | ✅ Pass | 1.027s |  |
| Function Argument Memory Test | ✅ Pass | 0.985s |  |
| Function Response Memory Test | ✅ Pass | 2.358s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.457s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.022s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.237s |  |
| Penetration Testing Methodology | ✅ Pass | 3.578s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.016s |  |
| SQL Injection Attack Type | ✅ Pass | 1.091s |  |
| Penetration Testing Framework | ✅ Pass | 2.961s |  |
| Web Application Security Scanner | ✅ Pass | 2.185s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.316s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.718s

---

### coder (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.075s |  |
| Text Transform Uppercase | ✅ Pass | 1.502s |  |
| Count from 1 to 5 | ✅ Pass | 1.886s |  |
| Math Calculation | ✅ Pass | 1.473s |  |
| Basic Echo Function | ✅ Pass | 1.718s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.899s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.529s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.765s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.904s |  |
| Search Query Function | ✅ Pass | 5.628s |  |
| Ask Advice Function | ✅ Pass | 2.626s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.451s |  |
| Basic Context Memory Test | ✅ Pass | 1.596s |  |
| Function Argument Memory Test | ✅ Pass | 0.997s |  |
| Function Response Memory Test | ✅ Pass | 1.117s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.988s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.574s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 10.725s |  |
| Penetration Testing Methodology | ✅ Pass | 4.963s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.065s |  |
| SQL Injection Attack Type | ✅ Pass | 7.389s |  |
| Penetration Testing Framework | ✅ Pass | 2.516s |  |
| Web Application Security Scanner | ✅ Pass | 4.771s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.395s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.121s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.067s

---

### installer (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.495s |  |
| Text Transform Uppercase | ✅ Pass | 1.566s |  |
| Count from 1 to 5 | ✅ Pass | 2.056s |  |
| Math Calculation | ✅ Pass | 3.841s |  |
| Basic Echo Function | ✅ Pass | 1.786s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.630s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.931s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.160s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.833s |  |
| Search Query Function | ✅ Pass | 2.009s |  |
| Ask Advice Function | ✅ Pass | 2.120s |  |
| Basic Context Memory Test | ✅ Pass | 1.881s |  |
| Function Argument Memory Test | ✅ Pass | 1.545s |  |
| Function Response Memory Test | ✅ Pass | 1.002s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 11.262s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.841s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.993s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.164s |  |
| Penetration Testing Methodology | ✅ Pass | 4.914s |  |
| SQL Injection Attack Type | ✅ Pass | 2.372s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.994s |  |
| Penetration Testing Framework | ✅ Pass | 3.002s |  |
| Web Application Security Scanner | ✅ Pass | 3.494s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.068s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.426s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.496s

---

### pentester (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.004s |  |
| Text Transform Uppercase | ✅ Pass | 2.030s |  |
| Count from 1 to 5 | ✅ Pass | 1.959s |  |
| Math Calculation | ✅ Pass | 1.660s |  |
| Basic Echo Function | ✅ Pass | 1.679s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.560s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.350s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.687s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.838s |  |
| Search Query Function | ✅ Pass | 1.699s |  |
| Ask Advice Function | ✅ Pass | 2.327s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.243s |  |
| Basic Context Memory Test | ✅ Pass | 1.626s |  |
| Function Argument Memory Test | ✅ Pass | 1.033s |  |
| Function Response Memory Test | ✅ Pass | 1.037s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.648s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.189s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.459s |  |
| Penetration Testing Methodology | ✅ Pass | 4.957s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.520s |  |
| SQL Injection Attack Type | ✅ Pass | 2.426s |  |
| Penetration Testing Framework | ✅ Pass | 2.777s |  |
| Web Application Security Scanner | ✅ Pass | 3.080s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.911s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.051s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 2.430s

---

