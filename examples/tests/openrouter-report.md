# LLM Agent Testing Report

Generated: Fri, 25 Sep 2026 00:19:55 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek/deepseek-v4-flash | true | 24/25 (96.00%) | 6.951s |
| simple_json | openai/gpt-4.1-mini | false | 8/8 (100.00%) | 0.956s |
| primary_agent | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.440s |
| assistant | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.487s |
| generator | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.415s |
| refiner | deepseek/deepseek-v4-pro | true | 25/25 (100.00%) | 3.951s |
| adviser | deepseek/deepseek-v4.1-flash | true | 16/16 (100.00%) | 1.638s |
| reflector | deepseek/deepseek-v4-flash | true | 16/16 (100.00%) | 5.014s |
| searcher | deepseek/deepseek-v4-flash | true | 25/25 (100.00%) | 5.923s |
| enricher | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.713s |
| coder | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.493s |
| installer | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.410s |
| pentester | deepseek/deepseek-v4.1-flash | true | 25/25 (100.00%) | 1.766s |

**Total**: 289/290 (99.66%) successful tests
**Overall average latency**: 2.768s

## Detailed Results

### simple (deepseek/deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.962s |  |
| Text Transform Uppercase | ✅ Pass | 2.182s |  |
| Count from 1 to 5 | ✅ Pass | 3.141s |  |
| Math Calculation | ✅ Pass | 2.129s |  |
| Basic Echo Function | ✅ Pass | 1.361s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.151s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 12.245s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.985s |  |
| Answer Stops At The Output Limit | ❌ Fail | 60.000s | error decoding response: context deadline exceeded \(Client\.Timeout or context cancellation while reading body\) |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 3.729s |  |
| JSON Response Function | ✅ Pass | 4.478s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.226s |  |
| Ask Advice Function | ✅ Pass | 5.452s |  |
| Function Argument Memory Test | ✅ Pass | 1.893s |  |
| Basic Context Memory Test | ✅ Pass | 4.813s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.453s |  |
| Function Response Memory Test | ✅ Pass | 7.909s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.399s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 18.887s |  |
| Penetration Testing Methodology | ✅ Pass | 1.342s |  |
| SQL Injection Attack Type | ✅ Pass | 1.977s |  |
| Penetration Testing Framework | ✅ Pass | 1.946s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.760s |  |
| Web Application Security Scanner | ✅ Pass | 7.326s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.004s |  |

**Summary**: 24/25 (96.00%) successful tests

**Average latency**: 6.951s

---

### simple_json (openai/gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.425s |  |
| Person Information JSON | ✅ Pass | 1.009s |  |
| Project Information JSON | ✅ Pass | 1.029s |  |
| User Profile JSON | ✅ Pass | 1.031s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.117s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.098s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output Refuses A Non-Object Schema | structured_output | ✅ Pass | 0.000s |  |
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.934s |  |

**Summary**: 8/8 (100.00%) successful tests

**Average latency**: 0.956s

---

### primary_agent (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Answer Stops At The Output Limit | ✅ Pass | 11.663s |  |
| Simple Math | ✅ Pass | 1.432s |  |
| Text Transform Uppercase | ✅ Pass | 0.597s |  |
| Count from 1 to 5 | ✅ Pass | 0.624s |  |
| Math Calculation | ✅ Pass | 0.499s |  |
| Basic Echo Function | ✅ Pass | 0.912s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.523s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.503s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.791s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.393s |  |
| Search Query Function | ✅ Pass | 1.282s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.755s |  |
| Ask Advice Function | ✅ Pass | 1.432s |  |
| Basic Context Memory Test | ✅ Pass | 0.753s |  |
| Function Argument Memory Test | ✅ Pass | 0.575s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.830s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.686s |  |
| Function Response Memory Test | ✅ Pass | 3.470s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.832s |  |
| Penetration Testing Methodology | ✅ Pass | 0.560s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.246s |  |
| Penetration Testing Framework | ✅ Pass | 0.631s |  |
| SQL Injection Attack Type | ✅ Pass | 0.778s |  |
| Web Application Security Scanner | ✅ Pass | 0.548s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.670s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.440s

