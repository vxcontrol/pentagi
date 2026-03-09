# LLM Agent Testing Report

Generated: Mon, 09 Mar 2026 18:50:45 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.203s |
| simple_json | Qwen/Qwen3.5-27B-FP8 | true | 5/5 (100.00%) | 0.192s |
| primary_agent | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.202s |
| assistant | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.202s |
| generator | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.204s |
| refiner | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.206s |
| adviser | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.204s |
| reflector | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.193s |
| searcher | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.201s |
| enricher | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.196s |
| coder | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.190s |
| installer | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.195s |
| pentester | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 0.196s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 0.199s

## Detailed Results

### simple (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.300s |  |
| Text Transform Uppercase | ✅ Pass | 0.211s |  |
| Count from 1 to 5 | ✅ Pass | 0.258s |  |
| Math Calculation | ✅ Pass | 0.203s |  |
| Basic Echo Function | ✅ Pass | 0.183s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.182s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.173s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.215s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.200s |  |
| Search Query Function | ✅ Pass | 0.182s |  |
| Ask Advice Function | ✅ Pass | 0.191s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.217s |  |
| Basic Context Memory Test | ✅ Pass | 0.182s |  |
| Function Argument Memory Test | ✅ Pass | 0.192s |  |
| Function Response Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.189s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.213s |  |
| Penetration Testing Methodology | ✅ Pass | 0.198s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.198s |  |
| SQL Injection Attack Type | ✅ Pass | 0.183s |  |
| Penetration Testing Framework | ✅ Pass | 0.198s |  |
| Web Application Security Scanner | ✅ Pass | 0.202s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.179s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.203s

---

### simple_json (Qwen/Qwen3.5-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.196s |  |
| Person Information JSON | ✅ Pass | 0.185s |  |
| Project Information JSON | ✅ Pass | 0.191s |  |
| User Profile JSON | ✅ Pass | 0.191s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.195s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.192s

---

### primary_agent (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.202s |  |
| Text Transform Uppercase | ✅ Pass | 0.243s |  |
| Count from 1 to 5 | ✅ Pass | 0.253s |  |
| Math Calculation | ✅ Pass | 0.185s |  |
| Basic Echo Function | ✅ Pass | 0.197s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.191s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.180s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.198s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.192s |  |
| Search Query Function | ✅ Pass | 0.188s |  |
| Ask Advice Function | ✅ Pass | 0.208s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.209s |  |
| Basic Context Memory Test | ✅ Pass | 0.180s |  |
| Function Argument Memory Test | ✅ Pass | 0.206s |  |
| Function Response Memory Test | ✅ Pass | 0.196s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.215s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.216s |  |
| Penetration Testing Methodology | ✅ Pass | 0.199s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.199s |  |
| SQL Injection Attack Type | ✅ Pass | 0.196s |  |
| Penetration Testing Framework | ✅ Pass | 0.207s |  |
| Web Application Security Scanner | ✅ Pass | 0.182s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.188s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.202s

---

### assistant (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.261s |  |
| Text Transform Uppercase | ✅ Pass | 0.243s |  |
| Count from 1 to 5 | ✅ Pass | 0.223s |  |
| Math Calculation | ✅ Pass | 0.178s |  |
| Basic Echo Function | ✅ Pass | 0.197s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.193s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.188s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.185s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.172s |  |
| Search Query Function | ✅ Pass | 0.198s |  |
| Ask Advice Function | ✅ Pass | 0.211s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.188s |  |
| Basic Context Memory Test | ✅ Pass | 0.189s |  |
| Function Argument Memory Test | ✅ Pass | 0.196s |  |
| Function Response Memory Test | ✅ Pass | 0.190s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.224s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Methodology | ✅ Pass | 0.188s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.205s |  |
| SQL Injection Attack Type | ✅ Pass | 0.212s |  |
| Penetration Testing Framework | ✅ Pass | 0.209s |  |
| Web Application Security Scanner | ✅ Pass | 0.191s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.186s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.202s

---

