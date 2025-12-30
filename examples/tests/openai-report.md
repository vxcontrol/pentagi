# LLM Agent Testing Report

Generated: Tue, 30 Dec 2025 21:52:12 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-4.1-mini | false | 23/23 (100.00%) | 1.193s |
| simple_json | gpt-4.1-mini | false | 5/5 (100.00%) | 1.685s |
| primary_agent | gpt-5 | true | 22/23 (95.65%) | 2.398s |
| assistant | gpt-5 | true | 23/23 (100.00%) | 4.374s |
| generator | gpt-5.2 | true | 23/23 (100.00%) | 1.171s |
| refiner | gpt-5 | true | 23/23 (100.00%) | 7.962s |
| adviser | gpt-5.2 | true | 23/23 (100.00%) | 1.088s |
| reflector | o4-mini | true | 23/23 (100.00%) | 1.961s |
| searcher | gpt-4.1-mini | false | 23/23 (100.00%) | 1.285s |
| enricher | gpt-4.1-mini | false | 23/23 (100.00%) | 1.196s |
| coder | gpt-5.2 | true | 23/23 (100.00%) | 1.143s |
| installer | o4-mini | true | 23/23 (100.00%) | 1.412s |
| pentester | o4-mini | true | 23/23 (100.00%) | 1.431s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 2.208s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.345s |  |
| Text Transform Uppercase | ✅ Pass | 0.938s |  |
| Count from 1 to 5 | ✅ Pass | 1.104s |  |
| Math Calculation | ✅ Pass | 0.844s |  |
| Basic Echo Function | ✅ Pass | 0.941s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.168s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.851s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.913s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.589s |  |
| Search Query Function | ✅ Pass | 0.978s |  |
| Ask Advice Function | ✅ Pass | 1.137s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.856s |  |
| Basic Context Memory Test | ✅ Pass | 1.028s |  |
| Function Argument Memory Test | ✅ Pass | 0.536s |  |
| Function Response Memory Test | ✅ Pass | 0.840s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.493s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.722s |  |
| Penetration Testing Methodology | ✅ Pass | 1.125s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.421s |  |
| SQL Injection Attack Type | ✅ Pass | 1.169s |  |
| Penetration Testing Framework | ✅ Pass | 1.364s |  |
| Web Application Security Scanner | ✅ Pass | 1.009s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.054s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.193s

---

### simple_json (gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.375s |  |
| Project Information JSON | ✅ Pass | 1.162s |  |
| User Profile JSON | ✅ Pass | 1.512s |  |
| Vulnerability Report Memory Test | ✅ Pass | 2.950s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.425s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.685s

---

### primary_agent (gpt-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.558s |  |
| Text Transform Uppercase | ✅ Pass | 1.614s |  |
| Count from 1 to 5 | ✅ Pass | 2.094s |  |
| Math Calculation | ✅ Pass | 1.280s |  |
| Basic Echo Function | ✅ Pass | 3.377s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.726s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.508s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.691s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.920s |  |
| Search Query Function | ❌ Fail | 2.593s | expected function 'search' not found in tool calls: expected function search not found in tool calls |
| Ask Advice Function | ✅ Pass | 1.493s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.694s |  |
| Basic Context Memory Test | ✅ Pass | 1.888s |  |
| Function Argument Memory Test | ✅ Pass | 1.593s |  |
| Function Response Memory Test | ✅ Pass | 1.861s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.580s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.997s |  |
| Penetration Testing Methodology | ✅ Pass | 2.824s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.094s |  |
| SQL Injection Attack Type | ✅ Pass | 3.401s |  |
| Penetration Testing Framework | ✅ Pass | 2.052s |  |
| Web Application Security Scanner | ✅ Pass | 2.026s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.283s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.398s

---

### assistant (gpt-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.219s |  |
| Count from 1 to 5 | ✅ Pass | 2.010s |  |
| Text Transform Uppercase | ✅ Pass | 3.753s |  |
| Math Calculation | ✅ Pass | 1.551s |  |
| Basic Echo Function | ✅ Pass | 4.355s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.983s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.594s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.638s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.264s |  |
| Ask Advice Function | ✅ Pass | 2.267s |  |
| Search Query Function | ✅ Pass | 5.773s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.102s |  |
| Basic Context Memory Test | ✅ Pass | 2.982s |  |
| Function Argument Memory Test | ✅ Pass | 4.308s |  |
| Function Response Memory Test | ✅ Pass | 2.774s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.913s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 9.296s |  |
| Penetration Testing Methodology | ✅ Pass | 4.770s |  |
| SQL Injection Attack Type | ✅ Pass | 2.563s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.417s |  |
| Penetration Testing Framework | ✅ Pass | 3.382s |  |
| Web Application Security Scanner | ✅ Pass | 3.693s |  |
| Penetration Testing Tool Selection | ✅ Pass | 8.991s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.374s

