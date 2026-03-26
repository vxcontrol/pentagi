# LLM Agent Testing Report

Generated: Mon, 23 Mar 2026 16:29:35 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | nemotron-3-super:cloud | false | 22/23 (95.65%) | 22.627s |
| simple_json | nemotron-3-super:cloud | false | 5/5 (100.00%) | 27.380s |
| primary_agent | qwen3-coder-next:cloud | false | 23/23 (100.00%) | 4.101s |
| assistant | nemotron-3-super:cloud | false | 23/23 (100.00%) | 22.773s |
| generator | qwen3-coder-next:cloud | false | 23/23 (100.00%) | 2.348s |
| refiner | glm-5:cloud | false | 23/23 (100.00%) | 7.964s |
| adviser | minimax-m2.7:cloud | false | 23/23 (100.00%) | 3.990s |
| reflector | glm-5:cloud | false | 23/23 (100.00%) | 9.201s |
| searcher | qwen3.5:397b-cloud | false | 20/23 (86.96%) | 46.174s |
| enricher | minimax-m2.7:cloud | false | 23/23 (100.00%) | 6.153s |
| coder | qwen3-coder-next:cloud | false | 23/23 (100.00%) | 3.223s |
| installer | devstral-2:123b-cloud | false | 23/23 (100.00%) | 8.049s |
| pentester | qwen3-coder-next:cloud | false | 23/23 (100.00%) | 2.227s |

**Total**: 277/281 (98.58%) successful tests
**Overall average latency**: 11.851s

## Detailed Results

### simple (nemotron-3-super:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.444s |  |
| Text Transform Uppercase | ✅ Pass | 10.749s |  |
| Count from 1 to 5 | ✅ Pass | 19.076s |  |
| Math Calculation | ✅ Pass | 6.549s |  |
| Basic Echo Function | ✅ Pass | 16.980s |  |
| Streaming Simple Math Streaming | ✅ Pass | 18.192s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 8.274s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 10.723s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 20.851s |  |
| Search Query Function | ✅ Pass | 16.304s |  |
| Ask Advice Function | ✅ Pass | 33.178s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 14.129s |  |
| Basic Context Memory Test | ✅ Pass | 25.373s |  |
| Function Argument Memory Test | ✅ Pass | 34.212s |  |
| Function Response Memory Test | ✅ Pass | 22.402s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 39.672s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 14.158s |  |
| Penetration Testing Methodology | ✅ Pass | 57.038s |  |
| Vulnerability Assessment Tools | ✅ Pass | 40.968s |  |
| SQL Injection Attack Type | ✅ Pass | 9.713s |  |
| Penetration Testing Framework | ✅ Pass | 65.247s |  |
| Web Application Security Scanner | ✅ Pass | 15.760s |  |
| Penetration Testing Tool Selection | ✅ Pass | 18.418s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 22.627s

---

### simple_json (nemotron-3-super:cloud)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 30.553s |  |
| Person Information JSON | ✅ Pass | 23.124s |  |
| Project Information JSON | ✅ Pass | 25.124s |  |
| User Profile JSON | ✅ Pass | 31.109s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 26.988s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 27.380s

---

### primary_agent (qwen3-coder-next:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.931s |  |
| Text Transform Uppercase | ✅ Pass | 0.763s |  |
| Count from 1 to 5 | ✅ Pass | 3.641s |  |
| Math Calculation | ✅ Pass | 0.932s |  |
| Basic Echo Function | ✅ Pass | 0.869s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.700s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.713s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.812s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.959s |  |
| Search Query Function | ✅ Pass | 0.933s |  |
| Ask Advice Function | ✅ Pass | 33.999s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.998s |  |
| Basic Context Memory Test | ✅ Pass | 8.687s |  |
| Function Argument Memory Test | ✅ Pass | 6.186s |  |
| Function Response Memory Test | ✅ Pass | 2.615s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.527s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.732s |  |
| Penetration Testing Methodology | ✅ Pass | 6.404s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.203s |  |
| SQL Injection Attack Type | ✅ Pass | 2.766s |  |
| Penetration Testing Framework | ✅ Pass | 8.059s |  |
| Web Application Security Scanner | ✅ Pass | 1.546s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.339s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.101s

