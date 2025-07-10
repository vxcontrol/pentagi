# LLM Agent Testing Report

Generated: Tue, 08 Jul 2025 21:30:07 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | openai/gpt-4.1-mini | false | 18/18 (100.00%) | 1.308s |
| simple_json | openai/gpt-4.1-mini | false | 4/4 (100.00%) | 1.397s |
| primary_agent | openai/o4-mini | true | 18/18 (100.00%) | 7.394s |
| assistant | x-ai/grok-3-beta | false | 18/18 (100.00%) | 1.682s |
| generator | anthropic/claude-3.7-sonnet:thinking | true | 18/18 (100.00%) | 3.977s |
| refiner | google/gemini-2.5-flash-preview:thinking | true | 18/18 (100.00%) | 3.116s |
| adviser | google/gemini-2.5-pro-preview | true | 18/18 (100.00%) | 7.895s |
| reflector | openai/gpt-4.1-mini | true | 18/18 (100.00%) | 1.309s |
| searcher | x-ai/grok-3-mini-beta | true | 18/18 (100.00%) | 3.216s |
| enricher | openai/gpt-4.1-mini | true | 18/18 (100.00%) | 1.124s |
| coder | anthropic/claude-3.7-sonnet:thinking | true | 18/18 (100.00%) | 3.950s |
| installer | google/gemini-2.5-flash-preview:thinking | true | 18/18 (100.00%) | 3.535s |
| pentester | x-ai/grok-3-mini-beta | true | 18/18 (100.00%) | 3.142s |

**Total**: 220/220 (100.00%) successful tests
**Overall average latency**: 3.433s

## Detailed Results

### simple (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.268s |  |
| Text Transform Uppercase | ✅ Pass | 0.751s |  |
| Count from 1 to 5 | ✅ Pass | 1.202s |  |
| Math Calculation | ✅ Pass | 1.066s |  |
| Basic Echo Function | ✅ Pass | 1.343s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.607s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.304s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.933s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.631s |  |
| Search Query Function | ✅ Pass | 0.826s |  |
| Ask Advice Function | ✅ Pass | 1.195s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.274s |  |
| Penetration Testing Methodology | ✅ Pass | 1.131s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.374s |  |
| SQL Injection Attack Type | ✅ Pass | 1.701s |  |
| Penetration Testing Framework | ✅ Pass | 1.804s |  |
| Web Application Security Scanner | ✅ Pass | 1.130s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.992s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.308s

---

### simple_json (openai/gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.738s |  |
| User Profile JSON | ✅ Pass | 1.001s |  |
| Project Information JSON | ✅ Pass | 1.658s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.190s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 1.397s

---

### primary_agent (openai/o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.503s |  |
| Text Transform Uppercase | ✅ Pass | 7.097s |  |
| Count from 1 to 5 | ✅ Pass | 5.256s |  |
| Math Calculation | ✅ Pass | 5.193s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.765s |  |
| Basic Echo Function | ✅ Pass | 19.006s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 7.875s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.954s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 20.401s |  |
| Search Query Function | ✅ Pass | 12.445s |  |
| Ask Advice Function | ✅ Pass | 3.023s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.096s |  |
| Penetration Testing Methodology | ✅ Pass | 6.596s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.239s |  |
| SQL Injection Attack Type | ✅ Pass | 5.250s |  |
| Penetration Testing Framework | ✅ Pass | 8.384s |  |
| Web Application Security Scanner | ✅ Pass | 5.048s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.948s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 7.394s

---

### assistant (x-ai/grok-3-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.030s |  |
| Text Transform Uppercase | ✅ Pass | 0.794s |  |
| Count from 1 to 5 | ✅ Pass | 0.660s |  |
| Math Calculation | ✅ Pass | 0.761s |  |
| Basic Echo Function | ✅ Pass | 1.079s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.613s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.705s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.226s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.944s |  |
| Search Query Function | ✅ Pass | 1.271s |  |
| Ask Advice Function | ✅ Pass | 1.858s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.213s |  |
| Penetration Testing Methodology | ✅ Pass | 5.893s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.545s |  |
| SQL Injection Attack Type | ✅ Pass | 0.886s |  |
| Penetration Testing Framework | ✅ Pass | 1.712s |  |
| Web Application Security Scanner | ✅ Pass | 1.512s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.571s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.682s