---

### generator (gpt-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.455s |  |
| Text Transform Uppercase | ✅ Pass | 0.882s |  |
| Count from 1 to 5 | ✅ Pass | 1.002s |  |
| Math Calculation | ✅ Pass | 0.958s |  |
| Basic Echo Function | ✅ Pass | 1.208s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.065s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.842s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.056s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.309s |  |
| Search Query Function | ✅ Pass | 2.105s |  |
| Ask Advice Function | ✅ Pass | 0.988s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.177s |  |
| Basic Context Memory Test | ✅ Pass | 1.060s |  |
| Function Argument Memory Test | ✅ Pass | 1.224s |  |
| Function Response Memory Test | ✅ Pass | 0.943s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.117s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.013s |  |
| Penetration Testing Methodology | ✅ Pass | 1.084s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.758s |  |
| SQL Injection Attack Type | ✅ Pass | 0.873s |  |
| Penetration Testing Framework | ✅ Pass | 1.014s |  |
| Web Application Security Scanner | ✅ Pass | 0.936s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.850s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.171s

---

### refiner (gpt-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.602s |  |
| Text Transform Uppercase | ✅ Pass | 4.781s |  |
| Count from 1 to 5 | ✅ Pass | 3.584s |  |
| Math Calculation | ✅ Pass | 2.556s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.796s |  |
| Basic Echo Function | ✅ Pass | 8.731s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.596s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.906s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 7.602s |  |
| Ask Advice Function | ✅ Pass | 2.785s |  |
| Search Query Function | ✅ Pass | 6.420s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.012s |  |
| Basic Context Memory Test | ✅ Pass | 4.574s |  |
| Function Argument Memory Test | ✅ Pass | 3.964s |  |
| Function Response Memory Test | ✅ Pass | 8.652s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 12.610s |  |
| Penetration Testing Methodology | ✅ Pass | 7.424s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 22.489s |  |
| Penetration Testing Framework | ✅ Pass | 4.448s |  |
| SQL Injection Attack Type | ✅ Pass | 10.906s |  |
| Web Application Security Scanner | ✅ Pass | 4.773s |  |
| Vulnerability Assessment Tools | ✅ Pass | 27.167s |  |
| Penetration Testing Tool Selection | ✅ Pass | 20.742s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.962s

---

### adviser (gpt-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.616s |  |
| Text Transform Uppercase | ✅ Pass | 0.809s |  |
| Count from 1 to 5 | ✅ Pass | 0.914s |  |
| Math Calculation | ✅ Pass | 0.984s |  |
| Basic Echo Function | ✅ Pass | 1.125s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.900s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.744s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.051s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.171s |  |
| Search Query Function | ✅ Pass | 1.160s |  |
| Ask Advice Function | ✅ Pass | 1.132s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.766s |  |
| Basic Context Memory Test | ✅ Pass | 0.986s |  |
| Function Argument Memory Test | ✅ Pass | 0.770s |  |
| Function Response Memory Test | ✅ Pass | 0.838s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.820s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.298s |  |
| Penetration Testing Methodology | ✅ Pass | 1.085s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.709s |  |
| SQL Injection Attack Type | ✅ Pass | 0.961s |  |
| Penetration Testing Framework | ✅ Pass | 1.019s |  |
| Web Application Security Scanner | ✅ Pass | 0.852s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.293s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.088s

---

