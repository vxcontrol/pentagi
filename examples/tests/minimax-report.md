# LLM Agent Testing Report

Generated: Tue, 21 Jul 2026 19:11:05 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | MiniMax-M2.7 | true | 23/23 (100.00%) | 3.034s |
| simple_json | MiniMax-M2.7 | true | 6/7 (85.71%) | 2.505s |
| primary_agent | MiniMax-M3 | true | 23/23 (100.00%) | 1.784s |
| assistant | MiniMax-M3 | true | 22/23 (95.65%) | 1.863s |
| generator | MiniMax-M3 | true | 22/23 (95.65%) | 1.774s |
| refiner | MiniMax-M3 | true | 22/23 (95.65%) | 1.857s |
| adviser | MiniMax-M3 | true | 22/23 (95.65%) | 1.618s |
| reflector | MiniMax-M3 | true | 23/23 (100.00%) | 3.042s |
| searcher | MiniMax-M3 | true | 23/23 (100.00%) | 2.713s |
| enricher | MiniMax-M3 | true | 23/23 (100.00%) | 2.287s |
| coder | MiniMax-M3 | true | 22/23 (95.65%) | 1.399s |
| installer | MiniMax-M3 | true | 23/23 (100.00%) | 0.224s |
| pentester | MiniMax-M3 | true | 23/23 (100.00%) | 0.222s |

**Total**: 277/283 (97.88%) successful tests
**Overall average latency**: 1.835s

## Detailed Results

### simple (MiniMax-M2.7)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.424s |  |
| Math Calculation | ✅ Pass | 1.493s |  |
| Count from 1 to 5 | ✅ Pass | 2.932s |  |
| Text Transform Uppercase | ✅ Pass | 5.049s |  |
| Basic Echo Function | ✅ Pass | 1.800s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.018s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.595s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.591s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.775s |  |
| Search Query Function | ✅ Pass | 1.895s |  |
| Ask Advice Function | ✅ Pass | 2.342s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.684s |  |
| Basic Context Memory Test | ✅ Pass | 2.271s |  |
| Function Argument Memory Test | ✅ Pass | 2.541s |  |
| Function Response Memory Test | ✅ Pass | 2.198s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.435s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.963s |  |
| Penetration Testing Methodology | ✅ Pass | 4.513s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.186s |  |
| SQL Injection Attack Type | ✅ Pass | 4.290s |  |
| Penetration Testing Framework | ✅ Pass | 6.556s |  |
| Web Application Security Scanner | ✅ Pass | 2.842s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.374s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.034s

---

### simple_json (MiniMax-M2.7)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 2.120s |  |
| Person Information JSON | ✅ Pass | 2.365s |  |
| User Profile JSON | ✅ Pass | 2.581s |  |
| Vulnerability Report Memory Test | ✅ Pass | 2.818s |  |
| Streaming Person Information JSON Streaming | ❌ Fail | 2.079s | invalid JSON response: invalid character '<' after top\-level value |
| JSON Array Response Without Schema | ✅ Pass | 2.834s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 2.732s |  |

**Summary**: 6/7 (85.71%) successful tests

**Average latency**: 2.505s

---

### primary_agent (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.002s |  |
| Text Transform Uppercase | ✅ Pass | 0.225s |  |
| Count from 1 to 5 | ✅ Pass | 0.207s |  |
| Math Calculation | ✅ Pass | 0.211s |  |
| Basic Echo Function | ✅ Pass | 0.211s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.205s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.265s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.278s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.211s |  |
| Search Query Function | ✅ Pass | 0.202s |  |
| Ask Advice Function | ✅ Pass | 0.208s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.257s |  |
| Basic Context Memory Test | ✅ Pass | 0.211s |  |
| Function Argument Memory Test | ✅ Pass | 1.470s |  |
| Function Response Memory Test | ✅ Pass | 1.604s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.260s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.735s |  |
| Penetration Testing Methodology | ✅ Pass | 4.491s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.840s |  |
| SQL Injection Attack Type | ✅ Pass | 3.323s |  |
| Penetration Testing Framework | ✅ Pass | 7.383s |  |
| Web Application Security Scanner | ✅ Pass | 1.496s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.715s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.784s

---

### assistant (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.002s |  |
| Text Transform Uppercase | ✅ Pass | 0.208s |  |
| Count from 1 to 5 | ✅ Pass | 0.207s |  |
| Math Calculation | ✅ Pass | 0.211s |  |
| Basic Echo Function | ✅ Pass | 0.207s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.207s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.261s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.273s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.217s |  |
| Search Query Function | ✅ Pass | 0.207s |  |
| Ask Advice Function | ✅ Pass | 0.207s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.257s |  |
| Basic Context Memory Test | ✅ Pass | 0.211s |  |
| Function Argument Memory Test | ✅ Pass | 1.350s |  |
| Function Response Memory Test | ✅ Pass | 1.193s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.322s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.172s |  |
| Penetration Testing Methodology | ✅ Pass | 5.842s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.280s |  |
| SQL Injection Attack Type | ❌ Fail | 4.207s | expected text 'injection' not found |
| Penetration Testing Framework | ✅ Pass | 7.334s |  |
| Web Application Security Scanner | ✅ Pass | 4.036s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.427s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.863s