---

### assistant (nemotron-3-super:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 13.460s |  |
| Text Transform Uppercase | ✅ Pass | 8.769s |  |
| Count from 1 to 5 | ✅ Pass | 6.944s |  |
| Math Calculation | ✅ Pass | 31.133s |  |
| Basic Echo Function | ✅ Pass | 10.710s |  |
| Streaming Simple Math Streaming | ✅ Pass | 16.396s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 7.617s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 13.783s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 20.992s |  |
| Search Query Function | ✅ Pass | 13.170s |  |
| Ask Advice Function | ✅ Pass | 17.042s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 18.533s |  |
| Basic Context Memory Test | ✅ Pass | 23.059s |  |
| Function Argument Memory Test | ✅ Pass | 25.397s |  |
| Function Response Memory Test | ✅ Pass | 24.149s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 39.970s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 8.683s |  |
| Penetration Testing Methodology | ✅ Pass | 63.505s |  |
| Vulnerability Assessment Tools | ✅ Pass | 34.659s |  |
| SQL Injection Attack Type | ✅ Pass | 26.478s |  |
| Penetration Testing Framework | ✅ Pass | 58.539s |  |
| Web Application Security Scanner | ✅ Pass | 33.683s |  |
| Penetration Testing Tool Selection | ✅ Pass | 7.095s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 22.773s

---

### generator (qwen3-coder-next:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.633s |  |
| Text Transform Uppercase | ✅ Pass | 0.664s |  |
| Count from 1 to 5 | ✅ Pass | 1.159s |  |
| Math Calculation | ✅ Pass | 7.640s |  |
| Basic Echo Function | ✅ Pass | 1.056s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.638s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.870s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.083s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.946s |  |
| Search Query Function | ✅ Pass | 0.880s |  |
| Ask Advice Function | ✅ Pass | 1.013s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.946s |  |
| Basic Context Memory Test | ✅ Pass | 1.315s |  |
| Function Argument Memory Test | ✅ Pass | 0.722s |  |
| Function Response Memory Test | ✅ Pass | 3.881s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.597s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.642s |  |
| Penetration Testing Methodology | ✅ Pass | 9.073s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.598s |  |
| SQL Injection Attack Type | ✅ Pass | 5.544s |  |
| Penetration Testing Framework | ✅ Pass | 1.304s |  |
| Web Application Security Scanner | ✅ Pass | 1.261s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.537s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.348s

---

### refiner (glm-5:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.032s |  |
| Text Transform Uppercase | ✅ Pass | 5.144s |  |
| Count from 1 to 5 | ✅ Pass | 1.324s |  |
| Math Calculation | ✅ Pass | 22.879s |  |
| Basic Echo Function | ✅ Pass | 2.930s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.628s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.727s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.291s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.312s |  |
| Search Query Function | ✅ Pass | 1.351s |  |
| Ask Advice Function | ✅ Pass | 1.539s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.405s |  |
| Basic Context Memory Test | ✅ Pass | 5.605s |  |
| Function Argument Memory Test | ✅ Pass | 2.879s |  |
| Function Response Memory Test | ✅ Pass | 1.468s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 15.864s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.101s |  |
| Penetration Testing Methodology | ✅ Pass | 16.016s |  |
| Vulnerability Assessment Tools | ✅ Pass | 24.194s |  |
| SQL Injection Attack Type | ✅ Pass | 7.143s |  |
| Penetration Testing Framework | ✅ Pass | 22.774s |  |
| Web Application Security Scanner | ✅ Pass | 34.765s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.788s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.964s

---

