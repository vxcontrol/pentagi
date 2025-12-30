# LLM Agent Testing Report

Generated: Tue, 30 Dec 2025 22:14:32 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-haiku-4-5 | false | 23/23 (100.00%) | 1.339s |
| simple_json | claude-haiku-4-5 | false | 5/5 (100.00%) | 0.758s |
| primary_agent | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.376s |
| assistant | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.323s |
| generator | claude-opus-4-5 | false | 23/23 (100.00%) | 2.899s |
| refiner | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.209s |
| adviser | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.223s |
| reflector | claude-haiku-4-5 | false | 23/23 (100.00%) | 1.220s |
| searcher | claude-haiku-4-5 | false | 23/23 (100.00%) | 1.238s |
| enricher | claude-haiku-4-5 | false | 23/23 (100.00%) | 1.143s |
| coder | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.276s |
| installer | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.240s |
| pentester | claude-sonnet-4-5 | false | 23/23 (100.00%) | 3.314s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 2.534s

## Detailed Results

### simple (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.001s |  |
| Text Transform Uppercase | ✅ Pass | 0.576s |  |
| Count from 1 to 5 | ✅ Pass | 0.686s |  |
| Math Calculation | ✅ Pass | 2.376s |  |
| Basic Echo Function | ✅ Pass | 0.725s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.566s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.851s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.810s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.829s |  |
| Search Query Function | ✅ Pass | 1.311s |  |
| Ask Advice Function | ✅ Pass | 1.074s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.783s |  |
| Basic Context Memory Test | ✅ Pass | 0.998s |  |
| Function Argument Memory Test | ✅ Pass | 0.648s |  |
| Function Response Memory Test | ✅ Pass | 2.339s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.719s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.140s |  |
| Penetration Testing Methodology | ✅ Pass | 2.811s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.837s |  |
| SQL Injection Attack Type | ✅ Pass | 0.776s |  |
| Penetration Testing Framework | ✅ Pass | 2.142s |  |
| Web Application Security Scanner | ✅ Pass | 1.934s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.848s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.339s

---

### simple_json (claude-haiku-4-5)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 0.621s |  |
| Person Information JSON | ✅ Pass | 0.650s |  |
| Vulnerability Report Memory Test | ✅ Pass | 0.971s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.622s |  |
| User Profile JSON | ✅ Pass | 0.922s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.758s

---

### primary_agent (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.558s |  |
| Text Transform Uppercase | ✅ Pass | 1.876s |  |
| Count from 1 to 5 | ✅ Pass | 4.449s |  |
| Math Calculation | ✅ Pass | 2.483s |  |
| Basic Echo Function | ✅ Pass | 2.762s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.078s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.081s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.891s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.868s |  |
| Search Query Function | ✅ Pass | 2.883s |  |
| Ask Advice Function | ✅ Pass | 2.883s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.741s |  |
| Basic Context Memory Test | ✅ Pass | 3.514s |  |
| Function Argument Memory Test | ✅ Pass | 2.625s |  |
| Function Response Memory Test | ✅ Pass | 2.324s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.062s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.608s |  |
| Penetration Testing Methodology | ✅ Pass | 6.620s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.029s |  |
| SQL Injection Attack Type | ✅ Pass | 2.821s |  |
| Penetration Testing Framework | ✅ Pass | 4.992s |  |
| Web Application Security Scanner | ✅ Pass | 4.303s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.189s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.376s

---

### assistant (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.590s |  |
| Text Transform Uppercase | ✅ Pass | 1.986s |  |
| Count from 1 to 5 | ✅ Pass | 3.074s |  |
| Math Calculation | ✅ Pass | 2.546s |  |
| Basic Echo Function | ✅ Pass | 2.864s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.375s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.203s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.058s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.752s |  |
| Search Query Function | ✅ Pass | 3.355s |  |
| Ask Advice Function | ✅ Pass | 2.816s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.136s |  |
| Basic Context Memory Test | ✅ Pass | 3.644s |  |
| Function Argument Memory Test | ✅ Pass | 2.722s |  |
| Function Response Memory Test | ✅ Pass | 2.100s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.201s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.592s |  |
| Penetration Testing Methodology | ✅ Pass | 5.465s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.056s |  |
| SQL Injection Attack Type | ✅ Pass | 3.729s |  |
| Penetration Testing Framework | ✅ Pass | 5.209s |  |
| Web Application Security Scanner | ✅ Pass | 5.148s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.788s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.323s

---