### reflector (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.896s |  |
| Text Transform Uppercase | ✅ Pass | 1.654s |  |
| Count from 1 to 5 | ✅ Pass | 1.474s |  |
| Math Calculation | ✅ Pass | 1.209s |  |
| Basic Echo Function | ✅ Pass | 2.228s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.244s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.376s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.889s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.780s |  |
| Search Query Function | ✅ Pass | 1.932s |  |
| Ask Advice Function | ✅ Pass | 2.172s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.279s |  |
| Basic Context Memory Test | ✅ Pass | 2.238s |  |
| Function Argument Memory Test | ✅ Pass | 1.562s |  |
| Function Response Memory Test | ✅ Pass | 1.500s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.101s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.689s |  |
| Penetration Testing Methodology | ✅ Pass | 1.587s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.039s |  |
| SQL Injection Attack Type | ✅ Pass | 2.628s |  |
| Penetration Testing Framework | ✅ Pass | 2.167s |  |
| Web Application Security Scanner | ✅ Pass | 1.770s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.672s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.961s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.554s |  |
| Text Transform Uppercase | ✅ Pass | 0.880s |  |
| Count from 1 to 5 | ✅ Pass | 2.007s |  |
| Math Calculation | ✅ Pass | 1.699s |  |
| Basic Echo Function | ✅ Pass | 0.960s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.831s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.981s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.857s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.016s |  |
| Search Query Function | ✅ Pass | 0.998s |  |
| Ask Advice Function | ✅ Pass | 1.350s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.828s |  |
| Basic Context Memory Test | ✅ Pass | 1.941s |  |
| Function Argument Memory Test | ✅ Pass | 0.699s |  |
| Function Response Memory Test | ✅ Pass | 1.742s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.480s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.766s |  |
| Penetration Testing Methodology | ✅ Pass | 0.992s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.541s |  |
| SQL Injection Attack Type | ✅ Pass | 1.234s |  |
| Penetration Testing Framework | ✅ Pass | 1.102s |  |
| Web Application Security Scanner | ✅ Pass | 1.026s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.054s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.285s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.305s |  |
| Text Transform Uppercase | ✅ Pass | 0.974s |  |
| Count from 1 to 5 | ✅ Pass | 1.050s |  |
| Math Calculation | ✅ Pass | 0.849s |  |
| Basic Echo Function | ✅ Pass | 1.069s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.771s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.799s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.864s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.451s |  |
| Search Query Function | ✅ Pass | 0.947s |  |
| Ask Advice Function | ✅ Pass | 0.990s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.003s |  |
| Basic Context Memory Test | ✅ Pass | 1.182s |  |
| Function Argument Memory Test | ✅ Pass | 0.587s |  |
| Function Response Memory Test | ✅ Pass | 0.891s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.908s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.697s |  |
| Penetration Testing Methodology | ✅ Pass | 1.092s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.689s |  |
| SQL Injection Attack Type | ✅ Pass | 1.191s |  |
| Penetration Testing Framework | ✅ Pass | 2.090s |  |
| Web Application Security Scanner | ✅ Pass | 1.004s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.096s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.196s

---

### coder (gpt-5.2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.792s |  |
| Text Transform Uppercase | ✅ Pass | 0.826s |  |
| Count from 1 to 5 | ✅ Pass | 1.000s |  |
| Math Calculation | ✅ Pass | 0.972s |  |
| Basic Echo Function | ✅ Pass | 1.012s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.005s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.965s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.252s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.141s |  |
| Search Query Function | ✅ Pass | 2.035s |  |
| Ask Advice Function | ✅ Pass | 1.069s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.970s |  |
| Basic Context Memory Test | ✅ Pass | 0.918s |  |
| Function Argument Memory Test | ✅ Pass | 0.991s |  |
| Function Response Memory Test | ✅ Pass | 1.008s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.441s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.198s |  |
| Penetration Testing Methodology | ✅ Pass | 0.735s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.374s |  |
| SQL Injection Attack Type | ✅ Pass | 0.970s |  |
| Penetration Testing Framework | ✅ Pass | 1.105s |  |
| Web Application Security Scanner | ✅ Pass | 1.026s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.481s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.143s

---

### installer (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.322s |  |
| Text Transform Uppercase | ✅ Pass | 0.969s |  |
| Count from 1 to 5 | ✅ Pass | 1.351s |  |
| Math Calculation | ✅ Pass | 1.119s |  |
| Basic Echo Function | ✅ Pass | 1.098s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.970s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.794s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.401s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.140s |  |
| Search Query Function | ✅ Pass | 1.055s |  |
| Ask Advice Function | ✅ Pass | 1.408s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.950s |  |
| Basic Context Memory Test | ✅ Pass | 1.476s |  |
| Function Argument Memory Test | ✅ Pass | 1.268s |  |
| Function Response Memory Test | ✅ Pass | 1.246s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.705s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.621s |  |
| Penetration Testing Methodology | ✅ Pass | 1.381s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.752s |  |
| SQL Injection Attack Type | ✅ Pass | 2.091s |  |
| Penetration Testing Framework | ✅ Pass | 1.734s |  |
| Web Application Security Scanner | ✅ Pass | 1.563s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.054s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.412s

---

### pentester (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.179s |  |
| Text Transform Uppercase | ✅ Pass | 1.753s |  |
| Count from 1 to 5 | ✅ Pass | 1.457s |  |
| Math Calculation | ✅ Pass | 1.092s |  |
| Basic Echo Function | ✅ Pass | 1.328s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.011s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.989s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.029s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.473s |  |
| Search Query Function | ✅ Pass | 1.118s |  |
| Ask Advice Function | ✅ Pass | 1.183s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.980s |  |
| Basic Context Memory Test | ✅ Pass | 1.691s |  |
| Function Argument Memory Test | ✅ Pass | 1.430s |  |
| Function Response Memory Test | ✅ Pass | 1.561s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.565s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.519s |  |
| Penetration Testing Methodology | ✅ Pass | 1.771s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.534s |  |
| SQL Injection Attack Type | ✅ Pass | 1.733s |  |
| Penetration Testing Framework | ✅ Pass | 1.761s |  |
| Web Application Security Scanner | ✅ Pass | 1.499s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.253s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.431s

---

