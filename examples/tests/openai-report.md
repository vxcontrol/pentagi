# LLM Agent Testing Report

Generated: Tue, 08 Jul 2025 21:25:46 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | gpt-4.1-mini | false | 18/18 (100.00%) | 1.065s |
| simple_json | gpt-4.1-mini | false | 4/4 (100.00%) | 1.015s |
| primary_agent | o4-mini | false | 18/18 (100.00%) | 2.326s |
| assistant | o4-mini | true | 18/18 (100.00%) | 3.586s |
| generator | o1 | true | 18/18 (100.00%) | 3.128s |
| refiner | gpt-4.1 | false | 18/18 (100.00%) | 1.024s |
| adviser | o4-mini | true | 18/18 (100.00%) | 4.385s |
| reflector | o4-mini | true | 18/18 (100.00%) | 3.380s |
| searcher | gpt-4.1-mini | false | 18/18 (100.00%) | 0.913s |
| enricher | gpt-4.1-mini | false | 18/18 (100.00%) | 0.897s |
| coder | gpt-4.1 | false | 18/18 (100.00%) | 0.923s |
| installer | gpt-4.1 | false | 18/18 (100.00%) | 0.766s |
| pentester | o4-mini | false | 18/18 (100.00%) | 2.127s |

**Total**: 220/220 (100.00%) successful tests
**Overall average latency**: 2.025s

## Detailed Results

### simple (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.558s |  |
| Text Transform Uppercase | ✅ Pass | 0.599s |  |
| Count from 1 to 5 | ✅ Pass | 0.921s |  |
| Math Calculation | ✅ Pass | 1.313s |  |
| Basic Echo Function | ✅ Pass | 0.740s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.577s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.696s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.764s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.797s |  |
| Search Query Function | ✅ Pass | 2.915s |  |
| Ask Advice Function | ✅ Pass | 1.086s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.862s |  |
| Penetration Testing Methodology | ✅ Pass | 0.838s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.701s |  |
| SQL Injection Attack Type | ✅ Pass | 0.935s |  |
| Penetration Testing Framework | ✅ Pass | 1.159s |  |
| Web Application Security Scanner | ✅ Pass | 0.853s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.851s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.065s

---

### simple_json (gpt-4.1-mini)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 0.883s |  |
| Person Information JSON | ✅ Pass | 1.461s |  |
| User Profile JSON | ✅ Pass | 0.866s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.846s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 1.015s

---

### primary_agent (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.914s |  |
| Text Transform Uppercase | ✅ Pass | 2.637s |  |
| Count from 1 to 5 | ✅ Pass | 1.717s |  |
| Math Calculation | ✅ Pass | 1.763s |  |
| Basic Echo Function | ✅ Pass | 1.910s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.858s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.096s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.249s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.397s |  |
| Search Query Function | ✅ Pass | 1.475s |  |
| Ask Advice Function | ✅ Pass | 1.952s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.485s |  |
| Penetration Testing Methodology | ✅ Pass | 2.619s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.531s |  |
| SQL Injection Attack Type | ✅ Pass | 2.825s |  |
| Penetration Testing Framework | ✅ Pass | 5.066s |  |
| Web Application Security Scanner | ✅ Pass | 2.406s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.969s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.326s

---

### assistant (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 7.605s |  |
| Text Transform Uppercase | ✅ Pass | 3.066s |  |
| Count from 1 to 5 | ✅ Pass | 2.029s |  |
| Math Calculation | ✅ Pass | 1.652s |  |
| Basic Echo Function | ✅ Pass | 3.493s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.992s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.878s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.035s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.213s |  |
| Search Query Function | ✅ Pass | 2.179s |  |
| Ask Advice Function | ✅ Pass | 3.083s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.438s |  |
| Penetration Testing Methodology | ✅ Pass | 7.206s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.747s |  |
| SQL Injection Attack Type | ✅ Pass | 4.267s |  |
| Penetration Testing Framework | ✅ Pass | 5.002s |  |
| Web Application Security Scanner | ✅ Pass | 4.282s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.377s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.586s

