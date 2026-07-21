# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 15:26:19 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.6-27B-FP8 | false | 23/23 (100.00%) | 0.877s |
| simple_json | Qwen/Qwen3.6-27B-FP8 | false | 7/7 (100.00%) | 0.781s |
| primary_agent | Qwen/Qwen3.6-27B-FP8 | true | 22/23 (95.65%) | 3.168s |
| assistant | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 3.532s |
| generator | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 3.354s |
| refiner | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 3.250s |
| adviser | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 5.741s |
| reflector | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 0.835s |
| searcher | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 0.875s |
| enricher | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 0.936s |
| coder | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 3.361s |
| installer | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 6.277s |
| pentester | Qwen/Qwen3.6-27B-FP8 | true | 23/23 (100.00%) | 2.999s |

**Total**: 282/283 (99.65%) successful tests
**Overall average latency**: 2.881s

## Detailed Results

### simple (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.624s |  |
| Text Transform Uppercase | ✅ Pass | 0.357s |  |
| Count from 1 to 5 | ✅ Pass | 0.476s |  |
| Math Calculation | ✅ Pass | 0.345s |  |
| Basic Echo Function | ✅ Pass | 0.661s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.336s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.418s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.602s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.783s |  |
| Search Query Function | ✅ Pass | 0.637s |  |
| Ask Advice Function | ✅ Pass | 0.924s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.604s |  |
| Basic Context Memory Test | ✅ Pass | 0.490s |  |
| Function Argument Memory Test | ✅ Pass | 0.395s |  |
| Function Response Memory Test | ✅ Pass | 0.369s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.372s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.393s |  |
| Penetration Testing Methodology | ✅ Pass | 2.749s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.405s |  |
| SQL Injection Attack Type | ✅ Pass | 0.390s |  |
| Penetration Testing Framework | ✅ Pass | 2.549s |  |
| Web Application Security Scanner | ✅ Pass | 1.275s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.998s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.877s

---

### simple_json (Qwen/Qwen3.6-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.454s |  |
| Project Information JSON | ✅ Pass | 0.656s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.517s |  |
| User Profile JSON | ✅ Pass | 0.559s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.678s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.107s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.495s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 0.781s

---

### primary_agent (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.314s |  |
| Text Transform Uppercase | ✅ Pass | 2.568s |  |
| Count from 1 to 5 | ✅ Pass | 2.903s |  |
| Math Calculation | ✅ Pass | 1.699s |  |
| Basic Echo Function | ✅ Pass | 2.587s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.727s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.023s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.244s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.301s |  |
| Search Query Function | ✅ Pass | 0.896s |  |
| Ask Advice Function | ✅ Pass | 2.751s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.396s |  |
| Basic Context Memory Test | ✅ Pass | 3.293s |  |
| Function Argument Memory Test | ✅ Pass | 2.259s |  |
| Function Response Memory Test | ✅ Pass | 1.446s |  |
| Cybersecurity Workflow Memory Test | ❌ Fail | 0.247s | API returned unexpected status code: 502 |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.357s |  |
| Penetration Testing Methodology | ✅ Pass | 7.403s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.894s |  |
| SQL Injection Attack Type | ✅ Pass | 3.029s |  |
| Penetration Testing Framework | ✅ Pass | 7.308s |  |
| Web Application Security Scanner | ✅ Pass | 6.238s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.982s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 3.168s

---

### assistant (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.041s |  |
| Text Transform Uppercase | ✅ Pass | 2.711s |  |
| Count from 1 to 5 | ✅ Pass | 3.078s |  |
| Math Calculation | ✅ Pass | 1.944s |  |
| Basic Echo Function | ✅ Pass | 1.700s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.641s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.790s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.759s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.203s |  |
| Search Query Function | ✅ Pass | 1.706s |  |
| Ask Advice Function | ✅ Pass | 2.062s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.181s |  |
| Basic Context Memory Test | ✅ Pass | 3.017s |  |
| Function Argument Memory Test | ✅ Pass | 1.716s |  |
| Function Response Memory Test | ✅ Pass | 2.329s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.284s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.690s |  |
| Penetration Testing Methodology | ✅ Pass | 7.370s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.975s |  |
| SQL Injection Attack Type | ✅ Pass | 3.552s |  |
| Penetration Testing Framework | ✅ Pass | 7.924s |  |
| Web Application Security Scanner | ✅ Pass | 6.412s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.139s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.532s

---

