# LLM Agent Testing Report

Generated: Thu, 23 Jul 2026 11:02:56 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 1.297s |
| simple_json | Qwen/Qwen3.6-27B-FP8 | false | 7/7 (100.00%) | 1.093s |
| primary_agent | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 4.874s |
| assistant | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 4.496s |
| generator | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 4.518s |
| refiner | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 5.409s |
| adviser | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 4.690s |
| reflector | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 2.113s |
| searcher | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.494s |
| enricher | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.270s |
| coder | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 4.309s |
| installer | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 4.625s |
| pentester | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 3.453s |

**Total**: 295/295 (100.00%) successful tests
**Overall average latency**: 3.325s

## Detailed Results

### simple (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.817s |  |
| Text Transform Uppercase | ✅ Pass | 0.873s |  |
| Count from 1 to 5 | ✅ Pass | 0.616s |  |
| Math Calculation | ✅ Pass | 0.523s |  |
| Basic Echo Function | ✅ Pass | 0.942s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.512s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.633s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.377s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.342s |  |
| Search Query Function | ✅ Pass | 0.890s |  |
| Ask Advice Function | ✅ Pass | 1.150s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.884s |  |
| Basic Context Memory Test | ✅ Pass | 0.635s |  |
| Function Argument Memory Test | ✅ Pass | 0.581s |  |
| Function Response Memory Test | ✅ Pass | 0.551s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.691s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.600s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.886s |  |
| Penetration Testing Methodology | ✅ Pass | 3.163s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.430s |  |
| SQL Injection Attack Type | ✅ Pass | 0.530s |  |
| Penetration Testing Framework | ✅ Pass | 3.282s |  |
| Web Application Security Scanner | ✅ Pass | 2.106s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.108s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.297s

---

### simple_json (Qwen/Qwen3.6-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 2.003s |  |
| Person Information JSON | ✅ Pass | 0.787s |  |
| Project Information JSON | ✅ Pass | 0.767s |  |
| User Profile JSON | ✅ Pass | 0.904s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.890s |  |
| JSON Array Response Without Schema | ✅ Pass | 1.350s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.944s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 1.093s

---

### primary_agent (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.840s |  |
| Text Transform Uppercase | ✅ Pass | 2.757s |  |
| Count from 1 to 5 | ✅ Pass | 23.297s |  |
| Math Calculation | ✅ Pass | 2.085s |  |
| Basic Echo Function | ✅ Pass | 1.593s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.733s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.092s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.554s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.371s |  |
| Search Query Function | ✅ Pass | 1.165s |  |
| Ask Advice Function | ✅ Pass | 3.085s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.586s |  |
| Basic Context Memory Test | ✅ Pass | 3.608s |  |
| Function Argument Memory Test | ✅ Pass | 1.422s |  |
| Function Response Memory Test | ✅ Pass | 3.631s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.060s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.691s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.489s |  |
| Penetration Testing Methodology | ✅ Pass | 9.435s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.149s |  |
| SQL Injection Attack Type | ✅ Pass | 4.313s |  |
| Penetration Testing Framework | ✅ Pass | 10.616s |  |
| Web Application Security Scanner | ✅ Pass | 7.340s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.046s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 4.874s

---

### assistant (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.507s |  |
| Text Transform Uppercase | ✅ Pass | 1.897s |  |
| Count from 1 to 5 | ✅ Pass | 24.442s |  |
| Math Calculation | ✅ Pass | 1.769s |  |
| Basic Echo Function | ✅ Pass | 1.990s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.866s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.995s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.554s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.256s |  |
| Search Query Function | ✅ Pass | 1.687s |  |
| Ask Advice Function | ✅ Pass | 2.096s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.609s |  |
| Basic Context Memory Test | ✅ Pass | 3.280s |  |
| Function Argument Memory Test | ✅ Pass | 1.519s |  |
| Function Response Memory Test | ✅ Pass | 2.905s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.206s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.167s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 8.111s |  |
| Penetration Testing Methodology | ✅ Pass | 8.559s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.973s |  |
| SQL Injection Attack Type | ✅ Pass | 3.183s |  |
| Penetration Testing Framework | ✅ Pass | 8.294s |  |
| Web Application Security Scanner | ✅ Pass | 5.902s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.132s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 4.496s

---