---

### generator (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.022s |  |
| Text Transform Uppercase | ✅ Pass | 0.215s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.214s |  |
| Basic Echo Function | ✅ Pass | 0.208s |  |
| Streaming Simple Math Streaming | ❌ Fail | 0.207s | expected text '6' not found |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.262s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.261s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.213s |  |
| Search Query Function | ✅ Pass | 0.204s |  |
| Ask Advice Function | ✅ Pass | 0.209s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.256s |  |
| Basic Context Memory Test | ✅ Pass | 0.213s |  |
| Function Argument Memory Test | ✅ Pass | 1.599s |  |
| Function Response Memory Test | ✅ Pass | 1.137s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.304s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.192s |  |
| Penetration Testing Methodology | ✅ Pass | 4.910s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.593s |  |
| SQL Injection Attack Type | ✅ Pass | 2.922s |  |
| Penetration Testing Framework | ✅ Pass | 5.884s |  |
| Web Application Security Scanner | ✅ Pass | 2.006s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.560s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.774s

---

### refiner (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.000s |  |
| Text Transform Uppercase | ✅ Pass | 0.207s |  |
| Count from 1 to 5 | ✅ Pass | 0.208s |  |
| Math Calculation | ✅ Pass | 0.211s |  |
| Basic Echo Function | ✅ Pass | 0.206s |  |
| Streaming Simple Math Streaming | ❌ Fail | 0.208s | expected text '6' not found |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.261s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.260s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.211s |  |
| Search Query Function | ✅ Pass | 0.207s |  |
| Ask Advice Function | ✅ Pass | 0.208s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.258s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 1.150s |  |
| Function Response Memory Test | ✅ Pass | 1.505s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.291s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.101s |  |
| Penetration Testing Methodology | ✅ Pass | 4.023s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.803s |  |
| SQL Injection Attack Type | ✅ Pass | 2.038s |  |
| Penetration Testing Framework | ✅ Pass | 6.667s |  |
| Web Application Security Scanner | ✅ Pass | 3.783s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.675s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.857s

---

### adviser (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.002s |  |
| Text Transform Uppercase | ✅ Pass | 0.211s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.218s |  |
| Basic Echo Function | ✅ Pass | 0.217s |  |
| Streaming Simple Math Streaming | ❌ Fail | 0.207s | expected text '6' not found |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.261s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.262s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.210s |  |
| Search Query Function | ✅ Pass | 0.207s |  |
| Ask Advice Function | ✅ Pass | 0.206s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.258s |  |
| Basic Context Memory Test | ✅ Pass | 0.204s |  |
| Function Argument Memory Test | ✅ Pass | 1.427s |  |
| Function Response Memory Test | ✅ Pass | 0.980s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.853s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.013s |  |
| Penetration Testing Methodology | ✅ Pass | 2.843s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.423s |  |
| SQL Injection Attack Type | ✅ Pass | 1.870s |  |
| Penetration Testing Framework | ✅ Pass | 7.770s |  |
| Web Application Security Scanner | ✅ Pass | 1.755s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.600s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.618s

---

### reflector (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.207s |  |
| Text Transform Uppercase | ✅ Pass | 1.806s |  |
| Count from 1 to 5 | ✅ Pass | 1.716s |  |
| Math Calculation | ✅ Pass | 1.479s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.874s |  |
| Basic Echo Function | ✅ Pass | 2.816s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.577s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.428s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.800s |  |
| Ask Advice Function | ✅ Pass | 1.774s |  |
| Search Query Function | ✅ Pass | 3.854s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.568s |  |
| Basic Context Memory Test | ✅ Pass | 1.478s |  |
| Function Argument Memory Test | ✅ Pass | 1.394s |  |
| Function Response Memory Test | ✅ Pass | 2.723s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.058s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.080s |  |
| Penetration Testing Methodology | ✅ Pass | 4.896s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.835s |  |
| SQL Injection Attack Type | ✅ Pass | 2.803s |  |
| Penetration Testing Framework | ✅ Pass | 7.024s |  |
| Web Application Security Scanner | ✅ Pass | 2.798s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.965s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 3.042s

---

