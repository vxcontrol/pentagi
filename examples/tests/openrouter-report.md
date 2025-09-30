# LLM Agent Testing Report

Generated: Tue, 30 Sep 2025 18:46:00 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | openai/gpt-4.1-mini | false | 23/23 (100.00%) | 1.594s |
| simple_json | openai/gpt-4.1-mini | false | 5/5 (100.00%) | 1.682s |
| primary_agent | openai/gpt-5 | true | 23/23 (100.00%) | 7.285s |
| assistant | openai/gpt-5 | true | 23/23 (100.00%) | 8.135s |
| generator | anthropic/claude-sonnet-4.5 | true | 23/23 (100.00%) | 4.525s |
| refiner | google/gemini-2.5-pro | true | 21/23 (91.30%) | 5.576s |
| adviser | google/gemini-2.5-pro | true | 22/23 (95.65%) | 5.532s |
| reflector | openai/gpt-4.1-mini | false | 23/23 (100.00%) | 1.556s |
| searcher | x-ai/grok-3-mini | true | 22/23 (95.65%) | 4.511s |
| enricher | openai/gpt-4.1-mini | true | 23/23 (100.00%) | 1.597s |
| coder | anthropic/claude-sonnet-4.5 | true | 23/23 (100.00%) | 4.445s |
| installer | google/gemini-2.5-flash | true | 23/23 (100.00%) | 3.276s |
| pentester | moonshotai/kimi-k2-0905 | true | 22/23 (95.65%) | 2.301s |

**Total**: 276/281 (98.22%) successful tests
**Overall average latency**: 4.150s

## Detailed Results

### simple (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 2.727s |  |
| Simple Math | ✅ Pass | 2.809s |  |
| Count from 1 to 5 | ✅ Pass | 3.158s |  |
| Math Calculation | ✅ Pass | 1.255s |  |
| Basic Echo Function | ✅ Pass | 1.112s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.109s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.179s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.270s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.334s |  |
| Search Query Function | ✅ Pass | 1.375s |  |
| Ask Advice Function | ✅ Pass | 1.433s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.436s |  |
| Basic Context Memory Test | ✅ Pass | 1.293s |  |
| Function Argument Memory Test | ✅ Pass | 1.326s |  |
| Function Response Memory Test | ✅ Pass | 1.378s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.802s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.454s |  |
| Penetration Testing Methodology | ✅ Pass | 1.216s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.509s |  |
| SQL Injection Attack Type | ✅ Pass | 2.427s |  |
| Penetration Testing Framework | ✅ Pass | 1.526s |  |
| Web Application Security Scanner | ✅ Pass | 1.093s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.419s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.594s

---

### simple_json (openai/gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 1.574s |  |
| User Profile JSON | ✅ Pass | 1.531s |  |
| Person Information JSON | ✅ Pass | 1.706s |  |
| Vulnerability Report Memory Test | ✅ Pass | 2.108s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.488s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.682s

---

### primary_agent (openai/gpt-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 5.678s |  |
| Simple Math | ✅ Pass | 6.979s |  |
| Math Calculation | ✅ Pass | 4.546s |  |
| Count from 1 to 5 | ✅ Pass | 8.078s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.289s |  |
| Basic Echo Function | ✅ Pass | 7.959s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.885s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 12.785s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.854s |  |
| Ask Advice Function | ✅ Pass | 2.945s |  |
| Basic Context Memory Test | ✅ Pass | 6.477s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 10.439s |  |
| Function Argument Memory Test | ✅ Pass | 5.706s |  |
| Search Query Function | ✅ Pass | 17.551s |  |
| Function Response Memory Test | ✅ Pass | 6.284s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.072s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 12.418s |  |
| Penetration Testing Methodology | ✅ Pass | 8.316s |  |
| SQL Injection Attack Type | ✅ Pass | 5.413s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.698s |  |
| Penetration Testing Framework | ✅ Pass | 5.109s |  |
| Web Application Security Scanner | ✅ Pass | 4.251s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.821s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.285s

---

### assistant (openai/gpt-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 4.176s |  |
| Simple Math | ✅ Pass | 4.241s |  |
| Count from 1 to 5 | ✅ Pass | 4.418s |  |
| Math Calculation | ✅ Pass | 2.466s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.288s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.402s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.997s |  |
| Basic Echo Function | ✅ Pass | 14.115s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Ask Advice Function | ✅ Pass | 3.039s |  |
| Search Query Function | ✅ Pass | 9.098s |  |
| Function Argument Memory Test | ✅ Pass | 3.562s |  |
| Basic Context Memory Test | ✅ Pass | 8.180s |  |
| Function Response Memory Test | ✅ Pass | 4.814s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 15.423s |  |
| JSON Response Function | ✅ Pass | 24.602s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.121s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.563s |  |
| SQL Injection Attack Type | ✅ Pass | 7.029s |  |
| Penetration Testing Methodology | ✅ Pass | 16.605s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.711s |  |
| Web Application Security Scanner | ✅ Pass | 3.749s |  |
| Penetration Testing Framework | ✅ Pass | 7.171s |  |
| Penetration Testing Tool Selection | ✅ Pass | 9.317s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 8.135s

