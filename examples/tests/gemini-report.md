# LLM Agent Testing Report

Generated: Mon, 14 Jul 2025 13:00:24 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gemini-2.0-flash-lite | false | 18/18 (100.00%) | 0.745s |
| simple_json | gemini-2.0-flash-lite | false | 4/4 (100.00%) | 0.718s |
| primary_agent | gemini-2.5-flash-lite-preview-06-17 | true | 18/18 (100.00%) | 1.786s |
| assistant | gemini-2.5-flash | true | 18/18 (100.00%) | 2.584s |
| generator | gemini-2.5-pro | true | 17/18 (94.44%) | 6.679s |
| refiner | gemini-2.0-flash | false | 18/18 (100.00%) | 0.650s |
| adviser | gemini-2.5-flash | true | 18/18 (100.00%) | 2.625s |
| reflector | gemini-2.5-flash | true | 18/18 (100.00%) | 2.387s |
| searcher | gemini-2.0-flash | false | 18/18 (100.00%) | 0.718s |
| enricher | gemini-2.0-flash | false | 18/18 (100.00%) | 0.655s |
| coder | gemini-2.5-pro | true | 15/18 (83.33%) | 7.145s |
| installer | gemini-2.5-flash-lite-preview-06-17 | true | 18/18 (100.00%) | 1.775s |
| pentester | gemini-2.5-flash-lite-preview-06-17 | true | 18/18 (100.00%) | 1.719s |

**Total**: 216/220 (98.18%) successful tests
**Overall average latency**: 2.424s

## Detailed Results

### simple (gemini-2.0-flash-lite)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.587s |  |
| Text Transform Uppercase | ✅ Pass | 0.596s |  |
| Count from 1 to 5 | ✅ Pass | 0.557s |  |
| Math Calculation | ✅ Pass | 0.629s |  |
| Basic Echo Function | ✅ Pass | 0.768s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.564s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.593s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.646s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.690s |  |
| Search Query Function | ✅ Pass | 0.734s |  |
| Ask Advice Function | ✅ Pass | 0.735s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.617s |  |
| Penetration Testing Methodology | ✅ Pass | 0.496s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.669s |  |
| SQL Injection Attack Type | ✅ Pass | 0.611s |  |
| Penetration Testing Framework | ✅ Pass | 0.532s |  |
| Web Application Security Scanner | ✅ Pass | 0.619s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.761s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.745s

---

### simple_json (gemini-2.0-flash-lite)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.732s |  |
| Project Information JSON | ✅ Pass | 0.659s |  |
| User Profile JSON | ✅ Pass | 0.674s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.807s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 0.718s

---

### primary_agent (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.296s |  |
| Text Transform Uppercase | ✅ Pass | 1.334s |  |
| Count from 1 to 5 | ✅ Pass | 1.678s |  |
| Math Calculation | ✅ Pass | 1.345s |  |
| Basic Echo Function | ✅ Pass | 1.323s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.176s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.349s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.338s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.534s |  |
| Search Query Function | ✅ Pass | 1.276s |  |
| Ask Advice Function | ✅ Pass | 1.401s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.226s |  |
| Penetration Testing Methodology | ✅ Pass | 1.543s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.191s |  |
| SQL Injection Attack Type | ✅ Pass | 1.400s |  |
| Penetration Testing Framework | ✅ Pass | 2.476s |  |
| Web Application Security Scanner | ✅ Pass | 2.746s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.514s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.786s

---

### assistant (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.507s |  |
| Text Transform Uppercase | ✅ Pass | 1.943s |  |
| Count from 1 to 5 | ✅ Pass | 1.492s |  |
| Math Calculation | ✅ Pass | 1.464s |  |
| Basic Echo Function | ✅ Pass | 1.769s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.497s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.508s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.433s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.521s |  |
| Search Query Function | ✅ Pass | 1.862s |  |
| Ask Advice Function | ✅ Pass | 1.759s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.510s |  |
| Penetration Testing Methodology | ✅ Pass | 1.739s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.602s |  |
| SQL Injection Attack Type | ✅ Pass | 6.575s |  |
| Penetration Testing Framework | ✅ Pass | 6.990s |  |
| Web Application Security Scanner | ✅ Pass | 2.875s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.466s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.584s

