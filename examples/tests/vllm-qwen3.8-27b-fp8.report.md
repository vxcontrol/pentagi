# LLM Agent Testing Report

Generated: Tue, 01 Sep 2026 23:37:44 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 1.785s |
| simple_json | Qwen/Qwen3.8-27B-FP8 | false | 7/7 (100.00%) | 1.118s |
| primary_agent | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 2.237s |
| assistant | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 2.523s |
| generator | Qwen/Qwen3.8-27B-FP8 | true | 23/24 (95.83%) | 2.683s |
| refiner | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 2.410s |
| adviser | Qwen/Qwen3.8-27B-FP8 | true | 23/24 (95.83%) | 2.376s |
| reflector | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 1.898s |
| searcher | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 1.802s |
| enricher | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 1.897s |
| coder | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 2.011s |
| installer | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 2.063s |
| pentester | Qwen/Qwen3.8-27B-FP8 | true | 24/24 (100.00%) | 2.354s |

**Total**: 293/295 (99.32%) successful tests
**Overall average latency**: 2.145s

## Detailed Results

### simple (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Math Calculation | ✅ Pass | 0.675s |  |
| Simple Math | ✅ Pass | 0.693s |  |
| Count from 1 to 5 | ✅ Pass | 0.781s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.634s |  |
| Basic Echo Function | ✅ Pass | 0.913s |  |
| Text Transform Uppercase | ✅ Pass | 2.412s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.155s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.483s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.943s |  |
| Search Query Function | ✅ Pass | 0.978s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.034s |  |
| Ask Advice Function | ✅ Pass | 1.539s |  |
| Function Response Memory Test | ✅ Pass | 0.548s |  |
| Basic Context Memory Test | ✅ Pass | 2.058s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.826s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.961s |  |
| Function Argument Memory Test | ✅ Pass | 3.931s |  |
| SQL Injection Attack Type | ✅ Pass | 0.437s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.282s |  |
| Penetration Testing Methodology | ✅ Pass | 2.616s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.596s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.094s |  |
| Web Application Security Scanner | ✅ Pass | 1.955s |  |
| Penetration Testing Framework | ✅ Pass | 4.273s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.785s

---

### simple_json (Qwen/Qwen3.8-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.692s |  |
| Person Information JSON | ✅ Pass | 0.966s |  |
| User Profile JSON | ✅ Pass | 0.909s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.870s |  |
| Project Information JSON | ✅ Pass | 1.149s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.343s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.892s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 1.118s

---

### primary_agent (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Count from 1 to 5 | ✅ Pass | 1.328s |  |
| Text Transform Uppercase | ✅ Pass | 1.328s |  |
| Simple Math | ✅ Pass | 1.349s |  |
| Math Calculation | ✅ Pass | 1.415s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.090s |  |
| Basic Echo Function | ✅ Pass | 1.235s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.622s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.944s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 1.443s |  |
| Ask Advice Function | ✅ Pass | 1.799s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.535s |  |
| JSON Response Function | ✅ Pass | 3.751s |  |
| Function Argument Memory Test | ✅ Pass | 1.457s |  |
| Basic Context Memory Test | ✅ Pass | 2.154s |  |
| Function Response Memory Test | ✅ Pass | 1.319s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.397s |  |
| Penetration Testing Methodology | ✅ Pass | 0.933s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.765s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.567s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.546s |  |
| SQL Injection Attack Type | ✅ Pass | 2.476s |  |
| Web Application Security Scanner | ✅ Pass | 1.527s |  |
| Penetration Testing Framework | ✅ Pass | 3.435s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.271s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.237s

---

### assistant (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.332s |  |
| Math Calculation | ✅ Pass | 1.334s |  |
| Basic Echo Function | ✅ Pass | 1.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.156s |  |
| Text Transform Uppercase | ✅ Pass | 2.722s |  |
| Count from 1 to 5 | ✅ Pass | 3.416s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.448s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.530s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 1.581s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.297s |  |
| JSON Response Function | ✅ Pass | 3.869s |  |
| Ask Advice Function | ✅ Pass | 2.206s |  |
| Basic Context Memory Test | ✅ Pass | 2.121s |  |
| Function Response Memory Test | ✅ Pass | 1.431s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.907s |  |
| Function Argument Memory Test | ✅ Pass | 4.516s |  |
| Penetration Testing Methodology | ✅ Pass | 1.172s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.219s |  |
| SQL Injection Attack Type | ✅ Pass | 1.288s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 6.439s |  |
| Penetration Testing Framework | ✅ Pass | 2.035s |  |
| Web Application Security Scanner | ✅ Pass | 1.611s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.740s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.979s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.523s

---