---

### generator (anthropic/claude-sonnet-4.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.796s |  |
| Text Transform Uppercase | ✅ Pass | 4.900s |  |
| Count from 1 to 5 | ✅ Pass | 3.211s |  |
| Math Calculation | ✅ Pass | 2.543s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.894s |  |
| Basic Echo Function | ✅ Pass | 3.969s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.810s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.255s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.649s |  |
| Search Query Function | ✅ Pass | 3.659s |  |
| Ask Advice Function | ✅ Pass | 3.011s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.007s |  |
| Basic Context Memory Test | ✅ Pass | 2.584s |  |
| Function Argument Memory Test | ✅ Pass | 3.795s |  |
| Function Response Memory Test | ✅ Pass | 3.613s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.593s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.289s |  |
| Penetration Testing Methodology | ✅ Pass | 11.070s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.953s |  |
| SQL Injection Attack Type | ✅ Pass | 4.623s |  |
| Web Application Security Scanner | ✅ Pass | 6.242s |  |
| Penetration Testing Framework | ✅ Pass | 9.207s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.393s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.525s

---

### refiner (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.765s |  |
| Text Transform Uppercase | ✅ Pass | 8.021s |  |
| Count from 1 to 5 | ✅ Pass | 5.828s |  |
| Math Calculation | ✅ Pass | 3.337s |  |
| Basic Echo Function | ✅ Pass | 3.749s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.356s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.657s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.911s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.989s |  |
| Search Query Function | ✅ Pass | 4.489s |  |
| Ask Advice Function | ✅ Pass | 3.256s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.329s |  |
| Function Argument Memory Test | ❌ Fail | 1.329s | expected text 'Go programming language' not found |
| Basic Context Memory Test | ✅ Pass | 4.987s |  |
| Function Response Memory Test | ❌ Fail | 1.624s | expected text '22' not found |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.209s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.357s |  |
| Penetration Testing Methodology | ✅ Pass | 9.174s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.362s |  |
| SQL Injection Attack Type | ✅ Pass | 7.982s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.862s |  |
| Web Application Security Scanner | ✅ Pass | 11.705s |  |
| Penetration Testing Framework | ✅ Pass | 15.968s |  |

**Summary**: 21/23 (91.30%) successful tests

**Average latency**: 5.576s

---

### adviser (google/gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.749s |  |
| Count from 1 to 5 | ✅ Pass | 4.983s |  |
| Text Transform Uppercase | ✅ Pass | 7.715s |  |
| Math Calculation | ✅ Pass | 3.160s |  |
| Basic Echo Function | ✅ Pass | 3.491s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.304s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.330s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.641s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.319s |  |
| Search Query Function | ✅ Pass | 3.352s |  |
| Ask Advice Function | ✅ Pass | 2.876s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.286s |  |
| Basic Context Memory Test | ✅ Pass | 5.184s |  |
| Function Argument Memory Test | ✅ Pass | 3.338s |  |
| Function Response Memory Test | ❌ Fail | 1.962s | expected text '22' not found |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.316s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.781s |  |
| Penetration Testing Methodology | ✅ Pass | 10.426s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.932s |  |
| SQL Injection Attack Type | ✅ Pass | 6.701s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.242s |  |
| Penetration Testing Framework | ✅ Pass | 13.500s |  |
| Web Application Security Scanner | ✅ Pass | 12.631s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 5.532s

---

### reflector (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.664s |  |
| Text Transform Uppercase | ✅ Pass | 3.352s |  |
| Count from 1 to 5 | ✅ Pass | 1.470s |  |
| Math Calculation | ✅ Pass | 1.184s |  |
| Basic Echo Function | ✅ Pass | 1.459s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.206s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.110s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.144s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.294s |  |
| Search Query Function | ✅ Pass | 1.555s |  |
| Ask Advice Function | ✅ Pass | 1.328s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.209s |  |
| Basic Context Memory Test | ✅ Pass | 1.465s |  |
| Function Argument Memory Test | ✅ Pass | 1.186s |  |
| Function Response Memory Test | ✅ Pass | 1.476s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.031s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.505s |  |
| Penetration Testing Methodology | ✅ Pass | 1.356s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.687s |  |
| SQL Injection Attack Type | ✅ Pass | 1.316s |  |
| Penetration Testing Framework | ✅ Pass | 1.093s |  |
| Web Application Security Scanner | ✅ Pass | 1.298s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.387s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.556s

---

