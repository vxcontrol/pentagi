# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 17:47:16 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek-chat | false | 23/23 (100.00%) | 5.311s |
| simple_json | deepseek-chat | false | 5/5 (100.00%) | 6.055s |
| primary_agent | deepseek-chat | false | 23/23 (100.00%) | 5.368s |
| assistant | deepseek-chat | false | 23/23 (100.00%) | 5.424s |
| generator | deepseek-chat | false | 23/23 (100.00%) | 5.229s |
| refiner | deepseek-chat | false | 23/23 (100.00%) | 5.565s |
| adviser | deepseek-chat | false | 23/23 (100.00%) | 5.247s |
| reflector | deepseek-chat | false | 23/23 (100.00%) | 5.271s |
| searcher | deepseek-chat | false | 23/23 (100.00%) | 5.625s |
| enricher | deepseek-chat | false | 23/23 (100.00%) | 5.513s |
| coder | deepseek-coder | false | 23/23 (100.00%) | 5.497s |
| installer | deepseek-coder | false | 23/23 (100.00%) | 5.568s |
| pentester | deepseek-chat | false | 23/23 (100.00%) | 5.420s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 5.431s

## Detailed Results

### simple (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.764s |  |
| Text Transform Uppercase | ✅ Pass | 4.524s |  |
| Count from 1 to 5 | ✅ Pass | 3.821s |  |
| Math Calculation | ✅ Pass | 2.964s |  |
| Basic Echo Function | ✅ Pass | 3.469s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.634s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.504s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.903s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.959s |  |
| Search Query Function | ✅ Pass | 4.433s |  |
| Ask Advice Function | ✅ Pass | 5.384s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 6.624s |  |
| Basic Context Memory Test | ✅ Pass | 5.520s |  |
| Function Argument Memory Test | ✅ Pass | 3.897s |  |
| Function Response Memory Test | ✅ Pass | 4.031s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.247s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.050s |  |
| Penetration Testing Methodology | ✅ Pass | 9.651s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.330s |  |
| SQL Injection Attack Type | ✅ Pass | 5.769s |  |
| Penetration Testing Framework | ✅ Pass | 7.152s |  |
| Web Application Security Scanner | ✅ Pass | 6.556s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.952s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.311s

---

### simple_json (deepseek-chat)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 4.846s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 4.705s |  |
| User Profile JSON | ✅ Pass | 4.834s |  |
| Project Information JSON | ✅ Pass | 5.575s |  |
| Vulnerability Report Memory Test | ✅ Pass | 10.313s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 6.055s

---

### primary_agent (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.672s |  |
| Text Transform Uppercase | ✅ Pass | 4.375s |  |
| Count from 1 to 5 | ✅ Pass | 4.845s |  |
| Math Calculation | ✅ Pass | 3.273s |  |
| Basic Echo Function | ✅ Pass | 2.789s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.038s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.051s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.513s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.225s |  |
| Search Query Function | ✅ Pass | 5.174s |  |
| Ask Advice Function | ✅ Pass | 5.114s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.113s |  |
| Basic Context Memory Test | ✅ Pass | 6.212s |  |
| Function Argument Memory Test | ✅ Pass | 3.835s |  |
| Function Response Memory Test | ✅ Pass | 3.563s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.881s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.325s |  |
| Penetration Testing Methodology | ✅ Pass | 9.404s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.911s |  |
| SQL Injection Attack Type | ✅ Pass | 5.832s |  |
| Penetration Testing Framework | ✅ Pass | 6.716s |  |
| Web Application Security Scanner | ✅ Pass | 6.720s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.862s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.368s

---

### assistant (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.686s |  |
| Text Transform Uppercase | ✅ Pass | 3.965s |  |
| Count from 1 to 5 | ✅ Pass | 4.470s |  |
| Math Calculation | ✅ Pass | 3.344s |  |
| Basic Echo Function | ✅ Pass | 2.617s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.421s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.428s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.557s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.913s |  |
| Search Query Function | ✅ Pass | 4.572s |  |
| Ask Advice Function | ✅ Pass | 5.513s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.304s |  |
| Basic Context Memory Test | ✅ Pass | 5.469s |  |
| Function Argument Memory Test | ✅ Pass | 4.242s |  |
| Function Response Memory Test | ✅ Pass | 3.458s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.731s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.415s |  |
| Penetration Testing Methodology | ✅ Pass | 9.544s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.469s |  |
| SQL Injection Attack Type | ✅ Pass | 6.214s |  |
| Penetration Testing Framework | ✅ Pass | 7.617s |  |
| Web Application Security Scanner | ✅ Pass | 8.104s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.687s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.424s

---

