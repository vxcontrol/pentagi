# LLM Agent Testing Report

Generated: Thu, 23 Jul 2026 11:03:48 UTC

## Overall Results

| Agent | Model | Reasoning | Success Rate | Average Latency |
|-------|-------|-----------|--------------|-----------------|
| simple | Qwen/Qwen3.6-27B-FP8 | false | 24/24 (100.00%) | 0.244s |
| simple_json | Qwen/Qwen3.6-27B-FP8 | false | 7/7 (100.00%) | 0.224s |
| primary_agent | DeepSeek-V4-Flash | true | 24/24 (100.00%) | 1.671s |
| assistant | DeepSeek-V4-Flash | true | 24/24 (100.00%) | 1.627s |
| generator | DeepSeek-V4-Flash | true | 24/24 (100.00%) | 1.526s |
| refiner | DeepSeek-V4-Flash | true | 24/24 (100.00%) | 1.381s |
| adviser | DeepSeek-V4-Flash | true | 24/24 (100.00%) | 1.255s |
| reflector | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.243s |
| searcher | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.238s |
| enricher | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.231s |
| coder | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.233s |
| installer | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.238s |
| pentester | Qwen/Qwen3.6-27B-FP8 | true | 24/24 (100.00%) | 0.239s |

**Total**: 295/295 (100.00%) successful tests
**Overall average latency**: 0.748s

## Detailed Results

### simple (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.265s |  |
| Text Transform Uppercase | ✅ Pass | 0.229s |  |
| Count from 1 to 5 | ✅ Pass | 0.207s |  |
| Math Calculation | ✅ Pass | 0.221s |  |
| Basic Echo Function | ✅ Pass | 0.251s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.224s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.233s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.215s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.218s |  |
| Search Query Function | ✅ Pass | 0.224s |  |
| Ask Advice Function | ✅ Pass | 0.308s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.234s |  |
| Basic Context Memory Test | ✅ Pass | 0.207s |  |
| Function Argument Memory Test | ✅ Pass | 0.295s |  |
| Function Response Memory Test | ✅ Pass | 0.337s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.224s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.214s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.451s |  |
| Penetration Testing Methodology | ✅ Pass | 0.222s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.214s |  |
| SQL Injection Attack Type | ✅ Pass | 0.220s |  |
| Penetration Testing Framework | ✅ Pass | 0.206s |  |
| Web Application Security Scanner | ✅ Pass | 0.216s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.219s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.244s

---

### simple_json (Qwen/Qwen3.6-27B-FP8)

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Vulnerability Report Memory Test | ✅ Pass | 0.213s |  |
| Person Information JSON | ✅ Pass | 0.208s |  |
| Project Information JSON | ✅ Pass | 0.213s |  |
| User Profile JSON | ✅ Pass | 0.223s |  |
| JSON Array Response Without Schema | ✅ Pass | 0.217s |  |
| Streaming Person Information JSON Streaming | ✅ Pass | 0.274s |  |

#### Capability Tests

| Test | Capability | Result | Latency | Note |
|------|------------|--------|---------|------|
| Structured Output With JSON Schema | structured_output | ✅ Pass | 0.214s |  |

**Summary**: 7/7 (100.00%) successful tests

**Average latency**: 0.224s

---

### primary_agent (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.238s |  |
| Text Transform Uppercase | ✅ Pass | 1.238s |  |
| Count from 1 to 5 | ✅ Pass | 1.104s |  |
| Math Calculation | ✅ Pass | 1.233s |  |
| Basic Echo Function | ✅ Pass | 1.589s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.987s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.189s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.393s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.553s |  |
| Search Query Function | ✅ Pass | 1.386s |  |
| Ask Advice Function | ✅ Pass | 2.022s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.370s |  |
| Basic Context Memory Test | ✅ Pass | 1.393s |  |
| Function Argument Memory Test | ✅ Pass | 1.112s |  |
| Function Response Memory Test | ✅ Pass | 1.150s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.412s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.437s |  |
| Penetration Testing Methodology | ✅ Pass | 1.728s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.727s |  |
| SQL Injection Attack Type | ✅ Pass | 1.461s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.942s |  |
| Penetration Testing Framework | ✅ Pass | 2.063s |  |
| Web Application Security Scanner | ✅ Pass | 1.749s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.612s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.671s

