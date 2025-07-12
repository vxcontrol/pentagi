# LLM Agent Testing Report

Generated: Thu, 10 Jul 2025 09:36:34 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | llama3.1:8b | false | 18/18 (100.00%) | 5.178s |
| simple_json | llama3.1:8b | false | 4/4 (100.00%) | 2.506s |
| primary_agent | llama3.1:8b | false | 18/18 (100.00%) | 5.021s |
| assistant | llama3.1:8b | false | 18/18 (100.00%) | 5.302s |
| generator | llama3.1:8b | false | 18/18 (100.00%) | 5.203s |
| refiner | llama3.1:8b | false | 18/18 (100.00%) | 4.985s |
| adviser | llama3.1:8b | false | 18/18 (100.00%) | 4.544s |
| reflector | llama3.1:8b | false | 18/18 (100.00%) | 4.768s |
| searcher | llama3.1:8b | false | 18/18 (100.00%) | 4.733s |
| enricher | llama3.1:8b | false | 18/18 (100.00%) | 4.763s |
| coder | llama3.1:8b | false | 18/18 (100.00%) | 4.821s |
| installer | llama3.1:8b | false | 18/18 (100.00%) | 4.822s |
| pentester | llama3.1:8b | false | 18/18 (100.00%) | 5.123s |

**Total**: 220/220 (100.00%) successful tests
**Overall average latency**: 4.894s

## Detailed Results

### simple (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.542s |  |
| Text Transform Uppercase | ✅ Pass | 1.079s |  |
| Count from 1 to 5 | ✅ Pass | 1.442s |  |
| Math Calculation | ✅ Pass | 0.972s |  |
| Basic Echo Function | ✅ Pass | 2.354s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.235s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.322s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.411s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.725s |  |
| Search Query Function | ✅ Pass | 2.785s |  |
| Ask Advice Function | ✅ Pass | 3.908s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.607s |  |
| Penetration Testing Methodology | ✅ Pass | 7.926s |  |
| Vulnerability Assessment Tools | ✅ Pass | 14.292s |  |
| SQL Injection Attack Type | ✅ Pass | 9.341s |  |
| Penetration Testing Framework | ✅ Pass | 8.506s |  |
| Web Application Security Scanner | ✅ Pass | 16.460s |  |
| Penetration Testing Tool Selection | ✅ Pass | 10.294s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.178s

---

### simple_json (llama3.1:8b)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Person Information JSON | ✅ Pass | 2.280s |  |
| Project Information JSON | ✅ Pass | 2.225s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 2.596s |  |
| User Profile JSON | ✅ Pass | 2.923s |  |

**Summary**: 4/4 (100.00%) successful tests

**Average latency**: 2.506s

---

### primary_agent (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.448s |  |
| Text Transform Uppercase | ✅ Pass | 1.072s |  |
| Count from 1 to 5 | ✅ Pass | 1.744s |  |
| Math Calculation | ✅ Pass | 0.975s |  |
| Basic Echo Function | ✅ Pass | 2.356s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.239s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.234s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.307s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.737s |  |
| Search Query Function | ✅ Pass | 2.772s |  |
| Ask Advice Function | ✅ Pass | 3.923s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 3.607s |  |
| Penetration Testing Methodology | ✅ Pass | 6.623s |  |
| Vulnerability Assessment Tools | ✅ Pass | 12.881s |  |
| SQL Injection Attack Type | ✅ Pass | 6.433s |  |
| Penetration Testing Framework | ✅ Pass | 10.277s |  |
| Web Application Security Scanner | ✅ Pass | 17.889s |  |
| Penetration Testing Tool Selection | ✅ Pass | 9.856s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.021s

---

### assistant (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.496s |  |
| Text Transform Uppercase | ✅ Pass | 1.138s |  |
| Count from 1 to 5 | ✅ Pass | 1.330s |  |
| Math Calculation | ✅ Pass | 0.938s |  |
| Basic Echo Function | ✅ Pass | 2.848s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.817s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.987s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.921s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.999s |  |
| Search Query Function | ✅ Pass | 2.466s |  |
| Ask Advice Function | ✅ Pass | 4.561s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.828s |  |
| Penetration Testing Methodology | ✅ Pass | 10.238s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.240s |  |
| SQL Injection Attack Type | ✅ Pass | 1.078s |  |
| Penetration Testing Framework | ✅ Pass | 16.862s |  |
| Web Application Security Scanner | ✅ Pass | 16.254s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.417s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.302s