### generator (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.717s |  |
| Text Transform Uppercase | ✅ Pass | 3.617s |  |
| Count from 1 to 5 | ✅ Pass | 4.246s |  |
| Math Calculation | ✅ Pass | 3.653s |  |
| Basic Echo Function | ✅ Pass | 3.080s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.909s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.088s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.582s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.792s |  |
| Search Query Function | ✅ Pass | 4.243s |  |
| Ask Advice Function | ✅ Pass | 5.467s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.125s |  |
| Basic Context Memory Test | ✅ Pass | 5.452s |  |
| Function Argument Memory Test | ✅ Pass | 3.968s |  |
| Function Response Memory Test | ✅ Pass | 3.586s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.853s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.614s |  |
| Penetration Testing Methodology | ✅ Pass | 9.258s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.043s |  |
| SQL Injection Attack Type | ✅ Pass | 5.662s |  |
| Penetration Testing Framework | ✅ Pass | 7.671s |  |
| Web Application Security Scanner | ✅ Pass | 7.155s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.480s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.229s

---

### refiner (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.333s |  |
| Text Transform Uppercase | ✅ Pass | 4.448s |  |
| Count from 1 to 5 | ✅ Pass | 3.710s |  |
| Math Calculation | ✅ Pass | 4.147s |  |
| Basic Echo Function | ✅ Pass | 3.304s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.781s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.982s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.683s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.776s |  |
| Search Query Function | ✅ Pass | 5.087s |  |
| Ask Advice Function | ✅ Pass | 5.447s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.247s |  |
| Basic Context Memory Test | ✅ Pass | 8.677s |  |
| Function Argument Memory Test | ✅ Pass | 4.532s |  |
| Function Response Memory Test | ✅ Pass | 4.411s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.120s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.760s |  |
| Penetration Testing Methodology | ✅ Pass | 8.653s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.457s |  |
| SQL Injection Attack Type | ✅ Pass | 4.741s |  |
| Penetration Testing Framework | ✅ Pass | 6.270s |  |
| Web Application Security Scanner | ✅ Pass | 6.718s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.708s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.565s

---

### adviser (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.991s |  |
| Text Transform Uppercase | ✅ Pass | 4.465s |  |
| Count from 1 to 5 | ✅ Pass | 3.697s |  |
| Math Calculation | ✅ Pass | 2.911s |  |
| Basic Echo Function | ✅ Pass | 3.647s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.682s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.747s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.787s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.283s |  |
| Search Query Function | ✅ Pass | 4.959s |  |
| Ask Advice Function | ✅ Pass | 5.555s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.493s |  |
| Basic Context Memory Test | ✅ Pass | 6.249s |  |
| Function Argument Memory Test | ✅ Pass | 3.804s |  |
| Function Response Memory Test | ✅ Pass | 3.557s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.113s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.564s |  |
| Penetration Testing Methodology | ✅ Pass | 8.741s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.528s |  |
| SQL Injection Attack Type | ✅ Pass | 4.858s |  |
| Penetration Testing Framework | ✅ Pass | 7.483s |  |
| Web Application Security Scanner | ✅ Pass | 6.759s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.798s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.247s

---

### reflector (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.482s |  |
| Text Transform Uppercase | ✅ Pass | 3.810s |  |
| Count from 1 to 5 | ✅ Pass | 3.679s |  |
| Math Calculation | ✅ Pass | 4.288s |  |
| Basic Echo Function | ✅ Pass | 2.900s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.166s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.882s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.877s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.838s |  |
| Search Query Function | ✅ Pass | 5.070s |  |
| Ask Advice Function | ✅ Pass | 5.312s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.392s |  |
| Basic Context Memory Test | ✅ Pass | 6.041s |  |
| Function Argument Memory Test | ✅ Pass | 3.810s |  |
| Function Response Memory Test | ✅ Pass | 3.559s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.409s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.099s |  |
| Penetration Testing Methodology | ✅ Pass | 9.142s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.554s |  |
| SQL Injection Attack Type | ✅ Pass | 5.351s |  |
| Penetration Testing Framework | ✅ Pass | 6.624s |  |
| Web Application Security Scanner | ✅ Pass | 6.054s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.888s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.271s

---

### searcher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.730s |  |
| Text Transform Uppercase | ✅ Pass | 4.388s |  |
| Count from 1 to 5 | ✅ Pass | 4.551s |  |
| Math Calculation | ✅ Pass | 4.270s |  |
| Basic Echo Function | ✅ Pass | 4.573s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.576s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.010s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.903s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.152s |  |
| Search Query Function | ✅ Pass | 4.998s |  |
| Ask Advice Function | ✅ Pass | 5.997s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.507s |  |
| Basic Context Memory Test | ✅ Pass | 5.518s |  |
| Function Argument Memory Test | ✅ Pass | 5.408s |  |
| Function Response Memory Test | ✅ Pass | 4.043s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.466s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.748s |  |
| Penetration Testing Methodology | ✅ Pass | 9.102s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.053s |  |
| SQL Injection Attack Type | ✅ Pass | 5.038s |  |
| Penetration Testing Framework | ✅ Pass | 7.892s |  |
| Web Application Security Scanner | ✅ Pass | 6.318s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.126s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.625s