### generator (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.585s |  |
| Text Transform Uppercase | ✅ Pass | 2.628s |  |
| Count from 1 to 5 | ✅ Pass | 23.202s |  |
| Math Calculation | ✅ Pass | 1.985s |  |
| Basic Echo Function | ✅ Pass | 1.791s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.590s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.357s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.504s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.982s |  |
| Search Query Function | ✅ Pass | 2.255s |  |
| Ask Advice Function | ✅ Pass | 2.174s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.719s |  |
| Basic Context Memory Test | ✅ Pass | 3.138s |  |
| Function Argument Memory Test | ✅ Pass | 1.350s |  |
| Function Response Memory Test | ✅ Pass | 2.353s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 6.086s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.940s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.938s |  |
| Penetration Testing Methodology | ✅ Pass | 10.433s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.994s |  |
| SQL Injection Attack Type | ✅ Pass | 3.726s |  |
| Penetration Testing Framework | ✅ Pass | 7.251s |  |
| Web Application Security Scanner | ✅ Pass | 4.251s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.175s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 4.518s

---

### refiner (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.699s |  |
| Text Transform Uppercase | ✅ Pass | 3.012s |  |
| Count from 1 to 5 | ✅ Pass | 23.392s |  |
| Math Calculation | ✅ Pass | 1.961s |  |
| Basic Echo Function | ✅ Pass | 2.761s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.349s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.090s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.226s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.152s |  |
| Search Query Function | ✅ Pass | 2.804s |  |
| Ask Advice Function | ✅ Pass | 3.465s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.743s |  |
| Basic Context Memory Test | ✅ Pass | 3.283s |  |
| Function Argument Memory Test | ✅ Pass | 1.567s |  |
| Function Response Memory Test | ✅ Pass | 1.550s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.854s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.232s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.511s |  |
| Penetration Testing Methodology | ✅ Pass | 9.651s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.079s |  |
| SQL Injection Attack Type | ✅ Pass | 15.503s |  |
| Penetration Testing Framework | ✅ Pass | 13.215s |  |
| Web Application Security Scanner | ✅ Pass | 5.997s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.713s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 5.409s

---

### adviser (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.787s |  |
| Text Transform Uppercase | ✅ Pass | 2.683s |  |
| Count from 1 to 5 | ✅ Pass | 22.764s |  |
| Math Calculation | ✅ Pass | 2.060s |  |
| Basic Echo Function | ✅ Pass | 2.330s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.256s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.909s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.587s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.519s |  |
| Search Query Function | ✅ Pass | 0.297s |  |
| Ask Advice Function | ✅ Pass | 4.167s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.687s |  |
| Basic Context Memory Test | ✅ Pass | 2.957s |  |
| Function Argument Memory Test | ✅ Pass | 2.440s |  |
| Function Response Memory Test | ✅ Pass | 3.161s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.586s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.017s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.254s |  |
| Penetration Testing Methodology | ✅ Pass | 8.891s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.617s |  |
| SQL Injection Attack Type | ✅ Pass | 3.427s |  |
| Penetration Testing Framework | ✅ Pass | 13.348s |  |
| Web Application Security Scanner | ✅ Pass | 5.351s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.462s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 4.690s

---

### reflector (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.611s |  |
| Text Transform Uppercase | ✅ Pass | 0.609s |  |
| Count from 1 to 5 | ✅ Pass | 20.127s |  |
| Math Calculation | ✅ Pass | 0.550s |  |
| Basic Echo Function | ✅ Pass | 1.434s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.559s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.751s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.474s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.402s |  |
| Search Query Function | ✅ Pass | 0.818s |  |
| Ask Advice Function | ✅ Pass | 1.805s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.035s |  |
| Basic Context Memory Test | ✅ Pass | 0.793s |  |
| Function Argument Memory Test | ✅ Pass | 0.591s |  |
| Function Response Memory Test | ✅ Pass | 1.043s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.845s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.819s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.852s |  |
| Penetration Testing Methodology | ✅ Pass | 2.214s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.514s |  |
| SQL Injection Attack Type | ✅ Pass | 1.162s |  |
| Penetration Testing Framework | ✅ Pass | 1.400s |  |
| Web Application Security Scanner | ✅ Pass | 1.947s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.350s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 2.113s

---

### searcher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.672s |  |
| Text Transform Uppercase | ✅ Pass | 0.208s |  |
| Count from 1 to 5 | ✅ Pass | 0.214s |  |
| Math Calculation | ✅ Pass | 0.217s |  |
| Basic Echo Function | ✅ Pass | 0.215s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.292s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.243s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.475s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.291s |  |
| Search Query Function | ✅ Pass | 0.317s |  |
| Ask Advice Function | ✅ Pass | 0.402s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.328s |  |
| Basic Context Memory Test | ✅ Pass | 0.214s |  |
| Function Argument Memory Test | ✅ Pass | 0.214s |  |
| Function Response Memory Test | ✅ Pass | 0.239s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.262s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.229s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.579s |  |
| Penetration Testing Methodology | ✅ Pass | 5.138s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.221s |  |
| SQL Injection Attack Type | ✅ Pass | 0.212s |  |
| Penetration Testing Framework | ✅ Pass | 0.214s |  |
| Web Application Security Scanner | ✅ Pass | 0.212s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.237s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.494s