---

### generator (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.271s |  |
| Text Transform Uppercase | ✅ Pass | 2.453s |  |
| Count from 1 to 5 | ✅ Pass | 3.420s |  |
| Math Calculation | ✅ Pass | 2.146s |  |
| Basic Echo Function | ✅ Pass | 3.565s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.516s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.694s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.940s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.952s |  |
| Search Query Function | ✅ Pass | 4.584s |  |
| Ask Advice Function | ✅ Pass | 3.105s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.309s |  |
| Penetration Testing Methodology | ✅ Pass | 4.811s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.859s |  |
| SQL Injection Attack Type | ✅ Pass | 6.968s |  |
| Penetration Testing Framework | ✅ Pass | 6.607s |  |
| Web Application Security Scanner | ✅ Pass | 3.498s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.887s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.977s

---

### refiner (google/gemini-2.5-flash-preview:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.550s |  |
| Text Transform Uppercase | ✅ Pass | 1.725s |  |
| Count from 1 to 5 | ✅ Pass | 2.639s |  |
| Math Calculation | ✅ Pass | 1.407s |  |
| Basic Echo Function | ✅ Pass | 2.411s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.448s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.375s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.669s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.770s |  |
| Search Query Function | ✅ Pass | 2.422s |  |
| Ask Advice Function | ✅ Pass | 2.587s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.939s |  |
| Penetration Testing Methodology | ✅ Pass | 2.805s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.344s |  |
| SQL Injection Attack Type | ✅ Pass | 2.498s |  |
| Penetration Testing Framework | ✅ Pass | 6.831s |  |
| Web Application Security Scanner | ✅ Pass | 6.315s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.336s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.116s

---

### adviser (google/gemini-2.5-pro-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.911s |  |
| Text Transform Uppercase | ✅ Pass | 4.328s |  |
| Count from 1 to 5 | ✅ Pass | 4.855s |  |
| Math Calculation | ✅ Pass | 3.843s |  |
| Basic Echo Function | ✅ Pass | 3.895s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.392s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.427s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.499s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.349s |  |
| Search Query Function | ✅ Pass | 3.029s |  |
| Ask Advice Function | ✅ Pass | 2.967s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.796s |  |
| Penetration Testing Methodology | ✅ Pass | 8.308s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.364s |  |
| SQL Injection Attack Type | ✅ Pass | 3.483s |  |
| Penetration Testing Framework | ✅ Pass | 10.389s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.935s |  |
| Web Application Security Scanner | ✅ Pass | 4.935s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 7.895s

---

### reflector (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.963s |  |
| Text Transform Uppercase | ✅ Pass | 1.036s |  |
| Count from 1 to 5 | ✅ Pass | 1.295s |  |
| Math Calculation | ✅ Pass | 1.102s |  |
| Basic Echo Function | ✅ Pass | 0.986s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.205s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.646s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.570s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.977s |  |
| Search Query Function | ✅ Pass | 1.370s |  |
| Ask Advice Function | ✅ Pass | 1.845s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.019s |  |
| Penetration Testing Methodology | ✅ Pass | 1.078s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.446s |  |
| SQL Injection Attack Type | ✅ Pass | 0.639s |  |
| Penetration Testing Framework | ✅ Pass | 2.082s |  |
| Web Application Security Scanner | ✅ Pass | 2.051s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.236s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.309s

---