### generator (claude-opus-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.695s |  |
| Text Transform Uppercase | ✅ Pass | 1.981s |  |
| Count from 1 to 5 | ✅ Pass | 2.460s |  |
| Math Calculation | ✅ Pass | 1.179s |  |
| Basic Echo Function | ✅ Pass | 1.559s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.143s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.804s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.890s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.891s |  |
| Search Query Function | ✅ Pass | 2.876s |  |
| Ask Advice Function | ✅ Pass | 2.945s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.717s |  |
| Basic Context Memory Test | ✅ Pass | 2.718s |  |
| Function Argument Memory Test | ✅ Pass | 2.300s |  |
| Function Response Memory Test | ✅ Pass | 2.419s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.930s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.464s |  |
| Penetration Testing Methodology | ✅ Pass | 4.912s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.733s |  |
| SQL Injection Attack Type | ✅ Pass | 1.584s |  |
| Penetration Testing Framework | ✅ Pass | 4.362s |  |
| Web Application Security Scanner | ✅ Pass | 4.495s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.620s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.899s

---

### refiner (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.700s |  |
| Text Transform Uppercase | ✅ Pass | 1.886s |  |
| Count from 1 to 5 | ✅ Pass | 3.384s |  |
| Math Calculation | ✅ Pass | 2.149s |  |
| Basic Echo Function | ✅ Pass | 2.941s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.650s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.801s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.251s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.987s |  |
| Search Query Function | ✅ Pass | 2.855s |  |
| Ask Advice Function | ✅ Pass | 2.645s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.079s |  |
| Basic Context Memory Test | ✅ Pass | 2.470s |  |
| Function Argument Memory Test | ✅ Pass | 2.448s |  |
| Function Response Memory Test | ✅ Pass | 2.222s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.854s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.038s |  |
| Penetration Testing Methodology | ✅ Pass | 5.538s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.097s |  |
| SQL Injection Attack Type | ✅ Pass | 3.137s |  |
| Penetration Testing Framework | ✅ Pass | 4.423s |  |
| Web Application Security Scanner | ✅ Pass | 3.067s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.175s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.209s

---

### adviser (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.948s |  |
| Text Transform Uppercase | ✅ Pass | 1.894s |  |
| Count from 1 to 5 | ✅ Pass | 3.073s |  |
| Math Calculation | ✅ Pass | 2.405s |  |
| Basic Echo Function | ✅ Pass | 2.886s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.090s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.961s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.062s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.087s |  |
| Search Query Function | ✅ Pass | 2.837s |  |
| Ask Advice Function | ✅ Pass | 2.883s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.102s |  |
| Basic Context Memory Test | ✅ Pass | 3.000s |  |
| Function Argument Memory Test | ✅ Pass | 2.411s |  |
| Function Response Memory Test | ✅ Pass | 2.719s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.520s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.498s |  |
| Penetration Testing Methodology | ✅ Pass | 5.616s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.760s |  |
| SQL Injection Attack Type | ✅ Pass | 2.884s |  |
| Penetration Testing Framework | ✅ Pass | 4.880s |  |
| Web Application Security Scanner | ✅ Pass | 3.798s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.801s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.223s

---

### reflector (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.073s |  |
| Text Transform Uppercase | ✅ Pass | 0.982s |  |
| Count from 1 to 5 | ✅ Pass | 0.642s |  |
| Math Calculation | ✅ Pass | 0.574s |  |
| Basic Echo Function | ✅ Pass | 1.530s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.724s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.807s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.849s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.897s |  |
| Search Query Function | ✅ Pass | 0.966s |  |
| Ask Advice Function | ✅ Pass | 0.849s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.969s |  |
| Basic Context Memory Test | ✅ Pass | 1.260s |  |
| Function Argument Memory Test | ✅ Pass | 0.716s |  |
| Function Response Memory Test | ✅ Pass | 0.608s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.244s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.857s |  |
| Penetration Testing Methodology | ✅ Pass | 2.898s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.842s |  |
| SQL Injection Attack Type | ✅ Pass | 0.853s |  |
| Penetration Testing Framework | ✅ Pass | 2.733s |  |
| Web Application Security Scanner | ✅ Pass | 2.249s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.923s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.220s

---

### searcher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.009s |  |
| Text Transform Uppercase | ✅ Pass | 1.748s |  |
| Count from 1 to 5 | ✅ Pass | 0.880s |  |
| Math Calculation | ✅ Pass | 0.646s |  |
| Basic Echo Function | ✅ Pass | 0.825s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.732s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.771s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.968s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.869s |  |
| Search Query Function | ✅ Pass | 1.389s |  |
| Ask Advice Function | ✅ Pass | 1.266s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.735s |  |
| Basic Context Memory Test | ✅ Pass | 1.074s |  |
| Function Argument Memory Test | ✅ Pass | 0.723s |  |
| Function Response Memory Test | ✅ Pass | 0.689s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.194s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.088s |  |
| Penetration Testing Methodology | ✅ Pass | 2.348s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.055s |  |
| SQL Injection Attack Type | ✅ Pass | 0.843s |  |
| Penetration Testing Framework | ✅ Pass | 2.565s |  |
| Web Application Security Scanner | ✅ Pass | 2.136s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.899s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.238s

