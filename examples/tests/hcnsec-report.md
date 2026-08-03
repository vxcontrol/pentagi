# LLM Agent Testing Report

Generated: Sat, 25 Jul 2026 11:53:24 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen3.6-35B-A3B | false | 23/24 (95.83%) | 1.683s |
| simple_json | Qwen3.6-35B-A3B | false | 7/7 (100.00%) | 1.942s |
| primary_agent | MiniMax-M2.7 | true | 23/24 (95.83%) | 17.986s |
| assistant | MiniMax-M2.7 | true | 23/24 (95.83%) | 0.353s |
| generator | MiniMax-M2.7 | true | 22/24 (91.67%) | 34.613s |
| refiner | MiniMax-M2.7 | true | 23/24 (95.83%) | 1.118s |
| adviser | MiniMax-M2.7 | true | 23/24 (95.83%) | 0.237s |
| reflector | Qwen3.6-35B-A3B | true | 23/24 (95.83%) | 1.516s |
| searcher | Qwen3.6-35B-A3B | true | 24/24 (100.00%) | 1.518s |
| enricher | Qwen3.6-35B-A3B | true | 23/24 (95.83%) | 1.768s |
| coder | MiniMax-M2.7 | true | 23/24 (95.83%) | 15.113s |
| installer | MiniMax-M2.7 | true | 23/24 (95.83%) | 0.304s |
| pentester | Qwen3.6-35B-A3B | true | 23/24 (95.83%) | 1.598s |

**Total**: 283/295 (95.93%) successful tests
**Overall average latency**: 6.376s

## Detailed Results

### simple (Qwen3.6-35B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.085s |  |
| Text Transform Uppercase | ✅ Pass | 0.216s |  |
| Count from 1 to 5 | ✅ Pass | 0.226s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.223s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.216s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.222s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.844s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.676s |  |
| Search Query Function | ✅ Pass | 1.321s |  |
| Ask Advice Function | ✅ Pass | 2.362s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.245s |  |
| Basic Context Memory Test | ✅ Pass | 1.757s |  |
| Function Argument Memory Test | ✅ Pass | 1.031s |  |
| Function Response Memory Test | ✅ Pass | 1.203s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 4.623s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.982s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 6.689s |  |
| Penetration Testing Methodology | ✅ Pass | 3.667s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.423s |  |
| SQL Injection Attack Type | ✅ Pass | 0.995s |  |
| Penetration Testing Framework | ✅ Pass | 2.298s |  |
| Web Application Security Scanner | ✅ Pass | 2.016s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.843s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 1.683s

---

### simple_json (Qwen3.6-35B-A3B)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.940s |  |
| Person Information JSON | ✅ Pass | 1.879s |  |
| Project Information JSON | ✅ Pass | 1.780s |  |
| User Profile JSON | ✅ Pass | 1.162s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.518s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 2.963s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 2.347s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 1.942s

---

### primary_agent (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.229s |  |
| Text Transform Uppercase | ✅ Pass | 0.211s |  |
| Count from 1 to 5 | ✅ Pass | 0.217s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.213s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.222s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.350s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 36.668s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 30.247s |  |
| Search Query Function | ✅ Pass | 26.245s |  |
| Ask Advice Function | ✅ Pass | 21.962s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 19.611s |  |
| Basic Context Memory Test | ✅ Pass | 12.075s |  |
| Function Argument Memory Test | ✅ Pass | 17.800s |  |
| Function Response Memory Test | ✅ Pass | 16.658s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 27.914s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 19.463s |  |
| Read a file, then edit it via unified diff | ❌ Fail | 45.071s | edit\_file's diff applied but did not produce "Priority: high" \(result: "Status: draft\nPriority: high\nPriority: low\n"\) |
| Penetration Testing Methodology | ✅ Pass | 47.195s |  |
| Vulnerability Assessment Tools | ✅ Pass | 40.454s |  |
| SQL Injection Attack Type | ✅ Pass | 36.091s |  |
| Penetration Testing Framework | ✅ Pass | 11.735s |  |
| Web Application Security Scanner | ✅ Pass | 9.673s |  |
| Penetration Testing Tool Selection | ✅ Pass | 11.147s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 17.986s

