# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 17:37:43 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | claude-3-5-haiku-20241022 | false | 23/23 (100.00%) | 1.775s |
| simple_json | claude-3-5-haiku-20241022 | false | 5/5 (100.00%) | 1.121s |
| primary_agent | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.321s |
| assistant | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.421s |
| generator | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.634s |
| refiner | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.400s |
| adviser | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.173s |
| reflector | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.273s |
| searcher | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.310s |
| enricher | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.283s |
| coder | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.255s |
| installer | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.080s |
| pentester | claude-sonnet-4-20250514 | false | 23/23 (100.00%) | 2.216s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 2.241s

## Detailed Results

### simple (claude-3-5-haiku-20241022)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.645s |  |
| Text Transform Uppercase | ✅ Pass | 0.545s |  |
| Count from 1 to 5 | ✅ Pass | 0.627s |  |
| Math Calculation | ✅ Pass | 0.693s |  |
| Basic Echo Function | ✅ Pass | 1.362s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.733s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.650s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.495s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.940s |  |
| Search Query Function | ✅ Pass | 1.455s |  |
| Ask Advice Function | ✅ Pass | 1.320s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.369s |  |
| Basic Context Memory Test | ✅ Pass | 0.865s |  |
| Function Argument Memory Test | ✅ Pass | 0.760s |  |
| Function Response Memory Test | ✅ Pass | 0.757s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.208s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.712s |  |
| Penetration Testing Methodology | ✅ Pass | 5.432s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.512s |  |
| SQL Injection Attack Type | ✅ Pass | 0.721s |  |
| Penetration Testing Framework | ✅ Pass | 4.370s |  |
| Web Application Security Scanner | ✅ Pass | 3.190s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.454s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.775s

---

### simple_json (claude-3-5-haiku-20241022)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 0.916s |  |
| Person Information JSON | ✅ Pass | 0.945s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.414s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.908s |  |
| User Profile JSON | ✅ Pass | 1.418s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.121s

---

### primary_agent (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.229s |  |
| Text Transform Uppercase | ✅ Pass | 1.855s |  |
| Count from 1 to 5 | ✅ Pass | 2.270s |  |
| Math Calculation | ✅ Pass | 2.889s |  |
| Basic Echo Function | ✅ Pass | 2.008s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.347s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.423s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.095s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.667s |  |
| Search Query Function | ✅ Pass | 1.745s |  |
| Ask Advice Function | ✅ Pass | 1.904s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.118s |  |
| Basic Context Memory Test | ✅ Pass | 3.057s |  |
| Function Argument Memory Test | ✅ Pass | 1.817s |  |
| Function Response Memory Test | ✅ Pass | 1.492s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.930s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.263s |  |
| Penetration Testing Methodology | ✅ Pass | 4.274s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.378s |  |
| SQL Injection Attack Type | ✅ Pass | 1.456s |  |
| Penetration Testing Framework | ✅ Pass | 3.957s |  |
| Web Application Security Scanner | ✅ Pass | 3.736s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.469s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.321s

---

### assistant (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.546s |  |
| Text Transform Uppercase | ✅ Pass | 2.031s |  |
| Count from 1 to 5 | ✅ Pass | 1.722s |  |
| Math Calculation | ✅ Pass | 1.743s |  |
| Basic Echo Function | ✅ Pass | 1.564s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.345s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.471s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.885s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.236s |  |
| Search Query Function | ✅ Pass | 2.799s |  |
| Ask Advice Function | ✅ Pass | 2.927s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.297s |  |
| Basic Context Memory Test | ✅ Pass | 1.644s |  |
| Function Argument Memory Test | ✅ Pass | 1.321s |  |
| Function Response Memory Test | ✅ Pass | 1.210s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.695s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.353s |  |
| Penetration Testing Methodology | ✅ Pass | 3.769s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.736s |  |
| SQL Injection Attack Type | ✅ Pass | 1.849s |  |
| Penetration Testing Framework | ✅ Pass | 2.546s |  |
| Web Application Security Scanner | ✅ Pass | 4.222s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.765s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.421s

---

