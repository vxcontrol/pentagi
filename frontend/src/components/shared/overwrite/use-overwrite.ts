import { useCallback, useEffect, useRef, useState } from 'react';

import type { OverwriteConflict } from './overwrite-dialog';

/**
 * Discriminated outcome of a server action that supports an overwrite flag.
 * Hooks wrapping REST calls return this so the orchestrator can branch
 * between success, a 409 that warrants a user prompt, and any other error
 * (which the hook is expected to surface as a toast on its own).
 *
 * `conflict` carries an optional `conflicts` array — callers that fan out
 * a batch into N parallel requests (e.g. multi-promote) can attach precise
 * per-item descriptors here. When omitted, the hook falls back to
 * `synthesizeFallbackConflicts` and finally to an anonymous descriptor so
 * the dialog still opens with the count-based copy.
 */
export type OverwriteOutcome =
    | { conflicts?: OverwriteConflict[]; kind: 'conflict' }
    | { kind: 'error' }
    | { kind: 'ok' };

/**
 * Anonymous fallback descriptor used when a 409 sneaks through after a clean
 * preflight and the caller didn't provide `synthesizeFallbackConflicts`. Falls
 * back to the count-based copy ("N items already exist...") in
 * `OverwriteDialog`.
 */
const ANONYMOUS_FALLBACK_CONFLICT: OverwriteConflict = {
    destination: '',
    destinationName: 'an item',
};

interface UseOverwriteOptions<TPlan> {
    /**
     * Execute the REST call. Receives the plan + a boolean `force` flag.
     * Should return a discriminated outcome — see `OverwriteOutcome`.
     */
    execute: (plan: TPlan, force: boolean) => Promise<OverwriteOutcome>;
    /**
     * Pure function: inspect the local snapshot and return any destinations
     * that would conflict. Empty array → primary execute proceeds with
     * `force=false`; non-empty → the OverwriteDialog is opened
     * pre-populated with these descriptors.
     */
    findConflicts: (plan: TPlan) => OverwriteConflict[];
    /** Fired only when `execute` resolves with `kind: 'ok'`. */
    onSuccess?: () => void;
    /**
     * Optional. Synthesizes conflict descriptors for the race-fallback case:
     * preflight returned [] but the server still answered 409. Defaults to a
     * single anonymous descriptor, which makes the dialog show the count-based
     * copy ("Some items already exist…") instead of a per-item name.
     */
    synthesizeFallbackConflicts?: (plan: TPlan) => OverwriteConflict[];
}

interface UseOverwriteResult<TPlan> {
    /** Live conflict descriptors. Wire to `<OverwriteDialog conflicts={…} />`. */
    conflicts: OverwriteConflict[];
    /**
     * Execute the action with `force=true` immediately, bypassing the
     * preflight and the conflict prompt. Wire to the secondary CTA.
     */
    forceExecute: (plan: TPlan) => Promise<void>;
    /** Wire to the `onReplaceAll` handler of `<OverwriteDialog />`. */
    handleReplaceAll: () => Promise<void>;
    /**
     * Execute the action with the preflight + race-fallback workflow. Wire
     * to the primary CTA.
     */
    primaryExecute: (plan: TPlan) => Promise<void>;
    /** Wire to the `onCancel` handler of `<OverwriteDialog />`. */
    resetConflicts: () => void;
}

/**
 * Orchestrates the canonical "primary CTA / `… with overwrite` CTA / Replace
 * all" workflow shared by the Pull, Attach, Promote, Move and Copy dialogs.
 *
 * Workflow:
 *   1. The user clicks the **primary CTA** → `primaryExecute(plan)` runs.
 *      `findConflicts` is consulted on the local snapshot first; if anything
 *      collides, the OverwriteDialog opens with those descriptors.
 *      Otherwise `execute(plan, false)` is dispatched. A 409 from the server
 *      (race) auto-opens the dialog with `synthesizeFallbackConflicts` (or an
 *      anonymous fallback).
 *   2. The user clicks the **`… with overwrite` CTA** → `forceExecute(plan)`
 *      runs `execute(plan, true)` straight away — no preflight, no prompt.
 *   3. Inside the conflict dialog, **Replace all** → `handleReplaceAll()`
 *      retries the cached plan with `force=true`. **Cancel** →
 *      `resetConflicts()` only closes the prompt; the parent dialog stays
 *      open so the user can change the destination and re-submit.
 *
 * Callbacks passed in `options` are read through a ref, so callers don't need
 * to wrap them in `useCallback`. This keeps the hook ergonomic at the call
 * site without sacrificing reference stability for the returned actions.
 */
export function useOverwrite<TPlan>(options: UseOverwriteOptions<TPlan>): UseOverwriteResult<TPlan> {
    const [conflicts, setConflicts] = useState<OverwriteConflict[]>([]);
    const [pendingPlan, setPendingPlan] = useState<null | TPlan>(null);

    // Read the latest options through a ref so callers can pass inline
    // arrow functions without re-creating the action callbacks every render.
    const optionsRef = useRef(options);

    useEffect(() => {
        optionsRef.current = options;
    }, [options]);

    const settle = useCallback(async (plan: TPlan, force: boolean): Promise<OverwriteOutcome> => {
        const outcome = await optionsRef.current.execute(plan, force);

        if (outcome.kind === 'ok') {
            setConflicts([]);
            setPendingPlan(null);
            optionsRef.current.onSuccess?.();
        }

        return outcome;
    }, []);

    const primaryExecute = useCallback(
        async (plan: TPlan): Promise<void> => {
            const detected = optionsRef.current.findConflicts(plan);

            if (detected.length > 0) {
                setConflicts(detected);
                setPendingPlan(plan);

                return;
            }

            const outcome = await settle(plan, false);

            if (outcome.kind !== 'conflict') {
                return;
            }

            // Resolution chain (most → least specific):
            //   1. precise descriptors returned by `execute` itself (e.g.
            //      multi-promote that fanned the batch into N parallel calls
            //      and knows exactly which ones came back as 409),
            //   2. caller-supplied synthesizer derived from the plan,
            //   3. an anonymous descriptor that triggers the dialog's
            //      count-based copy.
            const fromOutcome = outcome.conflicts ?? [];
            const synthesized = optionsRef.current.synthesizeFallbackConflicts?.(plan) ?? [];

            let final: OverwriteConflict[];

            if (fromOutcome.length > 0) {
                final = fromOutcome;
            } else if (synthesized.length > 0) {
                final = synthesized;
            } else {
                final = [ANONYMOUS_FALLBACK_CONFLICT];
            }

            setConflicts(final);
            setPendingPlan(plan);
        },
        [settle],
    );

    const forceExecute = useCallback(
        async (plan: TPlan): Promise<void> => {
            await settle(plan, true);
        },
        [settle],
    );

    const handleReplaceAll = useCallback(async (): Promise<void> => {
        if (pendingPlan === null) {
            return;
        }

        await settle(pendingPlan, true);
    }, [pendingPlan, settle]);

    const resetConflicts = useCallback(() => {
        setConflicts([]);
        setPendingPlan(null);
    }, []);

    return {
        conflicts,
        forceExecute,
        handleReplaceAll,
        primaryExecute,
        resetConflicts,
    };
}
