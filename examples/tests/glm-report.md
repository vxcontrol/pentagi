# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 17:03:13 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | glm-4.5-air | false | 22/23 (95.65%) | 0.236s |
| simple_json | glm-4.5-air | false | 6/7 (85.71%) | 1.870s |
| primary_agent | glm-5-turbo | true | 22/23 (95.65%) | 0.232s |
| assistant | glm-5-turbo | true | 22/23 (95.65%) | 0.233s |
| generator | glm-5.2 | true | 23/23 (100.00%) | 0.233s |
| refiner | glm-5.2 | true | 23/23 (100.00%) | 0.231s |
| adviser | glm-5.2 | true | 22/23 (95.65%) | 0.228s |
| reflector | glm-4.5-air | false | 22/23 (95.65%) | 0.230s |
| searcher | glm-4.5-air | false | 23/23 (100.00%) | 0.231s |
| enricher | glm-4.5-air | false | 23/23 (100.00%) | 0.210s |
| coder | glm-5.2 | true | 23/23 (100.00%) | 0.214s |
| installer | glm-4.5-air | true | 23/23 (100.00%) | 0.213s |
| pentester | glm-5.2 | true | 23/23 (100.00%) | 0.210s |

**Total**: 277/283 (97.88%) successful tests
**Overall average latency**: 0.266s

## Detailed Results

### simple (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.428s |  |
| Text Transform Uppercase | ✅ Pass | 0.216s |  |
| Count from 1 to 5 | ✅ Pass | 0.214s |  |
| Math Calculation | ✅ Pass | 0.201s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.203s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.215s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.220s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.212s |  |
| Search Query Function | ✅ Pass | 0.207s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.249s |  |
| Basic Context Memory Test | ✅ Pass | 0.208s |  |
| Function Argument Memory Test | ✅ Pass | 0.218s |  |
| Function Response Memory Test | ✅ Pass | 0.496s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.256s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.206s |  |
| Penetration Testing Methodology | ✅ Pass | 0.205s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.213s |  |
| SQL Injection Attack Type | ✅ Pass | 0.208s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.205s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.220s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.236s

---

### simple_json (glm-4.5-air)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.448s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.375s |  |
| Project Information JSON | ✅ Pass | 1.654s |  |
| User Profile JSON | ✅ Pass | 1.684s |  |
| JSON Array Response Without Schema | ✅ Pass | 2.093s |  |
| Vulnerability Report Memory Test | ✅ Pass | 3.363s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ❌ Fail | 1.471s | structured output: response validation failed \(provider=openai model=zai/glm\-4\.5\-air choice=0 stop\_reason=stop\): response is not a single JSO... |

**Summary**: 6/7 (85.71%) successful tests

**Average latency**: 1.870s

---

### primary_agent (glm-5-turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.423s |  |
| Text Transform Uppercase | ✅ Pass | 0.214s |  |
| Count from 1 to 5 | ✅ Pass | 0.219s |  |
| Math Calculation | ✅ Pass | 0.206s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.202s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.219s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.221s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.209s |  |
| Search Query Function | ✅ Pass | 0.205s |  |
| Ask Advice Function | ✅ Pass | 0.204s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.213s |  |
| Basic Context Memory Test | ✅ Pass | 0.205s |  |
| Function Argument Memory Test | ✅ Pass | 0.219s |  |
| Function Response Memory Test | ✅ Pass | 0.480s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.239s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.206s |  |
| Penetration Testing Methodology | ✅ Pass | 0.215s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.205s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.204s |  |
| Web Application Security Scanner | ✅ Pass | 0.207s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.208s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.232s

---

### assistant (glm-5-turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.399s |  |
| Text Transform Uppercase | ✅ Pass | 0.226s |  |
| Count from 1 to 5 | ✅ Pass | 0.216s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.207s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.204s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.219s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.225s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.212s |  |
| Search Query Function | ✅ Pass | 0.203s |  |
| Ask Advice Function | ✅ Pass | 0.199s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.217s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.213s |  |
| Function Response Memory Test | ✅ Pass | 0.466s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.284s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Methodology | ✅ Pass | 0.217s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.204s |  |
| Penetration Testing Framework | ✅ Pass | 0.204s |  |
| Web Application Security Scanner | ✅ Pass | 0.206s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.201s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.233s