### generator (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.677s |  |
| Text Transform Uppercase | ✅ Pass | 2.245s |  |
| Count from 1 to 5 | ✅ Pass | 3.411s |  |
| Math Calculation | ✅ Pass | 1.499s |  |
| Basic Echo Function | ✅ Pass | 2.209s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.010s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.073s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.806s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.294s |  |
| Search Query Function | ✅ Pass | 2.087s |  |
| Ask Advice Function | ✅ Pass | 2.326s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.099s |  |
| Basic Context Memory Test | ✅ Pass | 2.675s |  |
| Function Argument Memory Test | ✅ Pass | 2.940s |  |
| Function Response Memory Test | ✅ Pass | 4.411s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.787s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.556s |  |
| Penetration Testing Methodology | ✅ Pass | 7.681s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.115s |  |
| SQL Injection Attack Type | ✅ Pass | 3.597s |  |
| Penetration Testing Framework | ✅ Pass | 7.387s |  |
| Web Application Security Scanner | ✅ Pass | 5.811s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.429s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.354s

---

### refiner (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.045s |  |
| Text Transform Uppercase | ✅ Pass | 2.766s |  |
| Count from 1 to 5 | ✅ Pass | 3.423s |  |
| Math Calculation | ✅ Pass | 1.984s |  |
| Basic Echo Function | ✅ Pass | 2.168s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.710s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.720s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.724s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.737s |  |
| Search Query Function | ✅ Pass | 1.641s |  |
| Ask Advice Function | ✅ Pass | 1.517s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.751s |  |
| Basic Context Memory Test | ✅ Pass | 1.921s |  |
| Function Argument Memory Test | ✅ Pass | 3.008s |  |
| Function Response Memory Test | ✅ Pass | 1.607s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.890s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.124s |  |
| Penetration Testing Methodology | ✅ Pass | 6.267s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.601s |  |
| SQL Injection Attack Type | ✅ Pass | 3.598s |  |
| Penetration Testing Framework | ✅ Pass | 7.453s |  |
| Web Application Security Scanner | ✅ Pass | 5.439s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.652s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.250s

---

### adviser (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.796s |  |
| Text Transform Uppercase | ✅ Pass | 2.022s |  |
| Count from 1 to 5 | ✅ Pass | 2.724s |  |
| Math Calculation | ✅ Pass | 2.223s |  |
| Basic Echo Function | ✅ Pass | 1.523s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.985s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.725s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.368s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.680s |  |
| Search Query Function | ✅ Pass | 1.935s |  |
| Ask Advice Function | ✅ Pass | 1.873s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.135s |  |
| Basic Context Memory Test | ✅ Pass | 2.274s |  |
| Function Argument Memory Test | ✅ Pass | 1.882s |  |
| Function Response Memory Test | ✅ Pass | 1.363s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.059s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.751s |  |
| Penetration Testing Methodology | ✅ Pass | 8.470s |  |
| SQL Injection Attack Type | ✅ Pass | 3.770s |  |
| Penetration Testing Framework | ✅ Pass | 7.895s |  |
| Web Application Security Scanner | ✅ Pass | 4.795s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.278s |  |
| Vulnerability Assessment Tools | ✅ Pass | 69.509s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 5.741s

---

### reflector (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.624s |  |
| Text Transform Uppercase | ✅ Pass | 0.518s |  |
| Count from 1 to 5 | ✅ Pass | 0.484s |  |
| Math Calculation | ✅ Pass | 0.394s |  |
| Basic Echo Function | ✅ Pass | 0.672s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.380s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.453s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.751s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.903s |  |
| Search Query Function | ✅ Pass | 0.751s |  |
| Ask Advice Function | ✅ Pass | 0.954s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.844s |  |
| Basic Context Memory Test | ✅ Pass | 0.489s |  |
| Function Argument Memory Test | ✅ Pass | 0.435s |  |
| Function Response Memory Test | ✅ Pass | 0.405s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.411s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.476s |  |
| Penetration Testing Methodology | ✅ Pass | 2.687s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.877s |  |
| SQL Injection Attack Type | ✅ Pass | 0.364s |  |
| Penetration Testing Framework | ✅ Pass | 1.148s |  |
| Web Application Security Scanner | ✅ Pass | 1.275s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.904s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.835s

---

### searcher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.259s |  |
| Text Transform Uppercase | ✅ Pass | 0.397s |  |
| Count from 1 to 5 | ✅ Pass | 0.503s |  |
| Math Calculation | ✅ Pass | 0.367s |  |
| Basic Echo Function | ✅ Pass | 0.772s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.366s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.453s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.690s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.852s |  |
| Search Query Function | ✅ Pass | 0.829s |  |
| Ask Advice Function | ✅ Pass | 0.950s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.866s |  |
| Basic Context Memory Test | ✅ Pass | 0.521s |  |
| Function Argument Memory Test | ✅ Pass | 0.431s |  |
| Function Response Memory Test | ✅ Pass | 0.387s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.551s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.657s |  |
| Penetration Testing Methodology | ✅ Pass | 2.595s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.407s |  |
| SQL Injection Attack Type | ✅ Pass | 0.366s |  |
| Penetration Testing Framework | ✅ Pass | 1.379s |  |
| Web Application Security Scanner | ✅ Pass | 1.466s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.048s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.875s