---

### assistant (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Answer Stops At The Output Limit | ✅ Pass | 8.913s |  |
| Simple Math | ✅ Pass | 0.981s |  |
| Text Transform Uppercase | ✅ Pass | 0.585s |  |
| Count from 1 to 5 | ✅ Pass | 0.667s |  |
| Math Calculation | ✅ Pass | 0.532s |  |
| Basic Echo Function | ✅ Pass | 1.050s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.555s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.628s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.643s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.500s |  |
| Ask Advice Function | ✅ Pass | 1.837s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.909s |  |
| Basic Context Memory Test | ✅ Pass | 0.751s |  |
| Function Argument Memory Test | ✅ Pass | 0.563s |  |
| Search Query Function | ✅ Pass | 4.472s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.820s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.594s |  |
| Function Response Memory Test | ✅ Pass | 2.503s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.187s |  |
| Penetration Testing Methodology | ✅ Pass | 0.588s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.038s |  |
| SQL Injection Attack Type | ✅ Pass | 0.756s |  |
| Web Application Security Scanner | ✅ Pass | 0.586s |  |
| Penetration Testing Framework | ✅ Pass | 0.745s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.750s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.487s

---

### generator (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.643s |  |
| Text Transform Uppercase | ✅ Pass | 0.544s |  |
| Count from 1 to 5 | ✅ Pass | 0.445s |  |
| Math Calculation | ✅ Pass | 0.536s |  |
| Basic Echo Function | ✅ Pass | 0.804s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.046s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.554s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.617s |  |
| Answer Stops At The Output Limit | ✅ Pass | 8.750s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.486s |  |
| Search Query Function | ✅ Pass | 1.248s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.591s |  |
| Ask Advice Function | ✅ Pass | 1.395s |  |
| Function Argument Memory Test | ✅ Pass | 0.605s |  |
| Basic Context Memory Test | ✅ Pass | 0.890s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.932s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.642s |  |
| Function Response Memory Test | ✅ Pass | 4.514s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.694s |  |
| Penetration Testing Methodology | ✅ Pass | 0.570s |  |
| SQL Injection Attack Type | ✅ Pass | 0.788s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.178s |  |
| Penetration Testing Framework | ✅ Pass | 0.654s |  |
| Web Application Security Scanner | ✅ Pass | 0.538s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.689s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.415s

---

### refiner (deepseek/deepseek-v4-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.900s |  |
| Text Transform Uppercase | ✅ Pass | 2.325s |  |
| Count from 1 to 5 | ✅ Pass | 1.617s |  |
| Math Calculation | ✅ Pass | 1.970s |  |
| Basic Echo Function | ✅ Pass | 2.318s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.364s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.334s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.339s |  |
| Answer Stops At The Output Limit | ✅ Pass | 23.615s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.980s |  |
| Search Query Function | ✅ Pass | 2.740s |  |
| Ask Advice Function | ✅ Pass | 3.537s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.057s |  |
| Function Argument Memory Test | ✅ Pass | 3.092s |  |
| Basic Context Memory Test | ✅ Pass | 3.311s |  |
| Function Response Memory Test | ✅ Pass | 2.657s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.595s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.504s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 7.449s |  |
| Penetration Testing Methodology | ✅ Pass | 3.870s |  |
| SQL Injection Attack Type | ✅ Pass | 2.586s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.947s |  |
| Penetration Testing Framework | ✅ Pass | 4.256s |  |
| Web Application Security Scanner | ✅ Pass | 2.237s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.164s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 3.951s

---

