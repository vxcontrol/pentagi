# LLM Agent Testing Report

Generated: Sun, 15 Mar 2026 15:53:05 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 14.417s |
| simple_json | Qwen/Qwen3.5-27B-FP8 | false | 5/5 (100.00%) | 48.110s |
| primary_agent | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 40.143s |
| assistant | Qwen/Qwen3.5-27B-FP8 | true | 22/23 (95.65%) | 52.153s |
| generator | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 50.132s |
| refiner | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 48.599s |
| adviser | Qwen/Qwen3.5-27B-FP8 | true | 22/23 (95.65%) | 51.045s |
| reflector | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 20.053s |
| searcher | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 15.935s |
| enricher | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 17.074s |
| coder | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 54.885s |
| installer | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 55.538s |
| pentester | Qwen/Qwen3.5-27B-FP8 | true | 23/23 (100.00%) | 54.749s |

**Total**: 279/281 (99.29%) successful tests
**Overall average latency**: 39.712s

## Detailed Results

### simple (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.676s |  |
| Text Transform Uppercase | ✅ Pass | 0.659s |  |
| Count from 1 to 5 | ✅ Pass | 0.505s |  |
| Math Calculation | ✅ Pass | 16.490s |  |
| Basic Echo Function | ✅ Pass | 19.115s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.274s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 10.085s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.804s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 53.536s |  |
| Search Query Function | ✅ Pass | 0.823s |  |
| Ask Advice Function | ✅ Pass | 1.980s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 36.073s |  |
| Basic Context Memory Test | ✅ Pass | 1.551s |  |
| Function Argument Memory Test | ✅ Pass | 0.306s |  |
| Function Response Memory Test | ✅ Pass | 5.207s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 11.206s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.317s |  |
| Penetration Testing Methodology | ✅ Pass | 27.195s |  |
| Vulnerability Assessment Tools | ✅ Pass | 30.694s |  |
| SQL Injection Attack Type | ✅ Pass | 2.421s |  |
| Penetration Testing Framework | ✅ Pass | 54.771s |  |
| Web Application Security Scanner | ✅ Pass | 40.336s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.550s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 14.417s

---

### simple_json (Qwen/Qwen3.5-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 65.427s |  |
| Project Information JSON | ✅ Pass | 43.413s |  |
| Person Information JSON | ✅ Pass | 53.143s |  |
| User Profile JSON | ✅ Pass | 42.996s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 35.570s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 48.110s

---

### primary_agent (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 46.317s |  |
| Text Transform Uppercase | ✅ Pass | 3.599s |  |
| Count from 1 to 5 | ✅ Pass | 77.707s |  |
| Math Calculation | ✅ Pass | 49.296s |  |
| Basic Echo Function | ✅ Pass | 46.650s |  |
| Streaming Simple Math Streaming | ✅ Pass | 5.358s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.253s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 17.503s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 54.723s |  |
| Search Query Function | ✅ Pass | 1.968s |  |
| Ask Advice Function | ✅ Pass | 2.593s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 38.959s |  |
| Basic Context Memory Test | ✅ Pass | 37.345s |  |
| Function Argument Memory Test | ✅ Pass | 90.263s |  |
| Function Response Memory Test | ✅ Pass | 3.072s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 15.508s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.638s |  |
| Penetration Testing Methodology | ✅ Pass | 20.098s |  |
| Vulnerability Assessment Tools | ✅ Pass | 178.331s |  |
| SQL Injection Attack Type | ✅ Pass | 42.430s |  |
| Penetration Testing Framework | ✅ Pass | 50.972s |  |
| Web Application Security Scanner | ✅ Pass | 75.701s |  |
| Penetration Testing Tool Selection | ✅ Pass | 61.984s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 40.143s

---

