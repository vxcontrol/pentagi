# LLM Agent Testing Report

Generated: Tue, 08 Jul 2025 21:46:13 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | meta-llama/Llama-3.3-70B-Instruct | false | 17/18 (94.44%) | 1.984s |
| simple_json | google/gemma-3-27b-it | false | 4/4 (100.00%) | 1.838s |
| primary_agent | deepseek-ai/DeepSeek-R1 | true | 16/18 (88.89%) | 3.506s |
| assistant | deepseek-ai/DeepSeek-R1 | true | 17/18 (94.44%) | 3.015s |
| generator | deepseek-ai/DeepSeek-R1 | true | 17/18 (94.44%) | 3.372s |
| refiner | deepseek-ai/DeepSeek-R1 | true | 18/18 (100.00%) | 4.303s |
| adviser | deepseek-ai/DeepSeek-R1 | true | 16/18 (88.89%) | 3.191s |
| reflector | Qwen/Qwen3-30B-A3B | true | 18/18 (100.00%) | 3.080s |
| searcher | nvidia/Llama-3.1-Nemotron-70B-Instruct | true | 18/18 (100.00%) | 2.613s |
| enricher | nvidia/Llama-3.1-Nemotron-70B-Instruct | true | 18/18 (100.00%) | 2.619s |
| coder | deepseek-ai/DeepSeek-R1 | true | 18/18 (100.00%) | 3.716s |
| installer | nvidia/Llama-3.1-Nemotron-70B-Instruct | true | 18/18 (100.00%) | 2.624s |
| pentester | deepseek-ai/DeepSeek-R1 | true | 17/18 (94.44%) | 3.345s |

**Total**: 212/220 (96.36%) successful tests
**Overall average latency**: 3.091s

## Detailed Results

### simple (meta-llama/Llama-3.3-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.537s |  |
| Text Transform Uppercase | ✅ Pass | 1.095s |  |
| Count from 1 to 5 | ✅ Pass | 0.566s |  |
| Math Calculation | ✅ Pass | 0.509s |  |
| Basic Echo Function | ✅ Pass | 1.260s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.399s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.553s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.812s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.731s |  |
| Search Query Function | ❌ Fail | 0.802s | expected 1 tool calls, got 0 |
| Ask Advice Function | ✅ Pass | 0.987s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.751s |  |
| Penetration Testing Methodology | ✅ Pass | 2.949s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.987s |  |
| SQL Injection Attack Type | ✅ Pass | 0.356s |  |
| Penetration Testing Framework | ✅ Pass | 7.939s |  |
| Web Application Security Scanner | ✅ Pass | 5.569s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.907s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 1.984s

---

### simple_json (google/gemma-3-27b-it)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.913s |  |
| Project Information JSON | ✅ Pass | 1.887s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.344s |  |
| User Profile JSON | ✅ Pass | 2.204s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 1.838s

---

### primary_agent (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.576s |  |
| Text Transform Uppercase | ✅ Pass | 4.421s |  |
| Count from 1 to 5 | ✅ Pass | 2.009s |  |
| Math Calculation | ✅ Pass | 2.028s |  |
| Basic Echo Function | ✅ Pass | 2.377s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.646s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.857s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 2.840s | expected 1 tool calls, got 0 |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.971s |  |
| Search Query Function | ✅ Pass | 1.726s |  |
| Ask Advice Function | ✅ Pass | 2.033s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 1.582s | expected 1 tool calls, got 0 |
| Penetration Testing Methodology | ✅ Pass | 6.335s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.759s |  |
| SQL Injection Attack Type | ✅ Pass | 3.200s |  |
| Penetration Testing Framework | ✅ Pass | 8.376s |  |
| Web Application Security Scanner | ✅ Pass | 7.544s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.809s |  |

**Summary**: 16/18 (88.89%) successful tests

**Average latency**: 3.506s

---

