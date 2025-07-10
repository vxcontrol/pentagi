# LLM Agent Testing Report

Generated: Tue, 08 Jul 2025 21:37:50 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-4.1-mini | false | 18/18 (100.00%) | 0.908s |
| simple_json | gpt-4.1-mini | false | 4/4 (100.00%) | 1.254s |
| primary_agent | o3-mini | false | 18/18 (100.00%) | 2.301s |
| assistant | o3-mini | true | 18/18 (100.00%) | 2.889s |
| generator | o3-mini | true | 18/18 (100.00%) | 3.292s |
| refiner | gpt-4.1 | false | 18/18 (100.00%) | 0.946s |
| adviser | o3-mini | true | 18/18 (100.00%) | 2.925s |
| reflector | o3-mini | true | 18/18 (100.00%) | 3.114s |
| searcher | gpt-4.1-mini | false | 18/18 (100.00%) | 0.987s |
| enricher | gpt-4.1-mini | false | 18/18 (100.00%) | 0.918s |
| coder | gpt-4.1 | false | 18/18 (100.00%) | 1.037s |
| installer | gpt-4.1 | false | 18/18 (100.00%) | 0.895s |
| pentester | o3-mini | true | 18/18 (100.00%) | 2.058s |

**Total**: 220/220 (100.00%) successful tests
**Overall average latency**: 1.845s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.273s |  |
| Text Transform Uppercase | ✅ Pass | 0.886s |  |
| Count from 1 to 5 | ✅ Pass | 0.687s |  |
| Math Calculation | ✅ Pass | 0.563s |  |
| Basic Echo Function | ✅ Pass | 0.835s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.524s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.646s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.762s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.856s |  |
| Search Query Function | ✅ Pass | 1.236s |  |
| Ask Advice Function | ✅ Pass | 0.912s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.028s |  |
| Penetration Testing Methodology | ✅ Pass | 0.863s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.509s |  |
| SQL Injection Attack Type | ✅ Pass | 0.742s |  |
| Penetration Testing Framework | ✅ Pass | 1.130s |  |
| Web Application Security Scanner | ✅ Pass | 0.759s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.124s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.908s

---

### simple_json (gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 1.104s |  |
| User Profile JSON | ✅ Pass | 1.118s |  |
| Person Information JSON | ✅ Pass | 1.728s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 1.066s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 1.254s

---

### primary_agent (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.122s |  |
| Text Transform Uppercase | ✅ Pass | 2.739s |  |
| Count from 1 to 5 | ✅ Pass | 1.846s |  |
| Math Calculation | ✅ Pass | 1.468s |  |
| Basic Echo Function | ✅ Pass | 2.384s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.161s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.643s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.642s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.710s |  |
| Search Query Function | ✅ Pass | 1.354s |  |
| Ask Advice Function | ✅ Pass | 1.849s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.185s |  |
| Penetration Testing Methodology | ✅ Pass | 2.619s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.820s |  |
| SQL Injection Attack Type | ✅ Pass | 3.860s |  |
| Penetration Testing Framework | ✅ Pass | 2.999s |  |
| Web Application Security Scanner | ✅ Pass | 2.217s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.794s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.301s

---

### assistant (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.721s |  |
| Text Transform Uppercase | ✅ Pass | 2.801s |  |
| Count from 1 to 5 | ✅ Pass | 3.280s |  |
| Math Calculation | ✅ Pass | 2.702s |  |
| Basic Echo Function | ✅ Pass | 2.037s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.070s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.003s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.796s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.705s |  |
| Search Query Function | ✅ Pass | 1.963s |  |
| Ask Advice Function | ✅ Pass | 3.898s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.667s |  |
| Penetration Testing Methodology | ✅ Pass | 4.137s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.863s |  |
| SQL Injection Attack Type | ✅ Pass | 4.000s |  |
| Penetration Testing Framework | ✅ Pass | 4.767s |  |
| Web Application Security Scanner | ✅ Pass | 3.930s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.653s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.889s

---

### generator (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.825s |  |
| Text Transform Uppercase | ✅ Pass | 2.870s |  |
| Count from 1 to 5 | ✅ Pass | 3.381s |  |
| Math Calculation | ✅ Pass | 2.670s |  |
| Basic Echo Function | ✅ Pass | 2.788s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.463s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.603s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.499s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.586s |  |
| Search Query Function | ✅ Pass | 2.696s |  |
| Ask Advice Function | ✅ Pass | 2.697s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.963s |  |
| Penetration Testing Methodology | ✅ Pass | 3.747s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.585s |  |
| SQL Injection Attack Type | ✅ Pass | 6.148s |  |
| Penetration Testing Framework | ✅ Pass | 3.968s |  |
| Web Application Security Scanner | ✅ Pass | 4.753s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.012s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.292s

---

### refiner (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.478s |  |
| Text Transform Uppercase | ✅ Pass | 0.697s |  |
| Count from 1 to 5 | ✅ Pass | 0.710s |  |
| Math Calculation | ✅ Pass | 0.613s |  |
| Basic Echo Function | ✅ Pass | 0.647s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.730s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.600s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.641s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.655s |  |
| Search Query Function | ✅ Pass | 1.132s |  |
| Ask Advice Function | ✅ Pass | 0.749s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.825s |  |
| Penetration Testing Methodology | ✅ Pass | 1.662s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.748s |  |
| SQL Injection Attack Type | ✅ Pass | 1.005s |  |
| Penetration Testing Framework | ✅ Pass | 0.922s |  |
| Web Application Security Scanner | ✅ Pass | 1.126s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.084s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.946s

---

