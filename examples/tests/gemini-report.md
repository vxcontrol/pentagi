# LLM Agent Testing Report

Generated: Fri, 20 Feb 2026 14:59:29 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.5-flash-lite | false | 22/23 (95.65%) | 0.815s |
| simple_json | gemini-2.5-flash-lite | false | 5/5 (100.00%) | 0.861s |
| primary_agent | gemini-3-pro-preview | true | 23/23 (100.00%) | 9.396s |
| assistant | gemini-3-pro-preview | true | 23/23 (100.00%) | 10.774s |
| generator | gemini-3-pro-preview | true | 23/23 (100.00%) | 9.899s |
| refiner | gemini-3-pro-preview | true | 23/23 (100.00%) | 11.760s |
| adviser | gemini-3-pro-preview | true | 23/23 (100.00%) | 15.812s |
| reflector | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.086s |
| searcher | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.097s |
| enricher | gemini-3-flash-preview | true | 23/23 (100.00%) | 1.980s |
| coder | gemini-3-pro-preview | true | 23/23 (100.00%) | 9.836s |
| installer | gemini-3-flash-preview | true | 23/23 (100.00%) | 2.691s |
| pentester | gemini-3-pro-preview | true | 23/23 (100.00%) | 11.092s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 7.238s

## Detailed Results

### simple (gemini-2.5-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.271s |  |
| Text Transform Uppercase | ✅ Pass | 1.169s |  |
| Count from 1 to 5 | ✅ Pass | 0.737s |  |
| Math Calculation | ✅ Pass | 1.161s |  |
| Basic Echo Function | ✅ Pass | 0.744s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.481s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.427s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.776s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.545s |  |
| Search Query Function | ✅ Pass | 0.529s |  |
| Ask Advice Function | ✅ Pass | 0.592s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.690s |  |
| Basic Context Memory Test | ✅ Pass | 0.558s |  |
| Function Argument Memory Test | ✅ Pass | 1.009s |  |
| Function Response Memory Test | ✅ Pass | 0.333s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.531s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.479s |  |
| Penetration Testing Methodology | ✅ Pass | 2.316s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.248s |  |
| SQL Injection Attack Type | ✅ Pass | 1.044s |  |
| Penetration Testing Framework | ✅ Pass | 0.734s |  |
| Web Application Security Scanner | ✅ Pass | 0.831s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.527s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.815s

---

### simple_json (gemini-2.5-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.822s |  |
| Person Information JSON | ✅ Pass | 0.649s |  |
| Project Information JSON | ✅ Pass | 1.630s |  |
| User Profile JSON | ✅ Pass | 0.669s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.530s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.861s

---

### primary_agent (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.902s |  |
| Text Transform Uppercase | ✅ Pass | 6.076s |  |
| Count from 1 to 5 | ✅ Pass | 10.019s |  |
| Math Calculation | ✅ Pass | 6.523s |  |
| Basic Echo Function | ✅ Pass | 3.818s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.880s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.555s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.905s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 9.266s |  |
| Search Query Function | ✅ Pass | 15.721s |  |
| Ask Advice Function | ✅ Pass | 4.289s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.877s |  |
| Basic Context Memory Test | ✅ Pass | 7.126s |  |
| Function Argument Memory Test | ✅ Pass | 19.393s |  |
| Function Response Memory Test | ✅ Pass | 9.960s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 10.556s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.005s |  |
| Penetration Testing Methodology | ✅ Pass | 10.327s |  |
| Vulnerability Assessment Tools | ✅ Pass | 24.206s |  |
| SQL Injection Attack Type | ✅ Pass | 6.942s |  |
| Penetration Testing Framework | ✅ Pass | 13.874s |  |
| Web Application Security Scanner | ✅ Pass | 18.398s |  |
| Penetration Testing Tool Selection | ✅ Pass | 12.471s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.396s

---

### assistant (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.951s |  |
| Text Transform Uppercase | ✅ Pass | 6.829s |  |
| Count from 1 to 5 | ✅ Pass | 6.261s |  |
| Math Calculation | ✅ Pass | 4.240s |  |
| Basic Echo Function | ✅ Pass | 3.717s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.015s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 7.879s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 16.386s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.551s |  |
| Search Query Function | ✅ Pass | 10.526s |  |
| Ask Advice Function | ✅ Pass | 5.244s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 9.083s |  |
| Basic Context Memory Test | ✅ Pass | 9.976s |  |
| Function Argument Memory Test | ✅ Pass | 8.237s |  |
| Function Response Memory Test | ✅ Pass | 7.055s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 12.186s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.665s |  |
| Penetration Testing Methodology | ✅ Pass | 21.798s |  |
| Vulnerability Assessment Tools | ✅ Pass | 34.321s |  |
| SQL Injection Attack Type | ✅ Pass | 19.365s |  |
| Penetration Testing Framework | ✅ Pass | 20.504s |  |
| Web Application Security Scanner | ✅ Pass | 15.356s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.639s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 10.774s

---

