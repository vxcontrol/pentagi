# LLM Agent Testing Report

Generated: Sun, 20 Jul 2025 10:40:23 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | openai/gpt-4.1-mini | false | 23/23 (100.00%) | 0.910s |
| simple_json | openai/gpt-4.1-mini | false | 5/5 (100.00%) | 1.248s |
| primary_agent | openai/o4-mini | true | 23/23 (100.00%) | 4.344s |
| assistant | x-ai/grok-4 | true | 23/23 (100.00%) | 8.688s |
| generator | anthropic/claude-sonnet-4 | true | 23/23 (100.00%) | 3.963s |
| refiner | google/gemini-2.5-pro | true | 22/23 (95.65%) | 5.166s |
| adviser | google/gemini-2.5-pro | true | 22/23 (95.65%) | 4.937s |
| reflector | openai/gpt-4.1-mini | false | 23/23 (100.00%) | 0.844s |
| searcher | x-ai/grok-3-mini | true | 23/23 (100.00%) | 2.976s |
| enricher | openai/gpt-4.1-mini | true | 23/23 (100.00%) | 0.878s |
| coder | anthropic/claude-sonnet-4 | true | 23/23 (100.00%) | 3.897s |
| installer | google/gemini-2.5-flash | true | 21/23 (91.30%) | 2.699s |
| pentester | moonshotai/kimi-k2 | true | 22/23 (95.65%) | 1.320s |

**Total**: 276/281 (98.22%) successful tests
**Overall average latency**: 3.347s

## Detailed Results

### simple (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.989s |  |
| Text Transform Uppercase | ✅ Pass | 2.282s |  |
| Count from 1 to 5 | ✅ Pass | 0.762s |  |
| Math Calculation | ✅ Pass | 0.494s |  |
| Basic Echo Function | ✅ Pass | 0.808s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.484s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.617s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.832s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.814s |  |
| Search Query Function | ✅ Pass | 0.801s |  |
| Ask Advice Function | ✅ Pass | 0.909s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.599s |  |
| Basic Context Memory Test | ✅ Pass | 0.640s |  |
| Function Argument Memory Test | ✅ Pass | 0.494s |  |
| Function Response Memory Test | ✅ Pass | 0.365s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.151s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.465s |  |
| Penetration Testing Methodology | ✅ Pass | 0.748s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.672s |  |
| SQL Injection Attack Type | ✅ Pass | 0.568s |  |
| Penetration Testing Framework | ✅ Pass | 0.771s |  |
| Web Application Security Scanner | ✅ Pass | 1.831s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.832s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.910s

---

### simple_json (openai/gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 1.045s |  |
| User Profile JSON | ✅ Pass | 0.936s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.556s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.887s |  |
| Person Information JSON | ✅ Pass | 1.814s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.248s

---

### primary_agent (openai/o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.309s |  |
| Text Transform Uppercase | ✅ Pass | 4.023s |  |
| Count from 1 to 5 | ✅ Pass | 2.957s |  |
| Math Calculation | ✅ Pass | 2.364s |  |
| Basic Echo Function | ✅ Pass | 2.856s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.073s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.228s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.529s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.986s |  |
| Search Query Function | ✅ Pass | 4.343s |  |
| Ask Advice Function | ✅ Pass | 3.154s |  |
| Basic Context Memory Test | ✅ Pass | 4.983s |  |
| Function Argument Memory Test | ✅ Pass | 2.654s |  |
| Function Response Memory Test | ✅ Pass | 2.528s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 22.047s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.846s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.254s |  |
| Penetration Testing Methodology | ✅ Pass | 3.357s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.877s |  |
| SQL Injection Attack Type | ✅ Pass | 2.604s |  |
| Penetration Testing Framework | ✅ Pass | 4.444s |  |
| Web Application Security Scanner | ✅ Pass | 3.677s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.816s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.344s

---

### assistant (x-ai/grok-4)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.528s |  |
| Text Transform Uppercase | ✅ Pass | 3.418s |  |
| Count from 1 to 5 | ✅ Pass | 3.647s |  |
| Math Calculation | ✅ Pass | 4.332s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.827s |  |
| Basic Echo Function | ✅ Pass | 5.496s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.961s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 6.386s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.550s |  |
| Search Query Function | ✅ Pass | 5.066s |  |
| Ask Advice Function | ✅ Pass | 5.670s |  |
| Basic Context Memory Test | ✅ Pass | 3.228s |  |
| Function Response Memory Test | ✅ Pass | 9.694s |  |
| Function Argument Memory Test | ✅ Pass | 15.861s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 26.155s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.881s |  |
| Penetration Testing Methodology | ✅ Pass | 5.493s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 23.940s |  |
| SQL Injection Attack Type | ✅ Pass | 3.264s |  |
| Penetration Testing Framework | ✅ Pass | 8.263s |  |
| Vulnerability Assessment Tools | ✅ Pass | 31.030s |  |
| Web Application Security Scanner | ✅ Pass | 9.510s |  |
| Penetration Testing Tool Selection | ✅ Pass | 9.605s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.688s