### assistant (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 46.006s |  |
| Text Transform Uppercase | ✅ Pass | 4.004s |  |
| Count from 1 to 5 | ✅ Pass | 66.849s |  |
| Math Calculation | ✅ Pass | 46.121s |  |
| Basic Echo Function | ✅ Pass | 56.990s |  |
| Streaming Simple Math Streaming | ✅ Pass | 4.660s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 14.354s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 10.898s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 42.389s |  |
| Search Query Function | ✅ Pass | 1.672s |  |
| Ask Advice Function | ✅ Pass | 2.230s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 39.160s |  |
| Basic Context Memory Test | ✅ Pass | 30.534s |  |
| Function Argument Memory Test | ✅ Pass | 81.043s |  |
| Function Response Memory Test | ✅ Pass | 3.833s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 16.904s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 63.736s |  |
| Penetration Testing Methodology | ✅ Pass | 16.286s |  |
| SQL Injection Attack Type | ✅ Pass | 47.840s |  |
| Vulnerability Assessment Tools | ✅ Pass | 437.897s |  |
| Penetration Testing Framework | ✅ Pass | 37.735s |  |
| Web Application Security Scanner | ✅ Pass | 67.262s |  |
| Penetration Testing Tool Selection | ✅ Pass | 61.109s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 52.153s

---

### generator (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 45.295s |  |
| Text Transform Uppercase | ✅ Pass | 4.303s |  |
| Count from 1 to 5 | ✅ Pass | 59.045s |  |
| Math Calculation | ✅ Pass | 66.939s |  |
| Basic Echo Function | ✅ Pass | 34.650s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.368s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 11.610s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 67.278s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 40.016s |  |
| Search Query Function | ✅ Pass | 2.165s |  |
| Ask Advice Function | ✅ Pass | 2.305s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 14.124s |  |
| Basic Context Memory Test | ✅ Pass | 72.382s |  |
| Function Argument Memory Test | ✅ Pass | 70.020s |  |
| Function Response Memory Test | ✅ Pass | 2.337s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 16.301s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 54.758s |  |
| Penetration Testing Methodology | ✅ Pass | 20.628s |  |
| Vulnerability Assessment Tools | ✅ Pass | 342.586s |  |
| SQL Injection Attack Type | ✅ Pass | 78.224s |  |
| Penetration Testing Framework | ✅ Pass | 25.054s |  |
| Web Application Security Scanner | ✅ Pass | 58.230s |  |
| Penetration Testing Tool Selection | ✅ Pass | 61.399s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 50.132s

---

### refiner (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 45.873s |  |
| Text Transform Uppercase | ✅ Pass | 4.709s |  |
| Count from 1 to 5 | ✅ Pass | 62.154s |  |
| Math Calculation | ✅ Pass | 37.647s |  |
| Basic Echo Function | ✅ Pass | 37.346s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.470s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.181s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 68.362s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 37.968s |  |
| Search Query Function | ✅ Pass | 2.222s |  |
| Ask Advice Function | ✅ Pass | 59.187s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 13.251s |  |
| Basic Context Memory Test | ✅ Pass | 3.312s |  |
| Function Argument Memory Test | ✅ Pass | 67.582s |  |
| Function Response Memory Test | ✅ Pass | 2.802s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 16.111s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 72.918s |  |
| Penetration Testing Methodology | ✅ Pass | 14.653s |  |
| Vulnerability Assessment Tools | ✅ Pass | 345.334s |  |
| SQL Injection Attack Type | ✅ Pass | 82.511s |  |
| Penetration Testing Framework | ✅ Pass | 11.861s |  |
| Web Application Security Scanner | ✅ Pass | 48.987s |  |
| Penetration Testing Tool Selection | ✅ Pass | 75.316s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 48.599s

---

### adviser (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 46.219s |  |
| Text Transform Uppercase | ✅ Pass | 5.548s |  |
| Count from 1 to 5 | ✅ Pass | 70.699s |  |
| Math Calculation | ✅ Pass | 60.313s |  |
| Basic Echo Function | ✅ Pass | 39.678s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.608s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.352s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 68.389s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 33.105s |  |
| Search Query Function | ✅ Pass | 1.318s |  |
| Ask Advice Function | ✅ Pass | 58.002s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 10.630s |  |
| Basic Context Memory Test | ✅ Pass | 16.399s |  |
| Function Argument Memory Test | ✅ Pass | 62.118s |  |
| Function Response Memory Test | ✅ Pass | 2.379s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 16.699s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 64.896s |  |
| Penetration Testing Methodology | ✅ Pass | 15.182s |  |
| Vulnerability Assessment Tools | ❌ Fail | 314.943s | expected text 'network' not found |
| SQL Injection Attack Type | ✅ Pass | 51.973s |  |
| Penetration Testing Framework | ✅ Pass | 10.408s |  |
| Web Application Security Scanner | ✅ Pass | 140.844s |  |
| Penetration Testing Tool Selection | ✅ Pass | 76.315s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 51.045s

