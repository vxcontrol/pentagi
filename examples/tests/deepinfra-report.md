# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 21:30:02 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | MiniMaxAI/MiniMax-M3 | true | 23/23 (100.00%) | 0.213s |
| simple_json | moonshotai/Kimi-K2.6 | true | 7/7 (100.00%) | 0.213s |
| primary_agent | moonshotai/Kimi-K2.6 | true | 23/23 (100.00%) | 0.214s |
| assistant | moonshotai/Kimi-K2.6 | true | 23/23 (100.00%) | 0.213s |
| generator | zai-org/GLM-5.2 | true | 23/23 (100.00%) | 0.207s |
| refiner | zai-org/GLM-5.2 | true | 23/23 (100.00%) | 0.214s |
| adviser | MiniMaxAI/MiniMax-M3 | true | 23/23 (100.00%) | 0.216s |
| reflector | MiniMaxAI/MiniMax-M3 | true | 23/23 (100.00%) | 0.217s |
| searcher | MiniMaxAI/MiniMax-M3 | true | 23/23 (100.00%) | 0.217s |
| enricher | MiniMaxAI/MiniMax-M3 | true | 23/23 (100.00%) | 0.213s |
| coder | moonshotai/Kimi-K2.7-Code | true | 23/23 (100.00%) | 0.212s |
| installer | moonshotai/Kimi-K2.7-Code | true | 23/23 (100.00%) | 0.214s |
| pentester | moonshotai/Kimi-K2.6 | true | 23/23 (100.00%) | 0.210s |

**Total**: 283/283 (100.00%) successful tests
**Overall average latency**: 0.213s

## Detailed Results

### simple (MiniMaxAI/MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.254s |  |
| Text Transform Uppercase | ✅ Pass | 0.218s |  |
| Count from 1 to 5 | ✅ Pass | 0.223s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.204s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.204s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.208s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.207s |  |
| Search Query Function | ✅ Pass | 0.203s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.225s |  |
| Basic Context Memory Test | ✅ Pass | 0.210s |  |
| Function Argument Memory Test | ✅ Pass | 0.211s |  |
| Function Response Memory Test | ✅ Pass | 0.203s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.207s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.216s |  |
| Penetration Testing Methodology | ✅ Pass | 0.215s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.210s |  |
| SQL Injection Attack Type | ✅ Pass | 0.220s |  |
| Penetration Testing Framework | ✅ Pass | 0.201s |  |
| Web Application Security Scanner | ✅ Pass | 0.202s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.212s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.213s

---

### simple_json (moonshotai/Kimi-K2.6)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.214s |  |
| Person Information JSON | ✅ Pass | 0.215s |  |
| Project Information JSON | ✅ Pass | 0.214s |  |
| User Profile JSON | ✅ Pass | 0.216s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.207s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.209s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.211s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 0.213s

---

### primary_agent (moonshotai/Kimi-K2.6)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.261s |  |
| Text Transform Uppercase | ✅ Pass | 0.214s |  |
| Count from 1 to 5 | ✅ Pass | 0.211s |  |
| Math Calculation | ✅ Pass | 0.202s |  |
| Basic Echo Function | ✅ Pass | 0.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.206s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.226s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.215s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.203s |  |
| Search Query Function | ✅ Pass | 0.207s |  |
| Ask Advice Function | ✅ Pass | 0.212s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.230s |  |
| Basic Context Memory Test | ✅ Pass | 0.210s |  |
| Function Argument Memory Test | ✅ Pass | 0.209s |  |
| Function Response Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.208s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Methodology | ✅ Pass | 0.215s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.212s |  |
| SQL Injection Attack Type | ✅ Pass | 0.221s |  |
| Penetration Testing Framework | ✅ Pass | 0.203s |  |
| Web Application Security Scanner | ✅ Pass | 0.207s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.207s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.214s

---

### assistant (moonshotai/Kimi-K2.6)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.240s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.213s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.201s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.206s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.225s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.220s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.198s |  |
| Search Query Function | ✅ Pass | 0.208s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.241s |  |
| Basic Context Memory Test | ✅ Pass | 0.219s |  |
| Function Argument Memory Test | ✅ Pass | 0.206s |  |
| Function Response Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.208s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Methodology | ✅ Pass | 0.215s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.209s |  |
| SQL Injection Attack Type | ✅ Pass | 0.214s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.207s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.213s

