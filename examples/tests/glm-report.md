# LLM Agent Testing Report

Generated: Thu, 28 May 2026 12:25:41 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | glm-4.5-air | false | 23/23 (100.00%) | 3.213s |
| simple_json | glm-4.5-air | false | 5/5 (100.00%) | 3.923s |
| primary_agent | glm-5-turbo | true | 23/23 (100.00%) | 4.488s |
| assistant | glm-5-turbo | true | 22/23 (95.65%) | 3.383s |
| generator | glm-5.1 | true | 23/23 (100.00%) | 7.081s |
| refiner | glm-5.1 | true | 22/23 (95.65%) | 6.864s |
| adviser | glm-5.1 | true | 23/23 (100.00%) | 6.553s |
| reflector | glm-4.5-air | true | 23/23 (100.00%) | 2.760s |
| searcher | glm-4.5-air | true | 23/23 (100.00%) | 2.889s |
| enricher | glm-4.5-air | true | 23/23 (100.00%) | 3.001s |
| coder | glm-5.1 | true | 22/23 (95.65%) | 7.172s |
| installer | glm-4.5-air | true | 23/23 (100.00%) | 6.146s |
| pentester | glm-5.1 | true | 23/23 (100.00%) | 6.772s |

**Total**: 278/281 (98.93%) successful tests
**Overall average latency**: 5.007s

## Detailed Results

### simple (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.356s |  |
| Text Transform Uppercase | ✅ Pass | 1.140s |  |
| Count from 1 to 5 | ✅ Pass | 1.534s |  |
| Math Calculation | ✅ Pass | 1.975s |  |
| Basic Echo Function | ✅ Pass | 1.529s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.911s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.859s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.170s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.415s |  |
| Search Query Function | ✅ Pass | 3.528s |  |
| Ask Advice Function | ✅ Pass | 1.866s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.467s |  |
| Basic Context Memory Test | ✅ Pass | 1.478s |  |
| Function Argument Memory Test | ✅ Pass | 1.072s |  |
| Function Response Memory Test | ✅ Pass | 2.755s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.953s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.161s |  |
| Penetration Testing Methodology | ✅ Pass | 7.480s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.361s |  |
| SQL Injection Attack Type | ✅ Pass | 1.326s |  |
| Penetration Testing Framework | ✅ Pass | 7.132s |  |
| Web Application Security Scanner | ✅ Pass | 3.189s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.219s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.213s

---

### simple_json (glm-4.5-air)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 4.516s |  |
| Vulnerability Report Memory Test | ✅ Pass | 5.454s |  |
| Project Information JSON | ✅ Pass | 1.830s |  |
| User Profile JSON | ✅ Pass | 3.002s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 4.808s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 3.923s

---

### primary_agent (glm-5-turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.935s |  |
| Text Transform Uppercase | ✅ Pass | 3.695s |  |
| Count from 1 to 5 | ✅ Pass | 3.836s |  |
| Math Calculation | ✅ Pass | 3.798s |  |
| Basic Echo Function | ✅ Pass | 2.191s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.538s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.337s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.491s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.857s |  |
| Search Query Function | ✅ Pass | 1.869s |  |
| Ask Advice Function | ✅ Pass | 1.275s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.561s |  |
| Basic Context Memory Test | ✅ Pass | 5.012s |  |
| Function Argument Memory Test | ✅ Pass | 3.120s |  |
| Function Response Memory Test | ✅ Pass | 1.293s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.664s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.496s |  |
| Penetration Testing Methodology | ✅ Pass | 22.967s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.455s |  |
| SQL Injection Attack Type | ✅ Pass | 5.215s |  |
| Penetration Testing Framework | ✅ Pass | 5.249s |  |
| Web Application Security Scanner | ✅ Pass | 5.172s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.180s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.488s

---

### assistant (glm-5-turbo)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.980s |  |
| Text Transform Uppercase | ✅ Pass | 2.896s |  |
| Count from 1 to 5 | ✅ Pass | 2.572s |  |
| Math Calculation | ✅ Pass | 3.321s |  |
| Basic Echo Function | ✅ Pass | 1.431s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.935s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.138s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.242s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.378s |  |
| Search Query Function | ✅ Pass | 1.954s |  |
| Ask Advice Function | ✅ Pass | 1.638s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.390s |  |
| Basic Context Memory Test | ✅ Pass | 4.105s |  |
| Function Argument Memory Test | ✅ Pass | 1.651s |  |
| Function Response Memory Test | ✅ Pass | 2.104s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.155s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.861s |  |
| Penetration Testing Methodology | ✅ Pass | 5.751s |  |
| Vulnerability Assessment Tools | ❌ Fail | 6.224s | expected text 'network' not found |
| SQL Injection Attack Type | ✅ Pass | 4.402s |  |
| Penetration Testing Framework | ✅ Pass | 6.187s |  |
| Web Application Security Scanner | ✅ Pass | 5.577s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.904s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 3.383s