### generator (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.315s |  |
| Text Transform Uppercase | ✅ Pass | 0.227s |  |
| Count from 1 to 5 | ✅ Pass | 0.220s |  |
| Math Calculation | ✅ Pass | 0.185s |  |
| Basic Echo Function | ✅ Pass | 0.194s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.202s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.178s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.211s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.186s |  |
| Search Query Function | ✅ Pass | 0.199s |  |
| Ask Advice Function | ✅ Pass | 0.221s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.186s |  |
| Basic Context Memory Test | ✅ Pass | 0.181s |  |
| Function Argument Memory Test | ✅ Pass | 0.197s |  |
| Function Response Memory Test | ✅ Pass | 0.187s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.197s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.214s |  |
| Penetration Testing Methodology | ✅ Pass | 0.191s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.203s |  |
| SQL Injection Attack Type | ✅ Pass | 0.217s |  |
| Penetration Testing Framework | ✅ Pass | 0.193s |  |
| Web Application Security Scanner | ✅ Pass | 0.186s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.200s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.204s

---

### refiner (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.289s |  |
| Text Transform Uppercase | ✅ Pass | 0.240s |  |
| Count from 1 to 5 | ✅ Pass | 0.211s |  |
| Math Calculation | ✅ Pass | 0.193s |  |
| Basic Echo Function | ✅ Pass | 0.216s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.225s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.190s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.217s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.218s |  |
| Search Query Function | ✅ Pass | 0.201s |  |
| Ask Advice Function | ✅ Pass | 0.204s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.179s |  |
| Basic Context Memory Test | ✅ Pass | 0.187s |  |
| Function Argument Memory Test | ✅ Pass | 0.193s |  |
| Function Response Memory Test | ✅ Pass | 0.192s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.185s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.198s |  |
| Penetration Testing Methodology | ✅ Pass | 0.181s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.213s |  |
| SQL Injection Attack Type | ✅ Pass | 0.214s |  |
| Penetration Testing Framework | ✅ Pass | 0.182s |  |
| Web Application Security Scanner | ✅ Pass | 0.190s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.204s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.206s

---

### adviser (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.280s |  |
| Text Transform Uppercase | ✅ Pass | 0.186s |  |
| Count from 1 to 5 | ✅ Pass | 0.186s |  |
| Math Calculation | ✅ Pass | 0.195s |  |
| Basic Echo Function | ✅ Pass | 0.229s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.222s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.204s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.220s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.222s |  |
| Search Query Function | ✅ Pass | 0.204s |  |
| Ask Advice Function | ✅ Pass | 0.198s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.186s |  |
| Basic Context Memory Test | ✅ Pass | 0.188s |  |
| Function Argument Memory Test | ✅ Pass | 0.198s |  |
| Function Response Memory Test | ✅ Pass | 0.191s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.191s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.192s |  |
| Penetration Testing Methodology | ✅ Pass | 0.191s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.223s |  |
| SQL Injection Attack Type | ✅ Pass | 0.184s |  |
| Penetration Testing Framework | ✅ Pass | 0.193s |  |
| Web Application Security Scanner | ✅ Pass | 0.192s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.197s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.204s

---

### reflector (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.075s |  |
| Text Transform Uppercase | ✅ Pass | 0.191s |  |
| Count from 1 to 5 | ✅ Pass | 0.206s |  |
| Math Calculation | ✅ Pass | 0.194s |  |
| Basic Echo Function | ✅ Pass | 0.236s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.225s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.230s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.202s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.204s |  |
| Search Query Function | ✅ Pass | 0.208s |  |
| Ask Advice Function | ✅ Pass | 0.184s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.178s |  |
| Basic Context Memory Test | ✅ Pass | 0.203s |  |
| Function Argument Memory Test | ✅ Pass | 0.187s |  |
| Function Response Memory Test | ✅ Pass | 0.180s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.198s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.186s |  |
| Penetration Testing Methodology | ✅ Pass | 0.189s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.218s |  |
| SQL Injection Attack Type | ✅ Pass | 0.180s |  |
| Penetration Testing Framework | ✅ Pass | 0.182s |  |
| Web Application Security Scanner | ✅ Pass | 0.184s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.197s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.193s

---

### searcher (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.299s |  |
| Text Transform Uppercase | ✅ Pass | 0.194s |  |
| Count from 1 to 5 | ✅ Pass | 0.189s |  |
| Math Calculation | ✅ Pass | 0.184s |  |
| Basic Echo Function | ✅ Pass | 0.225s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.203s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.229s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.213s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.190s |  |
| Search Query Function | ✅ Pass | 0.203s |  |
| Ask Advice Function | ✅ Pass | 0.195s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.183s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.195s |  |
| Function Response Memory Test | ✅ Pass | 0.190s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.203s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.182s |  |
| Penetration Testing Methodology | ✅ Pass | 0.184s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.199s |  |
| SQL Injection Attack Type | ✅ Pass | 0.185s |  |
| Penetration Testing Framework | ✅ Pass | 0.187s |  |
| Web Application Security Scanner | ✅ Pass | 0.192s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.189s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.201s