### searcher (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.522s |  |
| Text Transform Uppercase | ✅ Pass | 1.797s |  |
| Count from 1 to 5 | ✅ Pass | 1.528s |  |
| Math Calculation | ✅ Pass | 1.399s |  |
| Basic Echo Function | ✅ Pass | 2.405s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.506s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.981s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.416s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.665s |  |
| Ask Advice Function | ✅ Pass | 1.628s |  |
| Search Query Function | ✅ Pass | 3.944s |  |
| Basic Context Memory Test | ✅ Pass | 1.394s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.683s |  |
| Function Argument Memory Test | ✅ Pass | 1.794s |  |
| Function Response Memory Test | ✅ Pass | 1.223s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.969s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.941s |  |
| Penetration Testing Methodology | ✅ Pass | 3.698s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.506s |  |
| SQL Injection Attack Type | ✅ Pass | 2.416s |  |
| Penetration Testing Framework | ✅ Pass | 6.985s |  |
| Web Application Security Scanner | ✅ Pass | 2.666s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.328s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.713s

---

### enricher (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.371s |  |
| Text Transform Uppercase | ✅ Pass | 2.029s |  |
| Count from 1 to 5 | ✅ Pass | 1.483s |  |
| Math Calculation | ✅ Pass | 1.719s |  |
| Basic Echo Function | ✅ Pass | 1.897s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.594s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.228s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.384s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.659s |  |
| Ask Advice Function | ✅ Pass | 1.517s |  |
| Search Query Function | ✅ Pass | 2.571s |  |
| Basic Context Memory Test | ✅ Pass | 1.365s |  |
| Function Argument Memory Test | ✅ Pass | 0.254s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.597s |  |
| Function Response Memory Test | ✅ Pass | 1.071s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.096s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.652s |  |
| Penetration Testing Methodology | ✅ Pass | 2.176s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.249s |  |
| SQL Injection Attack Type | ✅ Pass | 3.192s |  |
| Penetration Testing Framework | ✅ Pass | 3.008s |  |
| Web Application Security Scanner | ✅ Pass | 2.955s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.519s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.287s

---

### coder (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.237s |  |
| Text Transform Uppercase | ✅ Pass | 0.219s |  |
| Count from 1 to 5 | ✅ Pass | 0.209s |  |
| Math Calculation | ✅ Pass | 0.213s |  |
| Basic Echo Function | ✅ Pass | 0.216s |  |
| Streaming Simple Math Streaming | ❌ Fail | 0.211s | expected text '6' not found |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.255s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.257s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.204s |  |
| Search Query Function | ✅ Pass | 0.208s |  |
| Ask Advice Function | ✅ Pass | 0.208s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.255s |  |
| Basic Context Memory Test | ✅ Pass | 0.214s |  |
| Function Argument Memory Test | ✅ Pass | 0.250s |  |
| Function Response Memory Test | ✅ Pass | 0.838s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.135s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.314s |  |
| Penetration Testing Methodology | ✅ Pass | 3.940s |  |
| SQL Injection Attack Type | ✅ Pass | 0.209s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.416s |  |
| Penetration Testing Framework | ✅ Pass | 3.656s |  |
| Web Application Security Scanner | ✅ Pass | 2.992s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.512s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.399s

---

### installer (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.238s |  |
| Text Transform Uppercase | ✅ Pass | 0.208s |  |
| Count from 1 to 5 | ✅ Pass | 0.204s |  |
| Math Calculation | ✅ Pass | 0.208s |  |
| Basic Echo Function | ✅ Pass | 0.222s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.212s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.260s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.267s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.213s |  |
| Search Query Function | ✅ Pass | 0.209s |  |
| Ask Advice Function | ✅ Pass | 0.219s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.260s |  |
| Basic Context Memory Test | ✅ Pass | 0.221s |  |
| Function Argument Memory Test | ✅ Pass | 0.244s |  |
| Function Response Memory Test | ✅ Pass | 0.203s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.216s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.201s |  |
| Penetration Testing Methodology | ✅ Pass | 0.207s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.290s |  |
| SQL Injection Attack Type | ✅ Pass | 0.214s |  |
| Penetration Testing Framework | ✅ Pass | 0.212s |  |
| Web Application Security Scanner | ✅ Pass | 0.212s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.209s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.224s

---

### pentester (MiniMax-M3)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.241s |  |
| Text Transform Uppercase | ✅ Pass | 0.209s |  |
| Count from 1 to 5 | ✅ Pass | 0.212s |  |
| Math Calculation | ✅ Pass | 0.214s |  |
| Basic Echo Function | ✅ Pass | 0.202s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.212s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.251s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.272s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.211s |  |
| Search Query Function | ✅ Pass | 0.218s |  |
| Ask Advice Function | ✅ Pass | 0.212s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.262s |  |
| Basic Context Memory Test | ✅ Pass | 0.217s |  |
| Function Argument Memory Test | ✅ Pass | 0.201s |  |
| Function Response Memory Test | ✅ Pass | 0.257s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.211s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.205s |  |
| Penetration Testing Methodology | ✅ Pass | 0.208s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.215s |  |
| SQL Injection Attack Type | ✅ Pass | 0.212s |  |
| Penetration Testing Framework | ✅ Pass | 0.210s |  |
| Web Application Security Scanner | ✅ Pass | 0.223s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.226s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.222s

---

