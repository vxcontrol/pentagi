# LLM Agent Testing Report

Generated: Thu, 29 Jan 2026 17:36:55 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-haiku-4-5 | false | 23/23 (100.00%) | 1.239s |
| simple_json | claude-haiku-4-5 | false | 5/5 (100.00%) | 1.181s |
| primary_agent | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.542s |
| assistant | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.484s |
| generator | claude-opus-4-5 | true | 22/23 (95.65%) | 3.806s |
| refiner | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.512s |
| adviser | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.846s |
| reflector | claude-haiku-4-5 | true | 23/23 (100.00%) | 1.750s |
| searcher | claude-haiku-4-5 | true | 23/23 (100.00%) | 2.005s |
| enricher | claude-haiku-4-5 | true | 23/23 (100.00%) | 1.274s |
| coder | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.591s |
| installer | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.467s |
| pentester | claude-sonnet-4-5 | true | 23/23 (100.00%) | 3.390s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 2.878s

## Detailed Results

### simple (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.141s |  |
| Text Transform Uppercase | ✅ Pass | 0.758s |  |
| Count from 1 to 5 | ✅ Pass | 0.671s |  |
| Math Calculation | ✅ Pass | 0.670s |  |
| Basic Echo Function | ✅ Pass | 0.896s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.771s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.898s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.729s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.771s |  |
| Search Query Function | ✅ Pass | 1.083s |  |
| Ask Advice Function | ✅ Pass | 1.788s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.829s |  |
| Basic Context Memory Test | ✅ Pass | 0.800s |  |
| Function Argument Memory Test | ✅ Pass | 0.818s |  |
| Function Response Memory Test | ✅ Pass | 0.688s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.943s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.809s |  |
| Penetration Testing Methodology | ✅ Pass | 2.576s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.713s |  |
| SQL Injection Attack Type | ✅ Pass | 0.678s |  |
| Penetration Testing Framework | ✅ Pass | 2.781s |  |
| Web Application Security Scanner | ✅ Pass | 2.740s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.944s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.239s

---

### simple_json (claude-haiku-4-5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.094s |  |
| Project Information JSON | ✅ Pass | 0.980s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.344s |  |
| User Profile JSON | ✅ Pass | 1.302s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.181s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.181s

---

### primary_agent (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.570s |  |
| Text Transform Uppercase | ✅ Pass | 1.693s |  |
| Count from 1 to 5 | ✅ Pass | 2.039s |  |
| Math Calculation | ✅ Pass | 1.850s |  |
| Basic Echo Function | ✅ Pass | 2.055s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.790s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.265s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.289s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.807s |  |
| Search Query Function | ✅ Pass | 2.814s |  |
| Ask Advice Function | ✅ Pass | 2.926s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.151s |  |
| Basic Context Memory Test | ✅ Pass | 2.079s |  |
| Function Argument Memory Test | ✅ Pass | 2.591s |  |
| Function Response Memory Test | ✅ Pass | 2.866s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.508s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.358s |  |
| Penetration Testing Methodology | ✅ Pass | 9.430s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.419s |  |
| SQL Injection Attack Type | ✅ Pass | 3.307s |  |
| Penetration Testing Framework | ✅ Pass | 7.157s |  |
| Web Application Security Scanner | ✅ Pass | 5.690s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.796s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.542s

---

### assistant (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.461s |  |
| Text Transform Uppercase | ✅ Pass | 2.354s |  |
| Count from 1 to 5 | ✅ Pass | 2.622s |  |
| Math Calculation | ✅ Pass | 1.745s |  |
| Basic Echo Function | ✅ Pass | 2.333s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.756s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.387s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.253s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.586s |  |
| Search Query Function | ✅ Pass | 2.288s |  |
| Ask Advice Function | ✅ Pass | 2.792s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.011s |  |
| Basic Context Memory Test | ✅ Pass | 1.945s |  |
| Function Argument Memory Test | ✅ Pass | 2.272s |  |
| Function Response Memory Test | ✅ Pass | 2.866s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.404s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.110s |  |
| Penetration Testing Methodology | ✅ Pass | 9.989s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.019s |  |
| SQL Injection Attack Type | ✅ Pass | 3.210s |  |
| Penetration Testing Framework | ✅ Pass | 6.954s |  |
| Web Application Security Scanner | ✅ Pass | 6.797s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.967s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.484s

---