---

### generator (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 2.591s |  |
| Text Transform Uppercase | ✅ Pass | 1.133s |  |
| Count from 1 to 5 | ✅ Pass | 1.647s |  |
| Math Calculation | ✅ Pass | 0.941s |  |
| Basic Echo Function | ✅ Pass | 2.848s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.806s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.980s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.780s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.696s |  |
| Search Query Function | ✅ Pass | 2.263s |  |
| Ask Advice Function | ✅ Pass | 3.785s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 2.828s |  |
| Penetration Testing Methodology | ✅ Pass | 9.860s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.962s |  |
| SQL Injection Attack Type | ✅ Pass | 1.138s |  |
| Penetration Testing Framework | ✅ Pass | 18.316s |  |
| Web Application Security Scanner | ✅ Pass | 16.625s |  |
| Penetration Testing Tool Selection | ✅ Pass | 5.446s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.203s

---

### refiner (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.845s |  |
| Text Transform Uppercase | ✅ Pass | 0.888s |  |
| Count from 1 to 5 | ✅ Pass | 1.275s |  |
| Math Calculation | ✅ Pass | 1.018s |  |
| Basic Echo Function | ✅ Pass | 1.533s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.861s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.968s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.873s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.484s |  |
| Search Query Function | ✅ Pass | 2.040s |  |
| Ask Advice Function | ✅ Pass | 3.167s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.929s |  |
| Penetration Testing Methodology | ✅ Pass | 9.356s |  |
| Vulnerability Assessment Tools | ✅ Pass | 19.540s |  |
| SQL Injection Attack Type | ✅ Pass | 1.233s |  |
| Penetration Testing Framework | ✅ Pass | 16.387s |  |
| Web Application Security Scanner | ✅ Pass | 19.509s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.819s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.985s

---

### adviser (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.845s |  |
| Text Transform Uppercase | ✅ Pass | 0.902s |  |
| Count from 1 to 5 | ✅ Pass | 1.189s |  |
| Math Calculation | ✅ Pass | 1.032s |  |
| Basic Echo Function | ✅ Pass | 1.533s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.849s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.990s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.917s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.475s |  |
| Search Query Function | ✅ Pass | 2.029s |  |
| Ask Advice Function | ✅ Pass | 3.088s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.807s |  |
| Penetration Testing Methodology | ✅ Pass | 8.527s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.563s |  |
| SQL Injection Attack Type | ✅ Pass | 0.965s |  |
| Penetration Testing Framework | ✅ Pass | 16.038s |  |
| Web Application Security Scanner | ✅ Pass | 16.105s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.937s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.544s

---

### reflector (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.847s |  |
| Text Transform Uppercase | ✅ Pass | 0.855s |  |
| Count from 1 to 5 | ✅ Pass | 1.243s |  |
| Math Calculation | ✅ Pass | 0.890s |  |
| Basic Echo Function | ✅ Pass | 1.588s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.906s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.274s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.966s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.422s |  |
| Search Query Function | ✅ Pass | 1.905s |  |
| Ask Advice Function | ✅ Pass | 3.172s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.854s |  |
| Penetration Testing Methodology | ✅ Pass | 8.527s |  |
| Vulnerability Assessment Tools | ✅ Pass | 17.255s |  |
| SQL Injection Attack Type | ✅ Pass | 0.935s |  |
| Penetration Testing Framework | ✅ Pass | 14.785s |  |
| Web Application Security Scanner | ✅ Pass | 21.245s |  |
| Penetration Testing Tool Selection | ✅ Pass | 4.142s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.768s

---

