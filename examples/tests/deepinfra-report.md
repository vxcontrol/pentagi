# LLM Agent Testing Report

Generated: Tue, 30 Sep 2025 19:10:56 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3-Next-80B-A3B-Instruct | false | 23/23 (100.00%) | 1.284s |
| simple_json | Qwen/Qwen3-Next-80B-A3B-Instruct | false | 5/5 (100.00%) | 1.261s |
| primary_agent | moonshotai/Kimi-K2-Instruct-0905 | false | 22/23 (95.65%) | 1.406s |
| assistant | moonshotai/Kimi-K2-Instruct-0905 | true | 21/23 (91.30%) | 1.397s |
| generator | google/gemini-2.5-pro | true | 22/23 (95.65%) | 7.349s |
| refiner | deepseek-ai/DeepSeek-R1-0528-Turbo | true | 22/23 (95.65%) | 4.424s |
| adviser | google/gemini-2.5-pro | true | 23/23 (100.00%) | 6.986s |
| reflector | Qwen/Qwen3-Next-80B-A3B-Instruct | true | 23/23 (100.00%) | 1.277s |
| searcher | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 6.780s |
| enricher | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 6.705s |
| coder | anthropic/claude-4-sonnet | true | 23/23 (100.00%) | 2.953s |
| installer | google/gemini-2.5-flash | true | 23/23 (100.00%) | 2.703s |
| pentester | moonshotai/Kimi-K2-Instruct-0905 | true | 22/23 (95.65%) | 1.303s |

**Total**: 275/281 (97.86%) successful tests
**Overall average latency**: 3.670s

## Detailed Results

### simple (Qwen/Qwen3-Next-80B-A3B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.578s |  |
| Text Transform Uppercase | ✅ Pass | 2.580s |  |
| Count from 1 to 5 | ✅ Pass | 0.964s |  |
| Math Calculation | ✅ Pass | 0.900s |  |
| Basic Echo Function | ✅ Pass | 1.061s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.022s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.949s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.736s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.425s |  |
| Search Query Function | ✅ Pass | 1.348s |  |
| Ask Advice Function | ✅ Pass | 1.016s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.939s |  |
| Basic Context Memory Test | ✅ Pass | 1.248s |  |
| Function Argument Memory Test | ✅ Pass | 1.044s |  |
| Function Response Memory Test | ✅ Pass | 0.887s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.120s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.931s |  |
| Penetration Testing Methodology | ✅ Pass | 1.612s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.659s |  |
| SQL Injection Attack Type | ✅ Pass | 0.930s |  |
| Penetration Testing Framework | ✅ Pass | 1.154s |  |
| Web Application Security Scanner | ✅ Pass | 1.297s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.111s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.284s

---

### simple_json (Qwen/Qwen3-Next-80B-A3B-Instruct)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.469s |  |
| Person Information JSON | ✅ Pass | 1.179s |  |
| User Profile JSON | ✅ Pass | 1.124s |  |
| Project Information JSON | ✅ Pass | 1.246s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.283s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.261s

---

### primary_agent (moonshotai/Kimi-K2-Instruct-0905)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.603s |  |
| Text Transform Uppercase | ✅ Pass | 2.651s |  |
| Count from 1 to 5 | ✅ Pass | 0.755s |  |
| Math Calculation | ✅ Pass | 0.793s |  |
| Basic Echo Function | ✅ Pass | 1.418s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.783s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.653s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.644s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.157s |  |
| Search Query Function | ✅ Pass | 1.301s |  |
| Ask Advice Function | ✅ Pass | 1.400s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.985s |  |
| Basic Context Memory Test | ✅ Pass | 1.052s |  |
| Function Argument Memory Test | ✅ Pass | 1.118s |  |
| Function Response Memory Test | ✅ Pass | 0.731s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 2.041s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.632s |  |
| Penetration Testing Methodology | ✅ Pass | 1.588s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.885s |  |
| SQL Injection Attack Type | ✅ Pass | 0.796s |  |
| Penetration Testing Framework | ✅ Pass | 4.317s |  |
| Web Application Security Scanner | ✅ Pass | 0.679s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.336s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.406s

---

### assistant (moonshotai/Kimi-K2-Instruct-0905)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.648s |  |
| Text Transform Uppercase | ✅ Pass | 2.672s |  |
| Count from 1 to 5 | ✅ Pass | 0.759s |  |
| Math Calculation | ✅ Pass | 0.772s |  |
| Basic Echo Function | ❌ Fail | 1.552s | no tool calls found, expected at least 1 |
| Streaming Simple Math Streaming | ✅ Pass | 0.702s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.661s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.752s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.377s |  |
| Search Query Function | ✅ Pass | 2.002s |  |
| Ask Advice Function | ✅ Pass | 1.351s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.864s |  |
| Basic Context Memory Test | ✅ Pass | 1.075s |  |
| Function Argument Memory Test | ✅ Pass | 0.667s |  |
| Function Response Memory Test | ✅ Pass | 0.603s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 2.146s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.626s |  |
| Penetration Testing Methodology | ✅ Pass | 1.616s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.152s |  |
| SQL Injection Attack Type | ✅ Pass | 0.886s |  |
| Penetration Testing Framework | ✅ Pass | 3.288s |  |
| Web Application Security Scanner | ✅ Pass | 0.714s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.229s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 1.397s