---

### generator (zai-org/GLM-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.080s |  |
| Text Transform Uppercase | ✅ Pass | 0.217s |  |
| Count from 1 to 5 | ✅ Pass | 0.212s |  |
| Math Calculation | ✅ Pass | 0.222s |  |
| Basic Echo Function | ✅ Pass | 0.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.209s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.229s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.221s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.200s |  |
| Search Query Function | ✅ Pass | 0.206s |  |
| Ask Advice Function | ✅ Pass | 0.221s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.239s |  |
| Basic Context Memory Test | ✅ Pass | 0.212s |  |
| Function Argument Memory Test | ✅ Pass | 0.213s |  |
| Function Response Memory Test | ✅ Pass | 0.211s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.206s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Methodology | ✅ Pass | 0.213s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.210s |  |
| SQL Injection Attack Type | ✅ Pass | 0.208s |  |
| Penetration Testing Framework | ✅ Pass | 0.205s |  |
| Web Application Security Scanner | ✅ Pass | 0.204s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.207s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.207s

---

### refiner (zai-org/GLM-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.219s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.225s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.209s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.205s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.241s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.221s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.198s |  |
| Search Query Function | ✅ Pass | 0.210s |  |
| Ask Advice Function | ✅ Pass | 0.220s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.235s |  |
| Basic Context Memory Test | ✅ Pass | 0.202s |  |
| Function Argument Memory Test | ✅ Pass | 0.210s |  |
| Function Response Memory Test | ✅ Pass | 0.211s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.213s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.206s |  |
| Penetration Testing Methodology | ✅ Pass | 0.217s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.205s |  |
| SQL Injection Attack Type | ✅ Pass | 0.213s |  |
| Penetration Testing Framework | ✅ Pass | 0.212s |  |
| Web Application Security Scanner | ✅ Pass | 0.205s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.212s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.214s

---

### adviser (MiniMaxAI/MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.272s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.205s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.203s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.206s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.239s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.210s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.202s |  |
| Search Query Function | ✅ Pass | 0.234s |  |
| Ask Advice Function | ✅ Pass | 0.218s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.208s |  |
| Basic Context Memory Test | ✅ Pass | 0.203s |  |
| Function Argument Memory Test | ✅ Pass | 0.212s |  |
| Function Response Memory Test | ✅ Pass | 0.213s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.216s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Methodology | ✅ Pass | 0.213s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.217s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.206s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.215s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.216s

---

### reflector (MiniMaxAI/MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.269s |  |
| Text Transform Uppercase | ✅ Pass | 0.204s |  |
| Count from 1 to 5 | ✅ Pass | 0.211s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.203s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.210s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.236s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.216s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.197s |  |
| Search Query Function | ✅ Pass | 0.245s |  |
| Ask Advice Function | ✅ Pass | 0.218s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.208s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.215s |  |
| Function Response Memory Test | ✅ Pass | 0.206s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.269s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Methodology | ✅ Pass | 0.210s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.204s |  |
| SQL Injection Attack Type | ✅ Pass | 0.207s |  |
| Penetration Testing Framework | ✅ Pass | 0.210s |  |
| Web Application Security Scanner | ✅ Pass | 0.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.210s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.217s

---

### searcher (MiniMaxAI/MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.251s |  |
| Text Transform Uppercase | ✅ Pass | 0.218s |  |
| Count from 1 to 5 | ✅ Pass | 0.227s |  |
| Math Calculation | ✅ Pass | 0.217s |  |
| Basic Echo Function | ✅ Pass | 0.199s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.225s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.239s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.213s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.209s |  |
| Search Query Function | ✅ Pass | 0.236s |  |
| Ask Advice Function | ✅ Pass | 0.212s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.207s |  |
| Basic Context Memory Test | ✅ Pass | 0.214s |  |
| Function Argument Memory Test | ✅ Pass | 0.214s |  |
| Function Response Memory Test | ✅ Pass | 0.211s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.205s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.209s |  |
| Penetration Testing Methodology | ✅ Pass | 0.213s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.210s |  |
| SQL Injection Attack Type | ✅ Pass | 0.211s |  |
| Penetration Testing Framework | ✅ Pass | 0.215s |  |
| Web Application Security Scanner | ✅ Pass | 0.209s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.211s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.217s