---

### generator (o1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.039s |  |
| Text Transform Uppercase | ✅ Pass | 2.351s |  |
| Count from 1 to 5 | ✅ Pass | 3.450s |  |
| Math Calculation | ✅ Pass | 3.334s |  |
| Basic Echo Function | ✅ Pass | 1.935s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.859s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.794s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.596s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.846s |  |
| Search Query Function | ✅ Pass | 1.878s |  |
| Ask Advice Function | ✅ Pass | 2.005s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.654s |  |
| Penetration Testing Methodology | ✅ Pass | 2.737s |  |
| Vulnerability Assessment Tools | ✅ Pass | 6.704s |  |
| SQL Injection Attack Type | ✅ Pass | 2.990s |  |
| Penetration Testing Framework | ✅ Pass | 2.792s |  |
| Web Application Security Scanner | ✅ Pass | 3.937s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.397s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.128s

---

### refiner (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.252s |  |
| Text Transform Uppercase | ✅ Pass | 1.071s |  |
| Count from 1 to 5 | ✅ Pass | 0.635s |  |
| Math Calculation | ✅ Pass | 0.703s |  |
| Basic Echo Function | ✅ Pass | 0.789s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.615s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.586s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.692s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.255s |  |
| Search Query Function | ✅ Pass | 0.827s |  |
| Ask Advice Function | ✅ Pass | 0.923s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.950s |  |
| Penetration Testing Methodology | ✅ Pass | 1.751s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.476s |  |
| SQL Injection Attack Type | ✅ Pass | 0.822s |  |
| Penetration Testing Framework | ✅ Pass | 1.066s |  |
| Web Application Security Scanner | ✅ Pass | 2.077s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.939s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 1.024s

---

### adviser (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.973s |  |
| Text Transform Uppercase | ✅ Pass | 2.272s |  |
| Count from 1 to 5 | ✅ Pass | 2.690s |  |
| Math Calculation | ✅ Pass | 2.715s |  |
| Basic Echo Function | ✅ Pass | 3.357s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.875s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.081s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.457s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Search Query Function | ✅ Pass | 2.696s |  |
| Ask Advice Function | ✅ Pass | 2.630s |  |
| JSON Response Function | ✅ Pass | 18.825s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.864s |  |
| Penetration Testing Methodology | ✅ Pass | 6.513s |  |
| Vulnerability Assessment Tools | ✅ Pass | 5.925s |  |
| SQL Injection Attack Type | ✅ Pass | 3.627s |  |
| Penetration Testing Framework | ✅ Pass | 5.732s |  |
| Web Application Security Scanner | ✅ Pass | 5.752s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.940s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.385s

---

### reflector (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.038s |  |
| Text Transform Uppercase | ✅ Pass | 1.618s |  |
| Count from 1 to 5 | ✅ Pass | 3.994s |  |
| Math Calculation | ✅ Pass | 2.153s |  |
| Basic Echo Function | ✅ Pass | 3.778s |  |
| Streaming Simple Math Streaming | ✅ Pass | 2.919s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.889s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.707s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.598s |  |
| Search Query Function | ✅ Pass | 2.151s |  |
| Ask Advice Function | ✅ Pass | 2.531s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.962s |  |
| Penetration Testing Methodology | ✅ Pass | 3.389s |  |
| Vulnerability Assessment Tools | ✅ Pass | 8.344s |  |
| SQL Injection Attack Type | ✅ Pass | 3.381s |  |
| Penetration Testing Framework | ✅ Pass | 3.443s |  |
| Web Application Security Scanner | ✅ Pass | 8.382s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.545s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 3.380s

---