---

### generator (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.086s |  |
| Text Transform Uppercase | ✅ Pass | 5.375s |  |
| Count from 1 to 5 | ✅ Pass | 6.069s |  |
| Math Calculation | ✅ Pass | 5.985s |  |
| Basic Echo Function | ✅ Pass | 6.785s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.923s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.353s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 5.770s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.166s |  |
| Search Query Function | ✅ Pass | 4.078s |  |
| Ask Advice Function | ✅ Pass | 3.322s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.612s |  |
| Penetration Testing Methodology | ❌ Fail | 9.254s | expected text 'reconnaissance' not found |
| Vulnerability Assessment Tools | ✅ Pass | 15.114s |  |
| SQL Injection Attack Type | ✅ Pass | 6.687s |  |
| Penetration Testing Framework | ✅ Pass | 12.817s |  |
| Web Application Security Scanner | ✅ Pass | 13.566s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.253s |  |

**Summary**: 17/18 (94.44%) successful tests

**Average latency**: 6.679s

---

### refiner (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.541s |  |
| Text Transform Uppercase | ✅ Pass | 0.530s |  |
| Count from 1 to 5 | ✅ Pass | 0.525s |  |
| Math Calculation | ✅ Pass | 0.459s |  |
| Basic Echo Function | ✅ Pass | 0.584s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.453s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.476s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.612s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.737s |  |
| Search Query Function | ✅ Pass | 0.641s |  |
| Ask Advice Function | ✅ Pass | 0.767s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.785s |  |
| Penetration Testing Methodology | ✅ Pass | 0.510s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.702s |  |
| SQL Injection Attack Type | ✅ Pass | 0.510s |  |
| Penetration Testing Framework | ✅ Pass | 0.546s |  |
| Web Application Security Scanner | ✅ Pass | 0.684s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.625s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.650s

---

### adviser (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.566s |  |
| Text Transform Uppercase | ✅ Pass | 1.571s |  |
| Count from 1 to 5 | ✅ Pass | 1.673s |  |
| Math Calculation | ✅ Pass | 1.211s |  |
| Basic Echo Function | ✅ Pass | 2.036s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.193s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.414s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.510s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.551s |  |
| Search Query Function | ✅ Pass | 1.879s |  |
| Ask Advice Function | ✅ Pass | 1.869s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.400s |  |
| Penetration Testing Methodology | ✅ Pass | 2.635s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.017s |  |
| SQL Injection Attack Type | ✅ Pass | 6.949s |  |
| Penetration Testing Framework | ✅ Pass | 6.675s |  |
| Web Application Security Scanner | ✅ Pass | 3.474s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.620s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.625s

---

### reflector (gemini-2.5-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.457s |  |
| Text Transform Uppercase | ✅ Pass | 1.516s |  |
| Count from 1 to 5 | ✅ Pass | 1.391s |  |
| Math Calculation | ✅ Pass | 1.199s |  |
| Basic Echo Function | ✅ Pass | 1.834s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.586s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.355s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.385s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.718s |  |
| Search Query Function | ✅ Pass | 1.668s |  |
| Ask Advice Function | ✅ Pass | 1.760s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.472s |  |
| Penetration Testing Methodology | ✅ Pass | 2.242s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.719s |  |
| SQL Injection Attack Type | ✅ Pass | 3.863s |  |
| Penetration Testing Framework | ✅ Pass | 6.081s |  |
| Web Application Security Scanner | ✅ Pass | 4.535s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.182s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.387s

---

### searcher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.624s |  |
| Text Transform Uppercase | ✅ Pass | 0.620s |  |
| Count from 1 to 5 | ✅ Pass | 0.450s |  |
| Math Calculation | ✅ Pass | 0.633s |  |
| Basic Echo Function | ✅ Pass | 0.650s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.481s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.513s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.606s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.837s |  |
| Search Query Function | ✅ Pass | 0.614s |  |
| Ask Advice Function | ✅ Pass | 0.594s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.679s |  |
| Penetration Testing Methodology | ✅ Pass | 0.492s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.055s |  |
| SQL Injection Attack Type | ✅ Pass | 0.422s |  |
| Penetration Testing Framework | ✅ Pass | 0.878s |  |
| Web Application Security Scanner | ✅ Pass | 1.103s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.668s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.718s