---

### assistant (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.207s |  |
| Text Transform Uppercase | ✅ Pass | 0.224s |  |
| Count from 1 to 5 | ✅ Pass | 0.222s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.212s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.228s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.219s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.287s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.286s |  |
| Search Query Function | ✅ Pass | 0.314s |  |
| Ask Advice Function | ✅ Pass | 0.372s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.293s |  |
| Basic Context Memory Test | ✅ Pass | 0.315s |  |
| Function Argument Memory Test | ✅ Pass | 0.340s |  |
| Function Response Memory Test | ✅ Pass | 0.279s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.294s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.300s |  |
| Read a file, then edit it via unified diff | ❌ Fail | 0.495s | edit\_file's diff applied but did not produce "Priority: high" \(result: "Status: draft\nPriority: high\nPriority: low\n"\) |
| Penetration Testing Methodology | ✅ Pass | 1.560s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.355s |  |
| SQL Injection Attack Type | ✅ Pass | 0.274s |  |
| Penetration Testing Framework | ✅ Pass | 0.214s |  |
| Web Application Security Scanner | ✅ Pass | 0.673s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.277s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 0.353s

---

### generator (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 48.107s |  |
| Text Transform Uppercase | ✅ Pass | 20.175s |  |
| Count from 1 to 5 | ✅ Pass | 12.739s |  |
| Math Calculation | ✅ Pass | 16.767s |  |
| Basic Echo Function | ✅ Pass | 17.058s |  |
| Streaming Simple Math Streaming | ✅ Pass | 24.739s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 24.898s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 43.708s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ❌ Fail | 300.475s | API returned unexpected status code: 504 |
| Search Query Function | ✅ Pass | 22.717s |  |
| Ask Advice Function | ✅ Pass | 23.039s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 22.311s |  |
| Basic Context Memory Test | ✅ Pass | 13.975s |  |
| Function Argument Memory Test | ✅ Pass | 16.355s |  |
| Function Response Memory Test | ✅ Pass | 24.776s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 20.167s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 22.979s |  |
| Read a file, then edit it via unified diff | ❌ Fail | 47.010s | edit\_file's diff applied but did not produce "Priority: high" \(result: "Status: draft\nPriority: high\nPriority: low\n"\) |
| Penetration Testing Methodology | ✅ Pass | 23.258s |  |
| Vulnerability Assessment Tools | ✅ Pass | 32.940s |  |
| SQL Injection Attack Type | ✅ Pass | 11.754s |  |
| Penetration Testing Framework | ✅ Pass | 8.989s |  |
| Web Application Security Scanner | ✅ Pass | 13.227s |  |
| Penetration Testing Tool Selection | ✅ Pass | 18.538s |  |

**Summary**: 22/24 (91.67%) successful tests

**Average latency**: 34.613s

---

### refiner (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.293s |  |
| Text Transform Uppercase | ✅ Pass | 0.300s |  |
| Count from 1 to 5 | ✅ Pass | 0.305s |  |
| Math Calculation | ✅ Pass | 0.299s |  |
| Basic Echo Function | ✅ Pass | 0.279s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.299s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.288s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.287s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 18.541s |  |
| Search Query Function | ✅ Pass | 0.302s |  |
| Ask Advice Function | ✅ Pass | 0.358s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.584s |  |
| Basic Context Memory Test | ✅ Pass | 0.224s |  |
| Function Argument Memory Test | ✅ Pass | 0.286s |  |
| Function Response Memory Test | ✅ Pass | 0.290s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.280s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.312s |  |
| Read a file, then edit it via unified diff | ❌ Fail | 0.501s | edit\_file's diff applied but did not produce "Priority: high" \(result: "Status: draft\nPriority: high\nPriority: low\n"\) |
| Penetration Testing Methodology | ✅ Pass | 0.286s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.347s |  |
| SQL Injection Attack Type | ✅ Pass | 0.278s |  |
| Penetration Testing Framework | ✅ Pass | 0.285s |  |
| Web Application Security Scanner | ✅ Pass | 0.273s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.314s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 1.118s

