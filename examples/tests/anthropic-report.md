# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 17:29:51 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-haiku-4-5 | false | 23/23 (100.00%) | 1.551s |
| simple_json | claude-haiku-4-5 | false | 7/7 (100.00%) | 1.066s |
| primary_agent | claude-sonnet-5 | true | 24/24 (100.00%) | 2.675s |
| assistant | claude-sonnet-5 | true | 24/24 (100.00%) | 2.445s |
| generator | claude-opus-4-8 | true | 24/24 (100.00%) | 3.121s |
| refiner | claude-opus-4-8 | true | 24/24 (100.00%) | 3.083s |
| adviser | claude-sonnet-5 | true | 24/24 (100.00%) | 2.793s |
| reflector | claude-haiku-4-5 | true | 23/23 (100.00%) | 2.148s |
| searcher | claude-haiku-4-5 | true | 23/23 (100.00%) | 2.070s |
| enricher | claude-haiku-4-5 | true | 23/23 (100.00%) | 1.460s |
| coder | claude-sonnet-5 | true | 24/24 (100.00%) | 2.304s |
| installer | claude-sonnet-5 | true | 24/24 (100.00%) | 2.395s |
| pentester | claude-sonnet-5 | true | 24/24 (100.00%) | 2.439s |

**Total**: 291/291 (100.00%) successful tests
**Overall average latency**: 2.350s

## Detailed Results

### simple (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.474s |  |
| Text Transform Uppercase | ✅ Pass | 0.869s |  |
| Count from 1 to 5 | ✅ Pass | 0.887s |  |
| Math Calculation | ✅ Pass | 0.886s |  |
| Basic Echo Function | ✅ Pass | 1.734s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.051s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.887s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.527s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.323s |  |
| Search Query Function | ✅ Pass | 1.087s |  |
| Ask Advice Function | ✅ Pass | 1.464s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.469s |  |
| Basic Context Memory Test | ✅ Pass | 0.737s |  |
| Function Argument Memory Test | ✅ Pass | 0.737s |  |
| Function Response Memory Test | ✅ Pass | 0.862s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.389s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.887s |  |
| Penetration Testing Methodology | ✅ Pass | 3.168s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.400s |  |
| SQL Injection Attack Type | ✅ Pass | 0.935s |  |
| Penetration Testing Framework | ✅ Pass | 3.051s |  |
| Web Application Security Scanner | ✅ Pass | 2.223s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.616s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.551s

---

### simple_json (claude-haiku-4-5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.128s |  |
| Person Information JSON | ✅ Pass | 1.458s |  |
| Project Information JSON | ✅ Pass | 0.904s |  |
| User Profile JSON | ✅ Pass | 0.946s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.915s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.874s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 1.237s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 1.066s

---

### primary_agent (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.860s |  |
| Text Transform Uppercase | ✅ Pass | 1.435s |  |
| Count from 1 to 5 | ✅ Pass | 1.509s |  |
| Math Calculation | ✅ Pass | 1.558s |  |
| Basic Echo Function | ✅ Pass | 1.816s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.689s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.235s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.530s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.493s |  |
| Search Query Function | ✅ Pass | 1.741s |  |
| Ask Advice Function | ✅ Pass | 2.370s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.956s |  |
| Basic Context Memory Test | ✅ Pass | 1.412s |  |
| Function Argument Memory Test | ✅ Pass | 0.930s |  |
| Function Response Memory Test | ✅ Pass | 4.467s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.373s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.987s |  |
| Penetration Testing Methodology | ✅ Pass | 5.430s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.277s |  |
| SQL Injection Attack Type | ✅ Pass | 2.384s |  |
| Penetration Testing Framework | ✅ Pass | 2.964s |  |
| Web Application Security Scanner | ✅ Pass | 4.640s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.127s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.011s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.675s

---

### assistant (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.899s |  |
| Text Transform Uppercase | ✅ Pass | 1.592s |  |
| Count from 1 to 5 | ✅ Pass | 1.450s |  |
| Math Calculation | ✅ Pass | 1.483s |  |
| Basic Echo Function | ✅ Pass | 1.556s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.494s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.422s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.429s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.622s |  |
| Search Query Function | ✅ Pass | 1.584s |  |
| Ask Advice Function | ✅ Pass | 2.134s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.733s |  |
| Basic Context Memory Test | ✅ Pass | 1.910s |  |
| Function Argument Memory Test | ✅ Pass | 1.032s |  |
| Function Response Memory Test | ✅ Pass | 1.119s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.608s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.041s |  |
| Penetration Testing Methodology | ✅ Pass | 4.464s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.970s |  |
| SQL Injection Attack Type | ✅ Pass | 2.227s |  |
| Penetration Testing Framework | ✅ Pass | 2.750s |  |
| Web Application Security Scanner | ✅ Pass | 3.515s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.828s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.802s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.445s