### generator (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 1.330s |  |
| Math Calculation | ✅ Pass | 1.334s |  |
| Count from 1 to 5 | ✅ Pass | 1.364s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.084s |  |
| Basic Echo Function | ✅ Pass | 1.225s |  |
| Simple Math | ✅ Pass | 2.721s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.528s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.464s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 1.433s |  |
| JSON Response Function | ✅ Pass | 3.961s |  |
| Function Argument Memory Test | ✅ Pass | 1.427s |  |
| Ask Advice Function | ✅ Pass | 3.155s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.880s |  |
| Function Response Memory Test | ✅ Pass | 3.027s |  |
| Basic Context Memory Test | ✅ Pass | 4.700s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.436s |  |
| Penetration Testing Methodology | ✅ Pass | 1.045s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.671s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 6.220s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| SQL Injection Attack Type | ✅ Pass | 2.412s |  |
| Penetration Testing Framework | ✅ Pass | 1.787s |  |
| Web Application Security Scanner | ✅ Pass | 1.883s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.640s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.663s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 2.683s

---

### refiner (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 1.397s |  |
| Count from 1 to 5 | ✅ Pass | 1.534s |  |
| Simple Math | ✅ Pass | 2.358s |  |
| Basic Echo Function | ✅ Pass | 1.260s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.721s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.234s |  |
| Math Calculation | ✅ Pass | 4.525s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.785s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.654s |  |
| Search Query Function | ✅ Pass | 1.606s |  |
| Ask Advice Function | ✅ Pass | 2.278s |  |
| Function Argument Memory Test | ✅ Pass | 1.310s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.535s |  |
| Basic Context Memory Test | ✅ Pass | 2.234s |  |
| Function Response Memory Test | ✅ Pass | 1.151s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.745s |  |
| Penetration Testing Methodology | ✅ Pass | 1.788s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.988s |  |
| SQL Injection Attack Type | ✅ Pass | 1.054s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.005s |  |
| Penetration Testing Framework | ✅ Pass | 1.813s |  |
| Web Application Security Scanner | ✅ Pass | 1.433s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.758s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.658s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.410s

---

### adviser (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 1.088s |  |
| Simple Math | ✅ Pass | 1.466s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.894s |  |
| Count from 1 to 5 | ✅ Pass | 2.582s |  |
| Basic Echo Function | ✅ Pass | 1.406s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.589s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.498s |  |
| Math Calculation | ✅ Pass | 4.471s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 1.170s |  |
| JSON Response Function | ✅ Pass | 2.452s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.248s |  |
| Ask Advice Function | ✅ Pass | 1.952s |  |
| Basic Context Memory Test | ✅ Pass | 1.899s |  |
| Function Response Memory Test | ✅ Pass | 1.051s |  |
| Function Argument Memory Test | ✅ Pass | 3.146s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.829s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 4.554s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Read a file, then edit it via unified diff | ✅ Pass | 4.938s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.113s |  |
| SQL Injection Attack Type | ✅ Pass | 2.245s |  |
| Web Application Security Scanner | ✅ Pass | 1.385s |  |
| Penetration Testing Framework | ✅ Pass | 2.082s |  |
| Penetration Testing Methodology | ✅ Pass | 6.063s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.891s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 2.376s

---

### reflector (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 0.619s |  |
| Simple Math | ✅ Pass | 1.328s |  |
| Basic Echo Function | ✅ Pass | 0.920s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.597s |  |
| Count from 1 to 5 | ✅ Pass | 3.171s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.870s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.120s |  |
| Math Calculation | ✅ Pass | 4.451s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.961s |  |
| Search Query Function | ✅ Pass | 1.089s |  |
| Basic Context Memory Test | ✅ Pass | 0.588s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.033s |  |
| Ask Advice Function | ✅ Pass | 1.458s |  |
| Function Argument Memory Test | ✅ Pass | 0.564s |  |
| Function Response Memory Test | ✅ Pass | 0.548s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.729s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.398s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.892s |  |
| SQL Injection Attack Type | ✅ Pass | 0.883s |  |
| Web Application Security Scanner | ✅ Pass | 0.453s |  |
| Penetration Testing Methodology | ✅ Pass | 4.558s |  |
| Penetration Testing Framework | ✅ Pass | 2.611s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.081s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.624s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.898s

---

### searcher (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Count from 1 to 5 | ✅ Pass | 1.328s |  |
| Text Transform Uppercase | ✅ Pass | 1.329s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.669s |  |
| Basic Echo Function | ✅ Pass | 1.010s |  |
| Simple Math | ✅ Pass | 3.172s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.781s |  |
| Math Calculation | ✅ Pass | 4.365s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.054s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.179s |  |
| Search Query Function | ✅ Pass | 0.745s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.943s |  |
| Basic Context Memory Test | ✅ Pass | 0.563s |  |
| Ask Advice Function | ✅ Pass | 1.527s |  |
| Function Argument Memory Test | ✅ Pass | 0.629s |  |
| Function Response Memory Test | ✅ Pass | 0.546s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.816s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.085s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.082s |  |
| SQL Injection Attack Type | ✅ Pass | 0.832s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.904s |  |
| Penetration Testing Methodology | ✅ Pass | 4.750s |  |
| Web Application Security Scanner | ✅ Pass | 1.616s |  |
| Penetration Testing Framework | ✅ Pass | 2.233s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.082s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.802s

