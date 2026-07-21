# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 20:33:49 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | nemotron-3-nano:30b-cloud | false | 23/23 (100.00%) | 2.553s |
| simple_json | nemotron-3-nano:30b-cloud | false | 7/7 (100.00%) | 2.410s |
| primary_agent | kimi-k2.6:cloud | false | 23/23 (100.00%) | 2.718s |
| assistant | kimi-k2.6:cloud | false | 23/23 (100.00%) | 2.847s |
| generator | mistral-large-3:675b-cloud | false | 23/23 (100.00%) | 1.886s |
| refiner | nemotron-3-ultra:cloud | false | 23/23 (100.00%) | 5.631s |
| adviser | glm-5.1:cloud | false | 23/23 (100.00%) | 3.260s |
| reflector | nemotron-3-nano:30b-cloud | false | 23/23 (100.00%) | 2.096s |
| searcher | qwen3.5:397b-cloud | false | 23/23 (100.00%) | 5.856s |
| enricher | minimax-m2.7:cloud | false | 23/23 (100.00%) | 3.425s |
| coder | kimi-k2.7-code:cloud | false | 22/23 (95.65%) | 2.322s |
| installer | kimi-k2.7-code:cloud | false | 23/23 (100.00%) | 2.550s |
| pentester | deepseek-v4-pro:cloud | false | 23/23 (100.00%) | 1.824s |

**Total**: 282/283 (99.65%) successful tests
**Overall average latency**: 3.064s

## Detailed Results

### simple (nemotron-3-nano:30b-cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.189s |  |
| Count from 1 to 5 | ✅ Pass | 1.408s |  |
| Text Transform Uppercase | ✅ Pass | 9.093s |  |
| Math Calculation | ✅ Pass | 2.943s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.950s |  |
| Basic Echo Function | ✅ Pass | 6.563s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.039s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.664s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.877s |  |
| Search Query Function | ✅ Pass | 2.516s |  |
| Ask Advice Function | ✅ Pass | 1.608s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.250s |  |
| Basic Context Memory Test | ✅ Pass | 2.386s |  |
| Function Argument Memory Test | ✅ Pass | 2.295s |  |
| Function Response Memory Test | ✅ Pass | 1.072s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.079s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.494s |  |
| Penetration Testing Methodology | ✅ Pass | 1.433s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.904s |  |
| SQL Injection Attack Type | ✅ Pass | 2.914s |  |
| Penetration Testing Framework | ✅ Pass | 2.560s |  |
| Web Application Security Scanner | ✅ Pass | 1.253s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.227s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.553s

---

### simple_json (nemotron-3-nano:30b-cloud)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 1.468s |  |
| Project Information JSON | ✅ Pass | 1.558s |  |
| User Profile JSON | ✅ Pass | 1.822s |  |
| Vulnerability Report Memory Test | ✅ Pass | 3.089s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.354s |  |
| JSON Array Response Without Schema | ✅ Pass | 5.023s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 2.552s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 2.410s

---

### primary_agent (kimi-k2.6:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.671s |  |
| Text Transform Uppercase | ✅ Pass | 1.436s |  |
| Count from 1 to 5 | ✅ Pass | 2.245s |  |
| Math Calculation | ✅ Pass | 1.513s |  |
| Basic Echo Function | ✅ Pass | 1.420s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.155s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.994s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.485s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.924s |  |
| Search Query Function | ✅ Pass | 1.882s |  |
| Ask Advice Function | ✅ Pass | 1.658s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.927s |  |
| Basic Context Memory Test | ✅ Pass | 2.056s |  |
| Function Argument Memory Test | ✅ Pass | 3.303s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.341s |  |
| Function Response Memory Test | ✅ Pass | 6.369s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.381s |  |
| Penetration Testing Methodology | ✅ Pass | 3.671s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.404s |  |
| SQL Injection Attack Type | ✅ Pass | 2.150s |  |
| Penetration Testing Framework | ✅ Pass | 6.441s |  |
| Web Application Security Scanner | ✅ Pass | 2.062s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.018s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.718s

---

### assistant (kimi-k2.6:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.077s |  |
| Text Transform Uppercase | ✅ Pass | 1.443s |  |
| Count from 1 to 5 | ✅ Pass | 1.803s |  |
| Math Calculation | ✅ Pass | 1.713s |  |
| Basic Echo Function | ✅ Pass | 1.231s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.716s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.366s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.403s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.741s |  |
| Search Query Function | ✅ Pass | 1.544s |  |
| Ask Advice Function | ✅ Pass | 1.553s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.748s |  |
| Basic Context Memory Test | ✅ Pass | 1.659s |  |
| Function Argument Memory Test | ✅ Pass | 3.101s |  |
| Function Response Memory Test | ✅ Pass | 1.532s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.992s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.628s |  |
| Penetration Testing Methodology | ✅ Pass | 4.080s |  |
| SQL Injection Attack Type | ✅ Pass | 1.767s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.714s |  |
| Penetration Testing Framework | ✅ Pass | 7.542s |  |
| Web Application Security Scanner | ✅ Pass | 2.199s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.912s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.847s

