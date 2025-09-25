# LLM Agent Testing Report

Generated: Sat, 19 Jul 2025 17:34:06 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-4.1-mini | false | 23/23 (100.00%) | 0.780s |
| simple_json | gpt-4.1-mini | false | 5/5 (100.00%) | 1.057s |
| primary_agent | o4-mini | true | 22/23 (95.65%) | 1.666s |
| assistant | o4-mini | true | 23/23 (100.00%) | 2.582s |
| generator | o1 | true | 23/23 (100.00%) | 2.882s |
| refiner | gpt-4.1 | false | 23/23 (100.00%) | 1.478s |
| adviser | o4-mini | true | 23/23 (100.00%) | 2.375s |
| reflector | o4-mini | true | 23/23 (100.00%) | 2.327s |
| searcher | gpt-4.1-mini | false | 23/23 (100.00%) | 0.780s |
| enricher | gpt-4.1-mini | false | 23/23 (100.00%) | 0.673s |
| coder | gpt-4.1 | false | 23/23 (100.00%) | 1.319s |
| installer | gpt-4.1 | false | 23/23 (100.00%) | 1.211s |
| pentester | o4-mini | true | 22/23 (95.65%) | 1.597s |

**Total**: 279/281 (99.29%) successful tests
**Overall average latency**: 1.629s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.706s |  |
| Text Transform Uppercase | ✅ Pass | 0.408s |  |
| Count from 1 to 5 | ✅ Pass | 0.492s |  |
| Math Calculation | ✅ Pass | 0.474s |  |
| Basic Echo Function | ✅ Pass | 0.611s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.400s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.412s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.591s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.225s |  |
| Search Query Function | ✅ Pass | 0.916s |  |
| Ask Advice Function | ✅ Pass | 1.237s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.765s |  |
| Basic Context Memory Test | ✅ Pass | 0.606s |  |
| Function Argument Memory Test | ✅ Pass | 0.396s |  |
| Function Response Memory Test | ✅ Pass | 0.507s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.343s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.394s |  |
| Penetration Testing Methodology | ✅ Pass | 0.699s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.259s |  |
| SQL Injection Attack Type | ✅ Pass | 0.471s |  |
| Penetration Testing Framework | ✅ Pass | 0.719s |  |
| Web Application Security Scanner | ✅ Pass | 0.502s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.792s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.780s

---

### simple_json (gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 0.633s |  |
| Vulnerability Report Memory Test | ✅ Pass | 1.250s |  |
| Person Information JSON | ✅ Pass | 1.395s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.696s |  |
| User Profile JSON | ✅ Pass | 1.310s |  |

**Summary**: 5/5 (100.00%) successful tests

**Average latency**: 1.057s

---

### primary_agent (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.986s |  |
| Text Transform Uppercase | ✅ Pass | 1.375s |  |
| Math Calculation | ✅ Pass | 1.092s |  |
| Count from 1 to 5 | ✅ Pass | 3.523s |  |
| Basic Echo Function | ✅ Pass | 1.088s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.171s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.559s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.895s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.924s |  |
| Search Query Function | ✅ Pass | 1.222s |  |
| Ask Advice Function | ✅ Pass | 1.172s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.340s |  |
| Basic Context Memory Test | ✅ Pass | 1.613s |  |
| Function Argument Memory Test | ✅ Pass | 1.483s |  |
| Function Response Memory Test | ✅ Pass | 1.483s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.688s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.209s |  |
| Penetration Testing Methodology | ✅ Pass | 1.620s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.288s |  |
| SQL Injection Attack Type | ✅ Pass | 2.301s |  |
| Penetration Testing Framework | ✅ Pass | 2.256s |  |
| Web Application Security Scanner | ✅ Pass | 1.369s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.639s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.666s

---

### assistant (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.787s |  |
| Text Transform Uppercase | ✅ Pass | 1.239s |  |
| Count from 1 to 5 | ✅ Pass | 1.846s |  |
| Math Calculation | ✅ Pass | 1.242s |  |
| Basic Echo Function | ✅ Pass | 2.381s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.194s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.178s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.805s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 2.025s |  |
| Ask Advice Function | ✅ Pass | 2.158s |  |
| JSON Response Function | ✅ Pass | 8.600s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.553s |  |
| Basic Context Memory Test | ✅ Pass | 3.307s |  |
| Function Argument Memory Test | ✅ Pass | 2.096s |  |
| Function Response Memory Test | ✅ Pass | 1.774s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.961s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.801s |  |
| Penetration Testing Methodology | ✅ Pass | 1.855s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.216s |  |
| SQL Injection Attack Type | ✅ Pass | 2.243s |  |
| Penetration Testing Framework | ✅ Pass | 2.813s |  |
| Web Application Security Scanner | ✅ Pass | 1.693s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.614s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.582s

---