### searcher (x-ai/grok-3-mini-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.946s |  |
| Text Transform Uppercase | ✅ Pass | 1.747s |  |
| Count from 1 to 5 | ✅ Pass | 2.394s |  |
| Math Calculation | ✅ Pass | 1.388s |  |
| Basic Echo Function | ✅ Pass | 3.406s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.495s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.782s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.128s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.492s |  |
| Search Query Function | ✅ Pass | 4.034s |  |
| Ask Advice Function | ✅ Pass | 3.554s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.186s |  |
| Penetration Testing Methodology | ✅ Pass | 3.636s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.857s |  |
| SQL Injection Attack Type | ✅ Pass | 2.229s |  |
| Penetration Testing Framework | ✅ Pass | 4.880s |  |
| Web Application Security Scanner | ✅ Pass | 4.304s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.425s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.216s

---

### enricher (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.644s |  |
| Text Transform Uppercase | ✅ Pass | 0.625s |  |
| Count from 1 to 5 | ✅ Pass | 1.404s |  |
| Math Calculation | ✅ Pass | 0.681s |  |
| Basic Echo Function | ✅ Pass | 0.925s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.529s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.819s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.867s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.201s |  |
| Search Query Function | ✅ Pass | 0.879s |  |
| Ask Advice Function | ✅ Pass | 1.239s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.291s |  |
| Penetration Testing Methodology | ✅ Pass | 0.900s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.166s |  |
| SQL Injection Attack Type | ✅ Pass | 1.281s |  |
| Penetration Testing Framework | ✅ Pass | 1.391s |  |
| Web Application Security Scanner | ✅ Pass | 0.925s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.461s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.124s

---

### coder (anthropic/claude-3.7-sonnet:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.828s |  |
| Text Transform Uppercase | ✅ Pass | 3.137s |  |
| Count from 1 to 5 | ✅ Pass | 5.096s |  |
| Math Calculation | ✅ Pass | 2.659s |  |
| Basic Echo Function | ✅ Pass | 3.720s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.413s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.996s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.137s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.010s |  |
| Search Query Function | ✅ Pass | 3.594s |  |
| Ask Advice Function | ✅ Pass | 2.944s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.587s |  |
| Penetration Testing Methodology | ✅ Pass | 6.128s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.407s |  |
| SQL Injection Attack Type | ✅ Pass | 4.214s |  |
| Penetration Testing Framework | ✅ Pass | 5.801s |  |
| Web Application Security Scanner | ✅ Pass | 5.816s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.600s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.950s

---

### installer (google/gemini-2.5-flash-preview:thinking)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.560s |  |
| Text Transform Uppercase | ✅ Pass | 2.386s |  |
| Count from 1 to 5 | ✅ Pass | 2.466s |  |
| Math Calculation | ✅ Pass | 2.509s |  |
| Basic Echo Function | ✅ Pass | 2.665s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.554s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.734s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.640s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.092s |  |
| Search Query Function | ✅ Pass | 2.592s |  |
| Ask Advice Function | ✅ Pass | 2.777s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.839s |  |
| Penetration Testing Methodology | ✅ Pass | 4.037s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.558s |  |
| SQL Injection Attack Type | ✅ Pass | 5.398s |  |
| Penetration Testing Framework | ✅ Pass | 7.805s |  |
| Web Application Security Scanner | ✅ Pass | 6.249s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.768s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.535s

---

### pentester (x-ai/grok-3-mini-beta)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.636s |  |
| Text Transform Uppercase | ✅ Pass | 2.915s |  |
| Count from 1 to 5 | ✅ Pass | 3.184s |  |
| Math Calculation | ✅ Pass | 1.540s |  |
| Basic Echo Function | ✅ Pass | 3.418s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.391s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.034s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.223s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.020s |  |
| Search Query Function | ✅ Pass | 3.191s |  |
| Ask Advice Function | ✅ Pass | 3.599s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.569s |  |
| Penetration Testing Methodology | ✅ Pass | 2.752s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.419s |  |
| SQL Injection Attack Type | ✅ Pass | 1.873s |  |
| Penetration Testing Framework | ✅ Pass | 5.461s |  |
| Web Application Security Scanner | ✅ Pass | 3.248s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.077s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.142s

---