---

### enricher (MiniMaxAI/MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.203s |  |
| Text Transform Uppercase | ✅ Pass | 0.222s |  |
| Count from 1 to 5 | ✅ Pass | 0.218s |  |
| Math Calculation | ✅ Pass | 0.209s |  |
| Basic Echo Function | ✅ Pass | 0.206s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.212s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.228s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.214s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.210s |  |
| Search Query Function | ✅ Pass | 0.232s |  |
| Ask Advice Function | ✅ Pass | 0.205s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.206s |  |
| Basic Context Memory Test | ✅ Pass | 0.216s |  |
| Function Argument Memory Test | ✅ Pass | 0.214s |  |
| Function Response Memory Test | ✅ Pass | 0.213s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Methodology | ✅ Pass | 0.209s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.210s |  |
| SQL Injection Attack Type | ✅ Pass | 0.205s |  |
| Penetration Testing Framework | ✅ Pass | 0.207s |  |
| Web Application Security Scanner | ✅ Pass | 0.209s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.211s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.213s

---

### coder (moonshotai/Kimi-K2.7-Code)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.207s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.214s |  |
| Math Calculation | ✅ Pass | 0.211s |  |
| Basic Echo Function | ✅ Pass | 0.219s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.220s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.217s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.206s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.207s |  |
| Search Query Function | ✅ Pass | 0.216s |  |
| Ask Advice Function | ✅ Pass | 0.217s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.208s |  |
| Basic Context Memory Test | ✅ Pass | 0.215s |  |
| Function Argument Memory Test | ✅ Pass | 0.213s |  |
| Function Response Memory Test | ✅ Pass | 0.214s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.213s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.217s |  |
| Penetration Testing Methodology | ✅ Pass | 0.207s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.210s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.204s |  |
| Web Application Security Scanner | ✅ Pass | 0.203s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.211s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.212s

---

### installer (moonshotai/Kimi-K2.7-Code)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.208s |  |
| Text Transform Uppercase | ✅ Pass | 0.226s |  |
| Count from 1 to 5 | ✅ Pass | 0.224s |  |
| Math Calculation | ✅ Pass | 0.211s |  |
| Basic Echo Function | ✅ Pass | 0.213s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.216s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.216s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.211s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.209s |  |
| Search Query Function | ✅ Pass | 0.218s |  |
| Ask Advice Function | ✅ Pass | 0.213s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.232s |  |
| Basic Context Memory Test | ✅ Pass | 0.212s |  |
| Function Argument Memory Test | ✅ Pass | 0.203s |  |
| Function Response Memory Test | ✅ Pass | 0.209s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.215s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.219s |  |
| Penetration Testing Methodology | ✅ Pass | 0.207s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.213s |  |
| Penetration Testing Framework | ✅ Pass | 0.216s |  |
| Web Application Security Scanner | ✅ Pass | 0.204s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.212s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.214s

---

### pentester (moonshotai/Kimi-K2.6)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.208s |  |
| Text Transform Uppercase | ✅ Pass | 0.228s |  |
| Count from 1 to 5 | ✅ Pass | 0.212s |  |
| Math Calculation | ✅ Pass | 0.219s |  |
| Basic Echo Function | ✅ Pass | 0.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.207s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.207s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.217s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.203s |  |
| Search Query Function | ✅ Pass | 0.205s |  |
| Ask Advice Function | ✅ Pass | 0.211s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.207s |  |
| Basic Context Memory Test | ✅ Pass | 0.214s |  |
| Function Argument Memory Test | ✅ Pass | 0.204s |  |
| Function Response Memory Test | ✅ Pass | 0.206s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Methodology | ✅ Pass | 0.208s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.212s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.205s |  |
| Web Application Security Scanner | ✅ Pass | 0.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.209s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.210s

---