### generator (o1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Text Transform Uppercase | ✅ Pass | 1.587s |  |
| Simple Math | ✅ Pass | 4.813s |  |
| Count from 1 to 5 | ✅ Pass | 2.400s |  |
| Math Calculation | ✅ Pass | 2.553s |  |
| Basic Echo Function | ✅ Pass | 2.206s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.075s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.078s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.712s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.950s |  |
| Search Query Function | ✅ Pass | 3.545s |  |
| Ask Advice Function | ✅ Pass | 1.920s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.159s |  |
| Basic Context Memory Test | ✅ Pass | 2.708s |  |
| Function Response Memory Test | ✅ Pass | 2.391s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.559s |  |
| Function Argument Memory Test | ✅ Pass | 9.485s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.383s |  |
| Penetration Testing Methodology | ✅ Pass | 1.325s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.085s |  |
| SQL Injection Attack Type | ✅ Pass | 3.307s |  |
| Penetration Testing Framework | ✅ Pass | 2.229s |  |
| Web Application Security Scanner | ✅ Pass | 1.458s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.340s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.882s

---

### refiner (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.833s |  |
| Text Transform Uppercase | ✅ Pass | 0.398s |  |
| Count from 1 to 5 | ✅ Pass | 0.487s |  |
| Math Calculation | ✅ Pass | 0.435s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.526s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.475s |  |
| Basic Echo Function | ✅ Pass | 10.415s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.043s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.739s |  |
| Search Query Function | ✅ Pass | 0.656s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.653s |  |
| Ask Advice Function | ✅ Pass | 6.555s |  |
| Basic Context Memory Test | ✅ Pass | 0.756s |  |
| Function Argument Memory Test | ✅ Pass | 0.491s |  |
| Function Response Memory Test | ✅ Pass | 0.468s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.882s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.473s |  |
| Penetration Testing Methodology | ✅ Pass | 0.631s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.252s |  |
| SQL Injection Attack Type | ✅ Pass | 0.594s |  |
| Penetration Testing Framework | ✅ Pass | 0.563s |  |
| Web Application Security Scanner | ✅ Pass | 0.659s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.996s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.478s

---

### adviser (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.050s |  |
| Text Transform Uppercase | ✅ Pass | 2.436s |  |
| Count from 1 to 5 | ✅ Pass | 1.470s |  |
| Math Calculation | ✅ Pass | 1.041s |  |
| Basic Echo Function | ✅ Pass | 2.512s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.396s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.633s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.540s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.848s |  |
| Search Query Function | ✅ Pass | 2.046s |  |
| Ask Advice Function | ✅ Pass | 2.679s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.439s |  |
| Basic Context Memory Test | ✅ Pass | 1.794s |  |
| Function Argument Memory Test | ✅ Pass | 2.027s |  |
| Function Response Memory Test | ✅ Pass | 1.128s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.917s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 2.023s |  |
| Penetration Testing Methodology | ✅ Pass | 3.004s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.470s |  |
| SQL Injection Attack Type | ✅ Pass | 2.967s |  |
| Penetration Testing Framework | ✅ Pass | 2.959s |  |
| Web Application Security Scanner | ✅ Pass | 2.009s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.228s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.375s

---

### reflector (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.201s |  |
| Text Transform Uppercase | ✅ Pass | 1.419s |  |
| Count from 1 to 5 | ✅ Pass | 2.141s |  |
| Math Calculation | ✅ Pass | 1.407s |  |
| Basic Echo Function | ✅ Pass | 3.891s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.276s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.323s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.430s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.526s |  |
| Search Query Function | ✅ Pass | 1.386s |  |
| Ask Advice Function | ✅ Pass | 1.947s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.394s |  |
| Basic Context Memory Test | ✅ Pass | 3.876s |  |
| Function Argument Memory Test | ✅ Pass | 1.487s |  |
| Function Response Memory Test | ✅ Pass | 1.622s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 3.749s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.879s |  |
| Penetration Testing Methodology | ✅ Pass | 2.096s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.675s |  |
| SQL Injection Attack Type | ✅ Pass | 1.924s |  |
| Penetration Testing Framework | ✅ Pass | 2.378s |  |
| Web Application Security Scanner | ✅ Pass | 1.675s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.811s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 2.327s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.661s |  |
| Text Transform Uppercase | ✅ Pass | 0.508s |  |
| Count from 1 to 5 | ✅ Pass | 0.520s |  |
| Math Calculation | ✅ Pass | 0.800s |  |
| Basic Echo Function | ✅ Pass | 0.810s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.511s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.434s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.628s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.925s |  |
| Search Query Function | ✅ Pass | 1.037s |  |
| Ask Advice Function | ✅ Pass | 0.826s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.772s |  |
| Basic Context Memory Test | ✅ Pass | 0.612s |  |
| Function Argument Memory Test | ✅ Pass | 0.493s |  |
| Function Response Memory Test | ✅ Pass | 0.386s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.114s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.506s |  |
| Penetration Testing Methodology | ✅ Pass | 0.665s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.165s |  |
| SQL Injection Attack Type | ✅ Pass | 0.589s |  |
| Penetration Testing Framework | ✅ Pass | 0.649s |  |
| Web Application Security Scanner | ✅ Pass | 0.586s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.724s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.780s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.624s |  |
| Text Transform Uppercase | ✅ Pass | 0.438s |  |
| Count from 1 to 5 | ✅ Pass | 0.520s |  |
| Math Calculation | ✅ Pass | 0.493s |  |
| Basic Echo Function | ✅ Pass | 0.651s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.372s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.459s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.563s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.716s |  |
| Search Query Function | ✅ Pass | 0.848s |  |
| Ask Advice Function | ✅ Pass | 0.737s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.043s |  |
| Basic Context Memory Test | ✅ Pass | 0.544s |  |
| Function Argument Memory Test | ✅ Pass | 0.531s |  |
| Function Response Memory Test | ✅ Pass | 0.363s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.069s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.507s |  |
| Penetration Testing Methodology | ✅ Pass | 0.845s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.602s |  |
| SQL Injection Attack Type | ✅ Pass | 0.455s |  |
| Penetration Testing Framework | ✅ Pass | 0.817s |  |
| Web Application Security Scanner | ✅ Pass | 0.579s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.696s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 0.673s