---

### generator (mistral-large-3:675b-cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.057s |  |
| Text Transform Uppercase | ✅ Pass | 0.906s |  |
| Count from 1 to 5 | ✅ Pass | 1.016s |  |
| Math Calculation | ✅ Pass | 0.843s |  |
| Basic Echo Function | ✅ Pass | 1.010s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.334s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.931s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.129s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.149s |  |
| Search Query Function | ✅ Pass | 1.324s |  |
| Ask Advice Function | ✅ Pass | 1.236s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.226s |  |
| Basic Context Memory Test | ✅ Pass | 0.929s |  |
| Function Argument Memory Test | ✅ Pass | 0.803s |  |
| Function Response Memory Test | ✅ Pass | 0.882s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.688s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.908s |  |
| Penetration Testing Methodology | ✅ Pass | 1.975s |  |
| SQL Injection Attack Type | ✅ Pass | 0.980s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.025s |  |
| Penetration Testing Framework | ✅ Pass | 2.884s |  |
| Web Application Security Scanner | ✅ Pass | 2.011s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.118s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.886s

---

### refiner (nemotron-3-ultra:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.812s |  |
| Count from 1 to 5 | ✅ Pass | 4.445s |  |
| Math Calculation | ✅ Pass | 2.275s |  |
| Text Transform Uppercase | ✅ Pass | 12.666s |  |
| Basic Echo Function | ✅ Pass | 3.633s |  |
| Streaming Simple Math Streaming | ✅ Pass | 7.361s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.497s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 10.178s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 2.298s |  |
| Ask Advice Function | ✅ Pass | 4.959s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.867s |  |
| Basic Context Memory Test | ✅ Pass | 1.962s |  |
| Function Argument Memory Test | ✅ Pass | 1.149s |  |
| Function Response Memory Test | ✅ Pass | 1.863s |  |
| JSON Response Function | ✅ Pass | 31.697s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.205s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.164s |  |
| Penetration Testing Methodology | ✅ Pass | 4.731s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.192s |  |
| SQL Injection Attack Type | ✅ Pass | 1.133s |  |
| Penetration Testing Framework | ✅ Pass | 6.045s |  |
| Web Application Security Scanner | ✅ Pass | 1.541s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.820s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.631s

---

### adviser (glm-5.1:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.888s |  |
| Text Transform Uppercase | ✅ Pass | 3.135s |  |
| Count from 1 to 5 | ✅ Pass | 2.208s |  |
| Math Calculation | ✅ Pass | 1.384s |  |
| Basic Echo Function | ✅ Pass | 2.010s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.185s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.989s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.622s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.383s |  |
| Search Query Function | ✅ Pass | 3.991s |  |
| Ask Advice Function | ✅ Pass | 1.618s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.838s |  |
| Basic Context Memory Test | ✅ Pass | 4.070s |  |
| Function Argument Memory Test | ✅ Pass | 2.376s |  |
| Function Response Memory Test | ✅ Pass | 3.116s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.410s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.902s |  |
| Penetration Testing Methodology | ✅ Pass | 4.680s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.330s |  |
| SQL Injection Attack Type | ✅ Pass | 2.247s |  |
| Penetration Testing Framework | ✅ Pass | 9.490s |  |
| Web Application Security Scanner | ✅ Pass | 4.576s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.529s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.260s

---

### reflector (nemotron-3-nano:30b-cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.271s |  |
| Text Transform Uppercase | ✅ Pass | 1.096s |  |
| Count from 1 to 5 | ✅ Pass | 1.697s |  |
| Math Calculation | ✅ Pass | 1.010s |  |
| Basic Echo Function | ✅ Pass | 1.577s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.141s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.362s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.766s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.121s |  |
| Search Query Function | ✅ Pass | 1.656s |  |
| Ask Advice Function | ✅ Pass | 1.843s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.452s |  |
| Basic Context Memory Test | ✅ Pass | 2.101s |  |
| Function Argument Memory Test | ✅ Pass | 2.782s |  |
| Function Response Memory Test | ✅ Pass | 2.649s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.311s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.175s |  |
| Penetration Testing Methodology | ✅ Pass | 1.141s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.713s |  |
| SQL Injection Attack Type | ✅ Pass | 1.929s |  |
| Penetration Testing Framework | ✅ Pass | 1.143s |  |
| Web Application Security Scanner | ✅ Pass | 1.367s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.884s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.096s

---

