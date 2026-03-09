# LLM Agent Testing Report

Generated: Wed, 04 Mar 2026 22:36:05 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | kimi-k2-turbo-preview | false | 23/23 (100.00%) | 1.029s |
| simple_json | kimi-k2-turbo-preview | false | 5/5 (100.00%) | 1.090s |
| primary_agent | kimi-k2.5 | true | 23/23 (100.00%) | 4.379s |
| assistant | kimi-k2.5 | true | 23/23 (100.00%) | 4.599s |
| generator | kimi-k2.5 | true | 23/23 (100.00%) | 4.054s |
| refiner | kimi-k2.5 | true | 23/23 (100.00%) | 4.773s |
| adviser | kimi-k2.5 | true | 23/23 (100.00%) | 4.786s |
| reflector | kimi-k2-0905-preview | true | 23/23 (100.00%) | 2.573s |
| searcher | kimi-k2-0905-preview | true | 22/23 (95.65%) | 2.907s |
| enricher | kimi-k2-0905-preview | true | 23/23 (100.00%) | 2.275s |
| coder | kimi-k2.5 | true | 23/23 (100.00%) | 4.206s |
| installer | kimi-k2-turbo-preview | true | 23/23 (100.00%) | 0.918s |
| pentester | kimi-k2-turbo-preview | true | 23/23 (100.00%) | 0.901s |

**Total**: 280/281 (99.64%) successful tests
**Overall average latency**: 3.081s

## Detailed Results

### simple (kimi-k2-turbo-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.911s |  |
| Text Transform Uppercase | ✅ Pass | 0.743s |  |
| Count from 1 to 5 | ✅ Pass | 0.800s |  |
| Math Calculation | ✅ Pass | 0.691s |  |
| Basic Echo Function | ✅ Pass | 0.943s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.763s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.610s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.095s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.258s |  |
| Search Query Function | ✅ Pass | 0.699s |  |
| Ask Advice Function | ✅ Pass | 0.925s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.914s |  |
| Basic Context Memory Test | ✅ Pass | 0.908s |  |
| Function Argument Memory Test | ✅ Pass | 0.770s |  |
| Function Response Memory Test | ✅ Pass | 0.750s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.579s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.014s |  |
| Penetration Testing Methodology | ✅ Pass | 0.932s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.830s |  |
| SQL Injection Attack Type | ✅ Pass | 0.853s |  |
| Penetration Testing Framework | ✅ Pass | 1.045s |  |
| Web Application Security Scanner | ✅ Pass | 0.615s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.001s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.029s

---

### simple_json (kimi-k2-turbo-preview)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 0.962s |  |
| Project Information JSON | ✅ Pass | 0.950s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.834s |  |
| User Profile JSON | ✅ Pass | 1.014s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.687s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.090s

---

### primary_agent (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.913s |  |
| Text Transform Uppercase | ✅ Pass | 2.090s |  |
| Count from 1 to 5 | ✅ Pass | 4.081s |  |
| Math Calculation | ✅ Pass | 1.913s |  |
| Basic Echo Function | ✅ Pass | 1.683s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.689s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.210s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.376s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.979s |  |
| Search Query Function | ✅ Pass | 1.847s |  |
| Ask Advice Function | ✅ Pass | 3.195s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.966s |  |
| Basic Context Memory Test | ✅ Pass | 2.056s |  |
| Function Argument Memory Test | ✅ Pass | 3.404s |  |
| Function Response Memory Test | ✅ Pass | 2.744s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.534s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.884s |  |
| Penetration Testing Methodology | ✅ Pass | 13.876s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.282s |  |
| SQL Injection Attack Type | ✅ Pass | 3.591s |  |
| Penetration Testing Framework | ✅ Pass | 13.475s |  |
| Web Application Security Scanner | ✅ Pass | 10.430s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.495s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.379s

---

### assistant (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.791s |  |
| Text Transform Uppercase | ✅ Pass | 1.895s |  |
| Count from 1 to 5 | ✅ Pass | 2.390s |  |
| Math Calculation | ✅ Pass | 1.880s |  |
| Basic Echo Function | ✅ Pass | 2.028s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.733s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.070s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.019s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.252s |  |
| Search Query Function | ✅ Pass | 1.543s |  |
| Ask Advice Function | ✅ Pass | 2.579s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.891s |  |
| Basic Context Memory Test | ✅ Pass | 3.971s |  |
| Function Argument Memory Test | ✅ Pass | 3.501s |  |
| Function Response Memory Test | ✅ Pass | 2.208s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 4.784s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 4.306s |  |
| Vulnerability Assessment Tools | ✅ Pass | 10.359s |  |
| Penetration Testing Methodology | ✅ Pass | 17.251s |  |
| SQL Injection Attack Type | ✅ Pass | 2.463s |  |
| Penetration Testing Framework | ✅ Pass | 11.586s |  |
| Web Application Security Scanner | ✅ Pass | 13.252s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.006s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.599s