### generator (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.688s |  |
| Text Transform Uppercase | ✅ Pass | 4.714s |  |
| Count from 1 to 5 | ✅ Pass | 5.390s |  |
| Math Calculation | ✅ Pass | 9.395s |  |
| Basic Echo Function | ✅ Pass | 3.503s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.969s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.489s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.474s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.609s |  |
| Search Query Function | ✅ Pass | 6.792s |  |
| Ask Advice Function | ✅ Pass | 3.832s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.789s |  |
| Basic Context Memory Test | ✅ Pass | 6.405s |  |
| Function Argument Memory Test | ✅ Pass | 5.530s |  |
| Function Response Memory Test | ✅ Pass | 10.602s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.110s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 17.661s |  |
| Penetration Testing Methodology | ✅ Pass | 21.284s |  |
| Vulnerability Assessment Tools | ✅ Pass | 38.065s |  |
| SQL Injection Attack Type | ✅ Pass | 12.751s |  |
| Penetration Testing Framework | ✅ Pass | 15.421s |  |
| Web Application Security Scanner | ✅ Pass | 16.169s |  |
| Penetration Testing Tool Selection | ✅ Pass | 7.017s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.899s

---

### refiner (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.732s |  |
| Text Transform Uppercase | ✅ Pass | 8.816s |  |
| Count from 1 to 5 | ✅ Pass | 5.274s |  |
| Math Calculation | ✅ Pass | 6.910s |  |
| Basic Echo Function | ✅ Pass | 5.730s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.639s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.999s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.241s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.302s |  |
| Search Query Function | ✅ Pass | 4.317s |  |
| Ask Advice Function | ✅ Pass | 19.872s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.532s |  |
| Basic Context Memory Test | ✅ Pass | 7.685s |  |
| Function Argument Memory Test | ✅ Pass | 19.732s |  |
| Function Response Memory Test | ✅ Pass | 14.603s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.238s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 15.205s |  |
| Penetration Testing Methodology | ✅ Pass | 25.579s |  |
| Vulnerability Assessment Tools | ✅ Pass | 38.521s |  |
| SQL Injection Attack Type | ✅ Pass | 25.126s |  |
| Penetration Testing Framework | ✅ Pass | 13.120s |  |
| Web Application Security Scanner | ✅ Pass | 14.271s |  |
| Penetration Testing Tool Selection | ✅ Pass | 6.034s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 11.760s

---

### adviser (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.533s |  |
| Text Transform Uppercase | ✅ Pass | 14.028s |  |
| Count from 1 to 5 | ✅ Pass | 14.150s |  |
| Math Calculation | ✅ Pass | 7.374s |  |
| Basic Echo Function | ✅ Pass | 3.790s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.212s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.613s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.052s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 11.409s |  |
| Search Query Function | ✅ Pass | 4.777s |  |
| Ask Advice Function | ✅ Pass | 3.979s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.465s |  |
| Basic Context Memory Test | ✅ Pass | 6.198s |  |
| Function Argument Memory Test | ✅ Pass | 12.685s |  |
| Function Response Memory Test | ✅ Pass | 12.466s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.216s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.396s |  |
| Penetration Testing Methodology | ✅ Pass | 136.282s |  |
| Vulnerability Assessment Tools | ✅ Pass | 54.086s |  |
| SQL Injection Attack Type | ✅ Pass | 5.331s |  |
| Penetration Testing Framework | ✅ Pass | 19.267s |  |
| Web Application Security Scanner | ✅ Pass | 14.030s |  |
| Penetration Testing Tool Selection | ✅ Pass | 7.327s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 15.812s

---

### reflector (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.473s |  |
| Text Transform Uppercase | ✅ Pass | 1.471s |  |
| Count from 1 to 5 | ✅ Pass | 2.033s |  |
| Math Calculation | ✅ Pass | 1.304s |  |
| Basic Echo Function | ✅ Pass | 1.491s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.541s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.523s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.506s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.913s |  |
| Search Query Function | ✅ Pass | 1.711s |  |
| Ask Advice Function | ✅ Pass | 1.334s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.044s |  |
| Basic Context Memory Test | ✅ Pass | 1.560s |  |
| Function Argument Memory Test | ✅ Pass | 1.489s |  |
| Function Response Memory Test | ✅ Pass | 1.366s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.135s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.781s |  |
| Penetration Testing Methodology | ✅ Pass | 3.366s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.402s |  |
| SQL Injection Attack Type | ✅ Pass | 1.743s |  |
| Penetration Testing Framework | ✅ Pass | 3.847s |  |
| Web Application Security Scanner | ✅ Pass | 3.086s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.858s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.086s

---

