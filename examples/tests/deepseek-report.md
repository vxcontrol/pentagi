# LLM Agent Testing Report

Generated: Tue, 08 Jul 2025 21:52:20 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | deepseek-chat | false | 18/18 (100.00%) | 6.035s |
| simple_json | deepseek-chat | false | 4/4 (100.00%) | 4.235s |
| primary_agent | deepseek-chat | false | 18/18 (100.00%) | 5.777s |
| assistant | deepseek-chat | false | 18/18 (100.00%) | 5.673s |
| generator | deepseek-chat | false | 18/18 (100.00%) | 5.567s |
| refiner | deepseek-chat | false | 18/18 (100.00%) | 5.718s |
| adviser | deepseek-chat | false | 18/18 (100.00%) | 5.776s |
| reflector | deepseek-chat | false | 18/18 (100.00%) | 5.476s |
| searcher | deepseek-chat | false | 18/18 (100.00%) | 5.779s |
| enricher | deepseek-chat | false | 18/18 (100.00%) | 5.919s |
| coder | deepseek-coder | false | 18/18 (100.00%) | 6.060s |
| installer | deepseek-coder | false | 18/18 (100.00%) | 5.820s |
| pentester | deepseek-chat | false | 18/18 (100.00%) | 5.203s |

**Total**: 220/220 (100.00%) successful tests
**Overall average latency**: 5.706s

## Detailed Results

### simple (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.767s |  |
| Text Transform Uppercase | ✅ Pass | 3.980s |  |
| Count from 1 to 5 | ✅ Pass | 4.163s |  |
| Math Calculation | ✅ Pass | 3.807s |  |
| Basic Echo Function | ✅ Pass | 4.347s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.667s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.491s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.730s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.775s |  |
| Search Query Function | ✅ Pass | 4.006s |  |
| Ask Advice Function | ✅ Pass | 4.638s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.453s |  |
| Penetration Testing Methodology | ✅ Pass | 6.999s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.954s |  |
| SQL Injection Attack Type | ✅ Pass | 6.029s |  |
| Penetration Testing Framework | ✅ Pass | 14.651s |  |
| Web Application Security Scanner | ✅ Pass | 11.180s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.973s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 6.035s

---

### simple_json (deepseek-chat)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Project Information JSON | ✅ Pass | 3.019s |  |
| Person Information JSON | ✅ Pass | 4.924s |  |
| User Profile JSON | ✅ Pass | 4.891s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 4.102s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 4.235s

---

### primary_agent (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.177s |  |
| Text Transform Uppercase | ✅ Pass | 3.579s |  |
| Count from 1 to 5 | ✅ Pass | 3.183s |  |
| Math Calculation | ✅ Pass | 3.831s |  |
| Basic Echo Function | ✅ Pass | 4.351s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.485s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.650s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.242s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.437s |  |
| Search Query Function | ✅ Pass | 3.824s |  |
| Ask Advice Function | ✅ Pass | 4.403s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.011s |  |
| Penetration Testing Methodology | ✅ Pass | 8.408s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.510s |  |
| SQL Injection Attack Type | ✅ Pass | 5.139s |  |
| Penetration Testing Framework | ✅ Pass | 13.095s |  |
| Web Application Security Scanner | ✅ Pass | 10.198s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.451s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.777s

---

### assistant (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 5.030s |  |
| Text Transform Uppercase | ✅ Pass | 3.390s |  |
| Count from 1 to 5 | ✅ Pass | 4.192s |  |
| Math Calculation | ✅ Pass | 2.956s |  |
| Basic Echo Function | ✅ Pass | 4.129s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.309s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.217s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.106s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.947s |  |
| Search Query Function | ✅ Pass | 4.109s |  |
| Ask Advice Function | ✅ Pass | 4.772s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.805s |  |
| Penetration Testing Methodology | ✅ Pass | 6.635s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.035s |  |
| SQL Injection Attack Type | ✅ Pass | 5.550s |  |
| Penetration Testing Framework | ✅ Pass | 10.839s |  |
| Web Application Security Scanner | ✅ Pass | 11.729s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.356s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.673s