### searcher (x-ai/grok-3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.840s |  |
| Text Transform Uppercase | ✅ Pass | 5.601s |  |
| Count from 1 to 5 | ✅ Pass | 4.014s |  |
| Math Calculation | ✅ Pass | 3.175s |  |
| Basic Echo Function | ✅ Pass | 4.012s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.994s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.596s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.705s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.163s |  |
| Search Query Function | ✅ Pass | 3.663s |  |
| Ask Advice Function | ✅ Pass | 4.934s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.816s |  |
| Basic Context Memory Test | ✅ Pass | 3.479s |  |
| Function Argument Memory Test | ✅ Pass | 3.226s |  |
| Function Response Memory Test | ✅ Pass | 3.040s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.192s |  |
| Penetration Testing Methodology | ✅ Pass | 4.935s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 8.973s | expected text 'example\.com' not found |
| Vulnerability Assessment Tools | ✅ Pass | 6.358s |  |
| SQL Injection Attack Type | ✅ Pass | 3.042s |  |
| Penetration Testing Framework | ✅ Pass | 5.377s |  |
| Web Application Security Scanner | ✅ Pass | 4.338s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.267s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 4.511s

---

### enricher (openai/gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.796s |  |
| Text Transform Uppercase | ✅ Pass | 3.117s |  |
| Count from 1 to 5 | ✅ Pass | 1.902s |  |
| Math Calculation | ✅ Pass | 0.887s |  |
| Basic Echo Function | ✅ Pass | 1.260s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.943s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.273s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.393s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.519s |  |
| Search Query Function | ✅ Pass | 1.304s |  |
| Ask Advice Function | ✅ Pass | 1.661s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.592s |  |
| Basic Context Memory Test | ✅ Pass | 1.266s |  |
| Function Argument Memory Test | ✅ Pass | 1.239s |  |
| Function Response Memory Test | ✅ Pass | 1.617s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.076s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.278s |  |
| Penetration Testing Methodology | ✅ Pass | 1.934s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.300s |  |
| SQL Injection Attack Type | ✅ Pass | 1.211s |  |
| Penetration Testing Framework | ✅ Pass | 1.614s |  |
| Web Application Security Scanner | ✅ Pass | 1.195s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.334s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.597s

---

### coder (anthropic/claude-sonnet-4.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.233s |  |
| Text Transform Uppercase | ✅ Pass | 5.161s |  |
| Count from 1 to 5 | ✅ Pass | 3.227s |  |
| Math Calculation | ✅ Pass | 2.882s |  |
| Basic Echo Function | ✅ Pass | 3.143s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.506s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.003s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.763s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.419s |  |
| Search Query Function | ✅ Pass | 3.017s |  |
| Ask Advice Function | ✅ Pass | 2.999s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.992s |  |
| Basic Context Memory Test | ✅ Pass | 3.126s |  |
| Function Argument Memory Test | ✅ Pass | 3.670s |  |
| Function Response Memory Test | ✅ Pass | 3.248s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.631s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.598s |  |
| Penetration Testing Methodology | ✅ Pass | 11.220s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.139s |  |
| SQL Injection Attack Type | ✅ Pass | 4.317s |  |
| Penetration Testing Framework | ✅ Pass | 7.797s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.140s |  |
| Web Application Security Scanner | ✅ Pass | 8.004s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.445s

---

### installer (google/gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.212s |  |
| Text Transform Uppercase | ✅ Pass | 2.828s |  |
| Count from 1 to 5 | ✅ Pass | 0.800s |  |
| Math Calculation | ✅ Pass | 1.529s |  |
| Basic Echo Function | ✅ Pass | 1.841s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.011s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.480s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.747s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.678s |  |
| Search Query Function | ✅ Pass | 1.535s |  |
| Ask Advice Function | ✅ Pass | 2.439s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.017s |  |
| Basic Context Memory Test | ✅ Pass | 2.790s |  |
| Function Response Memory Test | ✅ Pass | 0.868s |  |
| Function Argument Memory Test | ✅ Pass | 3.503s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.933s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.969s |  |
| Penetration Testing Methodology | ✅ Pass | 6.046s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.005s |  |
| SQL Injection Attack Type | ✅ Pass | 2.731s |  |
| Web Application Security Scanner | ✅ Pass | 4.804s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.752s |  |
| Penetration Testing Framework | ✅ Pass | 13.820s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.276s

---

### pentester (moonshotai/kimi-k2-0905)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.485s |  |
| Text Transform Uppercase | ✅ Pass | 1.845s |  |
| Count from 1 to 5 | ✅ Pass | 1.481s |  |
| Math Calculation | ✅ Pass | 1.625s |  |
| Basic Echo Function | ✅ Pass | 1.611s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.693s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.843s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.580s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.278s |  |
| Search Query Function | ✅ Pass | 1.168s |  |
| Search Query Function Streaming | ❌ Fail | 0.747s | streaming tool call func returned an error: tool call name is required |
| Ask Advice Function | ✅ Pass | 4.458s |  |
| Basic Context Memory Test | ✅ Pass | 1.566s |  |
| Function Argument Memory Test | ✅ Pass | 1.148s |  |
| Function Response Memory Test | ✅ Pass | 0.811s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.292s |  |
| Penetration Testing Methodology | ✅ Pass | 2.511s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 11.318s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.730s |  |
| SQL Injection Attack Type | ✅ Pass | 3.724s |  |
| Penetration Testing Framework | ✅ Pass | 0.782s |  |
| Web Application Security Scanner | ✅ Pass | 3.533s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.674s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.301s

---