### assistant (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.902s |  |
| Text Transform Uppercase | ✅ Pass | 2.764s |  |
| Count from 1 to 5 | ✅ Pass | 1.921s |  |
| Math Calculation | ✅ Pass | 2.872s |  |
| Basic Echo Function | ✅ Pass | 1.809s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.893s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.752s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.430s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.031s |  |
| Search Query Function | ✅ Pass | 2.032s |  |
| Ask Advice Function | ✅ Pass | 2.786s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 1.386s | expected 1 tool calls, got 0 |
| Penetration Testing Methodology | ✅ Pass | 3.237s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.549s |  |
| SQL Injection Attack Type | ✅ Pass | 2.777s |  |
| Penetration Testing Framework | ✅ Pass | 6.083s |  |
| Web Application Security Scanner | ✅ Pass | 6.856s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.182s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 3.015s

---

### generator (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.655s |  |
| Text Transform Uppercase | ✅ Pass | 3.238s |  |
| Count from 1 to 5 | ✅ Pass | 1.859s |  |
| Math Calculation | ✅ Pass | 1.436s |  |
| Basic Echo Function | ✅ Pass | 2.039s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.765s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.196s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.493s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.574s |  |
| Search Query Function | ✅ Pass | 1.847s |  |
| Ask Advice Function | ✅ Pass | 3.135s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 2.009s | expected 1 tool calls, got 0 |
| Penetration Testing Methodology | ✅ Pass | 5.507s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.552s |  |
| SQL Injection Attack Type | ✅ Pass | 2.672s |  |
| Penetration Testing Framework | ✅ Pass | 6.392s |  |
| Web Application Security Scanner | ✅ Pass | 9.372s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.943s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 3.372s

---

### refiner (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.096s |  |
| Text Transform Uppercase | ✅ Pass | 3.034s |  |
| Count from 1 to 5 | ✅ Pass | 1.584s |  |
| Math Calculation | ✅ Pass | 1.627s |  |
| Basic Echo Function | ✅ Pass | 2.291s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.252s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.637s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.637s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.460s |  |
| Search Query Function | ✅ Pass | 1.384s |  |
| Ask Advice Function | ✅ Pass | 1.922s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.158s |  |
| Penetration Testing Methodology | ✅ Pass | 5.764s |  |
| SQL Injection Attack Type | ✅ Pass | 2.280s |  |
| Vulnerability Assessment Tools | ✅ Pass | 23.794s |  |
| Penetration Testing Framework | ✅ Pass | 7.377s |  |
| Web Application Security Scanner | ✅ Pass | 6.562s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.591s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.303s

---

### adviser (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.797s |  |
| Text Transform Uppercase | ✅ Pass | 2.702s |  |
| Count from 1 to 5 | ✅ Pass | 3.163s |  |
| Math Calculation | ✅ Pass | 1.739s |  |
| Basic Echo Function | ✅ Pass | 3.568s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.843s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.637s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 2.580s | expected 1 tool calls, got 0 |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.616s |  |
| Search Query Function | ✅ Pass | 1.346s |  |
| Ask Advice Function | ✅ Pass | 2.110s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 1.395s | expected 1 tool calls, got 0 |
| Penetration Testing Methodology | ✅ Pass | 6.210s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.656s |  |
| SQL Injection Attack Type | ✅ Pass | 2.503s |  |
| Penetration Testing Framework | ✅ Pass | 8.110s |  |
| Web Application Security Scanner | ✅ Pass | 5.684s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.766s |  |

**Summary**: 16/18 (88.89%) successful tests

**Average latency**: 3.191s

---

### reflector (Qwen/Qwen3-30B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.833s |  |
| Text Transform Uppercase | ✅ Pass | 2.917s |  |
| Count from 1 to 5 | ✅ Pass | 1.889s |  |
| Math Calculation | ✅ Pass | 3.159s |  |
| Basic Echo Function | ✅ Pass | 1.932s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.749s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.059s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.516s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.239s |  |
| Search Query Function | ✅ Pass | 1.487s |  |
| Ask Advice Function | ✅ Pass | 2.403s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.301s |  |
| Penetration Testing Methodology | ✅ Pass | 4.432s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.417s |  |
| SQL Injection Attack Type | ✅ Pass | 4.523s |  |
| Penetration Testing Framework | ✅ Pass | 6.187s |  |
| Web Application Security Scanner | ✅ Pass | 4.314s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.069s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.080s

---