---

### generator (anthropic/claude-sonnet-4)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.695s |  |
| Text Transform Uppercase | ✅ Pass | 2.374s |  |
| Count from 1 to 5 | ✅ Pass | 2.868s |  |
| Math Calculation | ✅ Pass | 2.312s |  |
| Basic Echo Function | ✅ Pass | 2.639s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.833s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.461s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.390s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.998s |  |
| Search Query Function | ✅ Pass | 3.218s |  |
| Ask Advice Function | ✅ Pass | 3.174s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.469s |  |
| Basic Context Memory Test | ✅ Pass | 2.277s |  |
| Function Argument Memory Test | ✅ Pass | 2.775s |  |
| Function Response Memory Test | ✅ Pass | 2.861s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.769s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.079s |  |
| Penetration Testing Methodology | ✅ Pass | 9.441s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.180s |  |
| SQL Injection Attack Type | ✅ Pass | 3.544s |  |
| Penetration Testing Framework | ✅ Pass | 11.923s |  |
| Web Application Security Scanner | ✅ Pass | 6.453s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.400s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.963s

---

### refiner (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.401s |  |
| Text Transform Uppercase | ✅ Pass | 4.303s |  |
| Count from 1 to 5 | ✅ Pass | 4.922s |  |
| Math Calculation | ✅ Pass | 2.887s |  |
| Basic Echo Function | ✅ Pass | 3.002s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.953s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.482s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.695s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.585s |  |
| Search Query Function | ✅ Pass | 2.323s |  |
| Ask Advice Function | ✅ Pass | 3.020s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.625s |  |
| Basic Context Memory Test | ✅ Pass | 4.572s |  |
| Function Argument Memory Test | ✅ Pass | 3.268s |  |
| Function Response Memory Test | ❌ Fail | 1.075s | expected text '22' not found |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.760s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.006s |  |
| Penetration Testing Methodology | ✅ Pass | 9.135s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.506s |  |
| SQL Injection Attack Type | ✅ Pass | 7.203s |  |
| Penetration Testing Framework | ✅ Pass | 13.738s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.564s |  |
| Web Application Security Scanner | ✅ Pass | 11.781s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.166s

---

### adviser (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.251s |  |
| Text Transform Uppercase | ✅ Pass | 4.290s |  |
| Count from 1 to 5 | ✅ Pass | 4.190s |  |
| Math Calculation | ✅ Pass | 2.833s |  |
| Basic Echo Function | ✅ Pass | 2.839s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.727s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.547s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.673s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.332s |  |
| Search Query Function | ✅ Pass | 2.175s |  |
| Ask Advice Function | ✅ Pass | 3.401s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.741s |  |
| Basic Context Memory Test | ✅ Pass | 4.090s |  |
| Function Argument Memory Test | ✅ Pass | 3.826s |  |
| Function Response Memory Test | ❌ Fail | 1.152s | expected text '22' not found |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.419s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.239s |  |
| Penetration Testing Methodology | ✅ Pass | 9.025s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.672s |  |
| SQL Injection Attack Type | ✅ Pass | 6.847s |  |
| Penetration Testing Framework | ✅ Pass | 14.066s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.670s |  |
| Web Application Security Scanner | ✅ Pass | 12.533s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.937s

---

### reflector (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.890s |  |
| Text Transform Uppercase | ✅ Pass | 1.145s |  |
| Count from 1 to 5 | ✅ Pass | 0.726s |  |
| Math Calculation | ✅ Pass | 0.476s |  |
| Basic Echo Function | ✅ Pass | 1.081s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.522s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.613s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.781s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.845s |  |
| Search Query Function | ✅ Pass | 0.677s |  |
| Ask Advice Function | ✅ Pass | 1.054s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.649s |  |
| Basic Context Memory Test | ✅ Pass | 1.024s |  |
| Function Argument Memory Test | ✅ Pass | 0.492s |  |
| Function Response Memory Test | ✅ Pass | 0.433s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.409s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.511s |  |
| Penetration Testing Methodology | ✅ Pass | 0.860s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.839s |  |
| SQL Injection Attack Type | ✅ Pass | 0.769s |  |
| Penetration Testing Framework | ✅ Pass | 0.893s |  |
| Web Application Security Scanner | ✅ Pass | 0.665s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.039s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.844s

---