---

### generator (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.339s |  |
| Text Transform Uppercase | ✅ Pass | 8.288s |  |
| Math Calculation | ✅ Pass | 3.081s |  |
| Count from 1 to 5 | ✅ Pass | 5.671s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.258s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.547s |  |
| Basic Echo Function | ✅ Pass | 8.848s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.689s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 4.440s |  |
| JSON Response Function | ✅ Pass | 8.128s |  |
| Ask Advice Function | ✅ Pass | 4.194s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.103s |  |
| Function Argument Memory Test | ✅ Pass | 4.669s |  |
| Basic Context Memory Test | ✅ Pass | 7.525s |  |
| Function Response Memory Test | ✅ Pass | 7.255s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.236s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.003s |  |
| Penetration Testing Methodology | ✅ Pass | 10.267s |  |
| SQL Injection Attack Type | ✅ Pass | 6.586s |  |
| Vulnerability Assessment Tools | ❌ Fail | 21.490s | expected text 'network' not found |
| Penetration Testing Tool Selection | ✅ Pass | 4.211s |  |
| Penetration Testing Framework | ✅ Pass | 18.186s |  |
| Web Application Security Scanner | ✅ Pass | 15.008s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 7.349s

---

### refiner (deepseek-ai/DeepSeek-R1-0528-Turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 3.027s |  |
| Count from 1 to 5 | ✅ Pass | 1.407s |  |
| Simple Math | ✅ Pass | 5.472s |  |
| Math Calculation | ✅ Pass | 2.814s |  |
| Basic Echo Function | ✅ Pass | 2.816s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.297s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.343s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.435s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.203s |  |
| Search Query Function | ✅ Pass | 3.924s |  |
| Ask Advice Function | ✅ Pass | 2.326s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 3.341s | no tool calls found, expected at least 1 |
| Basic Context Memory Test | ✅ Pass | 4.284s |  |
| Function Argument Memory Test | ✅ Pass | 2.489s |  |
| Function Response Memory Test | ✅ Pass | 1.821s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.045s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.472s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.717s |  |
| Penetration Testing Methodology | ✅ Pass | 8.188s |  |
| SQL Injection Attack Type | ✅ Pass | 6.380s |  |
| Penetration Testing Framework | ✅ Pass | 10.979s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.450s |  |
| Web Application Security Scanner | ✅ Pass | 10.505s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.424s

---

### adviser (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.136s |  |
| Text Transform Uppercase | ✅ Pass | 5.091s |  |
| Count from 1 to 5 | ✅ Pass | 5.334s |  |
| Math Calculation | ✅ Pass | 3.554s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.277s |  |
| Basic Echo Function | ✅ Pass | 5.468s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.383s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.941s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 4.618s |  |
| JSON Response Function | ✅ Pass | 8.534s |  |
| Ask Advice Function | ✅ Pass | 4.124s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.168s |  |
| Basic Context Memory Test | ✅ Pass | 5.123s |  |
| Function Argument Memory Test | ✅ Pass | 3.921s |  |
| Function Response Memory Test | ✅ Pass | 6.008s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.767s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.001s |  |
| Penetration Testing Methodology | ✅ Pass | 11.466s |  |
| SQL Injection Attack Type | ✅ Pass | 8.174s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.468s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.422s |  |
| Web Application Security Scanner | ✅ Pass | 15.610s |  |
| Penetration Testing Framework | ✅ Pass | 20.072s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.986s

---

### reflector (Qwen/Qwen3-Next-80B-A3B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.579s |  |
| Text Transform Uppercase | ✅ Pass | 1.261s |  |
| Count from 1 to 5 | ✅ Pass | 0.888s |  |
| Math Calculation | ✅ Pass | 0.900s |  |
| Basic Echo Function | ✅ Pass | 0.943s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.933s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.168s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.156s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.828s |  |
| Search Query Function | ✅ Pass | 1.027s |  |
| Ask Advice Function | ✅ Pass | 0.974s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.438s |  |
| Basic Context Memory Test | ✅ Pass | 0.978s |  |
| Function Argument Memory Test | ✅ Pass | 0.914s |  |
| Function Response Memory Test | ✅ Pass | 0.950s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.307s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.194s |  |
| Penetration Testing Methodology | ✅ Pass | 1.691s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.906s |  |
| SQL Injection Attack Type | ✅ Pass | 1.504s |  |
| Penetration Testing Framework | ✅ Pass | 1.183s |  |
| Web Application Security Scanner | ✅ Pass | 1.622s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.013s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.277s

---