---

### generator (claude-opus-4-8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.778s |  |
| Text Transform Uppercase | ✅ Pass | 1.634s |  |
| Count from 1 to 5 | ✅ Pass | 1.615s |  |
| Math Calculation | ✅ Pass | 1.425s |  |
| Basic Echo Function | ✅ Pass | 2.980s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.311s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.371s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.444s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.648s |  |
| Search Query Function | ✅ Pass | 2.628s |  |
| Ask Advice Function | ✅ Pass | 3.379s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.276s |  |
| Basic Context Memory Test | ✅ Pass | 1.516s |  |
| Function Argument Memory Test | ✅ Pass | 1.641s |  |
| Function Response Memory Test | ✅ Pass | 1.507s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.282s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.624s |  |
| Penetration Testing Methodology | ✅ Pass | 5.729s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.258s |  |
| SQL Injection Attack Type | ✅ Pass | 3.770s |  |
| Penetration Testing Framework | ✅ Pass | 3.937s |  |
| Web Application Security Scanner | ✅ Pass | 4.144s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.585s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 5.407s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 3.121s

---

### refiner (claude-opus-4-8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.490s |  |
| Text Transform Uppercase | ✅ Pass | 1.768s |  |
| Count from 1 to 5 | ✅ Pass | 1.285s |  |
| Math Calculation | ✅ Pass | 1.566s |  |
| Basic Echo Function | ✅ Pass | 2.123s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.846s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.477s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 7.217s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.953s |  |
| Search Query Function | ✅ Pass | 1.819s |  |
| Ask Advice Function | ✅ Pass | 2.126s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.757s |  |
| Basic Context Memory Test | ✅ Pass | 1.782s |  |
| Function Argument Memory Test | ✅ Pass | 1.609s |  |
| Function Response Memory Test | ✅ Pass | 1.416s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.534s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.721s |  |
| Penetration Testing Methodology | ✅ Pass | 7.345s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.771s |  |
| SQL Injection Attack Type | ✅ Pass | 3.520s |  |
| Penetration Testing Framework | ✅ Pass | 4.607s |  |
| Web Application Security Scanner | ✅ Pass | 3.308s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.268s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.670s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 3.083s

---

### adviser (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.247s |  |
| Text Transform Uppercase | ✅ Pass | 2.731s |  |
| Count from 1 to 5 | ✅ Pass | 3.645s |  |
| Math Calculation | ✅ Pass | 3.076s |  |
| Basic Echo Function | ✅ Pass | 1.611s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.498s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.317s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.079s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.630s |  |
| Search Query Function | ✅ Pass | 1.994s |  |
| Ask Advice Function | ✅ Pass | 1.866s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.537s |  |
| Basic Context Memory Test | ✅ Pass | 1.896s |  |
| Function Argument Memory Test | ✅ Pass | 2.105s |  |
| Function Response Memory Test | ✅ Pass | 0.987s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.039s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.355s |  |
| Penetration Testing Methodology | ✅ Pass | 5.592s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.804s |  |
| SQL Injection Attack Type | ✅ Pass | 2.132s |  |
| Penetration Testing Framework | ✅ Pass | 3.866s |  |
| Web Application Security Scanner | ✅ Pass | 3.457s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.638s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 3.929s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.793s

---

### reflector (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.383s |  |
| Text Transform Uppercase | ✅ Pass | 2.238s |  |
| Count from 1 to 5 | ✅ Pass | 1.640s |  |
| Math Calculation | ✅ Pass | 1.539s |  |
| Basic Echo Function | ✅ Pass | 1.184s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.151s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.121s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.233s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.245s |  |
| Search Query Function | ✅ Pass | 1.826s |  |
| Ask Advice Function | ✅ Pass | 2.019s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.377s |  |
| Basic Context Memory Test | ✅ Pass | 1.257s |  |
| Function Argument Memory Test | ✅ Pass | 1.366s |  |
| Function Response Memory Test | ✅ Pass | 1.849s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.053s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.621s |  |
| Penetration Testing Methodology | ✅ Pass | 5.110s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.665s |  |
| SQL Injection Attack Type | ✅ Pass | 2.887s |  |
| Penetration Testing Framework | ✅ Pass | 4.031s |  |
| Web Application Security Scanner | ✅ Pass | 2.893s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.711s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.148s

---

### searcher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.377s |  |
| Text Transform Uppercase | ✅ Pass | 1.572s |  |
| Count from 1 to 5 | ✅ Pass | 1.592s |  |
| Math Calculation | ✅ Pass | 1.708s |  |
| Basic Echo Function | ✅ Pass | 1.753s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.674s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.905s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.296s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.324s |  |
| Search Query Function | ✅ Pass | 2.001s |  |
| Ask Advice Function | ✅ Pass | 1.786s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.155s |  |
| Basic Context Memory Test | ✅ Pass | 1.404s |  |
| Function Argument Memory Test | ✅ Pass | 1.086s |  |
| Function Response Memory Test | ✅ Pass | 2.155s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.961s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.381s |  |
| Penetration Testing Methodology | ✅ Pass | 4.687s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.199s |  |
| SQL Injection Attack Type | ✅ Pass | 1.936s |  |
| Penetration Testing Framework | ✅ Pass | 4.645s |  |
| Web Application Security Scanner | ✅ Pass | 2.497s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.502s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.070s

