# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 12:05:49 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | meta-llama/Llama-3.3-70B-Instruct | true | 22/23 (95.65%) | 1.723s |
| simple_json | google/gemma-3-27b-it | false | 5/5 (100.00%) | 2.746s |
| primary_agent | deepseek-ai/DeepSeek-R1-0528-Turbo | true | 23/23 (100.00%) | 2.049s |
| assistant | deepseek-ai/DeepSeek-R1-0528-Turbo | true | 23/23 (100.00%) | 2.110s |
| generator | deepseek-ai/DeepSeek-R1-0528-Turbo | true | 23/23 (100.00%) | 1.876s |
| refiner | deepseek-ai/DeepSeek-R1-0528-Turbo | true | 23/23 (100.00%) | 1.982s |
| adviser | google/gemini-2.5-pro | true | 22/23 (95.65%) | 4.558s |
| reflector | google/gemini-2.5-flash | true | 23/23 (100.00%) | 2.152s |
| searcher | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 4.891s |
| enricher | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.066s |
| coder | anthropic/claude-4-sonnet | true | 23/23 (100.00%) | 2.381s |
| installer | google/gemini-2.5-flash | true | 23/23 (100.00%) | 2.117s |
| pentester | moonshotai/Kimi-K2-Instruct | true | 21/23 (91.30%) | 0.570s |

**Total**: 277/281 (98.58%) successful tests
**Overall average latency**: 2.625s

## Detailed Results

### simple (meta-llama/Llama-3.3-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.383s |  |
| Text Transform Uppercase | ✅ Pass | 0.346s |  |
| Count from 1 to 5 | ✅ Pass | 1.417s |  |
| Math Calculation | ✅ Pass | 0.297s |  |
| Basic Echo Function | ✅ Pass | 0.689s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.572s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.741s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.793s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.086s |  |
| Search Query Function | ❌ Fail | 1.074s | no tool calls found, expected at least 1 |
| Ask Advice Function | ✅ Pass | 1.437s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.885s |  |
| Basic Context Memory Test | ✅ Pass | 1.156s |  |
| Function Argument Memory Test | ✅ Pass | 0.359s |  |
| Function Response Memory Test | ✅ Pass | 0.670s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.096s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.337s |  |
| Penetration Testing Methodology | ✅ Pass | 7.562s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.391s |  |
| SQL Injection Attack Type | ✅ Pass | 0.286s |  |
| Penetration Testing Framework | ✅ Pass | 3.961s |  |
| Web Application Security Scanner | ✅ Pass | 4.478s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.613s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.723s

---

### simple_json (google/gemma-3-27b-it)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.795s |  |
| Project Information JSON | ✅ Pass | 0.626s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.051s |  |
| User Profile JSON | ✅ Pass | 3.410s |  |
| Vulnerability Report Memory Test | ✅ Pass | 7.845s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 2.746s

---

### primary_agent (deepseek-ai/DeepSeek-R1-0528-Turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.433s |  |
| Text Transform Uppercase | ✅ Pass | 0.863s |  |
| Count from 1 to 5 | ✅ Pass | 2.271s |  |
| Math Calculation | ✅ Pass | 1.670s |  |
| Basic Echo Function | ✅ Pass | 1.315s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.308s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.653s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.439s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.425s |  |
| Search Query Function | ✅ Pass | 1.201s |  |
| Ask Advice Function | ✅ Pass | 2.190s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.318s |  |
| Basic Context Memory Test | ✅ Pass | 1.404s |  |
| Function Argument Memory Test | ✅ Pass | 1.021s |  |
| Function Response Memory Test | ✅ Pass | 1.244s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.997s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.849s |  |
| Penetration Testing Methodology | ✅ Pass | 4.619s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.111s |  |
| SQL Injection Attack Type | ✅ Pass | 2.675s |  |
| Penetration Testing Framework | ✅ Pass | 5.233s |  |
| Web Application Security Scanner | ✅ Pass | 4.580s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.298s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.049s

---

### assistant (deepseek-ai/DeepSeek-R1-0528-Turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 0.810s |  |
| Simple Math | ✅ Pass | 3.463s |  |
| Count from 1 to 5 | ✅ Pass | 0.904s |  |
| Math Calculation | ✅ Pass | 1.906s |  |
| Basic Echo Function | ✅ Pass | 1.773s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.414s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.232s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.614s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.111s |  |
| Search Query Function | ✅ Pass | 1.386s |  |
| Ask Advice Function | ✅ Pass | 1.358s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.174s |  |
| Basic Context Memory Test | ✅ Pass | 1.397s |  |
| Function Argument Memory Test | ✅ Pass | 0.922s |  |
| Function Response Memory Test | ✅ Pass | 2.289s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.406s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.505s |  |
| Penetration Testing Methodology | ✅ Pass | 3.478s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.089s |  |
| SQL Injection Attack Type | ✅ Pass | 1.275s |  |
| Penetration Testing Framework | ✅ Pass | 3.592s |  |
| Web Application Security Scanner | ✅ Pass | 5.590s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.823s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.110s