### searcher (qwen3.5:397b-cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.078s |  |
| Text Transform Uppercase | ✅ Pass | 5.277s |  |
| Count from 1 to 5 | ✅ Pass | 5.322s |  |
| Math Calculation | ✅ Pass | 2.751s |  |
| Basic Echo Function | ✅ Pass | 1.776s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.460s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.329s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.876s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.141s |  |
| Search Query Function | ✅ Pass | 1.866s |  |
| Ask Advice Function | ✅ Pass | 2.291s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.823s |  |
| Basic Context Memory Test | ✅ Pass | 4.063s |  |
| Function Argument Memory Test | ✅ Pass | 1.898s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.962s |  |
| Function Response Memory Test | ✅ Pass | 11.805s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.704s |  |
| Penetration Testing Methodology | ✅ Pass | 10.296s |  |
| SQL Injection Attack Type | ✅ Pass | 8.448s |  |
| Penetration Testing Framework | ✅ Pass | 7.163s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.838s |  |
| Web Application Security Scanner | ✅ Pass | 8.118s |  |
| Vulnerability Assessment Tools | ✅ Pass | 34.393s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.856s

---

### enricher (minimax-m2.7:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.714s |  |
| Text Transform Uppercase | ✅ Pass | 3.532s |  |
| Count from 1 to 5 | ✅ Pass | 5.140s |  |
| Math Calculation | ✅ Pass | 1.952s |  |
| Basic Echo Function | ✅ Pass | 1.707s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.469s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.203s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.477s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.636s |  |
| Search Query Function | ✅ Pass | 2.099s |  |
| Ask Advice Function | ✅ Pass | 2.682s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.241s |  |
| Basic Context Memory Test | ✅ Pass | 1.996s |  |
| Function Argument Memory Test | ✅ Pass | 2.087s |  |
| Function Response Memory Test | ✅ Pass | 2.030s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.810s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.592s |  |
| Penetration Testing Methodology | ✅ Pass | 4.070s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.512s |  |
| SQL Injection Attack Type | ✅ Pass | 6.386s |  |
| Penetration Testing Framework | ✅ Pass | 6.082s |  |
| Web Application Security Scanner | ✅ Pass | 3.166s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.176s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.425s

---

### coder (kimi-k2.7-code:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.628s |  |
| Text Transform Uppercase | ✅ Pass | 1.266s |  |
| Count from 1 to 5 | ✅ Pass | 2.849s |  |
| Math Calculation | ✅ Pass | 4.151s |  |
| Basic Echo Function | ✅ Pass | 1.318s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.619s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.497s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.220s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.828s |  |
| Search Query Function | ✅ Pass | 1.327s |  |
| Ask Advice Function | ✅ Pass | 1.355s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.443s |  |
| Basic Context Memory Test | ✅ Pass | 1.541s |  |
| Function Argument Memory Test | ✅ Pass | 1.299s |  |
| Function Response Memory Test | ✅ Pass | 1.218s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.061s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.109s |  |
| Penetration Testing Methodology | ✅ Pass | 2.873s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.599s |  |
| SQL Injection Attack Type | ✅ Pass | 1.728s |  |
| Penetration Testing Framework | ✅ Pass | 2.782s |  |
| Web Application Security Scanner | ✅ Pass | 1.747s |  |
| Penetration Testing Tool Selection | ❌ Fail | 4.947s | expected function 'nmap' not found in tool calls: expected function nmap not found in tool calls |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.322s

---

### installer (kimi-k2.7-code:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.531s |  |
| Text Transform Uppercase | ✅ Pass | 1.436s |  |
| Count from 1 to 5 | ✅ Pass | 2.494s |  |
| Math Calculation | ✅ Pass | 3.434s |  |
| Basic Echo Function | ✅ Pass | 1.372s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.327s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.399s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.221s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.597s |  |
| Search Query Function | ✅ Pass | 1.715s |  |
| Ask Advice Function | ✅ Pass | 1.459s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.567s |  |
| Basic Context Memory Test | ✅ Pass | 1.722s |  |
| Function Argument Memory Test | ✅ Pass | 1.202s |  |
| Function Response Memory Test | ✅ Pass | 1.433s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.963s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.737s |  |
| Penetration Testing Methodology | ✅ Pass | 2.059s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.085s |  |
| SQL Injection Attack Type | ✅ Pass | 1.453s |  |
| Penetration Testing Framework | ✅ Pass | 2.092s |  |
| Web Application Security Scanner | ✅ Pass | 1.891s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.455s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.550s

---

### pentester (deepseek-v4-pro:cloud)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.062s |  |
| Text Transform Uppercase | ✅ Pass | 1.226s |  |
| Count from 1 to 5 | ✅ Pass | 1.536s |  |
| Math Calculation | ✅ Pass | 1.281s |  |
| Basic Echo Function | ✅ Pass | 1.337s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.349s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.193s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.361s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.357s |  |
| Search Query Function | ✅ Pass | 1.301s |  |
| Ask Advice Function | ✅ Pass | 1.337s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.356s |  |
| Basic Context Memory Test | ✅ Pass | 2.433s |  |
| Function Argument Memory Test | ✅ Pass | 1.422s |  |
| Function Response Memory Test | ✅ Pass | 1.598s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.345s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.396s |  |
| Penetration Testing Methodology | ✅ Pass | 6.055s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.079s |  |
| SQL Injection Attack Type | ✅ Pass | 1.424s |  |
| Penetration Testing Framework | ✅ Pass | 2.163s |  |
| Web Application Security Scanner | ✅ Pass | 1.613s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.724s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.824s

---

