# LLM Agent Testing Report

Generated: Tue, 30 Dec 2025 22:01:09 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 22/23 (95.65%) | 0.830s |
| simple_json | gemini-2.0-flash-lite | false | 5/5 (100.00%) | 0.817s |
| primary_agent | gemini-2.5-flash | true | 23/23 (100.00%) | 2.402s |
| assistant | gemini-2.5-flash | true | 23/23 (100.00%) | 2.424s |
| generator | gemini-2.5-pro | true | 23/23 (100.00%) | 5.460s |
| refiner | gemini-2.5-flash | true | 23/23 (100.00%) | 2.537s |
| adviser | gemini-2.5-flash | true | 23/23 (100.00%) | 2.705s |
| reflector | gemini-2.0-flash | false | 22/23 (95.65%) | 0.758s |
| searcher | gemini-2.0-flash | false | 22/23 (95.65%) | 0.730s |
| enricher | gemini-2.0-flash | false | 22/23 (95.65%) | 0.745s |
| coder | gemini-2.5-pro | true | 23/23 (100.00%) | 5.315s |
| installer | gemini-2.5-flash-lite | true | 23/23 (100.00%) | 2.495s |
| pentester | gemini-2.5-flash-lite | true | 23/23 (100.00%) | 2.259s |

**Total**: 277/281 (98.58%) successful tests
**Overall average latency**: 2.360s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.473s |  |
| Text Transform Uppercase | ✅ Pass | 0.567s |  |
| Count from 1 to 5 | ✅ Pass | 0.718s |  |
| Math Calculation | ✅ Pass | 0.695s |  |
| Basic Echo Function | ✅ Pass | 0.733s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.925s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.607s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.839s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.674s |  |
| Search Query Function | ✅ Pass | 0.706s |  |
| Ask Advice Function | ✅ Pass | 0.646s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.711s |  |
| Basic Context Memory Test | ✅ Pass | 0.613s |  |
| Function Argument Memory Test | ✅ Pass | 0.619s |  |
| Function Response Memory Test | ✅ Pass | 0.562s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.903s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.590s | expected text 'example\.com' not found |
| Penetration Testing Methodology | ✅ Pass | 1.345s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.097s |  |
| SQL Injection Attack Type | ✅ Pass | 0.589s |  |
| Penetration Testing Framework | ✅ Pass | 0.868s |  |
| Web Application Security Scanner | ✅ Pass | 0.902s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.694s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.830s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.034s |  |
| Person Information JSON | ✅ Pass | 0.743s |  |
| Project Information JSON | ✅ Pass | 0.741s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.600s |  |
| User Profile JSON | ✅ Pass | 0.965s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.817s

---

### primary_agent (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.858s |  |
| Text Transform Uppercase | ✅ Pass | 1.338s |  |
| Count from 1 to 5 | ✅ Pass | 0.470s |  |
| Math Calculation | ✅ Pass | 1.226s |  |
| Basic Echo Function | ✅ Pass | 1.886s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.579s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.421s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.340s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.799s |  |
| Search Query Function | ✅ Pass | 1.694s |  |
| Ask Advice Function | ✅ Pass | 2.014s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.306s |  |
| Basic Context Memory Test | ✅ Pass | 1.619s |  |
| Function Argument Memory Test | ✅ Pass | 1.700s |  |
| Function Response Memory Test | ✅ Pass | 1.597s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.701s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.022s |  |
| Penetration Testing Methodology | ✅ Pass | 3.501s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.559s |  |
| SQL Injection Attack Type | ✅ Pass | 2.175s |  |
| Penetration Testing Framework | ✅ Pass | 6.564s |  |
| Web Application Security Scanner | ✅ Pass | 6.140s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.723s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.402s

---

### assistant (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.588s |  |
| Text Transform Uppercase | ✅ Pass | 1.444s |  |
| Count from 1 to 5 | ✅ Pass | 0.433s |  |
| Math Calculation | ✅ Pass | 1.441s |  |
| Basic Echo Function | ✅ Pass | 1.676s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.437s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.325s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.112s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.321s |  |
| Search Query Function | ✅ Pass | 1.909s |  |
| Ask Advice Function | ✅ Pass | 1.903s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.604s |  |
| Basic Context Memory Test | ✅ Pass | 2.090s |  |
| Function Argument Memory Test | ✅ Pass | 1.324s |  |
| Function Response Memory Test | ✅ Pass | 2.190s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.301s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.382s |  |
| Penetration Testing Methodology | ✅ Pass | 2.703s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.559s |  |
| SQL Injection Attack Type | ✅ Pass | 4.031s |  |
| Penetration Testing Framework | ✅ Pass | 7.913s |  |
| Web Application Security Scanner | ✅ Pass | 5.580s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.483s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.424s