### generator (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 1.540s |  |
| Simple Math | ✅ Pass | 6.970s |  |
| Count from 1 to 5 | ✅ Pass | 1.405s |  |
| Math Calculation | ✅ Pass | 1.831s |  |
| Basic Echo Function | ✅ Pass | 2.134s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.529s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.094s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.876s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.130s |  |
| Search Query Function | ✅ Pass | 1.701s |  |
| Ask Advice Function | ✅ Pass | 2.396s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.912s |  |
| Basic Context Memory Test | ✅ Pass | 1.757s |  |
| Function Argument Memory Test | ✅ Pass | 2.499s |  |
| Function Response Memory Test | ✅ Pass | 1.479s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.460s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.257s |  |
| Penetration Testing Methodology | ✅ Pass | 4.027s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.885s |  |
| SQL Injection Attack Type | ✅ Pass | 1.375s |  |
| Penetration Testing Framework | ✅ Pass | 3.306s |  |
| Web Application Security Scanner | ✅ Pass | 3.395s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.607s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.634s

---

### refiner (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.278s |  |
| Text Transform Uppercase | ✅ Pass | 1.381s |  |
| Count from 1 to 5 | ✅ Pass | 1.474s |  |
| Math Calculation | ✅ Pass | 1.992s |  |
| Basic Echo Function | ✅ Pass | 1.858s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.756s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.366s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.006s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.380s |  |
| Search Query Function | ✅ Pass | 1.911s |  |
| Ask Advice Function | ✅ Pass | 4.252s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.158s |  |
| Basic Context Memory Test | ✅ Pass | 1.744s |  |
| Function Argument Memory Test | ✅ Pass | 1.543s |  |
| Function Response Memory Test | ✅ Pass | 1.615s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.379s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.462s |  |
| Penetration Testing Methodology | ✅ Pass | 7.839s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.170s |  |
| SQL Injection Attack Type | ✅ Pass | 1.281s |  |
| Penetration Testing Framework | ✅ Pass | 2.841s |  |
| Web Application Security Scanner | ✅ Pass | 3.031s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.478s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.400s

---

### adviser (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.291s |  |
| Text Transform Uppercase | ✅ Pass | 1.585s |  |
| Count from 1 to 5 | ✅ Pass | 1.649s |  |
| Math Calculation | ✅ Pass | 1.659s |  |
| Basic Echo Function | ✅ Pass | 2.023s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.537s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.608s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.778s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.093s |  |
| Search Query Function | ✅ Pass | 2.090s |  |
| Ask Advice Function | ✅ Pass | 2.140s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.390s |  |
| Basic Context Memory Test | ✅ Pass | 1.766s |  |
| Function Argument Memory Test | ✅ Pass | 1.637s |  |
| Function Response Memory Test | ✅ Pass | 1.481s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.765s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.836s |  |
| Penetration Testing Methodology | ✅ Pass | 3.969s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.440s |  |
| SQL Injection Attack Type | ✅ Pass | 1.268s |  |
| Penetration Testing Framework | ✅ Pass | 3.359s |  |
| Web Application Security Scanner | ✅ Pass | 2.706s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.894s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.173s

---

### reflector (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.391s |  |
| Text Transform Uppercase | ✅ Pass | 1.404s |  |
| Count from 1 to 5 | ✅ Pass | 2.017s |  |
| Math Calculation | ✅ Pass | 1.342s |  |
| Basic Echo Function | ✅ Pass | 2.861s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.668s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.572s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.843s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.955s |  |
| Search Query Function | ✅ Pass | 2.112s |  |
| Ask Advice Function | ✅ Pass | 2.254s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.134s |  |
| Basic Context Memory Test | ✅ Pass | 1.906s |  |
| Function Argument Memory Test | ✅ Pass | 1.531s |  |
| Function Response Memory Test | ✅ Pass | 1.708s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.168s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.392s |  |
| Penetration Testing Methodology | ✅ Pass | 4.646s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.123s |  |
| SQL Injection Attack Type | ✅ Pass | 1.421s |  |
| Penetration Testing Framework | ✅ Pass | 3.287s |  |
| Web Application Security Scanner | ✅ Pass | 3.795s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.734s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.273s

---

### searcher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.311s |  |
| Text Transform Uppercase | ✅ Pass | 1.556s |  |
| Count from 1 to 5 | ✅ Pass | 1.714s |  |
| Math Calculation | ✅ Pass | 1.524s |  |
| Basic Echo Function | ✅ Pass | 2.188s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.410s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.468s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.561s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.177s |  |
| Ask Advice Function | ✅ Pass | 1.990s |  |
| Search Query Function | ✅ Pass | 6.199s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.109s |  |
| Basic Context Memory Test | ✅ Pass | 1.421s |  |
| Function Argument Memory Test | ✅ Pass | 1.749s |  |
| Function Response Memory Test | ✅ Pass | 1.717s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.810s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.423s |  |
| Penetration Testing Methodology | ✅ Pass | 3.430s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.754s |  |
| SQL Injection Attack Type | ✅ Pass | 1.222s |  |
| Penetration Testing Framework | ✅ Pass | 3.153s |  |
| Web Application Security Scanner | ✅ Pass | 3.755s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.474s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.310s

