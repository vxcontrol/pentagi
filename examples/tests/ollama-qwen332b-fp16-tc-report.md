# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 21:18:34 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 7.029s |
| simple_json | qwen3:32b-fp16-tc | true | 5/5 (100.00%) | 6.073s |
| primary_agent | qwen3:32b-fp16-tc | true | 22/23 (95.65%) | 6.596s |
| assistant | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 7.374s |
| generator | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 6.395s |
| refiner | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 7.367s |
| adviser | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 7.065s |
| reflector | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 6.974s |
| searcher | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 6.736s |
| enricher | qwen3:32b-fp16-tc | true | 22/23 (95.65%) | 6.578s |
| coder | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 7.086s |
| installer | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 6.952s |
| pentester | qwen3:32b-fp16-tc | true | 23/23 (100.00%) | 7.140s |

**Total**: 279/281 (99.29%) successful tests
**Overall average latency**: 6.925s

## Detailed Results

### simple (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 8.338s |  |
| Text Transform Uppercase | ✅ Pass | 7.296s |  |
| Count from 1 to 5 | ✅ Pass | 4.953s |  |
| Math Calculation | ✅ Pass | 4.497s |  |
| Basic Echo Function | ✅ Pass | 3.309s |  |
| Streaming Simple Math Streaming | ✅ Pass | 9.747s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.707s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.416s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.116s |  |
| Search Query Function | ✅ Pass | 3.968s |  |
| Ask Advice Function | ✅ Pass | 4.378s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.260s |  |
| Basic Context Memory Test | ✅ Pass | 5.115s |  |
| Function Argument Memory Test | ✅ Pass | 6.694s |  |
| Function Response Memory Test | ✅ Pass | 4.319s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 14.800s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.892s |  |
| Penetration Testing Methodology | ✅ Pass | 9.962s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.119s |  |
| SQL Injection Attack Type | ✅ Pass | 7.969s |  |
| Penetration Testing Framework | ✅ Pass | 11.575s |  |
| Web Application Security Scanner | ✅ Pass | 14.461s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.754s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.029s

---

### simple_json (qwen3:32b-fp16-tc)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 5.983s |  |
| Person Information JSON | ✅ Pass | 4.989s |  |
| Project Information JSON | ✅ Pass | 6.805s |  |
| User Profile JSON | ✅ Pass | 6.519s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 6.068s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 6.073s

---

### primary_agent (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.310s |  |
| Text Transform Uppercase | ✅ Pass | 5.831s |  |
| Count from 1 to 5 | ✅ Pass | 5.415s |  |
| Math Calculation | ✅ Pass | 6.598s |  |
| Basic Echo Function | ✅ Pass | 3.385s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.320s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.366s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.270s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.612s |  |
| Search Query Function | ✅ Pass | 4.001s |  |
| Ask Advice Function | ✅ Pass | 4.534s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.845s |  |
| Basic Context Memory Test | ✅ Pass | 5.125s |  |
| Function Argument Memory Test | ✅ Pass | 4.316s |  |
| Function Response Memory Test | ✅ Pass | 3.577s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 11.379s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.505s |  |
| Penetration Testing Methodology | ✅ Pass | 11.729s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.465s |  |
| SQL Injection Attack Type | ✅ Pass | 7.851s |  |
| Penetration Testing Framework | ✅ Pass | 11.415s |  |
| Web Application Security Scanner | ✅ Pass | 12.780s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.079s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 6.596s

---

### assistant (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.196s |  |
| Text Transform Uppercase | ✅ Pass | 5.213s |  |
| Count from 1 to 5 | ✅ Pass | 3.672s |  |
| Math Calculation | ✅ Pass | 5.501s |  |
| Basic Echo Function | ✅ Pass | 3.435s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.058s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.833s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.393s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.098s |  |
| Search Query Function | ✅ Pass | 4.025s |  |
| Ask Advice Function | ✅ Pass | 5.241s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.946s |  |
| Basic Context Memory Test | ✅ Pass | 4.055s |  |
| Function Argument Memory Test | ✅ Pass | 7.927s |  |
| Function Response Memory Test | ✅ Pass | 21.505s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 10.776s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.533s |  |
| Penetration Testing Methodology | ✅ Pass | 11.291s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.372s |  |
| SQL Injection Attack Type | ✅ Pass | 10.011s |  |
| Penetration Testing Framework | ✅ Pass | 16.996s |  |
| Web Application Security Scanner | ✅ Pass | 6.533s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.978s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.374s

---