### searcher (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.360s |  |
| Text Transform Uppercase | ✅ Pass | 0.418s |  |
| Count from 1 to 5 | ✅ Pass | 0.578s |  |
| Math Calculation | ✅ Pass | 0.782s |  |
| Basic Echo Function | ✅ Pass | 0.826s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.381s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.709s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.576s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.848s |  |
| Search Query Function | ✅ Pass | 1.086s |  |
| Ask Advice Function | ✅ Pass | 1.055s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.509s |  |
| Penetration Testing Methodology | ✅ Pass | 4.621s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.218s |  |
| SQL Injection Attack Type | ✅ Pass | 0.873s |  |
| Penetration Testing Framework | ✅ Pass | 13.449s |  |
| Web Application Security Scanner | ✅ Pass | 6.639s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.093s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.613s

---

### enricher (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.470s |  |
| Text Transform Uppercase | ✅ Pass | 0.473s |  |
| Count from 1 to 5 | ✅ Pass | 0.574s |  |
| Math Calculation | ✅ Pass | 0.568s |  |
| Basic Echo Function | ✅ Pass | 0.983s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.364s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.337s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.556s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.855s |  |
| Search Query Function | ✅ Pass | 0.738s |  |
| Ask Advice Function | ✅ Pass | 1.027s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.509s |  |
| Penetration Testing Methodology | ✅ Pass | 4.041s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.011s |  |
| SQL Injection Attack Type | ✅ Pass | 0.842s |  |
| Penetration Testing Framework | ✅ Pass | 11.730s |  |
| Web Application Security Scanner | ✅ Pass | 10.032s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.031s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.619s

---

### coder (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.289s |  |
| Text Transform Uppercase | ✅ Pass | 1.708s |  |
| Count from 1 to 5 | ✅ Pass | 1.788s |  |
| Math Calculation | ✅ Pass | 2.868s |  |
| Basic Echo Function | ✅ Pass | 2.020s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.605s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.512s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.629s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.660s |  |
| Search Query Function | ✅ Pass | 1.372s |  |
| Ask Advice Function | ✅ Pass | 1.601s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.093s |  |
| Penetration Testing Methodology | ✅ Pass | 12.039s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.128s |  |
| SQL Injection Attack Type | ✅ Pass | 3.023s |  |
| Penetration Testing Framework | ✅ Pass | 11.055s |  |
| Web Application Security Scanner | ✅ Pass | 9.859s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.627s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.716s

---

### installer (nvidia/Llama-3.1-Nemotron-70B-Instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.361s |  |
| Text Transform Uppercase | ✅ Pass | 0.921s |  |
| Count from 1 to 5 | ✅ Pass | 0.938s |  |
| Math Calculation | ✅ Pass | 0.476s |  |
| Basic Echo Function | ✅ Pass | 1.121s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.384s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.355s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.518s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.138s |  |
| Search Query Function | ✅ Pass | 1.120s |  |
| Ask Advice Function | ✅ Pass | 1.293s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.477s |  |
| Penetration Testing Methodology | ✅ Pass | 4.959s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.399s |  |
| SQL Injection Attack Type | ✅ Pass | 0.823s |  |
| Penetration Testing Framework | ✅ Pass | 13.583s |  |
| Web Application Security Scanner | ✅ Pass | 6.359s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.004s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.624s

---

### pentester (deepseek-ai/DeepSeek-R1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.021s |  |
| Text Transform Uppercase | ✅ Pass | 3.398s |  |
| Count from 1 to 5 | ✅ Pass | 2.324s |  |
| Math Calculation | ✅ Pass | 2.707s |  |
| Basic Echo Function | ✅ Pass | 1.407s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.113s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.046s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.410s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.636s |  |
| Search Query Function | ✅ Pass | 2.407s |  |
| Ask Advice Function | ✅ Pass | 1.769s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 2.229s | expected 1 tool calls, got 0 |
| Penetration Testing Methodology | ✅ Pass | 5.018s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.909s |  |
| SQL Injection Attack Type | ✅ Pass | 4.176s |  |
| Penetration Testing Framework | ✅ Pass | 10.707s |  |
| Web Application Security Scanner | ✅ Pass | 5.503s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.412s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 3.345s

---