### adviser (minimax-m2.7:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.791s |  |
| Text Transform Uppercase | ✅ Pass | 1.907s |  |
| Count from 1 to 5 | ✅ Pass | 5.236s |  |
| Math Calculation | ✅ Pass | 4.146s |  |
| Basic Echo Function | ✅ Pass | 1.855s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.809s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.696s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 7.184s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.898s |  |
| Search Query Function | ✅ Pass | 4.785s |  |
| Ask Advice Function | ✅ Pass | 1.973s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.405s |  |
| Basic Context Memory Test | ✅ Pass | 2.504s |  |
| Function Argument Memory Test | ✅ Pass | 2.930s |  |
| Function Response Memory Test | ✅ Pass | 2.279s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.874s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.765s |  |
| Penetration Testing Methodology | ✅ Pass | 3.184s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.047s |  |
| SQL Injection Attack Type | ✅ Pass | 5.317s |  |
| Penetration Testing Framework | ✅ Pass | 5.817s |  |
| Web Application Security Scanner | ✅ Pass | 4.591s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.773s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.990s

---

### reflector (glm-5:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 18.770s |  |
| Text Transform Uppercase | ✅ Pass | 3.024s |  |
| Count from 1 to 5 | ✅ Pass | 1.798s |  |
| Math Calculation | ✅ Pass | 2.176s |  |
| Basic Echo Function | ✅ Pass | 4.612s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.214s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 17.066s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.913s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.675s |  |
| Search Query Function | ✅ Pass | 1.432s |  |
| Ask Advice Function | ✅ Pass | 1.765s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 11.441s |  |
| Basic Context Memory Test | ✅ Pass | 5.525s |  |
| Function Argument Memory Test | ✅ Pass | 10.351s |  |
| Function Response Memory Test | ✅ Pass | 3.459s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.899s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.117s |  |
| Penetration Testing Methodology | ✅ Pass | 10.196s |  |
| Vulnerability Assessment Tools | ✅ Pass | 63.102s |  |
| SQL Injection Attack Type | ✅ Pass | 4.516s |  |
| Penetration Testing Framework | ✅ Pass | 17.855s |  |
| Web Application Security Scanner | ✅ Pass | 9.839s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.867s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.201s

---

### searcher (qwen3.5:397b-cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.317s |  |
| Text Transform Uppercase | ✅ Pass | 1.459s |  |
| Count from 1 to 5 | ✅ Pass | 3.248s |  |
| Math Calculation | ✅ Pass | 1.474s |  |
| Basic Echo Function | ✅ Pass | 2.060s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.405s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.521s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.976s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.263s |  |
| Search Query Function | ✅ Pass | 0.955s |  |
| Ask Advice Function | ✅ Pass | 1.196s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.967s |  |
| Basic Context Memory Test | ✅ Pass | 2.587s |  |
| Function Argument Memory Test | ✅ Pass | 1.596s |  |
| Function Response Memory Test | ✅ Pass | 1.408s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.083s |  |
| Penetration Testing Methodology | ✅ Pass | 5.672s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 600.006s | Post "https://ollama\.com/api/chat?ts=1774282315": net/http: request canceled \(Client\.Timeout exceeded while awaiting headers\) |
| SQL Injection Attack Type | ✅ Pass | 4.083s |  |
| Vulnerability Assessment Tools | ❌ Fail | 391.303s | 500 Internal Server Error: Internal Server Error \(ref: 9d3370b7\-e2d2\-4086\-9f8b\-948f29fcd567\) |
| Penetration Testing Framework | ❌ Fail | 30.619s | expected text 'exploitation' not found |
| Web Application Security Scanner | ✅ Pass | 2.647s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.136s |  |

**Summary**: 20/23 (86.96%) successful tests

**Average latency**: 46.174s

---