---

### adviser (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.234s |  |
| Text Transform Uppercase | ✅ Pass | 0.233s |  |
| Count from 1 to 5 | ✅ Pass | 0.216s |  |
| Math Calculation | ✅ Pass | 0.205s |  |
| Basic Echo Function | ✅ Pass | 0.215s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.233s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.218s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.223s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.365s |  |
| Search Query Function | ✅ Pass | 0.208s |  |
| Ask Advice Function | ✅ Pass | 0.219s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.216s |  |
| Basic Context Memory Test | ✅ Pass | 0.209s |  |
| Function Argument Memory Test | ✅ Pass | 0.220s |  |
| Function Response Memory Test | ✅ Pass | 0.222s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.210s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.209s |  |
| Read a file, then edit it via unified diff | ❌ Fail | 0.465s | edit\_file's diff applied but did not produce "Priority: high" \(result: "Status: draft\nPriority: high\nPriority: low\n"\) |
| Penetration Testing Methodology | ✅ Pass | 0.242s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.284s |  |
| SQL Injection Attack Type | ✅ Pass | 0.210s |  |
| Penetration Testing Framework | ✅ Pass | 0.209s |  |
| Web Application Security Scanner | ✅ Pass | 0.212s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.210s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 0.237s

---

### reflector (Qwen3.6-35B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.455s |  |
| Text Transform Uppercase | ✅ Pass | 0.223s |  |
| Count from 1 to 5 | ✅ Pass | 0.219s |  |
| Math Calculation | ✅ Pass | 0.221s |  |
| Basic Echo Function | ✅ Pass | 0.212s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.227s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.215s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.970s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.154s |  |
| Search Query Function | ✅ Pass | 0.826s |  |
| Ask Advice Function | ✅ Pass | 1.286s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.734s |  |
| Basic Context Memory Test | ✅ Pass | 0.907s |  |
| Function Argument Memory Test | ✅ Pass | 1.674s |  |
| Function Response Memory Test | ✅ Pass | 1.166s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 2.852s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.307s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.336s |  |
| Penetration Testing Methodology | ✅ Pass | 3.367s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.843s |  |
| SQL Injection Attack Type | ✅ Pass | 2.740s |  |
| Penetration Testing Framework | ✅ Pass | 3.580s |  |
| Web Application Security Scanner | ✅ Pass | 1.280s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.574s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 1.516s

---

### searcher (Qwen3.6-35B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.227s |  |
| Text Transform Uppercase | ✅ Pass | 0.243s |  |
| Count from 1 to 5 | ✅ Pass | 0.222s |  |
| Math Calculation | ✅ Pass | 0.209s |  |
| Basic Echo Function | ✅ Pass | 0.209s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.222s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.223s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.962s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.839s |  |
| Search Query Function | ✅ Pass | 2.242s |  |
| Ask Advice Function | ✅ Pass | 0.889s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.945s |  |
| Basic Context Memory Test | ✅ Pass | 0.922s |  |
| Function Argument Memory Test | ✅ Pass | 2.475s |  |
| Function Response Memory Test | ✅ Pass | 1.163s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.711s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.014s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.999s |  |
| Penetration Testing Methodology | ✅ Pass | 4.439s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.656s |  |
| SQL Injection Attack Type | ✅ Pass | 1.181s |  |
| Penetration Testing Framework | ✅ Pass | 6.323s |  |
| Web Application Security Scanner | ✅ Pass | 1.865s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.238s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.518s

---

