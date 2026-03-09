# LLM Agent Testing Report

Generated: Thu, 05 Mar 2026 18:12:24 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-oss:120b | false | 23/23 (100.00%) | 1.398s |
| simple_json | gpt-oss:120b | false | 5/5 (100.00%) | 1.451s |
| primary_agent | gpt-oss:120b | false | 22/23 (95.65%) | 1.343s |
| assistant | gpt-oss:120b | false | 23/23 (100.00%) | 1.369s |
| generator | gpt-oss:120b | false | 22/23 (95.65%) | 1.339s |
| refiner | gpt-oss:120b | false | 23/23 (100.00%) | 1.285s |
| adviser | gpt-oss:120b | false | 23/23 (100.00%) | 1.240s |
| reflector | gpt-oss:120b | false | 23/23 (100.00%) | 1.229s |
| searcher | gpt-oss:120b | false | 22/23 (95.65%) | 1.180s |
| enricher | gpt-oss:120b | false | 22/23 (95.65%) | 1.281s |
| coder | gpt-oss:120b | false | 23/23 (100.00%) | 1.218s |
| installer | gpt-oss:120b | false | 23/23 (100.00%) | 1.260s |
| pentester | gpt-oss:120b | false | 22/23 (95.65%) | 1.203s |

**Total**: 276/281 (98.22%) successful tests
**Overall average latency**: 1.282s

## Detailed Results

### simple (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.491s |  |
| Text Transform Uppercase | ✅ Pass | 1.047s |  |
| Count from 1 to 5 | ✅ Pass | 0.896s |  |
| Math Calculation | ✅ Pass | 0.883s |  |
| Basic Echo Function | ✅ Pass | 0.963s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.998s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.044s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.429s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.137s |  |
| Search Query Function | ✅ Pass | 1.015s |  |
| Ask Advice Function | ✅ Pass | 1.187s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.969s |  |
| Basic Context Memory Test | ✅ Pass | 1.158s |  |
| Function Argument Memory Test | ✅ Pass | 1.198s |  |
| Function Response Memory Test | ✅ Pass | 1.100s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.276s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.169s |  |
| Penetration Testing Methodology | ✅ Pass | 1.275s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.485s |  |
| SQL Injection Attack Type | ✅ Pass | 1.718s |  |
| Penetration Testing Framework | ✅ Pass | 1.261s |  |
| Web Application Security Scanner | ✅ Pass | 1.225s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.212s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.398s

---

### simple_json (gpt-oss:120b)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 1.653s |  |
| Person Information JSON | ✅ Pass | 1.338s |  |
| Project Information JSON | ✅ Pass | 1.515s |  |
| User Profile JSON | ✅ Pass | 1.588s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.159s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.451s

---

### primary_agent (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.896s |  |
| Text Transform Uppercase | ✅ Pass | 1.101s |  |
| Count from 1 to 5 | ✅ Pass | 1.103s |  |
| Math Calculation | ✅ Pass | 1.249s |  |
| Basic Echo Function | ✅ Pass | 1.002s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.870s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.200s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.312s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.080s |  |
| Search Query Function | ✅ Pass | 0.969s |  |
| Ask Advice Function | ✅ Pass | 1.030s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.482s |  |
| Basic Context Memory Test | ✅ Pass | 1.174s |  |
| Function Argument Memory Test | ✅ Pass | 1.106s |  |
| Function Response Memory Test | ✅ Pass | 1.109s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.373s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.396s |  |
| Penetration Testing Methodology | ✅ Pass | 1.183s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.247s |  |
| SQL Injection Attack Type | ✅ Pass | 3.982s |  |
| Penetration Testing Framework | ✅ Pass | 1.473s |  |
| Web Application Security Scanner | ✅ Pass | 1.145s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.395s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.343s

---

### assistant (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.934s |  |
| Text Transform Uppercase | ✅ Pass | 0.921s |  |
| Count from 1 to 5 | ✅ Pass | 1.013s |  |
| Math Calculation | ✅ Pass | 0.898s |  |
| Basic Echo Function | ✅ Pass | 0.986s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.885s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.225s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.122s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.919s |  |
| Search Query Function | ✅ Pass | 1.270s |  |
| Ask Advice Function | ✅ Pass | 1.092s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.937s |  |
| Basic Context Memory Test | ✅ Pass | 1.179s |  |
| Function Argument Memory Test | ✅ Pass | 1.136s |  |
| Function Response Memory Test | ✅ Pass | 1.183s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.503s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.281s |  |
| Penetration Testing Methodology | ✅ Pass | 1.215s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.553s |  |
| SQL Injection Attack Type | ✅ Pass | 3.831s |  |
| Penetration Testing Framework | ✅ Pass | 1.037s |  |
| Web Application Security Scanner | ✅ Pass | 1.120s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.230s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.369s

