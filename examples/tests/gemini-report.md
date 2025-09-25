# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 12:15:20 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 23/23 (100.00%) | 0.628s |
| simple_json | gemini-2.0-flash-lite | false | 5/5 (100.00%) | 0.636s |
| primary_agent | gemini-2.5-flash-lite-preview-06-17 | true | 23/23 (100.00%) | 2.086s |
| assistant | gemini-2.5-flash | true | 23/23 (100.00%) | 2.429s |
| generator | gemini-2.5-pro | true | 23/23 (100.00%) | 6.644s |
| refiner | gemini-2.0-flash | false | 23/23 (100.00%) | 0.792s |
| adviser | gemini-2.5-flash | true | 23/23 (100.00%) | 2.617s |
| reflector | gemini-2.5-flash | true | 23/23 (100.00%) | 2.673s |
| searcher | gemini-2.0-flash | false | 23/23 (100.00%) | 0.827s |
| enricher | gemini-2.0-flash | false | 23/23 (100.00%) | 0.696s |
| coder | gemini-2.5-pro | true | 23/23 (100.00%) | 6.721s |
| installer | gemini-2.5-flash-lite-preview-06-17 | true | 23/23 (100.00%) | 2.077s |
| pentester | gemini-2.5-flash-lite-preview-06-17 | true | 23/23 (100.00%) | 2.115s |

**Total**: 281/281 (100.00%) successful tests
**Overall average latency**: 2.492s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.267s |  |
| Text Transform Uppercase | ✅ Pass | 0.454s |  |
| Count from 1 to 5 | ✅ Pass | 0.452s |  |
| Math Calculation | ✅ Pass | 0.511s |  |
| Basic Echo Function | ✅ Pass | 0.552s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.447s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.457s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.616s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.589s |  |
| Search Query Function | ✅ Pass | 0.666s |  |
| Ask Advice Function | ✅ Pass | 0.649s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.540s |  |
| Basic Context Memory Test | ✅ Pass | 0.603s |  |
| Function Argument Memory Test | ✅ Pass | 0.524s |  |
| Function Response Memory Test | ✅ Pass | 0.439s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.892s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.442s |  |
| Penetration Testing Methodology | ✅ Pass | 1.135s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.784s |  |
| SQL Injection Attack Type | ✅ Pass | 0.433s |  |
| Penetration Testing Framework | ✅ Pass | 0.733s |  |
| Web Application Security Scanner | ✅ Pass | 0.579s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.659s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.628s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.839s |  |
| Project Information JSON | ✅ Pass | 0.559s |  |
| Person Information JSON | ✅ Pass | 0.588s |  |
| User Profile JSON | ✅ Pass | 0.584s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.610s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 0.636s

---

### primary_agent (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.116s |  |
| Text Transform Uppercase | ✅ Pass | 1.375s |  |
| Count from 1 to 5 | ✅ Pass | 3.936s |  |
| Math Calculation | ✅ Pass | 1.435s |  |
| Basic Echo Function | ✅ Pass | 1.421s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.500s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.250s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.194s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.540s |  |
| Search Query Function | ✅ Pass | 1.162s |  |
| Ask Advice Function | ✅ Pass | 1.059s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.863s |  |
| Basic Context Memory Test | ✅ Pass | 1.659s |  |
| Function Argument Memory Test | ✅ Pass | 1.154s |  |
| Function Response Memory Test | ✅ Pass | 1.746s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.982s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.398s |  |
| Penetration Testing Methodology | ✅ Pass | 3.207s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.195s |  |
| SQL Injection Attack Type | ✅ Pass | 1.280s |  |
| Penetration Testing Framework | ✅ Pass | 3.925s |  |
| Web Application Security Scanner | ✅ Pass | 3.732s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.844s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.086s

---

### assistant (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.418s |  |
| Text Transform Uppercase | ✅ Pass | 1.517s |  |
| Count from 1 to 5 | ✅ Pass | 1.412s |  |
| Math Calculation | ✅ Pass | 1.460s |  |
| Basic Echo Function | ✅ Pass | 1.421s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.418s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.268s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.242s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.999s |  |
| Search Query Function | ✅ Pass | 1.593s |  |
| Ask Advice Function | ✅ Pass | 1.502s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.043s |  |
| Basic Context Memory Test | ✅ Pass | 1.682s |  |
| Function Argument Memory Test | ✅ Pass | 2.103s |  |
| Function Response Memory Test | ✅ Pass | 2.083s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.081s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.775s |  |
| Penetration Testing Methodology | ✅ Pass | 4.419s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.021s |  |
| SQL Injection Attack Type | ✅ Pass | 1.851s |  |
| Penetration Testing Framework | ✅ Pass | 6.956s |  |
| Web Application Security Scanner | ✅ Pass | 5.801s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.797s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.429s