---

### enricher (gemini-2.0-flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.450s |  |
| Text Transform Uppercase | ✅ Pass | 0.589s |  |
| Count from 1 to 5 | ✅ Pass | 0.742s |  |
| Math Calculation | ✅ Pass | 0.510s |  |
| Basic Echo Function | ✅ Pass | 0.603s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.409s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.414s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.634s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.727s |  |
| Search Query Function | ✅ Pass | 0.586s |  |
| Ask Advice Function | ✅ Pass | 0.713s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.773s |  |
| Penetration Testing Methodology | ✅ Pass | 0.479s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.703s |  |
| SQL Injection Attack Type | ✅ Pass | 0.404s |  |
| Penetration Testing Framework | ✅ Pass | 0.530s |  |
| Web Application Security Scanner | ✅ Pass | 0.856s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.661s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.655s

---

### coder (gemini-2.5-pro)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.092s |  |
| Text Transform Uppercase | ✅ Pass | 5.324s |  |
| Count from 1 to 5 | ✅ Pass | 6.242s |  |
| Math Calculation | ✅ Pass | 3.073s |  |
| Basic Echo Function | ✅ Pass | 3.908s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.346s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.459s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 6.040s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.247s |  |
| Search Query Function | ✅ Pass | 5.285s |  |
| Ask Advice Function | ✅ Pass | 3.976s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.065s |  |
| Penetration Testing Methodology | ❌ Fail | 12.625s | expected text 'reconnaissance' not found |
| Vulnerability Assessment Tools | ✅ Pass | 13.618s |  |
| SQL Injection Attack Type | ✅ Pass | 8.907s |  |
| Penetration Testing Framework | ❌ Fail | 11.341s | expected text 'exploitation' not found |
| Penetration Testing Tool Selection | ✅ Pass | 3.993s |  |
| Web Application Security Scanner | ❌ Fail | 24.062s | expected text 'web' not found |

**Summary**: 15/18 (83.33%) successful tests

**Average latency**: 7.145s

---

### installer (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.304s |  |
| Text Transform Uppercase | ✅ Pass | 1.444s |  |
| Count from 1 to 5 | ✅ Pass | 1.577s |  |
| Math Calculation | ✅ Pass | 1.055s |  |
| Basic Echo Function | ✅ Pass | 1.589s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.230s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.302s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.182s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.263s |  |
| Search Query Function | ✅ Pass | 1.535s |  |
| Ask Advice Function | ✅ Pass | 1.294s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.358s |  |
| Penetration Testing Methodology | ✅ Pass | 1.598s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.300s |  |
| SQL Injection Attack Type | ✅ Pass | 1.361s |  |
| Penetration Testing Framework | ✅ Pass | 3.316s |  |
| Web Application Security Scanner | ✅ Pass | 2.757s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.471s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.775s

---

### pentester (gemini-2.5-flash-lite-preview-06-17)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.385s |  |
| Text Transform Uppercase | ✅ Pass | 1.382s |  |
| Count from 1 to 5 | ✅ Pass | 1.213s |  |
| Math Calculation | ✅ Pass | 1.355s |  |
| Basic Echo Function | ✅ Pass | 1.771s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.029s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.119s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.461s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.498s |  |
| Search Query Function | ✅ Pass | 1.365s |  |
| Ask Advice Function | ✅ Pass | 1.436s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.196s |  |
| Penetration Testing Methodology | ✅ Pass | 1.331s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.293s |  |
| SQL Injection Attack Type | ✅ Pass | 1.377s |  |
| Penetration Testing Framework | ✅ Pass | 1.973s |  |
| Web Application Security Scanner | ✅ Pass | 1.988s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.758s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.719s

---