---

### generator (deepseek-ai/DeepSeek-R1-0528-Turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.446s |  |
| Text Transform Uppercase | ✅ Pass | 0.755s |  |
| Count from 1 to 5 | ✅ Pass | 1.042s |  |
| Math Calculation | ✅ Pass | 1.246s |  |
| Basic Echo Function | ✅ Pass | 1.466s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.119s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.645s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.436s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.876s |  |
| Search Query Function | ✅ Pass | 1.214s |  |
| Ask Advice Function | ✅ Pass | 2.018s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.645s |  |
| Basic Context Memory Test | ✅ Pass | 1.252s |  |
| Function Argument Memory Test | ✅ Pass | 0.872s |  |
| Function Response Memory Test | ✅ Pass | 1.082s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.526s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.129s |  |
| Penetration Testing Methodology | ✅ Pass | 3.908s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.284s |  |
| SQL Injection Attack Type | ✅ Pass | 1.991s |  |
| Penetration Testing Framework | ✅ Pass | 5.161s |  |
| Web Application Security Scanner | ✅ Pass | 2.621s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.411s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.876s

---

### refiner (deepseek-ai/DeepSeek-R1-0528-Turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.949s |  |
| Text Transform Uppercase | ✅ Pass | 0.461s |  |
| Count from 1 to 5 | ✅ Pass | 0.621s |  |
| Math Calculation | ✅ Pass | 1.755s |  |
| Basic Echo Function | ✅ Pass | 0.982s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.313s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.702s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.905s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.484s |  |
| Search Query Function | ✅ Pass | 0.964s |  |
| Ask Advice Function | ✅ Pass | 1.006s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.176s |  |
| Basic Context Memory Test | ✅ Pass | 2.296s |  |
| Function Argument Memory Test | ✅ Pass | 1.138s |  |
| Function Response Memory Test | ✅ Pass | 0.966s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.681s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.959s |  |
| Penetration Testing Methodology | ✅ Pass | 4.335s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.709s |  |
| SQL Injection Attack Type | ✅ Pass | 1.488s |  |
| Penetration Testing Framework | ✅ Pass | 6.190s |  |
| Web Application Security Scanner | ✅ Pass | 4.187s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.314s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.982s

---

### adviser (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.222s |  |
| Text Transform Uppercase | ✅ Pass | 3.274s |  |
| Count from 1 to 5 | ✅ Pass | 3.778s |  |
| Math Calculation | ✅ Pass | 2.690s |  |
| Basic Echo Function | ✅ Pass | 2.111s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.209s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.367s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.047s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 2.515s | API returned unexpected status code: 500: inference exception |
| Search Query Function | ✅ Pass | 2.145s |  |
| Ask Advice Function | ✅ Pass | 2.389s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.786s |  |
| Basic Context Memory Test | ✅ Pass | 4.467s |  |
| Function Argument Memory Test | ✅ Pass | 2.458s |  |
| Function Response Memory Test | ✅ Pass | 3.002s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.173s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.591s |  |
| Penetration Testing Methodology | ✅ Pass | 8.426s |  |
| SQL Injection Attack Type | ✅ Pass | 5.965s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.020s |  |
| Penetration Testing Framework | ✅ Pass | 12.528s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.958s |  |
| Web Application Security Scanner | ✅ Pass | 10.709s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.558s

---

### reflector (google/gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.908s |  |
| Text Transform Uppercase | ✅ Pass | 0.956s |  |
| Count from 1 to 5 | ✅ Pass | 1.120s |  |
| Math Calculation | ✅ Pass | 0.941s |  |
| Basic Echo Function | ✅ Pass | 1.203s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.964s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.075s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.465s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.345s |  |
| Search Query Function | ✅ Pass | 1.291s |  |
| Ask Advice Function | ✅ Pass | 1.536s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.662s |  |
| Basic Context Memory Test | ✅ Pass | 1.771s |  |
| Function Argument Memory Test | ✅ Pass | 1.358s |  |
| Function Response Memory Test | ✅ Pass | 1.572s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.063s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.773s |  |
| Penetration Testing Methodology | ✅ Pass | 2.470s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.989s |  |
| SQL Injection Attack Type | ✅ Pass | 1.716s |  |
| Penetration Testing Framework | ✅ Pass | 7.658s |  |
| Web Application Security Scanner | ✅ Pass | 7.415s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.238s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.152s

---