---

### generator (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.884s |  |
| Text Transform Uppercase | ✅ Pass | 1.102s |  |
| Count from 1 to 5 | ✅ Pass | 0.958s |  |
| Math Calculation | ✅ Pass | 1.046s |  |
| Basic Echo Function | ✅ Pass | 1.050s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.896s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.128s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.024s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.522s |  |
| Search Query Function | ✅ Pass | 1.082s |  |
| Ask Advice Function | ✅ Pass | 1.034s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.040s |  |
| Basic Context Memory Test | ✅ Pass | 1.403s |  |
| Function Argument Memory Test | ✅ Pass | 1.079s |  |
| Function Response Memory Test | ✅ Pass | 1.421s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.795s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.186s |  |
| Penetration Testing Methodology | ✅ Pass | 1.135s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.453s |  |
| SQL Injection Attack Type | ✅ Pass | 4.510s |  |
| Penetration Testing Framework | ✅ Pass | 1.658s |  |
| Web Application Security Scanner | ✅ Pass | 1.139s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.239s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.339s

---

### refiner (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.864s |  |
| Text Transform Uppercase | ✅ Pass | 1.707s |  |
| Count from 1 to 5 | ✅ Pass | 1.007s |  |
| Math Calculation | ✅ Pass | 1.004s |  |
| Basic Echo Function | ✅ Pass | 1.173s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.011s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.928s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.039s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.010s |  |
| Search Query Function | ✅ Pass | 1.068s |  |
| Ask Advice Function | ✅ Pass | 1.071s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.998s |  |
| Basic Context Memory Test | ✅ Pass | 1.102s |  |
| Function Argument Memory Test | ✅ Pass | 1.199s |  |
| Function Response Memory Test | ✅ Pass | 1.038s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.311s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.178s |  |
| Penetration Testing Methodology | ✅ Pass | 1.379s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.752s |  |
| SQL Injection Attack Type | ✅ Pass | 3.868s |  |
| Penetration Testing Framework | ✅ Pass | 1.064s |  |
| Web Application Security Scanner | ✅ Pass | 1.100s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.671s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.285s

---

### adviser (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.202s |  |
| Text Transform Uppercase | ✅ Pass | 0.921s |  |
| Count from 1 to 5 | ✅ Pass | 0.962s |  |
| Math Calculation | ✅ Pass | 0.960s |  |
| Basic Echo Function | ✅ Pass | 1.086s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.907s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.965s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.959s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.101s |  |
| Search Query Function | ✅ Pass | 1.005s |  |
| Ask Advice Function | ✅ Pass | 1.049s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.083s |  |
| Basic Context Memory Test | ✅ Pass | 1.114s |  |
| Function Argument Memory Test | ✅ Pass | 1.035s |  |
| Function Response Memory Test | ✅ Pass | 1.001s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.370s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.135s |  |
| Penetration Testing Methodology | ✅ Pass | 2.019s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.573s |  |
| SQL Injection Attack Type | ✅ Pass | 1.437s |  |
| Penetration Testing Framework | ✅ Pass | 1.378s |  |
| Web Application Security Scanner | ✅ Pass | 1.079s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.158s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.240s

---

### reflector (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.834s |  |
| Text Transform Uppercase | ✅ Pass | 1.040s |  |
| Count from 1 to 5 | ✅ Pass | 1.190s |  |
| Math Calculation | ✅ Pass | 0.915s |  |
| Basic Echo Function | ✅ Pass | 1.050s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.076s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.197s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.925s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.075s |  |
| Search Query Function | ✅ Pass | 1.052s |  |
| Ask Advice Function | ✅ Pass | 1.291s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.083s |  |
| Basic Context Memory Test | ✅ Pass | 1.799s |  |
| Function Argument Memory Test | ✅ Pass | 1.339s |  |
| Function Response Memory Test | ✅ Pass | 0.996s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.745s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.259s |  |
| Penetration Testing Methodology | ✅ Pass | 1.044s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.452s |  |
| SQL Injection Attack Type | ✅ Pass | 1.330s |  |
| Penetration Testing Framework | ✅ Pass | 0.976s |  |
| Web Application Security Scanner | ✅ Pass | 1.101s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.479s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.229s

---