---

### enricher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.385s |  |
| Text Transform Uppercase | ✅ Pass | 0.802s |  |
| Count from 1 to 5 | ✅ Pass | 0.872s |  |
| Math Calculation | ✅ Pass | 0.733s |  |
| Basic Echo Function | ✅ Pass | 1.510s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.317s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.655s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.929s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.797s |  |
| Search Query Function | ✅ Pass | 1.199s |  |
| Ask Advice Function | ✅ Pass | 1.205s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.019s |  |
| Basic Context Memory Test | ✅ Pass | 0.939s |  |
| Function Argument Memory Test | ✅ Pass | 1.526s |  |
| Function Response Memory Test | ✅ Pass | 0.811s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.631s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.880s |  |
| Penetration Testing Methodology | ✅ Pass | 3.295s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.795s |  |
| SQL Injection Attack Type | ✅ Pass | 0.937s |  |
| Penetration Testing Framework | ✅ Pass | 3.289s |  |
| Web Application Security Scanner | ✅ Pass | 2.153s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.890s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.460s

---

### coder (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.668s |  |
| Text Transform Uppercase | ✅ Pass | 1.633s |  |
| Count from 1 to 5 | ✅ Pass | 2.121s |  |
| Math Calculation | ✅ Pass | 1.835s |  |
| Basic Echo Function | ✅ Pass | 1.605s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.676s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.276s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.546s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.664s |  |
| Search Query Function | ✅ Pass | 1.798s |  |
| Ask Advice Function | ✅ Pass | 2.148s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.518s |  |
| Basic Context Memory Test | ✅ Pass | 1.983s |  |
| Function Argument Memory Test | ✅ Pass | 1.030s |  |
| Function Response Memory Test | ✅ Pass | 1.263s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.947s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.005s |  |
| Penetration Testing Methodology | ✅ Pass | 4.340s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.774s |  |
| SQL Injection Attack Type | ✅ Pass | 2.061s |  |
| Penetration Testing Framework | ✅ Pass | 3.558s |  |
| Web Application Security Scanner | ✅ Pass | 3.701s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.698s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.431s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.304s

---

### installer (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.433s |  |
| Text Transform Uppercase | ✅ Pass | 1.551s |  |
| Count from 1 to 5 | ✅ Pass | 1.496s |  |
| Math Calculation | ✅ Pass | 1.473s |  |
| Basic Echo Function | ✅ Pass | 2.275s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.716s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.359s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.503s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.619s |  |
| Search Query Function | ✅ Pass | 1.760s |  |
| Ask Advice Function | ✅ Pass | 1.729s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.610s |  |
| Basic Context Memory Test | ✅ Pass | 1.589s |  |
| Function Argument Memory Test | ✅ Pass | 0.962s |  |
| Function Response Memory Test | ✅ Pass | 1.103s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.898s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.935s |  |
| Penetration Testing Methodology | ✅ Pass | 4.798s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.985s |  |
| SQL Injection Attack Type | ✅ Pass | 2.534s |  |
| Penetration Testing Framework | ✅ Pass | 2.663s |  |
| Web Application Security Scanner | ✅ Pass | 4.330s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.633s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.510s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.395s

---

### pentester (claude-sonnet-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.701s |  |
| Text Transform Uppercase | ✅ Pass | 1.470s |  |
| Count from 1 to 5 | ✅ Pass | 1.607s |  |
| Math Calculation | ✅ Pass | 1.496s |  |
| Basic Echo Function | ✅ Pass | 1.715s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.927s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.357s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.543s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.735s |  |
| Search Query Function | ✅ Pass | 2.197s |  |
| Ask Advice Function | ✅ Pass | 2.169s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.399s |  |
| Basic Context Memory Test | ✅ Pass | 2.068s |  |
| Function Argument Memory Test | ✅ Pass | 1.980s |  |
| Function Response Memory Test | ✅ Pass | 1.073s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.179s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.145s |  |
| Penetration Testing Methodology | ✅ Pass | 3.816s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.313s |  |
| SQL Injection Attack Type | ✅ Pass | 2.516s |  |
| Penetration Testing Framework | ✅ Pass | 4.049s |  |
| Web Application Security Scanner | ✅ Pass | 3.485s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.157s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Adaptive Thinking Produces Reasoning | adaptive_thinking | ✅ Pass | 2.420s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.439s

---

