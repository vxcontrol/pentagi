# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 15:31:43 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.6-27B-FP8 | false | 23/23 (100.00%) | 0.804s |
| simple_json | Qwen/Qwen3.6-27B-FP8 | false | 7/7 (100.00%) | 0.773s |
| primary_agent | DeepSeek-V4-Flash | true | 22/23 (95.65%) | 1.232s |
| assistant | DeepSeek-V4-Flash | true | 23/23 (100.00%) | 1.231s |
| generator | DeepSeek-V4-Flash | true | 23/23 (100.00%) | 1.105s |
| refiner | DeepSeek-V4-Flash | true | 23/23 (100.00%) | 1.070s |
| adviser | DeepSeek-V4-Flash | true | 23/23 (100.00%) | 0.912s |
| reflector | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 0.875s |
| searcher | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 0.898s |
| enricher | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 0.929s |
| coder | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 3.311s |
| installer | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 3.519s |
| pentester | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 2.947s |

**Total**: 282/283 (99.65%) successful tests
**Overall average latency**: 1.550s

## Detailed Results

### simple (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.325s |  |
| Text Transform Uppercase | ✅ Pass | 0.389s |  |
| Count from 1 to 5 | ✅ Pass | 0.464s |  |
| Math Calculation | ✅ Pass | 0.360s |  |
| Basic Echo Function | ✅ Pass | 0.664s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.334s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.395s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.662s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.800s |  |
| Search Query Function | ✅ Pass | 0.593s |  |
| Ask Advice Function | ✅ Pass | 0.863s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.589s |  |
| Basic Context Memory Test | ✅ Pass | 0.482s |  |
| Function Argument Memory Test | ✅ Pass | 0.392s |  |
| Function Response Memory Test | ✅ Pass | 0.409s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.202s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.392s |  |
| Penetration Testing Methodology | ✅ Pass | 2.327s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.898s |  |
| SQL Injection Attack Type | ✅ Pass | 0.368s |  |
| Penetration Testing Framework | ✅ Pass | 2.080s |  |
| Web Application Security Scanner | ✅ Pass | 1.688s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.803s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.804s

---

### simple_json (Qwen/Qwen3.6-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.488s |  |
| Project Information JSON | ✅ Pass | 0.664s |  |
| User Profile JSON | ✅ Pass | 0.518s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.415s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.715s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.123s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.484s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 0.773s

---

### primary_agent (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.718s |  |
| Text Transform Uppercase | ✅ Pass | 0.878s |  |
| Count from 1 to 5 | ✅ Pass | 1.240s |  |
| Math Calculation | ✅ Pass | 0.766s |  |
| Basic Echo Function | ✅ Pass | 0.896s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.952s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.039s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.878s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.380s |  |
| Search Query Function | ✅ Pass | 1.391s |  |
| Ask Advice Function | ✅ Pass | 1.405s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.228s |  |
| Basic Context Memory Test | ✅ Pass | 1.470s |  |
| Function Argument Memory Test | ✅ Pass | 0.992s |  |
| Function Response Memory Test | ✅ Pass | 0.809s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.927s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.175s |  |
| Penetration Testing Methodology | ✅ Pass | 1.148s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.327s |  |
| SQL Injection Attack Type | ✅ Pass | 2.227s |  |
| Penetration Testing Framework | ✅ Pass | 1.028s |  |
| Web Application Security Scanner | ✅ Pass | 1.359s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.089s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.232s

---

### assistant (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.751s |  |
| Text Transform Uppercase | ✅ Pass | 0.844s |  |
| Count from 1 to 5 | ✅ Pass | 1.218s |  |
| Math Calculation | ✅ Pass | 0.758s |  |
| Basic Echo Function | ✅ Pass | 1.046s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.964s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.124s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.855s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.338s |  |
| Search Query Function | ✅ Pass | 1.337s |  |
| Ask Advice Function | ✅ Pass | 1.452s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.184s |  |
| Basic Context Memory Test | ✅ Pass | 1.431s |  |
| Function Argument Memory Test | ✅ Pass | 0.893s |  |
| Function Response Memory Test | ✅ Pass | 0.986s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.093s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.954s |  |
| Penetration Testing Methodology | ✅ Pass | 1.423s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.987s |  |
| SQL Injection Attack Type | ✅ Pass | 1.466s |  |
| Penetration Testing Framework | ✅ Pass | 1.117s |  |
| Web Application Security Scanner | ✅ Pass | 1.026s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.043s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.231s

---

