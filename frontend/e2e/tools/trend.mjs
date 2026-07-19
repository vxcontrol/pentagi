// Turns a Playwright JSON report into one trend record so a green/red gate also
// surfaces slow regressions: duration p95, per-spec outliers, and flaky/failed
// counts. Appends JSONL to the trend file (an artifact retained across runs).
//
// Usage: node trend.mjs [results.json] [trend.jsonl]
//   E2E_TREND_SHA / E2E_TREND_REF   optional provenance stamped into the record
import { appendFileSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const RESULTS = process.argv[2] ?? fileURLToPath(new URL('../test-results/results.json', import.meta.url));
const TREND = process.argv[3] ?? fileURLToPath(new URL('../test-results/trend.jsonl', import.meta.url));

const percentile = (values, p) => {
    if (!values.length) {
        return 0;
    }

    const sorted = [...values].sort((a, b) => a - b);
    const index = Math.min(sorted.length - 1, Math.ceil((p / 100) * sorted.length) - 1);

    return sorted[index];
};

const collectSpecs = (report) => {
    const specs = [];
    const walk = (suite) => {
        suite.suites?.forEach(walk);
        suite.specs?.forEach((spec) => {
            const ms = Math.max(0, ...(spec.tests ?? []).flatMap((test) => (test.results ?? []).map((r) => r.duration)));

            specs.push({ ms, title: spec.title });
        });
    };

    report.suites?.forEach(walk);

    return specs;
};

const report = JSON.parse(readFileSync(RESULTS, 'utf8'));
const specs = collectSpecs(report);
const durations = specs.map((spec) => spec.ms);
const { expected = 0, flaky = 0, skipped = 0, unexpected = 0 } = report.stats ?? {};

const record = {
    // A caller-provided timestamp keeps this pure (no Date.now in a determinism-
    // sensitive tool); falls back to the report's own start time.
    at: process.env.E2E_TREND_AT ?? report.stats?.startTime ?? '',
    durationTotalMs: Math.round(report.stats?.duration ?? 0),
    failed: unexpected,
    flaky,
    p50Ms: percentile(durations, 50),
    p95Ms: percentile(durations, 95),
    passed: expected,
    ref: process.env.E2E_TREND_REF ?? '',
    sha: process.env.E2E_TREND_SHA ?? '',
    skipped,
    slowest: [...specs].sort((a, b) => b.ms - a.ms).slice(0, 3),
    specCount: specs.length,
};

appendFileSync(TREND, `${JSON.stringify(record)}\n`);
console.log(
    `trend: ${record.specCount} specs, p50 ${record.p50Ms}ms, p95 ${record.p95Ms}ms, ` +
        `${record.passed} passed / ${record.failed} failed / ${record.flaky} flaky`,
);