---

### enricher (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.442s |  |
| Text Transform Uppercase | ✅ Pass | 2.035s |  |
| Count from 1 to 5 | ✅ Pass | 1.645s |  |
| Math Calculation | ✅ Pass | 1.961s |  |
| Basic Echo Function | ✅ Pass | 1.845s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.160s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.815s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.259s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.229s |  |
| Search Query Function | ✅ Pass | 2.001s |  |
| Ask Advice Function | ✅ Pass | 1.851s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.799s |  |
| Basic Context Memory Test | ✅ Pass | 1.369s |  |
| Function Argument Memory Test | ✅ Pass | 1.388s |  |
| Function Response Memory Test | ✅ Pass | 1.605s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.894s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.253s |  |
| Penetration Testing Methodology | ✅ Pass | 3.824s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.644s |  |
| SQL Injection Attack Type | ✅ Pass | 1.841s |  |
| Penetration Testing Framework | ✅ Pass | 3.826s |  |
| Web Application Security Scanner | ✅ Pass | 3.038s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.773s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.283s

---

### coder (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.577s |  |
| Text Transform Uppercase | ✅ Pass | 1.711s |  |
| Count from 1 to 5 | ✅ Pass | 1.964s |  |
| Math Calculation | ✅ Pass | 1.259s |  |
| Basic Echo Function | ✅ Pass | 2.009s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.745s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.544s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.936s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.711s |  |
| Search Query Function | ✅ Pass | 1.962s |  |
| Ask Advice Function | ✅ Pass | 2.404s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.581s |  |
| Basic Context Memory Test | ✅ Pass | 1.520s |  |
| Function Argument Memory Test | ✅ Pass | 1.706s |  |
| Function Response Memory Test | ✅ Pass | 1.637s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.423s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.296s |  |
| Penetration Testing Methodology | ✅ Pass | 3.956s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.808s |  |
| SQL Injection Attack Type | ✅ Pass | 1.049s |  |
| Penetration Testing Framework | ✅ Pass | 3.781s |  |
| Web Application Security Scanner | ✅ Pass | 3.583s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.684s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.255s

---

### installer (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.782s |  |
| Text Transform Uppercase | ✅ Pass | 1.440s |  |
| Count from 1 to 5 | ✅ Pass | 2.237s |  |
| Math Calculation | ✅ Pass | 1.488s |  |
| Basic Echo Function | ✅ Pass | 1.751s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.536s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.584s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.937s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.979s |  |
| Search Query Function | ✅ Pass | 2.061s |  |
| Ask Advice Function | ✅ Pass | 1.939s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.734s |  |
| Basic Context Memory Test | ✅ Pass | 1.866s |  |
| Function Argument Memory Test | ✅ Pass | 1.600s |  |
| Function Response Memory Test | ✅ Pass | 1.508s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.039s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.360s |  |
| Penetration Testing Methodology | ✅ Pass | 3.957s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.823s |  |
| SQL Injection Attack Type | ✅ Pass | 1.138s |  |
| Penetration Testing Framework | ✅ Pass | 2.859s |  |
| Web Application Security Scanner | ✅ Pass | 3.552s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.666s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.080s

---

### pentester (claude-sonnet-4-20250514)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.382s |  |
| Text Transform Uppercase | ✅ Pass | 1.478s |  |
| Count from 1 to 5 | ✅ Pass | 1.579s |  |
| Math Calculation | ✅ Pass | 1.301s |  |
| Basic Echo Function | ✅ Pass | 1.554s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.007s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.388s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.902s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.066s |  |
| Search Query Function | ✅ Pass | 1.862s |  |
| Ask Advice Function | ✅ Pass | 2.438s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.871s |  |
| Basic Context Memory Test | ✅ Pass | 1.855s |  |
| Function Argument Memory Test | ✅ Pass | 2.079s |  |
| Function Response Memory Test | ✅ Pass | 1.512s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.787s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.658s |  |
| Penetration Testing Methodology | ✅ Pass | 4.008s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.686s |  |
| SQL Injection Attack Type | ✅ Pass | 1.147s |  |
| Penetration Testing Framework | ✅ Pass | 5.001s |  |
| Web Application Security Scanner | ✅ Pass | 3.307s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.080s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.216s

---