### generator (claude-opus-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.321s |  |
| Text Transform Uppercase | ✅ Pass | 2.552s |  |
| Count from 1 to 5 | ✅ Pass | 2.619s |  |
| Math Calculation | ✅ Pass | 1.929s |  |
| Basic Echo Function | ✅ Pass | 2.562s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.360s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.331s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.475s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.969s |  |
| Search Query Function | ✅ Pass | 2.670s |  |
| Ask Advice Function | ✅ Pass | 3.021s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.796s |  |
| Basic Context Memory Test | ✅ Pass | 3.351s |  |
| Function Argument Memory Test | ✅ Pass | 2.976s |  |
| Function Response Memory Test | ✅ Pass | 2.796s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 3.838s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.065s |  |
| Penetration Testing Methodology | ✅ Pass | 10.315s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.043s |  |
| SQL Injection Attack Type | ✅ Pass | 3.519s |  |
| Penetration Testing Framework | ✅ Pass | 7.758s |  |
| Web Application Security Scanner | ✅ Pass | 7.778s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.486s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 3.806s

---

### refiner (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.404s |  |
| Text Transform Uppercase | ✅ Pass | 2.447s |  |
| Count from 1 to 5 | ✅ Pass | 2.261s |  |
| Math Calculation | ✅ Pass | 1.882s |  |
| Basic Echo Function | ✅ Pass | 2.163s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.718s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.771s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.208s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.094s |  |
| Search Query Function | ✅ Pass | 2.448s |  |
| Ask Advice Function | ✅ Pass | 2.774s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.988s |  |
| Basic Context Memory Test | ✅ Pass | 1.760s |  |
| Function Argument Memory Test | ✅ Pass | 2.900s |  |
| Function Response Memory Test | ✅ Pass | 2.596s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.326s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.033s |  |
| Penetration Testing Methodology | ✅ Pass | 7.834s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.705s |  |
| SQL Injection Attack Type | ✅ Pass | 2.849s |  |
| Penetration Testing Framework | ✅ Pass | 6.204s |  |
| Web Application Security Scanner | ✅ Pass | 8.261s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.140s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.512s

---

### adviser (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.245s |  |
| Text Transform Uppercase | ✅ Pass | 2.271s |  |
| Count from 1 to 5 | ✅ Pass | 2.626s |  |
| Math Calculation | ✅ Pass | 2.379s |  |
| Basic Echo Function | ✅ Pass | 2.195s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.984s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.513s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.233s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.467s |  |
| Search Query Function | ✅ Pass | 1.824s |  |
| Ask Advice Function | ✅ Pass | 2.907s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.212s |  |
| Basic Context Memory Test | ✅ Pass | 2.029s |  |
| Function Argument Memory Test | ✅ Pass | 3.062s |  |
| Function Response Memory Test | ✅ Pass | 2.245s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.578s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.265s |  |
| Penetration Testing Methodology | ✅ Pass | 10.992s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.983s |  |
| SQL Injection Attack Type | ✅ Pass | 3.834s |  |
| Penetration Testing Framework | ✅ Pass | 11.421s |  |
| Web Application Security Scanner | ✅ Pass | 5.404s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.776s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.846s

---

### reflector (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.054s |  |
| Text Transform Uppercase | ✅ Pass | 1.121s |  |
| Count from 1 to 5 | ✅ Pass | 1.197s |  |
| Math Calculation | ✅ Pass | 0.758s |  |
| Basic Echo Function | ✅ Pass | 0.950s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.839s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.877s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.067s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.055s |  |
| Search Query Function | ✅ Pass | 1.245s |  |
| Ask Advice Function | ✅ Pass | 1.219s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.137s |  |
| Basic Context Memory Test | ✅ Pass | 1.086s |  |
| Function Argument Memory Test | ✅ Pass | 1.374s |  |
| Function Response Memory Test | ✅ Pass | 1.842s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.876s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.243s |  |
| Penetration Testing Methodology | ✅ Pass | 5.903s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.234s |  |
| SQL Injection Attack Type | ✅ Pass | 1.474s |  |
| Penetration Testing Framework | ✅ Pass | 4.361s |  |
| Web Application Security Scanner | ✅ Pass | 2.855s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.474s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.750s

---

### searcher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.103s |  |
| Text Transform Uppercase | ✅ Pass | 0.949s |  |
| Count from 1 to 5 | ✅ Pass | 1.459s |  |
| Math Calculation | ✅ Pass | 0.803s |  |
| Basic Echo Function | ✅ Pass | 1.227s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.397s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.554s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.528s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.932s |  |
| Search Query Function | ✅ Pass | 1.231s |  |
| Ask Advice Function | ✅ Pass | 1.183s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.042s |  |
| Basic Context Memory Test | ✅ Pass | 1.258s |  |
| Function Argument Memory Test | ✅ Pass | 1.074s |  |
| Function Response Memory Test | ✅ Pass | 1.228s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.921s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.275s |  |
| Penetration Testing Methodology | ✅ Pass | 7.863s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.928s |  |
| SQL Injection Attack Type | ✅ Pass | 3.186s |  |
| Penetration Testing Framework | ✅ Pass | 3.992s |  |
| Web Application Security Scanner | ✅ Pass | 2.703s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.267s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.005s