### generator (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.751s |  |
| Text Transform Uppercase | ✅ Pass | 0.892s |  |
| Count from 1 to 5 | ✅ Pass | 1.239s |  |
| Math Calculation | ✅ Pass | 0.529s |  |
| Basic Echo Function | ✅ Pass | 0.931s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.930s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.001s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.803s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.227s |  |
| Search Query Function | ✅ Pass | 1.143s |  |
| Ask Advice Function | ✅ Pass | 1.388s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.190s |  |
| Basic Context Memory Test | ✅ Pass | 1.301s |  |
| Function Argument Memory Test | ✅ Pass | 0.814s |  |
| Function Response Memory Test | ✅ Pass | 0.751s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.916s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.925s |  |
| Penetration Testing Methodology | ✅ Pass | 1.189s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.771s |  |
| SQL Injection Attack Type | ✅ Pass | 1.098s |  |
| Penetration Testing Framework | ✅ Pass | 1.380s |  |
| Web Application Security Scanner | ✅ Pass | 0.901s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.345s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.105s

---

### refiner (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.751s |  |
| Text Transform Uppercase | ✅ Pass | 0.830s |  |
| Count from 1 to 5 | ✅ Pass | 1.033s |  |
| Math Calculation | ✅ Pass | 0.712s |  |
| Basic Echo Function | ✅ Pass | 0.933s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.859s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.916s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.925s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.245s |  |
| Search Query Function | ✅ Pass | 1.121s |  |
| Ask Advice Function | ✅ Pass | 1.264s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.069s |  |
| Basic Context Memory Test | ✅ Pass | 1.078s |  |
| Function Argument Memory Test | ✅ Pass | 0.742s |  |
| Function Response Memory Test | ✅ Pass | 0.770s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.902s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.831s |  |
| Penetration Testing Methodology | ✅ Pass | 1.419s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.832s |  |
| SQL Injection Attack Type | ✅ Pass | 0.764s |  |
| Penetration Testing Framework | ✅ Pass | 1.384s |  |
| Web Application Security Scanner | ✅ Pass | 0.969s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.244s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.070s

---

### adviser (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.679s |  |
| Text Transform Uppercase | ✅ Pass | 0.630s |  |
| Count from 1 to 5 | ✅ Pass | 0.857s |  |
| Math Calculation | ✅ Pass | 0.643s |  |
| Basic Echo Function | ✅ Pass | 0.729s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.544s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.801s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.858s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.020s |  |
| Search Query Function | ✅ Pass | 1.030s |  |
| Ask Advice Function | ✅ Pass | 1.093s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.834s |  |
| Basic Context Memory Test | ✅ Pass | 0.950s |  |
| Function Argument Memory Test | ✅ Pass | 0.713s |  |
| Function Response Memory Test | ✅ Pass | 0.678s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.634s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.687s |  |
| Penetration Testing Methodology | ✅ Pass | 1.078s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.025s |  |
| SQL Injection Attack Type | ✅ Pass | 0.666s |  |
| Penetration Testing Framework | ✅ Pass | 1.095s |  |
| Web Application Security Scanner | ✅ Pass | 0.852s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.876s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.912s

---

### reflector (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.325s |  |
| Text Transform Uppercase | ✅ Pass | 0.362s |  |
| Count from 1 to 5 | ✅ Pass | 0.503s |  |
| Math Calculation | ✅ Pass | 0.357s |  |
| Basic Echo Function | ✅ Pass | 0.850s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.337s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.433s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.753s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.869s |  |
| Search Query Function | ✅ Pass | 0.975s |  |
| Ask Advice Function | ✅ Pass | 0.934s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.085s |  |
| Basic Context Memory Test | ✅ Pass | 0.444s |  |
| Function Argument Memory Test | ✅ Pass | 0.441s |  |
| Function Response Memory Test | ✅ Pass | 0.382s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.539s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.513s |  |
| Penetration Testing Methodology | ✅ Pass | 2.543s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.088s |  |
| SQL Injection Attack Type | ✅ Pass | 0.346s |  |
| Penetration Testing Framework | ✅ Pass | 1.457s |  |
| Web Application Security Scanner | ✅ Pass | 1.632s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.940s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.875s

---

### searcher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.242s |  |
| Text Transform Uppercase | ✅ Pass | 0.523s |  |
| Count from 1 to 5 | ✅ Pass | 0.511s |  |
| Math Calculation | ✅ Pass | 0.356s |  |
| Basic Echo Function | ✅ Pass | 0.847s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.415s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.474s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.922s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.955s |  |
| Search Query Function | ✅ Pass | 0.924s |  |
| Ask Advice Function | ✅ Pass | 0.947s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.082s |  |
| Basic Context Memory Test | ✅ Pass | 0.439s |  |
| Function Argument Memory Test | ✅ Pass | 0.473s |  |
| Function Response Memory Test | ✅ Pass | 0.406s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.599s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.480s |  |
| Penetration Testing Methodology | ✅ Pass | 2.319s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.374s |  |
| SQL Injection Attack Type | ✅ Pass | 0.371s |  |
| Penetration Testing Framework | ✅ Pass | 1.580s |  |
| Web Application Security Scanner | ✅ Pass | 1.439s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.956s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.898s