### adviser (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.036s |  |
| Text Transform Uppercase | ✅ Pass | 2.555s |  |
| Count from 1 to 5 | ✅ Pass | 4.454s |  |
| Math Calculation | ✅ Pass | 1.452s |  |
| Basic Echo Function | ✅ Pass | 1.984s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.569s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.151s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.052s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.143s |  |
| Search Query Function | ✅ Pass | 3.494s |  |
| Ask Advice Function | ✅ Pass | 1.943s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.486s |  |
| Penetration Testing Methodology | ✅ Pass | 4.350s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.507s |  |
| SQL Injection Attack Type | ✅ Pass | 3.797s |  |
| Penetration Testing Framework | ✅ Pass | 5.259s |  |
| Web Application Security Scanner | ✅ Pass | 3.569s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.846s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.925s

---

### reflector (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.250s |  |
| Text Transform Uppercase | ✅ Pass | 2.648s |  |
| Count from 1 to 5 | ✅ Pass | 2.641s |  |
| Math Calculation | ✅ Pass | 2.044s |  |
| Basic Echo Function | ✅ Pass | 1.705s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.573s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.510s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.690s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.873s |  |
| Search Query Function | ✅ Pass | 1.982s |  |
| Ask Advice Function | ✅ Pass | 2.455s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.684s |  |
| Penetration Testing Methodology | ✅ Pass | 3.920s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.856s |  |
| SQL Injection Attack Type | ✅ Pass | 4.715s |  |
| Penetration Testing Framework | ✅ Pass | 6.356s |  |
| Web Application Security Scanner | ✅ Pass | 3.150s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.991s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.114s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.619s |  |
| Text Transform Uppercase | ✅ Pass | 0.695s |  |
| Count from 1 to 5 | ✅ Pass | 2.320s |  |
| Math Calculation | ✅ Pass | 0.649s |  |
| Basic Echo Function | ✅ Pass | 0.743s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.548s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.758s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.957s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.863s |  |
| Search Query Function | ✅ Pass | 1.296s |  |
| Ask Advice Function | ✅ Pass | 0.958s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.949s |  |
| Penetration Testing Methodology | ✅ Pass | 1.018s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.660s |  |
| SQL Injection Attack Type | ✅ Pass | 0.659s |  |
| Penetration Testing Framework | ✅ Pass | 1.245s |  |
| Web Application Security Scanner | ✅ Pass | 0.703s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.115s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.987s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.579s |  |
| Text Transform Uppercase | ✅ Pass | 0.597s |  |
| Count from 1 to 5 | ✅ Pass | 0.764s |  |
| Math Calculation | ✅ Pass | 0.672s |  |
| Basic Echo Function | ✅ Pass | 0.774s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.976s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.555s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.804s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.942s |  |
| Search Query Function | ✅ Pass | 1.126s |  |
| Ask Advice Function | ✅ Pass | 1.137s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.980s |  |
| Penetration Testing Methodology | ✅ Pass | 1.285s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.403s |  |
| SQL Injection Attack Type | ✅ Pass | 0.644s |  |
| Penetration Testing Framework | ✅ Pass | 1.558s |  |
| Web Application Security Scanner | ✅ Pass | 0.737s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.986s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.918s

---

### coder (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.674s |  |
| Text Transform Uppercase | ✅ Pass | 0.532s |  |
| Count from 1 to 5 | ✅ Pass | 0.812s |  |
| Math Calculation | ✅ Pass | 0.717s |  |
| Basic Echo Function | ✅ Pass | 0.964s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.605s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.781s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.461s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.057s |  |
| Search Query Function | ✅ Pass | 0.900s |  |
| Ask Advice Function | ✅ Pass | 0.858s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.743s |  |
| Penetration Testing Methodology | ✅ Pass | 0.828s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.720s |  |
| SQL Injection Attack Type | ✅ Pass | 1.334s |  |
| Penetration Testing Framework | ✅ Pass | 0.893s |  |
| Web Application Security Scanner | ✅ Pass | 2.926s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.861s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.037s

---

### installer (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.797s |  |
| Text Transform Uppercase | ✅ Pass | 0.581s |  |
| Count from 1 to 5 | ✅ Pass | 0.790s |  |
| Math Calculation | ✅ Pass | 0.672s |  |
| Basic Echo Function | ✅ Pass | 0.840s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.520s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.833s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.775s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.704s |  |
| Search Query Function | ✅ Pass | 0.870s |  |
| Ask Advice Function | ✅ Pass | 0.984s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.811s |  |
| Penetration Testing Methodology | ✅ Pass | 0.909s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.207s |  |
| SQL Injection Attack Type | ✅ Pass | 0.573s |  |
| Penetration Testing Framework | ✅ Pass | 1.137s |  |
| Web Application Security Scanner | ✅ Pass | 1.129s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.976s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.895s

---

### pentester (o3-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.667s |  |
| Text Transform Uppercase | ✅ Pass | 1.721s |  |
| Count from 1 to 5 | ✅ Pass | 3.443s |  |
| Math Calculation | ✅ Pass | 1.991s |  |
| Basic Echo Function | ✅ Pass | 1.893s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.574s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.834s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.627s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.144s |  |
| Search Query Function | ✅ Pass | 1.536s |  |
| Ask Advice Function | ✅ Pass | 1.837s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.336s |  |
| Penetration Testing Methodology | ✅ Pass | 1.977s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.612s |  |
| SQL Injection Attack Type | ✅ Pass | 1.965s |  |
| Penetration Testing Framework | ✅ Pass | 1.597s |  |
| Web Application Security Scanner | ✅ Pass | 3.125s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.162s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.058s

---