---

### generator (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.203s |  |
| Text Transform Uppercase | ✅ Pass | 1.889s |  |
| Count from 1 to 5 | ✅ Pass | 2.280s |  |
| Math Calculation | ✅ Pass | 1.852s |  |
| Basic Echo Function | ✅ Pass | 1.704s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.734s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.883s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.803s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.665s |  |
| Search Query Function | ✅ Pass | 1.833s |  |
| Ask Advice Function | ✅ Pass | 1.824s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.725s |  |
| Basic Context Memory Test | ✅ Pass | 2.329s |  |
| Function Argument Memory Test | ✅ Pass | 2.637s |  |
| Function Response Memory Test | ✅ Pass | 2.065s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.415s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.119s |  |
| Penetration Testing Methodology | ✅ Pass | 12.633s |  |
| SQL Injection Attack Type | ✅ Pass | 3.027s |  |
| Vulnerability Assessment Tools | ✅ Pass | 13.788s |  |
| Penetration Testing Framework | ✅ Pass | 11.570s |  |
| Web Application Security Scanner | ✅ Pass | 8.272s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.982s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.054s

---

### refiner (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.172s |  |
| Text Transform Uppercase | ✅ Pass | 3.542s |  |
| Count from 1 to 5 | ✅ Pass | 3.905s |  |
| Math Calculation | ✅ Pass | 2.205s |  |
| Basic Echo Function | ✅ Pass | 1.896s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.793s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.588s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.001s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.247s |  |
| Search Query Function | ✅ Pass | 1.763s |  |
| Ask Advice Function | ✅ Pass | 2.343s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.959s |  |
| Basic Context Memory Test | ✅ Pass | 2.718s |  |
| Function Argument Memory Test | ✅ Pass | 2.372s |  |
| Function Response Memory Test | ✅ Pass | 2.732s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 5.729s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.771s |  |
| Penetration Testing Methodology | ✅ Pass | 14.859s |  |
| Vulnerability Assessment Tools | ✅ Pass | 11.455s |  |
| SQL Injection Attack Type | ✅ Pass | 7.561s |  |
| Penetration Testing Framework | ✅ Pass | 11.828s |  |
| Web Application Security Scanner | ✅ Pass | 10.862s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.473s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.773s

---

### adviser (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.235s |  |
| Text Transform Uppercase | ✅ Pass | 2.341s |  |
| Count from 1 to 5 | ✅ Pass | 3.830s |  |
| Math Calculation | ✅ Pass | 2.093s |  |
| Basic Echo Function | ✅ Pass | 1.943s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.038s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.933s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.271s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.077s |  |
| Search Query Function | ✅ Pass | 1.605s |  |
| Ask Advice Function | ✅ Pass | 2.604s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.891s |  |
| Basic Context Memory Test | ✅ Pass | 3.162s |  |
| Function Argument Memory Test | ✅ Pass | 2.505s |  |
| Function Response Memory Test | ✅ Pass | 2.594s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.564s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.689s |  |
| Penetration Testing Methodology | ✅ Pass | 18.499s |  |
| SQL Injection Attack Type | ✅ Pass | 3.531s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.883s |  |
| Penetration Testing Framework | ✅ Pass | 16.455s |  |
| Web Application Security Scanner | ✅ Pass | 9.418s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.912s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.786s

---

### reflector (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.417s |  |
| Text Transform Uppercase | ✅ Pass | 1.378s |  |
| Count from 1 to 5 | ✅ Pass | 2.084s |  |
| Math Calculation | ✅ Pass | 1.037s |  |
| Basic Echo Function | ✅ Pass | 2.442s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.130s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.451s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.794s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.659s |  |
| Search Query Function | ✅ Pass | 2.429s |  |
| Ask Advice Function | ✅ Pass | 3.548s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.580s |  |
| Basic Context Memory Test | ✅ Pass | 1.911s |  |
| Function Argument Memory Test | ✅ Pass | 1.232s |  |
| Function Response Memory Test | ✅ Pass | 1.283s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 8.395s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.562s |  |
| Penetration Testing Methodology | ✅ Pass | 3.007s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.142s |  |
| SQL Injection Attack Type | ✅ Pass | 3.116s |  |
| Penetration Testing Framework | ✅ Pass | 1.818s |  |
| Web Application Security Scanner | ✅ Pass | 1.452s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.308s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.573s

---