---

### enricher (claude-haiku-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.694s |  |
| Text Transform Uppercase | ✅ Pass | 0.898s |  |
| Count from 1 to 5 | ✅ Pass | 0.816s |  |
| Math Calculation | ✅ Pass | 0.827s |  |
| Basic Echo Function | ✅ Pass | 0.826s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.717s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.800s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.082s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.887s |  |
| Search Query Function | ✅ Pass | 0.807s |  |
| Ask Advice Function | ✅ Pass | 0.919s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.867s |  |
| Basic Context Memory Test | ✅ Pass | 0.918s |  |
| Function Argument Memory Test | ✅ Pass | 0.711s |  |
| Function Response Memory Test | ✅ Pass | 1.245s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.138s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.906s |  |
| Penetration Testing Methodology | ✅ Pass | 2.450s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.452s |  |
| SQL Injection Attack Type | ✅ Pass | 0.906s |  |
| Penetration Testing Framework | ✅ Pass | 2.619s |  |
| Web Application Security Scanner | ✅ Pass | 1.938s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.857s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.143s

---

### coder (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.614s |  |
| Text Transform Uppercase | ✅ Pass | 2.102s |  |
| Count from 1 to 5 | ✅ Pass | 3.148s |  |
| Math Calculation | ✅ Pass | 3.010s |  |
| Basic Echo Function | ✅ Pass | 3.022s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.150s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.259s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.827s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.951s |  |
| Search Query Function | ✅ Pass | 3.003s |  |
| Ask Advice Function | ✅ Pass | 4.299s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.701s |  |
| Basic Context Memory Test | ✅ Pass | 2.697s |  |
| Function Argument Memory Test | ✅ Pass | 2.800s |  |
| Function Response Memory Test | ✅ Pass | 2.268s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.286s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.559s |  |
| Penetration Testing Methodology | ✅ Pass | 4.948s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.871s |  |
| SQL Injection Attack Type | ✅ Pass | 4.163s |  |
| Penetration Testing Framework | ✅ Pass | 4.058s |  |
| Web Application Security Scanner | ✅ Pass | 3.587s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.022s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.276s

---

### installer (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.442s |  |
| Text Transform Uppercase | ✅ Pass | 2.070s |  |
| Count from 1 to 5 | ✅ Pass | 3.320s |  |
| Math Calculation | ✅ Pass | 2.530s |  |
| Basic Echo Function | ✅ Pass | 2.892s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.337s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.165s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.728s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.819s |  |
| Search Query Function | ✅ Pass | 2.914s |  |
| Ask Advice Function | ✅ Pass | 2.922s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.084s |  |
| Basic Context Memory Test | ✅ Pass | 2.411s |  |
| Function Argument Memory Test | ✅ Pass | 2.959s |  |
| Function Response Memory Test | ✅ Pass | 2.431s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.410s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.624s |  |
| Penetration Testing Methodology | ✅ Pass | 6.428s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.695s |  |
| SQL Injection Attack Type | ✅ Pass | 2.965s |  |
| Penetration Testing Framework | ✅ Pass | 4.896s |  |
| Web Application Security Scanner | ✅ Pass | 3.397s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.067s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.240s

---

### pentester (claude-sonnet-4-5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.198s |  |
| Text Transform Uppercase | ✅ Pass | 2.201s |  |
| Count from 1 to 5 | ✅ Pass | 3.152s |  |
| Math Calculation | ✅ Pass | 2.462s |  |
| Basic Echo Function | ✅ Pass | 2.920s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.377s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.231s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.797s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.041s |  |
| Search Query Function | ✅ Pass | 3.350s |  |
| Ask Advice Function | ✅ Pass | 3.017s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.796s |  |
| Basic Context Memory Test | ✅ Pass | 3.022s |  |
| Function Argument Memory Test | ✅ Pass | 2.671s |  |
| Function Response Memory Test | ✅ Pass | 2.332s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.230s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.521s |  |
| Penetration Testing Methodology | ✅ Pass | 5.969s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.801s |  |
| SQL Injection Attack Type | ✅ Pass | 2.731s |  |
| Penetration Testing Framework | ✅ Pass | 5.749s |  |
| Web Application Security Scanner | ✅ Pass | 4.027s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.624s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.314s

---