---

### generator (glm-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.389s |  |
| Text Transform Uppercase | ✅ Pass | 0.223s |  |
| Count from 1 to 5 | ✅ Pass | 0.219s |  |
| Math Calculation | ✅ Pass | 0.208s |  |
| Basic Echo Function | ✅ Pass | 0.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.204s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.231s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.218s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.207s |  |
| Search Query Function | ✅ Pass | 0.203s |  |
| Ask Advice Function | ✅ Pass | 0.202s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.252s |  |
| Basic Context Memory Test | ✅ Pass | 0.214s |  |
| Function Argument Memory Test | ✅ Pass | 0.205s |  |
| Function Response Memory Test | ✅ Pass | 0.458s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.243s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.216s |  |
| Penetration Testing Methodology | ✅ Pass | 0.211s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.204s |  |
| SQL Injection Attack Type | ✅ Pass | 0.208s |  |
| Penetration Testing Framework | ✅ Pass | 0.205s |  |
| Web Application Security Scanner | ✅ Pass | 0.207s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.210s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.233s

---

### refiner (glm-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.418s |  |
| Text Transform Uppercase | ✅ Pass | 0.207s |  |
| Count from 1 to 5 | ✅ Pass | 0.215s |  |
| Math Calculation | ✅ Pass | 0.209s |  |
| Basic Echo Function | ✅ Pass | 0.201s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.211s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.225s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.205s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.201s |  |
| Search Query Function | ✅ Pass | 0.213s |  |
| Ask Advice Function | ✅ Pass | 0.199s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.252s |  |
| Basic Context Memory Test | ✅ Pass | 0.210s |  |
| Function Argument Memory Test | ✅ Pass | 0.206s |  |
| Function Response Memory Test | ✅ Pass | 0.444s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.206s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.243s |  |
| Penetration Testing Methodology | ✅ Pass | 0.216s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.205s |  |
| SQL Injection Attack Type | ✅ Pass | 0.209s |  |
| Penetration Testing Framework | ✅ Pass | 0.207s |  |
| Web Application Security Scanner | ✅ Pass | 0.208s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.200s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.231s

---

### adviser (glm-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.389s |  |
| Text Transform Uppercase | ✅ Pass | 0.208s |  |
| Count from 1 to 5 | ✅ Pass | 0.213s |  |
| Math Calculation | ✅ Pass | 0.203s |  |
| Basic Echo Function | ✅ Pass | 0.201s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.221s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.228s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.207s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.198s |  |
| Search Query Function | ✅ Pass | 0.202s |  |
| Ask Advice Function | ✅ Pass | 0.196s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.215s |  |
| Basic Context Memory Test | ✅ Pass | 0.208s |  |
| Function Argument Memory Test | ✅ Pass | 0.208s |  |
| Function Response Memory Test | ✅ Pass | 0.431s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.210s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.251s |  |
| Penetration Testing Methodology | ✅ Pass | 0.212s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.215s |  |
| SQL Injection Attack Type | ✅ Pass | 0.213s |  |
| Penetration Testing Framework | ✅ Pass | 0.202s |  |
| Web Application Security Scanner | ✅ Pass | 0.208s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.199s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.228s

---

### reflector (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.418s |  |
| Text Transform Uppercase | ✅ Pass | 0.211s |  |
| Count from 1 to 5 | ✅ Pass | 0.220s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.204s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.212s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.215s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.210s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.201s |  |
| Search Query Function | ✅ Pass | 0.202s |  |
| Ask Advice Function | ✅ Pass | 0.197s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.253s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.224s |  |
| Function Response Memory Test | ✅ Pass | 0.410s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 0.216s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.237s |  |
| Penetration Testing Methodology | ✅ Pass | 0.208s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.213s |  |
| SQL Injection Attack Type | ✅ Pass | 0.204s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.211s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.199s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.230s

---