---

### generator (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.810s |  |
| Text Transform Uppercase | ✅ Pass | 4.003s |  |
| Count from 1 to 5 | ✅ Pass | 3.116s |  |
| Math Calculation | ✅ Pass | 3.520s |  |
| Basic Echo Function | ✅ Pass | 3.447s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.103s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.690s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.689s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.756s |  |
| Search Query Function | ✅ Pass | 3.015s |  |
| Ask Advice Function | ✅ Pass | 2.642s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.456s |  |
| Basic Context Memory Test | ✅ Pass | 5.219s |  |
| Function Argument Memory Test | ✅ Pass | 3.941s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.406s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.949s |  |
| Function Response Memory Test | ✅ Pass | 13.391s |  |
| Penetration Testing Methodology | ✅ Pass | 10.591s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.674s |  |
| SQL Injection Attack Type | ✅ Pass | 6.398s |  |
| Penetration Testing Framework | ✅ Pass | 12.610s |  |
| Web Application Security Scanner | ✅ Pass | 9.891s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.255s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.460s

---

### refiner (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.882s |  |
| Text Transform Uppercase | ✅ Pass | 1.992s |  |
| Count from 1 to 5 | ✅ Pass | 1.626s |  |
| Math Calculation | ✅ Pass | 2.425s |  |
| Basic Echo Function | ✅ Pass | 1.470s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.199s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.265s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.099s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.510s |  |
| Search Query Function | ✅ Pass | 1.304s |  |
| Ask Advice Function | ✅ Pass | 2.104s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.265s |  |
| Basic Context Memory Test | ✅ Pass | 1.727s |  |
| Function Argument Memory Test | ✅ Pass | 1.741s |  |
| Function Response Memory Test | ✅ Pass | 2.150s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.820s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.218s |  |
| Penetration Testing Methodology | ✅ Pass | 3.723s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.481s |  |
| SQL Injection Attack Type | ✅ Pass | 2.281s |  |
| Penetration Testing Framework | ✅ Pass | 6.389s |  |
| Web Application Security Scanner | ✅ Pass | 5.752s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.927s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.537s

---

### adviser (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.013s |  |
| Text Transform Uppercase | ✅ Pass | 1.911s |  |
| Count from 1 to 5 | ✅ Pass | 0.831s |  |
| Math Calculation | ✅ Pass | 1.519s |  |
| Basic Echo Function | ✅ Pass | 1.537s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.150s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.127s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.577s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.305s |  |
| Search Query Function | ✅ Pass | 1.717s |  |
| Ask Advice Function | ✅ Pass | 2.033s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.126s |  |
| Basic Context Memory Test | ✅ Pass | 2.152s |  |
| Function Argument Memory Test | ✅ Pass | 1.534s |  |
| Function Response Memory Test | ✅ Pass | 3.335s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.171s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.072s |  |
| Penetration Testing Methodology | ✅ Pass | 3.474s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.504s |  |
| SQL Injection Attack Type | ✅ Pass | 3.773s |  |
| Penetration Testing Framework | ✅ Pass | 7.461s |  |
| Web Application Security Scanner | ✅ Pass | 6.781s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.097s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.705s

---

### reflector (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.442s |  |
| Text Transform Uppercase | ✅ Pass | 0.609s |  |
| Count from 1 to 5 | ✅ Pass | 0.516s |  |
| Math Calculation | ✅ Pass | 0.585s |  |
| Basic Echo Function | ✅ Pass | 0.661s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.625s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.576s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.783s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.739s |  |
| Search Query Function | ✅ Pass | 0.580s |  |
| Ask Advice Function | ✅ Pass | 0.700s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.744s |  |
| Basic Context Memory Test | ✅ Pass | 0.630s |  |
| Function Argument Memory Test | ✅ Pass | 0.652s |  |
| Function Response Memory Test | ✅ Pass | 0.614s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.839s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.475s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.723s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.045s |  |
| SQL Injection Attack Type | ✅ Pass | 0.410s |  |
| Penetration Testing Framework | ✅ Pass | 0.897s |  |
| Web Application Security Scanner | ✅ Pass | 0.787s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.783s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.758s

---

### searcher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.361s |  |
| Text Transform Uppercase | ✅ Pass | 0.568s |  |
| Count from 1 to 5 | ✅ Pass | 0.480s |  |
| Math Calculation | ✅ Pass | 0.611s |  |
| Basic Echo Function | ✅ Pass | 0.635s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.480s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.449s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.589s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.714s |  |
| Search Query Function | ✅ Pass | 0.550s |  |
| Ask Advice Function | ✅ Pass | 0.699s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.728s |  |
| Basic Context Memory Test | ✅ Pass | 0.578s |  |
| Function Argument Memory Test | ✅ Pass | 0.552s |  |
| Function Response Memory Test | ✅ Pass | 0.642s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.919s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.501s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.454s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.465s |  |
| SQL Injection Attack Type | ✅ Pass | 0.438s |  |
| Penetration Testing Framework | ✅ Pass | 0.673s |  |
| Web Application Security Scanner | ✅ Pass | 0.791s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.897s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.730s