### searcher (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.855s |  |
| Text Transform Uppercase | ✅ Pass | 0.900s |  |
| Count from 1 to 5 | ✅ Pass | 1.244s |  |
| Math Calculation | ✅ Pass | 0.801s |  |
| Basic Echo Function | ✅ Pass | 1.595s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.894s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.973s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.971s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.406s |  |
| Search Query Function | ✅ Pass | 1.894s |  |
| Ask Advice Function | ✅ Pass | 3.012s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.853s |  |
| Penetration Testing Methodology | ✅ Pass | 9.902s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.275s |  |
| SQL Injection Attack Type | ✅ Pass | 1.053s |  |
| Penetration Testing Framework | ✅ Pass | 15.739s |  |
| Web Application Security Scanner | ✅ Pass | 18.559s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.268s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.733s

---

### enricher (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.950s |  |
| Text Transform Uppercase | ✅ Pass | 0.949s |  |
| Count from 1 to 5 | ✅ Pass | 1.211s |  |
| Math Calculation | ✅ Pass | 0.834s |  |
| Basic Echo Function | ✅ Pass | 1.650s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.903s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.964s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.910s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.561s |  |
| Search Query Function | ✅ Pass | 1.933s |  |
| Ask Advice Function | ✅ Pass | 3.210s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.885s |  |
| Penetration Testing Methodology | ✅ Pass | 9.061s |  |
| Vulnerability Assessment Tools | ✅ Pass | 16.685s |  |
| SQL Injection Attack Type | ✅ Pass | 1.035s |  |
| Penetration Testing Framework | ✅ Pass | 15.353s |  |
| Web Application Security Scanner | ✅ Pass | 21.248s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.385s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.763s

---

### coder (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.852s |  |
| Text Transform Uppercase | ✅ Pass | 0.969s |  |
| Count from 1 to 5 | ✅ Pass | 1.211s |  |
| Math Calculation | ✅ Pass | 0.815s |  |
| Basic Echo Function | ✅ Pass | 1.659s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.904s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.065s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.900s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.342s |  |
| Search Query Function | ✅ Pass | 1.951s |  |
| Ask Advice Function | ✅ Pass | 3.213s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.872s |  |
| Penetration Testing Methodology | ✅ Pass | 10.494s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.283s |  |
| SQL Injection Attack Type | ✅ Pass | 1.018s |  |
| Penetration Testing Framework | ✅ Pass | 16.147s |  |
| Web Application Security Scanner | ✅ Pass | 18.765s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.312s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.821s

---

### installer (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.901s |  |
| Text Transform Uppercase | ✅ Pass | 0.979s |  |
| Count from 1 to 5 | ✅ Pass | 1.200s |  |
| Math Calculation | ✅ Pass | 0.909s |  |
| Basic Echo Function | ✅ Pass | 1.868s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.919s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.961s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.876s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 3.040s |  |
| Search Query Function | ✅ Pass | 2.012s |  |
| Ask Advice Function | ✅ Pass | 3.288s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.875s |  |
| Penetration Testing Methodology | ✅ Pass | 10.899s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.050s |  |
| SQL Injection Attack Type | ✅ Pass | 0.893s |  |
| Penetration Testing Framework | ✅ Pass | 16.285s |  |
| Web Application Security Scanner | ✅ Pass | 17.600s |  |
| Penetration Testing Tool Selection | ✅ Pass | 3.227s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 4.822s

---

### pentester (llama3.1:8b)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.893s |  |
| Text Transform Uppercase | ✅ Pass | 0.953s |  |
| Count from 1 to 5 | ✅ Pass | 1.199s |  |
| Math Calculation | ✅ Pass | 0.820s |  |
| Basic Echo Function | ✅ Pass | 1.886s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.924s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.956s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 2.180s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 2.624s |  |
| Search Query Function | ✅ Pass | 2.791s |  |
| Ask Advice Function | ✅ Pass | 3.289s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.882s |  |
| Penetration Testing Methodology | ✅ Pass | 9.101s |  |
| Vulnerability Assessment Tools | ✅ Pass | 18.899s |  |
| SQL Injection Attack Type | ✅ Pass | 0.875s |  |
| Penetration Testing Framework | ✅ Pass | 17.480s |  |
| Web Application Security Scanner | ✅ Pass | 22.453s |  |
| Penetration Testing Tool Selection | ✅ Pass | 2.993s |  |

**Summary**: 18/18 (100.00%) successful tests

**Average latency**: 5.123s

---