### searcher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.597s |  |
| Text Transform Uppercase | ✅ Pass | 0.736s |  |
| Count from 1 to 5 | ✅ Pass | 0.718s |  |
| Math Calculation | ✅ Pass | 0.860s |  |
| Basic Echo Function | ✅ Pass | 0.756s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.579s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.556s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.817s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.065s |  |
| Search Query Function | ✅ Pass | 0.812s |  |
| Ask Advice Function | ✅ Pass | 1.115s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.068s |  |
| Penetration Testing Methodology | ✅ Pass | 0.779s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.168s |  |
| SQL Injection Attack Type | ✅ Pass | 0.703s |  |
| Penetration Testing Framework | ✅ Pass | 1.118s |  |
| Web Application Security Scanner | ✅ Pass | 0.749s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.221s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.913s

---

### enricher (gpt-4.1-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.710s |  |
| Text Transform Uppercase | ✅ Pass | 0.586s |  |
| Count from 1 to 5 | ✅ Pass | 0.844s |  |
| Math Calculation | ✅ Pass | 0.600s |  |
| Basic Echo Function | ✅ Pass | 0.709s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.488s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.592s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.945s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.868s |  |
| Search Query Function | ✅ Pass | 0.755s |  |
| Ask Advice Function | ✅ Pass | 1.055s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.886s |  |
| Penetration Testing Methodology | ✅ Pass | 0.842s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.732s |  |
| SQL Injection Attack Type | ✅ Pass | 0.557s |  |
| Penetration Testing Framework | ✅ Pass | 1.129s |  |
| Web Application Security Scanner | ✅ Pass | 0.977s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.855s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.897s

---

### coder (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.901s |  |
| Text Transform Uppercase | ✅ Pass | 0.642s |  |
| Count from 1 to 5 | ✅ Pass | 0.837s |  |
| Math Calculation | ✅ Pass | 0.669s |  |
| Basic Echo Function | ✅ Pass | 0.687s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.569s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.529s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.777s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.958s |  |
| Search Query Function | ✅ Pass | 0.859s |  |
| Ask Advice Function | ✅ Pass | 1.267s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.645s |  |
| Penetration Testing Methodology | ✅ Pass | 0.937s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.352s |  |
| SQL Injection Attack Type | ✅ Pass | 0.814s |  |
| Penetration Testing Framework | ✅ Pass | 0.923s |  |
| Web Application Security Scanner | ✅ Pass | 1.514s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.722s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.923s

---

### installer (gpt-4.1)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.534s |  |
| Text Transform Uppercase | ✅ Pass | 0.686s |  |
| Count from 1 to 5 | ✅ Pass | 0.722s |  |
| Math Calculation | ✅ Pass | 0.615s |  |
| Basic Echo Function | ✅ Pass | 0.650s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.677s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.550s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.767s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.954s |  |
| Search Query Function | ✅ Pass | 0.676s |  |
| Ask Advice Function | ✅ Pass | 0.823s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.688s |  |
| Penetration Testing Methodology | ✅ Pass | 0.781s |  |
| Vulnerability Assessment Tools | ✅ Pass | 1.237s |  |
| SQL Injection Attack Type | ✅ Pass | 0.712s |  |
| Penetration Testing Framework | ✅ Pass | 0.955s |  |
| Web Application Security Scanner | ✅ Pass | 0.945s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.807s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 0.766s

---

### pentester (o4-mini)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.254s |  |
| Text Transform Uppercase | ✅ Pass | 1.637s |  |
| Count from 1 to 5 | ✅ Pass | 1.633s |  |
| Math Calculation | ✅ Pass | 1.225s |  |
| Basic Echo Function | ✅ Pass | 1.496s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.352s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.640s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.583s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.543s |  |
| Search Query Function | ✅ Pass | 1.749s |  |
| Ask Advice Function | ✅ Pass | 1.963s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.566s |  |
| Penetration Testing Methodology | ✅ Pass | 2.411s |  |
| Vulnerability Assessment Tools | ✅ Pass | 4.824s |  |
| SQL Injection Attack Type | ✅ Pass | 2.934s |  |
| Penetration Testing Framework | ✅ Pass | 3.227s |  |
| Web Application Security Scanner | ✅ Pass | 2.229s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.006s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 2.127s

---