---

### reflector (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.835s |  |
| Text Transform Uppercase | ✅ Pass | 1.439s |  |
| Count from 1 to 5 | ✅ Pass | 10.836s |  |
| Math Calculation | ✅ Pass | 14.311s |  |
| Basic Echo Function | ✅ Pass | 19.567s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.513s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.329s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 51.378s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 28.011s |  |
| Search Query Function | ✅ Pass | 0.793s |  |
| Ask Advice Function | ✅ Pass | 29.709s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 10.034s |  |
| Basic Context Memory Test | ✅ Pass | 1.082s |  |
| Function Argument Memory Test | ✅ Pass | 41.331s |  |
| Function Response Memory Test | ✅ Pass | 0.904s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 14.044s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 25.000s |  |
| Penetration Testing Methodology | ✅ Pass | 5.131s |  |
| Vulnerability Assessment Tools | ✅ Pass | 75.303s |  |
| SQL Injection Attack Type | ✅ Pass | 5.444s |  |
| Penetration Testing Framework | ✅ Pass | 55.241s |  |
| Web Application Security Scanner | ✅ Pass | 14.433s |  |
| Penetration Testing Tool Selection | ✅ Pass | 49.539s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 20.053s

---

### searcher (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 6.835s |  |
| Text Transform Uppercase | ✅ Pass | 0.289s |  |
| Count from 1 to 5 | ✅ Pass | 9.973s |  |
| Math Calculation | ✅ Pass | 13.611s |  |
| Basic Echo Function | ✅ Pass | 29.473s |  |
| Streaming Simple Math Streaming | ✅ Pass | 10.019s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.336s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 50.127s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.682s |  |
| Search Query Function | ✅ Pass | 0.708s |  |
| Ask Advice Function | ✅ Pass | 29.709s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 9.131s |  |
| Basic Context Memory Test | ✅ Pass | 0.501s |  |
| Function Argument Memory Test | ✅ Pass | 24.274s |  |
| Function Response Memory Test | ✅ Pass | 0.357s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 12.952s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 29.898s |  |
| Penetration Testing Methodology | ✅ Pass | 4.896s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.455s |  |
| SQL Injection Attack Type | ✅ Pass | 6.357s |  |
| Penetration Testing Framework | ✅ Pass | 34.478s |  |
| Web Application Security Scanner | ✅ Pass | 12.739s |  |
| Penetration Testing Tool Selection | ✅ Pass | 59.684s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 15.935s

---

### enricher (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 38.613s |  |
| Text Transform Uppercase | ✅ Pass | 0.288s |  |
| Count from 1 to 5 | ✅ Pass | 9.635s |  |
| Math Calculation | ✅ Pass | 5.748s |  |
| Basic Echo Function | ✅ Pass | 15.700s |  |
| Streaming Simple Math Streaming | ✅ Pass | 11.636s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.397s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 49.174s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.357s |  |
| Search Query Function | ✅ Pass | 1.935s |  |
| Ask Advice Function | ✅ Pass | 29.692s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 5.592s |  |
| Basic Context Memory Test | ✅ Pass | 0.396s |  |
| Function Argument Memory Test | ✅ Pass | 19.758s |  |
| Function Response Memory Test | ✅ Pass | 0.352s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.148s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 34.780s |  |
| Penetration Testing Methodology | ✅ Pass | 4.457s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.442s |  |
| SQL Injection Attack Type | ✅ Pass | 0.516s |  |
| Penetration Testing Framework | ✅ Pass | 63.726s |  |
| Web Application Security Scanner | ✅ Pass | 35.554s |  |
| Penetration Testing Tool Selection | ✅ Pass | 41.806s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 17.074s

