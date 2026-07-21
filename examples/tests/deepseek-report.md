# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 16:50:46 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek-v4-flash | true | 23/23 (100.00%) | 0.225s |
| simple_json | deepseek-v4-flash | true | 6/7 (85.71%) | 0.359s |
| primary_agent | deepseek-v4-pro | true | 23/23 (100.00%) | 0.222s |
| assistant | deepseek-v4-pro | true | 23/23 (100.00%) | 0.219s |
| generator | deepseek-v4-pro | true | 23/23 (100.00%) | 0.222s |
| refiner | deepseek-v4-pro | true | 23/23 (100.00%) | 0.220s |
| adviser | deepseek-v4-pro | true | 23/23 (100.00%) | 0.213s |
| reflector | deepseek-v4-flash | true | 23/23 (100.00%) | 0.218s |
| searcher | deepseek-v4-flash | true | 23/23 (100.00%) | 0.211s |
| enricher | deepseek-v4-flash | true | 23/23 (100.00%) | 0.213s |
| coder | deepseek-v4-pro | true | 23/23 (100.00%) | 0.217s |
| installer | deepseek-v4-flash | true | 23/23 (100.00%) | 0.217s |
| pentester | deepseek-v4-pro | true | 23/23 (100.00%) | 0.224s |

**Total**: 282/283 (99.65%) successful tests
**Overall average latency**: 0.222s

## Detailed Results

### simple (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.389s |  |
| Text Transform Uppercase | ✅ Pass | 0.231s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.210s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.205s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.222s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.210s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.212s |  |
| Search Query Function | ✅ Pass | 0.211s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.210s |  |
| Basic Context Memory Test | ✅ Pass | 0.205s |  |
| Function Argument Memory Test | ✅ Pass | 0.206s |  |
| Function Response Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.281s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Methodology | ✅ Pass | 0.267s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.165s |  |
| SQL Injection Attack Type | ✅ Pass | 0.199s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.173s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.334s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.225s

---