---

### enricher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.540s |  |
| Text Transform Uppercase | ✅ Pass | 3.596s |  |
| Count from 1 to 5 | ✅ Pass | 2.689s |  |
| Math Calculation | ✅ Pass | 3.260s |  |
| Basic Echo Function | ✅ Pass | 4.973s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.659s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.236s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 9.764s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.210s |  |
| Search Query Function | ✅ Pass | 6.859s |  |
| Ask Advice Function | ✅ Pass | 5.428s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.998s |  |
| Basic Context Memory Test | ✅ Pass | 5.319s |  |
| Function Argument Memory Test | ✅ Pass | 4.266s |  |
| Function Response Memory Test | ✅ Pass | 3.874s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.045s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.834s |  |
| Penetration Testing Methodology | ✅ Pass | 7.870s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.213s |  |
| SQL Injection Attack Type | ✅ Pass | 5.555s |  |
| Penetration Testing Framework | ✅ Pass | 6.570s |  |
| Web Application Security Scanner | ✅ Pass | 8.353s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.676s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.513s

---

### coder (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.420s |  |
| Text Transform Uppercase | ✅ Pass | 3.809s |  |
| Count from 1 to 5 | ✅ Pass | 3.425s |  |
| Math Calculation | ✅ Pass | 3.576s |  |
| Basic Echo Function | ✅ Pass | 5.434s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.735s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.722s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 6.069s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.113s |  |
| Search Query Function | ✅ Pass | 5.656s |  |
| Ask Advice Function | ✅ Pass | 5.661s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.103s |  |
| Basic Context Memory Test | ✅ Pass | 5.785s |  |
| Function Argument Memory Test | ✅ Pass | 5.092s |  |
| Function Response Memory Test | ✅ Pass | 3.643s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.078s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.726s |  |
| Penetration Testing Methodology | ✅ Pass | 9.194s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.973s |  |
| SQL Injection Attack Type | ✅ Pass | 5.889s |  |
| Penetration Testing Framework | ✅ Pass | 7.044s |  |
| Web Application Security Scanner | ✅ Pass | 6.330s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.951s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.497s

---

### installer (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.464s |  |
| Text Transform Uppercase | ✅ Pass | 3.784s |  |
| Count from 1 to 5 | ✅ Pass | 4.019s |  |
| Math Calculation | ✅ Pass | 2.930s |  |
| Basic Echo Function | ✅ Pass | 4.951s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.717s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.847s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 7.403s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.607s |  |
| Search Query Function | ✅ Pass | 4.932s |  |
| Ask Advice Function | ✅ Pass | 5.893s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.840s |  |
| Basic Context Memory Test | ✅ Pass | 5.465s |  |
| Function Argument Memory Test | ✅ Pass | 4.037s |  |
| Function Response Memory Test | ✅ Pass | 4.093s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.895s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.344s |  |
| Penetration Testing Methodology | ✅ Pass | 9.335s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.233s |  |
| SQL Injection Attack Type | ✅ Pass | 4.977s |  |
| Penetration Testing Framework | ✅ Pass | 5.985s |  |
| Web Application Security Scanner | ✅ Pass | 10.461s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.842s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.568s

---

### pentester (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.366s |  |
| Text Transform Uppercase | ✅ Pass | 3.691s |  |
| Count from 1 to 5 | ✅ Pass | 3.832s |  |
| Math Calculation | ✅ Pass | 3.527s |  |
| Basic Echo Function | ✅ Pass | 4.248s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.318s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.980s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.578s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.113s |  |
| Search Query Function | ✅ Pass | 4.782s |  |
| Ask Advice Function | ✅ Pass | 5.374s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.084s |  |
| Basic Context Memory Test | ✅ Pass | 5.872s |  |
| Function Argument Memory Test | ✅ Pass | 3.726s |  |
| Function Response Memory Test | ✅ Pass | 3.663s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.390s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.365s |  |
| Penetration Testing Methodology | ✅ Pass | 9.511s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.981s |  |
| SQL Injection Attack Type | ✅ Pass | 5.853s |  |
| Penetration Testing Framework | ✅ Pass | 5.901s |  |
| Web Application Security Scanner | ✅ Pass | 6.368s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.120s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.420s

---

