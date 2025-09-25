# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 19:43:32 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | llama3.1:8b | false | 23/23 (100.00%) | 0.641s |
| simple_json | llama3.1:8b | false | 5/5 (100.00%) | 0.514s |
| primary_agent | llama3.1:8b | false | 23/23 (100.00%) | 0.545s |
| assistant | llama3.1:8b | false | 23/23 (100.00%) | 0.543s |
| generator | llama3.1:8b | false | 23/23 (100.00%) | 0.512s |
| refiner | llama3.1:8b | false | 23/23 (100.00%) | 0.528s |
| adviser | llama3.1:8b | false | 23/23 (100.00%) | 0.538s |
| reflector | llama3.1:8b | false | 23/23 (100.00%) | 0.545s |
| searcher | llama3.1:8b | false | 23/23 (100.00%) | 0.533s |
| enricher | llama3.1:8b | false | 23/23 (100.00%) | 0.546s |
| coder | llama3.1:8b | false | 23/23 (100.00%) | 0.565s |
| installer | llama3.1:8b | false | 23/23 (100.00%) | 0.546s |
| pentester | llama3.1:8b | false | 23/23 (100.00%) | 0.543s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 0.548s

## Detailed Results

### simple (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.290s |  |
| Text Transform Uppercase | ✅ Pass | 0.323s |  |
| Count from 1 to 5 | ✅ Pass | 0.366s |  |
| Math Calculation | ✅ Pass | 0.314s |  |
| Basic Echo Function | ✅ Pass | 0.431s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.312s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.472s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.542s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.403s |  |
| Search Query Function | ✅ Pass | 0.411s |  |
| Ask Advice Function | ✅ Pass | 0.502s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.575s |  |
| Basic Context Memory Test | ✅ Pass | 0.457s |  |
| Function Argument Memory Test | ✅ Pass | 0.356s |  |
| Function Response Memory Test | ✅ Pass | 0.405s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.218s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.315s |  |
| Penetration Testing Methodology | ✅ Pass | 1.245s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.782s |  |
| SQL Injection Attack Type | ✅ Pass | 0.319s |  |
| Penetration Testing Framework | ✅ Pass | 1.296s |  |
| Web Application Security Scanner | ✅ Pass | 0.962s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.437s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.641s

---

### simple_json (llama3.1:8b)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.812s |  |
| Person Information JSON | ✅ Pass | 0.427s |  |
| Project Information JSON | ✅ Pass | 0.410s |  |
| User Profile JSON | ✅ Pass | 0.445s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.472s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.514s

---

### primary_agent (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.306s |  |
| Text Transform Uppercase | ✅ Pass | 0.313s |  |
| Count from 1 to 5 | ✅ Pass | 0.348s |  |
| Math Calculation | ✅ Pass | 0.306s |  |
| Basic Echo Function | ✅ Pass | 0.408s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.306s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.419s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.517s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.388s |  |
| Search Query Function | ✅ Pass | 0.401s |  |
| Ask Advice Function | ✅ Pass | 0.470s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.573s |  |
| Basic Context Memory Test | ✅ Pass | 0.633s |  |
| Function Argument Memory Test | ✅ Pass | 0.334s |  |
| Function Response Memory Test | ✅ Pass | 0.303s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.530s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.273s |  |
| Penetration Testing Methodology | ✅ Pass | 1.205s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.701s |  |
| SQL Injection Attack Type | ✅ Pass | 0.444s |  |
| Penetration Testing Framework | ✅ Pass | 1.015s |  |
| Web Application Security Scanner | ✅ Pass | 0.924s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.400s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.545s

---

### assistant (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.309s |  |
| Text Transform Uppercase | ✅ Pass | 0.321s |  |
| Count from 1 to 5 | ✅ Pass | 0.349s |  |
| Math Calculation | ✅ Pass | 0.303s |  |
| Basic Echo Function | ✅ Pass | 0.403s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.302s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.423s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.518s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.430s |  |
| Search Query Function | ✅ Pass | 0.401s |  |
| Ask Advice Function | ✅ Pass | 0.467s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.515s |  |
| Basic Context Memory Test | ✅ Pass | 0.638s |  |
| Function Argument Memory Test | ✅ Pass | 0.347s |  |
| Function Response Memory Test | ✅ Pass | 0.304s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.534s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.274s |  |
| Penetration Testing Methodology | ✅ Pass | 1.197s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.663s |  |
| SQL Injection Attack Type | ✅ Pass | 0.341s |  |
| Penetration Testing Framework | ✅ Pass | 1.142s |  |
| Web Application Security Scanner | ✅ Pass | 0.889s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.398s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.543s

---