---

### enricher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.158s |  |
| Text Transform Uppercase | ✅ Pass | 0.572s |  |
| Count from 1 to 5 | ✅ Pass | 0.489s |  |
| Math Calculation | ✅ Pass | 0.355s |  |
| Basic Echo Function | ✅ Pass | 0.844s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.413s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.461s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.749s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.951s |  |
| Search Query Function | ✅ Pass | 0.857s |  |
| Ask Advice Function | ✅ Pass | 0.876s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.056s |  |
| Basic Context Memory Test | ✅ Pass | 0.462s |  |
| Function Argument Memory Test | ✅ Pass | 0.495s |  |
| Function Response Memory Test | ✅ Pass | 0.407s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.392s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.461s |  |
| Penetration Testing Methodology | ✅ Pass | 2.876s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.880s |  |
| SQL Injection Attack Type | ✅ Pass | 0.361s |  |
| Penetration Testing Framework | ✅ Pass | 1.738s |  |
| Web Application Security Scanner | ✅ Pass | 1.509s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.996s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.929s

---

### coder (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.082s |  |
| Text Transform Uppercase | ✅ Pass | 2.219s |  |
| Count from 1 to 5 | ✅ Pass | 2.938s |  |
| Math Calculation | ✅ Pass | 2.041s |  |
| Basic Echo Function | ✅ Pass | 1.559s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.969s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.695s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.369s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.001s |  |
| Search Query Function | ✅ Pass | 1.574s |  |
| Ask Advice Function | ✅ Pass | 2.571s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.505s |  |
| Basic Context Memory Test | ✅ Pass | 2.486s |  |
| Function Argument Memory Test | ✅ Pass | 1.336s |  |
| Function Response Memory Test | ✅ Pass | 1.431s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.872s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.931s |  |
| Penetration Testing Methodology | ✅ Pass | 7.704s |  |
| SQL Injection Attack Type | ✅ Pass | 2.423s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.358s |  |
| Penetration Testing Framework | ✅ Pass | 7.475s |  |
| Web Application Security Scanner | ✅ Pass | 8.320s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.290s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.311s

---

### installer (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.807s |  |
| Text Transform Uppercase | ✅ Pass | 4.032s |  |
| Math Calculation | ✅ Pass | 1.855s |  |
| Basic Echo Function | ✅ Pass | 2.188s |  |
| Count from 1 to 5 | ✅ Pass | 6.222s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.774s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.077s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.594s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.802s |  |
| Search Query Function | ✅ Pass | 1.382s |  |
| Ask Advice Function | ✅ Pass | 2.326s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.422s |  |
| Basic Context Memory Test | ✅ Pass | 2.410s |  |
| Function Argument Memory Test | ✅ Pass | 1.547s |  |
| Function Response Memory Test | ✅ Pass | 1.870s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.757s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.061s |  |
| Penetration Testing Methodology | ✅ Pass | 6.035s |  |
| SQL Injection Attack Type | ✅ Pass | 3.266s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.942s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.049s |  |
| Web Application Security Scanner | ✅ Pass | 7.245s |  |
| Penetration Testing Framework | ✅ Pass | 12.260s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.519s

---

### pentester (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.800s |  |
| Text Transform Uppercase | ✅ Pass | 2.974s |  |
| Math Calculation | ✅ Pass | 1.781s |  |
| Count from 1 to 5 | ✅ Pass | 3.352s |  |
| Basic Echo Function | ✅ Pass | 1.788s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.382s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.549s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.524s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.694s |  |
| Search Query Function | ✅ Pass | 1.305s |  |
| Ask Advice Function | ✅ Pass | 3.077s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.206s |  |
| Function Argument Memory Test | ✅ Pass | 1.352s |  |
| Basic Context Memory Test | ✅ Pass | 3.221s |  |
| Function Response Memory Test | ✅ Pass | 1.188s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.963s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.485s |  |
| Penetration Testing Methodology | ✅ Pass | 6.373s |  |
| SQL Injection Attack Type | ✅ Pass | 3.415s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.750s |  |
| Penetration Testing Framework | ✅ Pass | 7.276s |  |
| Web Application Security Scanner | ✅ Pass | 6.321s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.003s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.947s

---