### searcher (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.910s |  |
| Text Transform Uppercase | ✅ Pass | 1.046s |  |
| Count from 1 to 5 | ✅ Pass | 0.902s |  |
| Math Calculation | ✅ Pass | 1.029s |  |
| Basic Echo Function | ✅ Pass | 1.376s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.394s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.670s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.105s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.188s |  |
| Search Query Function | ✅ Pass | 1.158s |  |
| Ask Advice Function | ✅ Pass | 1.071s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.843s |  |
| Basic Context Memory Test | ✅ Pass | 1.075s |  |
| Function Argument Memory Test | ✅ Pass | 1.027s |  |
| Function Response Memory Test | ✅ Pass | 1.038s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.625s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.159s |  |
| Penetration Testing Methodology | ✅ Pass | 1.303s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.499s |  |
| SQL Injection Attack Type | ✅ Pass | 1.109s |  |
| Penetration Testing Framework | ✅ Pass | 1.128s |  |
| Web Application Security Scanner | ✅ Pass | 1.189s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.284s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.180s

---

### enricher (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.643s |  |
| Text Transform Uppercase | ✅ Pass | 0.776s |  |
| Count from 1 to 5 | ✅ Pass | 1.023s |  |
| Math Calculation | ✅ Pass | 1.061s |  |
| Basic Echo Function | ✅ Pass | 0.918s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.944s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.206s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.036s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.993s |  |
| Search Query Function | ✅ Pass | 0.981s |  |
| Ask Advice Function | ✅ Pass | 2.644s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.904s |  |
| Basic Context Memory Test | ✅ Pass | 1.120s |  |
| Function Argument Memory Test | ✅ Pass | 1.211s |  |
| Function Response Memory Test | ✅ Pass | 0.987s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.877s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.384s |  |
| Penetration Testing Methodology | ✅ Pass | 1.118s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.482s |  |
| SQL Injection Attack Type | ✅ Pass | 1.726s |  |
| Penetration Testing Framework | ✅ Pass | 1.043s |  |
| Web Application Security Scanner | ✅ Pass | 2.143s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.231s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.281s

---

### coder (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.929s |  |
| Text Transform Uppercase | ✅ Pass | 0.941s |  |
| Count from 1 to 5 | ✅ Pass | 0.933s |  |
| Math Calculation | ✅ Pass | 1.311s |  |
| Basic Echo Function | ✅ Pass | 1.008s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.943s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.208s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.367s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.380s |  |
| Search Query Function | ✅ Pass | 0.942s |  |
| Ask Advice Function | ✅ Pass | 1.249s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.289s |  |
| Basic Context Memory Test | ✅ Pass | 1.099s |  |
| Function Argument Memory Test | ✅ Pass | 1.122s |  |
| Function Response Memory Test | ✅ Pass | 0.967s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.401s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.355s |  |
| Penetration Testing Methodology | ✅ Pass | 1.204s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.564s |  |
| SQL Injection Attack Type | ✅ Pass | 1.421s |  |
| Penetration Testing Framework | ✅ Pass | 1.507s |  |
| Web Application Security Scanner | ✅ Pass | 1.142s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.712s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.218s

---

### installer (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.848s |  |
| Text Transform Uppercase | ✅ Pass | 1.111s |  |
| Count from 1 to 5 | ✅ Pass | 0.968s |  |
| Math Calculation | ✅ Pass | 1.029s |  |
| Basic Echo Function | ✅ Pass | 0.983s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.016s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.492s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.627s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.052s |  |
| Search Query Function | ✅ Pass | 0.978s |  |
| Ask Advice Function | ✅ Pass | 1.163s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.407s |  |
| Basic Context Memory Test | ✅ Pass | 1.079s |  |
| Function Argument Memory Test | ✅ Pass | 1.316s |  |
| Function Response Memory Test | ✅ Pass | 1.072s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.321s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.255s |  |
| Penetration Testing Methodology | ✅ Pass | 1.210s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.539s |  |
| SQL Injection Attack Type | ✅ Pass | 1.214s |  |
| Penetration Testing Framework | ✅ Pass | 0.828s |  |
| Web Application Security Scanner | ✅ Pass | 1.242s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.221s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.260s

---

### pentester (gpt-oss:120b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.922s |  |
| Text Transform Uppercase | ✅ Pass | 1.372s |  |
| Count from 1 to 5 | ✅ Pass | 1.114s |  |
| Math Calculation | ✅ Pass | 0.914s |  |
| Basic Echo Function | ✅ Pass | 1.084s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.968s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.109s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.018s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.243s |  |
| Search Query Function | ✅ Pass | 1.003s |  |
| Ask Advice Function | ✅ Pass | 0.941s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.879s |  |
| Basic Context Memory Test | ✅ Pass | 1.544s |  |
| Function Argument Memory Test | ✅ Pass | 1.132s |  |
| Function Response Memory Test | ✅ Pass | 0.991s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.949s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.581s |  |
| Penetration Testing Methodology | ✅ Pass | 1.299s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.353s |  |
| SQL Injection Attack Type | ✅ Pass | 1.185s |  |
| Penetration Testing Framework | ✅ Pass | 1.473s |  |
| Web Application Security Scanner | ✅ Pass | 1.224s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.355s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.203s

---