### searcher (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.632s |  |
| Text Transform Uppercase | ✅ Pass | 5.010s |  |
| Count from 1 to 5 | ✅ Pass | 4.092s |  |
| Basic Echo Function | ✅ Pass | 3.654s |  |
| Math Calculation | ✅ Pass | 6.467s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.738s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.966s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.840s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.198s |  |
| Search Query Function | ✅ Pass | 3.664s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.344s |  |
| Ask Advice Function | ✅ Pass | 6.628s |  |
| Function Argument Memory Test | ✅ Pass | 4.112s |  |
| Basic Context Memory Test | ✅ Pass | 7.697s |  |
| Function Response Memory Test | ✅ Pass | 4.331s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.918s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.309s |  |
| Penetration Testing Methodology | ✅ Pass | 12.251s |  |
| SQL Injection Attack Type | ✅ Pass | 9.998s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.638s |  |
| Penetration Testing Framework | ✅ Pass | 12.965s |  |
| Web Application Security Scanner | ✅ Pass | 10.203s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.265s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.780s

---

### enricher (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 5.822s |  |
| Simple Math | ✅ Pass | 8.812s |  |
| Count from 1 to 5 | ✅ Pass | 4.942s |  |
| Math Calculation | ✅ Pass | 5.295s |  |
| Basic Echo Function | ✅ Pass | 3.634s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.727s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.957s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 7.226s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 3.308s |  |
| JSON Response Function | ✅ Pass | 7.863s |  |
| Ask Advice Function | ✅ Pass | 4.381s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.571s |  |
| Basic Context Memory Test | ✅ Pass | 7.343s |  |
| Function Response Memory Test | ✅ Pass | 4.464s |  |
| Function Argument Memory Test | ✅ Pass | 6.395s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.066s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.124s |  |
| Penetration Testing Methodology | ✅ Pass | 11.162s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.898s |  |
| SQL Injection Attack Type | ✅ Pass | 8.524s |  |
| Penetration Testing Framework | ✅ Pass | 11.747s |  |
| Web Application Security Scanner | ✅ Pass | 10.317s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.616s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.705s

---

### coder (anthropic/claude-4-sonnet)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.038s |  |
| Text Transform Uppercase | ✅ Pass | 1.810s |  |
| Count from 1 to 5 | ✅ Pass | 1.872s |  |
| Math Calculation | ✅ Pass | 2.400s |  |
| Basic Echo Function | ✅ Pass | 2.111s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.348s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.517s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.253s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.561s |  |
| Search Query Function | ✅ Pass | 2.016s |  |
| Ask Advice Function | ✅ Pass | 2.363s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.742s |  |
| Basic Context Memory Test | ✅ Pass | 2.941s |  |
| Function Argument Memory Test | ✅ Pass | 1.764s |  |
| Function Response Memory Test | ✅ Pass | 1.999s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.912s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.101s |  |
| Penetration Testing Methodology | ✅ Pass | 5.195s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.804s |  |
| SQL Injection Attack Type | ✅ Pass | 2.188s |  |
| Penetration Testing Framework | ✅ Pass | 5.985s |  |
| Web Application Security Scanner | ✅ Pass | 4.381s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.602s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.953s

---

### installer (google/gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.469s |  |
| Text Transform Uppercase | ✅ Pass | 2.376s |  |
| Count from 1 to 5 | ✅ Pass | 1.039s |  |
| Math Calculation | ✅ Pass | 1.579s |  |
| Basic Echo Function | ✅ Pass | 1.417s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.214s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.954s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.251s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.778s |  |
| Search Query Function | ✅ Pass | 2.765s |  |
| Ask Advice Function | ✅ Pass | 2.787s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.896s |  |
| Basic Context Memory Test | ✅ Pass | 3.006s |  |
| Function Argument Memory Test | ✅ Pass | 1.224s |  |
| Function Response Memory Test | ✅ Pass | 1.743s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.311s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.248s |  |
| Penetration Testing Methodology | ✅ Pass | 2.712s |  |
| SQL Injection Attack Type | ✅ Pass | 1.453s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.593s |  |
| Penetration Testing Framework | ✅ Pass | 4.916s |  |
| Web Application Security Scanner | ✅ Pass | 3.753s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.673s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.703s

---

### pentester (moonshotai/Kimi-K2-Instruct-0905)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.661s |  |
| Text Transform Uppercase | ✅ Pass | 0.641s |  |
| Count from 1 to 5 | ✅ Pass | 1.036s |  |
| Math Calculation | ✅ Pass | 0.594s |  |
| Basic Echo Function | ✅ Pass | 1.246s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.594s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.479s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.881s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.316s |  |
| Search Query Function | ✅ Pass | 1.443s |  |
| Ask Advice Function | ✅ Pass | 1.316s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.617s |  |
| Basic Context Memory Test | ✅ Pass | 0.840s |  |
| Function Argument Memory Test | ✅ Pass | 0.659s |  |
| Function Response Memory Test | ✅ Pass | 0.611s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 2.006s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.633s |  |
| Penetration Testing Methodology | ✅ Pass | 1.000s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.791s |  |
| SQL Injection Attack Type | ✅ Pass | 5.163s |  |
| Penetration Testing Framework | ✅ Pass | 0.761s |  |
| Web Application Security Scanner | ✅ Pass | 0.598s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.073s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.303s

---