---

### generator (glm-5.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.598s |  |
| Text Transform Uppercase | ✅ Pass | 4.686s |  |
| Count from 1 to 5 | ✅ Pass | 4.534s |  |
| Math Calculation | ✅ Pass | 4.466s |  |
| Basic Echo Function | ✅ Pass | 5.033s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.667s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.827s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.858s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.855s |  |
| Search Query Function | ✅ Pass | 5.443s |  |
| Ask Advice Function | ✅ Pass | 3.380s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.654s |  |
| Basic Context Memory Test | ✅ Pass | 7.745s |  |
| Function Argument Memory Test | ✅ Pass | 9.781s |  |
| Function Response Memory Test | ✅ Pass | 6.321s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.979s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.237s |  |
| Penetration Testing Methodology | ✅ Pass | 13.391s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.520s |  |
| SQL Injection Attack Type | ✅ Pass | 14.260s |  |
| Penetration Testing Framework | ✅ Pass | 7.151s |  |
| Web Application Security Scanner | ✅ Pass | 12.338s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.118s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.081s

---

### refiner (glm-5.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.346s |  |
| Text Transform Uppercase | ✅ Pass | 7.388s |  |
| Count from 1 to 5 | ✅ Pass | 7.160s |  |
| Math Calculation | ✅ Pass | 4.476s |  |
| Basic Echo Function | ✅ Pass | 7.189s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.466s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.994s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.243s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 6.498s |  |
| Search Query Function | ✅ Pass | 5.493s |  |
| Ask Advice Function | ✅ Pass | 4.021s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.686s |  |
| Basic Context Memory Test | ✅ Pass | 5.153s |  |
| Function Argument Memory Test | ✅ Pass | 5.468s |  |
| Function Response Memory Test | ✅ Pass | 3.245s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 7.345s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.404s |  |
| Penetration Testing Methodology | ✅ Pass | 12.575s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.644s |  |
| SQL Injection Attack Type | ✅ Pass | 5.347s |  |
| Penetration Testing Framework | ✅ Pass | 11.479s |  |
| Web Application Security Scanner | ✅ Pass | 12.258s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.992s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 6.864s

---

### adviser (glm-5.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.436s |  |
| Text Transform Uppercase | ✅ Pass | 4.739s |  |
| Count from 1 to 5 | ✅ Pass | 7.323s |  |
| Math Calculation | ✅ Pass | 3.884s |  |
| Basic Echo Function | ✅ Pass | 4.145s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.102s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.763s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.518s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.657s |  |
| Search Query Function | ✅ Pass | 3.205s |  |
| Ask Advice Function | ✅ Pass | 3.559s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.959s |  |
| Basic Context Memory Test | ✅ Pass | 6.817s |  |
| Function Argument Memory Test | ✅ Pass | 6.147s |  |
| Function Response Memory Test | ✅ Pass | 3.902s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.443s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.905s |  |
| Penetration Testing Methodology | ✅ Pass | 14.182s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.240s |  |
| SQL Injection Attack Type | ✅ Pass | 5.849s |  |
| Penetration Testing Framework | ✅ Pass | 11.161s |  |
| Web Application Security Scanner | ✅ Pass | 11.757s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.019s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.553s

---

### reflector (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.349s |  |
| Text Transform Uppercase | ✅ Pass | 0.952s |  |
| Count from 1 to 5 | ✅ Pass | 3.336s |  |
| Math Calculation | ✅ Pass | 0.828s |  |
| Basic Echo Function | ✅ Pass | 1.318s |  |
| Streaming Simple Math Streaming | ✅ Pass | 8.738s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.624s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.163s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.864s |  |
| Search Query Function | ✅ Pass | 1.759s |  |
| Ask Advice Function | ✅ Pass | 1.985s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.501s |  |
| Basic Context Memory Test | ✅ Pass | 1.776s |  |
| Function Argument Memory Test | ✅ Pass | 1.869s |  |
| Function Response Memory Test | ✅ Pass | 1.430s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.434s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.338s |  |
| Penetration Testing Methodology | ✅ Pass | 6.325s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.048s |  |
| SQL Injection Attack Type | ✅ Pass | 2.589s |  |
| Penetration Testing Framework | ✅ Pass | 7.158s |  |
| Web Application Security Scanner | ✅ Pass | 5.490s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.597s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.760s

---