---

### assistant (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.920s |  |
| Text Transform Uppercase | ✅ Pass | 1.028s |  |
| Math Calculation | ✅ Pass | 1.201s |  |
| Count from 1 to 5 | ✅ Pass | 1.935s |  |
| Basic Echo Function | ✅ Pass | 1.375s |  |
| Streaming Simple Math Streaming | ✅ Pass | 1.001s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.119s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.356s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.582s |  |
| Search Query Function | ✅ Pass | 1.358s |  |
| Ask Advice Function | ✅ Pass | 1.966s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.351s |  |
| Basic Context Memory Test | ✅ Pass | 1.114s |  |
| Function Argument Memory Test | ✅ Pass | 1.052s |  |
| Function Response Memory Test | ✅ Pass | 1.238s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.392s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.506s |  |
| Penetration Testing Methodology | ✅ Pass | 1.316s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.774s |  |
| SQL Injection Attack Type | ✅ Pass | 1.488s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.653s |  |
| Web Application Security Scanner | ✅ Pass | 1.626s |  |
| Penetration Testing Framework | ✅ Pass | 3.101s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.589s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.627s

---

### generator (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.828s |  |
| Text Transform Uppercase | ✅ Pass | 1.088s |  |
| Count from 1 to 5 | ✅ Pass | 1.407s |  |
| Math Calculation | ✅ Pass | 0.976s |  |
| Basic Echo Function | ✅ Pass | 1.545s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.875s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.162s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.274s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.400s |  |
| Search Query Function | ✅ Pass | 1.342s |  |
| Ask Advice Function | ✅ Pass | 1.860s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.190s |  |
| Basic Context Memory Test | ✅ Pass | 1.152s |  |
| Function Argument Memory Test | ✅ Pass | 0.959s |  |
| Function Response Memory Test | ✅ Pass | 0.881s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 1.242s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.492s |  |
| Penetration Testing Methodology | ✅ Pass | 1.616s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.804s |  |
| SQL Injection Attack Type | ✅ Pass | 0.966s |  |
| Vulnerability Assessment Tools | ✅ Pass | 3.905s |  |
| Penetration Testing Framework | ✅ Pass | 1.841s |  |
| Web Application Security Scanner | ✅ Pass | 1.273s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.538s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.526s

---

### refiner (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 1.224s |  |
| Text Transform Uppercase | ✅ Pass | 0.901s |  |
| Count from 1 to 5 | ✅ Pass | 1.372s |  |
| Math Calculation | ✅ Pass | 0.874s |  |
| Basic Echo Function | ✅ Pass | 1.513s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.725s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.123s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.377s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.472s |  |
| Search Query Function | ✅ Pass | 1.154s |  |
| Ask Advice Function | ✅ Pass | 1.727s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 1.455s |  |
| Basic Context Memory Test | ✅ Pass | 1.065s |  |
| Function Argument Memory Test | ✅ Pass | 0.940s |  |
| Function Response Memory Test | ✅ Pass | 0.840s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.239s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.333s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.754s |  |
| Penetration Testing Methodology | ✅ Pass | 1.504s |  |
| SQL Injection Attack Type | ✅ Pass | 0.221s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.540s |  |
| Penetration Testing Framework | ✅ Pass | 2.092s |  |
| Web Application Security Scanner | ✅ Pass | 1.218s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.460s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.381s

---

