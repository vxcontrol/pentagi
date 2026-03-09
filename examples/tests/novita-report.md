# LLM Agent Testing Report

Generated: Mon, 02 Mar 2026 15:08:50 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek/deepseek-v3.2 | false | 22/23 (95.65%) | 2.458s |
| simple_json | deepseek/deepseek-v3.2 | false | 5/5 (100.00%) | 2.148s |
| primary_agent | moonshotai/kimi-k2.5 | true | 22/23 (95.65%) | 2.658s |
| assistant | moonshotai/kimi-k2.5 | true | 22/23 (95.65%) | 3.286s |
| generator | moonshotai/kimi-k2.5 | true | 22/23 (95.65%) | 2.686s |
| refiner | moonshotai/kimi-k2.5 | true | 22/23 (95.65%) | 3.071s |
| adviser | zai-org/glm-5 | true | 23/23 (100.00%) | 9.204s |
| reflector | qwen/qwen3.5-35b-a3b | true | 23/23 (100.00%) | 3.375s |
| searcher | qwen/qwen3.5-35b-a3b | true | 22/23 (95.65%) | 3.648s |
| enricher | qwen/qwen3.5-35b-a3b | true | 23/23 (100.00%) | 3.332s |
| coder | moonshotai/kimi-k2.5 | true | 23/23 (100.00%) | 3.067s |
| installer | moonshotai/kimi-k2-instruct | true | 20/23 (86.96%) | 1.480s |
| pentester | moonshotai/kimi-k2.5 | true | 23/23 (100.00%) | 2.818s |

**Total**: 272/281 (96.80%) successful tests
**Overall average latency**: 3.401s

## Detailed Results

### simple (deepseek/deepseek-v3.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.288s |  |
| Text Transform Uppercase | ✅ Pass | 1.461s |  |
| Count from 1 to 5 | ✅ Pass | 1.353s |  |
| Math Calculation | ✅ Pass | 1.379s |  |
| Basic Echo Function | ✅ Pass | 2.869s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.182s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.654s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.997s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.244s |  |
| Search Query Function | ✅ Pass | 2.317s |  |
| Ask Advice Function | ✅ Pass | 3.651s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.320s |  |
| Basic Context Memory Test | ✅ Pass | 2.409s |  |
| Function Argument Memory Test | ✅ Pass | 1.077s |  |
| Function Response Memory Test | ✅ Pass | 1.354s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 4.833s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.674s |  |
| Penetration Testing Methodology | ✅ Pass | 2.582s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.695s |  |
| SQL Injection Attack Type | ✅ Pass | 1.572s |  |
| Penetration Testing Framework | ✅ Pass | 2.639s |  |
| Web Application Security Scanner | ✅ Pass | 1.808s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.174s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.458s

---

### simple_json (deepseek/deepseek-v3.2)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 2.759s |  |
| Person Information JSON | ✅ Pass | 1.929s |  |
| Project Information JSON | ✅ Pass | 1.792s |  |
| User Profile JSON | ✅ Pass | 2.154s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 2.102s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 2.148s

---

### primary_agent (moonshotai/kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.139s |  |
| Text Transform Uppercase | ✅ Pass | 0.784s |  |
| Count from 1 to 5 | ✅ Pass | 0.869s |  |
| Math Calculation | ✅ Pass | 1.179s |  |
| Basic Echo Function | ✅ Pass | 1.499s |  |
| Simple Math Streaming | ❌ Fail | 0.242s | API returned unexpected status code: 429: |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.832s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.742s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.169s |  |
| Search Query Function | ✅ Pass | 1.403s |  |
| Ask Advice Function | ✅ Pass | 1.781s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.215s |  |
| Basic Context Memory Test | ✅ Pass | 2.855s |  |
| Function Argument Memory Test | ✅ Pass | 1.847s |  |
| Function Response Memory Test | ✅ Pass | 2.417s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.549s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.638s |  |
| Penetration Testing Methodology | ✅ Pass | 3.428s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.630s |  |
| SQL Injection Attack Type | ✅ Pass | 1.428s |  |
| Penetration Testing Framework | ✅ Pass | 13.651s |  |
| Web Application Security Scanner | ✅ Pass | 2.220s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.612s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.658s

