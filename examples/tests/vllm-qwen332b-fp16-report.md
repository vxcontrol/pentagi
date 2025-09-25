# LLM Agent Testing Report

Generated: Sun, 20 Jul 2025 13:25:52 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.541s |
| simple_json | Qwen/Qwen3-32B | true | 5/5 (100.00%) | 5.902s |
| primary_agent | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.065s |
| assistant | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.675s |
| generator | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 4.901s |
| refiner | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.431s |
| adviser | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.354s |
| reflector | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.846s |
| searcher | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.446s |
| enricher | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 4.885s |
| coder | Qwen/Qwen3-32B | true | 22/23 (95.65%) | 5.096s |
| installer | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.414s |
| pentester | Qwen/Qwen3-32B | true | 23/23 (100.00%) | 5.113s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 5.324s

## Detailed Results

### simple (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 3.390s |  |
| Simple Math | ✅ Pass | 4.054s |  |
| Count from 1 to 5 | ✅ Pass | 3.719s |  |
| Math Calculation | ✅ Pass | 3.740s |  |
| Basic Echo Function | ✅ Pass | 3.433s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.312s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.693s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.028s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.639s |  |
| Search Query Function | ✅ Pass | 2.889s |  |
| Ask Advice Function | ✅ Pass | 3.303s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.517s |  |
| Basic Context Memory Test | ✅ Pass | 4.085s |  |
| Function Argument Memory Test | ✅ Pass | 2.836s |  |
| Function Response Memory Test | ✅ Pass | 3.312s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.214s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 11.265s |  |
| Penetration Testing Methodology | ✅ Pass | 8.611s |  |
| SQL Injection Attack Type | ✅ Pass | 7.502s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.404s |  |
| Penetration Testing Framework | ✅ Pass | 11.920s |  |
| Web Application Security Scanner | ✅ Pass | 11.111s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.457s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.541s

---

### simple_json (Qwen/Qwen3-32B)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 3.559s |  |
| Project Information JSON | ✅ Pass | 5.681s |  |
| Vulnerability Report Memory Test | ✅ Pass | 6.815s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 5.920s |  |
| User Profile JSON | ✅ Pass | 7.536s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 5.902s

---

### primary_agent (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.530s |  |
| Text Transform Uppercase | ✅ Pass | 4.919s |  |
| Count from 1 to 5 | ✅ Pass | 3.883s |  |
| Math Calculation | ✅ Pass | 3.759s |  |
| Basic Echo Function | ✅ Pass | 3.557s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.373s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.761s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.462s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.991s |  |
| Search Query Function | ✅ Pass | 2.167s |  |
| Ask Advice Function | ✅ Pass | 3.439s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.621s |  |
| Basic Context Memory Test | ✅ Pass | 4.612s |  |
| Function Argument Memory Test | ✅ Pass | 2.837s |  |
| Function Response Memory Test | ✅ Pass | 3.271s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.488s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.523s |  |
| Penetration Testing Methodology | ✅ Pass | 12.771s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.006s |  |
| SQL Injection Attack Type | ✅ Pass | 4.692s |  |
| Penetration Testing Framework | ✅ Pass | 10.874s |  |
| Web Application Security Scanner | ✅ Pass | 12.035s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.904s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.065s

---

### assistant (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 3.229s |  |
| Simple Math | ✅ Pass | 5.249s |  |
| Count from 1 to 5 | ✅ Pass | 2.442s |  |
| Math Calculation | ✅ Pass | 4.310s |  |
| Basic Echo Function | ✅ Pass | 3.051s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.167s |  |
| Streaming Simple Math Streaming | ✅ Pass | 8.278s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.520s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.899s |  |
| Search Query Function | ✅ Pass | 2.403s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.923s |  |
| Ask Advice Function | ✅ Pass | 9.095s |  |
| Function Argument Memory Test | ✅ Pass | 2.841s |  |
| Basic Context Memory Test | ✅ Pass | 4.914s |  |
| Function Response Memory Test | ✅ Pass | 3.712s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.188s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.495s |  |
| Penetration Testing Methodology | ✅ Pass | 9.632s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.700s |  |
| SQL Injection Attack Type | ✅ Pass | 8.125s |  |
| Penetration Testing Framework | ✅ Pass | 15.399s |  |
| Web Application Security Scanner | ✅ Pass | 10.554s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.393s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.675s

---

