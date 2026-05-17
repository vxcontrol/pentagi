import type { ComponentType } from 'react';

import { Loader2, Replace } from 'lucide-react';

import { Button } from '@/components/ui/button';

interface OverwriteButtonsProps {
    /**
     * When `true` both buttons are greyed-out and clicks are ignored. Use to
     * disable the CTAs based on form validity, selection emptiness, or any
     * other guard that doesn't carry a "request in flight" semantic.
     */
    isDisabled?: boolean;
    /**
     * When `true` both buttons are still disabled, but the primary one
     * additionally swaps its icon for a loading spinner. Wire this to the
     * `is{Verb}ing` flag from the corresponding action hook.
     */
    isProcessing?: boolean;
    onOverwrite: () => void;
    /** Optional primary action. When `primaryType="submit"` this is ignored and form submission is used instead. */
    onPrimary?: () => void;
    overwriteLabel: string;
    primaryIcon: ComponentType<{ className?: string }>;
    primaryLabel: string;
    /**
     * Render type for the primary button. Defaults to `"button"`. Pass `"submit"`
     * when the dialog uses `react-hook-form` and the primary CTA should submit
     * the form so validation runs first. The overwrite CTA is always a plain
     * button — explicit `with overwrite` is intentional and never gated by form
     * validation.
     */
    primaryType?: 'button' | 'submit';
}

/**
 * Two-CTA footer pattern shared by every overwrite-aware dialog (Pull, Attach,
 * Promote, Move, Copy):
 *
 *   [ Primary action ]  [ Action with overwrite ]
 *
 * The primary action runs the dialog's regular flow (preflight → execute with
 * `force=false`); the overwrite variant sends `force=true` straight away.
 * Visually the latter uses the destructive variant so users can spot it from
 * across the screen — overwriting is always destructive in our model.
 *
 * The component owns the spinner / icon swap, so callers don't repeat that
 * boilerplate in five different dialogs.
 */
export function OverwriteButtons({
    isDisabled,
    isProcessing,
    onOverwrite,
    onPrimary,
    overwriteLabel,
    primaryIcon: PrimaryIcon,
    primaryLabel,
    primaryType = 'button',
}: OverwriteButtonsProps) {
    const disabled = isDisabled || isProcessing;

    return (
        <>
            <Button
                disabled={disabled}
                onClick={primaryType === 'submit' ? undefined : onPrimary}
                type={primaryType}
            >
                {isProcessing ? <Loader2 className="animate-spin" /> : <PrimaryIcon />}
                {primaryLabel}
            </Button>
            <Button
                disabled={disabled}
                onClick={onOverwrite}
                type="button"
                variant="destructive"
            >
                <Replace />
                {overwriteLabel}
            </Button>
        </>
    );
}