---

### enricher (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Count from 1 to 5 | ✅ Pass | 1.065s |  |
| Text Transform Uppercase | ✅ Pass | 1.329s |  |
| Simple Math | ✅ Pass | 1.330s |  |
| Basic Echo Function | ✅ Pass | 0.991s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.425s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.542s |  |
| Math Calculation | ✅ Pass | 4.082s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.998s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.129s |  |
| Search Query Function | ✅ Pass | 1.139s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.890s |  |
| Basic Context Memory Test | ✅ Pass | 0.600s |  |
| Function Argument Memory Test | ✅ Pass | 0.544s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.560s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.222s |  |
| Function Response Memory Test | ✅ Pass | 2.842s |  |
| Ask Advice Function | ✅ Pass | 5.690s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.884s |  |
| SQL Injection Attack Type | ✅ Pass | 0.659s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.113s |  |
| Penetration Testing Methodology | ✅ Pass | 4.511s |  |
| Penetration Testing Framework | ✅ Pass | 2.157s |  |
| Web Application Security Scanner | ✅ Pass | 1.599s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.214s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.897s

---

### coder (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.404s |  |
| Count from 1 to 5 | ✅ Pass | 1.449s |  |
| Text Transform Uppercase | ✅ Pass | 2.357s |  |
| Basic Echo Function | ✅ Pass | 1.171s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.529s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.452s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.454s |  |
| Math Calculation | ✅ Pass | 4.059s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.662s |  |
| Search Query Function | ✅ Pass | 1.734s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.461s |  |
| Ask Advice Function | ✅ Pass | 1.834s |  |
| Basic Context Memory Test | ✅ Pass | 1.192s |  |
| Function Argument Memory Test | ✅ Pass | 1.309s |  |
| Function Response Memory Test | ✅ Pass | 1.084s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.879s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.429s |  |
| Penetration Testing Methodology | ✅ Pass | 1.896s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.157s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.604s |  |
| SQL Injection Attack Type | ✅ Pass | 1.394s |  |
| Penetration Testing Framework | ✅ Pass | 1.899s |  |
| Web Application Security Scanner | ✅ Pass | 1.226s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.624s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.011s

---

### installer (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.393s |  |
| Count from 1 to 5 | ✅ Pass | 1.496s |  |
| Text Transform Uppercase | ✅ Pass | 2.357s |  |
| Basic Echo Function | ✅ Pass | 1.316s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.327s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.171s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.592s |  |
| Math Calculation | ✅ Pass | 4.120s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.514s |  |
| Search Query Function | ✅ Pass | 1.360s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.297s |  |
| Ask Advice Function | ✅ Pass | 1.785s |  |
| Function Response Memory Test | ✅ Pass | 0.841s |  |
| Basic Context Memory Test | ✅ Pass | 1.683s |  |
| Function Argument Memory Test | ✅ Pass | 1.590s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.889s |  |
| Penetration Testing Methodology | ✅ Pass | 1.180s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.879s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.090s |  |
| SQL Injection Attack Type | ✅ Pass | 1.628s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.476s |  |
| Penetration Testing Framework | ✅ Pass | 1.645s |  |
| Web Application Security Scanner | ✅ Pass | 1.548s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.318s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.063s

---

### pentester (Qwen/Qwen3.8-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.328s |  |
| Count from 1 to 5 | ✅ Pass | 2.495s |  |
| Basic Echo Function | ✅ Pass | 1.184s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.231s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.407s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.194s |  |
| Text Transform Uppercase | ✅ Pass | 5.145s |  |
| Math Calculation | ✅ Pass | 4.120s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.439s |  |
| Search Query Function | ✅ Pass | 1.372s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.404s |  |
| Function Argument Memory Test | ✅ Pass | 1.428s |  |
| Ask Advice Function | ✅ Pass | 3.082s |  |
| Function Response Memory Test | ✅ Pass | 1.289s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.727s |  |
| Basic Context Memory Test | ✅ Pass | 4.196s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.987s |  |
| Penetration Testing Methodology | ✅ Pass | 1.242s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.143s |  |
| SQL Injection Attack Type | ✅ Pass | 1.953s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.018s |  |
| Penetration Testing Framework | ✅ Pass | 1.447s |  |
| Web Application Security Scanner | ✅ Pass | 1.384s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.277s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.354s

---

