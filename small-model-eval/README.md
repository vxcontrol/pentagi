# Small-model evaluation harness

This directory holds the measurement setup for tuning PentAGI with small local models (point 9 of the tuning guide). With small models the gap between a good and a bad configuration is far wider than with large ones, so every change in `SMALL_MODEL_*` or `SUMMARIZER_*` is validated here before it is trusted.

## What to measure

The metrics that matter for a pentest agent, in priority order:

- **Task completion.** Did the agent find the vulnerability / identify the port / retrieve the file the scenario defines as success?
- **Fact fidelity after compaction.** The single most useful metric for comparing two summarization strategies: ask the agent about details that live in sections the summarizer has already compacted ("which port was open?", "which Tomcat version?") and check the answers. It measures directly what compaction is costing you.
- **Malformed tool-call rate.** Fraction of tool calls that failed schema validation before the sanitize/repair path fixed them.
- **Steps to result.** Fewer is better; a rising step count is an early sign the model is losing the thread.
- **Out-of-scope actions blocked.** Count of actions the verifier stopped. This should be zero in a healthy run and non-zero only when a scenario deliberately tempts the agent out of scope.

## How to run

1. Prepare repeatable targets: known-vulnerable training machines or prepared containers, each with a clear expected outcome. Put one scenario file per target under `scenarios/`.
2. Set a baseline profile in `.env` (start from the values in `../backend/docs/config.md` plus `SMALL_MODEL_MODE=true`).
3. Run each scenario against a fresh flow and record the metrics above.
4. **Change one variable at a time** — compaction threshold, state format, number of exposed tools, the model used for summaries — then re-measure and compare. Two changes at once tell you nothing about which one acted.

## Finding the real context limit

Run the same scenario with `SMALL_MODEL_CTX_BUDGET_PERCENT` at 25, 40, and 60. The point where the success rate drops sharply is the model's real usable limit; stay one step below it.

## Scenario format

See `scenarios/example.yaml`. Each scenario names the target, the authorized scope (which becomes `SMALL_MODEL_SCOPE`), the success check, and the fact-fidelity questions with their expected answers.