---

### enricher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.217s |  |
| Text Transform Uppercase | ✅ Pass | 0.212s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.207s |  |
| Basic Echo Function | ✅ Pass | 0.221s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.274s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.261s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.477s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.253s |  |
| Search Query Function | ✅ Pass | 0.287s |  |
| Ask Advice Function | ✅ Pass | 0.235s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.232s |  |
| Basic Context Memory Test | ✅ Pass | 0.212s |  |
| Function Argument Memory Test | ✅ Pass | 0.288s |  |
| Function Response Memory Test | ✅ Pass | 0.445s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.409s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.313s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.428s |  |
| Penetration Testing Methodology | ✅ Pass | 0.222s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.214s |  |
| SQL Injection Attack Type | ✅ Pass | 0.206s |  |
| Penetration Testing Framework | ✅ Pass | 0.217s |  |
| Web Application Security Scanner | ✅ Pass | 0.210s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.214s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.270s

---

### coder (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.783s |  |
| Text Transform Uppercase | ✅ Pass | 2.497s |  |
| Count from 1 to 5 | ✅ Pass | 22.545s |  |
| Math Calculation | ✅ Pass | 2.090s |  |
| Basic Echo Function | ✅ Pass | 2.165s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.531s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.619s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.314s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.965s |  |
| Search Query Function | ✅ Pass | 1.762s |  |
| Ask Advice Function | ✅ Pass | 2.114s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.459s |  |
| Basic Context Memory Test | ✅ Pass | 3.015s |  |
| Function Argument Memory Test | ✅ Pass | 1.983s |  |
| Function Response Memory Test | ✅ Pass | 1.850s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.977s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 6.159s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.949s |  |
| Penetration Testing Methodology | ✅ Pass | 8.198s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.722s |  |
| SQL Injection Attack Type | ✅ Pass | 4.845s |  |
| Penetration Testing Framework | ✅ Pass | 7.508s |  |
| Web Application Security Scanner | ✅ Pass | 4.216s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.137s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 4.309s

---

### installer (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.641s |  |
| Text Transform Uppercase | ✅ Pass | 2.646s |  |
| Count from 1 to 5 | ✅ Pass | 22.984s |  |
| Math Calculation | ✅ Pass | 2.059s |  |
| Basic Echo Function | ✅ Pass | 1.856s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.449s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.693s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.316s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.602s |  |
| Search Query Function | ✅ Pass | 1.596s |  |
| Ask Advice Function | ✅ Pass | 2.780s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.641s |  |
| Basic Context Memory Test | ✅ Pass | 2.458s |  |
| Function Argument Memory Test | ✅ Pass | 1.962s |  |
| Function Response Memory Test | ✅ Pass | 1.824s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.969s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.246s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 5.041s |  |
| Penetration Testing Methodology | ✅ Pass | 16.871s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.280s |  |
| SQL Injection Attack Type | ✅ Pass | 3.475s |  |
| Penetration Testing Framework | ✅ Pass | 8.601s |  |
| Web Application Security Scanner | ✅ Pass | 5.177s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.830s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 4.625s

---

### pentester (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.630s |  |
| Text Transform Uppercase | ✅ Pass | 2.407s |  |
| Count from 1 to 5 | ✅ Pass | 3.551s |  |
| Math Calculation | ✅ Pass | 1.991s |  |
| Basic Echo Function | ✅ Pass | 1.816s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.833s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.556s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.526s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.596s |  |
| Search Query Function | ✅ Pass | 1.433s |  |
| Ask Advice Function | ✅ Pass | 2.162s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.209s |  |
| Basic Context Memory Test | ✅ Pass | 3.880s |  |
| Function Argument Memory Test | ✅ Pass | 1.940s |  |
| Function Response Memory Test | ✅ Pass | 1.501s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.826s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.473s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 4.081s |  |
| Penetration Testing Methodology | ✅ Pass | 11.146s |  |
| Vulnerability Assessment Tools | ✅ Pass | 7.677s |  |
| SQL Injection Attack Type | ✅ Pass | 3.017s |  |
| Penetration Testing Framework | ✅ Pass | 7.160s |  |
| Web Application Security Scanner | ✅ Pass | 5.180s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.267s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 3.453s

---