### generator (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 8.272s |  |
| Text Transform Uppercase | ✅ Pass | 6.269s |  |
| Count from 1 to 5 | ✅ Pass | 5.975s |  |
| Math Calculation | ✅ Pass | 5.078s |  |
| Basic Echo Function | ✅ Pass | 3.326s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.757s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.235s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.513s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.065s |  |
| Search Query Function | ✅ Pass | 2.729s |  |
| Ask Advice Function | ✅ Pass | 4.952s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.660s |  |
| Basic Context Memory Test | ✅ Pass | 4.273s |  |
| Function Argument Memory Test | ✅ Pass | 4.981s |  |
| Function Response Memory Test | ✅ Pass | 6.514s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.339s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.602s |  |
| Penetration Testing Methodology | ✅ Pass | 9.196s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.506s |  |
| SQL Injection Attack Type | ✅ Pass | 6.542s |  |
| Penetration Testing Framework | ✅ Pass | 11.258s |  |
| Web Application Security Scanner | ✅ Pass | 10.277s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.751s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.395s

---

### refiner (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 9.163s |  |
| Text Transform Uppercase | ✅ Pass | 6.860s |  |
| Count from 1 to 5 | ✅ Pass | 5.760s |  |
| Math Calculation | ✅ Pass | 6.596s |  |
| Basic Echo Function | ✅ Pass | 3.326s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.044s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.567s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.097s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.782s |  |
| Search Query Function | ✅ Pass | 2.999s |  |
| Ask Advice Function | ✅ Pass | 4.545s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.240s |  |
| Basic Context Memory Test | ✅ Pass | 3.805s |  |
| Function Argument Memory Test | ✅ Pass | 13.018s |  |
| Function Response Memory Test | ✅ Pass | 13.484s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 18.941s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.375s |  |
| Penetration Testing Methodology | ✅ Pass | 9.131s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.578s |  |
| SQL Injection Attack Type | ✅ Pass | 6.729s |  |
| Penetration Testing Framework | ✅ Pass | 9.926s |  |
| Web Application Security Scanner | ✅ Pass | 7.386s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.078s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.367s

---

### adviser (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.319s |  |
| Text Transform Uppercase | ✅ Pass | 4.659s |  |
| Count from 1 to 5 | ✅ Pass | 7.788s |  |
| Math Calculation | ✅ Pass | 5.550s |  |
| Basic Echo Function | ✅ Pass | 3.435s |  |
| Streaming Simple Math Streaming | ✅ Pass | 6.069s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.911s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.180s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.828s |  |
| Search Query Function | ✅ Pass | 2.529s |  |
| Ask Advice Function | ✅ Pass | 4.512s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.623s |  |
| Basic Context Memory Test | ✅ Pass | 3.887s |  |
| Function Argument Memory Test | ✅ Pass | 6.185s |  |
| Function Response Memory Test | ✅ Pass | 7.947s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 11.167s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.547s |  |
| Penetration Testing Methodology | ✅ Pass | 10.852s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.335s |  |
| SQL Injection Attack Type | ✅ Pass | 7.411s |  |
| Penetration Testing Framework | ✅ Pass | 12.124s |  |
| Web Application Security Scanner | ✅ Pass | 11.661s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.973s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.065s

---

### reflector (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 11.883s |  |
| Text Transform Uppercase | ✅ Pass | 4.865s |  |
| Count from 1 to 5 | ✅ Pass | 8.229s |  |
| Math Calculation | ✅ Pass | 5.889s |  |
| Basic Echo Function | ✅ Pass | 4.971s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.694s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.196s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.101s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.744s |  |
| Search Query Function | ✅ Pass | 4.014s |  |
| Ask Advice Function | ✅ Pass | 4.346s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.364s |  |
| Basic Context Memory Test | ✅ Pass | 4.967s |  |
| Function Argument Memory Test | ✅ Pass | 7.148s |  |
| Function Response Memory Test | ✅ Pass | 8.042s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 10.223s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.460s |  |
| Penetration Testing Methodology | ✅ Pass | 9.655s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.905s |  |
| SQL Injection Attack Type | ✅ Pass | 5.332s |  |
| Penetration Testing Framework | ✅ Pass | 13.050s |  |
| Web Application Security Scanner | ✅ Pass | 11.131s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.172s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.974s

---

### searcher (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.555s |  |
| Text Transform Uppercase | ✅ Pass | 4.951s |  |
| Count from 1 to 5 | ✅ Pass | 4.161s |  |
| Math Calculation | ✅ Pass | 3.418s |  |
| Basic Echo Function | ✅ Pass | 2.865s |  |
| Streaming Simple Math Streaming | ✅ Pass | 10.309s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.476s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.271s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.118s |  |
| Search Query Function | ✅ Pass | 5.118s |  |
| Ask Advice Function | ✅ Pass | 5.088s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.300s |  |
| Basic Context Memory Test | ✅ Pass | 4.086s |  |
| Function Argument Memory Test | ✅ Pass | 3.538s |  |
| Function Response Memory Test | ✅ Pass | 4.366s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 10.349s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.474s |  |
| Penetration Testing Methodology | ✅ Pass | 8.283s |  |
| Vulnerability Assessment Tools | ✅ Pass | 24.191s |  |
| SQL Injection Attack Type | ✅ Pass | 7.553s |  |
| Penetration Testing Framework | ✅ Pass | 12.626s |  |
| Web Application Security Scanner | ✅ Pass | 12.450s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.361s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.736s