---

### assistant (moonshotai/kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.280s |  |
| Text Transform Uppercase | ✅ Pass | 0.778s |  |
| Count from 1 to 5 | ✅ Pass | 1.746s |  |
| Math Calculation | ✅ Pass | 1.138s |  |
| Basic Echo Function | ✅ Pass | 1.738s |  |
| Simple Math Streaming | ❌ Fail | 0.282s | API returned unexpected status code: 429: |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.009s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.181s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.819s |  |
| Search Query Function | ✅ Pass | 2.981s |  |
| Ask Advice Function | ✅ Pass | 1.336s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.939s |  |
| Basic Context Memory Test | ✅ Pass | 3.554s |  |
| Function Argument Memory Test | ✅ Pass | 6.030s |  |
| Function Response Memory Test | ✅ Pass | 2.626s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.105s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.154s |  |
| Penetration Testing Methodology | ✅ Pass | 4.826s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.738s |  |
| SQL Injection Attack Type | ✅ Pass | 3.703s |  |
| Penetration Testing Framework | ✅ Pass | 13.123s |  |
| Web Application Security Scanner | ✅ Pass | 1.985s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.496s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 3.286s

---

### generator (moonshotai/kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.148s |  |
| Text Transform Uppercase | ✅ Pass | 0.817s |  |
| Count from 1 to 5 | ✅ Pass | 0.931s |  |
| Math Calculation | ✅ Pass | 1.174s |  |
| Basic Echo Function | ✅ Pass | 1.500s |  |
| Simple Math Streaming | ❌ Fail | 0.243s | API returned unexpected status code: 429: |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.691s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.370s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.270s |  |
| Search Query Function | ✅ Pass | 1.396s |  |
| Ask Advice Function | ✅ Pass | 1.799s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.598s |  |
| Basic Context Memory Test | ✅ Pass | 1.747s |  |
| Function Argument Memory Test | ✅ Pass | 1.747s |  |
| Function Response Memory Test | ✅ Pass | 0.828s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.171s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.225s |  |
| Penetration Testing Methodology | ✅ Pass | 3.684s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.069s |  |
| SQL Injection Attack Type | ✅ Pass | 1.440s |  |
| Penetration Testing Framework | ✅ Pass | 17.198s |  |
| Web Application Security Scanner | ✅ Pass | 1.930s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.790s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.686s

---

### refiner (moonshotai/kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.052s |  |
| Text Transform Uppercase | ✅ Pass | 1.245s |  |
| Count from 1 to 5 | ✅ Pass | 2.057s |  |
| Math Calculation | ✅ Pass | 1.189s |  |
| Basic Echo Function | ✅ Pass | 1.490s |  |
| Simple Math Streaming | ❌ Fail | 0.267s | API returned unexpected status code: 429: |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.289s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.278s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.328s |  |
| Search Query Function | ✅ Pass | 1.366s |  |
| Ask Advice Function | ✅ Pass | 1.970s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.560s |  |
| Basic Context Memory Test | ✅ Pass | 2.548s |  |
| Function Argument Memory Test | ✅ Pass | 5.412s |  |
| Function Response Memory Test | ✅ Pass | 1.155s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.038s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.165s |  |
| Penetration Testing Methodology | ✅ Pass | 3.756s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.463s |  |
| SQL Injection Attack Type | ✅ Pass | 5.187s |  |
| Penetration Testing Framework | ✅ Pass | 6.667s |  |
| Web Application Security Scanner | ✅ Pass | 2.588s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.554s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 3.071s

---