---

### generator (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.724s |  |
| Text Transform Uppercase | ✅ Pass | 5.532s |  |
| Count from 1 to 5 | ✅ Pass | 5.758s |  |
| Math Calculation | ✅ Pass | 2.834s |  |
| Basic Echo Function | ✅ Pass | 4.779s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.724s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.085s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.978s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 5.343s |  |
| Search Query Function | ✅ Pass | 2.901s |  |
| Ask Advice Function | ✅ Pass | 2.957s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.677s |  |
| Basic Context Memory Test | ✅ Pass | 5.643s |  |
| Function Argument Memory Test | ✅ Pass | 7.991s |  |
| Function Response Memory Test | ✅ Pass | 5.275s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.472s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 5.421s |  |
| Penetration Testing Methodology | ✅ Pass | 13.405s |  |
| SQL Injection Attack Type | ✅ Pass | 7.844s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.804s |  |
| Penetration Testing Framework | ✅ Pass | 16.143s |  |
| Web Application Security Scanner | ✅ Pass | 12.792s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.713s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.644s

---

### refiner (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.295s |  |
| Text Transform Uppercase | ✅ Pass | 0.574s |  |
| Count from 1 to 5 | ✅ Pass | 0.567s |  |
| Math Calculation | ✅ Pass | 0.515s |  |
| Basic Echo Function | ✅ Pass | 0.670s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.551s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.588s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.650s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.571s |  |
| Search Query Function | ✅ Pass | 0.687s |  |
| Ask Advice Function | ✅ Pass | 0.708s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.618s |  |
| Basic Context Memory Test | ✅ Pass | 0.634s |  |
| Function Argument Memory Test | ✅ Pass | 0.658s |  |
| Function Response Memory Test | ✅ Pass | 0.432s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.783s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.518s |  |
| Penetration Testing Methodology | ✅ Pass | 1.944s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.965s |  |
| SQL Injection Attack Type | ✅ Pass | 0.547s |  |
| Penetration Testing Framework | ✅ Pass | 1.033s |  |
| Web Application Security Scanner | ✅ Pass | 0.908s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.794s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.792s

---

### adviser (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.334s |  |
| Text Transform Uppercase | ✅ Pass | 1.786s |  |
| Count from 1 to 5 | ✅ Pass | 1.540s |  |
| Math Calculation | ✅ Pass | 1.511s |  |
| Basic Echo Function | ✅ Pass | 1.858s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.354s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.092s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.093s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.562s |  |
| Search Query Function | ✅ Pass | 1.508s |  |
| Ask Advice Function | ✅ Pass | 1.436s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.170s |  |
| Basic Context Memory Test | ✅ Pass | 1.733s |  |
| Function Argument Memory Test | ✅ Pass | 2.206s |  |
| Function Response Memory Test | ✅ Pass | 2.017s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.045s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.398s |  |
| Penetration Testing Methodology | ✅ Pass | 6.042s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.806s |  |
| SQL Injection Attack Type | ✅ Pass | 2.410s |  |
| Penetration Testing Framework | ✅ Pass | 8.296s |  |
| Web Application Security Scanner | ✅ Pass | 5.481s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.503s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.617s

---

### reflector (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.613s |  |
| Text Transform Uppercase | ✅ Pass | 1.616s |  |
| Count from 1 to 5 | ✅ Pass | 1.899s |  |
| Math Calculation | ✅ Pass | 1.329s |  |
| Basic Echo Function | ✅ Pass | 1.557s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.530s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.245s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.313s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.396s |  |
| Search Query Function | ✅ Pass | 1.367s |  |
| Ask Advice Function | ✅ Pass | 1.446s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.165s |  |
| Basic Context Memory Test | ✅ Pass | 1.369s |  |
| Function Argument Memory Test | ✅ Pass | 1.616s |  |
| Function Response Memory Test | ✅ Pass | 1.476s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.868s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.807s |  |
| Penetration Testing Methodology | ✅ Pass | 4.537s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.010s |  |
| SQL Injection Attack Type | ✅ Pass | 5.049s |  |
| Penetration Testing Framework | ✅ Pass | 8.524s |  |
| Web Application Security Scanner | ✅ Pass | 7.435s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.300s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.673s

---

### searcher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.405s |  |
| Text Transform Uppercase | ✅ Pass | 0.399s |  |
| Count from 1 to 5 | ✅ Pass | 0.656s |  |
| Math Calculation | ✅ Pass | 0.650s |  |
| Basic Echo Function | ✅ Pass | 0.726s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.562s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.589s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.663s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.586s |  |
| Search Query Function | ✅ Pass | 0.732s |  |
| Ask Advice Function | ✅ Pass | 0.740s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.698s |  |
| Basic Context Memory Test | ✅ Pass | 0.530s |  |
| Function Argument Memory Test | ✅ Pass | 0.603s |  |
| Function Response Memory Test | ✅ Pass | 0.563s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.733s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.554s |  |
| Penetration Testing Methodology | ✅ Pass | 1.459s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.166s |  |
| SQL Injection Attack Type | ✅ Pass | 0.478s |  |
| Penetration Testing Framework | ✅ Pass | 0.982s |  |
| Web Application Security Scanner | ✅ Pass | 1.829s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.714s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.827s