---

### enricher (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.379s |  |
| Text Transform Uppercase | ✅ Pass | 4.917s |  |
| Count from 1 to 5 | ✅ Pass | 4.756s |  |
| Math Calculation | ✅ Pass | 4.864s |  |
| Basic Echo Function | ✅ Pass | 3.522s |  |
| Streaming Simple Math Streaming | ✅ Pass | 7.023s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.537s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.523s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.829s |  |
| Search Query Function | ✅ Pass | 4.096s |  |
| Ask Advice Function | ✅ Pass | 5.520s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.941s |  |
| Basic Context Memory Test | ✅ Pass | 5.351s |  |
| Function Argument Memory Test | ✅ Pass | 4.132s |  |
| Function Response Memory Test | ✅ Pass | 4.927s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 12.571s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.384s |  |
| Penetration Testing Methodology | ✅ Pass | 16.065s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.939s |  |
| SQL Injection Attack Type | ✅ Pass | 7.665s |  |
| Penetration Testing Framework | ✅ Pass | 15.214s |  |
| Web Application Security Scanner | ✅ Pass | 10.771s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.347s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 6.578s

---

### coder (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.090s |  |
| Text Transform Uppercase | ✅ Pass | 5.618s |  |
| Count from 1 to 5 | ✅ Pass | 5.186s |  |
| Math Calculation | ✅ Pass | 7.975s |  |
| Basic Echo Function | ✅ Pass | 3.275s |  |
| Streaming Simple Math Streaming | ✅ Pass | 11.679s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.395s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.268s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.099s |  |
| Search Query Function | ✅ Pass | 4.003s |  |
| Ask Advice Function | ✅ Pass | 4.074s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.796s |  |
| Basic Context Memory Test | ✅ Pass | 4.440s |  |
| Function Argument Memory Test | ✅ Pass | 15.400s |  |
| Function Response Memory Test | ✅ Pass | 9.491s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.311s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.199s |  |
| Penetration Testing Methodology | ✅ Pass | 9.748s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.082s |  |
| SQL Injection Attack Type | ✅ Pass | 6.824s |  |
| Penetration Testing Framework | ✅ Pass | 10.664s |  |
| Web Application Security Scanner | ✅ Pass | 12.258s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.093s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.086s

---

### installer (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.516s |  |
| Text Transform Uppercase | ✅ Pass | 7.313s |  |
| Count from 1 to 5 | ✅ Pass | 6.568s |  |
| Math Calculation | ✅ Pass | 7.159s |  |
| Basic Echo Function | ✅ Pass | 3.013s |  |
| Streaming Simple Math Streaming | ✅ Pass | 10.104s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.982s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.514s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.809s |  |
| Search Query Function | ✅ Pass | 4.973s |  |
| Ask Advice Function | ✅ Pass | 5.545s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.130s |  |
| Basic Context Memory Test | ✅ Pass | 4.978s |  |
| Function Argument Memory Test | ✅ Pass | 5.363s |  |
| Function Response Memory Test | ✅ Pass | 7.220s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 11.346s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.498s |  |
| Penetration Testing Methodology | ✅ Pass | 8.142s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.207s |  |
| SQL Injection Attack Type | ✅ Pass | 9.205s |  |
| Penetration Testing Framework | ✅ Pass | 11.698s |  |
| Web Application Security Scanner | ✅ Pass | 11.854s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.745s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.952s

---

### pentester (qwen3:32b-fp16-tc)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.421s |  |
| Text Transform Uppercase | ✅ Pass | 5.115s |  |
| Count from 1 to 5 | ✅ Pass | 7.193s |  |
| Math Calculation | ✅ Pass | 3.295s |  |
| Basic Echo Function | ✅ Pass | 2.843s |  |
| Streaming Simple Math Streaming | ✅ Pass | 8.829s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.051s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.529s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.107s |  |
| Search Query Function | ✅ Pass | 4.109s |  |
| Ask Advice Function | ✅ Pass | 6.434s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.367s |  |
| Basic Context Memory Test | ✅ Pass | 3.979s |  |
| Function Argument Memory Test | ✅ Pass | 5.832s |  |
| Function Response Memory Test | ✅ Pass | 22.963s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.087s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.515s |  |
| Penetration Testing Methodology | ✅ Pass | 10.706s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.930s |  |
| SQL Injection Attack Type | ✅ Pass | 8.345s |  |
| Penetration Testing Framework | ✅ Pass | 11.880s |  |
| Web Application Security Scanner | ✅ Pass | 8.811s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.864s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 7.140s

---

