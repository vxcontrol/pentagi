import { Loader2 } from 'lucide-react';
import { useFormContext } from 'react-hook-form';

import { Button } from '@/components/ui/button';

interface FormSubmitButtonProps extends React.ComponentProps<typeof Button> {
    /**
     * Icon rendered to the left of `children` while the form is idle. During
     * submit it is swapped for a spinner so the button keeps its size. Omit
     * for plain text buttons — the spinner is then prepended in front of
     * `children` only while submitting.
     */
    icon?: React.ReactNode;
    /**
     * External submit-pending override. Wins over `formState.isSubmitting`
     * from the react-hook-form context. Use when the submit lives outside
     * `<FormProvider>` (attached via the HTML `form="…"` attribute) or when
     * the parent juggles multiple mutation flags that can't be reduced to a
     * single `isSubmitting`.
     */
    loading?: boolean;
    /**
     * When true (default), the button disables itself once the form has been
     * submitted with invalid values — mirrors our "click first, see errors,
     * then re-disable" UX. Pass `false` for flows that must accept further
     * submits even when validity has not yet been recomputed (e.g. forms that
     * run validation manually).
     */
    requireValid?: boolean;
}

/**
 * Submit button that subscribes to the nearest react-hook-form's
 * `formState.isSubmitting` (and `isValid` / `isSubmitted`) on its own — no
 * prop drilling. Accepts every prop the underlying `<Button>` does
 * (`variant`, `size`, `className`, `onClick`, `aria-label`, `form`, …) so it
 * works for plain text buttons, icon-only buttons, destructive actions, and
 * the `form="…"`-attached submits that sit outside the FormProvider tree.
 *
 * This is the react-hook-form-flavoured analogue of React 19's
 * `useFormStatus`: same idea (child component subscribes to the form
 * status), different source — our forms run through
 * `form.handleSubmit(...)`, not native `<form action>`.
 */
function FormSubmitButton({ children, icon, loading, requireValid = true, ...props }: FormSubmitButtonProps) {
    // `useFormContext` is safe to call outside a FormProvider — it returns
    // `null` instead of throwing. We branch on that so the same component
    // can drive both in-tree submits (read state from context) and
    // out-of-tree submits attached via `form="…"` (caller passes `loading`).
    const ctx = useFormContext();
    const formState = ctx?.formState;

    const isSubmitting = loading ?? formState?.isSubmitting ?? false;
    const isFormInvalid = !!formState && requireValid && formState.isSubmitted && !formState.isValid;
    const isDisabled = isSubmitting || isFormInvalid;

    const renderedIcon = isSubmitting ? <Loader2 className="size-4 animate-spin" /> : icon;

    return (
        <Button
            disabled={isDisabled}
            type="submit"
            {...props}
        >
            {renderedIcon}
            {children}
        </Button>
    );
}

export { FormSubmitButton };