### adviser (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.581s |  |
| Text Transform Uppercase | ✅ Pass | 0.791s |  |
| Count from 1 to 5 | ✅ Pass | 0.564s |  |
| Math Calculation | ✅ Pass | 0.539s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.551s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.638s |  |
| Answer Stops At The Output Limit | ✅ Pass | 10.924s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Basic Context Memory Test | ✅ Pass | 0.704s |  |
| Function Argument Memory Test | ✅ Pass | 0.618s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.679s |  |
| Function Response Memory Test | ✅ Pass | 6.070s |  |
| Penetration Testing Methodology | ✅ Pass | 0.544s |  |
| SQL Injection Attack Type | ✅ Pass | 0.662s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.212s |  |
| Penetration Testing Framework | ✅ Pass | 0.597s |  |
| Web Application Security Scanner | ✅ Pass | 0.527s |  |

**Summary**: 16/16 (100.00%) successful tests

**Average latency**: 1.638s

---

### reflector (deepseek/deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.263s |  |
| Text Transform Uppercase | ✅ Pass | 1.900s |  |
| Count from 1 to 5 | ✅ Pass | 3.623s |  |
| Math Calculation | ✅ Pass | 3.395s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.846s |  |
| Answer Stops At The Output Limit | ✅ Pass | 15.745s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.422s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Function Argument Memory Test | ✅ Pass | 3.887s |  |
| Function Response Memory Test | ✅ Pass | 1.511s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.758s |  |
| Basic Context Memory Test | ✅ Pass | 16.472s |  |
| Penetration Testing Methodology | ✅ Pass | 6.122s |  |
| SQL Injection Attack Type | ✅ Pass | 1.836s |  |
| Penetration Testing Framework | ✅ Pass | 1.880s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.204s |  |
| Web Application Security Scanner | ✅ Pass | 6.358s |  |

**Summary**: 16/16 (100.00%) successful tests

**Average latency**: 5.014s

---

### searcher (deepseek/deepseek-v4-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.969s |  |
| Text Transform Uppercase | ✅ Pass | 1.384s |  |
| Count from 1 to 5 | ✅ Pass | 1.430s |  |
| Math Calculation | ✅ Pass | 1.183s |  |
| Basic Echo Function | ✅ Pass | 1.725s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.285s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.156s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.048s |  |
| Answer Stops At The Output Limit | ✅ Pass | 54.525s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.752s |  |
| Search Query Function | ✅ Pass | 3.079s |  |
| Ask Advice Function | ✅ Pass | 4.625s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.470s |  |
| Basic Context Memory Test | ✅ Pass | 2.695s |  |
| Function Argument Memory Test | ✅ Pass | 3.231s |  |
| Function Response Memory Test | ✅ Pass | 2.230s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.792s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.151s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 9.748s |  |
| Penetration Testing Methodology | ✅ Pass | 3.122s |  |
| SQL Injection Attack Type | ✅ Pass | 3.863s |  |
| Penetration Testing Framework | ✅ Pass | 4.410s |  |
| Web Application Security Scanner | ✅ Pass | 1.780s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.590s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.818s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 5.923s

---

### enricher (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Answer Stops At The Output Limit | ✅ Pass | 9.909s |  |
| Simple Math | ✅ Pass | 1.422s |  |
| Text Transform Uppercase | ✅ Pass | 0.734s |  |
| Count from 1 to 5 | ✅ Pass | 0.577s |  |
| Math Calculation | ✅ Pass | 0.500s |  |
| Basic Echo Function | ✅ Pass | 1.321s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.780s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.574s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.685s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.078s |  |
| Search Query Function | ✅ Pass | 3.279s |  |
| Ask Advice Function | ✅ Pass | 1.273s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.668s |  |
| Basic Context Memory Test | ✅ Pass | 0.616s |  |
| Function Argument Memory Test | ✅ Pass | 0.583s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.825s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.597s |  |
| Function Response Memory Test | ✅ Pass | 9.266s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.274s |  |
| Penetration Testing Methodology | ✅ Pass | 0.710s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.416s |  |
| SQL Injection Attack Type | ✅ Pass | 0.739s |  |
| Penetration Testing Framework | ✅ Pass | 0.719s |  |
| Web Application Security Scanner | ✅ Pass | 0.577s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.698s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.713s