---

### enricher (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.257s |  |
| Text Transform Uppercase | ✅ Pass | 0.196s |  |
| Count from 1 to 5 | ✅ Pass | 0.190s |  |
| Math Calculation | ✅ Pass | 0.194s |  |
| Basic Echo Function | ✅ Pass | 0.193s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.192s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.213s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.196s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.213s |  |
| Search Query Function | ✅ Pass | 0.185s |  |
| Ask Advice Function | ✅ Pass | 0.184s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.185s |  |
| Basic Context Memory Test | ✅ Pass | 0.200s |  |
| Function Argument Memory Test | ✅ Pass | 0.190s |  |
| Function Response Memory Test | ✅ Pass | 0.199s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.193s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.185s |  |
| Penetration Testing Methodology | ✅ Pass | 0.194s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.186s |  |
| SQL Injection Attack Type | ✅ Pass | 0.181s |  |
| Penetration Testing Framework | ✅ Pass | 0.197s |  |
| Web Application Security Scanner | ✅ Pass | 0.195s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.186s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.196s

---

### coder (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.203s |  |
| Text Transform Uppercase | ✅ Pass | 0.206s |  |
| Count from 1 to 5 | ✅ Pass | 0.205s |  |
| Math Calculation | ✅ Pass | 0.192s |  |
| Basic Echo Function | ✅ Pass | 0.173s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.185s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.189s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.186s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.193s |  |
| Search Query Function | ✅ Pass | 0.187s |  |
| Ask Advice Function | ✅ Pass | 0.175s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.195s |  |
| Basic Context Memory Test | ✅ Pass | 0.180s |  |
| Function Argument Memory Test | ✅ Pass | 0.179s |  |
| Function Response Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.190s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.182s |  |
| Penetration Testing Methodology | ✅ Pass | 0.182s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.192s |  |
| SQL Injection Attack Type | ✅ Pass | 0.182s |  |
| Penetration Testing Framework | ✅ Pass | 0.197s |  |
| Web Application Security Scanner | ✅ Pass | 0.191s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.188s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.190s

---

### installer (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.190s |  |
| Text Transform Uppercase | ✅ Pass | 0.233s |  |
| Count from 1 to 5 | ✅ Pass | 0.195s |  |
| Math Calculation | ✅ Pass | 0.188s |  |
| Basic Echo Function | ✅ Pass | 0.172s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.180s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.232s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.197s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.219s |  |
| Search Query Function | ✅ Pass | 0.206s |  |
| Ask Advice Function | ✅ Pass | 0.182s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.186s |  |
| Basic Context Memory Test | ✅ Pass | 0.183s |  |
| Function Argument Memory Test | ✅ Pass | 0.183s |  |
| Function Response Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.200s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.192s |  |
| Penetration Testing Methodology | ✅ Pass | 0.182s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.195s |  |
| SQL Injection Attack Type | ✅ Pass | 0.182s |  |
| Penetration Testing Framework | ✅ Pass | 0.200s |  |
| Web Application Security Scanner | ✅ Pass | 0.185s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.194s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.195s

---

### pentester (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.194s |  |
| Text Transform Uppercase | ✅ Pass | 0.233s |  |
| Count from 1 to 5 | ✅ Pass | 0.210s |  |
| Math Calculation | ✅ Pass | 0.184s |  |
| Basic Echo Function | ✅ Pass | 0.185s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.178s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.210s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.178s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.191s |  |
| Search Query Function | ✅ Pass | 0.208s |  |
| Ask Advice Function | ✅ Pass | 0.199s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.173s |  |
| Basic Context Memory Test | ✅ Pass | 0.190s |  |
| Function Argument Memory Test | ✅ Pass | 0.199s |  |
| Function Response Memory Test | ✅ Pass | 0.180s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.212s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.193s |  |
| Penetration Testing Methodology | ✅ Pass | 0.196s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.197s |  |
| SQL Injection Attack Type | ✅ Pass | 0.213s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.185s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.186s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.196s

---