### adviser (zai-org/glm-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.329s |  |
| Text Transform Uppercase | ✅ Pass | 1.953s |  |
| Count from 1 to 5 | ✅ Pass | 6.180s |  |
| Math Calculation | ✅ Pass | 6.157s |  |
| Basic Echo Function | ✅ Pass | 4.979s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.658s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.993s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.868s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 6.030s |  |
| Search Query Function | ✅ Pass | 4.583s |  |
| Ask Advice Function | ✅ Pass | 3.598s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.290s |  |
| Basic Context Memory Test | ✅ Pass | 2.464s |  |
| Function Argument Memory Test | ✅ Pass | 4.195s |  |
| Function Response Memory Test | ✅ Pass | 2.751s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.420s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.993s |  |
| Penetration Testing Methodology | ✅ Pass | 17.127s |  |
| Vulnerability Assessment Tools | ✅ Pass | 56.258s |  |
| SQL Injection Attack Type | ✅ Pass | 7.521s |  |
| Penetration Testing Framework | ✅ Pass | 27.639s |  |
| Web Application Security Scanner | ✅ Pass | 22.824s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.873s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 9.204s

---

### reflector (qwen/qwen3.5-35b-a3b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.177s |  |
| Text Transform Uppercase | ✅ Pass | 1.761s |  |
| Count from 1 to 5 | ✅ Pass | 2.230s |  |
| Math Calculation | ✅ Pass | 2.020s |  |
| Basic Echo Function | ✅ Pass | 1.069s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.739s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.650s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.308s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.554s |  |
| Search Query Function | ✅ Pass | 1.355s |  |
| Ask Advice Function | ✅ Pass | 1.450s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.319s |  |
| Basic Context Memory Test | ✅ Pass | 3.054s |  |
| Function Argument Memory Test | ✅ Pass | 1.053s |  |
| Function Response Memory Test | ✅ Pass | 1.318s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.568s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.063s |  |
| Penetration Testing Methodology | ✅ Pass | 5.066s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.070s |  |
| SQL Injection Attack Type | ✅ Pass | 5.581s |  |
| Penetration Testing Framework | ✅ Pass | 12.000s |  |
| Web Application Security Scanner | ✅ Pass | 3.616s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.594s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.375s

---

### searcher (qwen/qwen3.5-35b-a3b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.605s |  |
| Text Transform Uppercase | ✅ Pass | 2.089s |  |
| Count from 1 to 5 | ✅ Pass | 8.093s |  |
| Math Calculation | ✅ Pass | 1.417s |  |
| Basic Echo Function | ✅ Pass | 1.186s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.070s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.164s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.441s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.266s |  |
| Search Query Function | ✅ Pass | 1.045s |  |
| Ask Advice Function | ✅ Pass | 1.284s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.722s |  |
| Basic Context Memory Test | ✅ Pass | 3.651s |  |
| Function Argument Memory Test | ✅ Pass | 6.143s |  |
| Function Response Memory Test | ✅ Pass | 5.972s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 2.381s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.153s |  |
| Penetration Testing Methodology | ✅ Pass | 6.202s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.668s |  |
| SQL Injection Attack Type | ✅ Pass | 2.414s |  |
| Penetration Testing Framework | ✅ Pass | 4.669s |  |
| Web Application Security Scanner | ✅ Pass | 3.988s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.274s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 3.648s

---

### enricher (qwen/qwen3.5-35b-a3b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.084s |  |
| Text Transform Uppercase | ✅ Pass | 2.561s |  |
| Count from 1 to 5 | ✅ Pass | 1.884s |  |
| Math Calculation | ✅ Pass | 1.308s |  |
| Basic Echo Function | ✅ Pass | 1.227s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.092s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 15.459s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.655s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.266s |  |
| Search Query Function | ✅ Pass | 1.511s |  |
| Ask Advice Function | ✅ Pass | 1.737s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.579s |  |
| Basic Context Memory Test | ✅ Pass | 2.538s |  |
| Function Argument Memory Test | ✅ Pass | 1.031s |  |
| Function Response Memory Test | ✅ Pass | 0.992s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.167s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.353s |  |
| Penetration Testing Methodology | ✅ Pass | 7.191s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.350s |  |
| SQL Injection Attack Type | ✅ Pass | 3.672s |  |
| Penetration Testing Framework | ✅ Pass | 3.777s |  |
| Web Application Security Scanner | ✅ Pass | 4.323s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.877s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.332s