---

### coder (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Answer Stops At The Output Limit | ✅ Pass | 9.525s |  |
| Simple Math | ✅ Pass | 1.269s |  |
| Text Transform Uppercase | ✅ Pass | 0.623s |  |
| Count from 1 to 5 | ✅ Pass | 0.528s |  |
| Math Calculation | ✅ Pass | 0.517s |  |
| Basic Echo Function | ✅ Pass | 0.746s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.603s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.519s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.952s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.999s |  |
| Search Query Function | ✅ Pass | 2.038s |  |
| Ask Advice Function | ✅ Pass | 1.264s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.818s |  |
| Basic Context Memory Test | ✅ Pass | 0.704s |  |
| Function Argument Memory Test | ✅ Pass | 0.557s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.841s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.636s |  |
| Function Response Memory Test | ✅ Pass | 5.634s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.195s |  |
| Penetration Testing Methodology | ✅ Pass | 0.581s |  |
| SQL Injection Attack Type | ✅ Pass | 0.736s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.222s |  |
| Penetration Testing Framework | ✅ Pass | 0.577s |  |
| Web Application Security Scanner | ✅ Pass | 0.561s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.659s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.493s

---

### installer (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.734s |  |
| Text Transform Uppercase | ✅ Pass | 0.637s |  |
| Answer Stops At The Output Limit | ✅ Pass | 8.143s |  |
| Count from 1 to 5 | ✅ Pass | 0.555s |  |
| Math Calculation | ✅ Pass | 0.569s |  |
| Basic Echo Function | ✅ Pass | 0.827s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.567s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.721s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.789s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.126s |  |
| Search Query Function | ✅ Pass | 1.252s |  |
| Ask Advice Function | ✅ Pass | 1.153s |  |
| Basic Context Memory Test | ✅ Pass | 0.930s |  |
| Function Argument Memory Test | ✅ Pass | 0.735s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.610s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.245s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.549s |  |
| Function Response Memory Test | ✅ Pass | 3.452s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 2.889s |  |
| Penetration Testing Methodology | ✅ Pass | 0.539s |  |
| SQL Injection Attack Type | ✅ Pass | 0.818s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.195s |  |
| Web Application Security Scanner | ✅ Pass | 0.544s |  |
| Penetration Testing Framework | ✅ Pass | 0.856s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.804s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.410s

---

### pentester (deepseek/deepseek-v4.1-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 9.296s |  |
| Answer Stops At The Output Limit | ✅ Pass | 9.559s |  |
| Text Transform Uppercase | ✅ Pass | 0.617s |  |
| Count from 1 to 5 | ✅ Pass | 0.563s |  |
| Math Calculation | ✅ Pass | 0.512s |  |
| Basic Echo Function | ✅ Pass | 0.797s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.539s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.666s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.749s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 1.325s |  |
| JSON Response Function | ✅ Pass | 1.461s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.588s |  |
| Basic Context Memory Test | ✅ Pass | 0.753s |  |
| Ask Advice Function | ✅ Pass | 1.604s |  |
| Function Argument Memory Test | ✅ Pass | 0.658s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.832s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.651s |  |
| Function Response Memory Test | ✅ Pass | 4.697s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.070s |  |
| Penetration Testing Methodology | ✅ Pass | 0.696s |  |
| SQL Injection Attack Type | ✅ Pass | 0.827s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.026s |  |
| Web Application Security Scanner | ✅ Pass | 0.583s |  |
| Penetration Testing Framework | ✅ Pass | 0.746s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.323s |  |

**Summary**: 25/25 (100.00%) successful tests

**Average latency**: 1.766s

---