### simple_json (deepseek-v4-flash)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.374s |  |
| Person Information JSON | ✅ Pass | 0.368s |  |
| Project Information JSON | ✅ Pass | 0.365s |  |
| User Profile JSON | ✅ Pass | 0.309s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.248s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.287s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ❌ Fail | 0.556s | API returned unexpected status code: 400: litellm\.BadRequestError: DeepseekException \- \{"error":\{"message":"This response\_format type is unava... |

**Summary**: 6/7 (85.71%) successful tests

**Average latency**: 0.359s

---

### primary_agent (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.422s |  |
| Text Transform Uppercase | ✅ Pass | 0.236s |  |
| Count from 1 to 5 | ✅ Pass | 0.205s |  |
| Math Calculation | ✅ Pass | 0.206s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.203s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.213s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.210s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.204s |  |
| Search Query Function | ✅ Pass | 0.207s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.211s |  |
| Basic Context Memory Test | ✅ Pass | 0.203s |  |
| Function Argument Memory Test | ✅ Pass | 0.213s |  |
| Function Response Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.204s |  |
| Penetration Testing Methodology | ✅ Pass | 0.250s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.188s |  |
| SQL Injection Attack Type | ✅ Pass | 0.198s |  |
| Penetration Testing Framework | ✅ Pass | 0.185s |  |
| Web Application Security Scanner | ✅ Pass | 0.178s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.315s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.222s

---

### assistant (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.399s |  |
| Text Transform Uppercase | ✅ Pass | 0.239s |  |
| Count from 1 to 5 | ✅ Pass | 0.204s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.205s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.205s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.203s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.210s |  |
| Search Query Function | ✅ Pass | 0.210s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.209s |  |
| Basic Context Memory Test | ✅ Pass | 0.208s |  |
| Function Argument Memory Test | ✅ Pass | 0.209s |  |
| Function Response Memory Test | ✅ Pass | 0.200s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.213s |  |
| Penetration Testing Methodology | ✅ Pass | 0.236s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.209s |  |
| SQL Injection Attack Type | ✅ Pass | 0.193s |  |
| Penetration Testing Framework | ✅ Pass | 0.137s |  |
| Web Application Security Scanner | ✅ Pass | 0.202s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.312s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.219s

---

### generator (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.389s |  |
| Text Transform Uppercase | ✅ Pass | 0.238s |  |
| Count from 1 to 5 | ✅ Pass | 0.206s |  |
| Math Calculation | ✅ Pass | 0.209s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.204s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.220s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.206s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.205s |  |
| Search Query Function | ✅ Pass | 0.210s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.206s |  |
| Basic Context Memory Test | ✅ Pass | 0.210s |  |
| Function Argument Memory Test | ✅ Pass | 0.206s |  |
| Function Response Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.214s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Methodology | ✅ Pass | 0.265s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.204s |  |
| SQL Injection Attack Type | ✅ Pass | 0.197s |  |
| Penetration Testing Framework | ✅ Pass | 0.169s |  |
| Web Application Security Scanner | ✅ Pass | 0.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.300s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.222s

---

### refiner (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.419s |  |
| Text Transform Uppercase | ✅ Pass | 0.204s |  |
| Count from 1 to 5 | ✅ Pass | 0.210s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.214s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.212s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.212s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.206s |  |
| Search Query Function | ✅ Pass | 0.206s |  |
| Ask Advice Function | ✅ Pass | 0.207s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.207s |  |
| Basic Context Memory Test | ✅ Pass | 0.206s |  |
| Function Argument Memory Test | ✅ Pass | 0.210s |  |
| Function Response Memory Test | ✅ Pass | 0.203s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.217s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Methodology | ✅ Pass | 0.256s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.194s |  |
| SQL Injection Attack Type | ✅ Pass | 0.183s |  |
| Penetration Testing Framework | ✅ Pass | 0.170s |  |
| Web Application Security Scanner | ✅ Pass | 0.203s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.287s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.220s

---

### adviser (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.389s |  |
| Text Transform Uppercase | ✅ Pass | 0.208s |  |
| Count from 1 to 5 | ✅ Pass | 0.214s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.208s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.202s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.218s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.213s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.207s |  |
| Search Query Function | ✅ Pass | 0.206s |  |
| Ask Advice Function | ✅ Pass | 0.207s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.203s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.209s |  |
| Function Response Memory Test | ✅ Pass | 0.209s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.221s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.214s |  |
| Penetration Testing Methodology | ✅ Pass | 0.213s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.200s |  |
| SQL Injection Attack Type | ✅ Pass | 0.136s |  |
| Penetration Testing Framework | ✅ Pass | 0.193s |  |
| Web Application Security Scanner | ✅ Pass | 0.197s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.215s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.213s

---

### reflector (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.423s |  |
| Text Transform Uppercase | ✅ Pass | 0.221s |  |
| Count from 1 to 5 | ✅ Pass | 0.216s |  |
| Math Calculation | ✅ Pass | 0.213s |  |
| Basic Echo Function | ✅ Pass | 0.208s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.219s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.226s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.217s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.204s |  |
| Search Query Function | ✅ Pass | 0.206s |  |
| Ask Advice Function | ✅ Pass | 0.203s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.203s |  |
| Basic Context Memory Test | ✅ Pass | 0.203s |  |
| Function Argument Memory Test | ✅ Pass | 0.202s |  |
| Function Response Memory Test | ✅ Pass | 0.209s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.221s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.201s |  |
| Penetration Testing Methodology | ✅ Pass | 0.228s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.168s |  |
| Penetration Testing Framework | ✅ Pass | 0.192s |  |
| Web Application Security Scanner | ✅ Pass | 0.202s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.196s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.218s

---

### searcher (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.413s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.217s |  |
| Math Calculation | ✅ Pass | 0.210s |  |
| Basic Echo Function | ✅ Pass | 0.210s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.218s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.219s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.215s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.205s |  |
| Search Query Function | ✅ Pass | 0.204s |  |
| Ask Advice Function | ✅ Pass | 0.211s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.209s |  |
| Basic Context Memory Test | ✅ Pass | 0.210s |  |
| Function Argument Memory Test | ✅ Pass | 0.201s |  |
| Function Response Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.216s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Methodology | ✅ Pass | 0.222s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.184s |  |
| SQL Injection Attack Type | ✅ Pass | 0.179s |  |
| Penetration Testing Framework | ✅ Pass | 0.197s |  |
| Web Application Security Scanner | ✅ Pass | 0.193s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.071s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.211s

---

### enricher (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.237s |  |
| Text Transform Uppercase | ✅ Pass | 0.218s |  |
| Count from 1 to 5 | ✅ Pass | 0.212s |  |
| Math Calculation | ✅ Pass | 0.209s |  |
| Basic Echo Function | ✅ Pass | 0.210s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.213s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.204s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.210s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.208s |  |
| Search Query Function | ✅ Pass | 0.203s |  |
| Ask Advice Function | ✅ Pass | 0.204s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.205s |  |
| Basic Context Memory Test | ✅ Pass | 0.208s |  |
| Function Argument Memory Test | ✅ Pass | 0.199s |  |
| Function Response Memory Test | ✅ Pass | 0.209s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.213s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.434s |  |
| Penetration Testing Methodology | ✅ Pass | 0.222s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.130s |  |
| SQL Injection Attack Type | ✅ Pass | 0.204s |  |
| Penetration Testing Framework | ✅ Pass | 0.195s |  |
| Web Application Security Scanner | ✅ Pass | 0.136s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.215s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.213s

---

### coder (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.208s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.208s |  |
| Basic Echo Function | ✅ Pass | 0.204s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.208s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.214s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.210s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.209s |  |
| Search Query Function | ✅ Pass | 0.203s |  |
| Ask Advice Function | ✅ Pass | 0.207s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.211s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.204s |  |
| Function Response Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.438s |  |
| Penetration Testing Methodology | ✅ Pass | 0.232s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.164s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.169s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.233s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.217s

---

### installer (deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.235s |  |
| Text Transform Uppercase | ✅ Pass | 0.218s |  |
| Count from 1 to 5 | ✅ Pass | 0.208s |  |
| Math Calculation | ✅ Pass | 0.208s |  |
| Basic Echo Function | ✅ Pass | 0.201s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.210s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.208s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.211s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.208s |  |
| Search Query Function | ✅ Pass | 0.216s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.202s |  |
| Basic Context Memory Test | ✅ Pass | 0.208s |  |
| Function Argument Memory Test | ✅ Pass | 0.205s |  |
| Function Response Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.208s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.418s |  |
| Penetration Testing Methodology | ✅ Pass | 0.227s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.182s |  |
| SQL Injection Attack Type | ✅ Pass | 0.211s |  |
| Penetration Testing Framework | ✅ Pass | 0.193s |  |
| Web Application Security Scanner | ✅ Pass | 0.172s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.219s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.217s

---

### pentester (deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.235s |  |
| Text Transform Uppercase | ✅ Pass | 0.224s |  |
| Count from 1 to 5 | ✅ Pass | 0.210s |  |
| Math Calculation | ✅ Pass | 0.208s |  |
| Basic Echo Function | ✅ Pass | 0.208s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.216s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.204s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.209s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.210s |  |
| Search Query Function | ✅ Pass | 0.216s |  |
| Ask Advice Function | ✅ Pass | 0.202s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.210s |  |
| Basic Context Memory Test | ✅ Pass | 0.204s |  |
| Function Argument Memory Test | ✅ Pass | 0.207s |  |
| Function Response Memory Test | ✅ Pass | 0.206s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.265s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.413s |  |
| Penetration Testing Methodology | ✅ Pass | 0.231s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.198s |  |
| SQL Injection Attack Type | ✅ Pass | 0.203s |  |
| Penetration Testing Framework | ✅ Pass | 0.140s |  |
| Web Application Security Scanner | ✅ Pass | 0.297s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.234s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.224s

---