### generator (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.303s |  |
| Text Transform Uppercase | ✅ Pass | 0.327s |  |
| Count from 1 to 5 | ✅ Pass | 0.346s |  |
| Math Calculation | ✅ Pass | 0.302s |  |
| Basic Echo Function | ✅ Pass | 0.404s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.304s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.418s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.519s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.439s |  |
| Search Query Function | ✅ Pass | 0.399s |  |
| Ask Advice Function | ✅ Pass | 0.470s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.511s |  |
| Basic Context Memory Test | ✅ Pass | 0.473s |  |
| Function Argument Memory Test | ✅ Pass | 0.294s |  |
| Function Response Memory Test | ✅ Pass | 0.305s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.530s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.279s |  |
| Penetration Testing Methodology | ✅ Pass | 0.812s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.864s |  |
| SQL Injection Attack Type | ✅ Pass | 0.305s |  |
| Penetration Testing Framework | ✅ Pass | 0.795s |  |
| Web Application Security Scanner | ✅ Pass | 0.970s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.398s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.512s

---

### refiner (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.301s |  |
| Text Transform Uppercase | ✅ Pass | 0.313s |  |
| Count from 1 to 5 | ✅ Pass | 0.350s |  |
| Math Calculation | ✅ Pass | 0.305s |  |
| Basic Echo Function | ✅ Pass | 0.405s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.304s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.420s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.520s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.428s |  |
| Search Query Function | ✅ Pass | 0.400s |  |
| Ask Advice Function | ✅ Pass | 0.468s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.509s |  |
| Basic Context Memory Test | ✅ Pass | 0.450s |  |
| Function Argument Memory Test | ✅ Pass | 0.339s |  |
| Function Response Memory Test | ✅ Pass | 0.300s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.529s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.274s |  |
| Penetration Testing Methodology | ✅ Pass | 1.232s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.385s |  |
| SQL Injection Attack Type | ✅ Pass | 0.397s |  |
| Penetration Testing Framework | ✅ Pass | 1.209s |  |
| Web Application Security Scanner | ✅ Pass | 0.906s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.397s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.528s

---

### adviser (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.307s |  |
| Text Transform Uppercase | ✅ Pass | 0.315s |  |
| Count from 1 to 5 | ✅ Pass | 0.349s |  |
| Math Calculation | ✅ Pass | 0.304s |  |
| Basic Echo Function | ✅ Pass | 0.406s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.301s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.421s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.517s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.432s |  |
| Search Query Function | ✅ Pass | 0.399s |  |
| Ask Advice Function | ✅ Pass | 0.470s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.508s |  |
| Basic Context Memory Test | ✅ Pass | 0.477s |  |
| Function Argument Memory Test | ✅ Pass | 0.339s |  |
| Function Response Memory Test | ✅ Pass | 0.303s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.532s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.275s |  |
| Penetration Testing Methodology | ✅ Pass | 1.082s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.479s |  |
| SQL Injection Attack Type | ✅ Pass | 0.315s |  |
| Penetration Testing Framework | ✅ Pass | 1.092s |  |
| Web Application Security Scanner | ✅ Pass | 1.331s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.400s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.538s

---

### reflector (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.303s |  |
| Text Transform Uppercase | ✅ Pass | 0.316s |  |
| Count from 1 to 5 | ✅ Pass | 0.356s |  |
| Math Calculation | ✅ Pass | 0.301s |  |
| Basic Echo Function | ✅ Pass | 0.401s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.307s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.418s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.518s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.424s |  |
| Search Query Function | ✅ Pass | 0.401s |  |
| Ask Advice Function | ✅ Pass | 0.467s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.511s |  |
| Basic Context Memory Test | ✅ Pass | 0.485s |  |
| Function Argument Memory Test | ✅ Pass | 0.366s |  |
| Function Response Memory Test | ✅ Pass | 0.307s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.542s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.277s |  |
| Penetration Testing Methodology | ✅ Pass | 1.486s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.552s |  |
| SQL Injection Attack Type | ✅ Pass | 0.313s |  |
| Penetration Testing Framework | ✅ Pass | 1.079s |  |
| Web Application Security Scanner | ✅ Pass | 0.999s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.399s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.545s

---

### searcher (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.307s |  |
| Text Transform Uppercase | ✅ Pass | 0.315s |  |
| Count from 1 to 5 | ✅ Pass | 0.343s |  |
| Math Calculation | ✅ Pass | 0.304s |  |
| Basic Echo Function | ✅ Pass | 0.407s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.300s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.422s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.517s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.430s |  |
| Search Query Function | ✅ Pass | 0.400s |  |
| Ask Advice Function | ✅ Pass | 0.468s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.516s |  |
| Basic Context Memory Test | ✅ Pass | 0.472s |  |
| Function Argument Memory Test | ✅ Pass | 0.352s |  |
| Function Response Memory Test | ✅ Pass | 0.302s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.528s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.276s |  |
| Penetration Testing Methodology | ✅ Pass | 1.057s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.729s |  |
| SQL Injection Attack Type | ✅ Pass | 0.444s |  |
| Penetration Testing Framework | ✅ Pass | 1.007s |  |
| Web Application Security Scanner | ✅ Pass | 0.888s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.468s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.533s