### searcher (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.418s |  |
| Text Transform Uppercase | ✅ Pass | 0.212s |  |
| Count from 1 to 5 | ✅ Pass | 0.218s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.209s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.215s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.203s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.204s |  |
| Search Query Function | ✅ Pass | 0.201s |  |
| Ask Advice Function | ✅ Pass | 0.203s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.229s |  |
| Basic Context Memory Test | ✅ Pass | 0.218s |  |
| Function Argument Memory Test | ✅ Pass | 0.207s |  |
| Function Response Memory Test | ✅ Pass | 0.400s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.243s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.239s |  |
| Penetration Testing Methodology | ✅ Pass | 0.210s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.204s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.209s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.231s

---

### enricher (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.211s |  |
| Text Transform Uppercase | ✅ Pass | 0.212s |  |
| Count from 1 to 5 | ✅ Pass | 0.206s |  |
| Math Calculation | ✅ Pass | 0.216s |  |
| Basic Echo Function | ✅ Pass | 0.208s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.201s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.208s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.204s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.202s |  |
| Search Query Function | ✅ Pass | 0.200s |  |
| Ask Advice Function | ✅ Pass | 0.202s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.219s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.210s |  |
| Function Response Memory Test | ✅ Pass | 0.208s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.239s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Methodology | ✅ Pass | 0.215s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.204s |  |
| SQL Injection Attack Type | ✅ Pass | 0.209s |  |
| Penetration Testing Framework | ✅ Pass | 0.208s |  |
| Web Application Security Scanner | ✅ Pass | 0.208s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.209s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.210s

---

### coder (glm-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.221s |  |
| Text Transform Uppercase | ✅ Pass | 0.216s |  |
| Count from 1 to 5 | ✅ Pass | 0.204s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.202s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.211s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.204s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.206s |  |
| Search Query Function | ✅ Pass | 0.205s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.218s |  |
| Basic Context Memory Test | ✅ Pass | 0.212s |  |
| Function Argument Memory Test | ✅ Pass | 0.220s |  |
| Function Response Memory Test | ✅ Pass | 0.259s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.237s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Methodology | ✅ Pass | 0.212s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.209s |  |
| Penetration Testing Framework | ✅ Pass | 0.210s |  |
| Web Application Security Scanner | ✅ Pass | 0.211s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.202s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.214s

---

### installer (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.219s |  |
| Text Transform Uppercase | ✅ Pass | 0.223s |  |
| Count from 1 to 5 | ✅ Pass | 0.202s |  |
| Math Calculation | ✅ Pass | 0.209s |  |
| Basic Echo Function | ✅ Pass | 0.205s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.207s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.210s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.203s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.205s |  |
| Search Query Function | ✅ Pass | 0.202s |  |
| Ask Advice Function | ✅ Pass | 0.210s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.260s |  |
| Basic Context Memory Test | ✅ Pass | 0.205s |  |
| Function Argument Memory Test | ✅ Pass | 0.204s |  |
| Function Response Memory Test | ✅ Pass | 0.247s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.237s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Methodology | ✅ Pass | 0.207s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.206s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.205s |  |
| Web Application Security Scanner | ✅ Pass | 0.217s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.202s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.213s

---

### pentester (glm-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.234s |  |
| Text Transform Uppercase | ✅ Pass | 0.221s |  |
| Count from 1 to 5 | ✅ Pass | 0.213s |  |
| Math Calculation | ✅ Pass | 0.212s |  |
| Basic Echo Function | ✅ Pass | 0.199s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.199s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.206s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.208s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.206s |  |
| Search Query Function | ✅ Pass | 0.200s |  |
| Ask Advice Function | ✅ Pass | 0.207s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.203s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.203s |  |
| Function Response Memory Test | ✅ Pass | 0.222s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.218s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Methodology | ✅ Pass | 0.212s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.204s |  |
| SQL Injection Attack Type | ✅ Pass | 0.208s |  |
| Penetration Testing Framework | ✅ Pass | 0.210s |  |
| Web Application Security Scanner | ✅ Pass | 0.218s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.207s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.210s

---