### searcher (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.957s |  |
| Text Transform Uppercase | ✅ Pass | 4.068s |  |
| Count from 1 to 5 | ✅ Pass | 3.145s |  |
| Math Calculation | ✅ Pass | 4.022s |  |
| Basic Echo Function | ✅ Pass | 2.484s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.094s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.468s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.877s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.443s |  |
| Search Query Function | ✅ Pass | 3.329s |  |
| Ask Advice Function | ✅ Pass | 4.545s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.468s |  |
| Basic Context Memory Test | ✅ Pass | 4.684s |  |
| Function Argument Memory Test | ✅ Pass | 3.013s |  |
| Function Response Memory Test | ✅ Pass | 2.463s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.347s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.868s |  |
| Penetration Testing Methodology | ✅ Pass | 6.279s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.722s |  |
| SQL Injection Attack Type | ✅ Pass | 9.415s |  |
| Penetration Testing Framework | ✅ Pass | 8.934s |  |
| Web Application Security Scanner | ✅ Pass | 6.818s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.051s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.891s

---

### enricher (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.682s |  |
| Text Transform Uppercase | ✅ Pass | 5.359s |  |
| Count from 1 to 5 | ✅ Pass | 3.636s |  |
| Math Calculation | ✅ Pass | 4.993s |  |
| Basic Echo Function | ✅ Pass | 2.686s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.960s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.663s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.722s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.386s |  |
| Search Query Function | ✅ Pass | 3.420s |  |
| Ask Advice Function | ✅ Pass | 4.952s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.582s |  |
| Basic Context Memory Test | ✅ Pass | 3.974s |  |
| Function Argument Memory Test | ✅ Pass | 3.322s |  |
| Function Response Memory Test | ✅ Pass | 2.609s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.247s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.154s |  |
| Penetration Testing Methodology | ✅ Pass | 7.772s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.260s |  |
| SQL Injection Attack Type | ✅ Pass | 8.574s |  |
| Penetration Testing Framework | ✅ Pass | 10.725s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.880s |  |
| Web Application Security Scanner | ✅ Pass | 9.961s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.066s

---

### coder (anthropic/claude-4-sonnet)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.318s |  |
| Text Transform Uppercase | ✅ Pass | 1.586s |  |
| Count from 1 to 5 | ✅ Pass | 1.331s |  |
| Math Calculation | ✅ Pass | 1.380s |  |
| Basic Echo Function | ✅ Pass | 2.162s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.687s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.401s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.677s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.224s |  |
| Search Query Function | ✅ Pass | 2.180s |  |
| Ask Advice Function | ✅ Pass | 2.221s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.380s |  |
| Basic Context Memory Test | ✅ Pass | 1.849s |  |
| Function Argument Memory Test | ✅ Pass | 2.217s |  |
| Function Response Memory Test | ✅ Pass | 1.443s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.963s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.584s |  |
| Penetration Testing Methodology | ✅ Pass | 5.247s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.846s |  |
| SQL Injection Attack Type | ✅ Pass | 1.549s |  |
| Penetration Testing Framework | ✅ Pass | 3.168s |  |
| Web Application Security Scanner | ✅ Pass | 4.361s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.978s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.381s

---

### installer (google/gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.124s |  |
| Text Transform Uppercase | ✅ Pass | 1.023s |  |
| Count from 1 to 5 | ✅ Pass | 1.142s |  |
| Math Calculation | ✅ Pass | 0.894s |  |
| Basic Echo Function | ✅ Pass | 1.291s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.026s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.065s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.412s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.187s |  |
| Search Query Function | ✅ Pass | 1.429s |  |
| Ask Advice Function | ✅ Pass | 1.499s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.003s |  |
| Basic Context Memory Test | ✅ Pass | 1.475s |  |
| Function Argument Memory Test | ✅ Pass | 1.286s |  |
| Function Response Memory Test | ✅ Pass | 1.757s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.004s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.497s |  |
| Penetration Testing Methodology | ✅ Pass | 3.134s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.979s |  |
| SQL Injection Attack Type | ✅ Pass | 1.601s |  |
| Penetration Testing Framework | ✅ Pass | 8.108s |  |
| Web Application Security Scanner | ✅ Pass | 6.488s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.266s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.117s

---

### pentester (moonshotai/Kimi-K2-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.348s |  |
| Text Transform Uppercase | ✅ Pass | 0.394s |  |
| Count from 1 to 5 | ✅ Pass | 0.462s |  |
| Math Calculation | ✅ Pass | 0.654s |  |
| Basic Echo Function | ✅ Pass | 0.640s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.357s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.393s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.580s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.286s |  |
| Search Query Function | ✅ Pass | 0.809s |  |
| Ask Advice Function | ✅ Pass | 0.751s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.744s |  |
| Basic Context Memory Test | ✅ Pass | 0.511s |  |
| Function Argument Memory Test | ✅ Pass | 0.416s |  |
| Function Response Memory Test | ✅ Pass | 0.334s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.161s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.372s |  |
| Penetration Testing Methodology | ✅ Pass | 0.410s |  |
| Vulnerability Assessment Tools | ❌ Fail | 0.382s | expected text 'network' not found |
| SQL Injection Attack Type | ✅ Pass | 0.508s |  |
| Penetration Testing Framework | ✅ Pass | 0.432s |  |
| Web Application Security Scanner | ✅ Pass | 0.387s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.771s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 0.570s

---