---

### enricher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.166s |  |
| Text Transform Uppercase | ✅ Pass | 0.501s |  |
| Count from 1 to 5 | ✅ Pass | 0.497s |  |
| Math Calculation | ✅ Pass | 0.366s |  |
| Basic Echo Function | ✅ Pass | 0.734s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.352s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.461s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.837s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.827s |  |
| Search Query Function | ✅ Pass | 0.790s |  |
| Ask Advice Function | ✅ Pass | 0.995s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.740s |  |
| Basic Context Memory Test | ✅ Pass | 0.533s |  |
| Function Argument Memory Test | ✅ Pass | 0.451s |  |
| Function Response Memory Test | ✅ Pass | 0.409s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.504s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.580s |  |
| Penetration Testing Methodology | ✅ Pass | 2.463s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.674s |  |
| SQL Injection Attack Type | ✅ Pass | 0.346s |  |
| Penetration Testing Framework | ✅ Pass | 2.681s |  |
| Web Application Security Scanner | ✅ Pass | 1.613s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.001s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.936s

---

### coder (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.135s |  |
| Text Transform Uppercase | ✅ Pass | 2.539s |  |
| Count from 1 to 5 | ✅ Pass | 3.738s |  |
| Math Calculation | ✅ Pass | 1.360s |  |
| Basic Echo Function | ✅ Pass | 2.107s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.081s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.350s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.815s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.649s |  |
| Search Query Function | ✅ Pass | 1.346s |  |
| Ask Advice Function | ✅ Pass | 2.244s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.088s |  |
| Basic Context Memory Test | ✅ Pass | 3.039s |  |
| Function Argument Memory Test | ✅ Pass | 1.380s |  |
| Function Response Memory Test | ✅ Pass | 3.934s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.733s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.180s |  |
| Penetration Testing Methodology | ✅ Pass | 6.927s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.605s |  |
| SQL Injection Attack Type | ✅ Pass | 3.516s |  |
| Penetration Testing Framework | ✅ Pass | 7.560s |  |
| Web Application Security Scanner | ✅ Pass | 4.755s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.213s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.361s

---

### installer (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.076s |  |
| Text Transform Uppercase | ✅ Pass | 2.101s |  |
| Count from 1 to 5 | ✅ Pass | 4.749s |  |
| Math Calculation | ✅ Pass | 1.669s |  |
| Basic Echo Function | ✅ Pass | 1.741s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.022s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.826s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.107s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.669s |  |
| Search Query Function | ✅ Pass | 1.274s |  |
| Ask Advice Function | ✅ Pass | 2.619s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.557s |  |
| Basic Context Memory Test | ✅ Pass | 2.334s |  |
| Function Argument Memory Test | ✅ Pass | 1.931s |  |
| Function Response Memory Test | ✅ Pass | 1.398s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.201s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.667s |  |
| Penetration Testing Methodology | ✅ Pass | 7.674s |  |
| SQL Injection Attack Type | ✅ Pass | 5.422s |  |
| Penetration Testing Framework | ✅ Pass | 11.689s |  |
| Web Application Security Scanner | ✅ Pass | 5.730s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.548s |  |
| Vulnerability Assessment Tools | ✅ Pass | 74.361s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.277s

---

### pentester (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.616s |  |
| Text Transform Uppercase | ✅ Pass | 2.141s |  |
| Count from 1 to 5 | ✅ Pass | 2.635s |  |
| Math Calculation | ✅ Pass | 1.734s |  |
| Basic Echo Function | ✅ Pass | 3.264s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.129s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.503s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.289s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.096s |  |
| Search Query Function | ✅ Pass | 1.516s |  |
| Ask Advice Function | ✅ Pass | 1.979s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.287s |  |
| Basic Context Memory Test | ✅ Pass | 2.120s |  |
| Function Argument Memory Test | ✅ Pass | 2.736s |  |
| Function Response Memory Test | ✅ Pass | 1.932s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.794s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.208s |  |
| Penetration Testing Methodology | ✅ Pass | 6.298s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.496s |  |
| SQL Injection Attack Type | ✅ Pass | 3.620s |  |
| Penetration Testing Framework | ✅ Pass | 6.025s |  |
| Web Application Security Scanner | ✅ Pass | 5.315s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.221s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.999s

---