### enricher (minimax-m2.7:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.381s |  |
| Text Transform Uppercase | ✅ Pass | 2.794s |  |
| Count from 1 to 5 | ✅ Pass | 4.367s |  |
| Math Calculation | ✅ Pass | 2.078s |  |
| Basic Echo Function | ✅ Pass | 1.977s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.884s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.745s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.099s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.883s |  |
| Search Query Function | ✅ Pass | 3.617s |  |
| Ask Advice Function | ✅ Pass | 1.873s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.564s |  |
| Basic Context Memory Test | ✅ Pass | 1.803s |  |
| Function Argument Memory Test | ✅ Pass | 2.661s |  |
| Function Response Memory Test | ✅ Pass | 2.227s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.442s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.429s |  |
| Penetration Testing Methodology | ✅ Pass | 5.753s |  |
| Vulnerability Assessment Tools | ✅ Pass | 70.743s |  |
| SQL Injection Attack Type | ✅ Pass | 4.696s |  |
| Penetration Testing Framework | ✅ Pass | 6.033s |  |
| Web Application Security Scanner | ✅ Pass | 3.398s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.064s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.153s

---

### coder (qwen3-coder-next:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.654s |  |
| Text Transform Uppercase | ✅ Pass | 12.005s |  |
| Count from 1 to 5 | ✅ Pass | 0.667s |  |
| Math Calculation | ✅ Pass | 0.633s |  |
| Basic Echo Function | ✅ Pass | 0.819s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.612s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.699s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.923s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.229s |  |
| Search Query Function | ✅ Pass | 0.873s |  |
| Ask Advice Function | ✅ Pass | 1.075s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.897s |  |
| Basic Context Memory Test | ✅ Pass | 1.072s |  |
| Function Argument Memory Test | ✅ Pass | 1.312s |  |
| Function Response Memory Test | ✅ Pass | 0.642s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.623s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.710s |  |
| Penetration Testing Methodology | ✅ Pass | 15.597s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.591s |  |
| SQL Injection Attack Type | ✅ Pass | 1.284s |  |
| Penetration Testing Framework | ✅ Pass | 2.334s |  |
| Web Application Security Scanner | ✅ Pass | 10.475s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.387s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.223s

---

### installer (devstral-2:123b-cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.698s |  |
| Text Transform Uppercase | ✅ Pass | 0.738s |  |
| Count from 1 to 5 | ✅ Pass | 0.793s |  |
| Math Calculation | ✅ Pass | 0.709s |  |
| Basic Echo Function | ✅ Pass | 0.842s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.028s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.740s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 158.484s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.955s |  |
| Search Query Function | ✅ Pass | 0.851s |  |
| Ask Advice Function | ✅ Pass | 1.002s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.813s |  |
| Basic Context Memory Test | ✅ Pass | 0.851s |  |
| Function Argument Memory Test | ✅ Pass | 0.757s |  |
| Function Response Memory Test | ✅ Pass | 0.761s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.261s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.730s |  |
| Penetration Testing Methodology | ✅ Pass | 2.598s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.323s |  |
| SQL Injection Attack Type | ✅ Pass | 1.026s |  |
| Penetration Testing Framework | ✅ Pass | 4.639s |  |
| Web Application Security Scanner | ✅ Pass | 1.504s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.016s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.049s

---

### pentester (qwen3-coder-next:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.654s |  |
| Text Transform Uppercase | ✅ Pass | 1.256s |  |
| Count from 1 to 5 | ✅ Pass | 0.716s |  |
| Math Calculation | ✅ Pass | 0.993s |  |
| Basic Echo Function | ✅ Pass | 0.798s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.657s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.658s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.956s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.019s |  |
| Search Query Function | ✅ Pass | 0.841s |  |
| Ask Advice Function | ✅ Pass | 8.525s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 7.950s |  |
| Basic Context Memory Test | ✅ Pass | 10.081s |  |
| Function Argument Memory Test | ✅ Pass | 0.815s |  |
| Function Response Memory Test | ✅ Pass | 0.675s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.659s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.804s |  |
| Penetration Testing Methodology | ✅ Pass | 1.960s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.314s |  |
| SQL Injection Attack Type | ✅ Pass | 1.424s |  |
| Penetration Testing Framework | ✅ Pass | 1.673s |  |
| Web Application Security Scanner | ✅ Pass | 1.148s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.646s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.227s

---