### generator (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.738s |  |
| Text Transform Uppercase | ✅ Pass | 3.894s |  |
| Count from 1 to 5 | ✅ Pass | 3.469s |  |
| Math Calculation | ✅ Pass | 3.868s |  |
| Basic Echo Function | ✅ Pass | 2.773s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.712s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.358s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.242s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.721s |  |
| Search Query Function | ✅ Pass | 2.043s |  |
| Ask Advice Function | ✅ Pass | 3.615s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.448s |  |
| Basic Context Memory Test | ✅ Pass | 3.412s |  |
| Function Argument Memory Test | ✅ Pass | 2.620s |  |
| Function Response Memory Test | ✅ Pass | 4.005s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.994s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.204s |  |
| Penetration Testing Methodology | ✅ Pass | 10.604s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.348s |  |
| SQL Injection Attack Type | ✅ Pass | 7.001s |  |
| Penetration Testing Framework | ✅ Pass | 10.667s |  |
| Web Application Security Scanner | ✅ Pass | 8.563s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.407s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.901s

---

### refiner (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 4.220s |  |
| Count from 1 to 5 | ✅ Pass | 3.778s |  |
| Simple Math | ✅ Pass | 9.303s |  |
| Math Calculation | ✅ Pass | 4.662s |  |
| Basic Echo Function | ✅ Pass | 1.995s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.482s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.006s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.547s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.625s |  |
| Search Query Function | ✅ Pass | 2.681s |  |
| Ask Advice Function | ✅ Pass | 4.481s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.585s |  |
| Basic Context Memory Test | ✅ Pass | 5.955s |  |
| Function Argument Memory Test | ✅ Pass | 3.295s |  |
| Function Response Memory Test | ✅ Pass | 3.521s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.420s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 10.841s |  |
| Penetration Testing Methodology | ✅ Pass | 9.209s |  |
| SQL Injection Attack Type | ✅ Pass | 4.656s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.565s |  |
| Penetration Testing Framework | ✅ Pass | 10.161s |  |
| Web Application Security Scanner | ✅ Pass | 7.496s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.405s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.431s

---

### adviser (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.330s |  |
| Text Transform Uppercase | ✅ Pass | 5.537s |  |
| Count from 1 to 5 | ✅ Pass | 3.508s |  |
| Math Calculation | ✅ Pass | 4.882s |  |
| Basic Echo Function | ✅ Pass | 2.013s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.580s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.061s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.403s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.030s |  |
| Search Query Function | ✅ Pass | 2.725s |  |
| Ask Advice Function | ✅ Pass | 5.240s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.777s |  |
| Basic Context Memory Test | ✅ Pass | 3.996s |  |
| Function Argument Memory Test | ✅ Pass | 2.958s |  |
| Function Response Memory Test | ✅ Pass | 2.726s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.670s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.165s |  |
| Penetration Testing Methodology | ✅ Pass | 8.884s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.437s |  |
| SQL Injection Attack Type | ✅ Pass | 8.106s |  |
| Penetration Testing Framework | ✅ Pass | 10.900s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.597s |  |
| Web Application Security Scanner | ✅ Pass | 9.604s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.354s

---

### reflector (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.660s |  |
| Text Transform Uppercase | ✅ Pass | 3.066s |  |
| Math Calculation | ✅ Pass | 4.272s |  |
| Count from 1 to 5 | ✅ Pass | 7.650s |  |
| Basic Echo Function | ✅ Pass | 3.022s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.243s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.105s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.524s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.421s |  |
| Search Query Function | ✅ Pass | 3.080s |  |
| Ask Advice Function | ✅ Pass | 5.933s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.787s |  |
| Basic Context Memory Test | ✅ Pass | 5.332s |  |
| Function Argument Memory Test | ✅ Pass | 4.597s |  |
| Function Response Memory Test | ✅ Pass | 2.669s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.569s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.404s |  |
| Penetration Testing Methodology | ✅ Pass | 13.903s |  |
| SQL Injection Attack Type | ✅ Pass | 6.160s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.057s |  |
| Penetration Testing Framework | ✅ Pass | 10.714s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.059s |  |
| Web Application Security Scanner | ✅ Pass | 9.230s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.846s

---

### searcher (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.066s |  |
| Text Transform Uppercase | ✅ Pass | 5.782s |  |
| Count from 1 to 5 | ✅ Pass | 5.111s |  |
| Math Calculation | ✅ Pass | 2.660s |  |
| Basic Echo Function | ✅ Pass | 3.483s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.673s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.049s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.389s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.432s |  |
| Search Query Function | ✅ Pass | 2.889s |  |
| Ask Advice Function | ✅ Pass | 5.132s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.584s |  |
| Basic Context Memory Test | ✅ Pass | 3.971s |  |
| Function Argument Memory Test | ✅ Pass | 3.713s |  |
| Function Response Memory Test | ✅ Pass | 3.510s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.261s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.650s |  |
| Penetration Testing Methodology | ✅ Pass | 9.116s |  |
| SQL Injection Attack Type | ✅ Pass | 6.415s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.681s |  |
| Penetration Testing Framework | ✅ Pass | 12.049s |  |
| Web Application Security Scanner | ✅ Pass | 9.638s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.984s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.446s