---

### enricher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.833s |  |
| Text Transform Uppercase | ✅ Pass | 0.743s |  |
| Count from 1 to 5 | ✅ Pass | 0.913s |  |
| Math Calculation | ✅ Pass | 0.633s |  |
| Basic Echo Function | ✅ Pass | 1.190s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.224s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.673s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.493s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.793s |  |
| Search Query Function | ✅ Pass | 0.886s |  |
| Ask Advice Function | ✅ Pass | 0.980s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.917s |  |
| Basic Context Memory Test | ✅ Pass | 0.795s |  |
| Function Argument Memory Test | ✅ Pass | 0.700s |  |
| Function Response Memory Test | ✅ Pass | 0.803s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.235s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.656s |  |
| Penetration Testing Methodology | ✅ Pass | 2.759s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.843s |  |
| SQL Injection Attack Type | ✅ Pass | 0.720s |  |
| Penetration Testing Framework | ✅ Pass | 4.301s |  |
| Web Application Security Scanner | ✅ Pass | 1.662s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.550s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.274s

---

### coder (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.400s |  |
| Text Transform Uppercase | ✅ Pass | 2.366s |  |
| Count from 1 to 5 | ✅ Pass | 2.549s |  |
| Math Calculation | ✅ Pass | 1.765s |  |
| Basic Echo Function | ✅ Pass | 2.240s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.441s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.591s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.306s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.554s |  |
| Search Query Function | ✅ Pass | 2.253s |  |
| Ask Advice Function | ✅ Pass | 2.974s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.105s |  |
| Basic Context Memory Test | ✅ Pass | 2.040s |  |
| Function Argument Memory Test | ✅ Pass | 2.789s |  |
| Function Response Memory Test | ✅ Pass | 2.121s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.801s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.181s |  |
| Penetration Testing Methodology | ✅ Pass | 10.242s |  |
| SQL Injection Attack Type | ✅ Pass | 3.511s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.006s |  |
| Penetration Testing Framework | ✅ Pass | 6.971s |  |
| Web Application Security Scanner | ✅ Pass | 7.101s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.264s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.591s

---

### installer (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.929s |  |
| Text Transform Uppercase | ✅ Pass | 2.695s |  |
| Count from 1 to 5 | ✅ Pass | 2.422s |  |
| Math Calculation | ✅ Pass | 1.983s |  |
| Basic Echo Function | ✅ Pass | 2.220s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.774s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.213s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.149s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.649s |  |
| Search Query Function | ✅ Pass | 2.268s |  |
| Ask Advice Function | ✅ Pass | 2.830s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.583s |  |
| Basic Context Memory Test | ✅ Pass | 2.037s |  |
| Function Argument Memory Test | ✅ Pass | 2.461s |  |
| Function Response Memory Test | ✅ Pass | 2.316s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.875s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.429s |  |
| Penetration Testing Methodology | ✅ Pass | 11.365s |  |
| SQL Injection Attack Type | ✅ Pass | 2.569s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.120s |  |
| Penetration Testing Framework | ✅ Pass | 6.013s |  |
| Web Application Security Scanner | ✅ Pass | 7.011s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.827s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.467s

---

### pentester (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.950s |  |
| Text Transform Uppercase | ✅ Pass | 2.156s |  |
| Count from 1 to 5 | ✅ Pass | 2.490s |  |
| Math Calculation | ✅ Pass | 1.791s |  |
| Basic Echo Function | ✅ Pass | 2.504s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.898s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.821s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.111s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.741s |  |
| Search Query Function | ✅ Pass | 2.284s |  |
| Ask Advice Function | ✅ Pass | 2.957s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.078s |  |
| Basic Context Memory Test | ✅ Pass | 2.196s |  |
| Function Argument Memory Test | ✅ Pass | 2.287s |  |
| Function Response Memory Test | ✅ Pass | 2.888s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.658s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.876s |  |
| Penetration Testing Methodology | ✅ Pass | 8.949s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.236s |  |
| SQL Injection Attack Type | ✅ Pass | 3.461s |  |
| Penetration Testing Framework | ✅ Pass | 6.547s |  |
| Web Application Security Scanner | ✅ Pass | 5.307s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.761s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.390s

---