### adviser (DeepSeek-V4-Flash)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.874s |  |
| Text Transform Uppercase | ✅ Pass | 0.964s |  |
| Count from 1 to 5 | ✅ Pass | 1.183s |  |
| Math Calculation | ✅ Pass | 0.690s |  |
| Basic Echo Function | ✅ Pass | 1.489s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.249s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 1.094s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 1.640s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 1.782s |  |
| Search Query Function | ✅ Pass | 1.120s |  |
| Ask Advice Function | ✅ Pass | 1.737s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.208s |  |
| Basic Context Memory Test | ✅ Pass | 1.212s |  |
| Function Argument Memory Test | ✅ Pass | 0.972s |  |
| Function Response Memory Test | ✅ Pass | 0.770s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.241s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 2.306s |  |
| Penetration Testing Methodology | ✅ Pass | 0.221s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 3.772s |  |
| SQL Injection Attack Type | ✅ Pass | 0.212s |  |
| Vulnerability Assessment Tools | ✅ Pass | 2.754s |  |
| Penetration Testing Framework | ✅ Pass | 2.173s |  |
| Web Application Security Scanner | ✅ Pass | 1.074s |  |
| Penetration Testing Tool Selection | ✅ Pass | 1.367s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 1.255s

---

### reflector (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.257s |  |
| Text Transform Uppercase | ✅ Pass | 0.211s |  |
| Count from 1 to 5 | ✅ Pass | 0.331s |  |
| Math Calculation | ✅ Pass | 0.231s |  |
| Basic Echo Function | ✅ Pass | 0.219s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.250s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.219s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.279s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.291s |  |
| Search Query Function | ✅ Pass | 0.216s |  |
| Ask Advice Function | ✅ Pass | 0.212s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.208s |  |
| Basic Context Memory Test | ✅ Pass | 0.215s |  |
| Function Argument Memory Test | ✅ Pass | 0.207s |  |
| Function Response Memory Test | ✅ Pass | 0.224s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.228s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.219s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.468s |  |
| Penetration Testing Methodology | ✅ Pass | 0.218s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.204s |  |
| SQL Injection Attack Type | ✅ Pass | 0.209s |  |
| Penetration Testing Framework | ✅ Pass | 0.224s |  |
| Web Application Security Scanner | ✅ Pass | 0.214s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.267s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.243s

---

### searcher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.262s |  |
| Text Transform Uppercase | ✅ Pass | 0.216s |  |
| Count from 1 to 5 | ✅ Pass | 0.289s |  |
| Math Calculation | ✅ Pass | 0.217s |  |
| Basic Echo Function | ✅ Pass | 0.214s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.223s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.229s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.287s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.224s |  |
| Search Query Function | ✅ Pass | 0.237s |  |
| Ask Advice Function | ✅ Pass | 0.232s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.230s |  |
| Basic Context Memory Test | ✅ Pass | 0.213s |  |
| Function Argument Memory Test | ✅ Pass | 0.210s |  |
| Function Response Memory Test | ✅ Pass | 0.217s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.225s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.224s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.436s |  |
| Penetration Testing Methodology | ✅ Pass | 0.218s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.207s |  |
| SQL Injection Attack Type | ✅ Pass | 0.212s |  |
| Penetration Testing Framework | ✅ Pass | 0.215s |  |
| Web Application Security Scanner | ✅ Pass | 0.230s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.244s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.238s

---

### enricher (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.217s |  |
| Text Transform Uppercase | ✅ Pass | 0.212s |  |
| Count from 1 to 5 | ✅ Pass | 0.248s |  |
| Math Calculation | ✅ Pass | 0.249s |  |
| Basic Echo Function | ✅ Pass | 0.211s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.240s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.225s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.280s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.250s |  |
| Search Query Function | ✅ Pass | 0.209s |  |
| Ask Advice Function | ✅ Pass | 0.214s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.215s |  |
| Basic Context Memory Test | ✅ Pass | 0.206s |  |
| Function Argument Memory Test | ✅ Pass | 0.205s |  |
| Function Response Memory Test | ✅ Pass | 0.204s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.210s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.212s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.429s |  |
| Penetration Testing Methodology | ✅ Pass | 0.209s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.207s |  |
| SQL Injection Attack Type | ✅ Pass | 0.209s |  |
| Penetration Testing Framework | ✅ Pass | 0.218s |  |
| Web Application Security Scanner | ✅ Pass | 0.216s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.242s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.231s

---