---

### enricher (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.715s |  |
| Text Transform Uppercase | ✅ Pass | 5.366s |  |
| Count from 1 to 5 | ✅ Pass | 5.232s |  |
| Math Calculation | ✅ Pass | 3.462s |  |
| Basic Echo Function | ✅ Pass | 2.620s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.736s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.986s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.728s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.076s |  |
| Search Query Function | ✅ Pass | 2.704s |  |
| Ask Advice Function | ✅ Pass | 3.854s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.521s |  |
| Basic Context Memory Test | ✅ Pass | 3.604s |  |
| Function Argument Memory Test | ✅ Pass | 4.046s |  |
| Function Response Memory Test | ✅ Pass | 2.543s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.109s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.119s |  |
| Penetration Testing Methodology | ✅ Pass | 6.827s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.096s |  |
| SQL Injection Attack Type | ✅ Pass | 6.466s |  |
| Penetration Testing Framework | ✅ Pass | 10.130s |  |
| Web Application Security Scanner | ✅ Pass | 7.226s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.184s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.885s

---

### coder (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.551s |  |
| Text Transform Uppercase | ✅ Pass | 4.442s |  |
| Count from 1 to 5 | ✅ Pass | 4.976s |  |
| Math Calculation | ✅ Pass | 3.041s |  |
| Basic Echo Function | ✅ Pass | 3.255s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.239s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.213s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.540s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.755s |  |
| Search Query Function | ❌ Fail | 2.246s | expected function 'search' not found in tool calls: expected function search not found in tool calls |
| Ask Advice Function | ✅ Pass | 3.146s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.317s |  |
| Basic Context Memory Test | ✅ Pass | 4.155s |  |
| Function Argument Memory Test | ✅ Pass | 2.891s |  |
| Function Response Memory Test | ✅ Pass | 7.612s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.979s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.565s |  |
| Penetration Testing Methodology | ✅ Pass | 6.892s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.097s |  |
| SQL Injection Attack Type | ✅ Pass | 7.142s |  |
| Penetration Testing Framework | ✅ Pass | 11.961s |  |
| Web Application Security Scanner | ✅ Pass | 9.644s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.542s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.096s

---

### installer (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.358s |  |
| Text Transform Uppercase | ✅ Pass | 4.053s |  |
| Count from 1 to 5 | ✅ Pass | 3.771s |  |
| Math Calculation | ✅ Pass | 3.097s |  |
| Basic Echo Function | ✅ Pass | 2.917s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.971s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.510s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.380s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.380s |  |
| Search Query Function | ✅ Pass | 2.594s |  |
| Ask Advice Function | ✅ Pass | 3.716s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.361s |  |
| Basic Context Memory Test | ✅ Pass | 4.104s |  |
| Function Argument Memory Test | ✅ Pass | 3.151s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.898s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.151s |  |
| Function Response Memory Test | ✅ Pass | 12.693s |  |
| Penetration Testing Methodology | ✅ Pass | 13.344s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.867s |  |
| SQL Injection Attack Type | ✅ Pass | 5.939s |  |
| Penetration Testing Framework | ✅ Pass | 9.416s |  |
| Web Application Security Scanner | ✅ Pass | 8.099s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.748s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.414s

---

### pentester (Qwen/Qwen3-32B)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.647s |  |
| Text Transform Uppercase | ✅ Pass | 6.479s |  |
| Count from 1 to 5 | ✅ Pass | 5.585s |  |
| Math Calculation | ✅ Pass | 4.859s |  |
| Basic Echo Function | ✅ Pass | 3.692s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.730s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.935s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.265s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.686s |  |
| Search Query Function | ✅ Pass | 2.996s |  |
| Ask Advice Function | ✅ Pass | 5.096s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.838s |  |
| Basic Context Memory Test | ✅ Pass | 4.337s |  |
| Function Argument Memory Test | ✅ Pass | 3.767s |  |
| Function Response Memory Test | ✅ Pass | 3.495s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.579s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.143s |  |
| Penetration Testing Methodology | ✅ Pass | 6.789s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.145s |  |
| SQL Injection Attack Type | ✅ Pass | 5.700s |  |
| Penetration Testing Framework | ✅ Pass | 7.321s |  |
| Web Application Security Scanner | ✅ Pass | 7.881s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.615s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.113s

---