---

### coder (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.562s |  |
| Text Transform Uppercase | ✅ Pass | 0.433s |  |
| Count from 1 to 5 | ✅ Pass | 1.227s |  |
| Math Calculation | ✅ Pass | 0.552s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.430s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.527s |  |
| Basic Echo Function | ✅ Pass | 9.753s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.297s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.510s |  |
| Search Query Function | ✅ Pass | 0.593s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.549s |  |
| Ask Advice Function | ✅ Pass | 5.503s |  |
| Basic Context Memory Test | ✅ Pass | 0.449s |  |
| Function Argument Memory Test | ✅ Pass | 0.534s |  |
| Function Response Memory Test | ✅ Pass | 0.612s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.669s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.525s |  |
| Penetration Testing Methodology | ✅ Pass | 1.356s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.959s |  |
| SQL Injection Attack Type | ✅ Pass | 0.467s |  |
| Penetration Testing Framework | ✅ Pass | 0.458s |  |
| Web Application Security Scanner | ✅ Pass | 0.708s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.649s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.319s

---

### installer (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.364s |  |
| Text Transform Uppercase | ✅ Pass | 0.422s |  |
| Count from 1 to 5 | ✅ Pass | 0.571s |  |
| Math Calculation | ✅ Pass | 0.347s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.513s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.442s |  |
| Basic Echo Function | ✅ Pass | 9.060s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.359s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.558s |  |
| Search Query Function | ✅ Pass | 0.796s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.689s |  |
| Ask Advice Function | ✅ Pass | 5.312s |  |
| Basic Context Memory Test | ✅ Pass | 0.690s |  |
| Function Argument Memory Test | ✅ Pass | 0.447s |  |
| Function Response Memory Test | ✅ Pass | 0.483s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 1.024s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.567s |  |
| Penetration Testing Methodology | ✅ Pass | 0.718s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.983s |  |
| SQL Injection Attack Type | ✅ Pass | 0.537s |  |
| Penetration Testing Framework | ✅ Pass | 0.657s |  |
| Web Application Security Scanner | ✅ Pass | 0.568s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.730s |  |

**Summary**: 23/23 (100.00%) successful tests

**Average latency**: 1.211s

---

### pentester (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.024s |  |
| Text Transform Uppercase | ✅ Pass | 1.239s |  |
| Count from 1 to 5 | ✅ Pass | 1.331s |  |
| Math Calculation | ✅ Pass | 1.345s |  |
| Basic Echo Function | ✅ Pass | 1.495s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.107s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.941s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.656s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.091s |  |
| Search Query Function | ✅ Pass | 1.109s |  |
| Ask Advice Function | ✅ Pass | 1.291s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.993s |  |
| Basic Context Memory Test | ✅ Pass | 1.600s |  |
| Function Argument Memory Test | ✅ Pass | 1.778s |  |
| Function Response Memory Test | ✅ Pass | 2.185s |  |
| Penetration Testing Memory with Tool Call | ❌ Fail | 1.617s | expected function 'generate\_report' not found in tool calls: expected function generate\_report not found in tool calls |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.497s |  |
| Penetration Testing Methodology | ✅ Pass | 1.607s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.159s |  |
| SQL Injection Attack Type | ✅ Pass | 1.839s |  |
| Penetration Testing Framework | ✅ Pass | 1.511s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.357s |  |
| Web Application Security Scanner | ✅ Pass | 3.946s |  |

**Summary**: 22/23 (95.65%) successful tests

**Average latency**: 1.597s

---