---

### generator (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 4.766s |  |
| Text Transform Uppercase | ✅ Pass | 3.578s |  |
| Count from 1 to 5 | ✅ Pass | 3.928s |  |
| Math Calculation | ✅ Pass | 3.107s |  |
| Basic Echo Function | ✅ Pass | 4.332s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.414s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.176s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.463s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.578s |  |
| Search Query Function | ✅ Pass | 3.988s |  |
| Ask Advice Function | ✅ Pass | 4.865s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.210s |  |
| Penetration Testing Methodology | ✅ Pass | 6.158s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.204s |  |
| SQL Injection Attack Type | ✅ Pass | 5.766s |  |
| Penetration Testing Framework | ✅ Pass | 12.134s |  |
| Web Application Security Scanner | ✅ Pass | 8.885s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.648s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.567s

---

### refiner (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.771s |  |
| Text Transform Uppercase | ✅ Pass | 3.789s |  |
| Count from 1 to 5 | ✅ Pass | 3.318s |  |
| Math Calculation | ✅ Pass | 3.410s |  |
| Basic Echo Function | ✅ Pass | 4.151s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.908s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.574s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.090s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.543s |  |
| Search Query Function | ✅ Pass | 4.441s |  |
| Ask Advice Function | ✅ Pass | 4.713s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.496s |  |
| Penetration Testing Methodology | ✅ Pass | 6.888s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.416s |  |
| SQL Injection Attack Type | ✅ Pass | 5.953s |  |
| Penetration Testing Framework | ✅ Pass | 13.605s |  |
| Web Application Security Scanner | ✅ Pass | 9.710s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.147s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.718s

---

### adviser (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.859s |  |
| Text Transform Uppercase | ✅ Pass | 3.058s |  |
| Count from 1 to 5 | ✅ Pass | 4.129s |  |
| Math Calculation | ✅ Pass | 3.580s |  |
| Basic Echo Function | ✅ Pass | 4.176s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.803s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.352s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.102s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.709s |  |
| Search Query Function | ✅ Pass | 4.214s |  |
| Ask Advice Function | ✅ Pass | 4.685s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.626s |  |
| Penetration Testing Methodology | ✅ Pass | 6.060s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.268s |  |
| SQL Injection Attack Type | ✅ Pass | 5.118s |  |
| Penetration Testing Framework | ✅ Pass | 11.966s |  |
| Web Application Security Scanner | ✅ Pass | 9.845s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.413s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.776s

---

### reflector (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.558s |  |
| Text Transform Uppercase | ✅ Pass | 3.383s |  |
| Count from 1 to 5 | ✅ Pass | 3.209s |  |
| Math Calculation | ✅ Pass | 3.686s |  |
| Basic Echo Function | ✅ Pass | 4.372s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.778s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 2.889s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.478s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.412s |  |
| Search Query Function | ✅ Pass | 4.183s |  |
| Ask Advice Function | ✅ Pass | 4.371s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.587s |  |
| Penetration Testing Methodology | ✅ Pass | 7.024s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.052s |  |
| SQL Injection Attack Type | ✅ Pass | 4.185s |  |
| Penetration Testing Framework | ✅ Pass | 10.406s |  |
| Web Application Security Scanner | ✅ Pass | 11.151s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.834s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.476s

---

### searcher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.359s |  |
| Text Transform Uppercase | ✅ Pass | 3.536s |  |
| Count from 1 to 5 | ✅ Pass | 3.431s |  |
| Math Calculation | ✅ Pass | 3.445s |  |
| Basic Echo Function | ✅ Pass | 3.871s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.699s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.693s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.636s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.658s |  |
| Search Query Function | ✅ Pass | 4.299s |  |
| Ask Advice Function | ✅ Pass | 5.249s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.909s |  |
| Penetration Testing Methodology | ✅ Pass | 7.192s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.373s |  |
| SQL Injection Attack Type | ✅ Pass | 6.005s |  |
| Penetration Testing Framework | ✅ Pass | 13.268s |  |
| Web Application Security Scanner | ✅ Pass | 11.441s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.951s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.779s