### searcher (x-ai/grok-3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.308s |  |
| Text Transform Uppercase | ✅ Pass | 2.117s |  |
| Count from 1 to 5 | ✅ Pass | 2.282s |  |
| Math Calculation | ✅ Pass | 1.465s |  |
| Basic Echo Function | ✅ Pass | 3.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.549s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.243s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.104s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.646s |  |
| Search Query Function | ✅ Pass | 3.187s |  |
| Ask Advice Function | ✅ Pass | 3.329s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.964s |  |
| Basic Context Memory Test | ✅ Pass | 1.811s |  |
| Function Argument Memory Test | ✅ Pass | 1.543s |  |
| Function Response Memory Test | ✅ Pass | 1.900s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.707s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.824s |  |
| Penetration Testing Methodology | ✅ Pass | 4.052s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.088s |  |
| SQL Injection Attack Type | ✅ Pass | 2.101s |  |
| Penetration Testing Framework | ✅ Pass | 4.140s |  |
| Web Application Security Scanner | ✅ Pass | 3.150s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.729s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.976s

---

### enricher (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.425s |  |
| Text Transform Uppercase | ✅ Pass | 0.523s |  |
| Count from 1 to 5 | ✅ Pass | 0.544s |  |
| Math Calculation | ✅ Pass | 0.426s |  |
| Basic Echo Function | ✅ Pass | 0.777s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.646s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.603s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.817s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.808s |  |
| Search Query Function | ✅ Pass | 0.677s |  |
| Ask Advice Function | ✅ Pass | 1.213s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.779s |  |
| Basic Context Memory Test | ✅ Pass | 0.557s |  |
| Function Argument Memory Test | ✅ Pass | 0.498s |  |
| Function Response Memory Test | ✅ Pass | 0.466s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.344s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.645s |  |
| Penetration Testing Methodology | ✅ Pass | 2.362s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.377s |  |
| SQL Injection Attack Type | ✅ Pass | 0.649s |  |
| Penetration Testing Framework | ✅ Pass | 0.751s |  |
| Web Application Security Scanner | ✅ Pass | 2.376s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.923s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.878s

---

### coder (anthropic/claude-sonnet-4)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.899s |  |
| Text Transform Uppercase | ✅ Pass | 2.616s |  |
| Count from 1 to 5 | ✅ Pass | 2.230s |  |
| Math Calculation | ✅ Pass | 2.269s |  |
| Basic Echo Function | ✅ Pass | 3.036s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.631s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.233s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.244s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.638s |  |
| Search Query Function | ✅ Pass | 3.794s |  |
| Ask Advice Function | ✅ Pass | 2.880s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.377s |  |
| Basic Context Memory Test | ✅ Pass | 2.781s |  |
| Function Argument Memory Test | ✅ Pass | 2.656s |  |
| Function Response Memory Test | ✅ Pass | 3.014s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.921s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.594s |  |
| Penetration Testing Methodology | ✅ Pass | 9.317s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.595s |  |
| SQL Injection Attack Type | ✅ Pass | 3.831s |  |
| Penetration Testing Framework | ✅ Pass | 9.788s |  |
| Web Application Security Scanner | ✅ Pass | 6.808s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.478s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.897s

---

### installer (google/gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.420s |  |
| Text Transform Uppercase | ✅ Pass | 1.344s |  |
| Count from 1 to 5 | ✅ Pass | 1.497s |  |
| Math Calculation | ✅ Pass | 1.329s |  |
| Basic Echo Function | ✅ Pass | 1.639s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.052s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.375s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.563s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.934s |  |
| Search Query Function | ✅ Pass | 1.394s |  |
| Ask Advice Function | ✅ Pass | 1.795s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.683s |  |
| Basic Context Memory Test | ✅ Pass | 2.428s |  |
| Function Argument Memory Test | ❌ Fail | 0.620s | expected text 'Go programming language' not found |
| Function Response Memory Test | ✅ Pass | 2.739s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.140s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.704s | expected text 'example\.com' not found |
| Penetration Testing Methodology | ✅ Pass | 7.060s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.403s |  |
| SQL Injection Attack Type | ✅ Pass | 3.628s |  |
| Penetration Testing Framework | ✅ Pass | 8.129s |  |
| Web Application Security Scanner | ✅ Pass | 7.590s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.595s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 2.699s

---

### pentester (moonshotai/kimi-k2)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.581s |  |
| Text Transform Uppercase | ✅ Pass | 0.465s |  |
| Count from 1 to 5 | ✅ Pass | 0.902s |  |
| Math Calculation | ✅ Pass | 0.757s |  |
| Basic Echo Function | ✅ Pass | 0.950s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.188s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.978s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.470s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.759s |  |
| Search Query Function | ✅ Pass | 1.785s |  |
| Ask Advice Function | ✅ Pass | 3.752s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.673s |  |
| Basic Context Memory Test | ✅ Pass | 0.811s |  |
| Function Argument Memory Test | ✅ Pass | 0.357s |  |
| Function Response Memory Test | ✅ Pass | 0.836s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.438s | no tool calls found, expected at least 1 |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.858s |  |
| Penetration Testing Methodology | ✅ Pass | 1.181s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.914s |  |
| SQL Injection Attack Type | ✅ Pass | 0.937s |  |
| Penetration Testing Framework | ✅ Pass | 1.793s |  |
| Web Application Security Scanner | ✅ Pass | 1.082s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.886s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.320s

---