### enricher (Qwen3.6-35B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.217s |  |
| Text Transform Uppercase | ✅ Pass | 0.262s |  |
| Count from 1 to 5 | ✅ Pass | 0.214s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.209s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.217s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.229s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.196s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.981s |  |
| Search Query Function | ✅ Pass | 3.136s |  |
| Ask Advice Function | ✅ Pass | 4.184s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.256s |  |
| Basic Context Memory Test | ✅ Pass | 0.997s |  |
| Function Argument Memory Test | ✅ Pass | 0.902s |  |
| Function Response Memory Test | ✅ Pass | 0.812s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 3.350s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.043s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.770s |  |
| Penetration Testing Methodology | ✅ Pass | 2.731s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.786s |  |
| SQL Injection Attack Type | ✅ Pass | 3.018s |  |
| Penetration Testing Framework | ✅ Pass | 2.320s |  |
| Web Application Security Scanner | ✅ Pass | 6.592s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.793s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 1.768s

---

### coder (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.212s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.213s |  |
| Math Calculation | ✅ Pass | 0.221s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.215s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 24.674s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 36.732s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 36.234s |  |
| Search Query Function | ❌ Fail | 21.186s | expected function 'search' not found in tool calls: expected function search not found in tool calls |
| Ask Advice Function | ✅ Pass | 13.571s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 18.545s |  |
| Basic Context Memory Test | ✅ Pass | 15.315s |  |
| Function Argument Memory Test | ✅ Pass | 13.221s |  |
| Function Response Memory Test | ✅ Pass | 14.308s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 17.397s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 16.827s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 47.220s |  |
| Penetration Testing Methodology | ✅ Pass | 23.848s |  |
| Vulnerability Assessment Tools | ✅ Pass | 24.060s |  |
| SQL Injection Attack Type | ✅ Pass | 7.113s |  |
| Penetration Testing Framework | ✅ Pass | 5.152s |  |
| Web Application Security Scanner | ✅ Pass | 10.533s |  |
| Penetration Testing Tool Selection | ✅ Pass | 15.472s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 15.113s

---

### installer (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.208s |  |
| Text Transform Uppercase | ✅ Pass | 0.272s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.208s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.507s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.291s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.280s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.277s |  |
| Search Query Function | ✅ Pass | 0.286s |  |
| Ask Advice Function | ✅ Pass | 0.289s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.302s |  |
| Basic Context Memory Test | ✅ Pass | 0.289s |  |
| Function Argument Memory Test | ✅ Pass | 0.290s |  |
| Function Response Memory Test | ✅ Pass | 0.291s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.378s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.301s |  |
| Read a file, then edit it via unified diff | ❌ Fail | 0.485s | edit\_file's diff applied but did not produce "Priority: high" \(result: "Status: draft\nPriority: high\nPriority: low\n"\) |
| Penetration Testing Methodology | ✅ Pass | 0.352s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.409s |  |
| SQL Injection Attack Type | ✅ Pass | 0.218s |  |
| Penetration Testing Framework | ✅ Pass | 0.342s |  |
| Web Application Security Scanner | ✅ Pass | 0.281s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.306s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 0.304s

---

### pentester (Qwen3.6-35B-A3B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.426s |  |
| Text Transform Uppercase | ✅ Pass | 0.218s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.213s |  |
| Basic Echo Function | ✅ Pass | 0.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.265s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.052s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.565s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.822s |  |
| Search Query Function | ✅ Pass | 1.764s |  |
| Ask Advice Function | ✅ Pass | 1.347s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.960s |  |
| Basic Context Memory Test | ✅ Pass | 2.841s |  |
| Function Argument Memory Test | ✅ Pass | 1.057s |  |
| Function Response Memory Test | ✅ Pass | 0.875s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 3.938s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.522s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.300s |  |
| Penetration Testing Methodology | ✅ Pass | 2.824s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.836s |  |
| SQL Injection Attack Type | ✅ Pass | 0.991s |  |
| Penetration Testing Framework | ✅ Pass | 2.814s |  |
| Web Application Security Scanner | ✅ Pass | 5.468s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.816s |  |

**Summary**: 23/24 (95.83%) successful tests

**Average latency**: 1.598s

---