---

### enricher (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.302s |  |
| Text Transform Uppercase | ✅ Pass | 0.317s |  |
| Count from 1 to 5 | ✅ Pass | 0.352s |  |
| Math Calculation | ✅ Pass | 0.264s |  |
| Basic Echo Function | ✅ Pass | 0.397s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.303s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.424s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.516s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.425s |  |
| Search Query Function | ✅ Pass | 0.400s |  |
| Ask Advice Function | ✅ Pass | 0.466s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.387s |  |
| Basic Context Memory Test | ✅ Pass | 0.484s |  |
| Function Argument Memory Test | ✅ Pass | 0.337s |  |
| Function Response Memory Test | ✅ Pass | 0.301s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.534s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.274s |  |
| Penetration Testing Methodology | ✅ Pass | 1.201s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.817s |  |
| SQL Injection Attack Type | ✅ Pass | 0.526s |  |
| Penetration Testing Framework | ✅ Pass | 1.105s |  |
| Web Application Security Scanner | ✅ Pass | 0.971s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.453s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.546s

---

### coder (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.312s |  |
| Text Transform Uppercase | ✅ Pass | 0.316s |  |
| Count from 1 to 5 | ✅ Pass | 0.349s |  |
| Math Calculation | ✅ Pass | 0.301s |  |
| Basic Echo Function | ✅ Pass | 0.401s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.305s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.425s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.518s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.429s |  |
| Search Query Function | ✅ Pass | 0.399s |  |
| Ask Advice Function | ✅ Pass | 0.469s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.556s |  |
| Basic Context Memory Test | ✅ Pass | 0.638s |  |
| Function Argument Memory Test | ✅ Pass | 0.380s |  |
| Function Response Memory Test | ✅ Pass | 0.310s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.530s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.275s |  |
| Penetration Testing Methodology | ✅ Pass | 1.201s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.092s |  |
| SQL Injection Attack Type | ✅ Pass | 0.315s |  |
| Penetration Testing Framework | ✅ Pass | 1.159s |  |
| Web Application Security Scanner | ✅ Pass | 0.896s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.403s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.565s

---

### installer (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.305s |  |
| Text Transform Uppercase | ✅ Pass | 0.315s |  |
| Count from 1 to 5 | ✅ Pass | 0.354s |  |
| Math Calculation | ✅ Pass | 0.303s |  |
| Basic Echo Function | ✅ Pass | 0.405s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.306s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.417s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.518s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.431s |  |
| Search Query Function | ✅ Pass | 0.398s |  |
| Ask Advice Function | ✅ Pass | 0.467s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.508s |  |
| Basic Context Memory Test | ✅ Pass | 0.639s |  |
| Function Argument Memory Test | ✅ Pass | 0.337s |  |
| Function Response Memory Test | ✅ Pass | 0.304s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.530s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.277s |  |
| Penetration Testing Methodology | ✅ Pass | 1.198s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.696s |  |
| SQL Injection Attack Type | ✅ Pass | 0.469s |  |
| Penetration Testing Framework | ✅ Pass | 1.076s |  |
| Web Application Security Scanner | ✅ Pass | 0.890s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.399s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.546s

---

### pentester (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.303s |  |
| Text Transform Uppercase | ✅ Pass | 0.316s |  |
| Count from 1 to 5 | ✅ Pass | 0.356s |  |
| Math Calculation | ✅ Pass | 0.302s |  |
| Basic Echo Function | ✅ Pass | 0.404s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.301s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.420s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.520s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.431s |  |
| Search Query Function | ✅ Pass | 0.399s |  |
| Ask Advice Function | ✅ Pass | 0.467s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.510s |  |
| Basic Context Memory Test | ✅ Pass | 0.505s |  |
| Function Argument Memory Test | ✅ Pass | 0.334s |  |
| Function Response Memory Test | ✅ Pass | 0.306s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.534s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.274s |  |
| Penetration Testing Methodology | ✅ Pass | 1.208s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.880s |  |
| SQL Injection Attack Type | ✅ Pass | 0.308s |  |
| Penetration Testing Framework | ✅ Pass | 0.987s |  |
| Web Application Security Scanner | ✅ Pass | 1.013s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.398s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.543s

---