---

### enricher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.667s |  |
| Text Transform Uppercase | ✅ Pass | 0.608s |  |
| Count from 1 to 5 | ✅ Pass | 0.577s |  |
| Math Calculation | ✅ Pass | 0.509s |  |
| Basic Echo Function | ✅ Pass | 0.616s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.578s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.548s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.752s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.764s |  |
| Search Query Function | ✅ Pass | 0.593s |  |
| Ask Advice Function | ✅ Pass | 0.720s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.765s |  |
| Basic Context Memory Test | ✅ Pass | 0.713s |  |
| Function Argument Memory Test | ✅ Pass | 0.597s |  |
| Function Response Memory Test | ✅ Pass | 0.593s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.935s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.457s | no content in generation response |
| Penetration Testing Methodology | ✅ Pass | 1.996s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.597s |  |
| SQL Injection Attack Type | ✅ Pass | 0.409s |  |
| Penetration Testing Framework | ✅ Pass | 0.698s |  |
| Web Application Security Scanner | ✅ Pass | 0.621s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.814s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 0.745s

---

### coder (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.022s |  |
| Text Transform Uppercase | ✅ Pass | 3.721s |  |
| Count from 1 to 5 | ✅ Pass | 5.274s |  |
| Math Calculation | ✅ Pass | 3.078s |  |
| Basic Echo Function | ✅ Pass | 2.992s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.619s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.444s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.204s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.866s |  |
| Search Query Function | ✅ Pass | 4.069s |  |
| Ask Advice Function | ✅ Pass | 3.009s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.730s |  |
| Basic Context Memory Test | ✅ Pass | 4.033s |  |
| Function Argument Memory Test | ✅ Pass | 4.490s |  |
| Function Response Memory Test | ✅ Pass | 4.497s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.694s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.035s |  |
| Penetration Testing Methodology | ✅ Pass | 11.176s |  |
| SQL Injection Attack Type | ✅ Pass | 4.041s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.667s |  |
| Penetration Testing Framework | ✅ Pass | 14.035s |  |
| Web Application Security Scanner | ✅ Pass | 9.590s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.944s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.315s

---

### installer (gemini-2.5-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.496s |  |
| Text Transform Uppercase | ✅ Pass | 2.012s |  |
| Count from 1 to 5 | ✅ Pass | 1.244s |  |
| Math Calculation | ✅ Pass | 2.384s |  |
| Basic Echo Function | ✅ Pass | 1.564s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.437s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.273s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.410s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.555s |  |
| Search Query Function | ✅ Pass | 1.336s |  |
| Ask Advice Function | ✅ Pass | 1.894s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.499s |  |
| Basic Context Memory Test | ✅ Pass | 1.901s |  |
| Function Argument Memory Test | ✅ Pass | 1.550s |  |
| Function Response Memory Test | ✅ Pass | 1.589s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.005s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.454s |  |
| Penetration Testing Methodology | ✅ Pass | 3.756s |  |
| SQL Injection Attack Type | ✅ Pass | 2.893s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.199s |  |
| Penetration Testing Framework | ✅ Pass | 4.908s |  |
| Web Application Security Scanner | ✅ Pass | 6.320s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.685s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.495s

---

### pentester (gemini-2.5-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.326s |  |
| Text Transform Uppercase | ✅ Pass | 1.316s |  |
| Count from 1 to 5 | ✅ Pass | 1.542s |  |
| Math Calculation | ✅ Pass | 1.345s |  |
| Basic Echo Function | ✅ Pass | 1.849s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.292s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.918s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.247s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.643s |  |
| Search Query Function | ✅ Pass | 1.087s |  |
| Ask Advice Function | ✅ Pass | 1.740s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.994s |  |
| Basic Context Memory Test | ✅ Pass | 1.939s |  |
| Function Argument Memory Test | ✅ Pass | 1.355s |  |
| Function Response Memory Test | ✅ Pass | 1.795s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.612s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.358s |  |
| Penetration Testing Methodology | ✅ Pass | 4.831s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.141s |  |
| SQL Injection Attack Type | ✅ Pass | 1.687s |  |
| Penetration Testing Framework | ✅ Pass | 5.425s |  |
| Web Application Security Scanner | ✅ Pass | 4.511s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.990s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.259s

---