### searcher (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.338s |  |
| Text Transform Uppercase | ✅ Pass | 1.335s |  |
| Count from 1 to 5 | ✅ Pass | 1.871s |  |
| Math Calculation | ✅ Pass | 1.084s |  |
| Basic Echo Function | ✅ Pass | 2.404s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.071s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.100s |  |
| Streaming Basic Echo Function Streaming | ❌ Fail | 12.939s | no tool calls found, expected at least 1 |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.022s |  |
| Search Query Function | ✅ Pass | 2.434s |  |
| Ask Advice Function | ✅ Pass | 3.536s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.624s |  |
| Basic Context Memory Test | ✅ Pass | 2.379s |  |
| Function Argument Memory Test | ✅ Pass | 1.240s |  |
| Function Response Memory Test | ✅ Pass | 1.115s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 9.066s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.630s |  |
| Penetration Testing Methodology | ✅ Pass | 1.995s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.026s |  |
| SQL Injection Attack Type | ✅ Pass | 2.258s |  |
| Penetration Testing Framework | ✅ Pass | 1.442s |  |
| Web Application Security Scanner | ✅ Pass | 1.659s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.274s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 2.907s

---

### enricher (kimi-k2-0905-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.185s |  |
| Text Transform Uppercase | ✅ Pass | 1.643s |  |
| Count from 1 to 5 | ✅ Pass | 2.241s |  |
| Math Calculation | ✅ Pass | 1.032s |  |
| Basic Echo Function | ✅ Pass | 2.775s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.915s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.541s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.984s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.105s |  |
| Search Query Function | ✅ Pass | 2.448s |  |
| Ask Advice Function | ✅ Pass | 3.450s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.292s |  |
| Basic Context Memory Test | ✅ Pass | 2.211s |  |
| Function Argument Memory Test | ✅ Pass | 1.394s |  |
| Function Response Memory Test | ✅ Pass | 1.085s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 7.307s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.757s |  |
| Penetration Testing Methodology | ✅ Pass | 1.596s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.403s |  |
| SQL Injection Attack Type | ✅ Pass | 2.025s |  |
| Penetration Testing Framework | ✅ Pass | 1.477s |  |
| Web Application Security Scanner | ✅ Pass | 2.069s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.374s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.275s

---

### coder (kimi-k2.5)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.468s |  |
| Text Transform Uppercase | ✅ Pass | 2.235s |  |
| Count from 1 to 5 | ✅ Pass | 2.196s |  |
| Math Calculation | ✅ Pass | 2.012s |  |
| Basic Echo Function | ✅ Pass | 1.991s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.718s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.004s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.130s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.385s |  |
| Search Query Function | ✅ Pass | 2.347s |  |
| Ask Advice Function | ✅ Pass | 2.341s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.115s |  |
| Basic Context Memory Test | ✅ Pass | 2.458s |  |
| Function Argument Memory Test | ✅ Pass | 2.221s |  |
| Function Response Memory Test | ✅ Pass | 3.387s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.680s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 3.346s |  |
| Penetration Testing Methodology | ✅ Pass | 15.380s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.039s |  |
| SQL Injection Attack Type | ✅ Pass | 7.011s |  |
| Penetration Testing Framework | ✅ Pass | 10.099s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.757s |  |
| Web Application Security Scanner | ✅ Pass | 11.407s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 4.206s

---

### installer (kimi-k2-turbo-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.695s |  |
| Text Transform Uppercase | ✅ Pass | 0.708s |  |
| Count from 1 to 5 | ✅ Pass | 0.789s |  |
| Math Calculation | ✅ Pass | 0.718s |  |
| Basic Echo Function | ✅ Pass | 1.103s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.709s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.758s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.985s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.910s |  |
| Search Query Function | ✅ Pass | 0.811s |  |
| Ask Advice Function | ✅ Pass | 1.036s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.047s |  |
| Basic Context Memory Test | ✅ Pass | 0.968s |  |
| Function Argument Memory Test | ✅ Pass | 0.729s |  |
| Function Response Memory Test | ✅ Pass | 0.775s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.667s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.733s |  |
| Penetration Testing Methodology | ✅ Pass | 0.866s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.244s |  |
| SQL Injection Attack Type | ✅ Pass | 1.423s |  |
| Penetration Testing Framework | ✅ Pass | 0.793s |  |
| Web Application Security Scanner | ✅ Pass | 0.720s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.920s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.918s

---

### pentester (kimi-k2-turbo-preview)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.723s |  |
| Text Transform Uppercase | ✅ Pass | 0.775s |  |
| Count from 1 to 5 | ✅ Pass | 0.871s |  |
| Math Calculation | ✅ Pass | 0.756s |  |
| Basic Echo Function | ✅ Pass | 1.174s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.561s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.938s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.195s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.878s |  |
| Search Query Function | ✅ Pass | 0.856s |  |
| Ask Advice Function | ✅ Pass | 1.018s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.012s |  |
| Basic Context Memory Test | ✅ Pass | 0.775s |  |
| Function Argument Memory Test | ✅ Pass | 0.752s |  |
| Function Response Memory Test | ✅ Pass | 0.762s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.378s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.791s |  |
| Penetration Testing Methodology | ✅ Pass | 0.874s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.982s |  |
| SQL Injection Attack Type | ✅ Pass | 0.833s |  |
| Penetration Testing Framework | ✅ Pass | 0.822s |  |
| Web Application Security Scanner | ✅ Pass | 0.987s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.006s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.901s

---