---

### enricher (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.335s |  |
| Text Transform Uppercase | ✅ Pass | 3.378s |  |
| Count from 1 to 5 | ✅ Pass | 3.747s |  |
| Math Calculation | ✅ Pass | 3.408s |  |
| Basic Echo Function | ✅ Pass | 4.023s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.784s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.223s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 3.976s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.813s |  |
| Search Query Function | ✅ Pass | 4.320s |  |
| Ask Advice Function | ✅ Pass | 4.389s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.390s |  |
| Penetration Testing Methodology | ✅ Pass | 8.176s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.848s |  |
| SQL Injection Attack Type | ✅ Pass | 5.214s |  |
| Penetration Testing Framework | ✅ Pass | 16.077s |  |
| Web Application Security Scanner | ✅ Pass | 10.176s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.251s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.919s

---

### coder (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.069s |  |
| Text Transform Uppercase | ✅ Pass | 4.379s |  |
| Count from 1 to 5 | ✅ Pass | 3.381s |  |
| Math Calculation | ✅ Pass | 3.684s |  |
| Basic Echo Function | ✅ Pass | 4.378s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.580s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 4.037s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.308s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.438s |  |
| Search Query Function | ✅ Pass | 4.182s |  |
| Ask Advice Function | ✅ Pass | 4.460s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.831s |  |
| Penetration Testing Methodology | ✅ Pass | 7.299s |  |
| Vulnerability Assessment Tools | ✅ Pass | 15.768s |  |
| SQL Injection Attack Type | ✅ Pass | 6.194s |  |
| Penetration Testing Framework | ✅ Pass | 16.445s |  |
| Web Application Security Scanner | ✅ Pass | 11.129s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.515s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 6.060s

---

### installer (deepseek-coder)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 3.018s |  |
| Text Transform Uppercase | ✅ Pass | 3.292s |  |
| Count from 1 to 5 | ✅ Pass | 3.435s |  |
| Math Calculation | ✅ Pass | 3.071s |  |
| Basic Echo Function | ✅ Pass | 4.182s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.511s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.356s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.030s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.630s |  |
| Search Query Function | ✅ Pass | 3.638s |  |
| Ask Advice Function | ✅ Pass | 5.281s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.999s |  |
| Penetration Testing Methodology | ✅ Pass | 8.868s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.602s |  |
| SQL Injection Attack Type | ✅ Pass | 5.462s |  |
| Penetration Testing Framework | ✅ Pass | 14.617s |  |
| Web Application Security Scanner | ✅ Pass | 10.780s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.972s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.820s

---

### pentester (deepseek-chat)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.980s |  |
| Text Transform Uppercase | ✅ Pass | 3.601s |  |
| Count from 1 to 5 | ✅ Pass | 3.835s |  |
| Math Calculation | ✅ Pass | 3.229s |  |
| Basic Echo Function | ✅ Pass | 4.034s |  |
| Streaming Simple Math Streaming | ✅ Pass | 3.092s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 3.300s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 4.760s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 4.865s |  |
| Search Query Function | ✅ Pass | 3.842s |  |
| Ask Advice Function | ✅ Pass | 4.761s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 4.847s |  |
| Penetration Testing Methodology | ✅ Pass | 7.469s |  |
| Vulnerability Assessment Tools | ✅ Pass | 9.532s |  |
| SQL Injection Attack Type | ✅ Pass | 5.392s |  |
| Penetration Testing Framework | ✅ Pass | 9.561s |  |
| Web Application Security Scanner | ✅ Pass | 9.919s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.623s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.203s

---