---

### coder (moonshotai/kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.418s |  |
| Text Transform Uppercase | ✅ Pass | 0.783s |  |
| Count from 1 to 5 | ✅ Pass | 1.644s |  |
| Math Calculation | ✅ Pass | 1.177s |  |
| Basic Echo Function | ✅ Pass | 1.741s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.137s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.736s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.831s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.516s |  |
| Search Query Function | ✅ Pass | 1.649s |  |
| Ask Advice Function | ✅ Pass | 1.316s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.804s |  |
| Basic Context Memory Test | ✅ Pass | 2.790s |  |
| Function Argument Memory Test | ✅ Pass | 1.301s |  |
| Function Response Memory Test | ✅ Pass | 3.491s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.182s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.580s |  |
| Penetration Testing Methodology | ✅ Pass | 4.092s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.292s |  |
| SQL Injection Attack Type | ✅ Pass | 1.417s |  |
| Penetration Testing Framework | ✅ Pass | 14.947s |  |
| Web Application Security Scanner | ✅ Pass | 4.822s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.867s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.067s

---

### installer (moonshotai/kimi-k2-instruct)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.799s |  |
| Text Transform Uppercase | ✅ Pass | 1.102s |  |
| Count from 1 to 5 | ✅ Pass | 1.179s |  |
| Math Calculation | ✅ Pass | 1.124s |  |
| Basic Echo Function | ✅ Pass | 1.514s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.115s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.312s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 1.588s | expected function 'echo' not found in tool calls: invalid JSON in tool call echo: unexpected end of JSON input |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.410s |  |
| Search Query Function | ✅ Pass | 1.787s |  |
| Ask Advice Function | ✅ Pass | 2.906s |  |
| Streaming Search Query Function Streaming | ❌ Fail | 1.552s | expected function 'search' not found in tool calls: invalid JSON in tool call search: unexpected end of JSON input |
| Basic Context Memory Test | ✅ Pass | 2.793s |  |
| Function Argument Memory Test | ✅ Pass | 1.024s |  |
| Function Response Memory Test | ✅ Pass | 0.989s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 2.065s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.145s |  |
| Penetration Testing Methodology | ✅ Pass | 1.163s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.252s |  |
| SQL Injection Attack Type | ✅ Pass | 1.508s |  |
| Penetration Testing Framework | ✅ Pass | 1.003s |  |
| Web Application Security Scanner | ✅ Pass | 1.105s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.601s |  |

**Summary**: 20/23 (86.96%) successful tests

**Average latency**: 1.480s

---

### pentester (moonshotai/kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.792s |  |
| Text Transform Uppercase | ✅ Pass | 0.903s |  |
| Count from 1 to 5 | ✅ Pass | 2.272s |  |
| Math Calculation | ✅ Pass | 3.338s |  |
| Basic Echo Function | ✅ Pass | 1.732s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.144s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.128s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.879s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.055s |  |
| Search Query Function | ✅ Pass | 2.093s |  |
| Ask Advice Function | ✅ Pass | 1.955s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.601s |  |
| Basic Context Memory Test | ✅ Pass | 2.653s |  |
| Function Argument Memory Test | ✅ Pass | 1.192s |  |
| Function Response Memory Test | ✅ Pass | 0.701s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.381s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.075s |  |
| Penetration Testing Methodology | ✅ Pass | 3.893s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.376s |  |
| SQL Injection Attack Type | ✅ Pass | 1.423s |  |
| Penetration Testing Framework | ✅ Pass | 16.347s |  |
| Web Application Security Scanner | ✅ Pass | 2.108s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.766s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.818s

---