### searcher (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.274s |  |
| Text Transform Uppercase | ✅ Pass | 1.602s |  |
| Count from 1 to 5 | ✅ Pass | 1.716s |  |
| Math Calculation | ✅ Pass | 2.345s |  |
| Basic Echo Function | ✅ Pass | 3.771s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.451s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.635s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.758s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.225s |  |
| Search Query Function | ✅ Pass | 1.191s |  |
| Ask Advice Function | ✅ Pass | 1.585s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.322s |  |
| Basic Context Memory Test | ✅ Pass | 1.296s |  |
| Function Argument Memory Test | ✅ Pass | 1.783s |  |
| Function Response Memory Test | ✅ Pass | 1.378s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.577s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.474s |  |
| Penetration Testing Methodology | ✅ Pass | 3.128s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.240s |  |
| SQL Injection Attack Type | ✅ Pass | 1.741s |  |
| Penetration Testing Framework | ✅ Pass | 2.960s |  |
| Web Application Security Scanner | ✅ Pass | 2.720s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.052s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.097s

---

### enricher (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.441s |  |
| Text Transform Uppercase | ✅ Pass | 1.158s |  |
| Count from 1 to 5 | ✅ Pass | 1.762s |  |
| Math Calculation | ✅ Pass | 1.307s |  |
| Basic Echo Function | ✅ Pass | 1.469s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.106s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.006s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.699s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.519s |  |
| Search Query Function | ✅ Pass | 1.387s |  |
| Ask Advice Function | ✅ Pass | 1.554s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.748s |  |
| Basic Context Memory Test | ✅ Pass | 1.542s |  |
| Function Argument Memory Test | ✅ Pass | 1.693s |  |
| Function Response Memory Test | ✅ Pass | 1.428s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.572s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.856s |  |
| Penetration Testing Methodology | ✅ Pass | 3.163s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.909s |  |
| SQL Injection Attack Type | ✅ Pass | 1.810s |  |
| Penetration Testing Framework | ✅ Pass | 3.942s |  |
| Web Application Security Scanner | ✅ Pass | 2.801s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.649s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.980s

---

### coder (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.675s |  |
| Text Transform Uppercase | ✅ Pass | 6.282s |  |
| Count from 1 to 5 | ✅ Pass | 8.972s |  |
| Math Calculation | ✅ Pass | 6.938s |  |
| Basic Echo Function | ✅ Pass | 5.178s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.589s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.106s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.562s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 19.068s |  |
| Search Query Function | ✅ Pass | 4.305s |  |
| Ask Advice Function | ✅ Pass | 10.809s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.311s |  |
| Basic Context Memory Test | ✅ Pass | 7.530s |  |
| Function Argument Memory Test | ✅ Pass | 11.571s |  |
| Function Response Memory Test | ✅ Pass | 4.128s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 21.307s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 11.266s |  |
| Penetration Testing Methodology | ✅ Pass | 17.362s |  |
| Vulnerability Assessment Tools | ✅ Pass | 23.019s |  |
| SQL Injection Attack Type | ✅ Pass | 14.319s |  |
| Penetration Testing Framework | ✅ Pass | 16.302s |  |
| Web Application Security Scanner | ✅ Pass | 11.089s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.520s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.836s

---

### installer (gemini-3-flash-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.172s |  |
| Text Transform Uppercase | ✅ Pass | 2.966s |  |
| Count from 1 to 5 | ✅ Pass | 3.724s |  |
| Math Calculation | ✅ Pass | 2.367s |  |
| Basic Echo Function | ✅ Pass | 1.233s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.199s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.735s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.303s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.436s |  |
| Search Query Function | ✅ Pass | 1.233s |  |
| Ask Advice Function | ✅ Pass | 1.209s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.955s |  |
| Basic Context Memory Test | ✅ Pass | 2.589s |  |
| Function Argument Memory Test | ✅ Pass | 0.954s |  |
| Function Response Memory Test | ✅ Pass | 1.172s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.842s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.480s |  |
| Penetration Testing Methodology | ✅ Pass | 5.546s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.478s |  |
| SQL Injection Attack Type | ✅ Pass | 2.437s |  |
| Penetration Testing Framework | ✅ Pass | 5.048s |  |
| Web Application Security Scanner | ✅ Pass | 5.197s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.599s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.691s

---

### pentester (gemini-3-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.558s |  |
| Text Transform Uppercase | ✅ Pass | 4.384s |  |
| Count from 1 to 5 | ✅ Pass | 10.981s |  |
| Math Calculation | ✅ Pass | 4.330s |  |
| Basic Echo Function | ✅ Pass | 4.095s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.161s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.698s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.388s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 14.763s |  |
| Search Query Function | ✅ Pass | 3.320s |  |
| Ask Advice Function | ✅ Pass | 5.861s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 11.046s |  |
| Basic Context Memory Test | ✅ Pass | 7.003s |  |
| Function Argument Memory Test | ✅ Pass | 8.292s |  |
| Function Response Memory Test | ✅ Pass | 5.524s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.609s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 10.527s |  |
| Penetration Testing Methodology | ✅ Pass | 31.041s |  |
| Vulnerability Assessment Tools | ✅ Pass | 45.351s |  |
| SQL Injection Attack Type | ✅ Pass | 13.882s |  |
| Penetration Testing Framework | ✅ Pass | 20.377s |  |
| Web Application Security Scanner | ✅ Pass | 15.410s |  |
| Penetration Testing Tool Selection | ✅ Pass | 13.497s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 11.092s

---