---

### enricher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.574s |  |
| Text Transform Uppercase | ✅ Pass | 0.494s |  |
| Count from 1 to 5 | ✅ Pass | 0.599s |  |
| Math Calculation | ✅ Pass | 0.513s |  |
| Basic Echo Function | ✅ Pass | 0.644s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.657s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.537s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.625s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.790s |  |
| Search Query Function | ✅ Pass | 0.628s |  |
| Ask Advice Function | ✅ Pass | 0.683s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.547s |  |
| Basic Context Memory Test | ✅ Pass | 0.498s |  |
| Function Argument Memory Test | ✅ Pass | 0.492s |  |
| Function Response Memory Test | ✅ Pass | 0.489s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.790s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.575s |  |
| Penetration Testing Methodology | ✅ Pass | 1.621s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.208s |  |
| SQL Injection Attack Type | ✅ Pass | 0.519s |  |
| Penetration Testing Framework | ✅ Pass | 0.599s |  |
| Web Application Security Scanner | ✅ Pass | 1.202s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.718s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.696s

---

### coder (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.041s |  |
| Text Transform Uppercase | ✅ Pass | 4.883s |  |
| Count from 1 to 5 | ✅ Pass | 6.256s |  |
| Math Calculation | ✅ Pass | 3.612s |  |
| Basic Echo Function | ✅ Pass | 3.570s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.161s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.176s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 6.678s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.663s |  |
| Search Query Function | ✅ Pass | 3.225s |  |
| Ask Advice Function | ✅ Pass | 3.247s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.221s |  |
| Basic Context Memory Test | ✅ Pass | 4.520s |  |
| Function Argument Memory Test | ✅ Pass | 3.173s |  |
| Function Response Memory Test | ✅ Pass | 5.907s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.887s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.801s |  |
| Penetration Testing Methodology | ✅ Pass | 11.351s |  |
| SQL Injection Attack Type | ✅ Pass | 9.503s |  |
| Vulnerability Assessment Tools | ✅ Pass | 20.720s |  |
| Penetration Testing Framework | ✅ Pass | 17.476s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.966s |  |
| Web Application Security Scanner | ✅ Pass | 13.531s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 6.721s

---

### installer (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.317s |  |
| Text Transform Uppercase | ✅ Pass | 1.268s |  |
| Count from 1 to 5 | ✅ Pass | 1.783s |  |
| Math Calculation | ✅ Pass | 1.350s |  |
| Basic Echo Function | ✅ Pass | 2.323s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.581s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.107s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.463s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.461s |  |
| Search Query Function | ✅ Pass | 1.321s |  |
| Ask Advice Function | ✅ Pass | 1.151s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.840s |  |
| Basic Context Memory Test | ✅ Pass | 1.226s |  |
| Function Argument Memory Test | ✅ Pass | 1.327s |  |
| Function Response Memory Test | ✅ Pass | 1.977s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.815s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.053s |  |
| Penetration Testing Methodology | ✅ Pass | 3.362s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.541s |  |
| SQL Injection Attack Type | ✅ Pass | 1.094s |  |
| Penetration Testing Framework | ✅ Pass | 4.847s |  |
| Web Application Security Scanner | ✅ Pass | 5.136s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.415s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.077s

---

### pentester (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.136s |  |
| Text Transform Uppercase | ✅ Pass | 1.220s |  |
| Count from 1 to 5 | ✅ Pass | 1.145s |  |
| Math Calculation | ✅ Pass | 1.479s |  |
| Basic Echo Function | ✅ Pass | 3.706s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.375s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.971s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.924s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.257s |  |
| Search Query Function | ✅ Pass | 1.378s |  |
| Ask Advice Function | ✅ Pass | 1.099s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.860s |  |
| Basic Context Memory Test | ✅ Pass | 1.215s |  |
| Function Argument Memory Test | ✅ Pass | 1.230s |  |
| Function Response Memory Test | ✅ Pass | 1.744s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.121s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.327s |  |
| Penetration Testing Methodology | ✅ Pass | 4.906s |  |
| SQL Injection Attack Type | ✅ Pass | 1.344s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.647s |  |
| Penetration Testing Framework | ✅ Pass | 4.079s |  |
| Web Application Security Scanner | ✅ Pass | 4.850s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.625s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.115s

---