### coder (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.215s |  |
| Text Transform Uppercase | ✅ Pass | 0.210s |  |
| Count from 1 to 5 | ✅ Pass | 0.229s |  |
| Math Calculation | ✅ Pass | 0.241s |  |
| Basic Echo Function | ✅ Pass | 0.219s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.243s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.212s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.265s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.225s |  |
| Search Query Function | ✅ Pass | 0.219s |  |
| Ask Advice Function | ✅ Pass | 0.211s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.209s |  |
| Basic Context Memory Test | ✅ Pass | 0.211s |  |
| Function Argument Memory Test | ✅ Pass | 0.210s |  |
| Function Response Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.267s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.226s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.430s |  |
| Penetration Testing Methodology | ✅ Pass | 0.212s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.214s |  |
| SQL Injection Attack Type | ✅ Pass | 0.213s |  |
| Penetration Testing Framework | ✅ Pass | 0.256s |  |
| Web Application Security Scanner | ✅ Pass | 0.217s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.222s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.233s

---

### installer (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.225s |  |
| Text Transform Uppercase | ✅ Pass | 0.212s |  |
| Count from 1 to 5 | ✅ Pass | 0.232s |  |
| Math Calculation | ✅ Pass | 0.215s |  |
| Basic Echo Function | ✅ Pass | 0.207s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.217s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.230s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.269s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.222s |  |
| Search Query Function | ✅ Pass | 0.208s |  |
| Ask Advice Function | ✅ Pass | 0.204s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.235s |  |
| Basic Context Memory Test | ✅ Pass | 0.429s |  |
| Function Argument Memory Test | ✅ Pass | 0.215s |  |
| Function Response Memory Test | ✅ Pass | 0.207s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.217s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.226s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.423s |  |
| Penetration Testing Methodology | ✅ Pass | 0.217s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.212s |  |
| SQL Injection Attack Type | ✅ Pass | 0.214s |  |
| Penetration Testing Framework | ✅ Pass | 0.217s |  |
| Web Application Security Scanner | ✅ Pass | 0.222s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.234s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.238s

---

### pentester (Qwen/Qwen3.6-27B-FP8)

#### Basic Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| Simple Math | ✅ Pass | 0.212s |  |
| Text Transform Uppercase | ✅ Pass | 0.207s |  |
| Count from 1 to 5 | ✅ Pass | 0.218s |  |
| Math Calculation | ✅ Pass | 0.235s |  |
| Basic Echo Function | ✅ Pass | 0.217s |  |
| Streaming Simple Math Streaming | ✅ Pass | 0.220s |  |
| Streaming Count from 1 to 3 Streaming | ✅ Pass | 0.208s |  |
| Streaming Basic Echo Function Streaming | ✅ Pass | 0.268s |  |

#### Advanced Tests

| Test | Result | Latency | Error |
|------|--------|---------|-------|
| JSON Response Function | ✅ Pass | 0.214s |  |
| Search Query Function | ✅ Pass | 0.293s |  |
| Ask Advice Function | ✅ Pass | 0.219s |  |
| Streaming Search Query Function Streaming | ✅ Pass | 0.222s |  |
| Basic Context Memory Test | ✅ Pass | 0.417s |  |
| Function Argument Memory Test | ✅ Pass | 0.209s |  |
| Function Response Memory Test | ✅ Pass | 0.210s |  |
| Penetration Testing Memory with Tool Call | ✅ Pass | 0.214s |  |
| Cybersecurity Workflow Memory Test | ✅ Pass | 0.219s |  |
| Read a file, then edit it via unified diff | ✅ Pass | 0.406s |  |
| Penetration Testing Methodology | ✅ Pass | 0.222s |  |
| Vulnerability Assessment Tools | ✅ Pass | 0.215s |  |
| SQL Injection Attack Type | ✅ Pass | 0.215s |  |
| Penetration Testing Framework | ✅ Pass | 0.217s |  |
| Web Application Security Scanner | ✅ Pass | 0.220s |  |
| Penetration Testing Tool Selection | ✅ Pass | 0.233s |  |

**Summary**: 24/24 (100.00%) successful tests

**Average latency**: 0.239s

---