---

### coder (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 39.481s |  |
| Text Transform Uppercase | ✅ Pass | 3.370s |  |
| Count from 1 to 5 | ✅ Pass | 78.615s |  |
| Math Calculation | ✅ Pass | 32.595s |  |
| Basic Echo Function | ✅ Pass | 16.466s |  |
| Streaming Simple Math Streaming | ✅ Pass | 14.916s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.345s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 74.466s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.116s |  |
| Search Query Function | ✅ Pass | 2.488s |  |
| Ask Advice Function | ✅ Pass | 61.793s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.247s |  |
| Basic Context Memory Test | ✅ Pass | 44.846s |  |
| Function Argument Memory Test | ✅ Pass | 11.874s |  |
| Function Response Memory Test | ✅ Pass | 1.275s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.892s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 82.178s |  |
| Penetration Testing Methodology | ✅ Pass | 154.106s |  |
| SQL Injection Attack Type | ✅ Pass | 19.059s |  |
| Vulnerability Assessment Tools | ✅ Pass | 298.388s |  |
| Penetration Testing Framework | ✅ Pass | 114.067s |  |
| Web Application Security Scanner | ✅ Pass | 150.870s |  |
| Penetration Testing Tool Selection | ✅ Pass | 47.893s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 54.885s

---

### installer (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 41.183s |  |
| Text Transform Uppercase | ✅ Pass | 4.185s |  |
| Count from 1 to 5 | ✅ Pass | 42.353s |  |
| Math Calculation | ✅ Pass | 36.572s |  |
| Basic Echo Function | ✅ Pass | 13.272s |  |
| Streaming Simple Math Streaming | ✅ Pass | 12.815s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 7.686s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 66.695s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.003s |  |
| Search Query Function | ✅ Pass | 2.361s |  |
| Ask Advice Function | ✅ Pass | 60.607s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.050s |  |
| Basic Context Memory Test | ✅ Pass | 29.974s |  |
| Function Argument Memory Test | ✅ Pass | 12.821s |  |
| Function Response Memory Test | ✅ Pass | 1.053s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.913s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 72.414s |  |
| Penetration Testing Methodology | ✅ Pass | 95.008s |  |
| SQL Injection Attack Type | ✅ Pass | 56.874s |  |
| Vulnerability Assessment Tools | ✅ Pass | 410.320s |  |
| Penetration Testing Framework | ✅ Pass | 113.426s |  |
| Web Application Security Scanner | ✅ Pass | 152.774s |  |
| Penetration Testing Tool Selection | ✅ Pass | 36.011s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 55.538s

---

### pentester (Qwen/Qwen3.5-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.622s |  |
| Text Transform Uppercase | ✅ Pass | 3.788s |  |
| Count from 1 to 5 | ✅ Pass | 61.459s |  |
| Math Calculation | ✅ Pass | 57.892s |  |
| Basic Echo Function | ✅ Pass | 10.738s |  |
| Streaming Simple Math Streaming | ✅ Pass | 13.100s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 5.418s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 54.160s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.458s |  |
| Search Query Function | ✅ Pass | 2.340s |  |
| Ask Advice Function | ✅ Pass | 60.580s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.590s |  |
| Basic Context Memory Test | ✅ Pass | 62.238s |  |
| Function Argument Memory Test | ✅ Pass | 7.654s |  |
| Function Response Memory Test | ✅ Pass | 1.304s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.581s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 71.250s |  |
| Penetration Testing Methodology | ✅ Pass | 152.774s |  |
| SQL Injection Attack Type | ✅ Pass | 82.404s |  |
| Vulnerability Assessment Tools | ✅ Pass | 294.588s |  |
| Penetration Testing Framework | ✅ Pass | 107.775s |  |
| Web Application Security Scanner | ✅ Pass | 166.016s |  |
| Penetration Testing Tool Selection | ✅ Pass | 31.482s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 54.749s

---