### searcher (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.144s |  |
| Text Transform Uppercase | ✅ Pass | 1.230s |  |
| Count from 1 to 5 | ✅ Pass | 2.526s |  |
| Math Calculation | ✅ Pass | 0.975s |  |
| Basic Echo Function | ✅ Pass | 1.962s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.438s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.872s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.207s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.837s |  |
| Search Query Function | ✅ Pass | 2.366s |  |
| Ask Advice Function | ✅ Pass | 1.564s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.558s |  |
| Basic Context Memory Test | ✅ Pass | 1.415s |  |
| Function Argument Memory Test | ✅ Pass | 1.913s |  |
| Function Response Memory Test | ✅ Pass | 1.785s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.208s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.507s |  |
| Penetration Testing Methodology | ✅ Pass | 5.352s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.510s |  |
| SQL Injection Attack Type | ✅ Pass | 2.434s |  |
| Penetration Testing Framework | ✅ Pass | 5.262s |  |
| Web Application Security Scanner | ✅ Pass | 2.488s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.891s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.889s

---

### enricher (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.166s |  |
| Text Transform Uppercase | ✅ Pass | 1.054s |  |
| Count from 1 to 5 | ✅ Pass | 1.600s |  |
| Math Calculation | ✅ Pass | 1.054s |  |
| Basic Echo Function | ✅ Pass | 1.737s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.306s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.486s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.403s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.452s |  |
| Search Query Function | ✅ Pass | 1.895s |  |
| Ask Advice Function | ✅ Pass | 1.631s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.771s |  |
| Basic Context Memory Test | ✅ Pass | 1.640s |  |
| Function Argument Memory Test | ✅ Pass | 2.034s |  |
| Function Response Memory Test | ✅ Pass | 1.623s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.927s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.506s |  |
| Penetration Testing Methodology | ✅ Pass | 8.173s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.181s |  |
| SQL Injection Attack Type | ✅ Pass | 1.488s |  |
| Penetration Testing Framework | ✅ Pass | 14.886s |  |
| Web Application Security Scanner | ✅ Pass | 2.242s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.759s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.001s

---

### coder (glm-5.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.213s |  |
| Text Transform Uppercase | ✅ Pass | 4.692s |  |
| Count from 1 to 5 | ✅ Pass | 5.795s |  |
| Math Calculation | ✅ Pass | 9.208s |  |
| Basic Echo Function | ✅ Pass | 5.473s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.389s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 8.238s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.129s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.931s |  |
| Search Query Function | ✅ Pass | 3.264s |  |
| Ask Advice Function | ✅ Pass | 7.299s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.841s |  |
| Basic Context Memory Test | ✅ Pass | 5.589s |  |
| Function Argument Memory Test | ✅ Pass | 6.503s |  |
| Function Response Memory Test | ✅ Pass | 3.301s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 6.407s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.279s |  |
| Penetration Testing Methodology | ✅ Pass | 17.570s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.116s |  |
| SQL Injection Attack Type | ✅ Pass | 5.596s |  |
| Penetration Testing Framework | ✅ Pass | 16.889s |  |
| Web Application Security Scanner | ✅ Pass | 11.733s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.497s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 7.172s

---

### installer (glm-4.5-air)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.253s |  |
| Text Transform Uppercase | ✅ Pass | 2.297s |  |
| Count from 1 to 5 | ✅ Pass | 3.713s |  |
| Math Calculation | ✅ Pass | 2.951s |  |
| Basic Echo Function | ✅ Pass | 2.057s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.884s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.290s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.606s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.104s |  |
| Search Query Function | ✅ Pass | 3.696s |  |
| Ask Advice Function | ✅ Pass | 3.538s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.188s |  |
| Basic Context Memory Test | ✅ Pass | 3.317s |  |
| Function Argument Memory Test | ✅ Pass | 1.915s |  |
| Function Response Memory Test | ✅ Pass | 3.620s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.784s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.012s |  |
| Penetration Testing Methodology | ✅ Pass | 14.559s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.449s |  |
| SQL Injection Attack Type | ✅ Pass | 8.793s |  |
| Penetration Testing Framework | ✅ Pass | 22.365s |  |
| Web Application Security Scanner | ✅ Pass | 15.217s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.740s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.146s

---

### pentester (glm-5.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.261s |  |
| Text Transform Uppercase | ✅ Pass | 6.631s |  |
| Count from 1 to 5 | ✅ Pass | 4.222s |  |
| Math Calculation | ✅ Pass | 3.124s |  |
| Basic Echo Function | ✅ Pass | 3.221s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.244s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.420s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.163s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.696s |  |
| Search Query Function | ✅ Pass | 2.983s |  |
| Ask Advice Function | ✅ Pass | 5.813s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 9.633s |  |
| Basic Context Memory Test | ✅ Pass | 4.770s |  |
| Function Argument Memory Test | ✅ Pass | 7.549s |  |
| Function Response Memory Test | ✅ Pass | 3.946s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.723s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.120s |  |
| Penetration Testing Methodology | ✅ Pass | 17.474s |  |
| Vulnerability Assessment Tools | ✅ Pass | 21.649s |  |
| SQL Injection Attack Type | ✅ Pass | 6.494s |  |
| Penetration Testing Framework | ✅ Pass | 12.428s |  |
| Web Application Security Scanner | ✅ Pass | 8.468s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.718s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.772s

---

